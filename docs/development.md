# Development

If you'd like to contribute to Regal, here are some pointers to help get you started. 

Before you start, the [architecture](./architecture) guide provides a useful overview of how Regal works, so you might
want to read that before diving into the code!

## Contributing New Rules

If you'd like to contribute a new built-in rule, the simplest way to get started is to run the `regal new rule` command.
This should be done from the top-level directory of the Regal repository, and would look something like this:

```shell
regal new rule --type builtin --category naming --name foo-bar-baz
```

This will create two files in `bundle/regal/rules/naming` (since `naming` was the category) — one for the rule and one
for testing it. The code here will be a pretty simple template, but contains all the required components for a built-in
rule. A good idea for learning more about what's needed is to take a look at some previous PRs adding new rules to
Regal.

## Building

Build the `regal` executable simply by running `go build`.

Occasionally you may want to run the `fetch_builtin_data.sh` script from inside the `build` directory. This will
populate the `data` directory with any data necessary for linting (such as the built-in function metadata from OPA).

## Testing

To run all the Rego unit tests, you can run the `regal test` command targeting the `bundle` directory:

```shell
regal test bundle
```

To run all tests — Go and Rego:

```shell
go test ./...
```

### E2E tests

End-to-End (E2E) tests assert the behaviour of the `regal` binary called with certain configs, and test files.
They expect a `regal` binary either in the top-level directory, or pointed at by `$REGAL_BIN`, and can be run
locally via

```shell
go test -tags e2e ./e2e
```

## Linting

Regal uses [golangci-lint](https://golangci-lint.run/) with most linters enabled. In order to check your code, run:

```shell
golangci-lint run ./...
```

In order to please the [gci](https://github.com/daixiang0/gci) linter, you may either manually order imports, or have
them automatically ordered and grouped by the tool:

```shell
gci write \
  -s standard \
  -s default \
  -s "prefix(github.com/open-policy-agent/opa)" \
  -s "prefix(github.com/styrainc/regal)" \
  -s blank \
  -s dot .
```
## Documentation

The table in the [Rules](../README.md#rules) section of the README is generated with the following command:

```shell
go run main.go table --write-to-readme bundle
```

## Community

If you'd like to discuss Regal development or just talk about Regal in general, please join us in the `#regal`
channel in the Styra Community [Slack](https://communityinviter.com/apps/styracommunity/signup)!
