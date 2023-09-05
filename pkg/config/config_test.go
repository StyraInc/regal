package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/util/test"

	rio "github.com/styrainc/regal/internal/io"
)

func TestFindRegalDirectory(t *testing.T) {
	t.Parallel()

	fs := map[string]string{"/foo/bar/baz/p.rego": ""}

	test.WithTempFS(fs, func(root string) {
		if err := os.Mkdir(filepath.Join(root, rio.PathSeparator, ".regal"), os.ModePerm); err != nil {
			t.Fatal(err)
		}

		path := filepath.Join(root, "/foo/bar/baz")

		_, err := FindRegalDirectory(path)
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
		_, err := FindRegalDirectory(path)
		if err == nil {
			t.Errorf("expected no config file to be found")
		}
	})
}

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

func TestMarshalConfig(t *testing.T) {
	t.Parallel()

	conf := Config{
		Rules: map[string]Category{
			"testing": {
				"foo": Rule{
					Level: "error",
					Ignore: &Ignore{
						Files: []string{"foo.rego"},
					},
					Extra: ExtraAttributes{
						"bar":    "baz",
						"ignore": "this should be removed by the marshaller",
					},
				},
			},
		},
	}

	bs, err := yaml.Marshal(conf)
	if err != nil {
		t.Fatal(err)
	}

	expect := `rules:
    testing:
        foo:
            bar: baz
            ignore:
                files:
                    - foo.rego
            level: error
`

	if string(bs) != expect {
		t.Errorf("expected %s, got %s", expect, string(bs))
	}
}
