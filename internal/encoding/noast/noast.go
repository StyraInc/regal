package noast

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/util"

	"github.com/open-policy-agent/regal/internal/encoding/util/write"
	"github.com/open-policy-agent/regal/internal/funsafe"
)

// Options provides JSON encoding and decoding options for OPA's
// ast.Values treated as "pure" values, e.g. not as AST nodes.
var Options = json.JoinOptions(
	json.WithMarshalers(json.JoinMarshalers(
		json.MarshalToFunc(ValueMarshalToFn),
	)),
	json.WithUnmarshalers(json.JoinUnmarshalers(
		json.UnmarshalFromFunc(TermUnmarshalFromFn),
		json.UnmarshalFromFunc(ValueUnmarshalFromFn),
		json.UnmarshalFromFunc(ObjectUnmarshalFromFn),
		json.UnmarshalFromFunc(ArrayUnmarshalFromFn),
	)),
)

// ValueMarshalToFn provides the most efficient methods for encoding and decoding
// OPA's ast.Values matching JSON types, as plain JSON. Use this for when you do
// **not** want RoAST.
func ValueMarshalToFn(enc *jsontext.Encoder, value ast.Value) error {
	switch v := value.(type) {
	case ast.Null, *ast.Null:
		return enc.WriteToken(jsontext.Null)
	case ast.Boolean:
		return enc.WriteToken(jsontext.Bool(bool(v)))
	case *ast.Boolean:
		return enc.WriteToken(jsontext.Bool(bool(*v)))
	case ast.Number:
		return enc.WriteValue(util.StringToByteSlice(v))
	case *ast.Number:
		return enc.WriteValue(util.StringToByteSlice((*v)))
	case ast.String:
		return write.String(enc, v)
	case *ast.String:
		return write.String(enc, *v)
	case *ast.Array:
		enc.WriteToken(jsontext.BeginArray)

		for i := range v.Len() {
			if err := ValueMarshalToFn(enc, v.Elem(i).Value); err != nil {
				return err
			}
		}

		enc.WriteToken(jsontext.EndArray)
	case ast.Set:
		enc.WriteToken(jsontext.BeginArray)

		for _, term := range v.Slice() {
			if err := ValueMarshalToFn(enc, term.Value); err != nil {
				return err
			}
		}

		enc.WriteToken(jsontext.EndArray)
	case ast.Object:
		enc.WriteToken(jsontext.BeginObject)

		if v.Len() > 0 {
			// Use of "funsafe" to allow traversing the object without
			// allocating, which isn't possible to do with any of the
			// public AST methods.. but needless to say, this should be
			// fixed in OPA rather than via hacks like this
			for _, elem := range funsafe.ObjectElems(v) {
				switch key := elem.Key().Value.(type) {
				case ast.String:
					write.String(enc, key)

					if err := ValueMarshalToFn(enc, elem.Value().Value); err != nil {
						return err
					}
				default:
					write.String(enc, key.String())

					if err := ValueMarshalToFn(enc, elem.Value().Value); err != nil {
						return err
					}
				}
			}
		}

		enc.WriteToken(jsontext.EndObject)
	default:
		return fmt.Errorf("can't encode value of type %s", ast.ValueName(v))
	}

	return nil
}

func TermUnmarshalFromFn(dec *jsontext.Decoder, term *ast.Term) error {
	return ValueUnmarshalFromFn(dec, &term.Value)
}

//nolint:gocritic // 'ptrToRefParam: consider `v' to be of non-pointer type' how?
func ValueUnmarshalFromFn(dec *jsontext.Decoder, v *ast.Value) (err error) {
	switch dec.PeekKind() {
	case jsontext.KindNull:
		if _, err = dec.ReadToken(); err == nil {
			*v = ast.NullValue
		}
	case jsontext.KindTrue:
		if _, err = dec.ReadToken(); err == nil {
			*v = ast.InternedBooleanTrueValue
		}
	case jsontext.KindFalse:
		if _, err = dec.ReadToken(); err == nil {
			*v = ast.InternedBooleanFalseValue
		}
	case jsontext.KindNumber:
		val, err := dec.ReadValue()
		if err != nil {
			return err
		}

		*v = ast.Number(val.String())
	case jsontext.KindString:
		val, err := dec.ReadValue()
		if err != nil {
			return err
		}

		dst, err := jsontext.AppendUnquote(make([]byte, 0, len(val)-2), val)
		if err != nil {
			return err
		}

		*v = ast.String(string(dst))
	case jsontext.KindBeginArray:
		arr := ast.NewArray()
		if err := json.UnmarshalDecode(dec, arr); err != nil {
			return err
		}

		*v = arr
	case jsontext.KindBeginObject:
		obj := ast.NewObject()
		if err := json.UnmarshalDecode(dec, &obj); err != nil {
			return err
		}

		*v = obj
	default:
		err = fmt.Errorf("unexpected JSON kind: %v", dec.PeekKind())
	}

	return err
}

func ArrayUnmarshalFromFn(dec *jsontext.Decoder, arr *ast.Array) (err error) {
	if _, err = dec.ReadToken(); err != nil { // '['
		return err
	}

	var (
		terms []*ast.Term
		term  *ast.Term
	)

	for err == nil {
		if dec.PeekKind() == jsontext.KindEndArray {
			_, err = dec.ReadToken() // ']'

			break
		}

		if term, err = readValueToTerm(dec); err == nil {
			terms = append(terms, term)
		}
	}

	if err == nil || errors.Is(err, io.EOF) {
		*arr = *ast.NewArray(terms...)
	}

	return err
}

//nolint:gocritic // 'ptrToRefParam: consider `obj' to be of non-pointer type' how?
func ObjectUnmarshalFromFn(dec *jsontext.Decoder, obj *ast.Object) (err error) {
	if _, err = dec.ReadToken(); err != nil { // '{'
		return err
	}

	var items [][2]*ast.Term

	for err == nil {
		var tok jsontext.Token

		if tok, err = dec.ReadToken(); err != nil || tok.Kind() == '}' {
			break
		}

		key := ast.InternedTerm(tok.String())

		value, err := readValueToTerm(dec)
		if err != nil {
			return err
		}

		items = append(items, [2]*ast.Term{key, value})
	}

	if err == nil || errors.Is(err, io.EOF) {
		*obj = ast.NewObject(items...)
	}

	return err
}

func readValueToTerm(dec *jsontext.Decoder) (*ast.Term, error) {
	value, err := dec.ReadValue()
	if err != nil {
		return nil, err
	}

	switch value.Kind() {
	case jsontext.KindNull:
		return ast.InternedNullTerm, nil
	case jsontext.KindTrue:
		return ast.InternedBooleanTrue, nil
	case jsontext.KindFalse:
		return ast.InternedBooleanFalse, nil
	case jsontext.KindNumber:
		if interned := ast.InternedIntNumberTermFromString(string(value)); interned != nil {
			return interned, nil
		}
	case jsontext.KindString:
		str, _ := strconv.Unquote(string(value))

		return ast.InternedTerm(str), nil
	case jsontext.KindBeginArray:
		if len(value) == 2 && value[0] == '[' && value[1] == ']' {
			return ast.InternedEmptyArray, nil
		}
	case jsontext.KindBeginObject:
		if len(value) == 2 && value[0] == '{' && value[1] == '}' {
			return ast.InternedEmptyObject, nil
		}
	}

	var val ast.Value
	if err := json.Unmarshal(value, &val, dec.Options()); err != nil {
		return nil, err
	}

	return ast.NewTerm(val), nil
}
