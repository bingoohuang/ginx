package anyfunc_test

import (
	"errors"
	"testing"

	"github.com/bingoohuang/ginx/pkg/adapt"
	"github.com/bingoohuang/ginx/pkg/anyfunc"
	"github.com/bingoohuang/ginx/pkg/gintest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	r := adapt.Adapt(gin.New())
	r.RegisterAdapter(anyfunc.NewAdapter())

	r.Any("/error", anyfunc.Fn(func() error { return errors.New("error occurred") }))
	r.GET("/ok", anyfunc.Fn(func() error { return nil }))
	r.GET("/url", anyfunc.Fn(func(c *gin.Context) (string, error) { return c.Request.URL.String(), nil }))

	rr := gintest.Get("/error", r)
	assert.Equal(t, 500, rr.StatusCode())
	assert.Equal(t, "error: error occurred", rr.Body())

	rr = gintest.Get("/ok", r)
	assert.Equal(t, 200, rr.StatusCode())

	rr = gintest.Get("/url", r)
	assert.Equal(t, 200, rr.StatusCode())
	assert.Equal(t, "/url", rr.Body())
}
