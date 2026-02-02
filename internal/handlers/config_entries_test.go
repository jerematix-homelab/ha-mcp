package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

func TestConfigEntryHandlers_RegisterTools(t *testing.T) {
	t.Parallel()

	registry := mcp.NewRegistry()
	h := NewConfigEntryHandlers()
	h.RegisterTools(registry)

	tools := registry.ListTools()
	if len(tools) != 2 {
		t.Errorf("RegisterTools() registered %d tools, want 2", len(tools))
	}

	// Verify expected tools exist
	toolMap := make(map[string]bool)
	for _, tool := range tools {
		toolMap[tool.Name] = true
	}

	expectedTools := []string{"list_config_entries", "get_config_entry"}
	for _, expected := range expectedTools {
		if !toolMap[expected] {
			t.Errorf("Expected tool %q not registered", expected)
		}
	}
}

func TestConfigEntryHandlers_ListConfigEntriesTool_Schema(t *testing.T) {
	t.Parallel()

	h := NewConfigEntryHandlers()
	tool := h.listConfigEntriesTool()

	verifyToolSchema(t, tool, toolSchemaExpectation{
		ExpectedName:    "list_config_entries",
		RequiredParams:  []string{},
		OptionalParams:  []string{"domain"},
		WantDescription: true,
	})
}

func TestConfigEntryHandlers_GetConfigEntryTool_Schema(t *testing.T) {
	t.Parallel()

	h := NewConfigEntryHandlers()
	tool := h.getConfigEntryTool()

	verifyToolSchema(t, tool, toolSchemaExpectation{
		ExpectedName:    "get_config_entry",
		RequiredParams:  []string{"entry_id"},
		OptionalParams:  []string{},
		WantDescription: true,
	})
}

func TestConfigEntryHandlers_HandleListConfigEntries(t *testing.T) {
	t.Parallel()

	h := NewConfigEntryHandlers()

	tests := []handlerTestCase{
		{
			name: "list all entries",
			args: map[string]any{},
			setupMock: func(m *UniversalMockClient) {
				m.GetConfigEntriesFn = func(_ context.Context, domain string) ([]homeassistant.ConfigEntryFull, error) {
					if domain != "" {
						t.Errorf("expected empty domain, got %q", domain)
					}
					return []homeassistant.ConfigEntryFull{
						{
							EntryID: "abc123",
							Domain:  "template",
							Title:   "My Template Sensor",
							State:   "loaded",
						},
						{
							EntryID: "def456",
							Domain:  "hue",
							Title:   "Philips Hue",
							State:   "loaded",
						},
					}, nil
				}
			},
			wantError:    false,
			wantContains: []string{"abc123", "template", "My Template Sensor", "def456", "hue"},
		},
		{
			name: "filter by domain",
			args: map[string]any{"domain": "template"},
			setupMock: func(m *UniversalMockClient) {
				m.GetConfigEntriesFn = func(_ context.Context, domain string) ([]homeassistant.ConfigEntryFull, error) {
					if domain != "template" {
						t.Errorf("expected domain 'template', got %q", domain)
					}
					return []homeassistant.ConfigEntryFull{
						{
							EntryID: "abc123",
							Domain:  "template",
							Title:   "My Template Sensor",
							State:   "loaded",
						},
					}, nil
				}
			},
			wantError:    false,
			wantContains: []string{"abc123", "template"},
		},
		{
			name: "empty results",
			args: map[string]any{},
			setupMock: func(m *UniversalMockClient) {
				m.GetConfigEntriesFn = func(_ context.Context, _ string) ([]homeassistant.ConfigEntryFull, error) {
					return []homeassistant.ConfigEntryFull{}, nil
				}
			},
			wantError:    false,
			wantContains: []string{"0 config entries"},
		},
		{
			name: "client error",
			args: map[string]any{},
			setupMock: func(m *UniversalMockClient) {
				m.GetConfigEntriesFn = func(_ context.Context, _ string) ([]homeassistant.ConfigEntryFull, error) {
					return nil, errors.New("connection failed")
				}
			},
			wantError:    true,
			wantContains: []string{"Error", "connection failed"},
		},
	}

	runHandlerTestCases(t, tests, h.handleListConfigEntries)
}

func TestConfigEntryHandlers_HandleGetConfigEntry(t *testing.T) {
	t.Parallel()

	h := NewConfigEntryHandlers()

	tests := []handlerTestCase{
		{
			name: "get entry with options",
			args: map[string]any{"entry_id": "abc123"},
			setupMock: func(m *UniversalMockClient) {
				m.GetConfigEntryFn = func(_ context.Context, entryID string) (*homeassistant.ConfigEntryFull, error) {
					if entryID != "abc123" {
						t.Errorf("expected entry_id 'abc123', got %q", entryID)
					}
					return &homeassistant.ConfigEntryFull{
						EntryID: "abc123",
						Domain:  "template",
						Title:   "My Template Sensor",
						State:   "loaded",
						Options: map[string]any{
							"name":  "Angeschaltete Lichter",
							"state": "{{ states.light | selectattr('state', 'eq', 'on') | list | count }}",
						},
					}, nil
				}
			},
			wantError:    false,
			wantContains: []string{"abc123", "template", "My Template Sensor", "state", "selectattr"},
		},
		{
			name: "entry not found",
			args: map[string]any{"entry_id": "nonexistent"},
			setupMock: func(m *UniversalMockClient) {
				m.GetConfigEntryFn = func(_ context.Context, _ string) (*homeassistant.ConfigEntryFull, error) {
					return nil, errors.New("config entry not found")
				}
			},
			wantError:    true,
			wantContains: []string{"Error", "config entry not found"},
		},
	}

	// Add required parameter tests
	tests = append(tests, paramRequiredTestCases("entry_id")...)

	runHandlerTestCases(t, tests, h.handleGetConfigEntry)
}

func TestRegisterConfigEntryTools(t *testing.T) {
	t.Parallel()

	registry := mcp.NewRegistry()
	RegisterConfigEntryTools(registry)

	tools := registry.ListTools()
	if len(tools) == 0 {
		t.Error("RegisterConfigEntryTools() registered no tools")
	}
}
