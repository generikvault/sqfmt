package main

import (
	"strings"
	"text/template"
)

// Data provides the data for the sql table markdown templates.
type Data struct {
	Tables []Table
}

// Table provides the data of a table for the sql table markdown templates.
type Table struct {
	Name string
	Cols []Col

	Group     string
	Reference string
}

// Col provides the data of a table coloumn for the sql table markdown templates.
type Col struct {
	Name            string
	Type            string
	Nullable        bool
	Default         interface{}
	PrimaryKeyIndex int

	ReferenceTableLink string

	ReferenceTable string
	ReferenceGroup string
}

var markdown = template.Must(template.New("md").Parse(`{{range $table := .Tables -}}
- [{{$table.Name}}](#{{$table.Reference}})
{{end -}}

{{range $table := .Tables}}
## {{$table.Name}}

{{range $col := $table.Cols -}}
- **{{$col.Name}}** 
	{{- if $col.Type}} *{{$col.Type}}
		{{- if $col.HasDefault}}({{$col.Default}}){{end}}*
	{{- end}}
	{{- if $col.ReferenceTable}} -> [{{$col.ReferenceTable}}]({{$col.ReferenceTableLink}}){{end}}
{{end}}
{{- end -}}`))

func (t *Table) deriveLinks() {
	for i := range t.Cols {
		c := &t.Cols[i]
		var file string
		if c.ReferenceGroup != t.Group {
			file = cfg.LinkPrefix + "Tabellen" + c.ReferenceGroup
			if !cfg.SkipLinkMDExtension {
				file += ".md"
			}
		}
		c.ReferenceTableLink = file + "#" + strings.ToLower(c.ReferenceTable) // TODO remove fixed dir reference
	}
}

// HasDefault returns true if the column has a default value.
func (c Col) HasDefault() bool {
	return c.Default != nil
}

func (d Data) tableByName(name string) (Table, bool) {
	for _, table := range d.Tables {
		if name == table.Name {
			return table, true
		}
	}
	return Table{}, false
}
