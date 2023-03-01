package linter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"

	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/util"
)

// Linter stores data to use for linting
type Linter struct {
	ruleBundles []*bundle.Bundle
}

// NewLinter creates a new Regal linter
func NewLinter() Linter {
	return Linter{}
}

// WithAddedBundle adds a bundle of rules and data to include in evaluation
func (l Linter) WithAddedBundle(b bundle.Bundle) Linter {
	l.ruleBundles = append(l.ruleBundles, &b)

	return l
}

// Lint runs the linter on provided policies
func (l Linter) Lint(ctx context.Context, result *loader.Result) {
	var regoArgs []func(*rego.Rego)
	regoArgs = append(regoArgs,
		rego.Query("report = data.regal.main.report"),
		rego.EnablePrintStatements(true),
		rego.PrintHook(topdown.NewPrintHook(os.Stderr)))

	if l.ruleBundles != nil {
		for _, ruleBundle := range l.ruleBundles {
			var bundleName string
			if metadataName, ok := ruleBundle.Manifest.Metadata["name"].(string); ok {
				bundleName = metadataName
			}
			regoArgs = append(regoArgs, rego.ParsedBundle(bundleName, ruleBundle))
		}
	}

	query, err := rego.New(regoArgs...).PrepareForEval(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Maintain order across runs
	modules := util.Keys(result.Modules)
	sort.Strings(modules)

	for _, name := range modules {
		module := result.Modules[name]
		astJSON := rio.MustJSON(module.Parsed)
		var input map[string]any
		if err = json.Unmarshal(astJSON, &input); err != nil {
			log.Fatal(err)
		}

		resultSet, err := query.Eval(ctx, rego.EvalInput(input))
		if err != nil {
			log.Fatal(err)
		}

		// TODO: Create a reporter interface and implementations
		for _, result := range resultSet {
			report := result.Bindings["report"].([]interface{})
			for _, entry := range report {
				violation := entry.(map[string]interface{})
				fmt.Printf("%v: %v\n", name, violation["description"])
			}
		}
	}
}
