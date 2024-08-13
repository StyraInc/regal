package bundles

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// FindInWorkspace finds all bundle roots and loads any data files under that
// root in the workspace.
func FindInWorkspace(workspacePath string) (map[string]any, error) {
	// bundleRoots is a list of all the .manifest files, relative to the
	// workspace
	bundleRoots := []string{}
	// dataFiles are collected during the walk of the workspace and loaded
	// into the relevant bundle later
	dataFiles := []string{}

	// Ensure that the workspace path always ends with a separator for consistency
	if !strings.HasSuffix(workspacePath, string(filepath.Separator)) {
		workspacePath += string(filepath.Separator)
	}

	err := filepath.WalkDir(workspacePath, func(path string, d os.DirEntry, err error) error {
		// These directories often have thousands of items we don't care about,
		// so don't even traverse them.
		if d.IsDir() && (d.Name() == ".git" || d.Name() == ".idea") {
			return filepath.SkipDir
		}

		// we need to traverse, but don't care about dirs themselves
		if d.IsDir() {
			return nil
		}

		// finding a .manifest file means we're at the root of a bundle
		if d.Name() == ".manifest" {
			relPath := strings.TrimPrefix(filepath.Dir(path), workspacePath)
			bundleRoots = append(bundleRoots, relPath)
		}

		// finding a data.* file means we have a data file to load into a bundle
		if d.Name() == "data.json" || d.Name() == "data.yaml" || d.Name() == "data.yml" {
			dataFiles = append(dataFiles, strings.TrimPrefix(path, workspacePath))
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk workspace: %w", err)
	}

	// bundleData will contain a map of bundle roots to all their data
	bundleData := make(map[string]any)

	for _, root := range bundleRoots {
		// bundleMap will contain all the data for this bundle root
		bundleMap := make(map[string]any)
		for _, dataFile := range dataFiles {
			if !strings.HasPrefix(dataFile, root) {
				continue
			}

			dir := filepath.Dir(strings.TrimPrefix(dataFile, root+string(filepath.Separator)))

			// if the data is not in the root of the bundle, then we need to
			// make a path for it in the current bundleMap
			if dir != "." {
				for _, part := range strings.Split(dir, string(filepath.Separator)) {
					if part == "" {
						continue
					}
					if _, exists := bundleMap[part]; !exists {
						bundleMap[part] = make(map[string]any)
					}
				}
			}

			filePath := filepath.Join(workspacePath, dataFile)
			fileContents, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
			}

			var parsedData any
			switch filepath.Ext(dataFile) {
			case ".json":
				if err := json.Unmarshal(fileContents, &parsedData); err != nil {
					return nil, fmt.Errorf("failed to parse JSON file %s: %w", filePath, err)
				}
			case ".yaml", ".yml":
				if err := yaml.Unmarshal(fileContents, &parsedData); err != nil {
					return nil, fmt.Errorf("failed to parse YAML file %s: %w", filePath, err)
				}
			}

			if filepath.Base(dir) == "." {
				// we cannot load multiple files into the root of a bundle
				if len(bundleMap) > 0 {
					return nil, fmt.Errorf("can't load multiple files into bundle root %s", root)
				}

				rootData, err := convert(parsedData)
				if err != nil {
					return nil, fmt.Errorf("failed to load structured data from file %s: %w", filePath, err)
				}

				switch v := rootData.(type) {
				case map[string]any:
					bundleMap = v
				case []any:
					bundleData[root] = v
				default:
					return nil, fmt.Errorf("unsupported type %T", v)
				}
			} else {
				// TODO: make this set recursively
				bundleMap[filepath.Base(dir)], err = convert(parsedData)
				if err != nil {
					return nil, fmt.Errorf("failed to load structured data from file %s: %w", filePath, err)
				}
			}
		}

		bundleData[root] = bundleMap
	}

	return bundleData, nil
}

// convert exists primarily to convert map[interface{}]interface to
// map[string]interface, but also passing through []any and map[string]any
func convert(input interface{}) (interface{}, error) {
	switch v := input.(type) {
	case map[interface{}]interface{}:
		output := make(map[string]interface{})
		for key, value := range v {
			strKey, ok := key.(string)
			if !ok {
				return nil, errors.New("all keys in the map must be strings")
			}
			output[strKey] = value
		}
		return output, nil
	case []any, map[string]any:
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported type %T", v)
	}
}
