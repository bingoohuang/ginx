package anyfunc

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

func singlePrimitiveValue(c *gin.Context, argIns []argIn) string {
	if countPrimitiveArgs(argIns) != 1 { // nolint:gomnd
		return ""
	}

	if len(c.Params) == 1 { // nolint:gomnd
		return c.Params[0].Value
	}

	q := c.Request.URL.Query()
	if len(q) == 1 { // nolint:gomnd
		for _, v := range q {
			return v[0]
		}
	}

	return ""
}

func countPrimitiveArgs(argIns []argIn) int {
	primitiveArgsNum := 0

	for i, arg := range argIns {
		switch arg.Kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Bool,
			reflect.String,
			reflect.Float32, reflect.Float64:
			argIns[i].PrimitiveIndex = primitiveArgsNum
			primitiveArgsNum++
		}
	}

	return primitiveArgsNum
}
