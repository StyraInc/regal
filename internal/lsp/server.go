//nolint:nilerr,nilnil
package lsp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anderseknert/roast/pkg/encoding"
	"github.com/sourcegraph/jsonrpc2"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"
	"github.com/open-policy-agent/opa/storage"

	rbundle "github.com/styrainc/regal/bundle"
	"github.com/styrainc/regal/internal/capabilities"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/lsp/bundles"
	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/commands"
	"github.com/styrainc/regal/internal/lsp/completions"
	"github.com/styrainc/regal/internal/lsp/completions/providers"
	lsconfig "github.com/styrainc/regal/internal/lsp/config"
	"github.com/styrainc/regal/internal/lsp/examples"
	"github.com/styrainc/regal/internal/lsp/hover"
	"github.com/styrainc/regal/internal/lsp/opa/oracle"
	"github.com/styrainc/regal/internal/lsp/rego"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/uri"
	rparse "github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/update"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/internal/web"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/fixer"
	"github.com/styrainc/regal/pkg/fixer/fileprovider"
	"github.com/styrainc/regal/pkg/fixer/fixes"
	"github.com/styrainc/regal/pkg/linter"
	"github.com/styrainc/regal/pkg/version"
)

const (
	methodTextDocumentPublishDiagnostics = "textDocument/publishDiagnostics"
	methodWorkspaceApplyEdit             = "workspace/applyEdit"

	ruleNameOPAFmt                   = "opa-fmt"
	ruleNameUseRegoV1                = "use-rego-v1"
	ruleNameUseAssignmentOperator    = "use-assignment-operator"
	ruleNameNoWhitespaceComment      = "no-whitespace-comment"
	ruleNameDirectoryPackageMismatch = "directory-package-mismatch"
)

type LanguageServerOptions struct {
	ErrorLog io.Writer
}

func NewLanguageServer(ctx context.Context, opts *LanguageServerOptions) *LanguageServer {
	c := cache.NewCache()
	store := NewRegalStore()

	ls := &LanguageServer{
		cache:                c,
		regoStore:            store,
		errorLog:             opts.ErrorLog,
		lintFileJobs:         make(chan lintFileJob, 10),
		lintWorkspaceJobs:    make(chan lintWorkspaceJob, 10),
		builtinsPositionJobs: make(chan lintFileJob, 10),
		commandRequest:       make(chan types.ExecuteCommandParams, 10),
		templateFile:         make(chan lintFileJob, 10),
		configWatcher:        lsconfig.NewWatcher(&lsconfig.WatcherOpts{ErrorWriter: opts.ErrorLog}),
		completionsManager:   completions.NewDefaultManager(ctx, c, store),
		webServer:            web.NewServer(c),
		loadedBuiltins:       make(map[string]map[string]*ast.Builtin),
	}

	return ls
}

type LanguageServer struct {
	errorLog io.Writer

	regoStore storage.Store
	conn      *jsonrpc2.Conn

	configWatcher  *lsconfig.Watcher
	loadedConfig   *config.Config
	loadedBuiltins map[string]map[string]*ast.Builtin

	clientInitializationOptions types.InitializationOptions

	cache       *cache.Cache
	bundleCache *bundles.Cache

	completionsManager *completions.Manager

	commandRequest       chan types.ExecuteCommandParams
	lintWorkspaceJobs    chan lintWorkspaceJob
	lintFileJobs         chan lintFileJob
	builtinsPositionJobs chan lintFileJob
	templateFile         chan lintFileJob

	webServer *web.Server

	workspaceRootURI string
	clientIdentifier clients.Identifier

	loadedBuiltinsLock sync.RWMutex

	loadedConfigLock sync.Mutex
}

// lintFileJob is sent to the lintFileJobs channel to trigger a
// diagnostic update for a file.
type lintFileJob struct {
	Reason string
	URI    string
}

// lintWorkspaceJob is sent to lintWorkspaceJobs when a full workspace
// diagnostic update is needed.
type lintWorkspaceJob struct {
	Reason string
	// OverwriteAggregates for a workspace is only run once at start up. All
	// later updates to aggregate state is made as files are changed.
	OverwriteAggregates bool
}

func (l *LanguageServer) Handle(
	ctx context.Context,
	conn *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	// null params are allowed, but only for certain methods
	if req.Params == nil && req.Method != "shutdown" && req.Method != "exit" {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	switch req.Method {
	case "initialize":
		return l.handleInitialize(ctx, conn, req)
	case "initialized":
		return l.handleInitialized(ctx, conn, req)
	case "textDocument/codeAction":
		return l.handleTextDocumentCodeAction(ctx, conn, req)
	case "textDocument/definition":
		return l.handleTextDocumentDefinition(ctx, conn, req)
	case "textDocument/diagnostic":
		return l.handleTextDocumentDiagnostic(ctx, conn, req)
	case "textDocument/didOpen":
		return l.handleTextDocumentDidOpen(ctx, conn, req)
	case "textDocument/didClose":
		return l.handleTextDocumentDidClose(ctx, conn, req)
	case "textDocument/didSave":
		return l.handleTextDocumentDidSave(ctx, conn, req)
	case "textDocument/documentSymbol":
		return l.handleTextDocumentDocumentSymbol(ctx, conn, req)
	case "textDocument/didChange":
		return l.handleTextDocumentDidChange(ctx, conn, req)
	case "textDocument/foldingRange":
		return l.handleTextDocumentFoldingRange(ctx, conn, req)
	case "textDocument/formatting":
		return l.handleTextDocumentFormatting(ctx, conn, req)
	case "textDocument/hover":
		return l.handleTextDocumentHover(ctx, conn, req)
	case "textDocument/inlayHint":
		return l.handleTextDocumentInlayHint(ctx, conn, req)
	case "textDocument/codeLens":
		return l.handleTextDocumentCodeLens(ctx, conn, req)
	case "textDocument/completion":
		return l.handleTextDocumentCompletion(ctx, conn, req)
	case "workspace/didChangeWatchedFiles":
		return l.handleWorkspaceDidChangeWatchedFiles(ctx, conn, req)
	case "workspace/diagnostic":
		return l.handleWorkspaceDiagnostic(ctx, conn, req)
	case "workspace/didRenameFiles":
		return l.handleWorkspaceDidRenameFiles(ctx, conn, req)
	case "workspace/didDeleteFiles":
		return l.handleWorkspaceDidDeleteFiles(ctx, conn, req)
	case "workspace/didCreateFiles":
		return l.handleWorkspaceDidCreateFiles(ctx, conn, req)
	case "workspace/executeCommand":
		return l.handleWorkspaceExecuteCommand(ctx, conn, req)
	case "workspace/symbol":
		return l.handleWorkspaceSymbol(ctx, conn, req)
	case "shutdown":
		// no-op as we wait for the exit signal before closing channel
		return struct{}{}, nil
	case "exit":
		// now we can close the channel, this will cause the program to exit and the
		// context for all workers to be cancelled
		if err := l.conn.Close(); err != nil {
			return nil, fmt.Errorf("failed to close connection: %w", err)
		}

		return struct{}{}, nil
	case "$/cancelRequest":
		// TODO: this is a no-op, but is something that we should implement
		// if we want to support longer running, client-triggered operations
		// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#dollarRequests
		return struct{}{}, nil
	}

	return nil, &jsonrpc2.Error{
		Code:    jsonrpc2.CodeMethodNotFound,
		Message: "method not supported: " + req.Method,
	}
}

func (l *LanguageServer) SetConn(conn *jsonrpc2.Conn) {
	l.conn = conn
}

func (l *LanguageServer) StartDiagnosticsWorker(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-l.lintFileJobs:
				fmt.Fprintln(l.errorLog, "lintFileJobs", filepath.Base(job.URI), job.Reason)
				bis := l.builtinsForCurrentCapabilities()

				// updateParse will not return an error when the parsing failed,
				// but only when it was impossible
				if _, err := updateParse(ctx, l.cache, l.regoStore, job.URI, bis); err != nil {
					l.logError(fmt.Errorf("failed to update module for %s: %w", job.URI, err))

					continue
				}

				// lint the file and send the diagnostics
				if err := updateFileDiagnostics(ctx, l.cache, l.getLoadedConfig(), job.URI, l.workspaceRootURI); err != nil {
					l.logError(fmt.Errorf("failed to update file diagnostics: %w", err))

					continue
				}

				if err := l.sendFileDiagnostics(ctx, job.URI); err != nil {
					l.logError(fmt.Errorf("failed to send diagnostic: %w", err))

					continue
				}

				l.lintWorkspaceJobs <- lintWorkspaceJob{
					Reason: fmt.Sprintf("file %s %s", job.URI, job.Reason),
					// this run is expected to used the cached aggregate state
					// for other files.
					// The aggregate state for this file will still be updated.
					OverwriteAggregates: false,
				}

				fmt.Fprintln(l.errorLog, "lintFileJobs", filepath.Base(job.URI), "done")
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case job := <-l.lintWorkspaceJobs:
			fmt.Fprintln(l.errorLog, "lintFileJobs", job.Reason, job.OverwriteAggregates)

			err := updateAllDiagnostics(
				ctx,
				l.cache,
				l.getLoadedConfig(),
				l.workspaceRootURI,
				// this is intended to only be set to true once at start up,
				// on following runs, cached aggregate data is used.
				job.OverwriteAggregates,
			)
			if err != nil {
				l.logError(fmt.Errorf("failed to update aggregate diagnostics (trigger): %w", err))
			}

			for fileURI := range l.cache.GetAllFiles() {
				fmt.Fprintln(l.errorLog, "post workspace send diag for file", filepath.Base(fileURI))

				if err := l.sendFileDiagnostics(ctx, fileURI); err != nil {
					l.logError(fmt.Errorf("failed to send diagnostic: %w", err))
				}
			}
		}
	}
}

func (l *LanguageServer) StartHoverWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-l.builtinsPositionJobs:
			fileURI := job.URI

			if l.ignoreURI(fileURI) {
				continue
			}

			if _, ok := l.cache.GetFileContents(fileURI); !ok {
				// If the file is not in the cache, exit early or else
				// we might accidentally put it in the cache after it's been
				// deleted: https://github.com/StyraInc/regal/issues/679
				continue
			}

			bis := l.builtinsForCurrentCapabilities()

			success, err := updateParse(ctx, l.cache, l.regoStore, fileURI, bis)
			if err != nil {
				l.logError(fmt.Errorf("failed to update parse: %w", err))

				continue
			}

			if !success {
				continue
			}

			if err = hover.UpdateBuiltinPositions(l.cache, fileURI, bis); err != nil {
				l.logError(fmt.Errorf("failed to update builtin positions: %w", err))

				continue
			}

			if err = hover.UpdateKeywordLocations(ctx, l.cache, fileURI); err != nil {
				l.logError(fmt.Errorf("failed to update keyword positions: %w", err))

				continue
			}
		}
	}
}

func (l *LanguageServer) getLoadedConfig() *config.Config {
	l.loadedConfigLock.Lock()
	defer l.loadedConfigLock.Unlock()

	return l.loadedConfig
}

func (l *LanguageServer) StartConfigWorker(ctx context.Context) {
	if err := l.configWatcher.Start(ctx); err != nil {
		l.logError(fmt.Errorf("failed to start config watcher: %w", err))

		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case path := <-l.configWatcher.Reload:
			configFileBs, err := os.ReadFile(path)
			if err != nil {
				l.logError(fmt.Errorf("failed to open config file: %w", err))

				continue
			}

			var userConfig config.Config

			if err = yaml.Unmarshal(configFileBs, &userConfig); err != nil && !errors.Is(err, io.EOF) {
				l.logError(fmt.Errorf("failed to reload config: %w", err))

				return
			}

			mergedConfig, err := config.LoadConfigWithDefaultsFromBundle(&rbundle.LoadedBundle, &userConfig)
			if err != nil {
				l.logError(fmt.Errorf("failed to load config: %w", err))

				return
			}

			// if the config is now blank, then we need to clear it
			l.loadedConfigLock.Lock()

			if errors.Is(err, io.EOF) {
				l.loadedConfig = nil
			} else {
				l.loadedConfig = &mergedConfig
			}

			l.loadedConfigLock.Unlock()

			// Capabilities URL may have changed, so we should reload it.
			cfg := l.getLoadedConfig()

			capsURL := "regal:///capabilities/default"
			if cfg != nil && cfg.CapabilitiesURL != "" {
				capsURL = cfg.CapabilitiesURL
			}

			caps, err := capabilities.Lookup(ctx, capsURL)
			if err != nil {
				l.logError(fmt.Errorf("failed to load capabilities for URL %q: %w", capsURL, err))

				return
			}

			bis := rego.BuiltinsForCapabilities(caps)

			l.loadedBuiltinsLock.Lock()
			l.loadedBuiltins[capsURL] = bis
			l.loadedBuiltinsLock.Unlock()

			// the config may now ignore files that existed in the cache before,
			// in which case we need to remove them to stop their contents being
			// used in other ls functions.
			for k := range l.cache.GetAllFiles() {
				if !l.ignoreURI(k) {
					continue
				}

				// move the contents to the ignored part of the cache
				contents, ok := l.cache.GetFileContents(k)
				if ok {
					l.cache.Delete(k)

					l.cache.SetIgnoredFileContents(k, contents)
				}

				if err := RemoveFileMod(ctx, l.regoStore, k); err != nil {
					l.logError(fmt.Errorf("failed to remove mod from store: %w", err))
				}
			}

			// when a file is 'unignored', we move its contents to the
			// standard file list if missing
			for k, v := range l.cache.GetAllIgnoredFiles() {
				if l.ignoreURI(k) {
					continue
				}

				// ignored contents will only be used when there is no existing content
				_, ok := l.cache.GetFileContents(k)
				if !ok {
					l.cache.SetFileContents(k, v)

					// updating the parse here will enable things like go-to definition
					// to start working right away without the need for a file content
					// update to run updateParse.
					if _, err = updateParse(ctx, l.cache, l.regoStore, k, bis); err != nil {
						l.logError(fmt.Errorf("failed to update parse for previously ignored file %q: %w", k, err))
					}
				}

				l.cache.ClearIgnoredFileContents(k)
			}

			//nolint:contextcheck
			go func() {
				if l.getLoadedConfig().Features.Remote.CheckVersion &&
					os.Getenv(update.CheckVersionDisableEnvVar) != "" {
					update.CheckAndWarn(update.Options{
						CurrentVersion: version.Version,
						CurrentTime:    time.Now().UTC(),
						Debug:          false,
						StateDir:       config.GlobalDir(),
					}, os.Stderr)
				}
			}()

			l.lintWorkspaceJobs <- lintWorkspaceJob{Reason: "config file changed"}
		case <-l.configWatcher.Drop:
			l.loadedConfigLock.Lock()
			l.loadedConfig = nil
			l.loadedConfigLock.Unlock()

			l.lintWorkspaceJobs <- lintWorkspaceJob{Reason: "config file dropped"}
		}
	}
}

var regalEvalUseAsInputComment = regexp.MustCompile(`^\s*regal eval:\s*use-as-input`)

func (l *LanguageServer) StartCommandWorker(ctx context.Context) { // nolint:maintidx
	// note, in this function conn.Call is used as the workspace/applyEdit message is a request, not a notification
	// as per the spec. In order to be 'routed' to the correct handler on the client it must have an ID
	// receive responses too.
	// Note however that the responses from the client are not needed by the server.
	for {
		select {
		case <-ctx.Done():
			return
		case params := <-l.commandRequest:
			var editParams *types.ApplyWorkspaceEditParams

			var err error

			var fixed bool

			switch params.Command {
			case "regal.fix.opa-fmt":
				fixed, editParams, err = l.fixEditParams(
					"Format using opa fmt",
					&fixes.Fmt{OPAFmtOpts: format.Opts{}},
					commands.ParseOptions{TargetArgIndex: 0},
					params,
				)
			case "regal.fix.use-rego-v1":
				fixed, editParams, err = l.fixEditParams(
					"Format for Rego v1 using opa-fmt",
					&fixes.Fmt{OPAFmtOpts: format.Opts{RegoVersion: ast.RegoV0CompatV1}},
					commands.ParseOptions{TargetArgIndex: 0},
					params,
				)
			case "regal.fix.use-assignment-operator":
				fixed, editParams, err = l.fixEditParams(
					"Replace = with := in assignment",
					&fixes.UseAssignmentOperator{},
					commands.ParseOptions{TargetArgIndex: 0, RowArgIndex: 1, ColArgIndex: 2},
					params,
				)
			case "regal.fix.no-whitespace-comment":
				fixed, editParams, err = l.fixEditParams(
					"Format comment to have leading whitespace",
					&fixes.NoWhitespaceComment{},
					commands.ParseOptions{TargetArgIndex: 0, RowArgIndex: 1, ColArgIndex: 2},
					params,
				)
			case "regal.fix.directory-package-mismatch":
				var renameParams types.ApplyWorkspaceRenameEditParams

				fileURL, ok := params.Arguments[0].(string)
				if !ok {
					l.logError(fmt.Errorf("expected first argument to be a string, got %T", params.Arguments[0]))

					break
				}

				renameParams, err = l.fixRenameParams(
					"Rename file to match package path",
					&fixes.DirectoryPackageMismatch{},
					fileURL,
				)
				if err != nil {
					l.logError(fmt.Errorf("failed to fix directory package mismatch: %w", err))

					break
				}

				if err := l.conn.Call(ctx, methodWorkspaceApplyEdit, renameParams, nil); err != nil {
					l.logError(fmt.Errorf("failed %s notify: %v", methodWorkspaceApplyEdit, err.Error()))
				}

				// clean up any empty edits dirs left over
				if len(renameParams.Edit.DocumentChanges) > 0 {
					dir := filepath.Dir(uri.ToPath(l.clientIdentifier, renameParams.Edit.DocumentChanges[0].OldURI))

					if err := util.DeleteEmptyDirs(dir); err != nil {
						l.logError(fmt.Errorf("failed to delete empty directories: %w", err))
					}
				}

				// handle this ourselves as it's a rename and not a content edit
				fixed = false
			case "regal.debug":
				if l.clientIdentifier != clients.IdentifierVSCode {
					l.logError(errors.New("regal.debug command is only supported in VSCode"))

					break
				}

				if len(params.Arguments) != 3 {
					l.logError(fmt.Errorf("expected three arguments, got %d", len(params.Arguments)))

					break
				}

				file, ok := params.Arguments[0].(string)
				if !ok {
					l.logError(fmt.Errorf("expected first argument to be a string, got %T", params.Arguments[0]))

					break
				}

				path, ok := params.Arguments[1].(string)
				if !ok {
					l.logError(fmt.Errorf("expected second argument to be a string, got %T", params.Arguments[1]))

					break
				}

				inputPath, _ := rio.FindInput(uri.ToPath(l.clientIdentifier, file), l.workspacePath())

				responseParams := map[string]any{
					"type":        "opa-debug",
					"name":        path,
					"request":     "launch",
					"command":     "eval",
					"query":       path,
					"enablePrint": true,
					"stopOnEntry": true,
					"inputPath":   inputPath,
				}

				responseResult := map[string]any{}

				if err = l.conn.Call(ctx, "regal/startDebugging", responseParams, &responseResult); err != nil {
					l.logError(fmt.Errorf("regal/startDebugging failed: %v", err.Error()))
				}
			case "regal.eval":
				if len(params.Arguments) != 3 {
					l.logError(fmt.Errorf("expected three arguments, got %d", len(params.Arguments)))

					break
				}

				file, ok := params.Arguments[0].(string)
				if !ok {
					l.logError(fmt.Errorf("expected first argument to be a string, got %T", params.Arguments[0]))

					break
				}

				path, ok := params.Arguments[1].(string)
				if !ok {
					l.logError(fmt.Errorf("expected second argument to be a string, got %T", params.Arguments[1]))

					break
				}

				line, ok := params.Arguments[2].(float64)
				if !ok {
					l.logError(fmt.Errorf("expected third argument to be a number, got %T", params.Arguments[2]))

					break
				}

				currentModule, ok := l.cache.GetModule(file)
				if !ok {
					l.logError(fmt.Errorf("failed to get module for file %q", file))

					break
				}

				currentContents, ok := l.cache.GetFileContents(file)
				if !ok {
					l.logError(fmt.Errorf("failed to get contents for file %q", file))

					break
				}

				allRuleHeadLocations, err := rego.AllRuleHeadLocations(ctx, filepath.Base(file), currentContents, currentModule)
				if err != nil {
					l.logError(fmt.Errorf("failed to get rule head locations: %w", err))

					break
				}

				// if there are none, then it's a package evaluation
				ruleHeadLocations := allRuleHeadLocations[path]

				var input io.Reader

				// When the first comment in the file is `regal eval: use-as-input`, the AST of that module is
				// used as the input rather than the contents of input.json. This is a development feature for
				// working on rules (built-in or custom), allowing querying the AST of the module directly.
				if len(currentModule.Comments) > 0 && regalEvalUseAsInputComment.Match(currentModule.Comments[0].Text) {
					inputMap, err := rparse.PrepareAST(file, currentContents, currentModule)
					if err != nil {
						l.logError(fmt.Errorf("failed to prepare module: %w", err))

						break
					}

					bs, err := encoding.JSON().Marshal(inputMap)
					if err != nil {
						l.logError(fmt.Errorf("failed to marshal module: %w", err))

						break
					}

					input = bytes.NewReader(bs)
				} else {
					// Normal mode — try to find the input.json file in the workspace and use as input
					_, input = rio.FindInput(uri.ToPath(l.clientIdentifier, file), l.workspacePath())
				}

				result, err := l.EvalWorkspacePath(ctx, path, input)
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to evaluate workspace path: %v\n", err)

					cleanedMessage := strings.Replace(err.Error(), l.workspaceRootURI+"/", "", 1)

					if err := l.conn.Notify(ctx, "window/showMessage", types.ShowMessageParams{
						Type:    1, // error
						Message: cleanedMessage,
					}); err != nil {
						l.logError(fmt.Errorf("failed to notify client of eval error: %w", err))
					}

					break
				}

				target := "package"
				if len(ruleHeadLocations) > 0 {
					target = strings.TrimPrefix(path, currentModule.Package.Path.String()+".")
				}

				if l.clientIdentifier == clients.IdentifierVSCode {
					responseParams := map[string]any{
						"result": result,
						"line":   line,
						"target": target,
						// only used when the target is 'package'
						"package": strings.TrimPrefix(currentModule.Package.Path.String(), "data."),
						// only used when the target is a rule
						"rule_head_locations": ruleHeadLocations,
					}

					responseResult := map[string]any{}

					if err = l.conn.Call(ctx, "regal/showEvalResult", responseParams, &responseResult); err != nil {
						l.logError(fmt.Errorf("regal/showEvalResult failed: %v", err.Error()))
					}
				} else {
					output := filepath.Join(l.workspacePath(), "output.json")

					var f *os.File

					f, err = os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)

					if err == nil {
						var jsonVal []byte

						value := result.Value
						if result.IsUndefined {
							// Display undefined as an empty object
							// we could also go with "<undefined>" or similar
							value = make(map[string]any)
						}

						json := encoding.JSON()

						jsonVal, err = json.MarshalIndent(value, "", "  ")
						if err == nil {
							// staticcheck thinks err here is never used, but I think that's false?
							_, err = f.Write(jsonVal) //nolint:staticcheck
						}

						f.Close()
					}
				}
			}

			if err != nil {
				l.logError(err)

				if err := l.conn.Notify(ctx, "window/showMessage", types.ShowMessageParams{
					Type:    1, // error
					Message: err.Error(),
				}); err != nil {
					l.logError(fmt.Errorf("failed to notify client of command error: %w", err))
				}

				break
			}

			if fixed {
				if err = l.conn.Call(ctx, methodWorkspaceApplyEdit, editParams, nil); err != nil {
					l.logError(fmt.Errorf("failed %s notify: %v", methodWorkspaceApplyEdit, err.Error()))
				}
			}
		}
	}
}

// StartWorkspaceStateWorker will poll for changes to the workspaces state that
// are not sent from the client. For example, when a file a is removed from the
// workspace after changing branch.
func (l *LanguageServer) StartWorkspaceStateWorker(ctx context.Context) {
	timer := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			// first clear files that are missing from the workspaceDir
			for fileURI := range l.cache.GetAllFiles() {
				filePath := uri.ToPath(l.clientIdentifier, fileURI)

				_, err := os.Stat(filePath)
				if !os.IsNotExist(err) {
					// if the file is not missing, we have no work to do
					continue
				}

				// clear the cache first,
				l.cache.Delete(fileURI)

				// then send the diagnostics message based on the cleared cache
				if err = l.sendFileDiagnostics(ctx, fileURI); err != nil {
					l.logError(fmt.Errorf("failed to send diagnostic: %w", err))
				}
			}

			// for this next operation, the workspace root must be set as it's
			// used to scan for new files.
			if l.workspaceRootURI == "" {
				continue
			}

			// next, check if there are any new files that are not ignored and
			// need to be loaded. We get new only so that files being worked
			// on are not loaded from disk during editing.
			changedOrNewURIs, err := l.loadWorkspaceContents(ctx, true)
			if err != nil {
				l.logError(fmt.Errorf("failed to refresh workspace contents: %w", err))

				continue
			}

			for _, cnURI := range changedOrNewURIs {
				l.lintFileJobs <- lintFileJob{
					URI:    cnURI,
					Reason: "internal/workspaceStateWorker/changedOrNewFile",
				}
			}
		}
	}
}

// StartTemplateWorker runs the process of the server that templates newly
// created Rego files.
func (l *LanguageServer) StartTemplateWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-l.templateFile:
			// determine the new contents for the file, if permitted
			newContents, err := l.templateContentsForFile(job.URI)
			if err != nil {
				l.logError(fmt.Errorf("failed to template new file: %w", err))

				continue
			}

			// set the contents of the new file in the cache immediately as
			// these must be update to date in order for fixRenameParams
			// to work
			l.cache.SetFileContents(job.URI, newContents)

			var edits []any

			edits = append(edits, types.TextDocumentEdit{
				TextDocument: types.OptionalVersionedTextDocumentIdentifier{URI: job.URI},
				Edits:        ComputeEdits("", newContents),
			})

			// determine if a rename is needed based on the new file contents
			renameParams, err := l.fixRenameParams(
				"Rename file to match package path",
				&fixes.DirectoryPackageMismatch{},
				job.URI,
			)
			if err != nil {
				l.logError(fmt.Errorf("failed to fix directory package mismatch: %w", err))

				continue
			}

			// move the file and clean up any empty directories ifd required
			fileURI := job.URI

			if len(renameParams.Edit.DocumentChanges) > 0 {
				edits = append(edits, renameParams.Edit.DocumentChanges[0])
			}

			// check if there are any dirs to clean
			if len(renameParams.Edit.DocumentChanges) > 0 {
				dirs, err := util.DirCleanUpPaths(
					uri.ToPath(l.clientIdentifier, renameParams.Edit.DocumentChanges[0].OldURI),
					[]string{
						// stop at the root
						l.workspacePath(),
						// also preserve any dirs needed for the new file
						uri.ToPath(l.clientIdentifier, renameParams.Edit.DocumentChanges[0].NewURI),
					},
				)
				if err != nil {
					l.logError(fmt.Errorf("failed to delete empty directories: %w", err))

					continue
				}

				for _, dir := range dirs {
					edits = append(
						edits,
						types.DeleteFile{
							Kind: "delete",
							URI:  uri.FromPath(l.clientIdentifier, dir),
							Options: &types.DeleteFileOptions{
								Recursive:         true,
								IgnoreIfNotExists: true,
							},
						},
					)
				}

				l.cache.Delete(renameParams.Edit.DocumentChanges[0].OldURI)
			}

			if err = l.conn.Call(ctx, methodWorkspaceApplyEdit, map[string]any{
				"label": "Template new Rego file",
				"edit": map[string]any{
					"documentChanges": edits,
				},
			}, nil); err != nil {
				l.logError(fmt.Errorf("failed %s notify: %v", methodWorkspaceApplyEdit, err.Error()))
			}

			// finally, trigger a diagnostics run for the new file
			updateEvent := lintFileJob{
				Reason: "internal/templateNewFile",
				URI:    fileURI,
			}

			l.lintFileJobs <- updateEvent
		}
	}
}

func (l *LanguageServer) StartWebServer(ctx context.Context) {
	l.webServer.Start(ctx)
}

func (l *LanguageServer) templateContentsForFile(fileURI string) (string, error) {
	content, ok := l.cache.GetFileContents(fileURI)
	if !ok {
		return "", fmt.Errorf("failed to get file contents for URI %q", fileURI)
	}

	if content != "" {
		return "", errors.New("file already has contents, templating not allowed")
	}

	diskContent, err := os.ReadFile(uri.ToPath(l.clientIdentifier, fileURI))
	if err == nil {
		// then we found the file on disk
		if string(diskContent) != "" {
			return "", errors.New("file on disk already has contents, templating not allowed")
		}
	}

	path := uri.ToPath(l.clientIdentifier, fileURI)
	dir := filepath.Dir(path)

	roots, err := config.GetPotentialRoots(uri.ToPath(l.clientIdentifier, fileURI))
	if err != nil {
		return "", fmt.Errorf("failed to get potential roots during templating of new file: %w", err)
	}

	longestPrefixRoot := ""

	for _, root := range roots {
		if strings.HasPrefix(dir, root) && len(root) > len(longestPrefixRoot) {
			longestPrefixRoot = root
		}
	}

	if longestPrefixRoot == "" {
		return "", fmt.Errorf("failed to find longest prefix root for templating of new file: %s", path)
	}

	parts := slices.Compact(strings.Split(strings.TrimPrefix(dir, longestPrefixRoot), string(os.PathSeparator)))

	var pkg string

	validPathComponentPattern := regexp.MustCompile(`^\w+[\w\-]*\w+$`)

	for _, part := range parts {
		if part == "" {
			continue
		}

		if !validPathComponentPattern.MatchString(part) {
			return "", fmt.Errorf("failed to template new file as package path contained invalid part: %s", part)
		}

		switch {
		case strings.Contains(part, "-"):
			pkg += fmt.Sprintf(`["%s"]`, part)
		case pkg == "":
			pkg += part
		default:
			pkg += "." + part
		}
	}

	// if we are in the root, then we can use main as a default
	if pkg == "" {
		pkg = "main"
	}

	if strings.HasSuffix(fileURI, "_test.rego") {
		pkg += "_test"
	}

	return fmt.Sprintf("package %s\n\nimport rego.v1\n", pkg), nil
}

func (l *LanguageServer) fixEditParams(
	label string,
	fix fixes.Fix,
	commandParseOpts commands.ParseOptions,
	params types.ExecuteCommandParams,
) (bool, *types.ApplyWorkspaceEditParams, error) {
	pr, err := commands.Parse(params, commandParseOpts)
	if err != nil {
		return false, nil, fmt.Errorf("failed to parse command params: %w", err)
	}

	oldContent, ok := l.cache.GetFileContents(pr.Target)
	if !ok {
		return false, nil, fmt.Errorf("could not get file contents for uri %q", pr.Target)
	}

	rto := &fixes.RuntimeOptions{
		BaseDir: uri.ToPath(l.clientIdentifier, l.workspaceRootURI),
	}
	if pr.Location != nil {
		rto.Locations = []ast.Location{*pr.Location}
	}

	fixResults, err := fix.Fix(
		&fixes.FixCandidate{
			Filename: filepath.Base(uri.ToPath(l.clientIdentifier, pr.Target)),
			Contents: []byte(oldContent),
		},
		rto,
	)
	if err != nil {
		return false, nil, fmt.Errorf("failed to fix: %w", err)
	}

	if len(fixResults) == 0 {
		return false, &types.ApplyWorkspaceEditParams{}, nil
	}

	editParams := &types.ApplyWorkspaceEditParams{
		Label: label,
		Edit: types.WorkspaceEdit{
			DocumentChanges: []types.TextDocumentEdit{
				{
					TextDocument: types.OptionalVersionedTextDocumentIdentifier{URI: pr.Target},
					Edits:        ComputeEdits(oldContent, string(fixResults[0].Contents)),
				},
			},
		},
	}

	return true, editParams, nil
}

func (l *LanguageServer) fixRenameParams(
	label string,
	fix fixes.Fix,
	fileURL string,
) (types.ApplyWorkspaceRenameEditParams, error) {
	contents, ok := l.cache.GetFileContents(fileURL)
	if !ok {
		return types.ApplyWorkspaceRenameEditParams{}, fmt.Errorf("failed to file contents %q", fileURL)
	}

	roots, err := config.GetPotentialRoots(l.workspacePath())
	if err != nil {
		return types.ApplyWorkspaceRenameEditParams{}, fmt.Errorf("failed to get potential roots: %w", err)
	}

	file := uri.ToPath(l.clientIdentifier, fileURL)
	baseDir := util.FindClosestMatchingRoot(file, roots)

	results, err := fix.Fix(
		&fixes.FixCandidate{
			Filename: file,
			Contents: []byte(contents),
		},
		&fixes.RuntimeOptions{
			Config:  l.getLoadedConfig(),
			BaseDir: baseDir,
		},
	)
	if err != nil {
		return types.ApplyWorkspaceRenameEditParams{}, fmt.Errorf("failed attempted fix: %w", err)
	}

	if len(results) == 0 {
		return types.ApplyWorkspaceRenameEditParams{
			Label: label,
			Edit:  types.WorkspaceRenameEdit{},
		}, nil
	}

	result := results[0]

	renameEdit := types.WorkspaceRenameEdit{
		DocumentChanges: []types.RenameFile{
			{
				Kind:   "rename",
				OldURI: uri.FromPath(l.clientIdentifier, result.Rename.FromPath),
				NewURI: uri.FromPath(l.clientIdentifier, result.Rename.ToPath),
				Options: &types.RenameFileOptions{
					Overwrite:      false,
					IgnoreIfExists: false,
				},
			},
		},
	}

	return types.ApplyWorkspaceRenameEditParams{
		Label: label,
		Edit:  renameEdit,
	}, nil
}

// processHoverContentUpdate updates information about built in, and keyword
// positions in the cache for use when handling hover requests.
func (l *LanguageServer) processHoverContentUpdate(ctx context.Context, fileURI string, content string) error {
	if l.ignoreURI(fileURI) {
		return nil
	}

	if _, ok := l.cache.GetFileContents(fileURI); !ok {
		// If the file is not in the cache, exit early or else
		// we might accidentally put it in the cache after it's been
		// deleted: https://github.com/StyraInc/regal/issues/679
		return nil
	}

	l.cache.SetFileContents(fileURI, content)

	bis := l.builtinsForCurrentCapabilities()

	success, err := updateParse(ctx, l.cache, l.regoStore, fileURI, bis)
	if err != nil {
		return fmt.Errorf("failed to update parse: %w", err)
	}

	if !success {
		return nil
	}

	if err = hover.UpdateBuiltinPositions(l.cache, fileURI, bis); err != nil {
		return fmt.Errorf("failed to update builtin positions: %w", err)
	}

	if err = hover.UpdateKeywordLocations(ctx, l.cache, fileURI); err != nil {
		return fmt.Errorf("failed to update keyword locations: %w", err)
	}

	return nil
}

func (l *LanguageServer) logError(err error) {
	if l.errorLog != nil {
		fmt.Fprintf(l.errorLog, "ERROR: %s\n", err)
	}
}

type HoverResponse struct {
	Contents types.MarkupContent `json:"contents"`
	Range    types.Range         `json:"range"`
}

func (l *LanguageServer) handleTextDocumentHover(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.TextDocumentHoverParams

	json := encoding.JSON()

	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	if l.ignoreURI(params.TextDocument.URI) {
		return nil, nil
	}

	// The Zed editor doesn't show CodeDescription.Href in diagnostic messages.
	// Instead, we hijack the hover request to show the documentation links
	// when there are violations present.
	violations, ok := l.cache.GetFileDiagnostics(params.TextDocument.URI)
	if l.clientIdentifier == clients.IdentifierZed && ok && len(violations) > 0 {
		var docSnippets []string

		var sharedRange types.Range

		for _, v := range violations {
			if v.Range.Start.Line == params.Position.Line &&
				v.Range.Start.Character <= params.Position.Character &&
				v.Range.End.Character >= params.Position.Character {
				// this is an approximation, if there are multiple violations on the same line
				// where hover loc is in their range, then they all just share a range as a
				// single range is needed in the hover response.
				sharedRange = v.Range
				docSnippets = append(docSnippets, fmt.Sprintf("[%s/%s](%s)", v.Source, v.Code, v.CodeDescription.Href))
			}
		}

		if len(docSnippets) > 1 {
			return HoverResponse{
				Contents: types.MarkupContent{
					Kind:  "markdown",
					Value: "Documentation links:\n\n* " + strings.Join(docSnippets, "\n* "),
				},
				Range: sharedRange,
			}, nil
		} else if len(docSnippets) == 1 {
			return HoverResponse{
				Contents: types.MarkupContent{
					Kind:  "markdown",
					Value: "Documentation: " + docSnippets[0],
				},
				Range: sharedRange,
			}, nil
		}
	}

	builtinsOnLine, ok := l.cache.GetBuiltinPositions(params.TextDocument.URI)
	// when no builtins are found, we can't return a useful hover response.
	// log the error, but return an empty struct to avoid an error being shown in the client.
	if !ok {
		l.logError(fmt.Errorf("could not get builtins for uri %q", params.TextDocument.URI))

		// return "null" as per the spec
		return nil, nil
	}

	for _, bp := range builtinsOnLine[params.Position.Line+1] {
		if params.Position.Character >= bp.Start-1 && params.Position.Character <= bp.End-1 {
			contents := hover.CreateHoverContent(bp.Builtin)

			return HoverResponse{
				Contents: types.MarkupContent{
					Kind:  "markdown",
					Value: contents,
				},
				Range: types.Range{
					Start: types.Position{Line: bp.Line - 1, Character: bp.Start - 1},
					End:   types.Position{Line: bp.Line - 1, Character: bp.End - 1},
				},
			}, nil
		}
	}

	keywordsOnLine, ok := l.cache.GetKeywordLocations(params.TextDocument.URI)
	if !ok {
		// when no keywords are found, we can't return a useful hover response.
		// return "null" as per the spec
		return nil, nil
	}

	for _, kp := range keywordsOnLine[params.Position.Line+1] {
		if params.Position.Character >= kp.Start-1 && params.Position.Character <= kp.End-1 {
			link, ok := examples.GetKeywordLink(kp.Name)
			if !ok {
				continue
			}

			contents := fmt.Sprintf(`### %s

[View examples](%s) for the '%s' keyword.
`, kp.Name, link, kp.Name)

			return HoverResponse{
				Contents: types.MarkupContent{
					Kind:  "markdown",
					Value: contents,
				},
				Range: types.Range{
					Start: types.Position{Line: kp.Line - 1, Character: kp.Start - 1},
					End:   types.Position{Line: kp.Line - 1, Character: kp.End - 1},
				},
			}, nil
		}
	}

	// return "null" as per the spec
	return nil, nil
}

func (l *LanguageServer) handleTextDocumentCodeAction(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.CodeActionParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	yes := true
	actions := make([]types.CodeAction, 0)

	if l.ignoreURI(params.TextDocument.URI) {
		return actions, nil
	}

	// only VS Code has the capability to open a provided URL, as far as we know
	// if we learn about others with this capability later, we should add them!
	if l.clientIdentifier == clients.IdentifierVSCode {
		explorerURL := l.webServer.GetBaseURL() + "/explorer" +
			strings.TrimPrefix(params.TextDocument.URI, l.workspaceRootURI)

		actions = append(actions, types.CodeAction{
			Title: "Explore compiler stages for this policy",
			Kind:  "source.explore",
			Command: types.Command{
				Title:     "Explore compiler stages for this policy",
				Command:   "vscode.open",
				Arguments: &[]any{explorerURL},
			},
		})
	}

	for _, diag := range params.Context.Diagnostics {
		switch diag.Code {
		case ruleNameOPAFmt:
			actions = append(actions, types.CodeAction{
				Title:       "Format using opa fmt",
				Kind:        "quickfix",
				Diagnostics: []types.Diagnostic{diag},
				IsPreferred: &yes,
				Command:     FmtCommand([]string{params.TextDocument.URI}),
			})
		case ruleNameUseRegoV1:
			actions = append(actions, types.CodeAction{
				Title:       "Format for Rego v1 using opa fmt",
				Kind:        "quickfix",
				Diagnostics: []types.Diagnostic{diag},
				IsPreferred: &yes,
				Command:     FmtV1Command([]string{params.TextDocument.URI}),
			})
		case ruleNameUseAssignmentOperator:
			actions = append(actions, types.CodeAction{
				Title:       "Replace = with := in assignment",
				Kind:        "quickfix",
				Diagnostics: []types.Diagnostic{diag},
				IsPreferred: &yes,
				Command: UseAssignmentOperatorCommand([]string{
					params.TextDocument.URI,
					strconv.FormatUint(uint64(diag.Range.Start.Line+1), 10),
					strconv.FormatUint(uint64(diag.Range.Start.Character+1), 10),
				}),
			})
		case ruleNameNoWhitespaceComment:
			actions = append(actions, types.CodeAction{
				Title:       "Format comment to have leading whitespace",
				Kind:        "quickfix",
				Diagnostics: []types.Diagnostic{diag},
				IsPreferred: &yes,
				Command: NoWhiteSpaceCommentCommand([]string{
					params.TextDocument.URI,
					strconv.FormatUint(uint64(diag.Range.Start.Line+1), 10),
					strconv.FormatUint(uint64(diag.Range.Start.Character+1), 10),
				}),
			})
		case ruleNameDirectoryPackageMismatch:
			actions = append(actions, types.CodeAction{
				Title:       "Move file so that directory structure mirrors package path",
				Kind:        "quickfix",
				Diagnostics: []types.Diagnostic{diag},
				IsPreferred: &yes,
				Command: DirectoryStructureMismatchCommand([]string{
					params.TextDocument.URI,
				}),
			})
		}

		if l.clientIdentifier == clients.IdentifierVSCode {
			// always show the docs link
			txt := "Show documentation for " + diag.Code
			actions = append(actions, types.CodeAction{
				Title:       txt,
				Kind:        "quickfix",
				Diagnostics: []types.Diagnostic{diag},
				IsPreferred: &yes,
				Command: types.Command{
					Title:     txt,
					Command:   "vscode.open",
					Arguments: &[]any{diag.CodeDescription.Href},
				},
			})
		}
	}

	return actions, nil
}

func (l *LanguageServer) handleWorkspaceExecuteCommand(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.ExecuteCommandParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	// this must not block, so we send the request to the worker on a buffered channel.
	// the response to the workspace/executeCommand request must be sent before the command is executed
	// so that the client can complete the request and be ready to receive the follow-on request for
	// workspace/applyEdit.
	l.commandRequest <- params

	// however, the contents of the response is not important
	return struct{}{}, nil
}

func (l *LanguageServer) handleTextDocumentInlayHint(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.TextDocumentInlayHintParams

	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	if l.ignoreURI(params.TextDocument.URI) {
		return []types.InlayHint{}, nil
	}

	bis := l.builtinsForCurrentCapabilities()

	// when a file cannot be parsed, we do a best effort attempt to provide inlay hints
	// by finding the location of the first parse error and attempting to parse up to that point
	parseErrors, ok := l.cache.GetParseErrors(params.TextDocument.URI)
	if ok && len(parseErrors) > 0 {
		contents, ok := l.cache.GetFileContents(params.TextDocument.URI)
		if !ok {
			// if there is no content, we can't even do a partial parse
			return []types.InlayHint{}, nil
		}

		return partialInlayHints(parseErrors, contents, params.TextDocument.URI, bis), nil
	}

	module, ok := l.cache.GetModule(params.TextDocument.URI)
	if !ok {
		l.logError(fmt.Errorf("failed to get inlay hint: no parsed module for uri %q", params.TextDocument.URI))

		return []types.InlayHint{}, nil
	}

	inlayHints := getInlayHints(module, bis)

	return inlayHints, nil
}

func (l *LanguageServer) handleTextDocumentCodeLens(
	ctx context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.CodeLensParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	module, ok := l.cache.GetModule(params.TextDocument.URI)
	if !ok {
		return nil, nil // return a null response, as per the spec
	}

	contents, ok := l.cache.GetFileContents(params.TextDocument.URI)
	if !ok {
		return nil, nil // return a null response, as per the spec
	}

	return rego.CodeLenses(ctx, params.TextDocument.URI, contents, module) //nolint:wrapcheck
}

func (l *LanguageServer) handleTextDocumentCompletion(
	ctx context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (any, error) {
	var params types.CompletionParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	// when config ignores a file, then we return an empty completion list
	// as a no-op.
	if l.ignoreURI(params.TextDocument.URI) {
		return types.CompletionList{
			IsIncomplete: false,
			Items:        []types.CompletionItem{},
		}, nil
	}

	// items is allocated here so that the return value is always a non-nil CompletionList
	items, err := l.completionsManager.Run(ctx, params, &providers.Options{
		ClientIdentifier: l.clientIdentifier,
		RootURI:          l.workspaceRootURI,
		Builtins:         l.builtinsForCurrentCapabilities(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find completions: %w", err)
	}

	// make sure the items is always [] instead of null as is required by the spec
	if items == nil {
		return types.CompletionList{
			IsIncomplete: false,
			Items:        make([]types.CompletionItem, 0),
		}, nil
	}

	return types.CompletionList{
		IsIncomplete: true,
		Items:        items,
	}, nil
}

func partialInlayHints(
	parseErrors []types.Diagnostic,
	contents,
	fileURI string,
	builtins map[string]*ast.Builtin,
) []types.InlayHint {
	firstErrorLine := uint(0)
	for _, parseError := range parseErrors {
		if parseError.Range.Start.Line > firstErrorLine {
			firstErrorLine = parseError.Range.Start.Line
		}
	}

	if firstErrorLine == 0 {
		// if there are parse errors from line 0, we skip doing anything
		return []types.InlayHint{}
	}

	if firstErrorLine > uint(len(strings.Split(contents, "\n"))) {
		// if the last valid line is beyond the end of the file, we exit as something is up
		return []types.InlayHint{}
	}

	// select the lines from the contents up to the first parse error
	lines := strings.Join(strings.Split(contents, "\n")[:firstErrorLine], "\n")

	// parse the part of the module that might work
	module, err := rparse.Module(fileURI, lines)
	if err != nil {
		// if we still can't parse the bit we hoped was valid, we exit as this is 'too hard'
		return []types.InlayHint{}
	}

	return getInlayHints(module, builtins)
}

func (l *LanguageServer) handleWorkspaceSymbol(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.WorkspaceSymbolParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	symbols := make([]types.WorkspaceSymbol, 0)
	contents := l.cache.GetAllFiles()

	// Note: currently ignoring params.Query, as the client seems to do a good
	// job of filtering anyway, and that would merely be an optimization here.
	// But perhaps a good one to do at some point, and I'm not sure all clients
	// do this filtering.

	bis := l.builtinsForCurrentCapabilities()

	for moduleURL, module := range l.cache.GetAllModules() {
		content := contents[moduleURL]
		docSyms := documentSymbols(content, module, bis)
		wrkSyms := make([]types.WorkspaceSymbol, 0)

		toWorkspaceSymbols(docSyms, moduleURL, &wrkSyms)

		symbols = append(symbols, wrkSyms...)
	}

	return symbols, nil
}

func (l *LanguageServer) handleTextDocumentDefinition(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.DefinitionParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	if l.ignoreURI(params.TextDocument.URI) {
		return nil, nil
	}

	contents, ok := l.cache.GetFileContents(params.TextDocument.URI)
	if !ok {
		return nil, fmt.Errorf("failed to get file contents for uri %q", params.TextDocument.URI)
	}

	modules, err := l.getFilteredModules()
	if err != nil {
		return nil, fmt.Errorf("failed to filter ignored paths: %w", err)
	}

	orc := oracle.New()
	query := oracle.DefinitionQuery{
		Filename: uri.ToPath(l.clientIdentifier, params.TextDocument.URI),
		Pos:      positionToOffset(contents, params.Position),
		Modules:  modules,
		Buffer:   []byte(contents),
	}

	definition, err := orc.FindDefinition(query)
	if err != nil {
		if errors.Is(err, oracle.ErrNoDefinitionFound) || errors.Is(err, oracle.ErrNoMatchFound) {
			// fail silently — the user could have clicked anywhere. return "null" as per the spec
			return nil, nil
		}

		l.logError(fmt.Errorf("failed to find definition: %w", err))

		// return "null" as per the spec
		return nil, nil
	}

	loc := types.Location{
		URI: uri.FromPath(l.clientIdentifier, definition.Result.File),
		Range: types.Range{
			Start: types.Position{Line: uint(definition.Result.Row - 1), Character: uint(definition.Result.Col - 1)},
			End:   types.Position{Line: uint(definition.Result.Row - 1), Character: uint(definition.Result.Col - 1)},
		},
	}

	return loc, nil
}

func (l *LanguageServer) handleTextDocumentDidOpen(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.TextDocumentDidOpenParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	// if the opened file is ignored in config, then we only store the
	// contents for file level operations like formatting.
	if l.ignoreURI(params.TextDocument.URI) {
		l.cache.SetIgnoredFileContents(
			params.TextDocument.URI,
			params.TextDocument.Text,
		)

		return struct{}{}, nil
	}

	l.cache.SetFileContents(params.TextDocument.URI, params.TextDocument.Text)

	job := lintFileJob{
		Reason: "textDocument/didOpen",
		URI:    params.TextDocument.URI,
	}

	l.lintFileJobs <- job

	l.builtinsPositionJobs <- job

	return struct{}{}, nil
}

func (l *LanguageServer) handleTextDocumentDidClose(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.TextDocumentDidCloseParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	// if the file being closed is ignored in config, then we
	// need to clear it from the ignored state in the cache.
	if l.ignoreURI(params.TextDocument.URI) {
		l.cache.Delete(params.TextDocument.URI)
	}

	return struct{}{}, nil
}

func (l *LanguageServer) handleTextDocumentDidChange(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.TextDocumentDidChangeParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	if len(params.ContentChanges) == 0 {
		return struct{}{}, nil
	}

	// if the changed file is ignored in config, then we only store the
	// contents for file level operations like formatting.
	if l.ignoreURI(params.TextDocument.URI) {
		l.cache.SetIgnoredFileContents(
			params.TextDocument.URI,
			params.ContentChanges[0].Text,
		)

		return struct{}{}, nil
	}

	if len(params.ContentChanges) < 1 {
		return struct{}{}, nil
	}

	l.cache.SetFileContents(params.TextDocument.URI, params.ContentChanges[0].Text)

	job := lintFileJob{
		Reason: "textDocument/didChange",
		URI:    params.TextDocument.URI,
	}

	l.lintFileJobs <- job
	l.builtinsPositionJobs <- job

	return struct{}{}, nil
}

func (l *LanguageServer) handleTextDocumentDidSave(
	ctx context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.TextDocumentDidSaveParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	if params.Text != nil && l.getLoadedConfig() == nil {
		if !strings.Contains(*params.Text, "\r\n") {
			return struct{}{}, nil
		}

		cfg := l.getLoadedConfig()

		enabled, err := linter.NewLinter().WithUserConfig(*cfg).DetermineEnabledRules(ctx)
		if err != nil {
			l.logError(fmt.Errorf("failed to determine enabled rules: %w", err))

			return struct{}{}, nil
		}

		formattingEnabled := false

		for _, rule := range enabled {
			if rule == "opa-fmt" || rule == "use-rego-v1" {
				formattingEnabled = true

				break
			}
		}

		if formattingEnabled {
			resp := types.ShowMessageParams{
				Type:    2, // warning
				Message: "CRLF line ending detected. Please change editor setting to use LF for line endings.",
			}

			if err := l.conn.Notify(ctx, "window/showMessage", resp); err != nil {
				l.logError(fmt.Errorf("failed to notify: %w", err))

				return struct{}{}, nil
			}
		}
	}

	return struct{}{}, nil
}

func (l *LanguageServer) handleTextDocumentDocumentSymbol(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.DocumentSymbolParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	if l.ignoreURI(params.TextDocument.URI) {
		return []types.DocumentSymbol{}, nil
	}

	contents, ok := l.cache.GetFileContents(params.TextDocument.URI)
	if !ok {
		l.logError(fmt.Errorf("failed to get file contents for uri %q", params.TextDocument.URI))

		return []types.DocumentSymbol{}, nil
	}

	module, ok := l.cache.GetModule(params.TextDocument.URI)
	if !ok {
		return []types.DocumentSymbol{}, nil
	}

	bis := l.builtinsForCurrentCapabilities()

	return documentSymbols(contents, module, bis), nil
}

func (l *LanguageServer) handleTextDocumentFoldingRange(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.FoldingRangeParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	module, ok := l.cache.GetModule(params.TextDocument.URI)
	if !ok {
		return []types.FoldingRange{}, nil
	}

	text, ok := l.cache.GetFileContents(params.TextDocument.URI)
	if !ok {
		return []types.FoldingRange{}, nil
	}

	return findFoldingRanges(text, module), nil
}

func (l *LanguageServer) handleTextDocumentFormatting(
	ctx context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.DocumentFormattingParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	var oldContent string

	var ok bool

	// Fetch the contents used for formatting from the appropriate cache location.
	if l.ignoreURI(params.TextDocument.URI) {
		oldContent, ok = l.cache.GetIgnoredFileContents(params.TextDocument.URI)
	} else {
		oldContent, ok = l.cache.GetFileContents(params.TextDocument.URI)
	}

	// if the file is empty, then the formatters will fail, so we template
	// instead
	if oldContent == "" {
		newContent, err := l.templateContentsForFile(params.TextDocument.URI)
		if err != nil {
			return nil, fmt.Errorf("failed to template contents as a templating fallback: %w", err)
		}

		l.cache.ClearFileDiagnostics()

		l.cache.SetFileContents(params.TextDocument.URI, newContent)

		updateEvent := lintFileJob{
			Reason: "internal/templateFormattingFallback",
			URI:    params.TextDocument.URI,
		}

		l.lintFileJobs <- updateEvent

		return ComputeEdits(oldContent, newContent), nil
	}

	if !ok {
		return nil, fmt.Errorf("failed to get file contents for uri %q", params.TextDocument.URI)
	}

	formatter := "opa-fmt"

	if l.clientInitializationOptions.Formatter != nil {
		formatter = *l.clientInitializationOptions.Formatter
	}

	var newContent []byte

	switch formatter {
	case "opa-fmt", "opa-fmt-rego-v1":
		opts := format.Opts{}
		if formatter == "opa-fmt-rego-v1" {
			opts.RegoVersion = ast.RegoV0CompatV1
		}

		f := &fixes.Fmt{OPAFmtOpts: opts}
		p := uri.ToPath(l.clientIdentifier, params.TextDocument.URI)

		fixResults, err := f.Fix(
			&fixes.FixCandidate{Filename: filepath.Base(p), Contents: []byte(oldContent)},
			&fixes.RuntimeOptions{
				BaseDir: l.workspacePath(),
			},
		)
		if err != nil {
			l.logError(fmt.Errorf("failed to format file: %w", err))

			// return "null" as per the spec
			return nil, nil
		}

		if len(fixResults) == 0 {
			return []types.TextEdit{}, nil
		}

		newContent = fixResults[0].Contents
	case "regal-fix":
		// set up an in-memory file provider to pass to the fixer for this one file
		memfp := fileprovider.NewInMemoryFileProvider(map[string][]byte{
			params.TextDocument.URI: []byte(oldContent),
		})

		input, err := memfp.ToInput()
		if err != nil {
			return nil, fmt.Errorf("failed to create fixer input: %w", err)
		}

		f := fixer.NewFixer()
		f.RegisterFixes(fixes.NewDefaultFormatterFixes()...)
		f.RegisterMandatoryFixes(
			&fixes.Fmt{
				NameOverride: "use-rego-v1",
				OPAFmtOpts: format.Opts{
					RegoVersion: ast.RegoV0CompatV1,
				},
			},
		)

		if roots, err := config.GetPotentialRoots(
			l.workspacePath(),
			uri.ToPath(l.clientIdentifier, params.TextDocument.URI),
		); err == nil {
			f.RegisterRoots(roots...)
		} else {
			return nil, fmt.Errorf("could not find potential roots: %w", err)
		}

		li := linter.NewLinter().
			WithInputModules(&input)

		if cfg := l.getLoadedConfig(); cfg != nil {
			li = li.WithUserConfig(*cfg)
		}

		fixReport, err := f.Fix(ctx, &li, memfp)
		if err != nil {
			return nil, fmt.Errorf("failed to format: %w", err)
		}

		if fixReport.TotalFixes() == 0 {
			return []types.TextEdit{}, nil
		}

		newContent, err = memfp.Get(params.TextDocument.URI)
		if err != nil {
			return nil, fmt.Errorf("failed to get formatted contents: %w", err)
		}
	default:
		return nil, fmt.Errorf("unrecognized formatter %q", formatter)
	}

	return ComputeEdits(oldContent, string(newContent)), nil
}

func (l *LanguageServer) handleWorkspaceDidCreateFiles(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.WorkspaceDidCreateFilesParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	if l.ignoreURI(params.Files[0].URI) {
		return struct{}{}, nil
	}

	for _, createOp := range params.Files {
		if _, _, err = cache.UpdateCacheForURIFromDisk(
			l.cache,
			uri.FromPath(l.clientIdentifier, createOp.URI),
			uri.ToPath(l.clientIdentifier, createOp.URI),
		); err != nil {
			return nil, fmt.Errorf("failed to update cache for uri %q: %w", createOp.URI, err)
		}

		job := lintFileJob{
			Reason: "textDocument/didCreate",
			URI:    createOp.URI,
		}

		l.lintFileJobs <- job
		l.builtinsPositionJobs <- job
		l.templateFile <- job
	}

	return struct{}{}, nil
}

func (l *LanguageServer) handleWorkspaceDidDeleteFiles(
	ctx context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.WorkspaceDidDeleteFilesParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	if l.ignoreURI(params.Files[0].URI) {
		return struct{}{}, nil
	}

	for _, deleteOp := range params.Files {
		l.cache.Delete(deleteOp.URI)

		if err := l.sendFileDiagnostics(ctx, deleteOp.URI); err != nil {
			l.logError(fmt.Errorf("failed to send diagnostic: %w", err))
		}
	}

	return struct{}{}, nil
}

func (l *LanguageServer) handleWorkspaceDidRenameFiles(
	ctx context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.WorkspaceDidRenameFilesParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	for _, renameOp := range params.Files {
		if l.ignoreURI(renameOp.OldURI) && l.ignoreURI(renameOp.NewURI) {
			continue
		}

		content, ok := l.cache.GetFileContents(renameOp.OldURI)
		// if the content is not in the cache then we can attempt to load from
		// the disk instead.
		if !ok || content == "" {
			_, content, err = cache.UpdateCacheForURIFromDisk(
				l.cache,
				uri.FromPath(l.clientIdentifier, renameOp.NewURI),
				uri.ToPath(l.clientIdentifier, renameOp.NewURI),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to update cache for uri %q: %w", renameOp.NewURI, err)
			}
		}

		// clear the cache and send diagnostics for the old URI to clear the client
		l.cache.Delete(renameOp.OldURI)

		if err := l.sendFileDiagnostics(ctx, renameOp.OldURI); err != nil {
			l.logError(fmt.Errorf("failed to send diagnostic: %w", err))
		}

		l.cache.SetFileContents(renameOp.NewURI, content)

		job := lintFileJob{
			Reason: "textDocument/didRename",
			URI:    renameOp.NewURI,
		}

		l.lintFileJobs <- job
		l.builtinsPositionJobs <- job
		// if the file being moved is empty, we template it too (if empty)
		l.templateFile <- job
	}

	return struct{}{}, nil
}

func (l *LanguageServer) handleWorkspaceDiagnostic(
	_ context.Context,
	_ *jsonrpc2.Conn,
	_ *jsonrpc2.Request,
) (result any, err error) {
	workspaceReport := types.WorkspaceDiagnosticReport{
		Items: make([]types.WorkspaceFullDocumentDiagnosticReport, 0),
	}

	// if the workspace root is not set, then we return an empty report
	// since we can't provide workspace diagnostics without a workspace root
	// being set. This is unset when the client in is in single file mode.
	if l.workspaceRootURI == "" {
		return workspaceReport, nil
	}

	wkspceDiags, ok := l.cache.GetFileDiagnostics(l.workspaceRootURI)
	if !ok {
		wkspceDiags = []types.Diagnostic{}
	}

	workspaceReport.Items = append(workspaceReport.Items, types.WorkspaceFullDocumentDiagnosticReport{
		URI:     l.workspaceRootURI,
		Kind:    "full",
		Version: nil,
		Items:   wkspceDiags,
	})

	return workspaceReport, nil
}

func (l *LanguageServer) handleInitialize(
	ctx context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (any, error) {
	var params types.InitializeParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	// params.RootURI is not expected to have a trailing slash, but if one is
	// present it will be removed for consistency.
	l.workspaceRootURI = strings.TrimSuffix(params.RootURI, string(os.PathSeparator))

	l.clientIdentifier = clients.DetermineClientIdentifier(params.ClientInfo.Name)

	if l.clientIdentifier == clients.IdentifierGeneric {
		l.logError(
			fmt.Errorf("unable to match client identifier for initializing client, using generic functionality: %s",
				params.ClientInfo.Name),
		)
	}

	l.webServer.SetClient(l.clientIdentifier)

	if params.InitializationOptions != nil {
		l.clientInitializationOptions = *params.InitializationOptions
	}

	regoFilter := types.FileOperationFilter{
		Scheme: "file",
		Pattern: types.FileOperationPattern{
			Glob: "**/*.rego",
		},
	}

	initializeResult := types.InitializeResult{
		Capabilities: types.ServerCapabilities{
			TextDocumentSyncOptions: types.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    1, // TODO: write logic to use 2, for incremental updates
				Save: types.TextDocumentSaveOptions{
					IncludeText: true,
				},
			},
			DiagnosticProvider: types.DiagnosticOptions{
				Identifier:            "rego",
				InterFileDependencies: true,
				WorkspaceDiagnostics:  true,
			},
			Workspace: types.WorkspaceOptions{
				FileOperations: types.FileOperationsServerCapabilities{
					DidCreate: types.FileOperationRegistrationOptions{
						Filters: []types.FileOperationFilter{regoFilter},
					},
					DidRename: types.FileOperationRegistrationOptions{
						Filters: []types.FileOperationFilter{regoFilter},
					},
					DidDelete: types.FileOperationRegistrationOptions{
						Filters: []types.FileOperationFilter{regoFilter},
					},
				},
			},
			InlayHintProvider: types.InlayHintOptions{
				ResolveProvider: false,
			},
			HoverProvider: true,
			CodeActionProvider: types.CodeActionOptions{
				CodeActionKinds: []string{
					"quickfix",
					"source.explore",
				},
			},
			ExecuteCommandProvider: types.ExecuteCommandOptions{
				Commands: []string{
					"regal.debug",
					"regal.eval",
					"regal.fix.opa-fmt",
					"regal.fix.use-rego-v1",
					"regal.fix.use-assignment-operator",
					"regal.fix.no-whitespace-comment",
					"regal.fix.directory-package-mismatch",
				},
			},
			DocumentFormattingProvider: true,
			FoldingRangeProvider:       true,
			DefinitionProvider:         true,
			DocumentSymbolProvider:     true,
			WorkspaceSymbolProvider:    true,
			CompletionProvider: types.CompletionOptions{
				ResolveProvider: false,
				CompletionItem: types.CompletionItemOptions{
					LabelDetailsSupport: true,
				},
			},
			CodeLensProvider: &types.CodeLensOptions{},
		},
	}

	if l.workspaceRootURI != "" {
		workspaceRootPath := l.workspacePath()

		l.bundleCache = bundles.NewCache(&bundles.CacheOptions{
			WorkspacePath: workspaceRootPath,
			ErrorLog:      l.errorLog,
		})

		configFile, err := config.FindConfig(workspaceRootPath)
		if err == nil {
			l.configWatcher.Watch(configFile.Name())
		}

		if _, err = l.loadWorkspaceContents(ctx, false); err != nil {
			return nil, fmt.Errorf("failed to load workspace contents: %w", err)
		}

		l.webServer.SetWorkspaceURI(l.workspaceRootURI)

		l.lintWorkspaceJobs <- lintWorkspaceJob{
			Reason: "server initialize",
			// 'OverwriteAggregates' is set to populate the cache's
			// initial aggregate state. Subsequent runs of lintWorkspaceJobs
			// will not set this and use the cached state.
			OverwriteAggregates: true,
		}
	}

	return initializeResult, nil
}

func (l *LanguageServer) loadWorkspaceContents(ctx context.Context, newOnly bool) ([]string, error) {
	workspaceRootPath := l.workspacePath()

	changedOrNewURIs := make([]string, 0)

	if err := rio.WalkFiles(workspaceRootPath, func(path string) error {
		// TODO(charlieegan3): make this configurable for things like .rq etc?
		if !strings.HasSuffix(path, ".rego") {
			return nil
		}

		fileURI := uri.FromPath(l.clientIdentifier, path)

		if l.ignoreURI(fileURI) {
			return nil
		}

		// if the caller has requested only new files, then we can exit early
		if contents, ok := l.cache.GetFileContents(fileURI); newOnly && ok {
			diskContents, err := os.ReadFile(fileURI)
			if err != nil {
				// then there is nothing we can do here
				return nil
			}

			if string(diskContents) == "" && contents == "" {
				// then there is nothing to be gained from loading from disk
				return nil
			}

			return nil
		}

		changed, _, err := cache.UpdateCacheForURIFromDisk(l.cache, fileURI, path)
		if err != nil {
			return fmt.Errorf("failed to update cache for uri %q: %w", path, err)
		}

		// there is no need to update the parse if the file contents
		// was not changed in the above operation.
		if !changed {
			return nil
		}

		bis := l.builtinsForCurrentCapabilities()

		if _, err = updateParse(ctx, l.cache, l.regoStore, fileURI, bis); err != nil {
			return fmt.Errorf("failed to update parse: %w", err)
		}

		changedOrNewURIs = append(changedOrNewURIs, fileURI)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk workspace dir %q: %w", workspaceRootPath, err)
	}

	if l.bundleCache != nil {
		if _, err := l.bundleCache.Refresh(); err != nil {
			return nil, fmt.Errorf("failed to refresh the bundle cache: %w", err)
		}
	}

	return changedOrNewURIs, nil
}

func (l *LanguageServer) handleInitialized(
	_ context.Context,
	_ *jsonrpc2.Conn,
	_ *jsonrpc2.Request,
) (result any, err error) {
	// if running without config, then we should send the diagnostic request now
	// otherwise it'll happen when the config is loaded
	if !l.configWatcher.IsWatching() {
		l.lintWorkspaceJobs <- lintWorkspaceJob{Reason: "server initialized"}
	}

	return struct{}{}, nil
}

func (*LanguageServer) handleTextDocumentDiagnostic(
	_ context.Context,
	_ *jsonrpc2.Conn,
	_ *jsonrpc2.Request,
) (result any, err error) {
	// this is a no-op. Because we accept the textDocument/didChange event, which contains the new content,
	// we don't need to do anything here as once the new content has been parsed, the diagnostics will be sent
	// on the channel regardless of this request.
	return nil, nil
}

func (l *LanguageServer) handleWorkspaceDidChangeWatchedFiles(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params types.WorkspaceDidChangeWatchedFilesParams
	if err := encoding.JSON().Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	// when a file is changed (saved), then we send trigger a full workspace lint
	regoFiles := make([]string, 0)

	for _, change := range params.Changes {
		if change.URI == "" || l.ignoreURI(change.URI) {
			continue
		}

		regoFiles = append(regoFiles, change.URI)
	}

	if len(regoFiles) > 0 {
		l.lintWorkspaceJobs <- lintWorkspaceJob{
			Reason: fmt.Sprintf("workspace/didChangeWatchedFiles (%s)", strings.Join(regoFiles, ", ")),
		}
	}

	return struct{}{}, nil
}

func (l *LanguageServer) sendFileDiagnostics(ctx context.Context, fileURI string) error {
	fileDiags, ok := l.cache.GetFileDiagnostics(fileURI)
	if !ok {
		fileDiags = []types.Diagnostic{}
	}

	resp := types.FileDiagnostics{
		URI:   fileURI,
		Items: fileDiags,
	}

	if err := l.conn.Notify(ctx, methodTextDocumentPublishDiagnostics, resp); err != nil {
		return fmt.Errorf("failed to notify: %w", err)
	}

	return nil
}

func (l *LanguageServer) getFilteredModules() (map[string]*ast.Module, error) {
	ignore := make([]string, 0)

	if cfg := l.getLoadedConfig(); cfg != nil && cfg.Ignore.Files != nil {
		ignore = cfg.Ignore.Files
	}

	allModules := l.cache.GetAllModules()
	paths := util.Keys(allModules)

	filtered, err := config.FilterIgnoredPaths(paths, ignore, false, l.workspaceRootURI)
	if err != nil {
		return nil, fmt.Errorf("failed to filter ignored paths: %w", err)
	}

	modules := make(map[string]*ast.Module, len(filtered))
	for _, path := range filtered {
		modules[path] = allModules[path]
	}

	return modules, nil
}

func (l *LanguageServer) ignoreURI(fileURI string) bool {
	cfg := l.getLoadedConfig()
	if cfg == nil {
		return false
	}

	paths, err := config.FilterIgnoredPaths(
		[]string{uri.ToPath(l.clientIdentifier, fileURI)},
		cfg.Ignore.Files,
		false,
		l.workspacePath(),
	)
	if err != nil || len(paths) == 0 {
		return true
	}

	return false
}

func (l *LanguageServer) workspacePath() string {
	return uri.ToPath(l.clientIdentifier, l.workspaceRootURI)
}

// builtinsForCurrentCapabilities returns the map of builtins for use
// in the server based on the currently loaded capabilities. If there is no
// config, then the default for the Regal OPA version is used.
func (l *LanguageServer) builtinsForCurrentCapabilities() map[string]*ast.Builtin {
	l.loadedBuiltinsLock.RLock()
	defer l.loadedBuiltinsLock.RUnlock()

	cfg := l.getLoadedConfig()
	if cfg == nil {
		return rego.BuiltinsForCapabilities(ast.CapabilitiesForThisVersion())
	}

	bis, ok := l.loadedBuiltins[cfg.CapabilitiesURL]
	if !ok {
		return rego.BuiltinsForCapabilities(ast.CapabilitiesForThisVersion())
	}

	return bis
}

func positionToOffset(text string, p types.Position) int {
	bytesRead := 0
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if line == "" {
			bytesRead++
		} else {
			bytesRead += len(line) + 1
		}

		//nolint:gosec
		if i == int(p.Line)-1 {
			return bytesRead + int(p.Character)
		}
	}

	return -1
}
