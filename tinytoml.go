// Package tinytoml provides a lightweight TOML (Tom's Obvious Minimal Language) parser
// and encoder for Go. It focuses on commonly used TOML features while maintaining
// strict parsing rules and predictable behavior.
//
// Features:
//   - Basic TOML types: strings, integers, floats, booleans, and homogeneous arrays
//   - Nested tables with dot notation
//   - Basic string escape sequences (\", \t, \n, \r, \)
//   - Comment support (# for both inline and full-line)
//   - Flexible whitespace handling
//
// Limitations:
//   - Arrays must contain elements of the same type
//   - No support for date/time formats
//   - No support for hex/octal/binary numbers or scientific notation
//   - No support for multi-line strings
//   - No support for inline tables
//   - No support for array of tables
//   - No support for Unicode escapes
//   - No support for +/- inf and nan floats

package tinytoml

import (
	"fmt"
	"strings"
)

// Error constants used throughout the package for consistent error messaging.
const (
	errNilValue           = "cannot marshal nil value"
	errUnsupported        = "unsupported type"
	errInvalidKey         = "invalid key format"
	errInvalidValue       = "invalid value format"
	errInvalidFormat      = "invalid TOML format"
	errInvalidTarget      = "unmarshal target must be a pointer to map[string]any"
	errMissingKey         = "missing key"
	errEmptyValue         = "empty value"
	errInvalidString      = "invalid string format"
	errInvalidNumber      = "invalid number format"
	errUnterminatedString = "unterminated string"
	errUnterminatedArray  = "unterminated array"
	errUnterminatedEscape = "unterminated escape sequence"
	errInvalidEscape      = "invalid escape sequence"
	errTypeMismatch       = "type mismatch in array"
	errReadFailed         = "failed to read input"
	errInvalidTableHeader = "invalid table header format"
	errInvalidTableName   = "invalid table name"
)

// tableGroup represents a TOML table/group with hierarchical structure.
// It maintains parent-child relationships and stores key-value pairs.
type tableGroup struct {
	name     string
	parent   *tableGroup
	children map[string]*tableGroup
	values   map[string]any
}

// newTableGroup creates a new table group with the given name and parent reference.
// It initializes empty maps for children and values.
func newTableGroup(name string, parent *tableGroup) *tableGroup {
	return &tableGroup{
		name:     name,
		parent:   parent,
		children: make(map[string]*tableGroup),
		values:   make(map[string]any),
	}
}

// errorf wraps errors with function context and optional additional context information.
// It creates a formatted error message that includes the function name and details.
func errorf(fn string, err error, context ...string) error {
	if len(context) > 0 {
		return fmt.Errorf("%s: %v [%s]", fn, err, strings.Join(context, ", "))
	}
	return fmt.Errorf("%s: %v", fn, err)
}

// isValidKey checks if a key name follows TOML specification.
// Valid keys contain only ASCII letters, digits, underscores, and hyphens,
// must start with a letter or underscore.
func isValidKey(key string) bool {
	if key == "" {
		return false
	}

	for i, r := range key {
		if i == 0 {
			if !isBareKeyStart(r) {
				return false
			}
		} else {
			if !isBareKeyChar(r) {
				return false
			}
		}
	}
	return true
}

// isBareKeyStart checks if a character is valid as first character of key.
// Only ASCII letters and underscore are valid starting characters.
func isBareKeyStart(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '_'
}

// isBareKeyChar checks if a character is valid within a key.
// Valid characters include ASCII letters, digits, underscore, and hyphen.
func isBareKeyChar(r rune) bool {
	return isBareKeyStart(r) || (r >= '0' && r <= '9') || r == '-'
}

// validateBareString checks if a string can be represented without quotes.
// Returns error if string contains whitespace or special characters.
func validateBareString(s string) error {
	if strings.ContainsAny(s, " \t\n\r\"\\#=[]") {
		return fmt.Errorf(errInvalidString)
	}
	return nil
}

// splitTableKey splits a table key into parts and validates each part.
// Returns error if any part is invalid according to TOML key rules.
func splitTableKey(key string) ([]string, error) {
	parts := strings.Split(key, ".")
	for _, part := range parts {
		if !isValidKey(part) {
			return nil, errorf("splitTableKey", fmt.Errorf(errInvalidTableName), part)
		}
	}
	return parts, nil
}

// flattenTable converts nested tableGroup structure to flat map.
// Recursively processes the table hierarchy to create a single-level map.
func flattenTable(group *tableGroup) map[string]any {
	result := make(map[string]any)

	// Add values
	for k, v := range group.values {
		result[k] = v
	}

	// Add children recursively
	for name, child := range group.children {
		childMap := flattenTable(child)
		if len(childMap) > 0 {
			result[name] = childMap
		}
	}

	return result
}
