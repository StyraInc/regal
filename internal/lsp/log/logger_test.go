package log

import "testing"

func TestShouldLog(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		Level                Level
		IncomingMessageLevel Level
		ExpectLog            bool
	}{
		"Off, no logs": {
			Level:                LevelOff,
			IncomingMessageLevel: LevelMessage,
			ExpectLog:            false,
		},
		"Off, no logs from debug": {
			Level:                LevelOff,
			IncomingMessageLevel: LevelDebug,
			ExpectLog:            false,
		},
		"Message, no logs from debug": {
			Level:                LevelMessage,
			IncomingMessageLevel: LevelDebug,
			ExpectLog:            false,
		},
		"Debug, logs from Message": {
			Level:                LevelDebug,
			IncomingMessageLevel: LevelMessage,
			ExpectLog:            true,
		},
		"Debug, all logs": {
			Level:                LevelDebug,
			IncomingMessageLevel: LevelDebug,
			ExpectLog:            true,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			t.Parallel()

			if got := tc.Level.ShouldLog(tc.IncomingMessageLevel); got != tc.ExpectLog {
				t.Errorf("expected %v, got %v", tc.ExpectLog, got)
			}
		})
	}
}
