package capabilities

import rego.v1

# This is here to make sure we deal with multiple notices correctly,
# and don't report duplicates multiple times.
has_key(map, key) if {
	_ = map[key]
}
