# Skill: Add a New Inference Backend

Use this workflow when implementing a new provider backend.

## Goal

Add a new backend implementation while preserving streaming behavior and API contracts.

## Steps

1. Add a new `BackendType` constant in `internal/model/model.go`.
2. Implement `StreamInfer(ctx, input, ch)` in `internal/backend/<provider>.go`.
3. Ensure the implementation:
   - closes output channel exactly once,
   - forwards tokens incrementally,
   - stops promptly on `ctx.Done()`,
   - returns descriptive errors.
4. Register the backend in `internal/backend/backend.go` factory.
5. Keep handler logic unchanged unless required for provider-specific options.
6. Update `README.md` with env vars and usage example.

## Validation

1. `go test ./...`
2. `go run main.go`
3. Register model with the new backend and run `/infer` end-to-end.
