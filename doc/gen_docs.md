# Generating Templated Docs For Your Own cobra.Command

Generating docs based on a template for a cobra command is incredibly easy. An example is as follows:

```go
package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func main() {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "my test program",
    }
    
    tpl := `
{{header .HeaderScale}} {{.Name}} - {{.Short}}

{{subHeader .HeaderScale}} How to use:
    {{.UseLine}}
`

    ct := template.Must(template.New("tpl").Funcs(template.FuncMap{
		"header": func(scale int) string {
			return strings.Repeat("#", scale + 2)
		},
		"subHeader": func(scale int) string {
			return "#" + strings.Repeat("#", scale + 2)
		},
	}).Parse(tpl))

	linkHandler := func(s string) string { return s }
	out := new(bytes.Buffer)

	err := GenDocsCustomTemplate(echoCmd, out, linkHandler, ct)
	if err != nil {
		log.Fatal(err)
	}
}
```

This will produce the output:

---
### test - my test program

#### How to use:
    test
---

The available fields for use in your template are:
```go
Name          string   // full path to the command
Short         string   // short description of the command
Long          string   // long description of the command
UseLine       string   // full usage for a given command (including parents)
Example       string   // examples of how to use the command
Flags         string   // default values of all non-inherited flags as a string
FlagSlice     []string // Flags represented as a slice
ParentFlags   string   // default values of all inherited flags as a string
ParentLink    string   // rendered internal link to the parent command
ChildrenLinks []string // rendered internal links to the child commands as a slice
RelatedLinks  []string // rendered internal links to the related commands as a slice
CommandLink   string   // rendered internal link to the command
HeaderScale   int      // integer scale indicating depth of the current command
AutoGenTag    string   // automatically generated tag by Cobra
```

The `linkHandler` can be used to customize the rendered internal links to the commands, given a filename:

```go
linkHandler := func(name string) string {
	base := strings.TrimSuffix(name, path.Ext(name))
	return "/commands/" + strings.ToLower(base) + "/"
}
```
