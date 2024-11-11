package main

import (
	"fmt"
	"github.com/LixenWraith/tinytoml"
	"log"
)

func main() {
	input := `# Basic types demonstration
name = "Complex \nApp"    # quoted string with space
env = production        # bare string
debug = true           # boolean true
maintenance = false    # boolean false
workers = 42           # positive integer
timeout = -30          # negative integer
rate = 3.14           # positive float
temp = -0.5           # negative float


# Array examples (one of each type)
ports = [8080, -6379]                      # mixed sign integers
rates = [1.5, -2.5]                        # mixed sign floats
flags = [true, false]                      # booleans
hosts = ["local host", bare_host]          # mixed quoted/bare strings

# Table examples
[server]
host = "localhost"
port = 8080

# Same table, different keys (should merge)
[server]
name = "main"
active = true

# Nested tables
[database.primary]
host = "db1"
ip = 2.33.45
port = 5432

[database.replica]
host = "db2"
host = "temp"         # first instance of key in the same group/subgroup is used, sebsequent definition of same key is ignored
port = 5433

# Dotted keys (alternative to nested tables)
queue.type = "redis"
queue.port = 6379

# Deeply nested example
[services.cache.redis]
host = "redis1"
port = 6379
slaves = ["redis2", "redis3"]
metrics = [1.1, -2.2, 1926.397247]
features = [true, false]`

	// Parse TOML
	var data map[string]any
	if err := tinytoml.Unmarshal([]byte(input), &data); err != nil {
		log.Fatalf("Unmarshal failed: %v", err)
	}

	fmt.Printf("Parsed structure:\n%#v\n\n", data)

	// Marshal back to TOML
	output, err := tinytoml.Marshal(data)
	if err != nil {
		log.Fatalf("Marshal failed: %v", err)
	}

	fmt.Printf("Generated TOML:\n%s\n", output)

	// Verify roundtrip by parsing again
	var verified map[string]any
	if err := tinytoml.Unmarshal(output, &verified); err != nil {
		log.Fatalf("Verification unmarshal failed: %v", err)
	}
}
