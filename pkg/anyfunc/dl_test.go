package anyfunc_test

import (
	"testing"

	"github.com/bingoohuang/ginx/pkg/adapt"
	"github.com/bingoohuang/ginx/pkg/anyfunc"
	"github.com/bingoohuang/ginx/pkg/gintest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDl(t *testing.T) {
	r := adapt.Adapt(gin.New())
	r.RegisterAdapter(anyfunc.NewAdapter())

	r.GET("/dl", anyfunc.Fn(func() anyfunc.DlFile {
		return anyfunc.DlFile{DiskFile: "testdata/hello.txt"}
	}))

	rr := gintest.Get("/dl", r)
	assert.Equal(t, 200, rr.StatusCode())
	assert.Equal(t, []string{"attachment; filename=hello.txt"}, rr.Header()["Content-Disposition"])
	assert.Equal(t, "Hello bingoohuang!", rr.Body())
}
