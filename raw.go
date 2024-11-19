package gojsonforms

import "encoding/json"

func BuildWithMap(schema map[string]interface{}, uiSchema map[string]interface{}, data map[string]interface{}, menu []MenuItem) (string, error) {
	schemaB, err := json.Marshal(schema)
	if err != nil {
		return "", err
	}

	uiSchemaB, err := json.Marshal(uiSchema)
	if err != nil {
		return "", err
	}

	dataB, err := json.Marshal(data)
	if err != err {
		return "", err
	}

	form, err := New(schemaB, uiSchemaB)
	if err != nil {
		return "", err
	}

	err = form.BindData(dataB)
	if err != nil {
		return "", err
	}

	form.SetMenu(menu)

	return form.Build()
}
