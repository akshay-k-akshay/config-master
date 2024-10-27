package configmaster

import (
	"os"
	"reflect"
	"testing"
)

func TestprocessRecursively(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		want    map[string]interface{}
		wantErr bool
		envVars map[string]string
	}{
		{
			name: "simple config map",
			config: map[string]interface{}{
				"foo": "bar",
			},
			want: map[string]interface{}{
				"foo": "bar",
			},
			wantErr: false,
		},
		{
			name: "nested config map",
			config: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz",
				},
			},
			want: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz",
				},
			},
			wantErr: false,
		},
		{
			name: "config map with slice",
			config: map[string]interface{}{
				"foo": []interface{}{"bar", "baz"},
			},
			want: map[string]interface{}{
				"foo": []interface{}{"bar", "baz"},
			},
			wantErr: false,
		},
		{
			name: "config map with environment variable",
			config: map[string]interface{}{
				"foo": map[string]interface{}{
					"env":     "FOO_ENV",
					"default": "bar",
				},
			},
			want: map[string]interface{}{
				"foo": "baz",
			},
			wantErr: false,
			envVars: map[string]string{
				"FOO_ENV": "baz",
			},
		},
		{
			name: "config map with invalid environment variable",
			config: map[string]interface{}{
				"foo": map[string]interface{}{
					"format":  []interface{}{"foo", "baz"},
					"env":     "FOO_ENV",
					"default": "bxb",
				},
			},
			want:    nil,
			wantErr: true,
			envVars: map[string]string{
				"FOO_ENV": "bar",
			},
		},
		{
			name: "config map with value not in expected format",
			config: map[string]interface{}{
				"foo": map[string]interface{}{
					"format":  []interface{}{"foo", "baz"},
					"default": "bxb",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}
			c := &Config{}
			got, err := c.processRecursively(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("processRecursively() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processRecursively() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewConfigWithValidJSONFile(t *testing.T) {
	// Create a temporary JSON file
	tmpFile, err := os.CreateTemp("", "config-master-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write a valid JSON config to the file
	configJSON := `{"foo": "bar"}`
	_, err = tmpFile.Write([]byte(configJSON))
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Test NewConfig with the JSON file
	cfg, err := NewConfig(tmpFile.Name())
	if err != nil {
		t.Errorf("NewConfig returned error: %v", err)
	}
	if cfg == nil {
		t.Errorf("NewConfig returned nil config")
	}
}

func TestNewConfigWithValidJSONMap(t *testing.T) {
	// Create a valid JSON map
	configMap := map[string]interface{}{
		"foo": "bar",
	}

	// Test NewConfig with the JSON map
	cfg, err := NewConfig(configMap)
	if err != nil {
		t.Errorf("NewConfig returned error: %v", err)
	}
	if cfg == nil {
		t.Errorf("NewConfig returned nil config")
	}
}

func TestNewConfigWithInvalidJSONFile(t *testing.T) {
	// Create a temporary JSON file with invalid JSON
	tmpFile, err := os.CreateTemp("", "config-master-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write invalid JSON to the file
	_, err = tmpFile.Write([]byte(`{"foo": "bar"`))
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Test NewConfig with the invalid JSON file
	_, err = NewConfig(tmpFile.Name())
	if err == nil {
		t.Errorf("NewConfig did not return error for invalid JSON file")
	}
}

func TestNewConfigWithNonExistentFile(t *testing.T) {
	// Test NewConfig with a non-existent file
	_, err := NewConfig("non-existent-file.json")
	if err == nil {
		t.Errorf("NewConfig did not return error for non-existent file")
	}
}

func TestNewConfigWithUnsupportedInputType(t *testing.T) {
	// Test NewConfig with an unsupported input type
	_, err := NewConfig(123)
	if err == nil {
		t.Errorf("NewConfig did not return error for unsupported input type")
	}
}

func TestNewConfigWithJSONFileContainingEnvironmentVariables(t *testing.T) {
	// Create a temporary JSON file with environment variables
	tmpFile, err := os.CreateTemp("", "config-master-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write a JSON config with environment variables to the file
	configJSON := `{"foo": {"env": "FOO"}}`
	_, err = tmpFile.Write([]byte(configJSON))
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Set the environment variable
	os.Setenv("FOO", "bar")

	// Test NewConfig with the JSON file
	cfg, err := NewConfig(tmpFile.Name())
	if err != nil {
		t.Errorf("NewConfig returned error: %v", err)
	}
	if cfg == nil {
		t.Errorf("NewConfig returned nil config")
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		key     string
		want    interface{}
		wantErr bool
	}{
		{
			name: "get top-level value",
			config: map[string]interface{}{
				"foo": "bar",
			},
			key:  "foo",
			want: "bar",
		},
		{
			name: "get nested value",
			config: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz",
				},
			},
			key:  "foo.bar",
			want: "baz",
		},
		{
			name: "get non-existent key",
			config: map[string]interface{}{
				"foo": "bar",
			},
			key:     "quux",
			want:    nil,
			wantErr: false,
		},
		{
			name: "get key with empty string",
			config: map[string]interface{}{
				"foo": "",
			},
			key:  "foo",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{data: tt.config}
			got := cfg.Get(tt.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNested(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want interface{}
		data map[string]interface{}
	}{
		{
			name: "valid key",
			key:  "foo.bar",
			want: "baz",
			data: map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}},
		},
		{
			name: "invalid key",
			key:  "foo.bar.baz",
			want: nil,
			data: map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}},
		},
		{
			name: "key with multiple parts",
			key:  "foo.bar.qux",
			want: "quux",
			data: map[string]interface{}{"foo": map[string]interface{}{"bar": map[string]interface{}{"qux": "quux"}}},
		},
		{
			name: "key with single part",
			key:  "foo",
			want: map[string]interface{}{"bar": "baz"},
			data: map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}},
		},
		{
			name: "key that does not exist",
			key:  "foo.bar.baz.qux",
			want: nil,
			data: map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{data: tt.data}
			got := c.getNested(tt.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNested(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestValidateAndSetValue(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]interface{}
		wantValue interface{}
		wantErr   bool
		envVars   map[string]string
	}{
		{
			name: "existing env var",
			config: map[string]interface{}{
				"env": "TEST_ENV",
			},
			wantValue: "test value",
			envVars: map[string]string{
				"TEST_ENV": "test value",
			},
		},
		{
			name: "non-existing env var with default",
			config: map[string]interface{}{
				"env":     "TEST_ENV",
				"default": "default value",
			},
			wantValue: "default value",
		},
		{
			name: "non-existing env var without default",
			config: map[string]interface{}{
				"env": "TEST_ENV",
			},
			wantValue: "",
		},
		{
			name: "invalid format",
			config: map[string]interface{}{
				"env":     "TEST_ENV",
				"format":  []interface{}{"valid value"},
				"default": "invalid value",
			},
			wantErr: true,
			envVars: map[string]string{
				"TEST_ENV": "invalid value",
			},
		},
		{
			name: "valid format",
			config: map[string]interface{}{
				"env":     "TEST_ENV",
				"format":  []interface{}{"valid value"},
				"default": "valid value",
			},
			wantValue: "valid value",
			envVars: map[string]string{
				"TEST_ENV": "valid value",
			},
		},
		{
			name: "missing env key",
			config: map[string]interface{}{
				"default": "default value",
			},
			wantValue: "default value",
		},
		{
			name: "missing format key",
			config: map[string]interface{}{
				"env":     "TEST_ENV",
				"default": "default value",
			},
			wantValue: "default value",
			envVars: map[string]string{
				"TEST_ENV": "default value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				// Unset environment variables
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			value, err := validateAndSetValue(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAndSetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(value, tt.wantValue) {
				t.Errorf("validateAndSetValue() value = %v, want %v", value, tt.wantValue)
			}
		})
	}
}

func TestIsValueInExpectedFormat(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		format   interface{}
		expected bool
	}{
		{
			name:     "value in list of accepted formats",
			value:    "foo",
			format:   []interface{}{"foo", "bar", "baz"},
			expected: true,
		},
		{
			name:     "value is a string",
			value:    "hello",
			format:   "string",
			expected: true,
		},
		{
			name:     "value is a boolean",
			value:    true,
			format:   "bool",
			expected: true,
		},
		{
			name:     "value is a float64",
			value:    3.14,
			format:   "float64",
			expected: true,
		},
		{
			name:     "value is an int",
			value:    42,
			format:   "int",
			expected: true,
		},
		{
			name:     "value is not in expected format",
			value:    "hello",
			format:   42,
			expected: false,
		},
		{
			name:     "format is not a list, string, boolean, float64, or int",
			value:    "hello",
			format:   struct{}{},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := isValueInExpectedFormat(test.value, test.format)
			if actual != test.expected {
				t.Errorf("expected %v, got %v", test.expected, actual)
			}
		})
	}
}

func TestIsNestedMap(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]interface{}
		want   bool
	}{
		{
			name: "nested key",
			config: map[string]interface{}{
				"foo": map[string]interface{}{"bar": "baz"},
			},
			want: true,
		},
		{
			name: "multi nested key",
			config: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{"baz": "qux"},
				},
			},
			want: true,
		},
		{
			name: "no nested key",
			config: map[string]interface{}{
				"foo": "bar",
			},
			want: false,
		},
		{
			name:   "empty map",
			config: map[string]interface{}{},
			want:   false,
		},
		{
			name:   "nil map",
			config: nil,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNestedMap(tt.config)
			if got != tt.want {
				t.Errorf("isNestedMap(%v) = %v, want %v", tt.config, got, tt.want)
			}
		})
	}
}
