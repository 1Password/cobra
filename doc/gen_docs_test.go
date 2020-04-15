package doc

import (
	"bytes"
	"fmt"
	"log"
	"path"
	"strings"
	"testing"
	"text/template"
)

func TestGenDocsCustomTemplate(t *testing.T) {

	cmdTpl := `
{{header .HeaderScale}} {{.Name}} - {{.Short}}

{{subHeader .HeaderScale}} How to Use

	{{.UseLine}}

{{subHeader .HeaderScale}} Synopsis
{{.Long}}

{{if ne .Example ""}}
{{subHeader .HeaderScale}} Examples

{{.Example}}{{end}}

{{if gt (len .FlagSlice) 1}}{{subHeader .HeaderScale}} Options

{{.Flags}}{{end}}

{{subHeader .HeaderScale}} Subcommands

{{range $childLink := .ChildrenLinks}}{{$childLink}}{{end}}
{{if gt (len .RelatedLinks) 0}}{{subHeader .HeaderScale}} Related commands

{{range $relatedLink := .RelatedLinks}}{{$relatedLink}}{{end}}{{end}}
`

	ct := template.Must(template.New("cmdTpl").Funcs(template.FuncMap{
		"header": func(scale int) string {
			return strings.Repeat("#", scale)
		},
		"subHeader": func(scale int) string {
			return "#" + strings.Repeat("#", scale)
		},
	}).Parse(cmdTpl))

	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, path.Ext(name))
		base = strings.Replace(base, "_", "-", -1)
		return fmt.Sprintf("#%s", base)
	}

	out := new(bytes.Buffer)

	err := GenDocsCustomTemplate(echoCmd, out, linkHandler, ct)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(out.String())
}
