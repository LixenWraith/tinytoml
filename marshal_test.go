package tinytoml

import (
	"reflect"
	"strings"
	"testing"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected string
		wantErr  bool
	}{
		{
			name: "basic types",
			input: map[string]any{
				"str":   "hello",
				"num":   int64(42),
				"float": 3.14,
				"bool":  true,
			},
			expected: "bool = true\nfloat = 3.14\nnum = 42\nstr = hello\n",
			wantErr:  false,
		},
		{
			name:     "empty map",
			input:    map[string]any{},
			expected: "",
			wantErr:  false,
		},
		{
			name: "nested tables",
			input: map[string]any{
				"database": map[string]any{
					"host": "localhost",
					"port": int64(5432),
					"primary": map[string]any{
						"user": "admin",
						"pass": "secret",
					},
				},
			},
			expected: `[database]
host = localhost
port = 5432

[database.primary]
pass = secret
user = admin
`,
			wantErr: false,
		},
		{
			name: "arrays of different types",
			input: map[string]any{
				"strings": []string{"a", "hello world", "bare_string"},
				"numbers": []int64{1, -2, 3},
				"floats":  []float64{1.1, -2.2, 3.3},
				"bools":   []bool{true, false, true},
			},
			expected: `bools = [true, false, true]
floats = [1.1, -2.2, 3.3]
numbers = [1, -2, 3]
strings = [a, "hello world", bare_string]
`,
			wantErr: false,
		},
		{
			name: "special string cases",
			input: map[string]any{
				"empty":     "",
				"spaces":    "hello world",
				"tabs":      "tab\there",
				"newlines":  "line\nbreak",
				"quotes":    "\"quoted\"",
				"backslash": "back\\slash",
				"true":      "true",  // should be quoted
				"false":     "false", // should be quoted
				"number":    "42",    // should be quoted
				"bare":      "simple",
			},
			expected: `backslash = "back\\slash"
bare = simple
empty = ""
false = "false"
newlines = "line\nbreak"
number = "42"
quotes = "\"quoted\""
spaces = "hello world"
tabs = "tab\there"
true = "true"
`,
			wantErr: false,
		},
		{
			name: "deep nesting",
			input: map[string]any{
				"a": map[string]any{
					"key": "value1",
				},
				"m": map[string]any{
					"key": "value2",
				},
				"z": map[string]any{
					"key": "value3",
				},
			},
			expected: `[a]
key = value1

[m]
key = value2

[z]
key = value3
`,
			wantErr: false,
		},
	}

	// Run positive tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			gotStr := string(got)
			if gotStr != tt.expected {
				t.Errorf("Marshal() \ngot  = %#v\nwant = %#v", gotStr, tt.expected)
			}

			// Verify round trip
			var unmarshaled map[string]any
			if err := Unmarshal(got, &unmarshaled); err != nil {
				t.Errorf("Round trip unmarshal failed: %v", err)
				return
			}

			if !reflect.DeepEqual(unmarshaled, tt.input) {
				t.Errorf("Round trip failed:\noriginal  = %#v\nround trip = %#v", tt.input, unmarshaled)
			}
		})
	}

	// Error test cases
	errorTests := []struct {
		name  string
		input any
	}{
		{
			name:  "nil value",
			input: nil,
		},
		{
			name: "invalid key type",
			input: map[int]string{
				1: "value",
			},
		},
		{
			name: "unsupported type",
			input: map[string]any{
				"channel": make(chan int),
			},
		},
		{
			name: "invalid nested type",
			input: map[string]any{
				"nested": struct{}{},
			},
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Marshal(tt.input); err == nil {
				t.Errorf("Marshal() expected error for input: %#v", tt.input)
			}
		})
	}
}

func TestMarshalTableOrder(t *testing.T) {
	input := map[string]any{
		"z": map[string]any{
			"key": "value3",
		},
		"a": map[string]any{
			"key": "value1",
		},
		"m": map[string]any{
			"key": "value2",
		},
	}

	got, err := Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Verify tables are in alphabetical order
	lines := strings.Split(string(got), "\n")
	var tables []string
	for _, line := range lines {
		if strings.HasPrefix(line, "[") {
			tables = append(tables, line)
		}
	}

	expected := []string{"[a]", "[m]", "[z]"}
	if !reflect.DeepEqual(tables, expected) {
		t.Errorf("Marshal() table order incorrect\ngot = %v\nwant = %v", tables, expected)
	}
}
