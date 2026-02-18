// Package handlers provides MCP tool handlers for Home Assistant operations.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/jsonpatch"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// Dashboard action constants.
const (
	dashboardActionList       = "list"
	dashboardActionGet        = "get"
	dashboardActionCreate     = "create"
	dashboardActionUpdate     = "update"
	dashboardActionDelete     = "delete"
	dashboardActionSaveConfig = "save_config"
	dashboardActionPatch      = "patch"
)

// DashboardHandlers provides handlers for dashboard-related MCP tools.
type DashboardHandlers struct{}

// NewDashboardHandlers creates a new DashboardHandlers instance.
func NewDashboardHandlers() *DashboardHandlers {
	return &DashboardHandlers{}
}

// RegisterTools registers the consolidated manage_dashboard tool with the registry.
func (h *DashboardHandlers) RegisterTools(registry *mcp.Registry) {
	registry.RegisterTool(h.manageDashboardTool(), h.handleManageDashboard)
}

// =============================================================================
// Tool Definition
// =============================================================================

func (h *DashboardHandlers) manageDashboardTool() mcp.Tool {
	schema := h.buildDashboardSchema()
	return mcp.Tool{
		Name: "manage_dashboard",
		Description: `Manage Home Assistant Lovelace dashboards - list, get, create, update, delete, or save configuration.

Actions:
- list: List all dashboards
- get: Get dashboard configuration with optional view filtering (url_path optional)
- create: Create a new dashboard with modern section-based layout (requires url_path, title, and mode)
- update: Update an existing dashboard (requires dashboard_id)
- delete: Delete a dashboard (requires dashboard_id)
- save_config: Save dashboard configuration (requires config, url_path optional)
- patch: Apply RFC 6902 JSON Patch operations to dashboard config (requires operations, url_path optional)

Note: Newly created dashboards use the modern "sections" layout format instead of the legacy "badges/cards" format.`,
		InputSchema: schema,
	}
}

func (h *DashboardHandlers) buildDashboardSchema() mcp.JSONSchema {
	return mcp.JSONSchema{
		Type:        "object",
		Description: "Dashboard management operation",
		Properties: map[string]mcp.JSONSchema{
			"action": {
				Type:        "string",
				Description: "Operation to perform: list, get, create, update, delete, save_config, patch",
				Enum:        []string{"list", "get", "create", "update", "delete", "save_config", "patch"},
			},
			"dashboard_id": {
				Type:        "string",
				Description: "HA-generated dashboard identifier (e.g., 'lovelace-xxxx'), returned by list/create. Required for update/delete.",
			},
			"url_path": {
				Type:        "string",
				Description: "Dashboard URL path (e.g., 'energy', 'leak-sensors'). Use hyphens for multi-word paths. Required for create. Optional for get/save_config (empty = default dashboard).",
			},
			"title": {
				Type:        "string",
				Description: "Dashboard title (required for create, optional for update)",
			},
			"icon": {
				Type:        "string",
				Description: "Dashboard icon (e.g., 'mdi:view-dashboard'). Omit rather than set empty.",
			},
			"mode": {
				Type:        "string",
				Description: "Dashboard mode: 'storage' or 'yaml'. Required for create.",
				Enum:        []string{"storage", "yaml"},
			},
			"require_admin": {
				Type:        "boolean",
				Description: "Whether dashboard requires admin access. Defaults to false for create.",
			},
			"show_in_sidebar": {
				Type:        "boolean",
				Description: "Whether to show dashboard in sidebar. Defaults to false for create.",
			},
			"config": {
				Type:        "object",
				Description: "Dashboard configuration object (for save_config action)",
			},
			"view": {
				Type:        "string",
				Description: "Filter by view path or title (for get action, case-insensitive, partial match)",
			},
			"verbose": {
				Type:        "boolean",
				Description: "If true, return full configuration including all cards (for get action). Default: false (compact overview)",
			},
			"format": {
				Type:        "string",
				Description: "Output format: 'natural' (default, human-readable) or 'json' (structured data)",
				Enum:        []string{"natural", "json"},
			},
			"operations": patchOperationsSchema(),
		},
		Required: []string{"action"},
	}
}

// =============================================================================
// Main Dispatcher
// =============================================================================

func (h *DashboardHandlers) handleManageDashboard(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	action, ok := args["action"].(string)
	if !ok {
		return errorResult("action is required and must be a string"), nil
	}

	switch action {
	case dashboardActionList:
		return h.handleList(ctx, client, args)
	case dashboardActionGet:
		return h.handleGet(ctx, client, args)
	case dashboardActionCreate:
		return h.handleCreate(ctx, client, args)
	case dashboardActionUpdate:
		return h.handleUpdate(ctx, client, args)
	case dashboardActionDelete:
		return h.handleDelete(ctx, client, args)
	case dashboardActionSaveConfig:
		return h.handleSaveConfig(ctx, client, args)
	case dashboardActionPatch:
		return h.handlePatch(ctx, client, args)
	default:
		return errorResult(fmt.Sprintf("invalid action: %s. Valid actions: list, get, create, update, delete, save_config, patch", action)), nil
	}
}

// =============================================================================
// Action Handlers
// =============================================================================

func (h *DashboardHandlers) handleList(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	dashboards, err := client.ListDashboards(ctx)
	if err != nil {
		return errorResult(fmt.Sprintf("error listing dashboards: %v", err)), nil
	}

	formatStr, _ := args["format"].(string)
	if formatStr == formatJSON {
		return h.formatListJSON(dashboards)
	}
	return h.formatListNatural(dashboards)
}

func (h *DashboardHandlers) handleGet(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	urlPath, _ := args["url_path"].(string)
	viewFilter, _ := args["view"].(string)
	verbose, _ := args["verbose"].(bool)

	config, err := client.GetLovelaceConfig(ctx, urlPath)
	if err != nil {
		return errorResult(fmt.Sprintf("error getting dashboard configuration: %v", err)), nil
	}

	views, _ := config["views"].([]any)

	if viewFilter != "" {
		return h.handleFilteredViews(views, viewFilter)
	}

	if verbose {
		summary := fmt.Sprintf("Lovelace configuration with %d views", len(views))
		return formatLovelaceResponse(config, summary)
	}

	return h.handleCompactViews(views)
}

func (h *DashboardHandlers) handleCreate(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	urlPath, ok := args["url_path"].(string)
	if !ok || urlPath == "" {
		return errorResult("url_path is required for create action"), nil
	}

	title, ok := args["title"].(string)
	if !ok || title == "" {
		return errorResult("title is required for create action"), nil
	}

	mode, ok := args["mode"].(string)
	if !ok || mode == "" {
		return errorResult("mode is required for create action (use 'storage' or 'yaml')"), nil
	}

	config := buildDashboardConfig(args)
	config.URLPath = urlPath
	config.Title = title

	// Default nil booleans to false for create
	if config.RequireAdmin == nil {
		falseVal := false
		config.RequireAdmin = &falseVal
	}
	if config.ShowInSidebar == nil {
		falseVal := false
		config.ShowInSidebar = &falseVal
	}

	dashboard, err := client.CreateDashboard(ctx, config)
	if err != nil {
		return errorResult(fmt.Sprintf("error creating dashboard: %v", err)), nil
	}

	// Initialize dashboard with modern section-based layout
	return h.initializeDashboardLayout(ctx, client, dashboard, urlPath, args)
}

func (h *DashboardHandlers) handleUpdate(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	dashboardID, ok := args["dashboard_id"].(string)
	if !ok || dashboardID == "" {
		return errorResult("dashboard_id is required for update action"), nil
	}

	config := buildDashboardConfig(args)

	dashboard, err := client.UpdateDashboard(ctx, dashboardID, config)
	if err != nil {
		return errorResult(fmt.Sprintf("error updating dashboard: %v", err)), nil
	}

	formatStr, _ := args["format"].(string)
	if formatStr == formatJSON {
		return h.formatDashboardJSON(dashboard, "Dashboard updated successfully")
	}
	return h.formatDashboardNatural(dashboard, "Dashboard updated successfully")
}

func (h *DashboardHandlers) handleDelete(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	dashboardID, ok := args["dashboard_id"].(string)
	if !ok || dashboardID == "" {
		return errorResult("dashboard_id is required for delete action"), nil
	}

	err := client.DeleteDashboard(ctx, dashboardID)
	if err != nil {
		return errorResult(fmt.Sprintf("error deleting dashboard: %v", err)), nil
	}

	return successResult(fmt.Sprintf("Dashboard deleted successfully: %s", dashboardID)), nil
}

func (h *DashboardHandlers) handleSaveConfig(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	config, ok := args["config"].(map[string]any)
	if !ok {
		return errorResult("config is required for save_config action"), nil
	}

	urlPath, _ := args["url_path"].(string)

	err := client.SaveLovelaceConfig(ctx, urlPath, config)
	if err != nil {
		return errorResult(fmt.Sprintf("error saving dashboard configuration: %v", err)), nil
	}

	target := "default dashboard"
	if urlPath != "" {
		target = fmt.Sprintf("dashboard '%s'", urlPath)
	}
	return successResult(fmt.Sprintf("Dashboard configuration saved successfully for %s", target)), nil
}

//nolint:funlen // Patch handler
func (h *DashboardHandlers) handlePatch(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	urlPath, _ := args["url_path"].(string)

	ops, errResult := parseOperations(args)
	if errResult != nil {
		return errResult, nil
	}

	config, err := client.GetLovelaceConfig(ctx, urlPath)
	if err != nil {
		return errorResult(fmt.Sprintf("error getting dashboard configuration: %v", err)), nil
	}

	patchedAny, patchErr := jsonpatch.Apply(config, ops)
	if patchErr != nil {
		return errorResult(fmt.Sprintf("error applying patch: %v", patchErr)), nil
	}

	patchedMap, ok := patchedAny.(map[string]any)
	if !ok {
		return errorResult("patch result must be an object"), nil
	}

	if err := client.SaveLovelaceConfig(ctx, urlPath, patchedMap); err != nil {
		return errorResult(fmt.Sprintf("error saving patched dashboard: %v", err)), nil
	}

	target := "default dashboard"
	if urlPath != "" {
		target = fmt.Sprintf("dashboard '%s'", urlPath)
	}

	return successResult(fmt.Sprintf("Dashboard patched successfully for %s (%d operations applied)", target, len(ops))), nil
}

// =============================================================================
// Helper Methods
// =============================================================================

// initializeDashboardLayout initializes a newly created dashboard with modern section-based layout.
func (h *DashboardHandlers) initializeDashboardLayout(
	ctx context.Context,
	client homeassistant.Client,
	dashboard *homeassistant.DashboardEntry,
	urlPath string,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	// Build modern section-based layout config
	defaultConfig := map[string]any{
		"views": []any{
			map[string]any{
				"title": "Home",
				"type":  "sections", // Modern section-based layout
				"sections": []any{
					map[string]any{
						"type":  "grid",
						"cards": []any{}, // Empty cards array - user can add cards later
					},
				},
			},
		},
	}

	// Wait briefly for dashboard to be fully initialized
	// This gives Home Assistant time to create the config entry
	time.Sleep(500 * time.Millisecond)

	// Save the default configuration with modern section layout
	// Non-fatal if this fails - dashboard was created successfully
	if err := client.SaveLovelaceConfig(ctx, urlPath, defaultConfig); err != nil {
		// Dashboard exists, but layout initialization failed
		// Return success with a note about manual configuration
		formatStr, _ := args["format"].(string)
		successMsg := fmt.Sprintf("Dashboard created successfully. Note: Could not initialize default layout (%v). Please configure views manually.", err)
		if formatStr == formatJSON {
			return h.formatDashboardJSON(dashboard, successMsg)
		}
		return h.formatDashboardNatural(dashboard, successMsg)
	}

	formatStr, _ := args["format"].(string)
	if formatStr == formatJSON {
		return h.formatDashboardJSON(dashboard, "Dashboard created successfully with modern section layout")
	}
	return h.formatDashboardNatural(dashboard, "Dashboard created successfully with modern section layout")
}

// =============================================================================
// Formatting Methods
// =============================================================================

func (h *DashboardHandlers) formatListJSON(dashboards []homeassistant.DashboardEntry) (*mcp.ToolsCallResult, error) {
	result := map[string]any{
		"count":      len(dashboards),
		"dashboards": dashboards,
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errorResult(fmt.Sprintf("error formatting dashboards: %v", err)), nil
	}

	summary := fmt.Sprintf("Found %d dashboards", len(dashboards))
	return successResult(summary + "\n\n" + string(output)), nil
}

func (h *DashboardHandlers) formatListNatural(dashboards []homeassistant.DashboardEntry) (*mcp.ToolsCallResult, error) {
	if len(dashboards) == 0 {
		return successResult("No dashboards found"), nil
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Found %d dashboard(s):\n", len(dashboards)))

	for i, d := range dashboards {
		var attrs []string
		if d.Icon != "" {
			attrs = append(attrs, fmt.Sprintf("icon: %s", d.Icon))
		}
		if d.RequireAdmin {
			attrs = append(attrs, "admin required")
		}
		if d.ShowInSidebar {
			attrs = append(attrs, "in sidebar")
		}

		attrStr := ""
		if len(attrs) > 0 {
			attrStr = fmt.Sprintf(" [%s]", strings.Join(attrs, ", "))
		}

		parts = append(parts, fmt.Sprintf("%d. %s (/%s) - ID: %s, Mode: %s%s",
			i+1, d.Title, d.URLPath, d.ID, d.Mode, attrStr))
	}

	return successResult(strings.Join(parts, "\n")), nil
}

func (h *DashboardHandlers) formatDashboardJSON(dashboard *homeassistant.DashboardEntry, message string) (*mcp.ToolsCallResult, error) {
	output, err := json.MarshalIndent(dashboard, "", "  ")
	if err != nil {
		return errorResult(fmt.Sprintf("error formatting dashboard: %v", err)), nil
	}

	return successResult(message + "\n\n" + string(output)), nil
}

func (h *DashboardHandlers) formatDashboardNatural(dashboard *homeassistant.DashboardEntry, message string) (*mcp.ToolsCallResult, error) {
	var parts []string
	parts = append(parts,
		message,
		"",
		fmt.Sprintf("Dashboard: %s", dashboard.Title),
		fmt.Sprintf("URL Path: /%s", dashboard.URLPath),
		fmt.Sprintf("ID: %s", dashboard.ID),
		fmt.Sprintf("Mode: %s", dashboard.Mode),
	)

	if dashboard.Icon != "" {
		parts = append(parts, fmt.Sprintf("Icon: %s", dashboard.Icon))
	}
	parts = append(parts,
		fmt.Sprintf("Require Admin: %t", dashboard.RequireAdmin),
		fmt.Sprintf("Show in Sidebar: %t", dashboard.ShowInSidebar),
	)

	return successResult(strings.Join(parts, "\n")), nil
}

// =============================================================================
// Helper Functions (preserved from original implementation)
// =============================================================================

// compactViewEntry represents a minimal view entry for compact output.
type compactViewEntry struct {
	Title      string `json:"title,omitempty"`
	Path       string `json:"path,omitempty"`
	Icon       string `json:"icon,omitempty"`
	CardCount  int    `json:"card_count"`
	BadgeCount int    `json:"badge_count,omitempty"`
	Subview    bool   `json:"subview,omitempty"`
}

// filterViewsByQuery filters views by title or path (case-insensitive, partial match).
func filterViewsByQuery(views []any, query string) []any {
	queryLower := strings.ToLower(query)
	filtered := make([]any, 0)

	for _, v := range views {
		viewMap, ok := v.(map[string]any)
		if !ok {
			continue
		}

		title, _ := viewMap["title"].(string)
		path, _ := viewMap["path"].(string)

		if strings.Contains(strings.ToLower(title), queryLower) ||
			strings.Contains(strings.ToLower(path), queryLower) {
			filtered = append(filtered, viewMap)
		}
	}

	return filtered
}

// countCardsInView counts all cards in a view, including cards in sections.
func countCardsInView(viewMap map[string]any) int {
	count := 0

	if cards, ok := viewMap["cards"].([]any); ok {
		count = len(cards)
	}

	if sections, ok := viewMap["sections"].([]any); ok {
		for _, section := range sections {
			if sectionMap, ok := section.(map[string]any); ok {
				if sectionCards, ok := sectionMap["cards"].([]any); ok {
					count += len(sectionCards)
				}
			}
		}
	}

	return count
}

// buildCompactViewEntry creates a compact view entry from a view map.
func buildCompactViewEntry(viewMap map[string]any) compactViewEntry {
	entry := compactViewEntry{}
	entry.Title, _ = viewMap["title"].(string)
	entry.Path, _ = viewMap["path"].(string)
	entry.Icon, _ = viewMap["icon"].(string)
	entry.Subview, _ = viewMap["subview"].(bool)
	entry.CardCount = countCardsInView(viewMap)

	if badges, ok := viewMap["badges"].([]any); ok {
		entry.BadgeCount = len(badges)
	}

	return entry
}

// buildCompactViews converts views to compact format.
func buildCompactViews(views []any) []compactViewEntry {
	compact := make([]compactViewEntry, 0, len(views))

	for _, v := range views {
		viewMap, ok := v.(map[string]any)
		if !ok {
			continue
		}
		compact = append(compact, buildCompactViewEntry(viewMap))
	}

	return compact
}

// formatLovelaceResponse creates a ToolsCallResult with JSON-formatted data.
func formatLovelaceResponse(data any, summary string) (*mcp.ToolsCallResult, error) {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{
				mcp.NewTextContent(fmt.Sprintf("Error formatting response: %v", err)),
			},
			IsError: true,
		}, nil
	}

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{
			mcp.NewTextContent(summary + "\n\n" + string(output)),
		},
	}, nil
}

// handleFilteredViews handles requests with a view filter.
func (h *DashboardHandlers) handleFilteredViews(views []any, filter string) (*mcp.ToolsCallResult, error) {
	filteredViews := filterViewsByQuery(views, filter)

	if len(filteredViews) == 0 {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{
				mcp.NewTextContent(fmt.Sprintf("No views found matching '%s'", filter)),
			},
		}, nil
	}

	summary := fmt.Sprintf("Found %d view(s) matching '%s'", len(filteredViews), filter)
	return formatLovelaceResponse(filteredViews, summary)
}

// handleCompactViews handles requests for compact view output.
func (h *DashboardHandlers) handleCompactViews(views []any) (*mcp.ToolsCallResult, error) {
	compact := buildCompactViews(views)
	summary := fmt.Sprintf("Found %d views", len(compact)) + VerboseHint
	return formatLovelaceResponse(compact, summary)
}

// buildDashboardConfig builds a DashboardConfig from args map.
func buildDashboardConfig(args map[string]any) homeassistant.DashboardConfig {
	config := homeassistant.DashboardConfig{}

	if title, ok := args["title"].(string); ok {
		config.Title = title
	}
	if icon, ok := args["icon"].(string); ok {
		config.Icon = icon
	}
	if mode, ok := args["mode"].(string); ok {
		config.Mode = mode
	}
	if requireAdmin, ok := args["require_admin"].(bool); ok {
		config.RequireAdmin = &requireAdmin
	}
	if showInSidebar, ok := args["show_in_sidebar"].(bool); ok {
		config.ShowInSidebar = &showInSidebar
	}

	return config
}
