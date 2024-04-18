package fixes

import (
	"bytes"
	"slices"
)

type NoWhitespaceComment struct{}

func (*NoWhitespaceComment) Key() string {
	return "no-whitespace-comment"
}

func (*NoWhitespaceComment) Fix(fc *FixCandidate, opts *RuntimeOptions) ([]FixResult, error) {
	lines := bytes.Split(fc.Contents, []byte("\n"))

	// this fix must have locations
	if len(opts.Locations) == 0 {
		return nil, nil
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

	return []FixResult{{Contents: bytes.Join(lines, []byte("\n"))}}, nil
}
