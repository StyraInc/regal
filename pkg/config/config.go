package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anderseknert/roast/pkg/encoding"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/ast"

	eopa_caps "github.com/styrainc/enterprise-opa/capabilities"

	rio "github.com/styrainc/regal/internal/io"
)

const (
	capabilitiesEngineOPA  = "opa"
	capabilitiesEngineEOPA = "eopa"
	keyIgnore              = "ignore"
	keyLevel               = "level"
)

type Config struct {
	Rules        map[string]Category `json:"rules"                  yaml:"rules"`
	Ignore       Ignore              `json:"ignore,omitempty"       yaml:"ignore,omitempty"`
	Capabilities *Capabilities       `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`

	// Defaults state is loaded from configuration under rules and so is not (un)marshalled
	// in the same way.
	Defaults Defaults `json:"-" yaml:"-"`

	Features *Features `json:"features,omitempty" yaml:"features,omitempty"`
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

type Features struct {
	Remote *RemoteFeatures `json:"remote,omitempty" yaml:"remote,omitempty"`
}

type RemoteFeatures struct {
	CheckVersion bool `json:"check-version,omitempty" yaml:"check-version,omitempty"`
}

func (d *Default) mapToConfig(result any) error {
	resultMap, ok := result.(map[string]any)
	if !ok {
		return errors.New("result was not a map")
	}

	level, ok := resultMap[keyLevel].(string)
	if ok {
		d.Level = level
	}

	return nil
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

		if len(parts) < 2 {
			return nil, errors.New("stopping as dir is root directory")
		}

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

	// place the global defaults at the top level under rules
	if config.Defaults.Global.Level != "" {
		r, ok := unstructuredConfig["rules"].(map[string]any)
		if !ok {
			return nil, errors.New("rules in config were not a map")
		}

		r["default"] = config.Defaults.Global
	}

	// place the category defaults under the respective category
	for categoryName, categoryDefault := range config.Defaults.Categories {
		rawRuleMap, ok := unstructuredConfig["rules"].(map[string]any)
		if !ok {
			return nil, errors.New("rules in config were not a map")
		}

		rawCategoryMap, ok := rawRuleMap[categoryName].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("category %s was not a map", categoryName)
		}

		rawCategoryMap["default"] = categoryDefault
	}

	if len(config.Ignore.Files) == 0 {
		delete(unstructuredConfig, keyIgnore)
	}

	return unstructuredConfig, nil
}

// unmarshallingIntermediary is used to contain config data in a format that is used during unmarshalling.
// The internally loaded config data layout differs from the user-defined YAML.
type marshallingIntermediary struct {
	// rules are unmarshalled as any since the defaulting needs to be extracted from here
	// and configured elsewhere in the struct.
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
	Features struct {
		RemoteFeatures struct {
			CheckVersion bool `yaml:"check_version"`
		} `yaml:"remote"`
	} `yaml:"features"`
}

func (config *Config) UnmarshalYAML(value *yaml.Node) error {
	var result marshallingIntermediary

	if err := value.Decode(&result); err != nil {
		return fmt.Errorf("unmarshalling config failed %w", err)
	}

	// this call will walk the rule config and load and defaults into the config
	err := extractDefaults(config, &result)
	if err != nil {
		return fmt.Errorf("extracting defaults failed: %w", err)
	}

	err = extractRules(config, &result)
	if err != nil {
		return fmt.Errorf("extracting rules failed: %w", err)
	}

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
		json := encoding.JSON()

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

	if capabilitiesEngine != "" && result.Capabilities.From.Engine == capabilitiesEngineEOPA {
		capabilities, err := eopa_caps.LoadCapabilitiesVersion(result.Capabilities.From.Version)
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

	// feature defaults
	if result.Features.RemoteFeatures.CheckVersion {
		config.Features = &Features{
			Remote: &RemoteFeatures{
				CheckVersion: true,
			},
		}
	}

	return nil
}

// extractRules is a helper to load rules from the raw config data.
func extractRules(config *Config, result *marshallingIntermediary) error {
	// in order to support wildcard 'default' configs, we
	// have some hooks in this unmarshalling process to load these.
	categoryMap := make(map[string]Category)

	for key, val := range result.Rules {
		if key == "default" {
			continue
		}

		rawRuleMap, ok := val.(map[string]any)
		if !ok {
			return fmt.Errorf("rules for category %s were not a map", key)
		}

		ruleMap := make(map[string]Rule)

		for ruleName, ruleData := range rawRuleMap {
			if ruleName == "default" {
				continue
			}

			var r Rule

			err := r.mapToConfig(ruleData)
			if err != nil {
				return fmt.Errorf("unmarshalling rule failed: %w", err)
			}

			ruleMap[ruleName] = r
		}

		categoryMap[key] = ruleMap
	}

	config.Rules = categoryMap

	return nil
}

// extractDefaults is a helper to load both global and category defaults from the raw config data.
func extractDefaults(c *Config, result *marshallingIntermediary) error {
	c.Defaults.Categories = make(map[string]Default)

	rawGlobalDefault, ok := result.Rules["default"]
	if ok {
		err := c.Defaults.Global.mapToConfig(rawGlobalDefault)
		if err != nil {
			return fmt.Errorf("unmarshalling global defaults failed: %w", err)
		}
	}

	for key, val := range result.Rules {
		rawRuleMap, ok := val.(map[string]any)
		if !ok {
			return fmt.Errorf("rules for category %s were not a map", key)
		}

		rawCategoryDefault, ok := rawRuleMap["default"]
		if ok {
			var categoryDefault Default

			err := categoryDefault.mapToConfig(rawCategoryDefault)
			if err != nil {
				return fmt.Errorf("unmarshalling category defaults failed: %w", err)
			}

			c.Defaults.Categories[key] = categoryDefault
		}
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

	json := encoding.JSON()

	return json.Marshal(&result) //nolint:wrapcheck
}

func (rule *Rule) UnmarshalJSON(data []byte) error {
	var result map[string]any

	json := encoding.JSON()

	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("unmarshalling rule failed %w", err)
	}

	return rule.mapToConfig(result)
}

func (rule Rule) MarshalYAML() (interface{}, error) {
	result := make(map[string]any)
	result[keyLevel] = rule.Level

	if rule.Ignore != nil && len(rule.Ignore.Files) != 0 {
		result[keyIgnore] = rule.Ignore
	}

	for key, val := range rule.Extra {
		if key != keyIgnore && key != keyLevel {
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
func (rule *Rule) mapToConfig(result any) error {
	ruleMap, ok := result.(map[string]any)
	if !ok {
		return errors.New("result was not a map")
	}

	level, ok := ruleMap[keyLevel].(string)
	if ok {
		rule.Level = level
	}

	if ignore, ok := ruleMap[keyIgnore]; ok {
		var dst Ignore

		err := rio.JSONRoundTrip(ignore, &dst)
		if err != nil {
			return fmt.Errorf("unmarshalling rule ignore failed: %w", err)
		}

		rule.Ignore = &dst
	}

	rule.Extra = ruleMap

	delete(rule.Extra, keyLevel)
	delete(rule.Extra, keyIgnore)

	return nil
}
