// Package handlers provides MCP tool handlers for Home Assistant operations.
package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// Area action constants.
const (
	areaActionList   = "list"
	areaActionGet    = "get"
	areaActionCreate = "create"
	areaActionUpdate = "update"
	areaActionDelete = "delete"

	formatJSON = "json"
)

// AreaHandlers provides handlers for area-related MCP tools.
type AreaHandlers struct{}

// NewAreaHandlers creates a new AreaHandlers instance.
func NewAreaHandlers() *AreaHandlers {
	return &AreaHandlers{}
}

// RegisterTools registers the consolidated manage_area tool with the registry.
func (h *AreaHandlers) RegisterTools(registry *mcp.Registry) {
	registry.RegisterTool(h.manageAreaTool(), h.handleManageArea)
}

// =============================================================================
// Tool Definition
// =============================================================================

func (h *AreaHandlers) manageAreaTool() mcp.Tool {
	schema := h.buildAreaSchema()
	return mcp.Tool{
		Name: "manage_area",
		Description: `Manage Home Assistant areas - list, get, create, update, or delete.

Actions:
- list: List all areas (optional filters: name_contains)
- get: Get details of a specific area with device/entity counts (requires area_id)
- create: Create a new area (requires name)
- update: Update an existing area (requires area_id)
- delete: Delete an area (requires area_id)`,
		InputSchema: schema,
	}
}

func (h *AreaHandlers) buildAreaSchema() mcp.JSONSchema {
	return mcp.JSONSchema{
		Type:        "object",
		Description: "Area management operation",
		Properties: map[string]mcp.JSONSchema{
			"action": {
				Type:        "string",
				Description: "Operation to perform: list, get, create, update, delete",
				Enum:        []string{"list", "get", "create", "update", "delete"},
			},
			"area_id": {
				Type:        "string",
				Description: "Area identifier (e.g., 'living_room'). Required for get/update/delete.",
			},
			"name": {
				Type:        "string",
				Description: "Area name (required for create, optional for update)",
			},
			"icon":    {Type: "string", Description: "Area icon (e.g., 'mdi:sofa')"},
			"picture": {Type: "string", Description: "Area picture URL"},
			"floor_id": {
				Type:        "string",
				Description: "Floor identifier this area belongs to",
			},
			"aliases": {
				Type:        "array",
				Description: "Alternative names for the area",
				Items: &mcp.JSONSchema{
					Type: "string",
				},
			},
			"labels": {
				Type:        "array",
				Description: "Labels for categorizing the area",
				Items: &mcp.JSONSchema{
					Type: "string",
				},
			},
			"name_contains": {
				Type:        "string",
				Description: "Filter by area name containing this string (for list action, case-insensitive)",
			},
			"format": {
				Type:        "string",
				Enum:        []string{"natural", "json"},
				Description: "Output format: 'natural' (default) for LLM-optimized text, 'json' for structured data",
			},
		},
		Required: []string{"action"},
	}
}

// =============================================================================
// Main Handler
// =============================================================================

func (h *AreaHandlers) handleManageArea(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	action, _ := args["action"].(string)
	if action == "" {
		return errorResult("action is required"), nil
	}

	switch action {
	case areaActionList:
		return h.handleList(ctx, client, args)
	case areaActionGet:
		return h.handleGet(ctx, client, args)
	case areaActionCreate:
		return h.handleCreate(ctx, client, args)
	case areaActionUpdate:
		return h.handleUpdate(ctx, client, args)
	case areaActionDelete:
		return h.handleDelete(ctx, client, args)
	default:
		return errorResult(fmt.Sprintf("invalid action: %s (must be list, get, create, update, or delete)", action)), nil
	}
}

// =============================================================================
// Action Handlers
// =============================================================================

func (h *AreaHandlers) handleList(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	areas, err := client.GetAreaRegistry(ctx)
	if err != nil {
		return errorResult(fmt.Sprintf("error listing areas: %v", err)), nil
	}

	// Apply name filter if provided
	nameContains, _ := args["name_contains"].(string)
	if nameContains != "" {
		filtered := make([]homeassistant.AreaRegistryEntry, 0)
		nameLower := strings.ToLower(nameContains)
		for _, area := range areas {
			if strings.Contains(strings.ToLower(area.Name), nameLower) ||
				strings.Contains(strings.ToLower(area.AreaID), nameLower) {
				filtered = append(filtered, area)
			}
		}
		areas = filtered
	}

	formatStr, _ := args["format"].(string)
	if formatStr == formatJSON {
		return h.formatListJSON(areas)
	}
	return h.formatListNatural(areas)
}

func (h *AreaHandlers) handleGet(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	areaID, ok := args["area_id"].(string)
	if !ok || areaID == "" {
		return errorResult("area_id is required for get action"), nil
	}

	// Get area from registry
	areas, err := client.GetAreaRegistry(ctx)
	if err != nil {
		return errorResult(fmt.Sprintf("error getting areas: %v", err)), nil
	}

	var found *homeassistant.AreaRegistryEntry
	for i := range areas {
		if areas[i].AreaID == areaID {
			found = &areas[i]
			break
		}
	}

	if found == nil {
		return errorResult(fmt.Sprintf("area not found: %s", areaID)), nil
	}

	// Enrich with device and entity counts
	deviceCount, entityCount := h.getAreaCounts(ctx, client, areaID)

	formatStr, _ := args["format"].(string)
	if formatStr == formatJSON {
		return h.formatDetailJSON(*found, deviceCount, entityCount)
	}
	return h.formatDetailNatural(*found, deviceCount, entityCount)
}

func (h *AreaHandlers) handleCreate(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return errorResult("name is required for create action"), nil
	}

	config := h.buildAreaConfig(args)
	config.Name = name

	entry, err := client.CreateArea(ctx, config)
	if err != nil {
		return errorResult(fmt.Sprintf("error creating area: %v", err)), nil
	}

	formatStr, _ := args["format"].(string)
	if formatStr == formatJSON {
		return h.formatDetailJSON(*entry, 0, 0)
	}
	return h.formatCreateNatural(*entry)
}

func (h *AreaHandlers) handleUpdate(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	areaID, ok := args["area_id"].(string)
	if !ok || areaID == "" {
		return errorResult("area_id is required for update action"), nil
	}

	config := h.buildAreaConfig(args)

	entry, err := client.UpdateArea(ctx, areaID, config)
	if err != nil {
		return errorResult(fmt.Sprintf("error updating area: %v", err)), nil
	}

	formatStr, _ := args["format"].(string)
	if formatStr == formatJSON {
		return h.formatDetailJSON(*entry, 0, 0)
	}
	return h.formatUpdateNatural(*entry)
}

func (h *AreaHandlers) handleDelete(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	areaID, ok := args["area_id"].(string)
	if !ok || areaID == "" {
		return errorResult("area_id is required for delete action"), nil
	}

	err := client.DeleteArea(ctx, areaID)
	if err != nil {
		return errorResult(fmt.Sprintf("error deleting area: %v", err)), nil
	}

	return successResult(fmt.Sprintf("Area '%s' deleted successfully", areaID)), nil
}

// =============================================================================
// Helper Functions
// =============================================================================

func (h *AreaHandlers) buildAreaConfig(args map[string]any) homeassistant.AreaConfig {
	cfg := homeassistant.AreaConfig{}

	if name, ok := args["name"].(string); ok && name != "" {
		cfg.Name = name
	}
	if icon, ok := args["icon"].(string); ok && icon != "" {
		cfg.Icon = icon
	}
	if picture, ok := args["picture"].(string); ok && picture != "" {
		cfg.Picture = picture
	}
	if floorID, ok := args["floor_id"].(string); ok && floorID != "" {
		cfg.FloorID = floorID
	}

	// Handle array fields
	cfg.Aliases = toStringArray(args["aliases"])
	cfg.Labels = toStringArray(args["labels"])

	return cfg
}

// toStringArray converts []any to []string, filtering out non-string values.
func toStringArray(value any) []string {
	arr, ok := value.([]any)
	if !ok {
		return nil
	}

	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func (h *AreaHandlers) getAreaCounts(ctx context.Context, client homeassistant.Client, areaID string) (int, int) {
	deviceCount := 0
	if devices, err := client.GetDeviceRegistry(ctx); err == nil {
		for _, d := range devices {
			if d.AreaID == areaID {
				deviceCount++
			}
		}
	}

	entityCount := 0
	if entities, err := client.GetEntityRegistry(ctx); err == nil {
		for _, e := range entities {
			if e.AreaID == areaID {
				entityCount++
			}
		}
	}

	return deviceCount, entityCount
}

// =============================================================================
// Formatting Functions (private, domain-specific)
// =============================================================================

func (h *AreaHandlers) formatListJSON(areas []homeassistant.AreaRegistryEntry) (*mcp.ToolsCallResult, error) {
	return jsonResult(map[string]any{
		"areas": areas,
		"count": len(areas),
	})
}

func (h *AreaHandlers) formatListNatural(areas []homeassistant.AreaRegistryEntry) (*mcp.ToolsCallResult, error) {
	if len(areas) == 0 {
		return successResult("No areas found"), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d area(s):\n\n", len(areas)))

	for _, area := range areas {
		output.WriteString(fmt.Sprintf("• %s (ID: %s)\n", area.Name, area.AreaID))
		if area.FloorID != "" {
			output.WriteString(fmt.Sprintf("  Floor: %s\n", area.FloorID))
		}
		if area.Icon != "" {
			output.WriteString(fmt.Sprintf("  Icon: %s\n", area.Icon))
		}
		if len(area.Aliases) > 0 {
			output.WriteString(fmt.Sprintf("  Aliases: %s\n", strings.Join(area.Aliases, ", ")))
		}
		if len(area.Labels) > 0 {
			output.WriteString(fmt.Sprintf("  Labels: %s\n", strings.Join(area.Labels, ", ")))
		}
	}

	return successResult(output.String()), nil
}

func (h *AreaHandlers) formatDetailJSON(area homeassistant.AreaRegistryEntry, deviceCount, entityCount int) (*mcp.ToolsCallResult, error) {
	result := map[string]any{
		"area_id": area.AreaID,
		"name":    area.Name,
	}

	if area.Icon != "" {
		result["icon"] = area.Icon
	}
	if area.Picture != "" {
		result["picture"] = area.Picture
	}
	if area.FloorID != "" {
		result["floor_id"] = area.FloorID
	}
	if len(area.Aliases) > 0 {
		result["aliases"] = area.Aliases
	}
	if len(area.Labels) > 0 {
		result["labels"] = area.Labels
	}
	if deviceCount > 0 || entityCount > 0 {
		result["device_count"] = deviceCount
		result["entity_count"] = entityCount
	}

	return jsonResult(result)
}

func (h *AreaHandlers) formatDetailNatural(area homeassistant.AreaRegistryEntry, deviceCount, entityCount int) (*mcp.ToolsCallResult, error) {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Area: %s\n", area.Name))
	output.WriteString(fmt.Sprintf("ID: %s\n", area.AreaID))

	if area.FloorID != "" {
		output.WriteString(fmt.Sprintf("Floor: %s\n", area.FloorID))
	}
	if area.Icon != "" {
		output.WriteString(fmt.Sprintf("Icon: %s\n", area.Icon))
	}
	if area.Picture != "" {
		output.WriteString(fmt.Sprintf("Picture: %s\n", area.Picture))
	}
	if len(area.Aliases) > 0 {
		output.WriteString(fmt.Sprintf("Aliases: %s\n", strings.Join(area.Aliases, ", ")))
	}
	if len(area.Labels) > 0 {
		output.WriteString(fmt.Sprintf("Labels: %s\n", strings.Join(area.Labels, ", ")))
	}
	if deviceCount > 0 || entityCount > 0 {
		output.WriteString(fmt.Sprintf("\nDevices: %d\n", deviceCount))
		output.WriteString(fmt.Sprintf("Entities: %d\n", entityCount))
	}

	return successResult(output.String()), nil
}

func (h *AreaHandlers) formatCreateNatural(area homeassistant.AreaRegistryEntry) (*mcp.ToolsCallResult, error) {
	return successResult(fmt.Sprintf("Area '%s' created successfully (ID: %s)", area.Name, area.AreaID)), nil
}

func (h *AreaHandlers) formatUpdateNatural(area homeassistant.AreaRegistryEntry) (*mcp.ToolsCallResult, error) {
	return successResult(fmt.Sprintf("Area '%s' updated successfully", area.Name)), nil
}
