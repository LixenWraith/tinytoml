# tinytoml

Package tinytoml provides a minimal TOML parser and encoder for configuration files.
Similar interface to encoding/json.

Supported Features:
- Basic types: string, number (int/float), boolean
- Table/group nesting up to 3 levels
- Single-line string values with escape sequences (\, \", \', \t)
- Both inline (#) and full-line comments
- Flexible whitespace around equals sign and line start
- Quoted string values (must be used for strings containing whitespace)
- Strict whitespace handling (unquoted values cannot contain whitespace)
- Integer overflow detection and number format validation
- Duplicate key detection (first occurrence used)

Limitations:
- No array/slice support
- No multi-line string support
- Limited escape sequence support (only \, \", \', \t)
- Maximum 3 levels of table/group nesting
- No support for custom time formats
- No support for hex/octal/binary number formats
- No scientific notation support for numbers
- Unquoted strings cannot contain whitespace (use quotes)

## Installation and use

Follow the example in src/cmd/main.go
Run build.sh bash script to build the example program and run the tests (linux).
