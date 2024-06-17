# METADATA
# description: |
#   base package for completion suggestion provider policies, and acts
#   like a router that'll collection suggestions from all provider policies
#   under regal.lsp.completion.providers
package regal.lsp.completion

import rego.v1

# METADATA
# description: main entry point for completion suggestions
# entrypoint: true
items contains data.regal.lsp.completion.providers[_].items[_]
