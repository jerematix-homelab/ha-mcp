package jsonpatch

import (
	"fmt"
	"strings"
)

// validOps is the set of valid RFC 6902 operation names.
var validOps = map[string]bool{
	"add":     true,
	"remove":  true,
	"replace": true,
	"move":    true,
	"copy":    true,
	"test":    true,
}

// Validate checks that all operations are well-formed per RFC 6902.
// It does not execute the operations; use Apply for that.
func Validate(ops []Operation) error {
	for i, op := range ops {
		if err := validateOne(op, i); err != nil {
			return err
		}
	}
	return nil
}

// validateOne validates a single operation at the given index.
func validateOne(op Operation, idx int) error {
	if op.Op == "" {
		return fmt.Errorf("op is required at index %d", idx)
	}
	if !validOps[op.Op] {
		return fmt.Errorf("invalid operation %q at index %d (valid: add, remove, replace, move, copy, test)", op.Op, idx)
	}
	if err := validatePath(op.Path, idx); err != nil {
		return err
	}
	if op.Op == "move" || op.Op == "copy" {
		if op.From == "" {
			return fmt.Errorf("from is required for %q operation at index %d", op.Op, idx)
		}
		if op.From != "" && !strings.HasPrefix(op.From, "/") {
			return fmt.Errorf("from must be a valid JSON Pointer starting with '/' at index %d", idx)
		}
	}
	return nil
}

// validatePath checks that the path is a valid RFC 6901 JSON Pointer.
// Non-empty paths must start with "/". Empty path refers to the root document.
func validatePath(path string, idx int) error {
	if path != "" && !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path %q must start with '/' at index %d", path, idx)
	}
	return nil
}
