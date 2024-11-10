// File: /tinytoml/marshal.go

package tinytoml

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Error constants for different error categories
const (
	ErrNilPointer    = "cannot marshal nil pointer"
	ErrMarshalStruct = "marshal target must be a struct"
)

// Marshal converts a struct to TOML format
func Marshal(v interface{}) ([]byte, error) {
	const fn = "Marshal"

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, errorf(fn, ErrNilPointer)
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, errorf(fn, ErrMarshalStruct)
	}

	m := &marshaler{
		groups: make(map[string]map[string]string),
	}

	if err := m.marshalStruct(val); err != nil {
		return nil, wrapf(fn, err)
	}

	return m.encode()
}

// marshaler holds marshaling state
type marshaler struct {
	groups map[string]map[string]string
}

// marshalStruct processes a struct value
func (m *marshaler) marshalStruct(val reflect.Value) error {
	const fn = "marshaler.marshalStruct"

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		tag := fieldType.Tag.Get("toml")
		if tag == "-" {
			continue
		}

		group, key := "", tag
		if idx := strings.Index(tag, "."); idx >= 0 {
			group = tag[:idx]
			key = tag[idx+1:]
		}

		// Validate group and key names
		if group != "" {
			for _, part := range strings.Split(group, ".") {
				if !isValidKey(part) {
					return errorf(fn, "%s: group '%s'", ErrInvalidKeyFormat, part)
				}
			}
		}
		if !isValidKey(key) {
			return errorf(fn, "%s: key '%s'", ErrInvalidKeyFormat, key)
		}

		if m.groups[group] == nil {
			m.groups[group] = make(map[string]string)
		}

		str, err := m.marshalValue(field)
		if err != nil {
			return wrapf(fn, err)
		}

		if _, exists := m.groups[group][key]; exists {
			return errorf(fn, "%s: '%s'", ErrDuplicateKey, key)
		}
		m.groups[group][key] = str
	}

	return nil
}

// marshalValue converts a reflect.Value to a TOML-compatible string
func (m *marshaler) marshalValue(v reflect.Value) (string, error) {
	const fn = "marshaler.marshalValue"

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		needsQuotes := strings.ContainsAny(s, "\"\\\n\t '") ||
			strings.Contains(s, "#") ||
			!isASCII(s) ||
			s == ""

		if needsQuotes {
			s = strings.ReplaceAll(s, "\\", "\\\\")
			s = strings.ReplaceAll(s, "\"", "\\\"")
			s = strings.ReplaceAll(s, "'", "\\'")
			s = strings.ReplaceAll(s, "\t", "\\t")
			return fmt.Sprintf("\"%s\"", s), nil
		}
		return s, nil

	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i := v.Int()
		if i > math.MaxInt64 || i < math.MinInt64 {
			return "", errorf(fn, ErrIntegerOverflow)
		}
		return strconv.FormatInt(i, 10), nil

	case reflect.Float32, reflect.Float64:
		f := v.Float()
		s := strconv.FormatFloat(f, 'f', -1, 64)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s, nil

	case reflect.Slice:
		return m.marshalArray(v)

	case reflect.Interface:
		if v.IsNil() {
			return "null", nil
		}
		return m.marshalValue(v.Elem())

	default:
		return "", errorf(fn, "%s: %v", ErrUnsupportedType, v.Type())
	}
}

// marshalArray marshals arrays
func (m *marshaler) marshalArray(v reflect.Value) (string, error) {
	const fn = "marshaler.marshalArray"

	if v.Len() == 0 {
		return "[]", nil
	}

	var elements []string
	for i := 0; i < v.Len(); i++ {
		elem, err := m.marshalValue(v.Index(i))
		if err != nil {
			return "", wrapf(fn, err)
		}
		elements = append(elements, elem)
	}

	return fmt.Sprintf("[%s]", strings.Join(elements, ", ")), nil
}

// isASCII checks if string contains non-ASCII characters
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// encode produces the final TOML document
func (m *marshaler) encode() ([]byte, error) {
	const fn = "marshaler.encode"

	var buf bytes.Buffer

	// Get all groups and sort
	var groups []string
	for group := range m.groups {
		groups = append(groups, group)
	}
	sort.Strings(groups)

	// Handle root group first
	if rootGroup, ok := m.groups[""]; ok && len(rootGroup) > 0 {
		if err := m.writeGroup(&buf, rootGroup); err != nil {
			return nil, wrapf(fn, err)
		}
		buf.WriteByte('\n') // Add newline after root group
	}

	// Write each group with proper spacing
	isFirst := true
	for _, group := range groups {
		if group == "" {
			continue
		}

		if !isFirst {
			buf.WriteByte('\n')
		}
		isFirst = false

		buf.WriteString(fmt.Sprintf("[%s]\n", group))
		if err := m.writeGroup(&buf, m.groups[group]); err != nil {
			return nil, wrapf(fn, err)
		}
	}

	return buf.Bytes(), nil
}

// writeGroup writes a group of key-value pairs
func (m *marshaler) writeGroup(buf *bytes.Buffer, values map[string]string) error {
	const fn = "marshaler.writeGroup"

	var keys []string
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := values[key]
		if value == "" {
			return errorf(fn, "%s: key '%s'", ErrEmptyValue, key)
		}
		if _, err := fmt.Fprintf(buf, "%s = %s\n", key, value); err != nil {
			return wrapf(fn, err)
		}
	}

	return nil
}

func MarshalIndent(v interface{}) ([]byte, error) {
	const fn = "MarshalIndent"

	data, err := Marshal(v)
	if err != nil {
		return nil, wrapf(fn, err)
	}

	var buf bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(data))
	inGroup := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Add newline before groups
		if strings.HasPrefix(trimmed, "[") {
			if inGroup {
				buf.WriteByte('\n')
			}
			inGroup = true
		}

		// Format arrays for better readability
		if strings.Contains(line, "[") && strings.Contains(line, "]") && strings.Contains(line, ",") {
			key := line[:strings.Index(line, "=")+1]
			arrayPart := strings.TrimSpace(line[strings.Index(line, "=")+1:])

			// Only format if it's a non-empty array
			if len(arrayPart) > 2 { // more than just "[]"
				elements := splitForIndent(arrayPart[1 : len(arrayPart)-1])
				buf.WriteString(key)
				buf.WriteString(" [\n")
				for i, elem := range elements {
					buf.WriteString("    ")
					buf.WriteString(strings.TrimSpace(elem))
					if i < len(elements)-1 {
						buf.WriteString(",")
					}
					buf.WriteByte('\n')
				}
				buf.WriteString("]")
			} else {
				buf.WriteString(line)
			}
		} else {
			buf.WriteString(line)
		}
		buf.WriteByte('\n')
	}

	if err := scanner.Err(); err != nil {
		return nil, wrapf(fn, err)
	}

	return buf.Bytes(), nil
}

// Helper function to split array elements for indentation
func splitForIndent(s string) []string {
	var result []string
	var current strings.Builder
	depth := 0
	inQuotes := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch ch {
		case '"':
			if i == 0 || s[i-1] != '\\' {
				inQuotes = !inQuotes
			}
			current.WriteByte(ch)
		case '[':
			if !inQuotes {
				depth++
			}
			current.WriteByte(ch)
		case ']':
			if !inQuotes {
				depth--
			}
			current.WriteByte(ch)
		case ',':
			if !inQuotes && depth == 0 {
				result = append(result, current.String())
				current.Reset()
			} else {
				current.WriteByte(ch)
			}
		default:
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}
