package httpgin

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/tenz-io/trackingo/logger"
	"github.com/tenz-io/trackingo/monitor"
	"gopkg.in/natefinch/lumberjack.v2"
	syslog "log"
	"net/http"
	"runtime/debug"
	"strings"
)

var (
	buildInMiddlewares = []ginFunc{
		applyAccessLog,
		applyTracking,
		applyTraffic,
		applyMetrics,
		applyTimeout,
		applyPanicRecovery,
	}
)

func applyAccessLog(cfg *Config) gin.HandlerFunc {
	if !cfg.EnableAccess {
		return func(context *gin.Context) {
			context.Next()
		}
	}

	if cfg.AccessLogbase == "" {
		cfg.AccessLogbase = "log"
	}

	filename := strings.Join([]string{cfg.AccessLogbase, "access.log"}, "/")
	syslog.Println("[httpgin] apply access log:", filename)

	accessLogger := &lumberjack.Logger{
		Filename:   filename,
		LocalTime:  true,
		MaxSize:    10,   // the maximum size of each log file (in megabytes)
		MaxBackups: 5,    // the maximum number of old log files to retain
		MaxAge:     30,   // the maximum number of days to retain old log files
		Compress:   true, // compress old log files with gzip
	}

	return gin.LoggerWithWriter(accessLogger)
}

func applyMetrics(cfg *Config) gin.HandlerFunc {
	if !cfg.EnableMetrics {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	syslog.Println("[httpgin] apply metrics")

	return func(c *gin.Context) {
		// get context from gin
		ctx := RequestContext(c)
		rec := monitor.BeginRecord(ctx, "total")
		defer func() {
			httpStatus := c.Writer.Status()
			rec.EndWithCode(httpStatus)
		}()

		c.Next()
	}
}

func applyTimeout(cfg *Config) gin.HandlerFunc {
	if cfg.Timeout <= 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	syslog.Println("[httpgin] apply timeout:", cfg.Timeout)

	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(RequestContext(c), cfg.Timeout)
		defer cancel()

		doneC := make(chan struct{})
		go func() {
			c.Next()
			close(doneC)
		}()

		select {
		case <-ctx.Done():
			c.AbortWithStatus(http.StatusRequestTimeout)
			return
		case <-doneC:
			// The request completed before the timeout
		}
	}
}

func applyPanicRecovery(cfg *Config) gin.HandlerFunc {
	syslog.Println("[httpgin] apply panic recover")

	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				syslog.Printf("panic recovery: %s, stacktrace: %s\n", r, string(debug.Stack()))
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next()
	}

	//return gin.Recovery()
}

func applyTracking(cfg *Config) gin.HandlerFunc {
	syslog.Println("[httpgin] apply tracking")

	return func(c *gin.Context) {
		url := c.Request.URL.Path
		ctx := RequestContext(c)

		// metrics
		ctx = monitor.InitSingleFlight(ctx, url)

		traceId := traceID()
		le := logger.WithFields(logger.Fields{
			"url": url,
		}).WithTracing(traceId)

		ctx = logger.WithLogger(ctx, le)

		te := logger.WithTrafficTracing(ctx, traceId).
			WithFields(logger.Fields{
				"url": url,
			}).
			WithIgnores(
				"password",
				"Authorization",
			)
		ctx = logger.WithTrafficEntry(ctx, te)

		WithContext(c, ctx)

		c.Next()
	}
}
