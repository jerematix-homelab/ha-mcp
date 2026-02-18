//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type BlueprintIntegrationTestSuite struct {
	IntegrationTestSuite
}

func TestBlueprintIntegration(t *testing.T) {
	suite.Run(t, new(BlueprintIntegrationTestSuite))
}

// TestListAutomationBlueprints verifies that we can list automation blueprints.
func (s *BlueprintIntegrationTestSuite) TestListAutomationBlueprints() {
	// Call blueprint/list WebSocket command
	response, err := s.Client().SendHACSCommand(s.Context(), "blueprint/list", map[string]any{
		"domain": "automation",
	})
	s.Require().NoError(err, "Failed to list automation blueprints")
	s.NotNil(response, "Response should not be nil")

	s.T().Log("Successfully listed automation blueprints")

	// Response should be a map of blueprint paths to metadata
	blueprintsMap, ok := response.(map[string]any)
	if ok && len(blueprintsMap) > 0 {
		s.T().Logf("Found %d automation blueprint(s)", len(blueprintsMap))
	} else {
		s.T().Log("No automation blueprints found (this is normal if none are imported)")
	}
}

// TestListScriptBlueprints verifies that we can list script blueprints.
func (s *BlueprintIntegrationTestSuite) TestListScriptBlueprints() {
	// Call blueprint/list WebSocket command
	response, err := s.Client().SendHACSCommand(s.Context(), "blueprint/list", map[string]any{
		"domain": "script",
	})
	s.Require().NoError(err, "Failed to list script blueprints")
	s.NotNil(response, "Response should not be nil")

	s.T().Log("Successfully listed script blueprints")

	// Response should be a map of blueprint paths to metadata
	blueprintsMap, ok := response.(map[string]any)
	if ok && len(blueprintsMap) > 0 {
		s.T().Logf("Found %d script blueprint(s)", len(blueprintsMap))
	} else {
		s.T().Log("No script blueprints found (this is normal if none are imported)")
	}
}
