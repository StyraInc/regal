# unused-return-value

**Summary**: Non-boolean return value unused

**Category**: Bugs

**Avoid**
```rego
package policy

import rego.v1

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
    unused-return-value:
      # one of "error", "warning", "ignore"
      level: error
```

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
