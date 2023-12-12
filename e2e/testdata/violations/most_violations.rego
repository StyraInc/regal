# METADATA
# description: |
#   This file is used for e2e tests of most of the provided linter rules. All violations can't
#   be tested in a single file, as some are mutually exlusive (i.e. implicit-future-keywords and
#   rule-named-if).
package all_violations

import rego.v1

# Imports

# avoid-importing-input
import input

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

constant_condition if {
	1 == 1
}

# METADATA
# invalid-metadata-attribute: true
should := "fail"

not_equals_in_loop if {
	"foo" != input.bar[_]
}

# rule-shadows-builtin
contains := true

top_level_iteration := input[_]

unused_return_value if {
	indexof("foo", "o")
}

zero_arity_function() := true

inconsistent_args(a, b) if {
	a == b
}

inconsistent_args(b, a) if {
	b == a
}

if_empty_object if {}

redundant_existence_check if {
	input.foo
	startswith(input.foo, "bar")
}

### Idiomatic ###

custom_has_key_construct(map, key) if {
	_ = map[key]
}

custom_in_construct(coll, item) if {
	item == coll[_]
}

use_some_for_output_vars if {
	input.foo[output_var]
}

non_raw_regex_pattern := regex.match("[0-9]", "1")

use_in_operator if {
	"item" == input.coll[_]
}

prefer_set_or_object_rule := {x | some x in input; x == "violation"}

equals_pattern_matching(x) := x == "x"

boolean_assignment := 1 < input.two

### Style ###

# avoid-get-and-list-prefix
get_foo(foo) := foo

# METADATA
# description: detached-metadata

annotation := "detached"

external_reference(_) if {
	data.foo
}

function_arg_return if {
	indexof("foo", "o", i)
	i == 1
}

line_length_should_be_no_longer_than_120_characters_but_this_line_is_really_long_and_will_exceed_that_limit_which_is := "bad"

#no-whitespace-comment

 opa_fmt := "fail"

preferSnakeCase := "fail"

# todo-comment

x := y if {
	y := 1
}

use_assignment = "oparator"

chained_rule_body if {
	input.x
} {
	input.y
}

rule_length if {
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
	input.x31
}

default_over_else := 1 if {
	input.x
} else := 3

default_over_not := input.foo

default_over_not := "foo" if not input.foo

unnecessary_some if {
	some "x" in ["x"]
}

yoda_condition if {
	"foo" == input.bar
}

### Testing ###

# this will also tringger the test-outside-test-package rule
test_identically_named_tests := true
test_identically_named_tests := true

todo_test_bad if {
	input.bad
}

print_or_trace_call if {
	print("forbidden!")
}

# metasyntactic variable
foo := "bar"

# dubious print sprintf
y if {
	print(sprintf("name is: %s domain is: %s", [input.name, input.domain]))
}
