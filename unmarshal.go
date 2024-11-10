// File: tinytoml/unmarshal.go

package tinytoml

import (
	"bufio"
	"io"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Additional error constants for unmarshaling
const (
	// Parse Errors
	ErrReadFailed        = "failed to read input"
	ErrUnterminatedArray = "unterminated array"
	ErrEmptyGroup        = "empty group name"
	ErrInvalidArray      = "invalid array format"
	ErrInvalidElement    = "invalid array element"

	// Decode Errors
	ErrInvalidTarget     = "decode target must be a non-nil pointer"
	ErrNotStruct         = "decode target must be a struct"
	ErrUnsupportedField  = "unsupported field type"
	ErrFieldOverflow     = "field value overflow"
	ErrArrayTypeMismatch = "array element type mismatch"
)

// Unmarshal parses TOML data into a struct
func Unmarshal(data []byte, v interface{}) error {
	const fn = "Unmarshal"

	p := &parser{
		groups:   make(map[string]map[string]Value),
		seenKeys: make(map[string]bool),
	}

	if err := p.parse(data); err != nil {
		return wrapf(fn, err)
	}

	if err := p.decode(v); err != nil {
		return wrapf(fn, err)
	}

	return nil
}

// parser holds the parsing state
type parser struct {
	groups   map[string]map[string]Value
	current  string          // Current group
	lineNum  int             // For error reporting
	seenKeys map[string]bool // Track duplicate keys
	// Array parsing state
	inArray    bool
	arrayKey   string
	arrayBuf   strings.Builder
	arrayDepth int
}

// parse processes TOML content
func (p *parser) parse(data []byte) error {
	const fn = "parser.parse"

	reader := bufio.NewReader(strings.NewReader(string(data)))
	p.current = "" // Root group
	p.groups[""] = make(map[string]Value)

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return errorf(fn, "%s: %v", ErrReadFailed, err)
		}

		p.lineNum++
		line = strings.TrimSpace(line)

		// Skip empty lines and full-line comments
		if line == "" || strings.HasPrefix(line, "#") {
			if err == io.EOF {
				break
			}
			continue
		}

		// Handle end-of-line comments
		if idx := strings.Index(line, " #"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		if err := p.parseLine(line); err != nil {
			return wrapf(fn, err)
		}

		if err == io.EOF {
			if p.inArray {
				return errorf(fn, "%s at line %d", ErrUnterminatedArray, p.lineNum)
			}
			break
		}
	}

	return nil
}

// parseLine handles a single line of TOML
func (p *parser) parseLine(line string) error {
	const fn = "parser.parseLine"

	line = strings.TrimSpace(line)

	// Skip empty lines and comments
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	// Handle end-of-line comments
	if idx := strings.Index(line, " #"); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}

	// Continue array parsing if we're in an array
	if p.inArray {
		if len(line) > 0 {
			if p.arrayBuf.Len() > 0 {
				p.arrayBuf.WriteByte(' ')
			}
			p.arrayBuf.WriteString(line)
		}

		// Count brackets
		for _, ch := range line {
			if ch == '[' {
				p.arrayDepth++
			} else if ch == ']' {
				p.arrayDepth--
				if p.arrayDepth == 0 {
					// Array complete
					arrayStr := p.arrayBuf.String()
					p.arrayBuf.Reset()
					p.inArray = false

					value, err := p.parseArray(arrayStr)
					if err != nil {
						return wrapf(fn, err)
					}

					// Split key into group.key parts
					keyParts := strings.Split(p.arrayKey, ".")
					actualKey := keyParts[len(keyParts)-1]
					groupName := p.current
					if len(keyParts) > 1 {
						groupName = strings.Join(keyParts[:len(keyParts)-1], ".")
					}

					// Store the value in correct group
					if p.groups[groupName] == nil {
						p.groups[groupName] = make(map[string]Value)
					}
					p.groups[groupName][actualKey] = value
					return nil
				}
			}
		}
		return nil
	}

	// Handle group headers
	if strings.HasPrefix(line, "[") {
		if !strings.HasSuffix(line, "]") {
			return errorf(fn, "%s at line %d", ErrInvalidGroupFormat, p.lineNum)
		}
		return p.parseGroup(line)
	}

	// Parse key-value pair
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return errorf(fn, "%s at line %d", ErrInvalidValueFormat, p.lineNum)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	if key == "" {
		return errorf(fn, "%s at line %d", ErrMissingKey, p.lineNum)
	}

	if value == "" {
		return errorf(fn, "%s at line %d", ErrEmptyValue, p.lineNum)
	}

	// Check if this starts an array
	if strings.HasPrefix(value, "[") {
		p.arrayKey = key
		p.arrayBuf.WriteString(value)
		p.inArray = true
		p.arrayDepth = 0 // Start at 0 since we'll count the first [ below

		for _, ch := range value {
			if ch == '[' {
				p.arrayDepth++
			} else if ch == ']' {
				p.arrayDepth--
				if p.arrayDepth == 0 {
					// Single-line array
					arrayStr := p.arrayBuf.String()
					p.arrayBuf.Reset()
					p.inArray = false

					value, err := p.parseArray(arrayStr)
					if err != nil {
						return wrapf(fn, err)
					}

					// Split key into group.key parts
					keyParts := strings.Split(key, ".")
					actualKey := keyParts[len(keyParts)-1]
					groupName := p.current
					if len(keyParts) > 1 {
						groupName = strings.Join(keyParts[:len(keyParts)-1], ".")
					}

					// Store the value in correct group
					if p.groups[groupName] == nil {
						p.groups[groupName] = make(map[string]Value)
					}
					p.groups[groupName][actualKey] = value
					return nil
				}
			}
		}
		return nil
	}

	return p.parseKeyValue(line)
}

// parseGroup handles a group declaration
func (p *parser) parseGroup(line string) error {
	const fn = "parser.parseGroup"

	group := strings.TrimSpace(line[1 : len(line)-1])
	if group == "" {
		return errorf(fn, "%s at line %d", ErrEmptyGroup, p.lineNum)
	}

	// Validate group name
	parts := strings.Split(group, ".")
	for _, part := range parts {
		if !isValidKey(part) {
			return errorf(fn, "%s: '%s' at line %d", ErrInvalidKeyFormat, part, p.lineNum)
		}
	}

	p.current = group

	// Check for duplicate groups
	if _, exists := p.groups[group]; exists {
		return errorf(fn, "%s: '%s' at line %d", ErrDuplicateKey, group, p.lineNum)
	}

	p.groups[group] = make(map[string]Value)
	return nil
}

// parseKeyValue handles a key-value pair
func (p *parser) parseKeyValue(line string) error {
	const fn = "parser.parseKeyValue"

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return errorf(fn, "%s at line %d", ErrInvalidValueFormat, p.lineNum)
	}

	key := strings.TrimSpace(parts[0])
	if key == "" {
		return errorf(fn, "%s at line %d", ErrMissingKey, p.lineNum)
	}

	// Split key into group.key parts
	keyParts := strings.Split(key, ".")
	actualKey := keyParts[len(keyParts)-1]
	groupName := p.current
	if len(keyParts) > 1 {
		groupName = strings.Join(keyParts[:len(keyParts)-1], ".")
	}

	// Create full key for duplicate checking
	fullKey := actualKey
	if groupName != "" {
		fullKey = groupName + "." + actualKey
	}

	// Skip if key already seen
	if p.seenKeys[fullKey] {
		return nil
	}
	p.seenKeys[fullKey] = true

	if !isValidKey(actualKey) {
		return errorf(fn, "%s: '%s' at line %d", ErrInvalidKeyFormat, actualKey, p.lineNum)
	}

	val := strings.TrimSpace(parts[1])
	if val == "" {
		return errorf(fn, "%s at line %d for key '%s'", ErrEmptyValue, p.lineNum, actualKey)
	}

	value, err := p.parseValue(val)
	if err != nil {
		return wrapf(fn, err)
	}
	if value.Type == TokenInvalid {
		return errorf(fn, "%s at line %d for key '%s'", ErrInvalidValueFormat, p.lineNum, actualKey)
	}

	// Store in correct group
	if p.groups[groupName] == nil {
		p.groups[groupName] = make(map[string]Value)
	}
	p.groups[groupName][actualKey] = value

	return nil
}

// parseValue determines the type and value of a TOML value
func (p *parser) parseValue(val string) (Value, error) {
	const fn = "parser.parseValue"

	val = strings.TrimSpace(val)

	if !utf8.ValidString(val) {
		return Value{Type: TokenInvalid}, errorf(fn, ErrInvalidUTF8)
	}

	// Handle arrays
	if strings.HasPrefix(val, "[") {
		return p.parseArray(val)
	}

	// Handle quoted strings
	if strings.HasPrefix(val, "\"") {
		if !strings.HasSuffix(val, "\"") {
			return Value{Type: TokenInvalid}, errorf(fn, ErrUnterminatedString)
		}

		if _, err := unescapeString(val); err != nil {
			return Value{Type: TokenInvalid}, wrapf(fn, err)
		}

		return Value{
			Type:  TokenString,
			Raw:   val,
			Group: p.current,
		}, nil
	}

	// Boolean
	if val == "true" || val == "false" {
		return Value{
			Type:  TokenBool,
			Raw:   val,
			Group: p.current,
		}, nil
	}

	// Try number
	if !strings.Contains(val, ".") {
		if _, err := parseInt(val); err != nil {
			if strings.Contains(err.Error(), "overflow") {
				return Value{Type: TokenInvalid}, wrapf(fn, err)
			}
		} else {
			return Value{
				Type:  TokenNumber,
				Raw:   val,
				Group: p.current,
			}, nil
		}
	} else {
		if _, err := parseFloat(val); err != nil {
			return Value{Type: TokenInvalid}, wrapf(fn, err)
		} else {
			return Value{
				Type:  TokenNumber,
				Raw:   val,
				Group: p.current,
			}, nil
		}
	}

	// Unquoted string validation
	if strings.ContainsAny(val, " \t'") {
		return Value{Type: TokenInvalid}, errorf(fn, "%s: contains whitespace or quotes", ErrInvalidValueFormat)
	}

	for _, r := range val {
		if !unicode.IsPrint(r) {
			return Value{Type: TokenInvalid}, errorf(fn, "%s: contains non-printable character", ErrInvalidValueFormat)
		}
	}

	// Unquoted string
	return Value{
		Type:  TokenString,
		Raw:   val,
		Group: p.current,
	}, nil
}

// parseArray parses arrays
func (p *parser) parseArray(val string) (Value, error) {
	const fn = "parser.parseArray"

	if !strings.HasSuffix(val, "]") {
		return Value{Type: TokenInvalid}, errorf(fn, ErrUnterminatedArray)
	}

	// Handle empty array
	if val == "[]" {
		return Value{
			Type:  TokenArray,
			Raw:   val,
			Group: p.current,
			Array: []Value{},
		}, nil
	}

	// Remove outer brackets and trim spaces
	content := strings.TrimSpace(val[1 : len(val)-1])
	elements := p.splitArrayElements(content)

	var values []Value
	for _, elem := range elements {
		elemVal, err := p.parseValue(elem)
		if err != nil {
			return Value{Type: TokenInvalid}, errorf(fn, "%s: %v", ErrInvalidElement, err)
		}
		elemVal.Raw = elem
		values = append(values, elemVal)
	}

	return Value{
		Type:  TokenArray,
		Raw:   val,
		Group: p.current,
		Array: values,
	}, nil
}

// splitArrayElements splits array elements
func (p *parser) splitArrayElements(input string) []string {
	var elements []string
	var current strings.Builder
	var depth int
	inQuotes := false

	for i := 0; i < len(input); i++ {
		ch := input[i]

		switch ch {
		case '"':
			if i == 0 || input[i-1] != '\\' {
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
				elements = append(elements, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteByte(ch)
			}
		default:
			current.WriteByte(ch)
		}
	}

	// Add the last element
	if current.Len() > 0 {
		elements = append(elements, strings.TrimSpace(current.String()))
	}

	return elements
}

// decode converts parsed TOML data into a struct
func (p *parser) decode(v interface{}) error {
	const fn = "parser.decode"

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errorf(fn, ErrInvalidTarget)
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return errorf(fn, ErrNotStruct)
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldVal := rv.Field(i)

		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("toml")
		if tag == "-" {
			continue
		}

		parts := strings.Split(tag, ".")
		if len(parts) == 0 {
			continue
		}

		var group, key string
		if len(parts) > 1 {
			group = strings.Join(parts[:len(parts)-1], ".")
			key = parts[len(parts)-1]
		} else {
			group = ""
			key = parts[0]
		}

		// Find the value in the correct group
		groupMap, ok := p.groups[group]
		if !ok {
			continue // Skip if group not found
		}

		val, ok := groupMap[key]
		if !ok {
			continue // Skip if key not found in group
		}

		if err := p.setField(fieldVal, val); err != nil {
			return errorf(fn, "field '%s': %s", field.Name, err)
		}
	}

	return nil
}

// setField sets a struct field to the parsed value
func (p *parser) setField(field reflect.Value, val Value) error {
	const fn = "parser.setField"

	switch field.Kind() {
	case reflect.Slice:
		if val.Type != TokenArray {
			return errorf(fn, ErrTypeMismatch)
		}

		// Create a new slice of the correct type
		sliceType := field.Type()
		elemType := sliceType.Elem()
		slice := reflect.MakeSlice(sliceType, 0, len(val.Array))

		for i, elem := range val.Array {
			newElem := reflect.New(elemType).Elem()
			if err := p.setArrayElement(newElem, elem); err != nil {
				return errorf(fn, "array element %d: %s", i, err)
			}
			slice = reflect.Append(slice, newElem)
		}

		field.Set(slice)
		return nil

	case reflect.String:
		str, err := val.GetString()
		if err != nil {
			return wrapf(fn, err)
		}
		field.SetString(str)

	case reflect.Bool:
		if val.Type != TokenBool {
			return errorf(fn, ErrTypeMismatch)
		}
		b, err := val.GetBool()
		if err != nil {
			return wrapf(fn, err)
		}
		field.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Type != TokenNumber {
			return errorf(fn, ErrTypeMismatch)
		}
		i, err := val.GetInt()
		if err != nil {
			return wrapf(fn, err)
		}
		if field.OverflowInt(i) {
			return errorf(fn, "%s: %v", ErrFieldOverflow, field.Type())
		}
		field.SetInt(i)

	case reflect.Float32, reflect.Float64:
		if val.Type != TokenNumber {
			return errorf(fn, ErrTypeMismatch)
		}
		f, err := val.GetFloat()
		if err != nil {
			return wrapf(fn, err)
		}
		if field.OverflowFloat(f) {
			return errorf(fn, "%s: %v", ErrFieldOverflow, field.Type())
		}
		field.SetFloat(f)

	default:
		return errorf(fn, "%s: %v", ErrUnsupportedField, field.Type())
	}

	return nil
}

// setArrayElement sets array elements
func (p *parser) setArrayElement(field reflect.Value, val Value) error {
	const fn = "parser.setArrayElement"

	switch field.Kind() {
	case reflect.Slice:
		if val.Type != TokenArray {
			return errorf(fn, "%s: expected array value", ErrArrayTypeMismatch)
		}
		sliceType := field.Type()
		elemType := sliceType.Elem()
		slice := reflect.MakeSlice(sliceType, 0, len(val.Array))

		for i, elem := range val.Array {
			newElem := reflect.New(elemType).Elem()
			if err := p.setArrayElement(newElem, elem); err != nil {
				if elemType.Kind() == reflect.Int && !strings.Contains(err.Error(), "overflow") {
					return errorf(fn, "%s: must be number", ErrInvalidElement)
				}
				return errorf(fn, "nested array element %d: %s", i, err)
			}
			slice = reflect.Append(slice, newElem)
		}
		field.Set(slice)
		return nil

	case reflect.Interface:
		var v interface{}
		switch val.Type {
		case TokenString:
			s, err := val.GetString()
			if err != nil {
				return wrapf(fn, err)
			}
			v = s
		case TokenNumber:
			if strings.Contains(val.Raw, ".") {
				f, err := val.GetFloat()
				if err != nil {
					return wrapf(fn, err)
				}
				v = f
			} else {
				i, err := val.GetInt()
				if err != nil {
					return wrapf(fn, err)
				}
				v = i
			}
		case TokenBool:
			b, err := val.GetBool()
			if err != nil {
				return wrapf(fn, err)
			}
			v = b
		case TokenArray:
			arr := make([]interface{}, len(val.Array))
			for i, elem := range val.Array {
				elemField := reflect.New(reflect.TypeOf((*interface{})(nil)).Elem()).Elem()
				if err := p.setArrayElement(elemField, elem); err != nil {
					return errorf(fn, "nested array element %d: %s", i, err)
				}
				arr[i] = elemField.Interface()
			}
			v = arr
		default:
			return errorf(fn, "%s: %v", ErrUnsupportedField, val.Type)
		}
		field.Set(reflect.ValueOf(v))
		return nil

	default:
		return p.setField(field, val)
	}
}
