# AGENTS.md

Repository-level guidance for AI coding agents.

## Goal

Keep `llm-gateway` stable, readable, and safe while making incremental improvements.

## Working Principles

- Make minimal, focused changes.
- Prefer backward-compatible API behavior.
- Preserve SSE streaming semantics and hot-update isolation.
- Keep all user-facing docs/comments/messages in English.
- Avoid broad refactors unless explicitly requested.

## Code Areas and Ownership Intent

- `internal/handler`: request validation, response format, streaming behavior.
- `internal/registry`: model lifecycle state and concurrency safety.
- `internal/backend`: provider integrations and token streaming.
- `internal/metrics`: metric names/labels are part of the public contract.

## Must-Not-Break Contracts

1. `/infer` streams SSE events (`data: ...`) and sends a terminal `done=true` event.
2. `trace_id` is returned in structured error payloads when available.
3. Hot updates affect only new requests.
4. Per-version concurrency checks return `429` when saturated.
5. Existing metric names and labels remain unchanged.

## Validation Before Finishing

```bash
go test ./...
go run main.go
```

Then verify:

- `GET /health`
- `GET /metrics`
- One end-to-end SSE call via `POST /infer`

## Security and OSS Hygiene

- Never commit secrets (`OPENAI_API_KEY`, `.env`, tokens).
- Keep examples using placeholder keys.
- Keep docs concise and runnable.
