# pointless-reassignment

**Summary**: Pointless reassignment of variable

**Category**: Style

**Avoid**
```rego
package policy

allow if {
    users := all_users
    any_admin(users)
}
```

**Prefer**
```rego
package policy

allow if {
    any_admin(all_users)
}
```

## Rationale

Values and variables are immutable in Rego, so reassigning the value of one variable to another only adds noise.

## Exceptions

Reassigning the value of a long reference often helps readability, and especially so when it needs to be referenced
multiple times:

```rego
package policy

allow if {
    users := input.context.permissions.users
    any_admin(users)
}
```

This rule does not consider such assignments violations.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  style:
    pointless-reassignment:
      # one of "error", "warning", "ignore"
      level: error
```

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
