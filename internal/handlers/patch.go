package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/zorak1103/ha-mcp/internal/jsonpatch"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// patchAction is the action constant for JSON Patch operations.
const patchAction = "patch"

// patchOperationsSchema returns the MCP JSONSchema for the operations parameter.
func patchOperationsSchema() mcp.JSONSchema {
	return mcp.JSONSchema{
		Type:        "array",
		Description: "RFC 6902 JSON Patch operations to apply",
		Items: &mcp.JSONSchema{
			Type:        "object",
			Description: "A single JSON Patch operation",
			Properties: map[string]mcp.JSONSchema{
				"op": {
					Type:        "string",
					Description: "Operation type",
					Enum:        []string{"add", "remove", "replace", "move", "copy", "test"},
				},
				"path": {
					Type:        "string",
					Description: "RFC 6901 JSON Pointer target path (e.g., '/triggers/2/entity_id')",
				},
				"value": {
					Description: "Value to use for add, replace, or test operations",
				},
				"from": {
					Type:        "string",
					Description: "Source path for move and copy operations",
				},
			},
			Required: []string{"op", "path"},
		},
	}
}

// parseOperations extracts and validates operations from MCP args.
// Returns nil result on success; returns an error result on validation failure.
func parseOperations(args map[string]any) ([]jsonpatch.Operation, *mcp.ToolsCallResult) {
	raw, ok := args["operations"]
	if !ok {
		return nil, errorResult("operations is required for patch action")
	}

	rawSlice, ok := raw.([]any)
	if !ok {
		return nil, errorResult("operations must be an array")
	}
	if len(rawSlice) == 0 {
		return nil, errorResult("operations must contain at least one operation")
	}

	ops := make([]jsonpatch.Operation, 0, len(rawSlice))
	for i, rawOp := range rawSlice {
		op, err := parseOneOperation(rawOp, i)
		if err != nil {
			return nil, errorResult(err.Error())
		}
		ops = append(ops, op)
	}

	if err := jsonpatch.Validate(ops); err != nil {
		return nil, errorResult(err.Error())
	}

	return ops, nil
}

// parseOneOperation converts a raw map[string]any to an Operation.
func parseOneOperation(raw any, idx int) (jsonpatch.Operation, error) {
	opMap, ok := raw.(map[string]any)
	if !ok {
		return jsonpatch.Operation{}, fmt.Errorf("operation at index %d must be an object", idx)
	}

	op := jsonpatch.Operation{
		Op:   getString(opMap, "op"),
		Path: getString(opMap, "path"),
		From: getString(opMap, "from"),
	}

	// Preserve Value even if nil (null is valid JSON Patch value)
	if v, hasValue := opMap["value"]; hasValue {
		op.Value = v
	}

	return op, nil
}

// configToMap converts a typed config struct to map[string]any via JSON round-trip.
func configToMap(config any) (map[string]any, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize config: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to deserialize config: %w", err)
	}
	return m, nil
}

// mapToStruct converts map[string]any back to a typed struct via JSON round-trip.
func mapToStruct(data map[string]any, target any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to serialize patched config: %w", err)
	}
	if err := json.Unmarshal(b, target); err != nil {
		return fmt.Errorf("failed to deserialize patched config: %w", err)
	}
	return nil
}
