package lsp

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/symbols"
)

func documentSymbols(
	contents string,
	module *ast.Module,
) []types.DocumentSymbol {
	// Only pkgSymbols would likely suffice, but we're keeping docSymbols around in case
	// we ever want to add more top-level symbols than the package.
	docSymbols := make([]types.DocumentSymbol, 0)
	pkgSymbols := make([]types.DocumentSymbol, 0)

	lines := strings.Split(contents, "\n")

	pkgRange := types.Range{
		Start: types.Position{Line: 0, Character: 0},
		End:   types.Position{Line: uint(len(lines) - 1), Character: uint(len(lines[len(lines)-1]))},
	}

	pkg := types.DocumentSymbol{
		Name:           module.Package.Path.String(),
		Kind:           symbols.Package,
		Range:          pkgRange,
		SelectionRange: pkgRange,
	}

	// Create groups of rules and functions sharing the same name
	ruleGroups := make(map[string][]*ast.Rule, len(module.Rules))

	for _, rule := range module.Rules {
		name := refToString(rule.Head.Ref())
		ruleGroups[name] = append(ruleGroups[name], rule)
	}

	for _, rules := range ruleGroups {
		if len(rules) == 1 {
			rule := rules[0]

			kind := symbols.Variable
			if isConstant(rule) {
				kind = symbols.Constant
			} else if rule.Head.Args != nil {
				kind = symbols.Function
			}

			ruleRange := locationToRange(rule.Location)
			ruleSymbol := types.DocumentSymbol{
				Name:           refToString(rule.Head.Ref()),
				Kind:           kind,
				Range:          ruleRange,
				SelectionRange: ruleRange,
			}

			if detail := getRuleDetail(rule); detail != "" {
				ruleSymbol.Detail = &detail
			}

			pkgSymbols = append(pkgSymbols, ruleSymbol)
		} else {
			groupFirstRange := locationToRange(rules[0].Location)
			groupLastRange := locationToRange(rules[len(rules)-1].Location)

			groupRange := types.Range{
				Start: groupFirstRange.Start,
				End:   groupLastRange.End,
			}

			kind := symbols.Variable
			if rules[0].Head.Args != nil {
				kind = symbols.Function
			}

			groupSymbol := types.DocumentSymbol{
				Name:           refToString(rules[0].Head.Ref()),
				Kind:           kind,
				Range:          groupRange,
				SelectionRange: groupRange,
			}

			detail := getRuleDetail(rules[0])
			if detail != "" {
				groupSymbol.Detail = &detail
			}

			children := make([]types.DocumentSymbol, 0, len(rules))

			for i, rule := range rules {
				childRange := locationToRange(rule.Location)
				childRule := types.DocumentSymbol{
					Name:           fmt.Sprintf("#%d", i+1),
					Kind:           kind,
					Range:          childRange,
					SelectionRange: childRange,
				}

				childDetail := getRuleDetail(rule)
				if childDetail != "" {
					childRule.Detail = &childDetail
				}

				children = append(children, childRule)
			}

			groupSymbol.Children = &children

			pkgSymbols = append(pkgSymbols, groupSymbol)
		}
	}

	if len(pkgSymbols) > 0 {
		pkg.Children = &pkgSymbols
	}

	docSymbols = append(docSymbols, pkg)

	return docSymbols
}

func locationToRange(location *ast.Location) types.Range {
	lines := bytes.Split(location.Text, []byte("\n"))

	var endLine uint
	if len(lines) == 1 {
		endLine = uint(location.Row - 1)
	} else {
		endLine = uint(location.Row-1) + uint(len(lines)-1)
	}

	return types.Range{
		Start: types.Position{
			Line:      uint(location.Row - 1),
			Character: uint(location.Col - 1),
		},
		End: types.Position{
			Line:      endLine,
			Character: uint(len(lines[len(lines)-1])),
		},
	}
}

func refToString(ref ast.Ref) string {
	sb := strings.Builder{}

	for i, part := range ref {
		if part.IsGround() {
			if i > 0 {
				sb.WriteString(".")
			}

			sb.WriteString(strings.Trim(part.Value.String(), `"`))
		} else {
			if i == 0 {
				sb.WriteString(strings.Trim(part.Value.String(), `"`))
			} else {
				sb.WriteString("[")
				sb.WriteString(strings.Trim(part.Value.String(), `"`))
				sb.WriteString("]")
			}
		}
	}

	return sb.String()
}

func getRuleDetail(rule *ast.Rule) string {
	if rule.Head.Args != nil {
		return "function" + rule.Head.Args.String()
	}

	if rule.Head.Key != nil && rule.Head.Value == nil {
		return "multi-value rule"
	}

	if rule.Head.Value == nil {
		return ""
	}

	detail := "single-value "

	if rule.Head.Key != nil {
		detail += "map "
	}

	detail += "rule"

	switch v := rule.Head.Value.Value.(type) {
	case ast.Boolean:
		if strings.HasPrefix(rule.Head.Ref()[0].String(), "test_") {
			detail += " (test)"
		} else {
			detail += " (boolean)"
		}
	case ast.Number:
		detail += " (number)"
	case ast.String:
		detail += " (string)"
	case *ast.Array, *ast.ArrayComprehension:
		detail += " (array)"
	case ast.Object, *ast.ObjectComprehension:
		detail += " (object)"
	case ast.Set, *ast.SetComprehension:
		detail += " (set)"
	case ast.Call:
		name := v[0].String()

		if builtin, ok := builtins[name]; ok {
			retType := builtin.Decl.NamedResult().String()

			detail += fmt.Sprintf(" (%s)", simplifyType(retType))
		}
	}

	return detail
}

// simplifyType removes anything but the base type from the type name.
func simplifyType(name string) string {
	result := name

	if strings.Contains(result, ":") {
		result = result[strings.Index(result, ":")+1:]
	}

	// silence gocritic linter here as strings.Index can in
	// fact *not* return -1 in these cases
	if strings.Contains(result, "[") {
		result = result[:strings.Index(result, "[")] //nolint:gocritic
	}

	if strings.Contains(result, "<") {
		result = result[:strings.Index(result, "<")] //nolint:gocritic
	}

	return strings.TrimSpace(result)
}

// isConstant returns true if the rule is a "constant" rule, i.e.
// one without conditions and scalar value in the head.
func isConstant(rule *ast.Rule) bool {
	isScalar := false

	if rule.Head.Value == nil {
		return false
	}

	switch rule.Head.Value.Value.(type) {
	case ast.Boolean, ast.Number, ast.String, ast.Null:
		isScalar = true
	}

	return isScalar &&
		rule.Head.Args == nil &&
		rule.Body.Equal(ast.NewBody(ast.NewExpr(ast.BooleanTerm(true)))) &&
		rule.Else == nil
}

func toWorkspaceSymbol(docSym types.DocumentSymbol, docURL string) types.WorkspaceSymbol {
	return types.WorkspaceSymbol{
		Name: docSym.Name,
		Kind: docSym.Kind,
		Location: types.Location{
			URI:   docURL,
			Range: docSym.Range,
		},
	}
}

func toWorkspaceSymbols(docSym []types.DocumentSymbol, docURL string, symbols *[]types.WorkspaceSymbol) {
	for _, sym := range docSym {
		// Only include the "main" symbol for incremental rules and functions
		// as numeric items isn't very useful in the workspace symbol list.
		if !strings.HasPrefix(sym.Name, "#") {
			*symbols = append(*symbols, toWorkspaceSymbol(sym, docURL))

			if sym.Children != nil {
				toWorkspaceSymbols(*sym.Children, docURL, symbols)
			}
		}
	}
}
