package tinytoml

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Marshal converts a Go value to TOML format.
// It supports basic types (string, int64, float64, bool),
// homogeneous arrays, and nested maps representing tables.
// Returns an error if the value cannot be marshaled according to TOML rules.
func Marshal(v any) ([]byte, error) {
	const fn = "Marshal"
	if v == nil {
		return nil, errorf(fn, fmt.Errorf(errNilValue))
	}

	data, err := anyToMap(v)
	if err != nil {
		return nil, errorf(fn, err)
	}

	m := &marshaler{
		buffer: &bytes.Buffer{},
	}

	if err := m.marshalValue(reflect.ValueOf(data), nil); err != nil {
		return nil, errorf(fn, err)
	}

	return m.buffer.Bytes(), nil
}

// anyToMap converts any Go value to map[string]any
func anyToMap(v any) (map[string]any, error) {
	const fn = "anyToMap"
	if v == nil {
		return nil, errorf(fn, fmt.Errorf(errNilValue))
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, errorf(fn, fmt.Errorf(errNilValue))
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		result := make(map[string]any)
		iter := val.MapRange()
		for iter.Next() {
			k := iter.Key()
			if k.Kind() != reflect.String {
				return nil, errorf(fn, fmt.Errorf(errInvalidString))
			}
			result[k.String()] = iter.Value().Interface()
		}
		return result, nil

	case reflect.Struct:
		result := make(map[string]any)
		t := val.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := val.Field(i)

			if !value.CanInterface() {
				continue
			}

			tag := field.Tag.Get("toml")
			if tag == "" || tag == "-" {
				continue
			}

			switch value.Kind() {
			case reflect.Struct:
				nested, err := anyToMap(value.Interface())
				if err != nil {
					return nil, err
				}
				result[tag] = nested
			case reflect.Slice:
				result[tag] = value.Interface()
			default:
				result[tag] = value.Interface()
			}
		}
		return result, nil

	default:
		return nil, errorf(fn, fmt.Errorf(errUnsupported))
	}
}

// marshaler handles the TOML encoding process and maintains the output buffer.
type marshaler struct {
	buffer *bytes.Buffer
}

// marshal converts a single value to its TOML representation.
// Handles basic types, arrays, and maps without table context.
// Returns error for unsupported types or invalid values.
func (m *marshaler) marshal(v reflect.Value) error {
	const fn = "marshaler.marshal"

	if v.Kind() == reflect.Interface && !v.IsNil() {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		// Maps without path context are treated as values
		if err := m.marshalMap(v, nil); err != nil {
			return errorf(fn, err)
		}
	case reflect.Slice:
		if err := m.marshalArray(v); err != nil {
			return errorf(fn, err)
		}
	case reflect.String:
		if err := m.marshalString(v.String()); err != nil {
			return errorf(fn, err)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if err := m.marshalInt(v.Int()); err != nil {
			return errorf(fn, err)
		}
	case reflect.Float32, reflect.Float64:
		if err := m.marshalFloat(v.Float()); err != nil {
			return errorf(fn, err)
		}
	case reflect.Bool:
		if err := m.marshalBool(v.Bool()); err != nil {
			return errorf(fn, err)
		}
	case reflect.Interface:
		if v.IsNil() {
			return nil
		}
		if err := m.marshal(v.Elem()); err != nil {
			return errorf(fn, err)
		}
	default:
		return errorf(fn, fmt.Errorf(errUnsupported), v.Type().String())
	}
	return nil
}

// marshalValue handles top-level value dispatch based on type.
// For maps, it maintains the table path context for proper TOML structure.
// Other types are marshaled directly as values.
func (m *marshaler) marshalValue(v reflect.Value, path []string) error {
	const fn = "marshaler.marshalValue"

	if v.Kind() == reflect.Interface && !v.IsNil() {
		v = v.Elem()
	}

	if v.Kind() == reflect.Map {
		if err := m.marshalMap(v, path); err != nil {
			return errorf(fn, err)
		}
		return nil
	}

	if err := m.marshal(v); err != nil {
		return errorf(fn, err)
	}
	return nil
}

// marshalMap converts a map to TOML format, handling both tables and key-value pairs.
// For table context (non-nil path), it generates appropriate table headers.
// Keys are processed in sorted order for consistent output.
func (m *marshaler) marshalMap(v reflect.Value, path []string) error {
	const fn = "marshaler.marshalMap"
	if v.Len() == 0 {
		return nil
	}

	// Get and sort keys
	keys := v.MapKeys()
	sortedKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		if k.Kind() != reflect.String {
			return errorf(fn, fmt.Errorf(errInvalidKey), errInvalidString)
		}
		key := k.String()
		if !isValidKey(key) {
			return errorf(fn, fmt.Errorf(errInvalidKey), key)
		}
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	// Group by tables and basic key-values
	tables := make(map[string]reflect.Value)
	basicKVs := make(map[string]reflect.Value)

	for _, key := range sortedKeys {
		value := v.MapIndex(reflect.ValueOf(key))
		if value.Kind() == reflect.Interface {
			value = value.Elem()
		}
		if value.Kind() == reflect.Map {
			tables[key] = value
			continue
		}
		basicKVs[key] = value
	}

	// Write table header only if there are values or it's a leaf table
	if len(path) > 0 && (len(basicKVs) > 0 || len(tables) == 0) {
		m.buffer.WriteString("[")
		m.buffer.WriteString(strings.Join(path, "."))
		m.buffer.WriteString("]\n")
	}

	// Write basic key-values
	if len(basicKVs) > 0 {
		for _, key := range sortedKeys {
			if value, ok := basicKVs[key]; ok {
				m.buffer.WriteString(key)
				m.buffer.WriteString(" = ")
				if err := m.marshalValue(value, nil); err != nil {
					return errorf(fn, err, fmt.Sprintf("key: %s", key))
				}
				m.buffer.WriteByte('\n')
			}
		}
		// Add single newline after basic key-values if non-empty tables follow
		if lenNonEmpty(tables) > 0 {
			m.buffer.WriteByte('\n')
		}
	}

	// Write tables
	if lenNonEmpty(tables) > 0 {
		// Process each table
		first := true
		for _, key := range sortedKeys {
			if value, ok := tables[key]; ok {
				if value.Len() == 0 {
					continue // Skip empty tables
				}
				if !first {
					// Only one newline between tables
					m.buffer.WriteByte('\n')
				}
				first = false
				newPath := append(path, key)
				if err := m.marshalValue(value, newPath); err != nil {
					return errorf(fn, err, fmt.Sprintf("table: %s", strings.Join(newPath, ".")))
				}
			}
		}
	}

	return nil
}

// marshalArray converts a slice to TOML array format.
// Verifies array homogeneity (all elements same type).
// Empty arrays are encoded as [].
func (m *marshaler) marshalArray(v reflect.Value) error {
	const fn = "marshaler.marshalArray"
	if v.Len() == 0 {
		m.buffer.WriteString("[]")
		return nil
	}

	// Verify all elements are same type
	elemType := v.Index(0).Type()
	for i := 1; i < v.Len(); i++ {
		if v.Index(i).Type() != elemType {
			return errorf(fn, fmt.Errorf(errTypeMismatch), fmt.Sprintf("element %d", i))
		}
	}

	m.buffer.WriteByte('[')

	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			m.buffer.WriteString(", ")
		}

		if err := m.marshal(v.Index(i)); err != nil {
			return errorf(fn, err, fmt.Sprintf("element %d", i))
		}
	}

	m.buffer.WriteByte(']')
	return nil
}

// marshalString converts a string to TOML format.
// Adds quotes and escapes special characters when necessary.
// Bare strings are used when possible for better readability.
func (m *marshaler) marshalString(s string) error {
	// Quote strings that:
	// - are empty
	// - contain special characters
	// - could be confused with other TOML types
	// - don't match bare key format
	needsQuotes := s == "" ||
		strings.ContainsAny(s, " \t\n\r\"\\#=[]") ||
		strings.EqualFold(s, "true") ||
		strings.EqualFold(s, "false") ||
		(len(s) > 0 && (s[0] == '-' || (s[0] >= '0' && s[0] <= '9')))

	if needsQuotes {
		m.buffer.WriteByte('"')
		for _, r := range s {
			switch r {
			case '"', '\\':
				m.buffer.WriteByte('\\')
				m.buffer.WriteRune(r)
			case '\t':
				m.buffer.WriteString("\\t")
			case '\n':
				m.buffer.WriteString("\\n")
			case '\r':
				m.buffer.WriteString("\\r")
			default:
				m.buffer.WriteRune(r)
			}
		}
		m.buffer.WriteByte('"')
	} else {
		m.buffer.WriteString(s)
	}
	return nil
}

// marshalInt converts an integer to TOML format.
// Verifies the value fits within int64 bounds.
// Numbers are written in decimal format.
func (m *marshaler) marshalInt(i int64) error {
	const fn = "marshaler.marshalInt"
	if i > math.MaxInt64 || i < math.MinInt64 {
		return errorf(fn, fmt.Errorf(errInvalidValue), fmt.Sprintf("integer overflow: %d", i))
	}
	m.buffer.WriteString(strconv.FormatInt(i, 10))
	return nil
}

// marshalFloat converts a floating-point number to TOML format.
// Uses decimal point notation with minimal precision.
// Ensures decimal point is present even for whole numbers.
func (m *marshaler) marshalFloat(f float64) error {
	// Use -1 precision to get minimal representation
	s := strconv.FormatFloat(f, 'f', -1, 64)
	// Ensure decimal point for whole numbers
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	m.buffer.WriteString(s)
	return nil
}

// marshalBool converts a boolean to TOML format.
// Outputs either "true" or "false".
func (m *marshaler) marshalBool(b bool) error {
	if b {
		m.buffer.WriteString("true")
	} else {
		m.buffer.WriteString("false")
	}
	return nil
}

// lenNotEmpty returns count of non-empty tables
func lenNonEmpty(tables map[string]reflect.Value) int {
	count := 0
	for _, v := range tables {
		if v.Len() > 0 {
			count++
		}
	}
	return count
}
