package io

import (
	"fmt"
	"io"
	files "io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/bundle"
	"github.com/open-policy-agent/opa/v1/loader/filter"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"
	outil "github.com/open-policy-agent/opa/v1/util"

	"github.com/styrainc/regal/pkg/roast/encoding"
)

const PathSeparator = string(os.PathSeparator)

// LoadRegalBundleFS loads bundle embedded from policy and data directory.
func LoadRegalBundleFS(fs files.FS) (bundle.Bundle, error) {
	embedLoader, err := bundle.NewFSLoader(fs)
	if err != nil {
		return bundle.Bundle{}, fmt.Errorf("failed to load bundle from filesystem: %w", err)
	}

	//nolint:wrapcheck
	return bundle.NewCustomReader(embedLoader.WithFilter(ExcludeTestFilter())).
		WithCapabilities(Capabilities()).
		WithSkipBundleVerification(true).
		WithProcessAnnotations(true).
		WithBundleName("regal").
		Read()
}

// LoadRegalBundlePath loads bundle from path.
func LoadRegalBundlePath(path string) (bundle.Bundle, error) {
	//nolint:wrapcheck
	return bundle.NewCustomReader(bundle.NewDirectoryLoader(path).WithFilter(ExcludeTestFilter())).
		WithCapabilities(Capabilities()).
		WithSkipBundleVerification(true).
		WithProcessAnnotations(true).
		WithBundleName("regal").
		Read()
}

// MustLoadRegalBundleFS loads bundle embedded from policy and data directory, exit on failure.
func MustLoadRegalBundleFS(fs files.FS) bundle.Bundle {
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
	_ = file.Close()
}

func ExcludeTestFilter() filter.LoaderFilter {
	return func(_ string, info files.FileInfo, _ int) bool {
		return strings.HasSuffix(info.Name(), "_test.rego") &&
			// (anderseknert): This is an outlier, but not sure we need something
			// more polished to deal with this for the time being.
			info.Name() != "todo_test.rego"
	}
}

// FindInputPath consults the filesystem and returns the input.json or input.yaml closes to the
// file provided as arguments.
func FindInputPath(file string, workspacePath string) string {
	relative := strings.TrimPrefix(file, workspacePath)
	components := strings.Split(filepath.Dir(relative), PathSeparator)

	for i := range components {
		current := components[:len(components)-i]
		prefix := filepath.Join(append([]string{workspacePath}, current...)...)

		for _, ext := range []string{"json", "yaml", "yml"} {
			inputPath := filepath.Join(prefix, "input."+ext)

			_, err := os.Stat(inputPath)
			if err == nil {
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
func FindInput(file string, workspacePath string) (string, map[string]any) {
	relative := strings.TrimPrefix(file, workspacePath)
	components := strings.Split(filepath.Dir(relative), PathSeparator)

	var (
		inputPath string
		content   []byte
	)

	for i := range components {
		current := components[:len(components)-i]

		inputPathJSON := filepath.Join(workspacePath, filepath.Join(current...), "input.json")

		f, err := os.Open(inputPathJSON)
		if err == nil {
			inputPath = inputPathJSON
			content, _ = io.ReadAll(f)

			break
		}

		inputPathYAML := filepath.Join(workspacePath, filepath.Join(current...), "input.yaml")

		f, err = os.Open(inputPathYAML)
		if err == nil {
			inputPath = inputPathYAML
			content, _ = io.ReadAll(f)

			break
		}
	}

	if inputPath == "" || content == nil {
		return "", nil
	}

	var input map[string]any

	if strings.HasSuffix(inputPath, ".json") {
		if err := encoding.JSON().Unmarshal(content, &input); err != nil {
			return "", nil
		}
	} else if strings.HasSuffix(inputPath, ".yaml") {
		if err := yaml.Unmarshal(content, &input); err != nil {
			return "", nil
		}
	}

	return inputPath, input
}

func IsSkipWalkDirectory(info files.DirEntry) bool {
	return info.IsDir() && (info.Name() == ".git" || info.Name() == ".idea" || info.Name() == "node_modules")
}

// WalkFiles walks the file system rooted at root, calling f for each file. This is
// a less ceremonious version of filepath.WalkDir where only file paths (not dirs)
// are passed to the callback, and where directories that should commonly  be ignored
// (.git, node_modules, etc.) are skipped.
func WalkFiles(root string, f func(path string) error) error {
	return filepath.WalkDir(root, func(path string, info os.DirEntry, _ error) error { //nolint:wrapcheck
		if IsSkipWalkDirectory(info) {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		return f(path)
	})
}

// FindManifestLocations walks the file system rooted at root, and returns the
// *relative* paths of directories containing a .manifest file.
func FindManifestLocations(root string) ([]string, error) {
	var foundBundleRoots []string

	if err := WalkFiles(root, func(path string) error {
		if filepath.Base(path) == ".manifest" {
			if rel, err := filepath.Rel(root, path); err == nil {
				foundBundleRoots = append(foundBundleRoots, filepath.Dir(rel))
			}
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk workspace path: %w", err)
	}

	return foundBundleRoots, nil
}

func ModulesFromCustomRuleFS(customRuleFS files.FS, rootPath string) (map[string]*ast.Module, error) {
	modules := make(map[string]*ast.Module)
	filter := ExcludeTestFilter()

	err := files.WalkDir(customRuleFS, rootPath, func(path string, d files.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk custom rule FS: %w", err)
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get info for custom rule file: %w", err)
		}

		if filter("", info, 0) {
			return nil
		}

		f, err := customRuleFS.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open custom rule file: %w", err)
		}
		defer f.Close()

		bs, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("failed to read custom rule file: %w", err)
		}

		m, err := ast.ParseModule(path, outil.ByteSliceToString(bs))
		if err != nil {
			return fmt.Errorf("failed to parse custom rule file %q: %w", path, err)
		}

		modules[path] = m

		return nil
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
