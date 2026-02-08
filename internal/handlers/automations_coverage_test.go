package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

func TestManageAutomation_Coverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         map[string]any
		setupClient  func() homeassistant.Client
		wantError    bool
		wantContains []string
	}{
		{
			name: "coverage with natural format",
			args: map[string]any{
				"action": "coverage",
				"format": "natural",
			},
			setupClient: func() homeassistant.Client {
				m := &UniversalMockClient{}

				// Mock GetStates - actionable entities
				m.GetStatesFn = func(_ context.Context) ([]homeassistant.Entity, error) {
					return []homeassistant.Entity{
						{EntityID: "light.living_room", State: "on", Attributes: map[string]any{"friendly_name": "Living Room"}},
						{EntityID: "light.bedroom", State: "off", Attributes: map[string]any{"friendly_name": "Bedroom"}},
						{EntityID: "switch.kitchen", State: "on", Attributes: map[string]any{"friendly_name": "Kitchen"}},
						{EntityID: "sensor.temperature", State: "22", Attributes: map[string]any{"friendly_name": "Temperature"}}, // Not actionable
					}, nil
				}

				// Mock Entity Registry
				m.GetEntityRegistryFn = func(_ context.Context) ([]homeassistant.EntityRegistryEntry, error) {
					return []homeassistant.EntityRegistryEntry{
						{EntityID: "light.living_room", AreaID: "living_room"},
						{EntityID: "light.bedroom", AreaID: "bedroom"},
						{EntityID: "switch.kitchen", AreaID: "kitchen"},
						{EntityID: "sensor.temperature", AreaID: "living_room"},
					}, nil
				}

				// Mock Device Registry
				m.GetDeviceRegistryFn = func(_ context.Context) ([]homeassistant.DeviceRegistryEntry, error) {
					return []homeassistant.DeviceRegistryEntry{}, nil
				}

				// Mock Area Registry
				m.GetAreaRegistryFn = func(_ context.Context) ([]homeassistant.AreaRegistryEntry, error) {
					return []homeassistant.AreaRegistryEntry{
						{AreaID: "living_room", Name: "Living Room"},
						{AreaID: "bedroom", Name: "Bedroom"},
						{AreaID: "kitchen", Name: "Kitchen"},
					}, nil
				}

				// Mock ListAutomations
				m.ListAutomationsFn = func(_ context.Context) ([]homeassistant.Automation, error) {
					return []homeassistant.Automation{
						{
							EntityID:     "automation.living_room_auto",
							FriendlyName: "Living Room Auto",
							Config: &homeassistant.AutomationConfig{
								ID:    "auto1",
								Alias: "Living Room Auto",
							},
						},
					}, nil
				}

				// Mock GetAutomation - references light.living_room
				m.GetAutomationFn = func(_ context.Context, _ string) (*homeassistant.Automation, error) {
					return &homeassistant.Automation{
						EntityID:     "automation.living_room_auto",
						FriendlyName: "Living Room Auto",
						Config: &homeassistant.AutomationConfig{
							ID:    "auto1",
							Alias: "Living Room Auto",
							Actions: []any{
								map[string]any{
									"service":   "light.turn_on",
									"entity_id": "light.living_room",
								},
							},
						},
					}, nil
				}

				return m
			},
			wantError:    false,
			wantContains: []string{"Coverage Analysis", "Overall Coverage", "Coverage by Area", "Uncovered Entities"},
		},
		{
			name: "coverage with json format",
			args: map[string]any{
				"action": "coverage",
				"format": "json",
			},
			setupClient: func() homeassistant.Client {
				m := &UniversalMockClient{}

				m.GetStatesFn = func(_ context.Context) ([]homeassistant.Entity, error) {
					return []homeassistant.Entity{
						{EntityID: "light.test", State: "on"},
					}, nil
				}

				m.GetEntityRegistryFn = func(_ context.Context) ([]homeassistant.EntityRegistryEntry, error) {
					return []homeassistant.EntityRegistryEntry{
						{EntityID: "light.test", AreaID: "test_area"},
					}, nil
				}

				m.GetDeviceRegistryFn = func(_ context.Context) ([]homeassistant.DeviceRegistryEntry, error) {
					return []homeassistant.DeviceRegistryEntry{}, nil
				}

				m.GetAreaRegistryFn = func(_ context.Context) ([]homeassistant.AreaRegistryEntry, error) {
					return []homeassistant.AreaRegistryEntry{
						{AreaID: "test_area", Name: "Test Area"},
					}, nil
				}

				m.ListAutomationsFn = func(_ context.Context) ([]homeassistant.Automation, error) {
					return []homeassistant.Automation{}, nil
				}

				return m
			},
			wantError:    false,
			wantContains: []string{"total_entities", "covered_entities", "coverage_percent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &AutomationHandlers{}
			client := tt.setupClient()

			result, err := h.handleManageAutomation(context.Background(), client, tt.args)
			if err != nil {
				t.Fatalf("handleManageAutomation() unexpected error = %v", err)
			}

			if result.IsError != tt.wantError {
				t.Errorf("IsError = %v, want %v. Content: %s", result.IsError, tt.wantError, result.Content[0].Text)
			}

			content := result.Content[0].Text
			for _, want := range tt.wantContains {
				if !strings.Contains(content, want) {
					t.Errorf("Expected content to contain %q, got: %s", want, content)
				}
			}
		})
	}
}
