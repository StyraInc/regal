package regal.config

# METADATA
# description: |
#   determines if file should be excluded, either because it's globally
#   ignored, or because the specific rule configuration excludes it
excluded_file(category, title, file) if {
	# regal ignore:external-reference
	some pattern in _global_ignore_patterns
	_exclude(pattern, file)
} else if {
	some pattern in for_rule(category, title).ignore.files
	_exclude(pattern, file)
}

_global_ignore_patterns := data.eval.params.ignore_files if {
	count(data.eval.params.ignore_files) > 0
} else := merged_config.ignore.files

# exclude imitates .gitignore pattern matching as best it can
# ref: https://git-scm.com/docs/gitignore#_pattern_format
_exclude(pattern, file) if {
	some p in _pattern_compiler(pattern)
	glob.match(p, ["/"], file)
}

# pattern_compiler transforms a glob pattern into a set of glob
# patterns to make the combined set behave as .gitignore
_pattern_compiler(pattern) := {pat |
	some p in _leading_doublestar_pattern(trim_prefix(_internal_slashes(pattern), "/"))
	some pat in _trailing_slash(p)
}

# Internal slashes means that the path is relative to root,
# if not it can appear anywhere in the hierarchy
#
# myfiledir and mydir/ turns into **/myfiledir and **/mydir/
# mydir/p and mydir/d/ are returned as is
_internal_slashes(pattern) := pattern if {
	s := substring(pattern, 0, count(pattern) - 1)
	contains(s, "/")
} else := concat("", ["**/", pattern])

# **/pattern might match my/dir/pattern and pattern
# So we branch it into itself and one with the leading **/ removed
_leading_doublestar_pattern(pattern) := {pattern, p} if {
	startswith(pattern, "**/")
	p := substring(pattern, 3, -1)
} else := {pattern}

# If a pattern does not end with a "/", then it can both
# - match a folder => pattern + "/**"
# - match a file => pattern
_trailing_slash(pattern) := {pattern, np} if {
	not endswith(pattern, "/")
	not endswith(pattern, "**")
	np := concat("", [pattern, "/**"])
} else := {np} if {
	endswith(pattern, "/")
	np := concat("", [pattern, "**"])
} else := {pattern}
