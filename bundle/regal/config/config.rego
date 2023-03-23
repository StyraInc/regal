package regal.config

default user_config := {}

user_config := data.regal_user_config

merged_config := object.union(data.regal.config.provided, user_config)

for_rule(metadata) := merged_config.rules[metadata.custom.category][metadata.title]
