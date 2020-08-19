package anyfn_test

import (
	"testing"

	"github.com/bingoohuang/ginx/pkg/adapt"
	"github.com/bingoohuang/ginx/pkg/anyfn"
	"github.com/bingoohuang/ginx/pkg/gintest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDl(t *testing.T) {
	af := anyfn.NewAdapter()
	r := adapt.Adapt(gin.New(), af)

	r.GET("/dl", af.F(func() anyfn.DlFile {
		return anyfn.DlFile{DiskFile: "testdata/hello.txt"}
	}))

	rr := gintest.Get("/dl", r)
	assert.Equal(t, 200, rr.StatusCode())
	assert.Equal(t, []string{"attachment; filename=hello.txt"}, rr.Header()["Content-Disposition"])
	assert.Equal(t, "Hello bingoohuang!", rr.Body())
}
