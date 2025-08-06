package lsp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/bundle"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/topdown/print"

	rbundle "github.com/styrainc/regal/bundle"
	"github.com/styrainc/regal/internal/lsp/log"
	"github.com/styrainc/regal/pkg/builtins"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/roast/transform"
)

var (
	emptyStringAnyMap = make(map[string]any, 0)

	workspaceBundleManifest = bundle.Manifest{
		// there is no data in this bundle so the roots are not used,
		// however, roots must be set.
		Roots:    &[]string{"workspace"},
		Metadata: map[string]any{"name": "workspace"},
	}
)

type EvalPathResult struct {
	Value       any                         `json:"value"`
	PrintOutput map[string]map[int][]string `json:"printOutput"`
	IsUndefined bool                        `json:"isUndefined"`
}

type PrintHook struct {
	Output map[string]map[int][]string
}

func (l *LanguageServer) Eval(
	ctx context.Context,
	query string,
	input map[string]any,
	printHook print.Hook,
	dataBundles map[string]bundle.Bundle,
) (rego.ResultSet, error) {
	modules := l.cache.GetAllModules()
	moduleFiles := make([]bundle.ModuleFile, 0, len(modules))

	var hasCustomRules bool

	for fileURI, module := range modules {
		moduleFiles = append(moduleFiles, bundle.ModuleFile{URL: fileURI, Parsed: module, Path: l.toPath(fileURI)})
		hasCustomRules = hasCustomRules || strings.Contains(module.Package.Path.String(), "custom.regal.rules")
	}

	allBundles := make(map[string]bundle.Bundle, len(dataBundles)+2)

	for k := range dataBundles {
		if dataBundles[k].Manifest.Roots == nil {
			l.logf(log.LevelMessage, "bundle %s has no roots and will be skipped", k)

			continue
		}

		allBundles[k] = dataBundles[k]
	}

	allBundles["workspace"] = bundle.Bundle{
		Manifest: workspaceBundleManifest,
		Modules:  moduleFiles,
		Data:     emptyStringAnyMap, // Data is sourced from the dataBundles instead
	}

	if hasCustomRules {
		// If someone evaluates a custom Regal rule, provide them the Regal bundle
		// in order to make all Regal functions available
		allBundles["regal"] = *rbundle.LoadedBundle()
	}

	regoArgs := prepareRegoArgs(ast.MustParseBody(query), allBundles, printHook, l.getLoadedConfig())

	// TODO: Let's try to avoid preparing on each eval, but only when the
	// contents of the workspace modules change, and before the user requests
	// an eval.
	pq, err := rego.New(regoArgs...).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed preparing query %s: %w", query, err)
	}

	if input != nil {
		if inputValue, err := transform.ToOPAInputValue(input); err != nil {
			return nil, fmt.Errorf("failed converting input to value: %w", err)
		} else {
			return pq.Eval(ctx, rego.EvalParsedInput(inputValue)) //nolint:wrapcheck
		}
	}

	return pq.Eval(ctx) //nolint:wrapcheck
}

func (l *LanguageServer) EvalWorkspacePath(
	ctx context.Context,
	query string,
	input map[string]any,
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

func prepareRegoArgs(
	query ast.Body,
	bundles map[string]bundle.Bundle,
	printHook print.Hook,
	cfg *config.Config,
) []func(*rego.Rego) {
	bundleArgs := make([]func(*rego.Rego), 0, len(bundles))
	// this copy is expensive, but I don't think we can avoid it
	//nolint:gocritic
	for key, b := range bundles {
		bundleArgs = append(bundleArgs, rego.ParsedBundle(key, &b))
	}

	args := []func(*rego.Rego){rego.ParsedQuery(query), rego.EnablePrintStatements(true), rego.PrintHook(printHook)}
	args = append(args, builtins.RegalBuiltinRegoFuncs...)
	args = append(args, bundleArgs...)

	var caps *config.Capabilities
	if cfg != nil && cfg.Capabilities != nil {
		caps = cfg.Capabilities
	} else {
		caps = config.CapabilitiesForThisVersion()
	}

	var evalConfig config.Config
	if cfg != nil {
		evalConfig = *cfg
	}

	internalBundle := &bundle.Bundle{
		Manifest: bundle.Manifest{
			Roots:    &[]string{"internal"},
			Metadata: map[string]any{"name": "internal"},
		},
		Data: map[string]any{
			"internal": map[string]any{
				"combined_config": config.ToMap(evalConfig),
				"capabilities":    caps,
			},
		},
	}

	return append(args, rego.ParsedBundle("internal", internalBundle))
}

func (h PrintHook) Print(ctx print.Context, msg string) error {
	if _, ok := h.Output[ctx.Location.File]; !ok {
		h.Output[ctx.Location.File] = make(map[int][]string)
	}

	h.Output[ctx.Location.File][ctx.Location.Row] = append(h.Output[ctx.Location.File][ctx.Location.Row], msg)

	return nil
}
