package model

import (
	"sync/atomic"
	"time"
)

// VersionStatus represents model version lifecycle status.
type VersionStatus string

const (
	StatusLoading    VersionStatus = "loading"    // Loading.
	StatusReady      VersionStatus = "ready"      // Ready for requests.
	StatusDeprecated VersionStatus = "deprecated" // Deprecated after update.
	StatusDeleted    VersionStatus = "deleted"    // Soft-deleted.
)

// BackendType enumerates backend drivers.
type BackendType string

const (
	BackendMock   BackendType = "mock"   // Mock backend.
	BackendOpenAI BackendType = "openai" // OpenAI API.
	BackendOllama BackendType = "ollama" // Local Ollama inference.
)

// ModelVersion stores per-version configuration and runtime state.
type ModelVersion struct {
	Version       string        `json:"version"`
	BackendType   BackendType   `json:"backend_type"`
	Status        VersionStatus `json:"status"`
	IsMock        bool          `json:"is_mock"`
	MaxConcurrent int           `json:"max_concurrent"` // Max concurrent requests; 0 means unlimited.
	ActiveConns   atomic.Int64  `json:"-"`              // Active requests (atomic counter).
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// ActiveConnCount returns current active connection count.
func (mv *ModelVersion) ActiveConnCount() int64 {
	return mv.ActiveConns.Load()
}

// Model contains model metadata and all versions.
type Model struct {
	Name     string                   `json:"name"`
	Versions map[string]*ModelVersion `json:"versions"`
}

// RegisterRequest is the model registration payload.
type RegisterRequest struct {
	ModelName     string      `json:"model_name" binding:"required"`
	Version       string      `json:"version" binding:"required"`
	BackendType   BackendType `json:"backend_type" binding:"required"`
	IsMock        bool        `json:"is_mock"`
	MaxConcurrent int         `json:"max_concurrent"` // Max concurrent requests; default 0 means unlimited.
}

// UpdateRequest is the model version hot-update payload.
type UpdateRequest struct {
	BackendType   BackendType `json:"backend_type"`
	IsMock        bool        `json:"is_mock"`
	MaxConcurrent int         `json:"max_concurrent"`
}

// InferRequest is the inference request payload.
type InferRequest struct {
	Model   string `json:"model" binding:"required"`
	Version string `json:"version" binding:"required"`
	Input   string `json:"input" binding:"required"`
}

// SSEEvent is one streamed SSE payload.
type SSEEvent struct {
	Token string `json:"token"`
	Done  bool   `json:"done"`
}

// ModelVersionView is the API view of a model version.
type ModelVersionView struct {
	Version       string        `json:"version"`
	BackendType   BackendType   `json:"backend_type"`
	Status        VersionStatus `json:"status"`
	IsMock        bool          `json:"is_mock"`
	MaxConcurrent int           `json:"max_concurrent"`
	ActiveConns   int64         `json:"active_conns"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// ToView converts ModelVersion to API view.
func (mv *ModelVersion) ToView() ModelVersionView {
	return ModelVersionView{
		Version:       mv.Version,
		BackendType:   mv.BackendType,
		Status:        mv.Status,
		IsMock:        mv.IsMock,
		MaxConcurrent: mv.MaxConcurrent,
		ActiveConns:   mv.ActiveConns.Load(),
		CreatedAt:     mv.CreatedAt,
		UpdatedAt:     mv.UpdatedAt,
	}
}

// ModelView is the API view of a model.
type ModelView struct {
	Name     string                      `json:"name"`
	Versions map[string]ModelVersionView `json:"versions"`
}

// ToView converts Model to API view.
func (m *Model) ToView() ModelView {
	view := ModelView{
		Name:     m.Name,
		Versions: make(map[string]ModelVersionView),
	}
	for k, v := range m.Versions {
		view.Versions[k] = v.ToView()
	}
	return view
}

// ErrorResponse is the unified error response payload.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id,omitempty"`
}
