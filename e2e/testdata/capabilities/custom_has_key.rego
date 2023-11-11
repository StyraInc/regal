package custom_has_key

import rego.v1

has_key(map, key) if {
	_ = map[key]
}
