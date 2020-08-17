package anyfunc

import "net/http"

// HTTPStatus defines the type of HTTP state.
type HTTPStatus int

func findStateCode(vs []interface{}) (int, []interface{}) {
	code := http.StatusOK
	vvs := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		if vv, ok := v.(HTTPStatus); ok {
			code = int(vv)
		} else {
			vvs = append(vvs, v)
		}
	}

	return code, vvs
}
