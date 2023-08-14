package regal.config

import future.keywords.if
import future.keywords.in

excluded_file(metadata, file) if {
	force_exclude_file(file)
} else if {
	rule_config := for_rule(metadata)
	ex := rule_config.ignore.files
	is_array(ex)
	some pattern in ex
	exclude(pattern, file)
} else := false

force_exclude_file(file) if {
	# regal ignore:external-reference
	some pattern in global_ignore_patterns
	exclude(pattern, file)
}

global_ignore_patterns := merged_config.ignore.files if {
	not data.eval.params.ignore_files
} else := data.eval.params.ignore_files

# exclude imitates Gits .gitignore pattern matching as best it can
# Ref: https://git-scm.com/docs/gitignore#_pattern_format
exclude(pattern, file) if {
	patterns := pattern_compiler(pattern)
	some p in patterns
	glob.match(p, ["/"], file)
} else := false

# pattern_compiler transforms a glob pattern into a set of glob patterns to make the
# combined set behave as Gits .gitignore
pattern_compiler(pattern) := ps1 if {
	p := internal_slashes(pattern)
	p1 := leading_slash(p)
	ps := leading_doublestar_pattern(p1)
	ps1 := {pat |
		some _p
		ps[_p]
		nps := trailing_slash(_p)
		nps[pat]
	}
}

# Internal slashes means that the path is relative to root,
# if not it can appear anywhere in the hierarchy
#
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

# If a pattern starts with a "/", the leading slash is ignored but according to
# the .gitignore rule of internal slashes, it is relative to root
leading_slash(pattern) := substring(pattern, 1, -1) if {
	startswith(pattern, "/")
} else := pattern
