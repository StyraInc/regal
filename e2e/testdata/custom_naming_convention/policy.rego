# package name must start with "acmecorp"
package custom_naming_convention

import rego.v1

# rules must either start with "_" or be named "allow"
naming_convention_fail if {
	input.foo == "bar"
}

_this_is_ok := true

allow := true
