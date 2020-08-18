package anyfn_test

import (
	"errors"
	"testing"

	"github.com/bingoohuang/ginx/pkg/adapt"
	"github.com/bingoohuang/ginx/pkg/anyfn"
	"github.com/bingoohuang/ginx/pkg/gintest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDirect(t *testing.T) {
	r := adapt.Adapt(gin.New())
	r.RegisterAdapter(anyfn.NewAdapter())

	r.GET("/direct1", anyfn.F(func() interface{} {
		return anyfn.DirectResponse{Code: 203}
	}))
	r.GET("/direct2", anyfn.F(func() interface{} {
		return &anyfn.DirectResponse{Code: 201, Error: errors.New("abc")}
	}))
	r.GET("/direct3", anyfn.F(func() interface{} {
		return &anyfn.DirectResponse{String: "ABC"}
	}))
	r.GET("/direct4", anyfn.F(func() interface{} {
		return &anyfn.DirectResponse{
			JSON: struct {
				Name string `json:"name"`
			}{
				Name: "ABC",
			},
			Header: map[string]string{"Xx-Server": "DDD"},
		}
	}))

	rr := gintest.Get("/direct1", r)
	assert.Equal(t, 203, rr.StatusCode())
	assert.Equal(t, "", rr.Body())

	rr = gintest.Get("/direct2", r)
	assert.Equal(t, 201, rr.StatusCode())
	assert.Equal(t, "abc", rr.Body())

	rr = gintest.Get("/direct3", r)
	assert.Equal(t, 200, rr.StatusCode())
	assert.Equal(t, `ABC`, rr.Body())

	rr = gintest.Get("/direct4", r)
	assert.Equal(t, 200, rr.StatusCode())
	assert.Equal(t, `{"name":"ABC"}`, rr.Body())
	assert.Equal(t, `DDD`, rr.Header()["Xx-Server"][0])
}
