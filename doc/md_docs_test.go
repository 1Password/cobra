package doc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/spf13/cobra"
)

func TestGenMdDoc(t *testing.T) {
	// We generate on subcommand so we have both subcommands and parents.
	buf := new(bytes.Buffer)
	if err := GenMarkdown(echoCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringContains(t, output, echoCmd.Long)
	checkStringContains(t, output, echoCmd.Example)
	checkStringContains(t, output, "boolone")
	checkStringContains(t, output, "rootflag")
	checkStringContains(t, output, rootCmd.Short)
	checkStringContains(t, output, echoSubCmd.Short)
	checkStringOmits(t, output, deprecatedCmd.Short)
	checkStringContains(t, output, "Options inherited from parent commands")
}

func TestGenMdNoHiddenParents(t *testing.T) {
	// We generate on subcommand so we have both subcommands and parents.
	for _, name := range []string{"rootflag", "strtwo"} {
		f := rootCmd.PersistentFlags().Lookup(name)
		f.Hidden = true
		defer func() { f.Hidden = false }()
	}
	buf := new(bytes.Buffer)
	if err := GenMarkdown(echoCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringContains(t, output, echoCmd.Long)
	checkStringContains(t, output, echoCmd.Example)
	checkStringContains(t, output, "boolone")
	checkStringOmits(t, output, "rootflag")
	checkStringContains(t, output, rootCmd.Short)
	checkStringContains(t, output, echoSubCmd.Short)
	checkStringOmits(t, output, deprecatedCmd.Short)
	checkStringOmits(t, output, "Options inherited from parent commands")
}

func TestGenMdNoTag(t *testing.T) {
	rootCmd.DisableAutoGenTag = true
	defer func() { rootCmd.DisableAutoGenTag = false }()

	buf := new(bytes.Buffer)
	if err := GenMarkdown(rootCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringOmits(t, output, "Auto generated")
}

func TestGenMdTree(t *testing.T) {
	c := &cobra.Command{Use: "do [OPTIONS] arg1 arg2"}
	tmpdir, err := ioutil.TempDir("", "test-gen-md-tree")
	if err != nil {
		t.Fatalf("Failed to create tmpdir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	if err := GenMarkdownTree(c, tmpdir); err != nil {
		t.Fatalf("GenMarkdownTree failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpdir, "do.md")); err != nil {
		t.Fatalf("Expected file 'do.md' to exist")
	}
}

func TestGenMdTemplate(t *testing.T) {
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

	err := GenMarkdownCustomTemplate(echoCmd, out, linkHandler, ct)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(out.String())
}

func BenchmarkGenMarkdownToFile(b *testing.B) {
	file, err := ioutil.TempFile("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := GenMarkdown(rootCmd, file); err != nil {
			b.Fatal(err)
		}
	}
}
