# annotation-without-metadata

**Summary**: Annotation without metadata

**Category**: Bugs

**Avoid**
```rego
package policy

import rego.v1

# description: allow allows
allow if {
    # ... some conditions
}
```

**Prefer**
```rego
package policy

import rego.v1

# METADATA
# description: allow allows
allow if {
    # ... some conditions
}
```

## Rationale

A comment that starts with `<annotation-attribute>:` but is not part of a metadata block is likely a mistake. Add
`# METADATA` above the line to turn it into a
[metadata](https://www.openpolicyagent.org/docs/latest/policy-language/#annotations) block.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  bugs:
    annotation-without-metadata:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [Annotations](https://www.openpolicyagent.org/docs/latest/policy-language/#annotations)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
