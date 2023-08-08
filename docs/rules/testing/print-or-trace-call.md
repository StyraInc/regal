# print-or-trace-call

**Summary**: Call to `print` or `trace` function

**Category**: Testing

**Avoid**
```rego
package policy

import future.keywords.contains
import future.keywords.if
import future.keywords.in

reasons contains sprintf("%q is a dog!", [user.name]) if {
    some user in input.users
    user.species == "canine"

    # Useful for debugging, but leave out before commiting
    print("user:", user)
}
```

## Rationale

The `print` function is really useful for development and debugging, but should normally not be included in production
policy. In order to be as useful for debugging purposes as possible, some performance optimizations are disabled when
`print` calls are encountered. Prefer decision logging in production.

The `trace` function serves no real purpose since the introduction of `print`, and should be considered deprecated.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  testing:
    print-or-trace-call:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Blog: [Introducing the OPA print function](https://blog.openpolicyagent.org/introducing-the-opa-print-function-809da6a13aee)
- OPA Docs: [Policy Reference: Debugging](https://www.openpolicyagent.org/docs/latest/policy-reference/#debugging)
- OPA Docs: [Decision Logs](https://www.openpolicyagent.org/docs/latest/management-decision-logs/)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
