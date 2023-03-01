# Regal

[![Build Status](https://github.com/styrainc/regal/workflows/Build/badge.svg?branch=main)](https://github.com/styrainc/regal/actions)
![OPA v0.49.2](https://openpolicyagent.org/badge/v0.49.2)

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

Run `regal` pointed at one or more files or directories to have them linted:

```shell
./regal policy/
```

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
