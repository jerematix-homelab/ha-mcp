package handlers

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/zorak1103/ha-mcp/internal/mcp"
)

func TestHACSHandlers_RegisterTools(t *testing.T) {
	t.Parallel()
	registry := mcp.NewRegistry()
	handlers := NewHACSHandlers()
	handlers.RegisterTools(registry)

	tool, found := registry.GetTool("manage_hacs")
	if !found {
		t.Fatal("manage_hacs tool not registered")
	}

	// Verify schema
	schema := tool.InputSchema
	if schema.Type != "object" {
		t.Errorf("expected schema type object, got %s", schema.Type)
	}

	// Verify required fields
	requiredFields := []string{"action"}
	for _, field := range requiredFields {
		found := false
		for _, req := range schema.Required {
			if req == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("required field %s not found in schema", field)
		}
	}

	// Verify action enum has 12 values
	actionProp := schema.Properties["action"]
	if len(actionProp.Enum) != 12 {
		t.Errorf("expected 12 actions, got %d", len(actionProp.Enum))
	}

	// Verify format enum
	formatProp := schema.Properties["format"]
	if len(formatProp.Enum) != 2 {
		t.Errorf("expected 2 format options, got %d", len(formatProp.Enum))
	}
}

func TestHACSHandlers_HandleManageHACS(t *testing.T) {
	t.Parallel()

	tests := []handlerTestCase{
		// =============================================================================
		// Read Actions - Info
		// =============================================================================
		{
			name: "info_natural_format",
			args: map[string]any{
				"action": "info",
				"format": "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(context.Context, string, map[string]any) (any, error) {
					return map[string]any{
						"version":       "1.34.0",
						"lovelace_mode": "storage",
					}, nil
				}
			},
			wantContains: []string{"Version:", "1.34.0", "Lovelace mode:", "storage"},
		},
		{
			name: "info_json_format",
			args: map[string]any{
				"action": "info",
				"format": "json",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(_ context.Context, _ string, _ map[string]any) (any, error) {
					return map[string]any{
						"version":       "1.34.0",
						"lovelace_mode": "storage",
					}, nil
				}
			},
			wantContains: []string{`"version"`, `"1.34.0"`, `"lovelace_mode"`, `"storage"`},
		},

		// =============================================================================
		// Read Actions - List
		// =============================================================================
		{
			name: "list_all_natural",
			args: map[string]any{
				"action": "list",
				"format": "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(_ context.Context, _ string, _ map[string]any) (any, error) {
					return []any{
						map[string]any{
							"id":        "123456",
							"name":      "hacs-frontend",
							"category":  "integration",
							"installed": true,
							"status":    "installed",
						},
					}, nil
				}
			},
			wantContains: []string{"hacs-frontend", "integration", "installed"},
		},
		{
			name: "list_installed_only",
			args: map[string]any{
				"action":         "list",
				"installed_only": true,
				"format":         "json",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(_ context.Context, _ string, data map[string]any) (any, error) {
					// Verify filter is passed
					if !data["installed_only"].(bool) {
						return nil, fmt.Errorf("expected installed_only filter")
					}
					return []any{}, nil
				}
			},
			wantContains: []string{"[]"},
		},
		{
			name: "list_pending_update",
			args: map[string]any{
				"action":         "list",
				"pending_update": true,
				"format":         "json",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(_ context.Context, _ string, data map[string]any) (any, error) {
					if !data["pending_update"].(bool) {
						return nil, fmt.Errorf("expected pending_update filter")
					}
					return []any{}, nil
				}
			},
			wantContains: []string{"[]"},
		},
		{
			name: "list_category_filter",
			args: map[string]any{
				"action":   "list",
				"category": "plugin",
				"format":   "json",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(_ context.Context, _ string, data map[string]any) (any, error) {
					if data["category"] != "plugin" {
						return nil, fmt.Errorf("expected category filter")
					}
					return []any{}, nil
				}
			},
			wantContains: []string{"[]"},
		},

		// =============================================================================
		// Read Actions - Get, Releases, Release Notes, Critical
		// =============================================================================
		{
			name: "get_repository_natural",
			args: map[string]any{
				"action":        "get",
				"repository_id": "123456",
				"format":        "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(_ context.Context, _ string, _ map[string]any) (any, error) {
					return map[string]any{
						"name":     "hacs-frontend",
						"category": "integration",
						"status":   "installed",
					}, nil
				}
			},
			wantContains: []string{"hacs-frontend", "integration", "installed"},
		},
		{
			name: "releases_json",
			args: map[string]any{
				"action":        "releases",
				"repository_id": "123456",
				"format":        "json",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(context.Context, string, map[string]any) (any, error) {
					return []any{
						map[string]any{"tag": "1.0.0"},
					}, nil
				}
			},
			wantContains: []string{`"tag"`, `"1.0.0"`},
		},
		{
			name: "release_notes_natural",
			args: map[string]any{
				"action":        "release_notes",
				"repository_id": "123456",
				"format":        "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(context.Context, string, map[string]any) (any, error) {
					return map[string]any{
						"tag":  "1.0.0",
						"body": "Bug fixes",
					}, nil
				}
			},
			wantContains: []string{"1.0.0", "Bug fixes"},
		},
		{
			name: "critical_list",
			args: map[string]any{
				"action": "critical",
				"format": "json",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(context.Context, string, map[string]any) (any, error) {
					return []any{}, nil
				}
			},
			wantContains: []string{"[]"},
		},

		// =============================================================================
		// Write Actions - Download, Uninstall, Add, Remove, Refresh, Toggle Beta
		// =============================================================================
		{
			name: "download_repository",
			args: map[string]any{
				"action":        "download",
				"repository_id": "123456",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(_ context.Context, command string, _ map[string]any) (any, error) {
					if command != "hacs/repository/download" {
						return nil, fmt.Errorf("wrong command: %s", command)
					}
					return map[string]any{"success": true}, nil
				}
			},
			wantContains: []string{"downloaded", "123456"},
		},
		{
			name: "download_with_version",
			args: map[string]any{
				"action":        "download",
				"repository_id": "123456",
				"version":       "1.0.0",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(_ context.Context, _ string, data map[string]any) (any, error) {
					if data["version"] != "1.0.0" {
						return nil, fmt.Errorf("expected version parameter")
					}
					return map[string]any{"success": true}, nil
				}
			},
			wantContains: []string{"downloaded", "123456", "1.0.0"},
		},
		{
			name: "uninstall_repository",
			args: map[string]any{
				"action":        "uninstall",
				"repository_id": "123456",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(_ context.Context, command string, _ map[string]any) (any, error) {
					if command != "hacs/repository/remove" {
						return nil, fmt.Errorf("wrong command: %s", command)
					}
					return map[string]any{"success": true}, nil
				}
			},
			wantContains: []string{"uninstalled", "123456"},
		},
		{
			name: "add_repository",
			args: map[string]any{
				"action":     "add_repository",
				"repository": "owner/repo",
				"category":   "integration",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(_ context.Context, command string, data map[string]any) (any, error) {
					if command != "hacs/repositories/add" {
						return nil, fmt.Errorf("wrong command: %s", command)
					}
					if data["repository"] != "owner/repo" || data["category"] != "integration" {
						return nil, fmt.Errorf("missing required parameters")
					}
					return map[string]any{"success": true}, nil
				}
			},
			wantContains: []string{"added", "owner/repo"},
		},
		{
			name: "remove_repository",
			args: map[string]any{
				"action":        "remove_repository",
				"repository_id": "123456",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(_ context.Context, command string, _ map[string]any) (any, error) {
					if command != "hacs/repositories/remove" {
						return nil, fmt.Errorf("wrong command: %s", command)
					}
					return map[string]any{"success": true}, nil
				}
			},
			wantContains: []string{"removed", "123456"},
		},
		{
			name: "refresh_repository",
			args: map[string]any{
				"action":        "refresh",
				"repository_id": "123456",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(_ context.Context, command string, _ map[string]any) (any, error) {
					if command != "hacs/repository/refresh" {
						return nil, fmt.Errorf("wrong command: %s", command)
					}
					return map[string]any{"success": true}, nil
				}
			},
			wantContains: []string{"refreshed", "123456"},
		},
		{
			name: "toggle_beta_enable",
			args: map[string]any{
				"action":        "toggle_beta",
				"repository_id": "123456",
				"show_beta":     true,
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(_ context.Context, command string, data map[string]any) (any, error) {
					if command != "hacs/repository/beta" {
						return nil, fmt.Errorf("wrong command: %s", command)
					}
					if data["show_beta"] != true {
						return nil, fmt.Errorf("expected show_beta parameter")
					}
					return map[string]any{"success": true}, nil
				}
			},
			wantContains: []string{"beta", "enabled", "123456"},
		},

		// =============================================================================
		// Error Cases
		// =============================================================================
		{
			name:         "missing_action",
			args:         map[string]any{},
			wantContains: []string{"action", "required"},
		},
		{
			name: "invalid_action",
			args: map[string]any{
				"action": "invalid",
			},
			wantContains: []string{"invalid", "action"},
		},
		{
			name: "get_missing_repository_id",
			args: map[string]any{
				"action": "get",
			},
			wantContains: []string{"repository_id", "required"},
		},
		{
			name: "add_missing_repository",
			args: map[string]any{
				"action":   "add_repository",
				"category": "integration",
			},
			wantContains: []string{"repository", "required"},
		},
		{
			name: "add_missing_category",
			args: map[string]any{
				"action":     "add_repository",
				"repository": "owner/repo",
			},
			wantContains: []string{"category", "required"},
		},
		{
			name: "toggle_beta_missing_show_beta",
			args: map[string]any{
				"action":        "toggle_beta",
				"repository_id": "123456",
			},
			wantContains: []string{"show_beta", "required"},
		},
		{
			name: "hacs_not_installed",
			args: map[string]any{
				"action": "info",
			},
			setupMock: func(m *UniversalMockClient) {
				m.SendHACSCommandFn = func(context.Context, string, map[string]any) (any, error) {
					return nil, fmt.Errorf("unknown_command")
				}
			},
			wantContains: []string{"HACS", "not installed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			mock := &UniversalMockClient{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			handlers := NewHACSHandlers()
			result, err := handlers.HandleManageHACS(ctx, mock, tt.args)

			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			resultText := getResultText(t, result)
			for _, want := range tt.wantContains {
				if !containsText(resultText, want) {
					t.Errorf("result does not contain %q:\n%s", want, resultText)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if containsText(resultText, notWant) {
					t.Errorf("result should not contain %q:\n%s", notWant, resultText)
				}
			}
		})
	}
}

func TestHACSHandlers_FormatOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		format       string
		data         any
		wantContains []string
	}{
		{
			name:   "natural_format_map",
			format: "natural",
			data: map[string]any{
				"version": "1.34.0",
				"status":  "ready",
			},
			wantContains: []string{"Version:", "1.34.0", "Status:", "ready"},
		},
		{
			name:   "json_format_map",
			format: "json",
			data: map[string]any{
				"version": "1.34.0",
			},
			wantContains: []string{`"version"`, `"1.34.0"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handlers := NewHACSHandlers()

			// Call a private formatter method through HandleManageHACS
			ctx := context.Background()
			mock := &UniversalMockClient{
				SendHACSCommandFn: func(context.Context, string, map[string]any) (any, error) {
					return tt.data, nil
				},
			}

			result, err := handlers.HandleManageHACS(ctx, mock, map[string]any{
				"action": "info",
				"format": tt.format,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			resultText := getResultText(t, result)
			for _, want := range tt.wantContains {
				if !containsText(resultText, want) {
					t.Errorf("result does not contain %q:\n%s", want, resultText)
				}
			}
		})
	}
}

// Helper to extract text from MCP result.
func getResultText(t *testing.T, result *mcp.ToolsCallResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		return ""
	}
	return result.Content[0].Text
}

// Helper to check if text contains substring (case-insensitive).
func containsText(text, substr string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(substr))
}
