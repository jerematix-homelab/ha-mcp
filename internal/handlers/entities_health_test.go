package handlers

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

func TestQueryEntities_Health_Analyze(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         map[string]any
		setupMock    func(*UniversalMockClient)
		wantError    bool
		wantContains []string
	}{
		{
			name: "health report natural format",
			args: map[string]any{
				"mode":   "health",
				"format": "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.GetStatesFn = func(_ context.Context) ([]homeassistant.Entity, error) {
					return []homeassistant.Entity{
						{EntityID: "sensor.unavailable_sensor", State: "unavailable"},
						{EntityID: "sensor.unknown_sensor", State: "unknown"},
						{EntityID: "sensor.working", State: "25"},
					}, nil
				}
				m.GetEntityRegistryFn = func(_ context.Context) ([]homeassistant.EntityRegistryEntry, error) {
					return []homeassistant.EntityRegistryEntry{
						{EntityID: "sensor.unavailable_sensor", Platform: "test"},
						{EntityID: "sensor.unknown_sensor", Platform: "test"},
						{EntityID: "sensor.working", Platform: "test"},
					}, nil
				}
				m.GetDeviceRegistryFn = func(_ context.Context) ([]homeassistant.DeviceRegistryEntry, error) {
					return []homeassistant.DeviceRegistryEntry{}, nil
				}
				m.GetAreaRegistryFn = func(_ context.Context) ([]homeassistant.AreaRegistryEntry, error) {
					return []homeassistant.AreaRegistryEntry{}, nil
				}
				m.GetConfigEntriesFn = func(_ context.Context, _ string) ([]homeassistant.ConfigEntryFull, error) {
					return []homeassistant.ConfigEntryFull{}, nil
				}
			},
			wantError:    false,
			wantContains: []string{"Entity Health Report", "Summary", "Unavailable", "Unknown"},
		},
		{
			name: "health report JSON format",
			args: map[string]any{
				"mode":   "health",
				"format": "json",
			},
			setupMock: func(m *UniversalMockClient) {
				m.GetStatesFn = func(_ context.Context) ([]homeassistant.Entity, error) {
					return []homeassistant.Entity{
						{EntityID: "sensor.unavailable_sensor", State: "unavailable"},
					}, nil
				}
				m.GetEntityRegistryFn = func(_ context.Context) ([]homeassistant.EntityRegistryEntry, error) {
					return []homeassistant.EntityRegistryEntry{
						{EntityID: "sensor.unavailable_sensor", Platform: "test"},
					}, nil
				}
				m.GetDeviceRegistryFn = func(_ context.Context) ([]homeassistant.DeviceRegistryEntry, error) {
					return []homeassistant.DeviceRegistryEntry{}, nil
				}
				m.GetAreaRegistryFn = func(_ context.Context) ([]homeassistant.AreaRegistryEntry, error) {
					return []homeassistant.AreaRegistryEntry{}, nil
				}
				m.GetConfigEntriesFn = func(_ context.Context, _ string) ([]homeassistant.ConfigEntryFull, error) {
					return []homeassistant.ConfigEntryFull{}, nil
				}
			},
			wantError:    false,
			wantContains: []string{`"issues"`, `"statistics"`, `"category":"unavailable"`},
		},
		{
			name: "single category filter",
			args: map[string]any{
				"mode":       "health",
				"categories": []any{"unavailable"},
				"format":     "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.GetStatesFn = func(_ context.Context) ([]homeassistant.Entity, error) {
					return []homeassistant.Entity{
						{EntityID: "sensor.unavailable_sensor", State: "unavailable"},
						{EntityID: "sensor.unknown_sensor", State: "unknown"},
					}, nil
				}
				m.GetEntityRegistryFn = func(_ context.Context) ([]homeassistant.EntityRegistryEntry, error) {
					return []homeassistant.EntityRegistryEntry{
						{EntityID: "sensor.unavailable_sensor", Platform: "test"},
						{EntityID: "sensor.unknown_sensor", Platform: "test"},
					}, nil
				}
				m.GetDeviceRegistryFn = func(_ context.Context) ([]homeassistant.DeviceRegistryEntry, error) {
					return []homeassistant.DeviceRegistryEntry{}, nil
				}
				m.GetAreaRegistryFn = func(_ context.Context) ([]homeassistant.AreaRegistryEntry, error) {
					return []homeassistant.AreaRegistryEntry{}, nil
				}
				m.GetConfigEntriesFn = func(_ context.Context, _ string) ([]homeassistant.ConfigEntryFull, error) {
					return []homeassistant.ConfigEntryFull{}, nil
				}
			},
			wantError:    false,
			wantContains: []string{"Unavailable", "unavailable_sensor"},
		},
		{
			name: "multiple categories filter",
			args: map[string]any{
				"mode":       "health",
				"categories": []any{"unavailable", "unknown"},
				"format":     "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.GetStatesFn = func(_ context.Context) ([]homeassistant.Entity, error) {
					return []homeassistant.Entity{
						{EntityID: "sensor.unavailable_sensor", State: "unavailable"},
						{EntityID: "sensor.unknown_sensor", State: "unknown"},
						{EntityID: "sensor.disabled_sensor", State: "25"},
					}, nil
				}
				m.GetEntityRegistryFn = func(_ context.Context) ([]homeassistant.EntityRegistryEntry, error) {
					return []homeassistant.EntityRegistryEntry{
						{EntityID: "sensor.unavailable_sensor", Platform: "test"},
						{EntityID: "sensor.unknown_sensor", Platform: "test"},
						{EntityID: "sensor.disabled_sensor", Platform: "test", DisabledBy: "user"},
					}, nil
				}
				m.GetDeviceRegistryFn = func(_ context.Context) ([]homeassistant.DeviceRegistryEntry, error) {
					return []homeassistant.DeviceRegistryEntry{}, nil
				}
				m.GetAreaRegistryFn = func(_ context.Context) ([]homeassistant.AreaRegistryEntry, error) {
					return []homeassistant.AreaRegistryEntry{}, nil
				}
				m.GetConfigEntriesFn = func(_ context.Context, _ string) ([]homeassistant.ConfigEntryFull, error) {
					return []homeassistant.ConfigEntryFull{}, nil
				}
			},
			wantError:    false,
			wantContains: []string{"Unavailable", "unavailable_sensor", "Unknown", "unknown_sensor"},
		},
		{
			name: "custom stale_days parameter",
			args: map[string]any{
				"mode":       "health",
				"stale_days": float64(60),
				"format":     "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				now := time.Now()
				m.GetStatesFn = func(_ context.Context) ([]homeassistant.Entity, error) {
					return []homeassistant.Entity{
						{
							EntityID:    "sensor.old_sensor",
							State:       "25",
							LastChanged: now.Add(-70 * 24 * time.Hour),
						},
					}, nil
				}
				m.GetEntityRegistryFn = func(_ context.Context) ([]homeassistant.EntityRegistryEntry, error) {
					return []homeassistant.EntityRegistryEntry{
						{EntityID: "sensor.old_sensor", Platform: "test"},
					}, nil
				}
				m.GetDeviceRegistryFn = func(_ context.Context) ([]homeassistant.DeviceRegistryEntry, error) {
					return []homeassistant.DeviceRegistryEntry{}, nil
				}
				m.GetAreaRegistryFn = func(_ context.Context) ([]homeassistant.AreaRegistryEntry, error) {
					return []homeassistant.AreaRegistryEntry{}, nil
				}
				m.GetConfigEntriesFn = func(_ context.Context, _ string) ([]homeassistant.ConfigEntryFull, error) {
					return []homeassistant.ConfigEntryFull{}, nil
				}
			},
			wantError:    false,
			wantContains: []string{"Stale", "60 days"},
		},
		{
			name: "no health issues detected",
			args: map[string]any{
				"mode":   "health",
				"format": "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.GetStatesFn = func(_ context.Context) ([]homeassistant.Entity, error) {
					return []homeassistant.Entity{
						{EntityID: "sensor.working", State: "25", LastChanged: time.Now()},
					}, nil
				}
				m.GetEntityRegistryFn = func(_ context.Context) ([]homeassistant.EntityRegistryEntry, error) {
					return []homeassistant.EntityRegistryEntry{
						{EntityID: "sensor.working", Platform: "test", DisabledBy: ""},
					}, nil
				}
				m.GetDeviceRegistryFn = func(_ context.Context) ([]homeassistant.DeviceRegistryEntry, error) {
					return []homeassistant.DeviceRegistryEntry{}, nil
				}
				m.GetAreaRegistryFn = func(_ context.Context) ([]homeassistant.AreaRegistryEntry, error) {
					return []homeassistant.AreaRegistryEntry{}, nil
				}
				m.GetConfigEntriesFn = func(_ context.Context, _ string) ([]homeassistant.ConfigEntryFull, error) {
					return []homeassistant.ConfigEntryFull{}, nil
				}
			},
			wantError:    false,
			wantContains: []string{"No health issues detected"},
		},
		{
			name: "disabled entities",
			args: map[string]any{
				"mode":   "health",
				"format": "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.GetStatesFn = func(_ context.Context) ([]homeassistant.Entity, error) {
					return []homeassistant.Entity{
						{EntityID: "sensor.disabled", State: "25"},
					}, nil
				}
				m.GetEntityRegistryFn = func(_ context.Context) ([]homeassistant.EntityRegistryEntry, error) {
					return []homeassistant.EntityRegistryEntry{
						{EntityID: "sensor.disabled", Platform: "test", DisabledBy: "user"},
					}, nil
				}
				m.GetDeviceRegistryFn = func(_ context.Context) ([]homeassistant.DeviceRegistryEntry, error) {
					return []homeassistant.DeviceRegistryEntry{}, nil
				}
				m.GetAreaRegistryFn = func(_ context.Context) ([]homeassistant.AreaRegistryEntry, error) {
					return []homeassistant.AreaRegistryEntry{}, nil
				}
				m.GetConfigEntriesFn = func(_ context.Context, _ string) ([]homeassistant.ConfigEntryFull, error) {
					return []homeassistant.ConfigEntryFull{}, nil
				}
			},
			wantError:    false,
			wantContains: []string{"Disabled", "sensor.disabled"},
		},
		{
			name: "orphaned integration",
			args: map[string]any{
				"mode":   "health",
				"format": "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.GetStatesFn = func(_ context.Context) ([]homeassistant.Entity, error) {
					return []homeassistant.Entity{
						{EntityID: "sensor.orphaned", State: "25"},
					}, nil
				}
				m.GetEntityRegistryFn = func(_ context.Context) ([]homeassistant.EntityRegistryEntry, error) {
					return []homeassistant.EntityRegistryEntry{
						{EntityID: "sensor.orphaned", Platform: "test", ConfigEntryID: "missing_entry"},
					}, nil
				}
				m.GetDeviceRegistryFn = func(_ context.Context) ([]homeassistant.DeviceRegistryEntry, error) {
					return []homeassistant.DeviceRegistryEntry{}, nil
				}
				m.GetAreaRegistryFn = func(_ context.Context) ([]homeassistant.AreaRegistryEntry, error) {
					return []homeassistant.AreaRegistryEntry{}, nil
				}
				m.GetConfigEntriesFn = func(_ context.Context, _ string) ([]homeassistant.ConfigEntryFull, error) {
					return []homeassistant.ConfigEntryFull{}, nil // Empty - no matching config entry
				}
			},
			wantError:    false,
			wantContains: []string{"Orphaned Integration", "sensor.orphaned"},
		},
		{
			name: "orphaned device",
			args: map[string]any{
				"mode":   "health",
				"format": "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.GetStatesFn = func(_ context.Context) ([]homeassistant.Entity, error) {
					return []homeassistant.Entity{
						{EntityID: "sensor.orphaned_device", State: "25"},
					}, nil
				}
				m.GetEntityRegistryFn = func(_ context.Context) ([]homeassistant.EntityRegistryEntry, error) {
					return []homeassistant.EntityRegistryEntry{
						{EntityID: "sensor.orphaned_device", Platform: "test", DeviceID: "missing_device"},
					}, nil
				}
				m.GetDeviceRegistryFn = func(_ context.Context) ([]homeassistant.DeviceRegistryEntry, error) {
					return []homeassistant.DeviceRegistryEntry{}, nil // Empty - no matching device
				}
				m.GetAreaRegistryFn = func(_ context.Context) ([]homeassistant.AreaRegistryEntry, error) {
					return []homeassistant.AreaRegistryEntry{}, nil
				}
				m.GetConfigEntriesFn = func(_ context.Context, _ string) ([]homeassistant.ConfigEntryFull, error) {
					return []homeassistant.ConfigEntryFull{}, nil
				}
			},
			wantError:    false,
			wantContains: []string{"Orphaned Device", "sensor.orphaned_device"},
		},
		{
			name: "integration error",
			args: map[string]any{
				"mode":   "health",
				"format": "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.GetStatesFn = func(_ context.Context) ([]homeassistant.Entity, error) {
					return []homeassistant.Entity{
						{EntityID: "sensor.error_integration", State: "25"},
					}, nil
				}
				m.GetEntityRegistryFn = func(_ context.Context) ([]homeassistant.EntityRegistryEntry, error) {
					return []homeassistant.EntityRegistryEntry{
						{EntityID: "sensor.error_integration", Platform: "test", ConfigEntryID: "error_entry"},
					}, nil
				}
				m.GetDeviceRegistryFn = func(_ context.Context) ([]homeassistant.DeviceRegistryEntry, error) {
					return []homeassistant.DeviceRegistryEntry{}, nil
				}
				m.GetAreaRegistryFn = func(_ context.Context) ([]homeassistant.AreaRegistryEntry, error) {
					return []homeassistant.AreaRegistryEntry{}, nil
				}
				m.GetConfigEntriesFn = func(_ context.Context, _ string) ([]homeassistant.ConfigEntryFull, error) {
					return []homeassistant.ConfigEntryFull{
						{EntryID: "error_entry", State: "setup_error"},
					}, nil
				}
			},
			wantError:    false,
			wantContains: []string{"Integration Error", "sensor.error_integration"},
		},
		{
			name: "registry only entities",
			args: map[string]any{
				"mode":   "health",
				"format": "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.GetStatesFn = func(_ context.Context) ([]homeassistant.Entity, error) {
					return []homeassistant.Entity{
						{EntityID: "sensor.has_state", State: "25"},
					}, nil
				}
				m.GetEntityRegistryFn = func(_ context.Context) ([]homeassistant.EntityRegistryEntry, error) {
					return []homeassistant.EntityRegistryEntry{
						{EntityID: "sensor.has_state", Platform: "test"},
						{EntityID: "sensor.registry_only", Platform: "test"},
					}, nil
				}
				m.GetDeviceRegistryFn = func(_ context.Context) ([]homeassistant.DeviceRegistryEntry, error) {
					return []homeassistant.DeviceRegistryEntry{}, nil
				}
				m.GetAreaRegistryFn = func(_ context.Context) ([]homeassistant.AreaRegistryEntry, error) {
					return []homeassistant.AreaRegistryEntry{}, nil
				}
				m.GetConfigEntriesFn = func(_ context.Context, _ string) ([]homeassistant.ConfigEntryFull, error) {
					return []homeassistant.ConfigEntryFull{}, nil
				}
			},
			wantError:    false,
			wantContains: []string{"Registry Only", "sensor.registry_only"},
		},
		{
			name: "domain filter",
			args: map[string]any{
				"mode":   "health",
				"domain": "sensor",
				"format": "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.GetStatesFn = func(_ context.Context) ([]homeassistant.Entity, error) {
					return []homeassistant.Entity{
						{EntityID: "sensor.unavailable_sensor", State: "unavailable"},
						{EntityID: "light.unavailable_light", State: "unavailable"},
					}, nil
				}
				m.GetEntityRegistryFn = func(_ context.Context) ([]homeassistant.EntityRegistryEntry, error) {
					return []homeassistant.EntityRegistryEntry{
						{EntityID: "sensor.unavailable_sensor", Platform: "test"},
						{EntityID: "light.unavailable_light", Platform: "test"},
					}, nil
				}
				m.GetDeviceRegistryFn = func(_ context.Context) ([]homeassistant.DeviceRegistryEntry, error) {
					return []homeassistant.DeviceRegistryEntry{}, nil
				}
				m.GetAreaRegistryFn = func(_ context.Context) ([]homeassistant.AreaRegistryEntry, error) {
					return []homeassistant.AreaRegistryEntry{}, nil
				}
				m.GetConfigEntriesFn = func(_ context.Context, _ string) ([]homeassistant.ConfigEntryFull, error) {
					return []homeassistant.ConfigEntryFull{}, nil
				}
			},
			wantError:    false,
			wantContains: []string{"sensor.unavailable_sensor"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockClient := &UniversalMockClient{}
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			handlers := NewConsolidatedEntityQueryHandlers()
			result, err := handlers.handleQueryEntities(context.Background(), mockClient, tt.args)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Content) == 0 {
				t.Fatal("expected content, got empty")
			}

			content := result.Content[0].Text
			for _, want := range tt.wantContains {
				if !strings.Contains(content, want) {
					t.Errorf("content missing %q, got:\n%s", want, content)
				}
			}
		})
	}
}

func TestQueryEntities_Health_Remove(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         map[string]any
		setupMock    func(*UniversalMockClient)
		wantError    bool
		wantContains []string
	}{
		{
			name: "remove with entity_ids",
			args: map[string]any{
				"mode":       "health",
				"action":     "remove",
				"entity_ids": []any{"sensor.dead1", "sensor.dead2"},
				"format":     "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.RemoveEntityRegistryEntryFn = func(context.Context, string) error {
					return nil
				}
			},
			wantError:    false,
			wantContains: []string{"Removed 2 entities", "sensor.dead1", "sensor.dead2"},
		},
		{
			name: "remove without entity_ids",
			args: map[string]any{
				"mode":   "health",
				"action": "remove",
			},
			wantError:    true,
			wantContains: []string{"entity_ids parameter is required"},
		},
		{
			name: "remove with partial failures",
			args: map[string]any{
				"mode":       "health",
				"action":     "remove",
				"entity_ids": []any{"sensor.success", "sensor.fail"},
				"format":     "natural",
			},
			setupMock: func(m *UniversalMockClient) {
				m.RemoveEntityRegistryEntryFn = func(_ context.Context, id string) error {
					if id == "sensor.fail" {
						return &homeassistant.APIError{StatusCode: 404, Message: "not found"}
					}
					return nil
				}
			},
			wantError:    false,
			wantContains: []string{"Removed 1", "sensor.success", "Failed 1", "sensor.fail"},
		},
		{
			name: "remove JSON format",
			args: map[string]any{
				"mode":       "health",
				"action":     "remove",
				"entity_ids": []any{"sensor.dead"},
				"format":     "json",
			},
			setupMock: func(m *UniversalMockClient) {
				m.RemoveEntityRegistryEntryFn = func(context.Context, string) error {
					return nil
				}
			},
			wantError:    false,
			wantContains: []string{`"removed"`, `"sensor.dead"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockClient := &UniversalMockClient{}
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			handlers := NewConsolidatedEntityQueryHandlers()
			result, err := handlers.handleQueryEntities(context.Background(), mockClient, tt.args)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil || len(result.Content) == 0 {
				t.Fatal("expected content, got empty")
			}

			if result.IsError != tt.wantError {
				t.Errorf("IsError = %v, want %v", result.IsError, tt.wantError)
			}

			content := result.Content[0].Text
			for _, want := range tt.wantContains {
				if !strings.Contains(content, want) {
					t.Errorf("content missing %q, got:\n%s", want, content)
				}
			}
		})
	}
}
