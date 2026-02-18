package handlers

import "github.com/zorak1103/ha-mcp/internal/mcp"

// Array mode constants for label/alias update operations.
const (
	arrayModeAdd     = "add"
	arrayModeRemove  = "remove"
	arrayModeReplace = "replace"
)

// applyArrayMode computes the result of applying a mode operation to an array field.
//   - add: append provided to current (dedup, current-first order)
//   - remove: subtract provided from current (silent no-op for missing items)
//   - replace: return provided as-is
func applyArrayMode(current, provided []string, mode string) []string {
	switch mode {
	case arrayModeRemove:
		return applyRemoveMode(current, provided)
	case arrayModeReplace:
		return provided
	default: // arrayModeAdd
		return applyAddMode(current, provided)
	}
}

// applyAddMode appends provided items to current, deduplicating (current-first order).
func applyAddMode(current, provided []string) []string {
	if len(provided) == 0 {
		return current
	}
	seen := make(map[string]bool, len(current)+len(provided))
	result := make([]string, 0, len(current)+len(provided))
	for _, s := range current {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	for _, s := range provided {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// applyRemoveMode removes provided items from current (silent no-op for missing items).
func applyRemoveMode(current, provided []string) []string {
	if len(current) == 0 {
		return current
	}
	removeSet := make(map[string]bool, len(provided))
	for _, s := range provided {
		removeSet[s] = true
	}
	result := make([]string, 0, len(current))
	for _, s := range current {
		if !removeSet[s] {
			result = append(result, s)
		}
	}
	return result
}

// getStringSlice extracts []string from args, preserving empty arrays.
// Returns (nil, false) if key not present or nil, ([]string{}, true) if empty
// array, (values, true) otherwise. Unlike toStringArray, this distinguishes
// "not provided" from "empty array provided".
func getStringSlice(args map[string]any, key string) ([]string, bool) {
	val, exists := args[key]
	if !exists || val == nil {
		return nil, false
	}
	arr, ok := val.([]any)
	if !ok {
		return nil, false
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}
	return result, true
}

// getArrayMode extracts the mode parameter from args with "add" as default.
func getArrayMode(args map[string]any, key string) string {
	if mode, ok := args[key].(string); ok && mode != "" {
		return mode
	}
	return arrayModeAdd
}

// arrayModeSchema returns the JSONSchema definition for a mode parameter.
func arrayModeSchema(fieldName string) mcp.JSONSchema {
	return mcp.JSONSchema{
		Type:        "string",
		Description: "How to apply " + fieldName + " changes: 'add' (default, append to existing), 'remove' (remove from existing), 'replace' (full replacement)",
		Enum:        []string{arrayModeAdd, arrayModeRemove, arrayModeReplace},
	}
}
