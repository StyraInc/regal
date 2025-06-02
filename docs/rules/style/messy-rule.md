> [!IMPORTANT]
> Please see [rules/style/messy-rule](https://docs.styra.com/regal/rules/style/messy-rule) on the Styra documentation website for the canonical representation of this page.

# messy-rule

**Summary**: Messy incremental rule

**Category**: Style

**Avoid**

```rego
package policy

allow if something

unrelated_rule if {
    # ...
}

allow if something_else
```

**Prefer**

```rego
package policy

allow if something

allow if something_else

unrelated_rule if {
    # ...
}
```

## Rationale

Rules that are defined incrementally should have their definitions grouped together, as this makes the code easier to
follow. While this is mostly a style preference, having incremental rules grouped also allows editors like VS Code to
"know" that the rules belong together, allowing them to be smarter when displaying the symbols of a workspace.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  style:
    messy-rule:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- GitHub: [Source Code](https://github.com/StyraInc/regal/blob/main/bundle/regal/rules/style/messy-rule/messy_rule.rego)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://inviter.co/styra)!
