package goformsbuilder

import (
	"embed"
	"html/template"
	"net/url"
	"strconv"
	"strings"
)

//go:embed html/*
var resources embed.FS

type SchemaJson map[string]interface{}

func BuildTemplate(schema SchemaJson, uiSchema UISchema) (string, error) {
	var builder strings.Builder

	find := func(path string) SchemaJson {
		pathSplits := strings.Split(path, "/")
		element := findElement(pathSplits[1:], schema)
		element["Label"] = pathSplits[len(pathSplits)-1]
		element["Scope"] = path

		return element
	}

	colWidth := func(scope string) int {
		return 12 / len(uiSchema.Elements.FindElementWithChild(scope))
	}

	tmpl, err := template.New("").Funcs(template.FuncMap{"find": find, "colWidth": colWidth}).ParseFS(resources, "html/*")
	if err != nil {
		return builder.String(), err
	}

	err = tmpl.ExecuteTemplate(&builder, "index.html", map[string]interface{}{
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
