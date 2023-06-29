# no-whitespace-comment

**Summary**: Comment should start with whitespace

**Category**: Style

**Avoid**

```rego
package policy

import future.keywords.if
import future.keywords.in

#Deny by default
default allow := false

#Allow only admins
allow if "admin" in input.user.roles
```

**Prefer**

```rego
package policy

import future.keywords.if
import future.keywords.in

# Deny by default
default allow := false

# Allow only admins
allow if "admin" in input.user.roles
```

## Rationale

Comments should be preceded by a single space, as this makes them easier to read.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  style:
    no-whitespace-comment:
      # one of "error", "warning", "ignore"
      level: error
```
