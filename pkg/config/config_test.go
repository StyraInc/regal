package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/util/test"

	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/testutil"
)

const levelError = "error"

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
		// ignore is empty and so should not be marshalled
		Ignore: Ignore{
			Files: []string{},
		},
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
		t.Errorf("expected:\n%sgot:\n%s", expect, string(bs))
	}
}

func TestUnmarshalMarshalConfigWithDefaultRuleConfigs(t *testing.T) {
	t.Parallel()

	bs := []byte(`
rules:
  default:
    level: ignore
  bugs:
    default:
      level: error
    constant-condition:
      level: ignore
  testing:
    print-or-trace-call:
      level: error
`)

	var originalConfig Config

	if err := yaml.Unmarshal(bs, &originalConfig); err != nil {
		t.Fatal(err)
	}

	if originalConfig.Defaults.Global.Level != "ignore" {
		t.Errorf("expected global default to be level ignore")
	}

	if _, unexpected := originalConfig.Rules["bugs"]["default"]; unexpected {
		t.Errorf("erroneous rule parsed, bugs.default should not exist")
	}

	if originalConfig.Defaults.Categories["bugs"].Level != levelError {
		t.Errorf("expected bugs default to be level error")
	}

	if originalConfig.Rules["testing"]["print-or-trace-call"].Level != levelError {
		t.Errorf("expected for testing.print-or-trace-call to be level error")
	}

	originalConfig.Capabilities = nil

	marshalledConfigBs := testutil.Must(yaml.Marshal(originalConfig))(t)

	var roundTrippedConfig Config
	if err := yaml.Unmarshal(marshalledConfigBs, &roundTrippedConfig); err != nil {
		t.Fatal(err)
	}

	if roundTrippedConfig.Defaults.Global.Level != "ignore" {
		t.Errorf("expected global default to be level ignore")
	}

	if roundTrippedConfig.Defaults.Categories["bugs"].Level != levelError {
		t.Errorf("expected bugs default to be level error")
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
capabilities:
  from:
    engine: opa
    version: v0.45.0
  plus:
    builtins:
      - name: ldap.query
        type: function
        decl:
          args:
            - type: string
        result:
          type: object
  minus:
    builtins:
      - name: http.send
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

	if exp, got := 183, len(conf.Capabilities.Builtins); exp != got {
		t.Errorf("expected %d builtins, got %d", exp, got)
	}

	expectedBuiltins := []string{"regex.match", "ldap.query"}

	for _, expectedBuiltin := range expectedBuiltins {
		expectedBuiltinFound := false

		for name := range conf.Capabilities.Builtins {
			if name == expectedBuiltin {
				expectedBuiltinFound = true

				break
			}
		}

		if !expectedBuiltinFound {
			t.Errorf("expected builtin %s to be found", expectedBuiltin)
		}
	}

	unexpectedBuiltins := []string{"http.send"}

	for _, unexpectedBuiltin := range unexpectedBuiltins {
		unexpectedBuiltinFound := false

		for name := range conf.Capabilities.Builtins {
			if name == unexpectedBuiltin {
				unexpectedBuiltinFound = true

				break
			}
		}

		if unexpectedBuiltinFound {
			t.Errorf("expected builtin %s to be removed", unexpectedBuiltin)
		}
	}
}

func TestUnmarshalConfigWithBuiltinsFile(t *testing.T) {
	t.Parallel()

	bs := []byte(`rules: {}
capabilities:
  from:
    file: "./fixtures/caps.json"
`)

	var conf Config

	if err := yaml.Unmarshal(bs, &conf); err != nil {
		t.Fatal(err)
	}

	if exp, got := 1, len(conf.Capabilities.Builtins); exp != got {
		t.Errorf("expected %d builtins, got %d", exp, got)
	}

	expectedBuiltins := []string{"wow"}

	for _, expectedBuiltin := range expectedBuiltins {
		expectedBuiltinFound := false

		for name := range conf.Capabilities.Builtins {
			if name == expectedBuiltin {
				expectedBuiltinFound = true

				break
			}
		}

		if !expectedBuiltinFound {
			t.Errorf("expected builtin %s to be found", expectedBuiltin)
		}
	}
}

func TestUnmarshalConfigDefaultCapabilities(t *testing.T) {
	t.Parallel()

	bs := []byte(`rules: {}
`)

	var conf Config

	if err := yaml.Unmarshal(bs, &conf); err != nil {
		t.Fatal(err)
	}

	caps := ast.CapabilitiesForThisVersion()

	if exp, got := len(caps.Builtins), len(conf.Capabilities.Builtins); exp != got {
		t.Errorf("expected %d builtins, got %d", exp, got)
	}

	// choose the first built-ins to check for to keep the test fast
	expectedBuiltins := []string{caps.Builtins[0].Name, caps.Builtins[1].Name}

	for _, expectedBuiltin := range expectedBuiltins {
		expectedBuiltinFound := false

		for name := range conf.Capabilities.Builtins {
			if name == expectedBuiltin {
				expectedBuiltinFound = true

				break
			}
		}

		if !expectedBuiltinFound {
			t.Errorf("expected builtin %s to be found", expectedBuiltin)
		}
	}
}
