//go:build integration

package integration

import (
	"testing"

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
