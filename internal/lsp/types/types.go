package types

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
	IsPreferred bool         `json:"isPreferred"`
	Command     Command      `json:"command"`
}

type Command struct {
	Title     string `json:"title"`
	Tooltip   string `json:"tooltip"`
	Command   string `json:"command"`
	Arguments []any  `json:"arguments"`
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

type TextDocumentSyncOptions struct {
	OpenClose bool `json:"openClose"`
	Change    uint `json:"change"`
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
	Range           Range           `json:"range"`
	Message         string          `json:"message"`
	Severity        uint            `json:"severity"`
	Source          string          `json:"source"`
	Code            string          `json:"code"`
	CodeDescription CodeDescription `json:"codeDescription"`
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

type TextDocumentItem struct {
	LanguageID string `json:"languageId"`
	Text       string `json:"text"`
	URI        string `json:"uri"`
	Version    uint   `json:"version"`
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
