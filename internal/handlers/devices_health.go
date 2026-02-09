package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// DeviceHealthReport represents the analysis results for device health.
type DeviceHealthReport struct {
	Issues     []DeviceHealthIssue    `json:"issues"`
	Statistics DeviceHealthStatistics `json:"statistics"`
}

// DeviceHealthIssue represents a single health issue detected for a device.
type DeviceHealthIssue struct {
	DeviceID     string `json:"device_id"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Details      string `json:"details"`
	Manufacturer string `json:"manufacturer,omitempty"`
}

// DeviceHealthStatistics provides aggregate counts from the health analysis.
type DeviceHealthStatistics struct {
	TotalDevices       int            `json:"total_devices"`
	HealthyDevices     int            `json:"healthy_devices"`
	ProblematicDevices int            `json:"problematic_devices"`
	ByCategory         map[string]int `json:"by_category"`
}

// DeviceRemoveResult represents the outcome of removing devices.
type DeviceRemoveResult struct {
	Successes []DeviceRemoveSuccess `json:"successes"`
	Failures  []DeviceRemoveFailure `json:"failures,omitempty"`
}

// DeviceRemoveSuccess represents a successful device removal.
type DeviceRemoveSuccess struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
}

// DeviceRemoveFailure represents a failure during device removal.
type DeviceRemoveFailure struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
	Error    string `json:"error"`
}

// handleDeviceHealth handles health mode analysis and cleanup operations for devices.
func (h *DeviceQueryHandlers) handleDeviceHealth(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	// Parse action parameter (default: analyze)
	action, _ := args["action"].(string)
	if action == "" {
		action = deviceHealthActionAnalyze
	}

	switch action {
	case deviceHealthActionAnalyze:
		return h.handleDeviceHealthAnalyze(ctx, client, args)
	case deviceHealthActionRemove:
		return h.handleDeviceHealthRemove(ctx, client, args)
	default:
		return errorResult(fmt.Sprintf("action must be one of: %s, %s", deviceHealthActionAnalyze, deviceHealthActionRemove)), nil
	}
}

func (h *DeviceQueryHandlers) handleDeviceHealthAnalyze(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	// Create snapshot for parallel fetching
	snapshot := CreateAnalysisSnapshot(ctx, client)

	// Fetch config entries (needed for orphaned_config_entry and config_entry_error detection)
	configEntries, _ := client.GetConfigEntries(ctx, "")
	configEntryMap := buildConfigEntryMap(configEntries)

	// Build entity-to-device map (for no_entities detection)
	entityDeviceMap := buildEntityDeviceMap(snapshot.EntityRegistry)

	// Parse filters
	categoryFilter := parseDeviceCategoriesFilter(args)
	manufacturerFilter, _ := args["manufacturer"].(string)

	// Detect issues
	var allIssues []DeviceHealthIssue
	for _, device := range snapshot.DeviceRegistry {
		// Apply manufacturer filter
		if manufacturerFilter != "" && device.Manufacturer != manufacturerFilter {
			continue
		}

		// Detect issues for this device
		issues := detectDeviceIssues(device, configEntryMap, entityDeviceMap, categoryFilter)
		allIssues = append(allIssues, issues...)
	}

	// Build report
	report := DeviceHealthReport{
		Issues: allIssues,
		Statistics: DeviceHealthStatistics{
			TotalDevices:       len(snapshot.DeviceRegistry),
			ProblematicDevices: countUniqueDevices(allIssues),
			ByCategory:         countByCategory(allIssues),
		},
	}
	report.Statistics.HealthyDevices = report.Statistics.TotalDevices - report.Statistics.ProblematicDevices

	// Format output
	format, _ := args["format"].(string)
	if format == "json" {
		return formatDeviceHealthReportJSON(report)
	}
	return formatDeviceHealthReportNatural(report)
}

func (h *DeviceQueryHandlers) handleDeviceHealthRemove(
	ctx context.Context,
	client homeassistant.Client,
	args map[string]any,
) (*mcp.ToolsCallResult, error) {
	// Parse device_ids parameter
	deviceIDs, err := parseDeviceIDsParam(args)
	if err != nil {
		return errorResult(err.Error()), nil
	}

	// Fetch device and config entry registries
	devices, err := client.GetDeviceRegistry(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get device registry: %w", err)
	}

	configEntries, err := client.GetConfigEntries(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get config entries: %w", err)
	}

	// Build maps
	deviceMap := buildDeviceMap(devices)
	configEntryMap := buildConfigEntryMap(configEntries)

	// Remove devices and collect results
	result := removeDevicesWithResults(ctx, client, deviceIDs, deviceMap, configEntryMap)

	// Format output
	format, _ := args["format"].(string)
	if format == "json" {
		return formatDeviceRemoveResultJSON(result)
	}
	return formatDeviceRemoveResultNatural(result)
}

func parseDeviceIDsParam(args map[string]any) ([]string, error) {
	deviceIDsRaw, ok := args["device_ids"]
	if !ok {
		return nil, fmt.Errorf("device_ids parameter is required for action=remove")
	}

	deviceIDsAny, ok := deviceIDsRaw.([]any)
	if !ok {
		return nil, fmt.Errorf("device_ids must be an array")
	}

	var deviceIDs []string
	for _, idAny := range deviceIDsAny {
		if idStr, ok := idAny.(string); ok {
			deviceIDs = append(deviceIDs, idStr)
		}
	}

	if len(deviceIDs) == 0 {
		return nil, fmt.Errorf("device_ids array is empty")
	}

	return deviceIDs, nil
}

func removeDevicesWithResults(
	ctx context.Context,
	client homeassistant.Client,
	deviceIDs []string,
	deviceMap map[string]homeassistant.DeviceRegistryEntry,
	configEntryMap map[string]homeassistant.ConfigEntryFull,
) DeviceRemoveResult {
	var successes []DeviceRemoveSuccess
	var failures []DeviceRemoveFailure

	for _, deviceID := range deviceIDs {
		device, found := deviceMap[deviceID]
		if !found {
			failures = append(failures, DeviceRemoveFailure{
				DeviceID: deviceID,
				Error:    "device not found",
			})
			continue
		}

		success, errMsg := removeDeviceConfigEntries(ctx, client, device, configEntryMap)
		if success {
			successes = append(successes, DeviceRemoveSuccess{
				DeviceID: deviceID,
				Name:     device.Name,
			})
		} else {
			failures = append(failures, DeviceRemoveFailure{
				DeviceID: deviceID,
				Name:     device.Name,
				Error:    errMsg,
			})
		}
	}

	return DeviceRemoveResult{
		Successes: successes,
		Failures:  failures,
	}
}

func removeDeviceConfigEntries(
	ctx context.Context,
	client homeassistant.Client,
	device homeassistant.DeviceRegistryEntry,
	configEntryMap map[string]homeassistant.ConfigEntryFull,
) (bool, string) {
	for _, ceID := range device.ConfigEntries {
		ce, ceFound := configEntryMap[ceID]
		if !ceFound {
			continue // Config entry already gone
		}

		if !ce.SupportsRemoveDevice {
			return false, fmt.Sprintf("integration %s does not support device removal", ce.Domain)
		}

		if err := client.RemoveDeviceConfigEntry(ctx, device.ID, ceID); err != nil {
			return false, err.Error()
		}
	}
	return true, ""
}

// detectDeviceIssues detects all issues for a single device.
func detectDeviceIssues(
	device homeassistant.DeviceRegistryEntry,
	configEntryMap map[string]homeassistant.ConfigEntryFull,
	entityDeviceMap map[string][]string,
	categoryFilter map[string]bool,
) []DeviceHealthIssue {
	var issues []DeviceHealthIssue

	issues = append(issues, detectDisabledDevice(device, categoryFilter)...)
	issues = append(issues, detectOrphanedConfigEntries(device, configEntryMap, categoryFilter)...)
	issues = append(issues, detectConfigEntryErrors(device, configEntryMap, categoryFilter)...)
	issues = append(issues, detectNoEntities(device, entityDeviceMap, categoryFilter)...)
	issues = append(issues, detectNoConfigEntries(device, categoryFilter)...)

	return issues
}

func detectDisabledDevice(device homeassistant.DeviceRegistryEntry, filter map[string]bool) []DeviceHealthIssue {
	if !shouldDetect(deviceCategoryDisabled, filter) || device.DisabledBy == "" {
		return nil
	}
	return []DeviceHealthIssue{{
		DeviceID:     device.ID,
		Name:         device.Name,
		Category:     deviceCategoryDisabled,
		Details:      fmt.Sprintf("disabled by %s", device.DisabledBy),
		Manufacturer: device.Manufacturer,
	}}
}

func detectOrphanedConfigEntries(
	device homeassistant.DeviceRegistryEntry,
	configEntryMap map[string]homeassistant.ConfigEntryFull,
	filter map[string]bool,
) []DeviceHealthIssue {
	if !shouldDetect(deviceCategoryOrphanedConfigEntry, filter) {
		return nil
	}

	var issues []DeviceHealthIssue
	for _, ceID := range device.ConfigEntries {
		if _, found := configEntryMap[ceID]; !found {
			issues = append(issues, DeviceHealthIssue{
				DeviceID:     device.ID,
				Name:         device.Name,
				Category:     deviceCategoryOrphanedConfigEntry,
				Details:      fmt.Sprintf("references non-existent config entry %s", ceID),
				Manufacturer: device.Manufacturer,
			})
		}
	}
	return issues
}

func detectConfigEntryErrors(
	device homeassistant.DeviceRegistryEntry,
	configEntryMap map[string]homeassistant.ConfigEntryFull,
	filter map[string]bool,
) []DeviceHealthIssue {
	if !shouldDetect(deviceCategoryConfigEntryError, filter) {
		return nil
	}

	var issues []DeviceHealthIssue
	for _, ceID := range device.ConfigEntries {
		if ce, found := configEntryMap[ceID]; found && ce.State != "loaded" {
			issues = append(issues, DeviceHealthIssue{
				DeviceID:     device.ID,
				Name:         device.Name,
				Category:     deviceCategoryConfigEntryError,
				Details:      fmt.Sprintf("config entry %s has state %s", ceID, ce.State),
				Manufacturer: device.Manufacturer,
			})
		}
	}
	return issues
}

func detectNoEntities(
	device homeassistant.DeviceRegistryEntry,
	entityDeviceMap map[string][]string,
	filter map[string]bool,
) []DeviceHealthIssue {
	if !shouldDetect(deviceCategoryNoEntities, filter) || len(entityDeviceMap[device.ID]) > 0 {
		return nil
	}
	return []DeviceHealthIssue{{
		DeviceID:     device.ID,
		Name:         device.Name,
		Category:     deviceCategoryNoEntities,
		Details:      "no entities reference this device",
		Manufacturer: device.Manufacturer,
	}}
}

func detectNoConfigEntries(device homeassistant.DeviceRegistryEntry, filter map[string]bool) []DeviceHealthIssue {
	if !shouldDetect(deviceCategoryNoConfigEntries, filter) || len(device.ConfigEntries) > 0 {
		return nil
	}
	return []DeviceHealthIssue{{
		DeviceID:     device.ID,
		Name:         device.Name,
		Category:     deviceCategoryNoConfigEntries,
		Details:      "device has no config entries",
		Manufacturer: device.Manufacturer,
	}}
}

// Helper functions

func parseDeviceCategoriesFilter(args map[string]any) map[string]bool {
	categoriesRaw, ok := args["categories"]
	if !ok {
		return nil
	}

	categoriesAny, ok := categoriesRaw.([]any)
	if !ok {
		return nil
	}

	categorySet := make(map[string]bool, len(categoriesAny))
	for _, catAny := range categoriesAny {
		if catStr, ok := catAny.(string); ok {
			categorySet[catStr] = true
		}
	}

	if len(categorySet) == 0 {
		return nil
	}
	return categorySet
}

func buildDeviceMap(devices []homeassistant.DeviceRegistryEntry) map[string]homeassistant.DeviceRegistryEntry {
	m := make(map[string]homeassistant.DeviceRegistryEntry, len(devices))
	for _, device := range devices {
		m[device.ID] = device
	}
	return m
}

func buildEntityDeviceMap(entities []homeassistant.EntityRegistryEntry) map[string][]string {
	m := make(map[string][]string)
	for _, entity := range entities {
		if entity.DeviceID != "" {
			m[entity.DeviceID] = append(m[entity.DeviceID], entity.EntityID)
		}
	}
	return m
}

func shouldDetect(category string, filter map[string]bool) bool {
	if filter == nil {
		return true // No filter = detect all
	}
	return filter[category]
}

func countUniqueDevices(issues []DeviceHealthIssue) int {
	seen := make(map[string]bool)
	for _, issue := range issues {
		seen[issue.DeviceID] = true
	}
	return len(seen)
}

func countByCategory(issues []DeviceHealthIssue) map[string]int {
	counts := make(map[string]int)
	for _, issue := range issues {
		counts[issue.Category]++
	}
	return counts
}

// Formatting functions

func formatDeviceHealthReportJSON(report DeviceHealthReport) (*mcp.ToolsCallResult, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report: %w", err)
	}

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(string(data))},
	}, nil
}

func formatDeviceHealthReportNatural(report DeviceHealthReport) (*mcp.ToolsCallResult, error) {
	var sb strings.Builder

	sb.WriteString("# Device Health Report\n\n")

	// Summary
	sb.WriteString("## Summary\n")
	sb.WriteString(fmt.Sprintf("- Total devices: %d\n", report.Statistics.TotalDevices))
	sb.WriteString(fmt.Sprintf("- Healthy devices: %d\n", report.Statistics.HealthyDevices))
	sb.WriteString(fmt.Sprintf("- Problematic devices: %d\n\n", report.Statistics.ProblematicDevices))

	if len(report.Issues) == 0 {
		sb.WriteString("No issues detected. All devices are healthy.\n")
		return &mcp.ToolsCallResult{
			Content: []mcp.ContentBlock{mcp.NewTextContent(sb.String())},
		}, nil
	}

	// Issues by category
	sb.WriteString("## Issues\n\n")

	// Group issues by category
	byCategory := make(map[string][]DeviceHealthIssue)
	for _, issue := range report.Issues {
		byCategory[issue.Category] = append(byCategory[issue.Category], issue)
	}

	// Display each category
	categoryNames := map[string]string{
		deviceCategoryDisabled:            "Disabled Devices",
		deviceCategoryOrphanedConfigEntry: "Orphaned Config Entries",
		deviceCategoryConfigEntryError:    "Config Entry Errors",
		deviceCategoryNoEntities:          "No Entities",
		deviceCategoryNoConfigEntries:     "No Config Entries",
	}

	for _, cat := range []string{
		deviceCategoryDisabled,
		deviceCategoryOrphanedConfigEntry,
		deviceCategoryConfigEntryError,
		deviceCategoryNoEntities,
		deviceCategoryNoConfigEntries,
	} {
		issues := byCategory[cat]
		if len(issues) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("### %s (%d)\n", categoryNames[cat], len(issues)))
		for _, issue := range issues {
			sb.WriteString(fmt.Sprintf("- **%s** (`%s`)", issue.Name, issue.DeviceID))
			if issue.Manufacturer != "" {
				sb.WriteString(fmt.Sprintf(" [%s]", issue.Manufacturer))
			}
			if issue.Details != "" {
				sb.WriteString(fmt.Sprintf(": %s", issue.Details))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(sb.String())},
	}, nil
}

func formatDeviceRemoveResultJSON(result DeviceRemoveResult) (*mcp.ToolsCallResult, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(string(data))},
	}, nil
}

func formatDeviceRemoveResultNatural(result DeviceRemoveResult) (*mcp.ToolsCallResult, error) {
	var sb strings.Builder

	sb.WriteString("# Device Removal Result\n\n")

	if len(result.Successes) > 0 {
		sb.WriteString(fmt.Sprintf("## Successfully Removed (%d)\n", len(result.Successes)))
		for _, success := range result.Successes {
			sb.WriteString(fmt.Sprintf("- **%s** (`%s`)\n", success.Name, success.DeviceID))
		}
		sb.WriteString("\n")
	}

	if len(result.Failures) > 0 {
		sb.WriteString(fmt.Sprintf("## Failed to Remove (%d)\n", len(result.Failures)))
		for _, failure := range result.Failures {
			sb.WriteString(fmt.Sprintf("- **%s** (`%s`): %s\n", failure.Name, failure.DeviceID, failure.Error))
		}
		sb.WriteString("\n")
	}

	if len(result.Successes) == 0 && len(result.Failures) == 0 {
		sb.WriteString("No devices were processed.\n")
	}

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{mcp.NewTextContent(sb.String())},
	}, nil
}
