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

# import-shadows-builtin
import data.http

# prefer-package-imports
# aggregate rule, so will only fail if more than one file is linted
import data.rule_named_if.allow

rule := "here"

import data.after.rule

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

use_some_for_output_vars {
	input.foo[output_var]
}

non_raw_regex_pattern := regex.match("[0-9]", "1")

use_in_operator {
	"item" == input.coll[_]
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

chained_rule_body {
	input.x
} {
	input.y
}

rule_length {
	input.x1
	input.x2
	input.x3
	input.x4
	input.x5
	input.x6
	input.x7
	input.x8
	input.x9
	input.x10
	input.x11
	input.x12
	input.x13
	input.x14
	input.x15
	input.x16
	input.x17
	input.x18
	input.x19
	input.x20
	input.x21
	input.x22
	input.x23
	input.x24
	input.x25
	input.x26
	input.x27
	input.x28
	input.x29
	input.x30
}

default_over_else := 1 {
	input.x
} else := 3

### Testing ###

# this will also tringger the test-outside-test-package rule
test_identically_named_tests := true
test_identically_named_tests := true

todo_test_bad {
	input.bad
}

print_or_trace_call {
	print("forbidden!")
}

# metasyntactic variable
foo := "bar"

# dubious print sprintf
y {
	print(sprintf("name is: %s domain is: %s", [input.name, input.domain]))
}
