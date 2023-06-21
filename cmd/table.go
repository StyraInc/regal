//nolint:wrapcheck
package cmd

import (
	"bytes"
	"errors"
	"io/fs"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"

	"github.com/styrainc/regal/internal/compile"
	"github.com/styrainc/regal/internal/docs"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/rules"
)

type tableCommandParams struct {
	writeToReadme bool
}

func init() {
	params := tableCommandParams{}
	parseCommand := &cobra.Command{
		Hidden: true,
		Use:    "table <path> [path [...]]",
		Long:   "Create a markdown table from rule annotations found in provided modules",

		PreRunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no files to parse for annotations provided")
			}

			return nil
		},

		Run: func(_ *cobra.Command, args []string) {
			if err := createTable(args, params); err != nil {
				log.SetOutput(os.Stderr)
				log.Println(err)
				os.Exit(1)
			}
		},
	}
	parseCommand.Flags().BoolVar(&params.writeToReadme, "write-to-readme", false, "Write table to README.md")
	RootCommand.AddCommand(parseCommand)
}

func createDocsURL(category, title string) string {
	return docs.DocsBaseURL + "/" + category + "/" + title + ".md"
}

func unquotedPath(path ast.Ref) []string {
	ret := make([]string, 0, len(path)-1)
	for _, ref := range path[1:] {
		ret = append(ret, strings.Trim(ref.String(), `"`))
	}

	return ret
}

func createTable(args []string, params tableCommandParams) error {
	result, err := loader.NewFileLoader().Filtered(args, func(abspath string, info fs.FileInfo, depth int) bool {
		return false
	})
	if err != nil {
		return err
	}

	modules := map[string]*ast.Module{}

	for path, file := range result.Modules {
		modules[path] = file.Parsed
	}

	compiler := compile.NewCompilerWithRegalBuiltins()
	compiler.Compile(modules)

	if compiler.Failed() {
		return compiler.Errors
	}

	flattened := compiler.GetAnnotationSet().Flatten()
	tableData := make([][]string, 0, len(flattened))

	traversedTitles := map[string]struct{}{}

	for _, entry := range flattened {
		annotations := entry.Annotations

		path := unquotedPath(entry.Path)

		if len(path) != 4 {
			continue
		}

		if path[0] != "regal" {
			continue
		}

		if path[1] != "rules" {
			continue
		}

		category := path[2]
		title := path[3]

		if _, ok := traversedTitles[title]; ok {
			continue
		}

		traversedTitles[title] = struct{}{}

		tableData = append(tableData, []string{
			category,
			"[" + title + "](" + createDocsURL(category, title) + ")",
			annotations.Description,
		})
	}

	// We currently don't include the severity level sourced from the provided config in the
	// table, as all rules default to error at this point. We might want to change that later.
	for _, rule := range rules.AllGoRules(config.Config{}) {
		tableData = append(tableData, []string{
			rule.Category(),
			"[" + rule.Name() + "](" + rule.Documentation() + ")",
			rule.Description(),
		})
	}

	sort.Slice(tableData, func(i, j int) bool {
		return tableData[i][0] < tableData[j][0]
	})

	return writeTable(tableData, params)
}

func writeTable(tableData [][]string, params tableCommandParams) error {
	var buf bytes.Buffer

	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Category", "Title", "Description"})
	table.SetAutoFormatHeaders(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetAutoWrapText(false)
	table.AppendBulk(tableData)
	table.Render()

	if !params.writeToReadme {
		_, err := os.Stdout.Write(buf.Bytes())

		return err
	}

	file, err := os.ReadFile("README.md")
	if err != nil {
		return err
	}

	first := strings.Split(string(file), "<!-- RULES_TABLE_START -->")[0]
	last := strings.Split(string(file), "<!-- RULES_TABLE_END -->")[1]

	newReadme := first + "<!-- RULES_TABLE_START -->\n\n" + buf.String() + "\n<!-- RULES_TABLE_END -->" + last

	return os.WriteFile("README.md", []byte(newReadme), 0o600)
}
