//Copyright 2015 Red Hat Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package doc

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

type CmdTemplate struct {
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
}

func generateCmdTemplate(cmd *cobra.Command, linkHandler func(string) string) *CmdTemplate {
	name := cmd.CommandPath()
	short := cmd.Short
	long := cmd.Long
	if len(long) == 0 {
		long = short
	}

	var useLine string
	useLine = cmd.UseLine()

	var example string
	if len(cmd.Example) > 0 {
		example = cmd.Example
	}

	buf := new(bytes.Buffer)

	var flagString string
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		flags.PrintDefaults()
		flagString = buf.String()
		buf.Reset()
	}

	var flagSlice []string
	flagSplitter := regexp.MustCompile("\n *-") // clips first '-' in flags starting the second flag
	numberOfFlags := strings.Count(flagString, "--")
	flagSlice = flagSplitter.Split(flagString, numberOfFlags)
	if flagSlice[len(flagSlice)-1] == "" {
		flagSlice = flagSlice[:len(flagSlice)-1]
	}
	for i, flag := range flagSlice {
		if i > 0 {
			flagSlice[i] = "-" + flag // add clipped '-' back to flags
		}
	}

	var parentFlagString string
	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		parentFlags.PrintDefaults()
		parentFlagString = buf.String()
		buf.Reset()
	}

	headerScale := 0
	var parentLink string
	if cmd.HasParent() {
		parent := cmd.Parent()
		pname := parent.CommandPath()
		link := pname + ".md"
		link = strings.Replace(link, " ", "_", -1)
		buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n", pname, linkHandler(link), parent.Short))
		parentLink = buf.String()
		buf.Reset()

		headerScale = 1
		for parent.HasParent() {
			headerScale += 1
			parent = parent.Parent()
		}
	}

	var childrenLinks []string
	children := cmd.Commands()
	sort.Sort(byName(children))

	for _, child := range children {
		var childLink string
		if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
			continue
		}
		cname := name + " " + child.Name()
		link := cname + ".md"
		link = strings.Replace(link, " ", "_", -1)
		buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n", cname, linkHandler(link), child.Short))

		childLink = buf.String()
		childrenLinks = append(childrenLinks, childLink)
		buf.Reset()
	}

	var relatedLinks []string
	relatedCmds := cmd.RelatedCommands()

	for _, relCmd := range relatedCmds {
		var relatedLink string
		if !relCmd.IsAvailableCommand() || relCmd.IsAdditionalHelpTopicCommand() {
			continue
		}
		rname := relCmd.CommandPath()
		link := rname + ".md"
		link = strings.Replace(link, " ", "_", -1)
		buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n", rname, linkHandler(link), relCmd.Short))

		relatedLink = buf.String()
		relatedLinks = append(relatedLinks, relatedLink)
		buf.Reset()
	}

	var commandLink string
	link := name + ".md"
	link = strings.Replace(link, " ", "_", -1)
	buf.WriteString(fmt.Sprintf("%s", linkHandler(link)))
	commandLink = buf.String()
	buf.Reset()

	return &CmdTemplate{
		Name:          name,
		Short:         short,
		Long:          long,
		UseLine:       useLine,
		Example:       example,
		Flags:         flagString,
		FlagSlice:     flagSlice,
		ParentFlags:   parentFlagString,
		ParentLink:    parentLink,
		ChildrenLinks: childrenLinks,
		RelatedLinks:  relatedLinks,
		CommandLink:   commandLink,
		HeaderScale:   headerScale,
	}
}

func printOptions(buf *bytes.Buffer, cmdStruct *CmdTemplate) error {
	if len(cmdStruct.Flags) > 0 {
		buf.WriteString(fmt.Sprintf("### Options\n\n```\n%s```\n\n", cmdStruct.Flags))
	}

	if len(cmdStruct.ParentFlags) > 0 {
		buf.WriteString(fmt.Sprintf("### Options inherited from parent commands\n\n```\n%s```\n\n", cmdStruct.ParentFlags))
	}
	return nil
}

// GenMarkdown creates markdown output.
func GenMarkdown(cmd *cobra.Command, w io.Writer) error {
	return GenMarkdownCustom(cmd, w, func(s string) string { return s })
}

// GenMarkdownCustom creates custom markdown output.
func GenMarkdownCustom(cmd *cobra.Command, w io.Writer, linkHandler func(string) string) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)

	cmdStruct := generateCmdTemplate(cmd, linkHandler)

	buf.WriteString("## " + cmdStruct.Name + "\n\n")
	buf.WriteString(cmdStruct.Short + "\n\n")
	buf.WriteString("### Synopsis\n\n")
	buf.WriteString(cmdStruct.Long + "\n\n")

	if cmd.Runnable() && len(cmdStruct.UseLine) > 0 {
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmdStruct.UseLine))
	}

	if len(cmdStruct.Example) > 0 {
		buf.WriteString("### Examples\n\n")
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmdStruct.Example))
	}

	if err := printOptions(buf, cmdStruct); err != nil {
		return err
	}
	if len(cmdStruct.ParentLink) > 0 || len(cmdStruct.ChildrenLinks) > 0 {
		buf.WriteString("### SEE ALSO\n\n")
		if len(cmdStruct.ParentLink) > 0 {
			buf.WriteString(cmdStruct.ParentLink)
			cmd.VisitParents(func(c *cobra.Command) {
				if c.DisableAutoGenTag {
					cmd.DisableAutoGenTag = c.DisableAutoGenTag
				}
			})
		}

		for _, childLink := range cmdStruct.ChildrenLinks {
			buf.WriteString(childLink)
		}
		buf.WriteString("\n")
	}

	if !cmd.DisableAutoGenTag {
		buf.WriteString("###### Auto generated by spf13/cobra on " + time.Now().Format("2-Jan-2006") + "\n")
	}
	_, err := buf.WriteTo(w)
	return err
}

func GenMarkdownCustomTemplate(cmd *cobra.Command, w io.Writer, linkHandler func(string) string, template *template.Template) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)

	cmdStruct := generateCmdTemplate(cmd, linkHandler)

	cmd.DisableAutoGenTag = true
	writeToTemplate(cmdStruct, template, buf)

	if !cmd.DisableAutoGenTag {
		buf.WriteString("###### Auto generated by spf13/cobra on " + time.Now().Format("2-Jan-2006") + "\n")
	}
	_, err := buf.WriteTo(w)
	return err
}

// GenMarkdownTree will generate a markdown page for this command and all
// descendants in the directory given. The header may be nil.
// This function may not work correctly if your command names have `-` in them.
// If you have `cmd` with two subcmds, `sub` and `sub-third`,
// and `sub` has a subcommand called `third`, it is undefined which
// help output will be in the file `cmd-sub-third.1`.
func GenMarkdownTree(cmd *cobra.Command, dir string) error {
	identity := func(s string) string { return s }
	emptyStr := func(s string) string { return "" }
	return GenMarkdownTreeCustom(cmd, dir, emptyStr, identity)
}

// GenMarkdownTreeCustom is the the same as GenMarkdownTree, but
// with custom filePrepender and linkHandler.
func GenMarkdownTreeCustom(cmd *cobra.Command, dir string, filePrepender, linkHandler func(string) string) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := GenMarkdownTreeCustom(c, dir, filePrepender, linkHandler); err != nil {
			return err
		}
	}

	basename := strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".md"
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.WriteString(f, filePrepender(filename)); err != nil {
		return err
	}
	if err := GenMarkdownCustom(cmd, f, linkHandler); err != nil {
		return err
	}
	return nil
}

func writeToTemplate(cmdStruct *CmdTemplate, template *template.Template, buf *bytes.Buffer) error {
	err := template.Execute(buf, cmdStruct)
	if err != nil {
		log.Println("executing template:", err)
	}
	return nil
}
