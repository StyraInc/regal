# implicit-future-keywords

**Summary**: Implicit future keywords

**Category**: Imports

**Avoid**
```rego
package policy

import future.keywords

report contains violation if {
    not "developer" in input.user.roles

    violation := "Required role 'developer' missing"
}
```

**Prefer**
```rego
package policy

import future.keywords.contains
import future.keywords.if
import future.keywords.in

report contains violation if {
    not "developer" in input.user.roles

    violation := "Required role 'developer' missing"
}
```

## Rationale

Using the "catch all" import of `future.keywords` is convenient, but it can lead to unexpected behavior. If future
versions of OPA introduces new keywords, there's always a risk that these keywords will conflict with existing rule and
variable names in your policy.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  imports:
    implicit-future-keywords:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- Rego Style Guide: [Use explicit imports for future keywords](https://github.com/StyraInc/rego-style-guide#use-explicit-imports-for-future-keywords)
- GitHub: [Source Code](https://github.com/StyraInc/regal/blob/main/bundle/regal/rules/imports/implicit-future-keywords/implicit_future_keywords.rego)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://inviter.co/styra)!
