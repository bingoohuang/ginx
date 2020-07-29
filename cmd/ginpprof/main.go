package main

import (
	"fmt"
	"net"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/bingoohuang/ginx/pkg/ginpprof"
)

func main() {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// automatically add routers for net/http/pprof
	// e.g. /debug/pprof, /debug/pprof/heap, etc.
	ginpprof.Wrap(router)

	// ginpprof also plays well with *gin.RouterGroup
	// group := router.Group("/debug/pprof")
	// ginpprof.WrapGroup(group)

	addr := FreeAddr()
	fmt.Println("start to listen on", addr)

	panic(router.Run(addr))
}

// FreeAddr asks the kernel for a free open port that is ready to use.
func FreeAddr() string {
	if v := os.Getenv("ADDR"); v != "" {
		return v
	}

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return ":10020"
	}

	_ = l.Close()

	return fmt.Sprintf(":%d", l.Addr().(*net.TCPAddr).Port)
}
