package cache

import (
	"fmt"
	"maps"
	"os"
	"sync"

	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/internal/lsp/types"
)

// Cache is used to store: current file contents (which includes unsaved changes), the latest parsed modules, and
// diagnostics for each file (including diagnostics gathered from linting files alongside other files).
type Cache struct {
	// fileContents is a map of file URI to raw file contents received from the client
	fileContents   map[string]string
	fileContentsMu sync.Mutex

	// modules is a map of file URI to parsed AST modules from the latest file contents value
	modules  map[string]*ast.Module
	moduleMu sync.Mutex

	// diagnosticsFile is a map of file URI to diagnostics for that file
	diagnosticsFile   map[string][]types.Diagnostic
	diagnosticsFileMu sync.Mutex

	// diagnosticsAggregate is a map of file URI to aggregate diagnostics for that file
	diagnosticsAggregate   map[string][]types.Diagnostic
	diagnosticsAggregateMu sync.Mutex

	// diagnosticsParseErrors is a map of file URI to parse errors for that file
	diagnosticsParseErrors map[string][]types.Diagnostic
	diagnosticsParseMu     sync.Mutex

	// builtinPositionsFile is a map of file URI to builtin positions for that file
	builtinPositionsFile map[string]map[uint][]types.BuiltinPosition
	builtinPositionsMu   sync.Mutex

	// fileRefs is a map of file URI to refs that are defined in that file. These are
	// intended to be used for completions in other files.
	// fileRefs is expected to be updated when a file is successfully parsed.
	fileRefs  map[string]map[string]types.Ref
	fileRefMu sync.Mutex

	// usedRefs is a map of file URI to a list of string ref names used in that file.
	// These are intended to be used for completions in that file.
	usedRefs   map[string][]string
	usedRefsMu sync.Mutex
}

func NewCache() *Cache {
	return &Cache{
		fileContents: make(map[string]string),
		modules:      make(map[string]*ast.Module),

		diagnosticsFile:        make(map[string][]types.Diagnostic),
		diagnosticsAggregate:   make(map[string][]types.Diagnostic),
		diagnosticsParseErrors: make(map[string][]types.Diagnostic),

		builtinPositionsFile: make(map[string]map[uint][]types.BuiltinPosition),

		fileRefs: make(map[string]map[string]types.Ref),

		usedRefs: make(map[string][]string),
	}
}

func (c *Cache) GetAllDiagnosticsForURI(uri string) []types.Diagnostic {
	parseDiags, ok := c.GetParseErrors(uri)
	if ok && len(parseDiags) > 0 {
		return parseDiags
	}

	allDiags := make([]types.Diagnostic, 0)

	aggDiags, ok := c.GetAggregateDiagnostics(uri)
	if ok {
		allDiags = append(allDiags, aggDiags...)
	}

	fileDiags, ok := c.GetFileDiagnostics(uri)
	if ok {
		allDiags = append(allDiags, fileDiags...)
	}

	return allDiags
}

func (c *Cache) GetAllFiles() map[string]string {
	c.fileContentsMu.Lock()
	defer c.fileContentsMu.Unlock()

	return maps.Clone(c.fileContents)
}

func (c *Cache) GetFileContents(uri string) (string, bool) {
	c.fileContentsMu.Lock()
	defer c.fileContentsMu.Unlock()

	val, ok := c.fileContents[uri]

	return val, ok
}

func (c *Cache) SetFileContents(uri string, content string) {
	c.fileContentsMu.Lock()
	defer c.fileContentsMu.Unlock()

	c.fileContents[uri] = content
}

func (c *Cache) GetAllModules() map[string]*ast.Module {
	c.moduleMu.Lock()
	defer c.moduleMu.Unlock()

	return maps.Clone(c.modules)
}

func (c *Cache) GetModule(uri string) (*ast.Module, bool) {
	c.moduleMu.Lock()
	defer c.moduleMu.Unlock()

	val, ok := c.modules[uri]

	return val, ok
}

func (c *Cache) SetModule(uri string, module *ast.Module) {
	c.moduleMu.Lock()
	defer c.moduleMu.Unlock()

	c.modules[uri] = module
}

func (c *Cache) GetFileDiagnostics(uri string) ([]types.Diagnostic, bool) {
	c.diagnosticsFileMu.Lock()
	defer c.diagnosticsFileMu.Unlock()

	val, ok := c.diagnosticsFile[uri]

	return val, ok
}

func (c *Cache) SetFileDiagnostics(uri string, diags []types.Diagnostic) {
	c.diagnosticsFileMu.Lock()
	defer c.diagnosticsFileMu.Unlock()

	c.diagnosticsFile[uri] = diags
}

func (c *Cache) ClearFileDiagnostics() {
	c.diagnosticsFileMu.Lock()
	defer c.diagnosticsFileMu.Unlock()

	c.diagnosticsFile = make(map[string][]types.Diagnostic)
}

func (c *Cache) GetAggregateDiagnostics(uri string) ([]types.Diagnostic, bool) {
	c.diagnosticsAggregateMu.Lock()
	defer c.diagnosticsAggregateMu.Unlock()

	val, ok := c.diagnosticsAggregate[uri]

	return val, ok
}

func (c *Cache) SetAggregateDiagnostics(uri string, diags []types.Diagnostic) {
	c.diagnosticsAggregateMu.Lock()
	defer c.diagnosticsAggregateMu.Unlock()

	c.diagnosticsAggregate[uri] = diags
}

func (c *Cache) ClearAggregateDiagnostics() {
	c.diagnosticsAggregateMu.Lock()
	defer c.diagnosticsAggregateMu.Unlock()

	c.diagnosticsAggregate = make(map[string][]types.Diagnostic)
}

func (c *Cache) GetParseErrors(uri string) ([]types.Diagnostic, bool) {
	c.diagnosticsParseMu.Lock()
	defer c.diagnosticsParseMu.Unlock()

	val, ok := c.diagnosticsParseErrors[uri]

	return val, ok
}

func (c *Cache) SetParseErrors(uri string, diags []types.Diagnostic) {
	c.diagnosticsParseMu.Lock()
	defer c.diagnosticsParseMu.Unlock()

	c.diagnosticsParseErrors[uri] = diags
}

func (c *Cache) GetBuiltinPositions(uri string) (map[uint][]types.BuiltinPosition, bool) {
	c.builtinPositionsMu.Lock()
	defer c.builtinPositionsMu.Unlock()

	val, ok := c.builtinPositionsFile[uri]

	return val, ok
}

func (c *Cache) SetBuiltinPositions(uri string, positions map[uint][]types.BuiltinPosition) {
	c.builtinPositionsMu.Lock()
	defer c.builtinPositionsMu.Unlock()

	c.builtinPositionsFile[uri] = positions
}

func (c *Cache) GetAllBuiltInPositions() map[string]map[uint][]types.BuiltinPosition {
	c.builtinPositionsMu.Lock()
	defer c.builtinPositionsMu.Unlock()

	return maps.Clone(c.builtinPositionsFile)
}

func (c *Cache) SetFileRefs(uri string, items map[string]types.Ref) {
	c.fileRefMu.Lock()
	defer c.fileRefMu.Unlock()

	c.fileRefs[uri] = items
}

func (c *Cache) GetFileRefs(uri string) map[string]types.Ref {
	c.fileRefMu.Lock()
	defer c.fileRefMu.Unlock()

	return c.fileRefs[uri]
}

func (c *Cache) GetAllFileRefs() map[string]map[string]types.Ref {
	c.fileRefMu.Lock()
	defer c.fileRefMu.Unlock()

	return maps.Clone(c.fileRefs)
}

func (c *Cache) SetUsedRefs(uri string, items []string) {
	c.usedRefsMu.Lock()
	defer c.usedRefsMu.Unlock()

	c.usedRefs[uri] = items
}

func (c *Cache) GetUsedRefs(uri string) ([]string, bool) {
	c.usedRefsMu.Lock()
	defer c.usedRefsMu.Unlock()

	refs, ok := c.usedRefs[uri]

	return refs, ok
}

// Delete removes all cached data for a given URI.
func (c *Cache) Delete(uri string) {
	c.fileContentsMu.Lock()
	delete(c.fileContents, uri)
	c.fileContentsMu.Unlock()

	c.moduleMu.Lock()
	delete(c.modules, uri)
	c.moduleMu.Unlock()

	c.diagnosticsFileMu.Lock()
	delete(c.diagnosticsFile, uri)
	c.diagnosticsFileMu.Unlock()

	c.diagnosticsAggregateMu.Lock()
	delete(c.diagnosticsAggregate, uri)
	c.diagnosticsAggregateMu.Unlock()

	c.diagnosticsParseMu.Lock()
	delete(c.diagnosticsParseErrors, uri)
	c.diagnosticsParseMu.Unlock()

	c.builtinPositionsMu.Lock()
	delete(c.builtinPositionsFile, uri)
	c.builtinPositionsMu.Unlock()

	c.fileRefMu.Lock()
	delete(c.fileRefs, uri)
	c.fileRefMu.Unlock()

	c.usedRefsMu.Lock()
	delete(c.usedRefs, uri)
	c.usedRefsMu.Unlock()
}

func UpdateCacheForURIFromDisk(cache *Cache, uri, path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	currentContent := string(content)

	cachedContent, ok := cache.GetFileContents(uri)
	if ok && cachedContent == currentContent {
		return cachedContent, nil
	}

	cache.SetFileContents(uri, currentContent)

	return currentContent, nil
}
