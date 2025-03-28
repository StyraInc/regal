# constant-condition

**Summary**: Constant condition

**Category**: Bugs

**Avoid**
```rego
package policy

allow if {
    1 == 1
}
```

**Prefer**
```rego
package policy

allow := true
```

## Rationale

While most often a mistake, constant conditions are sometimes used as placeholders, or "TODO logic". While this is
harmless, it has no place in production policy, and should be replaced or removed before deployment.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  bugs:
    constant-condition:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- GitHub: [Source Code](https://github.com/StyraInc/regal/blob/main/bundle/regal/rules/bugs/constant-condition/constant_condition.rego)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://inviter.co/styra)!
