# deprecated-builtin

**Summary**: Constant condition

**Category**: Bugs

**Avoid**
```rego
package policy

import future.keywords.if

# call to deprecated `any` built-in function
allow if any([input.user.is_admin, input.user.is_root])
```

**Prefer**
```rego
package policy

import future.keywords.if

allow if input.user.is_admin
allow if input.user.is_root
```

## Rationale

Calling deprecated built-in functions should always be avoided, and replacing them is usually trivial.
Refer to the OPA docs on [strict mode](https://www.openpolicyagent.org/docs/latest/policy-language/#strict-mode)
for more details on which built-in functions counts as deprecated.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  bugs:
    deprecated-builtin:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [Strict Mode](https://www.openpolicyagent.org/docs/latest/policy-language/#strict-mode)
- GitHub: [Source Code](https://github.com/StyraInc/regal/blob/main/bundle/regal/rules/bugs/deprecated-builtin/deprecated_builtin.rego)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
