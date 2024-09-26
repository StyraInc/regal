# leaked-internal-reference

**Summary**: Outside reference to internal rule or function

**Category**: Bugs

**Avoid**
```rego
package policy

# Import of rule or functions marked as internal
import data.users._all_users

allow if {
    # reference to rule or function marked as internal
    some role in data.permissions._roles
    # ...some conditions
}
```

## Rationale

OPA doesn't have a concept of "internal", or private rules and functions — and all rules can be queried or referenced
from the outside. Despite this fact, it has become a common convention to use an underscore prefix in the name of
rules and functions to indicate that they should be considered internal to the package that they're in:

```rego
# `allow` may be referenced from outside the package
allow if _user_is_developer

# `_user_is_developer` should not be referenced from outside the package
_user_is_developer if "developer" in input.users.roles
```

While this might seem like a pointless convention if it isn't enforced by OPA, it comes with a number of benefits:

- While OPA doesn't enforce it, other tools like linters can help with that. Like this rule does!
- It clearly communicates intent to other policy authors, and as a simple form of documentation
- Completion suggestions in editors can be filtered to exclude internal rules and functions
- Tools that render documentation from Rego policies and metadata annotations can exclude internal rules and functions
- Checking for unused rules and functions can be done much faster if they're known not to be referenced from outside

Do note that if you disagree with this rule, you don't need to disable it unless you use underscore prefixes to mean
something else. If you don't use underscore prefixes, nothing will be reported by this rule anyway. It does however
mean that the benefits listed above won't apply to your project.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  bugs:
    leaked-internal-reference:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- Rego Style Guide: [Optionally, use leading underscore for rules intended for internal use](https://docs.styra.com/opa/rego-style-guide#optionally-use-leading-underscore-for-rules-intended-for-internal-use)
- GitHub: [Source Code](https://github.com/StyraInc/regal/blob/main/bundle/regal/rules/bugs/leaked-internal-reference/leaked_internal_reference.rego)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
