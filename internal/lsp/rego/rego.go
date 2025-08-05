package rego

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"

	rbundle "github.com/styrainc/regal/bundle"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/pkg/builtins"
	"github.com/styrainc/regal/pkg/roast/encoding"
	"github.com/styrainc/regal/pkg/roast/rast"
	"github.com/styrainc/regal/pkg/roast/transform"
)

var (
	emptyResult = rego.Result{}

	errNoResults          = errors.New("no results returned from evaluation")
	errExcpectedOneResult = errors.New("expected exactly one result from evaluation")
	errExcpectedOneExpr   = errors.New("expected exactly one expression in result")
)

func init() {
	ast.InternStringTerm(
		// All keys from Code Actions
		"identifier", "workspace_root_uri", "web_server_base_uri", "client", "params", "start", "end",
		"textDocument", "context", "range", "uri", "diagnostics", "only", "triggerKind", "codeDescription",
		"message", "severity", "source", "code", "data", "title", "command", "kind", "isPreferred",
	)
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

			builtinCalls = append(builtinCalls, BuiltInCall{
				Builtin:  b,
				Location: terms[0].Location,
				Args:     terms[1:],
			})
		}

		return false
	})

	callVisitor.Walk(module)

	return builtinCalls
}

var (
	keywordsPreparedQuery          *rego.PreparedEvalQuery
	ruleHeadLocationsPreparedQuery *rego.PreparedEvalQuery
	codeLensPreparedQuery          *rego.PreparedEvalQuery
	codeActionPreparedQuery        *rego.PreparedEvalQuery
)

var preparedQueriesInitOnce sync.Once

type policy struct {
	module   *ast.Module
	fileName string
	contents string
}

func initialize() {
	keywordsPreparedQuery = createPreparedQuery("data.regal.ast.keywords")
	codeLensPreparedQuery = createPreparedQuery("data.regal.lsp.codelens.lenses")
	codeActionPreparedQuery = createPreparedQuery("data.regal.lsp.codeaction.actions")
	ruleHeadLocationsPreparedQuery = createPreparedQuery("data.regal.ast.rule_head_locations")
}

// AllKeywords returns all keywords in the module.
func AllKeywords(ctx context.Context, fileName, contents string, module *ast.Module) (map[string][]KeywordUse, error) {
	preparedQueriesInitOnce.Do(initialize)

	var keywords map[string][]KeywordUse

	value, err := queryToValue(ctx, keywordsPreparedQuery, policy{module, fileName, contents}, keywords)
	if err != nil {
		return nil, fmt.Errorf("failed querying code lenses: %w", err)
	}

	return value, nil
}

// AllRuleHeadLocations returns mapping of rules names to the head locations.
func AllRuleHeadLocations(ctx context.Context, fileName, contents string, module *ast.Module) (RuleHeads, error) {
	preparedQueriesInitOnce.Do(initialize)

	var ruleHeads RuleHeads

	value, err := queryToValue(ctx, ruleHeadLocationsPreparedQuery, policy{module, fileName, contents}, ruleHeads)
	if err != nil {
		return nil, fmt.Errorf("failed querying code lenses: %w", err)
	}

	return value, nil
}

// CodeLenses returns all code lenses in the module.
func CodeLenses(ctx context.Context, uri, contents string, module *ast.Module) ([]types.CodeLens, error) {
	preparedQueriesInitOnce.Do(initialize)

	var codeLenses []types.CodeLens

	value, err := queryToValue(ctx, codeLensPreparedQuery, policy{module, uri, contents}, codeLenses)
	if err != nil {
		return nil, fmt.Errorf("failed querying code lenses: %w", err)
	}

	return value, nil
}

// CodeActions returns all code actions in the module.
// Note that at least as of now, no code actions depend on the data in the module, so
// it is not passed as part of the input. This could change in the future.
func CodeActions(ctx context.Context, input CodeActionInput) ([]types.CodeAction, error) {
	preparedQueriesInitOnce.Do(initialize)

	var codeActions []types.CodeAction

	value, err := queryToValueWithParsedInput(ctx, codeActionPreparedQuery, rast.StructToValue(input), codeActions)
	if err != nil {
		return nil, fmt.Errorf("failed querying code lenses: %w", err)
	}

	return value, nil
}

func queryToValue[T any](ctx context.Context, pq *rego.PreparedEvalQuery, policy policy, toValue T) (T, error) {
	input, err := transform.ToAST(policy.fileName, policy.contents, policy.module, false)
	if err != nil {
		return toValue, fmt.Errorf("failed to prepare input: %w", err)
	}

	return queryToValueWithParsedInput(ctx, pq, input, toValue)
}

func queryToValueWithParsedInput[T any](
	ctx context.Context,
	pq *rego.PreparedEvalQuery,
	input ast.Value,
	toValue T,
) (T, error) {
	result, err := toValidResult(pq.Eval(ctx, rego.EvalParsedInput(input)))
	if err != nil {
		return toValue, err
	}

	if err := encoding.JSONRoundTrip(result.Expressions[0].Value, &toValue); err != nil {
		return toValue, fmt.Errorf("failed unmarshaling code lenses: %w", err)
	}

	return toValue, nil
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

func QueryRegalBundle(ctx context.Context, input ast.Value, pq rego.PreparedEvalQuery) (map[string]any, error) {
	result, err := pq.Eval(ctx, rego.EvalParsedInput(input))
	if err != nil {
		return nil, fmt.Errorf("failed evaluating query: %w", err)
	}

	if len(result) == 0 {
		return nil, errNoResults
	}

	return result[0].Bindings, nil
}

func createArgs(args ...func(*rego.Rego)) []func(*rego.Rego) {
	always := append([]func(*rego.Rego){
		rego.ParsedBundle("regal", &rbundle.LoadedBundle),
		rego.StoreReadAST(true),
	}, builtins.RegalBuiltinRegoFuncs...)

	return append(always, args...)
}

func createPreparedQuery(query string) *rego.PreparedEvalQuery {
	args := createArgs(rego.ParsedQuery(rast.RefStringToBody(query)))

	pq, err := rego.New(args...).PrepareForEval(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to prepare query %s: %v", query, err))
	}

	return &pq
}

func (c CodeActionInput) String() string { // For debugging only
	s, err := encoding.JSON().MarshalToString(&c)
	if err != nil {
		return fmt.Sprintf("CodeActionInput marshalling error: %v", err)
	}

	return s
}
