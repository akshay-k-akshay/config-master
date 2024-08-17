package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Config represents the configuration and provides methods to access it
type Config struct {
	data map[string]interface{}
}

// NewConfig creates a new Config instance from either a filename or a map
func NewConfig(input interface{}) (*Config, error) {
	var config map[string]interface{}

	if filename, ok := input.(string); ok {
		// If input is a filename
		file, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("error opening file: %w", err)
		}
		defer file.Close()

		byteValue, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}

		err = json.Unmarshal(byteValue, &config)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON from file: %w", err)
		}

	} else if configMap, ok := input.(map[string]interface{}); ok {
		// If input is already a map
		config = configMap

	} else {
		return nil, fmt.Errorf("unsupported input type: %T", input)
	}

	cfg := &Config{data: config}
	cfg.ProcessEnv() // Automatically process environment variables
	return cfg, nil
}

// Get retrieves a value from the configuration by key
func (c *Config) Get(key string) interface{} {
	return c.data[key]
}

// ProcessEnv replaces values associated with the 'env' key with environment variable values or fallback to default
func (c *Config) ProcessEnv() {
	c.processEnvRecursively(c.data)
}

// processEnvRecursively iterates over the config map and replaces 'env' key values with environment variable values or default
func (c *Config) processEnvRecursively(config map[string]interface{}) {
	for key, value := range config {
		switch v := value.(type) {
		case map[string]interface{}:
			// Process 'env' key if present
			if envKey, exists := v["env"].(string); exists {
				// Resolve environment variable
				if envValue, ok := os.LookupEnv(envKey); ok {
					config[key] = envValue
				} else {
					// If environment variable is not set, fall back to default
					if defaultValue, defaultExists := v["default"].(string); defaultExists {
						config[key] = defaultValue
					} else {
						config[key] = "" // or handle the missing default value case
					}
				}
			} else {
				// Recursively process nested maps
				c.processEnvRecursively(v)
			}
		case []interface{}:
			for i, item := range v {
				if subMap, ok := item.(map[string]interface{}); ok {
					c.processEnvRecursively(subMap)
					// Handle 'env' key at the array item level
					if envKey, exists := subMap["env"].(string); exists {
						if envValue, ok := os.LookupEnv(envKey); ok {
							v[i] = envValue
						} else {
							if defaultValue, defaultExists := subMap["default"].(string); defaultExists {
								v[i] = defaultValue
							} else {
								v[i] = "" // or handle the missing default value case
							}
						}
					}
				}
			}
		}
	}
}
