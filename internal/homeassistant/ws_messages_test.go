package homeassistant

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWSCommandWithPayload_MarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cmd     WSCommandWithPayload
		want    map[string]any
		wantErr bool
	}{
		{
			name: "basic command with payload",
			cmd: WSCommandWithPayload{
				ID:   1,
				Type: "test_command",
				Payload: map[string]any{
					"key1": "value1",
					"key2": float64(42),
				},
			},
			want: map[string]any{
				"id":   float64(1),
				"type": "test_command",
				"key1": "value1",
				"key2": float64(42),
			},
			wantErr: false,
		},
		{
			name: "command with empty payload",
			cmd: WSCommandWithPayload{
				ID:      2,
				Type:    "empty_payload",
				Payload: map[string]any{},
			},
			want: map[string]any{
				"id":   float64(2),
				"type": "empty_payload",
			},
			wantErr: false,
		},
		{
			name: "command with nil payload",
			cmd: WSCommandWithPayload{
				ID:      3,
				Type:    "nil_payload",
				Payload: nil,
			},
			want: map[string]any{
				"id":   float64(3),
				"type": "nil_payload",
			},
			wantErr: false,
		},
		{
			name: "command with nested payload",
			cmd: WSCommandWithPayload{
				ID:   4,
				Type: "nested_command",
				Payload: map[string]any{
					"nested": map[string]any{
						"inner_key": "inner_value",
					},
					"array": []string{"a", "b", "c"},
				},
			},
			want: map[string]any{
				"id":   float64(4),
				"type": "nested_command",
				"nested": map[string]any{
					"inner_key": "inner_value",
				},
				"array": []any{"a", "b", "c"},
			},
			wantErr: false,
		},
		{
			name: "command with boolean payload",
			cmd: WSCommandWithPayload{
				ID:   5,
				Type: "bool_command",
				Payload: map[string]any{
					"enabled":  true,
					"disabled": false,
				},
			},
			want: map[string]any{
				"id":       float64(5),
				"type":     "bool_command",
				"enabled":  true,
				"disabled": false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.cmd.MarshalJSON()

			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			var gotMap map[string]any
			if unmarshalErr := json.Unmarshal(got, &gotMap); unmarshalErr != nil {
				t.Errorf("Failed to unmarshal result: %v", unmarshalErr)
				return
			}

			if diff := cmp.Diff(tt.want, gotMap); diff != "" {
				t.Errorf("MarshalJSON() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseMessageType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    []byte
		want    string
		wantErr bool
	}{
		{
			name:    "valid auth_required message",
			data:    []byte(`{"type":"auth_required","ha_version":"2024.1.0"}`),
			want:    "auth_required",
			wantErr: false,
		},
		{
			name:    "valid auth_ok message",
			data:    []byte(`{"type":"auth_ok","ha_version":"2024.1.0"}`),
			want:    "auth_ok",
			wantErr: false,
		},
		{
			name:    "valid result message",
			data:    []byte(`{"id":1,"type":"result","success":true}`),
			want:    "result",
			wantErr: false,
		},
		{
			name:    "valid event message",
			data:    []byte(`{"id":1,"type":"event","event":{}}`),
			want:    "event",
			wantErr: false,
		},
		{
			name:    "empty type",
			data:    []byte(`{"type":"","other":"value"}`),
			want:    "",
			wantErr: false,
		},
		{
			name:    "missing type field",
			data:    []byte(`{"other":"value"}`),
			want:    "",
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    []byte(`{invalid json`),
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty JSON object",
			data:    []byte(`{}`),
			want:    "",
			wantErr: false,
		},
		{
			name:    "null value",
			data:    []byte(`null`),
			want:    "",
			wantErr: false,
		},
		{
			name:    "empty input",
			data:    []byte(``),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseMessageType(tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMessageType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("ParseMessageType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMessageID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    []byte
		want    int64
		wantErr bool
	}{
		{
			name:    "valid ID",
			data:    []byte(`{"id":1,"type":"result"}`),
			want:    1,
			wantErr: false,
		},
		{
			name:    "large ID",
			data:    []byte(`{"id":9223372036854775807,"type":"result"}`),
			want:    9223372036854775807,
			wantErr: false,
		},
		{
			name:    "zero ID",
			data:    []byte(`{"id":0,"type":"result"}`),
			want:    0,
			wantErr: false,
		},
		{
			name:    "negative ID",
			data:    []byte(`{"id":-1,"type":"result"}`),
			want:    -1,
			wantErr: false,
		},
		{
			name:    "missing ID field",
			data:    []byte(`{"type":"auth_required"}`),
			want:    0,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    []byte(`{invalid json`),
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty JSON object",
			data:    []byte(`{}`),
			want:    0,
			wantErr: false,
		},
		{
			name:    "null value",
			data:    []byte(`null`),
			want:    0,
			wantErr: false,
		},
		{
			name:    "empty input",
			data:    []byte(``),
			want:    0,
			wantErr: true,
		},
		{
			name:    "ID with other fields",
			data:    []byte(`{"id":42,"type":"event","success":true}`),
			want:    42,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseMessageID(tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMessageID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("ParseMessageID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWSMessage_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  WSMessage
	}{
		{
			name: "basic message",
			msg:  WSMessage{ID: 1, Type: "test"},
		},
		{
			name: "message without ID",
			msg:  WSMessage{Type: "auth"},
		},
		{
			name: "message with zero ID",
			msg:  WSMessage{ID: 0, Type: "zero"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			var got WSMessage
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if diff := cmp.Diff(tt.msg, got); diff != "" {
				t.Errorf("Roundtrip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWSAuthMessage_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	msg := WSAuthMessage{
		Type:        "auth",
		AccessToken: "test_token_123",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got WSAuthMessage
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if diff := cmp.Diff(msg, got); diff != "" {
		t.Errorf("Roundtrip mismatch (-want +got):\n%s", diff)
	}
}

func TestWSResultMessage_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  WSResultMessage
	}{
		{
			name: "successful result",
			msg: WSResultMessage{
				ID:      1,
				Type:    "result",
				Success: true,
				Result:  json.RawMessage(`{"key":"value"}`),
			},
		},
		{
			name: "error result",
			msg: WSResultMessage{
				ID:      2,
				Type:    "result",
				Success: false,
				Error: &WSError{
					Code:    "invalid_request",
					Message: "Invalid request",
				},
			},
		},
		{
			name: "result with nil error and empty result",
			msg: WSResultMessage{
				ID:      3,
				Type:    "result",
				Success: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			var got WSResultMessage
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if diff := cmp.Diff(tt.msg, got); diff != "" {
				t.Errorf("Roundtrip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWSAuthRequired_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	msg := WSAuthRequired{
		Type:      "auth_required",
		HAVersion: "2024.1.0",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got WSAuthRequired
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if diff := cmp.Diff(msg, got); diff != "" {
		t.Errorf("Roundtrip mismatch (-want +got):\n%s", diff)
	}
}

func TestWSAuthOK_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	msg := WSAuthOK{
		Type:      "auth_ok",
		HAVersion: "2024.1.0",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got WSAuthOK
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if diff := cmp.Diff(msg, got); diff != "" {
		t.Errorf("Roundtrip mismatch (-want +got):\n%s", diff)
	}
}

func TestWSAuthInvalid_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	msg := WSAuthInvalid{
		Type:    "auth_invalid",
		Message: "Invalid access token",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got WSAuthInvalid
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if diff := cmp.Diff(msg, got); diff != "" {
		t.Errorf("Roundtrip mismatch (-want +got):\n%s", diff)
	}
}

func TestWSCommand_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	msg := WSCommand{
		ID:   1,
		Type: "get_states",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got WSCommand
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if diff := cmp.Diff(msg, got); diff != "" {
		t.Errorf("Roundtrip mismatch (-want +got):\n%s", diff)
	}
}

func TestWSError_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	msg := WSError{
		Code:    "invalid_format",
		Message: "Invalid message format",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got WSError
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if diff := cmp.Diff(msg, got); diff != "" {
		t.Errorf("Roundtrip mismatch (-want +got):\n%s", diff)
	}
}
