package config

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/util/test"

	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/testutil"
	"github.com/styrainc/regal/internal/util"
)

const levelError = "error"

func TestFindRegalDirectory(t *testing.T) {
	t.Parallel()

	fs := map[string]string{"/foo/bar/baz/p.rego": ""}

	test.WithTempFS(fs, func(root string) {
		if err := os.Mkdir(filepath.Join(root, rio.PathSeparator, ".regal"), os.ModePerm); err != nil {
			t.Fatal(err)
		}

		path := filepath.Join(root, "foo", "bar", "baz")

		if _, err := FindRegalDirectory(path); err != nil {
			t.Error(err)
		}
	})

	fs = map[string]string{
		"/foo/bar/baz/p.rego": "",
		"/foo/bar/bax.json":   "",
	}

	test.WithTempFS(fs, func(root string) {
		path := filepath.Join(root, "foo", "bar", "baz")

		_, err := FindRegalDirectory(path)
		if err == nil {
			t.Errorf("expected no config file to be found")
		}
	})
}

func TestFindConfig(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		FS           map[string]string
		Error        string
		ExpectedName string
	}{
		"no config file": {
			FS: map[string]string{
				"/foo/bar/baz/p.rego": "",
				"/foo/bar/bax.json":   "",
			},
			Error: "could not find Regal config",
		},
		".regal/config.yaml": {
			FS: map[string]string{
				"/foo/bar/baz/p.rego":         "",
				"/foo/bar/.regal/config.yaml": "",
			},
			ExpectedName: "/foo/bar/.regal/config.yaml",
		},
		".regal/ dir missing config file": {
			FS: map[string]string{
				"/foo/bar/baz/p.rego":   "",
				"/foo/bar/.regal/.keep": "", // .keep file to ensure the dir is present
			},
			Error: "config file was not found in .regal directory",
		},
		".regal.yaml": {
			FS: map[string]string{
				"/foo/bar/baz/p.rego":  "",
				"/foo/bar/.regal.yaml": "",
			},
			ExpectedName: "/foo/bar/.regal.yaml",
		},
		".regal.yaml and .regal/config.yaml": {
			FS: map[string]string{
				"/foo/bar/baz/p.rego":         "",
				"/foo/bar/.regal.yaml":        "",
				"/foo/bar/.regal/config.yaml": "",
			},
			Error: "conflicting config files: both .regal directory and .regal.yaml found",
		},
		".regal.yaml with .regal/config.yaml at higher directory": {
			FS: map[string]string{
				"/foo/bar/baz/p.rego":  "",
				"/foo/bar/.regal.yaml": "",
				"/.regal/config.yaml":  "",
			},
			ExpectedName: "/foo/bar/.regal.yaml",
		},
		".regal/config.yaml with .regal.yaml at higher directory": {
			FS: map[string]string{
				"/foo/bar/baz/p.rego":         "",
				"/foo/bar/.regal/config.yaml": "",
				"/.regal.yaml":                "",
			},
			ExpectedName: "/foo/bar/.regal/config.yaml",
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			test.WithTempFS(testData.FS, func(root string) {
				path := filepath.Join(root, "foo", "bar", "baz")

				configFile, err := FindConfig(path)
				if testData.Error != "" {
					if err == nil {
						t.Fatalf("expected error %s, got nil", testData.Error)
					}

					if !strings.Contains(err.Error(), testData.Error) {
						t.Fatalf("expected error %q, got %q", testData.Error, err.Error())
					}
				} else if err != nil {
					t.Fatalf("expected no error, got %s", err)
				}

				if testData.ExpectedName != "" {
					if got, exp := strings.TrimPrefix(configFile.Name(), root), testData.ExpectedName; got != exp {
						t.Fatalf("expected config file %q, got %q", exp, got)
					}
				}
			})
		})
	}
}

func TestFindBundleRootDirectories(t *testing.T) {
	t.Parallel()

	cfg := `
project:
  roots:
  - foo/bar
  - baz
`

	fs := map[string]string{
		"/.regal/config.yaml":       cfg, // root from config
		"/.regal/rules/policy.rego": "",  // custom rules directory
		"/bundle/.manifest":         "",  // bundle from .manifest
		"/foo/bar/baz/policy.rego":  "",  // foo/bar from config
		"/baz":                      "",  // baz from config
	}

	test.WithTempFS(fs, func(root string) {
		locations, err := FindBundleRootDirectories(root)
		if err != nil {
			t.Error(err)
		}

		if len(locations) != 5 {
			t.Errorf("expected 5 locations, got %d", len(locations))
		}

		expected := util.Map(util.FilepathJoiner(root), []string{"", ".regal/rules", "baz", "bundle", "foo/bar"})

		if !slices.Equal(expected, locations) {
			t.Errorf("expected\n%s\ngot\n%s", strings.Join(expected, "\n"), strings.Join(locations, "\n"))
		}
	})
}

func TestFindBundleRootDirectoriesWithStandaloneConfig(t *testing.T) {
	t.Parallel()

	cfg := `
project:
  roots:
  - foo/bar
  - baz
`

	fs := map[string]string{
		"/.regal.yaml":             cfg, // root from config
		"/bundle/.manifest":        "",  // bundle from .manifest
		"/foo/bar/baz/policy.rego": "",  // foo/bar from config
		"/baz":                     "",  // baz from config
	}

	test.WithTempFS(fs, func(root string) {
		locations, err := FindBundleRootDirectories(root)
		if err != nil {
			t.Error(err)
		}

		if len(locations) != 4 {
			t.Errorf("expected 5 locations, got %d", len(locations))
		}

		expected := util.Map(util.FilepathJoiner(root), []string{"", "baz", "bundle", "foo/bar"})

		if !slices.Equal(expected, locations) {
			t.Errorf("expected\n%s\ngot\n%s", strings.Join(expected, "\n"), strings.Join(locations, "\n"))
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

func TestUnmarshalConfigWithNumericOPAVersion(t *testing.T) {
	t.Parallel()

	bs := []byte(`
capabilities:
  from:
    engine: opa
    version: 68
`)
	if err := yaml.Unmarshal(bs, &Config{}); err == nil ||
		err.Error() != "capabilities: from.version must be a string" {
		t.Errorf("expected error, got %v", err)
	}
}

func TestUnmarshalConfigWithMissingVPrefixOPAVersion(t *testing.T) {
	t.Parallel()

	bs := []byte(`
capabilities:
  from:
    engine: opa
    version: 0.68.0
`)
	if err := yaml.Unmarshal(bs, &Config{}); err == nil ||
		err.Error() != "capabilities: from.version must be a valid OPA version (with a 'v' prefix)" {
		t.Errorf("expected error, got %v", err)
	}
}

func TestUnmarshalProjectRootsAsStringOrObject(t *testing.T) {
	t.Parallel()

	bs := []byte(`project:
  roots:
    - foo/bar
    - baz
    - path: bar/baz
    - path: v1
      rego-version: 1
`)

	var conf Config

	if err := yaml.Unmarshal(bs, &conf); err != nil {
		t.Fatal(err)
	}

	version1 := 1

	expRoots := []Root{
		{Path: "foo/bar"},
		{Path: "baz"},
		{Path: "bar/baz"},
		{Path: "v1", RegoVersion: &version1},
	}

	roots := *conf.Project.Roots

	if len(roots) != len(expRoots) {
		t.Errorf("expected %d roots, got %d", len(expRoots), len(roots))
	}

	for i, expRoot := range expRoots {
		if roots[i].Path != expRoot.Path {
			t.Errorf("expected root path %v, got %v", expRoot.Path, roots[i].Path)
		}

		if expRoot.RegoVersion != nil {
			if roots[i].RegoVersion == nil {
				t.Errorf("expected root %v to have a rego version", expRoot.Path)
			} else if *roots[i].RegoVersion != *expRoot.RegoVersion {
				t.Errorf(
					"expected root %v rego version %v, got %v",
					expRoot.Path, *expRoot.RegoVersion, *roots[i].RegoVersion,
				)
			}
		}
	}
}

func TestAllRegoVersions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		Config   string
		FS       map[string]string
		Expected map[string]ast.RegoVersion
	}{
		"values from config": {
			Config: `project:
  rego-version: 0
  roots:
    - path: foo
      rego-version: 1
`,
			FS: map[string]string{
				"bar/baz/.manifest": `{"rego_version": 1}`,
			},
			Expected: map[string]ast.RegoVersion{
				"":        ast.RegoV0,
				"bar/baz": ast.RegoV1,
				"foo":     ast.RegoV1,
			},
		},
		"no config": {
			Config: "",
			FS: map[string]string{
				"bar/baz/.manifest": `{"rego_version": 1}`,
			},
			Expected: map[string]ast.RegoVersion{},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			var conf *Config

			if testData.Config != "" {
				var loadedConf Config
				if err := yaml.Unmarshal([]byte(testData.Config), &loadedConf); err != nil {
					t.Fatal(err)
				}

				conf = &loadedConf
			}

			test.WithTempFS(testData.FS, func(root string) {
				versions, err := AllRegoVersions(root, conf)
				if err != nil {
					t.Fatal(err)
				}

				if !maps.Equal(versions, testData.Expected) {
					t.Errorf("expected %v, got %v", testData.Expected, versions)
				}
			})
		})
	}
}
