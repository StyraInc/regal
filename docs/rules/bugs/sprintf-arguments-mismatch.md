# sprintf-arguments-mismatch

**Summary**: Mismatch in `sprintf` arguments count

**Category**: Bugs

**Avoid**
```rego
package policy

import rego.v1

max_issues := 1

report contains warning if {
    count(issues) > max_issues

    # two placeholders found in the string, but only one value in the array
    warning := sprintf("number of issues (%d) must not be higher than %d", [count(issues)])
}
```

**Prefer**
```rego
package policy

import rego.v1

max_issues := 1

report contains warning if {
    count(issues) > max_issues

    # two placeholders found in the string, and two values in the array
    warning := sprintf("number of issues (%d) must not be higher than %d", [count(issues), max_issues])
}
```

## Rationale

While the built-in `sprintf` function itself reports argument mismatches, it'll do so by returning a string containing
the error message rather than actually failing.

```shell
> opa eval -f pretty 'sprintf("%v %d", [1])'
"1 %!d(MISSING)"
```

While this is normally caught in development and testing, having this issue reported at "compile time", which ideally
is [directly in your editor](https://docs.styra.com/regal/language-server) as you work on your policy. This means less
time spent chasing down issues later, and a happier development experience.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  bugs:
    sprintf-arguments-mismatch:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [Built-in Functions: `sprintf`](https://www.openpolicyagent.org/docs/latest/policy-reference/#builtin-strings-sprintf)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
