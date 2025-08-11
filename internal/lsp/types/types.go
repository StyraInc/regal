package types

import (
	"github.com/open-policy-agent/opa/v1/ast"

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
	URI  string `json:"uri"`
	Type uint   `json:"type"`
}

type InitializationOptions struct {
	// Formatter specifies the formatter to use. Options: 'opa fmt' (default),
	// 'opa fmt --rego-v1' or 'regal fix'.
	Formatter *string `json:"formatter,omitempty"`
	// EnableDebugCodelens, if set, will enable debug codelens
	// when clients request code lenses for a file.
	EnableDebugCodelens *bool `json:"enableDebugCodelens,omitempty"`
	// EvalCodelensDisplayInline, if set, will show evaluation results natively
	// in the calling editor, rather than in an output file.
	EvalCodelensDisplayInline *bool `json:"evalCodelensDisplayInline,omitempty"`
}

type InitializeParams struct {
	InitializationOptions *InitializationOptions `json:"initializationOptions,omitempty"`
	ClientInfo            Client                 `json:"clientInfo"`
	Locale                string                 `json:"locale"`
	RootPath              string                 `json:"rootPath"`
	RootURI               string                 `json:"rootUri"`
	Trace                 string                 `json:"trace"`
	WorkspaceFolders      *[]WorkspaceFolder     `json:"workspaceFolders"`
	Capabilities          ClientCapabilities     `json:"capabilities"`
	ProcessID             int                    `json:"processId"`
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
	General   GeneralClientCapabilities      `json:"general"`
	Text      TextDocumentClientCapabilities `json:"textDocument"`
	Workspace WorkspaceClientCapabilities    `json:"workspace"`
	Window    WindowClientCapabilities       `json:"window"`
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
	Message string `json:"message"`
	Type    uint   `json:"type"`
}

type StaleRequestSupportClientCapabilities struct {
	RetryOnContentModified []string `json:"retryOnContentModified"`
	Cancel                 bool     `json:"cancel"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

type ServerCapabilities struct {
	CodeLensProvider           *CodeLensOptions        `json:"codeLensProvider,omitempty"`
	Workspace                  WorkspaceOptions        `json:"workspace"`
	DiagnosticProvider         DiagnosticOptions       `json:"diagnosticProvider"`
	CodeActionProvider         CodeActionOptions       `json:"codeActionProvider"`
	ExecuteCommandProvider     ExecuteCommandOptions   `json:"executeCommandProvider"`
	TextDocumentSyncOptions    TextDocumentSyncOptions `json:"textDocumentSync"`
	CompletionProvider         CompletionOptions       `json:"completionProvider"`
	InlayHintProvider          InlayHintOptions        `json:"inlayHintProvider"`
	HoverProvider              bool                    `json:"hoverProvider"`
	DocumentFormattingProvider bool                    `json:"documentFormattingProvider"`
	FoldingRangeProvider       bool                    `json:"foldingRangeProvider"`
	DocumentSymbolProvider     bool                    `json:"documentSymbolProvider"`
	WorkspaceSymbolProvider    bool                    `json:"workspaceSymbolProvider"`
	DefinitionProvider         bool                    `json:"definitionProvider"`
}

type CompletionOptions struct {
	CompletionItem    CompletionItemOptions `json:"completionItem"`
	ResolveProvider   bool                  `json:"resolveProvider"`
	TriggerCharacters []string              `json:"triggerCharacters,omitempty"`
}

type CompletionItemOptions struct {
	LabelDetailsSupport bool `json:"labelDetailsSupport"`
}

type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Context      CompletionContext      `json:"context"`
	Position     Position               `json:"position"`
	RegoVersion  ast.RegoVersion        `json:"regoVersion"`
}

type CompletionContext struct {
	TriggerCharacter string                 `json:"triggerCharacter"`
	TriggerKind      completion.TriggerKind `json:"triggerKind"`
}

type CompletionList struct {
	Items        []CompletionItem `json:"items"`
	IsIncomplete bool             `json:"isIncomplete"`
}

type CompletionItem struct {
	LabelDetails    *CompletionItemLabelDetails `json:"labelDetails,omitempty"`
	Documentation   *MarkupContent              `json:"documentation,omitempty"`
	TextEdit        *TextEdit                   `json:"textEdit,omitempty"`
	InserTextFormat *uint                       `json:"insertTextFormat,omitempty"`

	// Regal is used to store regal-specific metadata about the completion item.
	// This is not part of the LSP spec, but used in the manager to post process
	// items before returning them to the client.
	Regal  *CompletionItemRegalMetadata `json:"_regal,omitempty"`
	Label  string                       `json:"label"`
	Detail string                       `json:"detail"`
	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#completionItemKind
	Kind      completion.ItemKind `json:"kind"`
	Preselect bool                `json:"preselect"`

	// Mandatory is used to indicate that the completion item is mandatory and should be offered
	// as an exclusive completion. This is not part of the LSP spec, but used in regal providers
	// to indicate that the completion item is the only valid completion.
	Mandatory bool `json:"-"`
}

type CompletionItemRegalMetadata struct {
	Provider string `json:"provider"`
}

type CompletionItemLabelDetails struct {
	Description string `json:"description"`
	Detail      string `json:"detail"`
}

type WorkspaceFoldersServerCapabilities struct {
	Supported bool `json:"supported"`
}

type WorkspaceOptions struct {
	FileOperations   FileOperationsServerCapabilities   `json:"fileOperations"`
	WorkspaceFolders WorkspaceFoldersServerCapabilities `json:"workspaceFolders"`
}

type CodeActionOptions struct {
	CodeActionKinds []string `json:"codeActionKinds"`
}

type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Context      CodeActionContext      `json:"context"`
	Range        Range                  `json:"range"`
}

type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
	Only        []string     `json:"only,omitempty"`
	TriggerKind *uint8       `json:"triggerKind,omitempty"`
}

type CodeAction struct {
	Command     Command      `json:"command"`
	IsPreferred *bool        `json:"isPreferred,omitempty"`
	Title       string       `json:"title"`
	Kind        string       `json:"kind"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

type CodeLensOptions struct {
	ResolveProvider *bool `json:"resolveProvider,omitempty"`
}

type CodeLensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type CodeLens struct {
	Command *Command `json:"command,omitempty"`
	Data    *any     `json:"data,omitempty"`
	Range   Range    `json:"range"`
}

type Command struct {
	Arguments *[]any `json:"arguments,omitempty"`
	Title     string `json:"title"`
	Tooltip   string `json:"tooltip"`
	Command   string `json:"command"`
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

type ApplyWorkspaceRenameEditParams struct {
	Label string              `json:"label"`
	Edit  WorkspaceRenameEdit `json:"edit"`
}

type ApplyWorkspaceAnyEditParams struct {
	Label string           `json:"label"`
	Edit  WorkspaceAnyEdit `json:"edit"`
}
type WorkspaceAnyEdit struct {
	DocumentChanges []any `json:"documentChanges"`
}

type RenameFileOptions struct {
	Overwrite      bool `json:"overwrite"`
	IgnoreIfExists bool `json:"ignoreIfExists"`
}

type RenameFile struct {
	Options              *RenameFileOptions `json:"options,omitempty"`
	AnnotationIdentifier *string            `json:"annotationId,omitempty"`
	Kind                 string             `json:"kind"` // must always be "rename"
	OldURI               string             `json:"oldUri"`
	NewURI               string             `json:"newUri"`
}

type DeleteFileOptions struct {
	Recursive         bool `json:"recursive"`
	IgnoreIfNotExists bool `json:"ignoreIfNotExists"`
}

type DeleteFile struct {
	Options *DeleteFileOptions `json:"options,omitempty"`
	Kind    string             `json:"kind"` // must always be "delete"
	URI     string             `json:"uri"`
}

// WorkspaceRenameEdit is a WorkspaceEdit that is used for renaming files.
// Perhaps we should use generics and a union type here instead.
type WorkspaceRenameEdit struct {
	DocumentChanges []RenameFile `json:"documentChanges"`
}

type WorkspaceDeleteEdit struct {
	DocumentChanges []DeleteFile `json:"documentChanges"`
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
	NewText string `json:"newText"`
	Range   Range  `json:"range"`
}

type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DocumentSymbol struct {
	Detail         *string            `json:"detail,omitempty"`
	Children       *[]DocumentSymbol  `json:"children,omitempty"`
	Name           string             `json:"name"`
	Range          Range              `json:"range"`
	SelectionRange Range              `json:"selectionRange"`
	Kind           symbols.SymbolKind `json:"kind"`
}

type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}

type WorkspaceSymbol struct {
	ContainerName *string            `json:"containerName,omitempty"`
	Name          string             `json:"name"`
	Location      Location           `json:"location"`
	Kind          symbols.SymbolKind `json:"kind"`
}

type FoldingRangeParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type FoldingRange struct {
	StartCharacter *uint  `json:"startCharacter,omitempty"`
	EndCharacter   *uint  `json:"endCharacter,omitempty"`
	Kind           string `json:"kind"`
	StartLine      uint   `json:"startLine"`
	EndLine        uint   `json:"endLine"`
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
	Tooltip      MarkupContent `json:"tooltip"`
	Label        string        `json:"label"`
	Position     Position      `json:"position"`
	Kind         uint          `json:"kind"`
	PaddingLeft  bool          `json:"paddingLeft"`
	PaddingRight bool          `json:"paddingRight"`
}

type TextDocumentInlayHintParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
}

type TextDocumentSaveOptions struct {
	IncludeText bool `json:"includeText"`
}

type TextDocumentDidSaveParams struct {
	Text         *string                `json:"text,omitempty"`
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type TextDocumentSyncOptions struct {
	Change    uint                    `json:"change"`
	OpenClose bool                    `json:"openClose"`
	Save      TextDocumentSaveOptions `json:"save"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type OptionalVersionedTextDocumentIdentifier struct {
	// Version is optional (i.e. it can be null), but it cannot be undefined when used in some requests
	// (see workspace/applyEdit).
	Version *uint  `json:"version"`
	URI     string `json:"uri"`
}

type TextDocumentDidChangeParams struct {
	TextDocument   TextDocumentIdentifier           `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

type Diagnostic struct {
	CodeDescription *CodeDescription `json:"codeDescription,omitempty"`
	Message         string           `json:"message"`
	Source          *string          `json:"source,omitempty"`
	Code            string           `json:"code"` // spec says optional integer or string
	Range           Range            `json:"range"`
	Severity        *uint            `json:"severity,omitempty"`
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
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    Range         `json:"range"`
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

type iuint interface{ ~int | ~uint }

func RangeBetween[T1, T2, T3, T4 iuint](startLine T1, startCharacter T2, endLine T3, endCharacter T4) Range {
	return Range{
		Start: Position{Line: uint(startLine), Character: uint(startCharacter)},
		End:   Position{Line: uint(endLine), Character: uint(endCharacter)},
	}
}
