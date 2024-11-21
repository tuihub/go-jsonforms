package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	gojsonforms "github.com/TobiEiss/go-jsonforms"
	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	schema   = "testdata/basic/schema.json"
	uiSchema = "testdata/basic/uischema.json"
	data     = "testdata/basic/data.json"
)

func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {

		html, err := gojsonforms.NewBuilder().
			WithSchemaFile(schema).
			WithUISchemaFile(uiSchema).
			WithDataFile(data).
			Build(true)
		if err != nil {
			fmt.Println("Error:", err.Error())
		}

		fmt.Fprintf(w, html)
	})
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}

		result := gojsonforms.Verify(r.Form)
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}
		fmt.Println(string(jsonData))
	})

	log.Fatal(http.ListenAndServe("localhost:8080", router))
}
