package config

import (
	"encoding/json"
	"fmt"

	"github.com/styrainc/regal/internal/util"
)

type Config map[string]Category

type Category map[string]Rule

type ExtraAttributes map[string]any

type Rule struct {
	Enabled bool
	Extra   ExtraAttributes
}

func DefaultConfig(data map[string]any) (Config, error) {
	rulesRaw, err := util.SearchMap(data, []string{"data", "regal", "config", "rules"})
	if err != nil {
		return nil, err
	}

	rulesNode, ok := rulesRaw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected 'rules' of object type")
	}

	config := Config{}

	for cat, attributes := range rulesNode {
		category := make(Category)
		rules, ok := attributes.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected structure of category %v", cat)
		}
		for name, confRaw := range rules {
			rule := Rule{}
			conf, ok := confRaw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("unexpected structure of rule %v in category %v", name, cat)
			}
			enabled, ok := conf["enabled"].(bool)
			if !ok {
				return nil, fmt.Errorf("unexpected type for attribute 'enabled' in rule %v: must be boolean", name)
			}
			delete(conf, "enabled")

			rule.Enabled = enabled
			rule.Extra = conf

			category[name] = rule
		}
		config[cat] = category
	}

	return config, nil
}

func (rule Rule) MarshalJSON() ([]byte, error) {
	var result = make(map[string]any)
	result["enabled"] = rule.Enabled

	for key, val := range rule.Extra {
		result[key] = val
	}

	return json.Marshal(&result)
}

func (rule *Rule) UnmarshalJSON(data []byte) error {
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	enabled, ok := result["enabled"].(bool)
	if !ok {
		return fmt.Errorf("value of 'enabled' must be boolean")
	}
	delete(result, "enabled")

	rule.Enabled = enabled
	rule.Extra = result

	return nil
}
