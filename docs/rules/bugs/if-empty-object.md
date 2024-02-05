# if-empty-object

**Summary**: Empty object following `if`

**Category**: Bugs

**Avoid**
```rego
package policy

import rego.v1

allow if {}
```

## Rationale

An empty rule body would previously be considered an error by OPA. With the introduction, and use of the `if` keyword,
that is no longer the case. In fact, empty `{}` is not considered a rule body _at all_, but rather an empty object,
meaning that `if {}` will always evaluate. This is likely a mistake, and while hopefully caught by tests, should be
avoided.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  bugs:
    if-empty-object:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- Regal Docs: [constant-condition](https://docs.styra.com/regal/rules/bugs/constant-condition)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
