package util

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// RequestContentType returns the content type of the request.
func RequestContentType(r *http.Request) string {
	if r == nil {
		return ""
	}
	return r.Header.Get("Content-Type")
}

// ResponseContentType returns the content type of the response.
func ResponseContentType(r *http.Response) string {
	if r == nil {
		return ""
	}
	return r.Header.Get("Content-Type")
}

// CaptureRequest captures the request from http.Request.
// read the request body and return the bytes.
// the request body will be restored after the function returns.
func CaptureRequest(req *http.Request) []byte {
	if req == nil || req.Body == nil {
		return nil
	}

	bs, err := io.ReadAll(req.Body)
	if err != nil {
		return nil
	}

	_ = req.Body.Close()

	bsCopy := bytes.Clone(bs)
	defer func() {
		req.Body = io.NopCloser(bytes.NewBuffer(bs))
	}()

	return bsCopy
}

// CaptureResponse captures the response from http.Response.
// read the response body and return the bytes.
// the response body will be restored after the function returns.
func CaptureResponse(resp *http.Response) []byte {
	if resp == nil || resp.Body == nil {
		return nil
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	_ = resp.Body.Close()

	bsCopy := bytes.Clone(bs)
	defer func() {
		resp.Body = io.NopCloser(bytes.NewBuffer(bs))
	}()

	return bsCopy
}

// ReadableHttpBody returns the readable http body.
// if the content type is not json, xml, form, html, return nil.
func ReadableHttpBody(contentType string, body []byte) any {
	if contentType == "" {
		return nil
	}

	if len(body) == 0 {
		return nil
	}

	contentType = strings.ToLower(contentType)

	if !(strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "application/x-www-form-urlencoded") ||
		strings.HasPrefix(contentType, "text/xml") ||
		strings.HasPrefix(contentType, "text/html")) {
		// if not json, xml, form, html, return nil
		return nil
	}

	if strings.HasPrefix(contentType, "application/json") {
		var reqMap map[string]any
		if err := json.Unmarshal(body, &reqMap); err != nil {
			return nil
		}

		return reqMap
	}

	s := string(body)

	return If(len(s) > 256, s[:256]+"...", s)
}
