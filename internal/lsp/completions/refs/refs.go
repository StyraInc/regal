package refs

import (
	"fmt"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/ast"

	rast "github.com/styrainc/regal/internal/ast"
	"github.com/styrainc/regal/internal/lsp/types"
)

// ForModule returns a map of refs and details about them to be used in completions that
// were found in the given module.
func ForModule(module *ast.Module) map[string]types.Ref {
	modKey := module.Package.Path.String()

	// first, create a reference for the package using the metadata
	// if present
	packagePrettyName := strings.TrimPrefix(rast.RefToString(module.Package.Path), "data.")
	packageDescription := defaultDescription(packagePrettyName)

	packageAnnotation, ok := findAnnotationForPackage(module)
	if ok {
		packageDescription = documentAnnotatedRef(packageAnnotation)
	}

	items := map[string]types.Ref{
		modKey: {
			Label:       modKey,
			Kind:        types.Package,
			Detail:      "Package",
			Description: packageDescription,
		},
	}

	// Create groups of rules and functions sharing the same name
	ruleGroups := make(map[string][]*ast.Rule, len(module.Rules))

	for _, rule := range module.Rules {
		name := rast.RefToString(rule.Head.Ref())
		ruleGroups[name] = append(ruleGroups[name], rule)
	}

	for g, rs := range ruleGroups {
		// this should not happen, but we depend on rules being present below
		if len(rs) == 0 {
			continue
		}

		ruleKey := fmt.Sprintf("%s.%s", modKey, g)

		isConstant := true

		for _, r := range rs {
			if !rast.IsConstant(r) {
				isConstant = false

				break
			}
		}

		isFunc := false
		if rs[0].Head.Args != nil {
			isFunc = true
		}

		kind := types.Rule

		switch {
		case isConstant:
			kind = types.ConstantRule
		case isFunc:
			kind = types.Function
		}

		ruleDescription := defaultDescription(g)

		ruleAnnotation, ok := findAnnotationForRuleGroup(module, rs)
		if ok {
			ruleDescription = documentAnnotatedRef(ruleAnnotation)
		}

		items[ruleKey] = types.Ref{
			Kind:        kind,
			Label:       ruleKey,
			Detail:      rast.GetRuleDetail(rs[0]),
			Description: ruleDescription,
		}
	}

	return items
}

func defaultDescription(name string) string {
	return fmt.Sprintf(`# %s

See [METADATA Documentation](https://www.openpolicyagent.org/docs/latest/policy-language/#metadata)
to add more detail.`, name)
}

func findAnnotationForPackage(m *ast.Module) (*ast.Annotations, bool) {
	for _, a := range m.Annotations {
		if a.Scope == "package" {
			return a, true
		}
	}

	return nil, false
}

// findAnnotationForRuleGroup looks for an annotation on any of the rules in the group,
// if one is found, the first one is returned. Annotations are filtered based on the rule
// locations and the line on which the annotation ends.
func findAnnotationForRuleGroup(m *ast.Module, rs []*ast.Rule) (*ast.Annotations, bool) {
	// find all the starting locations of all the rules in the group
	ruleLocations := make([]int, len(rs))
	for i, r := range rs {
		ruleLocations[i] = r.Location.Row
	}

	// find any annotations that end on the line before any of the rules
	// and select the first one that matches
	for _, a := range m.Annotations {
		if a.Scope != "rule" {
			continue
		}

		if slices.Contains(ruleLocations, a.EndLoc().Row+1) {
			return a, true
		}
	}

	return nil, false
}

func documentAnnotatedRef(selectedAnnotation *ast.Annotations) string {
	var sb strings.Builder

	if selectedAnnotation.Title != "" {
		sb.WriteString("# ")
		sb.WriteString(selectedAnnotation.Title)
		sb.WriteString("\n\n")
	}

	if selectedAnnotation.Description != "" {
		sb.WriteString("**Description**:\n\n")
		sb.WriteString(selectedAnnotation.Description)
		sb.WriteString("\n\n")
	}

	if len(selectedAnnotation.Authors) > 0 {
		sb.WriteString("**Authors**:\n\n")

		for _, author := range selectedAnnotation.Authors {
			sb.WriteString("* ")

			if author.Name != "" {
				sb.WriteString(author.Name)
			}

			if author.Email != "" {
				sb.WriteString(" ")
				sb.WriteString("<")
				sb.WriteString(author.Email)
				sb.WriteString(">")
			}

			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	if len(selectedAnnotation.Organizations) > 0 {
		sb.WriteString("**Organizations**:\n\n")

		for _, org := range selectedAnnotation.Organizations {
			sb.WriteString("* ")
			sb.WriteString(org)
			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	if len(selectedAnnotation.RelatedResources) > 0 {
		sb.WriteString("**Related Resources**:\n\n")

		for _, resource := range selectedAnnotation.RelatedResources {
			sb.WriteString("* [")

			if resource.Description != "" {
				sb.WriteString(resource.Description)
			} else {
				sb.WriteString(strings.Replace(resource.Ref.String(), "http://", "", 1))
			}

			sb.WriteString("](")
			sb.WriteString(resource.Ref.String())
			sb.WriteString(")\n")
		}

		sb.WriteString("\n")
	}

	if len(selectedAnnotation.Custom) > 0 {
		sb.WriteString("**Custom**:\n\n```yaml\n")

		yamlBs, err := yaml.Marshal(selectedAnnotation.Custom)
		if err != nil {
			sb.WriteString("Error generating custom section")
		} else {
			sb.WriteString(strings.TrimSpace(string(yamlBs)))
		}

		sb.WriteString("\n```\n")
	}

	return sb.String()
}
