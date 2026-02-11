// Package handlers provides MCP tool handlers for Home Assistant operations.
package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/zorak1103/ha-mcp/internal/handlers/formatter"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// Label action constants.
const (
	labelActionList   = "list"
	labelActionGet    = "get"
	labelActionCreate = "create"
	labelActionUpdate = "update"
	labelActionDelete = "delete"
)

// LabelHandlers provides handlers for label-related MCP tools.
type LabelHandlers struct{}

// NewLabelHandlers creates a new LabelHandlers instance.
func NewLabelHandlers() *LabelHandlers {
	return &LabelHandlers{}
}

// RegisterTools registers the consolidated manage_label tool with the registry.
func (h *LabelHandlers) RegisterTools(registry *mcp.Registry) {
	registry.RegisterTool(h.manageLabelTool(), h.handleManageLabel)
}

// =============================================================================
// Tool Definition
// =============================================================================

func (h *LabelHandlers) manageLabelTool() mcp.Tool {
	schema := h.buildLabelSchema()
	return mcp.Tool{
		Name: "manage_label",
		Description: `Manage Home Assistant labels - list, get, create, update, or delete.

Labels are used to categorize and organize entities, devices, and areas.

Actions:
- list: List all labels (optional filters: name_contains)
- get: Get details of a specific label (requires label_id)
- create: Create a new label (requires name)
- update: Update an existing label (requires label_id)
- delete: Delete a label (requires label_id)`,
		InputSchema: schema,
	}
}

func (h *LabelHandlers) buildLabelSchema() mcp.JSONSchema {
	return mcp.JSONSchema{
		Type:        "object",
		Description: "Label management operation",
		Properties: map[string]mcp.JSONSchema{
			"action": {
				Type:        "string",
				Description: "Operation to perform: list, get, create, update, delete",
				Enum:        []string{"list", "get", "create", "update", "delete"},
			},
			"label_id": {
				Type:        "string",
				Description: "Label identifier. Required for get/update/delete.",
			},
			"name": {
				Type:        "string",
				Description: "Label name (required for create, optional for update)",
			},
			"color": {
				Type:        "string",
				Description: "Label color (hex format or named color)",
			},
			"icon": {
				Type:        "string",
				Description: "Label icon (e.g., 'mdi:label')",
			},
			"description": {
				Type:        "string",
				Description: "Label description",
			},
			"name_contains": {
				Type:        "string",
				Description: "Filter by label name containing this string (for list action, case-insensitive)",
			},
			"format": {
				Type:        "string",
				Description: "Output format: 'natural' (default) for LLM-optimized text, 'json' for structured data",
				Enum:        []string{"natural", "json"},
			},
		},
		Required: []string{"action"},
	}
}

// =============================================================================
// Handler: manage_label
// =============================================================================

func (h *LabelHandlers) handleManageLabel(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	action, _ := args["action"].(string)
	if action == "" {
		return errorResult("action is required"), nil
	}

	switch action {
	case labelActionList:
		return h.handleList(ctx, client, args)
	case labelActionGet:
		return h.handleGet(ctx, client, args)
	case labelActionCreate:
		return h.handleCreate(ctx, client, args)
	case labelActionUpdate:
		return h.handleUpdate(ctx, client, args)
	case labelActionDelete:
		return h.handleDelete(ctx, client, args)
	default:
		return errorResult(fmt.Sprintf("invalid action: %s (must be list, get, create, update, or delete)", action)), nil
	}
}

func (h *LabelHandlers) handleList(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	labels, err := client.GetLabelRegistry(ctx)
	if err != nil {
		return errorResult(fmt.Sprintf("error listing labels: %v", err)), nil
	}

	// Apply name filter if provided
	nameContains, _ := args["name_contains"].(string)
	if nameContains != "" {
		labels = h.filterLabelsByName(labels, nameContains)
	}

	formatStr, _ := args["format"].(string)
	format := formatter.ParseFormat(formatStr)

	if format == formatter.FormatJSON {
		return h.formatListJSON(labels)
	}
	return h.formatListNatural(labels)
}

func (h *LabelHandlers) handleGet(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	labelID, _ := args["label_id"].(string)
	if labelID == "" {
		return errorResult("label_id is required for get action"), nil
	}

	labels, err := client.GetLabelRegistry(ctx)
	if err != nil {
		return errorResult(fmt.Sprintf("error getting labels: %v", err)), nil
	}

	var label *homeassistant.LabelRegistryEntry
	for i := range labels {
		if labels[i].LabelID == labelID {
			label = &labels[i]
			break
		}
	}

	if label == nil {
		return errorResult(fmt.Sprintf("label not found: %s", labelID)), nil
	}

	formatStr, _ := args["format"].(string)
	format := formatter.ParseFormat(formatStr)

	if format == formatter.FormatJSON {
		return h.formatGetJSON(label)
	}
	return h.formatGetNatural(label)
}

func (h *LabelHandlers) handleCreate(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	name, _ := args["name"].(string)
	if name == "" {
		return errorResult("name is required for create action"), nil
	}

	config := h.buildLabelConfig(args)
	config.Name = name

	label, err := client.CreateLabel(ctx, config)
	if err != nil {
		return errorResult(fmt.Sprintf("error creating label: %v", err)), nil
	}

	return successResult(fmt.Sprintf("Label '%s' created successfully with ID: %s", label.Name, label.LabelID)), nil
}

func (h *LabelHandlers) handleUpdate(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	labelID, _ := args["label_id"].(string)
	if labelID == "" {
		return errorResult("label_id is required for update action"), nil
	}

	config := h.buildLabelConfig(args)

	label, err := client.UpdateLabel(ctx, labelID, config)
	if err != nil {
		return errorResult(fmt.Sprintf("error updating label: %v", err)), nil
	}

	return successResult(fmt.Sprintf("Label '%s' updated successfully", label.Name)), nil
}

func (h *LabelHandlers) handleDelete(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	labelID, _ := args["label_id"].(string)
	if labelID == "" {
		return errorResult("label_id is required for delete action"), nil
	}

	if err := client.DeleteLabel(ctx, labelID); err != nil {
		return errorResult(fmt.Sprintf("error deleting label: %v", err)), nil
	}

	return successResult(fmt.Sprintf("Label '%s' deleted successfully", labelID)), nil
}

// =============================================================================
// Helper Functions
// =============================================================================

func (h *LabelHandlers) buildLabelConfig(args map[string]any) homeassistant.LabelConfig {
	config := homeassistant.LabelConfig{}

	if name, ok := args["name"].(string); ok && name != "" {
		config.Name = name
	}
	if color, ok := args["color"].(string); ok && color != "" {
		config.Color = color
	}
	if icon, ok := args["icon"].(string); ok && icon != "" {
		config.Icon = icon
	}
	if description, ok := args["description"].(string); ok && description != "" {
		config.Description = description
	}

	return config
}

func (h *LabelHandlers) filterLabelsByName(labels []homeassistant.LabelRegistryEntry, nameContains string) []homeassistant.LabelRegistryEntry {
	filtered := make([]homeassistant.LabelRegistryEntry, 0)
	lowerSearch := strings.ToLower(nameContains)

	for _, label := range labels {
		if strings.Contains(strings.ToLower(label.Name), lowerSearch) {
			filtered = append(filtered, label)
		}
	}

	return filtered
}

// =============================================================================
// Formatting Methods
// =============================================================================

func (h *LabelHandlers) formatListJSON(labels []homeassistant.LabelRegistryEntry) (*mcp.ToolsCallResult, error) {
	return jsonResult(map[string]any{
		"labels": labels,
		"count":  len(labels),
	})
}

func (h *LabelHandlers) formatListNatural(labels []homeassistant.LabelRegistryEntry) (*mcp.ToolsCallResult, error) {
	if len(labels) == 0 {
		return successResult("No labels found."), nil
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Found %d label(s):\n", len(labels)))

	for _, label := range labels {
		line := fmt.Sprintf("• %s (ID: %s)", label.Name, label.LabelID)
		if label.Color != "" {
			line += fmt.Sprintf(" - Color: %s", label.Color)
		}
		if label.Icon != "" {
			line += fmt.Sprintf(" - Icon: %s", label.Icon)
		}
		if label.Description != "" {
			line += fmt.Sprintf("\n  Description: %s", label.Description)
		}
		parts = append(parts, line)
	}

	return successResult(strings.Join(parts, "\n")), nil
}

func (h *LabelHandlers) formatGetJSON(label *homeassistant.LabelRegistryEntry) (*mcp.ToolsCallResult, error) {
	return jsonResult(label)
}

func (h *LabelHandlers) formatGetNatural(label *homeassistant.LabelRegistryEntry) (*mcp.ToolsCallResult, error) {
	var parts []string
	parts = append(parts,
		fmt.Sprintf("Label: %s", label.Name),
		fmt.Sprintf("ID: %s", label.LabelID))

	if label.Color != "" {
		parts = append(parts, fmt.Sprintf("Color: %s", label.Color))
	}
	if label.Icon != "" {
		parts = append(parts, fmt.Sprintf("Icon: %s", label.Icon))
	}
	if label.Description != "" {
		parts = append(parts, fmt.Sprintf("Description: %s", label.Description))
	}

	return successResult(strings.Join(parts, "\n")), nil
}
