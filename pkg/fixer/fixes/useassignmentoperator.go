package fixes

import (
	"bytes"
)

type UseAssignmentOperator struct{}

func (*UseAssignmentOperator) Key() string {
	return "use-assignment-operator"
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

		if loc.Col != 1 {
			// current impl only understands the first column
			return false, nil, nil
		}

		line := lines[loc.Row-1]

		lines[loc.Row-1] = bytes.Replace(line, []byte("="), []byte(":="), 1)
	}

	return true, bytes.Join(lines, []byte("\n")), nil
}
