# prefer-snake-case

**Summary**: Prefer snake_case for names

**Category**: Style

**Avoid**
```rego
package policy

import rego.v1

# camelCase rule name
userIsAdmin if "admin" in input.user.roles
```

**Prefer**
```rego
package policy

import rego.v1

# snake_case rule name
user_is_admin if "admin" in input.user.roles
```

## Rationale

The built-in functions use `snake_case` for naming — follow that convention for your own rules, functions, and
variables, unless you have a really good reason not to.

## Exceptions

In many cases, you might not control the format of the `input` data — if the domain of a policy (e.g. Envoy)
mandates a different style, making an exception might seem reasonable. Adapting policy format after `input` is however
prone to inconsistencies, as you'll likely end up mixing different styles in the same policy (due to imports of common
code, etc).

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  style:
    prefer-snake-case:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- Rego Style Guide: [Prefer snake_case for rule names and variables](https://github.com/StyraInc/rego-style-guide#prefer-snake_case-for-rule-names-and-variables)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
