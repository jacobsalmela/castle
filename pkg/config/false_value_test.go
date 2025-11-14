package config

import (
	"os"
	"testing"
)

// TestFalseValuePreservation specifically tests that false values are preserved
// during config merge. This was a critical bug where false values were treated
// as "missing" and replaced with defaults.
func TestFalseValuePreservation(t *testing.T) {
	// Create a config file with lighting explicitly set to false
	tmpFile := "test_false_value.yml"
	defer os.Remove(tmpFile)

	configYAML := `screen:
  width: 160
  height: 96
  lighting: false
debug: false
`
	if err := os.WriteFile(tmpFile, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load with merge
	cfg, _, err := LoadConfigWithMerge(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfigWithMerge failed: %v", err)
	}

	// Critical: lighting must remain false, not be replaced with default true
	if cfg.Screen.Lighting != false {
		t.Errorf("Expected lighting to be preserved as false, got %v", cfg.Screen.Lighting)
	}

	// Debug must also remain false
	if cfg.Debug != false {
		t.Errorf("Expected debug to be preserved as false, got %v", cfg.Debug)
	}

	// Verify other fields got defaults
	if cfg.Entities.Knight == 0 {
		t.Error("Expected Entities.Knight to be populated with default")
	}
}

// TestZeroValuePreservation tests that zero values (0, "", etc.) are preserved
// when explicitly set in config.
func TestZeroValuePreservation(t *testing.T) {
	tmpFile := "test_zero_value.yml"
	defer os.Remove(tmpFile)

	configYAML := `hud:
  hudIconsX: 0
  barEndX1: 0
  maxTextWidth: 0
screen:
  width: 160
  height: 96
`
	if err := os.WriteFile(tmpFile, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, _, err := LoadConfigWithMerge(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfigWithMerge failed: %v", err)
	}

	// These zeros were explicitly set and should be preserved
	if cfg.Hud.HudIconsX != 0 {
		t.Errorf("Expected hudIconsX to be preserved as 0, got %d", cfg.Hud.HudIconsX)
	}
	if cfg.Hud.BarEndX1 != 0 {
		t.Errorf("Expected barEndX1 to be preserved as 0, got %d", cfg.Hud.BarEndX1)
	}
	if cfg.Hud.MaxTextWidth != 0 {
		t.Errorf("Expected maxTextWidth to be preserved as 0, got %d", cfg.Hud.MaxTextWidth)
	}

	// Fields not in YAML should get defaults
	if cfg.Hud.BarH == 0 {
		t.Error("Expected barH to be populated with default")
	}
}

// TestMissingFieldGetsDefault verifies that truly missing fields get defaults.
func TestMissingFieldGetsDefault(t *testing.T) {
	tmpFile := "test_missing_field.yml"
	defer os.Remove(tmpFile)

	// Config with lighting field completely missing
	configYAML := `screen:
  width: 160
  height: 96
`
	if err := os.WriteFile(tmpFile, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, result, err := LoadConfigWithMerge(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfigWithMerge failed: %v", err)
	}

	// Lighting was missing, should get default true
	if cfg.Screen.Lighting != lighting {
		t.Errorf("Expected lighting to be default %v, got %v", lighting, cfg.Screen.Lighting)
	}

	// Should be in NewFields list
	foundLighting := false
	for _, field := range result.NewFields {
		if field == "Screen.Lighting" {
			foundLighting = true
			break
		}
	}
	if !foundLighting {
		t.Error("Expected Screen.Lighting to be in NewFields list")
	}
}
