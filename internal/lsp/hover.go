package lsp

import (
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/types"

	"github.com/styrainc/regal/internal/lsp/cache"
	types2 "github.com/styrainc/regal/internal/lsp/types"
)

var builtins = builtinMap() //nolint:gochecknoglobals

var builtinHoverCache = make(map[*ast.Builtin]string) //nolint:gochecknoglobals

func builtinMap() map[string]*ast.Builtin {
	m := make(map[string]*ast.Builtin)
	for _, b := range ast.CapabilitiesForThisVersion().Builtins {
		m[b.Name] = b
	}

	return m
}

func builtinCategory(builtin *ast.Builtin) (category string) {
	if len(builtin.Categories) == 0 {
		if s := strings.Split(builtin.Name, "."); len(s) > 1 {
			category = s[0]
		} else {
			category = builtin.Name
		}
	} else {
		category = builtin.Categories[0]
	}

	return category
}

func writeFunctionSnippet(sb *strings.Builder, builtin *ast.Builtin) {
	sb.WriteString("```rego\n")

	resType := builtin.Decl.NamedResult()
	if n, ok := resType.(*types.NamedType); ok {
		sb.WriteString(n.Name)
	} else {
		sb.WriteString("output")
	}

	sb.WriteString(" := ")

	sb.WriteString(builtin.Name)
	sb.WriteString("(")

	for i, arg := range builtin.Decl.NamedFuncArgs().Args {
		if i > 0 {
			sb.WriteString(", ")
		}

		if n, ok := arg.(*types.NamedType); ok {
			sb.WriteString(n.Name)
		} else {
			sb.WriteString(arg.String())
		}
	}

	sb.WriteString(")\n```")
}

func createHoverContent(builtin *ast.Builtin) string {
	if content, ok := builtinHoverCache[builtin]; ok {
		return content
	}

	title := fmt.Sprintf(
		"[%s](https://www.openpolicyagent.org/docs/latest/policy-reference/#builtin-%s-%s)",
		builtin.Name,
		builtinCategory(builtin),
		strings.ReplaceAll(builtin.Name, ".", ""),
	)

	sb := &strings.Builder{}

	sb.WriteString("### ")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	writeFunctionSnippet(sb, builtin)

	sb.WriteString("\n\n")
	sb.WriteString(builtin.Description)
	sb.WriteString("\n")

	if len(builtin.Decl.FuncArgs().Args) == 0 {
		return sb.String()
	}

	sb.WriteString("\n\n#### Arguments\n\n")

	table := tablewriter.NewWriter(sb)

	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"Name", "Type", "Description"})
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetCenterSeparator("|") // Add Bulk Data

	argsData := make([][]string, 0)

	for _, arg := range builtin.Decl.NamedFuncArgs().Args {
		if n, ok := arg.(*types.NamedType); ok {
			argsData = append(argsData, []string{"`" + n.Name + "`", n.Type.String(), n.Descr})
		} else {
			argsData = append(argsData, []string{"`" + arg.String() + "`", "", ""})
		}
	}

	table.AppendBulk(argsData)
	table.Render()

	table.ClearRows()

	sb.WriteString("\n\nReturns ")

	ret := builtin.Decl.NamedResult()
	if n, ok := ret.(*types.NamedType); ok {
		sb.WriteString("`")
		sb.WriteString(n.Name)
		sb.WriteString("` of type `")
		sb.WriteString(n.Type.String())
		sb.WriteString("`: ")
		sb.WriteString(n.Descr)
	} else if ret != nil {
		sb.WriteString(ret.String())
	}

	sb.WriteString("\n")

	result := sb.String()

	builtinHoverCache[builtin] = result

	return result
}

func updateBuiltinPositions(cache *cache.Cache, uri string) error {
	module, ok := cache.GetModule(uri)
	if !ok {
		return fmt.Errorf("failed to update builtin positions: no parsed module for uri %q", uri)
	}

	builtinsOnLine := map[uint][]types2.BuiltinPosition{}

	for _, call := range AllBuiltinCalls(module) {
		line := uint(call.Location.Row)

		builtinsOnLine[line] = append(builtinsOnLine[line], types2.BuiltinPosition{
			Builtin: call.Builtin,
			Line:    line,
			Start:   uint(call.Location.Col),
			End:     uint(call.Location.Col + len(call.Builtin.Name)),
		})
	}

	cache.SetBuiltinPositions(uri, builtinsOnLine)

	return nil
}
