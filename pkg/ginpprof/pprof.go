package ginpprof

import (
	"fmt"
	"net/http/pprof"
	"strings"

	"github.com/gin-gonic/gin"
)

// Wrap adds several routes from package `net/http/pprof` to *gin.Engine object
func Wrap(router interface{}) {
	if r, ok := router.(*gin.Engine); ok {
		WrapGroup(&r.RouterGroup)
	} else if r, ok := router.(*gin.RouterGroup); ok {
		WrapGroup(r)
	} else {
		panic(fmt.Errorf("please wrap *gin.Engine or *gin.RouterGroup"))
	}
}

// WrapGroup adds several routes from package `net/http/pprof` to *gin.RouterGroup object
func WrapGroup(router *gin.RouterGroup) {
	basePath := strings.TrimSuffix(router.BasePath(), "/")

	var prefix string

	switch {
	case basePath == "":
		prefix = ""
	case strings.HasSuffix(basePath, "/debug"):
		prefix = "/debug"
	case strings.HasSuffix(basePath, "/debug/pprof"):
		prefix = "/debug/pprof"
	}

	for _, r := range routers {
		router.Handle(r.Method, strings.TrimPrefix(r.Path, prefix), r.Handler)
	}
}

var routers = []struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}{
	{"GET", "/debug/pprof/", func(c *gin.Context) {
		pprof.Index(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/heap", func(c *gin.Context) {
		pprof.Handler("heap").ServeHTTP(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/goroutine", func(c *gin.Context) {
		pprof.Handler("goroutine").ServeHTTP(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/allocs", func(c *gin.Context) {
		pprof.Handler("allocs").ServeHTTP(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/block", func(c *gin.Context) {
		pprof.Handler("block").ServeHTTP(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/threadcreate", func(c *gin.Context) {
		pprof.Handler("threadcreate").ServeHTTP(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/cmdline", func(c *gin.Context) {
		pprof.Cmdline(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/profile", func(c *gin.Context) {
		pprof.Profile(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/symbol", func(c *gin.Context) {
		pprof.Symbol(c.Writer, c.Request)
	}},
	{"POST", "/debug/pprof/symbol", func(c *gin.Context) {
		pprof.Symbol(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/trace", func(c *gin.Context) {
		pprof.Trace(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/mutex", func(c *gin.Context) {
		pprof.Handler("mutex").ServeHTTP(c.Writer, c.Request)
	}},
}
