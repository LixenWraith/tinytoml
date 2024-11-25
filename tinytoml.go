// Package tinytoml implements a minimalist TOML parser and encoder that supports
// core TOML functionality while maintaining simplicity.
//
// Features:
//   - Basic value types: strings, integers, floats, booleans
//   - Arrays of basic types, nested arrays, and mixed-type arrays
//   - Nested tables using dotted notation
//   - Dotted keys within tables (e.g. server.network.ip = "1.1.1.1")
//   - Struct tags for custom field names (e.g. `toml:"name"`)
//   - Comment handling (inline and single-line)
//   - Whitespace tolerance
//   - Table merging (last value wins)
//   - Basic string escape sequences (\n, \t, \r, \\)
//
// Limitations:
//   - No support for table arrays
//   - No support for hex, octal, binary, or exponential number formats
//   - No support for plus sign in front of numbers
//   - No multi-line keys or strings
//   - No inline table declarations
//   - No inline array declarations within tables
//   - No empty table declarations
//   - No datetime types
//   - No unicode escape sequences
//   - No key character escaping
//   - No literal strings (single quotes)
//   - Comments are discarded during parsing
//
// The package aims for simplicity over completeness, making it suitable for
// basic configuration needs while maintaining strict TOML compatibility
// within its supported feature set.
package tinytoml

import (
	"fmt"
	"reflect"
	"strings"
)

// Error constants used throughout the package for consistent error messaging.
const (
	errNilValue           = "cannot marshal nil value"
	errMissingKey         = "missing key"
	errMissingValue       = "missing value"
	errUnsupported        = "unsupported type"
	errInvalidKey         = "invalid key format"
	errInvalidValue       = "invalid value format"
	errInvalidFormat      = "invalid TOML format"
	errInvalidTarget      = "unmarshal target invalid"
	errInvalidString      = "invalid string format"
	errInvalidInteger     = "invalid integer format"
	errInvalidFloat       = "invalid float format"
	errInvalidBoolean     = "invalid boolean format"
	errUnterminatedString = "unterminated string"
	errUnterminatedArray  = "unterminated array"
	errUnterminatedEscape = "unterminated escape sequence"
	errInvalidEscape      = "invalid escape sequence"
	errInvalidTableName   = "invalid table name"
)

// SupportedTypes lists all Go types that can be marshaled/unmarshaled
// Includes basic types, composites and their variants
var SupportedTypes = []reflect.Kind{
	reflect.Map,
	reflect.String,
	reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
	reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
	reflect.Float32, reflect.Float64,
	reflect.Bool,
	reflect.Struct,
	reflect.Slice,
	reflect.Array,
	reflect.Interface,
}

// errorf formats an error with optional context information
// Prefixes the error with the calling function's name for tracing
func errorf(fn string, err error, context ...string) error {
	if len(context) > 0 {
		return fmt.Errorf("%s: %v [%s]", fn, err, strings.Join(context, ", "))
	}
	return fmt.Errorf("%s: %v", fn, err)
}

// isUnsupportedType checks if a reflect.Kind is not in SupportedTypes
func isUnsupportedType(t reflect.Kind) bool {
	for _, kind := range SupportedTypes {
		if kind == t {
			return false
		}
	}
	return true
}

// isAlpha checks if a character is a letter (A-Z, a-z)
func isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isNumeric checks if a character is a digit (0-9)
func isNumeric(c rune) bool {
	return c >= '0' && c <= '9'
}

// isValidKey checks if a string is a valid TOML key
// Must start with letter/underscore, followed by letters/numbers/dashes/underscores
func isValidKey(s string) bool {
	if len(s) == 0 {
		return false
	}

	firstChar := rune(s[0])
	if !isAlpha(firstChar) && firstChar != '_' {
		return false
	}

	for _, c := range s[1:] {
		if !isAlpha(c) && !isNumeric(c) && c != '-' && c != '_' && c != '.' {
			return false
		}
	}
	return true
}

// getBareValue unwraps interface values to their underlying type
func getBareValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Interface {
		return v.Elem()
	} else {
		return v
	}
}
