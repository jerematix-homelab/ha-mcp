package handlers

import (
	"context"
	"testing"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// TestManageTodoSchema verifies the schema for manage_todo tool.
func TestManageTodoSchema(t *testing.T) {
	t.Parallel()

	registry := mcp.NewRegistry()
	RegisterTodoTools(registry)

	tool, exists := registry.GetTool("manage_todo")
	if !exists {
		t.Fatal("manage_todo tool not registered")
	}

	// Verify basic properties
	if tool.Name != "manage_todo" {
		t.Errorf("tool.Name = %q, want %q", tool.Name, "manage_todo")
	}

	// Verify schema
	schema := tool.InputSchema
	props := schema.Properties

	// Check action field
	actionSchema, ok := props["action"]
	if !ok {
		t.Fatal("action property missing from schema")
	}
	if len(actionSchema.Enum) != 5 {
		t.Errorf("action enum count = %d, want 5 (list, get_items, add_item, update_item, remove_item)", len(actionSchema.Enum))
	}

	// Check status enum
	statusSchema, ok := props["status"]
	if !ok {
		t.Fatal("status property missing from schema")
	}
	if len(statusSchema.Enum) != 2 {
		t.Errorf("status enum count = %d, want 2 (needs_action, completed)", len(statusSchema.Enum))
	}
}

// TestManageTodo_List verifies list action.
func TestManageTodo_List(t *testing.T) {
	t.Parallel()

	client := &UniversalMockClient{
		GetStatesFn: func(context.Context) ([]homeassistant.Entity, error) {
			return []homeassistant.Entity{
				{
					EntityID: "todo.shopping_list",
					State:    "5",
					Attributes: map[string]any{
						"friendly_name": "Shopping List",
					},
				},
				{
					EntityID: "todo.tasks",
					State:    "2",
					Attributes: map[string]any{
						"friendly_name": "Tasks",
					},
				},
			}, nil
		},
	}

	handler := NewTodoHandlers()
	result, err := handler.HandleManageTodo(context.Background(), client, map[string]any{
		"action": "list",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected result content")
	}

	text := result.Content[0].Text
	if !contains(text, "Shopping List") {
		t.Errorf("result text does not contain 'Shopping List': %s", text)
	}
}

// TestManageTodo_GetItems verifies get_items action.
func TestManageTodo_GetItems(t *testing.T) {
	t.Parallel()

	client := &UniversalMockClient{
		CallServiceWithResponseFn: func(context.Context, string, string, map[string]any) (map[string]any, error) {
			return map[string]any{
				"todo.shopping_list": map[string]any{
					"items": []any{
						map[string]any{
							"uid":     "item1",
							"summary": "Buy milk",
							"status":  "needs_action",
						},
						map[string]any{
							"uid":     "item2",
							"summary": "Buy eggs",
							"status":  "completed",
						},
					},
				},
			}, nil
		},
	}

	handler := NewTodoHandlers()
	result, err := handler.HandleManageTodo(context.Background(), client, map[string]any{
		"action":    "get_items",
		"entity_id": "todo.shopping_list",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected result content")
	}

	text := result.Content[0].Text
	if !contains(text, "Buy milk") {
		t.Errorf("result text does not contain 'Buy milk': %s", text)
	}
}

// TestManageTodo_GetItemsStatusFilter verifies status_filter parameter.
func TestManageTodo_GetItemsStatusFilter(t *testing.T) {
	t.Parallel()

	var capturedData map[string]any
	client := &UniversalMockClient{
		CallServiceWithResponseFn: func(_ context.Context, _ string, _ string, data map[string]any) (map[string]any, error) {
			capturedData = data
			return map[string]any{
				"todo.tasks": map[string]any{
					"items": []any{},
				},
			}, nil
		},
	}

	handler := NewTodoHandlers()
	_, err := handler.HandleManageTodo(context.Background(), client, map[string]any{
		"action":        "get_items",
		"entity_id":     "todo.tasks",
		"status_filter": "needs_action",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify status parameter was passed
	if capturedData["status"] != "needs_action" {
		t.Errorf("status = %v, want 'needs_action'", capturedData["status"])
	}
}

// TestManageTodo_AddItem verifies add_item action.
func TestManageTodo_AddItem(t *testing.T) {
	t.Parallel()

	client := &UniversalMockClient{
		CallServiceFn: func(context.Context, string, string, map[string]any) ([]homeassistant.Entity, error) {
			return nil, nil
		},
	}

	handler := NewTodoHandlers()
	result, err := handler.HandleManageTodo(context.Background(), client, map[string]any{
		"action":    "add_item",
		"entity_id": "todo.shopping_list",
		"item":      "Buy bread",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected success result")
	}
}

// TestManageTodo_UpdateItem verifies update_item action.
func TestManageTodo_UpdateItem(t *testing.T) {
	t.Parallel()

	client := &UniversalMockClient{
		CallServiceFn: func(context.Context, string, string, map[string]any) ([]homeassistant.Entity, error) {
			return nil, nil
		},
	}

	handler := NewTodoHandlers()
	result, err := handler.HandleManageTodo(context.Background(), client, map[string]any{
		"action":    "update_item",
		"entity_id": "todo.shopping_list",
		"uid":       "item1",
		"status":    "completed",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected success result")
	}
}

// TestManageTodo_RemoveItem verifies remove_item action.
func TestManageTodo_RemoveItem(t *testing.T) {
	t.Parallel()

	client := &UniversalMockClient{
		CallServiceFn: func(context.Context, string, string, map[string]any) ([]homeassistant.Entity, error) {
			return nil, nil
		},
	}

	handler := NewTodoHandlers()
	result, err := handler.HandleManageTodo(context.Background(), client, map[string]any{
		"action":    "remove_item",
		"entity_id": "todo.shopping_list",
		"uid":       "item1",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected success result")
	}
}

// TestManageTodo_MissingRequiredParams verifies validation for actions requiring parameters.
func TestManageTodo_MissingRequiredParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args map[string]any
	}{
		{
			name: "get_items missing entity_id",
			args: map[string]any{
				"action": "get_items",
			},
		},
		{
			name: "add_item missing entity_id",
			args: map[string]any{
				"action": "add_item",
				"item":   "Test",
			},
		},
		{
			name: "add_item missing item",
			args: map[string]any{
				"action":    "add_item",
				"entity_id": "todo.test",
			},
		},
		{
			name: "update_item missing uid",
			args: map[string]any{
				"action":    "update_item",
				"entity_id": "todo.test",
			},
		},
		{
			name: "remove_item missing uid",
			args: map[string]any{
				"action":    "remove_item",
				"entity_id": "todo.test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &UniversalMockClient{}
			handler := NewTodoHandlers()

			result, err := handler.HandleManageTodo(context.Background(), client, tt.args)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil || !result.IsError {
				t.Error("expected error result")
			}
		})
	}
}
