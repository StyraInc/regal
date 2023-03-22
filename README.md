# Regal

[![Build Status](https://github.com/styrainc/regal/workflows/Build/badge.svg?branch=main)](https://github.com/styrainc/regal/actions)
![OPA v0.49.2](https://openpolicyagent.org/badge/v0.50.2)

Regal is a linter for Rego, with the goal of making your Rego magnificent!

> regal
>
> 1 : of, relating to, or suitable for a king
>
> 2 : of notable excellence or magnificence : splendid

\- [Merriam Webster](https://www.merriam-webster.com/dictionary/regal)

Regal rules are to as large extent as possible
[written in Rego](https://www.styra.com/blog/linting-rego-with-rego/) themselves,
using the JSON representation of the Rego abstract syntax tree (AST) as input, a
few additional custom built-in functions and some indexed data structures to help
with linting.

## Try it out!

Run `regal lint` pointed at one or more files or directories to have them linted:

```shell
regal lint policy/
```

## Rules

This table should be generated from metadata annotations, and inserted here as part of the build process.
See: https://github.com/StyraInc/regal/issues/36

| Category   | Rule                                                                                  | Description                     | Enabled |
|------------|---------------------------------------------------------------------------------------|---------------------------------|---------|
| Assignment | [use-assignment-operator](https://docs.styra.com/regal/rules/use-assignment-operator) | Prefer := over = for assignment | true    |
| Comments   | [todo-comment](https://docs.styra.com/regal/rules/todo-comment)                       | Avoid TODO comments             | true    |

## Configuration

A custom configuration file may be used to override the [default configuration](bundle/regal/config/data.yaml) options
provided by Regal. This is particularly useful for e.g. enabling/disabling certain rules, or to provide more
fine-grained options for the linter rules that support it.

**.regal/config.yaml**
```yaml
rules:
  style:
    prefer-snake-case:
      enabled: false
```

Regal will automatically search for a configuration file (`.regal/config.yaml`) in the current directory, and if not
found, traverse the parent directories either until either one is found, or the top of the directory hierarchy is 
reached. If no configuration file is found, Regal will use the default configuration.

A custom configuration may be also be provided using the `--config-file`/`-c` option for `regal lint`, which when 
provided will be used to override the default configuration.

## Custom Rules

Regal is built to be easily extended. Creating custom rules is a great way to enforce naming conventions, best practices
or more opinionated rules across teams and organizations. If you'd like to provide your own linter rules for a project,
you may do so by placing them in a `rules` directory inside the `.regal` directory preferably placed in the root of your
project (which is also where [custom configuration](#configuration) resides).

If you so prefer, custom rules may also be provided using the `--rules` option for `regal lint`, which may point either
to a Rego file, or a directory containing Rego files and potentially data (JSON or YAML).

### Developing Rules

Regal rules works primarily on the abstract syntax tree (AST) as parsed by OPA. The AST of each policy scanned will be
provided as input to the linter policies, and additional data useful in the context of linting, as well as some
purpose-built custom functions are made available in any Regal policy. 

If we were to write the simplest policy possible, it would contain nothing but a package declaration:

**policy.rego**
```rego
package policy
```

Using `opa parse --json policy.rego`, we're provided with the AST of the above policy:

```json
{
  "package": {
    "path": [
      {
        "type": "var",
        "value": "data"
      },
      {
        "type": "string",
        "value": "policy"
      }
    ]
  }
}
```

As trivial as may be, it's enough to build our first linter rule! Let's say we'd like to enforce a uniform naming
convention on any policy in a repository. Packages may be named anything, but must start with the name of the
organization (Acme Corp). So `package acme.corp.policy` should be allowed, but not `package policy` or
`package policy.acme.corp`. One exception: policy authors should be allowed to write policy for the `system.log` package
provided by OPA to allow
[masking](https://www.openpolicyagent.org/docs/latest/management-decision-logs/#masking-sensitive-data) sensitive data
from decision logs.

An example policy to implement this requirement might look something like this:

```rego
package custom.regal.rules.naming

import future.keywords.contains
import future.keywords.if

import data.regal

# METADATA
# title: acme-corp-package
# description: All packages must use "acme.corp" base name
# related_resources:
# - description: documentation
#   ref: https://www.acmecorp.example.org/docs/regal/package
# custom:
#   category: naming
report contains violation if {
	not acme_corp_package
	not system_log_package

	violation := regal.fail(rego.metadata.rule(), {})
}

acme_corp_package if {
	input["package"].path[1].value == "acme"
	input["package"].path[2].value == "corp"
}

system_log_package if {
	input["package"].path[1].value == "system"
	input["package"].path[2].value == "log"
}
```

Starting from top to bottom, these are the components comprising our custom rule:

1. The package of custom rules **must** start with `custom.regal.rules`, followed by the category of the rule.
2. Importing `data.regal` provides policy authors with custom built-in functions to help author rules, but is optional.
3. Regal rules make heavy use of [metadata annotations](https://www.openpolicyagent.org/docs/latest/annotations/) in
   order to document the purpose of the rule, along with any other information that could potentially be useful.
   All rules **must** have a `title`, a `description`, and a `category` (placed under the `custom` object). Providing
   links to additional documentation under `related_resources` is recommended, but not required.
4. Regal will evaluate any rule named `report` in each linter policy, so at least one `report` rule **must** be present.
5. In our example `report` rule, we evaluate another rule (`acme_corp_package`) in order to know if the package name 
   starts with `acme.corp`, and another rule (`system_log_package`) to know if it starts with `system.log`. If neither
   of the conditions are true, the rule fails and violation is created.
6. The violation is created by calling `regal.fail`, which takes the metadata from the rule and returns it, which will
   later be included in the final report provided by Regal.

## Development

### Building

1. Run the `build.sh` script to populate the `data` directory with any data necessary for
   linting (such as the built-in function metadata from OPA).
2. Build the `regal` executable by running `go build`.

### Testing

To run all Rego tests:

```shell
opa test policy data
```

### Authoring Rules

During development of Rego-based rules, you may want to test the policies in isolation — i.e. without running Regal.
Since Regal policies and data are kept in a regular
[bundle](https://www.openpolicyagent.org/docs/latest/management-bundles/) structure, this is simple. Given we want to 
test `p.rego` against the available set of rules, we can have OPA parse it and pipe the output to `opa eval` for
evaluation:

```shell
$ opa parse p.rego --format json | opa eval -f pretty -b bundle -I data.regal.main.report
[
  {
    "category": "variables",
    "description": "Unconditional assignment in rule body",
    "related_resources": [
      {
        "description": "documentation",
        "ref": "https://docs.styra.com/regal/rules/unconditional-assignment"
      }
    ],
    "title": "unconditional-assignment"
  }
]
```
