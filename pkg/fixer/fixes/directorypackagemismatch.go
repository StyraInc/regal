package fixes

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/pkg/config"
)

type DirectoryPackageMismatch struct {
	// DryRun indicates whether this fixer should perform the actual rename or not, and if true simply signals to the
	// client which file should be moved and to where. Whether to actually treat it as a "dry run" is up to the client.
	DryRun bool
}

func (*DirectoryPackageMismatch) Name() string {
	return "directory-package-mismatch"
}

// For now, just handle the "normal" set of characters, plus hyphens.
// We can broaden this later, but we should avoid characters that may have
// special meaning in files and directories.
var regularName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

// Fix moves a file to the correct directory based on the package name, and relative to the base
// directory provided in opts. If the file is already in the correct directory, no action is taken.
// If the DryRun field is set to true, the file is not moved by the fixer, but will rather have the
// old location and new location passed back to the client who will decide what to do with that.
func (d *DirectoryPackageMismatch) Fix(fc *FixCandidate, opts *RuntimeOptions) ([]FixResult, error) {
	pkgPath, err := getPackagePathDirectory(fc, opts.Config)
	if err != nil {
		return nil, err
	}

	rootPath := opts.BaseDir
	if rootPath == "" {
		rootPath = filepath.Dir(fc.Filename)
	}

	rootPath = filepath.Clean(rootPath)

	newPath := filepath.Join(rootPath, pkgPath, filepath.Base(fc.Filename))

	if newPath == fc.Filename {
		return nil, nil // File is where it should be. We are done!
	}

	if d.DryRun {
		newRel, err := filepath.Rel(rootPath, newPath)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate relative path: %w", err)
		}

		return []FixResult{{
			Title:    d.Name(),
			Root:     rootPath,
			Contents: fc.Contents,
			Rename: &Rename{
				FromPath: fc.Filename,
				ToPath:   filepath.Join(rootPath, newRel),
			},
		}}, nil
	}

	if err := recursiveMoveFile(fc.Filename, newPath); err != nil {
		return nil, fmt.Errorf("unexpected error when moving file %w", err)
	}

	if err := cleanOldPath(rootPath, fc.Filename); err != nil {
		return nil, fmt.Errorf("failed to clean up directory made empty by moving file %w", err)
	}

	return []FixResult{{
		Title:    d.Name(),
		Root:     rootPath,
		Contents: fc.Contents,
		Rename: &Rename{
			FromPath: fc.Filename,
			ToPath:   newPath, // TODO: should we check that this is relative to base somewhere?
		},
	}}, nil
}

func getPackagePathDirectory(fc *FixCandidate, config *config.Config) (string, error) {
	module, err := ast.ParseModule(fc.Filename, string(fc.Contents))
	if err != nil {
		return "", err // nolint:wrapcheck
	}

	parts := make([]string, len(module.Package.Path)-1)
	excludeTestSuffix := shouldExcludeTestSuffix(config)
	pathWithoutData := module.Package.Path[1:]

	for i, part := range pathWithoutData {
		text := strings.Trim(part.Value.String(), "\"")

		if !regularName.MatchString(text) {
			return "", fmt.Errorf("can only handle [a-zA-Z0-9_-] characters in package name, got: %s", text)
		}

		if i == len(pathWithoutData)-1 && excludeTestSuffix {
			text = strings.TrimSuffix(text, "_test")
		}

		parts[i] = text
	}

	return filepath.Join(parts...), nil
}

// nolint:wrapcheck
func recursiveMoveFile(oldPath, newPath string) error {
	_, err := os.Stat(filepath.Dir(newPath))
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	err = os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}

	return nil
}

// nolint:wrapcheck
func cleanOldPath(rootPath, oldPath string) error {
	// Remove empty directories
	dir := filepath.Dir(oldPath)

	f, err := os.Open(dir)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Readdirnames(1)
	if errors.Is(err, io.EOF) { // dir is empty
		if err := os.Remove(dir); err != nil {
			return err
		}

		// traverse one level up in the directory tree to see
		// if we've left more empty directories behind
		up, _ := filepath.Split(dir)

		if up != rootPath {
			return cleanOldPath(rootPath, up)
		}
	} else if err != nil {
		return err
	}

	return nil
}

func shouldExcludeTestSuffix(config *config.Config) bool {
	if category, ok := config.Rules["idiomatic"]; ok {
		if rule, ok := category["directory-package-mismatch"]; ok {
			if exclude, ok := rule.Extra["exclude-test-suffix"].(bool); ok {
				return exclude
			}
		}
	}

	// this is the default, and this should be unreachable provided that the
	// provided configuration was included (which it always would be)
	return true
}
