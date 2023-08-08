# file-missing-test-suffix

**Summary**: Files containing tests should have a `_test.rego` suffix

**Category**: Testing

## Rationale

In order to clearly communicate intent, and to avoid bundling tests with production policy, tests should be kept in a
separate file with a `_test.rego` suffix, and ideally prefixed with the same name as the policy the tests are targeting,
e.g. `policy.rego` and `policy_test.rego`.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  testing:
    file-missing-test-suffix:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [Policy Testing](https://www.openpolicyagent.org/docs/latest/policy-testing/)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
