# use-object-keys

**Summary**: Prefer to use `object.keys`

**Category**: Idiomatic

**Avoid**
```rego
package policy

keys := {k | some k, _ in input.object}

# or

keys := {k | some k; input.object[k]}
```

**Prefer**
```rego
package policy

keys := object.keys(input.object)
```

## Rationale

Instead of using a set comprehension to collect keys from an object, prefer to use the built-in function
[object.keys](https://www.openpolicyagent.org/docs/latest/policy-reference/#builtin-object-objectkeys).
This option is both more declarative and better conveys the intent of the code.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  idiomatic:
    use-object-keys:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [object.keys](https://www.openpolicyagent.org/docs/latest/policy-reference/#builtin-object-objectkeys)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
