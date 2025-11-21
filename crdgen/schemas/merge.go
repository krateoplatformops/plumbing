package schemas

import (
	"fmt"
	"reflect"

	"dario.cat/mergo"
)

func MergeTypes(types []*Type) (*Type, error) {
	if len(types) == 0 {
		return nil, ErrEmptyTypesList
	}

	result := &Type{}

	if isPrimitiveTypeList(types) {
		return result, nil
	}

	opts := []func(*mergo.Config){
		mergo.WithAppendSlice, // merge slice (per Required, Type ecc.)
		mergo.WithTransformers(typeListTransformer{}),
		mergo.WithTransformers(typeTransformer{}),
	}

	for _, t := range types {
		if err := mergo.Merge(result, t, opts...); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotMergeTypes, err)
		}
	}

	return result, nil
}

type typeTransformer struct{}

func (typeTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeOf(map[string]*Type{}) {
		return func(dst, src reflect.Value) error {
			if src.IsNil() {
				return nil
			}
			if dst.IsNil() {
				dst.Set(reflect.MakeMap(src.Type()))
			}
			for _, key := range src.MapKeys() {
				dst.SetMapIndex(key, src.MapIndex(key))
			}
			return nil
		}
	}
	return nil
}

type typeListTransformer struct{}

func (t typeListTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeOf(TypeList{}) {
		return func(dst, src reflect.Value) error {
			return nil
		}
	}

	return nil
}
