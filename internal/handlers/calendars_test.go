package handlers

import (
	"context"
	"testing"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// TestManageCalendarSchema verifies the schema for manage_calendar tool.
func TestManageCalendarSchema(t *testing.T) {
	t.Parallel()

	registry := mcp.NewRegistry()
	RegisterCalendarTools(registry)

	tool, exists := registry.GetTool("manage_calendar")
	if !exists {
		t.Fatal("manage_calendar tool not registered")
	}

	// Verify basic properties
	if tool.Name != "manage_calendar" {
		t.Errorf("tool.Name = %q, want %q", tool.Name, "manage_calendar")
	}

	// Verify schema
	schema := tool.InputSchema
	props := schema.Properties

	// Check action field
	actionSchema, ok := props["action"]
	if !ok {
		t.Fatal("action property missing from schema")
	}
	if len(actionSchema.Enum) != 4 {
		t.Errorf("action enum count = %d, want 4 (list, get_events, create_event, delete_event)", len(actionSchema.Enum))
	}
}

// TestManageCalendar_List verifies list action.
func TestManageCalendar_List(t *testing.T) {
	t.Parallel()

	client := &UniversalMockClient{
		GetCalendarsFn: func(context.Context) ([]homeassistant.CalendarEntry, error) {
			return []homeassistant.CalendarEntry{
				{EntityID: "calendar.holidays", Name: "Holidays"},
				{EntityID: "calendar.personal", Name: "Personal"},
			}, nil
		},
	}

	handler := NewCalendarHandlers()
	result, err := handler.HandleManageCalendar(context.Background(), client, map[string]any{
		"action": "list",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected result content")
	}

	text := result.Content[0].Text
	if !contains(text, "Holidays") {
		t.Errorf("result text does not contain 'Holidays': %s", text)
	}
}

// TestManageCalendar_GetEvents verifies get_events action.
func TestManageCalendar_GetEvents(t *testing.T) {
	t.Parallel()

	client := &UniversalMockClient{
		GetCalendarEventsFn: func(context.Context, string, string, string) ([]homeassistant.CalendarEvent, error) {
			return []homeassistant.CalendarEvent{
				{
					Start: homeassistant.CalendarDateTime{
						DateTime: "2024-01-15T10:00:00Z",
					},
					End: homeassistant.CalendarDateTime{
						DateTime: "2024-01-15T11:00:00Z",
					},
					Summary:     "Team Meeting",
					Description: "Weekly sync",
				},
			}, nil
		},
	}

	handler := NewCalendarHandlers()
	result, err := handler.HandleManageCalendar(context.Background(), client, map[string]any{
		"action":    "get_events",
		"entity_id": "calendar.work",
		"start":     "2024-01-15T00:00:00Z",
		"end":       "2024-01-16T00:00:00Z",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected result content")
	}

	text := result.Content[0].Text
	if !contains(text, "Team Meeting") {
		t.Errorf("result text does not contain 'Team Meeting': %s", text)
	}
}

// TestManageCalendar_GetEventsMissingParams verifies validation.
func TestManageCalendar_GetEventsMissingParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args map[string]any
	}{
		{
			name: "missing entity_id",
			args: map[string]any{
				"action": "get_events",
				"start":  "2024-01-15T00:00:00Z",
				"end":    "2024-01-16T00:00:00Z",
			},
		},
		{
			name: "missing start",
			args: map[string]any{
				"action":    "get_events",
				"entity_id": "calendar.work",
				"end":       "2024-01-16T00:00:00Z",
			},
		},
		{
			name: "missing end",
			args: map[string]any{
				"action":    "get_events",
				"entity_id": "calendar.work",
				"start":     "2024-01-15T00:00:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &UniversalMockClient{}
			handler := NewCalendarHandlers()

			result, err := handler.HandleManageCalendar(context.Background(), client, tt.args)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil || !result.IsError {
				t.Error("expected error result")
			}
		})
	}
}

// TestManageCalendar_CreateEvent verifies create_event action.
func TestManageCalendar_CreateEvent(t *testing.T) {
	t.Parallel()

	client := &UniversalMockClient{
		CallServiceFn: func(context.Context, string, string, map[string]any) ([]homeassistant.Entity, error) {
			return nil, nil
		},
	}

	handler := NewCalendarHandlers()
	result, err := handler.HandleManageCalendar(context.Background(), client, map[string]any{
		"action":          "create_event",
		"entity_id":       "calendar.personal",
		"summary":         "Doctor Appointment",
		"start_date_time": "2024-01-20T14:00:00Z",
		"end_date_time":   "2024-01-20T15:00:00Z",
		"description":     "Annual checkup",
		"location":        "Medical Center",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected success result")
	}

	text := result.Content[0].Text
	if !contains(text, "created") {
		t.Errorf("result does not indicate creation: %s", text)
	}
}

// TestManageCalendar_DeleteEvent verifies delete_event action.
func TestManageCalendar_DeleteEvent(t *testing.T) {
	t.Parallel()

	client := &UniversalMockClient{
		CallServiceFn: func(context.Context, string, string, map[string]any) ([]homeassistant.Entity, error) {
			return nil, nil
		},
	}

	handler := NewCalendarHandlers()
	result, err := handler.HandleManageCalendar(context.Background(), client, map[string]any{
		"action":    "delete_event",
		"entity_id": "calendar.personal",
		"uid":       "event123",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected success result")
	}

	text := result.Content[0].Text
	if !contains(text, "deleted") {
		t.Errorf("result does not indicate deletion: %s", text)
	}
}
