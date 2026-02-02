// Package handlers provides MCP tool handlers for Home Assistant operations.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// ConfigEntryHandlers provides MCP tool handlers for config entry operations.
type ConfigEntryHandlers struct{}

// NewConfigEntryHandlers creates a new ConfigEntryHandlers instance.
func NewConfigEntryHandlers() *ConfigEntryHandlers {
	return &ConfigEntryHandlers{}
}

// RegisterTools registers all config entry tools with the registry.
func (h *ConfigEntryHandlers) RegisterTools(registry *mcp.Registry) {
	registry.RegisterTool(h.listConfigEntriesTool(), h.handleListConfigEntries)
	registry.RegisterTool(h.getConfigEntryTool(), h.handleGetConfigEntry)
}

// listConfigEntriesTool returns the tool definition for listing config entries.
func (h *ConfigEntryHandlers) listConfigEntriesTool() mcp.Tool {
	return mcp.Tool{
		Name:        "list_config_entries",
		Description: "List entries in the Home Assistant config entry registry. Config entries store metadata about integrations and helpers (domain, title, state, entry_id). Use 'domain' filter to narrow down results (e.g., 'template', 'hue', 'zwave_js'). Note: Template definitions are not exposed through this API.",
		InputSchema: mcp.JSONSchema{
			Type:        "object",
			Description: "Filter options for config entries",
			Properties: map[string]mcp.JSONSchema{
				"domain": {
					Type:        "string",
					Description: "Filter by domain (e.g., 'template', 'hue', 'zwave_js'). If not specified, returns all config entries.",
				},
			},
		},
	}
}

// getConfigEntryTool returns the tool definition for getting a single config entry.
func (h *ConfigEntryHandlers) getConfigEntryTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_config_entry",
		Description: "Get a single config entry by its entry ID. Returns metadata about the config entry (domain, title, state, capabilities). Use list_entity_registry with verbose=true to find the config_entry_id for a specific entity. Note: Template definitions are stored but not exposed through this API.",
		InputSchema: mcp.JSONSchema{
			Type:        "object",
			Description: "Config entry identifier",
			Properties: map[string]mcp.JSONSchema{
				"entry_id": {
					Type:        "string",
					Description: "The config entry ID to retrieve (e.g., from entity registry's config_entry_id field)",
				},
			},
			Required: []string{"entry_id"},
		},
	}
}

// handleListConfigEntries handles requests to list config entries.
func (h *ConfigEntryHandlers) handleListConfigEntries(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	domain := getString(args, "domain")

	entries, err := client.GetConfigEntries(ctx, domain)
	if err != nil {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{
				mcp.NewTextContent(fmt.Sprintf("Error getting config entries: %v", err)),
			},
			IsError: true,
		}, nil
	}

	output, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{
				mcp.NewTextContent(fmt.Sprintf("Error formatting response: %v", err)),
			},
			IsError: true,
		}, nil
	}

	summary := fmt.Sprintf("Found %d config entries", len(entries))
	if domain != "" {
		summary += fmt.Sprintf(" for domain '%s'", domain)
	}

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{
			mcp.NewTextContent(summary + "\n\n" + string(output)),
		},
	}, nil
}

// handleGetConfigEntry handles requests to get a single config entry.
func (h *ConfigEntryHandlers) handleGetConfigEntry(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	entryID := getString(args, "entry_id")
	if entryID == "" {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{
				mcp.NewTextContent("entry_id is required"),
			},
			IsError: true,
		}, nil
	}

	entry, err := client.GetConfigEntry(ctx, entryID)
	if err != nil {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{
				mcp.NewTextContent(fmt.Sprintf("Error getting config entry: %v", err)),
			},
			IsError: true,
		}, nil
	}

	output, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{
				mcp.NewTextContent(fmt.Sprintf("Error formatting response: %v", err)),
			},
			IsError: true,
		}, nil
	}

	summary := fmt.Sprintf("Config entry '%s' (%s)", entry.Title, entry.Domain)
	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{
			mcp.NewTextContent(summary + "\n\n" + string(output)),
		},
	}, nil
}
