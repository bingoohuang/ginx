package hlog_test

import (
	"testing"

	"github.com/bingoohuang/ginx/pkg/adapt"
	"github.com/bingoohuang/ginx/pkg/anyfn"
	"github.com/bingoohuang/ginx/pkg/gintest"
	"github.com/bingoohuang/ginx/pkg/hlog"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	af := anyfn.NewAdapter()
	hf := hlog.NewAdapter()
	r := adapt.Adapt(gin.New(), af, hf)

	r.POST("/hello", af.F(func() string {
		return "Hello man!"
	}), hf.F(hf.Biz("你好啊")))

	r.POST("/world", af.F(func() string {
		return "Hello man!"
	}))

	r.POST("/bye", af.F(func() string {
		return "Hello man!"
	}), hf.F(hf.Ignore(true)))

	// r.Run(":8080")

	rr := gintest.Post("/hello", r)
	assert.Equal(t, "Hello man!", rr.Body())

	gintest.Post("/world", r)
	gintest.Post("/bye", r)
}
