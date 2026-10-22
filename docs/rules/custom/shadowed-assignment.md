# shadowed-assignment

**Summary**: Shadowed assignment

**Category**: Custom

**Automatically fixable**: No

**Avoid**
```rego
package policy

import rego.v1

x := 1

r if {
    x := 4
    x == 4
}
```

**Prefer**
```rego
package policy

import rego.v1

x := 1

r if {
    y := 4
    y == 4
}
```

## Rationale

Assigning to a variable in a rule body that has the same name as a package-level rule,
function, or imported identifier shadows the outer symbol. While permitted by OPA,
this can cause confusion as readers may mistake the assigned local variable for the
outer symbol, or assume modifications affect package state.

## Configuration Options

This linter rule provides the following [configuration options](https://www.openpolicyagent.org/projects/regal/configuration):

```yaml title=".regal/config.yaml or .regal.yaml"
rules:
  custom:
    shadowed-assignment:
      # one of "error", "warning", "ignore"
      level: ignore
```

## Related Resources

- GitHub: [Source Code](https://github.com/open-policy-agent/regal/blob/main/bundle/regal/rules/custom/shadowed-assignment/shadowed_assignment.rego)
