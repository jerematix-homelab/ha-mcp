// Package handlers provides MCP tool handlers for Home Assistant operations.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/zorak1103/ha-mcp/internal/handlers/formatter"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// ConsolidatedEntityQueryHandlers provides consolidated handlers for entity query operations.
// This replaces get_states, get_history, get_statistics, and list_domains tools.
type ConsolidatedEntityQueryHandlers struct{}

// compactEntityState represents a minimal entity state for compact output.
type compactEntityState struct {
	EntityID     string `json:"entity_id"`
	State        string `json:"state"`
	FriendlyName string `json:"friendly_name,omitempty"`
}

// stateFilterParams holds parsed filter parameters for entity states.
type stateFilterParams struct {
	domain         string
	areaID         string
	stateFilter    string
	stateNotFilter string
	nameContains   string
	format         formatter.Format
	verbose        bool
}

// compactHistoryEntry represents a minimal history entry for compact output.
type compactHistoryEntry struct {
	State       string `json:"state"`
	LastChanged string `json:"last_changed"`
}

// historyParams encapsulates all parsed parameters for history queries.
type historyParams struct {
	entityID    string
	startTime   time.Time
	endTime     time.Time
	stateFilter string
	limit       int
	format      formatter.Format
	verbose     bool
}

// historyResult encapsulates processed history data.
type historyResult struct {
	entries    []homeassistant.HistoryEntry
	totalCount int
}

// paginatedStatesResponse wraps state output with pagination metadata.
type paginatedStatesResponse struct {
	Items      json.RawMessage    `json:"items"`
	Pagination PaginationMetadata `json:"pagination"`
}

// NewConsolidatedEntityQueryHandlers creates a new ConsolidatedEntityQueryHandlers instance.
func NewConsolidatedEntityQueryHandlers() *ConsolidatedEntityQueryHandlers {
	return &ConsolidatedEntityQueryHandlers{}
}

// Helper functions for entity state filtering

// parseStateFilterParams extracts filter parameters from args.
func parseStateFilterParams(args map[string]any) stateFilterParams {
	return stateFilterParams{
		domain:         getStringArg(args, "domain"),
		areaID:         getStringArg(args, "area_id"),
		stateFilter:    getStringArg(args, "state"),
		stateNotFilter: getStringArg(args, "state_not"),
		nameContains:   getStringArg(args, "name_contains"),
		format:         formatter.ParseFormat(getStringArg(args, "format")),
		verbose:        getBoolArg(args, "verbose"),
	}
}

// buildEntityIDsInArea builds a set of entity IDs that belong to an area,
// either directly or via their device.
func buildEntityIDsInArea(ctx context.Context, client homeassistant.Client, areaID string) (map[string]bool, error) {
	entityIDsInArea := make(map[string]bool)

	// Load entity registry
	entityRegistry, err := client.GetEntityRegistry(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load entity registry: %w", err)
	}

	// Load device registry to check device-area mapping
	deviceRegistry, err := client.GetDeviceRegistry(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load device registry: %w", err)
	}

	// Build device ID to area ID mapping
	deviceIDsInArea := make(map[string]bool)
	for _, device := range deviceRegistry {
		if device.AreaID == areaID {
			deviceIDsInArea[device.ID] = true
		}
	}

	// Find entities in area (direct or via device)
	for _, entity := range entityRegistry {
		directMatch := entity.AreaID == areaID
		deviceMatch := entity.DeviceID != "" && deviceIDsInArea[entity.DeviceID]
		if directMatch || deviceMatch {
			entityIDsInArea[entity.EntityID] = true
		}
	}

	return entityIDsInArea, nil
}

// filterStates applies all filters to a list of entities.
// If entityIDsInArea is non-nil, only entities in that set are included.
func filterStates(states []homeassistant.Entity, params stateFilterParams, entityIDsInArea map[string]bool) []homeassistant.Entity {
	nameContainsLower := strings.ToLower(params.nameContains)
	filtered := make([]homeassistant.Entity, 0, len(states))

	for _, state := range states {
		if matchesStateFilters(state, params, nameContainsLower, entityIDsInArea) {
			filtered = append(filtered, state)
		}
	}

	return filtered
}

// matchesStateFilters checks if a single entity matches all filters.
// If entityIDsInArea is non-nil, entity must be in that set.
func matchesStateFilters(state homeassistant.Entity, params stateFilterParams, nameContainsLower string, entityIDsInArea map[string]bool) bool {
	// Check area filter first (if entityIDsInArea is provided)
	if entityIDsInArea != nil && !entityIDsInArea[state.EntityID] {
		return false
	}
	if params.domain != "" && !strings.HasPrefix(state.EntityID, params.domain+".") {
		return false
	}
	if params.stateFilter != "" && state.State != params.stateFilter {
		return false
	}
	if params.stateNotFilter != "" && state.State == params.stateNotFilter {
		return false
	}
	if params.nameContains != "" && !matchesNameFilter(state, nameContainsLower) {
		return false
	}
	return true
}

// matchesNameFilter checks if entity matches the name filter.
func matchesNameFilter(state homeassistant.Entity, nameContainsLower string) bool {
	if strings.Contains(strings.ToLower(state.EntityID), nameContainsLower) {
		return true
	}
	friendlyName, _ := state.Attributes["friendly_name"].(string)
	return strings.Contains(strings.ToLower(friendlyName), nameContainsLower)
}

// formatStatesOutput formats entity states based on verbose flag.
func formatStatesOutput(states []homeassistant.Entity, verbose bool) ([]byte, error) {
	if verbose {
		return json.MarshalIndent(states, "", "  ")
	}
	return json.MarshalIndent(toCompactStates(states), "", "  ")
}

// toCompactStates converts entities to compact format.
func toCompactStates(states []homeassistant.Entity) []compactEntityState {
	compact := make([]compactEntityState, 0, len(states))
	for _, state := range states {
		entry := compactEntityState{
			EntityID: state.EntityID,
			State:    state.State,
		}
		if friendlyName, ok := state.Attributes["friendly_name"].(string); ok {
			entry.FriendlyName = friendlyName
		}
		compact = append(compact, entry)
	}
	return compact
}

// buildStateFiltersMap creates a map of filter values for pagination hash.
func buildStateFiltersMap(params stateFilterParams) map[string]any {
	filters := make(map[string]any)
	if params.domain != "" {
		filters["domain"] = params.domain
	}
	if params.areaID != "" {
		filters["area_id"] = params.areaID
	}
	if params.stateFilter != "" {
		filters["state"] = params.stateFilter
	}
	if params.stateNotFilter != "" {
		filters["state_not"] = params.stateNotFilter
	}
	if params.nameContains != "" {
		filters["name_contains"] = params.nameContains
	}
	return filters
}

// buildPaginatedStatesResponse creates the final response JSON.
func buildPaginatedStatesResponse(paginated PaginatedResponse[homeassistant.Entity], itemsOutput []byte) []byte {
	// If no pagination was applied (limit=0), return items directly for backwards compatibility
	if paginated.Pagination.Limit == 0 {
		return itemsOutput
	}

	response := paginatedStatesResponse{
		Items:      itemsOutput,
		Pagination: paginated.Pagination,
	}
	result, _ := json.MarshalIndent(response, "", "  ")
	return result
}

// Helper functions for history queries

// parseHistoryParams extracts and validates all parameters from args.
func parseHistoryParams(args map[string]any) (*historyParams, error) {
	entityID, ok := args["entity_id"].(string)
	if !ok || entityID == "" {
		return nil, fmt.Errorf("entity_id is required")
	}

	startTime, endTime, err := parseTimeRange(args)
	if err != nil {
		return nil, err
	}

	stateFilter, _ := args["state"].(string)

	limit := 0
	if limitVal, ok := args["limit"].(float64); ok && limitVal > 0 {
		limit = int(limitVal)
	}

	formatStr, _ := args["format"].(string)
	verbose, _ := args["verbose"].(bool)

	return &historyParams{
		entityID:    entityID,
		startTime:   startTime,
		endTime:     endTime,
		stateFilter: stateFilter,
		limit:       limit,
		format:      formatter.ParseFormat(formatStr),
		verbose:     verbose,
	}, nil
}

// parseTimeRange parses start_time, end_time, and hours parameters.
func parseTimeRange(args map[string]any) (start, end time.Time, err error) {
	end = time.Now()
	start = end.Add(-24 * time.Hour)

	// 'hours' parameter takes precedence over 'start_time'
	if hours, ok := args["hours"].(float64); ok && hours > 0 {
		start = end.Add(-time.Duration(hours) * time.Hour)
	} else if startStr, ok := args["start_time"].(string); ok && startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid start_time format: %w", err)
		}
	}

	if endStr, ok := args["end_time"].(string); ok && endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end_time format: %w", err)
		}
	}

	return start, end, nil
}

// processHistoryEntries flattens, filters, and limits history entries.
func processHistoryEntries(
	history [][]homeassistant.HistoryEntry,
	stateFilter string,
	limit int,
) historyResult {
	// Flatten history (it's [][]HistoryEntry, typically with one inner array per entity)
	var entries []homeassistant.HistoryEntry
	for _, entityHistory := range history {
		entries = append(entries, entityHistory...)
	}

	// Apply state filter
	if stateFilter != "" {
		filtered := make([]homeassistant.HistoryEntry, 0, len(entries))
		for _, entry := range entries {
			if entry.State == stateFilter {
				filtered = append(filtered, entry)
			}
		}
		entries = filtered
	}

	totalCount := len(entries)

	// Apply limit (take most recent entries)
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}

	return historyResult{
		entries:    entries,
		totalCount: totalCount,
	}
}

// formatHistoryOutput formats entries based on verbose flag.
func formatHistoryOutput(entries []homeassistant.HistoryEntry, verbose bool) ([]byte, error) {
	if verbose {
		return json.MarshalIndent(entries, "", "  ")
	}

	compact := make([]compactHistoryEntry, 0, len(entries))
	for _, entry := range entries {
		compact = append(compact, compactHistoryEntry{
			State:       entry.State,
			LastChanged: entry.LastChangedTime().Format(time.RFC3339),
		})
	}

	return json.MarshalIndent(compact, "", "  ")
}

// buildHistorySummary creates the summary message for history results.
func buildHistorySummary(entityID string, result historyResult, stateFilter string, verbose bool) string {
	var summary string

	if result.totalCount > len(result.entries) {
		summary = fmt.Sprintf("Showing %d of %d history entries for %s (limited)",
			len(result.entries), result.totalCount, entityID)
	} else {
		summary = fmt.Sprintf("Found %d history entries for %s", len(result.entries), entityID)
	}

	if stateFilter != "" {
		summary += fmt.Sprintf(" (filtered by state='%s')", stateFilter)
	}

	if !verbose {
		summary += VerboseHint
	}

	return summary
}

// RegisterTools registers the consolidated query_entities tool.
func (h *ConsolidatedEntityQueryHandlers) RegisterTools(registry *mcp.Registry) {
	registry.RegisterTool(h.queryEntitiesTool(), h.handleQueryEntities)
}

// queryEntitiesTool returns the tool definition for the consolidated entity query tool.
func (h *ConsolidatedEntityQueryHandlers) queryEntitiesTool() mcp.Tool {
	return mcp.Tool{
		Name:        "query_entities",
		Description: queryEntitiesDescription(),
		InputSchema: mcp.JSONSchema{
			Type:       "object",
			Properties: queryEntitiesProperties(),
			Required:   []string{"mode"},
		},
	}
}

func queryEntitiesDescription() string {
	return `Query Home Assistant entities (states, history, statistics, domains).

Modes:
- current: Get current entity states with optional filters (domain, state, name_contains, pagination)
- history: Get historical state changes for an entity (requires entity_id)
- statistics: Get long-term statistics for entities (requires statistic_ids array)
- domains: List all available entity domains with counts

Examples:
- Get all lights: {"mode": "current", "domain": "light"}
- Get history: {"mode": "history", "entity_id": "sensor.temperature", "hours": 6}
- Get statistics: {"mode": "statistics", "statistic_ids": ["sensor.energy"]}
- List domains: {"mode": "domains"}`
}

//nolint:funlen // Schema definition naturally long
func queryEntitiesProperties() map[string]mcp.JSONSchema {
	return map[string]mcp.JSONSchema{
		"mode": {
			Type:        "string",
			Enum:        []string{"current", "history", "statistics", "domains"},
			Description: "Query mode: current (states), history, statistics, or domains",
		},
		"format": {
			Type:        "string",
			Enum:        []string{"natural", "json"},
			Description: "Output format: 'natural' for LLM-optimized human-readable output (default), 'json' for structured JSON",
		},
		// Current mode parameters
		"domain": {
			Type:        "string",
			Description: "Filter entities by domain (e.g., 'light', 'switch'). Only for mode=current",
		},
		"state": {
			Type:        "string",
			Description: "Filter by state value (e.g., 'on', 'off'). Only for mode=current",
		},
		"state_not": {
			Type:        "string",
			Description: "Exclude entities with this state. Only for mode=current",
		},
		"name_contains": {
			Type:        "string",
			Description: "Filter by entity_id or friendly_name containing string. Only for mode=current",
		},
		"area_id": {
			Type:        "string",
			Description: "Filter entities by area (matches entities directly in area or via device). Only for mode=current",
		},
		// History mode parameters
		"entity_id": {
			Type:        "string",
			Description: "Entity ID to query. Required for mode=history",
		},
		"start_time": {
			Type:        "string",
			Description: "Start time in RFC3339 format. Only for mode=history",
		},
		"end_time": {
			Type:        "string",
			Description: "End time in RFC3339 format. Only for mode=history",
		},
		"hours": {
			Type:        "number",
			Description: "Number of hours to look back. Only for mode=history",
		},
		// Statistics mode parameters
		"statistic_ids": {
			Type:        "array",
			Description: "Array of statistic IDs to retrieve. Required for mode=statistics",
		},
		"period": {
			Type:        "string",
			Description: "Statistics period: 5minute, hour, day, week, month. Only for mode=statistics",
		},
		// Common parameters
		"verbose": {
			Type:        "boolean",
			Description: "If true, return full details. Default: false (compact output)",
		},
		"limit": {
			Type:        "integer",
			Description: "Maximum number of results to return. Use with 'cursor' for pagination",
		},
		"cursor": {
			Type:        "string",
			Description: "Pagination cursor from previous response",
		},
	}
}

// handleQueryEntities handles the consolidated query_entities tool.
func (h *ConsolidatedEntityQueryHandlers) handleQueryEntities(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	mode, ok := args["mode"].(string)
	if !ok || mode == "" {
		return errorResult("mode parameter is required"), nil
	}

	switch mode {
	case "current":
		return h.handleCurrent(ctx, client, args)
	case "history":
		return h.handleHistory(ctx, client, args)
	case "statistics":
		return h.handleStatistics(ctx, client, args)
	case "domains":
		return h.handleDomains(ctx, client, args)
	default:
		return errorResult(fmt.Sprintf("Invalid mode %q. Must be one of: current, history, statistics, domains", mode)), nil
	}
}

// handleCurrent handles mode=current requests (adapted from handleGetStates).
func (h *ConsolidatedEntityQueryHandlers) handleCurrent(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	states, err := client.GetStates(ctx)
	if err != nil {
		return errorResult(fmt.Sprintf("Error getting states: %v", err)), nil
	}

	filterParams := parseStateFilterParams(args)

	// Load area filter if specified
	var entityIDsInArea map[string]bool
	if filterParams.areaID != "" {
		entityIDsInArea, err = buildEntityIDsInArea(ctx, client, filterParams.areaID)
		if err != nil {
			return errorResult(fmt.Sprintf("Error loading area filter: %v", err)), nil
		}
	}

	states = filterStates(states, filterParams, entityIDsInArea)

	// Sort by entity_id for stable pagination
	sort.Slice(states, func(i, j int) bool {
		return states[i].EntityID < states[j].EntityID
	})

	// Build filters map for pagination hash
	filtersMap := buildStateFiltersMap(filterParams)

	// Parse pagination params
	paginationParams, err := ParsePaginationParams(args, filtersMap)
	if err != nil {
		return errorResult(fmt.Sprintf("Error: %v", err)), nil
	}

	// Apply pagination
	paginated := ApplyPagination(states, paginationParams)

	// Use formatter based on format parameter
	if filterParams.format == formatter.FormatNatural {
		return h.formatStatesNatural(ctx, paginated, filterParams)
	}

	// JSON format (backward compatible)
	output, err := formatStatesOutput(paginated.Items, filterParams.verbose)
	if err != nil {
		return errorResult(fmt.Sprintf("Error formatting states: %v", err)), nil
	}

	summary := BuildPaginationSummary(paginated.Pagination, "entities")
	if !filterParams.verbose {
		summary += VerboseHint
	}

	// Build response with pagination metadata
	response := buildPaginatedStatesResponse(paginated, output)

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(summary + "\n\n" + string(response))},
	}, nil
}

// formatStatesNatural formats states using natural language formatter.
func (h *ConsolidatedEntityQueryHandlers) formatStatesNatural(
	ctx context.Context,
	paginated PaginatedResponse[homeassistant.Entity],
	params stateFilterParams,
) (*mcp.ToolsCallResult, error) {
	f := formatter.NewNaturalFormatter()

	opts := formatter.EntityListOptions{
		Verbose:        params.verbose,
		IncludeSummary: true,
		GroupByDomain:  !params.verbose, // Group by domain in compact mode
	}

	output, err := f.FormatEntities(ctx, paginated.Items, opts)
	if err != nil {
		return errorResult(fmt.Sprintf("Error formatting states: %v", err)), nil
	}

	// Add pagination info if paginated
	if paginated.Pagination.HasMore && paginated.Pagination.NextCursor != nil {
		output += fmt.Sprintf("\n\n[Page %d of results. Use cursor='%s' for next page]",
			(paginated.Pagination.Offset/paginated.Pagination.Limit)+1,
			*paginated.Pagination.NextCursor)
	}

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(output)},
	}, nil
}

// handleHistory handles mode=history requests (adapted from handleGetHistory).
func (h *ConsolidatedEntityQueryHandlers) handleHistory(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	params, err := parseHistoryParams(args)
	if err != nil {
		return errorResult(err.Error()), nil
	}

	history, err := client.GetHistory(ctx, params.entityID, params.startTime, params.endTime)
	if err != nil {
		return errorResult(fmt.Sprintf("Error getting history: %v", err)), nil
	}

	result := processHistoryEntries(history, params.stateFilter, params.limit)

	// Use formatter based on format parameter
	if params.format == formatter.FormatNatural {
		return h.formatHistoryNatural(ctx, params.entityID, result, params)
	}

	// JSON format (backward compatible)
	output, err := formatHistoryOutput(result.entries, params.verbose)
	if err != nil {
		return errorResult(fmt.Sprintf("Error formatting history: %v", err)), nil
	}

	summary := buildHistorySummary(params.entityID, result, params.stateFilter, params.verbose)

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(summary + "\n\n" + string(output))},
	}, nil
}

// formatHistoryNatural formats history using natural language formatter.
func (h *ConsolidatedEntityQueryHandlers) formatHistoryNatural(
	ctx context.Context,
	entityID string,
	result historyResult,
	params *historyParams,
) (*mcp.ToolsCallResult, error) {
	f := formatter.NewNaturalFormatter()

	opts := formatter.HistoryOptions{
		Verbose: params.verbose,
		Limit:   params.limit,
	}
	if opts.Limit == 0 {
		opts.Limit = 10 // Default limit for verbose natural language output
	}

	output, err := f.FormatHistory(ctx, entityID, result.entries, opts)
	if err != nil {
		return errorResult(fmt.Sprintf("Error formatting history: %v", err)), nil
	}

	// Add filter info if filtered
	if params.stateFilter != "" {
		output += fmt.Sprintf(" (filtered by state='%s')", params.stateFilter)
	}

	// Add truncation notice if applicable
	if result.totalCount > len(result.entries) {
		output += fmt.Sprintf("\n\n[Showing %d of %d entries. Use limit parameter for more.]",
			len(result.entries), result.totalCount)
	}

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(output)},
	}, nil
}

// handleStatistics handles mode=statistics requests (adapted from statistics.go).
func (h *ConsolidatedEntityQueryHandlers) handleStatistics(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	statIDs, err := parseStatisticIDs(args)
	if err != nil {
		return errorResult(err.Error()), nil
	}

	period := "hour"
	if p, ok := args["period"].(string); ok && p != "" {
		period = p
	}

	statistics, err := client.GetStatistics(ctx, statIDs, period)
	if err != nil {
		return errorResult(fmt.Sprintf("Error getting statistics: %v", err)), nil
	}

	filtersMap := buildStatisticsFiltersMap(statIDs, period)
	paginationParams, err := ParsePaginationParams(args, filtersMap)
	if err != nil {
		return errorResult(fmt.Sprintf("Error: %v", err)), nil
	}

	paginated := ApplyPagination(statistics, paginationParams)

	// Parse format parameter
	formatStr, _ := args["format"].(string)
	format := formatter.ParseFormat(formatStr)

	// Use natural format
	if format == formatter.FormatNatural {
		return h.formatStatisticsNatural(paginated, paginationParams)
	}

	// JSON format (backward compatible)
	result, err := json.MarshalIndent(paginated.Items, "", "  ")
	if err != nil {
		return errorResult(fmt.Sprintf("Error formatting statistics: %v", err)), nil
	}

	summary := BuildPaginationSummary(paginated.Pagination, "statistics results")
	response := buildPaginatedStatisticsResponse(paginated, result)

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(summary + "\n\n" + string(response))},
	}, nil
}

// formatStatisticsNatural formats statistics using natural language.
func (h *ConsolidatedEntityQueryHandlers) formatStatisticsNatural(
	paginated PaginatedResponse[homeassistant.StatisticsResult],
	paginationParams PaginationParams,
) (*mcp.ToolsCallResult, error) {
	var output strings.Builder

	// Group by statistic_id
	groupedStats := make(map[string][]homeassistant.StatisticsResult)
	for _, stat := range paginated.Items {
		groupedStats[stat.StatisticID] = append(groupedStats[stat.StatisticID], stat)
	}

	// Sort statistic IDs for consistent output
	var statIDs []string
	for id := range groupedStats {
		statIDs = append(statIDs, id)
	}
	sort.Strings(statIDs)

	fmt.Fprintf(&output, "Statistics for %d entities:\n\n", len(statIDs))

	for _, statID := range statIDs {
		stats := groupedStats[statID]
		fmt.Fprintf(&output, "## %s\n", statID)
		fmt.Fprintf(&output, "%d data points\n\n", len(stats))

		// Show recent entries
		showCount := min(5, len(stats))
		for i := len(stats) - showCount; i < len(stats); i++ {
			stat := stats[i]
			timestamp := stat.StartTime().Format(time.RFC3339)
			values := formatStatValues(stat)
			if values != "" {
				fmt.Fprintf(&output, "- %s: %s\n", timestamp, values)
			}
		}
		output.WriteString("\n")
	}

	// Add pagination info if paginated
	if paginated.Pagination.HasMore && paginated.Pagination.NextCursor != nil {
		fmt.Fprintf(&output, "\n[Page %d of results. Use cursor='%s' for next page]",
			(paginated.Pagination.Offset/paginationParams.Limit)+1,
			*paginated.Pagination.NextCursor)
	}

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(output.String())},
	}, nil
}

// formatStatValues formats individual stat fields.
func formatStatValues(stat homeassistant.StatisticsResult) string {
	var values []string

	if stat.Mean != nil {
		values = append(values, fmt.Sprintf("mean=%.2f", *stat.Mean))
	}
	if stat.Min != nil {
		values = append(values, fmt.Sprintf("min=%.2f", *stat.Min))
	}
	if stat.Max != nil {
		values = append(values, fmt.Sprintf("max=%.2f", *stat.Max))
	}
	if stat.Sum != nil {
		values = append(values, fmt.Sprintf("sum=%.2f", *stat.Sum))
	}
	if stat.State != nil {
		values = append(values, fmt.Sprintf("state=%.2f", *stat.State))
	}
	if stat.Change != nil {
		values = append(values, fmt.Sprintf("change=%.2f", *stat.Change))
	}

	return strings.Join(values, ", ")
}

// handleDomains handles mode=domains requests (adapted from handleListDomains).
func (h *ConsolidatedEntityQueryHandlers) handleDomains(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	states, err := client.GetStates(ctx)
	if err != nil {
		return errorResult(fmt.Sprintf("Error getting states: %v", err)), nil
	}

	// Extract unique domains using index lookup for efficiency
	domainSet := make(map[string]int)
	for _, state := range states {
		if idx := strings.Index(state.EntityID, "."); idx > 0 {
			domainSet[state.EntityID[:idx]]++
		}
	}

	// Parse format parameter
	formatStr, _ := args["format"].(string)
	format := formatter.ParseFormat(formatStr)

	// Natural format
	if format == formatter.FormatNatural {
		return h.formatDomainsNatural(domainSet, len(states)), nil
	}

	// JSON format (backward compatible)
	type domainInfo struct {
		Domain string `json:"domain"`
		Count  int    `json:"entity_count"`
	}

	domains := make([]domainInfo, 0, len(domainSet))
	for domain, count := range domainSet {
		domains = append(domains, domainInfo{
			Domain: domain,
			Count:  count,
		})
	}

	// Sort by domain name
	sort.Slice(domains, func(i, j int) bool {
		return domains[i].Domain < domains[j].Domain
	})

	output, err := json.MarshalIndent(domains, "", "  ")
	if err != nil {
		return errorResult(fmt.Sprintf("Error formatting domains: %v", err)), nil
	}

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(string(output))},
	}, nil
}

// formatDomainsNatural formats domains using natural language.
func (h *ConsolidatedEntityQueryHandlers) formatDomainsNatural(
	domainSet map[string]int,
	totalEntities int,
) *mcp.ToolsCallResult {
	var output strings.Builder

	fmt.Fprintf(&output, "%d domains, %d entities total\n\n", len(domainSet), totalEntities)

	// Sort domains alphabetically
	var domains []string
	for domain := range domainSet {
		domains = append(domains, domain)
	}
	sort.Strings(domains)

	for _, domain := range domains {
		count := domainSet[domain]
		fmt.Fprintf(&output, "- %s: %d entities\n", domain, count)
	}

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(output.String())},
	}
}

// parseStatisticIDs extracts and validates statistic_ids from args.
func parseStatisticIDs(args map[string]any) ([]string, error) {
	statIDsRaw, ok := args["statistic_ids"]
	if !ok {
		return nil, fmt.Errorf("statistic_ids is required")
	}

	statIDsSlice, ok := statIDsRaw.([]any)
	if !ok {
		return nil, fmt.Errorf("statistic_ids must be an array")
	}

	statIDs := make([]string, 0, len(statIDsSlice))
	for _, id := range statIDsSlice {
		if s, ok := id.(string); ok {
			statIDs = append(statIDs, s)
		}
	}

	if len(statIDs) == 0 {
		return nil, fmt.Errorf("at least one statistic_id is required")
	}

	return statIDs, nil
}

// buildStatisticsFiltersMap creates a map of filter values for pagination hash.
func buildStatisticsFiltersMap(statIDs []string, period string) map[string]any {
	filters := make(map[string]any)
	filters["statistic_ids"] = statIDs
	filters["period"] = period
	return filters
}

// paginatedStatisticsResponse wraps statistics output with pagination metadata.
type paginatedStatisticsResponse struct {
	Items      json.RawMessage    `json:"items"`
	Pagination PaginationMetadata `json:"pagination"`
}

// buildPaginatedStatisticsResponse creates the final response JSON.
func buildPaginatedStatisticsResponse(paginated PaginatedResponse[homeassistant.StatisticsResult], itemsOutput []byte) []byte {
	// If no pagination was applied (limit=0), return items directly for backwards compatibility
	if paginated.Pagination.Limit == 0 {
		return itemsOutput
	}

	response := paginatedStatisticsResponse{
		Items:      itemsOutput,
		Pagination: paginated.Pagination,
	}
	result, _ := json.MarshalIndent(response, "", "  ")
	return result
}

// RegisterConsolidatedEntityQueryTools registers the consolidated query_entities tool.
func RegisterConsolidatedEntityQueryTools(registry *mcp.Registry) {
	h := NewConsolidatedEntityQueryHandlers()
	h.RegisterTools(registry)
}
