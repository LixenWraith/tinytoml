# TinyTOML

A minimal TOML parser and encoder for Go that focuses on common configuration needs while maintaining strict TOML compatibility within its supported feature set.

## Features

- Basic TOML types:
  - Strings with escape sequences (\n, \t, \r, \\)
  - Numbers (integers and floats, with sign support)
  - Booleans
  - Arrays (homogeneous, nested, and mixed-type)
- Tables with dot notation
- Dotted keys within tables
- Table merging (last value wins)
- Struct tags (`toml:`) for custom field names
- Comment handling (inline and full-line)
- Flexible whitespace handling
- Type conversion following Go's standard rules
- Strict parsing rules with detailed error messages

## Installation

```bash
go get github.com/LixenWraith/tinytoml
```

## Implementation Details

### Limitations

- No support for:
  - Table arrays
  - Hex/octal/binary/exponential number formats
  - Multi-line keys or strings
  - Inline table declarations
  - Inline array declarations within tables
  - Empty table declarations
  - Datetime types
  - Unicode escape sequences
  - Key character escaping
  - Literal strings (single quotes)
  - Comments are discarded in parse, not supported in encode

### Implementation Choices

- Follows encoding/json-style interface for Marshal/Unmarshal
- Maps must have string keys
- Keys must start with letter/underscore, followed by letters/numbers/dashes/underscores
- Strings are always double-quoted
- Recursive handling of nested structures
- Integer bounds checking
- Float format validation
- Detailed error reporting

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/LixenWraith/tinytoml"
)

func main() {
    input := `
name = "Complex \nApp"
debug = true
workers = 42
rate = 3.14

# Array examples
ports = [8080, -6379]
hosts = ["local host", "bare_host"]

[server]
host = "localhost"
port = 8080

[database.primary]
host = "db1"
host.ip = "1.1.1.1"
host.port = 5432`

    var data any
    if err := tinytoml.Unmarshal([]byte(input), &data); err != nil {
        log.Fatal(err)
    }

    output, err := tinytoml.Marshal(data)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Generated TOML:\n%s", output)
}
```

### Configuration Management

TinyTOML is particularly useful for application configuration. Here's a typical usage pattern:

```go
type ServerConfig struct {
    Server struct {
        Host string `toml:"host"`
        Port int    `toml:"port"`
        Name string `toml:"name"`
    } `toml:"server"`
    
    Database struct {
        Host     string `toml:"host"`
        Port     int    `toml:"port"`
        User     string `toml:"user"`
        Password string `toml:"password"`
    } `toml:"database"`
}

func main() {
    var config ServerConfig
    
    // Load existing config or use defaults
    if data, err := os.ReadFile("config.toml"); err == nil {
        if err := tinytoml.Unmarshal(data, &config); err != nil {
            log.Fatal(err)
        }
    } else {
        config = getDefaultConfig()
    }

    // Save config
    data, err := tinytoml.Marshal(config)
    if err != nil {
        log.Fatal(err)
    }
    os.WriteFile("config.toml", data, 0644)
}
```

See [examples/] directory for more comprehensive examples including:
- Basic roundtrip conversion
- Default configuration management
- JSON/TOML configuration comparison

## API

### `Marshal(v any) ([]byte, error)`
Converts a Go value into TOML format. Supports structs, maps (with string keys), and basic types.

### `Unmarshal(data []byte, v any) error`
Parses TOML data into a Go value. Target must be a pointer to a struct or map.

## Error Handling

TinyTOML provides error messages with context:

```go
unmarshalErr := tinytoml.Unmarshal([]byte("[invalid table]"), &data)
// github.com/LixenWraith/tinytoml.tokenizeLine: invalid table name [line 1]

marshalErr := tinytoml.Marshal(make(chan int))
// github.com/LixenWraith/tinytoml.Marshal: unsupported type
```

## License

BSD-3