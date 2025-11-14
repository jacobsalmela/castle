package config

import (
	"os"
	"strings"
	"testing"
)

func TestUintSerialization(t *testing.T) {
	// Create a config with uint32 values
	cfg := NewDefaultConfig()

	// Save to temp file
	tmpFile := "test_uint_config.yml"
	defer os.Remove(tmpFile)

	if err := cfg.Save(tmpFile); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Read the file
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	content := string(data)

	// Verify uint values are saved as integers, not quoted strings
	testCases := []struct {
		field    string
		value    string
		shouldBe string
	}{
		{"knight", `knight: "26"`, "knight: 26"},
		{"ghoul", `ghoul: "27"`, "ghoul: 27"},
		{"gram", `gram: "87"`, "gram: 87"},
		{"chest", `chest: "149"`, "chest: 149"},
	}

	for _, tc := range testCases {
		// Check that value is NOT quoted (old bug)
		if strings.Contains(content, tc.value) {
			t.Errorf("Field %s is incorrectly quoted as string: found %q", tc.field, tc.value)
		}

		// Check that value IS unquoted (correct format)
		if !strings.Contains(content, tc.shouldBe) {
			t.Errorf("Field %s should be unquoted integer: expected %q", tc.field, tc.shouldBe)
		}
	}

	// Verify the YAML has !!int tags for entity values
	if !strings.Contains(content, "knight: 26") {
		t.Error("Expected 'knight: 26' (unquoted integer)")
	}
}

func TestUintLoadAndSave(t *testing.T) {
	tmpFile := "test_uint_roundtrip.yml"
	defer os.Remove(tmpFile)

	// Create and save config
	cfg1 := NewDefaultConfig()
	cfg1.Entities.Knight = 99 // Custom value

	if err := cfg1.Save(tmpFile); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config back
	cfg2, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify uint value preserved correctly
	if cfg2.Entities.Knight != 99 {
		t.Errorf("Expected Knight=99, got %d", cfg2.Entities.Knight)
	}

	// Verify it's still a uint32, not converted to string
	if cfg2.Entities.Knight == 0 {
		t.Error("Knight value lost - may have been converted to string")
	}

	// Read file and verify format
	data, _ := os.ReadFile(tmpFile)
	content := string(data)

	// Should be unquoted
	if strings.Contains(content, `knight: "99"`) {
		t.Error("Knight value incorrectly quoted")
	}

	// Should be unquoted integer
	if !strings.Contains(content, "knight: 99") {
		t.Error("Knight value should be unquoted integer")
	}
}

func TestAllEntityGIDsAreUint(t *testing.T) {
	cfg := NewDefaultConfig()

	// Verify all entity GIDs are non-zero
	entities := []struct {
		name  string
		value uint32
	}{
		{"Knight", cfg.Entities.Knight},
		{"Ghoul", cfg.Entities.Ghoul},
		{"Skeleman", cfg.Entities.Skeleman},
		{"Crawler", cfg.Entities.Crawler},
		{"Rat", cfg.Entities.Rat},
		{"Bat", cfg.Entities.Bat},
		{"Ent", cfg.Entities.Ent},
		{"Gram", cfg.Entities.Gram},
		{"Ferragus", cfg.Entities.Ferragus},
		{"Oscar", cfg.Entities.Oscar},
		{"Chest", cfg.Entities.Chest},
		{"Grave", cfg.Entities.Grave},
		{"Door", cfg.Entities.Door},
		{"Spike", cfg.Entities.Spike},
		{"FakeWall", cfg.Entities.FakeWall},
		{"Block", cfg.Entities.Block},
	}

	for _, e := range entities {
		if e.value == 0 {
			t.Errorf("Entity %s has zero GID (should have default value)", e.name)
		}
	}
}
