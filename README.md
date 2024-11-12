# TinyTOML

A minimal TOML parser and encoder for Go. TinyTOML provides a lightweight implementation focusing on the most commonly used TOML features while maintaining strict parsing rules.

The package is designed with simplicity and predictability in mind, to be used for the small/medium size app configuration file management. The limitations and deviations from TOML specs are outlined under the implementation details section below.

## Features

- Basic TOML types support:
  - Strings (bare and quoted)
  - Numbers (integers and floats)
  - Booleans
  - Arrays (homogeneous type only)
- Table/group nesting with dot notation
- Struct-based configuration handling with TOML tags
- Basic string escape sequences (`\"`, `\t`, `\n`, `\r`, `\\`)
- Comment support (both inline `#` and full-line)
- Flexible whitespace handling
- Strict parsing rules:
  - Integer overflow detection
  - Number format validation
  - String format validation
  - Duplicate key detection (first occurrence wins)

## Installation

```bash
go get github.com/LixenWraith/tinytoml
```

## Implementation Details

### Spec Deviations

- Arrays must be homogeneous (single type elements only)
- First occurrence of a key wins, subsequent duplicates ignored
- Table headers are merged, not overwritten
- Numbers with multiple dots are parsed as strings

### Implementation Choices

- Follows encoding/json-style interface for Marshal/Unmarshal
- Supports any type for stroage and struct tags
- Recursive handling of nested structures
- Type conversion follows Go's standard rules
- Keys must contain only ASCII letters, digits, underscore, and hyphen
- String values must be quoted if they contain spcee or any of special characters `,#=[]`
- String values can be unquoted if they contain only ASCII letters, digits, underscore, and hyphen
- Escape sequences in strings: `\", \t, \n, \r, \\`
- Integers checked for int64 bounds

### Unsupported Features

- Date/time formats
- Hex/octal/binary numbers
- Scientific notation
- Multi-line strings
- Inline tables
- Array of tables
- Unicode escapes
- +/- inf and nan floats

## Usage

### Roundtrip conversion

```go
package main


import (
    "fmt"
    "log"
    "github.com/LixenWraith/tinytoml"
)

func main() {
    input := `
# Basic key/values

name = "MyApp"
port = 8080
debug = true

# Arrays
hosts = ["localhost", "backup.local"]

# Tables
[database]
host = "localhost"
port = 5432


# Nested tables
[services.cache]
host = "redis.local"
port = 6379
`

    // Parse TOML
    var data any
    if err := tinytoml.Unmarshal([]byte(input), &data); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Parsed data:\n%+v\n\n", data)

    // Marshal back to TOML
    output, err := tinytoml.Marshal(data)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\nGenerated TOML:\n%s", output)
}
```

### Practical Example: Configuration Management

TinyTOML can be used effectively for application configuration management. Here's a practical example of managing server configuration:

```go
type ServerConfig struct {
    Server struct {
        Host string `toml:"host"`
        Port int    `toml:"port"`
    } `toml:"server"`
    
    Database struct {
        Host     string `toml:"host"`
        Port     int    `toml:"port"`
        User     string `toml:"user"`
        Password string `toml:"password"`
    } `toml:"database"`
}

func main() {
    // Load or create config
    if data, err := os.ReadFile("config.toml"); err == nil {
        // Load existing config and merge with defaults
        if err := tinytoml.Unmarshal(data, &config); err != nil {
            log.Fatal(err)
        }
    } else {
        // Create new config with defaults
        config = getDefaultConfig()
    }

    // Save complete config
    data, _ := tinytoml.Marshal(config)
    os.WriteFile("config.toml", data, 0644)
}
```

This pattern creates, or loads and rewrites a config file, a config file like:

```toml
[server]
host = "localhost"
port = 8080

[database]
host = "localhost"
port = 5432
user = "dbuser"
password = "dbpass"
```

See [examples/roundtrip/main.go] and [examples/default_config/main.go] and [examples/comparison_json/main.go] for complete working examples.

## API

### `Unmarshal(data []byte, v any) error`
Parses TOML data into a the provided type (struct or map)

### `Marshal(v any) ([]byte, error)`
Converts a Go struct to TOML format

## Error Handling

TinyTOML provides detailed error messages:

```go
var err error

err = tinytoml.Unmarshal([]byte(input), &data)

output, err = tinytoml.Marshal(data)
```

```
2024/11/11 16:16:24 Unmarshal error: Unmarshal: parser.parse: parser.parseLine: parser.parseArrayElements: type mismatch in array [element 1] [key "key"] [line 2]
2024/11/11 16:16:24 Marshal error: Marshal: anyToMap: unsupported type
```

## License

MIT License - see LICENSE file

