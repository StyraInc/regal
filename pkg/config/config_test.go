package config

import (
	"path/filepath"
	"testing"

	"github.com/open-policy-agent/opa/util/test"
)

func TestFindConfig(t *testing.T) {
	t.Parallel()

	fs := map[string]string{
		"/foo/bar/baz/p.rego":         "",
		"/foo/bar/.regal/config.yaml": "",
	}

	test.WithTempFS(fs, func(root string) {
		path := filepath.Join(root, "/foo/bar/baz")
		_, err := FindConfig(path)
		if err != nil {
			t.Error(err)
		}
	})

	fs = map[string]string{
		"/foo/bar/baz/p.rego": "",
		"/foo/bar/bax.json":   "",
	}

	test.WithTempFS(fs, func(root string) {
		path := filepath.Join(root, "/foo/bar/baz")
		_, err := FindConfig(path)
		if err == nil {
			t.Errorf("expected no config file to be found")
		}
	})
}
