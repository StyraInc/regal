package hover

import (
	"fmt"
	"strings"
	"sync"

	"github.com/olekukonko/tablewriter"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/types"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/rego"
	types2 "github.com/styrainc/regal/internal/lsp/types"
)

var builtinCache = make(map[*ast.Builtin]string) //nolint:gochecknoglobals

var builtinCacheLock = &sync.Mutex{} //nolint:gochecknoglobals

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

func CreateHoverContent(builtin *ast.Builtin) string {
	builtinCacheLock.Lock()
	if content, ok := builtinCache[builtin]; ok {
		builtinCacheLock.Unlock()

		return content
	}
	builtinCacheLock.Unlock()

	title := fmt.Sprintf(
		"[%s](https://www.openpolicyagent.org/docs/latest/policy-reference/#builtin-%s-%s)",
		builtin.Name,
		rego.BuiltinCategory(builtin),
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

	builtinCacheLock.Lock()
	builtinCache[builtin] = result
	builtinCacheLock.Unlock()

	return result
}

func UpdateBuiltinPositions(cache *cache.Cache, uri string) error {
	module, ok := cache.GetModule(uri)
	if !ok {
		return fmt.Errorf("failed to update builtin positions: no parsed module for uri %q", uri)
	}

	builtinsOnLine := map[uint][]types2.BuiltinPosition{}

	for _, call := range rego.AllBuiltinCalls(module) {
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
