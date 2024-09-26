package bundles

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadDataBundle(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path         string
		files        map[string]string
		expectedData any
	}{
		"simple bundle": {
			path: "foo",
			files: map[string]string{
				"foo/.manifest": `{"roots":["foo"]}`,
				"foo/data.json": `{"foo": "bar"}`,
			},
			expectedData: map[string]any{
				"foo": "bar",
			},
		},
		"nested bundle": {
			path: "foo",
			files: map[string]string{
				"foo/.manifest":     `{"roots":["foo", "bar"]}`,
				"foo/data.yml":      `foo: bar`,
				"foo/bar/data.yaml": `bar: baz`,
			},
			expectedData: map[string]any{
				"foo": "bar",
				"bar": map[string]any{
					"bar": "baz",
				},
			},
		},
		"array data": {
			path: "foo",
			files: map[string]string{
				"foo/.manifest":     `{"roots":["bar"]}`,
				"foo/bar/data.json": `[{"foo": "bar"}]`,
			},
			expectedData: map[string]any{
				"bar": []any{
					map[string]any{
						"foo": "bar",
					},
				},
			},
		},
		"rego files": {
			path: "foo",
			files: map[string]string{
				"foo/.manifest":  `{"roots":["foo"]}`,
				"food/rego.rego": `package foo`,
			},
			expectedData: map[string]any{},
		},
	}

	for testCase, testData := range testCases {
		t.Run(testCase, func(t *testing.T) {
			t.Parallel()

			workspacePath := t.TempDir()

			// create the workspace state
			for file, contents := range testData.files {
				filePath := filepath.Join(workspacePath, file)

				dir := filepath.Dir(filePath)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatalf("failed to create directory %s: %v", dir, err)
				}

				if err := os.WriteFile(filePath, []byte(contents), 0o600); err != nil {
					t.Fatalf("failed to write file %s: %v", filePath, err)
				}
			}

			b, err := LoadDataBundle(filepath.Join(workspacePath, testData.path))
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(b.Data, testData.expectedData) {
				t.Fatalf("expected data to be %v, but got %v", testData.expectedData, b.Data)
			}

			if len(b.Modules) != 0 {
				t.Fatalf("expected no modules, but got %d", len(b.Modules))
			}
		})
	}
}
