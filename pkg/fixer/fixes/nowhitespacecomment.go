package fixes

import (
	"errors"
	"strings"
)

type NoWhitespaceComment struct{}

func (*NoWhitespaceComment) Name() string {
	return "no-whitespace-comment"
}

func (n *NoWhitespaceComment) Fix(fc *FixCandidate, opts *RuntimeOptions) ([]FixResult, error) {
	lines := strings.Split(fc.Contents, "\n")

	if opts == nil {
		return nil, errors.New("missing runtime options")
	}

	fixed := false

	for _, loc := range opts.Locations {
		// unexpected line in file, skipping
		if loc.Row > len(lines) {
			continue
		}

		if loc.Column > len(lines[loc.Row-1]) || loc.Column < 1 {
			continue
		}

		line := lines[loc.Row-1]

		// unexpected character at location column, skipping
		if line[loc.Column-1] != byte('#') {
			continue
		}

		lines[loc.Row-1] = line[0:loc.Column] + " " + line[loc.Column:]
		fixed = true
	}

	if !fixed {
		return nil, nil
	}

	return []FixResult{{
		Title:    n.Name(),
		Root:     opts.BaseDir,
		Contents: strings.Join(lines, "\n"),
	}}, nil
}
