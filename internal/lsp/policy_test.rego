package policy_test

import data.policy

test_allow if {
	policy.allow with input.bar as "wee" with input.foo as "woo"
}
