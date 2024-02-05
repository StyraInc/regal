# test-outside-test-package

**Summary**: Test outside of test package

**Category**: Testing

**Avoid**
```rego
package policy

import rego.v1

allow if {
    "admin" in input.user.roles
}

# Tests in same package as policy
test_allow_if_admin {
    allow with input as {"user": {"roles": ["admin"]}}
}
```

**Prefer**
```rego
# Tests in separate package with _test suffix
package policy_test

import rego.v1

import data.policy

test_allow_if_admin {
    policy.allow with input as {"user": {"roles": ["admin"]}}
}
```

## Rationale

While OPA's test runner will evaluate any rules with a `test_` prefix, it is a good practice to clearly separate tests
from production policy. This is easily done by placing tests in a separate package with a `_test` suffix, and correctly
[naming](./file-missing-test-suffix.md) the test files.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  testing:
    test-outside-test-package:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [Policy Testing](https://www.openpolicyagent.org/docs/latest/policy-testing/)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
