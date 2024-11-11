package tinytoml

import (
	"strings"
	"testing"
)

func TestMarshalIndent(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		wantErr  bool
	}{
		{
			name: "basic types string quoting rules",
			input: map[string]interface{}{
				"bare_str":        "simple",         // Can be bare
				"quoted_space":    "hello world",    // Contains space
				"quoted_special":  "hash#symbol",    // Contains #
				"quoted_num":      "123",            // Looks like number
				"quoted_bool":     "true",           // Looks like boolean
				"quoted_brackets": "with[brackets]", // Contains []
				"quoted_equals":   "has=equals",     // Contains =
				"quoted_empty":    "",               // Empty string
				"quoted_negative": "-123",           // Starts with -
			},
			expected: `bare_str = simple
quoted_bool = "true"
quoted_brackets = "with[brackets]"
quoted_empty = ""
quoted_equals = "has=equals"
quoted_negative = "-123"
quoted_num = "123"
quoted_space = "hello world"
quoted_special = "hash#symbol"
`,
			wantErr: false,
		},
		{
			name: "escape sequence handling",
			input: map[string]interface{}{
				"escapes": map[string]interface{}{
					"tab":       "before\tafter",
					"newline":   "before\nafter",
					"quote":     "before\"after",
					"backslash": "before\\after",
					"return":    "before\rafter",
				},
			},
			expected: `[escapes]
backslash = "before\\after"
newline = "before\nafter"
quote = "before\"after"
return = "before\rafter"
tab = "before\tafter"
`,
			wantErr: false,
		},
		{
			name: "array indentation rules",
			input: map[string]interface{}{
				"empty":  []string{},
				"single": []string{"one"},
				"multi":  []string{"one", "two", "three"},
				"mixed_content": []string{
					"simple",
					"with space",
					"with\"quote",
					"with\nnewline",
				},
			},
			expected: `empty = []
mixed_content = [
    simple,
    "with space",
    "with\"quote",
    "with\nnewline"
]
multi = [
    one,
    two,
    three
]
single = [one]
`,
			wantErr: false,
		},
		{
			name: "nested table handling",
			input: map[string]interface{}{
				"server": map[string]interface{}{

					"simple": map[string]interface{}{
						"host": "localhost",
						"port": int64(8080),
					},
				},
			},
			expected: `[server.simple]
host = localhost
port = 8080
`,
			wantErr: false,
		},
		{
			name: "number formatting",
			input: map[string]interface{}{
				"numbers": map[string]interface{}{
					"integer":        int64(42),
					"negative":       int64(-42),
					"zero":           int64(0),
					"float_whole":    1.0,
					"float_frac":     3.14159,
					"float_negative": -3.14159,
					"float_zero":     0.0,
				},
			},
			expected: `[numbers]
float_frac = 3.14159
float_negative = -3.14159
float_whole = 1.0
float_zero = 0.0
integer = 42
negative = -42
zero = 0
`,
			wantErr: false,
		},
		{
			name: "array type consistency",
			input: map[string]interface{}{
				"arrays": map[string]interface{}{
					"strings": []string{"a", "b", "c"},
					"ints":    []int64{1, 2, 3},
					"floats":  []float64{1.1, 2.2, 3.3},
					"bools":   []bool{true, false, true},
				},
			},
			expected: `[arrays]
bools = [
    true,
    false,
    true
]
floats = [
    1.1,
    2.2,
    3.3
]
ints = [
    1,
    2,
    3
]
strings = [
    a,
    b,
    c
]
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
			name: "unsupported channel type",
			input: map[string]interface{}{
				"chan": make(chan int),
			},
		},
		{
			name: "invalid table key",
			input: map[string]interface{}{
				"@invalid": "value",
			},
		},
		{
			name: "invalid nested table key",
			input: map[string]interface{}{
				"table": map[string]interface{}{
					"@invalid": "value",
				},
			},
		},
	}

	// Run positive tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalIndent(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalIndent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if string(got) != tt.expected {
				t.Errorf("MarshalIndent() got vs want:\n=== GOT ===\n%s\n=== WANT ===\n%s",
					string(got), tt.expected)
			}
		})
	}

	// Run error tests
	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := MarshalIndent(tt.input); err == nil {
				t.Errorf("MarshalIndent() expected error for input: %#v", tt.input)
			}
		})
	}
}

func TestMarshalIndentFormatting(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		contains []string
		exclude  []string
	}{
		{
			name: "array indentation",
			input: map[string]interface{}{
				"multi": []string{"one", "two", "three"},
			},
			contains: []string{
				"multi = [\n", // Array opening on first line
				"    one,\n",  // 4-space indent with comma
				"    two,\n",  // 4-space indent with comma
				"    three\n", // 4-space indent no comma
				"]\n",         // Closing bracket
			},
			exclude: []string{
				"multi = [one, two, three]", // Single line array
				"[   one]",                  // Wrong indentation pattern
				"[     one]",                // Wrong indentation pattern
				"    one,two",               // Missing space after comma
			},
		},
		{
			name: "table formatting",
			input: map[string]interface{}{
				"table1": map[string]interface{}{
					"key": "value",
				},
				"table2": map[string]interface{}{
					"key": "value",
				},
			},
			contains: []string{
				"[table1]\n",    // Table header format
				"key = value\n", // Key-value format
				"\n[table2]\n",  // Newline between tables
			},
			exclude: []string{
				"[table1][table2]", // Adjacent tables
				"\n\n\n[table2]",   // Multiple blank lines
				" [table1] ",       // Extra spaces in header
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalIndent(tt.input)
			if err != nil {
				t.Fatalf("MarshalIndent() error = %v", err)
			}

			gotStr := string(got)

			for _, s := range tt.contains {
				if !strings.Contains(gotStr, s) {
					t.Errorf("Output missing expected string: %q", s)
				}
			}

			for _, s := range tt.exclude {
				if strings.Contains(gotStr, s) {
					t.Errorf("Output contains forbidden string: %q", s)
				}
			}
		})
	}
}

func TestMarshalIndentConsistency(t *testing.T) {
	input := map[string]interface{}{
		"table": map[string]interface{}{
			"array": []string{"one", "two", "three"},
			"nested": map[string]interface{}{
				"numbers": []int64{1, 2, 3},
			},
		},
	}

	// Test multiple marshals produce identical output
	var outputs []string
	for i := 0; i < 3; i++ {
		out, err := MarshalIndent(input)
		if err != nil {
			t.Fatalf("MarshalIndent() iteration %d failed: %v", i, err)
		}
		outputs = append(outputs, string(out))
	}

	for i := 1; i < len(outputs); i++ {
		if outputs[i] != outputs[0] {
			t.Errorf("Inconsistent output on iteration %d", i)
		}
	}
}
