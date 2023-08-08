# function-arg-return

**Summary**: Return value assigned in function argument

**Category**: Style

**Avoid**
```rego
package policy

import future.keywords.if

has_email(user) if {
    indexof(user.email, "@", i)
    i != -1
}
```

**Prefer**

```rego
package policy

import future.keywords.if

has_email(user) if {
    i := indexof(user.email, "@")
    i != -1
}
```

## Rationale

Older Rego policies sometimes contain an unusual way to declare where the return value of a function call should be
stored — the last argument of the function. True to it's Datalog roots, return values may be stored either using 
assignment (i.e. `:=`) or by appending a variable name to the argument list of a function. While both forms are valid,
using assignment `:=` consistently is preferred. 

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  style:
    function-arg-return:
      # one of "error", "warning", "ignore"
      level: error
      # list of function names to ignore
      # note that `print` is always ignored as it does not return a value
      except-functions:
        - walk
```

## Related Resources

- Rego Style Guide: [Avoid using the last argument for the return value](https://github.com/StyraInc/rego-style-guide#avoid-using-the-last-argument-for-the-return-value)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
