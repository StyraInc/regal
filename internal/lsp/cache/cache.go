package cache

import (
	"fmt"
	"maps"
	"os"
	"sync"

	"github.com/anderseknert/roast/pkg/util"

	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/pkg/report"
)

// Cache is used to store: current file contents (which includes unsaved changes), the latest parsed modules, and
// diagnostics for each file (including diagnostics gathered from linting files alongside other files).
type Cache struct {
	// fileContents is a map of file URI to raw file contents received from the client
	fileContents map[string]string

	// ignoredFileContents is a similar map of file URI to raw file contents
	// but it's not queried for project level operations like goto definition,
	// linting etc.
	// ignoredFileContents is also cleared on the delete operation.
	ignoredFileContents map[string]string

	// modules is a map of file URI to parsed AST modules from the latest file contents value
	modules map[string]*ast.Module

	// aggregateData stores the aggregate data from evaluations for each file.
	// This is used to cache the results of expensive evaluations and can be used
	// to update aggregate diagostics incrementally.
	aggregateData   map[string][]report.Aggregate
	aggregateDataMu sync.Mutex

	// diagnosticsFile is a map of file URI to diagnostics for that file
	diagnosticsFile map[string][]types.Diagnostic

	// diagnosticsParseErrors is a map of file URI to parse errors for that file
	diagnosticsParseErrors map[string][]types.Diagnostic

	// builtinPositionsFile is a map of file URI to builtin positions for that file
	builtinPositionsFile map[string]map[uint][]types.BuiltinPosition

	// keywordLocationsFile is a map of file URI to Rego keyword locations for that file
	// to be used for hover hints.
	keywordLocationsFile map[string]map[uint][]types.KeywordLocation

	// fileRefs is a map of file URI to refs that are defined in that file. These are
	// intended to be used for completions in other files.
	// fileRefs is expected to be updated when a file is successfully parsed.
	fileRefs       map[string]map[string]types.Ref
	fileContentsMu sync.Mutex

	ignoredFileContentsMu sync.Mutex

	moduleMu sync.Mutex

	diagnosticsFileMu sync.Mutex

	diagnosticsParseMu sync.Mutex

	builtinPositionsMu sync.Mutex

	keywordLocationsMu sync.Mutex

	fileRefMu sync.Mutex
}

func NewCache() *Cache {
	return &Cache{
		fileContents:        make(map[string]string),
		ignoredFileContents: make(map[string]string),

		modules: make(map[string]*ast.Module),

		aggregateData: make(map[string][]report.Aggregate),

		diagnosticsFile:        make(map[string][]types.Diagnostic),
		diagnosticsParseErrors: make(map[string][]types.Diagnostic),

		builtinPositionsFile: make(map[string]map[uint][]types.BuiltinPosition),
		keywordLocationsFile: make(map[string]map[uint][]types.KeywordLocation),

		fileRefs: make(map[string]map[string]types.Ref),
	}
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

// SetFileAggregates will only set aggregate data for the provided URI. Even if
// data for other files is provided, only the specified URI is updated.
func (c *Cache) SetFileAggregates(fileURI string, data map[string][]report.Aggregate) {
	c.aggregateDataMu.Lock()
	defer c.aggregateDataMu.Unlock()

	flattenedAggregates := make([]report.Aggregate, 0)

	for _, aggregates := range data {
		for _, aggregate := range aggregates {
			if aggregate.SourceFile() != fileURI {
				continue
			}

			flattenedAggregates = append(flattenedAggregates, aggregate)
		}
	}

	c.aggregateData[fileURI] = flattenedAggregates
}

func (c *Cache) SetAggregates(data map[string][]report.Aggregate) {
	c.aggregateDataMu.Lock()
	defer c.aggregateDataMu.Unlock()

	// clear the state
	c.aggregateData = make(map[string][]report.Aggregate)

	for _, aggregates := range data {
		for _, aggregate := range aggregates {
			c.aggregateData[aggregate.SourceFile()] = append(c.aggregateData[aggregate.SourceFile()], aggregate)
		}
	}
}

// GetFileAggregates is used to get aggregate data for a given list of files.
// This is only used in tests to validate the cache state.
func (c *Cache) GetFileAggregates(fileURIs ...string) map[string][]report.Aggregate {
	c.aggregateDataMu.Lock()
	defer c.aggregateDataMu.Unlock()

	includedFiles := make(map[string]struct{}, len(fileURIs))
	for _, fileURI := range fileURIs {
		includedFiles[fileURI] = struct{}{}
	}

	getAll := len(fileURIs) == 0

	allAggregates := make(map[string][]report.Aggregate)

	for sourceFile, aggregates := range c.aggregateData {
		if _, included := includedFiles[sourceFile]; !included && !getAll {
			continue
		}

		for _, aggregate := range aggregates {
			allAggregates[aggregate.IndexKey()] = append(allAggregates[aggregate.IndexKey()], aggregate)
		}
	}

	return allAggregates
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

// SetFileDiagnosticsForRules will perform a partial update of the diagnostics
// for a file given a list of evaluated rules.
func (c *Cache) SetFileDiagnosticsForRules(fileURI string, rules []string, diags []types.Diagnostic) {
	c.diagnosticsFileMu.Lock()
	defer c.diagnosticsFileMu.Unlock()

	ruleKeys := make(map[string]struct{}, len(rules))
	for _, rule := range rules {
		ruleKeys[rule] = struct{}{}
	}

	preservedDiagnostics := make([]types.Diagnostic, 0)

	for _, diag := range c.diagnosticsFile[fileURI] {
		if _, ok := ruleKeys[diag.Code]; !ok {
			preservedDiagnostics = append(preservedDiagnostics, diag)
		}
	}

	c.diagnosticsFile[fileURI] = append(preservedDiagnostics, diags...)
}

func (c *Cache) ClearFileDiagnostics() {
	c.diagnosticsFileMu.Lock()
	defer c.diagnosticsFileMu.Unlock()

	c.diagnosticsFile = make(map[string][]types.Diagnostic)
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

func (c *Cache) SetKeywordLocations(fileURI string, keywords map[uint][]types.KeywordLocation) {
	c.keywordLocationsMu.Lock()
	defer c.keywordLocationsMu.Unlock()

	c.keywordLocationsFile[fileURI] = keywords
}

func (c *Cache) GetKeywordLocations(fileURI string) (map[uint][]types.KeywordLocation, bool) {
	c.keywordLocationsMu.Lock()
	defer c.keywordLocationsMu.Unlock()

	val, ok := c.keywordLocationsFile[fileURI]

	return val, ok
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

	c.aggregateDataMu.Lock()
	delete(c.aggregateData, fileURI)
	c.aggregateDataMu.Unlock()

	c.diagnosticsFileMu.Lock()
	delete(c.diagnosticsFile, fileURI)
	c.diagnosticsFileMu.Unlock()

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

func UpdateCacheForURIFromDisk(cache *Cache, fileURI, path string) (bool, string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, "", fmt.Errorf("failed to read file: %w", err)
	}

	currentContent := util.ByteSliceToString(content)

	cachedContent, ok := cache.GetFileContents(fileURI)
	if ok && cachedContent == currentContent {
		return false, cachedContent, nil
	}

	cache.SetFileContents(fileURI, currentContent)

	return true, currentContent, nil
}
