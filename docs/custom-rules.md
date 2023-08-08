# Custom Rules

Regal is built to be easily extended. Creating custom rules is a great way to enforce naming conventions, best practices
or more opinionated rules across teams and organizations. If you'd like to provide your own linter rules for a project,
you may do so by placing them in a `rules` directory inside the `.regal` directory preferably placed in the root of your
project (which is also where custom configuration resides). The directory structure of a policy repository with custom
linter rules might then look something like this:

```text
.
├── .regal
│   ├── config.yaml
│   └── rules
│       ├── naming.rego
│       └── naming_test.rego
└── policy
    ├── authz.rego
    └── authz_test.rego
```

If you so prefer, custom rules may also be provided using the `--rules` option for `regal lint`, which may point either
to a Rego file, or a directory containing Rego files and potentially data (JSON or YAML).

## Developing Rules

Regal rules works primarily on the [abstract syntax tree](https://en.wikipedia.org/wiki/Abstract_syntax_tree) (AST) as
parsed by OPA, with a few custom additions. The AST of each policy scanned will be provided as input to the linter
policies, and additional data useful in the context of linting, as well as some purpose-built custom functions are made
available in any Regal policy.

If we were to write the simplest policy possible, and parse it using `opa parse`, it would contain nothing but a package
declaration:

**policy.rego**
```rego
package policy
```

Using `opa parse --format json --json-include locations policy.rego`, we're provided with the AST of the above policy:

```json
{
  "package": {
    "location": {
      "file": "policy.rego",
      "row": 1,
      "col": 1
    },
    "path": [
      {
        "location": {
          "file": "policy.rego",
          "row": 1,
          "col": 9
        },
        "type": "var",
        "value": "data"
      },
      {
        "location": {
          "file": "policy.rego",
          "row": 1,
          "col": 9
        },
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
# METADATA
# description: All packages must use "acme.corp" base name
# related_resources:
# - description: documentation
#   ref: https://www.acmecorp.example.org/docs/regal/package
# schemas:
# - input: schema.regal.ast
package custom.regal.rules.naming["acme-corp-package"]

import future.keywords.contains
import future.keywords.if

import data.regal.result

report contains violation if {
    not acme_corp_package
    not system_log_package

    violation := result.fail(rego.metadata.chain(), result.location(input["package"].path[1]))
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

1. The package of custom rules **must** start with `custom.regal.rules`, followed by the category of the rule, and the
   title (which is commonly quoted as rule names use `-` for spaces).
1. The `data.regal.result` provides some helpers for formatting the result of a violation for inclusion in a report.
1. Regal rules make heavy use of [metadata annotations](https://www.openpolicyagent.org/docs/latest/annotations/) in
   order to document the purpose of the rule, along with any other information that could potentially be useful.
   All rule packages **must** have a `description`. Providing links to additional documentation under
   `related_resources` is recommended, but not required.
1. Note the `schema` attribute present in the metadata annotation. Adding this is optional, but highly recommended, as
   it will make the compiler aware of the structure of the input, i.e. the AST. This allows the compiler to fail when
   unknown attributes are referenced, due to typos or other mistakes. The compiler will also fail when an attribute is
   referenced using a type it does not have, like referring to a string as if it was a number. Set to `schema.regal.ast`
   to use the AST schema provided by Regal.
1. Regal will evaluate any rule named `report` in each linter policy, so at least one `report` rule **must** be present.
1. In our example `report` rule, we evaluate another rule (`acme_corp_package`) in order to know if the package name
   starts with `acme.corp`, and another rule (`system_log_package`) to know if it starts with `system.log`. If neither
   of the conditions are true, the rule fails and violation is created.
1. The violation is created by calling `result.fail`, which takes the metadata from the package (using
   `rego.metadata.chain` which conveniently also includes the path of the package) and returns a result, which 
   will later be included in the final report provided by Regal.
1. The `result.location` helps extract the location from the element failing the test. Make sure to use it!

## Parsing and Testing

Regal provides a few tools mirrored from OPA in order to help test and debug custom rules. These are necessary since OPA
is not aware of the custom [built-in functions](#built-in-functions) included in Regal, and will fail when encountering
e.g. `regal.parse_module` in a custom linter policy. The following commands are included with Regal to help you author
custom rules:

- `regal parse` works similarly to `opa parse`, but will always output JSON and include location information, and any
  additional data added to the AST by Regal. Use this if you want to know exactly what the `input` will look like for
  any given policy, when provided to Regal for linting.
- `regal test` works like `opa test`, but aware of any custom Regal additions, and the schema used for the AST. Use this
  to test custom linter rules, e.g. `regal test .regal/rules`.

Note that the `print` built-in function is enabled for `regal test`. Good to use for quick debugging!

## Built-in Functions

Regal provides a few custom built-in functions tailor-made for linter policies.

### `regal.parse_module(filename, policy)`

Works just like `rego.parse_module`, but provides an AST including location information, and custom additions added
by Regal, like the text representation of each line in the original policy. This is useful for authoring tests to assert
linter rules work as expected.

### `regal.json_pretty(data)`

Printing nested objects and arrays is quite helpful for debugging AST nodes, but the standard representation — where
everything is displayed on a single line — not so much. This built-in allows marshalling JSON similar to `json.marshal`,
but with newlines and spaces added for a more pleasant experience.

In addition to this, Regal provides many helpful functions, rules and utilities in Rego. Browsing the source code of the
[regal.ast](https://github.com/StyraInc/regal/blob/main/bundle/regal/ast.rego) package to see what's available is
recommended!

## Community

If you'd like to discuss custom rules development or just talk about Regal in general, please join us in the `#regal`
channel in the Styra Community[Slack](https://communityinviter.com/apps/styracommunity/signup)!
