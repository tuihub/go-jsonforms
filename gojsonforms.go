package gojsonforms

import (
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
	uiSchema reader
	schema   reader
	data     reader
	menu     []models.MenuItem
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

	// uischema is necessary
	uiSchema, err := b.uiSchema.Read()
	if err != nil {
		return html, err
	}

	f, err := form.NewForm(schema, uiSchema)
	if err != nil {
		return html, err
	}

	if data, err := b.data.Read(); err != nil {
		return html, err
	} else if data != nil {
		f.BindData(data)
	}

	f.SetMenu(b.menu)

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
