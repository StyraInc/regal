package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sourcegraph/jsonrpc2"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"

	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/uri"
	"github.com/styrainc/regal/pkg/config"
)

const (
	methodTextDocumentPublishDiagnostics = "textDocument/publishDiagnostics"
	methodWorkspaceApplyEdit             = "workspace/applyEdit"
)

type LanguageServerOptions struct {
	ErrorLog io.Writer
}

func NewLanguageServer(opts *LanguageServerOptions) *LanguageServer {
	ls := &LanguageServer{
		cache:                      NewCache(),
		errorLog:                   opts.ErrorLog,
		diagnosticRequestFile:      make(chan fileUpdateEvent, 10),
		diagnosticRequestWorkspace: make(chan string, 10),
		builtinsPositionFile:       make(chan fileUpdateEvent, 10),
		commandRequest:             make(chan ExecuteCommandParams, 10),
	}

	return ls
}

type LanguageServer struct {
	cache *Cache

	conn *jsonrpc2.Conn

	errorLog io.Writer

	loadedConfig     *config.Config
	loadedConfigLock sync.Mutex

	diagnosticRequestFile      chan fileUpdateEvent
	diagnosticRequestWorkspace chan string

	builtinsPositionFile chan fileUpdateEvent

	commandRequest chan ExecuteCommandParams

	clientRootURI    string
	clientIdentifier clients.Identifier

	// workspaceMode is set to true when the ls is initialized with
	// a clientRootURI.
	workspaceMode bool
}

// fileUpdateEvent is sent to a channel when an update is required for a file.
type fileUpdateEvent struct {
	Reason  string
	URI     string
	Content string
}

func (l *LanguageServer) Handle(
	ctx context.Context,
	conn *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	switch req.Method {
	case "initialize":
		return l.handleInitialize(ctx, conn, req)
	case "initialized":
		return l.handleInitialized(ctx, conn, req)
	case "textDocument/codeAction":
		return l.handleTextDocumentCodeAction(ctx, conn, req)
	case "textDocument/diagnostic":
		return l.handleTextDocumentDiagnostic(ctx, conn, req)
	case "textDocument/didOpen":
		return l.handleTextDocumentDidOpen(ctx, conn, req)
	case "textDocument/didClose":
		return struct{}{}, nil
	case "textDocument/didChange":
		return l.handleTextDocumentDidChange(ctx, conn, req)
	case "textDocument/hover":
		return l.handleTextDocumentHover(ctx, conn, req)
	case "textDocument/inlayHint":
		return l.handleTextDocumentInlayHint(ctx, conn, req)
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
	case "shutdown":
		err = conn.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close connection: %w", err)
		}

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
			err = updateFileDiagnostics(ctx, l.cache, l.loadedConfig, evt.URI, l.clientRootURI)
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
			err := updateAllDiagnostics(ctx, l.cache, l.loadedConfig, l.clientRootURI)
			if err != nil {
				l.logError(fmt.Errorf("failed to update aggregate diagnostics (trigger): %w", err))
			}

			// send diagnostics for all files
			for uri := range l.cache.GetAllFiles() {
				err = l.sendFileDiagnostics(ctx, uri)
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
			_, err := l.processBuiltinsUpdate(ctx, evt.URI, evt.Content)
			if err != nil {
				l.logError(fmt.Errorf("failed to process builtin positions update: %w", err))
			}
		}
	}
}

func getTargetURIFromParams(params ExecuteCommandParams) (string, error) {
	if len(params.Arguments) == 0 {
		return "", fmt.Errorf("expected at least one argument in command %v", params.Arguments)
	}

	target, ok := params.Arguments[0].(string)
	if !ok {
		return "", fmt.Errorf("expected argument to be a string in command %v", params.Command)
	}

	return target, nil
}

func (l *LanguageServer) formatToEdits(params ExecuteCommandParams, opts format.Opts) ([]TextEdit, string, error) {
	target, err := getTargetURIFromParams(params)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get target uri: %w", err)
	}

	oldContent, ok := l.cache.GetFileContents(target)
	if !ok {
		return nil, target, fmt.Errorf("could not get file contents for uri %q", target)
	}

	newContent, err := Format(uri.ToPath(l.clientIdentifier, target), oldContent, opts)
	if err != nil {
		return nil, target, fmt.Errorf("failed to format file: %w", err)
	}

	return ComputeEdits(oldContent, newContent), target, nil
}

func (l *LanguageServer) StartCommandWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case params := <-l.commandRequest:
			switch params.Command {
			case "regal.fmt":
				edits, target, err := l.formatToEdits(params, format.Opts{})
				if err != nil {
					l.logError(err)

					break
				}

				editParams := ApplyWorkspaceEditParams{
					Label: "Format using opa fmt",
					Edit: WorkspaceEdit{
						DocumentChanges: []TextDocumentEdit{
							{
								TextDocument: OptionalVersionedTextDocumentIdentifier{URI: target},
								Edits:        edits,
							},
						},
					},
				}

				// note, here conn.Call is used as the workspace/applyEdit message is a request, not a notification
				// as per the spec. In order to be 'routed' to the correct handler on the client it must have an ID
				// receive responses too.
				err = l.conn.Call(
					ctx,
					methodWorkspaceApplyEdit,
					editParams,
					nil, // however, the response content is not important
				)
				if err != nil {
					l.logError(fmt.Errorf("failed %s notify: %v", methodWorkspaceApplyEdit, err.Error()))
				}
			case "regal.fmt.v1":
				edits, target, err := l.formatToEdits(params, format.Opts{RegoVersion: ast.RegoV0CompatV1})
				if err != nil {
					l.logError(err)

					break
				}

				editParams := ApplyWorkspaceEditParams{
					Label: "Format for Rego v1 using opa fmt",
					Edit: WorkspaceEdit{
						DocumentChanges: []TextDocumentEdit{
							{
								TextDocument: OptionalVersionedTextDocumentIdentifier{URI: target},
								Edits:        edits,
							},
						},
					},
				}

				// note, here conn.Call is used as the workspace/applyEdit message is a request, not a notification
				// as per the spec. In order to be 'routed' to the correct handler on the client it must have an ID
				// receive responses too.
				err = l.conn.Call(
					ctx,
					methodWorkspaceApplyEdit,
					editParams,
					nil, // however, the response content is not important
				)
				if err != nil {
					l.logError(fmt.Errorf("failed %s notify: %v", methodWorkspaceApplyEdit, err.Error()))
				}
			}
		}
	}
}

// processTextContentUpdate updates the cache with the new content for the file at the given URI, attempts to parse the
// file, and returns whether the parse was successful. If it was not successful, the parse errors will be sent
// on the diagnostic channel.
func (l *LanguageServer) processTextContentUpdate(
	ctx context.Context,
	uri string,
	content string,
) (bool, error) {
	currentContent, ok := l.cache.GetFileContents(uri)
	if ok && currentContent == content {
		return false, nil
	}

	l.cache.SetFileContents(uri, content)

	success, err := updateParse(l.cache, uri)
	if err != nil {
		return false, fmt.Errorf("failed to update parse: %w", err)
	}

	if success {
		return true, nil
	}

	err = l.sendFileDiagnostics(ctx, uri)
	if err != nil {
		return false, fmt.Errorf("failed to send diagnostic: %w", err)
	}

	return false, nil
}

func (l *LanguageServer) processBuiltinsUpdate(
	_ context.Context,
	uri string,
	content string,
) (bool, error) {
	l.cache.SetFileContents(uri, content)

	success, err := updateParse(l.cache, uri)
	if err != nil {
		return false, fmt.Errorf("failed to update parse: %w", err)
	}

	if !success {
		return false, nil
	}

	err = updateBuiltinPositions(l.cache, uri)

	return err == nil, err
}

func (l *LanguageServer) logError(err error) {
	if l.errorLog != nil {
		fmt.Fprintf(l.errorLog, "ERROR: %s\n", err)
	}
}

type HoverResponse struct {
	Contents MarkupContent `json:"contents"`
	Range    Range         `json:"range"`
}

func (l *LanguageServer) handleTextDocumentHover(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params TextDocumentHoverParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	builtinsOnLine, ok := l.cache.GetBuiltinPositions(params.TextDocument.URI)
	// when no builtins are found, we can't return a useful hover response.
	// log the error, but return an empty struct to avoid an error being shown in the client.
	if !ok {
		l.logError(fmt.Errorf("could not get builtins for uri %q", params.TextDocument.URI))

		return struct{}{}, nil
	}

	for _, bp := range builtinsOnLine[params.Position.Line+1] {
		if params.Position.Character >= bp.Start-1 && params.Position.Character <= bp.End-1 {
			contents := createHoverContent(bp.Builtin)

			return HoverResponse{
				Contents: MarkupContent{
					Kind:  "markdown",
					Value: contents,
				},
				Range: Range{
					Start: Position{Line: bp.Line - 1, Character: bp.Start - 1},
					End:   Position{Line: bp.Line - 1, Character: bp.End - 1},
				},
			}, nil
		}
	}

	return struct{}{}, nil
}

func (l *LanguageServer) handleTextDocumentCodeAction(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params CodeActionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	actions := make([]CodeAction, 0)

	for _, diag := range params.Context.Diagnostics {
		switch diag.Code {
		case "opa-fmt":
			actions = append(actions, CodeAction{
				Title:       "Format using opa fmt",
				Kind:        "quickfix",
				Diagnostics: []Diagnostic{diag},
				IsPreferred: true,
				Command:     FmtCommand([]string{params.TextDocument.URI}),
			})
		case "use-rego-v1":
			actions = append(actions, CodeAction{
				Title:       "Format for Rego v1 using opa fmt",
				Kind:        "quickfix",
				Diagnostics: []Diagnostic{diag},
				IsPreferred: true,
				Command:     FmtV1Command([]string{params.TextDocument.URI}),
			})
		}

		if l.clientIdentifier == clients.IdentifierVSCode {
			// always show the docs link
			txt := "Show documentation for " + diag.Code
			actions = append(actions, CodeAction{
				Title:       txt,
				Kind:        "quickfix",
				Diagnostics: []Diagnostic{diag},
				IsPreferred: true,
				Command: Command{
					Title:     txt,
					Command:   "vscode.open",
					Tooltip:   txt,
					Arguments: []any{diag.CodeDescription.Href},
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
	var params ExecuteCommandParams
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
	var params TextDocumentInlayHintParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	module, ok := l.cache.GetModule(params.TextDocument.URI)
	if !ok {
		l.logError(fmt.Errorf("failed to get module for uri %q", params.TextDocument.URI))

		return []InlayHint{}, nil
	}

	inlayHints := getInlayHints(module)

	return inlayHints, nil
}

func (l *LanguageServer) handleTextDocumentDidOpen(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params TextDocumentDidOpenParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
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

func (l *LanguageServer) handleTextDocumentDidChange(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params TextDocumentDidChangeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
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

func (l *LanguageServer) handleWorkspaceDidCreateFiles(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params WorkspaceDidCreateFilesParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	for _, createOp := range params.Files {
		_, err = updateCacheForURIFromDisk(
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
	var params WorkspaceDidDeleteFilesParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	for _, deleteOp := range params.Files {
		l.cache.Delete(deleteOp.URI)

		evt := fileUpdateEvent{
			Reason: "textDocument/didDelete",
			URI:    deleteOp.URI,
		}

		l.diagnosticRequestFile <- evt
		l.builtinsPositionFile <- evt
	}

	return struct{}{}, nil
}

func (l *LanguageServer) handleWorkspaceDidRenameFiles(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params WorkspaceDidRenameFilesParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	for _, renameOp := range params.Files {
		content, err := updateCacheForURIFromDisk(
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
	workspaceReport := WorkspaceDiagnosticReport{
		Items: make([]WorkspaceFullDocumentDiagnosticReport, 0),
	}

	// if we're not in workspace mode, we can't show anything here
	// since we don't have the URI of the workspace from initialize
	if !l.workspaceMode {
		return workspaceReport, nil
	}

	workspaceReport.Items = append(workspaceReport.Items, WorkspaceFullDocumentDiagnosticReport{
		URI:     l.clientRootURI,
		Kind:    "full",
		Version: nil,
		Items:   l.cache.GetAllDiagnosticsForURI(l.clientRootURI),
	})

	return workspaceReport, nil
}

func (l *LanguageServer) handleInitialize(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	var params InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	l.clientRootURI = params.RootURI
	l.clientIdentifier = clients.DetermineClientIdentifier(params.ClientInfo.Name)

	if l.clientIdentifier == clients.IdentifierGeneric {
		l.logError(
			fmt.Errorf("unable to match client identifier for initializing client, using generic functionality: %s",
				params.ClientInfo.Name),
		)
	}

	regoFilter := FileOperationFilter{
		Scheme: "file",
		Pattern: FileOperationPattern{
			Glob: "**/*.rego",
		},
	}

	result = InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSyncOptions: TextDocumentSyncOptions{
				OpenClose: true,
				Change:    1, // TODO: write logic to use 2, for incremental updates
			},
			DiagnosticProvider: DiagnosticOptions{
				Identifier:            "rego",
				InterFileDependencies: true,
				WorkspaceDiagnostics:  true,
			},
			Workspace: WorkspaceOptions{
				FileOperations: FileOperationsServerCapabilities{
					DidCreate: FileOperationRegistrationOptions{
						Filters: []FileOperationFilter{regoFilter},
					},
					DidRename: FileOperationRegistrationOptions{
						Filters: []FileOperationFilter{regoFilter},
					},
					DidDelete: FileOperationRegistrationOptions{
						Filters: []FileOperationFilter{regoFilter},
					},
				},
			},
			InlayHintProvider: InlayHintOptions{
				ResolveProvider: false,
			},
			HoverProvider: true,
			CodeActionProvider: CodeActionOptions{
				CodeActionKinds: []string{"quickfix"},
			},
			ExecuteCommandProvider: ExecuteCommandOptions{
				Commands: []string{"regal.fmt", "regal.fmt.v1"},
			},
		},
	}

	if l.clientRootURI != "" {
		l.workspaceMode = true

		err = l.loadWorkspaceContents()
		if err != nil {
			return nil, fmt.Errorf("failed to load workspace contents: %w", err)
		}

		// attempt to load the config as it is found on disk
		file, err := config.FindConfig(strings.TrimPrefix(l.clientRootURI, "file://"))
		if err == nil {
			l.reloadConfig(file, false)
		}
	}

	return result, nil
}

func (l *LanguageServer) loadWorkspaceContents() error {
	workspaceRootPath := uri.ToPath(l.clientIdentifier, l.clientRootURI)

	err := filepath.WalkDir(workspaceRootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk workspace dir %q: %w", path, err)
		}

		// TODO(charlieegan3): make this configurable for things like .rq etc?
		if d.IsDir() || !strings.HasSuffix(path, ".rego") {
			return nil
		}

		fileURI := uri.FromPath(l.clientIdentifier, path)

		_, err = updateCacheForURIFromDisk(l.cache, fileURI, path)
		if err != nil {
			return fmt.Errorf("failed to update cache for uri %q: %w", path, err)
		}

		_, err = updateParse(l.cache, fileURI)
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

func (l *LanguageServer) reloadConfig(configReader io.Reader, runWorkspaceDiagnostics bool) {
	l.loadedConfigLock.Lock()
	defer l.loadedConfigLock.Unlock()

	var loadedConfig config.Config

	err := yaml.NewDecoder(configReader).Decode(&loadedConfig)
	if err != nil && !errors.Is(err, io.EOF) {
		l.logError(fmt.Errorf("failed to reload config: %w", err))

		return
	}

	// if the config is now blank, then we need to clear it
	if errors.Is(err, io.EOF) {
		l.loadedConfig = nil
	} else {
		l.loadedConfig = &loadedConfig
	}

	// this can be set to false by callers to disable the running of diagnostics for the whole workspace.
	// this is intended to be used at start up when a workspace run is already going to be taking place.
	if runWorkspaceDiagnostics {
		l.diagnosticRequestWorkspace <- "config file changed"
	}
}

func (l *LanguageServer) handleInitialized(
	_ context.Context,
	_ *jsonrpc2.Conn,
	_ *jsonrpc2.Request,
) (result any, err error) {
	l.diagnosticRequestWorkspace <- "initialized"

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
	return struct{}{}, nil
}

func (l *LanguageServer) handleWorkspaceDidChangeWatchedFiles(
	_ context.Context,
	_ *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) (result any, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params WorkspaceDidChangeWatchedFilesParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	// when a file is changed (saved), then we send trigger a full workspace lint
	regoFiles := make([]string, 0)

	for _, change := range params.Changes {
		if change.URI == "" {
			continue
		}

		if strings.HasSuffix(change.URI, "/.regal/config.yaml") {
			// attempt to load the config as it is found on disk
			file, err := os.Open(strings.TrimPrefix(change.URI, "file://"))
			if err == nil {
				l.reloadConfig(file, true)
			}

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

func (l *LanguageServer) sendFileDiagnostics(ctx context.Context, uri string) error {
	resp := FileDiagnostics{
		Items: l.cache.GetAllDiagnosticsForURI(uri),
		URI:   uri,
	}

	err := l.conn.Notify(ctx, methodTextDocumentPublishDiagnostics, resp)
	if err != nil {
		return fmt.Errorf("failed to notify: %w", err)
	}

	return nil
}
