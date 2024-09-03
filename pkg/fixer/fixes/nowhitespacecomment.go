package fixes

import (
	"bytes"
	"errors"
	"slices"
)

type NoWhitespaceComment struct{}

func (*NoWhitespaceComment) Name() string {
	return "no-whitespace-comment"
}

func (n *NoWhitespaceComment) Fix(fc *FixCandidate, opts *RuntimeOptions) ([]FixResult, error) {
	lines := bytes.Split(fc.Contents, []byte("\n"))

	if opts == nil {
		return nil, errors.New("missing runtime options")
	}

	fixed := false

	for _, loc := range opts.Locations {
		// unexpected line in file, skipping
		if loc.Row > len(lines) {
			continue
		}

		if loc.Col > len(lines[loc.Row-1]) || loc.Col < 1 {
			continue
		}

		line := lines[loc.Row-1]

		// unexpected character at location column, skipping
		if line[loc.Col-1] != byte('#') {
			continue
		}

		lines[loc.Row-1] = slices.Concat(line[0:loc.Col], []byte(" "), line[loc.Col:])
		fixed = true
	}

	if !fixed {
		return nil, nil
	}

	return []FixResult{{
		Title:    n.Name(),
		Root:     opts.BaseDir,
		Contents: bytes.Join(lines, []byte("\n")),
	}}, nil
}
