# yoda-condition

**Summary**: Yoda condition, it is

**Category**: Style

**Avoid**
```rego
package policy

import rego.v1

allow if {
    "GET" == input.request.method
    "users" == input.request.path[0]
}
```

**Prefer**
```rego
package policy

import rego.v1

allow if {
    input.request.method == "GET"
    input.request.path[0] == "users"
}
```

## Rationale

Yoda conditions — expressions where the constant portion of a comparison is placed on the left-hand side of the
comparison — provide no benefits in Rego. They do however add a certain amount of cognitive overhead for most policy
authors in the galaxy.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  style:
    yoda-condition:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- Wikipedia: [Yoda conditions](https://en.wikipedia.org/wiki/Yoda_conditions)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
