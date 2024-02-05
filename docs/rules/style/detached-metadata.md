# detached-metadata

**Summary**: Detached metadata annotation

**Category**: Style

**Avoid**
```rego
package authz

import rego.v1

 # METADATA
 # description: allow any requests by admin users

allow if {
    "admin" in input.user.roles
}
```

**Prefer**
```rego
package authz

import rego.v1

# METADATA
# description: allow any requests by admin users
allow if {
    "admin" in input.user.roles
}
```

## Rationale

Metadata annotations should be placed directly above the package, rule or function they are annotating. While OPA
accepts any number of newlines between an annotation and the package/rule it applies to, this makes it difficult to
connect the two when reading the policy. Always optimize for readability!

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  style:
    detached-metadata:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [Annotations](https://www.openpolicyagent.org/docs/latest/policy-language/#annotations)
- OPA Docs: [Accessing Annotations](https://www.openpolicyagent.org/docs/latest/policy-language/#accessing-annotations)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
