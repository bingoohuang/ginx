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

func TestDirect(t *testing.T) {
	r := adapt.Adapt(gin.New())
	r.RegisterAdapter(anyfunc.NewAdapter())

	r.GET("/direct1", anyfunc.F(func() interface{} {
		return anyfunc.DirectResponse{Code: 203}
	}))
	r.GET("/direct2", anyfunc.F(func() interface{} {
		return &anyfunc.DirectResponse{Code: 201, Error: errors.New("abc")}
	}))

	rr := gintest.Get("/direct1", r)
	assert.Equal(t, 203, rr.StatusCode())
	assert.Equal(t, "", rr.Body())

	rr = gintest.Get("/direct2", r)
	assert.Equal(t, 201, rr.StatusCode())
	assert.Equal(t, "abc", rr.Body())
}
