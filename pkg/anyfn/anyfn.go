package anyfn

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/bingoohuang/ginx/pkg/adapt"

	"github.com/gin-gonic/gin"
)

// AdapterDealer is the dealer for a specified type.
type AdapterDealer interface {
	Deal(*gin.Context)
}

type Interceptor interface {
	StartRequest(path string, c *gin.Context, attrs map[string]interface{}) *gin.Context
	EndRequest()
}

type Adapter struct {
	Interceptor Interceptor
}

func (a *Adapter) Default(relativePath string) adapt.Handler {
	return nil
}

func (a *Adapter) Adapt(relativePath string, argV interface{}) adapt.Handler {
	anyF, ok := argV.(*anyF)
	if !ok {
		return nil
	}

	fv := reflect.ValueOf(anyF.F)

	return adapt.HandlerFunc(func(c *gin.Context) {
		if a.Interceptor != nil {
			c = a.Interceptor.StartRequest(relativePath, c, anyF.Option.Attrs)
			if c == nil {
				return
			}

			defer a.Interceptor.EndRequest()
		}

		if err := a.internalAdapter(c, fv, anyF); err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
		}
	})
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

type anyF struct {
	P      *Adapter
	F      interface{}
	Option *Option
}

func (a *anyF) Parent() adapt.Adapter { return a.P }

type Option struct {
	Before Before
	After  After
	Attrs  map[string]interface{}
}

type OptionFn func(*Option)

func (a *Adapter) Before(before Before) OptionFn {
	return func(f *Option) {
		f.Before = before
	}
}

func (a *Adapter) After(after After) OptionFn {
	return func(f *Option) {
		f.After = after
	}
}

func (a *Adapter) AttrMap(attrs map[string]interface{}) OptionFn {
	return func(f *Option) {
		for k, v := range attrs {
			f.Attrs[k] = v
		}
	}
}

func (a *Adapter) Attr(k string, v interface{}) OptionFn {
	return func(f *Option) {
		f.Attrs[k] = v
	}
}

func (a *Adapter) F(v interface{}, fns ...OptionFn) *anyF {
	option := &Option{
		Attrs: make(map[string]interface{}),
	}

	for _, fn := range fns {
		fn(option)
	}

	return &anyF{F: v, Option: option}
}

func (a *Adapter) internalAdapter(c *gin.Context, fv reflect.Value, anyF *anyF) error {
	argVs, err := a.createArgs(c, fv)
	if err != nil {
		return err
	}

	if err := a.before(argVs, anyF.Option.Before); err != nil {
		return err
	}

	r := fv.Call(argVs)

	if err := a.after(argVs, r, anyF.Option.After); err != nil {
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
