# Fixing Violations

For each violation Regal is able to detect, there is a documentation page
explaining the issue in detail and how to fix it. For example, here's the one for
the [`prefer-some-in-iteration`](/regal/rules/style/prefer-some-in-iteration) rule.

Some rules are automatically fixable, meaning that Regal can fix the violation
for you. To automatically fix all supported violations, run:

```shell
regal fix <path> [path [...]] [flags]
```

Currently, the following rules are automatically fixable:

- [opa-fmt](/regal/rules/style/opa-fmt)
- [use-rego-v1](/regal/rules/imports/use-rego-v1)
- [use-assignment-operator](/regal/rules/style/use-assignment-operator)
- [no-whitespace-comment](/regal/rules/style/no-whitespace-comment)

:::tip
Need to fix individual violations? Checkout the editors Regal supports
[here](/regal/editor-support).
:::

## Community

If you'd like to discuss Regal development or just talk about Regal in general, please join us in the `#regal`
channel in the Styra Community [Slack](https://communityinviter.com/apps/styracommunity/signup)!
