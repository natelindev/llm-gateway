package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"llm-gateway/internal/backend"
	"llm-gateway/internal/metrics"
	"llm-gateway/internal/middleware"
	"llm-gateway/internal/model"
	"llm-gateway/internal/registry"
)

// InferHandler serves inference endpoints.
type InferHandler struct {
	reg *registry.Registry
}

func NewInferHandler(reg *registry.Registry) *InferHandler {
	return &InferHandler{reg: reg}
}

// Infer streams inference responses.
// POST /infer
// Body: { "model": "chat-bot", "version": "v1", "input": "Tell me a joke" }
//
// Why SSE (Server-Sent Events):
// 1. It is an HTTP-native streaming protocol with no extra handshake.
// 2. It aligns with common LLM APIs (OpenAI, Claude).
// 3. It naturally supports one-way server-to-client token streaming.
// 4. It is easy to debug with curl (`curl -N`).
// 5. Browser EventSource supports auto-reconnect.
//
// Each token is sent as an SSE event:
//
//	data: {"token":"xxx","done":false}
//
// Final completion event:
//
//	data: {"token":"","done":true}
func (h *InferHandler) Infer(c *gin.Context) {
	var req model.InferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    400,
			Message: "invalid request payload: " + err.Error(),
			TraceID: middleware.GetTraceID(c),
		})
		return
	}

	start := time.Now()
	traceID := middleware.GetTraceID(c)

	// Read model version snapshot from registry.
	// Inference uses this snapshot, so hot-updates do not impact this request.
	mv, err := h.reg.GetVersion(req.Model, req.Version)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code:    404,
			Message: err.Error(),
			TraceID: traceID,
		})
		return
	}

	// Capacity control: enforce per-version max concurrency.
	if mv.MaxConcurrent > 0 {
		current := mv.ActiveConns.Load()
		if current >= int64(mv.MaxConcurrent) {
			metrics.InferTotal.WithLabelValues(req.Model, req.Version, "rejected").Inc()
			c.JSON(http.StatusTooManyRequests, model.ErrorResponse{
				Code:    429,
				Message: fmt.Sprintf("model %s version %s concurrency limit reached (%d/%d), please retry later", req.Model, req.Version, current, mv.MaxConcurrent),
				TraceID: traceID,
			})
			return
		}
	}

	// Increase active connection counters.
	mv.ActiveConns.Add(1)
	metrics.InferActive.WithLabelValues(req.Model, req.Version).Inc()
	defer func() {
		mv.ActiveConns.Add(-1)
		metrics.InferActive.WithLabelValues(req.Model, req.Version).Dec()
	}()

	// Create backend instance.
	b, err := backend.NewBackend(mv)
	if err != nil {
		metrics.InferTotal.WithLabelValues(req.Model, req.Version, "error").Inc()
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    500,
			Message: "failed to create backend: " + err.Error(),
			TraceID: traceID,
		})
		return
	}

	// Set SSE response headers.
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Trace-ID", traceID)

	// Create token channel and inference context.
	tokenCh := make(chan string, 10) // Buffered channel to avoid backend blocking.
	ctx := c.Request.Context()

	// Run inference in a goroutine.
	errCh := make(chan error, 1)
	go func() {
		errCh <- b.StreamInfer(ctx, req.Input, tokenCh)
	}()

	// Stream SSE events.
	c.Stream(func(w io.Writer) bool {
		select {
		case token, ok := <-tokenCh:
			if !ok {
				// Channel is closed; check inference result.
				if inferErr := <-errCh; inferErr != nil {
					log.Printf("[%s] inference error: %v", traceID, inferErr)
					// Send terminal error event.
					evt := model.SSEEvent{Token: "ERROR: " + inferErr.Error(), Done: true}
					data, _ := json.Marshal(evt)
					c.SSEvent("", string(data))
					metrics.InferTotal.WithLabelValues(req.Model, req.Version, "error").Inc()
				} else {
					// Send terminal completion event.
					evt := model.SSEEvent{Token: "", Done: true}
					data, _ := json.Marshal(evt)
					c.SSEvent("", string(data))
					metrics.InferTotal.WithLabelValues(req.Model, req.Version, "success").Inc()
				}
				// Record request latency.
				metrics.InferDuration.WithLabelValues(req.Model, req.Version).Observe(time.Since(start).Seconds())
				return false // Stop streaming.
			}
			// Send token event.
			evt := model.SSEEvent{Token: token, Done: false}
			data, _ := json.Marshal(evt)
			c.SSEvent("", string(data))
			return true // Continue streaming.

		case <-ctx.Done():
			// Client disconnected.
			log.Printf("[%s] client disconnected", traceID)
			metrics.InferTotal.WithLabelValues(req.Model, req.Version, "cancelled").Inc()
			metrics.InferDuration.WithLabelValues(req.Model, req.Version).Observe(time.Since(start).Seconds())
			return false
		}
	})
}
