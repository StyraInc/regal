package regal.capabilities

import data.regal.config

import rego.v1

default provided := {}

# METADATA
# description: |
#   The capabilities object for Regal itself. Use `config.capabilities`
#   to get the capabilities for the target environment (i.e. the policies
#   getting linted).
# scope: document
provided := data.internal.capabilities

has_object_keys if "object.keys" in object.keys(config.capabilities.builtins)

has_strings_count if "strings.count" in object.keys(config.capabilities.builtins)

# if if if!
has_if if "if" in config.capabilities.future_keywords

has_if if has_rego_v1_feature

has_contains if "contains" in config.capabilities.future_keywords

has_contains if has_rego_v1_feature

has_rego_v1_feature if "rego_v1_import" in config.capabilities.features
