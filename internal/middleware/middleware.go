package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const TraceIDKey = "trace_id"

// TraceID assigns a unique trace_id to each request.
// The trace_id is injected into Gin context and response headers for correlation.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		c.Set(TraceIDKey, traceID)
		c.Header("X-Trace-ID", traceID)
		c.Next()
	}
}

// Logger records request method, path, status code, and latency.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		traceID, _ := c.Get(TraceIDKey)
		log.Printf("[%s] %s %s %d %v",
			traceID, c.Request.Method, path, c.Writer.Status(), latency)
	}
}

// Recovery catches panics and returns HTTP 500 instead of crashing the process.
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

// GetTraceID returns trace_id from Gin context.
func GetTraceID(c *gin.Context) string {
	traceID, exists := c.Get(TraceIDKey)
	if !exists {
		return ""
	}
	return traceID.(string)
}
