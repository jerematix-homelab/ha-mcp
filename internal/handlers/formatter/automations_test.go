package formatter

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

func TestNaturalAutomationFormatter_FormatList(t *testing.T) {
	ctx := context.Background()
	f := NewNaturalAutomationFormatter()

	// Test data
	now := time.Now()
	fiveMinAgo := now.Add(-5 * time.Minute)
	oneHourAgo := now.Add(-1 * time.Hour)

	automations := []homeassistant.Automation{
		{
			EntityID:      "automation.motion_lights",
			State:         "on",
			FriendlyName:  "Motion Lights",
			LastTriggered: fiveMinAgo.Format(time.RFC3339),
		},
		{
			EntityID:      "automation.morning_routine",
			State:         "on",
			FriendlyName:  "Morning Routine",
			LastTriggered: oneHourAgo.Format(time.RFC3339),
		},
		{
			EntityID:     "automation.vacation_mode",
			State:        "off",
			FriendlyName: "Vacation Mode",
		},
	}

	configs := map[string]*homeassistant.AutomationConfig{
		"motion_lights": {
			Mode: "single",
		},
		"morning_routine": {
			Mode: "queued",
		},
		"vacation_mode": {
			Mode: "parallel",
		},
	}

	opts := AutomationListOptions{}
	result, err := f.FormatList(ctx, automations, configs, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify summary line
	if !strings.Contains(result, "3 automations") {
		t.Errorf("expected '3 automations' in output, got: %s", result)
	}
	if !strings.Contains(result, "2 enabled") {
		t.Errorf("expected '2 enabled' in output, got: %s", result)
	}
	if !strings.Contains(result, "1 disabled") {
		t.Errorf("expected '1 disabled' in output, got: %s", result)
	}

	// Verify mode breakdown
	if !strings.Contains(result, "By mode:") {
		t.Errorf("expected 'By mode:' in output, got: %s", result)
	}

	// Verify recently triggered section
	if !strings.Contains(result, "Recently triggered:") {
		t.Errorf("expected 'Recently triggered:' in output, got: %s", result)
	}
	if !strings.Contains(result, "Motion Lights") {
		t.Errorf("expected 'Motion Lights' in output, got: %s", result)
	}

	// Verify disabled section
	if !strings.Contains(result, "Disabled") {
		t.Errorf("expected 'Disabled' section in output, got: %s", result)
	}
	if !strings.Contains(result, "Vacation Mode") {
		t.Errorf("expected 'Vacation Mode' in disabled section, got: %s", result)
	}
}

func TestNaturalAutomationFormatter_FormatList_Empty(t *testing.T) {
	ctx := context.Background()
	f := NewNaturalAutomationFormatter()

	result, err := f.FormatList(ctx, nil, nil, AutomationListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != MsgNoAutomationsFound {
		t.Errorf("expected '%s', got: %s", MsgNoAutomationsFound, result)
	}
}

func TestNaturalAutomationFormatter_FormatList_Verbose(t *testing.T) {
	ctx := context.Background()
	f := NewNaturalAutomationFormatter()

	automations := []homeassistant.Automation{
		{
			EntityID:     "automation.test",
			State:        "on",
			FriendlyName: "Test Automation",
		},
	}

	configs := map[string]*homeassistant.AutomationConfig{
		"test": {
			Mode:     "single",
			Triggers: []any{map[string]any{"platform": "state", "entity_id": "binary_sensor.motion"}},
			Actions:  []any{map[string]any{"service": "light.turn_on", "target": map[string]any{"entity_id": "light.living"}}},
		},
	}

	opts := AutomationListOptions{Verbose: true}
	result, err := f.FormatList(ctx, automations, configs, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verbose should include triggers/actions info
	if !strings.Contains(result, "trigger") || !strings.Contains(result, "action") {
		t.Errorf("verbose output should mention triggers/actions, got: %s", result)
	}
}

func TestNaturalAutomationFormatter_FormatDetail(t *testing.T) {
	ctx := context.Background()
	f := NewNaturalAutomationFormatter()

	automation := homeassistant.Automation{
		EntityID:      "automation.motion_lights",
		State:         "on",
		FriendlyName:  "Motion Lights",
		LastTriggered: time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		Config: &homeassistant.AutomationConfig{
			ID:          "motion_lights",
			Alias:       "Motion Lights",
			Description: "Turn on lights when motion detected",
			Mode:        "single",
			Triggers: []any{
				map[string]any{
					"platform":  "state",
					"entity_id": "binary_sensor.motion_kitchen",
					"from":      "off",
					"to":        "on",
				},
			},
			Actions: []any{
				map[string]any{
					"service": "light.turn_on",
					"target":  map[string]any{"entity_id": "light.kitchen"},
					"data":    map[string]any{"brightness": 255},
				},
			},
		},
	}

	result, err := f.FormatDetail(ctx, automation)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify header
	if !strings.Contains(result, "Motion Lights") {
		t.Errorf("expected 'Motion Lights' in output, got: %s", result)
	}
	if !strings.Contains(result, "enabled") {
		t.Errorf("expected 'enabled' state in output, got: %s", result)
	}

	// Verify mode and last triggered
	if !strings.Contains(result, "Mode:") && !strings.Contains(result, "single") {
		t.Errorf("expected mode info in output, got: %s", result)
	}

	// Verify triggers section
	if !strings.Contains(result, "Trigger") {
		t.Errorf("expected 'Triggers' section in output, got: %s", result)
	}
	if !strings.Contains(result, "binary_sensor.motion_kitchen") {
		t.Errorf("expected trigger entity in output, got: %s", result)
	}

	// Verify actions section
	if !strings.Contains(result, "Action") {
		t.Errorf("expected 'Actions' section in output, got: %s", result)
	}
	if !strings.Contains(result, "light.turn_on") {
		t.Errorf("expected action service in output, got: %s", result)
	}
}

func TestNaturalAutomationFormatter_FormatDetail_NoConfig(t *testing.T) {
	ctx := context.Background()
	f := NewNaturalAutomationFormatter()

	automation := homeassistant.Automation{
		EntityID:     "automation.simple",
		State:        "on",
		FriendlyName: "Simple Automation",
	}

	result, err := f.FormatDetail(ctx, automation)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Simple Automation") {
		t.Errorf("expected 'Simple Automation' in output, got: %s", result)
	}
}

func TestNaturalAutomationFormatter_ImprovedTriggerFormatting(t *testing.T) {
	ctx := context.Background()
	f := NewNaturalAutomationFormatter()

	tests := []struct {
		name     string
		trigger  map[string]any
		wantText []string
	}{
		{
			name: "state trigger with only from",
			trigger: map[string]any{
				"platform":  "state",
				"entity_id": "light.test",
				"from":      "off",
			},
			wantText: []string{"state: light.test", "from: off"},
		},
		{
			name: "state trigger with only to",
			trigger: map[string]any{
				"platform":  "state",
				"entity_id": "light.test",
				"to":        "on",
			},
			wantText: []string{"state: light.test", "to: on"},
		},
		{
			name: "state trigger with for duration",
			trigger: map[string]any{
				"platform":  "state",
				"entity_id": "light.test",
				"to":        "on",
				"for":       "00:05:00",
			},
			wantText: []string{"state: light.test", "to: on", "for: 00:05:00"},
		},
		{
			name: "device trigger with type and subtype",
			trigger: map[string]any{
				"platform":  "device",
				"device_id": "device_123",
				"type":      "turned_on",
				"subtype":   "button_1",
			},
			wantText: []string{"device: device_123", "turned_on/button_1"},
		},
		{
			name: "template trigger with value",
			trigger: map[string]any{
				"platform":       "template",
				"value_template": "{{ states('sensor.temp') | float > 20 }}",
			},
			wantText: []string{"template:", "states('sensor.temp')"},
		},
		{
			name: "template trigger with long value",
			trigger: map[string]any{
				"platform":       "template",
				"value_template": "{{ states('sensor.temperature') | float > 20 and states('sensor.humidity') | float < 60 and is_state('binary_sensor.motion', 'on') }}",
			},
			wantText: []string{"template:", "..."},
		},
		{
			name: "numeric_state trigger with negative below threshold",
			trigger: map[string]any{
				"platform":  "numeric_state",
				"entity_id": "sensor.calc_netzleistung",
				"below":     float64(-1500),
			},
			wantText: []string{"numeric_state: sensor.calc_netzleistung", "below -1500"},
		},
		{
			name: "numeric_state trigger with for duration",
			trigger: map[string]any{
				"platform":  "numeric_state",
				"entity_id": "sensor.power",
				"below":     float64(-1500),
				"for":       "0:05:00",
			},
			wantText: []string{"below -1500", "for 0:05:00"},
		},
		{
			name: "numeric_state trigger with for as map",
			trigger: map[string]any{
				"platform":  "numeric_state",
				"entity_id": "sensor.power",
				"below":     float64(100),
				"for":       map[string]any{"hours": float64(0), "minutes": float64(5), "seconds": float64(0)},
			},
			wantText: []string{"below 100", "for 0:05:00"},
		},
		{
			name: "numeric_state trigger with template threshold",
			trigger: map[string]any{
				"platform":  "numeric_state",
				"entity_id": "sensor.power",
				"below":     "{{ states('input_number.threshold') }}",
			},
			wantText: []string{"below {{ states('input_number.threshold') }}"},
		},
		{
			name: "state trigger with for as map",
			trigger: map[string]any{
				"platform":  "state",
				"entity_id": "binary_sensor.motion",
				"to":        "off",
				"for":       map[string]any{"hours": float64(0), "minutes": float64(10), "seconds": float64(0)},
			},
			wantText: []string{"state: binary_sensor.motion", "to: off", "for: 0:10:00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			automation := homeassistant.Automation{
				EntityID:     "automation.test",
				State:        "on",
				FriendlyName: "Test",
				Config: &homeassistant.AutomationConfig{
					ID:       "test",
					Alias:    "Test",
					Triggers: []any{tt.trigger},
					Actions:  []any{},
				},
			}

			result, err := f.FormatDetail(ctx, automation)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, want := range tt.wantText {
				if !strings.Contains(result, want) {
					t.Errorf("expected %q in output, got: %s", want, result)
				}
			}
		})
	}
}

func TestNaturalAutomationFormatter_ImprovedActionFormatting(t *testing.T) {
	ctx := context.Background()
	f := NewNaturalAutomationFormatter()

	tests := []struct {
		name     string
		action   map[string]any
		wantText []string
	}{
		{
			name: "choose action with options",
			action: map[string]any{
				"choose": []any{
					map[string]any{
						"alias":      "Morning mode",
						"conditions": []any{},
						"sequence":   []any{},
					},
					map[string]any{
						"alias":      "Evening mode",
						"conditions": []any{},
						"sequence":   []any{},
					},
				},
			},
			wantText: []string{"choose:", "2 option"},
		},
		{
			name: "if action with conditions",
			action: map[string]any{
				"if": []any{
					map[string]any{"condition": "state", "entity_id": "light.test", "state": "on"},
				},
				"then": []any{},
			},
			wantText: []string{"conditional:", "if/then"},
		},
		{
			name: "modern action key with target entity_id",
			action: map[string]any{
				"action": "script.turn_on",
				"target": map[string]any{
					"entity_id": "script.kitchen_dishwasher_on",
				},
			},
			wantText: []string{"script.turn_on", "script.kitchen_dishwasher_on"},
		},
		{
			name: "modern action key without target",
			action: map[string]any{
				"action": "homeassistant.reload_config_entry",
			},
			wantText: []string{"homeassistant.reload_config_entry"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			automation := homeassistant.Automation{
				EntityID:     "automation.test",
				State:        "on",
				FriendlyName: "Test",
				Config: &homeassistant.AutomationConfig{
					ID:       "test",
					Alias:    "Test",
					Triggers: []any{},
					Actions:  []any{tt.action},
				},
			}

			result, err := f.FormatDetail(ctx, automation)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, want := range tt.wantText {
				if !strings.Contains(result, want) {
					t.Errorf("expected %q in output, got: %s", want, result)
				}
			}
		})
	}
}

func TestNaturalAutomationFormatter_NumericStateCondition(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	f := NewNaturalAutomationFormatter()

	tests := []struct {
		name      string
		condition map[string]any
		wantText  []string
	}{
		{
			name: "numeric_state condition with negative below",
			condition: map[string]any{
				"condition": "numeric_state",
				"entity_id": "sensor.power",
				"below":     float64(-500),
			},
			wantText: []string{"numeric_state: sensor.power", "below -500"},
		},
		{
			name: "numeric_state condition with above and below",
			condition: map[string]any{
				"condition": "numeric_state",
				"entity_id": "sensor.humidity",
				"above":     float64(30),
				"below":     float64(70),
			},
			wantText: []string{"numeric_state: sensor.humidity", "above 30", "below 70"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			automation := homeassistant.Automation{
				EntityID:     "automation.test",
				State:        "on",
				FriendlyName: "Test",
				Config: &homeassistant.AutomationConfig{
					ID:         "test",
					Alias:      "Test",
					Triggers:   []any{},
					Conditions: []any{tt.condition},
					Actions:    []any{},
				},
			}

			result, err := f.FormatDetail(ctx, automation)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, want := range tt.wantText {
				if !strings.Contains(result, want) {
					t.Errorf("expected %q in output, got: %s", want, result)
				}
			}
		})
	}
}

func TestJSONAutomationFormatter_FormatList(t *testing.T) {
	ctx := context.Background()
	f := NewJSONAutomationFormatter()

	automations := []homeassistant.Automation{
		{
			EntityID:     "automation.test",
			State:        "on",
			FriendlyName: "Test",
		},
	}

	result, err := f.FormatList(ctx, automations, nil, AutomationListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be valid JSON
	var parsed []map[string]any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Errorf("expected valid JSON array, got error: %v", err)
	}

	if len(parsed) != 1 {
		t.Errorf("expected 1 item in JSON, got: %d", len(parsed))
	}
}

func TestJSONAutomationFormatter_FormatList_Empty(t *testing.T) {
	ctx := context.Background()
	f := NewJSONAutomationFormatter()

	result, err := f.FormatList(ctx, nil, nil, AutomationListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be empty JSON array
	if result != "[]" {
		t.Errorf("expected '[]', got: %s", result)
	}
}

func TestJSONAutomationFormatter_FormatDetail(t *testing.T) {
	ctx := context.Background()
	f := NewJSONAutomationFormatter()

	automation := homeassistant.Automation{
		EntityID:     "automation.test",
		State:        "on",
		FriendlyName: "Test",
		Config: &homeassistant.AutomationConfig{
			Mode: "single",
		},
	}

	result, err := f.FormatDetail(ctx, automation)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be valid JSON
	var parsed map[string]any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Errorf("expected valid JSON object, got error: %v", err)
	}

	if parsed["entity_id"] != "automation.test" {
		t.Errorf("expected entity_id 'automation.test', got: %v", parsed["entity_id"])
	}
}

func TestNewAutomationFormatter(t *testing.T) {
	natural := NewAutomationFormatter(FormatNatural)
	if _, ok := natural.(*NaturalAutomationFormatter); !ok {
		t.Errorf("expected NaturalAutomationFormatter for natural format")
	}

	jsonFmt := NewAutomationFormatter(FormatJSON)
	if _, ok := jsonFmt.(*JSONAutomationFormatter); !ok {
		t.Errorf("expected JSONAutomationFormatter for json format")
	}

	defaultFmt := NewAutomationFormatter("")
	if _, ok := defaultFmt.(*NaturalAutomationFormatter); !ok {
		t.Errorf("expected NaturalAutomationFormatter for empty format")
	}
}

func TestFormatRepeatAction(t *testing.T) {
	tests := []struct {
		name        string
		repeat      any
		wantContain string
	}{
		{
			name:        "while loop with conditions and sequence",
			repeat:      map[string]any{"while": []any{"cond1", "cond2"}, "sequence": []any{"action1", "action2", "action3"}},
			wantContain: "while [2 condition(s)]: 3 action(s)",
		},
		{
			name:        "until loop",
			repeat:      map[string]any{"until": []any{"cond1"}, "sequence": []any{"action1"}},
			wantContain: "until [1 condition(s)]: 1 action(s)",
		},
		{
			name:        "count loop",
			repeat:      map[string]any{"count": 5, "sequence": []any{"action1", "action2"}},
			wantContain: "repeat 5 times: 2 action(s)",
		},
		{
			name:        "no type specified",
			repeat:      map[string]any{"sequence": []any{"action1"}},
			wantContain: "repeat: 1 action(s)",
		},
		{
			name:        "invalid repeat value falls back",
			repeat:      "not a map",
			wantContain: "repeat action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRepeatAction(tt.repeat)
			if !strings.Contains(result, tt.wantContain) {
				t.Errorf("expected %q to contain %q", result, tt.wantContain)
			}
		})
	}
}
