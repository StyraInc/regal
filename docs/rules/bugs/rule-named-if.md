# rule-named-if

**Summary**: Rule named `if`

**Category**: Bugs

**Avoid**
```rego
package policy

allow := true if {
    authorized
}
```

**Prefer**
```rego
package policy

import rego.v1

allow := true if {
    authorized
}
```

## Rationale

Forgetting to import the `if` keyword (using `import future.keywords.if`, or from OPA v0.59.0+ `import rego.v1`) is a
common mistake. While this often results in a parse error, there are some situations where the parser can't tell if the
`if` is intended to be used as the imported keyword, or a new rule named `if`. This is almost always a mistake, and if
it isn't — consider using a better name for your rule!

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  bugs:
    rule-named-if:
      # one of "error", "warning", "ignore"
      level: error
```

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
