# impossible-condition

**Summary**: Conflicting conditions in rule body

**Category**: Bugs

**Automatically fixable**: No

**Avoid**
```rego
package policy

import rego.v1

allow if {
    foo == "foo"
    foo == "bar"
}
```

**Prefer**
```rego
package policy

import rego.v1

allow if {
    foo in {"foo", "bar"}
}
```

## Rationale

When multiple equality conditions in the same rule body require the same variable or static reference
to equal different constant/scalar values, the rule can never evaluate to true. Such conflicting conditions
are dead code or indicate a logic bug that should be replaced with alternative conditions (such as the `in`
operator) or separate incremental rule bodies.

## Configuration Options

This linter rule provides the following [configuration options](https://www.openpolicyagent.org/projects/regal/configuration):

```yaml title=".regal/config.yaml or .regal.yaml"
rules:
  bugs:
    impossible-condition:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- GitHub: [Source Code](https://github.com/open-policy-agent/regal/blob/main/bundle/regal/rules/bugs/impossible-condition/impossible_condition.rego)
