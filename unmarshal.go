package tinytoml

import (
	"bufio"
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Unmarshal parses TOML data into a map[string]any.
// It supports basic TOML types (string, integer, float, boolean),
// homogeneous arrays, and nested tables using dot notation.
// Returns error if the TOML is invalid or cannot be parsed into the target type.
func Unmarshal(data []byte, v any) error {
	const fn = "Unmarshal"
	p := &parser{
		scanner: bufio.NewScanner(bytes.NewReader(data)),
		root:    newTableGroup("", nil),
		current: nil, // will be set to root during parse
		lineNum: 0,
	}

	if err := p.parse(); err != nil {
		return errorf(fn, err)
	}

	// Convert result to flat map
	result := flattenTable(p.root)

	// Set the result into the provided interface
	if m, ok := v.(*map[string]any); ok {
		*m = result
		return nil
	}
	return errorf(fn, fmt.Errorf(errInvalidTarget))
}

// parser maintains the state during TOML parsing.
// It tracks current position, table context, and builds the result structure.
type parser struct {
	scanner *bufio.Scanner
	root    *tableGroup
	current *tableGroup
	lineNum int
}

// parse is the main parsing loop that processes the TOML document.
// Handles comments, table headers, and key-value pairs.
// Maintains table context for proper nesting.
func (p *parser) parse() error {
	const fn = "parser.parse"
	p.current = p.root

	for p.scanner.Scan() {
		p.lineNum++
		line := p.scanner.Text()

		// Skip empty lines and comments
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle inline comments
		if idx := strings.Index(line, " #"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		// Handle table header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if err := p.parseTableHeader(line); err != nil {
				return errorf(fn, err, fmt.Sprintf("line %d", p.lineNum))
			}
			continue
		}

		// Parse key-value pair
		if err := p.parseLine(line); err != nil {
			return errorf(fn, err, fmt.Sprintf("line %d", p.lineNum))
		}
	}

	if err := p.scanner.Err(); err != nil {
		return errorf(fn, fmt.Errorf(errReadFailed), err.Error())
	}

	return nil
}

// parseTableHeader processes a table header line [table.name].
// Creates or navigates the table hierarchy as needed.
// Returns error if the table header format is invalid.
func (p *parser) parseTableHeader(line string) error {
	const fn = "parser.parseTableHeader"
	// Remove brackets and trim spaces
	tablePath := strings.TrimSpace(line[1 : len(line)-1])
	if tablePath == "" {
		return errorf(fn, fmt.Errorf(errInvalidTableHeader))
	}

	// Split and validate table path
	parts, err := splitTableKey(tablePath)
	if err != nil {
		return errorf(fn, err)
	}

	// Navigate/create table structure
	current := p.root
	for _, part := range parts {
		if child, exists := current.children[part]; exists {
			current = child
		} else {
			newGroup := newTableGroup(part, current)
			current.children[part] = newGroup
			current = newGroup
		}
	}
	p.current = current
	return nil
}

// parseLine processes a key-value pair line.
// Handles dotted keys by creating intermediate tables.
// Returns error if the line format or value is invalid.
func (p *parser) parseLine(line string) error {
	const fn = "parser.parseLine"
	// Split into key and value
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return errorf(fn, fmt.Errorf(errInvalidFormat))
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	if key == "" {
		return errorf(fn, fmt.Errorf(errMissingKey))
	}

	// Handle dotted keys
	keyParts := strings.Split(key, ".")
	for _, part := range keyParts {
		if !isValidKey(part) {
			return errorf(fn, fmt.Errorf(errInvalidKey), part)
		}
	}

	// If it's a dotted key, navigate to the correct table
	current := p.current
	if len(keyParts) > 1 {
		for _, part := range keyParts[:len(keyParts)-1] {
			if child, exists := current.children[part]; exists {
				current = child
			} else {
				newGroup := newTableGroup(part, current)
				current.children[part] = newGroup
				current = newGroup
			}
		}
		key = keyParts[len(keyParts)-1]
	}

	// Don't overwrite existing values (first occurrence wins)
	if _, exists := current.values[key]; exists {
		return nil
	}

	parsedValue, err := p.parseValue(value)
	if err != nil {
		return errorf(fn, err, fmt.Sprintf("key %q", key))
	}

	current.values[key] = parsedValue
	return nil
}

// parseValue converts a TOML value string to its Go representation.
// Handles basic types, quoted strings, and arrays.
// Returns error for invalid formats or unsupported types.
func (p *parser) parseValue(val string) (any, error) {
	const fn = "parser.parseValue"
	val = strings.TrimSpace(val)
	if val == "" {
		return nil, errorf(fn, fmt.Errorf(errEmptyValue))
	}

	// Handle array
	if strings.HasPrefix(val, "[") {
		return p.parseArray(val)
	}

	// Handle quoted string
	if strings.HasPrefix(val, "\"") {
		return p.parseQuotedString(val)
	}

	// Handle boolean
	if val == "true" {
		return true, nil
	}
	if val == "false" {
		return false, nil
	}

	// Handle number
	if v, err := p.parseNumber(val); err == nil {
		return v, nil
	}

	// Handle bare string
	if err := validateBareString(val); err != nil {
		return nil, errorf(fn, err)
	}
	return val, nil
}

// parseArray converts a TOML array string to a Go slice.
// Ensures array elements are of consistent type.
// Returns error if array format is invalid or elements are heterogeneous.
func (p *parser) parseArray(val string) (any, error) {
	const fn = "parser.parseArray"
	if !strings.HasSuffix(val, "]") {
		return nil, errorf(fn, fmt.Errorf(errUnterminatedArray))
	}

	if val == "[]" {
		return []any{}, nil
	}

	content := strings.TrimSpace(val[1 : len(val)-1])
	elements := splitArrayElements(content)
	if len(elements) == 0 {
		return []any{}, nil
	}

	firstElem, err := p.parseValue(elements[0])
	if err != nil {
		return nil, errorf(fn, err, "array element 0")
	}

	return p.parseArrayElements(elements, firstElem, reflect.TypeOf(firstElem))
}

// parseArrayElements parses an array of elements into a typed slice.
// It is a generic function that can handle different types (string, int64, float64, bool)
// based on the provided converter function.
//
// Type parameter T represents the target type for array elements.
//
// Parameters:
//   - elements: slice of string representations of array elements to parse
//   - converter: function that converts a parsed any value to type T
//     The converter returns the converted value and a boolean indicating success
//
// Returns:
//   - []T: slice of parsed and type-converted elements
//   - error: parsing or type conversion error with context
//
// The function ensures type consistency across all elements in the array
// by applying the converter function to each element. If any element fails
// to parse or convert to the target type, an error is returned with the
// element index for context.
//
// Example usage:
//
//	strings, err := p.parseArrayElements(elements, func(v any) (string, bool) {
//	    s, ok := v.(string)
//	    return s, ok
//	})
func (p *parser) parseArrayElements(elements []string, firstElem any, elementType reflect.Type) (any, error) {
	const fn = "parser.parseArrayElements"

	// Create typed slice using reflection
	result := reflect.MakeSlice(reflect.SliceOf(elementType), len(elements), len(elements))

	// Set first element
	result.Index(0).Set(reflect.ValueOf(firstElem))

	// Parse remaining elements
	for i := 1; i < len(elements); i++ {
		val, err := p.parseValue(elements[i])
		if err != nil {
			return nil, errorf(fn, err, fmt.Sprintf("element %d", i))
		}

		if !reflect.TypeOf(val).AssignableTo(elementType) {
			return nil, errorf(fn, fmt.Errorf(errTypeMismatch), fmt.Sprintf("element %d", i))
		}

		result.Index(i).Set(reflect.ValueOf(val))
	}

	return result.Interface(), nil
}

// splitArrayElements splits array elements handling quoted strings properly.
// Maintains proper handling of commas within quoted strings.
// Returns slice of individual element strings.
func splitArrayElements(input string) []string {
	var result []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(input); i++ {
		ch := input[i]
		switch ch {
		case '"':
			if i == 0 || input[i-1] != '\\' {
				inQuotes = !inQuotes
			}
			current.WriteByte(ch)
		case ',':
			if !inQuotes {
				if current.Len() > 0 {
					result = append(result, strings.TrimSpace(current.String()))
					current.Reset()
				}
			} else {
				current.WriteByte(ch)
			}
		default:
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	return result
}

// parseQuotedString processes a quoted string, handling escape sequences.
// Supports basic escapes: \", \\, \t, \n, \r
// Returns error for invalid escape sequences or unterminated strings.
func (p *parser) parseQuotedString(val string) (string, error) {
	const fn = "parser.parseQuotedString"
	if !strings.HasSuffix(val, "\"") {
		return "", errorf(fn, fmt.Errorf(errUnterminatedString))
	}

	// Remove quotes
	val = val[1 : len(val)-1]

	var result strings.Builder
	escaped := false

	for i := 0; i < len(val); i++ {
		c := val[i]
		if escaped {
			switch c {
			case '\\', '"':
				result.WriteByte(c)
			case 't':
				result.WriteByte('\t')
			case 'n':
				result.WriteByte('\n')
			case 'r':
				result.WriteByte('\r')
			default:
				return "", errorf(fn, fmt.Errorf(errInvalidEscape), string(c))
			}
			escaped = false
		} else if c == '\\' {
			escaped = true
		} else {
			result.WriteByte(c)
		}
	}

	if escaped {
		return "", errorf(fn, fmt.Errorf(errUnterminatedEscape))
	}

	return result.String(), nil
}

// parseNumber attempts to parse a string as number.
// Tries integer first, then float.
// Returns error if the string is not a valid number format.
func (p *parser) parseNumber(val string) (any, error) {
	const fn = "parser.parseNumber"
	// Try parsing as integer first
	if i, err := strconv.ParseInt(val, 10, 64); err == nil {
		return i, nil
	}

	// Try parsing as float
	if f, err := strconv.ParseFloat(val, 64); err == nil {
		return f, nil
	}

	return nil, errorf(fn, fmt.Errorf(errInvalidNumber))
}
