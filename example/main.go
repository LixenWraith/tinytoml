package main

import (
	"fmt"
	"log"

	"github.com/LixenWraith/tinytoml"
)

type ServerConfig struct {
	Host     string   `toml:"host"`
	Port     int      `toml:"port"`
	SSLPorts []int    `toml:"ssl_ports"`
	Regions  []string `toml:"regions"`
}

type DatabaseConfig struct {
	Primary struct {
		Host     string   `toml:"host"`
		Port     int      `toml:"port"`
		User     string   `toml:"user"`
		Password string   `toml:"password"`
		Replicas []string `toml:"replicas"`
	} `toml:"primary"`
	ReadOnly struct {
		Hosts    []string `toml:"hosts"`
		Port     int      `toml:"port"`
		User     string   `toml:"user"`
		Password string   `toml:"password"`
	} `toml:"readonly"`
}

type Config struct {
	AppName     string         `toml:"app_name"`
	Version     string         `toml:"version"`
	Debug       bool           `toml:"debug"`
	MaxWorkers  int            `toml:"max_workers"`
	RateLimit   float64        `toml:"rate_limit"`
	LogLevels   []string       `toml:"log_levels"`
	Tags        []string       `toml:"tags"`
	Server      ServerConfig   `toml:"server"`
	Database    DatabaseConfig `toml:"database"`
	Matrix      [][]int        `toml:"matrix"`      // Example of nested arrays
	Mixed       []interface{}  `toml:"mixed_array"` // Example of mixed type array
	Environment string         `toml:"env.type"`    // Example of dot notation without group
}

func main() {
	// Example TOML content demonstrating all supported features
	input := `# Application Configuration
app_name = "Complex Server"
version = "2.0.0"
debug = true
max_workers = 16
rate_limit = 1.5
log_levels = ["INFO", "WARN", "ERROR"]
tags = ["prod", "high-memory", "cluster-1"]

# Demonstrate nested arrays
matrix = [
   [1, 2, 3],
   [4, 5, 6],
   [7, 8, 9]
]

# Demonstrate mixed type array
mixed_array = ["string", 42, 3.14, true, [1, 2, 3]]

# Environment settings using dot notation
env.type = "production"

[server]
host = "localhost"
port = 8080
ssl_ports = [443, 8443, 9443]
regions = ["us-east", "us-west", "eu-central"]

[database.primary]
host = "db-master.internal"
port = 5432
user = "admin"
password = "super-secret"
replicas = [
   "db-replica-1.internal",
   "db-replica-2.internal",
   "db-replica-3.internal"
]

[database.readonly]
hosts = ["db-ro-1.internal", "db-ro-2.internal"]
port = 5432
user = "reader"
password = "read-only-pass"`

	var config Config
	if err := tinytoml.Unmarshal([]byte(input), &config); err != nil {
		log.Fatalf("Failed to unmarshal: %v", err)
	}

	// Print the parsed configuration
	fmt.Printf("Application: %s v%s\n", config.AppName, config.Version)
	fmt.Printf("Environment: %s\n", config.Environment)
	fmt.Printf("Debug: %v\n", config.Debug)
	fmt.Printf("Workers: %d\n", config.MaxWorkers)
	fmt.Printf("Rate Limit: %.2f\n", config.RateLimit)
	fmt.Printf("Log Levels: %v\n", config.LogLevels)
	fmt.Printf("Tags: %v\n", config.Tags)

	fmt.Printf("\nServer Configuration:\n")
	fmt.Printf("Host: %s\n", config.Server.Host)
	fmt.Printf("Port: %d\n", config.Server.Port)
	fmt.Printf("SSL Ports: %v\n", config.Server.SSLPorts)
	fmt.Printf("Regions: %v\n", config.Server.Regions)

	fmt.Printf("\nDatabase Configuration:\n")
	fmt.Printf("Primary DB: %s:%d\n", config.Database.Primary.Host, config.Database.Primary.Port)
	fmt.Printf("Primary Replicas: %v\n", config.Database.Primary.Replicas)
	fmt.Printf("ReadOnly Hosts: %v\n", config.Database.ReadOnly.Hosts)

	fmt.Printf("\nMatrix:\n")
	for _, row := range config.Matrix {
		fmt.Printf("%v\n", row)
	}

	fmt.Printf("\nMixed Array: %v\n", config.Mixed)

	// Marshal back to TOML with indentation
	output, err := tinytoml.MarshalIndent(config)
	if err != nil {
		log.Fatalf("Failed to marshal: %v", err)
	}

	fmt.Printf("\nGenerated TOML:\n%s", output)
}
