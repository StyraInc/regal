# prefer-package-imports

**Summary**: Prefer importing packages over rules

**Category**: Imports

**Type**: Aggregate - only runs when more than one file is provided for linting

**Avoid**
```rego
package policy

import future.keywords.in

# Rule imported directly
import data.users.first_names

has_waldo {
    # Not obvious where "first_names" comes from
    "Waldo" in first_names 
}
```

**Prefer**
```rego
package policy

import future.keywords.in

# Package imported rather than rule
import data.users

has_waldo {
    # Obvious where "first_names" comes from
    "Waldo" in users.first_names 
}
```

## Rationale

Importing packages and using the package name as a "namespace" for imported rules and functions tends to make your code
easier to follow. This is especially true for large policies, where the distance from the import to actual use may be
several hundreds of lines.

## Exceptions

Regal has no way of know whether an import points to a rule, function or some external data — only that it doesn't point
to a package. Use the `ignore-import-paths` configuration option if you want to make exceptions for e.g. imports of
external data, or use the various ignore options to ignore entire files.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  imports:
    prefer-package-imports:
      # one of "error", "warning", "ignore"
      level: error
      ignore-import-paths:
        # Make an exception for some specific import paths
        - data.permissions.admin.users
```

## Related Resources

- Rego Style Guide: [Prefer importing packages over rules and functions](https://github.com/StyraInc/rego-style-guide#prefer-importing-packages-over-rules-and-functions)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
