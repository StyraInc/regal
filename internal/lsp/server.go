//nolint:nilerr,nilnil,gochecknoglobals
package lsp

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/jsonrpc2"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/format"
	"github.com/open-policy-agent/opa/v1/storage"
	outil "github.com/open-policy-agent/opa/v1/util"

	rbundle "github.com/styrainc/regal/bundle"
	"github.com/styrainc/regal/internal/capabilities"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/lsp/bundles"
	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/completions"
	"github.com/styrainc/regal/internal/lsp/completions/providers"
	lsconfig "github.com/styrainc/regal/internal/lsp/config"
	"github.com/styrainc/regal/internal/lsp/examples"
	"github.com/styrainc/regal/internal/lsp/handler"
	"github.com/styrainc/regal/internal/lsp/hover"
	"github.com/styrainc/regal/internal/lsp/log"
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
	"github.com/styrainc/regal/pkg/report"
	"github.com/styrainc/regal/pkg/rules"
	"github.com/styrainc/regal/pkg/version"

	"github.com/styrainc/roast/pkg/encoding"
	"github.com/styrainc/roast/pkg/util/concurrent"
)

const (
	methodTextDocumentPublishDiagnostics = "textDocument/publishDiagnostics"
	methodWorkspaceApplyEdit             = "workspace/applyEdit"

	ruleNameOPAFmt                   = "opa-fmt"
	ruleNameUseRegoV1                = "use-rego-v1"
	ruleNameUseAssignmentOperator    = "use-assignment-operator"
	ruleNameNoWhitespaceComment      = "no-whitespace-comment"
	ruleNameDirectoryPackageMismatch = "directory-package-mismatch"
	ruleNameNonRawRegexPattern       = "non-raw-regex-pattern"
)

var (
	noCodeActions     = make([]types.CodeAction, 0)
	noCodeLenses      = make([]types.CodeLens, 0)
	noDocumentSymbols = make([]types.DocumentSymbol, 0)
	noCompletionItems = make([]types.CompletionItem, 0)
	noFoldingRanges   = make([]types.FoldingRange, 0)
	noDiagnostics     = make([]types.Diagnostic, 0)

	trueValue = true
	truePtr   = &trueValue

	orc = oracle.New()
)

type LanguageServerOptions struct {
	// LogWriter is the io.Writer where all logged messages will be written.
	LogWriter io.Writer

	// log.Level controls the verbosity of the logs, with log.LevelOff, no messages
	// are logged, log.LevelMessage logs only messages and errors, and log.LevelDebug
	// Logs all messages.
	LogLevel log.Level

	// WorkspaceDiagnosticsPoll, if set > 0 will cause a full workspace lint
	// to run on this interval. This is intended to be used where eventing
	// is not working, as expected. E.g. with a client that does not send
	// changes or when running in extremely slow environments like GHA with
	// the go race detector on. TODO, work out why this is required.
	WorkspaceDiagnosticsPoll time.Duration
}

func NewLanguageServer(ctx context.Context, opts *LanguageServerOptions) *LanguageServer {
	c := cache.NewCache()
	store := NewRegalStore()

	ls := &LanguageServer{
		cache:                       c,
		regoStore:                   store,
		logWriter:                   opts.LogWriter,
		logLevel:                    opts.LogLevel,
		lintFileJobs:                make(chan lintFileJob, 10),
		lintWorkspaceJobs:           make(chan lintWorkspaceJob, 10),
		builtinsPositionJobs:        make(chan lintFileJob, 10),
		commandRequest:              make(chan types.ExecuteCommandParams, 10),
		templateFileJobs:            make(chan lintFileJob, 10),
		completionsManager:          completions.NewDefaultManager(ctx, c, store),
		webServer:                   web.NewServer(c, opts.LogWriter, opts.LogLevel),
		loadedBuiltins:              concurrent.MapOf(make(map[string]map[string]*ast.Builtin)),
		workspaceDiagnosticsPoll:    opts.WorkspaceDiagnosticsPoll,
		loadedConfigAllRegoVersions: concurrent.MapOf(make(map[string]ast.RegoVersion)),
	}

	ls.configWatcher = lsconfig.NewWatcher(&lsconfig.WatcherOpts{LogFunc: ls.logf})

	return ls
}

type LanguageServer struct {
	logWriter io.Writer
	logLevel  log.Level

	regoStore storage.Store
	conn      *jsonrpc2.Conn

	configWatcher *lsconfig.Watcher
	loadedConfig  *config.Config
	// this is also used to lock the updates to the cache of enabled rules
	loadedConfigLock                     sync.RWMutex
	loadedConfigEnabledNonAggregateRules []string
	loadedConfigEnabledAggregateRules    []string
	loadedConfigAllRegoVersions          *concurrent.Map[string, ast.RegoVersion]
	loadedBuiltins                       *concurrent.Map[string, map[string]*ast.Builtin]

	clientInitializationOptions types.InitializationOptions

	cache       *cache.Cache
	bundleCache *bundles.Cache

	completionsManager *completions.Manager

	commandRequest       chan types.ExecuteCommandParams
	lintWorkspaceJobs    chan lintWorkspaceJob
	lintFileJobs         chan lintFileJob
	builtinsPositionJobs chan lintFileJob
	templateFileJobs     chan lintFileJob

	webServer *web.Server

	workspaceRootURI string
	clientIdentifier clients.Identifier

	workspaceDiagnosticsPoll time.Duration
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
	AggregateReportOnly bool
}

//nolint:wrapcheck
func (l *LanguageServer) Handle(ctx context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
	l.logf(log.LevelDebug, "received request: %s", req.Method)

	// null params are allowed, but only for certain methods
	if req.Params == nil && req.Method != "shutdown" && req.Method != "exit" {
		return nil, handler.ErrInvalidParams
	}

	switch req.Method {
	case "initialize":
		return handler.WithContextAndParams(ctx, req, l.handleInitialize)
	case "initialized":
		return l.handleInitialized()
	case "textDocument/codeAction":
		return handler.WithParams(req, l.handleTextDocumentCodeAction)
	case "textDocument/definition":
		return handler.WithParams(req, l.handleTextDocumentDefinition)
	case "textDocument/diagnostic":
		return l.handleTextDocumentDiagnostic()
	case "textDocument/didOpen":
		return handler.WithParams(req, l.handleTextDocumentDidOpen)
	case "textDocument/didClose":
		return handler.WithParams(req, l.handleTextDocumentDidClose)
	case "textDocument/didSave":
		return handler.WithContextAndParams(ctx, req, l.handleTextDocumentDidSave)
	case "textDocument/documentSymbol":
		return handler.WithParams(req, l.handleTextDocumentDocumentSymbol)
	case "textDocument/didChange":
		return handler.WithParams(req, l.handleTextDocumentDidChange)
	case "textDocument/foldingRange":
		return handler.WithParams(req, l.handleTextDocumentFoldingRange)
	case "textDocument/formatting":
		return handler.WithContextAndParams(ctx, req, l.handleTextDocumentFormatting)
	case "textDocument/hover":
		return handler.WithParams(req, l.handleTextDocumentHover)
	case "textDocument/inlayHint":
		return handler.WithParams(req, l.handleTextDocumentInlayHint)
	case "textDocument/codeLens":
		return handler.WithContextAndParams(ctx, req, l.handleTextDocumentCodeLens)
	case "textDocument/completion":
		return handler.WithContextAndParams(ctx, req, l.handleTextDocumentCompletion)
	case "workspace/didChangeWatchedFiles":
		return handler.WithParams(req, l.handleWorkspaceDidChangeWatchedFiles)
	case "workspace/diagnostic":
		return l.handleWorkspaceDiagnostic()
	case "workspace/didRenameFiles":
		return handler.WithContextAndParams(ctx, req, l.handleWorkspaceDidRenameFiles)
	case "workspace/didDeleteFiles":
		return handler.WithContextAndParams(ctx, req, l.handleWorkspaceDidDeleteFiles)
	case "workspace/didCreateFiles":
		return handler.WithParams(req, l.handleWorkspaceDidCreateFiles)
	case "workspace/executeCommand":
		return handler.WithParams(req, l.handleWorkspaceExecuteCommand)
	case "workspace/symbol":
		return l.handleWorkspaceSymbol()
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
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case job := <-l.lintFileJobs:
				l.logf(log.LevelDebug, "linting file %s (%s)", job.URI, job.Reason)
				bis := l.builtinsForCurrentCapabilities()

				// updateParse will not return an error when the parsing failed,
				// but only when it was impossible
				if _, err := updateParse(ctx, l.cache, l.regoStore, job.URI, bis, l.regoVersionForURI(job.URI)); err != nil {
					l.logf(log.LevelMessage, "failed to update module for %s: %s", job.URI, err)

					continue
				}

				// lint the file and send the diagnostics
				if err := updateFileDiagnostics(
					ctx,
					l.cache,
					l.getLoadedConfig(),
					job.URI,
					l.workspaceRootURI,
					// updateFileDiagnostics only ever updates the diagnostics
					// of non aggregate rules
					l.getEnabledNonAggregateRules(),
				); err != nil {
					l.logf(log.LevelMessage, "failed to update file diagnostics: %s", err)

					continue
				}

				if err := l.sendFileDiagnostics(ctx, job.URI); err != nil {
					l.logf(log.LevelMessage, "failed to send diagnostic: %s", err)

					continue
				}

				l.lintWorkspaceJobs <- lintWorkspaceJob{
					Reason: fmt.Sprintf("file %s %s", job.URI, job.Reason),
					// this run is expected to used the cached aggregate state
					// for other files.
					// The aggregate state for this file will still be updated.
					OverwriteAggregates: false,
					// when a file has changed, then there is no need to run
					// any other rules globally other than aggregate rules.
					AggregateReportOnly: true,
				}

				l.logf(log.LevelDebug, "linting file %s done", job.URI)
			}
		}
	}()

	wg.Add(1)

	workspaceLintRunBufferSize := 10
	workspaceLintRuns := make(chan lintWorkspaceJob, workspaceLintRunBufferSize)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case job := <-l.lintWorkspaceJobs:
				// AggregateReportOnly is set when updating aggregate
				// violations on character changes. Since these happen so
				// frequently, we stop adding to the channel if there already
				// jobs set to preserve performance
				if job.AggregateReportOnly && len(workspaceLintRuns) > workspaceLintRunBufferSize/2 {
					l.log(log.LevelDebug, "rate limiting aggregate reports")

					continue
				}

				workspaceLintRuns <- job
			}
		}
	}()

	if l.workspaceDiagnosticsPoll > 0 {
		wg.Add(1)

		ticker := time.NewTicker(l.workspaceDiagnosticsPoll)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					workspaceLintRuns <- lintWorkspaceJob{
						Reason:              "poll ticker",
						OverwriteAggregates: true,
					}
				}
			}
		}()
	}

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case job := <-workspaceLintRuns:
				l.logf(log.LevelDebug, "linting workspace: %#v", job)

				// if there are no parsed modules in the cache, then there is
				// no need to run the aggregate report. This can happen if the
				// server is very slow to start up.
				if len(l.cache.GetAllModules()) == 0 {
					continue
				}

				targetRules := l.getEnabledAggregateRules()
				if !job.AggregateReportOnly {
					targetRules = append(targetRules, l.getEnabledNonAggregateRules()...)
				}

				err := updateAllDiagnostics(
					ctx,
					l.cache,
					l.getLoadedConfig(),
					l.workspaceRootURI,
					// this is intended to only be set to true once at start up,
					// on following runs, cached aggregate data is used.
					job.OverwriteAggregates,
					job.AggregateReportOnly,
					targetRules,
				)
				if err != nil {
					l.logf(log.LevelMessage, "failed to update all diagnostics: %s", err)
				}

				for fileURI := range l.cache.GetAllFiles() {
					if err := l.sendFileDiagnostics(ctx, fileURI); err != nil {
						l.logf(log.LevelMessage, "failed to send diagnostic: %s", err)
					}
				}

				l.log(log.LevelDebug, "linting workspace done")
			}
		}
	}()

	<-ctx.Done()
	wg.Wait()
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

			success, err := updateParse(ctx, l.cache, l.regoStore, fileURI, bis, l.regoVersionForURI(fileURI))
			if err != nil {
				l.logf(log.LevelMessage, "failed to update parse: %s", err)

				continue
			}

			if !success {
				continue
			}

			if err = hover.UpdateBuiltinPositions(l.cache, fileURI, bis); err != nil {
				l.logf(log.LevelMessage, "failed to update builtin positions: %s", err)

				continue
			}

			if err = hover.UpdateKeywordLocations(ctx, l.cache, fileURI); err != nil {
				l.logf(log.LevelMessage, "failed to update keyword positions: %s", err)

				continue
			}
		}
	}
}

func (l *LanguageServer) getLoadedConfig() *config.Config {
	l.loadedConfigLock.RLock()
	defer l.loadedConfigLock.RUnlock()

	return l.loadedConfig
}

func (l *LanguageServer) getEnabledNonAggregateRules() []string {
	l.loadedConfigLock.RLock()
	defer l.loadedConfigLock.RUnlock()

	return l.loadedConfigEnabledNonAggregateRules
}

func (l *LanguageServer) getEnabledAggregateRules() []string {
	l.loadedConfigLock.RLock()
	defer l.loadedConfigLock.RUnlock()

	return l.loadedConfigEnabledAggregateRules
}

// loadEnabledRulesFromConfig is used to cache the enabled rules for the current
// config. These take some time to compute and only change when config changes,
// so we can store them on the server to speed up diagnostic runs.
func (l *LanguageServer) loadEnabledRulesFromConfig(ctx context.Context, cfg config.Config) error {
	lint := linter.NewLinter().WithUserConfig(cfg)

	enabledRules, err := lint.DetermineEnabledRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to determine enabled rules: %w", err)
	}

	enabledAggregateRules, err := lint.DetermineEnabledAggregateRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to determine enabled aggregate rules: %w", err)
	}

	l.loadedConfigLock.Lock()
	defer l.loadedConfigLock.Unlock()

	l.loadedConfigEnabledNonAggregateRules = []string{}

	for _, r := range enabledRules {
		if !slices.Contains(enabledAggregateRules, r) {
			l.loadedConfigEnabledNonAggregateRules = append(l.loadedConfigEnabledNonAggregateRules, r)
		}
	}

	l.loadedConfigEnabledAggregateRules = enabledAggregateRules

	return nil
}

func (l *LanguageServer) StartConfigWorker(ctx context.Context) {
	if err := l.configWatcher.Start(ctx); err != nil {
		l.logf(log.LevelMessage, "failed to start config watcher: %s", err)

		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case path := <-l.configWatcher.Reload:
			configFileBs, err := os.ReadFile(path)
			if err != nil {
				l.logf(log.LevelMessage, "failed to open config file: %s", err)

				continue
			}

			var userConfig config.Config

			// EOF errors are ignored here as then we just use the default config
			if err = yaml.Unmarshal(configFileBs, &userConfig); err != nil && !errors.Is(err, io.EOF) {
				l.logf(log.LevelMessage, "failed to reload config: %s", err)

				continue
			}

			mergedConfig, err := config.LoadConfigWithDefaultsFromBundle(&rbundle.LoadedBundle, &userConfig)
			if err != nil {
				l.logf(log.LevelMessage, "failed to load config: %s", err)

				continue
			}

			l.loadedConfigLock.Lock()
			l.loadedConfig = &mergedConfig
			l.loadedConfigLock.Unlock()

			// Rego versions may have changed, so reload them.
			allRegoVersions, err := config.AllRegoVersions(
				uri.ToPath(l.clientIdentifier, l.workspaceRootURI),
				l.getLoadedConfig(),
			)
			if err != nil {
				l.logf(log.LevelMessage, "failed to reload rego versions: %s", err)
			}

			l.loadedConfigAllRegoVersions.Clear()

			for k, v := range allRegoVersions {
				l.loadedConfigAllRegoVersions.Set(k, v)
			}

			// Enabled rules might have changed with the new config, so reload.
			err = l.loadEnabledRulesFromConfig(ctx, mergedConfig)
			if err != nil {
				l.logf(log.LevelMessage, "failed to cache enabled rules: %s", err)
			}

			// Capabilities URL may have changed, so we should reload it.
			cfg := l.getLoadedConfig()

			capsURL := "regal:///capabilities/default"
			if cfg != nil && cfg.CapabilitiesURL != "" {
				capsURL = cfg.CapabilitiesURL
			}

			caps, err := capabilities.Lookup(ctx, capsURL)
			if err != nil {
				l.logf(log.LevelMessage, "failed to load capabilities for URL %q: %s", capsURL, err)

				continue
			}

			bis := rego.BuiltinsForCapabilities(caps)

			l.loadedBuiltins.Set(capsURL, bis)

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
					l.logf(log.LevelMessage, "failed to remove mod from store: %s", err)
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
					if _, err = updateParse(ctx, l.cache, l.regoStore, k, bis, l.regoVersionForURI(k)); err != nil {
						l.logf(log.LevelMessage, "failed to update parse for previously ignored file %q: %s", k, err)
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
						StateDir:       config.GlobalConfigDir(true),
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

func (l *LanguageServer) StartCommandWorker(ctx context.Context) { //nolint:maintidx
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

			if len(params.Arguments) != 1 {
				l.logf(log.LevelMessage, "expected one argument, got %d", len(params.Arguments))

				continue
			}

			jsonData, ok := params.Arguments[0].(string)
			if !ok {
				l.logf(log.LevelMessage, "expected argument to be a json.RawMessage, got %T", params.Arguments[0])

				continue
			}

			var args commandArgs

			err = json.Unmarshal([]byte(jsonData), &args)
			if err != nil {
				l.logf(log.LevelMessage, "failed to unmarshal command arguments: %s", err)

				continue
			}

			var fixed bool

			switch params.Command {
			case "regal.fix.opa-fmt":
				fixed, editParams, err = l.fixEditParams(
					"Format using opa fmt",
					&fixes.Fmt{OPAFmtOpts: format.Opts{}},
					args,
				)
			case "regal.fix.use-rego-v1":
				fixed, editParams, err = l.fixEditParams(
					"Format for Rego v1 using opa-fmt",
					&fixes.Fmt{OPAFmtOpts: format.Opts{RegoVersion: ast.RegoV0CompatV1}},
					args,
				)
			case "regal.fix.use-assignment-operator":
				fixed, editParams, err = l.fixEditParams(
					"Replace = with := in assignment",
					&fixes.UseAssignmentOperator{},
					args,
				)
			case "regal.fix.no-whitespace-comment":
				fixed, editParams, err = l.fixEditParams(
					"Format comment to have leading whitespace",
					&fixes.NoWhitespaceComment{},
					args,
				)
			case "regal.fix.non-raw-regex-pattern":
				fixed, editParams, err = l.fixEditParams(
					"Replace \" with ` in regex pattern",
					&fixes.NonRawRegexPattern{},
					args,
				)
			case "regal.fix.directory-package-mismatch":
				params, err := l.fixRenameParams(
					"Rename file to match package path",
					&fixes.DirectoryPackageMismatch{},
					args.Target,
				)
				if err != nil {
					l.logf(log.LevelMessage, "failed to fix directory package mismatch: %s", err)

					break
				}

				if err := l.conn.Call(ctx, methodWorkspaceApplyEdit, params, nil); err != nil {
					l.logf(log.LevelMessage, "failed %s notify: %v", methodWorkspaceApplyEdit, err.Error())
				}

				// handle this ourselves as it's a rename and not a content edit
				fixed = false
			case "regal.debug":
				file := args.Target
				if file == "" {
					l.logf(log.LevelMessage, "expected command target to be set, got %q", file)

					break
				}

				path := args.QueryPath
				if path == "" {
					l.logf(log.LevelMessage, "expected command query path to be set, got %q", path)

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
					l.logf(log.LevelMessage, "regal/startDebugging failed: %v", err.Error())
				}
			case "regal.eval":
				file := args.Target
				if file == "" {
					l.logf(log.LevelMessage, "expected command target to be set, got %q", file)

					break
				}

				path := args.QueryPath
				if path == "" {
					l.logf(log.LevelMessage, "expected command query path to be set, got %q", path)

					break
				}

				currentContents, currentModule, ok := l.cache.GetContentAndModule(file)
				if !ok {
					l.logf(log.LevelMessage, "failed to get content or module for file %q", file)

					break
				}

				var allRuleHeadLocations rego.RuleHeads

				allRuleHeadLocations, err = rego.AllRuleHeadLocations(ctx, filepath.Base(file), currentContents, currentModule)
				if err != nil {
					l.logf(log.LevelMessage, "failed to get rule head locations: %s", err)

					break
				}

				// if there are none, then it's a package evaluation
				ruleHeadLocations := allRuleHeadLocations[path]

				var inputMap map[string]any

				// When the first comment in the file is `regal eval: use-as-input`, the AST of that module is
				// used as the input rather than the contents of input.json/yaml. This is a development feature for
				// working on rules (built-in or custom), allowing querying the AST of the module directly.
				if len(currentModule.Comments) > 0 && regalEvalUseAsInputComment.Match(currentModule.Comments[0].Text) {
					inputMap, err = rparse.PrepareAST(file, currentContents, currentModule)
					if err != nil {
						l.logf(log.LevelMessage, "failed to prepare module: %s", err)

						break
					}
				} else {
					// Normal mode — try to find the input.json/yaml file in the workspace and use as input
					// NOTE that we don't break on missing input, as some rules don't depend on that, and should
					// still be evaluable. We may consider returning some notice to the user though.
					_, inputMap = rio.FindInput(uri.ToPath(l.clientIdentifier, file), l.workspacePath())
				}

				var result EvalPathResult

				result, err = l.EvalWorkspacePath(ctx, path, inputMap)
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to evaluate workspace path: %v\n", err)

					cleanedMessage := strings.Replace(err.Error(), l.workspaceRootURI+"/", "", 1)

					if err := l.conn.Notify(ctx, "window/showMessage", types.ShowMessageParams{
						Type:    1, // error
						Message: cleanedMessage,
					}); err != nil {
						l.logf(log.LevelMessage, "failed to notify client of eval error: %s", err)
					}

					break
				}

				target := "package"
				if len(ruleHeadLocations) > 0 {
					target = strings.TrimPrefix(path, currentModule.Package.Path.String()+".")
				}

				if l.clientInitializationOptions.EvalCodelensDisplayInline != nil &&
					*l.clientInitializationOptions.EvalCodelensDisplayInline {
					responseParams := map[string]any{
						"result": result,
						"line":   args.Row,
						"target": target,
						// only used when the target is 'package'
						"package": strings.TrimPrefix(currentModule.Package.Path.String(), "data."),
						// only used when the target is a rule
						"rule_head_locations": ruleHeadLocations,
					}

					responseResult := map[string]any{}

					if err = l.conn.Call(ctx, "regal/showEvalResult", responseParams, &responseResult); err != nil {
						l.logf(log.LevelMessage, "regal/showEvalResult failed: %v", err.Error())
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
							_, err = f.Write(jsonVal)
						}

						f.Close()
					}
				}
			}

			if err != nil {
				l.logf(log.LevelMessage, "command failed: %s", err)

				if err := l.conn.Notify(ctx, "window/showMessage", types.ShowMessageParams{
					Type:    1, // error
					Message: err.Error(),
				}); err != nil {
					l.logf(log.LevelMessage, "failed to notify client of command error: %s", err)
				}

				break
			}

			if fixed {
				if err = l.conn.Call(ctx, methodWorkspaceApplyEdit, editParams, nil); err != nil {
					l.logf(log.LevelMessage, "failed %s notify: %v", methodWorkspaceApplyEdit, err.Error())
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
					l.logf(log.LevelMessage, "failed to send diagnostic: %s", err)
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
			newURIs, err := l.loadWorkspaceContents(ctx, true)
			if err != nil {
				l.logf(log.LevelMessage, "failed to refresh workspace contents: %s", err)

				continue
			}

			for _, cnURI := range newURIs {
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
		case job := <-l.templateFileJobs:
			// disable the templating feature for files in the workspace root.
			if filepath.Dir(uri.ToPath(l.clientIdentifier, job.URI)) ==
				uri.ToPath(l.clientIdentifier, l.workspaceRootURI) {
				continue
			}

			// determine the new contents for the file, if permitted
			newContents, err := l.templateContentsForFile(job.URI)
			if err != nil {
				l.logf(log.LevelMessage, "failed to template new file: %s", err)

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

			// set the cache contents so that the fix can access this content
			// when renaming the file if required.
			l.cache.SetFileContents(job.URI, newContents)

			// determine if a rename is needed based on the new file contents.
			// renameParams will be empty if there are no renames needed
			renameParams, err := l.fixRenameParams(
				"Template new Rego file",
				&fixes.DirectoryPackageMismatch{},
				job.URI,
			)
			if err != nil {
				l.logf(log.LevelMessage, "failed to get rename params: %s", err)

				continue
			}

			if err = l.conn.Call(ctx, methodWorkspaceApplyEdit, types.ApplyWorkspaceAnyEditParams{
				Label: renameParams.Label,
				Edit: types.WorkspaceAnyEdit{
					DocumentChanges: append(edits, renameParams.Edit.DocumentChanges...),
				},
			}, nil); err != nil {
				l.logf(log.LevelMessage, "failed %s notify: %v", methodWorkspaceApplyEdit, err.Error())
			}

			// finally, trigger a diagnostics run for the new file
			updateEvent := lintFileJob{
				Reason: "internal/templateNewFile",
				URI:    job.URI,
			}

			l.lintFileJobs <- updateEvent
		}
	}
}

func (l *LanguageServer) StartWebServer(ctx context.Context) {
	l.webServer.Start(ctx)
}

func (l *LanguageServer) templateContentsForFile(fileURI string) (string, error) {
	// this function should not be called with files in the root, but if it is,
	// then it is an error to prevent unwanted behavior.
	if filepath.Dir(uri.ToPath(l.clientIdentifier, fileURI)) ==
		uri.ToPath(l.clientIdentifier, l.workspaceRootURI) {
		return "", errors.New("this function does not template files in the workspace root")
	}

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
		if len(diskContent) > 0 {
			return "", errors.New("file on disk already has contents, templating not allowed")
		}
	}

	path := uri.ToPath(l.clientIdentifier, fileURI)
	dir := filepath.Dir(path)

	roots, err := config.GetPotentialRoots(uri.ToPath(l.clientIdentifier, fileURI))
	if err != nil {
		return "", fmt.Errorf("failed to get potential roots during templating of new file: %w", err)
	}

	// handle the case where the root is unknown by providing the server's root
	// dir as a defacto root. This allows templating of files when there is no
	// known root, but the package could be determined based on the file path
	// relative to the server's workspace root
	if len(roots) == 1 && roots[0] == dir {
		roots = []string{uri.ToPath(l.clientIdentifier, l.workspaceRootURI)}
	}

	roots = append(roots, uri.ToPath(l.clientIdentifier, l.workspaceRootURI))

	longestPrefixRoot := ""

	for _, root := range roots {
		if strings.HasPrefix(dir, root) && len(root) > len(longestPrefixRoot) {
			longestPrefixRoot = root
		}
	}

	if longestPrefixRoot == "" {
		return "", fmt.Errorf("failed to find longest prefix root for templating of new file: %s", path)
	}

	parts := slices.Compact(strings.Split(strings.TrimPrefix(dir, longestPrefixRoot), rio.PathSeparator))

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
	pkg = cmp.Or(pkg, "main")

	if strings.HasSuffix(fileURI, "_test.rego") {
		pkg += "_test"
	}

	version := l.regoVersionForURI(fileURI)

	if version == ast.RegoV0 {
		return fmt.Sprintf("package %s\n\nimport rego.v1\n", pkg), nil
	}

	return fmt.Sprintf("package %s\n\n", pkg), nil
}

func (l *LanguageServer) fixEditParams(
	label string,
	fix fixes.Fix,
	args commandArgs,
) (bool, *types.ApplyWorkspaceEditParams, error) {
	oldContent, ok := l.cache.GetFileContents(args.Target)
	if !ok {
		return false, nil, fmt.Errorf("could not get file contents for uri %q", args.Target)
	}

	rto := &fixes.RuntimeOptions{
		BaseDir: uri.ToPath(l.clientIdentifier, l.workspaceRootURI),
	}
	if args.Diagnostic != nil {
		rto.Locations = []report.Location{
			{
				Row:    util.SafeUintToInt(args.Diagnostic.Range.Start.Line + 1),
				Column: util.SafeUintToInt(args.Diagnostic.Range.Start.Character + 1),
				End: &report.Position{
					Row:    util.SafeUintToInt(args.Diagnostic.Range.End.Line + 1),
					Column: util.SafeUintToInt(args.Diagnostic.Range.End.Character + 1),
				},
			},
		}
	}

	fixResults, err := fix.Fix(
		&fixes.FixCandidate{
			Filename: filepath.Base(uri.ToPath(l.clientIdentifier, args.Target)),
			Contents: oldContent,
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
					TextDocument: types.OptionalVersionedTextDocumentIdentifier{URI: args.Target},
					Edits:        ComputeEdits(oldContent, fixResults[0].Contents),
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
) (types.ApplyWorkspaceAnyEditParams, error) {
	var result types.ApplyWorkspaceAnyEditParams

	roots, err := config.GetPotentialRoots(l.workspacePath())
	if err != nil {
		return types.ApplyWorkspaceAnyEditParams{}, fmt.Errorf("failed to get potential roots: %w", err)
	}

	f := fixer.NewFixer()
	f.RegisterRoots(roots...)
	f.RegisterFixes(fix)
	// the default for the LSP is to rename on conflict
	f.SetOnConflictOperation(fixer.OnConflictRename)

	violations := []report.Violation{
		{
			Title: fix.Name(),
			Location: report.Location{
				File: uri.ToPath(l.clientIdentifier, fileURL),
			},
		},
	}

	cfp := fileprovider.NewCacheFileProvider(l.cache, l.clientIdentifier)

	fixReport, err := f.FixViolations(violations, cfp, l.getLoadedConfig())
	if err != nil {
		return result, fmt.Errorf("failed to fix violations: %w", err)
	}

	ff := fixReport.FixedFiles()

	if len(ff) == 0 {
		return types.ApplyWorkspaceAnyEditParams{
			Label: label,
			Edit:  types.WorkspaceAnyEdit{},
		}, nil
	}

	// find the new file and the old location
	var fixedFile, oldFile string

	var found bool

	for _, f := range ff {
		var ok bool

		oldFile, ok = fixReport.OldPathForFile(f)
		if ok {
			fixedFile = f
			found = true

			break
		}
	}

	if !found {
		return types.ApplyWorkspaceAnyEditParams{
			Label: label,
			Edit:  types.WorkspaceAnyEdit{},
		}, errors.New("failed to find fixed file's old location")
	}

	oldURI := uri.FromPath(l.clientIdentifier, oldFile)
	newURI := uri.FromPath(l.clientIdentifier, fixedFile)

	// are there old dirs?
	dirs, err := util.DirCleanUpPaths(
		uri.ToPath(l.clientIdentifier, oldURI),
		[]string{
			// stop at the root
			l.workspacePath(),
			// also preserve any dirs needed for the new file
			uri.ToPath(l.clientIdentifier, newURI),
		},
	)
	if err != nil {
		return types.ApplyWorkspaceAnyEditParams{}, fmt.Errorf("failed to determine empty directories post rename: %w", err)
	}

	changes := make([]any, 0, len(dirs)+1)
	changes = append(changes, types.RenameFile{
		Kind:   "rename",
		OldURI: oldURI,
		NewURI: newURI,
		Options: &types.RenameFileOptions{
			Overwrite:      false,
			IgnoreIfExists: false,
		},
	})

	for _, dir := range dirs {
		changes = append(
			changes,
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

	l.cache.Delete(oldURI)

	return types.ApplyWorkspaceAnyEditParams{
		Label: label,
		Edit: types.WorkspaceAnyEdit{
			DocumentChanges: changes,
		},
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

	if success, err := updateParse(ctx, l.cache, l.regoStore, fileURI, bis, l.regoVersionForURI(fileURI)); err != nil {
		return fmt.Errorf("failed to update parse: %w", err)
	} else if !success {
		return nil
	}

	if err := hover.UpdateBuiltinPositions(l.cache, fileURI, bis); err != nil {
		return fmt.Errorf("failed to update builtin positions: %w", err)
	}

	if err := hover.UpdateKeywordLocations(ctx, l.cache, fileURI); err != nil {
		return fmt.Errorf("failed to update keyword locations: %w", err)
	}

	return nil
}

func (l *LanguageServer) logf(level log.Level, format string, args ...any) {
	l.log(level, fmt.Sprintf(format, args...))
}

func (l *LanguageServer) log(level log.Level, message string) {
	if !l.logLevel.ShouldLog(level) {
		return
	}

	if l.logWriter != nil {
		fmt.Fprintln(l.logWriter, message)
	}
}

type HoverResponse struct {
	Contents types.MarkupContent `json:"contents"`
	Range    types.Range         `json:"range"`
}

func (l *LanguageServer) handleTextDocumentHover(params types.TextDocumentHoverParams) (any, error) {
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
		l.logf(log.LevelMessage, "could not get builtins for uri %q", params.TextDocument.URI)

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

func (l *LanguageServer) handleTextDocumentCodeAction(params types.CodeActionParams) (any, error) {
	if l.ignoreURI(params.TextDocument.URI) {
		return noCodeActions, nil
	}

	actions := []types.CodeAction{}

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
				IsPreferred: truePtr,
				Command:     FmtCommand(params.TextDocument.URI),
			})
		case ruleNameUseRegoV1:
			actions = append(actions, types.CodeAction{
				Title:       "Format for Rego v1 using opa fmt",
				Kind:        "quickfix",
				Diagnostics: []types.Diagnostic{diag},
				IsPreferred: truePtr,
				Command:     FmtV1Command(params.TextDocument.URI),
			})
		case ruleNameUseAssignmentOperator:
			actions = append(actions, types.CodeAction{
				Title:       "Replace = with := in assignment",
				Kind:        "quickfix",
				Diagnostics: []types.Diagnostic{diag},
				IsPreferred: truePtr,
				Command:     UseAssignmentOperatorCommand(params.TextDocument.URI, diag),
			})
		case ruleNameNoWhitespaceComment:
			actions = append(actions, types.CodeAction{
				Title:       "Format comment to have leading whitespace",
				Kind:        "quickfix",
				Diagnostics: []types.Diagnostic{diag},
				IsPreferred: truePtr,
				Command:     NoWhiteSpaceCommentCommand(params.TextDocument.URI, diag),
			})
		case ruleNameDirectoryPackageMismatch:
			actions = append(actions, types.CodeAction{
				Title:       "Move file so that directory structure mirrors package path",
				Kind:        "quickfix",
				Diagnostics: []types.Diagnostic{diag},
				IsPreferred: truePtr,
				Command:     DirectoryStructureMismatchCommand(params.TextDocument.URI, diag),
			})
		case ruleNameNonRawRegexPattern:
			actions = append(actions, types.CodeAction{
				Title:       "Replace \" with ` in regex pattern",
				Kind:        "quickfix",
				Diagnostics: []types.Diagnostic{diag},
				IsPreferred: truePtr,
				Command:     NonRawRegexPatternCommand(params.TextDocument.URI, diag),
			})
		}

		if l.clientIdentifier == clients.IdentifierVSCode {
			// always show the docs link
			txt := "Show documentation for " + diag.Code
			actions = append(actions, types.CodeAction{
				Title:       txt,
				Kind:        "quickfix",
				Diagnostics: []types.Diagnostic{diag},
				IsPreferred: truePtr,
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

func (l *LanguageServer) handleWorkspaceExecuteCommand(params types.ExecuteCommandParams) (any, error) {
	// this must not block, so we send the request to the worker on a buffered channel.
	// the response to the workspace/executeCommand request must be sent before the command is executed
	// so that the client can complete the request and be ready to receive the follow-on request for
	// workspace/applyEdit.
	l.commandRequest <- params

	// however, the contents of the response is not important
	return struct{}{}, nil
}

func (l *LanguageServer) handleTextDocumentInlayHint(params types.TextDocumentInlayHintParams) (any, error) {
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

	// TODO: use GetContentAndModule here, or do we need to handle the cases separately?
	// file is blank, nothing to do
	if contents, ok := l.cache.GetFileContents(params.TextDocument.URI); ok && contents == "" {
		return []types.InlayHint{}, nil
	}

	module, ok := l.cache.GetModule(params.TextDocument.URI)
	if !ok {
		l.logf(log.LevelMessage, "failed to get inlay hint: no parsed module for uri %q", params.TextDocument.URI)

		return []types.InlayHint{}, nil
	}

	inlayHints := getInlayHints(module, bis)

	return inlayHints, nil
}

func (l *LanguageServer) handleTextDocumentCodeLens(ctx context.Context, params types.CodeLensParams) (any, error) {
	lastSuccessfullyParsedLineCount, everParsed := l.cache.GetSuccessfulParseLineCount(params.TextDocument.URI)

	// if the file has always been unparsable, we can return early
	// as there is no value to be gained from showing code lenses.
	if !everParsed {
		return noCodeLenses, nil
	}

	parseErrors, ok := l.cache.GetParseErrors(params.TextDocument.URI)
	if ok && len(parseErrors) > 0 {
		// if there are parse errors, but the line count is the same, then we
		// still show them based on the last parsed module.
		contents, ok := l.cache.GetFileContents(params.TextDocument.URI)
		if !ok {
			// we have no contents for an unknown reason
			return noCodeLenses, nil
		}

		if len(strings.Split(contents, "\n")) != lastSuccessfullyParsedLineCount {
			return noCodeLenses, nil
		}
	}

	contents, module, ok := l.cache.GetContentAndModule(params.TextDocument.URI)
	if !ok {
		return nil, nil // return a null response, as per the spec
	}

	lenses, err := rego.CodeLenses(ctx, params.TextDocument.URI, contents, module)
	if err != nil {
		return nil, fmt.Errorf("failed to get code lenses: %w", err)
	}

	if l.clientInitializationOptions.EnableDebugCodelens != nil &&
		*l.clientInitializationOptions.EnableDebugCodelens {
		return lenses, nil
	}

	// filter out `regal.debug` codelens
	filteredLenses := make([]types.CodeLens, 0, len(lenses))

	for _, lens := range lenses {
		if lens.Command.Command != "regal.debug" {
			filteredLenses = append(filteredLenses, lens)
		}
	}

	return filteredLenses, nil
}

func (l *LanguageServer) handleTextDocumentCompletion(ctx context.Context, params types.CompletionParams) (any, error) {
	// when config ignores a file, then we return an empty completion list  as a no-op.
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
		RegoVersion:      l.regoVersionForURI(params.TextDocument.URI),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find completions: %w", err)
	}

	if items == nil {
		// make sure the items is always [] instead of null as is required by the spec
		items = noCompletionItems
	}

	return types.CompletionList{
		IsIncomplete: items != nil,
		Items:        items,
	}, nil
}

var noInlayHints = make([]types.InlayHint, 0)

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

	if firstErrorLine == 0 || firstErrorLine > uint(len(strings.Split(contents, "\n"))) {
		// if there are parse errors from line 0, we skip doing anything
		// if the last valid line is beyond the end of the file, we exit as something is up
		return noInlayHints
	}

	// select the lines from the contents up to the first parse error
	lines := strings.Join(strings.Split(contents, "\n")[:firstErrorLine], "\n")

	// parse the part of the module that might work
	module, err := rparse.Module(fileURI, lines)
	if err != nil {
		// if we still can't parse the bit we hoped was valid, we exit as this is 'too hard'
		return noInlayHints
	}

	return getInlayHints(module, builtins)
}

// Note: currently ignoring params.Query, as the client seems to do a good
// job of filtering anyway, and that would merely be an optimization here.
// But perhaps a good one to do at some point, and I'm not sure all clients
// do this filtering.
func (l *LanguageServer) handleWorkspaceSymbol() (any, error) {
	symbols := make([]types.WorkspaceSymbol, 0)
	contents := l.cache.GetAllFiles()
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

func (l *LanguageServer) handleTextDocumentDefinition(params types.DefinitionParams) (any, error) {
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

	query := oracle.DefinitionQuery{
		Filename: uri.ToPath(l.clientIdentifier, params.TextDocument.URI),
		Pos:      positionToOffset(contents, params.Position),
		Modules:  modules,
		Buffer:   outil.StringToByteSlice(contents),
	}

	definition, err := orc.FindDefinition(query)
	if err != nil {
		if errors.Is(err, oracle.ErrNoDefinitionFound) || errors.Is(err, oracle.ErrNoMatchFound) {
			// fail silently — the user could have clicked anywhere. return "null" as per the spec
			return nil, nil
		}

		l.logf(log.LevelMessage, "failed to find definition: %s", err)

		// return "null" as per the spec
		return nil, nil
	}

	//nolint:gosec
	loc := types.Location{
		URI: uri.FromPath(l.clientIdentifier, definition.Result.File),
		Range: types.Range{
			Start: types.Position{Line: uint(definition.Result.Row - 1), Character: uint(definition.Result.Col - 1)},
			End:   types.Position{Line: uint(definition.Result.Row - 1), Character: uint(definition.Result.Col - 1)},
		},
	}

	return loc, nil
}

func (l *LanguageServer) handleTextDocumentDidOpen(params types.TextDocumentDidOpenParams) (any, error) {
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

func (l *LanguageServer) handleTextDocumentDidClose(params types.TextDocumentDidCloseParams) (any, error) {
	// if the file being closed is ignored in config, then we
	// need to clear it from the ignored state in the cache.
	if l.ignoreURI(params.TextDocument.URI) {
		l.cache.Delete(params.TextDocument.URI)
	}

	return struct{}{}, nil
}

func (l *LanguageServer) handleTextDocumentDidChange(params types.TextDocumentDidChangeParams) (any, error) {
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
	params types.TextDocumentDidSaveParams,
) (any, error) {
	if params.Text != nil && l.getLoadedConfig() == nil {
		if !strings.Contains(*params.Text, "\r\n") {
			return struct{}{}, nil
		}

		cfg := l.getLoadedConfig()

		enabled, err := linter.NewLinter().WithUserConfig(*cfg).DetermineEnabledRules(ctx)
		if err != nil {
			l.logf(log.LevelMessage, "failed to determine enabled rules: %s", err)

			return struct{}{}, nil
		}

		formattingEnabled := slices.ContainsFunc(enabled, func(rule string) bool {
			return rule == ruleNameOPAFmt || rule == ruleNameUseRegoV1
		})

		if formattingEnabled {
			resp := types.ShowMessageParams{
				Type:    2, // warning
				Message: "CRLF line ending detected. Please change editor setting to use LF for line endings.",
			}

			if err := l.conn.Notify(ctx, "window/showMessage", resp); err != nil {
				l.logf(log.LevelMessage, "failed to notify: %s", err)

				return struct{}{}, nil
			}
		}
	}

	return struct{}{}, nil
}

func (l *LanguageServer) handleTextDocumentDocumentSymbol(params types.DocumentSymbolParams) (any, error) {
	if l.ignoreURI(params.TextDocument.URI) {
		return noDocumentSymbols, nil
	}

	contents, module, ok := l.cache.GetContentAndModule(params.TextDocument.URI)
	if !ok {
		l.logf(log.LevelMessage, "failed to get file contents for uri %q", params.TextDocument.URI)

		return noDocumentSymbols, nil
	}

	bis := l.builtinsForCurrentCapabilities()

	return documentSymbols(contents, module, bis), nil
}

func (l *LanguageServer) handleTextDocumentFoldingRange(params types.FoldingRangeParams) (any, error) {
	text, module, ok := l.cache.GetContentAndModule(params.TextDocument.URI)
	if !ok {
		return noFoldingRanges, nil
	}

	return findFoldingRanges(text, module), nil
}

func (l *LanguageServer) handleTextDocumentFormatting(
	ctx context.Context,
	params types.DocumentFormattingParams,
) (any, error) {
	var oldContent string

	// Fetch the contents used for formatting from the appropriate cache location.
	if l.ignoreURI(params.TextDocument.URI) {
		oldContent, _ = l.cache.GetIgnoredFileContents(params.TextDocument.URI)
	} else {
		oldContent, _ = l.cache.GetFileContents(params.TextDocument.URI)
	}

	// if the file is empty, then the formatters will fail, so we template instead
	if oldContent == "" {
		// disable the templating feature for files in the workspace root.
		if filepath.Dir(uri.ToPath(l.clientIdentifier, params.TextDocument.URI)) ==
			uri.ToPath(l.clientIdentifier, l.workspaceRootURI) {
			return []types.TextEdit{}, nil
		}

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

	// opa-fmt is the default formatter if not set in the client options
	formatter := "opa-fmt"

	if l.clientInitializationOptions.Formatter != nil {
		formatter = *l.clientInitializationOptions.Formatter
	}

	var newContent string

	switch formatter {
	case "opa-fmt", "opa-fmt-rego-v1":
		opts := format.Opts{
			RegoVersion: l.regoVersionForURI(params.TextDocument.URI),
		}

		if formatter == "opa-fmt-rego-v1" {
			opts.RegoVersion = ast.RegoV0CompatV1
		}

		f := &fixes.Fmt{OPAFmtOpts: opts}
		p := uri.ToPath(l.clientIdentifier, params.TextDocument.URI)

		fixResults, err := f.Fix(
			&fixes.FixCandidate{Filename: filepath.Base(p), Contents: oldContent},
			&fixes.RuntimeOptions{
				BaseDir: l.workspacePath(),
			},
		)
		if err != nil {
			l.logf(log.LevelMessage, "failed to format file: %s", err)

			// return "null" as per the spec
			return nil, nil
		}

		if len(fixResults) == 0 {
			return []types.TextEdit{}, nil
		}

		newContent = fixResults[0].Contents
	case "regal-fix":
		// set up an in-memory file provider to pass to the fixer for this one file
		memfp := fileprovider.NewInMemoryFileProvider(map[string]string{
			params.TextDocument.URI: oldContent,
		})

		input, err := memfp.ToInput(l.loadedConfigAllRegoVersions.Clone())
		if err != nil {
			return nil, fmt.Errorf("failed to create fixer input: %w", err)
		}

		f := fixer.NewFixer()
		f.RegisterFixes(fixes.NewDefaultFormatterFixes()...)

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

	return ComputeEdits(oldContent, newContent), nil
}

func (l *LanguageServer) handleWorkspaceDidCreateFiles(params types.WorkspaceDidCreateFilesParams) (any, error) {
	if l.ignoreURI(params.Files[0].URI) {
		return struct{}{}, nil
	}

	for _, createOp := range params.Files {
		if _, _, err := cache.UpdateCacheForURIFromDisk(
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
		l.templateFileJobs <- job
	}

	return struct{}{}, nil
}

func (l *LanguageServer) handleWorkspaceDidDeleteFiles(
	ctx context.Context,
	params types.WorkspaceDidDeleteFilesParams,
) (any, error) {
	if l.ignoreURI(params.Files[0].URI) {
		return struct{}{}, nil
	}

	for _, deleteOp := range params.Files {
		l.cache.Delete(deleteOp.URI)

		if err := l.sendFileDiagnostics(ctx, deleteOp.URI); err != nil {
			l.logf(log.LevelMessage, "failed to send diagnostic: %s", err)
		}
	}

	return struct{}{}, nil
}

func (l *LanguageServer) handleWorkspaceDidRenameFiles(
	ctx context.Context,
	params types.WorkspaceDidRenameFilesParams,
) (any, error) {
	for _, renameOp := range params.Files {
		if l.ignoreURI(renameOp.OldURI) && l.ignoreURI(renameOp.NewURI) {
			continue
		}

		var err error

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

		if err = l.sendFileDiagnostics(ctx, renameOp.OldURI); err != nil {
			l.logf(log.LevelMessage, "failed to send diagnostic: %s", err)
		}

		if l.ignoreURI(renameOp.NewURI) {
			continue
		}

		l.cache.SetFileContents(renameOp.NewURI, content)

		job := lintFileJob{
			Reason: "textDocument/didRename",
			URI:    renameOp.NewURI,
		}

		l.lintFileJobs <- job
		l.builtinsPositionJobs <- job
		// if the file being moved is empty, we template it too (if empty)
		l.templateFileJobs <- job
	}

	return struct{}{}, nil
}

func (l *LanguageServer) handleWorkspaceDiagnostic() (any, error) {
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
		wkspceDiags = noDiagnostics
	}

	workspaceReport.Items = append(workspaceReport.Items, types.WorkspaceFullDocumentDiagnosticReport{
		URI:     l.workspaceRootURI,
		Kind:    "full",
		Version: nil,
		Items:   wkspceDiags,
	})

	return workspaceReport, nil
}

func (l *LanguageServer) handleInitialize(ctx context.Context, params types.InitializeParams) (any, error) {
	l.clientIdentifier = clients.DetermineClientIdentifier(params.ClientInfo.Name)

	// params.RootURI is not expected to have a trailing slash, but if one is
	// present it will be removed for consistency.
	rootURI := strings.TrimSuffix(params.RootURI, rio.PathSeparator)

	if rootURI == "" {
		return nil, errors.New("rootURI was not set by the client but is required")
	}

	workspaceRootPath := uri.ToPath(l.clientIdentifier, rootURI)

	configRoots, err := lsconfig.FindConfigRoots(workspaceRootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find config roots: %w", err)
	}

	l.workspaceRootURI = rootURI

	switch {
	case len(configRoots) > 1:
		l.logf(
			log.LevelMessage,
			"warning: multiple configuration root directories found in workspace:\n%s\nusing %q as workspace root directory",
			strings.Join(configRoots, "\n"),
			configRoots[0],
		)

		l.workspaceRootURI = uri.FromPath(l.clientIdentifier, configRoots[0])
	case len(configRoots) == 1:
		l.logf(log.LevelMessage, "using %q as workspace root directory", configRoots[0])

		l.workspaceRootURI = uri.FromPath(l.clientIdentifier, configRoots[0])
	default:
		l.logf(
			log.LevelMessage,
			"using supplied workspace root directory: %q, config may be inherited from parent directory",
			workspaceRootPath,
		)
	}

	if l.clientIdentifier == clients.IdentifierGeneric {
		l.logf(
			log.LevelMessage,
			"unable to match client identifier for initializing client, using generic functionality: %s",
			params.ClientInfo.Name,
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
				WorkspaceFolders: types.WorkspaceFoldersServerCapabilities{
					// NOTE(anders): The language server protocol doesn't go into detail about what this is meant to
					// entail, and there's nothing else in the request/response payloads that carry workspace folder
					// information. The best source I've found on the this topic is this example repo from VS Code,
					// where they have the client start one instance of the server per workspace folder:
					// https://github.com/microsoft/vscode-extension-samples/tree/main/lsp-multi-server-sample
					// That seems like a reasonable approach to take, and means we won't have to deal with workspace
					// folders throughout the rest of the codebase. But the question then is — what is the point of
					// this capability, and what does it mean to say we support it? Clearly we don't in the server as
					// *there is no way* to support it here.
					Supported: true,
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
					"regal.fix.non-raw-regex-pattern",
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

	defaultConfig, err := config.LoadConfigWithDefaultsFromBundle(&rbundle.LoadedBundle, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}

	l.loadedConfigLock.Lock()
	l.loadedConfig = &defaultConfig
	l.loadedConfigLock.Unlock()

	err = l.loadEnabledRulesFromConfig(ctx, defaultConfig)
	if err != nil {
		l.logf(log.LevelMessage, "failed to cache enabled rules: %s", err)
	}

	if l.workspaceRootURI != "" {
		workspaceRootPath := l.workspacePath()

		l.bundleCache = bundles.NewCache(&bundles.CacheOptions{
			WorkspacePath: workspaceRootPath,
			ErrorLog:      l.logWriter,
		})

		configFile, err := config.FindConfig(workspaceRootPath)

		globalConfigDir := config.GlobalConfigDir(false)

		switch {
		case err == nil:
			l.logf(log.LevelMessage, "using config file: %s", configFile.Name())
			l.configWatcher.Watch(configFile.Name())
		case globalConfigDir != "":
			globalConfigFile := filepath.Join(globalConfigDir, "config.yaml")
			// the file might not exist and we only want to log we're using the
			// global file if it does.
			if _, err = os.Stat(globalConfigFile); err == nil {
				l.logf(log.LevelMessage, "using global config file: %s", globalConfigFile)
				l.configWatcher.Watch(globalConfigFile)
			}
		default:
			l.logf(log.LevelMessage, "no config file found for workspace: %s", err)
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

			if len(diskContents) == 0 && contents == "" {
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

		if _, err = updateParse(ctx, l.cache, l.regoStore, fileURI, bis, l.regoVersionForURI(fileURI)); err != nil {
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

func (l *LanguageServer) handleInitialized() (any, error) {
	// if running without config, then we should send the diagnostic request now
	// otherwise it'll happen when the config is loaded
	if !l.configWatcher.IsWatching() {
		l.lintWorkspaceJobs <- lintWorkspaceJob{Reason: "server initialized"}
	}

	return struct{}{}, nil
}

func (*LanguageServer) handleTextDocumentDiagnostic() (any, error) {
	// this is a no-op. Because we accept the textDocument/didChange event, which contains the new content,
	// we don't need to do anything here as once the new content has been parsed, the diagnostics will be sent
	// on the channel regardless of this request.
	return nil, nil
}

func (l *LanguageServer) handleWorkspaceDidChangeWatchedFiles(
	params types.WorkspaceDidChangeWatchedFilesParams,
) (any, error) {
	// this handles the case of a new config file being created when one did
	// not exist before
	if len(params.Changes) > 0 && (strings.HasSuffix(params.Changes[0].URI, ".regal/config.yaml") ||
		strings.HasSuffix(params.Changes[0].URI, ".regal.yaml")) {
		configFile, err := config.FindConfig(l.workspacePath())
		if err == nil {
			l.configWatcher.Watch(configFile.Name())
		}
	}

	// when a file is changed (saved), then we send trigger a full workspace lint
	regoFiles := make([]string, 0, len(params.Changes))

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
	// first, set the diagnostics for the file to the current parse errors
	fileDiags, _ := l.cache.GetParseErrors(fileURI)

	// if there are no parse errors, then we can check for lint errors
	if len(fileDiags) == 0 {
		fileDiags, _ = l.cache.GetFileDiagnostics(fileURI)
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
	var ignore []string

	if cfg := l.getLoadedConfig(); cfg != nil && cfg.Ignore.Files != nil {
		ignore = cfg.Ignore.Files
	}

	allModules := l.cache.GetAllModules()
	paths := outil.Keys(allModules)

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
	// TODO(charlieegan3): make this configurable for things like .rq etc?
	if !strings.HasSuffix(fileURI, ".rego") {
		return true
	}

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

	return err != nil || len(paths) == 0
}

func (l *LanguageServer) workspacePath() string {
	return uri.ToPath(l.clientIdentifier, l.workspaceRootURI)
}

func (l *LanguageServer) regoVersionForURI(fileURI string) ast.RegoVersion {
	version := ast.RegoUndefined
	if l.loadedConfigAllRegoVersions != nil {
		version = rules.RegoVersionFromVersionsMap(
			l.loadedConfigAllRegoVersions.Clone(),
			strings.TrimPrefix(uri.ToPath(l.clientIdentifier, fileURI), uri.ToPath(l.clientIdentifier, l.workspaceRootURI)),
			ast.RegoUndefined,
		)
	}

	return version
}

// builtinsForCurrentCapabilities returns the map of builtins for use
// in the server based on the currently loaded capabilities. If there is no
// config, then the default for the Regal OPA version is used.
func (l *LanguageServer) builtinsForCurrentCapabilities() map[string]*ast.Builtin {
	cfg := l.getLoadedConfig()
	if cfg == nil {
		return rego.BuiltinsForCapabilities(ast.CapabilitiesForThisVersion())
	}

	bis, ok := l.loadedBuiltins.Get(cfg.CapabilitiesURL)
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
