<!-- markdownlint-disable MD041 -->

## Getting Started

### Download Regal

**MacOS and Linux**

```shell
brew install regal
```

<details>
  <summary><strong>Other Installation Options</strong></summary>

Please see [Packages](https://docs.styra.com/regal/adopters#packaging)
for a list of package repositories which distribute Regal.

Manual installation commands:

**MacOS (Apple Silicon)**

```shell
curl -L -o regal "https://github.com/open-policy-agent/regal/releases/latest/download/regal_Darwin_arm64"
```

**MacOS (x86_64)**

```shell
curl -L -o regal "https://github.com/open-policy-agent/regal/releases/latest/download/regal_Darwin_x86_64"
```

**Linux (x86_64)**

```shell
curl -L -o regal "https://github.com/open-policy-agent/regal/releases/latest/download/regal_Linux_x86_64"
chmod +x regal
```

**Windows**

```shell
curl.exe -L -o regal.exe "https://github.com/open-policy-agent/regal/releases/latest/download/regal_Windows_x86_64.exe"
```

**Docker**

```shell
docker pull ghcr.io/styrainc/regal:latest
```

See all versions, and checksum files, at the Regal [releases](https://github.com/open-policy-agent/regal/releases/)
page, and published Docker images at the [packages](https://github.com/open-policy-agent/regal/pkgs/container/regal)
page.

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
