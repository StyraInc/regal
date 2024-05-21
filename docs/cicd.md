# Running Regal in CI/CD pipeline(s)

Its possible to use Regal to lint your Rego policies in your CI/CD pipeline(s)! This document will guide you on how to do so.

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

### GitLab CICD

// TODO