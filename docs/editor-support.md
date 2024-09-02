# Editor support

## Visual Studio Code

[vscode-opa](https://marketplace.visualstudio.com/items?itemName=tsandall.opa) -
the official OPA extension for Visual Studio Code - now supports the Regal language server.

To see Regal linting as you work, install the extension at version `0.13.3` or later
and open a workspace with Rego files.

The plugin will automatically find and use [Regal config](https://docs.styra.com/regal#configuration).

## Zed

[Zed](https://zed.dev) is a modern open-source code editor with focus on performance and simplicity.

Zed supports Rego via Regal and the [zed-rego](https://github.com/StyraInc/zed-rego) extension developed by the Styra
community. The extension provides syntax highlighting, linting, and most of the other language server features provided
by Regal.

## Neovim

There are a number of different plugins available for [Neovim](https://neovim.io/) which integrate
with language servers using the Language Server Protocol.

Generally, the Regal binary should be [installed](https://docs.styra.com/regal#getting-started)
first. [`mason.vim`](https://github.com/williamboman/mason.nvim) users can install the
Regal binary with `:MasonInstall regal`
([package definition](https://github.com/mason-org/mason-registry/blob/2024-07-23-asian-hate/packages/regal/package.yaml)).

Below are a number of different plugin options to configure a language server
client for Regal in Neovim.

### nvim-lspconfig

[nvim-lspconfig](https://github.com/neovim/nvim-lspconfig) has native support for the
Regal language server. Use the configuration below to configure Regal:

```lua
require('lspconfig').regal.setup()
```

### none-ls

[none-ls](https://github.com/nvimtools/none-ls.nvim) - Use Neovim as a
language server to inject LSP diagnostics, code actions, and more via Lua.

Minimal installation via [VimPlug](https://github.com/junegunn/vim-plug)

```vim
Plug 'nvim-lua/plenary.nvim'
Plug 'nvimtools/none-ls.nvim'

lua <<EOF
local null_ls = require("null-ls")
null_ls.setup {
    sources = { null_ls.builtins.diagnostics.regal }
}
EOF
```

Using sample rego file `test.rego` with following content

```rego
package test

default allowRbac := true
```

Example of the diagnostics in as shown in the UI:

![regal in none-ls](./assets/editors-neovim.png)

### nvim-cmp

[nvim-cmp](https://github.com/hrsh7th/nvim-cmp) supports the adding of language
servers as a source.

To use Regal with `nvim-cmp`, it is recommended that you use
the [`nvim-lspconfig` source](https://github.com/hrsh7th/cmp-nvim-lsp) and
follow the instructions above to configure `nvim-lspconfig`.

### Other plugins

To see live linting of Rego, your plugin must support
[`textDocument/diagnostic`](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_diagnostic)
messages.

There are many language server integrations for Neovim, if you'd like to see
another one listed, please [open an issue](https://github.com/StyraInc/regal/issues/new)
or drop us a message in [Slack](http://communityinviter.com/apps/styracommunity/signup).
