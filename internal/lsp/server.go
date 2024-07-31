//nolint:nilerr,nilnil
package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/jsonrpc2"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"
	"github.com/open-policy-agent/opa/storage"

	"github.com/styrainc/regal/bundle"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/commands"
	"github.com/styrainc/regal/internal/lsp/completions"
	"github.com/styrainc/regal/internal/lsp/completions/providers"
	lsconfig "github.com/styrainc/regal/internal/lsp/config"
	"github.com/styrainc/regal/internal/lsp/examples"
	"github.com/styrainc/regal/internal/lsp/hover"
	"github.com/styrainc/regal/internal/lsp/opa/oracle"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/uri"
	"github.com/styrainc/regal/internal/lsp/workspace"
	rparse "github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/update"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/fixer/fixes"
	"github.com/styrainc/regal/pkg/linter"
	"github.com/styrainc/regal/pkg/version"
)

const (
	methodTextDocumentPublishDiagnostics = "textDocument/publishDiagnostics"
	methodWorkspaceApplyEdit             = "workspace/applyEdit"

	ruleNameOPAFmt    = "opa-fmt"
	ruleNameUseRegoV1 = "use-rego-v1"
)

type LanguageServerOptions struct {
	ErrorLog io.Writer
}

func NewLanguageServer(opts *LanguageServerOptions) *LanguageServer {
	c := cache.NewCache()
	store := NewRegalStore()

	ls := &LanguageServer{
		cache:                      c,
		regoStore:                  store,
		errorLog:                   opts.ErrorLog,
		diagnosticRequestFile:      make(chan fileUpdateEvent, 10),
		diagnosticRequestWorkspace: make(chan string, 10),
		builtinsPositionFile:       make(chan fileUpdateEvent, 10),
		commandRequest:             make(chan types.ExecuteCommandParams, 10),
		configWatcher:              lsconfig.NewWatcher(&lsconfig.WatcherOpts{ErrorWriter: opts.ErrorLog}),
		completionsManager:         completions.NewDefaultManager(c, store),
	}

	return ls
}

type LanguageServer struct {
	conn *jsonrpc2.Conn

	errorLog io.Writer

	configWatcher    *lsconfig.Watcher
	loadedConfig     *config.Config
	loadedConfigLock sync.Mutex

	workspaceDirWatcher *workspace.DirWatcher

	workspaceRootURI string
	clientIdentifier clients.Identifier

	cache     *cache.Cache
	regoStore storage.Store

	completionsManager *completions.Manager

	diagnosticRequestFile      chan fileUpdateEvent
	diagnosticRequestWorkspace chan string
	builtinsPositionFile       chan fileUpdateEvent
	commandRequest             chan types.ExecuteCommandParams
}

// fileUpdateEvent is sent to a channel when an update is required for a file.
type fileUpdateEvent struct {
	Reason  string
	URI     string
	OldURI  string
	Content string
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
		err := l.conn.Close()
		if err != nil {
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
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-l.diagnosticRequestFile:
			// if file has been deleted, clear diagnostics in the client
			if evt.Reason == "textDocument/didDelete" {
				err := l.sendFileDiagnostics(ctx, evt.URI)
				if err != nil {
					l.logError(fmt.Errorf("failed to send diagnostic: %w", err))
				}

				continue
			}

			if evt.Reason == "textDocument/didRename" && evt.OldURI != "" {
				err := l.sendFileDiagnostics(ctx, evt.OldURI)
				if err != nil {
					l.logError(fmt.Errorf("failed to send diagnostic: %w", err))
				}

				continue
			}

			// if there is new content, we need to update the parse errors or module first
			success, err := l.processTextContentUpdate(ctx, evt.URI, evt.Content)
			if err != nil {
				l.logError(fmt.Errorf("failed to process text content update: %w", err))

				continue
			}

			if !success {
				continue
			}

			// otherwise, lint the file and send the diagnostics
			err = updateFileDiagnostics(ctx, l.cache, l.loadedConfig, evt.URI, l.workspaceRootURI)
			if err != nil {
				l.logError(fmt.Errorf("failed to update file diagnostics: %w", err))
			}

			err = l.sendFileDiagnostics(ctx, evt.URI)
			if err != nil {
				l.logError(fmt.Errorf("failed to send diagnostic: %w", err))
			}

			// if the file has agg diagnostics, we trigger a run for the workspace as by changing this file,
			// these may now be out of date
			aggDiags, ok := l.cache.GetAggregateDiagnostics(evt.URI)
			if ok && len(aggDiags) > 0 {
				l.diagnosticRequestWorkspace <- fmt.Sprintf("file %q with aggregate violation changed", evt.URI)
			}
		case <-l.diagnosticRequestWorkspace:
			// results will be sent in response to the next workspace/diagnostics request
			err := updateAllDiagnostics(ctx, l.cache, l.loadedConfig, l.workspaceRootURI)
			if err != nil {
				l.logError(fmt.Errorf("failed to update aggregate diagnostics (trigger): %w", err))
			}

			// send diagnostics for all files
			for fileURI := range l.cache.GetAllFiles() {
				err = l.sendFileDiagnostics(ctx, fileURI)
				if err != nil {
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
		case evt := <-l.builtinsPositionFile:
			err := l.processHoverContentUpdate(ctx, evt.URI, evt.Content)
			if err != nil {
				l.logError(fmt.Errorf("failed to process builtin positions update: %w", err))
			}
		}
	}
}

func (l *LanguageServer) StartConfigWorker(ctx context.Context) {
	err := l.configWatcher.Start(ctx)
	if err != nil {
		l.logError(fmt.Errorf("failed to start config watcher: %w", err))

		return
	}

	regalRules, err := rio.LoadRegalBundleFS(bundle.Bundle)
	if err != nil {
		l.logError(fmt.Errorf("failed to load regal bundle for defaulting of user config: %w", err))

		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case path := <-l.configWatcher.Reload:
			configFile, err := os.Open(path)
			if err != nil {
				l.logError(fmt.Errorf("failed to open config file: %w", err))

				continue
			}

			var userConfig config.Config

			err = yaml.NewDecoder(configFile).Decode(&userConfig)
			if err != nil && !errors.Is(err, io.EOF) {
				l.logError(fmt.Errorf("failed to reload config: %w", err))

				return
			}

			mergedConfig, err := config.LoadConfigWithDefaultsFromBundle(&regalRules, &userConfig)
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

				err := RemoveFileMod(ctx, l.regoStore, k)
				if err != nil {
					l.logError(fmt.Errorf("failed to remove mod from store: %w", err))
				}
			}

			// when a file is 'unignored', we move it's contents to the
			// standard file list if missing
			for k, v := range l.cache.GetAllIgnoredFiles() {
				if l.ignoreURI(k) {
					continue
				}

				// ignored contents will only be used when there is no existing content
				_, ok := l.cache.GetFileContents(k)
				if !ok {
					l.cache.SetFileContents(k, v)

					// updating the parse here will enable things like go-to defn
					// to start working right away without the need for a file content
					// update to run updateParse.
					_, err = updateParse(ctx, l.cache, l.regoStore, k)
					if err != nil {
						l.logError(fmt.Errorf("failed to update parse for previously ignored file %q: %w", k, err))
					}
				}

				l.cache.ClearIgnoredFileContents(k)
			}

			//nolint:contextcheck
			go func() {
				if l.loadedConfig.Features.Remote.CheckVersion &&
					os.Getenv(update.CheckVersionDisableEnvVar) != "" {
					update.CheckAndWarn(update.Options{
						CurrentVersion: version.Version,
						CurrentTime:    time.Now().UTC(),
						Debug:          false,
						StateDir:       config.GlobalDir(),
					}, os.Stderr)
				}
			}()

			l.diagnosticRequestWorkspace <- "config file changed"
		case <-l.configWatcher.Drop:
			l.loadedConfigLock.Lock()
			l.loadedConfig = nil
			l.loadedConfigLock.Unlock()

			l.diagnosticRequestWorkspace <- "config file dropped"
		}
	}
}

func (l *LanguageServer) StartWorkspaceStateWorker(ctx context.Context) {
	ready := make(chan struct{})
	go func() {
		for l.workspaceRootURI == "" {
			time.Sleep(300 * time.Millisecond)
		}
		ready <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		return
	case <-ready:
	}

	l.logError(fmt.Errorf("starting workspace state worker"))

	dw, err := workspace.NewDirWatcher(&workspace.DirWatcherOpts{
		ErrorLog:        l.errorLog,
		PollingInterval: 500 * time.Millisecond,
		RootPath:        uri.ToPath(l.clientIdentifier, l.workspaceRootURI),
	})
	if err != nil {
		l.logError(fmt.Errorf("failed to create dir watcher: %w", err))
	}

	l.workspaceDirWatcher = dw

	l.workspaceDirWatcher.Start(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case path := <-dw.CreateChan:
			l.logError(fmt.Errorf("file created: %s", path))
		case path := <-dw.UpdateChan:
			l.logError(fmt.Errorf("file updated: %s", path))
		case path := <-dw.RemoveChan:
			l.logError(fmt.Errorf("file removed: %s", path))
		}
	}
}

func (l *LanguageServer) StartCommandWorker(ctx context.Context) {
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
			}

			if err != nil {
				l.logError(err)

				break
			}

			if fixed {
				err = l.conn.Call(ctx, methodWorkspaceApplyEdit, editParams, nil)
				if err != nil {
					l.logError(fmt.Errorf("failed %s notify: %v", methodWorkspaceApplyEdit, err.Error()))
				}
			}
		}
	}
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

	var rto *fixes.RuntimeOptions
	if pr.Location != nil {
		rto = &fixes.RuntimeOptions{Locations: []ast.Location{*pr.Location}}
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

// processTextContentUpdate updates the cache with the new content for the file at the given URI, attempts to parse the
// file, and returns whether the parse was successful. If it was not successful, the parse errors will be sent
// on the diagnostic channel.
func (l *LanguageServer) processTextContentUpdate(
	ctx context.Context,
	fileURI string,
	content string,
) (bool, error) {
	if l.ignoreURI(fileURI) {
		return false, nil
	}

	currentContent, ok := l.cache.GetFileContents(fileURI)
	if ok && currentContent == content {
		return false, nil
	}

	l.cache.SetFileContents(fileURI, content)

	success, err := updateParse(ctx, l.cache, l.regoStore, fileURI)
	if err != nil {
		return false, fmt.Errorf("failed to update parse: %w", err)
	}

	if success {
		return true, nil
	}

	err = l.sendFileDiagnostics(ctx, fileURI)
	if err != nil {
		return false, fmt.Errorf("failed to send diagnostic: %w", err)
	}

	return false, nil
}

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

	success, err := updateParse(ctx, l.cache, l.regoStore, fileURI)
	if err != nil {
		return fmt.Errorf("failed to update parse: %w", err)
	}

	if !success {
		return nil
	}

	err = hover.UpdateBuiltinPositions(l.cache, fileURI)
	if err != nil {
		return fmt.Errorf("failed to update builtin positions: %w", err)
	}

	err = hover.UpdateKeywordLocations(ctx, l.cache, fileURI)
	if err != nil {
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
	// when no keywords are found, we can't return a useful hover response.
	if !ok {
		l.logError(fmt.Errorf("could not get keywords for uri %q", params.TextDocument.URI))

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
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	yes := true
	actions := make([]types.CodeAction, 0)

	if l.ignoreURI(params.TextDocument.URI) {
		return actions, nil
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
		case "use-assignment-operator":
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
		case "no-whitespace-comment":
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
	if err := json.Unmarshal(*req.Params, &params); err != nil {
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
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	if l.ignoreURI(params.TextDocument.URI) {
		return []types.InlayHint{}, nil
	}

	// when a file cannot be parsed, we do a best effort attempt to provide inlay hints
	// by finding the location of the first parse error and attempting to parse up to that point
	parseErrors, ok := l.cache.GetParseErrors(params.TextDocument.URI)
	if ok && len(parseErrors) > 0 {
		contents, ok := l.cache.GetFileContents(params.TextDocument.URI)
		if !ok {
			// if there is no content, we can't even do a partial parse
			return []types.InlayHint{}, nil
		}

		return partialInlayHints(parseErrors, contents, params.TextDocument.URI), nil
	}

	module, ok := l.cache.GetModule(params.TextDocument.URI)
	if !ok {
		l.logError(fmt.Errorf("failed to get inlay hint: no parsed module for uri %q", params.TextDocument.URI))

		return []types.InlayHint{}, nil
	}

	inlayHints := getInlayHints(module)

	return inlayHints, nil
}

func (l *LanguageServer) handleTextDocumentCompletion(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (any, error) {
	var params types.CompletionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
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
	items, err := l.completionsManager.Run(params, &providers.Options{
		ClientIdentifier: l.clientIdentifier,
		RootURI:          l.workspaceRootURI,
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

func partialInlayHints(parseErrors []types.Diagnostic, contents, fileURI string) []types.InlayHint {
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

	return getInlayHints(module)
}

func (l *LanguageServer) handleWorkspaceSymbol(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.WorkspaceSymbolParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	symbols := make([]types.WorkspaceSymbol, 0)
	contents := l.cache.GetAllFiles()

	// Note: currently ignoring params.Query, as the client seems to do a good
	// job of filtering anyway, and that would merely be an optimization here.
	// But perhaps a good one to do at some point, and I'm not sure all clients
	// do this filtering.

	for moduleURL, module := range l.cache.GetAllModules() {
		content := contents[moduleURL]
		docSyms := documentSymbols(content, module)
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
	if err := json.Unmarshal(*req.Params, &params); err != nil {
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
	if err := json.Unmarshal(*req.Params, &params); err != nil {
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

	evt := fileUpdateEvent{
		Reason:  "textDocument/didOpen",
		URI:     params.TextDocument.URI,
		Content: params.TextDocument.Text,
	}

	l.diagnosticRequestFile <- evt

	l.builtinsPositionFile <- evt

	return struct{}{}, nil
}

func (l *LanguageServer) handleTextDocumentDidClose(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.TextDocumentDidCloseParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
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
	if err := json.Unmarshal(*req.Params, &params); err != nil {
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

	evt := fileUpdateEvent{
		Reason:  "textDocument/didChange",
		URI:     params.TextDocument.URI,
		Content: params.ContentChanges[0].Text,
	}

	l.diagnosticRequestFile <- evt
	l.builtinsPositionFile <- evt

	return struct{}{}, nil
}

func (l *LanguageServer) handleTextDocumentDidSave(
	ctx context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.TextDocumentDidSaveParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	if params.Text != nil && l.loadedConfig != nil {
		if !strings.Contains(*params.Text, "\r\n") {
			return struct{}{}, nil
		}

		enabled, err := linter.NewLinter().WithUserConfig(*l.loadedConfig).DetermineEnabledRules(ctx)
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

			err := l.conn.Notify(ctx, "window/showMessage", resp)
			if err != nil {
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
	if err := json.Unmarshal(*req.Params, &params); err != nil {
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
		l.logError(fmt.Errorf("failed to get module for uri %q", params.TextDocument.URI))

		return []types.DocumentSymbol{}, nil
	}

	return documentSymbols(contents, module), nil
}

func (l *LanguageServer) handleTextDocumentFoldingRange(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.FoldingRangeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
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
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.DocumentFormattingParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	if warnings := validateFormattingOptions(params.Options); len(warnings) > 0 {
		l.logError(fmt.Errorf("formatting params validation warnings: %v", warnings))
	}

	var oldContent string

	var ok bool

	// Fetch the contents used for formatting from the appropriate cache location.
	if l.ignoreURI(params.TextDocument.URI) {
		oldContent, ok = l.cache.GetIgnoredFileContents(params.TextDocument.URI)
	} else {
		oldContent, ok = l.cache.GetFileContents(params.TextDocument.URI)
	}

	if !ok {
		return nil, fmt.Errorf("failed to get file contents for uri %q", params.TextDocument.URI)
	}

	f := &fixes.Fmt{OPAFmtOpts: format.Opts{}}

	fixResults, err := f.Fix(&fixes.FixCandidate{
		Filename: filepath.Base(uri.ToPath(l.clientIdentifier, params.TextDocument.URI)),
		Contents: []byte(oldContent),
	}, nil)
	if err != nil {
		l.logError(fmt.Errorf("failed to format file: %w", err))

		// return "null" as per the spec
		return nil, nil
	}

	if len(fixResults) == 0 {
		return []types.TextEdit{}, nil
	}

	return ComputeEdits(oldContent, string(fixResults[0].Contents)), nil
}

func (l *LanguageServer) handleWorkspaceDidCreateFiles(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.WorkspaceDidCreateFilesParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	if l.ignoreURI(params.Files[0].URI) {
		return struct{}{}, nil
	}

	for _, createOp := range params.Files {
		_, err = cache.UpdateCacheForURIFromDisk(
			l.cache,
			uri.FromPath(l.clientIdentifier, createOp.URI),
			uri.ToPath(l.clientIdentifier, createOp.URI),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update cache for uri %q: %w", createOp.URI, err)
		}

		evt := fileUpdateEvent{
			Reason: "textDocument/didCreate",
			URI:    createOp.URI,
		}

		l.diagnosticRequestFile <- evt
		l.builtinsPositionFile <- evt
	}

	return struct{}{}, nil
}

func (l *LanguageServer) handleWorkspaceDidDeleteFiles(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.WorkspaceDidDeleteFilesParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	if l.ignoreURI(params.Files[0].URI) {
		return struct{}{}, nil
	}

	for _, deleteOp := range params.Files {
		l.cache.Delete(deleteOp.URI)

		evt := fileUpdateEvent{
			Reason: "textDocument/didDelete",
			URI:    deleteOp.URI,
		}

		l.diagnosticRequestFile <- evt
	}

	return struct{}{}, nil
}

func (l *LanguageServer) handleWorkspaceDidRenameFiles(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.WorkspaceDidRenameFilesParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	for _, renameOp := range params.Files {
		if l.ignoreURI(renameOp.OldURI) && l.ignoreURI(renameOp.NewURI) {
			continue
		}

		content, err := cache.UpdateCacheForURIFromDisk(
			l.cache,
			uri.FromPath(l.clientIdentifier, renameOp.NewURI),
			uri.ToPath(l.clientIdentifier, renameOp.NewURI),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update cache for uri %q: %w", renameOp.NewURI, err)
		}

		l.cache.Delete(renameOp.OldURI)

		evt := fileUpdateEvent{
			Reason:  "textDocument/didRename",
			URI:     renameOp.NewURI,
			OldURI:  renameOp.OldURI,
			Content: content,
		}

		l.diagnosticRequestFile <- evt
		l.builtinsPositionFile <- evt
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

	workspaceReport.Items = append(workspaceReport.Items, types.WorkspaceFullDocumentDiagnosticReport{
		URI:     l.workspaceRootURI,
		Kind:    "full",
		Version: nil,
		Items:   l.cache.GetAllDiagnosticsForURI(l.workspaceRootURI),
	})

	return workspaceReport, nil
}

func (l *LanguageServer) handleInitialize(
	ctx context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params types.InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	l.workspaceRootURI = params.RootURI
	l.clientIdentifier = clients.DetermineClientIdentifier(params.ClientInfo.Name)

	if l.clientIdentifier == clients.IdentifierGeneric {
		l.logError(
			fmt.Errorf("unable to match client identifier for initializing client, using generic functionality: %s",
				params.ClientInfo.Name),
		)
	}

	regoFilter := types.FileOperationFilter{
		Scheme: "file",
		Pattern: types.FileOperationPattern{
			Glob: "**/*.rego",
		},
	}

	result = types.InitializeResult{
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
				CodeActionKinds: []string{"quickfix"},
			},
			ExecuteCommandProvider: types.ExecuteCommandOptions{
				Commands: []string{
					"regal.fix.opa-fmt",
					"regal.fix.use-rego-v1",
					"regal.fix.use-assignment-operator",
					"regal.fix.no-whitespace-comment",
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
		},
	}

	if l.workspaceRootURI != "" {
		configFile, err := config.FindConfig(uri.ToPath(l.clientIdentifier, l.workspaceRootURI))
		if err == nil {
			l.configWatcher.Watch(configFile.Name())
		}

		err = l.loadWorkspaceContents(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load workspace contents: %w", err)
		}

		l.diagnosticRequestWorkspace <- "server initialize"
	}

	return result, nil
}

func (l *LanguageServer) loadWorkspaceContents(ctx context.Context) error {
	workspaceRootPath := uri.ToPath(l.clientIdentifier, l.workspaceRootURI)

	err := filepath.WalkDir(workspaceRootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk workspace dir %q: %w", path, err)
		}

		// TODO(charlieegan3): make this configurable for things like .rq etc?
		if d.IsDir() || !strings.HasSuffix(path, ".rego") {
			return nil
		}

		fileURI := uri.FromPath(l.clientIdentifier, path)

		if l.ignoreURI(fileURI) {
			return nil
		}

		_, err = cache.UpdateCacheForURIFromDisk(l.cache, fileURI, path)
		if err != nil {
			return fmt.Errorf("failed to update cache for uri %q: %w", path, err)
		}

		_, err = updateParse(ctx, l.cache, l.regoStore, fileURI)
		if err != nil {
			return fmt.Errorf("failed to update parse: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk workspace dir %q: %w", workspaceRootPath, err)
	}

	return nil
}

func (l *LanguageServer) handleInitialized(
	_ context.Context,
	_ *jsonrpc2.Conn,
	_ *jsonrpc2.Request,
) (result any, err error) {
	// if running without config, then we should send the diagnostic request now
	// otherwise it'll happen when the config is loaded
	if !l.configWatcher.IsWatching() {
		l.diagnosticRequestWorkspace <- "initialized"
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
	if err := json.Unmarshal(*req.Params, &params); err != nil {
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
		l.diagnosticRequestWorkspace <- fmt.Sprintf(
			"workspace/didChangeWatchedFiles (%s)", strings.Join(regoFiles, ", "))
	}

	return struct{}{}, nil
}

func (l *LanguageServer) sendFileDiagnostics(ctx context.Context, fileURI string) error {
	resp := types.FileDiagnostics{
		Items: l.cache.GetAllDiagnosticsForURI(fileURI),
		URI:   fileURI,
	}

	err := l.conn.Notify(ctx, methodTextDocumentPublishDiagnostics, resp)
	if err != nil {
		return fmt.Errorf("failed to notify: %w", err)
	}

	return nil
}

func (l *LanguageServer) getFilteredModules() (map[string]*ast.Module, error) {
	ignore := make([]string, 0)

	if l.loadedConfig != nil && l.loadedConfig.Ignore.Files != nil {
		ignore = l.loadedConfig.Ignore.Files
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
	if l.loadedConfig == nil {
		return false
	}

	paths, err := config.FilterIgnoredPaths(
		[]string{uri.ToPath(l.clientIdentifier, fileURI)},
		l.loadedConfig.Ignore.Files,
		false,
		uri.ToPath(l.clientIdentifier, l.workspaceRootURI),
	)
	if err != nil || len(paths) == 0 {
		return true
	}

	return false
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

		if i == int(p.Line)-1 {
			return bytesRead + int(p.Character)
		}
	}

	return -1
}
