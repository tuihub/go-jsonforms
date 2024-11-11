package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	gojsonforms "github.com/TobiEiss/go-jsonforms"
	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var screens = map[string]struct {
	Schema   string
	UISchema string
}{
	"Basic": {
		Schema:   "testdata/basic/schema.json",
		UISchema: "testdata/basic/uischema.json",
	},
	"Control": {
		Schema:   "testdata/control/schema.json",
		UISchema: "testdata/control/uischema.json",
	},
}

func main() {
	schemaData, err := os.ReadFile(screens["Basic"].Schema)
	if err != nil {
		panic(err)
	}

	uiSchemaData, err := os.ReadFile(screens["Basic"].UISchema)
	if err != nil {
		panic(err)
	}

	var schema gojsonforms.SchemaJson
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		panic(err)
	}

	var uischema gojsonforms.UISchema
	if err := json.Unmarshal(uiSchemaData, &uischema); err != nil {
		panic(err)
	}

	html, err := gojsonforms.BuildTemplate(schema, uischema)
	if err != nil {
		panic(err)
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, html)
	})
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}

		result := gojsonforms.ReadForm(r.Form)
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}
		fmt.Println(string(jsonData))
	})

	log.Fatal(http.ListenAndServe("localhost:8080", router))
}
