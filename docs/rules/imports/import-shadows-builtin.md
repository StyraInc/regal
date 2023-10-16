# import-shadows-builtin

**Summary**: Import shadows built-in namespace

**Category**: Imports

**Avoid**
```rego
package policy

# Shadows the built-in `print` function
import data.print

# Shadows the built-in `http.send` function
import input.attributes.http
```

**Prefer**
To either use different names for your packages, or use import aliases to avoid shadowing built-ins.
```rego
package policy

# Using a different package name
import data.printer

# Using an alias
import input.attributes.http as http_attributes
```

## Rationale

OPA will not complain about an import shadowing the name or the "namespace" (i.e. `array` in `array.slice`) until a
conflicting built-in function is used in the same policy. Preventing this to happen in the first place is a better
option!

Why does this happen? The OPA compiler rewrites any import used in a policy, so that the shorthand form expands to its
longer form. Provided a simple policy like this:

```rego
package policy

import future.keywords.if

import data.http

allow if {
    http.send({"method": "GET", "url": "https://example.com"})
}
```

The compiler will go ahead and rewrite the `http.send` call using the import:

```rego
allow if {
    data.http.send({"method": "GET", "url": "https://example.com"})
}
```

This is obviously not what the policy author intended.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  imports:
    import-shadows-builtin:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [Built-in Functions](https://www.openpolicyagent.org/docs/latest/policy-reference/#built-in-functions)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
