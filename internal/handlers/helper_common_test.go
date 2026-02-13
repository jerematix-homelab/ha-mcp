package handlers

import (
	"testing"
)

func TestVerboseHint(t *testing.T) {
	t.Parallel()

	if VerboseHint == "" {
		t.Error("VerboseHint is empty, want non-empty")
	}
}

func TestHelperPlatforms(t *testing.T) {
	t.Parallel()

	if len(HelperPlatforms) == 0 {
		t.Error("HelperPlatforms is empty, want non-empty")
	}

	// Check that expected platforms are present
	expectedPlatforms := []string{
		"input_boolean",
		"input_number",
		"input_text",
		"input_select",
		"input_datetime",
		"input_button",
		"counter",
		"timer",
		"schedule",
	}

	for _, expected := range expectedPlatforms {
		found := false
		for _, platform := range HelperPlatforms {
			if platform == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("HelperPlatforms missing expected platform %q", expected)
		}
	}
}

func TestParseHelperEntityID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		entityID     string
		wantPlatform string
		wantID       string
	}{
		{
			name:         "input_boolean entity",
			entityID:     "input_boolean.my_switch",
			wantPlatform: "input_boolean",
			wantID:       "my_switch",
		},
		{
			name:         "input_number entity",
			entityID:     "input_number.temperature_setting",
			wantPlatform: "input_number",
			wantID:       "temperature_setting",
		},
		{
			name:         "input_text entity",
			entityID:     "input_text.user_name",
			wantPlatform: "input_text",
			wantID:       "user_name",
		},
		{
			name:         "input_select entity",
			entityID:     "input_select.mode_selector",
			wantPlatform: "input_select",
			wantID:       "mode_selector",
		},
		{
			name:         "input_datetime entity",
			entityID:     "input_datetime.alarm_time",
			wantPlatform: "input_datetime",
			wantID:       "alarm_time",
		},
		{
			name:         "input_button entity",
			entityID:     "input_button.restart_button",
			wantPlatform: "input_button",
			wantID:       "restart_button",
		},
		{
			name:         "counter entity",
			entityID:     "counter.visit_count",
			wantPlatform: "counter",
			wantID:       "visit_count",
		},
		{
			name:         "timer entity",
			entityID:     "timer.cooking_timer",
			wantPlatform: "timer",
			wantID:       "cooking_timer",
		},
		{
			name:         "schedule entity",
			entityID:     "schedule.weekly_schedule",
			wantPlatform: "schedule",
			wantID:       "weekly_schedule",
		},
		{
			name:         "group entity",
			entityID:     "group.all_lights",
			wantPlatform: "group",
			wantID:       "all_lights",
		},
		{
			name:         "threshold entity",
			entityID:     "threshold.temperature_high",
			wantPlatform: "threshold",
			wantID:       "temperature_high",
		},
		{
			name:         "derivative entity",
			entityID:     "derivative.power_rate",
			wantPlatform: "derivative",
			wantID:       "power_rate",
		},
		{
			name:         "integration entity",
			entityID:     "integration.energy_total",
			wantPlatform: "integration",
			wantID:       "energy_total",
		},
		{
			name:         "template entity",
			entityID:     "template.calculated_value",
			wantPlatform: "template",
			wantID:       "calculated_value",
		},
		{
			name:         "unknown platform",
			entityID:     "light.living_room",
			wantPlatform: "",
			wantID:       "",
		},
		{
			name:         "empty entity_id",
			entityID:     "",
			wantPlatform: "",
			wantID:       "",
		},
		{
			name:         "no dot separator",
			entityID:     "input_boolean_my_switch",
			wantPlatform: "",
			wantID:       "",
		},
		{
			name:         "entity with multiple dots",
			entityID:     "input_boolean.my.switch.test",
			wantPlatform: "input_boolean",
			wantID:       "my.switch.test",
		},
		{
			name:         "entity with underscores",
			entityID:     "input_number.my_long_entity_name",
			wantPlatform: "input_number",
			wantID:       "my_long_entity_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotPlatform, gotID := ParseHelperEntityID(tt.entityID)

			if gotPlatform != tt.wantPlatform {
				t.Errorf("ParseHelperEntityID(%q) platform = %q, want %q", tt.entityID, gotPlatform, tt.wantPlatform)
			}
			if gotID != tt.wantID {
				t.Errorf("ParseHelperEntityID(%q) id = %q, want %q", tt.entityID, gotID, tt.wantID)
			}
		})
	}
}
