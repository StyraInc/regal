package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	rio "github.com/styrainc/regal/internal/io"
)

type Config struct {
	Rules  map[string]Category `json:"rules"            yaml:"rules"`
	Ignore Ignore              `json:"ignore,omitempty" yaml:"ignore,omitempty"`
}

type Category map[string]Rule

type Ignore struct {
	Files []string `json:"files,omitempty" yaml:"files,omitempty"`
}

type ExtraAttributes map[string]any

type Rule struct {
	Level  string
	Ignore Ignore `json:"ignore,omitempty" yaml:"ignore,omitempty"`
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

	for {
		searchPath := filepath.Join(rio.PathSeparator, dir, regalDirName)
		regalDir, err := os.Open(searchPath)

		if err == nil {
			rdInfo, err := regalDir.Stat()
			if err == nil && rdInfo.IsDir() {
				return regalDir, nil
			}
		}

		if searchPath == rio.PathSeparator+regalDirName {
			// Stop traversing at the root path
			return nil, fmt.Errorf("can't traverse past root directory %w", err)
		}

		// Move up one level in the directory tree
		parts := strings.Split(dir, rio.PathSeparator)
		parts = parts[:len(parts)-1]
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
	var confMap map[string]any

	rio.MustJSONRoundTrip(config, &confMap)

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

var errLevelMustBeString = errors.New("value of 'level' must be string")

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

	if rule.Ignore.Files != nil {
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

func (rule *Rule) mapToConfig(result map[string]any) error {
	level, ok := result["level"].(string)
	if !ok {
		return errLevelMustBeString
	}

	if ignore, ok := result["ignore"]; ok {
		var dst Ignore

		err := rio.JSONRoundTrip(ignore, &dst)
		if err != nil {
			return fmt.Errorf("unmarshalling rule ignore failed: %w", err)
		}

		rule.Ignore = dst
	}

	delete(result, "level")

	rule.Level = level
	rule.Extra = result

	return nil
}
