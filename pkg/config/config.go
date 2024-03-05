package config

import (
	"encoding/json"
	"errors"
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
	Capabilities *Capabilities       `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`

	// Defaults state is loaded from configuration under rules and so is not (un)marshalled
	// in the same way.
	Defaults Defaults `json:"-"     yaml:"-"`
}

type Category map[string]Rule

// Defaults is used to store information about global and category
// defaults for rules.
type Defaults struct {
	Global     Default
	Categories map[string]Default
}

// Default represents global or category settings for rules,
// currently only the level is supported.
type Default struct {
	Level string `json:"level" yaml:"level"`
}

type Ignore struct {
	Files []string `json:"files,omitempty" yaml:"files,omitempty"`
}

type ExtraAttributes map[string]any

type Rule struct {
	Level  string
	Ignore *Ignore `json:"ignore,omitempty" yaml:"ignore,omitempty"`
	Extra  ExtraAttributes
}

type Capabilities struct {
	Builtins       map[string]*Builtin `json:"builtins"        yaml:"builtins"`
	FutureKeywords []string            `json:"future_keywords" yaml:"future_keywords"`
	Features       []string            `json:"features"        yaml:"features"`
}

type Decl struct {
	Args   []string `json:"args"   yaml:"args"`
	Result string   `json:"result" yaml:"result"`
}

type Builtin struct {
	Decl Decl `json:"decl" yaml:"decl"`
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

func (config Config) MarshalYAML() (any, error) {

	var unstructuredConfig map[string]any
	err := rio.JSONRoundTrip(config, &unstructuredConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to created unstructured config: %w", err)
	}

	fmt.Println(unstructuredConfig)

	// place the global defaults at the top level under rules
	if config.Defaults.Global.Level != "" {
		unstructuredConfig["rules"].(map[string]any)["default"] = config.Defaults.Global
	}

	// place the category defaults under the respective category
	for categoryName, categoryDefault := range config.Defaults.Categories {
		unstructuredConfig["rules"].(map[string]any)[categoryName].(map[string]any)["default"] =
			categoryDefault
	}

	if _, ok := unstructuredConfig["ignore"]; ok {
		delete(unstructuredConfig, "ignore")
	}

	return unstructuredConfig, nil
}

func (config *Config) UnmarshalYAML(value *yaml.Node) error {
	var result struct {
		Rules        map[string]any `yaml:"rules"`
		Ignore       Ignore         `yaml:"ignore"`
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

	// in order to support wildcard 'default' configs, we
	// have some hooks in this unmarshalling process to load these.
	categoryMap := make(map[string]Category)
	config.Defaults.Categories = make(map[string]Default)
	for key, val := range result.Rules {
		if key == "default" {
			var defaults Default
			err := rio.JSONRoundTrip(val, &defaults)
			if err != nil {
				return fmt.Errorf("unmarshalling defaults failed: %w", err)
			}

			config.Defaults.Global = defaults
			continue
		}

		// default configs are also supported within
		ruleMap := make(map[string]Rule)
		for ruleName, ruleData := range val.(map[string]any) {
			if ruleName == "default" {
				var defaults Default
				err := rio.JSONRoundTrip(ruleData, &defaults)
				if err != nil {
					return fmt.Errorf("unmarshalling defaults failed: %w", err)
				}

				config.Defaults.Categories[key] = defaults
				continue
			}

			var rule Rule
			err := rio.JSONRoundTrip(ruleData, &rule)
			if err != nil {
				return fmt.Errorf("unmarshalling rule failed: %w", err)
			}
			ruleMap[ruleName] = rule
		}

		categoryMap[key] = ruleMap
	}

	config.Rules = categoryMap
	config.Ignore = result.Ignore

	capabilitiesFile := result.Capabilities.From.File
	capabilitiesEngine := result.Capabilities.From.Engine
	capabilitiesEngineVersion := result.Capabilities.From.Version

	if capabilitiesFile != "" && capabilitiesEngine != "" {
		return errors.New("capabilities from.file and from.engine are mutually exclusive")
	}

	if capabilitiesEngine != "" && capabilitiesEngineVersion == "" {
		return errors.New("please set the version for the engine from which to load capabilities from")
	}

	if capabilitiesFile != "" {
		bs, err := os.ReadFile(capabilitiesFile)
		if err != nil {
			return fmt.Errorf("failed to load capabilities file: %w", err)
		}

		opaCaps := ast.Capabilities{}

		err = json.Unmarshal(bs, &opaCaps)
		if err != nil {
			return fmt.Errorf("failed to unmarshal capabilities file contents: %w", err)
		}

		config.Capabilities = fromOPACapabilities(opaCaps)
	}

	if capabilitiesEngine != "" && result.Capabilities.From.Engine == capabilitiesEngineOPA {
		capabilities, err := ast.LoadCapabilitiesVersion(result.Capabilities.From.Version)
		if err != nil {
			return fmt.Errorf("loading capabilities failed: %w", err)
		}

		config.Capabilities = fromOPACapabilities(*capabilities)
	}

	// by default, use the capabilities from the current OPA
	if capabilitiesEngine == "" && capabilitiesFile == "" {
		config.Capabilities = CapabilitiesForThisVersion()
	}

	// remove any builtins referenced in the minus config
	for _, minusBuiltin := range result.Capabilities.Minus.Builtins {
		delete(config.Capabilities.Builtins, minusBuiltin.Name)
	}

	// add any builtins referenced in the plus config
	for _, plusBuiltin := range result.Capabilities.Plus.Builtins {
		config.Capabilities.Builtins[plusBuiltin.Name] = fromOPABuiltin(*plusBuiltin)
	}

	return nil
}

// CapabilitiesForThisVersion returns the capabilities for the current OPA version Regal depends on.
func CapabilitiesForThisVersion() *Capabilities {
	return fromOPACapabilities(*ast.CapabilitiesForThisVersion())
}

func fromOPABuiltin(builtin ast.Builtin) *Builtin {
	funcArgs := builtin.Decl.FuncArgs().Args
	args := make([]string, len(funcArgs))

	for i, arg := range funcArgs {
		args[i] = arg.String()
	}

	rb := &Builtin{Decl: Decl{Args: args}}

	if builtin.Decl != nil && builtin.Decl.Result() != nil {
		// internal.print has no result
		rb.Decl.Result = builtin.Decl.Result().String()
	}

	return rb
}

func fromOPACapabilities(capabilities ast.Capabilities) *Capabilities {
	var result Capabilities

	result.Builtins = make(map[string]*Builtin)

	for _, builtin := range capabilities.Builtins {
		result.Builtins[builtin.Name] = fromOPABuiltin(*builtin)
	}

	result.FutureKeywords = capabilities.FutureKeywords
	result.Features = capabilities.Features

	return &result
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
	result, err := rule.MarshalYAML()

	if err != nil {
		return nil, fmt.Errorf("marshalling rule failed %w", err)
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
