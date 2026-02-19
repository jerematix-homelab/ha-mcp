//go:build integration

package integration

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TraceIntegrationTestSuite struct {
	IntegrationTestSuite
}

func TestTraceIntegration(t *testing.T) {
	suite.Run(t, new(TraceIntegrationTestSuite))
}

// TestListAutomationTraces verifies that we can list automation traces.
func (s *TraceIntegrationTestSuite) TestListAutomationTraces() {
	// Call trace/list WebSocket command for automations
	response, err := s.Client().SendHACSCommand(s.Context(), "trace/list", map[string]any{
		"domain": "automation",
	})
	s.Require().NoError(err, "Failed to list automation traces")
	s.NotNil(response, "Response should not be nil")

	s.T().Log("Successfully listed automation traces")

	// Note: We don't assert on trace count as it depends on automation executions
	// This test verifies the API integration works correctly
}

// TestListScriptTraces verifies that we can list script traces.
func (s *TraceIntegrationTestSuite) TestListScriptTraces() {
	// Call trace/list WebSocket command for scripts
	response, err := s.Client().SendHACSCommand(s.Context(), "trace/list", map[string]any{
		"domain": "script",
	})
	s.Require().NoError(err, "Failed to list script traces")
	s.NotNil(response, "Response should not be nil")

	s.T().Log("Successfully listed script traces")
}

// TestListAllTraces verifies listing traces for both domains.
func (s *TraceIntegrationTestSuite) TestListAllTraces() {
	// Note: trace/list requires domain parameter, so test both domains separately

	// List automation traces
	response, err := s.Client().SendHACSCommand(s.Context(), "trace/list", map[string]any{
		"domain": "automation",
	})
	s.Require().NoError(err, "Failed to list automation traces")
	s.NotNil(response, "Automation traces response should not be nil")

	// List script traces
	response, err = s.Client().SendHACSCommand(s.Context(), "trace/list", map[string]any{
		"domain": "script",
	})
	s.Require().NoError(err, "Failed to list script traces")
	s.NotNil(response, "Script traces response should not be nil")

	s.T().Log("Successfully listed traces for both automation and script domains")
}

// TestGetTrace verifies that we can get a specific trace by item_id and run_id.
// It lists automation traces first to find a valid run_id and item_id, then
// calls trace/get. The test is skipped gracefully if no traces are available.
func (s *TraceIntegrationTestSuite) TestGetTrace() {
	// List automation traces to find a valid item_id and run_id
	response, err := s.Client().SendHACSCommand(s.Context(), "trace/list", map[string]any{
		"domain": "automation",
	})
	s.Require().NoError(err, "Failed to list automation traces")

	// Parse response as a list of traces
	traces, ok := response.([]any)
	if !ok || len(traces) == 0 {
		s.T().Skip("Skipping TestGetTrace: no automation traces available")
		return
	}

	// Extract item_id and run_id from the first trace
	firstTrace, ok := traces[0].(map[string]any)
	if !ok {
		s.T().Skip("Skipping TestGetTrace: invalid trace format")
		return
	}

	itemID, _ := firstTrace["item_id"].(string)
	runID, _ := firstTrace["run_id"].(string)

	if itemID == "" || runID == "" {
		s.T().Skip("Skipping TestGetTrace: trace missing item_id or run_id")
		return
	}

	s.T().Logf("Getting trace for item_id=%s, run_id=%s", itemID, runID)

	// Call trace/get with the extracted item_id and run_id
	traceResponse, err := s.Client().SendHACSCommand(s.Context(), "trace/get", map[string]any{
		"domain":  "automation",
		"item_id": itemID,
		"run_id":  runID,
	})
	s.Require().NoError(err, "Failed to get automation trace")
	s.NotNil(traceResponse, "Trace response should not be nil")

	s.T().Logf("Successfully retrieved trace for %s", itemID)
}

// TestDebugAutomation verifies that all 4 API calls used by the debug action work against real HA.
// It finds an existing automation, then exercises GetAutomation, trace/list, GetState, and GetLogbook.
func (s *TraceIntegrationTestSuite) TestDebugAutomation() {
	// Find an existing automation to debug
	automations, err := s.Client().ListAutomations(s.Context())
	s.Require().NoError(err, "Failed to list automations")

	if len(automations) == 0 {
		s.T().Skip("Skipping TestDebugAutomation: no automations available")
		return
	}

	// Pick the first automation
	testAuto := automations[0]
	entityID := testAuto.EntityID
	configID := strings.TrimPrefix(entityID, "automation.")

	s.T().Logf("Debugging automation: %s", entityID)

	// 1. Verify GetAutomation works (fetches config, triggers, mode, etc.)
	fullAuto, err := s.Client().GetAutomation(s.Context(), configID)
	s.Require().NoError(err, "GetAutomation should succeed")
	s.Require().NotNil(fullAuto, "GetAutomation should return automation")
	s.T().Logf("Automation state: %s, config: %v", fullAuto.State, fullAuto.Config != nil)

	// 2. Verify trace/list works for this automation
	traceResponse, err := s.Client().SendHACSCommand(s.Context(), "trace/list", map[string]any{
		"domain":  "automation",
		"item_id": entityID,
	})
	s.Require().NoError(err, "trace/list should succeed")
	s.NotNil(traceResponse, "trace/list response should not be nil")
	s.T().Logf("trace/list response type: %T", traceResponse)

	// 3. Verify GetState works (for trigger entity states)
	entityState, err := s.Client().GetState(s.Context(), entityID)
	s.Require().NoError(err, "GetState for automation should succeed")
	s.NotNil(entityState, "Entity state should not be nil")
	s.T().Logf("Automation entity state: %s", entityState.State)

	// 4. Verify GetLogbook works for the automation
	now := time.Now()
	startTime := now.Add(-6 * time.Hour)
	logbook, err := s.Client().GetLogbook(s.Context(), startTime.Format(time.RFC3339), now.Format(time.RFC3339), entityID)
	s.Require().NoError(err, "GetLogbook should succeed")
	s.T().Logf("Found %d logbook entries for %s in last 6 hours", len(logbook), entityID)

	s.T().Log("All 4 debug action API calls succeeded")
}
