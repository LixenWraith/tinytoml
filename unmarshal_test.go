package tinytoml

import (
	"reflect"
	"strings"
	"testing"
)

func TestUnmarshal_SingleValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     map[string]any
		wantErr  bool
		errormsg string
	}{
		{
			name:     "simple string",
			input:    `name = "value"`,
			want:     map[string]any{"name": "value"},
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "integer value",
			input:    "count = 42",
			want:     map[string]any{"count": int64(42)},
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "integer value",
			input:    "count = +42",
			want:     map[string]any{"count": int64(42)},
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "boolean value",
			input:    "active = true",
			want:     map[string]any{"active": true},
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "float value",
			input:    "price = -19.99",
			want:     map[string]any{"price": -19.99},
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "float value",
			input:    "price = +19.99",
			want:     map[string]any{"price": 19.99},
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "empty input",
			input:    "",
			want:     nil,
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "bad float",
			input:    `bad_float = 12.5e9`,
			want:     map[string]any{"name": "value"},
			wantErr:  true,
			errormsg: "",
		},
		{
			name:     "bad integer",
			input:    `bad_int = -129 9`,
			want:     map[string]any{"name": "value"},
			wantErr:  true,
			errormsg: errInvalidFormat,
		},
		{
			name:     "invalid format",
			input:    "invalid line without equals",
			want:     nil,
			wantErr:  true,
			errormsg: errInvalidFormat,
		},
		{
			name:     "invalid key",
			input:    "123invalid = \"value\"",
			want:     nil,
			wantErr:  true,
			errormsg: errInvalidKey,
		},
		{
			name:     "missing value",
			input:    "key = ",
			want:     nil,
			wantErr:  true,
			errormsg: errMissingValue,
		},
		{
			name: "multiple key-value pairs",
			input: `name = "value"
                    count = 42
                    active = true`,
			want: map[string]any{
				"name":   "value",
				"count":  int64(42),
				"active": true,
			},
			wantErr: false,
		},
		{
			name: "with comments and empty lines",
			input: `
                    # This is a comment
                    name = "value"
                    
                    count = 42  # inline comment
                    active = true`,
			want: map[string]any{
				"name":   "value",
				"count":  int64(42),
				"active": true,
			},
			wantErr: false,
		},
		{
			name: "with escape sequences",
			input: `message = "line1\nline2\tindented\r\n"
                    path = "C:\\Program Files\\App"`,
			want: map[string]any{
				"message": "line1\nline2\tindented\r\n",
				"path":    "C:\\Program Files\\App",
			},
			wantErr: false,
		},
		{
			name:     "invalid escape sequence",
			input:    `message = "invalid\k"`,
			want:     nil,
			wantErr:  true,
			errormsg: errInvalidEscape,
		},
		{
			name:     "valid string array",
			input:    `files = ["readme.txt", "operation.log", "data1234.txt"]`,
			want:     map[string]any{"files": []any{"readme.txt", "operation.log", "data1234.txt"}},
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "valid number array",
			input:    "numbers = [1, 2, 3, 4, 5]",
			want:     map[string]any{"numbers": []any{int64(1), int64(2), int64(3), int64(4), int64(5)}},
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "mixed type array",
			input:    `mixed = [1, "two", 3]`,
			want:     map[string]any{"mixed": []any{int64(1), "two", int64(3)}},
			wantErr:  false,
			errormsg: "",
		},
		{
			name: "error: mixed type array",
			input: `mixed = [1, "two", 3]
number = 192837194
name = "value"`,
			want:     map[string]any{"mixed": []any{int64(1), "two", int64(3)}, "name": "value", "number": int64(192837194)},
			wantErr:  false,
			errormsg: "",
		},
		{
			name:     "error: invalid array syntax",
			input:    "invalid = [1, 2, 3",
			want:     nil,
			wantErr:  true,
			errormsg: errUnterminatedArray,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got any
			err := Unmarshal([]byte(tt.input), &got)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Unmarshal() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.errormsg) {
					t.Errorf("Unmarshal() error = %v, want error containing %v", err, tt.errormsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want == nil && got != nil {
				t.Errorf("Unmarshal() = %v, want nil", got)
				return
			}

			if tt.want != nil {
				gotMap, ok := got.(map[string]any)
				if !ok {
					t.Errorf("Unmarshal() result is not a map[string]any, got %T", got)
					return
				}

				if !reflect.DeepEqual(gotMap, tt.want) {
					t.Errorf("Unmarshal() = %v, want %v", gotMap, tt.want)
				}
			}
		})
	}
}

func TestUnmarshalTables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]any
		wantErr  bool
		errormsg string
	}{
		{
			name: "basic table",
			input: `[server]
name = "web"
port = 8080`,
			expected: map[string]any{
				"server": map[string]any{
					"name": "web",
					"port": int64(8080),
				},
			},
			wantErr: false,
		},
		{
			name: "multiple tables",
			input: `[server]
name = "web"

[database]
host = "localhost"`,
			expected: map[string]any{
				"server": map[string]any{
					"name": "web",
				},
				"database": map[string]any{
					"host": "localhost",
				},
			},
			wantErr: false,
		},
		{
			name: "table with comments",
			input: `[server] # Web Server Config
name = "web" # primary server`,
			expected: map[string]any{
				"server": map[string]any{
					"name": "web",
				},
			},
			wantErr: false,
		},
		{
			name: "table with whitespace",
			input: `[    server    ]
name = "web"`,
			expected: map[string]any{
				"server": map[string]any{
					"name": "web",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid table name with spaces",
			input: `[server name]
host = "localhost"`,
			wantErr:  true,
			errormsg: errInvalidTableName,
		},
		{
			name: "empty table name",
			input: `[]
name = "test"`,
			wantErr:  true,
			errormsg: errInvalidTableName,
		},
		{
			name: "duplicate tables merge",
			input: `[server]
name = "web1"
port = 8080

[server]
name = "web2"`,
			expected: map[string]any{
				"server": map[string]any{
					"name": "web2",
					"port": int64(8080),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid table syntax",
			input: `[server
name = "web"`,
			wantErr:  true,
			errormsg: errInvalidFormat,
		},
		{
			name: "comment-only table line",
			input: `[server] # comment
# another comment
name = "web"`,
			expected: map[string]any{
				"server": map[string]any{
					"name": "web",
				},
			},
			wantErr: false,
		},
		{
			name: "quoted strings in comments",
			input: `[server] # don't "parse" this
name = "web" # or "this" one`,
			expected: map[string]any{
				"server": map[string]any{
					"name": "web",
				},
			},
			wantErr: false,
		},
		{
			name: "tables with empty lines and comments",
			input: `
# header comment
[server]

# mid comment
name = "web"

# another comment

[database]
host = "localhost"
`,
			expected: map[string]any{
				"server": map[string]any{
					"name": "web",
				},
				"database": map[string]any{
					"host": "localhost",
				},
			},
			wantErr: false,
		},
		{
			name: "basic dotted table",
			input: `[server.network]
ip = "1.2.3.4"
port = 8080`,
			expected: map[string]any{
				"server": map[string]any{
					"network": map[string]any{
						"ip":   "1.2.3.4",
						"port": int64(8080),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple nested tables",
			input: `[server.network]
ip = "1.2.3.4"

[server.config]
timeout = 30

[server.network.ssl]
enabled = true`,
			expected: map[string]any{
				"server": map[string]any{
					"network": map[string]any{
						"ip": "1.2.3.4",
						"ssl": map[string]any{
							"enabled": true,
						},
					},
					"config": map[string]any{
						"timeout": int64(30),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "merge dotted tables",
			input: `[server]
name = "main"

[server.network]
ip = "1.2.3.4"

[server]
id = "srv1"`,
			expected: map[string]any{
				"server": map[string]any{
					"name": "main",
					"id":   "srv1",
					"network": map[string]any{
						"ip": "1.2.3.4",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "deep nesting",
			input: `[a.b.c.d.e]
value = true`,
			expected: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": map[string]any{
							"d": map[string]any{
								"e": map[string]any{
									"value": true,
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "dotted key nested table",
			input: `[server]
network.ip = "1.1.1.1"`,
			expected: map[string]any{
				"server": map[string]any{
					"network": map[string]any{
						"ip": "1.1.1.1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid empty segment",
			input: `[server..network]
ip = "1.2.3.4"`,
			wantErr:  true,
			errormsg: errInvalidTableName,
		},
		{
			name: "invalid segment name",
			input: `[server.123network]
ip = "1.2.3.4"`,
			wantErr:  true,
			errormsg: errInvalidTableName,
		},
		{
			name: "dotted name with spaces",
			input: `[server. network]
ip = "1.2.3.4"`,
			wantErr:  true,
			errormsg: errInvalidTableName,
		},
		{
			name: "out of order table definition",
			input: `[server.network.ssl]
enabled = true

[server]
name = "main"

[server.network]
ip = "1.2.3.4"`,
			expected: map[string]any{
				"server": map[string]any{
					"name": "main",
					"network": map[string]any{
						"ip": "1.2.3.4",
						"ssl": map[string]any{
							"enabled": true,
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got map[string]any
			err := Unmarshal([]byte(tt.input), &got)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Unmarshal() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.errormsg) {
					t.Errorf("Unmarshal() error = %v, want error containing %v", err, tt.errormsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Unmarshal() = %v, want %v", got, tt.expected)
			}
		})
	}
}
