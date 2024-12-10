package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"

	"github.com/open-policy-agent/opa/v1/bundle"

	rio "github.com/styrainc/regal/internal/io"
)

func FilterIgnoredPaths(paths, ignore []string, checkFileExists bool, pathPrefix string) ([]string, error) {
	// - special case for stdin, return as is
	if len(paths) == 1 && paths[0] == "-" {
		return paths, nil
	}

	// if set, pathPrefix is normalized to end with a platform appropriate separator
	if pathPrefix != "" && !strings.HasSuffix(pathPrefix, rio.PathSeparator) {
		pathPrefix += rio.PathSeparator
	}

	if checkFileExists {
		filtered := make([]string, 0, len(paths))

		if err := walkPaths(paths, func(path string, info os.DirEntry, err error) error {
			if rio.IsSkipWalkDirectory(info) {
				return filepath.SkipDir
			}

			if !info.IsDir() && strings.HasSuffix(path, bundle.RegoExt) {
				filtered = append(filtered, path)
			}

			return err
		}); err != nil {
			return nil, fmt.Errorf("failed to filter paths:\n%w", err)
		}

		return filterPaths(filtered, ignore, pathPrefix)
	}

	if len(ignore) == 0 {
		return paths, nil
	}

	return filterPaths(paths, ignore, pathPrefix)
}

func walkPaths(paths []string, filter func(path string, info os.DirEntry, err error) error) error {
	var errs error

	for _, path := range paths {
		// We need to stat the initial set of paths, as filepath.WalkDir
		// will panic on non-existent paths.
		if _, err := os.Stat(path); err != nil {
			errs = errors.Join(errs, err)

			continue
		}

		if err := filepath.WalkDir(path, filter); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func filterPaths(policyPaths []string, ignore []string, pathPrefix string) ([]string, error) {
	filtered := make([]string, 0, len(policyPaths))

outer:
	for _, f := range policyPaths {
		for _, pattern := range ignore {
			if pattern == "" {
				continue
			}

			excluded, err := excludeFile(pattern, f, pathPrefix)
			if err != nil {
				return nil, fmt.Errorf("failed to check for exclusion using pattern %s: %w", pattern, err)
			}

			if excluded {
				continue outer
			}
		}

		filtered = append(filtered, f)
	}

	return filtered, nil
}

// excludeFile imitates the pattern matching of .gitignore files
// See `exclusion.rego` for details on the implementation.
func excludeFile(pattern, filename, pathPrefix string) (bool, error) {
	n := len(pattern)

	if pathPrefix != "" {
		filename = strings.TrimPrefix(filename, pathPrefix)
	}

	// Internal slashes means path is relative to root, otherwise it can
	// appear anywhere in the directory (--> **/)
	if !strings.Contains(pattern[:n-1], "/") {
		pattern = "**/" + pattern
	}

	// Leading slash?
	pattern = strings.TrimPrefix(pattern, "/")

	// Leading double-star?
	var ps []string
	if strings.HasPrefix(pattern, "**/") {
		ps = []string{pattern, strings.TrimPrefix(pattern, "**/")}
	} else {
		ps = []string{pattern}
	}

	var ps1 []string

	// trailing slash?
	for _, p := range ps {
		switch {
		case strings.HasSuffix(p, "/"):
			ps1 = append(ps1, p+"**")
		case !strings.HasSuffix(p, "/") && !strings.HasSuffix(p, "**"):
			ps1 = append(ps1, p, p+"/**")
		default:
			ps1 = append(ps1, p)
		}
	}

	// Loop through patterns and return true on first match
	for _, p := range ps1 {
		g, err := glob.Compile(p, '/')
		if err != nil {
			return false, fmt.Errorf("failed to compile pattern %s: %w", p, err)
		}

		if g.Match(filename) {
			return true, nil
		}
	}

	return false, nil
}
