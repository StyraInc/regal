<!-- Please see docs/readme-sections/ to update readme content -->

<!-- markdownlint-disable MD041 -->

# Regal


<!-- markdownlint-disable MD041 -->

[![Build Status](https://github.com/styrainc/regal/workflows/Build/badge.svg)](https://github.com/styrainc/regal/actions)
![OPA v1.5.0](https://openpolicyagent.org/badge/v1.5.0)
[![codecov](https://codecov.io/github/StyraInc/regal/graph/badge.svg?token=EQK01YF3X3)](https://codecov.io/github/StyraInc/regal)
[![Downloads](https://img.shields.io/github/downloads/styrainc/regal/total.svg)](https://github.com/StyraInc/regal/releases)


<!-- markdownlint-disable MD041 -->

Regal is a linter and language server for [Rego](https://www.openpolicyagent.org/docs/policy-language/), making
your Rego magnificent, and you the ruler of rules!

With its extensive set of linter rules, documentation and editor integrations, Regal is the perfect companion for policy
development, whether you're an experienced Rego developer or just starting out.


<!-- markdownlint-disable MD033 -->
<!-- markdownlint-disable MD041 -->

<img
  src="/docs/assets/regal-banner.png"
  alt="illustration of a viking representing the Regal logo"
  width="150px" />

> regal
>
> adj : of notable excellence or magnificence : splendid

\- [Merriam Webster](https://www.merriam-webster.com/dictionary/regal)


<!-- markdownlint-disable MD041 -->

## Goals

- Deliver an outstanding policy development experience by providing the best possible tools for that purpose
- Identify common mistakes, bugs and inefficiencies in Rego policies, and suggest better approaches
- Provide advice on [best practices](https://github.com/StyraInc/rego-style-guide), coding style, and tooling
- Allow users, teams and organizations to enforce custom rules on their policy code


<!-- markdownlint-disable MD041 -->

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

See the [adopters](https://docs.styra.com/regal/adopters) file for more Regal users.


<!-- markdownlint-disable MD041 -->

## Getting Started

### Download Regal

**MacOS and Linux**

```shell
brew install styrainc/packages/regal
```

<details>
  <summary><strong>Other Installation Options</strong></summary>

Please see [Packages](https://docs.styra.com/regal/adopters#packaging)
for a list of package repositories which distribute Regal.

Manual installation commands:

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

### Using Regal in Your Editor

Linting from the command line is a great way to get started with Regal, and even for some experienced developers
the preferred way to work with the linter. However, not only is Regal a linter, but a full-fledged development
companion for Rego development!

Integrating Regal in your favorite editor means you'll get immediate feedback from the linter as you work on your
policies. More than that, it'll unlock a whole new set of features that leverage Regal's
[language server](https://docs.styra.com/regal/language-server),
like context-aware completion suggestions, informative tooltips on hover,
or go-to-definition.

Elevate your policy development experience with Regal in VS Code, Neovim, Zed, Helix
and more on our [Editor Support page](https://docs.styra.com/regal/editor-support)!

To learn more about the features provided by the Regal language server, see the
[Language Server](https://docs.styra.com/regal/language-server) page.

### Using Regal in Your Build Pipeline

To ensure Regal's rules are enforced consistently in your project or organization,
we've made it easy to run Regal as part of your builds.
See the docs on [Using Regal in your build pipeline](https://docs.styra.com/regal/cicd) to learn more
about how to set up Regal to lint your policies on every commit or pull request.


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


<!-- markdownlint-disable MD041 -->

## OPA Check and Strict Mode

OPA itself provides a "linter" of sorts, via the `opa check` command and its `--strict` flag. This checks the provided
Rego files not only for syntax errors, but also for OPA
[strict mode](https://www.openpolicyagent.org/docs/policy-language/#strict-mode) violations. Most of the strict
mode checks from before OPA 1.0 have now been made default checks in OPA, and only two additional checks are currently
provided by the `--strict` flag. Those are both important checks not covered by Regal though, so our recommendation is
to run `opa check --strict` against your policies before linting with Regal.


<!-- markdownlint-disable MD041 -->

## Regal Language Server

In order to support linting directly in editors and IDE's, Regal implements parts of the
[Language Server Protocol](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/)
(LSP). With Regal installed and available on your `$PATH`, editors like VS Code (using the
[OPA extension](https://github.com/open-policy-agent/vscode-opa)) and Zed (using the
[zed-rego](https://github.com/StyraInc/zed-rego) extension) can leverage Regal for diagnostics, i.e. linting,
and have the results displayed directly in your editor as you work on your Rego policies. The Regal LSP implementation
doesn't stop at linting though — it'll also provide features like tooltips on hover, go to definition, and document
symbols helping you easily navigate the Rego code in your workspace.

The Regal language server currently supports the following LSP features:

- [x] [Diagnostics](https://docs.styra.com/regal/language-server#diagnostics) (linting)
- [x] [Hover](https://docs.styra.com/regal/language-server#hover)
      (for inline docs on built-in functions)
- [x] [Go to definition](https://docs.styra.com/regal/language-server#go-to-definition)
      (ctrl/cmd + click on a reference to go to definition)
- [x] [Folding ranges](https://docs.styra.com/regal/language-server#folding-ranges)
      (expand/collapse blocks, imports, comments)
- [x] [Document and workspace symbols](https://docs.styra.com/regal/language-server#document-and-workspace-symbols)
      (navigate to rules, functions, packages)
- [x] [Inlay hints](https://docs.styra.com/regal/language-server#inlay-hints)
      (show names of built-in function arguments next to their values)
- [x] [Formatting](https://docs.styra.com/regal/language-server#formatting)
- [x] [Code completions](https://docs.styra.com/regal/language-server#code-completions)
- [x] [Code actions](https://docs.styra.com/regal/language-server#code-actions)
      (quick fixes for linting issues)
  - [x] [opa-fmt](https://docs.styra.com/regal/rules/style/opa-fmt)
  - [x] [use-rego-v1](https://docs.styra.com/regal/rules/imports/use-rego-v1)
  - [x] [use-assignment-operator](https://docs.styra.com/regal/rules/style/use-assignment-operator)
  - [x] [no-whitespace-comment](https://docs.styra.com/regal/rules/style/no-whitespace-comment)
  - [x] [directory-package-mismatch](https://docs.styra.com/regal/rules/idiomatic/directory-package-mismatch)
- [x] [Code lenses](https://docs.styra.com/regal/language-server#code-lenses-evaluation)
      (click to evaluate any package or rule directly in the editor)

See the
[documentation page for the language server](https://github.com/StyraInc/regal/blob/main/docs/language-server.md)
for an extensive overview of all features, and their meaning.

See the [Editor Support](https://docs.styra.com/regal/editor-support)
page for information about Regal support in different editors.


<!-- markdownlint-disable MD041 -->

## Regal and OPA 1.0+

Starting from version v0.30.0, Regal supports working with both
[OPA 1.0+](https://blog.openpolicyagent.org/announcing-opa-1-0-a-new-standard-for-policy-as-code-a6d8427ee828)
policies and Rego from earlier versions of OPA. While everything should work without additional configuration,
we recommend checking out our documentation on using Regal with [OPA 1.0](https://docs.styra.com/regal/opa-one-dot-zero)
and later for the best possible experience managing projects of any given Rego version, or even a mix of them.


<!-- If updating, please check resources-website.md too -->
<!-- markdownlint-disable MD041 -->

## Resources

### Documentation

Please see [Regal's Documentation Site](https://docs.styra.com/regal) for the
canonical documentation of Regal.

[Contributing](https://github.com/StyraInc/regal/blob/main/docs/CONTRIBUTING.md)
contains information about how to hack on Regal itself.

### Talks

- [OPA Maintainer Track, featuring Regal](https://www.youtube.com/watch?v=XtA-NKoJDaI), KubeCon London, 2025
- [Regal the Rego Linter](https://www.youtube.com/watch?v=Xx8npd2TQJ0&t=2567s), CNCF London meetup, June 2023
  [![Regal the Rego Linter](/docs/assets/regal_cncf_london.png)](https://www.youtube.com/watch?v=Xx8npd2TQJ0&t=2567s)

### Blogs and Articles

- [Guarding the Guardrails - Introducing Regal the Rego Linter](https://www.styra.com/blog/guarding-the-guardrails-introducing-regal-the-rego-linter/)
  by Anders Eknert ([@anderseknert](https://github.com/anderseknert))
- [Scaling Open Source Community by Getting Closer to Users](https://thenewstack.io/scaling-open-source-community-by-getting-closer-to-users/)
  by Charlie Egan ([@charlieegan3](https://github.com/charlieegan3))
- [Renovating Rego](https://www.styra.com/blog/renovating-rego/) by Anders Eknert ([@anderseknert](https://github.com/anderseknert))
- [Linting Rego with... Rego!](https://www.styra.com/blog/linting-rego-with-rego/) by Anders Eknert ([@anderseknert](https://github.com/anderseknert))
- [Regal: Rego(OPA)用リンタの導入手順](https://zenn.dev/erueru_tech/articles/6cfb886d92858a) by Jun Fujita ([@erueru-tech](https://github.com/erueru-tech))
- [Regal を使って Rego を Lint する](https://tech.dentsusoken.com/entry/2024/12/05/Regal_%E3%82%92%E4%BD%BF%E3%81%A3%E3%81%A6_Rego_%E3%82%92_Lint_%E3%81%99%E3%82%8B)
  by Shibata Takao ([@shibata.takao](https://shodo.ink/@shibata.takao/))


<!-- markdownlint-disable MD041 -->

## Status

Regal is currently in beta. End-users should not expect any drastic changes, but any API may change without notice.
If you want to embed Regal in another project or product, please reach out!


<!-- markdownlint-disable MD041 -->

## Roadmap

The current Roadmap items are all related to the preparation for
[Regal 1.0](https://github.com/StyraInc/regal/issues/979):

- [ ] [Go API: Refactor the Location object in Violation (#1554)](https://github.com/StyraInc/regal/issues/1554)
- [ ] [Rego API: Provide a stable and well-documented Rego API (#1555)](https://github.com/StyraInc/regal/issues/1555)
- [ ] [Go API: Audit and reduce the public Go API surface (#1556)](https://github.com/StyraInc/regal/issues/1556)
- [ ] [Custom Rules: Tighten up Authoring experience (#1559)](https://github.com/StyraInc/regal/issues/1559)
- [ ] [docs: Improve automated documentation generation (#1557)](https://github.com/StyraInc/regal/issues/1557)
- [ ] [docs: Break down README into smaller units (#1558)](https://github.com/StyraInc/regal/issues/1558)
- [ ] [lsp: Support a JetBrains LSP client (#1560)](https://github.com/StyraInc/regal/issues/1560)

If there's something you'd like to have added to the roadmap, either open an issue, or reach out in the community Slack!


<!-- markdownlint-disable MD041 -->

## Community

For questions, discussions and announcements related to Styra products, services and open source projects, please join
the Styra community on [Slack](https://inviter.co/styra)!


