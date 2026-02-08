// Package handlers provides MCP tool handlers for Home Assistant operations.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zorak1103/ha-mcp/internal/handlers/formatter"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

const configKeyEntityID = "entity_id"
const configKeyTarget = "target"

// EntityHandlers provides MCP tool handlers for entity operations.
type EntityHandlers struct{}

// NewEntityHandlers creates a new EntityHandlers instance.
func NewEntityHandlers() *EntityHandlers {
	return &EntityHandlers{}
}

// RegisterTools registers all entity-related tools with the registry.
func (h *EntityHandlers) RegisterTools(registry *mcp.Registry) {
	registry.RegisterTool(h.getStateTool(), h.handleGetState)
	registry.RegisterTool(h.getEntityDependenciesTool(), h.handleGetEntityDependencies)
}

func (h *EntityHandlers) getStateTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_state",
		Description: "Get the state of one or more entities. By default returns natural language format optimized for LLMs. Use 'format=json' for structured data.",
		InputSchema: mcp.JSONSchema{
			Type:        "object",
			Description: "Parameters for getting entity state(s)",
			Properties: map[string]mcp.JSONSchema{
				"entity_id": {
					Type:        "string",
					Description: "Single entity ID (e.g., 'light.living_room'). Use entity_id OR entity_ids, not both.",
				},
				"entity_ids": {
					Type:        "array",
					Description: "Array of entity IDs for batch query (e.g., ['light.living_room', 'light.bedroom']). Use entity_id OR entity_ids, not both.",
				},
				"format": {
					Type:        "string",
					Description: "Output format: 'natural' (default, human-readable LLM-optimized) or 'json' (structured data)",
					Enum:        []string{"natural", "json"},
				},
			},
		},
	}
}

func (h *EntityHandlers) getEntityDependenciesTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_entity_dependencies",
		Description: "Find all automations that use a specific entity. Shows where the entity is used as trigger, condition, or action target. Useful for understanding the impact of changing or removing an entity.",
		InputSchema: mcp.JSONSchema{
			Type:        "object",
			Description: "Parameters for finding entity dependencies",
			Properties: map[string]mcp.JSONSchema{
				"entity_id": {
					Type:        "string",
					Description: "The entity ID to search for (e.g., 'binary_sensor.motion_living_room')",
				},
			},
			Required: []string{"entity_id"},
		},
	}
}

// getStringArg safely extracts a string argument.
func getStringArg(args map[string]any, key string) string {
	val, _ := args[key].(string)
	return val
}

// getBoolArg safely extracts a boolean argument.
func getBoolArg(args map[string]any, key string) bool {
	val, _ := args[key].(bool)
	return val
}

func (h *EntityHandlers) handleGetState(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	entityID, hasSingle := args["entity_id"].(string)
	entityIDs, hasBatch := args["entity_ids"]

	// Validate: exactly one of entity_id or entity_ids must be provided
	if hasSingle && hasBatch {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{mcp.NewTextContent("Cannot specify both entity_id and entity_ids. Use one or the other.")},
			IsError: true,
		}, nil
	}

	if !hasSingle && !hasBatch {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{mcp.NewTextContent("entity_id or entity_ids is required")},
			IsError: true,
		}, nil
	}

	// Single entity mode
	if hasSingle {
		if entityID == "" {
			return &mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{mcp.NewTextContent("entity_id is required")},
				IsError: true,
			}, nil
		}
		return h.handleGetStateSingle(ctx, client, entityID, args)
	}

	// Batch mode
	return h.handleGetStateBatch(ctx, client, entityIDs, args)
}

func (h *EntityHandlers) handleGetStateSingle(ctx context.Context, client homeassistant.Client, entityID string, args map[string]any) (*mcp.ToolsCallResult, error) {
	state, err := client.GetState(ctx, entityID)
	if err != nil {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{mcp.NewTextContent(fmt.Sprintf("Error getting state: %v", err))},
			IsError: true,
		}, nil
	}

	// Use formatter based on format parameter
	format := formatter.ParseFormat(getStringArg(args, "format"))
	f := formatter.New(format)

	output, err := f.FormatEntity(ctx, *state)
	if err != nil {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{mcp.NewTextContent(fmt.Sprintf("Error formatting state: %v", err))},
			IsError: true,
		}, nil
	}

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(output)},
	}, nil
}

func (h *EntityHandlers) handleGetStateBatch(ctx context.Context, client homeassistant.Client, entityIDsArg any, args map[string]any) (*mcp.ToolsCallResult, error) {
	// Parse entity_ids array
	entityIDsArray, ok := entityIDsArg.([]any)
	if !ok {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{mcp.NewTextContent("entity_ids must be an array")},
			IsError: true,
		}, nil
	}

	var entityIDs []string
	for _, id := range entityIDsArray {
		if idStr, ok := id.(string); ok && idStr != "" {
			entityIDs = append(entityIDs, idStr)
		}
	}

	if len(entityIDs) == 0 {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{mcp.NewTextContent("entity_ids array is empty")},
			IsError: true,
		}, nil
	}

	// Get all states and filter
	allStates, err := client.GetStates(ctx)
	if err != nil {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{mcp.NewTextContent(fmt.Sprintf("Error getting states: %v", err))},
			IsError: true,
		}, nil
	}

	// Build map for quick lookup
	stateMap := make(map[string]homeassistant.Entity)
	for _, state := range allStates {
		stateMap[state.EntityID] = state
	}

	// Collect states in order requested
	var foundStates []homeassistant.Entity
	var notFound []string
	for _, entityID := range entityIDs {
		if state, ok := stateMap[entityID]; ok {
			foundStates = append(foundStates, state)
		} else {
			notFound = append(notFound, entityID)
		}
	}

	// Use formatter based on format parameter
	format := formatter.ParseFormat(getStringArg(args, "format"))

	if format == formatter.FormatNatural {
		return h.formatBatchNatural(ctx, foundStates, notFound)
	}

	return h.formatBatchJSON(foundStates, notFound)
}

func (h *EntityHandlers) formatBatchNatural(ctx context.Context, states []homeassistant.Entity, notFound []string) (*mcp.ToolsCallResult, error) {
	var output strings.Builder
	f := formatter.New(formatter.FormatNatural)

	for _, state := range states {
		line, err := f.FormatEntity(ctx, state)
		if err != nil {
			return &mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{mcp.NewTextContent(fmt.Sprintf("Error formatting state: %v", err))},
				IsError: true,
			}, nil
		}
		output.WriteString(line)
		output.WriteString("\n")
	}

	// Add not found entities
	for _, entityID := range notFound {
		fmt.Fprintf(&output, "%s: not found\n", entityID)
	}

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(strings.TrimSuffix(output.String(), "\n"))},
	}, nil
}

func (h *EntityHandlers) formatBatchJSON(states []homeassistant.Entity, notFound []string) (*mcp.ToolsCallResult, error) {
	type batchResult struct {
		States   []homeassistant.Entity `json:"states"`
		NotFound []string               `json:"not_found,omitempty"`
	}

	result := batchResult{
		States:   states,
		NotFound: notFound,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{mcp.NewTextContent(fmt.Sprintf("Error formatting JSON: %v", err))},
			IsError: true,
		}, nil
	}

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(string(data))},
	}, nil
}

// entityDependency represents where an entity is used in an automation.
type entityDependency struct {
	AutomationID    string   `json:"automation_id"`
	AutomationAlias string   `json:"automation_alias"`
	UsedIn          []string `json:"used_in"` // "trigger", "condition", "action"
}

// entityDependenciesResult is the result of get_entity_dependencies.
type entityDependenciesResult struct {
	EntityID    string             `json:"entity_id"`
	Automations []entityDependency `json:"automations"`
	TotalUsages int                `json:"total_usages"`
}

func (h *EntityHandlers) handleGetEntityDependencies(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	entityID, ok := args["entity_id"].(string)
	if !ok || entityID == "" {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{mcp.NewTextContent("entity_id is required")},
			IsError: true,
		}, nil
	}

	automations, err := client.ListAutomations(ctx)
	if err != nil {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{mcp.NewTextContent(fmt.Sprintf("Error listing automations: %v", err))},
			IsError: true,
		}, nil
	}

	dependencies := findEntityDependencies(ctx, client, automations, entityID)

	result := entityDependenciesResult{
		EntityID:    entityID,
		Automations: dependencies,
		TotalUsages: len(dependencies),
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{mcp.NewTextContent(fmt.Sprintf("Error formatting result: %v", err))},
			IsError: true,
		}, nil
	}

	summary := fmt.Sprintf("Found %d automations using '%s'", len(dependencies), entityID)

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(summary + "\n\n" + string(output))},
	}, nil
}

// findEntityDependencies searches all automations for entity usage.
func findEntityDependencies(ctx context.Context, client homeassistant.Client, automations []homeassistant.Automation, entityID string) []entityDependency {
	var dependencies []entityDependency

	for _, auto := range automations {
		autoID := strings.TrimPrefix(auto.EntityID, "automation.")
		fullAuto, err := client.GetAutomation(ctx, autoID)
		if err != nil {
			continue
		}

		if dep := checkAutomationForEntity(fullAuto, autoID, entityID); dep != nil {
			dependencies = append(dependencies, *dep)
		}
	}

	return dependencies
}

// checkAutomationForEntity checks if an automation uses the entity.
func checkAutomationForEntity(auto *homeassistant.Automation, autoID, entityID string) *entityDependency {
	usedIn := findEntityUsageLocations(auto.Config, entityID)
	if len(usedIn) == 0 {
		return nil
	}

	alias := auto.FriendlyName
	if auto.Config != nil && auto.Config.Alias != "" {
		alias = auto.Config.Alias
	}

	return &entityDependency{
		AutomationID:    autoID,
		AutomationAlias: alias,
		UsedIn:          usedIn,
	}
}

// findEntityUsageLocations finds where an entity is used in automation config.
func findEntityUsageLocations(config *homeassistant.AutomationConfig, entityID string) []string {
	var usedIn []string

	if config == nil {
		return usedIn
	}

	if searchEntityInConfig(config.Triggers, entityID) {
		usedIn = append(usedIn, "trigger")
	}
	if searchEntityInConfig(config.Conditions, entityID) {
		usedIn = append(usedIn, "condition")
	}
	if searchEntityInConfig(config.Actions, entityID) {
		usedIn = append(usedIn, "action")
	}

	return usedIn
}

// searchEntityInConfig recursively searches for an entity ID in a config structure.
func searchEntityInConfig(config any, entityID string) bool {
	if config == nil {
		return false
	}

	switch v := config.(type) {
	case string:
		return v == entityID
	case []any:
		return searchEntityInSlice(v, entityID)
	case map[string]any:
		return searchEntityInMap(v, entityID)
	}

	return false
}

// searchEntityInSlice searches for an entity ID in a slice.
func searchEntityInSlice(items []any, entityID string) bool {
	for _, item := range items {
		if searchEntityInConfig(item, entityID) {
			return true
		}
	}
	return false
}

// searchEntityInMap searches for an entity ID in a map structure.
func searchEntityInMap(m map[string]any, entityID string) bool {
	for key, val := range m {
		if searchEntityInMapEntry(key, val, entityID) {
			return true
		}
	}
	return false
}

// searchEntityInMapEntry checks a single map entry for entity ID references.
func searchEntityInMapEntry(key string, val any, entityID string) bool {
	// Check direct entity_id field
	if key == configKeyEntityID {
		return searchEntityInConfig(val, entityID)
	}

	// Check nested entity_id in target or data fields
	if key == configKeyTarget || key == "data" {
		if found := searchEntityInNestedMap(val, entityID); found {
			return true
		}
	}

	// Recursively search in nested structures
	return searchEntityInConfig(val, entityID)
}

// searchEntityInNestedMap checks for entity_id in a nested map.
func searchEntityInNestedMap(val any, entityID string) bool {
	nestedMap, ok := val.(map[string]any)
	if !ok {
		return false
	}
	return searchEntityInConfig(nestedMap[configKeyEntityID], entityID)
}
