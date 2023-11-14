# use-in-operator

**Summary**: Use `in` to check for membership

**Category**: Idiomatic

**Avoid**
```rego
package policy

import rego.v1

# "Old" way of checking for membership - iteration + comparison
allow if {
    "admin" == input.user.roles[_]
}
```

**Prefer**
```rego
package policy

import rego.v1

allow if {
    "admin" in input.user.roles
}
```

## Rationale

Using `in` for membership checks clearly communicates intent, and is less prone to errors. This is especially true when
checking if something is **not** part of a collection.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  idiomatic:
    use-in-operator:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- Rego Style Guide: [Use `in` to check for membership](https://github.com/StyraInc/rego-style-guide#use-in-to-check-for-membership)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
