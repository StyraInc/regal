package rules

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/util"

	"github.com/styrainc/regal/internal/parse"
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

type regoFile struct {
	name   string
	parsed *ast.Module
	raw    []byte
}

// NewInput creates a new Input from a set of modules.
func NewInput(fileContent map[string]string, modules map[string]*ast.Module) Input {
	// Maintain order across runs
	filenames := util.KeysSorted(modules)

	return Input{
		FileContent: fileContent,
		FileNames:   filenames,
		Modules:     modules,
	}
}

// InputFromPaths creates a new Input from a set of file or directory paths. Note that this function assumes that the
// paths point to valid Rego files. Use config.FilterIgnoredPaths to filter out unwanted content *before* calling this
// function. When the versionsMap is not nil/empty, files in a directory matching a key in the map will be parsed with
// the corresponding Rego version. If not provided, the file may be parsed multiple times in order to determine the
// version (best-effort and may include false positives).
func InputFromPaths(paths []string, prefix string, versionsMap map[string]ast.RegoVersion) (Input, error) {
	if len(paths) == 1 && paths[0] == "-" {
		return inputFromStdin()
	}

	fileContent := make(map[string]string, len(paths))
	modules := make(map[string]*ast.Module, len(paths))

	var versionedDirs []string

	if len(versionsMap) > 0 {
		versionedDirs = util.Keys(versionsMap)
		// Sort directories by length, so that the most specific path is found first
		slices.Sort(versionedDirs)
		slices.Reverse(versionedDirs)
	}

	var mu sync.Mutex

	var wg sync.WaitGroup

	wg.Add(len(paths))

	errors := make([]error, 0, len(paths))

	for _, path := range paths {
		go func(path string) {
			defer wg.Done()

			parserOptions := parse.ParserOptions()

			parserOptions.RegoVersion = RegoVersionFromVersionsMap(
				versionsMap,
				strings.TrimPrefix(path, prefix),
				ast.RegoUndefined,
			)

			result, err := regoWithOpts(path, parserOptions)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors = append(errors, err)

				return
			}

			fileContent[result.name] = util.ByteSliceToString(result.raw)
			modules[result.name] = result.parsed
		}(path)
	}

	wg.Wait()

	if len(errors) > 0 {
		return Input{}, fmt.Errorf("failed to parse %d module(s) — first error: %w", len(errors), errors[0])
	}

	return NewInput(fileContent, modules), nil
}

// InputFromMap creates a new Input from a map of file paths to their contents.
// This function uses a vesrionsMap to determine the parser version for each
// file before parsing the module.
func InputFromMap(files map[string]string, versionsMap map[string]ast.RegoVersion) (Input, error) {
	fileContent := make(map[string]string, len(files))
	modules := make(map[string]*ast.Module, len(files))
	parserOptions := parse.ParserOptions()

	for path, content := range files {
		fileContent[path] = content

		parserOptions.RegoVersion = RegoVersionFromVersionsMap(versionsMap, path, ast.RegoUndefined)

		mod, err := parse.ModuleWithOpts(path, content, parserOptions)
		if err != nil {
			return Input{}, fmt.Errorf("failed to parse module %s: %w", path, err)
		}

		modules[path] = mod
	}

	return NewInput(fileContent, modules), nil
}

// RegoVersionFromVersionsMap takes a mapping of file path prefixes, typically
// representing the roots of the project, and the expected Rego version for
// each. Using this, it finds the longest matching prefix for the given filename
// and returns the defaultVersion if to matching prefix is found.
func RegoVersionFromVersionsMap(
	versionsMap map[string]ast.RegoVersion,
	filename string,
	defaultVersion ast.RegoVersion,
) ast.RegoVersion {
	if len(versionsMap) == 0 {
		return defaultVersion
	}

	selectedVersion := defaultVersion

	var longestMatch int

	dir := filepath.Dir(filename)

	for versionedDir := range versionsMap {
		matchingVersionedDir := path.Join("/", versionedDir, "/")

		if strings.HasPrefix(dir+"/", matchingVersionedDir) {
			// >= as the versioned dir might be "" for the project root
			if len(versionedDir) >= longestMatch {
				longestMatch = len(versionedDir)
				selectedVersion = versionsMap[versionedDir]
			}
		}
	}

	return selectedVersion
}

func regoWithOpts(path string, opts ast.ParserOptions) (*regoFile, error) {
	path = filepath.Clean(path)

	bs, err := os.ReadFile(path)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	mod, err := parse.ModuleWithOpts(path, util.ByteSliceToString(bs), opts)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &regoFile{name: path, raw: bs, parsed: mod}, nil
}

func inputFromStdin() (Input, error) {
	// Ideally, we'd just pass the reader to OPA, but as the parser materializes
	// the input immediately anyway, there's currently no benefit to doing so.
	bs, err := io.ReadAll(os.Stdin)
	if err != nil {
		return Input{}, fmt.Errorf("failed to read from reader: %w", err)
	}

	policy := util.ByteSliceToString(bs)

	module, err := parse.ModuleUnknownVersionWithOpts("stdin", policy, parse.ParserOptions())
	if err != nil {
		return Input{}, fmt.Errorf("failed to parse module from stdin: %w", err)
	}

	return Input{
		FileContent: map[string]string{"stdin": policy},
		Modules:     map[string]*ast.Module{"stdin": module},
	}, nil
}

// InputFromText creates a new Input from raw Rego text.
func InputFromText(fileName, text string) (Input, error) {
	mod, err := parse.Module(fileName, text)
	if err != nil {
		return Input{}, fmt.Errorf("failed to parse module: %w", err)
	}

	return NewInput(map[string]string{fileName: text}, map[string]*ast.Module{fileName: mod}), nil
}

// InputFromTextWithOptions creates a new Input from raw Rego text while
// respecting the provided options.
func InputFromTextWithOptions(fileName, text string, opts ast.ParserOptions) (Input, error) {
	mod, err := ast.ParseModuleWithOpts(fileName, text, opts)
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
