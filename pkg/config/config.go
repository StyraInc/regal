package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/anderseknert/roast/pkg/encoding"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/internal/capabilities"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/util"
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

	CapabilitiesURL string `json:"capabilities_url,omitempty" yaml:"capabilities_url,omitempty"`

	Project *Project `json:"project,omitempty" yaml:"project,omitempty"`
}

type Project struct {
	Roots []string `json:"roots" yaml:"roots"`
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

// FindBundleRootDirectories finds all bundle root directories from the provided path,
// which **must** be an absolute path. Bundle root directories may be found either by:
//
// - Configuration (`project.roots`)
// - By the presence of a .manifest file anywhere under the path
// - By the presence of a .regal directory anywhere under or above the path ... TODO (anders): might reconsider "above"?
//
// All returned paths are absolute paths. If the provided path itself
// is determined to be a bundle root directory it will be included in the result.
func FindBundleRootDirectories(path string) ([]string, error) {
	var foundBundleRoots []string

	// This will traverse the tree **upwards** searching for a .regal directory
	regalDir, err := FindRegalDirectory(path)
	if err == nil {
		roots, err := rootsFromRegalDirectory(regalDir)
		if err != nil {
			return nil, fmt.Errorf("failed to get roots from .regal directory: %w", err)
		}

		foundBundleRoots = append(foundBundleRoots, roots...)
	}

	// This will traverse the tree **downwards** searching for .regal directories
	// Not using rio.WalkFiles here as we're specifically looking for directories
	if err := filepath.WalkDir(path, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk path: %w", err)
		}

		if info.IsDir() && info.Name() == regalDirName {
			// Opening files as part of walking is generally not a good idea...
			// but I think we can assume the number of .regal directories in a project
			// is limited to a reasonable number.
			rd, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open .regal directory: %w", err)
			}

			defer rd.Close()

			roots, err := rootsFromRegalDirectory(rd)
			if err != nil {
				return fmt.Errorf("failed to get roots from .regal directory: %w", err)
			}

			foundBundleRoots = append(foundBundleRoots, roots...)
		}

		// rather than calling rio.FindManifestLocations later, let's
		// check for .manifest directories as part of the same walk
		if !info.IsDir() && info.Name() == ".manifest" {
			foundBundleRoots = append(foundBundleRoots, filepath.Dir(path))
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk path: %w", err)
	}

	slices.Sort(foundBundleRoots)

	return slices.Compact(foundBundleRoots), nil
}

func rootsFromRegalDirectory(regalDir *os.File) ([]string, error) {
	foundBundleRoots := make([]string, 0)

	defer regalDir.Close()

	parent, _ := filepath.Split(regalDir.Name())

	parent = filepath.Clean(parent)

	// add the parent directory of .regal
	foundBundleRoots = append(foundBundleRoots, parent)

	file, err := os.ReadFile(filepath.Join(regalDir.Name(), "config.yaml"))
	if err == nil {
		var conf Config

		err = yaml.Unmarshal(file, &conf)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
		}

		if conf.Project != nil {
			foundBundleRoots = append(foundBundleRoots, util.Map(util.FilepathJoiner(parent), conf.Project.Roots)...)
		}
	}

	return foundBundleRoots, nil
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
			URL     string `yaml:"url"`
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
	Project *Project `yaml:"project"`
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
	capabilitiesURL := result.Capabilities.From.URL

	// Capabilities can be specified by an engine+version combo, a local
	// file path, or a URL. These cannot be mixed and matched.
	if capabilitiesURL != "" && capabilitiesFile != "" {
		return errors.New("capabilities from.url and from.file are mutually exclusive")
	}

	if capabilitiesURL != "" && capabilitiesEngine != "" {
		return errors.New("capabilities from.url and from.engine are mutually exclusive")
	}

	if capabilitiesURL != "" && capabilitiesEngineVersion != "" {
		return errors.New("capabilities from.url and from.version are mutually exclusive")
	}

	if capabilitiesFile != "" && capabilitiesEngine != "" {
		return errors.New("capabilities from.file and from.engine are mutually exclusive")
	}

	if capabilitiesEngine != "" && capabilitiesEngineVersion == "" {
		// Although regal:///capabilities/{engine} is valid and refers
		// to the latest version for that engine, we'll keep the
		// existing (pre-capabilities.Lookup()) behavior in place and
		// disallow that when using the engine key.
		return errors.New("please set the version for the engine from which to load capabilities from")
	}

	if capabilitiesEngine != "" {
		capabilitiesURL = "regal:///capabilities/" + capabilitiesEngine + "/" + capabilitiesEngineVersion
	}

	if capabilitiesFile != "" {
		absfp, err := filepath.Abs(capabilitiesFile)
		if err != nil {
			return fmt.Errorf(
				"unable to load capabilities from '%s', failed to determine absolute path: %w",
				capabilitiesFile,
				err,
			)
		}

		capabilitiesURL = "file://" + absfp
	}

	if capabilitiesEngine == "" && capabilitiesFile == "" && capabilitiesURL == "" {
		capabilitiesURL = "regal:///capabilities/default"
	}

	opaCaps, err := capabilities.Lookup(context.Background(), capabilitiesURL)
	if err != nil {
		return fmt.Errorf("failed to load capabilities: %w", err)
	}

	config.Capabilities = fromOPACapabilities(*opaCaps)

	// This is used in the LSP to load the OPA capabilities, since the
	// capabilities version in the user-facing config does not contain all
	// of the information that the LSP needs.
	config.CapabilitiesURL = capabilitiesURL

	config.Project = result.Project

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

func GetPotentialRoots(paths ...string) ([]string, error) {
	var err error

	dirMap := make(map[string]struct{})

	absDirPaths := make([]string, len(paths))

	for i, path := range paths {
		abs := path

		if !filepath.IsAbs(abs) {
			abs, err = filepath.Abs(path)
			if err != nil {
				return nil, fmt.Errorf("failed to get absolute path for %s: %w", path, err)
			}
		}

		if isDir(abs) {
			absDirPaths[i] = abs
		} else {
			absDirPaths[i] = filepath.Dir(abs)
		}
	}

	for _, dir := range absDirPaths {
		brds, err := FindBundleRootDirectories(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to find bundle root directories in %s: %w", dir, err)
		}

		for _, brd := range brds {
			dirMap[brd] = struct{}{}
		}
	}

	foundRoots := util.Keys(dirMap)
	if len(foundRoots) == 0 {
		return absDirPaths, nil
	}

	return foundRoots, nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}
