# Regal

[![Build Status](https://github.com/styrainc/regal/workflows/Build/badge.svg?branch=main)](https://github.com/styrainc/regal/actions)
![OPA v0.58.0](https://openpolicyagent.org/badge/v0.58.0)

Regal is a linter for [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/), with the goal of making your
Rego magnificent!

<img
  src="/docs/assets/regal-banner.png"
  alt="illustration of a viking representing the Regal logo"
  width="150px" />

> regal
>
> adj : of notable excellence or magnificence : splendid

\- [Merriam Webster](https://www.merriam-webster.com/dictionary/regal)


## Goals

- Identify common mistakes, bugs and inefficiencies in Rego policies, and suggest better approaches
- Provide advice on [best practices](https://github.com/StyraInc/rego-style-guide), coding style, and tooling
- Allow users, teams and organizations to enforce custom rules on their policy code

Regal rules are to as large extent as possible
[written in Rego](https://www.styra.com/blog/linting-rego-with-rego/) themselves,
using the JSON representation of the Rego abstract syntax tree (AST) as input, a
few additional custom built-in functions and some indexed data structures to help
with linting.

## What People Say About Regal

> I really like that at each release of Regal I learn something new!
> Of all the linters I'm exposed to, Regal is probably the most instructive one.

— Leonardo Taccari, [NetBSD](https://www.netbsd.org/)

> Reviewing the Regal rules documentation. Pure gold.

— Dima Korolev, [Miro](https://miro.com/)

> Such an awesome project!

— Shawn McGuire, [Atlassian](https://www.atlassian.com/)

## Getting Started

### Download Regal

**MacOS and Linux**
```shell
brew install styrainc/packages/regal
```

<details>
  <summary><strong>Manual download options</strong></summary>

**MacOS (Apple Silicon)**
```shell
curl -L -o regal "https://github.com/StyraInc/regal/releases/latest/download/regal_Darwin_arm64"
```

**MacOS (x86_64)**
```shell
curl -L -o regal "https://github.com/StyraInc/regal/releases/latest/download/regal_Darwin_x86_64"
```

**Linux (x86_64)**
```shell
curl -L -o regal "https://github.com/StyraInc/regal/releases/latest/download/regal_Linux_x86_64"
chmod +x regal
```

**Windows**
```shell
curl.exe -L -o regal.exe "https://github.com/StyraInc/regal/releases/latest/download/regal_Windows_x86_64.exe"
```

**Docker**
```shell
docker pull ghcr.io/styrainc/regal:latest
```

See all versions, and checksum files, at the Regal [releases](https://github.com/StyraInc/regal/releases/) page, and
published Docker images at the [packages](https://github.com/StyraInc/regal/pkgs/container/regal) page.

</details>

### Try it out!

First, author some Rego!

**policy/authz.rego**
```rego
package authz

import future.keywords

default allow = false

deny if {
	"admin" != input.user.roles[_]
}

allow if not deny
```

Next, run `regal lint` pointed at one or more files or directories to have them linted.

```shell
regal lint policy/
```
```text
Rule:         	not-equals-in-loop
Description:  	Use of != in loop
Category:     	bugs
Location:     	policy/authz.rego:8:10
Text:         	"admin" != input.user.roles[_]
Documentation:	https://docs.styra.com/regal/rules/bugs/not-equals-in-loop

Rule:         	implicit-future-keywords
Description:  	Use explicit future keyword imports
Category:     	imports
Location:     	policy/authz.rego:3:8
Text:         	import future.keywords
Documentation:	https://docs.styra.com/regal/rules/imports/implicit-future-keywords

Rule:         	use-assignment-operator
Description:  	Prefer := over = for assignment
Category:     	style
Location:     	policy/authz.rego:5:1
Text:         	default allow = false
Documentation:	https://docs.styra.com/regal/rules/style/use-assignment-operator

1 file linted. 3 violations found.
```
<br />

> **Note**
> If you're running Regal on an existing policy library, you may want to disable the `style` category initially, as it
> will likely generate a lot of violations. You can do this by passing the `--disable-category style` flag to
> `regal lint`.

### GitHub Actions

If you'd like to run Regal in GitHub actions, please consider using [`setup-regal`](https://github.com/StyraInc/setup-regal).
A simple `.github/workflows/lint.yml` to run regal on PRs could look like this, where `policy` contains Rego files:

```yaml
name: Regal Lint
on:
  pull_request:
jobs:
  lint-rego:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: StyraInc/setup-regal@main
      with:
        version: latest

    - name: Lint
      run: regal lint --format=github ./policy
```

Please see [`setup-regal`](https://github.com/StyraInc/setup-regal) for more information.

## Rules

Regal comes with a set of built-in rules, grouped by category.

- **bugs**: Common mistakes, potential bugs and inefficiencies in Rego policies.
- **custom**: Custom, rules where enforcement can be adjusted to match your preferences.
- **idiomatic**: Suggestions for more idiomatic constructs.
- **imports**: Best practices for imports.
- **style**: [Rego Style Guide](https://github.com/StyraInc/rego-style-guide) rules.
- **testing**: Rules for testing and development.

The following rules are currently available:

<!-- RULES_TABLE_START -->

| Category  |                                                Title                                                |                        Description                        |
|-----------|-----------------------------------------------------------------------------------------------------|-----------------------------------------------------------|
| bugs      | [constant-condition](https://docs.styra.com/regal/rules/bugs/constant-condition)                    | Constant condition                                        |
| bugs      | [invalid-metadata-attribute](https://docs.styra.com/regal/rules/bugs/invalid-metadata-attribute)    | Invalid attribute in metadata annotation                  |
| bugs      | [not-equals-in-loop](https://docs.styra.com/regal/rules/bugs/not-equals-in-loop)                    | Use of != in loop                                         |
| bugs      | [rule-named-if](https://docs.styra.com/regal/rules/bugs/rule-named-if)                              | Rule named "if"                                           |
| bugs      | [rule-shadows-builtin](https://docs.styra.com/regal/rules/bugs/rule-shadows-builtin)                | Rule name shadows built-in                                |
| bugs      | [top-level-iteration](https://docs.styra.com/regal/rules/bugs/top-level-iteration)                  | Iteration in top-level assignment                         |
| bugs      | [unused-return-value](https://docs.styra.com/regal/rules/bugs/unused-return-value)                  | Non-boolean return value unused                           |
| bugs      | [zero-arity-function](https://docs.styra.com/regal/rules/bugs/zero-arity-function)                  | Avoid functions without args                              |
| custom    | [forbidden-function-call](https://docs.styra.com/regal/rules/custom/forbidden-function-call)        | Forbidden function call                                   |
| custom    | [naming-convention](https://docs.styra.com/regal/rules/custom/naming-convention)                    | Naming convention violation                               |
| custom    | [one-liner-rule](https://docs.styra.com/regal/rules/custom/one-liner-rule)                          | Rule body could be made a one-liner                       |
| custom    | [prefer-value-in-head](https://docs.styra.com/regal/rules/custom/prefer-value-in-head)              | Prefer value in rule head                                 |
| idiomatic | [custom-has-key-construct](https://docs.styra.com/regal/rules/idiomatic/custom-has-key-construct)   | Custom function may be replaced by `in` and `object.keys` |
| idiomatic | [custom-in-construct](https://docs.styra.com/regal/rules/idiomatic/custom-in-construct)             | Custom function may be replaced by `in` keyword           |
| idiomatic | [equals-pattern-matching](https://docs.styra.com/regal/rules/idiomatic/equals-pattern-matching)     | Prefer pattern matching in function arguments             |
| idiomatic | [no-defined-entrypoint](https://docs.styra.com/regal/rules/idiomatic/no-defined-entrypoint)         | Missing entrypoint annotation                             |
| idiomatic | [non-raw-regex-pattern](https://docs.styra.com/regal/rules/idiomatic/non-raw-regex-pattern)         | Use raw strings for regex patterns                        |
| idiomatic | [prefer-set-or-object-rule](https://docs.styra.com/regal/rules/idiomatic/prefer-set-or-object-rule) | Prefer set or object rule over comprehension              |
| idiomatic | [use-in-operator](https://docs.styra.com/regal/rules/idiomatic/use-in-operator)                     | Use in to check for membership                            |
| idiomatic | [use-some-for-output-vars](https://docs.styra.com/regal/rules/idiomatic/use-some-for-output-vars)   | Use `some` to declare output variables                    |
| imports   | [avoid-importing-input](https://docs.styra.com/regal/rules/imports/avoid-importing-input)           | Avoid importing input                                     |
| imports   | [implicit-future-keywords](https://docs.styra.com/regal/rules/imports/implicit-future-keywords)     | Use explicit future keyword imports                       |
| imports   | [import-after-rule](https://docs.styra.com/regal/rules/imports/import-after-rule)                   | Import declared after rule                                |
| imports   | [import-shadows-builtin](https://docs.styra.com/regal/rules/imports/import-shadows-builtin)         | Import shadows built-in namespace                         |
| imports   | [import-shadows-import](https://docs.styra.com/regal/rules/imports/import-shadows-import)           | Import shadows another import                             |
| imports   | [prefer-package-imports](https://docs.styra.com/regal/rules/imports/prefer-package-imports)         | Prefer importing packages over rules                      |
| imports   | [redundant-alias](https://docs.styra.com/regal/rules/imports/redundant-alias)                       | Redundant alias                                           |
| imports   | [redundant-data-import](https://docs.styra.com/regal/rules/imports/redundant-data-import)           | Redundant import of data                                  |
| style     | [avoid-get-and-list-prefix](https://docs.styra.com/regal/rules/style/avoid-get-and-list-prefix)     | Avoid get_ and list_ prefix for rules and functions       |
| style     | [chained-rule-body](https://docs.styra.com/regal/rules/style/chained-rule-body)                     | Avoid chaining rule bodies                                |
| style     | [default-over-else](https://docs.styra.com/regal/rules/style/default-over-else)                     | Prefer default assignment over fallback else              |
| style     | [detached-metadata](https://docs.styra.com/regal/rules/style/detached-metadata)                     | Detached metadata annotation                              |
| style     | [external-reference](https://docs.styra.com/regal/rules/style/external-reference)                   | Reference to input, data or rule ref in function body     |
| style     | [file-length](https://docs.styra.com/regal/rules/style/file-length)                                 | Max file length exceeded                                  |
| style     | [function-arg-return](https://docs.styra.com/regal/rules/style/function-arg-return)                 | Function argument used for return value                   |
| style     | [line-length](https://docs.styra.com/regal/rules/style/line-length)                                 | Line too long                                             |
| style     | [no-whitespace-comment](https://docs.styra.com/regal/rules/style/no-whitespace-comment)             | Comment should start with whitespace                      |
| style     | [opa-fmt](https://docs.styra.com/regal/rules/style/opa-fmt)                                         | File should be formatted with `opa fmt`                   |
| style     | [prefer-snake-case](https://docs.styra.com/regal/rules/style/prefer-snake-case)                     | Prefer snake_case for names                               |
| style     | [prefer-some-in-iteration](https://docs.styra.com/regal/rules/style/prefer-some-in-iteration)       | Prefer `some .. in` for iteration                         |
| style     | [rule-length](https://docs.styra.com/regal/rules/style/rule-length)                                 | Max rule length exceeded                                  |
| style     | [todo-comment](https://docs.styra.com/regal/rules/style/todo-comment)                               | Avoid TODO comments                                       |
| style     | [unconditional-assignment](https://docs.styra.com/regal/rules/style/unconditional-assignment)       | Unconditional assignment in rule body                     |
| style     | [unnecessary-some](https://docs.styra.com/regal/rules/style/unnecessary-some)                       | Unnecessary use of `some`                                 |
| style     | [use-assignment-operator](https://docs.styra.com/regal/rules/style/use-assignment-operator)         | Prefer := over = for assignment                           |
| testing   | [dubious-print-sprintf](https://docs.styra.com/regal/rules/testing/dubious-print-sprintf)           | Dubious use of print and sprintf                          |
| testing   | [file-missing-test-suffix](https://docs.styra.com/regal/rules/testing/file-missing-test-suffix)     | Files containing tests should have a _test.rego suffix    |
| testing   | [identically-named-tests](https://docs.styra.com/regal/rules/testing/identically-named-tests)       | Multiple tests with same name                             |
| testing   | [metasyntactic-variable](https://docs.styra.com/regal/rules/testing/metasyntactic-variable)         | Metasyntactic variable name                               |
| testing   | [print-or-trace-call](https://docs.styra.com/regal/rules/testing/print-or-trace-call)               | Call to print or trace function                           |
| testing   | [test-outside-test-package](https://docs.styra.com/regal/rules/testing/test-outside-test-package)   | Test outside of test package                              |
| testing   | [todo-test](https://docs.styra.com/regal/rules/testing/todo-test)                                   | TODO test encountered                                     |

<!-- RULES_TABLE_END -->

By default, all rules except for those in the `custom` category are currently **enabled**.

#### Aggregate Rules

Most Regal rules will use data only from a single file at a time, with no consideration for other files. A few rules
however require data from multiple files, and will therefore collect, or aggregate, data from all files provided for
linting. These rules are called *aggregate rules*, and will only be run when there is more than one file to lint, such
as when linting a directory or a whole policy repository. One example of such a rule is the `prefer-package-imports`
rule, which will aggregate package names and imports from all provided policies in order to determine if any imports
are pointing to rules or functions rather than packages. You normally won't need to care about this distinction other
than being aware of the fact that some linter rules won't be run when linting a single file.

If you'd like to see more rules, please [open an issue](https://github.com/StyraInc/regal/issues) for your feature
request, or better yet, submit a PR! See the [custom rules](/docs/custom-rules.md) page for more information on how to
develop your own rules, for yourself or for inclusion in Regal.

### Custom Rules

The `custom` category is a special one, as the rules in this category allow you to enforce rules that are specific to
your project, team or organization. This typically includes things like naming conventions, where you might want to
ensure that, for example, all package names adhere to an organizational standard, like having a prefix matching the
organization name.

Since these rules require configuration provided by the user, or are more opinionated than other rules, they are
disabled by default. In order to enable them, see the configuration options available for each rule for how to configure
them according to your requirements.

For more advanced requirements, see the guide on writing [custom rules](/docs/custom-rules.md) in Rego.

## Configuration

A custom configuration file may be used to override the [default configuration](https://github.com/StyraInc/regal/blob/main/bundle/regal/config/provided/data.yaml)
options provided by Regal. The most common use case for this is to change the severity level of a rule. These three
levels are available:

- `ignore`  — disable the rule entirely
- `warning` — report the violation without changing the exit code of the lint command
- `error`   — report the violation and have the lint command exit with a non-zero exit code (default)

Additionally, some rules may have configuration options of their own. See the documentation page for a rule to learn
more about it.

**.regal/config.yaml**
```yaml
# Files can be excluded from all lint
# rules according to glob-patterns
ignore:
  files:
    - file1.rego
    - "*_tmp.rego"
rules:
  style:
    todo-comment:
      # don't report on todo comments
      level: ignore
    line-length:
      # custom rule configuration
      max-line-length: 100
      # warn on too long lines, but don't fail
      level: warning
    opa-fmt:
      # not needed as error is the default, but
      # being explicit won't hurt
      level: error
      # Files can be ignored.
      # In this example, test files are ignored
      ignore:
        files:
          - "*_test.rego"
  custom:
    # custom rule configuration
    naming-convention:
      level: error
      conventions:
        # ensure all package names start with "acmecorp" or "system"
        - pattern: '^acmecorp\.[a-z_\.]+$|^system\.[a-z_\.]+$'
          targets:
            - package
```

Regal will automatically search for a configuration file (`.regal/config.yaml`) in the current directory, and if not
found, traverse the parent directories either until either one is found, or the top of the directory hierarchy is
reached. If no configuration file is found, Regal will use the default configuration.

A custom configuration may be also be provided using the `--config-file`/`-c` option for `regal lint`, which when
provided will be used to override the default configuration.

### CLI flags

For development, rules may also quickly be enabled or disabled using the relevant CLI flags for the `regal lint`
command.

- `--disable-all` disables **all** rules
- `--disable-category` disables all rules in a category, overriding `--enable-all` (may be repeated)
- `--disable` disables a specific rule, overriding `--enable-all` and `--enable-category` (may be repeated)
- `--enable-all` enables **all** rules
- `--enable-category` enables all rules in a category, overriding `--disable-all` (may be repeated)
- `--enable` enables a specific rule, overriding `--disable-all` and `--disable-category` (may be repeated)
- `--ignore-files` ignores files using glob patterns, overriding `ignore` in the config file (may be repeated)

All CLI flags override configuration provided in file.

## Exit Codes

Exit codes are used to indicate the result of the `lint` command. The `--fail-level` provided for `regal lint` may be
used to change the exit code behavior, and allows a value of either `warning` or `error` (default).

If `--fail-level error` is supplied, exit code will be zero even if warnings are present:

- `0`: no errors were found
- `0`: one or more warnings were found
- `3`: one or more errors were found

This is the default behavior.

If `--fail-level warning` is supplied, warnings will result in a non-zero exit code:

- `0`: no errors or warnings were found
- `2`: one or more warnings were found
- `3`: one or more errors were found

## Inline Ignore Directives

If you'd like to ignore a specific violation, you can add an ignore directive above the line in question:

```rego
package policy

# regal ignore:prefer-snake-case
camelCase := "yes"
```

The format of an ignore directive is `regal ignore:<rule-name>,<rule-name>...`, where `<rule-name>` is the name of the
rule to ignore. Multiple rules may be added to the same ignore directive, separated by commas.

Note that at this point in time, Regal only considers the line following the ignore directive, i.e. it does not ignore
entire blocks of code (like rules, functions or even packages). See [configuration](#configuration) if you want to
ignore certain rules altogether.

## Output Formats

The `regal lint` command allows specifying the output format by using the `--format` flag. The available output formats
are:

- `pretty` (default) - Human-readable table-like output where each violation is printed with a detailed explanation
- `compact` - Human-readable output where each violation is printed on a single line
- `json` - JSON output, suitable for programmatic consumption
- `github` - GitHub [workflow command](https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions)
  output, ideal for use in GitHub Actions. Annotates PRs and creates a
  [job summary](https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-job-summary)
  from the linter report

## Resources

### Documentation

- [Custom Rules](/docs/custom-rules.md) describes how to develop your own linter rules
- [Architecture](/docs/architecture.md) provides a high-level technical overview of how Regal works
- [Development](/docs/development.md) contains information about how to hack on Regal itself
- [Integration](/docs/integration.md) describes how to integrate Regal in your Go application
- [Rego Style Guide](/docs/rego-style-guide.md) contains notes on implementing the
  [Rego Style Guide](https://github.com/StyraInc/rego-style-guide) rules
- [Pre-Commit Hooks](/docs/pre-commit-hooks.md) describes how to use Regal in pre-commit hooks

### Talks

[Regal the Rego Linter](https://www.youtube.com/watch?v=Xx8npd2TQJ0&t=2567s), CNCF London meetup, June 2023
[![Regal the Rego Linter](/docs/assets/regal_cncf_london.png)](https://www.youtube.com/watch?v=Xx8npd2TQJ0&t=2567s)

### Blogs and Articles

- [Guarding the Guardrails - Introducing Regal the Rego Linter](https://www.styra.com/blog/guarding-the-guardrails-introducing-regal-the-rego-linter/)
- [Scaling Open Source Community by Getting Closer to Users](https://thenewstack.io/scaling-open-source-community-by-getting-closer-to-users/)

## Status

Regal is currently in beta. End-users should not expect any drastic changes, but any API may change without notice.
If you want to embed Regal in another project or product, please reach out!

## Roadmap

The roadmap for Regal currently looks like this:

- [x] 50 rules!
- [x] Add `custom` (or `organizational`, `opinionated`, or...) category for built-in "custom", or
      [organizational rules](https://github.com/StyraInc/regal/issues/48), to enforce things like naming conventions.
      The most common customizations should not require writing custom rules, but be made available in configuration.
- [x] Simplify custom rules authoring by providing
      [command for scaffolding](https://github.com/StyraInc/regal/issues/206)
- [ ] Make more rules consider [nested](https://github.com/StyraInc/regal/issues/82) AST nodes
- [x] [GitHub Action](https://github.com/StyraInc/setup-regal)

The roadmap is updated when all the current items have been completed.

## Community

For questions, discussions and announcements related to Styra products, services and open source projects, please join
the Styra community on [Slack](https://communityinviter.com/apps/styracommunity/signup)!
