package formatter

import (
	"testing"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Format
	}{
		{
			name:     "natural explicit",
			input:    "natural",
			expected: FormatNatural,
		},
		{
			name:     "json explicit",
			input:    "json",
			expected: FormatJSON,
		},
		{
			name:     "empty string defaults to natural",
			input:    "",
			expected: FormatNatural,
		},
		{
			name:     "unknown defaults to natural",
			input:    "xml",
			expected: FormatNatural,
		},
		{
			name:     "case sensitive - JSON not recognized",
			input:    "JSON",
			expected: FormatNatural,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseFormat(tt.input)
			if result != tt.expected {
				t.Errorf("ParseFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		format         Format
		expectedType   string
		notExpectedNil bool
	}{
		{
			name:           "natural format creates NaturalFormatter",
			format:         FormatNatural,
			expectedType:   "*formatter.NaturalFormatter",
			notExpectedNil: true,
		},
		{
			name:           "json format creates JSONFormatter",
			format:         FormatJSON,
			expectedType:   "*formatter.JSONFormatter",
			notExpectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := New(tt.format)
			if result == nil {
				t.Fatal("New() returned nil")
			}

			switch tt.format {
			case FormatNatural:
				if _, ok := result.(*NaturalFormatter); !ok {
					t.Errorf("New(%v) did not return *NaturalFormatter", tt.format)
				}
			case FormatJSON:
				if _, ok := result.(*JSONFormatter); !ok {
					t.Errorf("New(%v) did not return *JSONFormatter", tt.format)
				}
			}
		})
	}
}

func TestNewRegistryFormatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format Format
	}{
		{name: "natural", format: FormatNatural},
		{name: "json", format: FormatJSON},
		{name: "unknown defaults to natural", format: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NewRegistryFormatter(tt.format)
			if result == nil {
				t.Fatal("NewRegistryFormatter() returned nil")
			}

			switch tt.format {
			case FormatJSON:
				if _, ok := result.(*JSONRegistryFormatter); !ok {
					t.Errorf("NewRegistryFormatter(%v) did not return *JSONRegistryFormatter", tt.format)
				}
			default:
				if _, ok := result.(*NaturalRegistryFormatter); !ok {
					t.Errorf("NewRegistryFormatter(%v) did not return *NaturalRegistryFormatter", tt.format)
				}
			}
		})
	}
}

func TestNewTargetFormatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format Format
	}{
		{name: "natural", format: FormatNatural},
		{name: "json", format: FormatJSON},
		{name: "unknown defaults to natural", format: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NewTargetFormatter(tt.format)
			if result == nil {
				t.Fatal("NewTargetFormatter() returned nil")
			}

			switch tt.format {
			case FormatJSON:
				if _, ok := result.(*JSONTargetFormatter); !ok {
					t.Errorf("NewTargetFormatter(%v) did not return *JSONTargetFormatter", tt.format)
				}
			default:
				if _, ok := result.(*NaturalTargetFormatter); !ok {
					t.Errorf("NewTargetFormatter(%v) did not return *NaturalTargetFormatter", tt.format)
				}
			}
		})
	}
}
