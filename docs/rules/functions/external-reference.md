# external-reference

**Summary**: Reference to input, data or rule ref in function body

**Category**: Functions

**Avoid**
```rego
package policy

# Depends on both `input` and `data`
is_preferred_login_method(method) if {
    preferred_login_methods := {login_method |
        some login_method in data.authentication.all_login_methods
        login_method in input.user.login_methods
    }
    method in preferred_login_methods
}
```

**Prefer**

```rego
package policy

# Depends only on function arguments
is_preferred_login_method(method, user, all_login_methods) if {
    preferred_login_methods := {login_method |
        some login_method in all_login_methods
        login_method in user.login_methods
    }
    method in preferred_login_methods
}
```

## Rationale

What separates functions from rules is that they accept arguments. While a function too may reference anything from
`input`, `data` or other rules declared in a policy, these references create dependencies that aren't obvious simply by
checking the function signature, and it makes it harder to reuse that function in other contexts. Additionally,
functions that only depend on their arguments are easier to test standalone.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  functions:
    external-reference:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- Rego Style Guide: [Prefer using arguments over input, data or rule references](https://github.com/StyraInc/rego-style-guide#prefer-using-arguments-over-input-data-or-rule-references)
