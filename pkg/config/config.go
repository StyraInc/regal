package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/ast"

	rio "github.com/styrainc/regal/internal/io"
)

const capabilitiesEngineOPA = "opa"

type Config struct {
	Rules        map[string]Category `json:"rules"                  yaml:"rules"`
	Ignore       Ignore              `json:"ignore,omitempty"       yaml:"ignore,omitempty"`
	Capabilities ast.Capabilities    `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
}

func (config *Config) UnmarshalYAML(value *yaml.Node) error {
	var result struct {
		Rules        map[string]Category `yaml:"rules"`
		Ignore       Ignore              `yaml:"ignore"`
		Capabilities struct {
			From struct {
				Engine  string `yaml:"engine"`
				Version string `yaml:"version"`
				File    string `yaml:"file"`
			} `yaml:"from"`
			Plus struct {
				Builtins []*ast.Builtin `yaml:"builtins"`
			} `yaml:"plus"`
			Minus struct {
				Builtins []struct {
					Name string `yaml:"name"`
				} `yaml:"builtins"`
			} `yaml:"minus"`
		} `yaml:"capabilities"`
	}

	if err := value.Decode(&result); err != nil {
		return fmt.Errorf("unmarshalling config failed %w", err)
	}

	config.Rules = result.Rules
	config.Ignore = result.Ignore

	capabilitiesFile := result.Capabilities.From.File
	capabilitiesEngine := result.Capabilities.From.Engine
	capabilitiesEngineVersion := result.Capabilities.From.Version

	if capabilitiesFile != "" && capabilitiesEngine != "" {
		return fmt.Errorf("capabilities from.file and from.engine are mutually exclusive")
	}

	if capabilitiesEngine != "" && capabilitiesEngineVersion == "" {
		return fmt.Errorf("please set the version for the engine from which to load capabilities from")
	}

	if capabilitiesFile != "" {
		bs, err := os.ReadFile(capabilitiesFile)
		if err != nil {
			return fmt.Errorf("failed to load capabilities file: %w", err)
		}

		err = json.Unmarshal(bs, &config.Capabilities)
		if err != nil {
			return fmt.Errorf("failed to unmarshal capabilities file contents: %w", err)
		}
	}

	if capabilitiesEngine != "" && result.Capabilities.From.Engine == capabilitiesEngineOPA {
		capabilities, err := ast.LoadCapabilitiesVersion(result.Capabilities.From.Version)
		if err != nil {
			return fmt.Errorf("loading capabilities failed: %w", err)
		}

		config.Capabilities = *capabilities
	}

	// by default, use the capabilities from the current OPA
	if capabilitiesEngine == "" && capabilitiesFile == "" {
		config.Capabilities = *ast.CapabilitiesForThisVersion()
	}

	// remove any builtins referenced in the minus config
	for i, builtin := range config.Capabilities.Builtins {
		for _, minusBuiltin := range result.Capabilities.Minus.Builtins {
			if minusBuiltin.Name == builtin.Name {
				config.Capabilities.Builtins = append(config.Capabilities.Builtins[:i], config.Capabilities.Builtins[i+1:]...)
			}
		}
	}

	config.Capabilities.Builtins = append(config.Capabilities.Builtins, result.Capabilities.Plus.Builtins...)

	return nil
}

type Category map[string]Rule

type Ignore struct {
	Files []string `json:"files,omitempty" yaml:"files,omitempty"`
}

type ExtraAttributes map[string]any

type Rule struct {
	Level  string
	Ignore *Ignore `json:"ignore,omitempty" yaml:"ignore,omitempty"`
	Extra  ExtraAttributes
}

const (
	regalDirName   = ".regal"
	configFileName = "config.yaml"
)

// FindRegalDirectory searches for a .regal directory first in the directory of path, and if not found,
// in the parent directory, and if not found, in the parent's parent directory, and so on.
func FindRegalDirectory(path string) (*os.File, error) {
	finfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %v: %w", path, err)
	}

	dir := path

	if !finfo.IsDir() {
		dir = filepath.Dir(path)
	}

	// returns e.g. "C:" on windows, "" on other platforms
	volume := filepath.VolumeName(dir)

	for {
		var searchPath string
		if volume == "" {
			searchPath = filepath.Join(rio.PathSeparator, dir, regalDirName)
		} else {
			searchPath = filepath.Join(dir, regalDirName)
		}

		regalDir, err := os.Open(searchPath)

		if err == nil {
			rdInfo, err := regalDir.Stat()
			if err == nil && rdInfo.IsDir() {
				return regalDir, nil
			}
		}

		if searchPath == volume+rio.PathSeparator+regalDirName {
			// Stop traversing at the root path
			return nil, fmt.Errorf("can't traverse past root directory %w", err)
		}

		// Move up one level in the directory tree
		parts := strings.Split(dir, rio.PathSeparator)
		parts = parts[:len(parts)-1]

		if parts[0] == volume {
			parts[0] = volume + rio.PathSeparator
		}

		dir = filepath.Join(parts...)
	}
}

func FindConfig(path string) (*os.File, error) {
	regalDir, err := FindRegalDirectory(path)
	if err != nil {
		return nil, fmt.Errorf("could not find .regal directory: %w", err)
	}

	return os.Open(filepath.Join(regalDir.Name(), rio.PathSeparator, configFileName)) //nolint:wrapcheck
}

func FromMap(confMap map[string]any) (Config, error) {
	var conf Config

	err := rio.JSONRoundTrip(confMap, &conf)
	if err != nil {
		return conf, fmt.Errorf("failed to convert config map to config struct: %w", err)
	}

	return conf, nil
}

func ToMap(config Config) map[string]any {
	confMap := make(map[string]any)

	rio.MustJSONRoundTrip(config, &confMap)

	// Not sure why `omitempty` doesn't do the trick here, but having `ignore: {}` in the config for each
	// rule is annoying noice when printed from Rego.
	for categoryName, category := range config.Rules {
		for ruleName, rule := range category {
			if rule.Ignore == nil {
				// casts should be safe here as the structure is copied from the config struct
				//nolint:forcetypeassert
				delete(confMap["rules"].(map[string]any)[categoryName].(map[string]any)[ruleName].(map[string]any), "ignore")
			}
		}
	}

	return confMap
}

func (rule Rule) MarshalJSON() ([]byte, error) {
	result := make(map[string]any)
	result["level"] = rule.Level
	result["ignore"] = rule.Ignore

	for key, val := range rule.Extra {
		result[key] = val
	}

	return json.Marshal(&result) //nolint:wrapcheck
}

func (rule *Rule) UnmarshalJSON(data []byte) error {
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("unmarshalling rule failed %w", err)
	}

	return rule.mapToConfig(result)
}

func (rule Rule) MarshalYAML() (interface{}, error) {
	result := make(map[string]any)
	result["level"] = rule.Level

	if rule.Ignore != nil && len(rule.Ignore.Files) != 0 {
		result["ignore"] = rule.Ignore
	}

	for key, val := range rule.Extra {
		if key != "ignore" && key != "level" {
			result[key] = val
		}
	}

	return result, nil
}

func (rule *Rule) UnmarshalYAML(value *yaml.Node) error {
	var result map[string]any
	if err := value.Decode(&result); err != nil {
		return fmt.Errorf("unmarshalling rule failed %w", err)
	}

	return rule.mapToConfig(result) //nolint:errcheck
}

// Note that this function will mutate the result map. This isn't a problem right now
// as we only use this after unmarshalling, but if we use this for other purposes later
// we need to make a copy of the map first.
func (rule *Rule) mapToConfig(result map[string]any) error {
	level, ok := result["level"].(string)
	if ok {
		rule.Level = level
	}

	if ignore, ok := result["ignore"]; ok {
		var dst Ignore

		err := rio.JSONRoundTrip(ignore, &dst)
		if err != nil {
			return fmt.Errorf("unmarshalling rule ignore failed: %w", err)
		}

		rule.Ignore = &dst
	}

	rule.Extra = result

	delete(rule.Extra, "level")
	delete(rule.Extra, "ignore")

	return nil
}
