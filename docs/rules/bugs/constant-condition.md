# constant-condition

**Summary**: Constant condition

**Category**: Bugs

**Avoid**
```rego
package policy

import future.keywords.if

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
