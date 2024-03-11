# Editor support

## Visual Studio Code

[vscode-opa](https://marketplace.visualstudio.com/items?itemName=tsandall.opa) -
the official OPA extension for Visual Studio Code - now supports the Regal language server.

To see Regal linting as you work, install the extension at version `0.13.3` or later
and open a workspace with Rego files.

The plugin will automatically find and use
[Regal config](https://docs.styra.com/regal#configuration).

## Neovim via none-ls

[none-ls](https://github.com/nvimtools/none-ls.nvim) - Use Neovim as a language server to inject LSP diagnostics,
code actions, and more via Lua.

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

diagnostics may look like this.

![regal in none-ls](./assets/editors-neovim.png)
