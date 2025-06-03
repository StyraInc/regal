<!-- markdownlint-disable MD041 -->

## OPA Check and Strict Mode

OPA itself provides a "linter" of sorts, via the `opa check` command and its `--strict` flag. This checks the provided
Rego files not only for syntax errors, but also for OPA
[strict mode](https://www.openpolicyagent.org/docs/policy-language/#strict-mode) violations. Most of the strict
mode checks from before OPA 1.0 have now been made default checks in OPA, and only two additional checks are currently
provided by the `--strict` flag. Those are both important checks not covered by Regal though, so our recommendation is
to run `opa check --strict` against your policies before linting with Regal.
