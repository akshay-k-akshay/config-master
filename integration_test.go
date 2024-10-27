package configmaster

import (
	"reflect"
	"testing"
)

func TestWithBasicFlatJson(t *testing.T) {
	directMap := map[string]interface{}{
		"foo": "bar",
	}
	config, err := NewConfig(directMap)
	if err != nil {
		t.Fatalf(`NewConfig() = %v, want nil`, err)
	}
	value := config.Get("foo")
	want := `bar`
	if value != want {
		t.Fatalf(`config.Get("foo") should be "%v", got "%v"`, want, value)
	}
}

func TestWithBasicJson(t *testing.T) {
	directMap := map[string]interface{}{
		"foo": map[string]interface{}{
			"default": "bar",
			"env":     "FOO",
		},
	}
	config, err := NewConfig(directMap)
	if err != nil {
		t.Fatalf(`NewConfig() = %v, want nil`, err)
	}
	value := config.Get("foo")
	want := `bar`
	if value != want {
		t.Fatalf(`config.Get("foo") should be "%v", got "%v"`, want, value)
	}
}

func TestWithValueFromEnv(t *testing.T) {
	directMap := map[string]interface{}{
		"foo": map[string]interface{}{
			"default": "bar",
			"env":     "FOO_ENV",
		},
	}
	// set Env VAlue
	want := "bar from env"
	t.Setenv("FOO_ENV", want)
	config, err := NewConfig(directMap)
	if err != nil {
		t.Fatalf(`NewConfig() = %v, want nil`, err)
	}
	value := config.Get("foo")
	if value != want {
		t.Fatalf(`config.Get("foo") should be "%v", got "%v"`, want, value)
	}
}

func TestWithValueIsInFormat(t *testing.T) {
	directMap := map[string]interface{}{
		"foo": map[string]interface{}{
			"format":  []interface{}{"bar", "baz", "foo"},
			"default": "bar test",
			"env":     "FOO",
		},
	}
	// set Env VAlue
	config, err := NewConfig(directMap)
	if err == nil {
		t.Fatalf(`should throw error if value is not in format %v`, err)
	}

	if config != nil {
		t.Fatalf("should not return config as value is not in format")
	}
}

func TestWithEnvValueIsInFormat(t *testing.T) {
	directMap := map[string]interface{}{
		"foo": map[string]interface{}{
			"format":  []interface{}{"bar", "baz", "foo"},
			"default": "bar",
			"env":     "FOO_ENV",
		},
	}
	// set Env VAlue
	t.Setenv("FOO_ENV", "bar test")
	config, err := NewConfig(directMap)
	if err == nil {
		t.Fatalf(`should throw error if value is not in format %v`, err)
	}

	if config != nil {
		t.Fatalf("should not return config as value is not in format")
	}
}

func TestWithNestedConfig(t *testing.T) {
	directMap := map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": map[string]interface{}{
				"format":  []interface{}{"bar", "baz", "foo"},
				"default": "bar",
				"env":     "BAR",
			},
			"quux": map[string]interface{}{
				"format":  []interface{}{"bar", "baz", "foo"},
				"default": "baz",
				"env":     "QUUX",
			},
		},
	}
	config, err := NewConfig(directMap)

	if err != nil {
		t.Fatalf(`should throw error if value is not in format %v`, err)
	}

	value := config.Get("foo")
	want := map[string]interface{}{
		"bar":  "bar",
		"quux": "baz",
	}
	if !reflect.DeepEqual(value, want) {
		t.Fatalf(`config.Get("foo") should be "%v", got "%v"`, want, value)
	}
}

func TestGetNestedKey(t *testing.T) {
	directMap := map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": map[string]interface{}{
				"format":  []interface{}{"bar", "baz", "foo"},
				"default": "baz",
				"env":     "QUUX",
			},
		},
	}
	config, err := NewConfig(directMap)

	if err != nil {
		t.Fatalf(`should throw error if value is not in format %v`, err)
	}

	value := config.Get("foo.bar")

	want := "baz"
	if value != want {
		t.Fatalf(`config.Get("foo") should be "%v", got "%v"`, want, value)
	}
}

func TestNewConfigWithJsonFile(t *testing.T) {
	config, err := NewConfig("./sample-config.json")
	if err != nil {
		t.Fatalf(`should throw error if value is not in format %v`, err)
	}
	value := config.Get("foo")

	want := "bar"
	if value != want {
		t.Fatalf(`config.Get("foo") should be "%v", got "%v"`, want, value)
	}
}

func TestNewConfigWithAdvancedJsonFile(t *testing.T) {
	t.Setenv("FOO", "bar")
	t.Setenv("QUX_BAR", "BAR-QUX")
	t.Setenv("XYZ_BAR", "XYZ-BAR")

	config, err := NewConfig("./advanced-config.json")
	if err != nil {
		t.Fatalf(`should throw error if value is not in format %v`, err)
	}
	value := config.Get("foo")

	want := "bar"
	if value != want {
		t.Fatalf(`config.Get("foo") should be "%v", got "%v"`, want, value)
	}

	want = "FOO"
	value = config.Get("baz.xyz")
	if value != want {
		t.Fatalf(`config.Get("baz.xyz") should be "%v", got "%v"`, want, value)
	}

	want = "FOO-QUX"
	value = config.Get("qux.foo")
	if value != want {
		t.Fatalf(`config.Get("qux.foo") should be "%v", got "%v"`, want, value)
	}

	want = "BAR-QUX"
	value = config.Get("qux.bar")
	if value != want {
		t.Fatalf(`config.Get("qux.bar") should be "%v", got "%v"`, want, value)
	}

	wantStrings := []interface{}{"FOO-QUUX", "FOO-QUUX-BAZ"}
	value = config.Get("quux")
	if !reflect.DeepEqual(value, wantStrings) {
		t.Fatalf(`config.Get("quux") should be "%v", got "%v"`, wantStrings, value)
	}

	wantStrings = []interface{}{
		map[string]interface{}{
			"foo": "FOO-XYZ",
		},
		map[string]interface{}{
			"bar": "XYZ-BAR",
		},
	}
	value = config.Get("xyz")
	if !reflect.DeepEqual(value, wantStrings) {
		t.Fatalf(`config.Get("xyz") should be "%v", got "%v"`, wantStrings, value)
	}

}
