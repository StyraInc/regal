# if-object-literal

**Summary**: Object literal following `if`

**Category**: Bugs

**Avoid**
```rego
package policy

import rego.v1

# {} interpreted as object, not a rule body
allow if {}

allow if {
    # perhaps meant to be comparison?
    # but this too is an object
    input.x: 10
}
```

## Rationale

An object literal immediately following an `if` is almost certainly a mistake, and the intention was likely to express
a rule body in its place. This isn't too common, but can happen when either an empty object `{}` is all that follows the
`if`, or an expression in the "body" mistakenly is written as a `key: value` pair.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  bugs:
    if-object-literal:
      # one of "error", "warning", "ignore"
      level: error
```

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
