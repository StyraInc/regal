# use-contains

**Summary**: Use the `contains` keyword

**Category**: Idiomatic

**Avoid**
```rego
package policy

import future.keywords.in

report[item] if {
    some item in input.items
    startswith(item, "report")
}

# unconditinally add an item to report
report["report1"]
```

**Prefer**
```rego
package policy

import future.keywords.contains
import future.keywords.if
import future.keywords.in

report contains item if {
    some item in input.items
    startswith(item, "report")
}

# unconditinally add an item to report
report contains "report1"
```

## Rationale

The `contains` keyword helps to clearly distinguish *multi-value rules* (or "partial rules") from
single-value rules ("complete rules"). Just like the `if` keyword, `contains` additionally makes the rule read the same
way in English as OPA interprets its meaning — a set that contains one or more values given some (optional) conditions.

**Note**: don't forget to `import future.keywords.contains`! Or from OPA v0.59.0 and onwards, `import rego.v1`.

**Tip**: When either of the imports mentioned above are found in a Rego file, the `contains` keyword will be inserted
automatically at any applicable location by the `opa fmt` tool.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  idiomatic:
    use-contains:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- Regal Docs: [use-if](https://docs.styra.com/regal/rules/idiomatic/use-if)
- OPA Docs: [Future Keywords](https://www.openpolicyagent.org/docs/latest/policy-language/#future-keywords)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
