package anyfn

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

// AdapterDealer is the dealer for a specified type.
type AdapterDealer interface {
	Deal(*gin.Context)
}

type Adapter struct {
}

func NewAdapter() *Adapter {
	adapter := &Adapter{}

	return adapter
}

type Before interface {
	// Do will be called Before the adaptee invoking.
	Do(args []interface{}) error
}

type BeforeFn func(args []interface{}) error

func (b BeforeFn) Do(args []interface{}) error { return b(args) }

type After interface {
	// Do will be called Before the adaptee invoking.
	Do(args []interface{}, results []interface{}) error
}

type AfterFn func(args []interface{}, results []interface{}) error

func (b AfterFn) Do(args []interface{}, results []interface{}) error { return b(args, results) }

type AnyF struct {
	F      interface{}
	Before Before
	After  After
}

func F(v interface{}) *AnyF {
	return &AnyF{F: v}
}

func F3(v interface{}, before Before, after After) *AnyF {
	return &AnyF{F: v, Before: before, After: after}
}

var AnyFType = reflect.TypeOf((*AnyF)(nil))

func (a *Adapter) Support(relativePath string, arg interface{}) bool {
	return reflect.TypeOf(arg) == AnyFType
}

func (a *Adapter) Adapt(relativePath string, argV interface{}) gin.HandlerFunc {
	anyF := argV.(*AnyF)
	fv := reflect.ValueOf(anyF.F)

	return func(c *gin.Context) {
		if err := a.internalAdapter(c, fv, anyF); err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
		}
	}
}

func (a *Adapter) internalAdapter(c *gin.Context, fv reflect.Value, anyF *AnyF) error {
	argVs, err := a.createArgs(c, fv)
	if err != nil {
		return err
	}

	if err := a.before(argVs, anyF.Before); err != nil {
		return err
	}

	r := fv.Call(argVs)

	if err := a.after(argVs, r, anyF.After); err != nil {
		return err
	}

	return a.processOut(c, fv, r)
}

func (a *Adapter) before(v []reflect.Value, f Before) error {
	if f == nil {
		return nil
	}

	return f.Do(Values(v).Interface())
}

func (a *Adapter) after(v, results []reflect.Value, f After) error {
	if f == nil {
		return nil
	}

	return f.Do(Values(v).Interface(), Values(results).Interface())
}

type Values []reflect.Value

func (v Values) Interface() []interface{} {
	args := make([]interface{}, len(v))
	for i, a := range v {
		args[i] = a.Interface()
	}

	return args
}

func (a *Adapter) processOut(c *gin.Context, fv reflect.Value, r []reflect.Value) error {
	ft := fv.Type()
	numOut := ft.NumOut()

	if numOut == 0 {
		return nil
	}

	if AsError(ft.Out(numOut - 1)) { // nolint:gomnd
		if !r[numOut-1].IsNil() {
			return r[numOut-1].Interface().(error)
		}

		numOut-- // drop the error returned by the adapted.
	}

	a.succProcess(c, numOut, r)

	return nil
}

func (a *Adapter) succProcess(c *gin.Context, numOut int, r []reflect.Value) {
	vs := make([]interface{}, numOut)

	for i := 0; i < numOut; i++ {
		vs[i] = r[i].Interface()
		if v, ok := vs[i].(AdapterDealer); ok {
			v.Deal(c)
			return
		}
	}

	succProcessorInternal(c, vs...)
}

func succProcessorInternal(g *gin.Context, vs ...interface{}) {
	code, vs := findStateCode(vs)

	if len(vs) == 0 {
		g.Status(code)
		return
	}

	if len(vs) == 1 {
		respondOut1(g, vs, code)
		return
	}

	m := make(map[string]interface{})

	for _, v := range vs {
		m[reflect.TypeOf(v).String()] = v
	}

	g.JSON(code, m)
}

func respondOut1(g *gin.Context, vs []interface{}, code int) {
	switch v0 := vs[0]; reflect.Indirect(reflect.ValueOf(v0)).Kind() {
	case reflect.Struct, reflect.Map:
		g.JSON(code, v0)
	default:
		g.String(code, "%v", v0)
	}
}

func (a *Adapter) createArgs(c *gin.Context, fv reflect.Value) (v []reflect.Value, err error) {
	ft := fv.Type()
	argIns := parseArgIns(ft)
	v = make([]reflect.Value, ft.NumIn())
	singleArgValue := singlePrimitiveValue(c, argIns)

	for i, arg := range argIns {
		if v[i], err = a.createArgValue(c, arg, singleArgValue); err != nil {
			return nil, err
		}
	}

	return v, err
}

func (a *Adapter) createArgValue(c *gin.Context, arg argIn, singleArgValue string) (reflect.Value, error) {
	switch arg.Kind {
	case reflect.Struct:
		v, err := a.processStruct(c, arg)
		if err != nil {
			return reflect.Value{}, err
		}

		return ConvertPtr(arg.Ptr, v), nil
	case reflect.Interface:
		if arg.Type == HTTPResponseWriterType {
			return reflect.ValueOf(c.Writer), nil
		}
	}

	if arg.PrimitiveIndex < 0 {
		return reflect.Value{}, fmt.Errorf("unable to parse arg%d for %s", arg.Index, arg.Type)
	}

	if singleArgValue != "" {
		return arg.convertValue(singleArgValue)
	}

	return reflect.Zero(arg.Type), nil
}

var (
	GinContextType         = reflect.TypeOf((*gin.Context)(nil)).Elem()
	HTTPRequestType        = reflect.TypeOf((*http.Request)(nil)).Elem()
	HTTPResponseWriterType = reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
)

func (a *Adapter) processStruct(c *gin.Context, arg argIn) (reflect.Value, error) {
	if arg.Ptr && arg.Type == GinContextType { // 直接注入gin.Context
		return reflect.ValueOf(c), nil
	}

	if arg.Ptr && arg.Type == HTTPRequestType {
		return reflect.ValueOf(c.Request), nil
	}

	for _, v := range c.Keys {
		if arg.Type == NonPtrTypeOf(v) {
			return reflect.ValueOf(v), nil
		}
	}

	argValue := reflect.New(arg.Type)
	if err := c.ShouldBind(argValue.Interface()); err != nil {
		return reflect.Value{}, &AdapterError{Err: err, Context: "ShouldBind"}
	}

	return argValue, nil
}
