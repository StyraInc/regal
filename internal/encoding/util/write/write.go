package write

import (
	"encoding/json/jsontext"
	"encoding/json/v2"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/util"
)

var locationKeyArr = [...]byte{'"', 'l', 'o', 'c', 'a', 't', 'i', 'o', 'n', '"'}

type marshalToFn[T any] func(*jsontext.Encoder, T) error

func ArrayField[T any](enc *jsontext.Encoder, name string, vals []T) error {
	if len(vals) == 0 {
		return nil
	}

	return ArrayFieldFn(enc, name, vals, marshalEncodeUnvariadic)
}

func ArrayOfString(enc *jsontext.Encoder, arr []string) (err error) {
	if err = enc.WriteToken(jsontext.BeginArray); err == nil {
		if err = Foreach(enc, arr, String); err == nil {
			err = enc.WriteToken(jsontext.EndArray)
		}
	}

	return err
}

func ArrayOfStringField(enc *jsontext.Encoder, name string, vals []string) (err error) {
	if len(vals) > 0 {
		if err = String(enc, name); err == nil {
			err = ArrayOfString(enc, vals)
		}
	}

	return err
}

func ArrayFieldFn[T any](enc *jsontext.Encoder, name string, vals []T, f marshalToFn[T]) (err error) {
	if len(vals) > 0 {
		if err = String(enc, name); err == nil {
			err = ArrayToFn(enc, vals, f)
		}
	}

	return err
}

func ArrayFieldTo[T json.MarshalerTo](enc *jsontext.Encoder, name string, vals []T) (err error) {
	if len(vals) > 0 {
		if err = String(enc, name); err == nil {
			err = ArrayTo(enc, vals)
		}
	}

	return err
}

func ArrayFieldToFn[T any](enc *jsontext.Encoder, name string, vals []T, f marshalToFn[T]) (err error) {
	if len(vals) > 0 {
		if err = String(enc, name); err == nil {
			err = ArrayToFn(enc, vals, f)
		}
	}

	return err
}

func ArrayTo[T json.MarshalerTo](enc *jsontext.Encoder, arr []T) error {
	enc.WriteToken(jsontext.BeginArray)

	for _, item := range arr {
		if err := item.MarshalJSONTo(enc); err != nil {
			return err
		}
	}

	return enc.WriteToken(jsontext.EndArray)
}

func ArrayToFn[T any](enc *jsontext.Encoder, arr []T, f marshalToFn[T]) (err error) {
	enc.WriteToken(jsontext.BeginArray)

	if err = Foreach(enc, arr, f); err == nil {
		err = enc.WriteToken(jsontext.EndArray)
	}

	return err
}

func Map[T any](enc *jsontext.Encoder, m map[string]T) error {
	if len(m) == 0 {
		return nil
	}

	enc.WriteToken(jsontext.BeginObject)

	for k, v := range m {
		_ = String(enc, k)
		_ = json.MarshalEncode(enc, v)
	}

	return enc.WriteToken(jsontext.EndObject)
}

func MapField[T any](enc *jsontext.Encoder, name string, m map[string]T) (err error) {
	if len(m) > 0 {
		if err = String(enc, name); err == nil {
			err = Map(enc, m)
		}
	}

	return err
}

func ObjectStart(enc *jsontext.Encoder, loc *ast.Location) (err error) {
	enc.WriteToken(jsontext.BeginObject)

	if loc != nil {
		enc.WriteValue(locationKeyArr[:])

		endRow, endCol := loc.End()

		buf := append(enc.AvailableBuffer(), '"')
		buf = append(util.AppendInt(buf, loc.Row), ':')
		buf = append(util.AppendInt(buf, loc.Col), ':')
		buf = append(util.AppendInt(buf, endRow), ':')
		buf = append(util.AppendInt(buf, endCol), '"')

		return enc.WriteValue(buf)
	}

	return err
}

func Tokens(enc *jsontext.Encoder, tokens ...jsontext.Token) error {
	for _, token := range tokens {
		if err := enc.WriteToken(token); err != nil {
			return err
		}
	}

	return nil
}

func WhenTrue(enc *jsontext.Encoder, condition bool, name string) (err error) {
	if condition {
		if err = String(enc, name); err == nil {
			err = enc.WriteToken(jsontext.True)
		}
	}

	return err
}

func StringField[T ~string](enc *jsontext.Encoder, name string, val T) error {
	if val == "" {
		return nil
	}

	return Tokens(enc, jsontext.String(name), jsontext.String(string(val)))
}

func Field[T any](enc *jsontext.Encoder, name string, val T) error {
	return FieldFn(enc, name, val, marshalEncodeUnvariadic)
}

func FieldFn[T any](enc *jsontext.Encoder, name string, val T, f marshalToFn[T]) (err error) {
	if err = String(enc, name); err == nil {
		err = f(enc, val)
	}

	return err
}

func FieldPtrFn[T any](enc *jsontext.Encoder, name string, val *T, f marshalToFn[*T]) error {
	if val == nil {
		return nil
	}

	return FieldFn(enc, name, val, f)
}

func String[T ~string](enc *jsontext.Encoder, val T) error {
	return enc.WriteToken(jsontext.String(string(val)))
}

func Foreach[T any](enc *jsontext.Encoder, arr []T, f marshalToFn[T]) (err error) {
	for _, item := range arr {
		if err = f(enc, item); err != nil {
			break
		}
	}

	return err
}

func marshalEncodeUnvariadic[T any](enc *jsontext.Encoder, val T) error {
	return json.MarshalEncode(enc, val)
}
