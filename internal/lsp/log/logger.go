package log

import (
	"fmt"
	"io"
	"sync"
)

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

type Logger struct {
	io.Writer

	Level Level
	rwm   sync.RWMutex
}

func (l Level) String() string {
	return [...]string{"Off", "Messages", "Debug"}[l]
}

func (l Level) ShouldLog(incoming Level) bool {
	return l >= incoming
}

func TraceValueToLevel(value string) (Level, error) {
	switch value {
	case "off":
		return LevelOff, nil
	case "messages":
		return LevelMessage, nil
	case "verbose":
		return LevelDebug, nil
	default:
		return LevelOff, fmt.Errorf("trace value must be one of 'off', 'messages', 'verbose', got: %s,", value)
	}
}

func NewLogger(level Level, writer io.Writer) *Logger {
	return &Logger{Level: level, Writer: writer, rwm: sync.RWMutex{}}
}

func NoOpLogger() *Logger {
	return &Logger{Level: LevelOff, Writer: io.Discard, rwm: sync.RWMutex{}}
}

func (l *Logger) SetLevel(level Level) {
	l.rwm.Lock()
	l.Level = level
	l.rwm.Unlock()
}

func (l *Logger) Message(message string, args ...any) {
	l.logf(LevelMessage, message, args...)
}

func (l *Logger) Debug(message string, args ...any) {
	l.logf(LevelDebug, message, args...)
}

func (l *Logger) logf(lvl Level, message string, args ...any) {
	l.rwm.RLock()

	if l.Level.ShouldLog(lvl) && l.Writer != nil {
		if len(args) > 0 {
			message = fmt.Sprintf(message, args...)
		}

		fmt.Fprintln(l.Writer, message)
	}

	l.rwm.RUnlock()
}
