# redundant-data-import

**Summary**: Redundant import of data

**Category**: Imports

**Avoid**
```rego
package policy

import data
```

## Rationale

Just like `input`, `data` is always globally available and does not need to be imported.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  imports:
    redundant-data-import:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- GitHub: [Source Code](https://github.com/open-policy-agent/regal/blob/main/bundle/regal/rules/imports/redundant-data-import/redundant_data_import.rego)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://inviter.co/styra)!
