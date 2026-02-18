package handlers

import (
	"context"
	"testing"

	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// TestManageBlueprintSchema verifies the schema for manage_blueprint tool.
func TestManageBlueprintSchema(t *testing.T) {
	t.Parallel()

	registry := mcp.NewRegistry()
	RegisterBlueprintTools(registry)

	tool, exists := registry.GetTool("manage_blueprint")
	if !exists {
		t.Fatal("manage_blueprint tool not registered")
	}

	// Verify basic properties
	if tool.Name != "manage_blueprint" {
		t.Errorf("tool.Name = %q, want %q", tool.Name, "manage_blueprint")
	}
	if tool.Description == "" {
		t.Error("tool.Description is empty")
	}

	// Verify schema properties
	schema := tool.InputSchema
	props := schema.Properties

	// Check action field
	actionSchema, ok := props["action"]
	if !ok {
		t.Fatal("action property missing from schema")
	}
	if actionSchema.Type != "string" {
		t.Errorf("action type = %q, want %q", actionSchema.Type, "string")
	}
	if len(actionSchema.Enum) != 2 {
		t.Errorf("action enum count = %d, want 2 (list, import)", len(actionSchema.Enum))
	}

	// Check domain field
	domainSchema, ok := props["domain"]
	if !ok {
		t.Fatal("domain property missing from schema")
	}
	if len(domainSchema.Enum) != 2 {
		t.Errorf("domain enum count = %d, want 2 (automation, script)", len(domainSchema.Enum))
	}

	// Check format field
	formatSchema, ok := props["format"]
	if !ok {
		t.Fatal("format property missing from schema")
	}
	if len(formatSchema.Enum) != 2 {
		t.Errorf("format enum count = %d, want 2", len(formatSchema.Enum))
	}

	// Check required fields
	if len(schema.Required) != 1 {
		t.Errorf("required count = %d, want 1 (action)", len(schema.Required))
	}
	if schema.Required[0] != "action" {
		t.Errorf("required[0] = %q, want %q", schema.Required[0], "action")
	}
}

// TestManageBlueprint_MissingAction verifies validation when action is missing.
func TestManageBlueprint_MissingAction(t *testing.T) {
	t.Parallel()

	client := &UniversalMockClient{}
	handler := NewBlueprintHandlers()

	result, err := handler.HandleManageBlueprint(context.Background(), client, map[string]any{})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected result with error message")
	}
	if !result.IsError {
		t.Error("expected IsError to be true")
	}
}

// TestManageBlueprint_InvalidAction verifies validation for invalid action.
func TestManageBlueprint_InvalidAction(t *testing.T) {
	t.Parallel()

	client := &UniversalMockClient{}
	handler := NewBlueprintHandlers()

	result, err := handler.HandleManageBlueprint(context.Background(), client, map[string]any{
		"action": "invalid_action",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected result with error message")
	}
	if !result.IsError {
		t.Error("expected IsError to be true")
	}
}

// TestManageBlueprint_List verifies list action.
func TestManageBlueprint_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         map[string]any
		mockResponse any
		wantContain  string
	}{
		{
			name: "list automation blueprints natural format",
			args: map[string]any{
				"action": "list",
				"domain": "automation",
			},
			mockResponse: map[string]any{
				"blueprints/automation/homeassistant/motion_light.yaml": map[string]any{
					"metadata": map[string]any{
						"name":   "Motion-activated Light",
						"source": "builtin",
					},
				},
			},
			wantContain: "Motion-activated Light",
		},
		{
			name: "list script blueprints json format",
			args: map[string]any{
				"action": "list",
				"domain": "script",
				"format": "json",
			},
			mockResponse: map[string]any{
				"blueprints/script/custom/notify.yaml": map[string]any{
					"metadata": map[string]any{
						"name": "Notification Script",
					},
				},
			},
			wantContain: "notify.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &UniversalMockClient{
				SendHACSCommandFn: func(context.Context, string, map[string]any) (any, error) {
					return tt.mockResponse, nil
				},
			}

			handler := NewBlueprintHandlers()
			result, err := handler.HandleManageBlueprint(context.Background(), client, tt.args)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil || len(result.Content) == 0 {
				t.Fatal("expected result content")
			}

			text := result.Content[0].Text
			if !contains(text, tt.wantContain) {
				t.Errorf("result text does not contain %q: %s", tt.wantContain, text)
			}
		})
	}
}

// TestManageBlueprint_Import verifies import action.
func TestManageBlueprint_Import(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		wantContain string
	}{
		{
			name: "successful import",
			args: map[string]any{
				"action": "import",
				"url":    "https://example.com/blueprint.yaml",
			},
			wantErr:     false,
			wantContain: "successfully imported",
		},
		{
			name: "missing url",
			args: map[string]any{
				"action": "import",
			},
			wantErr:     true,
			wantContain: "url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &UniversalMockClient{
				SendHACSCommandFn: func(context.Context, string, map[string]any) (any, error) {
					return map[string]any{"success": true}, nil
				},
			}

			handler := NewBlueprintHandlers()
			result, err := handler.HandleManageBlueprint(context.Background(), client, tt.args)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil || len(result.Content) == 0 {
				t.Fatal("expected result content")
			}

			if result.IsError != tt.wantErr {
				t.Errorf("IsError = %v, want %v", result.IsError, tt.wantErr)
			}

			text := result.Content[0].Text
			if !contains(text, tt.wantContain) {
				t.Errorf("result text does not contain %q: %s", tt.wantContain, text)
			}
		})
	}
}

// TestManageBlueprint_ListMissingDomain verifies that domain is required for list.
func TestManageBlueprint_ListMissingDomain(t *testing.T) {
	t.Parallel()

	client := &UniversalMockClient{}
	handler := NewBlueprintHandlers()

	result, err := handler.HandleManageBlueprint(context.Background(), client, map[string]any{
		"action": "list",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected result with error message")
	}
	if !result.IsError {
		t.Error("expected IsError to be true")
	}
}
