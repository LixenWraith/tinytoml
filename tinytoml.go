// File: /tinytoml/tinytoml.go

/*
Package tinytoml provides a minimal TOML parser and encoder for configuration files.

Supported Features:
- Basic types: string, number (int/float), boolean
- Arrays with mixed types
- Nested arrays
- Unlimited table/group nesting
- Single-line string values with escape sequences (\, \", \', \t)
- Both inline (#) and full-line comments
- Flexible whitespace around equals sign and line start
- Quoted string values (must be used for strings containing whitespace)
- Strict whitespace handling (unquoted values cannot contain whitespace)
- Integer overflow detection and number format validation
- Duplicate key detection (first occurrence used)

Limitations:
- No multi-line string support
- Limited escape sequence support (only \, \", \', \t)
- No support for custom time formats
- No support for hex/octal/binary number formats
- No scientific notation support for numbers
- Unquoted strings cannot contain whitespace (use quotes)
*/

package tinytoml

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Error constants for different error categories
const (
	// Syntax Errors
	ErrInvalidKeyFormat   = "invalid key format"
	ErrInvalidGroupFormat = "invalid group format"
	ErrInvalidValueFormat = "invalid value format"
	ErrUnterminatedString = "unterminated string"
	ErrInvalidEscape      = "invalid escape sequence"

	// Type Errors
	ErrTypeMismatch    = "type mismatch"
	ErrUnsupportedType = "unsupported type"

	// Value Errors
	ErrIntegerOverflow = "integer overflow"
	ErrInvalidNumber   = "invalid number format"
	ErrInvalidBoolean  = "invalid boolean value"
	ErrEmptyValue      = "empty value"
	ErrInvalidUTF8     = "invalid UTF-8 encoding"

	// Structure Errors
	ErrDuplicateKey = "duplicate key"
	ErrMissingKey   = "missing key"
)

// TokenType represents the type of TOML value
type TokenType int

const (
	TokenInvalid TokenType = iota
	TokenString
	TokenNumber
	TokenBool
	TokenArray
)

// Value represents a TOML value with type information
type Value struct {
	Type  TokenType // Type of the value
	Raw   string    // Raw string representation
	Group string    // Group this value belongs to
	Array []Value   // Array elements if Type is TokenArray
}

// wrapf adds function context to errors
func wrapf(fn string, err error) error {
	return fmt.Errorf("%s: %w", fn, err)
}

// errorf creates new errors with function context
func errorf(fn string, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %s", fn, msg)
}

// GetString returns string value with validation
func (v *Value) GetString() (string, error) {
	const fn = "Value.GetString"
	if v.Type != TokenString {
		return "", errorf(fn, ErrTypeMismatch)
	}
	val, err := unescapeString(v.Raw)
	if err != nil {
		return "", wrapf(fn, err)
	}
	return val, nil
}

// GetBool returns boolean value with validation
func (v *Value) GetBool() (bool, error) {
	const fn = "Value.GetBool"
	if v.Type != TokenBool {
		return false, errorf(fn, ErrTypeMismatch)
	}
	switch v.Raw {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errorf(fn, ErrInvalidBoolean)
	}
}

// GetInt returns integer value with validation and overflow checking
func (v *Value) GetInt() (int64, error) {
	const fn = "Value.GetInt"
	if v.Type != TokenNumber {
		return 0, errorf(fn, ErrTypeMismatch)
	}
	val, err := parseInt(v.Raw)
	if err != nil {
		return 0, wrapf(fn, err)
	}
	return val, nil
}

// GetFloat returns float value with validation
func (v *Value) GetFloat() (float64, error) {
	const fn = "Value.GetFloat"
	if v.Type != TokenNumber {
		return 0, errorf(fn, ErrTypeMismatch)
	}
	val, err := parseFloat(v.Raw)
	if err != nil {
		return 0, wrapf(fn, err)
	}
	return val, nil
}

// GetArray returns array value with validation
func (v *Value) GetArray() ([]Value, error) {
	const fn = "Value.GetArray"
	if v.Type != TokenArray {
		return nil, errorf(fn, ErrTypeMismatch)
	}
	return v.Array, nil
}

// isValidKey checks if a key name follows TOML specification
func isValidKey(key string) bool {
	if len(key) == 0 {
		return false
	}

	// Check first character
	if !isValidKeyStart(key[0]) {
		return false
	}

	// Check rest of the characters
	dots := 0
	for i := 1; i < len(key); i++ {
		if key[i] == '.' {
			if dots > 0 { // Check consecutive dots
				return false
			}
			dots++
			continue
		}
		if !isValidKeyChar(key[i]) {
			return false
		}
		dots = 0
	}

	return dots == 0 // Key can't end with a dot
}

// isValidKeyStart checks if a character is valid as first character of key
func isValidKeyStart(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

// isValidKeyChar checks if a character is valid within a key
func isValidKeyChar(c byte) bool {
	return isValidKeyStart(c) ||
		(c >= '0' && c <= '9') ||
		c == '-'
}

// unescapeString handles basic string escape sequences
func unescapeString(s string) (string, error) {
	const fn = "unescapeString"

	var result strings.Builder
	escaped := false
	isQuoted := false

	// Handle quoted strings
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
		isQuoted = true
	}

	for i := 0; i < len(s); {
		r := rune(s[i])
		size := 1
		if r >= utf8.RuneSelf {
			r, size = utf8.DecodeRuneInString(s[i:])
			if r == utf8.RuneError {
				return "", errorf(fn, ErrInvalidUTF8)
			}
		}

		if escaped {
			switch r {
			case '\\', '"', '\'', 't':
				if r == 't' {
					result.WriteRune('\t')
				} else {
					result.WriteRune(r)
				}
			default:
				return "", errorf(fn, "%s: \\%c", ErrInvalidEscape, r)
			}
			escaped = false
			i += size
			continue
		}

		if r == '\\' {
			escaped = true
			i++
			continue
		}

		if isQuoted || (r != ' ' && r != '\t') {
			result.WriteRune(r)
		}
		i += size
	}

	if escaped {
		return "", errorf(fn, ErrUnterminatedString)
	}

	return result.String(), nil
}

// parseInt is a helper to parse integer values with overflow checking
func parseInt(s string) (int64, error) {
	const fn = "parseInt"

	var val int64
	var neg bool

	if len(s) == 0 {
		return 0, errorf(fn, ErrEmptyValue)
	}

	// Handle negative numbers
	if s[0] == '-' {
		neg = true
		s = s[1:]
	}

	// Parse digits
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errorf(fn, "%s: %s", ErrInvalidNumber, s)
		}

		// Check for overflow
		if val > (1<<63-1)/10 {
			return 0, errorf(fn, "%s: %s", ErrIntegerOverflow, s)
		}

		digit := int64(c - '0')
		val = val*10 + digit

		// Check for overflow after addition
		if val < 0 {
			return 0, errorf(fn, "%s: %s", ErrIntegerOverflow, s)
		}
	}

	if neg {
		val = -val
		if val > 0 {
			return 0, errorf(fn, "%s: %s", ErrIntegerOverflow, s)
		}
	}

	return val, nil
}

// parseFloat is a helper to parse float values with format validation
func parseFloat(s string) (float64, error) {
	const fn = "parseFloat"

	// Special check for invalid format like "123."
	if strings.HasSuffix(s, ".") {
		return 0, errorf(fn, "%s: trailing decimal point", ErrInvalidNumber)
	}

	var intPart int64
	var fracPart int64
	var fracDiv float64 = 1
	var neg bool

	if len(s) == 0 {
		return 0, errorf(fn, ErrEmptyValue)
	}

	// Handle negative numbers
	if s[0] == '-' {
		neg = true
		s = s[1:]
	}

	// Split on decimal point
	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return 0, errorf(fn, "%s: multiple decimal points", ErrInvalidNumber)
	}

	// Parse integer part
	var err error
	intPart, err = parseInt(parts[0])
	if err != nil {
		return 0, wrapf(fn, err)
	}

	// Parse fractional part if exists
	if len(parts) == 2 {
		for _, c := range parts[1] {
			if c < '0' || c > '9' {
				return 0, errorf(fn, "%s: invalid fractional part", ErrInvalidNumber)
			}
			fracPart = fracPart*10 + int64(c-'0')
			fracDiv *= 10
		}
	}

	result := float64(intPart) + float64(fracPart)/fracDiv
	if neg {
		result = -result
	}

	return result, nil
}
