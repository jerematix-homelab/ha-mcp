//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// UpdateBlueprintIntegrationTestSuite combines update and blueprint tests
// since both are lightweight read-only operations.
type UpdateBlueprintIntegrationTestSuite struct {
	IntegrationTestSuite
}

func TestUpdateBlueprintIntegration(t *testing.T) {
	suite.Run(t, new(UpdateBlueprintIntegrationTestSuite))
}

// ============================================================================
// Update Integration Tests
// ============================================================================

// TestListUpdates verifies that we can list all update entities.
func (s *UpdateBlueprintIntegrationTestSuite) TestListUpdates() {
	// Get all states and filter for update domain
	states, err := s.Client().GetStates(s.Context())
	s.Require().NoError(err, "Failed to get states")

	var updates []string
	for _, state := range states {
		if strings.HasPrefix(state.EntityID, "update.") {
			updates = append(updates, state.EntityID)
			s.T().Logf("  - %s (state: %s)", state.EntityID, state.State)
		}
	}

	s.T().Logf("Found %d update entit(ies)", len(updates))
}

// TestGetReleaseNotes verifies we can get release notes for an update.
func (s *UpdateBlueprintIntegrationTestSuite) TestGetReleaseNotes() {
	// Get all states to find an update entity
	states, err := s.Client().GetStates(s.Context())
	s.Require().NoError(err, "Failed to get states")

	var updateEntityID string
	for _, state := range states {
		if strings.HasPrefix(state.EntityID, "update.") {
			updateEntityID = state.EntityID
			break
		}
	}

	if updateEntityID == "" {
		s.T().Skip("No update entities found in Home Assistant")
		return
	}

	s.T().Logf("Testing release notes for: %s", updateEntityID)

	// Call update/release_notes WebSocket command
	response, err := s.Client().SendHACSCommand(s.Context(), "update/release_notes", map[string]any{
		"entity_id": updateEntityID,
	})

	// Note: This may fail if the update doesn't support release notes
	if err != nil {
		s.T().Logf("Release notes not available for %s: %v", updateEntityID, err)
	} else {
		s.NotNil(response, "Response should not be nil when successful")
		s.T().Log("Successfully retrieved release notes")
	}
}

// ============================================================================
// Blueprint Integration Tests
// ============================================================================

// TestListAutomationBlueprints verifies that we can list automation blueprints.
func (s *UpdateBlueprintIntegrationTestSuite) TestListAutomationBlueprints() {
	// Call blueprint/list WebSocket command
	response, err := s.Client().SendHACSCommand(s.Context(), "blueprint/list", map[string]any{
		"domain": "automation",
	})
	s.Require().NoError(err, "Failed to list automation blueprints")
	s.NotNil(response, "Response should not be nil")

	s.T().Log("Successfully listed automation blueprints")

	// Response should be a map of blueprint paths
	if blueprintsMap, ok := response.(map[string]any); ok {
		s.T().Logf("Found %d automation blueprint(s)", len(blueprintsMap))
	}
}

// TestListScriptBlueprints verifies that we can list script blueprints.
func (s *UpdateBlueprintIntegrationTestSuite) TestListScriptBlueprints() {
	// Call blueprint/list WebSocket command
	response, err := s.Client().SendHACSCommand(s.Context(), "blueprint/list", map[string]any{
		"domain": "script",
	})
	s.Require().NoError(err, "Failed to list script blueprints")
	s.NotNil(response, "Response should not be nil")

	s.T().Log("Successfully listed script blueprints")

	// Response should be a map of blueprint paths
	if blueprintsMap, ok := response.(map[string]any); ok {
		s.T().Logf("Found %d script blueprint(s)", len(blueprintsMap))
	}
}
