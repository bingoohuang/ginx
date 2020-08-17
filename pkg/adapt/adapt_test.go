package adapt_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bingoohuang/ginx/pkg/adapt"
	"github.com/bingoohuang/ginx/pkg/gintest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBeforeAdapt(t *testing.T) {
	r := gin.New()

	// This handler will match /user/john but will not match /user/ or /user
	r.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.String(http.StatusOK, "Hello %s", name)
	})

	// r.Run(":8080")

	rr := gintest.Get("/user/bingoohuang", r)
	assert.Equal(t, "Hello bingoohuang", rr.Body())
}

func TestAdapt(t *testing.T) {
	r := adapt.Adapt(gin.New())
	r.RegisterAdapter(func(f func(string) string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.String(http.StatusOK, f(StringArg(c)))
		}
	})

	// This handler will match /user/john but will not match /user/ or /user
	r.GET("/user/:name", func(name string) string {
		return fmt.Sprintf("Hello %s", name)
	})

	// r.Run(":8080")

	rr := gintest.Get("/user/bingoohuang", r)
	assert.Equal(t, "Hello bingoohuang", rr.Body())
}

func StringArg(c *gin.Context) string {
	if len(c.Params) == 1 {
		return c.Params[0].Value
	}

	if q := c.Request.URL.Query(); len(q) == 1 {
		for _, v := range q {
			return v[0]
		}
	}

	return ""
}
