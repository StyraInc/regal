# METADATA
# description: |
#   This file is used for e2e tests of most of the provided linter rules. All violations can't
#   be tested in a single file, as some are mutually exlusive (i.e. implicit-future-keywords and
#   rule-named-if).
package all_violations

# Imports

# avoid-importing-input
import input

# implict-future-keywords
import future.keywords

# import-shadows-import
import data.foo
import data.foo

import data.redundant.alias as alias

# redundant-data-import
import data

### Bugs ###

constant_condition {
	1 == 1
}

# METADATA
# invalid-metadata-attribute: true
should := "fail"

not_equals_in_loop {
	"foo" != input.bar[_]
}

# rule-shadows-builtin
contains := true

top_level_iteration := input[_]

unused_return_value {
	indexof("foo", "o")
}

### Idiomatic ###

custom_has_key_construct(map, key) {
	_ = map[key]
}

custom_in_construct(coll, item) {
	item == coll[_]
}

### Style ###

# avoid-get-and-list-prefix
get_foo(foo) := foo

# METADATA
# description: detached-metadata

annotation := "detached"

external_reference(_) {
	data.foo
}

function_arg_return {
	indexof("foo", "o", i)
	i == 1
}

line_length_should_be_no_longer_than_120_characters_but_this_line_is_really_long_and_will_exceed_that_limit_which_is := "bad"

#no-whitespace-comment

 opa_fmt := "fail"

preferSnakeCase := "fail"

# todo-comment

x := y {
	y := 1
}

use_assignment = "oparator"

use_in_operator {
	"item" == input.coll[_]
}

# this will also tringger the test-outside-test-package rule
test_identically_named_tests := true
test_identically_named_tests := true

todo_test_bad {
	input.bad
}

print_or_trace_call {
	print("forbidden!")
}

non_raw_regex_pattern := regex.match("[0-9]", "1")

use_some_for_output_vars {
	input.foo[output_var]
}

# metasyntactic variable
foo := "bar"

chained_rule_body {
	input.x
} {
	input.y
}

# dubious print sprintf
y {
	print(sprintf("name is: %s domain is: %s", [input.name, input.domain]))
}