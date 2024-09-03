package fixes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/styrainc/regal/pkg/config"
)

func TestFixDirectoryPackageMismatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name              string
		files             map[string]string
		expected          map[string]string
		wantErr           string
		includeTestSuffix bool // inverse of exclude-text-suffix to avoid providing the default everywhere
	}{
		{
			name: "files are moved according to their package",
			files: map[string]string{
				"main.rego":    "package main",
				"foo/bar.rego": "package bar",
				"foo/baz.rego": "package foo.baz",
			},
			expected: map[string]string{
				"main/main.rego":   "package main",
				"bar/bar.rego":     "package bar",
				"foo/baz/baz.rego": "package foo.baz",
			},
		},
		{
			name: "empty directories are cleaned up",
			files: map[string]string{
				"foo/bar.rego": "package bar",
			},
			expected: map[string]string{
				"bar/bar.rego": "package bar",
			},
		},
		{
			name: "non-empty directories are not cleaned up",
			files: map[string]string{
				"foo/bar.rego":  "package bar",
				"foo/README.md": "This is docs!",
			},
			expected: map[string]string{
				"bar/bar.rego":  "package bar",
				"foo/README.md": "This is docs!",
			},
		},
		{
			name: "nested directories are created correctly",
			files: map[string]string{
				"baz.rego": "package foo.bar.baz",
				"qux.rego": "package foo.bar",
			},
			expected: map[string]string{
				"foo/bar/baz/baz.rego": "package foo.bar.baz",
				"foo/bar/qux.rego":     "package foo.bar",
			},
		},
		{
			name: "nested directories are cleaned up correctly",
			files: map[string]string{
				"foo/bar/qux/yoo/foo.rego": "package foo",
				"foo/bar/foo.bar.rego":     "package foo.bar",
			},
			expected: map[string]string{
				"foo/foo.rego":         "package foo",
				"foo/bar/foo.bar.rego": "package foo.bar",
			},
		},
		{
			name: "package name with hyphens creates directories with hyphens",
			files: map[string]string{
				"foo.rego": `package foo["bar-baz"].qux`,
			},
			expected: map[string]string{
				"foo/bar-baz/qux/foo.rego": `package foo["bar-baz"].qux`,
			},
		},
		{
			name: "package name with special characters returns an error",
			files: map[string]string{
				"foo.rego": `package foo["bar/baz"].qux`,
			},
			expected: map[string]string{
				"foo.rego": `package foo["bar/baz"].qux`,
			},
			wantErr: "can only handle [a-zA-Z0-9_-] characters in package name, got: bar/baz",
		},
		{
			name: "package names with _test suffix are by default treated as if without the suffix",
			files: map[string]string{
				"foo_test.rego": "package foo_test",
				"foo.rego":      "package foo",
			},
			expected: map[string]string{
				"foo/foo_test.rego": "package foo_test",
				"foo/foo.rego":      "package foo",
			},
		},
		{
			name: "package names with _test suffix are accounted for when config exclude-test-suffix is false",
			files: map[string]string{
				"foo_test.rego": "package foo_test",
				"foo.rego":      "package foo",
			},
			expected: map[string]string{
				"foo_test/foo_test.rego": "package foo_test",
				"foo/foo.rego":           "package foo",
			},
			includeTestSuffix: true,
		},
		{
			name: "nested package names with _test suffix are accounted for when config exclude-test-suffix is false",
			files: map[string]string{
				"foo_test.rego":    "package foo.bar.baz_test",
				"foo.rego":         "package foo.bar.baz",
				"qux.rego":         "package foo[\"bar-baz\"].qux",
				"bar/bar/bar.rego": "package foo[\"bar-baz\"].qux_test",
			},
			expected: map[string]string{
				"foo/bar/baz_test/foo_test.rego": "package foo.bar.baz_test",
				"foo/bar/baz/foo.rego":           "package foo.bar.baz",
				"foo/bar-baz/qux/qux.rego":       "package foo[\"bar-baz\"].qux",
				"foo/bar-baz/qux_test/bar.rego":  "package foo[\"bar-baz\"].qux_test",
			},
			includeTestSuffix: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			baseDir := filepath.Join(tmpDir, "bundle")

			writeFiles(t, baseDir, tc.files)

			dpm := DirectoryPackageMismatch{}

			for file, contents := range tc.files {
				if !strings.HasSuffix(file, ".rego") {
					continue
				}

				_, err := dpm.Fix(&FixCandidate{
					Filename: filepath.Join(baseDir, file),
					Contents: []byte(contents),
				}, &RuntimeOptions{
					BaseDir: baseDir,
					Config:  configWithExcludeTestSuffix(!tc.includeTestSuffix),
				})
				if err != nil {
					if tc.wantErr == "" {
						t.Errorf("failed to fix %s: %v", file, err)
					}

					if strings.Contains(err.Error(), tc.wantErr) {
						continue
					} else {
						t.Errorf("expected error to contain %q, got %q", tc.wantErr, err.Error())
					}
				}
			}

			result := readFiles(t, baseDir)

			if len(result) != len(tc.expected) {
				t.Errorf("expected %d files, got %d", len(tc.expected), len(result))
			}

			for file, contents := range tc.expected {
				if result[file] != contents {
					t.Errorf("expected %s to be %s, got %s", file, contents, result[file])
				}
			}
		})
	}
}

func writeFiles(t *testing.T, path string, files map[string]string) {
	t.Helper()

	for file, contents := range files {
		filePath := filepath.Join(path, file)

		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}

		err := os.WriteFile(filePath, []byte(contents), 0o600)
		if err != nil {
			t.Fatalf("failed to write file %s: %v", filePath, err)
		}
	}
}

// recursively traverse dir entry and create a map[string]string of all files
// where the key is the full path, and the value is the file contents.
func readFiles(t *testing.T, baseDir string) map[string]string {
	t.Helper()

	files := make(map[string]string)

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			t.Fatalf("failed to walk path %s: %v", path, err)
		}

		if info.IsDir() {
			return nil
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read file %s: %v", path, err)
		}

		relPath, _ := filepath.Rel(baseDir, path)
		files[relPath] = string(contents)

		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk path %s: %v", baseDir, err)
	}

	return files
}

func configWithExcludeTestSuffix(exclude bool) *config.Config {
	return &config.Config{
		Rules: map[string]config.Category{
			"idiomatic": {
				"directory-package-mismatch": config.Rule{
					Level: "ignore",
					Extra: map[string]any{
						"exclude-test-suffix": exclude,
					},
				},
			},
		},
	}
}
