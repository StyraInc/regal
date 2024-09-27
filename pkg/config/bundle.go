package config

import (
	"errors"
	"fmt"
	"strings"

	"dario.cat/mergo"

	"github.com/open-policy-agent/opa/bundle"

	"github.com/styrainc/regal/internal/util"
)

func LoadConfigWithDefaultsFromBundle(regalBundle *bundle.Bundle, userConfig *Config) (Config, error) {
	path := []string{"regal", "config", "provided"}

	bundled, err := util.SearchMap(regalBundle.Data, path)
	if err != nil {
		return Config{}, fmt.Errorf("config path not found %s: %w", strings.Join(path, "."), err)
	}

	bundledConf, ok := bundled.(map[string]any)
	if !ok {
		return Config{}, errors.New("expected 'rules' of object type")
	}

	defaultConfig, err := FromMap(bundledConf)
	if err != nil {
		return Config{}, fmt.Errorf("failed to convert config from map: %w", err)
	}

	if userConfig == nil {
		defaultConfig.Capabilities = CapabilitiesForThisVersion()

		return defaultConfig, nil
	}

	providedRuleLevels := providedConfLevels(&defaultConfig)

	if err = mergo.Merge(&defaultConfig, userConfig, mergo.WithOverride); err != nil {
		return Config{}, fmt.Errorf("failed to merge user config: %w", err)
	}

	if defaultConfig.Capabilities == nil {
		defaultConfig.Capabilities = CapabilitiesForThisVersion()
	}

	// adopt user rule levels based on config and defaults
	// If the user configuration contains rules with the level unset, copy the level from the provided configuration
	extractUserRuleLevels(userConfig, &defaultConfig, providedRuleLevels)

	return defaultConfig, nil
}

// extractUserRuleLevels uses defaulting config and per-rule levels from user configuration to set the level for each
// rule.
func extractUserRuleLevels(userConfig *Config, mergedConf *Config, providedRuleLevels map[string]string) {
	for categoryName, rulesByCategory := range mergedConf.Rules {
		for ruleName, rule := range rulesByCategory {
			var providedLevel string

			var ok bool

			if providedLevel, ok = providedRuleLevels[ruleName]; !ok {
				continue
			}

			// use the level from the provided configuration as the fallback
			selectedRuleLevel := providedLevel

			var userHasConfiguredRule bool

			if _, ok := userConfig.Rules[categoryName]; ok {
				_, userHasConfiguredRule = userConfig.Rules[categoryName][ruleName]
			}

			if userHasConfiguredRule && userConfig.Rules[categoryName][ruleName].Level != "" {
				// if the user config has a level for the rule, use that
				selectedRuleLevel = userConfig.Rules[categoryName][ruleName].Level
			} else if categoryDefault, ok := mergedConf.Defaults.Categories[categoryName]; ok {
				// if the config has a default level for the category, use that
				if categoryDefault.Level != "" {
					selectedRuleLevel = categoryDefault.Level
				}
			} else if mergedConf.Defaults.Global.Level != "" {
				// if the config has a global default level, use that
				selectedRuleLevel = mergedConf.Defaults.Global.Level
			}

			rule.Level = selectedRuleLevel
			mergedConf.Rules[categoryName][ruleName] = rule
		}
	}
}

// Copy the level of each rule from the provided configuration.
func providedConfLevels(conf *Config) map[string]string {
	ruleLevels := make(map[string]string)

	for categoryName, rulesByCategory := range conf.Rules {
		for ruleName := range rulesByCategory {
			// Note that this assumes all rules have unique names,
			// which we'll likely always want for provided rules
			ruleLevels[ruleName] = conf.Rules[categoryName][ruleName].Level
		}
	}

	return ruleLevels
}
