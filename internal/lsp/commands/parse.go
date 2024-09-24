package commands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/internal/lsp/types"
)

type ParseOptions struct {
	TargetArgIndex int
	RowArgIndex    int
	ColArgIndex    int
}

type ParseResult struct {
	Location *ast.Location
	Target   string
}

// Parse is responsible for extracting the target and location from the given params command params sent from the client
// after acting on a Code Action.
func Parse(params types.ExecuteCommandParams, opts ParseOptions) (*ParseResult, error) {
	if len(params.Arguments) == 0 {
		return nil, errors.New("no args supplied")
	}

	target := ""

	if opts.TargetArgIndex < len(params.Arguments) {
		target = fmt.Sprintf("%s", params.Arguments[opts.TargetArgIndex])
	}

	// we can't extract a location from the same location as the target, so location arg positions
	// must not have been set in the opts.
	if opts.RowArgIndex == opts.TargetArgIndex {
		return &ParseResult{
			Target: target,
		}, nil
	}

	var loc *ast.Location

	if opts.RowArgIndex < len(params.Arguments) && opts.ColArgIndex < len(params.Arguments) {
		var row, col int

		switch v := params.Arguments[opts.RowArgIndex].(type) {
		case int:
			row = v
		case string:
			var err error

			row, err = strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("failed to parse row: %w", err)
			}
		default:
			return nil, fmt.Errorf("unexpected type for row: %T", params.Arguments[opts.RowArgIndex])
		}

		switch v := params.Arguments[opts.ColArgIndex].(type) {
		case int:
			col = v
		case string:
			var err error

			col, err = strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("failed to parse col: %w", err)
			}
		default:
			return nil, fmt.Errorf("unexpected type for col: %T", params.Arguments[opts.ColArgIndex])
		}

		loc = &ast.Location{
			Row: row,
			Col: col,
		}
	}

	return &ParseResult{
		Target:   target,
		Location: loc,
	}, nil
}
