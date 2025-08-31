package gojsonforms

import (
	"embed"
	"net/url"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/TobiEiss/go-jsonforms/internal/form"
	"github.com/TobiEiss/go-jsonforms/models"
)

type MenuItem struct {
	Link    string
	Titel   string
	Current bool
}

type builder struct {
	uiSchema           reader
	schema             reader
	data               reader
	menu               []models.MenuItem
	postLink           string
	cssPath            string
	logoPath           string
	confirmation       models.Confirmation
	customTemplateFS   embed.FS
	customTemplateDir  string
	useCustomTemplates bool
	customTemplateExt  string
}

type reader struct {
	Bytes []byte
	Map   map[string]interface{}
	File  string
}

type FormBuilder interface {
	WithUISchemaBytes(uiSchema []byte) *FormBuilder
	WithUISchemaMap(uiSchema map[string]interface{}) *FormBuilder
	WithUISchemaFile(filePath string) *FormBuilder

	WithSchemaBytes(schema []byte) *FormBuilder
	WithSchemaMap(schema map[string]interface{}) *FormBuilder
	WithSchemaFile(filePath string) *FormBuilder

	WithDataBytes(data []byte) *FormBuilder
	WithDataMap(data map[string]interface{}) *FormBuilder
	WithDataFile(filePath string) *FormBuilder

	WithMenu(menu []MenuItem) *FormBuilder
	WithCssPath(cssPath string) *FormBuilder
	WithLogoPath(logoPath string) *FormBuilder
	WithPostLink(link string) *FormBuilder
	WithConfirmation(confirmation models.Confirmation) *FormBuilder
	WithCustomTemplateFS(templateFS embed.FS) *FormBuilder
	WithCustomTemplateDir(templateDir string) *FormBuilder

	GetUISchema() []byte

	Build() (string, error)
}

func (r *reader) Read() (*gabs.Container, error) {
	if r.Bytes != nil {
		return gabs.ParseJSON(r.Bytes)
	} else if r.Map != nil {
		return gabs.Wrap(r.Map), nil
	} else if r.File != "" {
		return gabs.ParseJSONFile(r.File)
	}
	return nil, nil
}

func NewBuilder() *builder {
	return &builder{}
}

func (b *builder) Build(withIndex bool) (string, error) {
	var html string

	// schema is necessary
	schema, err := b.schema.Read()
	if err != nil {
		return html, err
	}

	// uischema is optional - generate default if not provided
	uiSchema, err := b.uiSchema.Read()
	if err != nil {
		return html, err
	}

	// if no uiSchema provided, generate default from schema
	if uiSchema == nil {
		uiSchema, err = generateDefaultUISchema(schema)
		if err != nil {
			return html, err
		}
	}

	var f *form.Form
	if b.useCustomTemplates {
		f, err = form.NewFormWithCustomTemplates(schema, uiSchema, b.customTemplateFS, b.customTemplateDir, b.useCustomTemplates)
	} else {
		f, err = form.NewForm(schema, uiSchema)
	}
	if err != nil {
		return html, err
	}

	if data, err := b.data.Read(); err != nil {
		return html, err
	} else if data != nil {
		f.BindData(data)
	}

	f.SetMenu(b.menu)
	f.SetCSS(b.cssPath)
	f.SetLogo(b.logoPath)
	f.SetPostLink(b.postLink)
	f.SetConfirmation(b.confirmation)
	f.SetCustomTemplateExt(b.customTemplateExt)

	if withIndex {
		return f.BuildIndex()
	}
	return f.BuildContent()
}

func Verify(urlForm url.Values) interface{} {
	return form.ReadForm(urlForm).Data()
}

func (b *builder) WithUISchemaBytes(uiSchema []byte) *builder {
	b.uiSchema.Bytes = uiSchema
	return b
}

func (b *builder) WithUISchemaMap(uiSchema map[string]interface{}) *builder {
	b.uiSchema.Map = uiSchema
	return b
}

func (b *builder) WithUISchemaFile(uiSchema string) *builder {
	b.uiSchema.File = uiSchema
	return b
}

func (b *builder) WithSchemaBytes(schema []byte) *builder {
	b.schema.Bytes = schema
	return b
}

func (b *builder) WithSchemaMap(schema map[string]interface{}) *builder {
	b.schema.Map = schema
	return b
}

func (b *builder) WithSchemaFile(schema string) *builder {
	b.schema.File = schema
	return b
}

func (b *builder) WithDataBytes(data []byte) *builder {
	b.data.Bytes = data
	return b
}

func (b *builder) WithDataMap(data map[string]interface{}) *builder {
	b.data.Map = data
	return b
}

func (b *builder) WithDataFile(data string) *builder {
	b.data.File = data
	return b
}

func (b *builder) WithMenu(menu []models.MenuItem) *builder {
	b.menu = menu
	return b
}

func (b *builder) WithCss(cssPath string) *builder {
	b.cssPath = cssPath
	return b
}

func (b *builder) WithLogo(logoPath string) *builder {
	b.logoPath = logoPath
	return b
}

func (b *builder) WithPostLink(link string) *builder {
	b.postLink = link
	return b
}

func (b *builder) WithConfirmation(c models.Confirmation) *builder {
	b.confirmation = c
	return b
}

func (b *builder) WithCustomTemplateFS(templateDir string, templateFS embed.FS) *builder {
	b.customTemplateDir = templateDir
	b.customTemplateFS = templateFS
	b.useCustomTemplates = true
	return b
}

func (b *builder) WithCustomTemplateExt(ext string) *builder {
	b.customTemplateExt = ext
	return b
}

// generateDefaultUISchema creates a default UI schema from a JSON schema
func generateDefaultUISchema(schema *gabs.Container) (*gabs.Container, error) {
	return generateUISchemaFromProperties(schema, "")
}

// generateUISchemaFromProperties recursively generates UI schema from properties
func generateUISchemaFromProperties(schema *gabs.Container, basePath string) (*gabs.Container, error) {
	defaultUISchema := gabs.New()

	// Set the root layout type
	defaultUISchema.SetP("VerticalLayout", "type")

	// Initialize elements array
	elements := make([]interface{}, 0)

	// Get properties from schema
	properties := schema.Path("properties")
	if properties != nil {
		for propertyName := range properties.ChildrenMap() {
			// Get the property schema
			propertySchema := properties.Path(propertyName)

			// Build the scope path
			var scope string
			if basePath == "" {
				scope = "#/properties/" + propertyName
			} else {
				scope = basePath + "/properties/" + propertyName
			}

			// Get the type of the property
			propertyType := propertySchema.Path("type").Data()

			if propertyType == "object" {
				// Handle nested objects
				nestedProperties := propertySchema.Path("properties")
				if nestedProperties != nil {
					// Create a group for nested object
					group := map[string]interface{}{
						"type":  "Group",
						"scope": scope,
					}

					// Add title from schema or generate from property name
					if title := propertySchema.Path("title").Data(); title != nil {
						group["label"] = title
					} else {
						group["label"] = propertyName
					}

					// Recursively generate UI schema for nested properties
					nestedElements := make([]interface{}, 0)
					for nestedPropertyName := range nestedProperties.ChildrenMap() {
						nestedPropertySchema := nestedProperties.Path(nestedPropertyName)
						nestedScope := scope + "/properties/" + nestedPropertyName

						nestedControl := map[string]interface{}{
							"type":  "Control",
							"scope": nestedScope,
						}

						// Add title from schema if available
						if nestedTitle := nestedPropertySchema.Path("title").Data(); nestedTitle != nil {
							nestedControl["title"] = nestedTitle
						}

						nestedElements = append(nestedElements, nestedControl)
					}

					group["elements"] = nestedElements
					elements = append(elements, group)
				}
			} else if propertyType == "array" {
				// Handle arrays
				control := map[string]interface{}{
					"type":  "Control",
					"scope": scope,
				}

				// Add title from schema or generate from property name
				if title := propertySchema.Path("title").Data(); title != nil {
					control["title"] = title
				}

				// Check if array items have object type for special handling
				itemsSchema := propertySchema.Path("items")
				if itemsSchema != nil {
					options := map[string]interface{}{}

					if itemsSchema.Path("type").Data() == "object" {
						// For array of objects, create detail layout
						itemsProperties := itemsSchema.Path("properties")
						if itemsProperties != nil {
							detailElements := make([]interface{}, 0)

							for itemPropertyName := range itemsProperties.ChildrenMap() {
								itemPropertySchema := itemsProperties.Path(itemPropertyName)
								itemControl := map[string]interface{}{
									"type":  "Control",
									"scope": scope + "/items/properties/" + itemPropertyName,
								}

								// Add title from schema if available
								if itemTitle := itemPropertySchema.Path("title").Data(); itemTitle != nil {
									itemControl["title"] = itemTitle
								}

								detailElements = append(detailElements, itemControl)
							}

							options["detail"] = map[string]interface{}{
								"type":     "VerticalLayout",
								"elements": detailElements,
							}
						}
					} else {
						// For arrays of primitive types, use simple detail
						options["detail"] = "GENERATED"
					}

					control["options"] = options
				}

				elements = append(elements, control)
			} else {
				// Handle primitive types (string, number, integer, boolean, etc.)
				control := map[string]interface{}{
					"type":  "Control",
					"scope": scope,
				}

				// Add title from schema if available
				if title := propertySchema.Path("title").Data(); title != nil {
					control["title"] = title
				}

				elements = append(elements, control)
			}
		}
	}

	// Set elements in the UI schema
	defaultUISchema.SetP(elements, "elements")

	return defaultUISchema, nil
}
