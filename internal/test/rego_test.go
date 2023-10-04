package test

import (
	"context"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/open-policy-agent/opa/tester"

	"github.com/styrainc/regal/internal/compile"
	"github.com/styrainc/regal/internal/testutil"
	"github.com/styrainc/regal/pkg/builtins"
)

func TestRunRegoUnitTests(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	bdl := filepath.Join("..", "..", "bundle")

	bundle := testutil.Must(tester.LoadBundles([]string{bdl}, func(abspath string, info fs.FileInfo, depth int) bool {
		return false
	}))(t)

	store := inmem.NewWithOpts(inmem.OptRoundTripOnWrite(false))
	txn := testutil.Must(store.NewTransaction(ctx, storage.WriteParams))(t)

	t.Cleanup(func() {
		store.Abort(ctx, txn)
	})

	compiler := compile.NewCompilerWithRegalBuiltins().
		WithSchemas(compile.RegalSchemaSet()).
		WithUseTypeCheckAnnotations(true).
		WithEnablePrintStatements(true)

	runner := tester.NewRunner().
		SetCompiler(compiler).
		SetStore(store).
		SetRuntime(ast.NewTerm(ast.NewObject())).
		SetBundles(bundle).
		// TODO: Not needed?
		AddCustomBuiltins(builtins.TestContextBuiltins())

	ch := testutil.Must(runner.RunTests(ctx, txn))(t)

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
