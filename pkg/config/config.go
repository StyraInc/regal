package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config map[string]Category

type Category map[string]Rule

type ExtraAttributes map[string]any

type Rule struct {
	Enabled bool
	Extra   ExtraAttributes
}

const (
	configFileRelLocation = ".regal/config.yaml"
	pathSeparator         = string(os.PathSeparator)
)

// FindConfig searches for .regal/config.yaml first in the directory of path,
// and if not found, in the parent directory, and if not found, in the parent's
// parent directory, and so on.
func FindConfig(path string) (*os.File, error) {
	finfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %v: %w", path, err)
	}

	dir := path

	if !finfo.IsDir() {
		dir = filepath.Dir(path)
	}

	for {
		searchPath := filepath.Join(pathSeparator, dir, configFileRelLocation)
		config, err := os.Open(searchPath)

		if err == nil {
			return config, nil
		}

		if searchPath == pathSeparator+configFileRelLocation {
			// Stop traversing at the root path
			return nil, fmt.Errorf("can't traverse past root directory %w", err)
		}

		// Move up one level in the directory tree
		parts := strings.Split(dir, pathSeparator)
		parts = parts[:len(parts)-1]
		dir = filepath.Join(parts...)
	}
}

func (rule Rule) MarshalJSON() ([]byte, error) {
	result := make(map[string]any)
	result["enabled"] = rule.Enabled

	for key, val := range rule.Extra {
		result[key] = val
	}

	//nolint:wrapcheck
	return json.Marshal(&result)
}

var errEnabledMustBeBoolean = errors.New("value of 'enabled' must be boolean")

func (rule *Rule) UnmarshalJSON(data []byte) error {
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("unmarshalling rule failed %w", err)
	}

	enabled, ok := result["enabled"].(bool)
	if !ok {
		return errEnabledMustBeBoolean
	}

	delete(result, "enabled")

	rule.Enabled = enabled
	rule.Extra = result

	return nil
}
