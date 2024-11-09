// File: tinytoml/src/pkg/tinytomll/tinytoml_test.go

package tinytoml

import (
	"reflect"
	"strings"
	"testing"
)

type TestConfig struct {
	Title    string  `toml:"title"`
	Port     int     `toml:"port"`
	Debug    bool    `toml:"debug"`
	Rate     float64 `toml:"rate"`
	Host     string  `toml:"server.host"`
	SSLPort  int     `toml:"server.ssl_port"`
	User     string  `toml:"db.user"`
	Password string  `toml:"db.password"`
	Tabs     string  `toml:"tabs"`
	Quoted   string  `toml:"quoted"`
	Escaped  string  `toml:"escaped"`
	SpaceVal string  `toml:"space_val"`
	UTF8     string  `toml:"utf8"`
}

func TestUnmarshal(t *testing.T) {
	input := `# Test configuration
title = "Test App"
port = 8080
debug = true
rate = 0.75
tabs = "value\twith\ttabs"
quoted = "quoted \"string\" value"
escaped = "escaped\\value"
space_val = "value with spaces"
utf8 = "测试 UTF-8"

[server]
host = "localhost"
ssl_port = 443  # Default SSL port

[db]
user = "admin"
password = "secret"`

	expected := TestConfig{
		Title:    "Test App",
		Port:     8080,
		Debug:    true,
		Rate:     0.75,
		Host:     "localhost",
		SSLPort:  443,
		User:     "admin",
		Password: "secret",
		Tabs:     "value\twith\ttabs",
		Quoted:   "quoted \"string\" value",
		Escaped:  "escaped\\value",
		SpaceVal: "value with spaces",
		UTF8:     "测试 UTF-8",
	}

	var config TestConfig
	if err := Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Unmarshal result mismatch:\nGot: %+v\nWant: %+v", config, expected)
	}
}

func TestMarshal(t *testing.T) {
	config := TestConfig{
		Title:    "Test App",
		Port:     8080,
		Debug:    true,
		Rate:     0.75,
		Host:     "localhost",
		SSLPort:  443,
		User:     "admin",
		Password: "secret",
		Tabs:     "value\twith\ttabs",
		Quoted:   "quoted \"string\" value",
		Escaped:  "escaped\\value",
		SpaceVal: "value with spaces",
		UTF8:     "测试 UTF-8",
	}

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded TestConfig
	if err := Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal marshaled data: %v", err)
	}

	if !reflect.DeepEqual(config, decoded) {
		t.Errorf("Marshal/Unmarshal roundtrip mismatch:\nGot: %+v\nWant: %+v", decoded, config)
	}
}

func TestMarshalIndent(t *testing.T) {
	config := TestConfig{
		Title:    "Test App",
		Port:     8080,
		Debug:    true,
		Rate:     0.75,
		Host:     "localhost",
		SSLPort:  443,
		User:     "admin",
		Password: "secret",
		Tabs:     "value\twith\ttabs",
		Quoted:   "quoted \"string\" value",
		Escaped:  "escaped\\value",
		SpaceVal: "value with spaces",
		UTF8:     "测试 UTF-8",
	}

	data, err := MarshalIndent(config)
	if err != nil {
		t.Fatalf("Failed to marshal with indent: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "[") {
			if i > 0 && lines[i-1] != "" {
				t.Errorf("Expected empty line before group at line %d", i+1)
			}
		}
	}

	var decoded TestConfig
	if err := Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal indented data: %v", err)
	}

	if !reflect.DeepEqual(config, decoded) {
		t.Errorf("MarshalIndent/Unmarshal roundtrip mismatch:\nGot: %+v\nWant: %+v", decoded, config)
	}
}

func TestErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		errLine  int
		errMatch string
	}{
		{
			name:    "Empty input",
			input:   "",
			wantErr: false,
		},
		{
			name:     "Invalid group format",
			input:    "[invalid",
			wantErr:  true,
			errLine:  1,
			errMatch: "invalid group format",
		},
		{
			name:     "Empty group",
			input:    "[]",
			wantErr:  true,
			errLine:  1,
			errMatch: "empty group name",
		},
		{
			name:     "Invalid group name",
			input:    "[123invalid]",
			wantErr:  true,
			errLine:  1,
			errMatch: "invalid group name",
		},
		{
			name:     "Duplicate group",
			input:    "[group]\nkey=1\n[group]\nkey=2",
			wantErr:  true,
			errLine:  3,
			errMatch: "duplicate group",
		},
		{
			name: "Excessive group nesting",
			input: `[a.b.c.d]
key = "value"`,
			wantErr:  true,
			errLine:  1,
			errMatch: "nesting exceeds maximum depth",
		},
		{
			name:     "Invalid key-value format",
			input:    "key",
			wantErr:  true,
			errLine:  1,
			errMatch: "invalid key-value format",
		},
		{
			name:     "Empty key",
			input:    "= value",
			wantErr:  true,
			errLine:  1,
			errMatch: "empty key",
		},
		{
			name:     "Invalid key name",
			input:    "123key = value",
			wantErr:  true,
			errLine:  1,
			errMatch: "invalid key",
		},
		{
			name:     "Empty value",
			input:    "key = ",
			wantErr:  true,
			errLine:  1,
			errMatch: "empty value",
		},
		{
			name:     "Invalid comment format",
			input:    "key = value#comment",
			wantErr:  true,
			errLine:  1,
			errMatch: "invalid comment format",
		},
		{
			name: "Comments only",
			input: `# Just a comment
# Another comment`,
			wantErr: false,
		},
		{
			name:     "Invalid number",
			input:    "number = 12.34.56",
			wantErr:  false,
			errLine:  1,
			errMatch: "",
		},
		{
			name:     "Integer overflow",
			input:    "number = 9223372036854775808",
			wantErr:  true,
			errLine:  1,
			errMatch: "integer overflow",
		},
		{
			name:     "Unterminated string",
			input:    `key = "unterminated`,
			wantErr:  true,
			errLine:  1,
			errMatch: "invalid value format",
		},
		{
			name:     "Invalid escape sequence",
			input:    `key = "invalid\x"`,
			wantErr:  true,
			errLine:  1,
			errMatch: "invalid value format",
		},
		{
			name:     "Invalid UTF-8",
			input:    "key = \xFF\xFF",
			wantErr:  true,
			errLine:  1,
			errMatch: "invalid value format",
		},
		{
			name:     "Unquoted string with space",
			input:    "key = has space",
			wantErr:  true,
			errLine:  1,
			errMatch: "unquoted value contains whitespace",
		},
		{
			name:     "Single quote in unquoted string",
			input:    "key = don't",
			wantErr:  true,
			errLine:  1,
			errMatch: "unquoted value contains whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config TestConfig
			err := Unmarshal([]byte(tt.input), &config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				if perr, ok := err.(*ParseError); ok {
					if tt.errLine > 0 && perr.Line != tt.errLine {
						t.Errorf("Wrong error line: got %d, want %d", perr.Line, tt.errLine)
					}
					if tt.errMatch != "" && !strings.Contains(perr.Error(), tt.errMatch) {
						t.Errorf("Error message doesn't match: got %q, want to contain %q", perr.Error(), tt.errMatch)
					}
				}
			}
		})
	}
}

func TestValueConversion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, TestConfig)
	}{
		{
			name: "Integer conversion",
			input: `port = 8080
server.ssl_port = 443`,
			check: func(t *testing.T, c TestConfig) {
				if c.Port != 8080 || c.SSLPort != 443 {
					t.Errorf("Integer conversion failed: port=%d, ssl_port=%d", c.Port, c.SSLPort)
				}
			},
		},
		{
			name:  "Boolean conversion",
			input: `debug = true`,
			check: func(t *testing.T, c TestConfig) {
				if !c.Debug {
					t.Error("Boolean conversion failed")
				}
			},
		},
		{
			name:  "Float conversion",
			input: `rate = 0.75`,
			check: func(t *testing.T, c TestConfig) {
				if c.Rate != 0.75 {
					t.Errorf("Float conversion failed: got %f, want 0.75", c.Rate)
				}
			},
		},
		{
			name: "String variations",
			input: `
title = "Test"
db.password = "secret"
space_val = "value with spaces"
tabs = "value\twith\ttabs"
quoted = "quoted \"string\" value"
escaped = "escaped\\value"
utf8 = "测试 UTF-8"`,
			check: func(t *testing.T, c TestConfig) {
				cases := []struct {
					got, want, field string
				}{
					{c.Title, "Test", "title"},
					{c.Password, "secret", "password"},
					{c.SpaceVal, "value with spaces", "space_val"},
					{c.Tabs, "value\twith\ttabs", "tabs"},
					{c.Quoted, "quoted \"string\" value", "quoted"},
					{c.Escaped, "escaped\\value", "escaped"},
					{c.UTF8, "测试 UTF-8", "utf8"},
				}
				for _, tc := range cases {
					if tc.got != tc.want {
						t.Errorf("String conversion failed for %s: got %q, want %q", tc.field, tc.got, tc.want)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config TestConfig
			if err := Unmarshal([]byte(tt.input), &config); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}
			tt.check(t, config)
		})
	}
}

func TestDuplicateKeys(t *testing.T) {
	input := `
key = "first"
key = "second"
[group]
nested = "first"
nested = "second"
`
	type DupConfig struct {
		Key    string `toml:"key"`
		Nested string `toml:"group.nested"`
	}

	var config DupConfig
	if err := Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if config.Key != "first" {
		t.Errorf("Duplicate key handling failed: got %q, want %q", config.Key, "first")
	}
	if config.Nested != "first" {
		t.Errorf("Duplicate nested key handling failed: got %q, want %q", config.Nested, "first")
	}
}
