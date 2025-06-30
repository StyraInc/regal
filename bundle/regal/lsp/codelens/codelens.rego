# METADATA
# description: |
#   the code lens provider decides where code lenses should be placed in the given input file
# schemas:
#   - input: schema.regal.ast
package regal.lsp.codelens

import data.regal.ast
import data.regal.result
import data.regal.util

import data.regal.lsp.util.location

# code lenses are displayed in the order they come back in the returned
# array, and 'evaluate' somehow feels better to the left of 'debug'
# METADATA
# description: contains code lenses determined for module
lenses := array.concat(
	[l | some l in _eval_lenses],
	[l | some l in _debug_lenses],
)

_eval_lenses contains {
	"range": location.to_range(result.location(input.package).location),
	"command": {
		"title": "Evaluate",
		"command": "regal.eval",
		"arguments": [json.marshal({
			"target": input.regal.file.name,
			"path": ast.ref_to_string(input.package.path),
			"row": util.to_location_object(input.package.location).row,
		})],
	},
}

_eval_lenses contains _rule_lens(input.regal.file.name, rule, "regal.eval", "Evaluate") if some rule in ast.rules

_debug_lenses contains {
	"range": location.to_range(result.location(input.package).location),
	"command": {
		"title": "Debug",
		"command": "regal.debug",
		"arguments": [json.marshal({
			"target": input.regal.file.name,
			"path": ast.ref_to_string(input.package.path),
			"row": util.to_location_object(input.package.location).row,
		})],
	},
}

_debug_lenses contains _rule_lens(input.regal.file.name, rule, "regal.debug", "Debug") if {
	some rule in ast.rules

	# no need to add a debug lens for a rule like `pi := 3.14`
	not _unconditional_constant(rule)
}

_rule_lens(filename, rule, command, title) := {
	"range": location.to_range(result.location(rule).location),
	"command": {
		"title": title,
		"command": command,
		"arguments": [json.marshal({
			"target": filename,
			"path": concat(".", [
				ast.ref_to_string(input.package.path),
				ast.ref_static_to_string(rule.head.ref),
			]),
			"row": util.to_location_object(rule.head.location).row,
		})],
	},
}

_unconditional_constant(rule) if {
	not rule.body
	ast.is_constant(rule.head.value)
}
