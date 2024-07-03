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

	// ignoredFileContents is a similar map of file URI to raw file contents
	// but it's not queried for project level operations like goto definition,
	// linting etc.
	// ignoredFileContents is also cleared on the delete operation.
	ignoredFileContents   map[string]string
	ignoredFileContentsMu sync.Mutex

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
}

func NewCache() *Cache {
	return &Cache{
		fileContents:        make(map[string]string),
		ignoredFileContents: make(map[string]string),

		modules: make(map[string]*ast.Module),

		diagnosticsFile:        make(map[string][]types.Diagnostic),
		diagnosticsAggregate:   make(map[string][]types.Diagnostic),
		diagnosticsParseErrors: make(map[string][]types.Diagnostic),

		builtinPositionsFile: make(map[string]map[uint][]types.BuiltinPosition),

		fileRefs: make(map[string]map[string]types.Ref),
	}
}

func (c *Cache) GetAllDiagnosticsForURI(fileURI string) []types.Diagnostic {
	parseDiags, ok := c.GetParseErrors(fileURI)
	if ok && len(parseDiags) > 0 {
		return parseDiags
	}

	allDiags := make([]types.Diagnostic, 0)

	aggDiags, ok := c.GetAggregateDiagnostics(fileURI)
	if ok {
		allDiags = append(allDiags, aggDiags...)
	}

	fileDiags, ok := c.GetFileDiagnostics(fileURI)
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

func (c *Cache) GetFileContents(fileURI string) (string, bool) {
	c.fileContentsMu.Lock()
	defer c.fileContentsMu.Unlock()

	val, ok := c.fileContents[fileURI]

	return val, ok
}

func (c *Cache) SetFileContents(fileURI string, content string) {
	c.fileContentsMu.Lock()
	defer c.fileContentsMu.Unlock()

	c.fileContents[fileURI] = content
}

func (c *Cache) GetIgnoredFileContents(fileURI string) (string, bool) {
	c.ignoredFileContentsMu.Lock()
	defer c.ignoredFileContentsMu.Unlock()

	val, ok := c.ignoredFileContents[fileURI]

	return val, ok
}

func (c *Cache) SetIgnoredFileContents(fileURI string, content string) {
	c.ignoredFileContentsMu.Lock()
	defer c.ignoredFileContentsMu.Unlock()

	c.ignoredFileContents[fileURI] = content
}

func (c *Cache) GetAllIgnoredFiles() map[string]string {
	c.ignoredFileContentsMu.Lock()
	defer c.ignoredFileContentsMu.Unlock()

	return maps.Clone(c.ignoredFileContents)
}

func (c *Cache) ClearIgnoredFileContents(fileURI string) {
	c.ignoredFileContentsMu.Lock()
	defer c.ignoredFileContentsMu.Unlock()

	delete(c.ignoredFileContents, fileURI)
}

func (c *Cache) GetAllModules() map[string]*ast.Module {
	c.moduleMu.Lock()
	defer c.moduleMu.Unlock()

	return maps.Clone(c.modules)
}

func (c *Cache) GetModule(fileURI string) (*ast.Module, bool) {
	c.moduleMu.Lock()
	defer c.moduleMu.Unlock()

	val, ok := c.modules[fileURI]

	return val, ok
}

func (c *Cache) SetModule(fileURI string, module *ast.Module) {
	c.moduleMu.Lock()
	defer c.moduleMu.Unlock()

	c.modules[fileURI] = module
}

func (c *Cache) GetFileDiagnostics(uri string) ([]types.Diagnostic, bool) {
	c.diagnosticsFileMu.Lock()
	defer c.diagnosticsFileMu.Unlock()

	val, ok := c.diagnosticsFile[uri]

	return val, ok
}

func (c *Cache) SetFileDiagnostics(fileURI string, diags []types.Diagnostic) {
	c.diagnosticsFileMu.Lock()
	defer c.diagnosticsFileMu.Unlock()

	c.diagnosticsFile[fileURI] = diags
}

func (c *Cache) ClearFileDiagnostics() {
	c.diagnosticsFileMu.Lock()
	defer c.diagnosticsFileMu.Unlock()

	c.diagnosticsFile = make(map[string][]types.Diagnostic)
}

func (c *Cache) GetAggregateDiagnostics(fileURI string) ([]types.Diagnostic, bool) {
	c.diagnosticsAggregateMu.Lock()
	defer c.diagnosticsAggregateMu.Unlock()

	val, ok := c.diagnosticsAggregate[fileURI]

	return val, ok
}

func (c *Cache) SetAggregateDiagnostics(fileURI string, diags []types.Diagnostic) {
	c.diagnosticsAggregateMu.Lock()
	defer c.diagnosticsAggregateMu.Unlock()

	c.diagnosticsAggregate[fileURI] = diags
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

func (c *Cache) SetParseErrors(fileURI string, diags []types.Diagnostic) {
	c.diagnosticsParseMu.Lock()
	defer c.diagnosticsParseMu.Unlock()

	c.diagnosticsParseErrors[fileURI] = diags
}

func (c *Cache) GetBuiltinPositions(fileURI string) (map[uint][]types.BuiltinPosition, bool) {
	c.builtinPositionsMu.Lock()
	defer c.builtinPositionsMu.Unlock()

	val, ok := c.builtinPositionsFile[fileURI]

	return val, ok
}

func (c *Cache) SetBuiltinPositions(fileURI string, positions map[uint][]types.BuiltinPosition) {
	c.builtinPositionsMu.Lock()
	defer c.builtinPositionsMu.Unlock()

	c.builtinPositionsFile[fileURI] = positions
}

func (c *Cache) GetAllBuiltInPositions() map[string]map[uint][]types.BuiltinPosition {
	c.builtinPositionsMu.Lock()
	defer c.builtinPositionsMu.Unlock()

	return maps.Clone(c.builtinPositionsFile)
}

func (c *Cache) SetFileRefs(fileURI string, items map[string]types.Ref) {
	c.fileRefMu.Lock()
	defer c.fileRefMu.Unlock()

	c.fileRefs[fileURI] = items
}

func (c *Cache) GetFileRefs(fileURI string) map[string]types.Ref {
	c.fileRefMu.Lock()
	defer c.fileRefMu.Unlock()

	return c.fileRefs[fileURI]
}

func (c *Cache) GetAllFileRefs() map[string]map[string]types.Ref {
	c.fileRefMu.Lock()
	defer c.fileRefMu.Unlock()

	return maps.Clone(c.fileRefs)
}

// Delete removes all cached data for a given URI. Ignored file contents are
// also removed if found for a matching URI.
func (c *Cache) Delete(fileURI string) {
	c.fileContentsMu.Lock()
	delete(c.fileContents, fileURI)
	c.fileContentsMu.Unlock()

	c.moduleMu.Lock()
	delete(c.modules, fileURI)
	c.moduleMu.Unlock()

	c.diagnosticsFileMu.Lock()
	delete(c.diagnosticsFile, fileURI)
	c.diagnosticsFileMu.Unlock()

	c.diagnosticsAggregateMu.Lock()
	delete(c.diagnosticsAggregate, fileURI)
	c.diagnosticsAggregateMu.Unlock()

	c.diagnosticsParseMu.Lock()
	delete(c.diagnosticsParseErrors, fileURI)
	c.diagnosticsParseMu.Unlock()

	c.builtinPositionsMu.Lock()
	delete(c.builtinPositionsFile, fileURI)
	c.builtinPositionsMu.Unlock()

	c.fileRefMu.Lock()
	delete(c.fileRefs, fileURI)
	c.fileRefMu.Unlock()

	c.ignoredFileContentsMu.Lock()
	delete(c.ignoredFileContents, fileURI)
	c.ignoredFileContentsMu.Unlock()
}

func UpdateCacheForURIFromDisk(cache *Cache, fileURI, path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	currentContent := string(content)

	cachedContent, ok := cache.GetFileContents(fileURI)
	if ok && cachedContent == currentContent {
		return cachedContent, nil
	}

	cache.SetFileContents(fileURI, currentContent)

	return currentContent, nil
}
