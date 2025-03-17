# import-after-rule

**Summary**: Import declared after rule

**Category**: Imports

**Avoid**
```rego
package policy

required_role := "developer"

import data.identity.users
```

**Prefer**
```rego
package policy

import data.identity.users

required_role := "developer"
```

## Rationale

Imports should be declared at the top of a policy, and before any rules. This makes it easy to quickly see the
dependencies imported in the policy simply by looking at the top of the file.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  imports:
    import-after-rule:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- GitHub: [Source Code](https://github.com/StyraInc/regal/blob/main/bundle/regal/rules/imports/import-after-rule/import_after_rule.rego)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://inviter.co/styra)!
