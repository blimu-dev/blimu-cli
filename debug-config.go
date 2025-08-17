package main

import (
	"fmt"
	"log"

	"github.com/blimu-dev/blimu-cli/pkg/config"
)

func main() {
	// Load the config
	blimuConfig, err := config.LoadBlimuConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Debug the loaded config
	fmt.Printf("Resources: %+v\n", blimuConfig.Resources)
	fmt.Printf("Plans: %+v\n", blimuConfig.Plans)
	fmt.Printf("Entitlements: %+v\n", blimuConfig.Entitlements)
	fmt.Printf("Features: %+v\n", blimuConfig.Features)

	// Convert to JSON
	jsonData, err := blimuConfig.MergeToJSON()
	if err != nil {
		log.Fatalf("Failed to convert to JSON: %v", err)
	}

	fmt.Printf("JSON Output:\n%s\n", string(jsonData))
}
