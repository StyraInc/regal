package lsp

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/anderseknert/roast/pkg/encoding"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown/print"

	"github.com/styrainc/regal/internal/lsp/uri"
	"github.com/styrainc/regal/pkg/builtins"
)

func (l *LanguageServer) Eval(
	ctx context.Context,
	query string,
	input io.Reader,
	printHook print.Hook,
	dataBundles map[string]bundle.Bundle,
) (rego.ResultSet, error) {
	modules := l.cache.GetAllModules()
	moduleFiles := make([]bundle.ModuleFile, 0, len(modules))

	for fileURI, module := range modules {
		moduleFiles = append(moduleFiles, bundle.ModuleFile{
			URL:    fileURI,
			Parsed: module,
			Path:   uri.ToPath(l.clientIdentifier, fileURI),
		})
	}

	allBundles := make(map[string]bundle.Bundle)

	for k, v := range dataBundles {
		if v.Manifest.Roots == nil {
			l.logError(fmt.Errorf("bundle %s has no roots and will be skipped", k))

			continue
		}

		allBundles[k] = v
	}

	allBundles["workspace"] = bundle.Bundle{
		Manifest: bundle.Manifest{
			// there is no data in this bundle so the roots are not used,
			// however, roots must be set.
			Roots:    &[]string{"workspace"},
			Metadata: map[string]any{"name": "workspace"},
		},
		Modules: moduleFiles,
		// Data is all sourced from the dataBundles instead
		Data: make(map[string]any),
	}

	regoArgs := prepareRegoArgs(ast.MustParseBody(query), allBundles, printHook)

	pq, err := rego.New(regoArgs...).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed preparing query %s: %w", query, err)
	}

	if input != nil {
		inputMap := make(map[string]any)

		in, err := io.ReadAll(input)
		if err != nil {
			return nil, fmt.Errorf("failed reading input: %w", err)
		}

		json := encoding.JSON()

		err = json.Unmarshal(in, &inputMap)
		if err != nil {
			return nil, fmt.Errorf("failed unmarshalling input: %w", err)
		}

		return pq.Eval(ctx, rego.EvalInput(inputMap)) //nolint:wrapcheck
	}

	return pq.Eval(ctx) //nolint:wrapcheck
}

type EvalPathResult struct {
	Value       any                         `json:"value"`
	IsUndefined bool                        `json:"isUndefined"`
	PrintOutput map[string]map[int][]string `json:"printOutput"`
}

func (l *LanguageServer) EvalWorkspacePath(
	ctx context.Context,
	query string,
	input io.Reader,
) (EvalPathResult, error) {
	resultQuery := "result := " + query

	hook := PrintHook{Output: make(map[string]map[int][]string)}

	var bs map[string]bundle.Bundle
	if l.bundleCache != nil {
		bs = l.bundleCache.All()
	}

	result, err := l.Eval(ctx, resultQuery, input, hook, bs)
	if err != nil {
		return EvalPathResult{}, fmt.Errorf("failed evaluating query: %w", err)
	}

	if len(result) == 0 {
		return EvalPathResult{IsUndefined: true, PrintOutput: hook.Output}, nil
	}

	res, ok := result[0].Bindings["result"]
	if !ok {
		return EvalPathResult{}, errors.New("expected result in bindings, didn't get it")
	}

	return EvalPathResult{Value: res, PrintOutput: hook.Output}, nil
}

func prepareRegoArgs(query ast.Body, bundles map[string]bundle.Bundle, printHook print.Hook) []func(*rego.Rego) {
	bundleArgs := make([]func(*rego.Rego), 0, len(bundles))
	for key, b := range bundles {
		bundleArgs = append(bundleArgs, rego.ParsedBundle(key, &b))
	}

	baseArgs := []func(*rego.Rego){
		rego.ParsedQuery(query),
		rego.Function2(builtins.RegalParseModuleMeta, builtins.RegalParseModule),
		rego.Function1(builtins.RegalLastMeta, builtins.RegalLast),
		rego.EnablePrintStatements(true),
		rego.PrintHook(printHook),
	}

	return append(baseArgs, bundleArgs...)
}

type PrintHook struct {
	Output map[string]map[int][]string
}

func (h PrintHook) Print(ctx print.Context, msg string) error {
	if _, ok := h.Output[ctx.Location.File]; !ok {
		h.Output[ctx.Location.File] = make(map[int][]string)
	}

	h.Output[ctx.Location.File][ctx.Location.Row] = append(h.Output[ctx.Location.File][ctx.Location.Row], msg)

	return nil
}
