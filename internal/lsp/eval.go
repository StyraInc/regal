package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"

	"github.com/styrainc/regal/internal/lsp/uri"
	"github.com/styrainc/regal/pkg/builtins"
)

func (l *LanguageServer) Eval(ctx context.Context, query string, input string) (rego.ResultSet, error) {
	modules := l.cache.GetAllModules()
	moduleFiles := make([]bundle.ModuleFile, 0, len(modules))

	for fileURI, module := range modules {
		moduleFiles = append(moduleFiles, bundle.ModuleFile{
			URL:    fileURI,
			Parsed: module,
			Path:   uri.ToPath(l.clientIdentifier, fileURI),
		})
	}

	bd := bundle.Bundle{
		Data: make(map[string]any),
		Manifest: bundle.Manifest{
			Roots:    &[]string{""},
			Metadata: map[string]any{"name": "workspace"},
		},
		Modules: moduleFiles,
	}

	regoArgs := prepareRegoArgs(ast.MustParseBody(query), bd)

	pq, err := rego.New(regoArgs...).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed preparing query: %w", err)
	}

	if input != "" {
		inputMap := make(map[string]any)

		err = json.Unmarshal([]byte(input), &inputMap)
		if err != nil {
			return nil, fmt.Errorf("failed unmarshalling input: %w", err)
		}

		return pq.Eval(ctx, rego.EvalInput(inputMap)) //nolint:wrapcheck
	}

	return pq.Eval(ctx) //nolint:wrapcheck
}

type EvalPathResult struct {
	Value       any  `json:"value"`
	IsUndefined bool `json:"isUndefined"`
}

func FindInput(file string, workspacePath string) string {
	relative := strings.TrimPrefix(file, workspacePath)
	components := strings.Split(path.Dir(relative), string(filepath.Separator))

	for i := range len(components) {
		inputPath := path.Join(workspacePath, path.Join(components[:len(components)-i]...), "input.json")

		if input, err := os.ReadFile(inputPath); err == nil {
			return string(input)
		}
	}

	return ""
}

func (l *LanguageServer) EvalWorkspacePath(ctx context.Context, query string, input string) (EvalPathResult, error) {
	resultQuery := "result := " + query

	result, err := l.Eval(ctx, resultQuery, input)
	if err != nil {
		return EvalPathResult{}, fmt.Errorf("failed evaluating query: %w", err)
	}

	if len(result) == 0 {
		return EvalPathResult{IsUndefined: true}, nil
	}

	res, ok := result[0].Bindings["result"]
	if !ok {
		return EvalPathResult{}, errors.New("expected result in bindings, didn't get it")
	}

	return EvalPathResult{Value: res}, nil
}

func prepareRegoArgs(query ast.Body, bd bundle.Bundle) []func(*rego.Rego) {
	return []func(*rego.Rego){
		rego.ParsedQuery(query),
		rego.ParsedBundle("workspace", &bd),
		rego.Function2(builtins.RegalParseModuleMeta, builtins.RegalParseModule),
		rego.Function1(builtins.RegalLastMeta, builtins.RegalLast),
		rego.EnablePrintStatements(true),
		rego.PrintHook(topdown.NewPrintHook(os.Stderr)),
	}
}
