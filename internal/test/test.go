//nolint:gochecknoglobals
package test

import (
	"sync"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"

	"github.com/styrainc/regal/internal/embeds"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/pkg/rules"
)

var once sync.Once

var regalBundle *bundle.Bundle

// GetRegalBundle allows reusing the same Regal rule bundle in tests
// without having to reload and compile it each time (i.e. a singleton)
// Note that tests making use of this must *not* make any modifications
// to the contents of the bundle.
func GetRegalBundle(t *testing.T) bundle.Bundle {
	t.Helper()

	once.Do(func() {
		rb := rio.MustLoadRegalBundleFS(embeds.EmbedBundleFS)
		regalBundle = &rb
	})

	return *regalBundle
}

func InputPolicy(filename string, policy string) rules.Input {
	content := map[string]string{filename: policy}
	modules := map[string]*ast.Module{filename: parse.MustParseModule(policy)}

	return rules.NewInput(content, modules)
}
