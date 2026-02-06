// Package handlers provides MCP tool handlers for Home Assistant operations.
package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

// Tests use UniversalMockClient from testing_helpers_test.go

func TestHandleGetState(t *testing.T) {
	t.Parallel()

	testEntityData := &homeassistant.Entity{
		EntityID: "light.living_room",
		State:    "on",
		Attributes: map[string]any{
			"friendly_name": "Living Room Light",
			"brightness":    255,
		},
		LastChanged: time.Now(),
		LastUpdated: time.Now(),
	}

	tests := []struct {
		name         string
		args         map[string]any
		setupMock    func(*UniversalMockClient)
		wantError    bool
		wantContains []string
	}{
		{
			name: "success - returns entity state",
			args: map[string]any{"entity_id": "light.living_room", "format": "json"},
			setupMock: func(m *UniversalMockClient) {
				m.GetStateFn = func(_ context.Context, _ string) (*homeassistant.Entity, error) {
					return testEntityData, nil
				}
			},
			wantError:    false,
			wantContains: []string{"light.living_room", "on", "Living Room Light", "brightness", "255"},
		},
		{
			name:         "error - missing entity_id",
			args:         map[string]any{},
			setupMock:    nil,
			wantError:    true,
			wantContains: []string{"entity_id is required"},
		},
		{
			name:         "error - empty entity_id",
			args:         map[string]any{"entity_id": ""},
			setupMock:    nil,
			wantError:    true,
			wantContains: []string{"entity_id is required"},
		},
		{
			name: "error - client error",
			args: map[string]any{"entity_id": "light.nonexistent"},
			setupMock: func(m *UniversalMockClient) {
				m.GetStateFn = func(_ context.Context, _ string) (*homeassistant.Entity, error) {
					return nil, errors.New("entity not found")
				}
			},
			wantError:    true,
			wantContains: []string{"Error getting state", "entity not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &EntityHandlers{}
			client := &UniversalMockClient{}
			if tt.setupMock != nil {
				tt.setupMock(client)
			}

			result, err := h.handleGetState(context.Background(), client, tt.args)
			if err != nil {
				t.Fatalf("handleGetState() unexpected error = %v", err)
			}

			if result.IsError != tt.wantError {
				t.Errorf("IsError = %v, want %v", result.IsError, tt.wantError)
			}

			if len(result.Content) == 0 {
				t.Fatal("handleGetState() returned no content")
			}

			content := result.Content[0].Text
			assertContainsAll(t, content, tt.wantContains)
		})
	}
}
