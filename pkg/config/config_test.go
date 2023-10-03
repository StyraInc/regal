package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/util/test"

	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/testutil"
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

	bs := testutil.Must(yaml.Marshal(conf))(t)

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

func TestUnmarshalConfig(t *testing.T) {
	t.Parallel()

	bs := []byte(`rules:
  testing:
    foo:
      bar: baz
      ignore:
        files:
          - foo.rego
      level: error
`)

	var conf Config

	if err := yaml.Unmarshal(bs, &conf); err != nil {
		t.Fatal(err)
	}

	if conf.Rules["testing"]["foo"].Level != "error" {
		t.Errorf("expected level to be error")
	}

	if conf.Rules["testing"]["foo"].Ignore == nil {
		t.Errorf("expected ignore attribute to be set")
	}

	if len(conf.Rules["testing"]["foo"].Ignore.Files) != 1 {
		t.Errorf("expected ignore files to be set")
	}

	if conf.Rules["testing"]["foo"].Ignore.Files[0] != "foo.rego" {
		t.Errorf("expected ignore files to contain foo.rego")
	}

	if conf.Rules["testing"]["foo"].Extra["bar"] != "baz" {
		t.Errorf("expected extra attribute 'bar' to be baz")
	}

	if conf.Rules["testing"]["foo"].Extra["ignore"] != nil {
		t.Errorf("expected extra attribute 'ignore' to be removed")
	}

	if conf.Rules["testing"]["foo"].Extra["level"] != nil {
		t.Errorf("expected extra attribute 'level' to be removed")
	}
}
