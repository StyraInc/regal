package rego

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"

	rbundle "github.com/styrainc/regal/bundle"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/uri"
	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/pkg/builtins"

	"github.com/styrainc/roast/pkg/encoding"
	"github.com/styrainc/roast/pkg/transform"
)

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
)

var preparedQueriesInitOnce sync.Once

type policy struct {
	module   *ast.Module
	fileName string
	contents string
}

func initialize() {
	createArgs := func(args ...func(*rego.Rego)) []func(*rego.Rego) {
		always := append([]func(*rego.Rego){
			rego.ParsedBundle("regal", &rbundle.LoadedBundle),
			rego.StoreReadAST(true),
		}, builtins.RegalBuiltinRegoFuncs...)

		return append(always, args...)
	}

	keywordRegoArgs := createArgs(rego.Query("data.regal.ast.keywords"))

	kwpq, err := rego.New(keywordRegoArgs...).PrepareForEval(context.Background())
	if err != nil {
		panic(err)
	}

	keywordsPreparedQuery = &kwpq

	ruleHeadLocationsRegoArgs := createArgs(rego.Query("data.regal.ast.rule_head_locations"))

	rhlpq, err := rego.New(ruleHeadLocationsRegoArgs...).PrepareForEval(context.Background())
	if err != nil {
		panic(err)
	}

	ruleHeadLocationsPreparedQuery = &rhlpq

	codeLensRegoArgs := createArgs(rego.Query("data.regal.lsp.codelens.lenses"))

	clpq, err := rego.New(codeLensRegoArgs...).PrepareForEval(context.Background())
	if err != nil {
		panic(err)
	}

	codeLensPreparedQuery = &clpq
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

func queryToValue[T any](ctx context.Context, pq *rego.PreparedEvalQuery, policy policy, toValue T) (T, error) {
	input, err := parse.PrepareAST(policy.fileName, policy.contents, policy.module)
	if err != nil {
		return toValue, fmt.Errorf("failed to prepare input: %w", err)
	}

	inputValue, err := transform.ToOPAInputValue(input)
	if err != nil {
		return toValue, fmt.Errorf("failed converting input to value: %w", err)
	}

	result, err := toValidResult(pq.Eval(ctx, rego.EvalParsedInput(inputValue)))
	if err != nil {
		return toValue, err
	}

	if err := encoding.JSONRoundTrip(result.Expressions[0].Value, &toValue); err != nil {
		return toValue, fmt.Errorf("failed unmarshaling code lenses: %w", err)
	}

	return toValue, nil
}

func toValidResult(rs rego.ResultSet, err error) (rego.Result, error) {
	if err != nil {
		return rego.Result{}, fmt.Errorf("evaluation failed: %w", err)
	}

	if len(rs) == 0 {
		return rego.Result{}, errors.New("no results returned from evaluation")
	}

	if len(rs) != 1 {
		return rego.Result{}, errors.New("expected exactly one result from evaluation")
	}

	if len(rs[0].Expressions) != 1 {
		return rego.Result{}, errors.New("expected exactly one expression in result")
	}

	return rs[0], nil
}

// ToInput prepares a module with Regal additions to be used as input for evaluation.
func ToInput(
	fileURI string,
	cid clients.Identifier,
	content string,
	context map[string]any,
) (map[string]any, error) {
	path := uri.ToPath(cid, fileURI)

	input := map[string]any{
		"regal": map[string]any{
			"file": map[string]any{
				"name":  path,
				"uri":   fileURI,
				"lines": strings.Split(content, "\n"),
			},
			"context": context,
		},
	}

	if regal, ok := input["regal"].(map[string]any); ok {
		if f, ok := regal["file"].(map[string]any); ok {
			f["uri"] = fileURI
		}

		regal["client_id"] = cid
	}

	return SetInputContext(input, context), nil
}

func SetInputContext(input map[string]any, context map[string]any) map[string]any {
	if regal, ok := input["regal"].(map[string]any); ok {
		regal["context"] = context
	}

	return input
}

func QueryRegalBundle(ctx context.Context, input map[string]any, pq rego.PreparedEvalQuery) (map[string]any, error) {
	inputValue, err := transform.ToOPAInputValue(input)
	if err != nil {
		return nil, fmt.Errorf("failed converting input map to value: %w", err)
	}

	result, err := pq.Eval(ctx, rego.EvalParsedInput(inputValue))
	if err != nil {
		return nil, fmt.Errorf("failed evaluating query: %w", err)
	}

	if len(result) == 0 {
		return nil, errors.New("expected result from evaluation, didn't get it")
	}

	return result[0].Bindings, nil
}
