package main

import (
	"github.com/LixenWraith/tinytoml"
	"log"
)

func main() {
	// Unmarshal error - invalid TOML format
	input := `
[application server]  # Error: space inside table name not acceptable
`
	var data any
	if err := tinytoml.Unmarshal([]byte(input), &data); err != nil {
		log.Printf("Unmarshal error: %v\n", err)
	}

	// Marshal error - unsupported type
	invalidData := make(chan int) // channels are unsupported
	if _, err := tinytoml.Marshal(invalidData); err != nil {
		log.Printf("Marshal error: %v\n", err)
	}
}
