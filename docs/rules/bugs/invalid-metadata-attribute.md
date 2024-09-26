# invalid-metadata-attribute

**Summary**: Invalid attribute in metadata annotation

**Category**: Bugs

**Avoid**
```rego
# METADATA
# title: Main policy routing requests to other policies based on input
# category: Routing
package router
```

**Prefer**
```rego
# METADATA
# title: Main policy routing requests to other policies based on input
# custom:
#   category: Routing
package router
```

## Rationale

Metadata comments should follow the schema expected by
[annotations](https://www.openpolicyagent.org/docs/latest/policy-language/#annotations). Custom attributes, like
`category` above, should be placed under the `custom` key, which is a map of arbitrary key-value pairs.

While arbitrary attributes are accepted, they will not be treated as metadata annotations but regular comments, and as
such won't be available to other tools that
[process annotations](https://www.openpolicyagent.org/docs/latest/policy-language/#accessing-annotations).
These tools include built-in functions like
[rego.metadata.rule](https://www.openpolicyagent.org/docs/latest/policy-reference/#builtin-rego-regometadatarule) and
[rego.metadata.chain](https://www.openpolicyagent.org/docs/latest/policy-reference/#builtin-rego-regometadatachain).

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  bugs:
    invalid-metadata-attribute:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [Annotations](https://www.openpolicyagent.org/docs/latest/policy-language/#annotations)
- OPA Docs: [Accessing Annotations](https://www.openpolicyagent.org/docs/latest/policy-language/#accessing-annotations)
- GitHub: [Source Code](https://github.com/StyraInc/regal/blob/main/bundle/regal/rules/bugs/invalid-metadata-attribute/invalid_metadata_attribute.rego)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
