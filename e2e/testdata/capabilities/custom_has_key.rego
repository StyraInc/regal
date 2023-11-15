package custom_has_key

import future.keywords.if

has_key(map, key) if {
	_ = map[key]
}
