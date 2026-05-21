// Package jsonvalue implements a generic `collections.ValueCodec[T]` backed
// by encoding/json. Used by the v0 keepers in place of `codec.CollValue[T]`
// — the latter requires proto-generated types, which we don't have until
// the buf pipeline is wired.
//
// JSON is slower and bulkier on-disk than proto, but it's correct, well-
// understood, and lets the chain compile + run for devnet purposes. When
// proto-gen lands, swap `jsonvalue.Codec[T]()` → `codec.CollValue[T](cdc)`
// in each keeper. The interface is identical, so the call-site change is
// one word.
package jsonvalue

import (
	"encoding/json"
	"fmt"
	"reflect"

	"cosmossdk.io/collections/codec"
)

// Codec returns a ValueCodec[T] for any type that round-trips through
// encoding/json.
func Codec[T any]() codec.ValueCodec[T] {
	return jsonCodec[T]{}
}

type jsonCodec[T any] struct{}

func (jsonCodec[T]) Encode(v T) ([]byte, error)        { return json.Marshal(v) }
func (jsonCodec[T]) EncodeJSON(v T) ([]byte, error)    { return json.Marshal(v) }
func (jsonCodec[T]) Decode(b []byte) (T, error)        { var v T; err := json.Unmarshal(b, &v); return v, err }
func (jsonCodec[T]) DecodeJSON(b []byte) (T, error)    { var v T; err := json.Unmarshal(b, &v); return v, err }
func (c jsonCodec[T]) Stringify(v T) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(b)
}
func (jsonCodec[T]) ValueType() string {
	var zero T
	return "json:" + reflect.TypeOf(zero).String()
}
