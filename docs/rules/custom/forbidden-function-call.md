# forbidden-function-call

**Summary**: Forbidden function call

**Category**: Custom

## Description

This custom rule allows providing Regal a list of
[built-in functions](https://www.openpolicyagent.org/docs/latest/policy-reference/#built-in-functions) that should be
considered forbidden. Any call to a function in the list will be reported as a violation.

Another, more advanced, option to achieve the same result is the
[capabilities](https://www.openpolicyagent.org/docs/latest/deployments/#capabilities) feature in OPA. While a more
capable option, allowing things like:

- Adding new custom built-in functions that OPA should be aware of
- Disabling certain features not necessarily being built-in functions, like "future" keywords
- List allowed hosts in network calls

...it is also more demanding to configure and maintain. If you're already using the capabilities feature
to forbid certain functions as part of your policy development process, there's no need to enable this rule.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  custom:
    forbidden-function-call:
      # note that all rules in the "custom" category are disabled by default
      # (i.e. level "ignore") as some configuration needs to be provided by
      # the user (i.e. you!) in order for them to be useful.
      #
      # one of "error", "warning", "ignore"
      level: error
      # Just an example — no functions forbidden by default
      forbidden-functions:
        # Prefer to use asymmetric algorithms
        - io.jwt.verify_hs256
        - io.jwt.verify_hs384
        - io.jwt.verify_hs512
```

## Related Resources

- OPA Docs: [Capabilities](https://www.openpolicyagent.org/docs/latest/deployments/#capabilities)
- OPA Docs: [Built-in Functions](https://www.openpolicyagent.org/docs/latest/policy-reference/#built-in-functions)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
