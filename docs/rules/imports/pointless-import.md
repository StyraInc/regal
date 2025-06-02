> [!IMPORTANT]
> Please see [rules/imports/pointless-import](https://docs.styra.com/regal/rules/imports/pointless-import) on the Styra documentation website for the canonical representation of this page.

# pointless-import

**Summary**: Importing own package is pointless

**Category**: Imports

**Avoid**
```rego
package policy

# pointless, as policy is the own package
import data.policy

# pointless, as rules in own package can be referenced without the import
import data.policy.rule
```

**Prefer**
```rego
package policy
```

## Rationale

There's no point importing the own package, or rules from the same package, as both can be referenced just as well
without the import.

## Exceptions

While it may not be the best way use a reference from the same package, longer references than the package, or the
package plus a rule, are at least not pointless, and as such not flagged by this rule.

```rego
package policy

# this is allowed, but consider using the reference directly rather than importing it
import data.policy.a.b.c
```

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  imports:
    pointless-import:
      # one of "error", "warning", "ignore"
      level: error
```

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://inviter.co/styra)!
