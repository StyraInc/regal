# prefer-some-in-iteration

**Summary**: Prefer `some .. in` for iteration

**Category**: Style

**Avoid**
```rego
package policy

engineering_roles = {"engineer", "dba", "developer"}

engineers[employee] {
    employee := data.employees[_]
    employee.role in engineering_roles
}
```

**Prefer**
```rego
package policy

import rego.v1

engineering_roles = {"engineer", "dba", "developer"}

engineers[employee] {
    some employee in data.employees
    employee.role in engineering_roles
}
```

## Rationale

Using the `some .. in` construct for iteration removes ambiguity around iteration vs. membership checks, and is
generally more pleasant to read. Consider the following example:

```rego
some_condition if {
    other_rule[user]
    # ...
}
```

Are we iterating users over a partial "other_rule" here, or checking if the set contains a user defined elsewhere?
Or is `other_rule` a map-generating rule, and we're checking for the existence of a key? We won't know without looking
elsewhere in the code. Using `some .. in` removes this ambiguity, and makes the intent clear without having to jump
around in the policy.

## Exceptions

Deeply nested iteration is often easier to read using the more compact form.

```rego
package policy

import rego.v1

# These rules are equivalent, but the more compact form is arguably easier to read

any_user_is_admin if {
    some user in input.users
    some attribute in user.attributes
    some role in attribute.roles
    role == "admin"
}

any_user_is_admin if {
    input.users[_].attributes[_].roles[_] == "admin"
}

# Using "if", we may even omit the brackets for single line rules
any_user_is_admin if input.users[_].attributes[_].roles[_] == "admin"
```

The `ignore-nesting-level` configuration option allows setting the threshold for nesting. Any level of nesting
**equal or greater than** the threshold won't be considered a violation. The default setting of `2` allows all _nested_
iteration, but not e.g. `my_array[x]`.

**Note:** not all nesting is _iteration_! The following example is considered to have a nesting level of `1`, as only
one of the variables (including wildcards: `_`) is an output variable bound in iteration:

```rego
package policy

example_users[user] {
    domain := "example.com"
    user := input.sites[domain].users[_]
}
```

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  style:
    prefer-some-in-iteration:
      # one of "error", "warning", "ignore"
      level: error
      # except iteration if nested at or above the level i.e. setting of
      # '2' will allow `input[_].users[_]` but not `input[_]`
      ignore-nesting-level: 2
      # except iteration over items with sub-attributes, like
      # `name := input.users[_].name`
      # default is true
      ignore-if-sub-attribute: true
```

## Related Resources

- Rego Style Guide: [Prefer some .. in for iteration](https://github.com/StyraInc/rego-style-guide#prefer-some--in-for-iteration)
- Regal Docs: [Use `some` to declare output variables](https://docs.styra.com/regal/rules/idiomatic/use-some-for-output-vars)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
