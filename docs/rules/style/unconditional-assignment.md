# unconditional-assignment

**Summary**: Unconditional assignment in rule body

**Category**: Style

**Avoid**
```rego
package policy

full_name := name {
    name := concat(", ", [input.first_name, input.last_name])
}

divide_by_ten(x) := y {
    y := x / 10
}
```

**Prefer**
```rego
package policy

full_name := concat(", ", [input.first_name, input.last_name])

divide_by_ten(x) := x / 10
```

## Rationale

Rules that return values unconditionally should place the assignment directly in the rule head, as doing so in the rule
body adds unnecessary noise.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  style:
    unconditional-assignment:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- Rego Style Guide: [Prefer unconditional assignment in rule head over rule body](https://github.com/StyraInc/rego-style-guide#prefer-unconditional-assignment-in-rule-head-over-rule-body)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
