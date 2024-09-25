package goformsbuilder

import (
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"strings"
)

//go:embed html/*
var resources embed.FS

var templates = template.Must(template.New("").Funcs(templateFuncs).ParseFS(resources, "html/*"))
var templateFuncs = template.FuncMap{
	"find": func(object any, path string) string {
		fmt.Printf("object: %v\n", object)
		fmt.Printf("path: %v\n", path)
		return "abc"
	},
}

type SchemaJson map[string]interface{}

type Property struct {
	Type        string   `json:"type"`
	MinLength   int      `json:"minLength,omitempty"`
	MaxLength   int      `json:"maxLength,omitempty"`
	Format      string   `json:"format,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description,omitempty"`
}

type UISchema struct {
	Type     string      `json:"type"`
	Elements []UIElement `json:"elements"`
}

type UIElement struct {
	Type        string      `json:"type"`
	Scope       string      `json:"scope,omitempty"`
	Text        string      `json:"text,omitempty"`
	Elements    []UIElement `json:"elements,omitempty"`
	Suggestions []string    `json:"suggestion,omitempty"`
}

func BuildTemplate(schema SchemaJson, uiSchema UISchema) (string, error) {
	var builder strings.Builder

	find := func(path string) SchemaJson {
		pathSplits := strings.Split(path, "/")
		element := findElement(pathSplits[1:], schema)
		element["Label"] = pathSplits[len(pathSplits)-1]
		element["Scope"] = path

		return element
	}

	err := templates.Funcs(template.FuncMap{"find": find}).ExecuteTemplate(&builder, "index.html", map[string]interface{}{
		"UISchema": uiSchema,
		"Schema":   schema,
	})

	return builder.String(), err
}

func ReadForm(form url.Values) map[string]interface{} {
	result := map[string]interface{}{}

	for key, value := range form {
		path := strings.TrimPrefix(key, "#/")
		keys := strings.Split(path, "/")

		val := value[0]

		if numVal, err := strconv.Atoi(val); err == nil {
			setNestedKey(result, keys, numVal)
		} else {
			setNestedKey(result, keys, val)
		}
	}

	return result
}

func findElement(path []string, s map[string]interface{}) SchemaJson {
	if len(path) == 0 {
		return s
	}

	return findElement(path[1:], s[path[0]].(map[string]interface{}))
}

func setNestedKey(data map[string]interface{}, path []string, value interface{}) {
	if path[0] == "properties" {
		path = path[1:]
	}

	if len(path) == 1 {
		data[path[0]] = value
		return
	}

	if _, ok := data[path[0]]; !ok {
		data[path[0]] = make(map[string]interface{})
	}

	setNestedKey(data[path[0]].(map[string]interface{}), path[1:], value)
}
