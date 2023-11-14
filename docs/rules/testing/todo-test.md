# todo-test

**Summary**: TODO test encountered

**Category**: Testing

**Avoid**
```rego
package policy_test

import rego.v1

import data.policy

# Make sure this passes
todo_test_allow_if_admin {
    policy.allow with input as {"user": {"roles": ["admin"]}}
}
```

## Rationale

Writing TODO tests by prefixing `todo_` to any test is a good way to keep track of tests that need to be written while
developing policy. They are however not to be committed, and should be removed before submitting the change for review.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  testing:
    todo-test:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [Policy Testing](https://www.openpolicyagent.org/docs/latest/policy-testing/)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
