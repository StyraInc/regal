<!-- markdownlint-disable MD041 -->

## Rules

Regal comes with a set of built-in rules, grouped by category.

- [**bugs**](https://docs.styra.com/regal/rules/bugs):
  Common mistakes, potential bugs and inefficiencies in Rego policies.
- [**custom**](https://docs.styra.com/regal/rules/custom):
  Custom rules where enforcement can be adjusted to match your preferences.
- [**idiomatic**](https://docs.styra.com/regal/rules/idiomatic):
  Suggestions for more idiomatic constructs.
- [**imports**](https://docs.styra.com/regal/rules/imports):
  Best practices for imports.
- [**style**](https://docs.styra.com/regal/rules/style):
  [Rego Style Guide](https://github.com/StyraInc/rego-style-guide) rules.
- [**testing**](https://docs.styra.com/regal/rules/testing):
  Rules for testing and development.
- [**performance**](https://docs.styra.com/regal/rules/performance):
  Rules for improving performance of policies.

Browse [All Available Rules](https://docs.styra.com/regal/rules).

Rules in all categories except for those in `custom` are **enabled** by default. Some rules however — like
`use-contains` and `use-if` — are conditionally enabled only when a version of OPA/Rego before 1.0 is targeted. See the
configuration options below if you want to use Regal to lint "legacy" policies.

**Aggregate Rules**

Most Regal rules will use data only from a single file at a time, with no consideration for other files. A few rules
however require data from multiple files, and will therefore collect, or aggregate, data from all files provided for
linting. These rules are called _aggregate rules_, and will only be run when there is more than one file to lint, such
as when linting a directory or a whole policy repository. One example of such a rule is the `prefer-package-imports`
rule, which will aggregate package names and imports from all provided policies in order to determine if any imports
are pointing to rules or functions rather than packages. You normally won't need to care about this distinction other
than being aware of the fact that some linter rules won't be run when linting a single file.

If you'd like to see more rules, please [open an issue](https://github.com/StyraInc/regal/issues) for your feature
request, or better yet, submit a PR! See the
[custom rules](https://docs.styra.com/regal/custom-rules)
page for more information on how to develop your own rules, for yourself or for
inclusion in Regal.

### Custom Rules

The `custom` category is a special one, as the rules in this category allow you to enforce rules that are specific to
your project, team or organization. This typically includes things like naming conventions, where you might want to
ensure that, for example, all package names adhere to an organizational standard, like having a prefix matching the
organization name.

Since these rules require configuration provided by the user, or are more opinionated than other rules, they are
disabled by default. In order to enable them, see the configuration options available for each rule for how to configure
them according to your requirements.

For more advanced requirements, see the guide on writing [custom rules](https://docs.styra.com/regal/custom-rules) in Rego.
