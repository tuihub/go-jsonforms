package form

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/TobiEiss/go-jsonforms/models"
)

//go:embed html/*
var resources embed.FS

var funcs = template.FuncMap{
	"json": func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	},
}

type Form struct {
	schema             *gabs.Container
	uiSchema           *gabs.Container
	data               *gabs.Container
	menu               []models.MenuItem
	postLink           string
	cssPath            string
	logoPath           string
	confirmation       models.Confirmation
	customTemplateFS   embed.FS
	customTemplateDir  string
	useCustomTemplates bool
}

func NewForm(schema, uiSchema *gabs.Container) (*Form, error) {
	form := &Form{schema: schema, uiSchema: uiSchema}
	err := form.setup()
	return form, err
}

func NewFormWithCustomTemplates(schema, uiSchema *gabs.Container, templateFS embed.FS, templateDir string, useCustom bool) (*Form, error) {
	form := &Form{
		schema:             schema,
		uiSchema:           uiSchema,
		customTemplateFS:   templateFS,
		customTemplateDir:  templateDir,
		useCustomTemplates: useCustom,
	}
	err := form.setup()
	return form, err
}

func (f *Form) setup() error {
	var err error

	// add schema-information as schema to every control
	iterateObj(f.uiSchema, "type", "Control", func(c *gabs.Container) {
		scope, ok := c.Path("scope").Data().(string)
		if !ok {
			err = errors.New(fmt.Sprintf("in %v is no scope", c))
		}

		for k, v := range f.schema.Path(gabsPath(scope, true)).ChildrenMap() {
			// "simple" (not nested) object
			if len(v.Children()) == 0 {
				c.SetP(v, fmt.Sprintf("schema.%s", k))
			}
			// arrays (for e.g. items)
			if _, err := v.ArrayCount(); err == nil {
				c.SetP(v, fmt.Sprintf("schema.%s", k))
			}
		}
	})

	// add HTML-col-tag
	iterateObj(f.uiSchema, "type", nil, func(c *gabs.Container) {
		cType, ok := c.Path("type").Data().(string)
		if !ok {
			return
		}

		if cType != "HorizontalLayout" && cType != "VerticalLayout" {
			return
		}

		arrayCount, err := c.ArrayCountP("elements")
		if err != nil {
			return
		}

		col := 12
		if cType == "HorizontalLayout" {
			col = 12 / arrayCount
		}

		tag := fmt.Sprintf(" column col-%d", col)
		for i := range arrayCount {
			c.SetP(tag, fmt.Sprintf("elements.%d.schema.col", i))
		}
	})

	return err
}

func (f *Form) BindData(data *gabs.Container) error {
	var err error

	f.data = data

	// build multiple items for arrays
	arrayObj, _ := gabs.New().Set("array")
	iterateObj(f.uiSchema, "schema.type", arrayObj, func(c *gabs.Container) {
		scope, ok := c.Path("scope").Data().(string)
		if !ok {
			err = errors.New(fmt.Sprintf("can't find scope for %v", c))
		}

		// how many data are there
		arrayCount, err := f.data.ArrayCountP(gabsPath(scope, false))
		if err != nil {
			return
		}

		// create new array
		origin := c.Path("options.detail").String()
		for i := range arrayCount {
			copy := strings.ReplaceAll(origin, "items/", fmt.Sprintf("%d/", i))
			newDetails, _ := gabs.ParseJSON([]byte(copy))
			c.ArrayAppendP(newDetails, "options.details")
		}
		c.DeleteP("options.detail")
	})

	// I don't know why....
	f.uiSchema, _ = gabs.ParseJSON([]byte(f.uiSchema.String()))

	// array-select options
	iterateObj(f.uiSchema, "schema.type", "array-select", func(c *gabs.Container) {
		scope, ok := c.Path("scope").Data().(string)
		if !ok {
			err = errors.New(fmt.Sprintf("in %v is no scope", c))
		}

		// create map in data
		iterateArray(f.data, gabsPath(scope, false), func(dataElement *gabs.Container) {
			label := []string{}
			iterateArray(c, "options.elementLabelProps", func(labelElement *gabs.Container) {
				path := labelElement.Data().(string)
				label = append(label, dataElement.Path(gabsPath(path, false)).Data().(string))
			})
			c.SetP(dataElement, fmt.Sprintf("data.%s", strings.Join(label, " ")))
		})

		// clean pahts for scope in detail.elements
		iterateObj(c.Path("options.detail.elements"), "type", "Control", func(control *gabs.Container) {
			scope, ok := control.Path("scope").Data().(string)
			if !ok {
				err = errors.New(fmt.Sprintf("in %v is no scope", control))
			}

			control.SetP(strings.SplitN(gabsPath(scope, false), "items.", 2)[1], "scope")
		})
	})

	// add data to every control
	iterateObj(f.uiSchema, "type", "Control", func(c *gabs.Container) {
		// ignore array-controls
		schemaType := c.Path("schema.type").Data()
		if reflect.DeepEqual(schemaType, "array") || reflect.DeepEqual(schemaType, "array-select") {
			return
		}

		scope, ok := c.Path("scope").Data().(string)
		if !ok {
			err = errors.New(fmt.Sprintf("in %v is no scope", c))
		}

		if data := f.data.Path(gabsPath(scope, false)).Data(); data != nil {
			c.SetP(data, "data")
		}
	})
	return err
}

func (f *Form) SetMenu(menu []models.MenuItem) {
	f.menu = menu
}

func (f *Form) SetCSS(cssPath string) {
	f.cssPath = cssPath
}

func (f *Form) SetLogo(logoPath string) {
	f.logoPath = logoPath
}

func (f *Form) SetPostLink(link string) {
	f.postLink = link
}

func (f *Form) SetConfirmation(c models.Confirmation) {
	f.confirmation = c
}

func (f *Form) BuildContent() (string, error) {
	return f.build("raw.html")
}

func (f *Form) BuildIndex() (string, error) {
	return f.build("index.html")
}

func (f *Form) build(file string) (string, error) {
	var builder strings.Builder
	var err error
	var tmpl *template.Template

	if f.useCustomTemplates && f.customTemplateDir != "" {
		tmpl, err = template.New("").Funcs(funcs).ParseGlob(f.customTemplateDir + "/*.html")
	} else {
		// Use default embedded templates
		tmpl, err = template.New("").Funcs(funcs).ParseFS(resources, "html/*")
	}

	if err != nil {
		return builder.String(), err
	}

	var uischema map[string]interface{}
	if err := json.Unmarshal(f.uiSchema.Bytes(), &uischema); err != nil {
		return "", err
	}

	err = tmpl.ExecuteTemplate(&builder, file, map[string]interface{}{
		"UISchema":     uischema,
		"Menu":         f.menu,
		"Css":          f.cssPath,
		"Logo":         f.logoPath,
		"PostLink":     f.postLink,
		"Confirmation": f.confirmation,
	})

	return builder.String(), err
}

func ReadForm(urlForm url.Values) *gabs.Container {
	jsonObj := gabs.New()

	for key, value := range urlForm {
		path := gabsPath(key, false)

		val := value[0]
		if numVal, err := strconv.Atoi(val); err == nil {
			jsonObj.SetP(numVal, path)
		} else {
			jsonObj.SetP(val, path)
		}
	}

	return jsonObj
}

func (form *Form) UISchema() []byte {
	return form.uiSchema.Bytes()
}

// iterateObj searches for a key and value. If value is empty, it looks only for the key
func iterateObj(container *gabs.Container, key string, value any, operate func(c *gabs.Container)) {
	val := container.Path(key).Data()
	if val != nil && (value == nil || reflect.DeepEqual(val, value)) {
		operate(container)
	}

	for _, child := range container.Children() {
		iterateObj(child, key, value, operate)
	}
}

func gabsPath(scope string, withProperties bool) string {
	scope = strings.Trim(scope, "#/")
	if !withProperties {
		scope = strings.ReplaceAll(scope, "properties/", "")
	}
	path := strings.ReplaceAll(scope, "/", ".")
	return path
}

func iterateArray(container *gabs.Container, path string, operate func(*gabs.Container)) error {
	numberOfItems, err := container.ArrayCountP(path)
	if err != nil {
		fmt.Printf("err: %s %v\n", path, err)
		return err
	}

	for i := range numberOfItems {
		element, err := container.ArrayElementP(i, path)
		if err != nil {
			return err
		}
		operate(element)
	}
	return nil
}
