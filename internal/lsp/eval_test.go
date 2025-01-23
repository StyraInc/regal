package lsp

import (
	"context"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/lsp/log"
	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/testutil"
)

func TestEvalWorkspacePath(t *testing.T) {
	t.Parallel()

	logger := newTestLogger(t)

	ls := NewLanguageServer(
		context.Background(),
		&LanguageServerOptions{LogWriter: logger, LogLevel: log.LevelDebug},
	)

	policy1 := `package policy1

	import data.policy2

	default allow := false

	allow if policy2.allow
	`

	policy2 := `package policy2

	allow if input.exists
	`

	module1 := parse.MustParseModule(policy1)
	module2 := parse.MustParseModule(policy2)

	ls.cache.SetFileContents("file://policy1.rego", policy1)
	ls.cache.SetFileContents("file://policy2.rego", policy2)
	ls.cache.SetModule("file://policy1.rego", module1)
	ls.cache.SetModule("file://policy2.rego", module2)

	input := map[string]any{
		"exists": true,
	}

	res, err := ls.EvalWorkspacePath(context.TODO(), "data.policy1.allow", input)
	if err != nil {
		t.Fatal(err)
	}

	if val, ok := res.Value.(bool); !ok || val != true {
		t.Fatalf("expected true, got false")
	}
}

func TestEvalWorkspacePathInternalData(t *testing.T) {
	t.Parallel()

	logger := newTestLogger(t)

	ls := NewLanguageServer(
		context.Background(),
		&LanguageServerOptions{LogWriter: logger, LogLevel: log.LevelDebug},
	)

	res, err := ls.EvalWorkspacePath(context.TODO(), "object.keys(data.internal)", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	val, ok := res.Value.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", res.Value)
	}

	valStr := make([]string, 0, len(val))

	for _, v := range val {
		str, ok := v.(string)
		if !ok {
			t.Fatalf("expected string, got %T", v)
		}

		valStr = append(valStr, str)
	}

	slices.Sort(valStr)

	exp := []string{"capabilities", "combined_config"}

	if !slices.Equal(valStr, exp) {
		t.Fatalf("expected %v, got %v", exp, valStr)
	}
}

func TestFindInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		fileType    string
		fileContent string
	}{
		{"json", `{"x": true}`},
		{"yaml", "x: true"},
	}

	for _, tc := range cases {
		t.Run(tc.fileType, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()

			workspacePath := filepath.Join(tmpDir, "workspace")
			file := filepath.Join(tmpDir, "workspace", "foo", "bar", "baz.rego")

			testutil.MustMkdirAll(t, workspacePath, "foo", "bar")

			path, content := rio.FindInput(file, workspacePath)
			if path != "" || content != nil {
				t.Fatalf("did not expect to find input.%s", tc.fileType)
			}

			createWithContent(t, tmpDir+"/workspace/foo/bar/input."+tc.fileType, tc.fileContent)

			path, content = rio.FindInput(file, workspacePath)
			if path != workspacePath+"/foo/bar/input."+tc.fileType || !maps.Equal(content, map[string]any{"x": true}) {
				t.Errorf(`expected input {"x": true} at, got %s`, content)
			}

			testutil.MustRemove(t, tmpDir+"/workspace/foo/bar/input."+tc.fileType)

			createWithContent(t, tmpDir+"/workspace/input."+tc.fileType, tc.fileContent)

			path, content = rio.FindInput(file, workspacePath)
			if path != workspacePath+"/input."+tc.fileType || !maps.Equal(content, map[string]any{"x": true}) {
				t.Errorf(`expected input {"x": true} at, got %s`, content)
			}
		})
	}
}

func createWithContent(t *testing.T, path string, content string) {
	t.Helper()

	f := testutil.Must(os.Create(path))(t)
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
}
