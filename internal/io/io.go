package io

import (
	"errors"
	"fmt"
	"io"
	files "io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
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

// FileWalkerFilter is a function that filters files and directories during a walk.
type FileWalkerFilter func(path string, info os.DirEntry) bool

// FileWalker is a utility for walking directories and files, applying
// provided filters before processing each file or directory.
type FileWalker struct {
	// root is the root directory to start walking from.
	root string
	// filters are the filters to apply to each file or directory.
	// If any filter returns true, the file or directory is skipped.
	filters []FileWalkerFilter
	// skipFunc is a special filter that allows skipping entire directory
	// entire directory trees (like .git, node_modules, etc.).
	skipFunc FileWalkerFilter
	// statRoot indicates whether to stat the root directory before walking.
	statRoot bool
}

// FileWalkerReducer extends FileWalker to allow reducing the results of the walk.
type FileWalkerReducer[T any] struct {
	*FileWalker

	initial T
}

// DirectoryFilter is a filter for filtering directories.
func DirectoryFilter(_ string, info os.DirEntry) bool {
	return info.IsDir()
}

// FileNameFilter filters files by their exact name, not counting the path.
func FileNameFilter(names ...string) FileWalkerFilter {
	return func(_ string, info os.DirEntry) bool {
		return slices.ContainsFunc(names, func(name string) bool {
			return info.Name() == name
		})
	}
}

// SuffixesFilter filters any path that has a suffix matching any of the provided suffixes.
func SuffixesFilter(suffixes ...string) FileWalkerFilter {
	return func(_ string, info os.DirEntry) bool {
		return slices.ContainsFunc(suffixes, func(suffix string) bool {
			return strings.HasSuffix(info.Name(), suffix)
		})
	}
}

// NegateFilter negates the result of the provided filter, so that e.g.
// NegateFilter(DirectoryFilter) will return true for files and false for directories.
func NegateFilter(filter FileWalkerFilter) FileWalkerFilter {
	return func(path string, info os.DirEntry) bool {
		return !filter(path, info)
	}
}

// DefaultSkipDirectories is a default skip function that skips common directories
// that are not relevant for most file walks, such as .git, .idea, and node_modules.
func DefaultSkipDirectories(_ string, info files.DirEntry) bool {
	name := info.Name()

	return info.IsDir() && (name == ".git" || name == ".idea" || name == "node_modules")
}

// NewFileWalker creates a new FileWalker with the specified root directory.
func NewFileWalker(root string) *FileWalker {
	return &FileWalker{
		root: root,
	}
}

// WithFilters adds filters to the FileWalker.
func (fw *FileWalker) WithFilters(filters ...FileWalkerFilter) *FileWalker {
	fw.filters = filters

	return fw
}

// WithSkipFunc sets the skip function for the FileWalker. This is a special
// filter that allows skipping traversal of entire directory trees (like
// .git, node_modules, etc.).
func (fw *FileWalker) WithSkipFunc(skipFunc FileWalkerFilter) *FileWalker {
	fw.skipFunc = skipFunc

	return fw
}

// WithStatBeforeWalk sets whether to stat the root directory before walking.
// The underlying filepath.WalkDir implementation panics on non-existent paths,
// so this is useful when the input isn't guaranteed to exist.
func (fw *FileWalker) WithStatBeforeWalk(statRoot bool) *FileWalker {
	fw.statRoot = statRoot

	return fw
}

// Walk walks the file system rooted at the root, calling f for each file
// not filtered out by the filters.
func (fw *FileWalker) Walk(f func(string) error) error {
	if err := fw.validate(); err != nil {
		return err
	}

	return filepath.WalkDir(fw.root, fw.walker(f))
}

// WalkFS walks the file system rooted at the root using the provided files.FS,
// calling f for each file not filtered out by the filters.
func (fw *FileWalker) WalkFS(fs files.FS, f func(string) error) error {
	if err := fw.validate(); err != nil {
		return err
	}

	return files.WalkDir(fs, fw.root, fw.walker(f))
}

func (fw *FileWalker) validate() error {
	if fw.root == "" {
		return errors.New("root path is empty")
	}

	if fw.statRoot {
		if _, err := os.Stat(fw.root); err != nil {
			return err
		}
	}

	return nil
}

func (fw *FileWalker) walker(f func(string) error) func(path string, d files.DirEntry, err error) error {
	return func(path string, d files.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk directory %s: %w", path, err)
		}

		if fw.skipFunc != nil && fw.skipFunc(path, d) {
			return files.SkipDir
		}

		for _, filter := range fw.filters {
			if filter(path, d) {
				return nil
			}
		}

		return f(path)
	}
}

// NewFileWalkReducer creates a new FileWalkerReducer with the specified root
// directory and initial value for the accumulator.
func NewFileWalkReducer[T any](root string, into T) *FileWalkerReducer[T] {
	return &FileWalkerReducer[T]{
		FileWalker: &FileWalker{root: root},
		initial:    into,
	}
}

// WithFilters adds filters to the FileWalkerReducer.
func (fwr *FileWalkerReducer[T]) WithFilters(filters ...FileWalkerFilter) *FileWalkerReducer[T] {
	fwr.filters = append(fwr.filters, filters...)

	return fwr
}

// WithSkipFunc sets the skip function for the FileWalkerReducer.
func (fwr *FileWalkerReducer[T]) WithSkipFunc(skipFunc FileWalkerFilter) *FileWalkerReducer[T] {
	fwr.skipFunc = skipFunc

	return fwr
}

// WithStatBeforeWalk sets whether to stat the root directory before walking.
// The underlying filepath.WalkDir implementation panics on non-existent paths,
// so this is useful when the input isn't guaranteed to exist.
func (fwr *FileWalkerReducer[T]) WithStatBeforeWalk(statRoot bool) *FileWalkerReducer[T] {
	fwr.statRoot = statRoot

	return fwr
}

// Reduce walks the file system rooted at fwr.Root, calling f for each file.
// The function f receives the path of the file and the current value of the
// accumulator. It returns the new value of the accumulator and an error if any.
func (fwr *FileWalkerReducer[T]) Reduce(f func(string, T) (T, error)) (T, error) {
	curr := fwr.initial

	err := fwr.Walk(func(path string) error {
		var err error

		curr, err = f(path, curr)

		return err
	})

	return curr, err
}

// PathAppendReducer is a simple reducer function that appends the
// given path to the current slice of strings.
func PathAppendReducer(path string, curr []string) ([]string, error) {
	return append(curr, path), nil
}

const PathSeparator = string(os.PathSeparator)

// LoadRegalBundleFS loads bundle embedded from policy and data directory.
func LoadRegalBundleFS(fs files.FS) (bundle.Bundle, error) {
	embedLoader, err := bundle.NewFSLoader(fs)
	if err != nil {
		return bundle.Bundle{}, fmt.Errorf("failed to load bundle from filesystem: %w", err)
	}

	//nolint:wrapcheck
	return bundle.NewCustomReader(embedLoader.WithFilter(ExcludeTestLegacyFilter())).
		WithCapabilities(Capabilities()).
		WithSkipBundleVerification(true).
		WithProcessAnnotations(true).
		WithBundleName("regal").
		Read()
}

// LoadRegalBundlePath loads bundle from path.
func LoadRegalBundlePath(path string) (bundle.Bundle, error) {
	//nolint:wrapcheck
	return bundle.NewCustomReader(bundle.NewDirectoryLoader(path).WithFilter(ExcludeTestLegacyFilter())).
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

func ExcludeTestLegacyFilter() filter.LoaderFilter {
	return func(_ string, info files.FileInfo, _ int) bool {
		return strings.HasSuffix(info.Name(), "_test.rego") &&
			// (anderseknert): This is an outlier, but not sure we need something
			// more polished to deal with this for the time being.
			info.Name() != "todo_test.rego"
	}
}

func ExcludeTestsFilter(_ string, info os.DirEntry) bool {
	return strings.HasSuffix(info.Name(), "_test.rego") && info.Name() != "todo_test.rego"
}

func IsDir(path string) bool {
	info, err := os.Stat(path)

	return err == nil && info.IsDir()
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
	components := strings.Split(filepath.Dir(relative), PathSeparator)
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
	components := strings.Split(filepath.Dir(relative), PathSeparator)
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

// WalkFiles walks the file system rooted at root, calling f for each file. This is
// a less ceremonious version of filepath.WalkDir where only file paths (not dirs)
// are passed to the callback, and where directories that should commonly  be ignored
// (.git, node_modules, etc.) are skipped.
func WalkFiles(root string, f func(path string) error) error {
	return NewFileWalker(root).
		WithFilters(DirectoryFilter).
		WithSkipFunc(DefaultSkipDirectories).
		Walk(f)
}

// FindManifestLocations walks the file system rooted at root, and returns the
// *relative* paths of directories containing a .manifest file.
func FindManifestLocations(root string) ([]string, error) {
	var foundBundleRoots []string

	return NewFileWalkReducer(root, foundBundleRoots).
		WithSkipFunc(DefaultSkipDirectories).
		WithFilters(DirectoryFilter, NegateFilter(FileNameFilter(".manifest"))).
		Reduce(func(path string, curr []string) ([]string, error) {
			rel, err := filepath.Rel(root, path)
			if err == nil {
				curr = append(curr, filepath.Dir(rel))
			}

			return curr, err
		})
}

func ModulesFromCustomRuleFS(customRuleFS files.FS, rootPath string) (map[string]*ast.Module, error) {
	modules := make(map[string]*ast.Module)

	err := NewFileWalker(rootPath).
		WithFilters(DirectoryFilter, ExcludeTestsFilter).
		WalkFS(customRuleFS, func(path string) error {
			bs, err := FSReadAll(customRuleFS, path)
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

func FSReadAll(fs files.FS, path string) ([]byte, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", path, err)
	}
	defer f.Close()

	bs, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", path, err)
	}

	return bs, nil
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
