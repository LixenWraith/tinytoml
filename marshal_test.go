package tinytoml

import (
	"reflect"
	"strings"
	"testing"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		wantErr  bool
	}{
		{
			name: "basic types",
			input: map[string]interface{}{
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
			input:    map[string]interface{}{},
			expected: "",
			wantErr:  false,
		},
		{
			name: "nested tables",
			input: map[string]interface{}{
				"database": map[string]interface{}{
					"host": "localhost",
					"port": int64(5432),
					"primary": map[string]interface{}{
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
			input: map[string]interface{}{
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
			input: map[string]interface{}{
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
			input: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": map[string]interface{}{
							"d": map[string]interface{}{
								"key": "value",
							},
						},
					},
				},
			},
			expected: `[a.b.c.d]
key = value
`,
			wantErr: false,
		},
		{
			name: "multiple tables same level",
			input: map[string]interface{}{
				"server1": map[string]interface{}{
					"host": "host1",
					"port": int64(8081),
				},
				"server2": map[string]interface{}{
					"host": "host2",
					"port": int64(8082),
				},
			},
			expected: `[server1]
host = host1
port = 8081

[server2]
host = host2
port = 8082
`,
			wantErr: false,
		},
	}

	errorTests := []struct {
		name  string
		input interface{}
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
			input: map[string]interface{}{
				"channel": make(chan int),
			},
		},
		{
			name: "invalid nested type",
			input: map[string]interface{}{
				"nested": struct{}{},
			},
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

			// Normalize line endings and compare
			gotStr := string(got)
			if gotStr != tt.expected {
				t.Errorf("Marshal() \ngot  = %#v\nwant = %#v", gotStr, tt.expected)
			}

			// Verify round trip
			var unmarshaled map[string]interface{}
			if err := Unmarshal(got, &unmarshaled); err != nil {
				t.Errorf("Round trip unmarshal failed: %v", err)
				return
			}

			if !reflect.DeepEqual(unmarshaled, tt.input) {
				t.Errorf("Round trip failed:\noriginal  = %#v\nround trip = %#v", tt.input, unmarshaled)
			}
		})
	}

	// Run error tests
	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Marshal(tt.input); err == nil {
				t.Errorf("Marshal() expected error for input: %#v", tt.input)
			}
		})
	}
}

func TestMarshalFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{
			name:     "integer float",
			input:    42.0,
			expected: "42.0",
		},
		{
			name:     "decimal float",
			input:    3.14159,
			expected: "3.14159",
		},
		{
			name:     "negative float",
			input:    -2.5,
			expected: "-2.5",
		},
		{
			name:     "zero float",
			input:    0.0,
			expected: "0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]interface{}{"value": tt.input}
			got, err := Marshal(input)
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			expected := "value = " + tt.expected + "\n"
			if string(got) != expected {
				t.Errorf("Marshal() got = %v, want %v", string(got), expected)
			}
		})
	}
}

func TestMarshalString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "special chars",
			input:    "tab\tnewline\nquote\"slash\\",
			expected: "\"tab\\tnewline\\nquote\\\"slash\\\\\"",
		},
		{
			name:     "bare string",
			input:    "simple",
			expected: "simple",
		},
		{
			name:     "empty string",
			input:    "",
			expected: `""`,
		},
		{
			name:     "space containing",
			input:    "hello world",
			expected: `"hello world"`,
		},
		{
			name:     "number-like",
			input:    "123",
			expected: `"123"`,
		},
		{
			name:     "boolean-like",
			input:    "true",
			expected: `"true"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]interface{}{"value": tt.input}
			got, err := Marshal(input)
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			expected := "value = " + tt.expected + "\n"
			if string(got) != expected {
				t.Errorf("Marshal() got = %v, want %v", string(got), expected)
			}
		})
	}
}

func TestMarshalTableOrder(t *testing.T) {
	input := map[string]interface{}{
		"z": map[string]interface{}{
			"key": "value3",
		},
		"a": map[string]interface{}{
			"key": "value1",
		},
		"m": map[string]interface{}{
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

	if !reflect.DeepEqual(tables, []string{"[a]", "[m]", "[z]"}) {
		t.Errorf("Marshal() table order incorrect, got = %v", tables)
	}
}
