package anyfunc_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bingoohuang/ginx/pkg/adapt"
	"github.com/bingoohuang/ginx/pkg/anyfunc"
	"github.com/bingoohuang/ginx/pkg/gintest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAdaptAnyFunc(t *testing.T) {
	adapter := anyfunc.NewAdapter()

	r := adapt.Adapt(gin.New())
	r.RegisterAdapter(func(f func(string) string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.String(http.StatusOK, f(StringArg(c)))
		}
	})
	r.RegisterAdapter(adapter)

	// This handler will match /user/john but will not match /user/ or /user
	r.GET("/user/:name", func(name string) string {
		return fmt.Sprintf("Hello %s", name)
	})

	type MyObject struct {
		Name string
	}

	r.POST("/MyObject1", anyfunc.Fn(func(m MyObject) string {
		return "Object: " + m.Name
	}))

	r.POST("/MyObject2", anyfunc.Fn(func(m *MyObject) string {
		return "Object: " + m.Name
	}))

	// r.Run(":8080")

	rr := gintest.Get("/user/bingoohuang", r)
	assert.Equal(t, "Hello bingoohuang", rr.Body())

	rr = gintest.Post("/MyObject1", r, gintest.JSONVar(MyObject{Name: "bingoo"}))
	assert.Equal(t, "Object: bingoo", rr.Body())
	rr = gintest.Post("/MyObject2", r, gintest.JSONVar(MyObject{Name: "bingoo2"}))
	assert.Equal(t, "Object: bingoo2", rr.Body())
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
