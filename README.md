# Regal

[![Build Status](https://github.com/styrainc/regal/workflows/Build/badge.svg?branch=main)](https://github.com/styrainc/regal/actions)
![OPA v0.52.0](https://openpolicyagent.org/badge/v0.52.0)

Regal is a linter for Rego, with the goal of making your Rego magnificent!

> regal
>
> 1 : of, relating to, or suitable for a king
>
> 2 : of notable excellence or magnificence : splendid

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

## Status

Regal is currently in an alpha stage. While we'd like to think of it as helpful already, **every** feature, API, name,
config attribute and paragraph of documentation is subject to change. If you'd still like to use it, we'd love to hear
what you think!

## Try it out!

Run `regal lint` pointed at one or more files or directories to have them linted:

```shell
regal lint policy/
```

## Rules

Regal comes with a set of built-in rules. The following rules are currently available:

<!-- RULES_TABLE_START -->

|  Category  |                                           Title                                           |                      Description                       |
|------------|-------------------------------------------------------------------------------------------|--------------------------------------------------------|
| assignment | [use-assignment-operator](https://docs.styra.com/regal/rules/use-assignment-operator)     | Prefer := over = for assignment                        |
| bugs       | [constant-condition](https://docs.styra.com/regal/rules/constant-condition)               | Constant condition                                     |
| bugs       | [top-level-iteration](https://docs.styra.com/regal/rules/top-level-iteration)             | Iteration in top-level assignment                      |
| comments   | [todo-comment](https://docs.styra.com/regal/rules/todo-comment)                           | Avoid TODO comments                                    |
| functions  | [external-reference](https://docs.styra.com/regal/rules/external-reference)               | Reference to input, data or rule ref in function body  |
| imports    | [implicit-future-keywords](https://docs.styra.com/regal/rules/implicit-future-keywords)   | Use explicit future keyword imports                    |
| imports    | [avoid-importing-input](https://docs.styra.com/regal/rules/avoid-importing-input)         | Avoid importing input                                  |
| imports    | [redundant-data-import](https://docs.styra.com/regal/rules/redundant-data-import)         | Redundant import of data                               |
| imports    | [import-shadows-import](https://docs.styra.com/regal/rules/import-shadows-import)         | Import shadows another import                          |
| rules      | [avoid-get-and-list-prefix](https://docs.styra.com/regal/rules/avoid-get-and-list-prefix) | Avoid get_ and list_ prefix for rules and functions    |
| rules      | [rule-shadows-builtin](https://docs.styra.com/regal/rules/rule-shadows-builtin)           | Rule name shadows built-in                             |
| style      | [prefer-snake-case](https://docs.styra.com/regal/rules/prefer-snake-case)                 | Prefer snake_case for names                            |
| style      | [use-in-operator](https://docs.styra.com/regal/rules/use-in-operator)                     | Use in to check for membership                         |
| style      | [line-length](https://docs.styra.com/regal/rules/line-length)                             | Line too long                                          |
| style      | [opa-fmt](https://docs.styra.com/regal/rules/opa-fmt)                                     | File should be formatted with `opa fmt`                |
| testing    | [test-outside-test-package](https://docs.styra.com/regal/rules/test-outside-test-package) | Test outside of test package                           |
| testing    | [file-missing-test-suffix](https://docs.styra.com/regal/rules/file-missing-test-suffix)   | Files containing tests should have a _test.rego suffix |
| testing    | [identically-named-tests](https://docs.styra.com/regal/rules/identically-named-tests)     | Multiple tests with same name                          |
| testing    | [todo-test](https://docs.styra.com/regal/rules/todo-test)                                 | TODO test encountered                                  |
| variables  | [unconditional-assignment](https://docs.styra.com/regal/rules/unconditional-assignment)   | Unconditional assignment in rule body                  |

<!-- RULES_TABLE_END -->

If you'd like to see more rules, please [open an issue](https://github.com/StyraInc/regal/issues) for your feature
request, or better yet, submit a PR! See the [custom rules](#custom-rules) section for more information on how to
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

## Community

For questions, discussions and announcements related to Styra products, services and open source projects, please join 
the Styra community on [Slack](https://join.slack.com/t/styracommunity/shared_invite/zt-1p81qz8g4-t2OLKbvw0J5ibdcNc62~6Q)!
