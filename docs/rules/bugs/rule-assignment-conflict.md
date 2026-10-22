# rule-assignment-conflict

**Summary**: Rule assignment conflict

**Category**: Bugs

**Avoid**
```rego
package policy

# two unconditional definitions for the same rule
x := 1
x := 2
```

**Prefer**
```rego
package policy

# use conditions or default to define rule values
default x := 1

x := 2 if input.condition
```

## Rationale

Complete rules must evaluate to a single value. When multiple unconditional definitions are provided for the
same rule (or function with catch-all arguments), they will either fail at evaluation time with
`eval_conflict_error: complete rules must not produce multiple outputs` (if the values differ), or introduce
redundant and confusing assignments.

If a fallback value is needed, use the `default` keyword rather than multiple unconditional definitions.
If different values should be returned depending on conditions, use conditional rule definitions (`if`).

## Configuration Options

This linter rule provides the following [configuration options](https://www.openpolicyagent.org/projects/regal/configuration):

```yaml title=".regal/config.yaml or .regal.yaml"
rules:
  bugs:
    rule-assignment-conflict:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [Complete Definitions](https://www.openpolicyagent.org/docs/policy-language/#complete-definitions)
- GitHub: [Source Code](https://github.com/open-policy-agent/regal/blob/main/bundle/regal/rules/bugs/rule-assignment-conflict/rule_assignment_conflict.rego)
