//nolint:gochecknoglobals
package transforms

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/util"

	_ "github.com/styrainc/regal/pkg/roast/intern"
)

// AnyToValue converts a native Go value x to a Value.
// This is an optimized version of the same function in the OPA codebase,
// and optimized in a way that makes it useful only for a map[string]any
// unmarshaled from RoAST JSON. Don't use it for anything else.
func AnyToValue(x any) (ast.Value, error) {
	switch x := x.(type) {
	case nil:
		return ast.NullValue, nil
	case bool:
		return ast.InternedTerm(x).Value, nil
	case float64:
		ix := int(x)
		if x == float64(ix) {
			return ast.InternedTerm(ix).Value, nil
		}

		return ast.Number(strconv.FormatFloat(x, 'g', -1, 64)), nil
	case json.Number:
		if interned := ast.InternedTerm(string(x)); interned != nil {
			return interned.Value, nil
		}

		return ast.Number(x), nil
	case string:
		return ast.InternedTerm(x).Value, nil
	case []string:
		if len(x) == 0 {
			return ast.InternedEmptyArrayValue, nil
		}

		r := util.NewPtrSlice[ast.Term](len(x))

		for i, s := range x {
			r[i].Value = ast.InternedTerm(s).Value
		}

		return ast.NewArray(r...), nil
	case []any:
		if len(x) == 0 {
			return ast.InternedEmptyArrayValue, nil
		}

		r := util.NewPtrSlice[ast.Term](len(x))

		for i, e := range x {
			e, err := AnyToValue(e)
			if err != nil {
				return nil, err
			}

			r[i].Value = e
		}

		return ast.NewArray(r...), nil
	case map[string]any:
		if len(x) == 0 {
			return ast.InternedEmptyObject.Value, nil
		}

		kvs := util.NewPtrSlice[ast.Term](len(x) * 2)
		idx := 0

		for k, v := range x {
			kvs[idx].Value = ast.InternedTerm(k).Value

			v, err := AnyToValue(v)
			if err != nil {
				return nil, err
			}

			kvs[idx+1].Value = v

			idx += 2
		}

		tuples := make([][2]*ast.Term, len(kvs)/2)
		for i := 0; i < len(kvs); i += 2 {
			tuples[i/2] = *(*[2]*ast.Term)(kvs[i : i+2])
		}

		return ast.NewObject(tuples...), nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", x)
	}
}
