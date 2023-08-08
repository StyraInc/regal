# import-shadows-import

**Summary**: Import shadows import

**Category**: Imports

**Avoid**
```rego
package policy

import data.permissions
import data.users

# Already imported
import data.permissions
```

**Prefer**
```rego
package policy

import data.permissions
import data.users
```

## Rationale

Duplicate imports are redundant, and while harmless, should just be removed.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  imports:
    import-shadows-import:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [Strict Mode](https://www.openpolicyagent.org/docs/latest/policy-language/#strict-mode)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
