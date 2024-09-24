# METADATA
# description: utility module to help determine if an update of Regal should be recommended
package update

import rego.v1

default needs_update := false

# METADATA
# description: true if current version is behind latest version
# scope: document
needs_update if {
	current_version := trim(input.current_version, "v")
	latest_version := trim(input.latest_version, "v")

	semver.is_valid(current_version)
	semver.compare(current_version, latest_version) == -1
}
