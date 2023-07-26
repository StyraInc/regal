//nolint:gochecknoglobals
package test

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"

	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/pkg/rules"
)

const regalBundleDir = "../../bundle"

var once sync.Once

var regalBundle *bundle.Bundle

// GetRegalBundle allows reusing the same Regal rule bundle in tests
// without having to reload it from disk each time (i.e. a singleton)
// Note that tests making use of this must *not* make any modifications
// to the contents of the bundle.
func GetRegalBundle(t *testing.T) bundle.Bundle {
	t.Helper()

	once.Do(func() {
		absRegalBundleDir, err := filepath.Abs(regalBundleDir)
		if err != nil {
			t.Fatal(err)
		}
		rb := rio.MustLoadRegalBundlePath(absRegalBundleDir)
		regalBundle = &rb
	})

	return *regalBundle
}

func InputPolicy(filename string, policy string) rules.Input {
	content := map[string]string{filename: policy}
	modules := map[string]*ast.Module{filename: parse.MustParseModule(policy)}

	return rules.NewInput(content, modules)
}

// InputBundle allows for constructing a policy bundle
// for testing rules that perform linting of an entire policy bundle.
// policyBundle represents the bundle as a map from filename to filecontent.
func InputBundle(policyBundle map[string]string) rules.Input {
	content := make(map[string]string)
	modules := make(map[string]*ast.Module)

	for filename, filecontent := range policyBundle {
		content[filename] = filecontent
		modules[filename] = parse.MustParseModule(filecontent)
	}

	return rules.NewInput(content, modules)
}
