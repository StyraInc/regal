package io

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/bundle"
	ofilter "github.com/open-policy-agent/opa/v1/loader/filter"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"
	outil "github.com/open-policy-agent/opa/v1/util"

	"github.com/open-policy-agent/regal/internal/io/files"
	"github.com/open-policy-agent/regal/internal/io/files/filter"
	"github.com/open-policy-agent/regal/pkg/roast/encoding"
)

// Getwd returns the current working directory, or an empty string if it cannot be determined.
func Getwd() string {
	wd, _ := os.Getwd()

	return wd
}

// LoadRegalBundleFS loads bundle embedded from policy and data directory.
func LoadRegalBundleFS(fs fs.FS) (*bundle.Bundle, error) {
	embedLoader, err := bundle.NewFSLoader(fs)
	if err != nil {
		return nil, fmt.Errorf("failed to load bundle from filesystem: %w", err)
	}

	b, err := bundle.NewCustomReader(embedLoader.WithFilter(ExcludeTestLegacyFilter())).
		WithCapabilities(Capabilities()).
		WithSkipBundleVerification(true).
		WithProcessAnnotations(true).
		WithBundleName("regal").
		Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read bundle from filesystem: %w", err)
	}

	return &b, nil
}

// LoadRegalBundlePath loads bundle from path.
func LoadRegalBundlePath(path string) (*bundle.Bundle, error) {
	root, err := os.OpenRoot(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open root: %w", err)
	}

	return LoadRegalBundleFS(root.FS())
}

// MustLoadRegalBundleFS loads bundle embedded from policy and data directory, exit on failure.
func MustLoadRegalBundleFS(fs fs.FS) *bundle.Bundle {
	regalBundle, err := LoadRegalBundleFS(fs)
	if err != nil {
		log.Fatal(err)
	}

	return regalBundle
}

// ToMap convert any value to map[string]any, or panics on failure.
func ToMap(a any) map[string]any {
	r := make(map[string]any)

	encoding.MustJSONRoundTrip(a, &r)

	return r
}

// CloseFileIgnore closes file ignoring errors, mainly for deferred cleanup.
func CloseFileIgnore(file *os.File) {
	if file != nil {
		_ = file.Close()
	}
}

func ExcludeTestLegacyFilter() ofilter.LoaderFilter {
	return func(_ string, info fs.FileInfo, _ int) bool {
		return strings.HasSuffix(info.Name(), "_test.rego") &&
			// (anderseknert): This is an outlier, but not sure we need something
			// more polished to deal with this for the time being.
			info.Name() != "todo_test.rego"
	}
}

func IsDir(path string) bool {
	info, err := os.Stat(path)

	return err == nil && info.IsDir()
}

func IsFile(path string) bool {
	info, err := os.Stat(path)

	return err == nil && !info.IsDir()
}

func Exists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

// WithCreateRecursive creates a file at the given path, ensuring that all parent directories
// are created recursively. It then calls the provided function with the created file as an argument
// before closing the file.
func WithCreateRecursive(path string, fn func(f *os.File) error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o770); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return fn(file)
}

// FindInputPath consults the filesystem and returns the input.json or input.yaml closes to the
// file provided as arguments.
func FindInputPath(file string, workspacePath string) string {
	relative := strings.TrimPrefix(file, workspacePath)
	components := strings.Split(filepath.Dir(relative), string(os.PathSeparator))
	supported := []string{"input.json", "input.yaml"}

	for i := range components {
		current := components[:len(components)-i]

		prefix := filepath.Join(append([]string{workspacePath}, current...)...)
		for _, name := range supported {
			inputPath := filepath.Join(prefix, name)
			if _, err := os.Stat(inputPath); err == nil {
				return inputPath
			}
		}
	}

	return ""
}

// FindInput finds input.json or input.yaml file in workspace closest to the file, and returns
// both the location and the contents of the file (as map), or an empty string and nil if not found.
// Note that:
// - This function doesn't do error handling. If the file can't be read, nothing is returned.
// - While the input data theoretically could be anything JSON/YAML value, we only support an object.
func FindInput(file string, workspacePath string) (inputPath string, input map[string]any) {
	relative := strings.TrimPrefix(file, workspacePath)
	components := strings.Split(filepath.Dir(relative), string(os.PathSeparator))
	supported := []string{"input.json", "input.yaml"}

	for i := range components {
		current := filepath.Join(components[:len(components)-i]...)
		for _, name := range supported {
			inputPath := filepath.Join(workspacePath, current, name)
			if content, err := os.ReadFile(inputPath); err == nil {
				if err = unmarshallerFor(name)(content, &input); err == nil {
					return inputPath, input
				}
			}
		}
	}

	return "", nil
}

func unmarshallerFor(name string) func([]byte, any) error {
	switch name {
	case "input.json":
		return encoding.JSON().Unmarshal
	case "input.yaml", "input.yml":
		return yaml.Unmarshal
	}

	panic("no decoder for file type: " + name)
}

// FindManifestLocations walks the file system rooted at root, and returns the
// *relative* paths of directories containing a .manifest file.
func FindManifestLocations(root string) ([]string, error) {
	var foundBundleRoots []string

	return files.DefaultWalkReducer(root, foundBundleRoots).
		WithFilters(filter.Not(filter.Filenames(".manifest"))).
		Reduce(func(path string, curr []string) ([]string, error) {
			rel, err := filepath.Rel(root, path)
			if err == nil {
				curr = append(curr, filepath.Dir(rel))
			}

			return curr, err
		})
}

func ModulesFromCustomRuleFS(customRuleFS fs.FS, rootPath string) (map[string]*ast.Module, error) {
	modules, err := files.DefaultWalkReducer(rootPath, make(map[string]*ast.Module)).
		WithFilters(filter.RegoTests).
		ReduceFS(customRuleFS, func(path string, modules map[string]*ast.Module) (map[string]*ast.Module, error) {
			bs, err := fs.ReadFile(customRuleFS, path)
			if err != nil {
				return modules, fmt.Errorf("failed to read custom rule file: %w", err)
			}

			m, err := ast.ParseModule(path, outil.ByteSliceToString(bs))
			if err != nil {
				return modules, fmt.Errorf("failed to parse custom rule file %q: %w", path, err)
			}

			modules[path] = m

			return modules, nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to walk custom rule FS: %w", err)
	}

	return modules, nil
}

// NOTE: These are mirrored here merely to provide correct capabilities for the
// parser. Can't import other packages from here as almost all of them depend on
// this package. We should probably move these definitions to somewhere where all
// packages can import them from.

var regalParseModuleMeta = &rego.Function{
	Name: "regal.parse_module",
	Decl: types.NewFunction(
		types.Args(
			types.Named("filename", types.S).Description("file name to attach to AST nodes' locations"),
			types.Named("rego", types.S).Description("Rego module"),
		),
		types.Named("output", types.NewObject(nil, types.NewDynamicProperty(types.S, types.A))),
	),
}

var regalLastMeta = &rego.Function{
	Name: "regal.last",
	Decl: types.NewFunction(
		types.Args(
			types.Named("array", types.NewArray(nil, types.A)).
				Description("performance optimized last index retrieval"),
		),
		types.Named("element", types.A),
	),
}

var OPACapabilities = ast.CapabilitiesForThisVersion()

var Capabilities = sync.OnceValue(capabilities)

func capabilities() *ast.Capabilities {
	cpy := *OPACapabilities
	cpy.Builtins = append(cpy.Builtins,
		&ast.Builtin{
			Name: regalParseModuleMeta.Name,
			Decl: regalParseModuleMeta.Decl,
		},
		&ast.Builtin{
			Name: regalLastMeta.Name,
			Decl: regalLastMeta.Decl,
		})

	return &cpy
}
