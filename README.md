# Regal

[![Build Status](https://github.com/styrainc/regal/workflows/Build/badge.svg?branch=main)](https://github.com/styrainc/regal/actions)
![OPA v0.64.1](https://openpolicyagent.org/badge/v0.64.1)

Regal is a linter and language server for [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/), helping
you write better policies and have fun while doing it!

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

> I am really impressed with Regal. It has helped me write more expressive and deterministic Rego.

— Jimmy Ray, [Boeing](https://www.boeing.com/)

See the [adopters](/docs/adopters.md) file for more Regal users.

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

import rego.v1

default allow = false

allow if {
    isEmployee
    "developer" in input.user.roles
}

isEmployee if regex.match("@acmecorp\\.com$", input.user.email)
```

Next, run `regal lint` pointed at one or more files or directories to have them linted.

```shell
regal lint policy/
```
<!-- markdownlint-capture -->
<!-- markdownlint-disable MD010 -->
```text
Rule:         	non-raw-regex-pattern
Description:  	Use raw strings for regex patterns
Category:     	idiomatic
Location:     	policy/authz.rego:12:27
Text:         	isEmployee if regex.match("@acmecorp\\.com$", input.user.email)
Documentation:	https://docs.styra.com/regal/rules/idiomatic/non-raw-regex-pattern

Rule:         	use-assignment-operator
Description:  	Prefer := over = for assignment
Category:     	style
Location:     	policy/authz.rego:5:1
Text:         	default allow = false
Documentation:	https://docs.styra.com/regal/rules/style/use-assignment-operator

Rule:         	prefer-snake-case
Description:  	Prefer snake_case for names
Category:     	style
Location:     	policy/authz.rego:12:1
Text:         	isEmployee if regex.match("@acmecorp\\.com$", input.user.email)
Documentation:	https://docs.styra.com/regal/rules/style/prefer-snake-case

1 file linted. 3 violations found.
```
<!-- markdownlint-restore -->
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
    - uses: StyraInc/setup-regal@v1
      with:
        # For production workflows, use a specific version, like v0.16.0
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

|  Category   |                                                 Title                                                 |                        Description                        |
|-------------|-------------------------------------------------------------------------------------------------------|-----------------------------------------------------------|
| bugs        | [constant-condition](https://docs.styra.com/regal/rules/bugs/constant-condition)                      | Constant condition                                        |
| bugs        | [deprecated-builtin](https://docs.styra.com/regal/rules/bugs/deprecated-builtin)                      | Avoid using deprecated built-in functions                 |
| bugs        | [duplicate-rule](https://docs.styra.com/regal/rules/bugs/duplicate-rule)                              | Duplicate rule                                            |
| bugs        | [if-empty-object](https://docs.styra.com/regal/rules/bugs/if-empty-object)                            | Empty object following `if`                               |
| bugs        | [impossible-not](https://docs.styra.com/regal/rules/bugs/impossible-not)                              | Impossible `not` condition                                |
| bugs        | [inconsistent-args](https://docs.styra.com/regal/rules/bugs/inconsistent-args)                        | Inconsistently named function arguments                   |
| bugs        | [invalid-metadata-attribute](https://docs.styra.com/regal/rules/bugs/invalid-metadata-attribute)      | Invalid attribute in metadata annotation                  |
| bugs        | [not-equals-in-loop](https://docs.styra.com/regal/rules/bugs/not-equals-in-loop)                      | Use of != in loop                                         |
| bugs        | [redundant-existence-check](https://docs.styra.com/regal/rules/bugs/redundant-existence-check)        | Redundant existence check                                 |
| bugs        | [rule-named-if](https://docs.styra.com/regal/rules/bugs/rule-named-if)                                | Rule named "if"                                           |
| bugs        | [rule-shadows-builtin](https://docs.styra.com/regal/rules/bugs/rule-shadows-builtin)                  | Rule name shadows built-in                                |
| bugs        | [top-level-iteration](https://docs.styra.com/regal/rules/bugs/top-level-iteration)                    | Iteration in top-level assignment                         |
| bugs        | [unassigned-return-value](https://docs.styra.com/regal/rules/bugs/unassigned-return-value)            | Non-boolean return value unassigned                       |
| bugs        | [zero-arity-function](https://docs.styra.com/regal/rules/bugs/zero-arity-function)                    | Avoid functions without args                              |
| custom      | [forbidden-function-call](https://docs.styra.com/regal/rules/custom/forbidden-function-call)          | Forbidden function call                                   |
| custom      | [naming-convention](https://docs.styra.com/regal/rules/custom/naming-convention)                      | Naming convention violation                               |
| custom      | [one-liner-rule](https://docs.styra.com/regal/rules/custom/one-liner-rule)                            | Rule body could be made a one-liner                       |
| custom      | [prefer-value-in-head](https://docs.styra.com/regal/rules/custom/prefer-value-in-head)                | Prefer value in rule head                                 |
| idiomatic   | [boolean-assignment](https://docs.styra.com/regal/rules/idiomatic/boolean-assignment)                 | Prefer `if` over boolean assignment                       |
| idiomatic   | [custom-has-key-construct](https://docs.styra.com/regal/rules/idiomatic/custom-has-key-construct)     | Custom function may be replaced by `in` and `object.keys` |
| idiomatic   | [custom-in-construct](https://docs.styra.com/regal/rules/idiomatic/custom-in-construct)               | Custom function may be replaced by `in` keyword           |
| idiomatic   | [equals-pattern-matching](https://docs.styra.com/regal/rules/idiomatic/equals-pattern-matching)       | Prefer pattern matching in function arguments             |
| idiomatic   | [no-defined-entrypoint](https://docs.styra.com/regal/rules/idiomatic/no-defined-entrypoint)           | Missing entrypoint annotation                             |
| idiomatic   | [non-raw-regex-pattern](https://docs.styra.com/regal/rules/idiomatic/non-raw-regex-pattern)           | Use raw strings for regex patterns                        |
| idiomatic   | [prefer-set-or-object-rule](https://docs.styra.com/regal/rules/idiomatic/prefer-set-or-object-rule)   | Prefer set or object rule over comprehension              |
| idiomatic   | [use-contains](https://docs.styra.com/regal/rules/idiomatic/use-contains)                             | Use the `contains` keyword                                |
| idiomatic   | [use-if](https://docs.styra.com/regal/rules/idiomatic/use-if)                                         | Use the `if` keyword                                      |
| idiomatic   | [use-in-operator](https://docs.styra.com/regal/rules/idiomatic/use-in-operator)                       | Use in to check for membership                            |
| idiomatic   | [use-some-for-output-vars](https://docs.styra.com/regal/rules/idiomatic/use-some-for-output-vars)     | Use `some` to declare output variables                    |
| imports     | [avoid-importing-input](https://docs.styra.com/regal/rules/imports/avoid-importing-input)             | Avoid importing input                                     |
| imports     | [circular-import](https://docs.styra.com/regal/rules/imports/circular-import)                         | Circular import                                           |
| imports     | [ignored-import](https://docs.styra.com/regal/rules/imports/ignored-import)                           | Reference ignores import                                  |
| imports     | [implicit-future-keywords](https://docs.styra.com/regal/rules/imports/implicit-future-keywords)       | Use explicit future keyword imports                       |
| imports     | [import-after-rule](https://docs.styra.com/regal/rules/imports/import-after-rule)                     | Import declared after rule                                |
| imports     | [import-shadows-builtin](https://docs.styra.com/regal/rules/imports/import-shadows-builtin)           | Import shadows built-in namespace                         |
| imports     | [import-shadows-import](https://docs.styra.com/regal/rules/imports/import-shadows-import)             | Import shadows another import                             |
| imports     | [prefer-package-imports](https://docs.styra.com/regal/rules/imports/prefer-package-imports)           | Prefer importing packages over rules                      |
| imports     | [redundant-alias](https://docs.styra.com/regal/rules/imports/redundant-alias)                         | Redundant alias                                           |
| imports     | [redundant-data-import](https://docs.styra.com/regal/rules/imports/redundant-data-import)             | Redundant import of data                                  |
| imports     | [unresolved-import](https://docs.styra.com/regal/rules/imports/unresolved-import)                     | Unresolved import                                         |
| imports     | [use-rego-v1](https://docs.styra.com/regal/rules/imports/use-rego-v1)                                 | Use `import rego.v1`                                      |
| performance | [with-outside-test-context](https://docs.styra.com/regal/rules/performance/with-outside-test-context) | `with` used outside test context                          |
| style       | [avoid-get-and-list-prefix](https://docs.styra.com/regal/rules/style/avoid-get-and-list-prefix)       | Avoid `get_` and `list_` prefix for rules and functions   |
| style       | [chained-rule-body](https://docs.styra.com/regal/rules/style/chained-rule-body)                       | Avoid chaining rule bodies                                |
| style       | [default-over-else](https://docs.styra.com/regal/rules/style/default-over-else)                       | Prefer default assignment over fallback else              |
| style       | [default-over-not](https://docs.styra.com/regal/rules/style/default-over-not)                         | Prefer default assignment over negated condition          |
| style       | [detached-metadata](https://docs.styra.com/regal/rules/style/detached-metadata)                       | Detached metadata annotation                              |
| style       | [double-negative](https://docs.styra.com/regal/rules/style/double-negative)                           | Avoid double negatives                                    |
| style       | [external-reference](https://docs.styra.com/regal/rules/style/external-reference)                     | Reference to input, data or rule ref in function body     |
| style       | [file-length](https://docs.styra.com/regal/rules/style/file-length)                                   | Max file length exceeded                                  |
| style       | [function-arg-return](https://docs.styra.com/regal/rules/style/function-arg-return)                   | Function argument used for return value                   |
| style       | [line-length](https://docs.styra.com/regal/rules/style/line-length)                                   | Line too long                                             |
| style       | [no-whitespace-comment](https://docs.styra.com/regal/rules/style/no-whitespace-comment)               | Comment should start with whitespace                      |
| style       | [opa-fmt](https://docs.styra.com/regal/rules/style/opa-fmt)                                           | File should be formatted with `opa fmt`                   |
| style       | [prefer-snake-case](https://docs.styra.com/regal/rules/style/prefer-snake-case)                       | Prefer snake_case for names                               |
| style       | [prefer-some-in-iteration](https://docs.styra.com/regal/rules/style/prefer-some-in-iteration)         | Prefer `some .. in` for iteration                         |
| style       | [rule-length](https://docs.styra.com/regal/rules/style/rule-length)                                   | Max rule length exceeded                                  |
| style       | [rule-name-repeats-package](https://docs.styra.com/regal/rules/style/rule-name-repeats-package)       | Rule name repeats package                                 |
| style       | [todo-comment](https://docs.styra.com/regal/rules/style/todo-comment)                                 | Avoid TODO comments                                       |
| style       | [unconditional-assignment](https://docs.styra.com/regal/rules/style/unconditional-assignment)         | Unconditional assignment in rule body                     |
| style       | [unnecessary-some](https://docs.styra.com/regal/rules/style/unnecessary-some)                         | Unnecessary use of `some`                                 |
| style       | [use-assignment-operator](https://docs.styra.com/regal/rules/style/use-assignment-operator)           | Prefer := over = for assignment                           |
| style       | [yoda-condition](https://docs.styra.com/regal/rules/style/yoda-condition)                             | Yoda condition                                            |
| testing     | [dubious-print-sprintf](https://docs.styra.com/regal/rules/testing/dubious-print-sprintf)             | Dubious use of print and sprintf                          |
| testing     | [file-missing-test-suffix](https://docs.styra.com/regal/rules/testing/file-missing-test-suffix)       | Files containing tests should have a _test.rego suffix    |
| testing     | [identically-named-tests](https://docs.styra.com/regal/rules/testing/identically-named-tests)         | Multiple tests with same name                             |
| testing     | [metasyntactic-variable](https://docs.styra.com/regal/rules/testing/metasyntactic-variable)           | Metasyntactic variable name                               |
| testing     | [print-or-trace-call](https://docs.styra.com/regal/rules/testing/print-or-trace-call)                 | Call to print or trace function                           |
| testing     | [test-outside-test-package](https://docs.styra.com/regal/rules/testing/test-outside-test-package)     | Test outside of test package                              |
| testing     | [todo-test](https://docs.styra.com/regal/rules/testing/todo-test)                                     | TODO test encountered                                     |

<!-- RULES_TABLE_END -->

By default, all rules except for those in the `custom` category are currently **enabled**.

**Aggregate Rules**

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
      # files can be ignored for any individual rule
      # in this example, test files are ignored
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

capabilities:
  from:
    # optionally configure Regal to target a specific version of OPA
    # this will disable rules that has dependencies to e.g. built-in
    # functions or features not supported by the given version
    #
    # if not provided, Regal will use the capabilities of the latest
    # version of OPA available at the time of the Regal release
    engine: opa
    version: v0.58.0

ignore:
  # files can be excluded from all lint rules according to glob-patterns
  files:
    - file1.rego
    - "*_tmp.rego"
```

Regal will automatically search for a configuration file (`.regal/config.yaml`) in the current directory, and if not
found, traverse the parent directories either until either one is found, or the top of the directory hierarchy is
reached. If no configuration file is found, Regal will use the default configuration.

A custom configuration may be also be provided using the `--config-file`/`-c` option for `regal lint`, which when
provided will be used to override the default configuration.

## Ignoring Rules

If one of Regal's rules doesn't align with your team's preferences, don't worry! Regal is not meant to be the law,
and some rules may not make sense for your project, or parts of it.
Regal provides several different methods to ignore rules with varying precedence.
The available methods are (ranked highest to lowest precedence):

- [Inline Ignore Directives](#inline-ignore-directives) cannot be overridden by any other method.
- Enabling or Disabling Rules with CLI flags.
  - Enabling or Disabling Rules with `--enable` and `--disable` CLI flags.
  - Enabling or Disabling Rules with `--enable-category` and `--disable-category` CLI flags.
  - Enabling or Disabling All Rules with `--enable-all` and `--disable-all` CLI flags.
  - See [Ignoring Rules via CLI Flags](#ignoring-rules-via-cli-flags) for more details.
- [Ignoring a Rule In Config](#ignoring-a-rule-in-config)
- [Ignoring a Category In Config](#ignoring-a-category-in-config)
- [Ignoring All Rules In Config](#ignoring-all-rules-in-config)

In summary, the CLI flags will override any configuration provided in the file, and inline ignore directives for a
specific line will override any other method.

It's also possible to ignore messages on a per-file basis. The available methods are (ranked High to Lowest precedence):

- Using the `--ignore-files` CLI flag.
  See [Ignoring Rules via CLI Flags](#ignoring-rules-via-cli-flags).
- [Ignoring Files Globally](#ignoring-files-globally) or
  [Ignoring a Rule in Some Files](#ignoring-a-rule-in-some-files).

### Ignoring a Rule In Config

If you want to ignore a rule, set its level to `ignore` in the configuration file:

```yaml
rules:
  style:
    prefer-snake-case:
      # At example.com, we use camel case to comply with our naming conventions
      level: ignore
```

### Ignoring a Category In Config

If you want to ignore a category of rules, set its default level to `ignore` in the configuration file:

```yaml
rules:
  style:
    default:
      level: ignore
```

### Ignoring All Rules In Config

If you want to ignore all rules, set the default level to `ignore` in the configuration file:

```yaml
rules:
  default:
    level: ignore
  # then you can re-enable specific rules or categories
  testing:
    default:
      level: error
  style:
    opa-fmt:
      level: error
```

**Tip**: providing a comment on ignored rules is a good way to communicate why the decision was made.

### Ignoring a Rule in Some Files

You can use the `ignore` attribute inside any rule configuration to provide a list of files, or patterns, that should
be ignored for that rule:

```yaml
rules:
  style:
    line-length:
      level: error
      ignore:
        files:
          # ignore line length in test files to accommodate messy test data
          - "*_test.rego"
          # specific file used only for testing
          - "scratch.rego"
```

### Ignoring Files Globally

If you want to ignore certain files for all rules, you can use the global ignore attribute in your configuration file:

```yaml
ignore:
  files:
    - file1.rego
    - "*_tmp.rego"
```

### Inline Ignore Directives

If you'd like to ignore a specific violation in a file, you can add an ignore directive above the line in question, or
alternatively on the same line to the right of the expression:

```rego
package policy

import rego.v1

# regal ignore:prefer-snake-case
camelCase := "yes"

list_users contains user if { # regal ignore:avoid-get-and-list-prefix
    some user in data.db.users
    # ...
}
```

The format of an ignore directive is `regal ignore:<rule-name>,<rule-name>...`, where `<rule-name>` is the name of the
rule to ignore. Multiple rules may be added to the same ignore directive, separated by commas.

Note that at this point in time, Regal only considers the same line or the line following the ignore directive, i.e. it
does not apply to entire blocks of code (like rules, functions or even packages). See [configuration](#configuration)
if you want to ignore certain rules altogether.

### Ignoring Rules via CLI Flags

For development and testing, rules or classes of rules may quickly be enabled or disabled using the relevant CLI flags
for the `regal lint` command:

- `--disable-all` disables **all** rules
- `--disable-category` disables all rules in a category, overriding `--enable-all` (may be repeated)
- `--disable` disables a specific rule, overriding `--enable-all` and `--enable-category` (may be repeated)
- `--enable-all` enables **all** rules
- `--enable-category` enables all rules in a category, overriding `--disable-all` (may be repeated)
- `--enable` enables a specific rule, overriding `--disable-all` and `--disable-category` (may be repeated)
- `--ignore-files` ignores files using glob patterns, overriding `ignore` in the config file (may be repeated)

**Note:** all CLI flags override configuration provided in file.

## Capabilities

By default, Regal will lint your policies using the
[capabilities](https://www.openpolicyagent.org/docs/latest/deployments/#capabilities) of the latest version of OPA
known to Regal (i.e. the latest version of OPA at the time Regal was released). Sometimes you might want to tell Regal
that some rules aren't applicable to your project (yet!). As an example, if you're running OPA v0.46.0, you likely won't
be helped by the [custom-has-key](https://docs.styra.com/regal/rules/idiomatic/custom-has-key-construct) rule, as it
suggests using the `object.keys` built-in function introduced in OPA v0.47.0. The opposite could also be true —
sometimes new versions of OPA will invalidate rules that applied to older versions. An example of this is the upcoming
introduction of `import rego.v1`, which will make
[implicit-future-keywords](https://docs.styra.com/regal/rules/imports/implicit-future-keywords) obsolete, as importing
`rego.v1` automatically imports all "future" functions.

Capabilities help you tell Regal which features to take into account, and rules with dependencies to capabilities
not available or not applicable in the given version will be skipped.

If you'd like to target a specific version of OPA, you can include a `capabilities` section in your configuration,
providing either a specific `version` of an `engine` (currently only `opa` supported):

```yaml
capabilities:
  from:
    engine: opa
    version: v0.58.0
```

You can also choose to import capabilities from a file:

```yaml
capabilities:
  from:
    file: build/capabilities.json
```

You can use `plus` and `minus` to add or remove built-in functions from the given set of capabilities:

```yaml
capabilities:
  from:
    engine: opa
    version: v0.58.0
  minus:
    builtins:
      # exclude rules that depend on the http.send built-in function
      - name: http.send
  plus:
    builtins:
      # make Regal aware of a custom "ldap.query" function
      - name: ldap.query
        type: function
        decl:
          args:
            - type: string
        result:
          type: object
```

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
- `sarif` - [SARIF](https://sarifweb.azurewebsites.net/) JSON output, for consumption by tools processing code analysis
  reports

## OPA Check and Strict Mode

Linting with Regal assumes syntactically correct Rego. If there are errors parsing any files during linting, the
process is aborted and any parser errors are logged similarly to OPA. OPA itself provides a "linter" of sorts,
via the `opa check` comand and its `--strict` flag. This checks the provided Rego files not only for syntax errors,
but also for OPA [strict mode](https://www.openpolicyagent.org/docs/latest/policy-language/#strict-mode) violations.

> **Note** It is recommended to run `opa check --strict` as part of your policy build process, and address any violations
> reported there before running Regal. Why both commands? Couldn't the strict mode checks be integrated in Regal?
> That would certainly be an option. However, most of the strict mode checks will be made default / mandatory as part
> of a future OPA 1.0 release, at which point they'd be made immediately obsolete as part of Regal. There are a few
> strict mode checks that likely will remain optional in OPA, and we may choose to integrate them into Regal in the
> future.
>
> Until then, the recommendation is to run both `opa check --strict` and `regal lint` as part of your policy build
> and test process.

## Regal Language Server

In order to support linting directly in editors and IDE's, Regal implements parts of the
[Language Server Protocol](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/)
(LSP). With Regal installed and available on your `$PATH`, editors like VS Code (using the
[OPA extension](https://github.com/open-policy-agent/vscode-opa)) can leverage Regal for diagnostics, i.e. linting,
and have the results displayed directly in your editor as you work on your Rego policies. The Regal LSP implementation
doesn't stop at linting though — it'll also provide features like tooltips on hover, go to definition, and document
symbols helping you easily navigate the Rego code in your workspace.

The Regal language server currently supports the following LSP features:

- [x] Diagnostics (linting)
- [x] Hover (for inline docs on built-in functions)
- [x] Go to definition (ctrl/cmd + click on a reference to go to definition)
- [x] Folding ranges (expand/collapse blocks, imports, comments)
- [x] Document and workspace symbols (navigate to rules, functions, packages)
- [x] Inlay hints (show names of built-in function arguments next to their values)
- [x] Formatting
- [x] Code actions (quick fixes for linting issues)
  - [x] [opa-fmt](https://docs.styra.com/regal/rules/style/opa-fmt)
  - [x] [use-rego-v1](https://docs.styra.com/regal/rules/imports/use-rego-v1)
  - [x] [use-assignment-operator](https://docs.styra.com/regal/rules/style/use-assignment-operator)
  - [x] [no-whitespace-comment](https://docs.styra.com/regal/rules/style/no-whitespace-comment)

See the [editor Support](/docs/editor-support.md) page for information about Regal support in different editors.

## Resources

### Documentation

- [Custom Rules](/docs/custom-rules.md) describes how to develop your own linter rules
- [Architecture](/docs/architecture.md) provides a high-level technical overview of how Regal works
- [Development](/docs/development.md) contains information about how to hack on Regal itself
- [Go Integration](/docs/integration.md) describes how to integrate Regal in your Go application
- [Rego Style Guide](/docs/rego-style-guide.md) contains notes on implementing the
  [Rego Style Guide](https://github.com/StyraInc/rego-style-guide) rules
- [Pre-Commit Hooks](/docs/pre-commit-hooks.md) describes how to use Regal in pre-commit hooks
- [Editor Support](/docs/editor-support.md) contains information about editor support for Regal

### Talks

[Regal the Rego Linter](https://www.youtube.com/watch?v=Xx8npd2TQJ0&t=2567s), CNCF London meetup, June 2023
[![Regal the Rego Linter](/docs/assets/regal_cncf_london.png)](https://www.youtube.com/watch?v=Xx8npd2TQJ0&t=2567s)

### Blogs and Articles

- [Guarding the Guardrails - Introducing Regal the Rego Linter](https://www.styra.com/blog/guarding-the-guardrails-introducing-regal-the-rego-linter/)
  by @charlieegan3
- [Scaling Open Source Community by Getting Closer to Users](https://thenewstack.io/scaling-open-source-community-by-getting-closer-to-users/)
  by @anderseknert
- [Linting Rego with... Rego!](https://www.styra.com/blog/linting-rego-with-rego/) by @anderseknert

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
