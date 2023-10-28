package httpgin

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/tenz-io/trackingo/logger"
	"io"
	syslog "log"
	"strings"
	"time"
)

func applyTraffic(cfg *Config) gin.HandlerFunc {
	if !cfg.EnableTraffic {
		return func(context *gin.Context) {
			context.Next()
		}
	}

	syslog.Println("[middleware] apply traffic logging")

	return func(c *gin.Context) {
		ctx := RequestContext(c)
		begin := time.Now()

		reqCopy := captureRequest(c)

		// hijack response writer
		rw := &responseWrapper{c.Writer, bytes.NewBuffer([]byte{})}
		c.Writer = rw

		defer func() {
			c.Writer = rw.ResponseWriter

			logger.TrafficEntryFromContext(ctx).
				DataWith(&logger.Traffic{
					Typ:  logger.TrafficTypAccess,
					Cmd:  "gateway",
					Code: c.Writer.Status(),
					Cost: time.Since(begin),
					Req:  reqCopy,
					Resp: captureResponse(c, rw.buffer.Bytes()),
				}, logger.Fields{
					"method":      c.Request.Method,
					"client":      c.ClientIP(),
					"query":       c.Request.URL.Query(),
					"req_header":  c.Request.Header,
					"req_size":    c.Request.ContentLength,
					"resp_header": c.Writer.Header(),
					"resp_size":   c.Writer.Size(),
				})
		}()

		c.Next()
	}
}

// capture http body from gin context request
// input is gin.Context
// output is any
// when context-type is application/json, return map[string]any
// when context-type is application/x-www-form-urlencoded, return map[string]string
// the other case, return nil
func captureRequest(c *gin.Context) (res any) {
	var body []byte
	var err error
	contentType := strings.ToLower(c.ContentType())

	ctx := c.Request.Context()
	le := logger.FromContext(ctx).WithFields(logger.Fields{
		"Content-Type": contentType,
	})
	defer func() {
		le.WithError(err).
			Debug("capture request")
	}()

	if strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "text/xml") ||
		strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		body, err = io.ReadAll(c.Request.Body)
		if err != nil {
			return nil
		}

		// clone body for reset body
		bs := bytes.Clone(body)
		defer func() {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bs))
		}()
	}

	if len(body) == 0 {
		return nil
	}

	if strings.HasPrefix(contentType, "application/json") {
		var req map[string]any
		if err = json.Unmarshal(body, &req); err != nil {
			return nil
		}

		return req
	}

	return string(body)
}

// captureResponse capture response from gin context writer
// input is gin.Context
// output is any
// when response writer context-type is application/json, return map[string]any
func captureResponse(c *gin.Context, bs []byte) (res any) {
	var (
		err         error
		contentType string
		ctx         = c.Request.Context()
		le          = logger.FromContext(ctx)
	)

	defer func() {
		le.WithError(err).
			WithFields(logger.Fields{
				"Content-Type": contentType,
			}).Debug("capture response")
	}()

	if len(bs) == 0 {
		return nil
	}

	if c.Writer == nil {
		return "<nil writer>"
	}

	contentType = strings.ToLower(c.Writer.Header().Get("Content-Type"))

	if strings.HasPrefix(contentType, "application/json") {
		var resp map[string]any
		if err = json.Unmarshal(bs, &resp); err != nil {
			return nil
		}
		return resp
	} else if strings.HasPrefix(contentType, "text/plain") ||
		strings.HasPrefix(contentType, "text/xml") {
		return string(bs)
	} else {
		return "<unsupported capture content-type>"
	}
}

type responseWrapper struct {
	gin.ResponseWriter
	buffer *bytes.Buffer
}

func (rw *responseWrapper) Write(data []byte) (int, error) {
	// Capture the response body
	written, err := rw.ResponseWriter.Write(data)
	rw.buffer.Write(data)
	return written, err
}
