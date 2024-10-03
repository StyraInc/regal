package lsp

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/parse"
)

func TestEvalWorkspacePath(t *testing.T) {
	t.Parallel()

	ls := NewLanguageServer(context.Background(), &LanguageServerOptions{ErrorLog: os.Stderr})

	policy1 := `package policy1

	import rego.v1

	import data.policy2

	default allow := false

	allow if policy2.allow
	`

	policy2 := `package policy2

	import rego.v1

	allow if input.exists
	`

	module1 := parse.MustParseModule(policy1)
	module2 := parse.MustParseModule(policy2)

	ls.cache.SetFileContents("file://policy1.rego", policy1)
	ls.cache.SetFileContents("file://policy2.rego", policy2)
	ls.cache.SetModule("file://policy1.rego", module1)
	ls.cache.SetModule("file://policy2.rego", module2)

	input := strings.NewReader(`{"exists": true}`)

	res, err := ls.EvalWorkspacePath(context.TODO(), "data.policy1.allow", input)
	if err != nil {
		t.Fatal(err)
	}

	if val, ok := res.Value.(bool); !ok || val != true {
		t.Fatalf("expected true, got false")
	}
}

func TestFindInput(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	workspacePath := tmpDir + "/workspace"
	file := tmpDir + "/workspace/foo/bar/baz.rego"

	if err := os.MkdirAll(workspacePath+"/foo/bar", 0o755); err != nil {
		t.Fatal(err)
	}

	if readInputString(t, file, workspacePath) != "" {
		t.Fatalf("did not expect to find input.json")
	}

	content := `{"x": 1}`

	createWithContent(t, tmpDir+"/workspace/foo/bar/input.json", content)

	if res := readInputString(t, file, workspacePath); res != content {
		t.Errorf("expected input at %s, got %s", content, res)
	}

	if err := os.Remove(tmpDir + "/workspace/foo/bar/input.json"); err != nil {
		t.Fatal(err)
	}

	createWithContent(t, tmpDir+"/workspace/input.json", content)

	if res := readInputString(t, file, workspacePath); res != content {
		t.Errorf("expected input at %s, got %s", content, res)
	}
}

func createWithContent(t *testing.T, path string, content string) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	if _, err = f.WriteString(content); err != nil {
		t.Fatal(err)
	}
}

func readInputString(t *testing.T, file, workspacePath string) string {
	t.Helper()

	_, input := rio.FindInput(file, workspacePath)

	if input == nil {
		return ""
	}

	bs, err := io.ReadAll(input)
	if err != nil {
		t.Fatal(err)
	}

	if bs == nil {
		return ""
	}

	return string(bs)
}
