package hlog

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"strings"
)

// At returns the element of index i in the slice s.
func At(s []string, i int) string {
	if i < len(s) {
		return s[i]
	}

	return ""
}

// Abbreviate abbreviates a string using ellipses.
func Abbreviate(str string, maxWidth int) string {
	size := len(str)
	if str == "" || maxWidth < 4 || size <= maxWidth {
		return str
	}

	return str[:maxWidth-3] + ("...")
}

// IPAddrFromRemoteAddr parses the IP Address.
// Request.RemoteAddress contains port, which we want to remove i.e.: "[::1]:58292" => "[::1]".
func IPAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}

	return s[:idx]
}

// GetRemoteAddress returns ip address of the client making the request, taking into account http proxies.
func GetRemoteAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")

	if hdrRealIP == "" && hdrForwardedFor == "" {
		return IPAddrFromRemoteAddr(r.RemoteAddr)
	}

	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}

		return parts[0]
	}

	return hdrRealIP
}

// IsWsRequest return true if this request is a websocket request.
func IsWsRequest(url string) bool {
	return strings.HasPrefix(url, "/ws/")
}

// PeekBody peeks the maxSize body from the request limit to maxSize bytes.
func PeekBody(r *http.Request, maxSize int) []byte {
	if r.Body == nil {
		return nil
	}

	buf := bufio.NewReader(r.Body)
	// And now set a new body, which will simulate the same rowsData we read:
	r.Body = ioutil.NopCloser(buf)

	// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
	// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
	// response body. A request body larger than that will now result in
	// Decode() returning a "http: request body too large" error.
	// r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	// Work / inspect body. You may even modify it!

	peek, _ := buf.Peek(maxSize)

	return peek
}
