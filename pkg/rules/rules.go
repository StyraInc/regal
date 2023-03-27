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
	// FileBytes carries the raw bytes of each file
	FileBytes map[string][]byte
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
func NewInput(fileBytes map[string][]byte, modules map[string]*ast.Module) Input {
	// Maintain order across runs
	filenames := util.Keys(modules)
	sort.Strings(filenames)

	return Input{
		FileBytes: fileBytes,
		FileNames: filenames,
		Modules:   modules,
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

	filebytes := make(map[string][]byte, len(policyPaths))
	modules := make(map[string]*ast.Module, len(policyPaths))

	for _, path := range policyPaths {
		result, err := loader.RegoWithOpts(path, parse.ParserOptions())
		if err != nil {
			// TODO: Keep running and collect errors instead?
			return Input{}, err //nolint:wrapcheck
		}

		filebytes[result.Name] = result.Raw
		modules[result.Name] = result.Parsed
	}

	return NewInput(filebytes, modules), nil
}
