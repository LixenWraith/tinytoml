package tinytoml

import (
	"bytes"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestMarshal(t *testing.T) {
	pc, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(pc).Name()

	type Simple struct {
		Name   string
		Count  int
		Active bool
	}

	type Nested struct {
		Info    Simple
		Tags    []string
		Numbers []int
	}

	type Complex struct {
		Basic    Simple
		Details  Nested
		Matrix   [][]int
		Settings map[string]any
	}

	tests := []struct {
		name     string
		input    any
		expected string
		wantErr  bool
		errormsg string
	}{
		{
			name: "marshal any value",
			input: map[string]any{
				"Foo":  "Ba\nr",
				"Bar":  true,
				"Baz":  12,
				"Jazz": 19.5924,
			},
			expected: "Bar = true\nBaz = 12\nFoo = \"Ba\\nr\"\nJazz = 19.5924\n",
			wantErr:  false,
			errormsg: "",
		},
		{
			name: "marshal sort values",
			input: map[string]string{
				"Foo": "Bar",
				"Bar": "Foo",
			},
			expected: "Bar = \"Foo\"\nFoo = \"Bar\"\n",
			wantErr:  false,
			errormsg: "",
		},
		{
			name: "marshal number value",
			input: map[string]any{
				"Fizz": 64,
			},
			expected: `Fizz = 64
`,
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "marshal empty map",
			input:    map[string]any{},
			expected: "",
			wantErr:  false,
			errormsg: "",
		},
		{
			name: "marshal nested map",
			input: map[string]any{
				"Foo": map[string]any{
					"Bar": "Baz",
					"Baz": "Bar",
				},
			},
			expected: "[Foo]\nBar = \"Baz\"\nBaz = \"Bar\"\n",
			wantErr:  false,
			errormsg: "",
		},
		{
			name: "marshal deep nested map",
			input: map[string]any{
				"Foo": map[string]any{
					"FooBar": map[string]any{
						"Bar": "Baz",
						"Baz": "Bar",
					},
				},
			},
			expected: "[Foo]\n[Foo.FooBar]\nBar = \"Baz\"\nBaz = \"Bar\"\n",
			wantErr:  false,
			errormsg: "",
		},
		{
			name: "marshal complex nested map",
			input: map[string]any{
				"Foo": map[string]any{
					"Fizz": 12,
					"Jazz": "Buzz",
					"FooBar": map[string]any{
						"Bar": "Baz",
						"Baz": "Bar",
					},
				},
			},
			expected: "[Foo]\nFizz = 12\nJazz = \"Buzz\"\n[Foo.FooBar]\nBar = \"Baz\"\nBaz = \"Bar\"\n",
			wantErr:  false,
			errormsg: "",
		},
		{
			name: "marshal simple struct",
			input: Simple{
				Name:   "test",
				Count:  42,
				Active: true,
			},
			expected: "Active = true\nCount = 42\nName = \"test\"\n",
			wantErr:  false,
			errormsg: "",
		},
		{
			name: "marshal simple array",
			input: map[string][]string{
				"Tags": {"one", "two", "three"},
			},
			expected: "Tags = [\"one\", \"two\", \"three\"]\n",
			wantErr:  false,
			errormsg: "",
		},
		{
			name: "marshal nested struct",
			input: Nested{
				Info: Simple{
					Name:   "nested",
					Count:  123,
					Active: false,
				},
				Tags:    []string{"tag1", "tag2"},
				Numbers: []int{1, 2, 3},
			},
			expected: "Numbers = [1, 2, 3]\nTags = [\"tag1\", \"tag2\"]\n[Info]\nActive = false\nCount = 123\nName = \"nested\"\n",
			wantErr:  false,
			errormsg: "",
		},
		{
			name: "marshal nested arrays",
			input: map[string][][]int{
				"Matrix": {{1, 2}, {3, 4}, {5, 6}},
			},
			expected: "Matrix = [[1, 2], [3, 4], [5, 6]]\n",
			wantErr:  false,
			errormsg: "",
		},
		{
			name: "marshal complex nested struct",
			input: Complex{
				Basic: Simple{
					Name:   "root",
					Count:  100,
					Active: true,
				},
				Details: Nested{
					Info: Simple{
						Name:   "detail",
						Count:  50,
						Active: false,
					},
					Tags:    []string{"important", "critical"},
					Numbers: []int{10, 20, 30},
				},
				Matrix: [][]int{{1, 2}, {3, 4}},
				Settings: map[string]any{
					"Enabled": true,
					"Port":    8080,
				},
			},
			expected: "Matrix = [[1, 2], [3, 4]]\n[Basic]\nActive = true\nCount = 100\nName = \"root\"\n[Details]\nNumbers = [10, 20, 30]\nTags = [\"important\", \"critical\"]\n[Details.Info]\nActive = false\nCount = 50\nName = \"detail\"\n[Settings]\nEnabled = true\nPort = 8080\n",
			wantErr:  false,
			errormsg: "",
		},
		{
			name: "marshal struct with TOML tags and nesting",
			input: struct {
				Basic struct {
					Name    string `toml:"name"`
					Ignored string `toml:"-"`
					Value   int
				} `toml:"basic_config"`
				Details struct {
					Tags    []string `toml:"tag_list"`
					Active  bool     `toml:"enabled"`
					Hidden  bool     `toml:"-"`
					Default string
				} `toml:"detail_config"`
			}{
				Basic: struct {
					Name    string `toml:"name"`
					Ignored string `toml:"-"`
					Value   int
				}{
					Name:    "test",
					Ignored: "hidden",
					Value:   42,
				},
				Details: struct {
					Tags    []string `toml:"tag_list"`
					Active  bool     `toml:"enabled"`
					Hidden  bool     `toml:"-"`
					Default string
				}{
					Tags:    []string{"a", "b", "c"},
					Active:  true,
					Hidden:  true,
					Default: "default",
				},
			},
			expected: `[basic_config]
name = "test"
Value = 42
[detail_config]
Default = "default"
enabled = true
tag_list = ["a", "b", "c"]
`,
			wantErr:  false,
			errormsg: "",
		},
		{
			name: "marshal mixed array",
			input: map[string]any{
				"Mixed": []any{"string", 42, 3.14, true},
			},
			expected: "Mixed = [\"string\", 42, 3.14, true]\n",
			wantErr:  false,
			errormsg: "",
		},
		{
			name: "marshal array with unsupported type",
			input: map[string]any{
				"Invalid": []any{"string", map[string]string{"key": "value"}},
			},
			expected: "",
			wantErr:  true,
			errormsg: errUnsupported,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := Marshal(test.input)

			/*
				fmt.Printf("got  bytes: %v\n", []byte(result))
				fmt.Printf("want bytes: %v\n", []byte(test.expected))
			*/

			if test.wantErr {
				if err == nil {
					t.Errorf("-- %s failed: want error but got none.\n- input: %v\n- want: %s\n- got : %s\n\n", fn, test.input, test.expected, result)
					return
				}

				if !strings.Contains(err.Error(), test.errormsg) {
					t.Errorf("-- %s failed: got wrong error.\n- input: %v\n- want: %s\n- got: %s\n- error: %s\n\n", fn, test.input, test.expected, result, err.Error())
					return
				}
				return
			}

			if err != nil {
				t.Errorf("-- %s failed: want no error but got one.\n- input: %v\n- want: %s\n- got : %s\n- error: %s\n\n", fn, test.input, test.expected, result, err.Error())
				return
			}

			if string(result) != test.expected {
				t.Errorf("-- %s failed: wrong result.\n- input: %v\n- want: %s\n- got: %s\n\n", fn, test.input, test.expected, result)
				return
			}
		})
	}
}

func Test_isUnsupportedTypeError(t *testing.T) {
	pc, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(pc).Name()

	tests := []struct {
		name     string
		input    any
		expected string
		wantErr  bool
		errormsg string
	}{
		{
			name:     "marshal nil",
			input:    nil,
			expected: "",
			wantErr:  true,
			errormsg: errNilValue,
		},
		{
			name:     "marshal unsupported type",
			input:    make(chan int),
			expected: "",
			wantErr:  true,
			errormsg: errUnsupported,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := Marshal(test.input)

			if test.wantErr {
				if err == nil {
					t.Errorf("-- %s failed: want error but got none.\n- input: %v\n- want: %s\n- got : %s\n\n", fn, test.input, test.expected, result)
					return
				}

				if !strings.Contains(err.Error(), test.errormsg) {
					t.Errorf("-- %s failed: got wrong error.\n- input: %v\n- want: %s\n- got: %s\n- error: %s\n\n", fn, test.input, test.expected, result, err.Error())
					return
				}
				return
			}

			if err != nil {
				t.Errorf("-- %s failed: want no error but got one.\n- input: %v\n- want: %s\n- got : %s\n- error: %s\n\n", fn, test.input, test.expected, result, err.Error())
				return
			}

			if string(result) != test.expected {
				t.Errorf("-- %s failed: wrong result.\n- input: %v\n- want: %s\n- got: %s\n\n", fn, test.input, test.expected, result)
				return
			}
		})
	}
}

func Test_marshaller_marshalString(t *testing.T) {
	pc, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(pc).Name()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
		errormsg string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "\"hello\"",
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "string with escape characters",
			input:    "hello\tworld\n",
			expected: "\"hello\\tworld\\n\"",
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "string with quotes",
			input:    `hello "world"`,
			expected: `"hello \"world\""`,
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "string with backslash",
			input:    `hello\world`,
			expected: `"hello\\world"`,
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "string with multiple escapes",
			input:    "tab:\t newline:\n quote:\" backslash:\\",
			expected: "\"tab:\\t newline:\\n quote:\\\" backslash:\\\\\"",
			wantErr:  false,
			errormsg: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := &marshaller{
				buffer: &bytes.Buffer{},
				path:   []string{},
				depth:  0,
			}

			err := m.marshalString(reflect.ValueOf(test.input))

			result := m.buffer.String()

			if test.wantErr {
				if err == nil {
					t.Errorf("-- %s failed: want error but got none.\n- input: %v\n- want: %s\n- got : %s\n\n", fn, test.input, test.expected, result)
					return
				}

				if !strings.Contains(err.Error(), test.errormsg) {
					t.Errorf("-- %s failed: got wrong error.\n- input: %v\n- want: %s\n- got: %s\n- error: %s\n\n", fn, test.input, test.expected, result, err.Error())
					return
				}
				return
			}

			if err != nil {
				t.Errorf("-- %s failed: want no error but got one.\n- input: %v\n- want: %s\n- got : %s\n- error: %s\n\n", fn, test.input, test.expected, result, err.Error())
				return
			}

			if string(result) != test.expected {
				t.Errorf("-- %s failed: wrong result.\n- input: %v\n- want: %s\n- got: %s\n\n", fn, test.input, test.expected, result)
				return
			}
		})
	}
}
