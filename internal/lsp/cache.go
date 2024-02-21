package lsp

import (
	"fmt"
	"os"
	"sync"

	"github.com/open-policy-agent/opa/ast"
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
	diagnosticsFile   map[string][]Diagnostic
	diagnosticsFileMu sync.Mutex

	// diagnosticsAggregate is a map of file URI to aggregate diagnostics for that file
	diagnosticsAggregate   map[string][]Diagnostic
	diagnosticsAggregateMu sync.Mutex

	// diagnosticsParseErrors is a map of file URI to parse errors for that file
	diagnosticsParseErrors map[string][]Diagnostic
	diagnosticsParseMu     sync.Mutex
}

func NewCache() *Cache {
	return &Cache{
		fileContents: make(map[string]string),
		modules:      make(map[string]*ast.Module),

		diagnosticsFile:        make(map[string][]Diagnostic),
		diagnosticsAggregate:   make(map[string][]Diagnostic),
		diagnosticsParseErrors: make(map[string][]Diagnostic),
	}
}

func (c *Cache) GetAllDiagnosticsForURI(uri string) []Diagnostic {
	parseDiags, ok := c.GetParseErrors(uri)
	if ok && len(parseDiags) > 0 {
		return parseDiags
	}

	allDiags := make([]Diagnostic, 0)

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

	return c.fileContents
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

	return c.modules
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

func (c *Cache) GetFileDiagnostics(uri string) ([]Diagnostic, bool) {
	c.diagnosticsFileMu.Lock()
	defer c.diagnosticsFileMu.Unlock()

	val, ok := c.diagnosticsFile[uri]

	return val, ok
}

func (c *Cache) SetFileDiagnostics(uri string, diags []Diagnostic) {
	c.diagnosticsFileMu.Lock()
	defer c.diagnosticsFileMu.Unlock()

	c.diagnosticsFile[uri] = diags
}

func (c *Cache) ClearFileDiagnostics() {
	c.diagnosticsFileMu.Lock()
	defer c.diagnosticsFileMu.Unlock()

	c.diagnosticsFile = make(map[string][]Diagnostic)
}

func (c *Cache) GetAggregateDiagnostics(uri string) ([]Diagnostic, bool) {
	c.diagnosticsAggregateMu.Lock()
	defer c.diagnosticsAggregateMu.Unlock()

	val, ok := c.diagnosticsAggregate[uri]

	return val, ok
}

func (c *Cache) SetAggregateDiagnostics(uri string, diags []Diagnostic) {
	c.diagnosticsAggregateMu.Lock()
	defer c.diagnosticsAggregateMu.Unlock()

	c.diagnosticsAggregate[uri] = diags
}

func (c *Cache) ClearAggregateDiagnostics() {
	c.diagnosticsAggregateMu.Lock()
	defer c.diagnosticsAggregateMu.Unlock()

	c.diagnosticsAggregate = make(map[string][]Diagnostic)
}

func (c *Cache) GetParseErrors(uri string) ([]Diagnostic, bool) {
	c.diagnosticsParseMu.Lock()
	defer c.diagnosticsParseMu.Unlock()

	val, ok := c.diagnosticsParseErrors[uri]

	return val, ok
}

func (c *Cache) SetParseErrors(uri string, diags []Diagnostic) {
	c.diagnosticsParseMu.Lock()
	defer c.diagnosticsParseMu.Unlock()

	c.diagnosticsParseErrors[uri] = diags
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
}

func updateCacheForURIFromDisk(cache *Cache, uri, path string) (string, error) {
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
