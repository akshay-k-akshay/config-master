// Package configmaster provides a way to manage configuration data from various sources.
package configmaster

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// Config holds the configuration data.
type Config struct {
	data map[string]interface{}
}

// NewConfig creates a new Config instance from various input types (file path or map).
func NewConfig(input interface{}) (*Config, error) {
	// Parse the input to extract configuration data.
	config, err := parseInput(input)
	if err != nil {
		return nil, err
	}

	// Create a new Config instance with the parsed configuration data.
	cfg := &Config{data: config}

	// Process the configuration data recursively to resolve any nested maps and validate the data against the expected formats.
	cfg.data, err = cfg.processRecursively(cfg.data)
	if err != nil {
		return nil, fmt.Errorf("[Config-Master]: %w", err)
	}

	return cfg, nil
}

// parseInput parses the input to extract configuration data.
func parseInput(input interface{}) (map[string]interface{}, error) {
	// Check if the input is a string, map, or something else.
	switch input := input.(type) {
	case string:
		// If the input is a string, read the JSON configuration from the file and return it.
		return parseFromFile(input)
	case map[string]interface{}:
		// If the input is a map, return it as is.
		return input, nil
	default:
		// If the input is something else, return an error.
		return nil, fmt.Errorf("unsupported input type: %T", input)
	}
}

// parseFromFile reads and parses the JSON configuration from a file.
func parseFromFile(filename string) (map[string]interface{}, error) {
	// Open the file and read its contents.
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Read the file contents into a byte slice.
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Unmarshal the byte slice into a map.
	var config map[string]interface{}
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON from file: %w", err)
	}

	// Return the parsed configuration data.
	return config, nil
}

// Get retrieves a value from the configuration data by its key.
func (c *Config) Get(key string) interface{} {
	// Check if the key contains a dot separator.
	if strings.Contains(key, ".") {
		// If the key contains a dot separator, retrieve the nested value using the getNested method.
		return c.getNested(key)
	}
	// If the key does not contain a dot separator, retrieve the value from the top-level configuration data.
	return c.data[key]
}

// getNested retrieves a nested value from the configuration data.
func (c *Config) getNested(key string) interface{} {
	// Split the key into parts based on the dot separator.
	parts := strings.Split(key, ".")

	// Start with the top-level configuration data.
	var value interface{} = c.data

	// Traverse through the configuration data using each part of the key.
	for _, part := range parts {
		// Attempt to access the next level of the configuration data.
		mapValue, ok := value.(map[string]interface{})
		if !ok {
			// Return nil if any part of the key path is invalid.
			return nil
		}

		// Check if the next part of the key is in the configuration data.
		value, ok = mapValue[part]
		if !ok || value == nil {
			// Return nil if any part of the key path is invalid.
			return nil
		}
	}
	// Return the final value found at the end of the key path.
	return value
}

// contains checks if a slice contains a specific value.
func contains[T comparable](slice []T, value T) bool {
	// Iterate over the slice and check if the value is present.
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// getDefaultValue retrieves the default value from the configuration data if it exists.
func getDefaultValue(config map[string]interface{}) interface{} {
	// Check if the default value exists in the configuration data.
	if defaultValue, exists := config["default"]; exists {
		return defaultValue
	}
	// Return an empty string if the default value does not exist.
	return ""
}

// validateAndSetValue validates the configuration data against the expected format and sets the value accordingly.
func validateAndSetValue(config map[string]interface{}) (interface{}, error) {
	// Initialize the value to an empty string.
	var value interface{}

	// Check if the environment variable exists.
	if envKey, exists := config["env"].(string); exists {
		if envValue, exists := os.LookupEnv(envKey); exists {
			// If the environment variable exists, set the value to the environment variable's value.
			value = envValue
		} else {
			// If the environment variable does not exist, set the value to the default value.
			value = getDefaultValue(config)
		}
	} else if _, exists := config["default"]; exists {
		// If the default value exists, set the value to the default value.
		value = getDefaultValue(config)
	} else {
		// If the value is not in the expected format, return what we have.
		return config, nil
	}

	// Check if the expected format exists in the configuration data.
	if expectedFormat, exists := config["format"]; exists {
		// Check if the value is in the expected format.
		if err := isValueInExpectedFormat(value, expectedFormat); err != nil {
			return nil, err
		}
	}

	// Return the validated and set value.
	return value, nil
}

// isValueInExpectedFormat checks if a value is in the expected format.
func isValueInExpectedFormat(value interface{}, format interface{}) error {
	// Get the type of the value.
	valueType := reflect.TypeOf(value)

	// Check if the format is a slice or a string.
	switch format := format.(type) {
	case []interface{}:
		// Check if the value is in the slice of expected formats.
		if !contains(format, value) {
			errorMessage := fmt.Sprintf("value is not in the expected format. Expected formats: %v", format)
			return errors.New(errorMessage)
		}
	case string:
		// Check if the value matches the expected format string.
		switch strings.ToLower(format) {
		case "string":

			if valueType != reflect.TypeOf("") {
				return errors.New("value is not a string")
			}
		case "bool":
			if valueType != reflect.TypeOf(true) {
				return errors.New("value is not a boolean")
			}
		case "float64":
			if valueType != reflect.TypeOf(float64(0)) {
				return errors.New("value is not a float64")
			}
		case "int":
			if valueType != reflect.TypeOf(int(0)) {
				return errors.New("value is not an int")
			}
		}
	default:
		return errors.New("invalid format")
	}

	// Return false if the value is not in the expected format.
	return nil
}

// isNestedMap checks if a map is a nested map or not.
func isNestedMap(config map[string]interface{}) bool {
	// Iterate over all keys in the configuration data.
	for key := range config {
		// Check if this map contains another map.
		if _, ok := config[key].(map[string]interface{}); ok {
			return true
		}
	}
	// Return false if no nested key is found.
	return false
}

// processRecursively processes the configuration data recursively to resolve any nested maps and validate the data against the expected formats.
func (c *Config) processRecursively(config map[string]interface{}) (map[string]interface{}, error) {
	// Create a new map to store the processed configuration data.
	processedConfig := make(map[string]interface{})

	// Iterate over all keys in the configuration data.
	for key, value := range config {
		// Check if the value is a nested map.
		switch typedValue := value.(type) {
		case map[string]interface{}:
			// Check if the map is a nested map or not.
			if !isNestedMap(typedValue) {
				// If the map is not a nested map, validate and set the value using the validateAndSetValue method.
				var err error
				processedConfig[key], err = validateAndSetValue(typedValue)
				if err != nil {
					return nil, err
				}
			} else {
				// If the map is a nested map, recursively process the nested map using the processRecursively method.
				nestedConfig, err := c.processRecursively(typedValue)
				if err != nil {
					return nil, err
				}
				processedConfig[key] = nestedConfig
			}
		case []interface{}:
			// If the value is a slice, process each item in the slice recursively.
			processedSlice := make([]interface{}, len(typedValue))
			for index, item := range typedValue {
				switch nestedItem := item.(type) {
				case map[string]interface{}:
					// If an item is a nested map, recursively process the nested map using the processRecursively method.
					processedItem, err := c.processRecursively(nestedItem)
					if err != nil {
						return nil, err
					}
					processedSlice[index] = processedItem
				default:
					// If an item is not a nested map, add it to the processed slice as is.
					processedSlice[index] = nestedItem
				}
			}
			processedConfig[key] = processedSlice
		default:
			// If the value is not a nested map or a slice, add it to the processed map as is
			processedConfig[key] = value
		}
	}
	return processedConfig, nil
}
