# non-raw-regex-pattern

**Summary**: Use raw strings for regex patterns

**Category**: Idiomatic

**Avoid**
```rego
all_digits if {
    regex.match("[\\d]+", "12345")
}
```

**Prefer**
```rego
all_digits if {
    regex.match(`[\d]+`, "12345")
}
```

## Rationale

[Raw strings](https://www.openpolicyagent.org/docs/edge/policy-language/#strings) are interpreted literally, allowing
you to avoid having to escape special characters like `\` in your regex patterns. Using raw strings for regex patterns
additionally makes them easier to identify as such.

## Limitations

This rule currently only scans regex string literals in the place of the `pattern` argument of the various
[regex built-in functions](https://www.openpolicyagent.org/docs/latest/policy-reference/#regex). It will not **not**
try to "resolve" patterns assigned to variables. The following example would as such not render a warning:

```rego
package policy

import rego.v1

# Pattern assigned to variable
pattern := "[\\d]+"

# This won't trigger a violation
allow if regex.match(pattern, "12345")
```

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  idiomatic:
    non-raw-regex-pattern:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- Rego Style Guide: [Use raw strings for regex patterns](https://github.com/StyraInc/rego-style-guide#use-raw-strings-for-regex-patterns)
- OPA Docs: [Regex Functions Reference](https://www.openpolicyagent.org/docs/latest/policy-reference/#regex)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
