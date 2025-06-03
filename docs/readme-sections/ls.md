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
