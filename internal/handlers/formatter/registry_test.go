package formatter

import (
	"context"
	"strings"
	"testing"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

func TestNaturalRegistryFormatter_FormatEntityRegistry_Empty(t *testing.T) {
	f := NewNaturalRegistryFormatter()
	result, err := f.FormatEntityRegistry(context.Background(), nil, RegistryOptions{})
	if err != nil {
		t.Fatalf("FormatEntityRegistry() error = %v", err)
	}
	if !strings.Contains(result, "No entities found") {
		t.Errorf("FormatEntityRegistry() = %q, want to contain 'No entities found'", result)
	}
}

func TestNaturalRegistryFormatter_FormatEntityRegistry_WithEntries(t *testing.T) {
	f := NewNaturalRegistryFormatter()

	entries := []homeassistant.EntityRegistryEntry{
		{EntityID: "light.living_room", Platform: "hue", Name: "Living Room Light"},
		{EntityID: "light.bedroom", Platform: "hue", Name: "Bedroom Light"},
		{EntityID: "sensor.temperature", Platform: "mqtt", Name: "Temperature"},
		{EntityID: "sensor.humidity", Platform: "mqtt", Name: "Humidity", DisabledBy: "user"},
	}

	result, err := f.FormatEntityRegistry(context.Background(), entries, RegistryOptions{
		Verbose: false,
	})
	if err != nil {
		t.Fatalf("FormatEntityRegistry() error = %v", err)
	}

	// Should contain total count
	if !strings.Contains(result, "4 entities") {
		t.Errorf("FormatEntityRegistry() should contain total count, got %q", result)
	}

	// Should contain enabled count (3 enabled, 1 disabled)
	if !strings.Contains(result, "3 enabled") {
		t.Errorf("FormatEntityRegistry() should show enabled count, got %q", result)
	}

	// Should contain domain breakdown
	if !strings.Contains(result, "light: 2") {
		t.Errorf("FormatEntityRegistry() should contain light count, got %q", result)
	}
}

func TestNaturalRegistryFormatter_FormatEntityRegistry_Verbose(t *testing.T) {
	f := NewNaturalRegistryFormatter()

	entries := []homeassistant.EntityRegistryEntry{
		{EntityID: "light.living_room", Platform: "hue", Name: "Living Room Light", AreaID: "living_room"},
		{EntityID: "sensor.temperature", Platform: "mqtt", Name: "Temperature"},
	}

	result, err := f.FormatEntityRegistry(context.Background(), entries, RegistryOptions{
		Verbose: true,
	})
	if err != nil {
		t.Fatalf("FormatEntityRegistry() error = %v", err)
	}

	// Should contain entity details
	if !strings.Contains(result, "light.living_room") {
		t.Errorf("FormatEntityRegistry() should contain entity_id, got %q", result)
	}
	if !strings.Contains(result, "Living Room Light") {
		t.Errorf("FormatEntityRegistry() should contain friendly name, got %q", result)
	}
	if !strings.Contains(result, "living_room") {
		t.Errorf("FormatEntityRegistry() should contain area, got %q", result)
	}
}

func TestNaturalRegistryFormatter_FormatDeviceRegistry_Empty(t *testing.T) {
	f := NewNaturalRegistryFormatter()
	result, err := f.FormatDeviceRegistry(context.Background(), nil, RegistryOptions{})
	if err != nil {
		t.Fatalf("FormatDeviceRegistry() error = %v", err)
	}
	if !strings.Contains(result, "No devices found") {
		t.Errorf("FormatDeviceRegistry() = %q, want to contain 'No devices found'", result)
	}
}

func TestNaturalRegistryFormatter_FormatDeviceRegistry_WithEntries(t *testing.T) {
	f := NewNaturalRegistryFormatter()

	entries := []homeassistant.DeviceRegistryEntry{
		{ID: "dev1", Name: "Hue Bridge", Manufacturer: "Philips", Model: "BSB002"},
		{ID: "dev2", Name: "Living Room Bulb", Manufacturer: "Philips", Model: "LWA001"},
		{ID: "dev3", Name: "Smart Plug", Manufacturer: "Sonoff", Model: "S31"},
	}

	result, err := f.FormatDeviceRegistry(context.Background(), entries, RegistryOptions{})
	if err != nil {
		t.Fatalf("FormatDeviceRegistry() error = %v", err)
	}

	// Should contain total count
	if !strings.Contains(result, "3 devices") {
		t.Errorf("FormatDeviceRegistry() should contain total count, got %q", result)
	}

	// Should contain manufacturer breakdown
	if !strings.Contains(result, "Philips: 2") {
		t.Errorf("FormatDeviceRegistry() should contain Philips count, got %q", result)
	}
	if !strings.Contains(result, "Sonoff: 1") {
		t.Errorf("FormatDeviceRegistry() should contain Sonoff count, got %q", result)
	}
}

func TestNaturalRegistryFormatter_FormatDeviceRegistry_Verbose(t *testing.T) {
	f := NewNaturalRegistryFormatter()

	entries := []homeassistant.DeviceRegistryEntry{
		{ID: "dev1", Name: "Hue Bridge", Manufacturer: "Philips", Model: "BSB002", AreaID: "living_room"},
	}

	result, err := f.FormatDeviceRegistry(context.Background(), entries, RegistryOptions{
		Verbose: true,
	})
	if err != nil {
		t.Fatalf("FormatDeviceRegistry() error = %v", err)
	}

	// Should contain device details
	if !strings.Contains(result, "Hue Bridge") {
		t.Errorf("FormatDeviceRegistry() should contain device name, got %q", result)
	}
	if !strings.Contains(result, "BSB002") {
		t.Errorf("FormatDeviceRegistry() should contain model, got %q", result)
	}
}

func TestNaturalRegistryFormatter_FormatAreaRegistry_Empty(t *testing.T) {
	f := NewNaturalRegistryFormatter()
	result, err := f.FormatAreaRegistry(context.Background(), nil)
	if err != nil {
		t.Fatalf("FormatAreaRegistry() error = %v", err)
	}
	if !strings.Contains(result, "No areas found") {
		t.Errorf("FormatAreaRegistry() = %q, want to contain 'No areas found'", result)
	}
}

func TestNaturalRegistryFormatter_FormatAreaRegistry_WithEntries(t *testing.T) {
	f := NewNaturalRegistryFormatter()

	entries := []homeassistant.AreaRegistryEntry{
		{AreaID: "living_room", Name: "Living Room"},
		{AreaID: "kitchen", Name: "Kitchen"},
		{AreaID: "bedroom", Name: "Bedroom"},
	}

	result, err := f.FormatAreaRegistry(context.Background(), entries)
	if err != nil {
		t.Fatalf("FormatAreaRegistry() error = %v", err)
	}

	// Should contain total count
	if !strings.Contains(result, "3 areas") {
		t.Errorf("FormatAreaRegistry() should contain total count, got %q", result)
	}

	// Should list areas
	if !strings.Contains(result, "Living Room") {
		t.Errorf("FormatAreaRegistry() should contain area name, got %q", result)
	}
	if !strings.Contains(result, "living_room") {
		t.Errorf("FormatAreaRegistry() should contain area_id, got %q", result)
	}
}

func TestNaturalRegistryFormatter_FormatAllRegistries(t *testing.T) {
	f := NewNaturalRegistryFormatter()

	entities := []homeassistant.EntityRegistryEntry{
		{EntityID: "light.living_room", Platform: "hue"},
		{EntityID: "sensor.temperature", Platform: "mqtt"},
	}

	devices := []homeassistant.DeviceRegistryEntry{
		{ID: "dev1", Name: "Hue Bridge", Manufacturer: "Philips"},
	}

	areas := []homeassistant.AreaRegistryEntry{
		{AreaID: "living_room", Name: "Living Room"},
	}

	result, err := f.FormatAllRegistries(context.Background(), entities, devices, areas, RegistryOptions{})
	if err != nil {
		t.Fatalf("FormatAllRegistries() error = %v", err)
	}

	// Should contain all sections
	if !strings.Contains(result, "Entities") {
		t.Errorf("FormatAllRegistries() should contain Entities section, got %q", result)
	}
	if !strings.Contains(result, "Devices") {
		t.Errorf("FormatAllRegistries() should contain Devices section, got %q", result)
	}
	if !strings.Contains(result, "Areas") {
		t.Errorf("FormatAllRegistries() should contain Areas section, got %q", result)
	}
}

func TestJSONRegistryFormatter_FormatEntityRegistry(t *testing.T) {
	f := NewJSONRegistryFormatter()

	entries := []homeassistant.EntityRegistryEntry{
		{EntityID: "light.living_room", Platform: "hue", Name: "Living Room Light"},
	}

	result, err := f.FormatEntityRegistry(context.Background(), entries, RegistryOptions{})
	if err != nil {
		t.Fatalf("FormatEntityRegistry() error = %v", err)
	}

	// Should be valid JSON with entity_id
	if !strings.Contains(result, `"entity_id"`) {
		t.Errorf("FormatEntityRegistry() should contain entity_id JSON field, got %q", result)
	}
	if !strings.Contains(result, `"light.living_room"`) {
		t.Errorf("FormatEntityRegistry() should contain entity_id value, got %q", result)
	}
}

func TestJSONRegistryFormatter_FormatDeviceRegistry(t *testing.T) {
	f := NewJSONRegistryFormatter()

	entries := []homeassistant.DeviceRegistryEntry{
		{ID: "dev1", Name: "Hue Bridge", Manufacturer: "Philips"},
	}

	result, err := f.FormatDeviceRegistry(context.Background(), entries, RegistryOptions{})
	if err != nil {
		t.Fatalf("FormatDeviceRegistry() error = %v", err)
	}

	// Should be valid JSON
	if !strings.Contains(result, `"id"`) {
		t.Errorf("FormatDeviceRegistry() should contain id JSON field, got %q", result)
	}
	if !strings.Contains(result, `"dev1"`) {
		t.Errorf("FormatDeviceRegistry() should contain device id value, got %q", result)
	}
}

func TestJSONRegistryFormatter_FormatAreaRegistry(t *testing.T) {
	f := NewJSONRegistryFormatter()

	entries := []homeassistant.AreaRegistryEntry{
		{AreaID: "living_room", Name: "Living Room"},
	}

	result, err := f.FormatAreaRegistry(context.Background(), entries)
	if err != nil {
		t.Fatalf("FormatAreaRegistry() error = %v", err)
	}

	// Should be valid JSON
	if !strings.Contains(result, `"area_id"`) {
		t.Errorf("FormatAreaRegistry() should contain area_id JSON field, got %q", result)
	}
	if !strings.Contains(result, `"living_room"`) {
		t.Errorf("FormatAreaRegistry() should contain area_id value, got %q", result)
	}
}

func TestJSONRegistryFormatter_FormatAreaRegistry_Nil(t *testing.T) {
	f := NewJSONRegistryFormatter()

	result, err := f.FormatAreaRegistry(context.Background(), nil)
	if err != nil {
		t.Fatalf("FormatAreaRegistry() error = %v", err)
	}
	// nil should produce empty JSON array
	if !strings.Contains(result, "[]") {
		t.Errorf("FormatAreaRegistry(nil) = %q, want empty JSON array", result)
	}
}

func TestJSONRegistryFormatter_FormatDeviceRegistry_WithEntityMap(t *testing.T) {
	f := NewJSONRegistryFormatter()

	entries := []homeassistant.DeviceRegistryEntry{
		{ID: "dev1", Name: "Hue Bridge", Manufacturer: "Philips"},
	}

	opts := RegistryOptions{
		EntityMap: map[string][]EntityInfo{
			"dev1": {
				{EntityID: "light.hue_1"},
			},
		},
	}

	result, err := f.FormatDeviceRegistry(context.Background(), entries, opts)
	if err != nil {
		t.Fatalf("FormatDeviceRegistry() error = %v", err)
	}

	// Should include entities key
	if !strings.Contains(result, `"entities"`) {
		t.Errorf("FormatDeviceRegistry() should contain entities key, got %q", result)
	}
	if !strings.Contains(result, "light.hue_1") {
		t.Errorf("FormatDeviceRegistry() should contain entity_id, got %q", result)
	}
}

func TestJSONRegistryFormatter_FormatDeviceRegistry_Nil(t *testing.T) {
	f := NewJSONRegistryFormatter()

	result, err := f.FormatDeviceRegistry(context.Background(), nil, RegistryOptions{})
	if err != nil {
		t.Fatalf("FormatDeviceRegistry() error = %v", err)
	}
	if !strings.Contains(result, "[]") {
		t.Errorf("FormatDeviceRegistry(nil) = %q, want empty JSON array", result)
	}
}

func TestJSONRegistryFormatter_FormatAllRegistries(t *testing.T) {
	f := NewJSONRegistryFormatter()

	entities := []homeassistant.EntityRegistryEntry{
		{EntityID: "light.living_room", Platform: "hue"},
	}
	devices := []homeassistant.DeviceRegistryEntry{
		{ID: "dev1", Name: "Hue Bridge"},
	}
	areas := []homeassistant.AreaRegistryEntry{
		{AreaID: "living_room", Name: "Living Room"},
	}

	result, err := f.FormatAllRegistries(context.Background(), entities, devices, areas, RegistryOptions{})
	if err != nil {
		t.Fatalf("FormatAllRegistries() error = %v", err)
	}

	// Should be JSON with all three keys
	if !strings.Contains(result, `"entities"`) {
		t.Errorf("FormatAllRegistries() should contain entities key, got %q", result)
	}
	if !strings.Contains(result, `"devices"`) {
		t.Errorf("FormatAllRegistries() should contain devices key, got %q", result)
	}
	if !strings.Contains(result, `"areas"`) {
		t.Errorf("FormatAllRegistries() should contain areas key, got %q", result)
	}
}

func TestJSONRegistryFormatter_FormatAllRegistries_Nil(t *testing.T) {
	f := NewJSONRegistryFormatter()

	result, err := f.FormatAllRegistries(context.Background(), nil, nil, nil, RegistryOptions{})
	if err != nil {
		t.Fatalf("FormatAllRegistries(nil) error = %v", err)
	}
	if !strings.Contains(result, `"entities"`) {
		t.Errorf("FormatAllRegistries(nil) should contain entities key, got %q", result)
	}
}

func TestNaturalRegistryFormatter_FormatDeviceRegistry_WithEntityMap(t *testing.T) {
	f := NewNaturalRegistryFormatter()

	entries := []homeassistant.DeviceRegistryEntry{
		{ID: "dev1", Name: "Hue Bridge", Manufacturer: "Philips", Model: "BSB002"},
	}

	opts := RegistryOptions{
		Verbose: true,
		EntityMap: map[string][]EntityInfo{
			"dev1": {
				{EntityID: "light.hue_1"},
				{EntityID: "light.hue_2"},
			},
		},
	}

	result, err := f.FormatDeviceRegistry(context.Background(), entries, opts)
	if err != nil {
		t.Fatalf("FormatDeviceRegistry() error = %v", err)
	}

	if !strings.Contains(result, "Hue Bridge") {
		t.Errorf("FormatDeviceRegistry() should contain device name, got %q", result)
	}
	if !strings.Contains(result, "light.hue_1") {
		t.Errorf("FormatDeviceRegistry() should contain entity, got %q", result)
	}
}
