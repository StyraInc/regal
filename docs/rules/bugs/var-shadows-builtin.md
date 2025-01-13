# var-shadows-builtin

**Summary**: Variable shadows built-in

**Category**: Bugs

**Avoid**
```rego
package policy

# variable `http` shadows `http.send` built-in function
allow if {
    http := startswith(input.url, "http://")
    # do something with http
}
```

**Prefer**
```rego
package policy

# variable `is_http` doesn't shadow any built-in function
allow if {
    is_http := startswith(input.url, "http://")
    # do something with is_http
}
```

## Rationale

Using the name of built-in functions or operators as variable names can lead to confusion and unexpected behavior.
A variable that shadows a built-in function (or the namespace of a function, like `http` in `http.send`) prevents any
function in that namespace to be used later in the rule. Avoid this!

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  bugs:
    var-shadows-builtin:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [Built-in Functions](https://www.openpolicyagent.org/docs/latest/policy-reference/#built-in-functions)
- OPA Repo: [builtin_metadata.json](https://github.com/open-policy-agent/opa/blob/main/builtin_metadata.json)
- Regal Docs: [rule-shadows-builtin](https://docs.styra.com/regal/rules/bugs/rule-shadows-builtin)
- GitHub: [Source Code](https://github.com/StyraInc/regal/blob/main/bundle/regal/rules/bugs/var-shadows-builtin/var_shadows_builtin.rego)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
