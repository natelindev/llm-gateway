# CLAUDE.md

This file provides working context and guardrails for Claude-like coding agents in this repository.

## Project Overview

- Name: `llm-gateway`
- Language: Go
- Runtime: Gin HTTP server
- Purpose: multi-model inference gateway with SSE streaming, runtime model registration, and version hot updates.

## Architecture

- `main.go`: process entrypoint and server startup.
- `internal/router`: route wiring and middleware setup.
- `internal/handler`: HTTP handlers (`/models`, `/infer`).
- `internal/registry`: thread-safe in-memory model/version registry.
- `internal/backend`: backend abstraction + implementations (`mock`, `openai`).
- `internal/metrics`: Prometheus metrics definitions.
- `internal/middleware`: trace ID, logging, and recovery middleware.
- `internal/model`: request/response and domain models.

## Core Behavior to Preserve

1. Inference response is streamed over SSE.
2. Hot updates affect new requests only and do not interrupt active streams.
3. Per-version concurrency limits are enforced.
4. Error responses include `trace_id` when available.
5. Prometheus metrics remain stable and backward compatible.

## Local Development Commands

```bash
go mod tidy
go run main.go
go test ./...
```

## Environment Variables

- `PORT` (default: `8080`)
- `OPENAI_API_KEY` (required for `openai` backend)
- `OPENAI_BASE_URL` (optional)
- `OPENAI_MODEL` (optional)

## Agent Rules

- Prefer small, focused changes.
- Keep code and comments in English.
- Keep API behavior backward compatible unless asked to change it.
- Avoid introducing global mutable state unless necessary.
- Preserve concurrency safety (`RWMutex`, `atomic`) when touching registry/runtime state.
- If adding a new backend, implement `backend.Backend` and wire in `NewBackend` factory.

## Validation Checklist

Before finishing significant changes:

1. `go test ./...`
2. Start service with `go run main.go`
3. Verify `/health` and `/metrics`
4. Verify one `/infer` SSE run with a mock model
