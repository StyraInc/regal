package rules

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/loader"

	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/util"
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

// InputFromPaths creates a new Input from a set of file or directory paths.
func InputFromPaths(paths []string) (Input, error) {
	policyPaths, err := loader.FilteredPaths(paths, func(_ string, info os.FileInfo, depth int) bool {
		return !info.IsDir() && !strings.HasSuffix(info.Name(), bundle.RegoExt)
	})
	if err != nil {
		return Input{}, fmt.Errorf("failed to load policy from provided args: %w", err)
	}

	fileContent := make(map[string]string, len(policyPaths))
	modules := make(map[string]*ast.Module, len(policyPaths))

	for _, path := range policyPaths {
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
