package main

import (
	"fmt"
	"net"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/bingoohuang/ginx/pkg/ginpprof"
)

func main() {
	r := gin.Default()
	r.Use(gin.Logger())
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// automatically add routers for net/http/pprof
	// e.g. /debug/pprof, /debug/pprof/heap, etc.
	ginpprof.Wrap(r)

	// Authorization group
	// authorized := r.Group("/", AuthRequired())
	// exactly the same as:
	authorized := r.Group("/")
	// per group middleware! in this case we use the custom created
	// AuthRequired() middleware just in the "authorized" group.
	authorized.Use(AuthRequired())
	{
		authorized.POST("/login", func(c *gin.Context) {})
		authorized.POST("/submit", func(c *gin.Context) {})
		authorized.POST("/read", func(c *gin.Context) {})

		// nested group
		testing := authorized.Group("/testing")
		testing.GET("/analytics", func(c *gin.Context) {})
	}

	// ginpprof also plays well with *gin.RouterGroup
	// group := r.Group("/debug/pprof")
	// ginpprof.WrapGroup(group)

	addr := FreeAddr()
	fmt.Println("start to listen on", addr)

	panic(r.Run(addr))
}

func AuthRequired() gin.HandlerFunc {
	return nil
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
