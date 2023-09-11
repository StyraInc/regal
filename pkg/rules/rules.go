package rules

import (
	"context"
	"fmt"
	"sort"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"

	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/report"
)

// Input represents the input for a linter evaluation.
type Input struct {
	// FileNames is used to maintain consistent order between runs.
	FileNames []string
	// FileContent carries the string contents of each file
	FileContent map[string]string
	// Modules is the set of modules to lint.
	Modules map[string]*ast.Module
}

// Rule represents a linter rule.
type Rule interface {
	// Run runs the rule on the provided input.
	Run(context.Context, Input) (*report.Report, error)
	// Name returns the name of the rule.
	Name() string
	// Category returns the category of the rule.
	Category() string
	// Description returns the description of the rule.
	Description() string
	// Documentation returns the documentation URL for the rule.
	Documentation() string
	// Config returns the provided configuration for the rule
	Config() config.Rule
}

// NewInput creates a new Input from a set of modules.
func NewInput(fileContent map[string]string, modules map[string]*ast.Module) Input {
	// Maintain order across runs
	filenames := util.Keys(modules)
	sort.Strings(filenames)

	return Input{
		FileContent: fileContent,
		FileNames:   filenames,
		Modules:     modules,
	}
}

// InputFromPaths creates a new Input from a set of file or directory paths. Note that this function assumes that the
// paths point to valid Rego files. Use config.FilterIgnoredPaths to filter out unwanted content *before* calling this
// function.
func InputFromPaths(paths []string) (Input, error) {
	fileContent := make(map[string]string, len(paths))
	modules := make(map[string]*ast.Module, len(paths))

	for _, path := range paths {
		result, err := loader.RegoWithOpts(path, parse.ParserOptions())
		if err != nil {
			// TODO: Keep running and collect errors instead?
			return Input{}, err //nolint:wrapcheck
		}

		fileContent[result.Name] = string(result.Raw)
		modules[result.Name] = result.Parsed
	}

	return NewInput(fileContent, modules), nil
}

// InputFromText creates a new Input from raw Rego text.
func InputFromText(fileName, text string) (Input, error) {
	mod, err := parse.Module(fileName, text)
	if err != nil {
		return Input{}, fmt.Errorf("failed to parse module: %w", err)
	}

	return NewInput(map[string]string{fileName: text}, map[string]*ast.Module{fileName: mod}), nil
}

func AllGoRules(conf config.Config) []Rule {
	return []Rule{
		NewOpaFmtRule(conf),
	}
}
