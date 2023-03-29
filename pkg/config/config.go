package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	rio "github.com/styrainc/regal/internal/io"
)

type Config struct {
	Rules map[string]Category `json:"rules"`
}

type Category map[string]Rule

type ExtraAttributes map[string]any

type Rule struct {
	Enabled bool
	Extra   ExtraAttributes
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

func (rule Rule) MarshalJSON() ([]byte, error) {
	result := make(map[string]any)
	result["enabled"] = rule.Enabled

	for key, val := range rule.Extra {
		result[key] = val
	}

	return json.Marshal(&result) //nolint:wrapcheck
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
