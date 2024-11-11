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

## Usage

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


    // Parse TOML into map
    var config map[string]any
    if err := tinytoml.Unmarshal([]byte(input), &config); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%+v\n", config)

    // Marshal back to TOML
    output, err := tinytoml.Marshal(config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\nGenerated TOML:\n%s", output)
}
```

## Implementation Details

### Spec Deviations

- Arrays must be homogeneous (single type elements only)
- First occurrence of a key wins, subsequent duplicates ignored
- Table headers are merged, not overwritten
- Numbers with multiple dots are parsed as strings

### Implementation Choices

- Keys must contain only ASCII letters, digits, underscore, and hyphen
- String values must be quoted if they contain spcee or any of special characters `,#=[]`
- String values can be unquoted if they contain only ASCII letters, digits, underscore, and hyphen
- Escape sequences in strings: `\", \t, \n, \r, \\`
- Integers checked for int64 bounds
- During Marshal, map[string]interface{} with no entries produces no output
- MarshalIndent uses 4-space indentation for arrays

### Unsupported Features

- Date/time formats
- Hex/octal/binary numbers
- Scientific notation
- Multi-line strings
- Inline tables
- Array of tables
- Unicode escapes
- +/- inf and nan floats

## API

### `Unmarshal(data []byte, v any) error`
Parses TOML data into a map[string]interface{}

### `Marshal(v any) ([]byte, error)`
Converts a Go value to TOML format

### `MarshalIndent(v any) ([]byte, error)`
Like Marshal but adds consistent indentation for improved readability

## Error Handling


TinyTOML provides detailed error messages:

```go
if err := tinytoml.Unmarshal(data, &config); err != nil {
    fmt.Printf("Error: %v\n", err)
}
```

## License

MIT License - see LICENSE file

