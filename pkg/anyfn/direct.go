package anyfn

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DirectResponse represents the direct response.
type DirectResponse struct {
	Code  int
	Error error
}

func (d DirectResponse) Deal(c *gin.Context) {
	if d.Code == 0 {
		d.Code = http.StatusOK
	}

	errString := ""

	if d.Error != nil {
		errString = d.Error.Error()
	}

	c.String(d.Code, errString)
}
