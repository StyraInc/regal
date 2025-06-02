> [!IMPORTANT]
> Please see [rules/idiomatic/in-wildcard-key](https://docs.styra.com/regal/rules/idiomatic/in-wildcard-key) on the Styra documentation website for the canonical representation of this page.

# in-wildcard-key

**Summary**: Unnecessary wildcard key

**Category**: Idiomatic

**Avoid**
```rego
package policy

allow if {
    # since only the value is used, we don't need to iterate the keys
    some _, user in input.users

    # do something with each user
}
```

**Prefer**
```rego
package policy

allow if {
    some user in input.users

    # do something with each user
}
```

## Rationale

The `some .. in` iteration form can either iterate only values:

```rego
some value in object
```

Or keys and values:

```rego
some key, value in object
```

Using a wildcard variable for the key in the key-value form is thus unnecessary, and:

```rego
some _, value in object
```

Can simply be replaced by:

```rego
some value in object

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  idiomatic:
    in-wildcard-key:
      # one of "error", "warning", "ignore"
      level: error
```

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://inviter.co/styra)!
