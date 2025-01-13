# import-shadows-import

**Summary**: Import shadows import

**Category**: Imports

## Notice: Rule made obsolete by OPA 1.0

Since Regal v0.30.0, this rule is only enabled for projects that have either been explicitly configured to target
versions of OPA before 1.0, or if no configuration is provided — where Regal is able to determine that an older version
of OPA/Rego is being targeted. Consult the documentation on Regal's
[configuration](https://docs.styra.com/regal#configuration) for information on how to best work with older versions of
OPA and Rego.

Since OPA v1.0, this rule is automatically disabled as OPA itself now forbids this, and shadowed imports will result in
a parse error.

**Avoid**
```rego
package policy

import data.permissions
import data.users

# Already imported
import data.permissions
```

**Prefer**
```rego
package policy

import data.permissions
import data.users
```

## Rationale

Duplicate imports are redundant, and while harmless, should just be removed.

## Configuration Options

This linter rule provides the following configuration options:

```yaml
rules:
  imports:
    import-shadows-import:
      # one of "error", "warning", "ignore"
      level: error
```

## Related Resources

- OPA Docs: [Strict Mode](https://www.openpolicyagent.org/docs/latest/policy-language/#strict-mode)
- GitHub: [Source Code](https://github.com/StyraInc/regal/blob/main/bundle/regal/rules/imports/import-shadows-import/import_shadows_import.rego)

## Community

If you think you've found a problem with this rule or its documentation, would like to suggest improvements, new rules,
or just talk about Regal in general, please join us in the `#regal` channel in the Styra Community
[Slack](https://communityinviter.com/apps/styracommunity/signup)!
