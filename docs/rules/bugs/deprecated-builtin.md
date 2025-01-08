# deprecated-builtin

**Summary**: Constant condition

**Category**: Bugs

## Notice: Rule disabled with OPA 1.0

Since Regal v0.30.0, this rule is only enabled for projects that have either been explicitly configured to target
versions of OPA before 1.0, or if no configuration is provided — where Regal is able to determine that an older version
of OPA/Rego is being targeted. Consult the documentation on Regal's
[configuration](https://docs.styra.com/regal#configuration) for information on how to best work with older versions of
OPA and Rego.

Since OPA v1.0, this rule is automatically disabled, as there currently are no deprecated built-in functions
in that version, and trying to use a previously deprecated function will result in a parser error. Note however that
this may change if later OPA versions deprecate current built-in functions. If/when that happens, this rule will be
re-enabled.

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
