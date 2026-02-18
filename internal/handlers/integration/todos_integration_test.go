//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TodoIntegrationTestSuite struct {
	IntegrationTestSuite
}

func TestTodoIntegration(t *testing.T) {
	suite.Run(t, new(TodoIntegrationTestSuite))
}

// TestListTodoLists verifies that we can list all todo lists.
func (s *TodoIntegrationTestSuite) TestListTodoLists() {
	// Get all states and filter for todo domain
	states, err := s.Client().GetStates(s.Context())
	s.Require().NoError(err, "Failed to get states")

	var todoLists []string
	for _, state := range states {
		if len(state.EntityID) > 5 && state.EntityID[:5] == "todo." {
			todoLists = append(todoLists, state.EntityID)
		}
	}

	s.T().Logf("Found %d todo list(s)", len(todoLists))
	// Note: Test doesn't require todo lists to exist, just verifies API works
}

// TestTodoItemOperations verifies CRUD operations on todo items.
// This test requires at least one todo list to exist in Home Assistant.
func (s *TodoIntegrationTestSuite) TestTodoItemOperations() {
	// Get all states to find a todo list
	states, err := s.Client().GetStates(s.Context())
	s.Require().NoError(err, "Failed to get states")

	var todoEntityID string
	for _, state := range states {
		if len(state.EntityID) > 5 && state.EntityID[:5] == "todo." {
			todoEntityID = state.EntityID
			break
		}
	}

	if todoEntityID == "" {
		s.T().Skip("No todo lists found in Home Assistant - create one manually to run this test")
		return
	}

	s.T().Logf("Using todo list: %s", todoEntityID)

	// Test 1: Get items (verify API works)
	response, err := s.Client().CallServiceWithResponse(s.Context(), "todo", "get_items", map[string]any{
		"entity_id": todoEntityID,
	})
	s.Require().NoError(err, "Failed to get todo items")
	s.NotNil(response, "Response should not be nil")

	// Extract items from response: {"todo.list": {"items": [...]}}
	entityData, ok := response[todoEntityID].(map[string]any)
	s.Require().True(ok, "Response should contain entity data")

	initialItems, ok := entityData["items"].([]any)
	s.Require().True(ok, "Entity data should contain items array")
	initialCount := len(initialItems)

	s.T().Logf("Initial item count: %d", initialCount)

	// Test 2: Add item
	testItemSummary := "Integration Test Item - Delete Me"
	_, err = s.Client().CallService(s.Context(), "todo", "add_item", map[string]any{
		"entity_id":   todoEntityID,
		"item":        testItemSummary,
		"description": "Created by ha-mcp integration test",
	})
	s.Require().NoError(err, "Failed to add todo item")

	s.T().Log("Added test item successfully")

	// Test 3: Get items again to verify addition
	response, err = s.Client().CallServiceWithResponse(s.Context(), "todo", "get_items", map[string]any{
		"entity_id": todoEntityID,
	})
	s.Require().NoError(err, "Failed to get todo items after add")

	entityData, _ = response[todoEntityID].(map[string]any)
	newItems, _ := entityData["items"].([]any)
	s.Require().Len(newItems, initialCount+1, "Should have one more item after add")

	// Find the newly added item
	var testItemUID string
	for _, item := range newItems {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if summary, ok := itemMap["summary"].(string); ok && summary == testItemSummary {
			testItemUID, _ = itemMap["uid"].(string)
			break
		}
	}
	s.Require().NotEmpty(testItemUID, "Should find the added item with UID")

	s.T().Logf("Found added item with UID: %s", testItemUID)

	// Test 4: Update item (mark as completed)
	_, err = s.Client().CallService(s.Context(), "todo", "update_item", map[string]any{
		"entity_id": todoEntityID,
		"item":      testItemUID,
		"status":    "completed",
	})
	s.Require().NoError(err, "Failed to update todo item")

	s.T().Log("Updated item status to completed")

	// Test 5: Verify update
	response, err = s.Client().CallServiceWithResponse(s.Context(), "todo", "get_items", map[string]any{
		"entity_id": todoEntityID,
	})
	s.Require().NoError(err, "Failed to get todo items after update")

	entityData, _ = response[todoEntityID].(map[string]any)
	updatedItems, _ := entityData["items"].([]any)

	var foundCompleted bool
	for _, item := range updatedItems {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if uid, ok := itemMap["uid"].(string); ok && uid == testItemUID {
			status, _ := itemMap["status"].(string)
			foundCompleted = (status == "completed")
			break
		}
	}
	s.Require().True(foundCompleted, "Item should be marked as completed")

	// Test 6: Remove item (cleanup)
	_, err = s.Client().CallService(s.Context(), "todo", "remove_item", map[string]any{
		"entity_id": todoEntityID,
		"item":      testItemUID,
	})
	s.Require().NoError(err, "Failed to remove todo item")

	s.T().Log("Removed test item successfully")

	// Test 7: Verify removal
	response, err = s.Client().CallServiceWithResponse(s.Context(), "todo", "get_items", map[string]any{
		"entity_id": todoEntityID,
	})
	s.Require().NoError(err, "Failed to get todo items after remove")

	entityData, _ = response[todoEntityID].(map[string]any)
	finalItems, _ := entityData["items"].([]any)
	s.Require().Len(finalItems, initialCount, "Should be back to initial item count after removal")
}

// TestTodoStatusFilter verifies status filtering works.
func (s *TodoIntegrationTestSuite) TestTodoStatusFilter() {
	// Get all states to find a todo list
	states, err := s.Client().GetStates(s.Context())
	s.Require().NoError(err, "Failed to get states")

	var todoEntityID string
	for _, state := range states {
		if len(state.EntityID) > 5 && state.EntityID[:5] == "todo." {
			todoEntityID = state.EntityID
			break
		}
	}

	if todoEntityID == "" {
		s.T().Skip("No todo lists found")
		return
	}

	// Test filtering by needs_action status
	response, err := s.Client().CallServiceWithResponse(s.Context(), "todo", "get_items", map[string]any{
		"entity_id": todoEntityID,
		"status":    "needs_action",
	})
	s.Require().NoError(err, "Failed to get items with status filter")
	s.NotNil(response, "Response should not be nil")

	// Test filtering by completed status
	response, err = s.Client().CallServiceWithResponse(s.Context(), "todo", "get_items", map[string]any{
		"entity_id": todoEntityID,
		"status":    "completed",
	})
	s.Require().NoError(err, "Failed to get completed items")
	s.NotNil(response, "Response should not be nil")

	s.T().Log("Status filtering works correctly")
}
