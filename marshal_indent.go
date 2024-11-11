package tinytoml

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// MarshalIndent returns a prettified TOML representation of v with consistent
// indentation and formatting. Arrays longer than one line are split with each
// element on its own line. Table headers are separated by newlines for readability.
//
// Formatting rules:
//   - Arrays with multiple elements are split across lines with 4-space indentation
//   - Table headers are separated by blank lines
//   - Key-value pairs maintain original formatting unless containing multi-line arrays
//   - Proper indentation is maintained for nested structures
//   - Output maintains consistent ordering of elements
func MarshalIndent(v any) ([]byte, error) {
	const fn = "MarshalIndent"

	data, err := Marshal(v)
	if err != nil {
		return nil, errorf(fn, err)
	}

	var buf bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()

		// Format arrays that contain multiple elements
		if strings.Contains(line, "=") {
			key := line[:strings.Index(line, "=")+1]
			arrayPart := strings.TrimSpace(line[strings.Index(line, "=")+1:])

			if strings.HasPrefix(arrayPart, "[") && strings.HasSuffix(arrayPart, "]") && strings.Contains(arrayPart, ",") {
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
		return nil, errorf(fn, fmt.Errorf(errReadFailed), err.Error())
	}

	return buf.Bytes(), nil
}

// splitForIndent splits array elements for indented formatting.
// Handles nested arrays and quoted strings properly when splitting.
// Maintains proper nesting depth to split only at top-level commas.
// Used internally by MarshalIndent to format arrays.
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
