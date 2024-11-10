// File: tintomy/example/main.go

package main

import (
	"fmt"
	"log"

	"github.com/LixenWraith/tinytoml"
)

type Config struct {
	Title    string  `toml:"title"`
	Port     int     `toml:"port"`
	Debug    bool    `toml:"debug"`
	Rate     float64 `toml:"rate"`
	Host     string  `toml:"server.host"`
	SSLPort  int     `toml:"server.ssl_port"`
	User     string  `toml:"db.user"`
	Password string  `toml:"db.password"`
}

func main() {
	// Example TOML content with proper string escaping
	input := `# Server Configuration
title = "Test Server"
port = 8080
debug = true
rate = 0.75

[server]

host = "localhost"
ssl_port = 443  # Default SSL port

[db]
user = "admin"
password = "secret123"  # Should be in env var`

	var config Config
	if err := tinytoml.Unmarshal([]byte(input), &config); err != nil {
		log.Fatalf("Failed to unmarshal: %v", err)
	}

	fmt.Printf("Loaded config: %+v\n", config)

	// Marshal back to TOML with proper indentation
	output, err := tinytoml.MarshalIndent(config)
	if err != nil {
		log.Fatalf("Failed to marshal: %v", err)
	}

	fmt.Printf("\nGenerated TOML:\n%s", output)
}
