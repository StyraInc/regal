package fixes

import (
	"bytes"
	"errors"
	"slices"
)

type UseAssignmentOperator struct{}

func (*UseAssignmentOperator) Name() string {
	return "use-assignment-operator"
}

func (u *UseAssignmentOperator) Fix(fc *FixCandidate, opts *RuntimeOptions) ([]FixResult, error) {
	lines := bytes.Split(fc.Contents, []byte("\n"))

	if opts == nil {
		return nil, errors.New("missing runtime options")
	}

	fixed := false

	for _, loc := range opts.Locations {
		if loc.Row > len(lines) {
			continue
		}

		line := lines[loc.Row-1]

		if loc.Col-1 < 0 || loc.Col-1 >= len(line) {
			continue
		}

		// unexpected character at location column, skipping
		if line[loc.Col-1] != byte('=') {
			continue
		}

		lines[loc.Row-1] = slices.Concat(line[0:loc.Col-1], []byte(":"), line[loc.Col-1:])
		fixed = true
	}

	if !fixed {
		return nil, nil
	}

	return []FixResult{{
		Title:    u.Name(),
		Root:     opts.BaseDir,
		Contents: bytes.Join(lines, []byte("\n")),
	}}, nil
}
