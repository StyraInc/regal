package test

import (
	"context"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/open-policy-agent/opa/tester"

	"github.com/styrainc/regal/pkg/builtins"
)

func TestRunRegoUnitTests(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	bdl := filepath.Join("..", "..", "bundle")

	bundle, err := tester.LoadBundles([]string{bdl}, func(abspath string, info fs.FileInfo, depth int) bool {
		return false
	})
	if err != nil {
		t.Fatal(err)
	}

	store := inmem.NewWithOpts(inmem.OptRoundTripOnWrite(false))

	txn, err := store.NewTransaction(ctx, storage.WriteParams)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		store.Abort(ctx, txn)
	})

	runner := tester.NewRunner().
		SetStore(store).
		CapturePrintOutput(true).
		SetRuntime(ast.NewTerm(ast.NewObject())).
		SetBundles(bundle).
		AddCustomBuiltins(customBuiltins())

	ch, err := runner.RunTests(ctx, txn)
	if err != nil {
		t.Fatal(err)
	}

	for r := range ch {
		rc := r
		t.Run(rc.Name, func(t *testing.T) {
			t.Parallel()
			if rc.Fail {
				t.Errorf("%v\n%v", string(rc.Output), rc.Location)
			}
		})
	}
}

func customBuiltins() []*tester.Builtin {
	return []*tester.Builtin{
		{
			Decl: &ast.Builtin{
				Name: builtins.RegalParseModuleMeta.Name,
				Decl: builtins.RegalParseModuleMeta.Decl,
			},
			Func: rego.Function2(builtins.RegalParseModuleMeta, builtins.RegalParseModule),
		},
	}
}
