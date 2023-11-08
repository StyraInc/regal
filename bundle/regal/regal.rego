# METADATA
# scope: subpackages
# authors:
# - Styra
# related_resources:
# - https://www.styra.com
# schemas:
# - input: schema.regal.ast
package regal

default capabilities := {}

# METADATA
# description: |
#   The capabilities object for Regal itself. Use `config.capabilities`
#   to get the capabilities for the target environment (i.e. the policies
#   getting linted).
capabilities := data.internal.capabilities
