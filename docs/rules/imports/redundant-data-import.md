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
