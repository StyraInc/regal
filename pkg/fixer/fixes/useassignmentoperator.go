package fixes

import (
	"errors"
	"strings"
)

type UseAssignmentOperator struct{}

func (*UseAssignmentOperator) Name() string {
	return "use-assignment-operator"
}

func (u *UseAssignmentOperator) Fix(fc *FixCandidate, opts *RuntimeOptions) ([]FixResult, error) {
	lines := strings.Split(fc.Contents, "\n")

	if opts == nil {
		return nil, errors.New("missing runtime options")
	}

	fixed := false

	for _, loc := range opts.Locations {
		if loc.Row > len(lines) {
			continue
		}

		line := lines[loc.Row-1]

		if loc.Column-1 < 0 || loc.Column-1 >= len(line) {
			continue
		}

		// unexpected character at location column, skipping
		if line[loc.Column-1] != '=' {
			continue
		}

		lines[loc.Row-1] = line[0:loc.Column-1] + ":" + line[loc.Column-1:]
		fixed = true
	}

	if !fixed {
		return nil, nil
	}

	return []FixResult{{
		Title:    u.Name(),
		Root:     opts.BaseDir,
		Contents: strings.Join(lines, "\n"),
	}}, nil
}
