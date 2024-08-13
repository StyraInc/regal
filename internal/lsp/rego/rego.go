package rego

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"

	rbundle "github.com/styrainc/regal/bundle"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/uri"
	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/pkg/builtins"
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
	return types.Position{
		Line:      uint(loc.Row - 1),
		Character: uint(loc.Col - 1),
	}
}

func LocationFromPosition(pos types.Position) *ast.Location {
	return &ast.Location{
		Row: int(pos.Line + 1),
		Col: int(pos.Character + 1),
	}
}

// AllBuiltinCalls returns all built-in calls in the module, excluding operators
// and any other function identified by an infix.
func AllBuiltinCalls(module *ast.Module) []BuiltInCall {
	builtinCalls := make([]BuiltInCall, 0)

	callVisitor := ast.NewGenericVisitor(func(x interface{}) bool {
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

		if b, ok := BuiltIns[terms[0].Value.String()]; ok {
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

//nolint:gochecknoglobals
var keywordsPreparedQuery *rego.PreparedEvalQuery

//nolint:gochecknoglobals
var ruleHeadLocationsPreparedQuery *rego.PreparedEvalQuery

//nolint:gochecknoglobals
var preparedQueriesInitOnce sync.Once

func initialize() {
	regalRules := rio.MustLoadRegalBundleFS(rbundle.Bundle)

	sharedRegoArgs := []func(*rego.Rego){
		rego.ParsedBundle("regal", &regalRules),
		rego.Function2(builtins.RegalParseModuleMeta, builtins.RegalParseModule),
		rego.Function1(builtins.RegalLastMeta, builtins.RegalLast),
	}

	keywordRegoArgs := append(sharedRegoArgs, rego.Query("data.regal.ast.keywords"))

	kwpq, err := rego.New(keywordRegoArgs...).PrepareForEval(context.Background())
	if err != nil {
		panic(err)
	}

	keywordsPreparedQuery = &kwpq

	ruleHeadLocationsRegoArgs := append(sharedRegoArgs, rego.Query("data.regal.ast.rule_head_locations"))

	rhlpq, err := rego.New(ruleHeadLocationsRegoArgs...).PrepareForEval(context.Background())
	if err != nil {
		panic(err)
	}

	ruleHeadLocationsPreparedQuery = &rhlpq
}

// AllKeywords returns all keywords in the module.
func AllKeywords(ctx context.Context, fileName, contents string, module *ast.Module) (map[string][]KeywordUse, error) {
	preparedQueriesInitOnce.Do(initialize)

	enhancedInput, err := parse.PrepareAST(fileName, contents, module)
	if err != nil {
		return nil, fmt.Errorf("failed enhancing input: %w", err)
	}

	rs, err := keywordsPreparedQuery.Eval(ctx, rego.EvalInput(enhancedInput))
	if err != nil {
		return nil, fmt.Errorf("failed evaluating keywords: %w", err)
	}

	if len(rs) != 1 {
		return nil, errors.New("expected exactly one result from evaluation")
	}

	if len(rs[0].Expressions) != 1 {
		return nil, errors.New("expected exactly one expression in result")
	}

	var result map[string][]KeywordUse

	err = rio.JSONRoundTrip(rs[0].Expressions[0].Value, &result)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshaling keywords: %w", err)
	}

	return result, nil
}

// AllRuleHeadLocations returns mapping of rules names to the head locations.
func AllRuleHeadLocations(ctx context.Context, fileName, contents string, module *ast.Module) (RuleHeads, error) {
	preparedQueriesInitOnce.Do(initialize)

	enhancedInput, err := parse.PrepareAST(fileName, contents, module)
	if err != nil {
		return nil, fmt.Errorf("failed enhancing input: %w", err)
	}

	rs, err := ruleHeadLocationsPreparedQuery.Eval(ctx, rego.EvalInput(enhancedInput))
	if err != nil {
		return nil, fmt.Errorf("failed evaluating keywords: %w", err)
	}

	if len(rs) == 0 {
		return nil, errors.New("no results returned from evaluation")
	}

	if len(rs) != 1 {
		return nil, errors.New("expected exactly one result from evaluation")
	}

	if len(rs[0].Expressions) != 1 {
		return nil, errors.New("expected exactly one expression in result")
	}

	var result RuleHeads

	err = rio.JSONRoundTrip(rs[0].Expressions[0].Value, &result)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshaling keywords: %w", err)
	}

	return result, nil
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

func QueryRegalBundle(input map[string]any, pq rego.PreparedEvalQuery) (map[string]any, error) {
	result, err := pq.Eval(context.Background(), rego.EvalInput(input))
	if err != nil {
		return nil, fmt.Errorf("failed evaluating query: %w", err)
	}

	if len(result) == 0 {
		return nil, errors.New("expected result from evaluation, didn't get it")
	}

	return result[0].Bindings, nil
}
