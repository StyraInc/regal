# opa-fmt

**Summary**: File should be formatted with `opa fmt`

**Category**: Style

**Avoid**

Inconsistent style across policy files and repositories.

## Rationale

The `opa fmt` tool ensures consistent formatting across teams and projects. Unified formatting is a big win, and saves a
lot of time in code reviews arguing over details around style.

A good idea could be to run `opa fmt --write` on save, which can be configured in most editors.

**Tip**: `opa fmt` uses tabs for indentation. By default, GitHub uses 8 spaces to display tabs, which is arguably a bit
much. You can change this preference for your account in `github.com/settings/appearance`, or provide an `.editorconfig`
file in your policy repository, which will be used by GitHub (and other tools) to properly display your Rego files:

```ini
[*.rego]
end_of_line = lf
insert_final_newline = true
charset = utf-8
indent_style = tab
indent_size = 4
```

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  style:
    opa-fmt:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs [CLI Reference `opa fmt`](https://www.openpolicyagent.org/docs/latest/cli/#opa-fmt)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
