package regal.config

import future.keywords.if
import future.keywords.in

excluded_file(rule_config, file) if {
	force_exclude_file(file)
} else if {
	ex := rule_config.ignore
	is_array(ex)
	some pattern in ex
	exclude(pattern, file)
} else := false

force_exclude_file(file) if {
	# regal ignore:external-reference
	some pattern in global_ignore_patterns
	exclude(pattern, file)
}

global_ignore_patterns := merged_config.ignore if {
	not data.eval.params.ignore
} else := data.eval.params.ignore

# exclude imitates Gits .gitignore pattern matching as best it can
# Ref: https://git-scm.com/docs/gitignore#_pattern_format
exclude(pattern, file) if {
	patterns := pattern_compiler(pattern)
	some p in patterns
	glob.match(p, ["/"], file)
} else := false

# pattern_compiler transforms a glob pattern into a set of glob patterns to make the
# groups behave as Gits .gitignore
pattern_compiler(pattern) := ps1 if {
	p := internal_slashes(pattern)
	ps := leading_doublestar_pattern(p)
	ps1 := {pat | some _p; ps[_p]; nps := trailing_slash(_p); nps[pat]}
}

# myfiledir and mydir/ turns into **/myfiledir and **/mydir/
# mydir/p and mydir/d/ are returned as is
internal_slashes(pattern) := pattern if {
	s := substring(pattern, 0, count(pattern) - 1)
	contains(s, "/")
} else := concat("", ["**/", pattern])

# **/pattern might match my/dir/pattern and pattern
# So we branch it into itself and one with the leading **/ removed
leading_doublestar_pattern(pattern) := {pattern, p} if {
	startswith(pattern, "**/")
	p := substring(pattern, 3, -1)
} else := {pattern}

# If a pattern does not end with a "/", then it can both
# - match a folder => pattern + "/**"
# - match a file => pattern
trailing_slash(pattern) := {pattern, np} if {
	not endswith(pattern, "/")
	not endswith(pattern, "**")
	np := concat("", [pattern, "/**"])
} else := {np} if {
	endswith(pattern, "/")
	np := concat("", [pattern, "**"])
} else := {pattern}
