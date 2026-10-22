# redundant-future-keywords

**Summary**: Redundant import of future keywords

**Category**: Imports

**Avoid**
```rego
package policy

import future.keywords.every
import future.keywords.in
```

**Prefer**
```rego
package policy

# 'every' and 'in' are standard keywords in recent OPA versions and do not require import
```

## Rationale

In recent versions of OPA (and Rego v1), keywords such as `contains`, `every`, `if`, and `in` have become standard built-in language keywords. Importing them via `future.keywords.<keyword>` is redundant and no longer required.

Only experimental or newly introduced future keywords (such as `and`, `not`, and `or` where supported by target capabilities) should be imported via `future.keywords`.

## Configuration Options

This linter rule provides the following [configuration options](https://www.openpolicyagent.org/projects/regal/configuration):

```yaml title=".regal/config.yaml or .regal.yaml"
rules:
  imports:
    redundant-future-keywords:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- GitHub: [Source Code](https://github.com/open-policy-agent/regal/blob/main/bundle/regal/rules/imports/redundant-future-keywords/redundant_future_keywords.rego)
