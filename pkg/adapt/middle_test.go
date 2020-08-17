package adapt_test

import (
	"testing"

	"github.com/bingoohuang/ginx/pkg/adapt"
	"github.com/bingoohuang/ginx/pkg/anyfunc"
	"github.com/bingoohuang/ginx/pkg/gintest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type AuthUser struct {
	Name string
}

func TestMiddleware(t *testing.T) {
	r := adapt.Adapt(gin.New())
	r.RegisterAdapter(anyfunc.NewAdapter())

	r.Use(func(c *gin.Context) {
		c.Set("AuthUser", AuthUser{Name: "TestAuthUser"})
	})

	doTest(t, r)
}

func TestMiddlewarePtr(t *testing.T) {
	r := adapt.Adapt(gin.New())
	r.RegisterAdapter(anyfunc.NewAdapter())

	r.Use(func(c *gin.Context) {
		c.Set("AuthUser", &AuthUser{Name: "TestAuthUser"})
	})

	doTest(t, r)
}

func doTest(t *testing.T, r *adapt.Adaptee) {
	r.GET("/GetAge1/:name", anyfunc.Fn(func(user AuthUser, name string) string {
		return user.Name + "/" + name
	}))
	r.GET("/GetAge2/:name", anyfunc.Fn(func(name string, user AuthUser) string {
		return user.Name + "/" + name
	}))
	r.GET("/GetAge3/:name", anyfunc.Fn(func(user *AuthUser, name string) string {
		return user.Name + "/" + name
	}))
	r.GET("/GetAge4/:name", anyfunc.Fn(func(name string, user *AuthUser) string {
		return user.Name + "/" + name
	}))

	// r.Run(":8080")

	rr := gintest.Get("/GetAge1/bingoohuang", r)
	assert.Equal(t, "TestAuthUser/bingoohuang", rr.Body())
	rr = gintest.Get("/GetAge2/bingoohuang", r)
	assert.Equal(t, "TestAuthUser/bingoohuang", rr.Body())
	rr = gintest.Get("/GetAge3/bingoohuang", r)
	assert.Equal(t, "TestAuthUser/bingoohuang", rr.Body())
	rr = gintest.Get("/GetAge4/bingoohuang", r)
	assert.Equal(t, "TestAuthUser/bingoohuang", rr.Body())
}
