package tinytoml

import (
	"reflect"
	"strings"
	"testing"
)

// Test structures for different cases
type BasicTypes struct {
	String       string  `toml:"string"`
	StringEmpty  string  `toml:"string_empty"`
	StringQuoted string  `toml:"string_quoted"`
	Int          int     `toml:"int"`
	IntNeg       int     `toml:"int_neg"`
	Float        float64 `toml:"float"`
	FloatNeg     float64 `toml:"float_neg"`
	Bool         bool    `toml:"bool"`
}

type StringVariations struct {
	Basic       string `toml:"basic"`
	WithSpaces  string `toml:"with_spaces"`
	WithQuotes  string `toml:"with_quotes"`
	WithEscapes string `toml:"with_escapes"`
	WithTabs    string `toml:"with_tabs"`
	WithUnicode string `toml:"with_unicode"`
	Empty       string `toml:"empty"`
}

type ArrayTypes struct {
	Strings    []string      `toml:"strings"`
	Ints       []int         `toml:"ints"`
	Floats     []float64     `toml:"floats"`
	Bools      []bool        `toml:"bools"`
	Mixed      []interface{} `toml:"mixed"`
	Empty      []int         `toml:"empty"`
	NestedInts [][]int       `toml:"nested_ints"`
	NestedMix  []interface{} `toml:"nested_mix"`
}

type NestedGroups struct {
	One       string `toml:"one.value"`
	TwoA      string `toml:"one.two.a"`
	TwoB      int    `toml:"one.two.b"`
	ThreeA    bool   `toml:"one.two.three.a"`
	ThreeB    string `toml:"one.two.three.b"`
	FourArray []int  `toml:"one.two.three.four.array"`
}

// Test functions organized by feature
func TestBasicTypes(t *testing.T) {
	input := `
string = "basic string"
string_empty = ""
string_quoted = "string with \"quotes\""
int = 42
int_neg = -42
float = 3.14159
float_neg = -3.14159
bool = true
`
	var config BasicTypes
	if err := Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("Failed to unmarshal basic types: %v", err)
	}

	expected := BasicTypes{
		String:       "basic string",
		StringEmpty:  "",
		StringQuoted: "string with \"quotes\"",
		Int:          42,
		IntNeg:       -42,
		Float:        3.14159,
		FloatNeg:     -3.14159,
		Bool:         true,
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Basic types mismatch:\nGot: %+v\nWant: %+v", config, expected)
	}

	// Test marshal/unmarshal roundtrip
	output, err := Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal basic types: %v", err)
	}

	var decoded BasicTypes
	if err := Unmarshal(output, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal marshaled basic types: %v", err)
	}

	if !reflect.DeepEqual(decoded, expected) {
		t.Errorf("Basic types roundtrip mismatch:\nGot: %+v\nWant: %+v", decoded, expected)
	}
}

func TestStringVariations(t *testing.T) {
	input := `
basic = "simple string"
with_spaces = "string with spaces and  multiple   spaces"
with_quotes = "string with \"double\" and 'single' quotes"
with_escapes = "escaped\\backslash\\and\"quotes"
with_tabs = "tabs\tand\tmore\ttabs"
with_unicode = "unicode chars: 测试 UTF-8"
empty = ""
`
	var config StringVariations
	if err := Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("Failed to unmarshal string variations: %v", err)
	}

	expected := StringVariations{
		Basic:       "simple string",
		WithSpaces:  "string with spaces and  multiple   spaces",
		WithQuotes:  "string with \"double\" and 'single' quotes",
		WithEscapes: "escaped\\backslash\\and\"quotes",
		WithTabs:    "tabs\tand\tmore\ttabs",
		WithUnicode: "unicode chars: 测试 UTF-8",
		Empty:       "",
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("String variations mismatch:\nGot: %+v\nWant: %+v", config, expected)
	}
}

func TestArrayTypes(t *testing.T) {
	input := `
strings = ["one", "two", "three", "with spaces", "with\"quotes", "with\\escape"]
ints = [1, 2, 3, -4, 5, 0]
floats = [1.1, 2.2, -3.3, 0.0, 123.456]
bools = [true, false, true, true]
mixed = ["string", 42, 3.14, true, [1, 2, 3]]
empty = []
nested_ints = [[1, 2], [3, 4], [5, 6]]
nested_mix = [[1, "two"], [true, 3.14], ["nested", ["deep", 42]]]
`
	var config ArrayTypes
	if err := Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("Failed to unmarshal array types: %v", err)
	}

	expected := ArrayTypes{
		Strings:    []string{"one", "two", "three", "with spaces", "with\"quotes", "with\\escape"},
		Ints:       []int{1, 2, 3, -4, 5, 0},
		Floats:     []float64{1.1, 2.2, -3.3, 0.0, 123.456},
		Bools:      []bool{true, false, true, true},
		Mixed:      []interface{}{"string", int64(42), 3.14, true, []interface{}{int64(1), int64(2), int64(3)}},
		Empty:      []int{},
		NestedInts: [][]int{{1, 2}, {3, 4}, {5, 6}},
		NestedMix: []interface{}{
			[]interface{}{int64(1), "two"},
			[]interface{}{true, 3.14},
			[]interface{}{"nested", []interface{}{"deep", int64(42)}},
		},
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Array types mismatch:\nGot: %+v\nWant: %+v", config, expected)
	}
}

func TestNestedGroups(t *testing.T) {
	input := `
[one]
value = "root value"

[one.two]
a = "level two a"
b = 42

[one.two.three]
a = true
b = "level three b"

[one.two.three.four]
array = [1, 2, 3]
`
	var config NestedGroups
	if err := Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("Failed to unmarshal nested groups: %v", err)
	}

	expected := NestedGroups{
		One:       "root value",
		TwoA:      "level two a",
		TwoB:      42,
		ThreeA:    true,
		ThreeB:    "level three b",
		FourArray: []int{1, 2, 3},
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Nested groups mismatch:\nGot: %+v\nWant: %+v", config, expected)
	}
}

func TestCommentHandling(t *testing.T) {
	input := `
# Full line comment
string = "value"  # End of line comment

[group]  # Group comment
key = "value"  # Another comment

# Comment before group
[nested.group]  # Group with comment
# Comment before key
key = "value"  # Comment after value
`
	type CommentConfig struct {
		String string `toml:"string"`
		Key1   string `toml:"group.key"`
		Key2   string `toml:"nested.group.key"`
	}

	expected := CommentConfig{
		String: "value",
		Key1:   "value",
		Key2:   "value",
	}

	var config CommentConfig
	if err := Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("Failed to unmarshal with comments: %v", err)
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Comment handling mismatch:\nGot: %+v\nWant: %+v", config, expected)
	}
}

func TestWhitespaceHandling(t *testing.T) {
	input := `
   string = "value"
       indented = "value"
key="no-spaces"
  spaces   =   "value"   
[group]
 key = "value"
   nested   =   "value"  
`
	type WhitespaceConfig struct {
		String   string `toml:"string"`
		Indented string `toml:"indented"`
		NoSpace  string `toml:"key"`
		Spaces   string `toml:"spaces"`
		Key      string `toml:"group.key"`
		Nested   string `toml:"group.nested"`
	}

	expected := WhitespaceConfig{
		String:   "value",
		Indented: "value",
		NoSpace:  "no-spaces",
		Spaces:   "value",
		Key:      "value",
		Nested:   "value",
	}

	var config WhitespaceConfig
	if err := Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("Failed to unmarshal with whitespace variations: %v", err)
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Whitespace handling mismatch:\nGot: %+v\nWant: %+v", config, expected)
	}
}

func TestDuplicateHandling(t *testing.T) {
	input := `
string = "first"
string = "second"
[group]
key = "first"
key = "second"
[group.nested]
key = "first"
key = "second"
`
	type DuplicateConfig struct {
		String string `toml:"string"`
		Key1   string `toml:"group.key"`
		Key2   string `toml:"group.nested.key"`
	}

	expected := DuplicateConfig{
		String: "first",
		Key1:   "first",
		Key2:   "first",
	}

	var config DuplicateConfig
	if err := Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("Failed to unmarshal with duplicates: %v", err)
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Duplicate handling mismatch:\nGot: %+v\nWant: %+v", config, expected)
	}
}

func TestEdgeCases(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected interface{}
		config   interface{}
	}{
		{
			name:     "empty input",
			input:    "",
			config:   &struct{}{},
			expected: &struct{}{},
		},
		{
			name: "only comments",
			input: `# Comment 1
                  # Comment 2
                  # Comment 3`,
			config:   &struct{}{},
			expected: &struct{}{},
		},
		{
			name: "empty groups",
			input: `[group1]
                  [group2]
                  [group.nested]`,
			config: &struct {
				Value string `toml:"group1.value"`
			}{},
			expected: &struct {
				Value string `toml:"group1.value"`
			}{},
		},
		{
			name: "special characters in strings",
			input: `
special = "tab\t tab"
quotes = "\"quotes' quotes\""
backslash = "back\\slash"
mixed = "tab\t\"quote'\\mixed"`,
			config: &struct {
				Special   string `toml:"special"`
				Quotes    string `toml:"quotes"`
				Backslash string `toml:"backslash"`
				Mixed     string `toml:"mixed"`
			}{},
			expected: &struct {
				Special   string `toml:"special"`
				Quotes    string `toml:"quotes"`
				Backslash string `toml:"backslash"`
				Mixed     string `toml:"mixed"`
			}{
				Special:   "tab\t tab",
				Quotes:    "\"quotes' quotes\"",
				Backslash: "back\\slash",
				Mixed:     "tab\t\"quote'\\mixed",
			},
		},
		{
			name: "number variations",
			input: `
int = 0
neg = -0
big = 999999
small = -999999
float = 0.0
float_neg = -0.0
float_big = 123.456
float_small = -123.456`,
			config: &struct {
				Int        int     `toml:"int"`
				Neg        int     `toml:"neg"`
				Big        int     `toml:"big"`
				Small      int     `toml:"small"`
				Float      float64 `toml:"float"`
				FloatNeg   float64 `toml:"float_neg"`
				FloatBig   float64 `toml:"float_big"`
				FloatSmall float64 `toml:"float_small"`
			}{},
			expected: &struct {
				Int        int     `toml:"int"`
				Neg        int     `toml:"neg"`
				Big        int     `toml:"big"`
				Small      int     `toml:"small"`
				Float      float64 `toml:"float"`
				FloatNeg   float64 `toml:"float_neg"`
				FloatBig   float64 `toml:"float_big"`
				FloatSmall float64 `toml:"float_small"`
			}{
				Int:        0,
				Neg:        0,
				Big:        999999,
				Small:      -999999,
				Float:      0.0,
				FloatNeg:   0.0,
				FloatBig:   123.456,
				FloatSmall: -123.456,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := Unmarshal([]byte(tc.input), tc.config); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if !reflect.DeepEqual(tc.config, tc.expected) {
				t.Errorf("Mismatch:\nGot: %+v\nWant: %+v", tc.config, tc.expected)
			}
		})
	}
}

func TestErrorCases(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		config  interface{}
		wantErr string
	}{
		{
			name:    "nil pointer",
			input:   "key = \"value\"",
			config:  nil,
			wantErr: "decode target must be a non-nil pointer",
		},
		{
			name:    "non-pointer",
			input:   "key = \"value\"",
			config:  struct{}{},
			wantErr: "decode target must be a non-nil pointer",
		},
		{
			name:    "non-struct pointer",
			input:   "key = \"value\"",
			config:  new(string),
			wantErr: "decode target must be a struct",
		},
		{
			name:    "invalid group format",
			input:   "[invalid",
			config:  &struct{}{},
			wantErr: "invalid group format",
		},
		{
			name:    "empty group name",
			input:   "[]",
			config:  &struct{}{},
			wantErr: "empty group name",
		},
		{
			name:    "invalid key format",
			input:   "invalid",
			config:  &struct{}{},
			wantErr: "invalid key-value format",
		},
		{
			name:    "empty value",
			input:   "key = ",
			config:  &struct{}{},
			wantErr: "empty value",
		},
		{
			name:    "unterminated string",
			input:   "key = \"unterminated",
			config:  &struct{}{},
			wantErr: "invalid value format",
		},
		{
			name:    "invalid escape",
			input:   "key = \"invalid\\x\"",
			config:  &struct{}{},
			wantErr: "invalid value format",
		},
		{
			name:    "unquoted string with space",
			input:   "key = has space",
			config:  &struct{}{},
			wantErr: "unquoted value contains whitespace",
		},
		{
			name:  "integer overflow",
			input: "num = 9223372036854775808", // MaxInt64 + 1
			config: &struct {
				Num int64 `toml:"num"`
			}{},
			wantErr: "integer overflow",
		},
		{
			name:  "invalid array element",
			input: "arr = [1, \"string\", true]",
			config: &struct {
				Arr []int `toml:"arr"`
			}{},
			wantErr: "array element must be number",
		},
		{
			name:  "invalid float format",
			input: "num = 123.",
			config: &struct {
				Num float64 `toml:"num"`
			}{},
			wantErr: "invalid float format",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Unmarshal([]byte(tc.input), tc.config)
			if err == nil {
				t.Fatal("Expected error but got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Wrong error:\ngot:  %v\nwant: %v", err, tc.wantErr)
			}
		})
	}
}

func TestMarshalRoundtrip(t *testing.T) {
	type CompleteConfig struct {
		String string        `toml:"string"`
		Int    int           `toml:"int"`
		Float  float64       `toml:"float"`
		Bool   bool          `toml:"bool"`
		Array  []string      `toml:"array"`
		Mixed  []interface{} `toml:"mixed"`
		Nested string        `toml:"group.nested"`
	}

	original := CompleteConfig{
		String: "test \"string\"",
		Int:    42,
		Float:  3.14159,
		Bool:   true,
		Array:  []string{"one", "two", "three"},
		Mixed:  []interface{}{"string", int64(42), 3.14, true},
		Nested: "nested value",
	}

	// Marshal
	data, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded CompleteConfig
	if err := Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Compare
	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("Roundtrip mismatch:\nGot:  %+v\nWant: %+v", decoded, original)
	}

	// Test MarshalIndent
	indented, err := MarshalIndent(original)
	if err != nil {
		t.Fatalf("MarshalIndent failed: %v", err)
	}

	var indentDecoded CompleteConfig
	if err := Unmarshal(indented, &indentDecoded); err != nil {
		t.Fatalf("Unmarshal of indented failed: %v", err)
	}

	if !reflect.DeepEqual(original, indentDecoded) {
		t.Errorf("Indented roundtrip mismatch:\nGot:  %+v\nWant: %+v", indentDecoded, original)
	}
}
