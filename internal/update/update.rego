package update

import rego.v1

current_version := trim(input.current_version, "v")

latest_version := trim(input.latest_version, "v")

default needs_update := false

needs_update if {
	semver.is_valid(current_version)
	semver.compare(current_version, latest_version) == -1
}
