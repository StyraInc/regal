# ignored-import

**Summary**: Reference ignores import

**Category**: Imports

**Avoid**
```rego
package policy

import data.authz.roles

allow if {
    some role in input.user.roles
    # data.authz.roles has been imported, but the import is ignored here
    role in data.authz.roles.admin_roles
}
```

**Prefer**
```rego
package policy

import data.authz.roles

allow if {
    some role in input.user.roles
    # imported data.authz.roles used
    role in roles.admin_roles
}
```

## Rationale

Imports tend to make long, nested references more readable, and encourages reuse of common logic. Using a full reference
(like `data.users.permissions`) despite having previously imported the reference, or parts of it (like `data.users`)
defeats the purpose of the import, and you're better off referring to the import directly.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  imports:
    ignored-import:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- GitHub: [Source Code](https://github.com/StyraInc/regal/blob/main/bundle/regal/rules/imports/ignored-import/ignored_import.rego)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
