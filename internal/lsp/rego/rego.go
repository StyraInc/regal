package rego

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/bundle"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/storage"

	rbundle "github.com/styrainc/regal/bundle"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/pkg/builtins"
	"github.com/styrainc/regal/pkg/roast/encoding"
	"github.com/styrainc/regal/pkg/roast/rast"
	"github.com/styrainc/regal/pkg/roast/transform"
	"github.com/styrainc/regal/pkg/roast/util/concurrent"
)

var (
	emptyResult            = rego.Result{}
	errNoResults           = errors.New("no results returned from evaluation")
	errExcpectedOneResult  = errors.New("expected exactly one result from evaluation")
	errExcpectedOneExpr    = errors.New("expected exactly one expression in result")
	keywordsQuery          = "data.regal.ast.keywords"
	codeLensQuery          = "data.regal.lsp.codelens.lenses"
	codeActionQuery        = "data.regal.lsp.codeaction.actions"
	ruleHeadLocationsQuery = "data.regal.ast.rule_head_locations"
	prepared               = concurrent.MapOf(make(map[string]*cachedQuery, 5))
	simpleRefPattern       = regexp.MustCompile(`^[a-zA-Z\.]$`)
)

type regoOptions = []func(*rego.Rego)

type cachedQuery struct {
	body     ast.Body
	prepared *rego.PreparedEvalQuery
	store    storage.Store
}

func init() {
	ast.InternStringTerm(
		// All keys from Code Actions
		"identifier", "workspace_root_uri", "web_server_base_uri", "client", "params", "start", "end",
		"textDocument", "context", "range", "uri", "diagnostics", "only", "triggerKind", "codeDescription",
		"message", "severity", "source", "code", "data", "title", "command", "kind", "isPreferred",
	)
}

type policy struct {
	module   *ast.Module
	fileName string
	contents string
}

type BuiltInCall struct {
	Builtin  *ast.Builtin
	Location *ast.Location
	Args     []*ast.Term
}

type KeywordUse struct {
	Name     string             `json:"name"`
	Location KeywordUseLocation `json:"location"`
}

type RuleHeads map[string][]*ast.Location

type KeywordUseLocation struct {
	Row uint `json:"row"`
	Col uint `json:"col"`
}

type Client struct {
	Identifier            clients.Identifier           `json:"identifier"`
	InitializationOptions *types.InitializationOptions `json:"init_options,omitempty"`
}

type Environment struct {
	WorkspaceRootURI string `json:"workspace_root_uri"`
	WebServerBaseURI string `json:"web_server_base_uri"`
}

type RegalContext struct {
	Client      Client      `json:"client"`
	Environment Environment `json:"environment"`
}

type CodeActionInput struct {
	Regal  RegalContext           `json:"regal"`
	Params types.CodeActionParams `json:"params"`
}

func (c CodeActionInput) String() string { // For debugging only
	s, err := encoding.JSON().MarshalToString(&c)
	if err != nil {
		return fmt.Sprintf("CodeActionInput marshalling error: %v", err)
	}

	return s
}

func PositionFromLocation(loc *ast.Location) types.Position {
	//nolint:gosec
	return types.Position{
		Line:      uint(loc.Row - 1),
		Character: uint(loc.Col - 1),
	}
}

func LocationFromPosition(pos types.Position) *ast.Location {
	return &ast.Location{
		Row: int(pos.Line + 1),      //nolint: gosec
		Col: int(pos.Character + 1), //nolint: gosec
	}
}

// AllBuiltinCalls returns all built-in calls in the module, excluding operators
// and any other function identified by an infix.
func AllBuiltinCalls(module *ast.Module, builtins map[string]*ast.Builtin) []BuiltInCall {
	builtinCalls := make([]BuiltInCall, 0)

	callVisitor := ast.NewGenericVisitor(func(x any) bool {
		var terms []*ast.Term

		switch node := x.(type) {
		case ast.Call:
			terms = node
		case *ast.Expr:
			if call, ok := node.Terms.([]*ast.Term); ok {
				terms = call
			}
		default:
			return false
		}

		if len(terms) == 0 {
			return false
		}

		if b, ok := builtins[terms[0].Value.String()]; ok {
			// Exclude operators and similar builtins
			if b.Infix != "" {
				return false
			}

			builtinCalls = append(builtinCalls, BuiltInCall{Builtin: b, Location: terms[0].Location, Args: terms[1:]})
		}

		return false
	})

	callVisitor.Walk(module)

	return builtinCalls
}

// AllKeywords returns all keywords in the module.
func AllKeywords(ctx context.Context, fileName, contents string, module *ast.Module) (map[string][]KeywordUse, error) {
	var keywords map[string][]KeywordUse

	err := policyToValue(ctx, keywordsQuery, policy{module, fileName, contents}, &keywords)
	if err != nil {
		return nil, fmt.Errorf("failed querying for all keywords: %w", err)
	}

	return keywords, nil
}

// AllRuleHeadLocations returns mapping of rules names to the head locations.
func AllRuleHeadLocations(ctx context.Context, fileName, contents string, module *ast.Module) (RuleHeads, error) {
	var ruleHeads RuleHeads

	err := policyToValue(ctx, ruleHeadLocationsQuery, policy{module, fileName, contents}, &ruleHeads)
	if err != nil {
		return nil, fmt.Errorf("failed querying for rule head locations: %w", err)
	}

	return ruleHeads, nil
}

// CodeLenses returns all code lenses in the module.
func CodeLenses(ctx context.Context, uri, contents string, module *ast.Module) ([]types.CodeLens, error) {
	var codeLenses []types.CodeLens

	err := policyToValue(ctx, codeLensQuery, policy{module, uri, contents}, &codeLenses)
	if err != nil {
		return nil, fmt.Errorf("failed querying code lenses: %w", err)
	}

	return codeLenses, nil
}

// CodeActions returns all code actions in the module.
// Note that at least as of now, no code actions depend on the data in the module, so
// it is not passed as part of the input. This could change in the future.
func CodeActions(ctx context.Context, input CodeActionInput) ([]types.CodeAction, error) {
	var codeActions []types.CodeAction

	err := CachedQueryEval(ctx, codeActionQuery, rast.StructToValue(input), &codeActions)
	if err != nil {
		return nil, fmt.Errorf("failed querying code actions: %w", err)
	}

	return codeActions, nil
}

func policyToValue[T any](ctx context.Context, query string, policy policy, toValue *T) error {
	input, err := transform.ToAST(policy.fileName, policy.contents, policy.module, false)
	if err != nil {
		return fmt.Errorf("failed to prepare input: %w", err)
	}

	return CachedQueryEval(ctx, query, input, toValue)
}

func CachedQueryEval[T any](ctx context.Context, query string, input ast.Value, toValue *T) error {
	cq, err := getOrSetCachedQuery(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed preparing query: %w", err)
	}

	result, err := toValidResult(cq.prepared.Eval(ctx, rego.EvalParsedInput(input)))
	if err != nil {
		return err
	}

	// TODO: let's do something more roubust than this. but fine for now
	fromValue := result.Expressions[0].Value
	if strings.Contains(query, "=") {
		// We don't know the name of the binding here, so just assume there's only one
		for _, v := range result.Bindings {
			fromValue = v

			break
		}
	}

	if err := encoding.JSONRoundTrip(fromValue, toValue); err != nil {
		return fmt.Errorf("failed unmarshaling value: %w", err)
	}

	return nil
}

func toValidResult(rs rego.ResultSet, err error) (rego.Result, error) {
	switch {
	case err != nil:
		return emptyResult, fmt.Errorf("evaluation failed: %w", err)
	case len(rs) == 0:
		return emptyResult, errNoResults
	case len(rs) != 1:
		return emptyResult, errExcpectedOneResult
	case len(rs[0].Expressions) != 1:
		return emptyResult, errExcpectedOneExpr
	}

	return rs[0], nil
}

func prepareQueryArgs(
	ctx context.Context,
	query ast.Body,
	store storage.Store,
	regalBundle *bundle.Bundle,
) (regoOptions, storage.Transaction) {
	var txn storage.Transaction

	args := append([]func(*rego.Rego){
		rego.ParsedQuery(query),
		rego.ParsedBundle("regal", regalBundle),
	}, builtins.RegalBuiltinRegoFuncs...)

	if store != nil {
		txn, _ = store.NewTransaction(ctx, storage.WriteParams)
		args = append(args, rego.Store(store), rego.Transaction(txn))
	} else {
		args = append(args, rego.StoreReadAST(true))
	}

	return args, txn
}

func getOrSetCachedQuery(ctx context.Context, query string, store storage.Store) (*cachedQuery, error) {
	cq, ok := prepared.Get(query)
	if !ok {
		parsedQuery := ParseQuery(query)

		pq, err := prepareQuery(ctx, parsedQuery, store)
		if err != nil {
			return nil, fmt.Errorf("failed preparing query %q: %w", query, err)
		}

		cq = &cachedQuery{body: parsedQuery, prepared: pq, store: store}

		prepared.Set(query, cq)

		return cq, nil
	}

	if isBundleDevelopmentMode() {
		// In dev mode, we always prepare the query to ensure changes in the bundle are reflected
		// immediately. We can however reuse the query and the store (if set).
		pq, err := prepareQuery(ctx, cq.body, cq.store)
		if err != nil {
			return nil, fmt.Errorf("failed preparing query %q: %w", query, err)
		}

		cq.prepared = pq
	}

	return cq, nil
}

func StoreCachedQuery(ctx context.Context, query string, store storage.Store) error {
	parsedQuery := ParseQuery(query)

	pq, err := prepareQuery(ctx, parsedQuery, store)
	if err != nil {
		return fmt.Errorf("failed preparing query %q: %w", query, err)
	}

	prepared.Set(query, &cachedQuery{body: parsedQuery, prepared: pq, store: store})

	return nil
}

func prepareQuery(ctx context.Context, query ast.Body, store storage.Store) (*rego.PreparedEvalQuery, error) {
	args, txn := prepareQueryArgs(ctx, query, store, rbundle.LoadedBundle())

	// Note that we currently don't provide metrics or profiling here, and
	// most likely we should — need to consider how to best make that conditional
	// and how to present it if enabled.
	pq, err := rego.New(args...).PrepareForEval(ctx)
	if err != nil {
		if store != nil {
			store.Abort(ctx, txn)
		}

		if isBundleDevelopmentMode() {
			// Try falling back to the embedded bundle, or else we'll
			// easily have errors popping up as notifications, making it
			// really hard to fix the issue that broke the query (like a parse error)
			args, txn = prepareQueryArgs(ctx, query, store, rbundle.EmbeddedBundle())
			if pq, err = rego.New(args...).PrepareForEval(ctx); err == nil {
				if store != nil && txn != nil {
					if err = store.Commit(ctx, txn); err != nil {
						return nil, err
					}
				}

				return &pq, nil
			}

			if store != nil {
				store.Abort(ctx, txn)
			}
		}

		return nil, err
	}

	if store != nil && txn != nil {
		if err = store.Commit(ctx, txn); err != nil {
			return nil, err
		}
	}

	return &pq, nil
}

func ParseQuery(query string) ast.Body {
	// Try cheap parsing if possible
	if simpleRefPattern.MatchString(query) {
		return rast.RefStringToBody(query)
	}

	return ast.MustParseBody(query)
}

func isBundleDevelopmentMode() bool {
	return os.Getenv("REGAL_BUNDLE_PATH") != ""
}
