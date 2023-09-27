# line-length

**Summary**: Line too long

**Category**: Style

**Avoid**

Excessive line length.

## Rationale

Rego does not have many nested constructs, and long lines of code are thus almost never needed. If you find yourself
close to the maximum line length, consider refactoring your policy.

The default maximum line length is 120 characters.

## Exceptions

On a few rare occasions, a single word — like a really long URL in a metadata annotation — can't possibly be made any 
shorter. Using an ignore directive isn't an option in that context, and ignoring the whole file is rarely what you'll
want. The `non-breakable-word-threshold` configuration option allows defining a threshold length for when a single word
should be considered so long that the line length rule should ignore the line entirely.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules: 
  style:
    line-length:
      # one of "error", "warning", "ignore"
      level: error
      # maximum line length
      max-line-length: 120
      # if any single word on a line exceeds this length, ignore it
      non-breakable-word-threshold: 100
```

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
