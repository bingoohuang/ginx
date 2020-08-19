package hlog

import (
	"fmt"

	"github.com/bingoohuang/ginx/pkg/adapt"
	"github.com/gin-gonic/gin"
)

type Adapter struct {
}

func NewAdapter() *Adapter {
	adapter := &Adapter{}

	return adapter
}

type Middle struct {
	hlog         *hlog
	relativePath string
}

func (m *Middle) Handle(c *gin.Context) {
}

func (m *Middle) Before(c *gin.Context) (after adapt.Handler) {
	if m.hlog.Option.Ignore {
		return nil
	}

	fmt.Println("hlog before " + m.hlog.Option.Biz)

	return adapt.HandlerFunc(func(c *gin.Context) {
		fmt.Println("hlog after " + m.hlog.Option.Biz)
	})
}

func (a *Adapter) Default(relativePath string) adapt.Handler {
	return &Middle{
		relativePath: relativePath,
		hlog: &hlog{
			Option: &Option{},
		},
	}
}

func (a *Adapter) Adapt(relativePath string, argV interface{}) adapt.Handler {
	hlog, ok := argV.(*hlog)
	if !ok {
		return nil
	}

	middle := &Middle{
		relativePath: relativePath,
		hlog:         hlog,
	}

	return middle
}

type Option struct {
	Biz    string
	Ignore bool
}

type OptionFn func(option *Option)

func (a *Adapter) Biz(biz string) OptionFn {
	return func(option *Option) {
		option.Biz = biz
	}
}

func (a *Adapter) Ignore(b bool) OptionFn {
	return func(option *Option) {
		option.Ignore = b
	}
}

type hlog struct {
	Option *Option
	P      *Adapter
}

func (h *hlog) Parent() adapt.Adapter { return h.P }

func (a *Adapter) F(fns ...OptionFn) *hlog {
	o := &Option{}

	for _, f := range fns {
		f(o)
	}

	return &hlog{
		P:      a,
		Option: o,
	}
}
