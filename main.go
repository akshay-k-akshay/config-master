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

// Config represents the configuration and provides methods to access it
type Config struct {
	data map[string]interface{}
}

// NewConfig creates a new Config instance from either a filename or a map
// If input is a string, it is interpreted as a filename and the contents of the file are read and unmarshalled as JSON.
// If input is already a map, it is used directly.
// The method returns a pointer to the new Config instance and an error which is nil if everything went well.
func NewConfig(input interface{}) (*Config, error) {
	var config map[string]interface{}

	// This switch statement handles the two possible types of input.
	switch input := input.(type) {
	case string:
		// If input is a filename
		file, err := os.Open(input)
		if err != nil {
			return nil, fmt.Errorf("error opening file: %w", err)
		}
		defer file.Close()

		var byteValue []byte
		// Read the contents of the file
		byteValue, err = io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}

		// Unmarshal the JSON from the file
		err = json.Unmarshal(byteValue, &config)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON from file: %w", err)
		}
	case map[string]interface{}:
		// If input is already a map
		config = input
	default:
		// If the input is some other type, return an error
		return nil, fmt.Errorf("unsupported input type: %T", input)
	}

	// Create a new Config instance
	cfg := &Config{data: config}

	// Process environment variables recursively
	var err error
	cfg.data, err = cfg.processEnvRecursively(cfg.data)
	if err != nil {
		return nil, fmt.Errorf("[Config-Master]: %w", err)
	}

	// Return the new Config instance and nil error
	return cfg, nil
}

// Get retrieves a value from the configuration by key.
//
// If the key contains a dot, it is treated as a path to a nested value, and the getNested method is used to retrieve it.
// Otherwise, it returns the value directly associated with the key in the top-level data map.
func (c *Config) Get(key string) interface{} {
	// Check if the key is a path to a nested value
	if strings.Contains(key, ".") {
		return c.getNested(key)
	}
	// Return the value directly from the top-level data map
	return c.data[key]
}

// getNested retrieves a value from a nested map structure using a dot-separated key.
// It returns nil if any part of the key does not exist in the map.
//
// The method takes a dot-separated key and traverses the nested map structure using
// each part of the key. It returns the value found at the end of the key path, or
// nil if any part of the key path is invalid.
func (c *Config) getNested(key string) interface{} {
	// Split the key into parts based on the dot separator
	parts := strings.Split(key, ".")

	// Start with the top-level data map
	var value interface{} = c.data

	// Traverse through the map using each part of the key
	for _, part := range parts {
		// Attempt to access the next level of the map

		// Attempt to access the next level of the map
		mapValue, ok := value.(map[string]interface{})
		if !ok {
			return nil
		}

		// Check if the next part of the key is in the map
		value, ok = mapValue[part]
		// Return nil if any part of the key path is invalid
		if !ok || value == nil {
			return nil
		}
	}
	// Return the final value found at the end of the key path
	return value
}

// contains checks if a value is present in a slice.
// It returns true if the value is found, otherwise false.
//
// This function is type-safe and works with slices of any comparable type.
func contains[T comparable](slice []T, value T) bool {
	// Iterate over the slice and check if the value is present
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// getDefaultValue returns the default value from a configuration map.
// If the default value does not exist, an empty string is returned.
func getDefaultValue(config map[string]interface{}) interface{} {
	defaultValue, exists := config["default"]
	if exists {
		return defaultValue
	}
	return ""
}

// The method takes a map with at least one of the following keys:
//   - env: the name of an environment variable
//   - default: a default value to use if the environment variable does not exist
//   - format: a list of accepted formats for the value
func validateAndSetValue(config map[string]interface{}) (interface{}, error) {
	var value interface{}

	// Check if the environment variable exists
	if envKey, exists := config["env"].(string); exists {
		if envValue, exists := os.LookupEnv(envKey); exists {
			value = envValue
		} else {
			value = getDefaultValue(config)
		}
	} else {
		value = getDefaultValue(config)
	}

	// Check if the value is in the expected format
	if expectedFormat, exists := config["format"]; exists {
		if !isValueInExpectedFormat(value, expectedFormat) {
			return nil, errors.New("value is not in the expected format")
		}
	}

	return value, nil
}

// isValueInExpectedFormat checks if a value is in the expected format.
// It verifies if the value matches any of the accepted formats provided in the format parameter.
// The function returns true if the value is in the expected format, otherwise it returns false.
func isValueInExpectedFormat(value interface{}, format interface{}) bool {
	// Get the type of the value
	valueType := reflect.TypeOf(value)

	// Check if the format is a list of accepted formats
	if formats, ok := format.([]interface{}); ok {
		// Check if the value is in the list of accepted formats
		return contains(formats, value)
	}

	// Check if the format is a string
	if _, ok := format.(string); ok {
		// Check if the value is a string
		return valueType == reflect.TypeOf("")
	}

	// Check if the format is a boolean
	if _, ok := format.(bool); ok {
		// Check if the value is a boolean
		return valueType == reflect.TypeOf(true)
	}

	// Check if the format is a float64
	if _, ok := format.(float64); ok {
		// Check if the value is a float64
		return valueType == reflect.TypeOf(float64(0))
	}

	// Check if the format is an int
	if _, ok := format.(int); ok {
		// Check if the value is an int
		return valueType == reflect.TypeOf(int(0))
	}

	// If none of the above conditions are met, return false
	return false
}

// processEnvRecursively iterates over the config map and replaces 'env' key values with environment variable values or default
// It returns a new map with the replaced values and an error if any of the values are not in their specified format
func (c *Config) processEnvRecursively(config map[string]interface{}) (map[string]interface{}, error) {
	// Create a new map to store the processed values
	processedConfig := make(map[string]interface{})

	// Iterate over the config map
	for key, value := range config {
		// Check if the value is a nested map
		switch typedValue := value.(type) {
		case map[string]interface{}:
			// If the nested map has an 'env' key, replace the value with the environment variable value or default
			if hasKey(typedValue, "default") || hasKey(typedValue, "env") {
				processedValue, err := validateAndSetValue(typedValue)
				if err != nil {
					return nil, err
				}
				processedConfig[key] = processedValue
			} else {
				// If the nested map does not have an 'env' key, recursively process the nested map
				nestedConfig, err := c.processEnvRecursively(typedValue)
				if err != nil {
					return nil, err
				}
				processedConfig[key] = nestedConfig
			}
		case []interface{}:
			// Check if the value is a slice
			processedSlice := make([]interface{}, len(typedValue))
			// Iterate over the slice and check if any of the items are nested maps
			for index, item := range typedValue {
				switch nestedItem := item.(type) {
				case map[string]interface{}:
					// If an item is a nested map, recursively process the nested map
					processedItem, err := c.processEnvRecursively(nestedItem)
					if err != nil {
						return nil, err
					}
					processedSlice[index] = processedItem
				default:
					// If an item is not a nested map, add it to the processed slice as is
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

// hasKey checks if a map has a given key
//
// This method takes a map and a key as input and returns true if the key exists in the map, and false otherwise.
func hasKey(config map[string]interface{}, key string) bool {
	_, exists := config[key]
	return exists
}
