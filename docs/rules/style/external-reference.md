# external-reference

**Summary**: External reference in function

**Category**: Style

**Avoid**
```rego
package policy

import rego.v1

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

import rego.v1

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

## Exceptions

Rego does not provide first-class functions — functions can't be passed as arguments to other functions. Therefore, this
rule allows functions to freely reference (i.e. call) _other functions_, whether built-in functions, or custom functions
defined in the same package or elsewhere, and these do not count as "external references" simply because there is not
other way to import them into the function body.

```rego
package policy

import rego.v1

first_name(full_name) := capitalized {
    first_name := split(full_name, " ")[0]

    # while data.utils.capitalize is an external reference, it's not flagged
    # as such, since there is no way to import it via function arguments
    capitalized := data.utils.capitalize(first_name)
}
```

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  style:
    external-reference:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- Rego Style Guide: [Prefer using arguments over input, data or rule references](https://github.com/StyraInc/rego-style-guide#prefer-using-arguments-over-input-data-or-rule-references)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
