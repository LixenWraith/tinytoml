// Package tinytoml provides a simplified TOML encoder and decoder
package tinytoml

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// Marshal converts a Go value into TOML format.
// It supports basic types (string, int, float, bool), arrays, and nested structures.
// Maps must have string keys. Struct fields can use 'toml' tags for customization.
func Marshal(v any) ([]byte, error) {
	pc, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(pc).Name()

	if v == nil {
		return nil, errorf(fn, fmt.Errorf(errNilValue))
	}

	input := getBareValue(reflect.ValueOf(v))

	if input.Kind() != reflect.Struct && input.Kind() != reflect.Map {
		return nil, errorf(fn, fmt.Errorf(errUnsupported))
	}

	m := &marshaller{
		buffer: &bytes.Buffer{},
		path:   []string{},
		depth:  0,
	}

	if err := m.marshalValue(input); err != nil {
		return m.buffer.Bytes(), errorf(fn, err)
	}
	return m.buffer.Bytes(), nil
}

// marshaller handles the TOML encoding process by maintaining the current state
// including output buffer, current table path and nesting depth
type marshaller struct {
	buffer *bytes.Buffer
	path   []string
	depth  int
}

// marshalValue encodes a reflect.Value into TOML format based on its kind.
// It handles basic types, arrays, maps and structs recursively.
func (m *marshaller) marshalValue(v reflect.Value) error {
	pc, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(pc).Name()

	if isUnsupportedType(getBareValue(v).Kind()) {
		return errorf(fn, fmt.Errorf(errUnsupported))
	}

	switch v.Kind() {
	case reflect.Struct:
		if err := m.marshalStruct(v); err != nil {
			return errorf(fn, err)
		}
	case reflect.Map:
		if err := m.marshalMap(v); err != nil {
			return errorf(fn, err)
		}
	case reflect.Slice, reflect.Array:
		if err := m.marshalSlice(v); err != nil {
			return errorf(fn, err)
		}
	case reflect.String:
		if err := m.marshalString(v); err != nil {
			return errorf(fn, err)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if err := m.marshalInt(v); err != nil {
			return errorf(fn, err)
		}
	case reflect.Float32, reflect.Float64:
		if err := m.marshalFloat(v); err != nil {
			return errorf(fn, err)
		}
	case reflect.Bool:
		if err := m.marshalBool(v); err != nil {
			return errorf(fn, err)
		}
	default:
		return errorf(fn, fmt.Errorf(errUnsupported))
	}
	return nil
}

// marshalStruct encodes a struct into TOML format.
// Fields are sorted alphabetically and nested structures create new tables.
// It respects toml tags for field names and skip directives.
func (m *marshaller) marshalStruct(v reflect.Value) error {
	pc, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(pc).Name()

	t := v.Type()
	type fieldInfo struct {
		tomlName  string
		fieldName string
	}
	sortedFields := []fieldInfo{}
	sortedNestedFields := []fieldInfo{}

	// Collect and sort field names
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		tomlName, include := getFieldName(field)
		if !include {
			continue
		}

		fieldValue := getBareValue(v.Field(i))
		info := fieldInfo{tomlName: tomlName, fieldName: field.Name}

		if fieldValue.Kind() == reflect.Map || fieldValue.Kind() == reflect.Struct {
			sortedNestedFields = append(sortedNestedFields, info)
		} else {
			sortedFields = append(sortedFields, info)
		}
	}
	sort.Slice(sortedFields, func(i, j int) bool {
		return strings.ToLower(sortedFields[i].tomlName) < strings.ToLower(sortedFields[j].tomlName)
	})
	sort.Slice(sortedNestedFields, func(i, j int) bool {
		return strings.ToLower(sortedNestedFields[i].tomlName) < strings.ToLower(sortedNestedFields[j].tomlName)
	})

	// Marshal non-nested fields
	for _, info := range sortedFields {
		value := getBareValue(v.FieldByName(info.fieldName))

		m.buffer.WriteString(info.tomlName)
		m.buffer.WriteString(" = ")
		if err := m.marshalValue(value); err != nil {
			return errorf(fn, err)
		}
		m.buffer.WriteString("\n")
	}

	// Marshal nested fields
	for _, info := range sortedNestedFields {
		m.pushLevel(info.tomlName)

		m.buffer.WriteString("[")
		m.buffer.WriteString(strings.Join(m.path, "."))
		m.buffer.WriteString("]\n")

		value := getBareValue(v.FieldByName(info.fieldName))
		if err := m.marshalValue(value); err != nil {
			return errorf(fn, err)
		}

		m.popLevel()
	}

	return nil
}

// marshalMap processes and encodes a map value into TOML format.
// Keys must be strings and are sorted alphabetically.
// Nested maps and structs create new tables with dotted notation.
func (m *marshaller) marshalMap(v reflect.Value) error {
	pc, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(pc).Name()

	if v.Len() == 0 || v.IsNil() {
		return nil
	}

	hasNestedValue := func(v reflect.Value) bool {
		if v.Kind() == reflect.Map || v.Kind() == reflect.Struct {
			return true
		}
		return false
	}

	sortedKeys := []string{}
	sortedNestedKeys := []string{}

	keys := v.MapKeys()
	for _, k := range keys {
		if k.Kind() != reflect.String {
			return errorf(fn, fmt.Errorf(errInvalidKey), errInvalidString)
		}
		key := k.String()
		if !isValidKey(key) {
			return errorf(fn, fmt.Errorf(errInvalidKey), key)
		}
		if hasNestedValue(getBareValue(v.MapIndex(k))) {
			sortedNestedKeys = append(sortedNestedKeys, key)
		} else {
			sortedKeys = append(sortedKeys, key)
		}
	}
	sort.Strings(sortedKeys)
	sort.Strings(sortedNestedKeys)

	for _, key := range sortedKeys {
		value := getBareValue(v.MapIndex(reflect.ValueOf(key)))

		m.buffer.WriteString(key)
		m.buffer.WriteString(" = ")
		if err := m.marshalValue(value); err != nil {
			return errorf(fn, err)
		}
		m.buffer.WriteString("\n")
	}

	for _, key := range sortedNestedKeys {
		m.pushLevel(key)

		m.buffer.WriteString("[")
		m.buffer.WriteString(strings.Join(m.path, "."))
		m.buffer.WriteString("]\n")

		value := getBareValue(v.MapIndex(reflect.ValueOf(key)))

		if err := m.marshalValue(value); err != nil {
			return errorf(fn, err)
		}
		m.popLevel()
	}
	return nil
}

// marshalSlice converts a slice or array into TOML array format.
// Empty slices are encoded as []. Elements are comma-separated.
func (m *marshaller) marshalSlice(v reflect.Value) error {
	pc, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(pc).Name()

	if v.Len() == 0 {
		m.buffer.WriteString("[]")
		return nil
	}

	m.buffer.WriteString("[")

	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			m.buffer.WriteString(", ")
		}

		elem := getBareValue(v.Index(i))
		if isUnsupportedType(elem.Kind()) {
			return errorf(fn, fmt.Errorf(errUnsupported))
		}
		if elem.Kind() == reflect.Map || elem.Kind() == reflect.Struct {
			return errorf(fn, fmt.Errorf(errUnsupported))
		}

		if err := m.marshalValue(elem); err != nil {
			return errorf(fn, err)
		}
	}

	m.buffer.WriteString("]")
	return nil
}

// marshalString encodes a string value with proper escaping.
// Handles special characters: tab, newline, carriage return, quote, backslash
func (m *marshaller) marshalString(v reflect.Value) error {
	m.buffer.WriteByte('"')
	for _, c := range v.String() {
		switch c {
		case '\t':
			m.buffer.WriteByte('\\')
			m.buffer.WriteByte('t')
		case '\n':
			m.buffer.WriteByte('\\')
			m.buffer.WriteByte('n')
		case '\r':
			m.buffer.WriteByte('\\')
			m.buffer.WriteByte('r')
		case '"':
			m.buffer.WriteByte('\\')
			m.buffer.WriteByte('"')
		case '\\':
			m.buffer.WriteByte('\\')
			m.buffer.WriteByte('\\')
		default:
			m.buffer.WriteRune(c)
		}
	}
	m.buffer.WriteByte('"')
	return nil
}

// marshalInt formats an integer value (signed or unsigned) in base 10
func (m *marshaller) marshalInt(v reflect.Value) error {
	m.buffer.WriteString(strconv.FormatInt(v.Int(), 10))
	return nil
}

// marshalFloat formats a floating-point number with decimal point
// Ensures at least one decimal place is always present (e.g. 1.0 not 1)
func (m *marshaller) marshalFloat(v reflect.Value) error {
	s := strconv.FormatFloat(v.Float(), 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	m.buffer.WriteString(s)
	return nil
}

// marshalBool converts boolean value to "true" or "false" string
func (m *marshaller) marshalBool(v reflect.Value) error {
	if v.Bool() {
		m.buffer.WriteString("true")
	} else {
		m.buffer.WriteString("false")
	}
	return nil
}

// pushLevel adds a new table segment to the current path and increases depth
func (m *marshaller) pushLevel(key string) {
	m.path = append(m.path, key)
	m.depth++
	return
}

// popLevel removes the last table segment and decreases depth
func (m *marshaller) popLevel() {
	m.depth--
	m.path = m.path[:len(m.path)-1]
	return
}

// getFieldName extracts the TOML key name from struct field tags
// Returns the tag value if present, field name otherwise
// Second return value indicates if field should be included
func getFieldName(field reflect.StructField) (string, bool) {
	if tag, ok := field.Tag.Lookup("toml"); ok {
		if tag == "-" {
			return "", false // Skip this field
		}
		parts := strings.Split(tag, ",")
		if parts[0] != "" {
			return parts[0], true
		}
	}
	return field.Name, true
}
