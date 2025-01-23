package lsp

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"

	rast "github.com/styrainc/regal/internal/ast"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/symbols"
)

func documentSymbols(
	contents string,
	module *ast.Module,
	builtins map[string]*ast.Builtin,
) []types.DocumentSymbol {
	// Only pkgSymbols would likely suffice, but we're keeping docSymbols around in case
	// we ever want to add more top-level symbols than the package.
	docSymbols := make([]types.DocumentSymbol, 0)
	pkgSymbols := make([]types.DocumentSymbol, 0)

	lines := strings.Split(contents, "\n")

	//nolint:gosec
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
		name := rule.Head.Ref().String()
		ruleGroups[name] = append(ruleGroups[name], rule)
	}

	for _, rules := range ruleGroups {
		if len(rules) == 1 {
			rule := rules[0]

			kind := symbols.Variable
			if rast.IsConstant(rule) {
				kind = symbols.Constant
			} else if rule.Head.Args != nil {
				kind = symbols.Function
			}

			ruleRange := locationToRange(rule.Location)
			ruleSymbol := types.DocumentSymbol{
				Name:           rule.Head.Ref().String(),
				Kind:           kind,
				Range:          ruleRange,
				SelectionRange: ruleRange,
			}

			if detail := rast.GetRuleDetail(rule, builtins); detail != "" {
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
				Name:           rules[0].Head.Ref().String(),
				Kind:           kind,
				Range:          groupRange,
				SelectionRange: groupRange,
			}

			detail := rast.GetRuleDetail(rules[0], builtins)
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

				childDetail := rast.GetRuleDetail(rule, builtins)
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

//nolint:gosec
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
