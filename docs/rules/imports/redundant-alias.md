# redundant-alias

**Summary**: Redundant alias

**Category**: Imports

**Avoid**
```rego
package policy

import data.users.permissions as permissions
```

**Prefer**
```rego
package policy

import data.users.permissions
```

## Rationale

The last component of an import path is always made referencable by its name inside the package in which it's imported.
Using an alias with the same name is thus redundant, and should be omitted.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  imports:
    redundant-alias:
      # one of "error", "warning", "ignore"
      level: error
```

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
