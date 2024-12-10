package providers

import (
	"context"
	"errors"
	"fmt"

	"github.com/anderseknert/roast/pkg/encoding"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/storage"

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

// NewPolicy creates a new Policy provider. This provider is distinctly different from the other providers
// as it acts like the entrypoint for all Rego-based providers, and not a single provider "function" like
// the Go providers do.
func NewPolicy(ctx context.Context, store storage.Store) *Policy {
	pq, err := prepareQuery(ctx, store, "completions := data.regal.lsp.completion.items")
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

func (p *Policy) Run(
	ctx context.Context,
	c *cache.Cache,
	params types.CompletionParams,
	opts *Options,
) ([]types.CompletionItem, error) {
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
	inputContext["path_separator"] = rio.PathSeparator

	workspacePath := uri.ToPath(opts.ClientIdentifier, opts.RootURI)

	inputDotJSONPath, inputDotJSONContent := rio.FindInput(
		uri.ToPath(opts.ClientIdentifier, params.TextDocument.URI),
		workspacePath,
	)

	if inputDotJSONPath != "" && inputDotJSONContent != nil {
		inputContext["input_dot_json_path"] = inputDotJSONPath
		inputContext["input_dot_json"] = inputDotJSONContent
	}

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

	result, err := rego2.QueryRegalBundle(ctx, input, p.pq)
	if err != nil {
		return nil, fmt.Errorf("failed querying regal bundle: %w", err)
	}

	completions := make([]types.CompletionItem, 8)

	if err := encoding.JSONRoundTrip(result["completions"], &completions); err != nil {
		return nil, fmt.Errorf("failed converting completions: %w", err)
	}

	return completions, nil
}

func prepareQuery(ctx context.Context, store storage.Store, query string) (*rego.PreparedEvalQuery, error) {
	regoArgs := prepareRegoArgs(store, ast.MustParseBody(query))

	txn, err := store.NewTransaction(ctx, storage.WriteParams)
	if err != nil {
		return nil, fmt.Errorf("failed creating transaction: %w", err)
	}

	regoArgs = append(regoArgs, rego.Transaction(txn))

	// Note that we currently don't provide metrics or profiling here, and
	// most likely we should — need to consider how to best make that conditional
	// and how to present it if enabled.
	pq, err := rego.New(regoArgs...).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed preparing query: %s, %w", query, err)
	}

	if err = store.Commit(ctx, txn); err != nil {
		return nil, fmt.Errorf("failed committing transaction: %w", err)
	}

	return &pq, nil
}

func prepareRegoArgs(store storage.Store, query ast.Body) []func(*rego.Rego) {
	return []func(*rego.Rego){
		rego.StoreReadAST(true),
		rego.Store(store),
		rego.ParsedQuery(query),
		rego.ParsedBundle("regal", &rbundle.LoadedBundle),
		rego.Function2(builtins.RegalParseModuleMeta, builtins.RegalParseModule),
		rego.Function1(builtins.RegalLastMeta, builtins.RegalLast),
		// Uncomment for development
		// rego.EnablePrintStatements(true),
		// rego.PrintHook(topdown.NewPrintHook(os.Stderr)),
	}
}
