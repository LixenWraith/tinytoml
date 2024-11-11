package tinytoml

import (
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name: "basic types",
			input: `str = "hello"
num = 42
float = 3.14
bool = true`,
			expected: map[string]interface{}{
				"str":   "hello",
				"num":   int64(42),
				"float": 3.14,
				"bool":  true,
			},
			wantErr: false,
		},
		{
			name:     "empty document",
			input:    "",
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name: "complex nested tables",
			input: `
[database]
host = "localhost"
port = 5432

[database.primary]
user = "admin"
pass = "secret"

[database.replica]
user = "readonly"
enabled = true`,
			expected: map[string]interface{}{
				"database": map[string]interface{}{
					"host": "localhost",
					"port": int64(5432),
					"primary": map[string]interface{}{
						"user": "admin",
						"pass": "secret",
					},
					"replica": map[string]interface{}{
						"user":    "readonly",
						"enabled": true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mixed arrays",
			input: `
strings = ["a", "hello world", bare_string]
numbers = [1, -2, 3]
floats = [1.1, -2.2, 3.3]
bools = [true, false, true]`,
			expected: map[string]interface{}{
				"strings": []string{"a", "hello world", "bare_string"},
				"numbers": []int64{1, -2, 3},
				"floats":  []float64{1.1, -2.2, 3.3},
				"bools":   []bool{true, false, true},
			},
			wantErr: false,
		},
		{
			name: "dotted keys",
			input: `
server.host = "example.com"
server.port = 8080
database.credentials.username = "admin"
database.credentials.password = "secret"`,
			expected: map[string]interface{}{
				"server": map[string]interface{}{
					"host": "example.com",
					"port": int64(8080),
				},
				"database": map[string]interface{}{
					"credentials": map[string]interface{}{
						"username": "admin",
						"password": "secret",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "comments and whitespace",
			input: `
# Full line comment
key1 = "value1"  # Inline comment

  # Indented comment
  key2 = 42  # Number comment

[section]  # Section comment
key3 = true  # Boolean comment`,
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": int64(42),
				"section": map[string]interface{}{
					"key3": true,
				},
			},
			wantErr: false,
		},
		{
			name: "escape sequences",
			input: `
str1 = "tab:\t"
str2 = "newline:\n"
str3 = "quote:\"" 
str4 = "backslash:\\"`,
			expected: map[string]interface{}{
				"str1": "tab:\t",
				"str2": "newline:\n",
				"str3": "quote:\"",
				"str4": "backslash:\\",
			},
			wantErr: false,
		},
	}

	errorTests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid table header",
			input: `[invalid.[table]`,
		},
		{
			name:  "heterogeneous array",
			input: `mixed = [1, "string", true]`,
		},
		{
			name:  "invalid escape sequence",
			input: `bad = "invalid\escape"`,
		},
		{
			name:  "unterminated string",
			input: `str = "unterminated`,
		},
		{
			name:  "unterminated array",
			input: `arr = [1, 2, 3`,
		},
		{
			name:  "missing value",
			input: `key =`,
		},
	}

	// Run positive tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got map[string]interface{}
			err := Unmarshal([]byte(tt.input), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Unmarshal()\ngot  = %#v\nwant = %#v", got, tt.expected)
			}
		})
	}

	// Run error tests
	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			var got map[string]interface{}
			if err := Unmarshal([]byte(tt.input), &got); err == nil {
				t.Errorf("Unmarshal() expected error for input: %s", tt.input)
			}
		})
	}
}

func TestUnmarshalEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		target  interface{}
		wantErr bool
	}{
		{
			name:    "nil target",
			input:   `key = "value"`,
			target:  nil,
			wantErr: true,
		},
		{
			name:    "non-pointer target",
			input:   `key = "value"`,
			target:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name:    "wrong target type",
			input:   `key = "value"`,
			target:  new(string),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.input), tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
