# unused-return-value

**Summary**: Non-boolean return value unused

**Category**: Bugs

**Avoid**
```rego
package policy

import future.keywords.if

allow if {
    # return value unused
    lower(input.user.name)
    # ...
}
```

**Prefer**
```rego
package policy

allow if {
    # return value assigned
    name_lower := lower(input.user.name)
    # ...
}
```

## Rationale

Calling a built-in function that returns a non-boolean value without actually *using* the returned value is almost
always a mistake. Only return of `false` or undefined will cause evaluation to halt, so a function that e.g. always
returns a string will always be evaluated as "truthy". But more importantly — not handling the return value in that case
is almost certainly a mistake.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  bugs:
    return-value-unused:
      # one of "error", "warning", "ignore"
      level: error
```
