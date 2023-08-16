# METADATA
# description: |
#   Fetch the builtin metadata from the OPA repo, and filter out anything but
#   the args and return value objects of each bultin function. We'll make this
#   data available in Regal policies under `data.opa.builtins`
package build.metadata

import future.keywords.if
import future.keywords.in

response := http.send({
	"method": "get",
	"url": "https://raw.githubusercontent.com/open-policy-agent/opa/main/builtin_metadata.json",
	"force_json_decode": true,
	"raise_error": false,
})

something_wrong if {
	response.status_code != 200
	print("error in response")
	print(response)
}

builtin_metadata[builtin] := object.filter(attributes, ["args", "result"]) if {
	not something_wrong
	some builtin, attributes in object.remove(response.body, ["_categories"])
}
