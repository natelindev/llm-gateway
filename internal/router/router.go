package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"llm-gateway/internal/handler"
	"llm-gateway/internal/middleware"
	"llm-gateway/internal/registry"
)

// Setup wires routes and middleware.
// Route groups follow RESTful conventions:
//   - /models: model CRUD operations
//   - /infer: inference entrypoint
//   - /metrics: Prometheus endpoint
//   - /health: health check
func Setup(reg *registry.Registry) *gin.Engine {
	r := gin.New()

	// Middleware chain: trace_id -> logger -> recovery.
	r.Use(middleware.TraceID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// Health check endpoint.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Prometheus metrics endpoint.
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Initialize handlers.
	modelHandler := handler.NewModelHandler(reg)
	inferHandler := handler.NewInferHandler(reg)

	// Model management routes.
	r.POST("/models", modelHandler.Register)
	r.GET("/models", modelHandler.List)
	r.GET("/models/:name", modelHandler.Get)
	r.PUT("/models/:name/version/:v", modelHandler.Update)
	r.DELETE("/models/:name/version/:v", modelHandler.Delete)

	// Inference route.
	r.POST("/infer", inferHandler.Infer)

	return r
}
