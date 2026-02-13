package handlers

import (
	"strings"
	"testing"
)

func TestValidateEntityID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		entityID    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid entity ID",
			entityID: "light.living_room",
			wantErr:  false,
		},
		{
			name:     "valid entity ID with numbers",
			entityID: "sensor.temperature_1",
			wantErr:  false,
		},
		{
			name:     "valid input_number entity",
			entityID: "input_number.target_temp",
			wantErr:  false,
		},
		{
			name:        "empty entity ID",
			entityID:    "",
			wantErr:     true,
			errContains: "required",
		},
		{
			name:        "missing dot separator",
			entityID:    "lightliving_room",
			wantErr:     true,
			errContains: "domain.object_id",
		},
		{
			name:        "uppercase characters",
			entityID:    "Light.Living_Room",
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "spaces in entity ID",
			entityID:    "light.living room",
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "special characters",
			entityID:    "light.living-room",
			wantErr:     true,
			errContains: "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateEntityID(tt.entityID)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		min       float64
		max       float64
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid range min < max",
			min:       0,
			max:       100,
			fieldName: "input_number",
			wantErr:   false,
		},
		{
			name:      "valid range min = max",
			min:       50,
			max:       50,
			fieldName: "input_number",
			wantErr:   false,
		},
		{
			name:      "valid range with negative values",
			min:       -100,
			max:       100,
			fieldName: "temperature",
			wantErr:   false,
		},
		{
			name:      "invalid range min > max",
			min:       100,
			max:       0,
			fieldName: "input_number",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateRange(tt.min, tt.max, tt.fieldName)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetOptionalString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		args   map[string]any
		key    string
		want   string
		wantOK bool
	}{
		{
			name:   "key present with value",
			args:   map[string]any{"name": "test"},
			key:    "name",
			want:   "test",
			wantOK: true,
		},
		{
			name:   "key missing",
			args:   map[string]any{"other": "value"},
			key:    "name",
			want:   "",
			wantOK: false,
		},
		{
			name:   "key present but empty",
			args:   map[string]any{"name": ""},
			key:    "name",
			want:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := GetOptionalString(tt.args, tt.key)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, wantOK = %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
