package types

import (
	"testing"

	"github.com/open-policy-agent/regal/internal/util"
)

func TestPositionToOffset(t *testing.T) {
	t.Parallel()

	text := "line1\nline2\nline3"

	for line := range uint32(2) {
		for char := range uint32(5) {
			pos := Position{Line: line, Character: char}
			exp := line*6 + char
			got := util.IntTo[uint32](pos.ToOffset(text))

			if exp != got {
				t.Fatalf("expected offset for line %d char %d to be %d, got %d", line, char, exp, got)
			}
		}
	}
}
