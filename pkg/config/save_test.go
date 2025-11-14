package config

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestSavePreservesFalseValues tests that saving a config preserves false boolean values.
// This is critical to ensure the Save() operation doesn't lose false values.
func TestSavePreservesFalseValues(t *testing.T) {
	tmpFile := "test_save_false.yml"
	defer os.Remove(tmpFile)

	// Create a config with false values
	cfg := NewDefaultConfig()
	cfg.Screen.Lighting = false
	cfg.Debug = false

	// Save it
	if err := cfg.Save(tmpFile); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Read back the raw YAML
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	// Parse the YAML to check actual values
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to parse saved YAML: %v", err)
	}

	// Check that lighting is false in the saved file
	screen, ok := parsed["screen"].(map[string]interface{})
	if !ok {
		t.Fatal("screen section not found in saved config")
	}

	lightingValue, exists := screen["lighting"]
	if !exists {
		t.Error("lighting field not found in saved config")
	}
	if lightingValue != false {
		t.Errorf("Expected lighting to be false in saved file, got %v", lightingValue)
	}

	// Check that debug is false
	debugValue, exists := parsed["debug"]
	if !exists {
		t.Error("debug field not found in saved config")
	}
	if debugValue != false {
		t.Errorf("Expected debug to be false in saved file, got %v", debugValue)
	}
}

// TestSaveThenLoadPreservesFalse tests the full cycle: save config with false, then load it back.
func TestSaveThenLoadPreservesFalse(t *testing.T) {
	tmpFile := "test_save_load_false.yml"
	defer os.Remove(tmpFile)

	// Create and save config with false values
	cfg := NewDefaultConfig()
	cfg.Screen.Lighting = false
	cfg.Debug = false
	cfg.Stats.PlayerInvulnerable = false

	if err := cfg.Save(tmpFile); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load it back with merge
	loaded, _, err := LoadConfigWithMerge(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify false values are preserved
	if loaded.Screen.Lighting != false {
		t.Errorf("Expected lighting to be false after load, got %v", loaded.Screen.Lighting)
	}
	if loaded.Debug != false {
		t.Errorf("Expected debug to be false after load, got %v", loaded.Debug)
	}
	if loaded.Stats.PlayerInvulnerable != false {
		t.Errorf("Expected playerInvulnerable to be false after load, got %v", loaded.Stats.PlayerInvulnerable)
	}
}
