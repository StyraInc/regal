package providers

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/topdown"

	rbundle "github.com/styrainc/regal/bundle"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/lsp/cache"
	rego2 "github.com/styrainc/regal/internal/lsp/rego"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/uri"
	"github.com/styrainc/regal/pkg/builtins"
)

// Policy provides suggestions that have been determined by Rego policy.
type Policy struct {
	pq rego.PreparedEvalQuery
}

//nolint:gochecknoglobals
var regalRules = func() bundle.Bundle {
	regalRules := rio.MustLoadRegalBundleFS(rbundle.Bundle)

	return regalRules
}()

// NewPolicy creates a new Policy provider. This provider is distinctly different from the other providers
// as it acts like the entrypoint for all Rego-based providers, and not a single provider "function" like
// the Go providers do.
func NewPolicy(store storage.Store) *Policy {
	pq, err := prepareQuery(store, "completions := data.regal.lsp.completion.items")
	if err != nil {
		panic(fmt.Sprintf("failed preparing query for static bundle: %v", err))
	}

	return &Policy{
		pq: *pq,
	}
}

func (*Policy) Name() string {
	return "policy"
}

func (p *Policy) Run(c *cache.Cache, params types.CompletionParams, opts *Options) ([]types.CompletionItem, error) {
	if opts == nil {
		return nil, errors.New("options must be provided")
	}

	content, ok := c.GetFileContents(params.TextDocument.URI)
	if !ok {
		return nil, fmt.Errorf("could not get file contents for: %s", params.TextDocument.URI)
	}

	location := rego2.LocationFromPosition(params.Position)
	inputContext := make(map[string]any)
	inputContext["location"] = map[string]any{
		"row": location.Row,
		"col": location.Col,
	}
	inputContext["client_identifier"] = opts.ClientIdentifier
	inputContext["workspace_root"] = uri.ToPath(opts.ClientIdentifier, opts.RootURI)
	inputContext["path_separator"] = string(os.PathSeparator)

	input, err := rego2.ToInput(
		params.TextDocument.URI,
		opts.ClientIdentifier,
		content,
		inputContext,
	)
	if err != nil {
		// parser error could be due to work in progress, so just return an empty list here
		return []types.CompletionItem{}, nil //nolint: nilerr
	}

	result, err := rego2.QueryRegalBundle(input, p.pq)
	if err != nil {
		return nil, fmt.Errorf("failed querying regal bundle: %w", err)
	}

	completions := make([]types.CompletionItem, 8)

	err = rio.JSONRoundTrip(result["completions"], &completions)
	if err != nil {
		return nil, fmt.Errorf("failed converting completions: %w", err)
	}

	return completions, nil
}

func prepareQuery(store storage.Store, query string) (*rego.PreparedEvalQuery, error) {
	regoArgs := prepareRegoArgs(store, ast.MustParseBody(query))

	txn, err := store.NewTransaction(context.TODO(), storage.WriteParams)
	if err != nil {
		return nil, fmt.Errorf("failed creating transaction: %w", err)
	}

	regoArgs = append(regoArgs, rego.Transaction(txn))

	// Note that we currently don't provide metrics or profiling here, and
	// most likely we should — need to consider how to best make that conditional
	// and how to present it if enabled.
	pq, err := rego.New(regoArgs...).PrepareForEval(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed preparing query: %w", err)
	}

	err = store.Commit(context.Background(), txn)
	if err != nil {
		return nil, fmt.Errorf("failed committing transaction: %w", err)
	}

	return &pq, nil
}

func prepareRegoArgs(store storage.Store, query ast.Body) []func(*rego.Rego) {
	return []func(*rego.Rego){
		rego.Store(store),
		rego.ParsedQuery(query),
		rego.ParsedBundle("regal", &regalRules),
		rego.Function2(builtins.RegalParseModuleMeta, builtins.RegalParseModule),
		rego.Function1(builtins.RegalLastMeta, builtins.RegalLast),
		// TODO: remove later
		rego.EnablePrintStatements(true),
		rego.PrintHook(topdown.NewPrintHook(os.Stderr)),
	}
}
