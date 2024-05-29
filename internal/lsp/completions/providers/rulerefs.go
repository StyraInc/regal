package providers

import (
	"strings"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

// RuleRefs is a completion provider that returns completions for
// rules found in:
// - the current file
// - imported packages
// - any other files in the workspace under data.
type RuleRefs struct{}

func (*RuleRefs) Run(
	c *cache.Cache,
	params types.CompletionParams,
	_ *Options,
) ([]types.CompletionItem, error) {
	fileURI := params.TextDocument.URI

	lines, currentLine := completionLineHelper(c, fileURI, params.Position.Line)
	if len(lines) < 1 || currentLine == "" {
		return []types.CompletionItem{}, nil
	}

	// TODO: Share and improve this logic, currently shared with the builtins provider
	if !strings.Contains(currentLine, " if ") && // if after if keyword
		!strings.Contains(currentLine, " contains ") && // if after contains
		!strings.Contains(currentLine, " else ") && // if after else
		!strings.Contains(currentLine, "= ") && // if after assignment
		!patternRuleBody.MatchString(currentLine) { // if in rule body
		return nil, nil
	}

	words := patternWhiteSpace.Split(currentLine, -1)
	lastWord := words[len(words)-1]

	// some version of a parsed mod is needed here to filter refs to suggest
	// based on import statements
	mod, ok := c.GetModule(fileURI)
	if !ok {
		return nil, nil
	}

	refsForContext := make(map[string]types.Ref)

	for uri, refs := range c.GetAllFileRefs() {
		for _, ref := range refs {
			// we are not interested in packages here, only rules
			if ref.Kind == types.Package {
				continue
			}

			// don't suggest refs of "private" rules, even if this
			// is only just a naming convention
			if strings.Contains(ref.Label, "._") {
				continue
			}

			// for refs from imported packages, we need to strip the start of the
			// package string, e.g. data.foo.bar -> bar
			key := ref.Label
			for _, i := range mod.Imports {
				if k := attemptToStripImportPrefix(key, i.Path.String()); k != "" {
					key = k

					break
				}
			}

			// suggest rules from the current file without any package prefix
			if uri == fileURI {
				parts := strings.Split(key, ".")
				key = parts[len(parts)-1]
			}

			// only suggest refs that match the last word the user has typed.
			if !strings.HasPrefix(key, lastWord) {
				continue
			}

			refsForContext[key] = ref
		}
	}

	// Generate a list of package names from longer rule names.
	// For example, if the rules data.foo.bar and data.foo.baz
	// are defined, an author should still see data.foo suggested
	// as a partial path leading to the rules bar and baz.
	for key := range refsForContext {
		parts := strings.Split(key, ".")
		// starting at 1 to skip 'data'
		for i := 1; i < len(parts)-1; i++ {
			partialKey := strings.Join(parts[:i+1], ".")
			// only insert the new partial key if there is no full package
			// ref that matches it
			if _, ok := refsForContext[partialKey]; !ok {
				refsForContext[partialKey] = types.Ref{
					Label:       partialKey,
					Description: "Partial",
				}
			}
		}
	}

	// refs are grouped by 'depth', where depth is the number of dots in the
	// ref string. This is a simplification, but allows shorter, higher level
	// refs to be suggested first.
	depths, byDepth := groupKeyedRefsByDepth(refsForContext)

	items := make([]types.CompletionItem, 0)
	for _, depth := range depths {
		// items are added in groups of depth until there more then 10 items.
		if len(items) > 10 {
			continue
		}

		itemsForDepth, ok := byDepth[depth]
		if !ok {
			continue
		}

		for key, ref := range itemsForDepth {
			symbol := completion.Variable
			detail := "Rule"

			switch {
			case ref.Kind == types.ConstantRule:
				symbol = completion.Constant
				detail = "Constant Rule"
			case ref.Kind == types.Function:
				symbol = completion.Function
				detail = "Function"
			case ref.Description == "Partial":
				detail = "Partial path suggestion, continue typing for more suggestions."
				symbol = completion.Module
				ref.Description = ""
			}

			items = append(items, types.CompletionItem{
				Label:  key,
				Kind:   symbol,
				Detail: detail,
				Documentation: &types.MarkupContent{
					Kind:  "markdown",
					Value: ref.Description,
				},
				TextEdit: &types.TextEdit{
					Range: types.Range{
						Start: types.Position{
							Line:      params.Position.Line,
							Character: params.Position.Character - uint(len(lastWord)),
						},
						End: types.Position{
							Line:      params.Position.Line,
							Character: uint(len(currentLine)),
						},
					},
					NewText: key,
				},
			})
		}
	}

	return items, nil
}

func attemptToStripImportPrefix(key, importKey string) string {
	// we can only strip the import prefix if the rule key starts with the
	// import key
	if !strings.HasPrefix(key, importKey) {
		return ""
	}

	importKeyParts := strings.Split(importKey, ".")

	// if for some reason we have a key shorter than 'data.foo', we don't know
	// how to handle this
	if len(importKeyParts) < 2 {
		return ""
	}

	// strippablePrefix is all but the last part of the module path
	// for example 'data.foo.bar.baz' -> 'data.foo.bar.'.
	// This is what the author would need to type to access rules
	// from the imported package.
	strippablePrefix := strings.Join(importKeyParts[0:len(importKeyParts)-1], ".") + "."

	return strings.TrimPrefix(key, strippablePrefix)
}
