# avoid-get-and-list-prefix

**Summary**: Avoid `get_` and `list_` prefix for rules and functions

**Category**: Style

**Avoid**
```rego
package policy

get_first_name(user) := split(user.name, " ")[0]

# Partial rule, so a set of users is to be expected
list_developers contains user if {
    some user in data.application.users
    user.type == "developer"
}
```

**Prefer**
```rego
package policy

# "get" is implied
first_name(user) := split(user.name, " ")[0]

# Partial rule, so a set of users is to be expected
developers contains user if {
    some user in data.application.users
    user.type == "developer"
}
```

**Exceptions**

Using `is_`, or `has_` for boolean helper functions, like `is_admin(user)` may be easier to comprehend than
`admin(user)`.

## Rationale

Since Rego evaluation is generally free of side effects, any rule or function is essentially a "getter". Adding a
`get_` prefix to a rule or function (like `get_resources`) thus adds little of value compared to just naming it
`resources`. Additionally, the type and return value of the rule should serve to tell whether a rule might return a
single value (i.e. a complete rule) or a collection (a partial rule).

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  style:
    avoid-get-and-list-prefix:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- Rego Style Guide: [Avoid prefixing rules and functions with `get_` or `list_`](https://github.com/StyraInc/rego-style-guide#avoid-prefixing-rules-and-functions-with-get_-or-list_)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
