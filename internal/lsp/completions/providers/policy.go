package providers

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/storage"

	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/rego"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/uri"
	"github.com/styrainc/regal/pkg/roast/transform"
)

const completionsQuery = "data.regal.lsp.completion.items"

// Policy provides suggestions that have been determined by Rego policy.
type Policy struct{}

// NewPolicy creates a new Policy provider. This provider is distinctly different from the other providers
// as it acts like the entrypoint for all Rego-based providers, and not a single provider "function" like
// the Go providers do.
func NewPolicy(ctx context.Context, store storage.Store) *Policy {
	if err := rego.StoreCachedQuery(ctx, completionsQuery, store); err != nil {
		log.Fatalf("failed to start policy completions provider: %v", err)
	}

	return &Policy{}
}

func (*Policy) Name() string {
	return "policy"
}

func (*Policy) Run(
	ctx context.Context,
	c *cache.Cache,
	params types.CompletionParams,
	opts *Options,
) ([]types.CompletionItem, error) {
	// TODO: Merge this into the rego package
	if opts == nil {
		return nil, errors.New("options must be provided")
	}

	content, ok := c.GetFileContents(params.TextDocument.URI)
	if !ok {
		return nil, fmt.Errorf("could not get file contents for: %s", params.TextDocument.URI)
	}

	// input.regal.context
	location := rego.LocationFromPosition(params.Position)
	regalContext := ast.NewObject(
		ast.Item(ast.InternedTerm("location"), ast.ObjectTerm(
			ast.Item(ast.InternedTerm("row"), ast.InternedTerm(location.Row)),
			ast.Item(ast.InternedTerm("col"), ast.InternedTerm(location.Col)),
		)),
		ast.Item(ast.InternedTerm("client_identifier"), ast.InternedTerm(int(opts.ClientIdentifier))),
		ast.Item(ast.InternedTerm("workspace_root"), ast.InternedTerm(opts.RootURI)),
	)

	path := uri.ToPath(opts.ClientIdentifier, params.TextDocument.URI)

	// TODO: Avoid the intermediate map[string]any step and unmarshal directly into ast.Value.
	inputDotJSONPath, inputDotJSONContent := rio.FindInput(path, uri.ToPath(opts.ClientIdentifier, opts.RootURI))
	if inputDotJSONPath != "" && inputDotJSONContent != nil {
		inputDotJSONValue, err := transform.ToOPAInputValue(inputDotJSONContent)
		if err != nil {
			return nil, fmt.Errorf("failed converting input dot JSON content to value: %w", err)
		}

		regalContext.Insert(ast.InternedTerm("input_dot_json_path"), ast.InternedTerm(inputDotJSONPath))
		regalContext.Insert(ast.InternedTerm("input_dot_json"), ast.NewTerm(inputDotJSONValue))
	}

	// TODO: Schemas from annotations to be used for completions on types, etc.

	// input.regal
	regalObj := transform.RegalContext(path, content, opts.RegoVersion.String())
	regalObj.Insert(ast.InternedTerm("context"), ast.NewTerm(regalContext))

	fileRef := ast.Ref{ast.InternedTerm("file")}
	fileObj, _ := regalObj.Find(fileRef)
	//nolint:forcetypeassert
	fileObj.(ast.Object).Insert(ast.InternedTerm("uri"), ast.InternedTerm(params.TextDocument.URI))

	input := ast.NewObject(ast.Item(ast.InternedTerm("regal"), ast.NewTerm(regalObj)))

	var completions []types.CompletionItem

	if err := rego.CachedQueryEval(ctx, completionsQuery, input, &completions); err != nil {
		return nil, fmt.Errorf("failed querying for completion suggestions: %w", err)
	}

	return completions, nil
}
