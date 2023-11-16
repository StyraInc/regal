package regal.util

import rego.v1

is_snake_case(str) if str == lower(str)
