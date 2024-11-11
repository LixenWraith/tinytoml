package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/LixenWraith/tinytoml"
)

type Config struct {
	Name    string   `json:"name"    toml:"name"`
	Port    int      `json:"port"    toml:"port"`
	Tags    []string `json:"tags"    toml:"tags"`
	Enabled bool     `json:"enabled" toml:"enabled"`
}

func main() {
	// JSON Example
	jsonData := `{
        "name": "test-app",
        "port": 8080,
        "tags": ["dev", "test"],
        "enabled": true
    }`

	var jsonConfig Config
	if err := json.Unmarshal([]byte(jsonData), &jsonConfig); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("JSON Config: %+v\n", jsonConfig)

	// TOML Example - now just as simple!
	tomlData := `
name = "test-app"
port = 8080
tags = ["dev", "test"]
enabled = true
`
	var tomlConfig Config
	if err := tinytoml.Unmarshal([]byte(tomlData), &tomlConfig); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("TOML Config: %+v\n", tomlConfig)

	// Both can marshal back just as easily
	jsonBytes, _ := json.Marshal(jsonConfig)
	tomlBytes, _ := tinytoml.Marshal(tomlConfig)

	fmt.Printf("\nJSON output:\n%s\n", jsonBytes)
	fmt.Printf("\nTOML output:\n%s\n", tomlBytes)
}
