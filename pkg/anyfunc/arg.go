package anyfunc

import (
	"fmt"
	"reflect"

	"github.com/bingoohuang/ginx/pkg/cast"
)

type argIn struct {
	Index          int
	Type           reflect.Type
	Kind           reflect.Kind
	Ptr            bool
	PrimitiveIndex int
}

func parseArgIns(ft reflect.Type) []argIn {
	numIn := ft.NumIn()
	argIns := make([]argIn, numIn)

	for i := 0; i < numIn; i++ {
		argIns[i] = parseArgs(ft, i)
	}

	return argIns
}

func parseArgs(ft reflect.Type, argIndex int) argIn {
	argType := ft.In(argIndex)
	ptr := argType.Kind() == reflect.Ptr

	if ptr {
		argType = argType.Elem()
	}

	return argIn{Index: argIndex, Type: argType, Kind: argType.Kind(), Ptr: ptr, PrimitiveIndex: -1}
}

func (arg argIn) convertValue(s string) (reflect.Value, error) {
	v, err := cast.To(s, arg.Type)
	if err != nil {
		return reflect.Value{}, &AdapterError{
			Err:     err,
			Context: fmt.Sprintf("To %s to %v", s, arg.Type),
		}
	}

	return ConvertPtr(arg.Ptr, v), nil
}
