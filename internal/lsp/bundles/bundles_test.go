package bundles

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFindInWorkspace(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	files := map[string]string{
		"foo/bar/.manifest":         ``, // contents of manifests not used currently
		"foo/bar/baz/data.json":     `{"file1": true}`,
		"bar/foo/.manifest":         ``,
		"bar/foo/baz/data.yaml":     `file2: true`,
		"bar/foo/bar/bax/data.yaml": `file5: true`,
		"baz/.manifest":             ``,
		"baz/data.yml":              `file3: true`,
		"baz/foo/data.json":         `[{"file4": true}]`,
	}

	for file, contents := range files {
		dir := filepath.Join(tempDir, filepath.Dir(file))

		err := os.MkdirAll(dir, 0o777)
		if err != nil {
			t.Fatal(err)
		}

		err = os.WriteFile(filepath.Join(tempDir, file), []byte(contents), 0o777)
		if err != nil {
			t.Fatal(err)
		}
	}

	expectedBundles := map[string]any{
		"foo/bar": map[string]any{
			"baz": map[string]any{
				"file1": true,
			},
		},
		"bar/foo": map[string]any{
			"baz": map[string]any{
				"file2": true,
			},
		},
		"baz": map[string]any{
			"file3": true,
			"foo": []map[string]any{
				{
					"file4": true,
				},
			},
		},
	}

	expectedJSON, err := json.MarshalIndent(expectedBundles, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	bundles, err := FindInWorkspace(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	actualJSON, _ := json.MarshalIndent(bundles, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if exp, got := string(expectedJSON), string(actualJSON); exp != got {
		t.Fatalf("expected %s, got %s", exp, got)
	}
}
