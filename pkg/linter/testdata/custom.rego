package custom.regal.rules.naming

import future.keywords.contains
import future.keywords.if

import data.regal

# METADATA
# title: acme-corp-package
# description: All packages must use "acme.corp" base name
# related_resources:
# - description: documentation
#   ref: https://www.acmecorp.example.org/docs/regal/package
# custom:
#   category: naming
report contains violation if {
	not acme_corp_package
	not system_log_package

	violation := regal.fail(rego.metadata.rule(), {})
}

acme_corp_package if {
	input["package"].path[1].value == "acme"
	input["package"].path[2].value == "corp"
}

system_log_package if {
	input["package"].path[1].value == "system"
	input["package"].path[2].value == "log"
}
