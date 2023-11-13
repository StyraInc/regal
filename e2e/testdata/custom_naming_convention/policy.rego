# package name must start with "acmecorp"
package this.fails

import future.keywords.if

# rules must either start with "_" or be named "allow"
naming_convention_fail if {
	input.foo == "bar"
}

_this_is_ok := true

allow := true
