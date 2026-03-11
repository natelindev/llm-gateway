package backend

import (
	"context"
	"fmt"

	"llm-gateway/internal/model"
)

// Backend is the unified inference backend interface.
// All backends (mock/openai/ollama) implement this interface so callers are
// decoupled from backend-specific details.
// StreamInfer sends generated tokens to the channel; the caller forwards them.
type Backend interface {
	StreamInfer(ctx context.Context, input string, ch chan<- string) error
}

// NewBackend creates a backend instance based on backend_type.
func NewBackend(mv *model.ModelVersion) (Backend, error) {
	switch mv.BackendType {
	case model.BackendMock:
		return NewMockBackend(), nil
	case model.BackendOpenAI:
		return NewOpenAIBackend(), nil
	case model.BackendOllama:
		return nil, fmt.Errorf("ollama backend is not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported backend type: %s", mv.BackendType)
	}
}
