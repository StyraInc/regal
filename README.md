# Regal

[![Build Status](https://github.com/styrainc/regal/workflows/Build/badge.svg?branch=main)](https://github.com/styrainc/regal/actions)
![OPA v0.53.1](https://openpolicyagent.org/badge/v0.53.1)

Regal is a linter for Rego, with the goal of making your Rego magnificent!

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

See all versions, and checksum files, at the Regal [releases](https://github.com/StyraInc/regal/releases/) page.

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
Documentation:	https://github.com/StyraInc/regal/blob/main/docs/rules/bugs/not-equals-in-loop.md

Rule:         	implicit-future-keywords
Description:  	Use explicit future keyword imports
Category:     	imports
Location:     	policy/authz.rego:3:8
Text:         	import future.keywords
Documentation:	https://github.com/StyraInc/regal/blob/main/docs/rules/imports/implicit-future-keywords.md

Rule:         	use-assignment-operator
Description:  	Prefer := over = for assignment
Category:     	style
Location:     	policy/authz.rego:5:1
Text:         	default allow = false
Documentation:	https://github.com/StyraInc/regal/blob/main/docs/rules/style/use-assignment-operator.md

1 file linted. 3 violations found.
```
<br />

> **Note**
> If you're running Regal on an existing policy library, you may want to disable the `style` category initially, as it
> will likely generate a lot of violations. You can do this by passing the `--disable-category style` flag to
> `regal lint`.

## Rules

Regal comes with a set of built-in rules, grouped by category.

- **bugs**: Common mistakes, potential bugs and inefficiencies in Rego policies.
- **imports**: Best practices for imports.
- **style**: [Rego Style Guide](https://github.com/StyraInc/rego-style-guide) rules.
- **testing**: Rules for testing and development.

The following rules are currently available:

<!-- RULES_TABLE_START -->

| Category |                                                          Title                                                           |                      Description                       |
|----------|--------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------|
| bugs     | [constant-condition](https://github.com/StyraInc/regal/blob/main/docs/rules/bugs/constant-condition.md)                  | Constant condition                                     |
| bugs     | [top-level-iteration](https://github.com/StyraInc/regal/blob/main/docs/rules/bugs/top-level-iteration.md)                | Iteration in top-level assignment                      |
| bugs     | [unused-return-value](https://github.com/StyraInc/regal/blob/main/docs/rules/bugs/unused-return-value.md)                | Non-boolean return value unused                        |
| bugs     | [not-equals-in-loop](https://github.com/StyraInc/regal/blob/main/docs/rules/bugs/not-equals-in-loop.md)                  | Use of != in loop                                      |
| bugs     | [rule-shadows-builtin](https://github.com/StyraInc/regal/blob/main/docs/rules/bugs/rule-shadows-builtin.md)              | Rule name shadows built-in                             |
| bugs     | [rule-named-if](https://github.com/StyraInc/regal/blob/main/docs/rules/bugs/rule-named-if.md)                            | Rule named "if"                                        |
| imports  | [implicit-future-keywords](https://github.com/StyraInc/regal/blob/main/docs/rules/imports/implicit-future-keywords.md)   | Use explicit future keyword imports                    |
| imports  | [avoid-importing-input](https://github.com/StyraInc/regal/blob/main/docs/rules/imports/avoid-importing-input.md)         | Avoid importing input                                  |
| imports  | [redundant-data-import](https://github.com/StyraInc/regal/blob/main/docs/rules/imports/redundant-data-import.md)         | Redundant import of data                               |
| imports  | [import-shadows-import](https://github.com/StyraInc/regal/blob/main/docs/rules/imports/import-shadows-import.md)         | Import shadows another import                          |
| imports  | [redundant-alias](https://github.com/StyraInc/regal/blob/main/docs/rules/imports/redundant-alias.md)                     | Redundant alias                                        |
| style    | [use-in-operator](https://github.com/StyraInc/regal/blob/main/docs/rules/style/use-in-operator.md)                       | Use in to check for membership                         |
| style    | [unconditional-assignment](https://github.com/StyraInc/regal/blob/main/docs/rules/style/unconditional-assignment.md)     | Unconditional assignment in rule body                  |
| style    | [line-length](https://github.com/StyraInc/regal/blob/main/docs/rules/style/line-length.md)                               | Line too long                                          |
| style    | [use-assignment-operator](https://github.com/StyraInc/regal/blob/main/docs/rules/style/use-assignment-operator.md)       | Prefer := over = for assignment                        |
| style    | [todo-comment](https://github.com/StyraInc/regal/blob/main/docs/rules/style/todo-comment.md)                             | Avoid TODO comments                                    |
| style    | [external-reference](https://github.com/StyraInc/regal/blob/main/docs/rules/style/external-reference.md)                 | Reference to input, data or rule ref in function body  |
| style    | [avoid-get-and-list-prefix](https://github.com/StyraInc/regal/blob/main/docs/rules/style/avoid-get-and-list-prefix.md)   | Avoid get_ and list_ prefix for rules and functions    |
| style    | [prefer-snake-case](https://github.com/StyraInc/regal/blob/main/docs/rules/style/prefer-snake-case.md)                   | Prefer snake_case for names                            |
| style    | [function-arg-return](https://github.com/StyraInc/regal/blob/main/docs/rules/style/function-arg-return.md)               | Function argument used for return value                |
| style    | [opa-fmt](https://github.com/StyraInc/regal/blob/main/docs/rules/style/opa-fmt.md)                                       | File should be formatted with `opa fmt`                |
| testing  | [file-missing-test-suffix](https://github.com/StyraInc/regal/blob/main/docs/rules/testing/file-missing-test-suffix.md)   | Files containing tests should have a _test.rego suffix |
| testing  | [identically-named-tests](https://github.com/StyraInc/regal/blob/main/docs/rules/testing/identically-named-tests.md)     | Multiple tests with same name                          |
| testing  | [todo-test](https://github.com/StyraInc/regal/blob/main/docs/rules/testing/todo-test.md)                                 | TODO test encountered                                  |
| testing  | [print-or-trace-call](https://github.com/StyraInc/regal/blob/main/docs/rules/testing/print-or-trace-call.md)             | Call to print or trace function                        |
| testing  | [test-outside-test-package](https://github.com/StyraInc/regal/blob/main/docs/rules/testing/test-outside-test-package.md) | Test outside of test package                           |

<!-- RULES_TABLE_END -->

By default, all rules are currently **enabled**.

If you'd like to see more rules, please [open an issue](https://github.com/StyraInc/regal/issues) for your feature
request, or better yet, submit a PR! See the [custom rules](/docs/custom-rules) page for more information on how to
develop your own rules, for yourself or for inclusion in Regal.

## Configuration

A custom configuration file may be used to override the [default configuration](bundle/regal/config/provided/data.yaml)
options provided by Regal. The most common use case for this is to change the severity level of a rule. These three
levels are available:

- `ignore`  — disable the rule entirely
- `warning` — report the violation without changing the exit code of the lint command
- `error`   — report the violation and have the lint command exit with a non-zero exit code (default)

Additionally, some rules may have configuration options of their own. See the documentation page for a rule to learn
more about it.

**.regal/config.yaml**
```yaml
rules:
  comments:
    todo-comment:
      # don't report on todo comments
      level: ignore
  style:
    line-length:
      # custom rule configuration
      max-line-length: 100
      # warn on too long lines, but don't fail
      level: warning
    opa-fmt:
      # not needed as error is the default, but
      # being explicit won't hurt
      level: error
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
entire blocks of code (like rules, or even packages). See [configuration](#configuration) if you want to ignore certain
rules altogether.

## Documentation

- [Custom Rules](/docs/custom-rules) describes how to develop your own rules
- [Development](/docs/development) for info about how to hack on Regal itself
- [Rego Style Guide](/docs/rego-style-guide) contains notes on implementing the [Rego Style Guide](https://github.com/StyraInc/rego-style-guide) rules

## Status

Regal is currently in beta. End-users should not expect any drastic changes, but any API may change without notice.

## Roadmap

- [ ] More rules!
- [ ] Add `custom` category for built-in "custom", or customizable rules, to enforce things like naming conventions
- [ ] Simplify custom rules authoring by providing command for scaffolding
- [ ] Improvements to assist writing rules that can't be enforced using the AST alone
- [ ] Make more rules consider nesting
- [ ] GitHub Action
- [ ] VS Code extension

## Community

For questions, discussions and announcements related to Styra products, services and open source projects, please join
the Styra community on [Slack](https://communityinviter.com/apps/styracommunity/signup)!
