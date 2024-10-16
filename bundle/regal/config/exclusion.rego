package regal.config

import rego.v1

# METADATA
# description: |
#   determines if file should be excluded, either because it's globally
#   ignored, or because the specific rule configuration excludes it
excluded_file(category, title, file) if {
	# regal ignore:external-reference
	some pattern in _global_ignore_patterns
	_exclude(pattern, _relative_to_pattern(pattern, file))
} else if {
	some pattern in for_rule(category, title).ignore.files
	_exclude(pattern, _relative_to_pattern(pattern, file))
}

# NOTE
# this is an awful hack which is needed when the language server
# invokes linting, as it currently will provide filenames in the
# form of URIs rather than as relative paths... and as we do not
# know the base/workspace path here, we can only try to make the
# path relative to the *pattern* providedm which is something...
#
# pattern: foo/**/bar.rego
# file:    file://my/workspace/foo/baz/bar.rego
# returns: foo/baz/bar.rego
#
_relative_to_pattern(pattern, file) := relative if {
	startswith(file, "file://")

	absolute := trim_suffix(trim_prefix(file, "file://"), "/")
	file_parts := indexof_n(absolute, "/")

	relative := substring(
		absolute,
		array.slice(
			file_parts, (count(file_parts) - strings.count(pattern, "/")) - 1,
			count(file_parts),
		)[0] + 1,
		-1,
	)
} else := file

_global_ignore_patterns := data.eval.params.ignore_files

_global_ignore_patterns := merged_config.ignore.files if not data.eval.params.ignore_files

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
