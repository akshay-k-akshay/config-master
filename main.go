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
	err := cfg.ProcessEnv() // Automatically process environment variables
	if err != nil {
		return nil, fmt.Errorf("[Config-Master]: %w", err)
	}
	return cfg, nil
}

// Get retrieves a value from the configuration by key
func (c *Config) Get(key string) interface{} {
	if strings.Contains(key, ".") {
		ar := strings.Split(key, ".")
		var o interface{}
		o = c.data
		for _, k := range ar {
			o = o.(map[string]interface{})[k]
			if o == nil {
				return nil
			}
		}
		return o
	}
	return c.data[key]
}

// ProcessEnv replaces values associated with the 'env' key with environment variable values or fallback to default
func (c *Config) ProcessEnv() error {
	err := c.processEnvRecursively(c.data)
	if err != nil {
		return err
	}
	return nil
}

func contains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func setValue(envKey string, v map[string]interface{}) (interface{}, error) {
	var value interface{}
	if envValue, ok := os.LookupEnv(envKey); ok {
		value = envValue
	} else {
		if defaultValue, defaultExists := v["default"].(string); defaultExists {
			value = defaultValue
		} else {
			value = ""
		}
	}
	valueType := reflect.TypeOf(value)
	if format, formatExists := v["format"]; formatExists {
		switch v := format.(type) {
		case []interface{}:
			if !contains(v, value) {
				return 0, errors.New("value is not in format")
			}
		case string:
			if valueType != reflect.TypeOf("") {
				return 0, errors.New("value should be string")
			}
		case bool:
			if valueType != reflect.TypeOf(true) && valueType != reflect.TypeOf(false) {
				return 0, errors.New("value should be bool")
			}
		case float64:
			if valueType != reflect.TypeOf(float64(0)) {
				return 0, errors.New("value should be float64")
			}
		case int:
			if valueType != reflect.TypeOf(int(0)) {
				return 0, errors.New("value should be int")
			}
		}
	}
	return value, nil
}

// processEnvRecursively iterates over the config map and replaces 'env' key values with environment variable values or default
func (c *Config) processEnvRecursively(config map[string]interface{}) error {
	for key, value := range config {
		switch v := value.(type) {
		case map[string]interface{}:
			if envKey, exists := v["env"].(string); exists {
				value, err := setValue(envKey, v)
				if err != nil {
					return err
				}
				config[key] = value
			} else {
				c.processEnvRecursively(v)
			}
		default:
			config[key] = value
		}
	}

	return nil
}
