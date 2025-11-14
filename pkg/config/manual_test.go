//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jacobsalmela/castle/pkg/config"
)

// This is a manual test to verify that false values are preserved during config merge.
// Run with: go run pkg/config/manual_test.go
func main() {
	configPath := "config.yml"

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config.yml not found at %s", configPath)
	}

	fmt.Println("Loading config with merge...")
	cfg, result, err := config.LoadConfigWithMerge(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("\nMerge Results:\n")
	fmt.Printf("  New fields added: %d\n", len(result.NewFields))
	if len(result.NewFields) > 0 {
		for _, field := range result.NewFields {
			fmt.Printf("    - %s\n", field)
		}
	}
	fmt.Printf("  Deprecated fields: %d\n", len(result.DeprecatedFields))
	fmt.Printf("  Updated fields: %d\n", len(result.UpdatedFields))

	fmt.Printf("\nCritical Values:\n")
	fmt.Printf("  Screen.Lighting: %v\n", cfg.Screen.Lighting)
	fmt.Printf("  Debug: %v\n", cfg.Debug)
	fmt.Printf("  Screen.Width: %d\n", cfg.Screen.Width)
	fmt.Printf("  Screen.Height: %d\n", cfg.Screen.Height)

	fmt.Printf("\nEntity GIDs:\n")
	fmt.Printf("  Knight: %d\n", cfg.Entities.Knight)
	fmt.Printf("  Ghoul: %d\n", cfg.Entities.Ghoul)

	// Read the config file to show what's in it
	fmt.Printf("\nCurrent config.yml lighting value:\n")
	data, _ := os.ReadFile(configPath)
	for i, line := range []string{} {
		if i > 10 {
			break
		}
		fmt.Println(line)
	}

	// Show first few lines with lighting
	content := string(data)
	lines := []byte{}
	inScreen := false
	lineCount := 0
	for _, b := range content {
		lines = append(lines, byte(b))
		if b == '\n' {
			line := string(lines)
			if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
				if inScreen {
					break
				}
				if len(line) > 6 && line[:6] == "screen" {
					inScreen = true
				}
			}
			if inScreen {
				fmt.Print(line)
				lineCount++
				if lineCount > 4 {
					break
				}
			}
			lines = []byte{}
		}
	}

	fmt.Printf("\n✅ Test complete! If lighting is false in config.yml, it should show false above.\n")
}
