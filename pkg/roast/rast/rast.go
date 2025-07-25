// Package providing tools for working with Rego's AST library (not RoAST)
package rast

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
)

// UnquotedPath returns a slice of strings from a path without quotes.
// e.g. data.foo["bar"] -> ["foo", "bar"], note that the data is not included.
func UnquotedPath(path ast.Ref) []string {
	ret := make([]string, 0, len(path)-1)
	for _, ref := range path[1:] {
		ret = append(ret, strings.Trim(ref.String(), `"`))
	}

	return ret
}

// IsBodyGenerated checks if the body of a rule is generated by the parser.
func IsBodyGenerated(rule *ast.Rule) bool {
	if rule.Default {
		return true
	}

	if len(rule.Body) == 0 {
		return true
	}

	if rule.Head == nil {
		return false
	}

	if rule.Body[0] != nil && rule.Body[0].Location == rule.Location {
		return true
	}

	if rule.Body[0] != nil && rule.Head.Value != nil && rule.Body[0].Location == rule.Head.Value.Location {
		return true
	}

	if rule.Head.Key != nil &&
		rule.Body[0].Location.Row == rule.Head.Key.Location.Row &&
		rule.Body[0].Location.Col < rule.Head.Key.Location.Col {
		// This is a quirk in the original AST — the generated body will have a location
		// set before the key, i.e. "message"
		return true
	}

	return false
}

// RefStringToBody converts a simple dot-delimited string path to an ast.Body.
// This is a lightweight alternative to ast.ParseBody that avoids the overhead of parsing,
// and benefits from using interned terms when possible. It is also nowhere near as competent,
// and can only handle simple string paths without vars, numbers, etc. Suitable for use with
// e.g. rego.ParsedQuery and other places where a simple ref is needed. Do *NOT* use the returned
// ast.Body anywhere it might be mutated (like having location data added), as that modifies the
// globally interned terms.
//
// Implementations tested:
// -----------------------
// 333.6 ns/op	     472 B/op	      19 allocs/op - SplitSeq
// 330.7 ns/op	     496 B/op	      16 allocs/op - Split
// 269.1 ns/op	     400 B/op	      15 allocs/op - IndexOf for loop (current).
func RefStringToBody(path string) ast.Body {
	var i int
	if i = strings.Index(path, "."); i == -1 {
		return ast.NewBody(ast.NewExpr(ast.RefTerm(refHeadTerm(path))))
	}

	terms := append(make([]*ast.Term, 0, strings.Count(path, ".")+1), refHeadTerm(path[:i]))

	for {
		path = path[i+1:]
		if i = strings.Index(path, "."); i == -1 {
			if len(path) > 0 {
				terms = append(terms, ast.InternedTerm(path))
			}

			break
		}

		terms = append(terms, ast.InternedTerm(path[:i]))
	}

	return ast.NewBody(ast.NewExpr(ast.RefTerm(terms...)))
}

// RefStringToRef converts a simple dot-delimited string path to an ast.Ref in the most
// efficient way possible, using interned terms where applicable. See RefStringToBody for
// more details on the limitations of this function.
func RefStringToRef(path string) ast.Ref {
	var i int
	if i = strings.Index(path, "."); i == -1 {
		return ast.Ref([]*ast.Term{refHeadTerm(path)})
	}

	terms := append(make([]*ast.Term, 0, strings.Count(path, ".")+1), refHeadTerm(path[:i]))

	for {
		path = path[i+1:]
		if i = strings.Index(path, "."); i == -1 {
			if len(path) > 0 {
				terms = append(terms, ast.InternedTerm(path))
			}

			break
		}

		terms = append(terms, ast.InternedTerm(path[:i]))
	}

	return ast.Ref(terms)
}

// LinesArrayTerm converts a string with newlines into an ast.Term array holding each line.
func LinesArrayTerm(content string) *ast.Term {
	parts := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	terms := make([]*ast.Term, len(parts))

	for i := range parts {
		terms[i] = ast.InternedTerm(parts[i])
	}

	return ast.ArrayTerm(terms...)
}

func refHeadTerm(name string) *ast.Term {
	switch name {
	case "data":
		return ast.DefaultRootDocument
	case "input":
		return ast.InputRootDocument
	default:
		return ast.VarTerm(name)
	}
}

// StructToValue converts a struct to ast.Value using 'json' struct tags (e.g., `json:"field,omitempty"`)
// but without an expensive JSON roundtrip.
// Experimental: this is new and not yet battle-tested, so use with caution.
func StructToValue(input any) ast.Value {
	v := reflect.ValueOf(input)
	t := reflect.TypeOf(input)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	kvs := make([][2]*ast.Term, 0, t.NumField())

	for i := range t.NumField() {
		field := t.Field(i)

		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		value := v.Field(i)

		if strings.Contains(tag, ",") {
			parts := strings.Split(tag, ",")
			tag = parts[0]

			omitempty := slices.Contains(parts[1:], "omitempty")
			if omitempty && isZeroValue(value) {
				continue
			}
		}

		kvs = append(kvs, ast.Item(ast.InternedTerm(tag), ast.NewTerm(toAstValue(value.Interface()))))
	}

	return ast.NewObject(kvs...)
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.IsNil() || v.Len() == 0
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Struct:
		for i := range v.NumField() {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}

		return true
	}

	return false
}

func toAstValue(v any) ast.Value {
	if v == nil {
		return ast.NullValue
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return ast.NullValue
		}

		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return StructToValue(rv.Interface())
	case reflect.Slice, reflect.Array:
		l := rv.Len()
		if l == 0 {
			return ast.InternedEmptyArrayValue
		}

		arr := make([]*ast.Term, 0, l)
		for i := range l {
			arr = append(arr, internedAny(rv.Index(i).Interface()))
		}

		return ast.NewArray(arr...)
	case reflect.Map:
		kvs := make([][2]*ast.Term, 0, rv.Len())

		for _, key := range rv.MapKeys() {
			var k *ast.Term

			ki := key.Interface()
			if s, ok := ki.(string); ok {
				k = ast.InternedTerm(s)
			} else {
				k = ast.InternedTerm(fmt.Sprintf("%v", ki))
			}

			kvs = append(kvs, [2]*ast.Term{k, internedAny(rv.MapIndex(key).Interface())})
		}

		return ast.NewObject(kvs...)
	case reflect.String:
		return ast.String(rv.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return ast.Number(strconv.FormatInt(rv.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return ast.Number(strconv.FormatUint(rv.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		return ast.Number(fmt.Sprintf("%v", rv.Float()))
	case reflect.Bool:
		return ast.InternedTerm(rv.Bool()).Value
	}
	// Fallback: string representation
	//nolint:forbidigo
	fmt.Println("WARNING: Unsupported type for conversion to ast.Value:", rv.Kind())

	return ast.String(fmt.Sprintf("%v", v))
}

func internedAny(v any) *ast.Term {
	switch value := v.(type) {
	case bool:
		return ast.InternedTerm(value)
	case string:
		return ast.InternedTerm(value)
	case int:
		return ast.InternedTerm(value)
	case uint:
		return ast.InternedTerm(int(value)) // //nolint:gosec
	case int64:
		return ast.InternedTerm(int(value))
	case float64:
		return ast.FloatNumberTerm(value)
	default:
		return ast.NewTerm(toAstValue(v))
	}
}
