package test

import (
	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/pkg/rules"
)

func InputPolicy(filename string, policy string) rules.Input {
	content := map[string]string{filename: policy}
	modules := map[string]*ast.Module{filename: parse.MustParseModule(policy)}

	return rules.NewInput(content, modules)
}
