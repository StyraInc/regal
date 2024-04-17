package fixes

import (
	"bytes"
	"slices"
)

type UseAssignmentOperator struct{}

func (*UseAssignmentOperator) Key() string {
	return "use-assignment-operator"
}

func (*UseAssignmentOperator) WholeFile() bool {
	return false
}

func (*UseAssignmentOperator) Fix(in []byte, opts *RuntimeOptions) (bool, []byte, error) {
	lines := bytes.Split(in, []byte("\n"))

	// this fix must have locations
	if len(opts.Locations) == 0 {
		return false, nil, nil
	}

	for _, loc := range opts.Locations {
		if loc.Row > len(lines) {
			return false, nil, nil
		}

		line := lines[loc.Row-1]

		// unexpected character at location column, skipping
		if line[loc.Col-1] != byte('=') {
			continue
		}

		lines[loc.Row-1] = slices.Concat(line[0:loc.Col-1], []byte(":"), line[loc.Col-1:])
	}

	return true, bytes.Join(lines, []byte("\n")), nil
}
