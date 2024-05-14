# {{.NameOriginal}}

**Summary**: ADD DESCRIPTION HERE

**Category**: {{.Category}}

**Avoid**
```rego
package policy

# ... ADD CODE TO AVOID HERE
```

**Prefer**
```rego
package policy

# ... ADD CODE TO PREFER HERE
```

## Rationale

ADD RATIONALE HERE

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  {{.Category}}:
    {{.NameOriginal}}:
      # one of "error", "warning", "ignore"
      level: error
```

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
