package log

// Level controls the level of server logging and corresponds to TraceValue in
// the LSP spec:
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#traceValue.
type Level int

const (
	// LevelOff is used to disable logging completely.
	LevelOff Level = iota
	// LogLevelMessage are intended to contain errors and other messages that
	// should be shown in normal operation.
	LevelMessage
	// LogLevelDebug is includes LogLevelMessage, but also information that is
	// not expected to be useful unless debugging the server.
	LevelDebug
)

func (l Level) String() string {
	return [...]string{"Off", "Messages", "Debug"}[l]
}

func (l Level) ShouldLog(incoming Level) bool {
	if l == LevelOff {
		return false
	}

	if l == LevelDebug {
		return true
	}

	return l == LevelMessage && incoming == LevelMessage
}
