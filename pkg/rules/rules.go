package rules

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"

	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/report"
)

// Input represents the input for a linter evaluation.
type Input struct {
	// FileContent carries the string contents of each file
	FileContent map[string]string
	// Modules is the set of modules to lint.
	Modules map[string]*ast.Module
	// FileNames is used to maintain consistent order between runs.
	FileNames []string
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
	if len(paths) == 1 && paths[0] == "-" {
		return inputFromStdin()
	}

	return inputFromPathsFS(os.DirFS("/"), paths)
}

func inputFromPathsFS(fsys fs.FS, paths []string) (Input, error) {
	fileContent := make(map[string]string, len(paths))
	modules := make(map[string]*ast.Module, len(paths))

	var mu sync.Mutex

	var wg sync.WaitGroup

	wg.Add(len(paths))

	errors := make([]error, 0, len(paths))

	for _, path := range paths {
		go func(path string) {
			defer wg.Done()

			result, err := readAndParse(fsys, path)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors = append(errors, err)
				return
			}

			fileContent[result.Name] = string(result.Raw)
			modules[result.Name] = result.Parsed
		}(path)
	}

	wg.Wait()

	if len(errors) > 0 {
		return Input{}, fmt.Errorf("failed to parse %d module(s) — first error: %w", len(errors), errors[0])
	}

	return NewInput(fileContent, modules), nil
}

func readAndParse(fsys fs.FS, path string) (*loader.RegoFile, error) {
	if path[0] != '/' { // make relative OS path absolute
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}

		path = filepath.Join(cwd, path)
	}

	if path[0] == '/' {
		path = path[1:] // fs.FS doesn't consider paths starting with '/' valid: https://pkg.go.dev/io/fs#ValidPath
	}

	bs, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("read file from %v%s: %w", fsys, path, err)
	}

	module, err := ast.ParseModuleWithOpts(path, string(bs), parse.ParserOptions())
	if err != nil {
		return nil, fmt.Errorf("parse file from %v%s: %w", fsys, path, err)
	}

	return &loader.RegoFile{
		Name:   path,
		Parsed: module,
		Raw:    bs,
	}, nil
}

func inputFromStdin() (Input, error) {
	// Ideally, we'd just pass the reader to OPA, but as the parser materializes
	// the input immediately anyway, there's currently no benefit to doing so.
	bs, err := io.ReadAll(os.Stdin)
	if err != nil {
		return Input{}, fmt.Errorf("failed to read from reader: %w", err)
	}

	policy := string(bs)

	module, err := parse.Module("stdin", policy)
	if err != nil {
		return Input{}, fmt.Errorf("failed to parse module from stdin: %w", err)
	}

	return NewInput(map[string]string{"stdin": policy}, map[string]*ast.Module{"stdin": module}), nil
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
