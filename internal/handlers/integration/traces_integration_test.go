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
