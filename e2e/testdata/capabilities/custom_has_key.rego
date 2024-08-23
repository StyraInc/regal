package capabilities

import rego.v1

# custom_has_key
has_key(map, key) if {
	_ = map[key]
}
