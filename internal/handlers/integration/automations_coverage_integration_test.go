//go:build integration

package integration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AutomationCoverageIntegrationTestSuite struct {
	AutomationTestSuite
}

func TestAutomationCoverageIntegration(t *testing.T) {
	suite.Run(t, new(AutomationCoverageIntegrationTestSuite))
}

func (s *AutomationCoverageIntegrationTestSuite) TestCoverageAnalysisDataFetch() {
	// This test verifies that all API calls needed for coverage analysis work correctly
	// Coverage analysis is a handler-level feature, so we just verify data fetching

	// Verify we can fetch automations
	automations, err := s.Client().ListAutomations(s.Context())
	s.Require().NoError(err, "Should be able to list automations")
	s.T().Logf("Found %d automations", len(automations))

	// Verify we can fetch all states
	states, err := s.Client().GetStates(s.Context())
	s.Require().NoError(err, "Should be able to get states")
	s.T().Logf("Found %d entities", len(states))

	// Verify we can fetch area registry
	areas, err := s.Client().GetAreaRegistry(s.Context())
	s.Require().NoError(err, "Should be able to get area registry")
	s.T().Logf("Found %d areas", len(areas))

	// Verify we can fetch entity registry
	entityRegistry, err := s.Client().GetEntityRegistry(s.Context())
	s.Require().NoError(err, "Should be able to get entity registry")
	s.T().Logf("Found %d entity registry entries", len(entityRegistry))

	// If we have automations, verify we can get their full configs
	if len(automations) > 0 {
		testAutomation := automations[0]
		// Extract ID from entity_id (remove "automation." prefix)
		automationID := testAutomation.EntityID[len("automation."):]

		fullAuto, err := s.Client().GetAutomation(s.Context(), automationID)
		s.Require().NoError(err, "Should be able to get automation details")
		s.NotNil(fullAuto)
		s.NotNil(fullAuto.Config, "Automation should have config")
		s.T().Logf("Successfully fetched automation config for: %s", testAutomation.EntityID)
	}

	s.T().Log("All data fetching operations for coverage analysis succeeded")
}

func (s *AutomationCoverageIntegrationTestSuite) TestCoverageDataAvailability() {
	// Verify that coverage analysis can fetch automation configs with all needed data
	// This test ensures GetAutomation returns full config needed for entity extraction

	// Get automations
	automations, err := s.Client().ListAutomations(s.Context())
	s.Require().NoError(err, "Should be able to list automations")

	if len(automations) == 0 {
		s.T().Skip("No automations available to test coverage analysis")
		return
	}

	// Pick first automation and get its full config
	testAutomation := automations[0]
	automationID := testAutomation.EntityID[len("automation."):]

	fullAuto, err := s.Client().GetAutomation(s.Context(), automationID)
	s.Require().NoError(err, "Should be able to get automation details")
	s.Require().NotNil(fullAuto)
	s.Require().NotNil(fullAuto.Config, "Automation should have config")

	// Verify config has expected fields for coverage analysis
	s.T().Logf("Automation %s has config with ID: %s", testAutomation.EntityID, fullAuto.Config.ID)

	// Verify we can serialize config (needed for entity extraction)
	configJSON, err := json.Marshal(fullAuto.Config)
	s.Require().NoError(err, "Config should be serializable")
	s.NotEmpty(configJSON, "Config JSON should not be empty")

	s.T().Logf("Successfully retrieved automation config (%d bytes)", len(configJSON))
}
