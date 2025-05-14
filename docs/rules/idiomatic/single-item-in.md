# single-item-in

**Summary**: Avoid `in` for single item collection

**Category**: Idiomatic

**Avoid**
```rego
package policy

allow if input.role in {"admin"}
```

**Prefer**
```rego
package policy

allow if input.role == "admin"
```

## Rationale

Using `in` on a single-item collection (array, set or object) is a convoluted way of checking for equality. Better
then to check for equality directly! Besides being more obvious, equality checks are also subject to rule indexing,
whereas `in` checks currently aren't.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  idiomatic:
    single-item-in:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [Use indexed statements](https://www.openpolicyagent.org/docs/latest/policy-performance/#use-indexed-statements)
- GitHub: [Source Code](https://github.com/StyraInc/regal/blob/main/bundle/regal/rules/idiomatic/single-item-in/single_item_in.rego)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://inviter.co/styra)!
