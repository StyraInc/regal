# one-liner-rule

**Summary**: Rule body could be made a one-liner

**Category**: Custom

**Avoid**
```rego
package policy

import rego.v1

allow if {
    is_admin
}

is_admin if {
    "admin" in input.user.roles
}
```

**Prefer**
```rego
package policy

import rego.v1

allow if is_admin

is_admin if "admin" in input.user.roles
```

## Rationale

Rules with only a single expression in the body may omit the curly braces around the body, and be written as a
one-liner. This makes simple rules read more like English, and will have more rules fit on the screen.

As with other rules in the `custom` category, this is not necessarily a general recommendation, but a style preference
teams or organizations might want to standardize on. As such, it must be enabled via configuration.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  custom:
    one-liner-rule:
      # note that all rules in the "custom" category are disabled by default
      # (i.e. level "ignore")
      #
      # one of "error", "warning", "ignore"
      level: error
```

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
