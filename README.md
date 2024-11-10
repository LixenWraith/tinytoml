# TinyTOML

A minimal TOML parser and encoder for Go configuration files. TinyTOML provides a lightweight implementation focusing on the most commonly used TOML features while maintaining strict parsing rules.

## Features


- Basic TOML types support:
  - Strings (single-line)
  - Numbers (integers and floats)
  - Booleans
  - Arrays (including nested and mixed-type arrays)
- Unlimited table/group nesting
- String escape sequences (`\`, `\"`, `\'`, `\t`)
- Full comment support (both inline `#` and full-line)
- Flexible whitespace handling around equals sign and line starts
- Strict parsing rules:
  - Integer overflow detection
  - Number format validation
  - Whitespace validation in unquoted strings
  - Duplicate key detection (first occurrence used)

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


type Config struct {
    AppName string   `toml:"app_name"`
    Port    int      `toml:"port"`
    Debug   bool     `toml:"debug"`
    Tags    []string `toml:"tags"`
    DB      struct {
        Host     string  `toml:"host"`
        Port     int     `toml:"port"`
        Replicas []string `toml:"replicas"`
    } `toml:"database"`
}


func main() {
    input := `

# Application config
app_name = "MyApp"
port = 8080
debug = true

tags = ["prod", "v1", "critical"]

[database]
host = "localhost"
port = 5432
replicas = [
    "replica1.local",

    "replica2.local"
]`

    var config Config
    if err := tinytoml.Unmarshal([]byte(input), &config); err != nil {
        log.Fatal(err)

    }

    fmt.Printf("%+v\n", config)


    // Marshal back to TOML
    output, err := tinytoml.MarshalIndent(config)
    if err != nil {
        log.Fatal(err)

    }

    fmt.Printf("\nGenerated TOML:\n%s", output)
}
```


## Limitations

TinyTOML intentionally omits some TOML features to maintain simplicity:

- No multi-line string support

- Limited escape sequence support (only `\`, `\"`, `\'`, `\t`)
- No support for custom time formats
- No support for hex/octal/binary number formats
- No scientific notation support for numbers
- Unquoted strings cannot contain whitespace (must use quotes)


## Design Principles

1. **Simplicity**: Focus on the most commonly used TOML features
2. **Strictness**: Enforce strict parsing rules to prevent ambiguous configurations
3. **Predictability**: Clear behavior for edge cases (e.g., duplicate keys)

4. **Type Safety**: Strong type checking and overflow protection
5. **Readability**: Clean, well-formatted output with proper indentation


## API

### `Unmarshal(data []byte, v interface{}) error`
Parses TOML data into a struct

### `Marshal(v interface{}) ([]byte, error)`
Converts a struct to TOML format

### `MarshalIndent(v interface{}) ([]byte, error)`
Converts a struct to TOML format with proper indentation

## Error Handling


TinyTOML provides detailed error messages including line numbers and context:

```go
if err := tinytoml.Unmarshal(data, &config); err != nil {
    switch e := err.(type) {
    case *tinytoml.ParseError:
        fmt.Printf("Parse error at line %d: %v\n", e.Line, e.Message)
    default:
        fmt.Printf("Error: %v\n", err)

    }
}
```

## License

MIT License - see LICENSE file


