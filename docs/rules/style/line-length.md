# line-length

**Summary**: Line too long

**Category**: Style

**Avoid**

Excessive line length.

## Rationale

Rego does not have many nested constructs, and long lines of code are thus almost never needed. If you find yourself
close to the maximum line length, consider refactoring your policy.

The default maximum line length is 120 characters.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  style:
    line-length:
      # one of "error", "warning", "ignore"
      level: error
      # maximum line length
      max-line-length: 120
```
