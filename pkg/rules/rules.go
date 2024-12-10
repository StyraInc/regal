package rules

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"

	rutil "github.com/anderseknert/roast/pkg/util"

	"github.com/open-policy-agent/opa/v1/ast"

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
	// RegoVersions stores best-effort version information from the parser.
	RegoVersions map[string]ast.RegoVersion
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
	// best-effort guess from parsing
	version ast.RegoVersion
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

func NewInputWithVersions(
	fileContent map[string]string,
	modules map[string]*ast.Module,
	versions map[string]ast.RegoVersion,
) Input {
	// Maintain order across runs
	filenames := util.Keys(modules)
	sort.Strings(filenames)

	return Input{
		FileContent:  fileContent,
		FileNames:    filenames,
		Modules:      modules,
		RegoVersions: versions,
	}
}

// InputFromPaths creates a new Input from a set of file or directory paths. Note that this function assumes that the
// paths point to valid Rego files. Use config.FilterIgnoredPaths to filter out unwanted content *before* calling this
// function.
func InputFromPaths(paths []string, versionsMap map[string]ast.RegoVersion) (Input, error) {
	if len(paths) == 1 && paths[0] == "-" {
		return inputFromStdin()
	}

	fileContent := make(map[string]string, len(paths))
	modules := make(map[string]*ast.Module, len(paths))
	versions := make(map[string]ast.RegoVersion, len(paths))

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

	parserOptions := parse.ParserOptions()

	for _, path := range paths {
		go func(path string) {
			defer wg.Done()

			parserOptions.RegoVersion = ast.RegoUndefined

			// Check if the path matches any directory where a specific Rego version is set,
			// and if so use that instead of having to parse the file (potentially multiple times)
			// in order to determine the Rego version.
			// If a project-wide version has been set, it'll be found under the path "", which will
			// always be the last entry in versionedDirs, and only match if no specific directory
			// matches.
			if len(versionsMap) > 0 {
				dir := filepath.Dir(path)
				for _, versionedDir := range versionedDirs {
					if strings.HasPrefix(dir, versionedDir) {
						parserOptions.RegoVersion = versionsMap[versionedDir]

						break
					}
				}
			}

			result, err := regoWithOpts(path, parserOptions)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors = append(errors, err)

				return
			}

			fileContent[result.name] = rutil.ByteSliceToString(result.raw)
			modules[result.name] = result.parsed
			versions[result.name] = result.version
		}(path)
	}

	wg.Wait()

	if len(errors) > 0 {
		return Input{}, fmt.Errorf("failed to parse %d module(s) — first error: %w", len(errors), errors[0])
	}

	return NewInputWithVersions(fileContent, modules, versions), nil
}

func regoWithOpts(path string, opts ast.ParserOptions) (*regoFile, error) {
	path = filepath.Clean(path)

	bs, err := os.ReadFile(path)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	regoFile := regoFile{
		name: path,
		raw:  bs,
	}

	policy := rutil.ByteSliceToString(bs)

	mod, version, err := parse.ModuleWithOpts(path, policy, opts)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	regoFile.parsed = mod
	regoFile.version = version

	return &regoFile, nil
}

func inputFromStdin() (Input, error) {
	// Ideally, we'd just pass the reader to OPA, but as the parser materializes
	// the input immediately anyway, there's currently no benefit to doing so.
	bs, err := io.ReadAll(os.Stdin)
	if err != nil {
		return Input{}, fmt.Errorf("failed to read from reader: %w", err)
	}

	policy := rutil.ByteSliceToString(bs)

	module, version, err := parse.ModuleUnknownVersionWithOpts("stdin", policy, parse.ParserOptions())
	if err != nil {
		return Input{}, fmt.Errorf("failed to parse module from stdin: %w", err)
	}

	return Input{
		FileContent:  map[string]string{"stdin": policy},
		Modules:      map[string]*ast.Module{"stdin": module},
		RegoVersions: map[string]ast.RegoVersion{"stdin": version},
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
