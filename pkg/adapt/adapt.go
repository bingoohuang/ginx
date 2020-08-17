package adapt

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

type Adapter interface {
	Support(interface{}) bool
	Adapt(arg interface{}) gin.HandlerFunc
}

type Adaptee struct {
	Router       Gin
	adapterFuncs map[reflect.Type]*adapterFuncItem
	adapters     []Adapter
}

type adapterFuncItem struct {
	adapterFunc reflect.Value
}

type Gin interface {
	gin.IRoutes
	http.Handler
}

func (i *adapterFuncItem) invoke(adapteeFn interface{}) gin.HandlerFunc {
	args := []reflect.Value{reflect.ValueOf(adapteeFn)}
	result := i.adapterFunc.Call(args)
	return result[0].Convert(ginHandlerFuncType).Interface().(gin.HandlerFunc)
}

func (a *Adaptee) Use(f func(c *gin.Context)) {
	a.Router.Use(f)
}

func (a *Adaptee) Any(relativePath string, args ...interface{}) {
	a.Router.Any(relativePath, a.createHandlerFuncs(args)...)
}

func (a *Adaptee) POST(relativePath string, args ...interface{}) {
	a.Router.POST(relativePath, a.createHandlerFuncs(args)...)
}

func (a *Adaptee) GET(relativePath string, args ...interface{}) {
	a.Router.GET(relativePath, a.createHandlerFuncs(args)...)
}

func (a *Adaptee) DELETE(relativePath string, args ...interface{}) {
	a.Router.DELETE(relativePath, a.createHandlerFuncs(args)...)
}

func (a *Adaptee) PUT(relativePath string, args ...interface{}) {
	a.Router.PUT(relativePath, a.createHandlerFuncs(args)...)
}

func (a *Adaptee) PATCH(relativePath string, args ...interface{}) {
	a.Router.PATCH(relativePath, a.createHandlerFuncs(args)...)
}

func (a *Adaptee) OPTIONS(relativePath string, args ...interface{}) {
	a.Router.OPTIONS(relativePath, a.createHandlerFuncs(args)...)
}

func (a *Adaptee) HEAD(relativePath string, args ...interface{}) {
	a.Router.HEAD(relativePath, a.createHandlerFuncs(args)...)
}

func (a *Adaptee) createHandlerFuncs(args []interface{}) []gin.HandlerFunc {
	hfs := make([]gin.HandlerFunc, 0, len(args))

	for _, arg := range args {
		if hf := a.adapt(arg); hf != nil {
			hfs = append(hfs, hf)
		}
	}

	return hfs
}

func (a *Adaptee) adapt(arg interface{}) gin.HandlerFunc {
	if f, ok := arg.(gin.HandlerFunc); ok {
		return f
	}

	if fn := a.findAdapterFunc(arg); fn != nil {
		return fn
	}

	return a.findAdapter(arg)
}

func (a *Adaptee) findAdapterFunc(arg interface{}) gin.HandlerFunc {
	argType := reflect.TypeOf(arg)

	for funcType, funcItem := range a.adapterFuncs {
		if argType.ConvertibleTo(funcType) {
			return funcItem.invoke(arg)
		}
	}

	return nil
}

func (a *Adaptee) findAdapter(arg interface{}) gin.HandlerFunc {
	for _, adapter := range a.adapters {
		if adapter.Support(arg) {
			return adapter.Adapt(arg)
		}
	}

	return nil
}

func Adapt(router *gin.Engine) *Adaptee {
	return &Adaptee{
		Router:       router,
		adapterFuncs: make(map[reflect.Type]*adapterFuncItem),
	}
}

func (a *Adaptee) ServeHTTP(r http.ResponseWriter, w *http.Request) {
	a.Router.ServeHTTP(r, w)
}

var ginHandlerFuncType = reflect.TypeOf(gin.HandlerFunc(nil))

func (a *Adaptee) RegisterAdapter(adapter interface{}) {
	if v, ok := adapter.(Adapter); ok {
		a.adapters = append(a.adapters, v)
		return
	}

	adapterValue := reflect.ValueOf(adapter)
	t := adapterValue.Type()

	if t.Kind() != reflect.Func {
		panic(fmt.Errorf("register method should use a func"))
	}

	if t.NumIn() != 1 || t.In(0).Kind() != reflect.Func {
		panic(fmt.Errorf("register method should use a func which inputs a func"))
	}

	if t.NumOut() != 1 || !t.Out(0).ConvertibleTo(ginHandlerFuncType) {
		panic(fmt.Errorf("register method should use a func which returns gin.HandlerFunc"))
	}

	a.adapterFuncs[t.In(0)] = &adapterFuncItem{
		adapterFunc: adapterValue,
	}
}
