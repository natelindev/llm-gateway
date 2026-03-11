# Skill: Extend Inference API Safely

Use this workflow when modifying `/infer` behavior or payloads.

## Safety Constraints

- Preserve existing request fields unless a breaking change is explicitly requested.
- Preserve SSE event shape and terminal done event.
- Preserve `trace_id` handling in structured errors.

## Steps

1. Update request/response models in `internal/model/model.go`.
2. Update validation and runtime behavior in `internal/handler/infer_handler.go`.
3. Check metrics emission paths for success/error/cancel/reject status.
4. Update docs and curl examples in `README.md`.

## Validation

1. `go test ./...`
2. Manual SSE test: `curl -N -X POST http://localhost:8080/infer ...`
3. Verify terminal `done=true` event arrives in all non-cancel scenarios.
