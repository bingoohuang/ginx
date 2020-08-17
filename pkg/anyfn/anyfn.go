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

type anyfn struct {
	v interface{}
}

func F(v interface{}) *anyfn {
	return &anyfn{v: v}
}

var Type = reflect.TypeOf((*anyfn)(nil))

func (a *Adapter) Support(arg interface{}) bool {
	return reflect.TypeOf(arg) == Type
}

func NewAdapter() *Adapter {
	adapter := &Adapter{}

	return adapter
}

func (a *Adapter) Adapt(argV interface{}) gin.HandlerFunc {
	arg := argV.(*anyfn).v
	fv := reflect.ValueOf(arg)

	return func(c *gin.Context) {
		if err := a.internalAdapter(c, fv); err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
		}
	}
}

func (a *Adapter) internalAdapter(c *gin.Context, fv reflect.Value) error {
	argVs, err := a.createArgs(c, fv)
	if err != nil {
		return err
	}

	r := fv.Call(argVs)

	return a.processOut(c, fv, r)
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
	if arg.Kind == reflect.Struct {
		v, err := a.processStruct(c, arg)
		if err != nil {
			return reflect.Value{}, err
		}

		return ConvertPtr(arg.Ptr, v), nil
	}

	if arg.PrimitiveIndex < 0 {
		return reflect.Value{}, fmt.Errorf("unable to parse arg%d for %s", arg.Index, arg.Type)
	}

	if singleArgValue != "" {
		return arg.convertValue(singleArgValue)
	}

	return reflect.Zero(arg.Type), nil
}

func (a *Adapter) processStruct(c *gin.Context, arg argIn) (reflect.Value, error) {
	if arg.Ptr && arg.Type == NonPtrTypeOf(c) { // 直接注入gin.Context
		return reflect.ValueOf(c), nil
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
