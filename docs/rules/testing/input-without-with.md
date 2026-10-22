# input-without-with

**Summary**: Reference to `input` in test without using `with`

**Category**: Testing

**Automatically fixable**: No

**Avoid**
```rego
package policy_test

import rego.v1

test_something if {
    my_function(input) == "yes"
}
```

**Prefer**
```rego
package policy_test

import rego.v1

test_something if {
    my_function(input) == "yes" with input as {"foo": "bar"}
}
```

## Rationale

When evaluating policy tests using OPA or Regal, test rules are expected to provide mock data
for `input` using `with` statements (e.g. `with input as ...`). Referencing `input` directly in a
test without mocking it via `with` relies on an empty or ambient input document, which is almost
certainly unintentional and can lead to brittle or unrepeatable tests.

## Configuration Options

This linter rule provides the following [configuration options](https://www.openpolicyagent.org/projects/regal/configuration):

```yaml title=".regal/config.yaml or .regal.yaml"
rules:
  testing:
    input-without-with:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- GitHub: [Source Code](https://github.com/open-policy-agent/regal/blob/main/bundle/regal/rules/testing/input-without-with/input_without_with.rego)
