# rule-length

**Summary**: Max rule length exceeded

**Category**: Style

**Avoid**

Having too much logic placed in a single rule body.

**Prefer**

To use helper rules and functions to compose your rules.

## Rationale

Splitting up large rules into smaller ones, and liberally using helper rules and functions, makes your policy easier for
others to read and understand, and for yourself and your team to maintain.

Note that this rule only counts the number of lines of a rule, and currently does not take into account the actual
content inside of it. Neither does it try to analyze the complexity of the code in the rule.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  style:
    rule-length:
      # one of "error", "warning", "ignore"
      level: error
      # default limit is 30 lines
      max-rule-length: 30
      # whether to count comments as lines
      # by default, this is set to false
      count-comments: false
      # except rules with empty bodies from this rule, as they're
      # likely an assignment of long values rather than a "rule"
      # with conditions:
      #
      # users := [
      #     {"username": "ted"},
      #     {"username": "alice"},
      #     {"username": "bob"},
      #     # ... many more lines
      # ]
      #
      # the default value is true
      except-empty-body: true
```

## Related Resources

- Regal Docs: [file-length](https://docs.styra.com/regal/rules/style/file-length)
- Regal Docs: [line-length](https://docs.styra.com/regal/rules/style/line-length)
- GitHub: [Source Code](https://github.com/StyraInc/regal/blob/main/bundle/regal/rules/style/rule-length/rule_length.rego)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
