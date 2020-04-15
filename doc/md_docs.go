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
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func printOptions(buf *bytes.Buffer, cmdStruct *CmdOutline) error {
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

	cmdStruct := generateCmdOutline(cmd, linkHandler)

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
		buf.WriteString("######" + cmdStruct.AutoGenTag)
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
