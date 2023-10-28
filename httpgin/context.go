package httpgin

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strings"
)

// RequestContext returns the value associated with this context for key, or nil
func RequestContext(c *gin.Context) context.Context {
	if c == nil {
		return context.Background()
	}
	return c.Request.Context()
}

// WithContext returns a copy of parent in which the value associated with key is val.
func WithContext(c *gin.Context, ctx context.Context) {
	if c == nil {
		return
	}
	c.Request = c.Request.WithContext(ctx)
}

// newUUID generate a uuid string
func newUUID() string {
	return uuid.NewString()
}

// traceID generate a trace id
//
//	uuid string, remove the '-' in the uuid string
func traceID() string {
	return strings.ReplaceAll(newUUID(), "-", "")
}
