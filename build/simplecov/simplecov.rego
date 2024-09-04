package build.simplecov

import rego.v1

from_opa := {"coverage": coverage}

coverage[file] := {"lines": to_lines(report)} if some file, report in input.files

covered_map(report) := cm if {
	covered := object.get(report, "covered", [])
	cm := {line: 1 |
		some item in covered
		some line in numbers.range(item.start.row, item.end.row)
	}
}

not_covered_map(report) := ncm if {
	not_covered := object.get(report, "not_covered", [])
	ncm := {line: 0 |
		some item in not_covered
		some line in numbers.range(item.start.row, item.end.row)
	}
}

to_lines(report) := lines if {
	cm := covered_map(report)
	ncm := not_covered_map(report)
	keys := sort([line | some line, _ in object.union(cm, ncm)])
	last := keys[count(keys) - 1]

	lines := [value |
		some i in numbers.range(1, last)
		value := to_value(cm, ncm, i)
	]
}

to_value(cm, _, line) := 1 if cm[line]

to_value(_, ncm, line) := 0 if ncm[line]

to_value(cm, ncm, line) := null if {
	not cm[line]
	not ncm[line]
}
