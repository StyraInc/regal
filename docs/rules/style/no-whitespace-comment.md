# no-whitespace-comment

**Summary**: Comment should start with whitespace

**Category**: Style

**Automatically fixable**: [Yes](/regal/fixing)

**Avoid**

```rego
package policy

#Deny by default
default allow := false

#Allow only admins
allow if "admin" in input.user.roles
```

**Prefer**

```rego
package policy

# Deny by default
default allow := false

# Allow only admins
allow if "admin" in input.user.roles
```

## Rationale

Comments should be preceded by a single space, as this makes them easier to read.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  style:
    no-whitespace-comment:
      # one of "error", "warning", "ignore"
      level: error
      # optional pattern to except from this rule
      # this example would allow comments like "#--"
      # use or (`|`) to separate multiple patterns
      except-pattern: '^--'
```

## Related Resources

- GitHub: [Source Code](https://github.com/StyraInc/regal/blob/main/bundle/regal/rules/style/no-whitespace-comment/no_whitespace_comment.rego)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
