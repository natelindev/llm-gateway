package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"llm-gateway/internal/metrics"
	"llm-gateway/internal/middleware"
	"llm-gateway/internal/model"
	"llm-gateway/internal/registry"
)

// ModelHandler serves model management endpoints.
type ModelHandler struct {
	reg *registry.Registry
}

func NewModelHandler(reg *registry.Registry) *ModelHandler {
	return &ModelHandler{reg: reg}
}

// Register creates a new model version.
// POST /models
// Body: { "model_name": "chat-bot", "version": "v1", "backend_type": "mock", "is_mock": true, "max_concurrent": 10 }
func (h *ModelHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    400,
			Message: "invalid request payload: " + err.Error(),
			TraceID: middleware.GetTraceID(c),
		})
		return
	}

	mv, err := h.reg.Register(req)
	if err != nil {
		c.JSON(http.StatusConflict, model.ErrorResponse{
			Code:    409,
			Message: err.Error(),
			TraceID: middleware.GetTraceID(c),
		})
		return
	}

	// Update Prometheus gauge.
	metrics.ModelRegistered.Inc()

	c.JSON(http.StatusCreated, gin.H{
		"message": "model registered successfully",
		"data":    mv.ToView(),
	})
}

// List returns all models and version states.
// GET /models
func (h *ModelHandler) List(c *gin.Context) {
	models := h.reg.ListAll()
	views := make([]model.ModelView, 0, len(models))
	for _, m := range models {
		views = append(views, m.ToView())
	}
	c.JSON(http.StatusOK, gin.H{
		"data": views,
	})
}

// Get returns details for a model.
// GET /models/:name
func (h *ModelHandler) Get(c *gin.Context) {
	name := c.Param("name")
	m, err := h.reg.GetModel(name)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code:    404,
			Message: err.Error(),
			TraceID: middleware.GetTraceID(c),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": m.ToView(),
	})
}

// Update hot-updates a model version.
// PUT /models/:name/version/:v
// New requests use updated configuration immediately; existing requests keep
// their current snapshot.
func (h *ModelHandler) Update(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("v")

	var req model.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    400,
			Message: "invalid request payload: " + err.Error(),
			TraceID: middleware.GetTraceID(c),
		})
		return
	}

	mv, err := h.reg.UpdateVersion(name, version, req)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code:    404,
			Message: err.Error(),
			TraceID: middleware.GetTraceID(c),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "model version updated successfully",
		"data":    mv.ToView(),
	})
}

// Delete soft-deletes a model version.
// DELETE /models/:name/version/:v
func (h *ModelHandler) Delete(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("v")

	if err := h.reg.DeleteVersion(name, version); err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code:    404,
			Message: err.Error(),
			TraceID: middleware.GetTraceID(c),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "model version deleted",
	})
}
