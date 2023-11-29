# redundant-existence-check

**Summary**: Redundant existence check

**Category**: Bugs

**Avoid**
```rego
package policy

import rego.v1

employee if {
    input.user.email
    endswith(input.user.email, "@acmecorp.com")
}
```

**Prefer**
```rego
package policy

import rego.v1

employee if {
    endswith(input.user.email, "@acmecorp.com")
}

# alternatively

employee if endswith(input.user.email, "@acmecorp.com")
```

## Rationale

Checking that a reference (like `input.user.email`) is defined before immediately using it is redundant. If the
reference is undefined, the next expression will fail anyway, as the value will be checked before the rest of the
expression is evaluated. While an extra check doesn't "hurt", it also serves no purpose, similarly to an unused
variable.

**Note**: This rule only applies to references that are immediately used in the next expression. If the reference is
used later in the rule, it won't be flagged. While the existence check _could_ be redundant even in that case, it could
also be used to avoid making some expensive computation, an `http.send` call, or whatnot.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  bugs:
    redundant-existence-check:
      # one of "error", "warning", "ignore"
      level: error
```

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
