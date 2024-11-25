// Package tinytoml provides a simplified TOML encoder and decoder
package tinytoml

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/mitchellh/mapstructure"
)

// Unmarshal parses TOML data into a Go value.
// The target must be a pointer to a struct or map.
// It supports basic types, arrays, and nested structures through tables.
func Unmarshal(data []byte, v any) error {
	pc, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(pc).Name()

	if len(data) == 0 {
		return nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errorf(fn, fmt.Errorf(errInvalidTarget), "type", reflect.TypeOf(rv).String(), "value", reflect.ValueOf(rv).String())
	}

	result := make(map[string]any)
	currentTable := result
	var currentTablePath []string // Track current table context
	lines := bytes.Split(data, []byte("\n"))

	// getOrCreateTable ensures a table path exists, creating missing tables
	// Returns the innermost table for the given path
	getOrCreateTable := func(path []string) (map[string]any, error) {
		current := result
		for _, segment := range path {
			next, ok := current[segment]
			if !ok {
				// Create intermediate table
				m := make(map[string]any)
				current[segment] = m
				current = m
				continue
			}

			if m, ok := next.(map[string]any); ok {
				current = m
			} else {
				return nil, errorf(fn, fmt.Errorf(errInvalidFormat), "type", reflect.TypeOf(m).String(), "value", reflect.ValueOf(m).String())
			}
		}
		return current, nil // Return the current map instead of error
	}

	for lineNum, l := range lines {
		tokens, err := tokenizeLine(string(l))
		if err != nil {
			return errorf(fn, err, append([]string{fmt.Sprintf("line %d", lineNum+1), "tokens"}, func(t []token) []string {
				v := make([]string, len(t))
				for i, tt := range t {
					v[i] = tt.value
				}
				return v
			}(tokens)...)...)
		}

		// Skip empty lines
		if len(tokens) == 0 {
			continue
		}

		if tokens[0].typ == tokenTable {
			segments := strings.Split(tokens[0].value, ".")
			table, err := getOrCreateTable(segments)
			if err != nil {
				return err
			}
			currentTable = table
			currentTablePath = segments
			continue
		}

		// Validate basic key-value structure
		if len(tokens) < 3 || tokens[0].typ != tokenKey || tokens[1].typ != tokenEquals {
			if len(tokens) > 0 && tokens[0].typ != tokenKey {
				return errorf(fn, fmt.Errorf(errMissingKey))
			}
			if len(tokens) > 1 && tokens[1].typ == tokenEquals && len(tokens) < 3 {
				return errorf(fn, fmt.Errorf(errMissingValue))
			}
			return errorf(fn, fmt.Errorf(errInvalidFormat))
		}

		key := tokens[0].value
		if !isValidKey(key) {
			return errorf(fn, fmt.Errorf(errInvalidKey))
		}

		// Parse value based on token type
		value, err := parseValue(tokens[2])
		if err != nil {
			return errorf(fn, err)
		}

		// Check for unexpected tokens after value
		if len(tokens) > 3 {
			return errorf(fn, fmt.Errorf(errInvalidFormat), tokens[0].value, tokens[1].value, tokens[2].value)
		}

		if strings.Contains(key, ".") {
			segments, err := getTableSegments(key)
			if err != nil {
				return errorf(fn, err)
			}

			parentPath := segments[:len(segments)-1]
			finalKey := segments[len(segments)-1]

			var targetTable map[string]any
			if len(parentPath) > 0 {
				// Create full path by combining current table path with parent path
				fullPath := append(currentTablePath, parentPath...)
				targetTable, err = getOrCreateTable(fullPath)
				if err != nil {
					return err
				}
			} else {
				targetTable = currentTable
			}

			targetTable[finalKey] = value
		} else {
			currentTable[key] = value
		}
	}

	// Use mapstructure to decode the map into the target variable
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  v,
		TagName: "toml",
	})
	if err != nil {
		return errorf(fn, err)
	}

	err = decoder.Decode(result)
	if err != nil {
		return errorf(fn, err)
	}

	return nil
}

// parseValue converts a token into its corresponding Go value
// based on the token type (string, integer, float, boolean, array)
func parseValue(t token) (any, error) {
	pc, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(pc).Name()

	switch t.typ {
	case tokenString:
		return t.value, nil
	case tokenFloat:
		if strings.Count(t.value, ".") == 1 {
			if v, err := strconv.ParseFloat(t.value, 64); err == nil {
				return v, nil
			}
		} else {
			return nil, errorf(fn, fmt.Errorf(errInvalidFloat), t.value)
		}
	case tokenInteger:
		if strings.Count(t.value, ".") == 0 {
			if v, err := strconv.ParseInt(t.value, 10, 64); err == nil {
				return v, nil
			}
		} else {
			return nil, errorf(fn, fmt.Errorf(errInvalidInteger), t.value)
		}
	case tokenBoolean:
		return t.value == "true", nil
	case tokenArray:
		return parseArray(t.value)
	default:
		return nil, errorf(fn, fmt.Errorf(errInvalidValue), "default", t.value)
	}
	return nil, errorf(fn, fmt.Errorf(errInvalidValue), "outside", t.value)
}

// parseArray processes array contents into a slice of interface values
// Handles strings, booleans, integers and floats as element types
func parseArray(s string) ([]any, error) {
	pc, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(pc).Name()

	elements := strings.Split(s, ",")
	var result []any

	for _, elem := range elements {
		elem = strings.TrimSpace(elem)
		if elem == "" {
			continue
		}

		var value any
		if strings.HasPrefix(elem, "\"") && strings.HasSuffix(elem, "\"") {
			value = elem[1 : len(elem)-1]
			if _, ok := value.(string); !ok {
				return nil, errorf(fn, fmt.Errorf(errInvalidString))
			}
		} else if elem == "true" || elem == "false" {
			value = elem == "true"
			if _, ok := value.(bool); !ok {
				return nil, errorf(fn, fmt.Errorf(errInvalidBoolean))
			}
		} else if v, err := strconv.ParseInt(elem, 10, 64); err == nil {
			value = v
			if _, ok := value.(int64); !ok {
				return nil, errorf(fn, fmt.Errorf(errInvalidInteger))
			}
		} else if v, err := strconv.ParseFloat(elem, 64); err == nil {
			value = v
			if _, ok := value.(float64); !ok {
				return nil, errorf(fn, fmt.Errorf(errInvalidFloat))
			}
		} else {
			return nil, errorf(fn, fmt.Errorf(errInvalidValue), "array", elem)
		}

		result = append(result, value)
	}

	return result, nil
}

// tokenType represents different kinds of TOML syntax elements
type tokenType int

const (
	tokenError tokenType = iota
	tokenKey
	tokenEquals
	tokenString
	tokenFloat
	tokenInteger
	tokenBoolean
	tokenArray
	tokenTable
)

// token represents a parsed TOML syntax element with its type and value
type token struct {
	typ   tokenType
	value string
}

// tokenizeLine breaks a TOML line into tokens for parsing
// It handles key-value pairs, table headers, and different value types
func tokenizeLine(line string) ([]token, error) {
	pc, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(pc).Name()

	var tokens []token
	var buf strings.Builder
	inString := false
	inValue := false
	inArray := false
	arrayStart := -1

	// Clean the line from whitespaces and comments
	line = cleanLine(line)
	if line == "" {
		return nil, nil
	}

	// Check for table header
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
		tableName := strings.TrimSpace(line[1 : len(line)-1])
		segments, err := getTableSegments(tableName)
		if err != nil {
			return nil, errorf(fn, err, "table name", tableName)
		}
		return []token{{typ: tokenTable, value: strings.Join(segments, ".")}}, nil
	}

	for i := 0; i < len(line); {
		r := rune(line[i])

		// Skip whitespace between tokens (but not in strings)
		if !inString && unicode.IsSpace(r) {
			i++
			continue
		}

		// Handle equals sign
		if r == '=' {
			if buf.Len() > 0 {
				tokens = append(tokens, token{typ: tokenKey, value: buf.String()})
				buf.Reset()
			}
			tokens = append(tokens, token{typ: tokenEquals})
			inValue = true
			i++
			continue
		}

		// Handle array start
		if r == '[' && inValue && !inString && !inArray {
			inArray = true
			arrayStart = i
			bracketCount := 1
			for i++; i < len(line); i++ {
				if line[i] == '[' {
					bracketCount++
				} else if line[i] == ']' {
					bracketCount--
					if bracketCount == 0 {
						arrayContent := strings.TrimSpace(line[arrayStart+1 : i])
						tokens = append(tokens, token{typ: tokenArray, value: arrayContent})
						inArray = false
						inValue = false
						i++
						break
					}
				}
			}
			if bracketCount != 0 {
				return nil, errorf(fn, fmt.Errorf(errUnterminatedArray))
			}
			continue
		}

		// String handling
		if r == '"' {
			if !inString {
				inString = true
				inValue = true
				i++
				continue
			}

			// Check if this quote is escaped
			if i > 0 && line[i-1] == '\\' {
				buf.WriteRune(r)
				i++
				continue
			}

			// End of string
			tokens = append(tokens, token{typ: tokenString, value: buf.String()})
			buf.Reset()
			inString = false
			i++
			continue
		}

		if inString {
			// Handle escape sequences
			if r == '\\' && i+1 < len(line) {
				if i+1 >= len(line) {
					return nil, fmt.Errorf(errUnterminatedEscape)
				}
				next := rune(line[i+1])
				switch next {
				case 't':
					buf.WriteRune('\t')
				case 'n':
					buf.WriteRune('\n')
				case 'r':
					buf.WriteRune('\r')
				case '\\':
					buf.WriteRune('\\')
				default:
					return nil, errorf(fn, fmt.Errorf(errInvalidEscape))
				}
				i += 2
				continue
			}
			buf.WriteRune(r)
			i++
			continue
		}

		// Handle non-string values
		if inValue && buf.Len() == 0 {
			// Boolean
			if strings.HasPrefix(line[i:], "true") {
				tokens = append(tokens, token{typ: tokenBoolean, value: "true"})
				i += 4
				continue
			}
			if strings.HasPrefix(line[i:], "false") {
				tokens = append(tokens, token{typ: tokenBoolean, value: "false"})
				i += 5
				continue
			}

			// Number (will be parsed later)
			if unicode.IsDigit(r) || r == '-' || r == '+' {
				start := i
				dotCount := 0
				hasDigit := false

				// Handle leading sign
				if r == '-' || r == '+' {
					i++
				}

				// Scan the rest
				for i < len(line) {
					c := line[i]
					if unicode.IsDigit(rune(c)) {
						hasDigit = true
						i++
					} else if c == '.' {
						dotCount++
						if dotCount > 1 {
							return nil, errorf(fn, fmt.Errorf(errInvalidFloat))
						}
						i++
					} else {
						break
					}
				}

				if !hasDigit {
					return nil, errorf(fn, fmt.Errorf(errInvalidValue))
				}

				value := line[start:i]
				if dotCount == 0 {
					tokens = append(tokens, token{typ: tokenInteger, value: value})
				} else {
					tokens = append(tokens, token{typ: tokenFloat, value: value})
				}
				continue
			}
		}

		// Building key or other token
		buf.WriteRune(r)
		i++
	}

	// Check for unterminated array
	if inArray {
		return nil, errorf(fn, fmt.Errorf(errUnterminatedArray))
	}

	// Add final token if buffer not empty
	if buf.Len() > 0 {
		if inString {
			return nil, errorf(fn, fmt.Errorf(errUnterminatedString))
		}
		tokens = append(tokens, token{typ: tokenKey, value: buf.String()})
	}

	return tokens, nil
}

// cleanLine removes comments and trims whitespace from a TOML line
// Preserves text within strings, including comment characters
func cleanLine(line string) string {
	var buf strings.Builder
	inString := false

	for i := 0; i < len(line); i++ {
		c := rune(line[i])

		// Handle string content
		if c == '"' {
			if i > 0 && line[i-1] == '\\' {
				buf.WriteRune(c)
				continue
			}
			inString = !inString
			buf.WriteRune(c)
			continue
		}

		// Handle comment outside string
		if c == '#' && !inString {
			break
		}

		buf.WriteRune(c)
	}

	return strings.TrimSpace(buf.String())
}

// getTableSegments splits a table name into its dot-separated segments
// Validates each segment as a valid TOML key
func getTableSegments(tableName string) ([]string, error) {
	segments := strings.Split(tableName, ".")
	for _, segment := range segments {
		if strings.Contains(segment, " ") || !isValidKey(segment) {
			return nil, fmt.Errorf(errInvalidTableName)
		}
	}
	return segments, nil
}
