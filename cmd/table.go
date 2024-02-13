//nolint:wrapcheck
package cmd

import (
	"bytes"
	"errors"
	"io"
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
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/rules"
)

type tableCommandParams struct {
	compareToReadme bool
	writeToReadme   bool
}

func init() {
	params := tableCommandParams{}
	tableCommand := &cobra.Command{
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
			buf := must(createTable(args))

			var err error
			switch {
			case params.writeToReadme:
				err = writeToREADME(buf)
			case params.compareToReadme:
				err = compareToREADME(buf)
			default:
				_, err = os.Stdout.Write(must(io.ReadAll(buf)))
			}

			if err != nil {
				log.Fatal(err)
			}
		},
	}
	tableCommand.Flags().BoolVar(&params.compareToReadme, "compare-to-readme", false,
		"Compare generated table to that in README.md, and exit with non-zero status if they differ")
	tableCommand.Flags().BoolVar(&params.writeToReadme, "write-to-readme", false,
		"Write table to README.md")
	RootCommand.AddCommand(tableCommand)
}

func unquotedPath(path ast.Ref) []string {
	ret := make([]string, 0, len(path)-1)
	for _, ref := range path[1:] {
		ret = append(ret, strings.Trim(ref.String(), `"`))
	}

	return ret
}

func createTable(args []string) (io.Reader, error) {
	result, err := loader.NewFileLoader().Filtered(args, func(_abspath string, _ fs.FileInfo, _depth int) bool {
		return false
	})
	if err != nil {
		return nil, err
	}

	modules := map[string]*ast.Module{}

	for path, file := range result.Modules {
		modules[path] = file.Parsed
	}

	compiler := compile.NewCompilerWithRegalBuiltins()
	compiler.Compile(modules)

	if compiler.Failed() {
		return nil, compiler.Errors
	}

	flattened := compiler.GetAnnotationSet().Flatten()

	tableMap := map[string][][]string{}

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

		tableMap[category] = append(tableMap[category], []string{
			category,
			"[" + title + "](" + docs.CreateDocsURL(category, title) + ")",
			annotations.Description,
		})
	}

	for _, rule := range rules.AllGoRules(config.Config{}) {
		tableMap[rule.Category()] = append(tableMap[rule.Category()], []string{
			rule.Category(),
			"[" + rule.Name() + "](" + rule.Documentation() + ")",
			rule.Description(),
		})
	}

	// Sort the list of rules in each category by name
	for category := range tableMap {
		sort.Slice(tableMap[category], func(i, j int) bool {
			return tableMap[category][i][1] < tableMap[category][j][1]
		})
	}

	// And sort the categories themselves
	categories := util.Keys(tableMap)

	sort.Strings(categories)

	tableData := make([][]string, 0, len(flattened))

	for _, category := range categories {
		tableData = append(tableData, tableMap[category]...)
	}

	return renderTable(tableData), nil
}

func renderTable(tableData [][]string) io.Reader {
	var buf bytes.Buffer

	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Category", "Title", "Description"})
	table.SetAutoFormatHeaders(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetAutoWrapText(false)
	table.AppendBulk(tableData)
	table.Render()

	return &buf
}

func renderREADME(r io.Reader) string {
	file := string(must(os.ReadFile("README.md")))

	first := strings.Split(file, "<!-- RULES_TABLE_START -->")[0]
	last := strings.Split(file, "<!-- RULES_TABLE_END -->")[1]

	table := must(io.ReadAll(r))

	newReadme := first + "<!-- RULES_TABLE_START -->\n\n" + string(table) + "\n<!-- RULES_TABLE_END -->" + last

	return newReadme
}

func writeToREADME(r io.Reader) error {
	newReadme := renderREADME(r)

	return os.WriteFile("README.md", []byte(newReadme), 0o600)
}

func compareToREADME(r io.Reader) error {
	oldReadme := must(os.ReadFile("README.md"))
	newReadme := renderREADME(r)

	if string(oldReadme) != newReadme {
		return errors.New(
			"table in README.md is out of date. Run `go run main.go table --write-to-readme` to have it updated, " +
				"then include the change in your commit",
		)
	}

	return nil
}

func must[V any](value V, err error) V {
	if err != nil {
		log.Fatal(err)
	}

	return value
}
