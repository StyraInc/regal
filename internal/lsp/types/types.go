package types

import (
	"github.com/styrainc/regal/internal/lsp/types/completion"
	"github.com/styrainc/regal/internal/lsp/types/symbols"
)

type FileDiagnostics struct {
	URI   string       `json:"uri"`
	Items []Diagnostic `json:"diagnostics"`
}

type WorkspaceDidChangeWatchedFilesParams struct {
	Changes []FileEvent `json:"changes"`
}

type FileEvent struct {
	Type uint   `json:"type"`
	URI  string `json:"uri"`
}

type InitializeParams struct {
	ProcessID        int                `json:"processId"`
	ClientInfo       Client             `json:"clientInfo"`
	Locale           string             `json:"locale"`
	RootPath         string             `json:"rootPath"`
	RootURI          string             `json:"rootUri"`
	Capabilities     ClientCapabilities `json:"capabilities"`
	Trace            string             `json:"trace"`
	WorkspaceFolders []WorkspaceFolder  `json:"workspaceFolders"`
}

type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

type Client struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ClientCapabilities struct {
	Workspace WorkspaceClientCapabilities    `json:"workspace"`
	Text      TextDocumentClientCapabilities `json:"textDocument"`
	Window    WindowClientCapabilities       `json:"window"`
	General   GeneralClientCapabilities      `json:"general"`
}

type WorkspaceClientCapabilities struct {
	Diagnostics DiagnosticWorkspaceClientCapabilities `json:"diagnostics"`
}

type DiagnosticWorkspaceClientCapabilities struct {
	RefreshSupport bool `json:"refreshSupport"`
}

type TextDocumentClientCapabilities struct {
	Diagnostic DiagnosticClientCapabilities `json:"diagnostic"`
}

type DiagnosticClientCapabilities struct {
	DynamicRegistration    bool `json:"dynamicRegistration"`
	RelatedDocumentSupport bool `json:"relatedDocumentSupport"`
}

type WindowClientCapabilities struct {
	WorkDoneProgress bool `json:"workDoneProgress"`
}

type GeneralClientCapabilities struct {
	StaleRequestSupport StaleRequestSupportClientCapabilities `json:"staleRequestSupport"`
}

type ShowMessageParams struct {
	Type    uint   `json:"type"`
	Message string `json:"message"`
}

type StaleRequestSupportClientCapabilities struct {
	Cancel                  bool     `json:"cancel"`
	RetryOnContentModifieds []string `json:"retryOnContentModified"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

type ServerCapabilities struct {
	TextDocumentSyncOptions    TextDocumentSyncOptions `json:"textDocumentSync"`
	DiagnosticProvider         DiagnosticOptions       `json:"diagnosticProvider"`
	Workspace                  WorkspaceOptions        `json:"workspace"`
	InlayHintProvider          InlayHintOptions        `json:"inlayHintProvider"`
	HoverProvider              bool                    `json:"hoverProvider"`
	CodeActionProvider         CodeActionOptions       `json:"codeActionProvider"`
	ExecuteCommandProvider     ExecuteCommandOptions   `json:"executeCommandProvider"`
	DocumentFormattingProvider bool                    `json:"documentFormattingProvider"`
	FoldingRangeProvider       bool                    `json:"foldingRangeProvider"`
	DocumentSymbolProvider     bool                    `json:"documentSymbolProvider"`
	WorkspaceSymbolProvider    bool                    `json:"workspaceSymbolProvider"`
	DefinitionProvider         bool                    `json:"definitionProvider"`
	CompletionProvider         CompletionOptions       `json:"completionProvider"`
}

type CompletionOptions struct {
	CompletionItem  CompletionItemOptions `json:"completionItem"`
	ResolveProvider bool                  `json:"resolveProvider"`
}

type CompletionItemOptions struct {
	LabelDetailsSupport bool `json:"labelDetailsSupport"`
}

type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      CompletionContext      `json:"context"`
}

type CompletionContext struct {
	TriggerKind      completion.TriggerKind `json:"triggerKind"`
	TriggerCharacter string                 `json:"triggerCharacter"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

type CompletionItem struct {
	Label        string                      `json:"label"`
	LabelDetails *CompletionItemLabelDetails `json:"labelDetails,omitempty"`
	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#completionItemKind
	Kind            completion.ItemKind `json:"kind"`
	Detail          string              `json:"detail"`
	Documentation   *MarkupContent      `json:"documentation,omitempty"`
	Preselect       bool                `json:"preselect"`
	TextEdit        *TextEdit           `json:"textEdit,omitempty"`
	InserTextFormat *uint               `json:"insertTextFormat,omitempty"`

	// Mandatory is used to indicate that the completion item is mandatory and should be offered
	// as an exclusive completion. This is not part of the LSP spec, but used in regal providers
	// to indicate that the completion item is the only valid completion.
	Mandatory bool `json:"-"`

	// Regal is used to store regal-specific metadata about the completion item.
	// This is not part of the LSP spec, but used in the manager to post process
	// items before returning them to the client.
	Regal *CompletionItemRegalMetadata `json:"_regal,omitempty"`
}

type CompletionItemRegalMetadata struct {
	Provider string `json:"provider"`
}

type CompletionItemLabelDetails struct {
	Description string `json:"description"`
	Detail      string `json:"detail"`
}

type WorkspaceOptions struct {
	FileOperations FileOperationsServerCapabilities `json:"fileOperations"`
}

type CodeActionOptions struct {
	CodeActionKinds []string `json:"codeActionKinds"`
}

type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type CodeAction struct {
	Title       string       `json:"title"`
	Kind        string       `json:"kind"`
	Diagnostics []Diagnostic `json:"diagnostics"`
	IsPreferred *bool        `json:"isPreferred,omitempty"`
	Command     Command      `json:"command"`
}

type Command struct {
	Title     string `json:"title"`
	Tooltip   string `json:"tooltip"`
	Command   string `json:"command"`
	Arguments *[]any `json:"arguments,omitempty"`
}

type ExecuteCommandOptions struct {
	Commands []string `json:"commands"`
}

type ExecuteCommandParams struct {
	Command   string `json:"command"`
	Arguments []any  `json:"arguments"`
}

type ApplyWorkspaceEditParams struct {
	Label string        `json:"label"`
	Edit  WorkspaceEdit `json:"edit"`
}

type WorkspaceEdit struct {
	DocumentChanges []TextDocumentEdit `json:"documentChanges"`
}

type TextDocumentEdit struct {
	// TextDocument is the document to change. Not that this could be versioned,
	// (OptionalVersionedTextDocumentIdentifier) but we currently don't use that.
	TextDocument OptionalVersionedTextDocumentIdentifier `json:"textDocument"`
	Edits        []TextEdit                              `json:"edits"`
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DocumentSymbol struct {
	Name           string             `json:"name"`
	Detail         *string            `json:"detail,omitempty"`
	Kind           symbols.SymbolKind `json:"kind"`
	Range          Range              `json:"range"`
	SelectionRange Range              `json:"selectionRange"`
	Children       *[]DocumentSymbol  `json:"children,omitempty"`
}

type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}

type WorkspaceSymbol struct {
	Name          string             `json:"name"`
	Kind          symbols.SymbolKind `json:"kind"`
	Location      Location           `json:"location"`
	ContainerName *string            `json:"containerName,omitempty"`
}

type FoldingRangeParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type FoldingRange struct {
	StartLine      uint   `json:"startLine"`
	StartCharacter *uint  `json:"startCharacter,omitempty"`
	EndLine        uint   `json:"endLine"`
	EndCharacter   *uint  `json:"endCharacter,omitempty"`
	Kind           string `json:"kind"`
}

type FormattingOptions struct {
	TabSize                uint `json:"tabSize"`
	InsertSpaces           bool `json:"insertSpaces"`
	TrimTrailingWhitespace bool `json:"trimTrailingWhitespace"`
	InsertFinalNewline     bool `json:"insertFinalNewline"`
	TrimFinalNewlines      bool `json:"trimFinalNewlines"`
}

type FileOperationsServerCapabilities struct {
	DidCreate FileOperationRegistrationOptions `json:"didCreate"`
	DidRename FileOperationRegistrationOptions `json:"didRename"`
	DidDelete FileOperationRegistrationOptions `json:"didDelete"`
}

type FileOperationRegistrationOptions struct {
	Filters []FileOperationFilter `json:"filters"`
}

type FileOperationFilter struct {
	Scheme  string               `json:"scheme"`
	Pattern FileOperationPattern `json:"pattern"`
}
type FileOperationPattern struct {
	Glob string `json:"glob"`
}

type DiagnosticOptions struct {
	Identifier            string `json:"identifier"`
	InterFileDependencies bool   `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool   `json:"workspaceDiagnostics"`
}

type InlayHintOptions struct {
	ResolveProvider bool `json:"resolveProvider"`
}

type InlayHint struct {
	Position     Position      `json:"position"`
	Label        string        `json:"label"`
	Kind         uint          `json:"kind"`
	PaddingLeft  bool          `json:"paddingLeft"`
	PaddingRight bool          `json:"paddingRight"`
	Tooltip      MarkupContent `json:"tooltip"`
}

type TextDocumentInlayHintParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
}

type TextDocumentSaveOptions struct {
	IncludeText bool `json:"includeText"`
}

type TextDocumentDidSaveParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Text         *string                `json:"text,omitempty"`
}

type TextDocumentSyncOptions struct {
	OpenClose bool                    `json:"openClose"`
	Change    uint                    `json:"change"`
	Save      TextDocumentSaveOptions `json:"save"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type OptionalVersionedTextDocumentIdentifier struct {
	URI string `json:"uri"`
	// Version is optional (i.e. it can be null), but it cannot be undefined when used in some requests
	// (see workspace/applyEdit).
	Version *uint `json:"version"`
}

type TextDocumentDidChangeParams struct {
	TextDocument   TextDocumentIdentifier           `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

type Diagnostic struct {
	Range           Range            `json:"range"`
	Message         string           `json:"message"`
	Severity        uint             `json:"severity"`
	Source          string           `json:"source"`
	Code            string           `json:"code"`
	CodeDescription *CodeDescription `json:"codeDescription,omitempty"`
}

type CodeDescription struct {
	Href string `json:"href"`
}

type DiagnosticCode struct {
	Value  string `json:"value"`
	Target string `json:"target"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      uint `json:"line"`
	Character uint `json:"character"`
}

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type TextDocumentDidOpenParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type TextDocumentDidCloseParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type TextDocumentItem struct {
	LanguageID string `json:"languageId"`
	Text       string `json:"text"`
	URI        string `json:"uri"`
	Version    uint   `json:"version"`
}

type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type TextDocumentHoverParams struct {
	Position     Position               `json:"position"`
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type WorkspaceDidCreateFilesParams struct {
	Files []WorkspaceDidCreateFilesParamsCreatedFile `json:"files"`
}

type WorkspaceDidCreateFilesParamsCreatedFile struct {
	URI string `json:"uri"`
}

type WorkspaceDidDeleteFilesParams struct {
	Files []WorkspaceDidDeleteFilesParamsDeletedFile `json:"files"`
}

type WorkspaceDidDeleteFilesParamsDeletedFile struct {
	URI string `json:"uri"`
}

type WorkspaceDidRenameFilesParams struct {
	Files []WorkspaceDidRenameFilesParamsFileRename `json:"files"`
}

type WorkspaceDidRenameFilesParamsFileRename struct {
	NewURI string `json:"newUri"`
	OldURI string `json:"oldUri"`
}

type WorkspaceDiagnosticReport struct {
	Items []WorkspaceFullDocumentDiagnosticReport `json:"items"`
}

type WorkspaceFullDocumentDiagnosticReport struct {
	URI     string       `json:"uri"`
	Version *uint        `json:"version"`
	Kind    string       `json:"kind"` // full, or incremental. We always use full
	Items   []Diagnostic `json:"items"`
}
