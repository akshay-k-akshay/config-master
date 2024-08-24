# Config Master

<!-- <img align="right" width="159px" src=""> -->

<!-- [![Build Status](https://github.com/akshay-k-akshay/config-master/workflows/Run%20Tests/badge.svg?branch=master)](https://github.com/akshay-k-akshay/config-master/actions?query=branch%3Amaster) -->
[![Go Report Card](https://goreportcard.com/badge/github.com/akshay-k-akshay/config-master)](https://goreportcard.com/report/github.com/akshay-k-akshay/config-master)
[![Go Reference](https://pkg.go.dev/badge/github.com/akshay-k-akshay/config-master?status.svg)](https://pkg.go.dev/github.com/akshay-k-akshay/config-master?tab=doc)
[![Release](https://img.shields.io/github/release/akshay-k-akshay/config-master.svg?style=flat-square)](https://github.com/akshay-k-akshay/config-master/releases)

In [Go](https://go.dev/), a schema-based configuration management approach enhances the standard pattern of configuring applications, making it more robust and accessible to collaborators who may not want to dive into the code to inspect or modify settings. By introducing a configuration schema, this method provides clear documentation and context for each setting, enabling validation and early failure detection when configurations go wrong. This schema defines the expected configuration keys, their types, default values, and constraints, ensuring that configurations are both flexible and reliable. Additionally, it supports environment variable substitution and default handling, allowing for environment-specific configurations while maintaining sensible defaults. This approach simplifies the configuration process, making it easier for project collaborators to understand and manage settings.

**ConfigMaster's key features are:**

- ***Loading and merging***: configurations are loaded from disk or inline and merged
- ***Nested structure***: keys and values can be organized in a tree structure
- ***Environmental variables***: values can be derived from environmental variables
- ***Validation***: configurations are validated against your schema (presence checking, type checking, custom checking), generating an error report with all errors that are found
- ***Comments allowed***: schema and configuration files can be either in the JSON format or in the newer JSON5 format, so comments are welcome

## Getting started

### Prerequisites

ConfigMaster requires [Go](https://go.dev/) version [1.23](https://go.dev/doc/devel/release#go1.23.0) or above.

### Getting ConfigMaster

With [Go's module support](https://go.dev/wiki/Modules#how-to-use-modules), `go [build|run|test]` automatically fetches the necessary dependencies when you add the import in your code:

```sh
import "github.com/akshay-k-akshay/config-master"
```

Alternatively, use `go get`:

```sh
go get -u github.com/akshay-k-akshay/config-master
```

### Running configMaster

A basic example: 

```go
package config

import configmaster "github.com/akshay-k-akshay/config-master"

var config *configmaster.Config

func init() {
	var err error
  configMap := map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": map[string]interface{}{
				"format":  []interface{}{"bar", "baz", "foo"},
				"default": "baz",
				"env":     "QUUX",
        "doc":     "some description"
			},
		},
	}

	config, err = configmaster.NewConfig()
	if err != nil {
		panic(err)
	}
}

func get(key string) interface{} {
	return config.Get(key)
}

```

A basic example: with json file

```go
package config

import configmaster "github.com/akshay-k-akshay/config-master"

var config *configmaster.Config

func init() {
	var err error
	config, err = configmaster.NewConfig("config.json")
	if err != nil {
		panic(err)
	}
}

func get(key string) interface{} {
	return config.Get(key)
}

```

example json file config.json
```json
{
  "foo": {
    "format": ["bar", "baz", "foo"],
    "default": "bar",
    "env": "FOO"
  }
}
```
To run the code, use the `go run` command, like:

```sh
$ go run example.go
```