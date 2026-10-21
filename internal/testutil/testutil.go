package testutil

import (
	"encoding/json/jsontext"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/open-policy-agent/regal/internal/encoding/roast"
	"github.com/open-policy-agent/regal/internal/test/must"
)

func MustBeOK[T any](x T, ok bool) func(testing.TB) T {
	return func(tb testing.TB) T {
		tb.Helper()

		if !ok {
			tb.Fatal("expected ok to be true, got false")
		}

		return x
	}
}

func ErrMustContain(err error, substr string) func(testing.TB) {
	return func(tb testing.TB) {
		tb.Helper()

		if err == nil {
			tb.Fatal("expected error got nil")
		} else if !strings.Contains(err.Error(), substr) {
			tb.Fatalf("expected error to contain %q, got %q", substr, err.Error())
		}
	}
}

func TempDirectoryOf(tb testing.TB, files map[string]string) string {
	tb.Helper()

	tmpDir := tb.TempDir()

	for file, contents := range files {
		path := filepath.Join(tmpDir, file)

		must.MkdirAll(tb, filepath.Dir(path))
		must.WriteFile(tb, path, []byte(contents))
	}

	return tmpDir
}

func MustUnmarshalYAML[T any](tb testing.TB, data []byte) T {
	tb.Helper()

	var result T
	if err := yaml.Unmarshal(data, &result); err != nil {
		tb.Fatalf("failed to unmarshal YAML: %v", err)
	}

	return result
}

func ToJSONRawMessage(tb testing.TB, msg any) *jsontext.Value {
	tb.Helper()

	return new(jsontext.Value(must.Marshal(tb, msg, roast.Options)))
}
