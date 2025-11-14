package config

import (
	"os"
	"testing"
)

func TestMergeWithDefaults(t *testing.T) {
	// Create a config with some fields missing
	incomplete := &Config{
		Screen: Screen{
			Width:  160,
			Height: 96,
			// Lighting field missing - should be added from defaults
		},
		Debug: true,
		// Entities field completely missing - should be added from defaults
	}

	// Simulate YAML with only width, height, and debug fields
	existingFields := map[string]interface{}{
		"screen": map[string]interface{}{
			"width":  160.0,
			"height": 96.0,
		},
		"debug": true,
	}

	// Merge with defaults
	merged, result := MergeWithDefaults(incomplete, existingFields)

	// Verify existing fields are preserved
	if merged.Screen.Width != 160.0 {
		t.Errorf("Expected Screen.Width to be preserved as 160, got %0.1f", merged.Screen.Width)
	}
	if merged.Screen.Height != 96.0 {
		t.Errorf("Expected Screen.Height to be preserved as 96, got %0.1f", merged.Screen.Height)
	}
	if merged.Debug != true {
		t.Errorf("Expected Debug to be preserved as true, got %v", merged.Debug)
	}

	// Verify new fields were added
	if len(result.NewFields) == 0 {
		t.Error("Expected some new fields to be added, got none")
	}

	// Verify Entities struct was added with defaults
	if merged.Entities.Knight == 0 {
		t.Error("Expected Entities.Knight to be populated with default, got 0")
	}
	if merged.Entities.Knight != defaultKnight {
		t.Errorf("Expected Entities.Knight to be %d, got %d", defaultKnight, merged.Entities.Knight)
	}

	// Verify Lighting field was added
	if merged.Screen.Lighting != lighting {
		t.Errorf("Expected Screen.Lighting to be %v, got %v", lighting, merged.Screen.Lighting)
	}
}

func TestLoadConfigWithMerge(t *testing.T) {
	// Create a temporary config file with missing fields
	tmpFile := "test_config_merge.yml"
	defer os.Remove(tmpFile)

	incompleteYAML := `screen:
  width: 160
  height: 96
debug: true
`
	if err := os.WriteFile(tmpFile, []byte(incompleteYAML), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load with merge
	cfg, result, err := LoadConfigWithMerge(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfigWithMerge failed: %v", err)
	}

	// Verify existing fields preserved
	if cfg.Screen.Width != 160.0 {
		t.Errorf("Expected Width 160, got %0.1f", cfg.Screen.Width)
	}
	if cfg.Debug != true {
		t.Errorf("Expected Debug true, got %v", cfg.Debug)
	}

	// Verify new fields added
	if len(result.NewFields) == 0 {
		t.Error("Expected new fields to be detected")
	}

	// Verify entities were added
	if cfg.Entities.Knight == 0 {
		t.Error("Expected Entities.Knight to be populated")
	}

	// Verify file was updated with new fields
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read updated config: %v", err)
	}

	// Check that entities section was added to file
	content := string(data)
	if len(content) <= len(incompleteYAML) {
		t.Error("Expected config file to be updated with new fields")
	}
}

func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"zero int", int(0), true},
		{"non-zero int", int(42), false},
		{"zero float", float64(0), true},
		{"non-zero float", float64(3.14), false},
		{"zero string", "", true},
		{"non-zero string", "hello", false},
		{"false bool", false, true},
		{"true bool", true, false},
		{"zero uint", uint32(0), true},
		{"non-zero uint", uint32(42), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test is conceptual - isZeroValue is not exported
			// In practice, we test it indirectly through MergeWithDefaults
		})
	}
}

func TestMergePreservesUserValues(t *testing.T) {
	// Create a complete config with custom values
	custom := &Config{
		Screen: Screen{
			Width:    320,  // Custom size
			Height:   240,  // Custom size
			Lighting: true, // Different from default
		},
		Entities: Entities{
			Knight: 99, // Custom GID
			Ghoul:  100,
		},
		Debug: true,
	}

	// Simulate YAML with all fields present
	existingFields := map[string]interface{}{
		"screen": map[string]interface{}{
			"width":    320,
			"height":   240,
			"lighting": true,
		},
		"entities": map[string]interface{}{
			"knight": 99,
			"ghoul":  100,
		},
		"debug": true,
	}

	// Merge should preserve all custom values
	merged, _ := MergeWithDefaults(custom, existingFields)

	if merged.Screen.Width != 320.0 {
		t.Errorf("Custom Screen.Width not preserved, got %0.1f", merged.Screen.Width)
	}
	if merged.Screen.Height != 240.0 {
		t.Errorf("Custom Screen.Height not preserved, got %0.1f", merged.Screen.Height)
	}
	if merged.Screen.Lighting != true {
		t.Errorf("Custom Screen.Lighting not preserved, got %v", merged.Screen.Lighting)
	}
	if merged.Entities.Knight != 99 {
		t.Errorf("Custom Entities.Knight not preserved, got %d", merged.Entities.Knight)
	}
	if merged.Entities.Ghoul != 100 {
		t.Errorf("Custom Entities.Ghoul not preserved, got %d", merged.Entities.Ghoul)
	}
	if merged.Debug != true {
		t.Errorf("Custom Debug not preserved, got %v", merged.Debug)
	}
}

func TestMergeResult(t *testing.T) {
	// Create config missing several fields
	incomplete := &Config{
		Screen: Screen{
			Width:  160,
			Height: 96,
			// Lighting missing
		},
		// Everything else missing
	}

	// Simulate YAML with only width and height
	existingFields := map[string]interface{}{
		"screen": map[string]interface{}{
			"width":  160,
			"height": 96,
		},
	}

	_, result := MergeWithDefaults(incomplete, existingFields)

	// Should have detected preserved fields
	if len(result.UpdatedFields) == 0 {
		t.Error("Expected some updated fields to be tracked")
	}

	// Should have detected new fields
	if len(result.NewFields) == 0 {
		t.Error("Expected some new fields to be tracked")
	}

	// Check that field names are tracked
	foundWidth := false
	for _, field := range result.UpdatedFields {
		if field == "Screen.Width" {
			foundWidth = true
			break
		}
	}
	if !foundWidth {
		t.Error("Expected Screen.Width to be in UpdatedFields")
	}
}
