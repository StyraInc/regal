package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sourcegraph/jsonrpc2"
	"gopkg.in/yaml.v3"

	"github.com/styrainc/regal/pkg/config"
)

type LanguageServerOptions struct {
	ErrorLog       *os.File
	VerboseLogging bool
}

func NewLanguageServer(opts *LanguageServerOptions) *LanguageServer {
	ls := &LanguageServer{
		cache:                      NewCache(),
		errorLog:                   opts.ErrorLog,
		diagnosticRequestFile:      make(chan fileDiagnosticRequiredEvent, 10),
		diagnosticRequestWorkspace: make(chan string, 10),
		verboseLogging:             opts.VerboseLogging,
	}

	return ls
}

type LanguageServer struct {
	cache *Cache

	conn *jsonrpc2.Conn

	errorLog       *os.File
	verboseLogging bool

	loadedConfig     *config.Config
	loadedConfigLock sync.Mutex

	diagnosticRequestFile      chan fileDiagnosticRequiredEvent
	diagnosticRequestWorkspace chan string

	clientRootURI string
}

func (l *LanguageServer) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	l.logInboundMessage(req.Method, string(*req.Params))

	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	switch req.Method {
	case "initialize":
		return l.handleInitialize(ctx, conn, req)
	case "initialized":
		return l.handleInitialized(ctx, conn, req)
	case "textDocument/diagnostic":
		return l.handleTextDocumentDiagnostic(ctx, conn, req)
	case "textDocument/didOpen":
		return l.handleTextDocumentDidOpen(ctx, conn, req)
	case "textDocument/didClose":
		return nil, nil
	case "textDocument/didChange":
		return l.handleTextDocumentDidChange(ctx, conn, req)
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
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
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
			case evt := <-l.diagnosticRequestFile:
				l.log(fmt.Sprintf("diagnostic request for %s: %q", evt.URI, evt.Reason))

				// if there is new content, we need to update the parse errors or module first
				if evt.Content != "" {
					currentContent, ok := l.cache.GetFileContents(evt.URI)
					if ok && currentContent == evt.Content {
						continue
					}

					l.cache.SetFileContents(evt.URI, evt.Content)

					success, err := updateParse(l.cache, evt.URI)
					if err != nil {
						l.logError(fmt.Errorf("failed to update parse: %w", err))
					}

					// if there are parse errors, just send these
					if !success {
						err = l.sendFileDiagnostics(ctx, evt.URI)
						if err != nil {
							l.logError(fmt.Errorf("failed to send diagnostic: %w", err))
						}
						continue
					}
				}

				// otherwise, lint the file and send the diagnostics
				err := updateFileDiagnostics(ctx, l.cache, l.loadedConfig, evt.URI)
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
			case reason := <-l.diagnosticRequestWorkspace:
				l.log(fmt.Sprintf("diagnostic request workspace: %q", reason))

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
	}()
}

func (l *LanguageServer) logError(err error) {
	if l.errorLog != nil {
		fmt.Fprintf(l.errorLog, "ERROR: %s\n", err)
	}
}

func (l *LanguageServer) log(msg string) {
	if l.errorLog != nil {
		fmt.Fprintf(l.errorLog, "%s\n", msg)
	}
}

func (l *LanguageServer) logInboundMessage(method string, message any) {
	if !l.verboseLogging {
		return
	}

	strMessage, ok := message.(string)
	if !ok {
		bs, err := json.Marshal(message)
		if err != nil {
			l.logError(fmt.Errorf("failed to marshal request: %w", err))
			return
		}

		strMessage = string(bs)
	}

	if l.errorLog != nil {
		fmt.Fprintf(l.errorLog, "->(%s): %s\n", method, strMessage)
	}
}

func (l *LanguageServer) logOutboundMessage(method string, message any) {
	if !l.verboseLogging {
		return
	}

	strMessage, ok := message.(string)
	if !ok {
		bs, err := json.Marshal(message)
		if err != nil {
			l.logError(fmt.Errorf("failed to marshal response: %w", err))
			return
		}

		strMessage = string(bs)
	}

	if l.errorLog != nil {
		fmt.Fprintf(l.errorLog, "<-(%s): %s\n", method, strMessage)
	}
}

// fileDiagnosticRequiredEvent is sent to the diagnosticRequestFile channel when
// diagnostics are required for a file
type fileDiagnosticRequiredEvent struct {
	Reason  string
	URI     string
	Content string
}

func (l *LanguageServer) handleTextDocumentDidOpen(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	var params TextDocumentDidOpenParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	l.diagnosticRequestFile <- fileDiagnosticRequiredEvent{
		Reason:  "textDocument/didOpen",
		URI:     params.TextDocument.URI,
		Content: params.TextDocument.Text,
	}

	return nil, nil
}

type TextDocumentDidChangeParams struct {
	TextDocument   TextDocumentIdentifier           `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

func (l *LanguageServer) handleTextDocumentDidChange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	var params TextDocumentDidChangeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	l.diagnosticRequestFile <- fileDiagnosticRequiredEvent{
		Reason:  "textDocument/didChange",
		URI:     params.TextDocument.URI,
		Content: params.ContentChanges[0].Text,
	}

	return nil, nil
}

func (l *LanguageServer) handleWorkspaceDidCreateFiles(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	var params WorkspaceDidCreateFilesParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	for _, createOp := range params.Files {
		_, _, err = updateCacheForURIFromDisk(l.cache, createOp.URI)
		if err != nil {
			return nil, fmt.Errorf("failed to update cache for uri %q: %w", createOp.URI, err)
		}

		l.diagnosticRequestFile <- fileDiagnosticRequiredEvent{
			Reason: "textDocument/didCreate",
			URI:    createOp.URI,
		}
	}

	return nil, nil
}

func (l *LanguageServer) handleWorkspaceDidDeleteFiles(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	var params WorkspaceDidDeleteFilesParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	for _, deleteOp := range params.Files {
		l.cache.Delete(deleteOp.URI)

		l.diagnosticRequestFile <- fileDiagnosticRequiredEvent{
			Reason: "textDocument/didDelete",
			URI:    deleteOp.URI,
		}
	}

	return nil, nil
}
func (l *LanguageServer) handleWorkspaceDidRenameFiles(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	var params WorkspaceDidRenameFilesParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	for _, renameOp := range params.Files {
		content, _, err := updateCacheForURIFromDisk(l.cache, renameOp.NewURI)
		if err != nil {
			return nil, fmt.Errorf("failed to update cache for uri %q: %w", renameOp.NewURI, err)
		}
		l.cache.Delete(renameOp.OldURI)

		l.diagnosticRequestFile <- fileDiagnosticRequiredEvent{
			Reason:  "textDocument/didRename",
			URI:     renameOp.NewURI,
			Content: content,
		}
	}

	return nil, nil
}

func (l *LanguageServer) handleWorkspaceDiagnostic(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	workspaceReport := WorkspaceDiagnosticReport{
		Items: make([]WorkspaceFullDocumentDiagnosticReport, 0),
	}

	workspaceReport.Items = append(workspaceReport.Items, WorkspaceFullDocumentDiagnosticReport{
		URI:     l.clientRootURI,
		Kind:    "full",
		Version: nil,
		Items:   l.cache.GetAllDiagnosticsForURI(l.clientRootURI),
	})

	l.logOutboundMessage("workspace/diagnostic", workspaceReport)

	return workspaceReport, nil
}

func (l *LanguageServer) handleInitialize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	var params InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	l.clientRootURI = params.RootURI

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
		},
	}

	folderURI, err := url.Parse(l.clientRootURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI: %w", err)
	}

	// load the rego source files into the cache
	err = filepath.Walk(folderURI.Path, func(path string, info os.FileInfo, err error) error {
		stat, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat %q: %w", path, err)
		}
		if stat.IsDir() {
			return nil
		}

		// TODO(charlieegan3): make this configurable for things like .rq etc?
		if !strings.HasSuffix(path, ".rego") {
			return nil
		}

		_, _, err = updateCacheForURIFromDisk(l.cache, fmt.Sprintf("file://%s", path))
		if err != nil {
			return fmt.Errorf("failed to update cache for uri %q: %w", path, err)
		}

		_, err = updateParse(l.cache, fmt.Sprintf("file://%s", path))
		if err != nil {
			return fmt.Errorf("failed to update parse: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk workspace dir %q: %w", folderURI.Path, err)
	}

	// attempt to load the config as it is found on disk
	file, err := config.FindConfig(strings.TrimPrefix(l.clientRootURI, "file://"))
	if err == nil {
		l.reloadConfig(file, false)
	}

	l.logOutboundMessage("initialize", result)

	return result, nil
}

func (l *LanguageServer) reloadConfig(configReader io.Reader, runWorkspaceDiagnostics bool) {
	l.loadedConfigLock.Lock()
	defer l.loadedConfigLock.Unlock()

	var loadedConfig config.Config
	err := yaml.NewDecoder(configReader).Decode(&loadedConfig)
	if err != io.EOF && err != nil {
		l.logError(fmt.Errorf("failed to reload config: %w", err))
		return
	}

	// if the config is now blank, then we need to clear it
	if err == io.EOF {
		l.loadedConfig = nil
	} else {
		l.loadedConfig = &loadedConfig
	}

	// this can be set to false by callers to disable the running of diagnostics for the whole workspace.
	// this is intended to be used at start up when a workspace run is already going to be taking place.
	if runWorkspaceDiagnostics {
		l.diagnosticRequestWorkspace <- fmt.Sprintf("config file changed")
	}
}

func (l *LanguageServer) handleInitialized(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {

	l.diagnosticRequestWorkspace <- "initialized"

	return nil, nil
}

func (l *LanguageServer) handleTextDocumentDiagnostic(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	// this is a no-op. Because we accept the textDocument/didChange event, which contains the new content,
	// we don't need to do anything here as once the new content has been parsed, the diagnostics will be sent
	// on the channel regardless of this request.
	return nil, nil
}

func (l *LanguageServer) handleWorkspaceDidChangeWatchedFiles(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params WorkspaceDidChangeWatchedFilesParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	// when a file is changed (saved), then we send trigger a full workspace lint
	var regoFiles []string
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
		l.diagnosticRequestWorkspace <- fmt.Sprintf("workspace/didChangeWatchedFiles (%s)", strings.Join(regoFiles, ", "))
	}

	return nil, nil
}

func (l *LanguageServer) sendFileDiagnostics(ctx context.Context, uri string) error {

	resp := FileDiagnostics{
		Items: l.cache.GetAllDiagnosticsForURI(uri),
		URI:   uri,
	}

	l.logOutboundMessage("textDocument/publishDiagnostics", resp)

	err := l.conn.Notify(ctx, "textDocument/publishDiagnostics", resp)
	if err != nil {
		return fmt.Errorf("failed to notify: %w", err)
	}

	return nil
}
