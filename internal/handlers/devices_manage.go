package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// Device action constants.
const (
	deviceActionGet    = "get"
	deviceActionUpdate = "update"
)

// DeviceManageHandlers provides handlers for device registry management.
type DeviceManageHandlers struct{}

// NewDeviceManageHandlers creates a new DeviceManageHandlers instance.
func NewDeviceManageHandlers() *DeviceManageHandlers {
	return &DeviceManageHandlers{}
}

// RegisterTools registers the manage_device tool with the registry.
func (h *DeviceManageHandlers) RegisterTools(registry *mcp.Registry) {
	registry.RegisterTool(h.manageDeviceTool(), h.handleManageDevice)
}

// =============================================================================
// Tool Definition
// =============================================================================

func (h *DeviceManageHandlers) manageDeviceTool() mcp.Tool {
	return mcp.Tool{
		Name: "manage_device",
		Description: `Manage Home Assistant device registry entries - get or update.

Actions:
- get: Get device registry details (requires device_id)
- update: Update device registry entry (requires device_id and at least one field to update)

Safe fields that can be updated:
- name_by_user: Custom display name (empty string removes override)
- area_id: Area assignment (empty string removes)
- disabled_by: 'user' to disable, 'none' to enable
- labels: Array of label strings (empty array clears)`,
		InputSchema: mcp.JSONSchema{
			Type:        "object",
			Description: "Device registry management operation",
			Properties: map[string]mcp.JSONSchema{
				"action": {
					Type:        "string",
					Description: "Operation to perform: get, update",
					Enum:        []string{"get", "update"},
				},
				"device_id": {
					Type:        "string",
					Description: "Device ID. Required for get/update.",
				},
				"name_by_user": {
					Type:        "string",
					Description: "Custom display name (update only, empty string removes override)",
				},
				"area_id": {
					Type:        "string",
					Description: "Area ID (update only, empty string removes assignment)",
				},
				"disabled_by": {
					Type:        "string",
					Description: "Disable status (update only): 'user' to disable, 'none' to enable",
					Enum:        []string{"user", "none"},
				},
				"labels": {
					Type:        "array",
					Description: "Labels array (update only, empty array clears)",
					Items:       &mcp.JSONSchema{Type: "string"},
				},
				"format": {
					Type:        "string",
					Description: "Output format: 'natural' (human-readable, default) or 'json' (structured)",
					Enum:        []string{"natural", "json"},
				},
			},
			Required: []string{"action"},
		},
	}
}

// =============================================================================
// Handler Implementation
// =============================================================================

func (h *DeviceManageHandlers) handleManageDevice(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	action, ok := args["action"].(string)
	if !ok || action == "" {
		return errorResult("action is required and must be a string (get or update)"), nil
	}

	format := "natural"
	if f, ok := args["format"].(string); ok && f != "" {
		format = f
	}

	switch action {
	case deviceActionGet:
		return h.handleGetDevice(ctx, client, args, format)
	case deviceActionUpdate:
		return h.handleUpdateDevice(ctx, client, args, format)
	default:
		return errorResult(fmt.Sprintf("unsupported action '%s'. Valid actions: get, update", action)), nil
	}
}

func (h *DeviceManageHandlers) handleGetDevice(ctx context.Context, client homeassistant.Client, args map[string]any, format string) (*mcp.ToolsCallResult, error) {
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return errorResult("device_id is required for get action"), nil
	}

	// Get device registry
	registry, err := client.GetDeviceRegistry(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get device registry: %w", err)
	}

	// Find device
	for _, entry := range registry {
		if entry.ID == deviceID {
			if format == formatJSON {
				return h.formatDeviceJSON(&entry)
			}
			return h.formatDeviceNatural(&entry), nil
		}
	}

	return errorResult(fmt.Sprintf("device '%s' not found in registry", deviceID)), nil
}

func (h *DeviceManageHandlers) handleUpdateDevice(ctx context.Context, client homeassistant.Client, args map[string]any, format string) (*mcp.ToolsCallResult, error) {
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return errorResult("device_id is required for update action"), nil
	}

	// Build update config
	config, hasFields := h.buildDeviceUpdateConfig(args)
	if !hasFields {
		return errorResult("at least one field must be provided for update (name_by_user, area_id, disabled_by, labels)"), nil
	}

	// Update device
	updated, err := client.UpdateDeviceRegistryEntry(ctx, deviceID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	if format == formatJSON {
		return h.formatDeviceJSON(updated)
	}
	return h.formatDeviceNaturalWithSuccess(updated), nil
}

// =============================================================================
// Helper Functions
// =============================================================================

func (h *DeviceManageHandlers) buildDeviceUpdateConfig(args map[string]any) (homeassistant.DeviceRegistryUpdateConfig, bool) {
	config := homeassistant.DeviceRegistryUpdateConfig{}
	hasFields := false

	if nameByUser, ok := args["name_by_user"].(string); ok {
		config.NameByUser = &nameByUser
		hasFields = true
	}

	if areaID, ok := args["area_id"].(string); ok {
		config.AreaID = &areaID
		hasFields = true
	}

	if disabledBy, ok := args["disabled_by"].(string); ok {
		// Map "none" to empty string for HA API
		if disabledBy == "none" {
			disabledBy = ""
		}
		config.DisabledBy = &disabledBy
		hasFields = true
	}

	if labels, ok := args["labels"].([]any); ok {
		config.Labels = toStringArray(labels)
		hasFields = true
	}

	return config, hasFields
}

// =============================================================================
// Formatters
// =============================================================================

func (h *DeviceManageHandlers) formatDeviceNatural(entry *homeassistant.DeviceRegistryEntry) *mcp.ToolsCallResult {
	var parts []string
	parts = append(parts, fmt.Sprintf("Device ID: %s", entry.ID))

	if entry.Name != "" {
		parts = append(parts, fmt.Sprintf("Name: %s", entry.Name))
	}

	if entry.NameByUser != "" {
		parts = append(parts, fmt.Sprintf("Name by user: %s", entry.NameByUser))
	}

	if entry.Manufacturer != "" {
		parts = append(parts, fmt.Sprintf("Manufacturer: %s", entry.Manufacturer))
	}

	if string(entry.Model) != "" {
		parts = append(parts, fmt.Sprintf("Model: %s", string(entry.Model)))
	}

	if entry.AreaID != "" {
		parts = append(parts, fmt.Sprintf("Area ID: %s", entry.AreaID))
	}

	if entry.DisabledBy != "" {
		parts = append(parts, fmt.Sprintf("Disabled by: %s", entry.DisabledBy))
	}

	if len(entry.Labels) > 0 {
		parts = append(parts, fmt.Sprintf("Labels: %s", strings.Join(entry.Labels, ", ")))
	}

	if string(entry.SWVersion) != "" {
		parts = append(parts, fmt.Sprintf("SW Version: %s", string(entry.SWVersion)))
	}

	if string(entry.HWVersion) != "" {
		parts = append(parts, fmt.Sprintf("HW Version: %s", string(entry.HWVersion)))
	}

	if len(entry.ConfigEntries) > 0 {
		parts = append(parts, fmt.Sprintf("Config Entries: %d", len(entry.ConfigEntries)))
	}

	return textResult(strings.Join(parts, "\n"))
}

func (h *DeviceManageHandlers) formatDeviceNaturalWithSuccess(entry *homeassistant.DeviceRegistryEntry) *mcp.ToolsCallResult {
	details := h.formatDeviceNatural(entry).Content[0].Text
	return textResult(fmt.Sprintf("Device '%s' updated successfully.\n\n%s", entry.ID, details))
}

func (h *DeviceManageHandlers) formatDeviceJSON(entry *homeassistant.DeviceRegistryEntry) (*mcp.ToolsCallResult, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal device: %w", err)
	}
	return textResult(string(data)), nil
}
