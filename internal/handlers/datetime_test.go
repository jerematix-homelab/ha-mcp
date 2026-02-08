package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

func TestDatetimeHandlers_RegisterTools(t *testing.T) {
	t.Parallel()

	h := NewDatetimeHandlers()
	registry := mcp.NewRegistry()
	h.RegisterTools(registry)

	tools := registry.ListTools()
	if len(tools) != 1 {
		t.Errorf("RegisterTools() registered %d tools, want 1", len(tools))
	}

	found := false
	for _, tool := range tools {
		if tool.Name == "get_datetime" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected get_datetime tool to be registered")
	}
}

func TestDatetimeHandlers_GetDatetimeTool(t *testing.T) {
	t.Parallel()

	h := NewDatetimeHandlers()
	tool := h.getDatetimeTool()

	verifyToolSchema(t, tool, toolSchemaExpectation{
		ExpectedName:    "get_datetime",
		RequiredParams:  []string{},
		OptionalParams:  []string{"timezone"},
		WantDescription: true,
	})
}

func TestDatetimeHandlers_HandleGetDatetime(t *testing.T) {
	t.Parallel()

	// Fixed reference time for testing: 2024-01-15 14:30:00 UTC
	referenceTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

	tests := []handlerTestCase{
		{
			name: "default timezone from HA config",
			args: map[string]any{},
			setupMock: func(m *UniversalMockClient) {
				m.GetConfigFn = func(_ context.Context) (*homeassistant.Config, error) {
					return &homeassistant.Config{
						TimeZone: "Europe/Berlin",
					}, nil
				}
			},
			wantContains: []string{
				"Date:",
				"Time:",
				"Timezone: Europe/Berlin",
				"ISO 8601:",
				"Unix timestamp:",
				"Day of week:",
			},
		},
		{
			name: "custom timezone override",
			args: map[string]any{
				"timezone": "America/New_York",
			},
			wantContains: []string{
				"Timezone: America/New_York",
			},
		},
		{
			name: "UTC timezone",
			args: map[string]any{
				"timezone": "UTC",
			},
			wantContains: []string{"Timezone: UTC"},
		},
		{
			name: "invalid timezone",
			args: map[string]any{
				"timezone": "Invalid/Timezone",
			},
			wantError: true,
		},
		{
			name: "API error on GetConfig",
			args: map[string]any{},
			setupMock: func(m *UniversalMockClient) {
				m.GetConfigFn = func(_ context.Context) (*homeassistant.Config, error) {
					return nil, errors.New("API error")
				}
			},
			wantError: true,
		},
		{
			name: "invalid timezone in HA config",
			args: map[string]any{},
			setupMock: func(m *UniversalMockClient) {
				m.GetConfigFn = func(_ context.Context) (*homeassistant.Config, error) {
					return &homeassistant.Config{
						TimeZone: "Invalid/Zone",
					}, nil
				}
			},
			wantError: true,
		},
	}

	h := NewDatetimeHandlers()

	// Override time function for consistent testing
	h.nowFunc = func() time.Time { return referenceTime }

	runHandlerTestCases(t, tests, h.HandleGetDatetime)
}
