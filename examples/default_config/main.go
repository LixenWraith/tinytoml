package main

import (
	"fmt"
	"log"
	"os"

	"github.com/LixenWraith/tinytoml"
)

// ServerConfig represents a typical server application configuration
type ServerConfig struct {
	// Basic server settings
	Server struct {
		Host string `toml:"host"` // Server host address
		Port int64  `toml:"port"` // Server port number
		Name string `toml:"name"` // Server instance name
		Mode string `toml:"mode"` // Running mode (development/production)
	} `toml:"server"`

	// TLS/SSL configuration
	TLS struct {
		Enabled  bool   `toml:"enabled"`   // Enable/disable TLS
		CertFile string `toml:"cert_file"` // Path to certificate file
		KeyFile  string `toml:"key_file"`  // Path to private key file
	} `toml:"tls"`

	// Database connection settings
	Database struct {
		Host     string `toml:"host"`     // Database host
		Port     int64  `toml:"port"`     // Database port
		Name     string `toml:"name"`     // Database name
		User     string `toml:"user"`     // Database user
		Password string `toml:"password"` // Database password
		Pool     struct {
			MaxOpen int64 `toml:"max_open"` // Maximum open connections
			MaxIdle int64 `toml:"max_idle"` // Maximum idle connections
		} `toml:"pool"`
	} `toml:"database"`

	// API related settings
	API struct {
		Prefix      string   `toml:"prefix"`       // API route prefix
		Timeout     int64    `toml:"timeout"`      // Request timeout in seconds
		RateLimit   int64    `toml:"rate_limit"`   // Requests per minute
		CorsOrigins []string `toml:"cors_origins"` // Allowed CORS origins
	} `toml:"api"`
}

// getDefaultConfig returns a ServerConfig with default values
func getDefaultConfig() ServerConfig {
	config := ServerConfig{}

	// Set default values for server
	config.Server.Host = "localhost"
	config.Server.Port = 8080
	config.Server.Name = "app-server"
	config.Server.Mode = "development"

	// Set default TLS configuration
	config.TLS.Enabled = false
	config.TLS.CertFile = "cert/server.crt"
	config.TLS.KeyFile = "cert/server.key"

	// Set default database configuration
	config.Database.Host = "localhost"
	config.Database.Port = 5432
	config.Database.Name = "appdb"
	config.Database.User = "dbuser"
	config.Database.Password = "dbpass"
	config.Database.Pool.MaxOpen = 10
	config.Database.Pool.MaxIdle = 5

	// Set default API configuration
	config.API.Prefix = "/api/v1"
	config.API.Timeout = 30
	config.API.RateLimit = 100
	config.API.CorsOrigins = []string{"http://localhost:3000", "https://app.example.com"}

	return config
}

func main() {
	const configFile = "./examples/default_config/config.toml"
	var config ServerConfig

	// Check if config file exists and load it
	if data, err := os.ReadFile(configFile); err == nil {
		fmt.Println("Loading existing config file...")
		if err := tinytoml.Unmarshal(data, &config); err != nil {
			log.Fatalf("Failed to parse config file: %v", err)
		}
	} else {
		fmt.Println("Config file not found, using default values...")
		config = getDefaultConfig()
	}

	// Write config to file
	data, err := tinytoml.Marshal(config)
	if err != nil {
		log.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		log.Fatalf("Failed to write config file: %v", err)
	}
	fmt.Printf("Configuration saved to %s\n", configFile)
}

/* Example output for first run (no config file):
Config file not found, using default values...
Configuration saved to config.toml

Example output for subsequent runs (with existing config):
Loading existing config file...
Configuration saved to config.toml

Generated config.toml will contain:
[server]
host = "localhost"
port = 8080
name = "app-server"
mode = "development"

[tls]
enabled = false
cert_file = "cert/server.crt"
key_file = "cert/server.key"

[database]
host = "localhost"
port = 5432
name = "appdb"
user = "dbuser"
password = "dbpass"

[database.pool]
max_open = 10
max_idle = 5

[api]
prefix = "/api/v1"
timeout = 30
rate_limit = 100
cors_origins = ["http://localhost:3000", "https://app.example.com"]

If you modify config.toml and remove some fields,
the next run will restore them with default values
while preserving any customized values you set.
*/
