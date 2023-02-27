package regal.main

import future.keywords.contains
import future.keywords.if
import future.keywords.in

report contains violation if {
	violation := data.regal.rules[_].violation[_]
}
