package adapt

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

type Adapter interface {
	Support(relativePath string, arg interface{}) bool
	Adapt(relativePath string, arg interface{}) gin.HandlerFunc
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
	gin.IRouter
	http.Handler
}

func (i *adapterFuncItem) invoke(adapteeFn interface{}) gin.HandlerFunc {
	args := []reflect.Value{reflect.ValueOf(adapteeFn)}
	result := i.adapterFunc.Call(args)
	return result[0].Convert(GinHandlerFuncType).Interface().(gin.HandlerFunc)
}

func (a *Adaptee) createHandlerFuncs(relativePath string, args []interface{}) []gin.HandlerFunc {
	hfs := make([]gin.HandlerFunc, 0, len(args))

	for _, arg := range args {
		if hf := a.adapt(relativePath, arg); hf != nil {
			hfs = append(hfs, hf)
		}
	}

	return hfs
}

func (a *Adaptee) adapt(relativePath string, arg interface{}) gin.HandlerFunc {
	if f := a.findAdapterFunc(arg); f != nil {
		return f
	}

	if f := a.findAdapter(relativePath, arg); f != nil {
		return f
	}

	if v := reflect.ValueOf(arg); v.Type().ConvertibleTo(GinHandlerFuncType) {
		return v.Convert(GinHandlerFuncType).Interface().(gin.HandlerFunc)
	}

	return nil
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

func (a *Adaptee) findAdapter(relativePath string, arg interface{}) gin.HandlerFunc {
	for _, v := range a.adapters {
		if v.Support(relativePath, arg) {
			return v.Adapt(relativePath, arg)
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

var GinHandlerFuncType = reflect.TypeOf(gin.HandlerFunc(nil))

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

	if t.NumOut() != 1 || !t.Out(0).ConvertibleTo(GinHandlerFuncType) {
		panic(fmt.Errorf("register method should use a func which returns gin.HandlerFunc"))
	}

	a.adapterFuncs[t.In(0)] = &adapterFuncItem{
		adapterFunc: adapterValue,
	}
}

func (a *Adaptee) Use(f func(c *gin.Context)) {
	a.Router.Use(f)
}

func (a *Adaptee) Any(relativePath string, args ...interface{}) {
	a.Router.Any(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *Adaptee) POST(relativePath string, args ...interface{}) {
	a.Router.POST(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *Adaptee) GET(relativePath string, args ...interface{}) {
	a.Router.GET(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *Adaptee) DELETE(relativePath string, args ...interface{}) {
	a.Router.DELETE(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *Adaptee) PUT(relativePath string, args ...interface{}) {
	a.Router.PUT(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *Adaptee) PATCH(relativePath string, args ...interface{}) {
	a.Router.PATCH(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *Adaptee) OPTIONS(relativePath string, args ...interface{}) {
	a.Router.OPTIONS(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *Adaptee) HEAD(relativePath string, args ...interface{}) {
	a.Router.HEAD(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *Adaptee) Group(relativePath string, args ...interface{}) *AdapteeGroup {
	g := a.Router.Group(relativePath, a.createHandlerFuncs(relativePath, args)...)
	return &AdapteeGroup{
		Adaptee:     a,
		RouterGroup: g,
	}
}

type AdapteeGroup struct {
	*Adaptee
	*gin.RouterGroup
}

func (a *AdapteeGroup) Use(f func(c *gin.Context)) {
	a.RouterGroup.Use(f)
}

func (a *AdapteeGroup) Any(relativePath string, args ...interface{}) {
	a.RouterGroup.Any(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *AdapteeGroup) POST(relativePath string, args ...interface{}) {
	a.RouterGroup.POST(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *AdapteeGroup) GET(relativePath string, args ...interface{}) {
	a.RouterGroup.GET(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *AdapteeGroup) DELETE(relativePath string, args ...interface{}) {
	a.Router.DELETE(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *AdapteeGroup) PUT(relativePath string, args ...interface{}) {
	a.RouterGroup.PUT(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *AdapteeGroup) PATCH(relativePath string, args ...interface{}) {
	a.RouterGroup.PATCH(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *AdapteeGroup) OPTIONS(relativePath string, args ...interface{}) {
	a.RouterGroup.OPTIONS(relativePath, a.createHandlerFuncs(relativePath, args)...)
}

func (a *AdapteeGroup) HEAD(relativePath string, args ...interface{}) {
	a.RouterGroup.HEAD(relativePath, a.createHandlerFuncs(relativePath, args)...)
}
