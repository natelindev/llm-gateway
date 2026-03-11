package registry

import (
	"fmt"
	"sync"
	"time"

	"llm-gateway/internal/model"
)

// Registry is a thread-safe in-memory model registry.
// RWMutex enables concurrent reads and exclusive writes.
type Registry struct {
	mu     sync.RWMutex
	models map[string]*model.Model
}

// NewRegistry creates a new model registry.
func NewRegistry() *Registry {
	return &Registry{
		models: make(map[string]*model.Model),
	}
}

// Register registers a new model version.
// If the model does not exist, it is created.
// Returns an error when the same version already exists.
func (r *Registry) Register(req model.RegisterRequest) (*model.ModelVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get or create the model entry.
	m, exists := r.models[req.ModelName]
	if !exists {
		m = &model.Model{
			Name:     req.ModelName,
			Versions: make(map[string]*model.ModelVersion),
		}
		r.models[req.ModelName] = m
	}

	// Ensure version does not already exist.
	if _, vExists := m.Versions[req.Version]; vExists {
		return nil, fmt.Errorf("model %s version %s already exists", req.ModelName, req.Version)
	}

	// Create new version.
	now := time.Now()
	mv := &model.ModelVersion{
		Version:       req.Version,
		BackendType:   req.BackendType,
		Status:        model.StatusReady, // Ready immediately after registration.
		IsMock:        req.IsMock,
		MaxConcurrent: req.MaxConcurrent,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	m.Versions[req.Version] = mv
	return mv, nil
}

// GetVersion returns a pointer snapshot of a specific model version.
// Only versions in ready status are available for new requests.
func (r *Registry) GetVersion(name, version string) (*model.ModelVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, exists := r.models[name]
	if !exists {
		return nil, fmt.Errorf("model %s does not exist", name)
	}

	mv, vExists := m.Versions[version]
	if !vExists {
		return nil, fmt.Errorf("model %s version %s does not exist", name, version)
	}

	if mv.Status != model.StatusReady {
		return nil, fmt.Errorf("model %s version %s is %s and unavailable", name, version, mv.Status)
	}

	return mv, nil
}

// GetModel returns all metadata and versions for a model.
func (r *Registry) GetModel(name string) (*model.Model, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, exists := r.models[name]
	if !exists {
		return nil, fmt.Errorf("model %s does not exist", name)
	}

	return m, nil
}

// ListAll returns all models.
func (r *Registry) ListAll() []*model.Model {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*model.Model, 0, len(r.models))
	for _, m := range r.models {
		result = append(result, m)
	}
	return result
}

// UpdateVersion hot-updates version configuration.
// Existing requests keep running with the snapshot they already hold;
// new requests use updated configuration.
func (r *Registry) UpdateVersion(name, version string, req model.UpdateRequest) (*model.ModelVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, exists := r.models[name]
	if !exists {
		return nil, fmt.Errorf("model %s does not exist", name)
	}

	mv, vExists := m.Versions[version]
	if !vExists {
		return nil, fmt.Errorf("model %s version %s does not exist", name, version)
	}

	if mv.Status == model.StatusDeleted {
		return nil, fmt.Errorf("model %s version %s is deleted and cannot be updated", name, version)
	}

	// Apply updates.
	if req.BackendType != "" {
		mv.BackendType = req.BackendType
	}
	mv.IsMock = req.IsMock
	if req.MaxConcurrent >= 0 {
		mv.MaxConcurrent = req.MaxConcurrent
	}
	mv.Status = model.StatusReady
	mv.UpdatedAt = time.Now()

	return mv, nil
}

// DeleteVersion soft-deletes a model version.
// Deleted versions reject new requests but do not interrupt active ones.
func (r *Registry) DeleteVersion(name, version string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, exists := r.models[name]
	if !exists {
		return fmt.Errorf("model %s does not exist", name)
	}

	mv, vExists := m.Versions[version]
	if !vExists {
		return fmt.Errorf("model %s version %s does not exist", name, version)
	}

	mv.Status = model.StatusDeleted
	mv.UpdatedAt = time.Now()
	return nil
}
