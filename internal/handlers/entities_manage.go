package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// Entity action constants.
const (
	entityActionGet    = "get"
	entityActionUpdate = "update"

	formatNatural = "natural"
	noneValue     = "none"
)

// EntityManageHandlers provides handlers for entity registry management.
type EntityManageHandlers struct{}

// NewEntityManageHandlers creates a new EntityManageHandlers instance.
func NewEntityManageHandlers() *EntityManageHandlers {
	return &EntityManageHandlers{}
}

// RegisterTools registers the manage_entity tool with the registry.
func (h *EntityManageHandlers) RegisterTools(registry *mcp.Registry) {
	registry.RegisterTool(h.manageEntityTool(), h.handleManageEntity)
}

// =============================================================================
// Tool Definition
// =============================================================================

func (h *EntityManageHandlers) manageEntityTool() mcp.Tool {
	return mcp.Tool{
		Name: "manage_entity",
		Description: `Manage Home Assistant entity registry entries - get or update.

Actions:
- get: Get entity registry details (requires entity_id)
- update: Update entity registry entry (requires entity_id and at least one field to update)

Safe fields that can be updated:
- name: Custom display name (empty string removes override)
- icon: Custom icon like 'mdi:lightbulb' (empty string removes override)
- area_id: Area assignment (empty string removes)
- disabled_by: 'user' to disable, 'none' to enable
- hidden_by: 'user' to hide, 'none' to show
- labels: Array of label strings (empty array clears)
- aliases: Array of alternative name strings (empty array clears)`,
		InputSchema: h.buildEntityManageSchema(),
	}
}

func (h *EntityManageHandlers) buildEntityManageSchema() mcp.JSONSchema {
	return mcp.JSONSchema{
		Type:        "object",
		Description: "Entity registry management operation",
		Properties: map[string]mcp.JSONSchema{
			"action": {
				Type:        "string",
				Description: "Operation to perform: get, update",
				Enum:        []string{"get", "update"},
			},
			"entity_id": {
				Type:        "string",
				Description: "Entity ID (e.g., 'light.living_room'). Required for get/update.",
			},
			"name": {
				Type:        "string",
				Description: "Custom display name (update only, empty string removes override)",
			},
			"icon": {
				Type:        "string",
				Description: "Custom icon (update only, e.g., 'mdi:lightbulb', empty string removes override)",
			},
			"area_id": {
				Type:        "string",
				Description: "Area ID (update only, empty string removes assignment)",
			},
			"disabled_by": {
				Type:        "string",
				Description: "Disable status (update only): 'user' to disable, 'none' to enable",
			},
			"hidden_by": {
				Type:        "string",
				Description: "Hidden status (update only): 'user' to hide, 'none' to show",
			},
			"labels": {
				Type:        "array",
				Description: "Labels array (update only, empty array clears)",
				Items:       &mcp.JSONSchema{Type: "string"},
			},
			"aliases": {
				Type:        "array",
				Description: "Aliases array (update only, empty array clears)",
				Items:       &mcp.JSONSchema{Type: "string"},
			},
			"format": {
				Type:        "string",
				Description: "Output format: 'natural' (human-readable, default) or 'json' (structured)",
				Enum:        []string{"natural", "json"},
			},
		},
	}
}

// =============================================================================
// Handler Implementation
// =============================================================================

func (h *EntityManageHandlers) handleManageEntity(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	action, ok := args["action"].(string)
	if !ok || action == "" {
		return errorResult("action is required and must be a string (get or update)"), nil
	}

	format := formatNatural
	if f, ok := args["format"].(string); ok && f != "" {
		format = f
	}

	switch action {
	case entityActionGet:
		return h.handleGetEntity(ctx, client, args, format)
	case entityActionUpdate:
		return h.handleUpdateEntity(ctx, client, args, format)
	default:
		return errorResult(fmt.Sprintf("unsupported action '%s'. Valid actions: get, update", action)), nil
	}
}

func (h *EntityManageHandlers) handleGetEntity(ctx context.Context, client homeassistant.Client, args map[string]any, format string) (*mcp.ToolsCallResult, error) {
	entityID, ok := args["entity_id"].(string)
	if !ok || entityID == "" {
		return errorResult("entity_id is required for get action"), nil
	}

	// Get entity registry
	registry, err := client.GetEntityRegistry(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity registry: %w", err)
	}

	// Find entity
	for _, entry := range registry {
		if entry.EntityID == entityID {
			if format == formatJSON {
				return h.formatEntityJSON(&entry)
			}
			return h.formatEntityNatural(&entry), nil
		}
	}

	return errorResult(fmt.Sprintf("entity '%s' not found in registry", entityID)), nil
}

func (h *EntityManageHandlers) handleUpdateEntity(ctx context.Context, client homeassistant.Client, args map[string]any, format string) (*mcp.ToolsCallResult, error) {
	entityID, ok := args["entity_id"].(string)
	if !ok || entityID == "" {
		return errorResult("entity_id is required for update action"), nil
	}

	// Build update config
	config, hasFields := h.buildEntityUpdateConfig(args)
	if !hasFields {
		return errorResult("at least one field must be provided for update (name, icon, area_id, disabled_by, hidden_by, labels, aliases)"), nil
	}

	// Update entity
	updated, err := client.UpdateEntityRegistryEntry(ctx, entityID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	if format == formatJSON {
		return h.formatEntityJSON(updated)
	}
	return h.formatEntityNaturalWithSuccess(updated), nil
}

// =============================================================================
// Helper Functions
// =============================================================================

func (h *EntityManageHandlers) buildEntityUpdateConfig(args map[string]any) (homeassistant.EntityRegistryUpdateConfig, bool) {
	config := homeassistant.EntityRegistryUpdateConfig{}
	hasFields := false

	if name, ok := args["name"].(string); ok {
		config.Name = &name
		hasFields = true
	}

	if icon, ok := args["icon"].(string); ok {
		config.Icon = &icon
		hasFields = true
	}

	if areaID, ok := args["area_id"].(string); ok {
		config.AreaID = &areaID
		hasFields = true
	}

	if disabledBy, ok := args["disabled_by"].(string); ok {
		// Map "none" to empty string for HA API
		if disabledBy == noneValue {
			disabledBy = ""
		}
		config.DisabledBy = &disabledBy
		hasFields = true
	}

	if hiddenBy, ok := args["hidden_by"].(string); ok {
		// Map "none" to empty string for HA API
		if hiddenBy == noneValue {
			hiddenBy = ""
		}
		config.HiddenBy = &hiddenBy
		hasFields = true
	}

	if labels, ok := args["labels"].([]any); ok {
		config.Labels = toStringArray(labels)
		hasFields = true
	}

	if aliases, ok := args["aliases"].([]any); ok {
		config.Aliases = toStringArray(aliases)
		hasFields = true
	}

	return config, hasFields
}

// =============================================================================
// Formatters
// =============================================================================

func (h *EntityManageHandlers) formatEntityNatural(entry *homeassistant.EntityRegistryEntry) *mcp.ToolsCallResult {
	var parts []string
	parts = append(parts, fmt.Sprintf("Entity ID: %s", entry.EntityID))

	if entry.Name != "" {
		parts = append(parts, fmt.Sprintf("Name: %s", entry.Name))
	}

	if entry.Platform != "" {
		parts = append(parts, fmt.Sprintf("Platform: %s", entry.Platform))
	}

	if entry.AreaID != "" {
		parts = append(parts, fmt.Sprintf("Area ID: %s", entry.AreaID))
	}

	if entry.Icon != "" {
		parts = append(parts, fmt.Sprintf("Icon: %s", entry.Icon))
	}

	if entry.DisabledBy != "" {
		parts = append(parts, fmt.Sprintf("Disabled by: %s", entry.DisabledBy))
	}

	if entry.HiddenBy != "" {
		parts = append(parts, fmt.Sprintf("Hidden by: %s", entry.HiddenBy))
	}

	if len(entry.Labels) > 0 {
		parts = append(parts, fmt.Sprintf("Labels: %s", strings.Join(entry.Labels, ", ")))
	}

	if len(entry.Aliases) > 0 {
		parts = append(parts, fmt.Sprintf("Aliases: %s", strings.Join(entry.Aliases, ", ")))
	}

	if entry.DeviceID != "" {
		parts = append(parts, fmt.Sprintf("Device ID: %s", entry.DeviceID))
	}

	if entry.ConfigEntryID != "" {
		parts = append(parts, fmt.Sprintf("Config Entry ID: %s", entry.ConfigEntryID))
	}

	return textResult(strings.Join(parts, "\n"))
}

func (h *EntityManageHandlers) formatEntityNaturalWithSuccess(entry *homeassistant.EntityRegistryEntry) *mcp.ToolsCallResult {
	details := h.formatEntityNatural(entry).Content[0].Text
	return textResult(fmt.Sprintf("Entity '%s' updated successfully.\n\n%s", entry.EntityID, details))
}

func (h *EntityManageHandlers) formatEntityJSON(entry *homeassistant.EntityRegistryEntry) (*mcp.ToolsCallResult, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}
	return textResult(string(data)), nil
}
