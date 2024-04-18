package fixes

import (
	"bytes"
	"slices"
)

type UseAssignmentOperator struct{}

func (*UseAssignmentOperator) Name() string {
	return "use-assignment-operator"
}

func (*UseAssignmentOperator) Fix(fc *FixCandidate, opts *RuntimeOptions) ([]FixResult, error) {
	lines := bytes.Split(fc.Contents, []byte("\n"))

	// this fix must have locations
	if len(opts.Locations) == 0 {
		return nil, nil
	}

	fixed := false

	for _, loc := range opts.Locations {
		if loc.Row > len(lines) {
			continue
		}

		line := lines[loc.Row-1]

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

	return []FixResult{{Contents: bytes.Join(lines, []byte("\n"))}}, nil
}
