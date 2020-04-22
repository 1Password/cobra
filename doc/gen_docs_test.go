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

{{if ne .Example ""}}{{subHeader .HeaderScale}} Examples

{{.Example}}{{end}}

{{if gt (len .FlagSlice) 1}}{{subHeader .HeaderScale}} Flags

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

	expectedOutput := `
# root echo - Echo anything to the screen

## How to Use

	root echo [string to echo] [flags]

## Synopsis
an utterly useless command for testing

## Examples

Just run cobra-test echo

## Flags

  -b, --boolone          help message for flag boolone (default true)
  -h, --help             help for echo
  -i, --intone int       help message for flag intone (default 123)
  -p, --persistentbool   help message for flag persistentbool
  -s, --strone string    help message for flag strone (default "one")

## Subcommands

* [root echo echosub](#root-echo-echosub)	 - second sub command for echo
* [root echo times](#root-echo-times)	 - Echo anything to the screen more times

## Related commands

* [root print](#root-print)	 - Print anything to the screen

`

	if !strings.EqualFold(out.String(), expectedOutput) {
		t.Error("Generated output did not match expected output")
	}
}
