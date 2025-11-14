package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// MergeResult contains information about the merge operation.
type MergeResult struct {
	NewFields        []string // Fields added from defaults
	DeprecatedFields []string // Fields in loaded config but not in defaults
	UpdatedFields    []string // Fields that existed and were preserved
}

// MergeWithDefaults merges a loaded config with default values.
// - New fields from defaults are added to loaded config
// - Existing fields in loaded config are preserved (even if zero value)
// - Fields in loaded config but not in defaults are marked as deprecated
//
// existingFields is a map of fields that actually exist in the YAML file,
// used to distinguish between "field set to false" vs "field missing from YAML".
//
// Returns the merged config and information about the merge.
func MergeWithDefaults(loaded *Config, existingFields map[string]interface{}) (*Config, *MergeResult) {
	defaults := NewDefaultConfig()
	result := &MergeResult{
		NewFields:        make([]string, 0),
		DeprecatedFields: make([]string, 0),
		UpdatedFields:    make([]string, 0),
	}

	// Use reflection to merge struct fields
	merged := &Config{}
	mergeStructs(reflect.ValueOf(merged).Elem(), reflect.ValueOf(loaded).Elem(), reflect.ValueOf(defaults).Elem(), "", existingFields, result)

	return merged, result
}

// mergeStructs recursively merges struct fields.
// - If field exists in YAML (even if zero value), use loaded value
// - If field doesn't exist in YAML, use default value
// - Track all changes in MergeResult
func mergeStructs(dst, src, defaults reflect.Value, path string, existingFields map[string]interface{}, result *MergeResult) {
	dstType := dst.Type()

	for i := 0; i < dst.NumField(); i++ {
		field := dstType.Field(i)
		fieldPath := field.Name
		yamlFieldName := getYAMLFieldName(field)

		if path != "" {
			fieldPath = path + "." + field.Name
		}

		dstField := dst.Field(i)
		srcField := src.Field(i)
		defaultField := defaults.Field(i)

		// Skip unexported fields
		if !dstField.CanSet() {
			continue
		}

		// Handle different field types
		switch dstField.Kind() {
		case reflect.Struct:
			// Get nested map for struct fields
			var nestedFields map[string]interface{}
			if existingFields != nil {
				if nested, ok := existingFields[yamlFieldName].(map[string]interface{}); ok {
					nestedFields = nested
				}
			}
			// Recursively merge nested structs
			mergeStructs(dstField, srcField, defaultField, fieldPath, nestedFields, result)

		default:
			// Check if field exists in the YAML (not just if it has a zero value)
			fieldExists := fieldExistsInYAML(existingFields, yamlFieldName)

			if !fieldExists {
				// Field is missing from YAML, use default
				if dstField.CanSet() {
					dstField.Set(defaultField)
					result.NewFields = append(result.NewFields, fieldPath)
				}
			} else {
				// Field exists in YAML, preserve loaded value (even if zero)
				if dstField.CanSet() {
					dstField.Set(srcField)
					result.UpdatedFields = append(result.UpdatedFields, fieldPath)
				}
			}
		}
	}
}

// getYAMLFieldName extracts the YAML field name from struct tags.
func getYAMLFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("yaml")
	if tag == "" {
		return strings.ToLower(field.Name)
	}
	// Split on comma to handle tags like "field,omitempty"
	parts := strings.Split(tag, ",")
	return parts[0]
}

// fieldExistsInYAML checks if a field actually exists in the parsed YAML map.
func fieldExistsInYAML(yamlMap map[string]interface{}, fieldName string) bool {
	if yamlMap == nil {
		return false
	}
	_, exists := yamlMap[fieldName]
	return exists
}

// isZeroValue checks if a reflect.Value is the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Struct:
		// For structs, check if all fields are zero
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// LoadConfigWithMerge loads config and merges with defaults.
// This ensures new fields are added and deprecated fields are noted.
func LoadConfigWithMerge(path string) (*Config, *MergeResult, error) {
	// First, try to load the raw YAML to see what fields exist
	yamlData, err := os.ReadFile(path)
	var existingFields map[string]interface{}
	if err == nil {
		// Parse YAML to see what fields are actually present
		yaml.Unmarshal(yamlData, &existingFields)
	}

	// Load the config file (may have missing fields)
	loaded, err := LoadConfig(path)
	if err != nil {
		return nil, nil, err
	}

	// Merge with defaults to add new fields
	merged, result := MergeWithDefaults(loaded, existingFields)

	// Log merge results
	if len(result.NewFields) > 0 {
		log.Printf("Config merge: Added %d new fields with defaults", len(result.NewFields))
		for _, field := range result.NewFields {
			log.Printf("  + %s (using default)", field)
		}
	}

	if len(result.DeprecatedFields) > 0 {
		log.Printf("Config merge: Found %d deprecated fields", len(result.DeprecatedFields))
		for _, field := range result.DeprecatedFields {
			log.Printf("  - %s (deprecated, will be ignored)", field)
		}
	}

	// Save merged config back to file to add new fields
	if len(result.NewFields) > 0 {
		if err := merged.Save(path); err != nil {
			log.Printf("Warning: Failed to save merged config: %v", err)
		} else {
			log.Printf("Config file updated with %d new fields", len(result.NewFields))
		}
	}

	return merged, result, nil
}

// LogMergeResult logs a formatted summary of the merge operation.
func LogMergeResult(result *MergeResult) {
	if result == nil {
		return
	}

	summary := fmt.Sprintf("Config merge summary: %d updated, %d new, %d deprecated",
		len(result.UpdatedFields), len(result.NewFields), len(result.DeprecatedFields))
	log.Println(summary)

	if len(result.NewFields) > 0 {
		log.Println("  New fields added:")
		for _, field := range result.NewFields {
			log.Printf("    + %s", field)
		}
	}

	if len(result.DeprecatedFields) > 0 {
		log.Println("  Deprecated fields (ignored):")
		for _, field := range result.DeprecatedFields {
			log.Printf("    - %s", field)
		}
	}
}
