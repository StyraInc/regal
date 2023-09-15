# METADATA
# description: Collect data in aggregates and validate it
package custom.regal.rules.testcase["aggregates"]

import future.keywords
import data.regal.result

aggregate contains entry if {
    entry := { "file" : input.regal.file.name }
}

report contains violation if {
	not two_files_processed
    violation := result.fail(rego.metadata.chain(), {})
}

two_files_processed {
	files := [x | x = input.aggregate[_].file]
	count(files) == 2
}
