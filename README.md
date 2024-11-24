# go-jsonforms

A Go implementation of [JSONForms](https://jsonforms.io/) - a framework for building forms based on JSON Schema and UI Schema.

[![Go Reference](https://pkg.go.dev/badge/github.com/TobiEiss/go-jsonforms.svg)](https://pkg.go.dev/github.com/TobiEiss/go-jsonforms)
[![Go Report Card](https://goreportcard.com/badge/github.com/TobiEiss/go-jsonforms)](https://goreportcard.com/report/github.com/TobiEiss/go-jsonforms)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- Generate HTML forms from JSON Schema and UI Schema
- Support for various form layouts (Vertical, Horizontal)
- Built-in menu system for multi-page forms
- Multiple input methods (JSON files, bytes, or Go maps)
- Embedded HTML templates
- Form data verification

## Installation

```bash
go get github.com/TobiEiss/go-jsonforms
```

## Quick Start

```go
package main

import (
    "github.com/TobiEiss/go-jsonforms"
)

func main() {
    // Define your schema
    schema := map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "name": map[string]interface{}{
                "type": "string",
                "description": "Please enter your name",
            },
        },
    }

    // Define UI schema
    uiSchema := map[string]interface{}{
        "type": "VerticalLayout",
        "elements": []map[string]interface{}{
            {
                "type": "Control",
                "scope": "#/properties/name",
            },
        },
    }

    // Create and render the form
    builder := gojsonforms.NewBuilder()
    html, err := builder.
        WithSchemaMap(schema).
        WithUISchemaMap(uiSchema).
        Build(false)
    if err != nil {
        panic(err)
    }

    // Use the generated HTML in your application
    fmt.Println(html)
}
```

## Documentation

### Builder Methods

- `WithSchemaMap(schema map[string]interface{})`: Set schema using a Go map
- `WithSchemaBytes(schema []byte)`: Set schema using JSON bytes
- `WithSchemaFile(filepath string)`: Set schema from a JSON file
- `WithUISchemaMap(uiSchema map[string]interface{})`: Set UI schema using a Go map
- `WithUISchemaBytes(uiSchema []byte)`: Set UI schema using JSON bytes
- `WithUISchemaFile(filepath string)`: Set UI schema from a JSON file
- `WithDataMap(data map[string]interface{})`: Set initial data using a Go map
- `WithDataBytes(data []byte)`: Set initial data using JSON bytes
- `WithDataFile(filepath string)`: Set initial data from a JSON file
- `WithMenu(menu []MenuItem)`: Add navigation menu items
- `Build(withIndex bool)`: Generate the HTML form

### Examples

Check the [example](./example) directory for complete working examples:
- [Basic Form](./example/basic/main.go)
- [Multi-screen Form](./example/multiScreen/main.go)

### Schema Examples

<details>
<summary>Basic Schema Example</summary>

```json
{
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "Please enter your name"
    },
    "age": {
      "type": "integer",
      "description": "Please enter your age"
    }
  },
  "required": ["name"]
}
```
</details>

<details>
<summary>UI Schema Example</summary>

```json
{
  "type": "VerticalLayout",
  "elements": [
    {
      "type": "Control",
      "scope": "#/properties/name"
    },
    {
      "type": "Control",
      "scope": "#/properties/age"
    }
  ]
}
```
</details>

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [JSONForms](https://jsonforms.io/) for the original implementation and specification
- [Gabs](https://github.com/Jeffail/gabs) for JSON parsing
