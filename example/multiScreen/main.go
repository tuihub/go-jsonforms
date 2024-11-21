package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	gojsonforms "github.com/TobiEiss/go-jsonforms"
	"github.com/TobiEiss/go-jsonforms/models"
	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var menu = []models.MenuItem{
	{
		Link:  "basic",
		Titel: "Basic",
	},
	{
		Link:  "control",
		Titel: "Control",
	},
	{
		Link:  "array",
		Titel: "Array Forms",
	},
}

func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/{screen:(basic|control|array)*}", func(w http.ResponseWriter, r *http.Request) {
		screenID := chi.URLParam(r, "screen")
		if screenID == "" {
			screenID = "basic"
		}

		for i := range menu {
			menu[i].Current = (menu[i].Link == screenID)
		}

		html, err := gojsonforms.NewBuilder().
			WithSchemaFile(fmt.Sprintf("testdata/%s/schema.json", screenID)).
			WithUISchemaFile(fmt.Sprintf("testdata/%s/uischema.json", screenID)).
			WithDataFile(fmt.Sprintf("testdata/%s/data.json", screenID)).
			WithMenu(menu).
			Build(true)
		if err != nil {
			panic(err)
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
