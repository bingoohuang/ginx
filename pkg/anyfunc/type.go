package anyfunc

import "reflect"

// 参考 https://github.com/uber-go/dig/blob/master/types.go
// nolint:gochecknoglobals
var (
	// ErrType defines the error's type
	ErrType = reflect.TypeOf((*error)(nil)).Elem()
)

// ImplType tells src whether it implements target type.
func ImplType(src, target reflect.Type) bool {
	if src == target || src.Kind() == reflect.Ptr && src.Elem() == target {
		return true
	}

	if target.Kind() != reflect.Interface {
		return false
	}

	return reflect.PtrTo(src).Implements(target)
}

// IsError tells t whether it is error type exactly.
func IsError(t reflect.Type) bool { return t == ErrType }

// AsError tells t whether it implements error type exactly.
func AsError(t reflect.Type) bool { return ImplType(t, ErrType) }
