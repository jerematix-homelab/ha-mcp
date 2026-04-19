package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// Action constants for manage_blueprint tool.
const (
	blueprintActionList   = "list"
	blueprintActionImport = "import"

	blueprintURLMaxLength = 2048
)

// loopbackHostnames are hostname literals that resolve to loopback and must be blocked.
var loopbackHostnames = map[string]bool{
	"localhost":     true,
	"ip6-localhost": true,
	"ip6-loopback":  true,
}

// validateBlueprintURL rejects URLs that could enable SSRF when forwarded to HA's blueprint/import.
// It enforces https-only and blocks literal private/loopback/link-local IP addresses.
// Note: hostname-based DNS rebinding is not prevented here because HA performs the actual fetch.
func validateBlueprintURL(rawURL string) error {
	if len(rawURL) > blueprintURLMaxLength {
		return fmt.Errorf("url too long (max %d characters)", blueprintURLMaxLength)
	}
	u, err := url.Parse(rawURL)
	if err != nil || !u.IsAbs() || u.Host == "" {
		return fmt.Errorf("invalid url: must be an absolute https:// URL")
	}
	if u.Scheme != "https" {
		return fmt.Errorf("url must use https scheme, got %q", u.Scheme)
	}
	hostname := u.Hostname()
	if loopbackHostnames[strings.ToLower(hostname)] {
		return fmt.Errorf("url targets a loopback address")
	}
	ip := net.ParseIP(hostname)
	if ip == nil {
		return nil
	}
	switch {
	case ip.IsLoopback():
		return fmt.Errorf("url targets a loopback address")
	case ip.IsPrivate():
		return fmt.Errorf("url targets a private network address")
	case ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast():
		return fmt.Errorf("url targets a link-local address")
	case ip.IsUnspecified():
		return fmt.Errorf("url targets an unspecified address")
	case ip.IsMulticast():
		return fmt.Errorf("url targets a multicast address")
	}
	return nil
}

// BlueprintHandlers provides handlers for blueprint operations.
type BlueprintHandlers struct{}

// NewBlueprintHandlers creates a new blueprint handlers instance.
func NewBlueprintHandlers() *BlueprintHandlers {
	return &BlueprintHandlers{}
}

// RegisterBlueprintTools registers blueprint-related tools with the MCP registry.
func RegisterBlueprintTools(registry *mcp.Registry) {
	handler := NewBlueprintHandlers()

	registry.RegisterTool(mcp.Tool{
		Name:        "manage_blueprint",
		Description: "Manage Home Assistant blueprints for automations and scripts. Supports list (view available blueprints) and import (import from URL).",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.JSONSchema{
				"action": {
					Type:        "string",
					Description: "Action to perform: 'list' (list available blueprints) or 'import' (import blueprint from URL).",
					Enum:        []string{blueprintActionList, blueprintActionImport},
				},
				"domain": {
					Type:        "string",
					Description: "Domain for blueprints: 'automation' or 'script' (required for 'list' action).",
					Enum:        []string{traceDomainAutomation, traceDomainScript},
				},
				"url": {
					Type:        "string",
					Description: "Blueprint URL (required for 'import' action). Must use https:// and point to a publicly reachable host. Example: https://raw.githubusercontent.com/<user>/<repo>/main/blueprint.yaml",
				},
				"format": {
					Type:        "string",
					Description: "Output format: 'natural' (default, human-readable) or 'json' (structured JSON).",
					Enum:        []string{"natural", "json"},
					Default:     "natural",
				},
			},
			Required: []string{"action"},
		},
	}, handler.HandleManageBlueprint)
}

// HandleManageBlueprint handles the manage_blueprint tool invocation.
func (h *BlueprintHandlers) HandleManageBlueprint(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	// Extract action
	action, _ := args["action"].(string)
	if action == "" {
		return errorResult("action parameter is required and must be 'list' or 'import'"), nil
	}

	// Extract format (default: natural)
	format, _ := args["format"].(string)
	if format == "" {
		format = formatNatural
	}

	// Route to action handler
	switch action {
	case blueprintActionList:
		return h.handleListBlueprints(ctx, client, args, format)
	case blueprintActionImport:
		return h.handleImportBlueprint(ctx, client, args, format)
	default:
		return errorResult(fmt.Sprintf("invalid action %q, must be one of: list, import", action)), nil
	}
}

// handleListBlueprints lists all blueprints for a domain.
func (h *BlueprintHandlers) handleListBlueprints(ctx context.Context, client homeassistant.Client, args map[string]any, format string) (*mcp.ToolsCallResult, error) {
	// Extract domain (required for list)
	domain, _ := args["domain"].(string)
	if domain == "" {
		return errorResult("domain is required for list action"), nil
	}

	// Build command data
	data := map[string]any{
		"domain": domain,
	}

	// Call blueprint/list WebSocket command
	response, err := client.SendHACSCommand(ctx, "blueprint/list", data)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to list blueprints: %v", err)), nil
	}

	// Format output
	if format == formatJSON {
		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return errorResult(fmt.Sprintf("failed to marshal blueprints: %v", err)), nil
		}
		return successResult(string(jsonData)), nil
	}

	// Natural format
	return successResult(h.formatBlueprintsNatural(response, domain)), nil
}

// handleImportBlueprint imports a blueprint from a URL.
func (h *BlueprintHandlers) handleImportBlueprint(ctx context.Context, client homeassistant.Client, args map[string]any, format string) (*mcp.ToolsCallResult, error) {
	// Validate required parameters
	blueprintURL, _ := args["url"].(string)
	if blueprintURL == "" {
		return errorResult("url is required for import action"), nil
	}
	if err := validateBlueprintURL(blueprintURL); err != nil {
		return errorResult(fmt.Sprintf("invalid blueprint url: %v", err)), nil
	}

	// Build command data
	data := map[string]any{
		"url": blueprintURL,
	}

	// Call blueprint/import WebSocket command
	response, err := client.SendHACSCommand(ctx, "blueprint/import", data)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to import blueprint: %v", err)), nil
	}

	// Format output
	if format == formatJSON {
		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return errorResult(fmt.Sprintf("failed to marshal import result: %v", err)), nil
		}
		return successResult(string(jsonData)), nil
	}

	// Natural format
	return successResult(fmt.Sprintf("Blueprint successfully imported from: %s", blueprintURL)), nil
}

// formatBlueprintsNatural formats blueprints in natural language.
func (h *BlueprintHandlers) formatBlueprintsNatural(response any, domain string) string {
	blueprintsMap, ok := response.(map[string]any)
	if !ok || len(blueprintsMap) == 0 {
		return fmt.Sprintf("No %s blueprints found.", domain)
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Found %d %s blueprint(s):", len(blueprintsMap), domain))

	i := 1
	for path, data := range blueprintsMap {
		parts = append(parts, fmt.Sprintf("\n%d. Path: %s", i, path))

		if dataMap, ok := data.(map[string]any); ok {
			if metadata, ok := dataMap["metadata"].(map[string]any); ok {
				name := getMapString(metadata, "name", "")
				source := getMapString(metadata, "source", "")

				if name != "" {
					parts = append(parts, fmt.Sprintf("   Name: %s", name))
				}
				if source != "" {
					parts = append(parts, fmt.Sprintf("   Source: %s", source))
				}
			}
		}
		i++
	}

	return strings.Join(parts, "\n")
}
