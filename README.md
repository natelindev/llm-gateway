# LLM Gateway

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](go.mod)

A production-minded, lightweight multi-model inference gateway written in Go.

LLM Gateway focuses on API stability and operational safety: model/version registration at runtime, SSE streaming, hot updates without interrupting in-flight requests, per-version concurrency controls, and Prometheus metrics.

## Why This Project

- Keep inference access behind one stable API.
- Roll out model/version updates safely while traffic is running.
- Observe throughput and latency with Prometheus-ready metrics.
- Start local in minutes with a built-in `mock` backend.

## Features

- Register and manage multiple models and versions at runtime.
- Stream inference output over SSE (`POST /infer`).
- Hot-update model version configs without breaking in-flight requests.
- Enforce per-version concurrency limits with `429` responses.
- Integrate multiple backends (`mock`, `openai`; `ollama` is reserved as a placeholder).
- Export metrics at `/metrics`.

## Tech Stack

- Go
- Gin
- SSE (Server-Sent Events)
- Prometheus client
- `sync.RWMutex` + `atomic.Int64`

## Project Structure

```text
llm-gateway/
├── main.go
├── internal/
│   ├── model/model.go
│   ├── registry/registry.go
│   ├── backend/
│   │   ├── backend.go
│   │   ├── mock.go
│   │   └── openai.go
│   ├── handler/
│   │   ├── model_handler.go
│   │   └── infer_handler.go
│   ├── router/router.go
│   ├── middleware/middleware.go
│   └── metrics/metrics.go
├── CLAUDE.md
├── AGENTS.md
└── .ai/
```

## Quick Start

```bash
go mod tidy
go run main.go
```

Default port is `8080`. Override with `PORT`:

```bash
PORT=9090 go run main.go
```

### Optional: OpenAI Backend

```bash
export OPENAI_API_KEY="sk-xxx"
export OPENAI_BASE_URL="https://api.openai.com/v1"  # optional
export OPENAI_MODEL="gpt-4o-mini"                    # optional
go run main.go
```

## API Examples

### Register a Model Version

```bash
curl -X POST http://localhost:8080/models \
  -H "Content-Type: application/json" \
  -d '{
    "model_name": "chat-bot",
    "version": "v1",
    "backend_type": "mock",
    "is_mock": true,
    "max_concurrent": 10
  }'
```

### List Models

```bash
curl http://localhost:8080/models | jq
```

### Stream Inference (SSE)

```bash
curl -N -X POST http://localhost:8080/infer \
  -H "Content-Type: application/json" \
  -d '{
    "model": "chat-bot",
    "version": "v1",
    "input": "Tell me a joke"
  }'
```

Example event stream:

```text
data: {"token":"This","done":false}

data: {"token":" is","done":false}

...

data: {"token":"","done":true}
```

### Hot-Update a Model Version

```bash
curl -X PUT http://localhost:8080/models/chat-bot/version/v1 \
  -H "Content-Type: application/json" \
  -d '{
    "backend_type": "mock",
    "is_mock": true,
    "max_concurrent": 50
  }'
```

### Delete a Model Version (Soft Delete)

```bash
curl -X DELETE http://localhost:8080/models/chat-bot/version/v1
```

### Health and Metrics

```bash
curl http://localhost:8080/health
curl http://localhost:8080/metrics
```

## Run Tests

```bash
go test ./...
```

## Metrics

- `llm_gateway_infer_total` (counter: model/version/status)
- `llm_gateway_infer_duration_seconds` (histogram)
- `llm_gateway_infer_active` (gauge)
- `llm_gateway_models_registered` (gauge)

## Design Notes

- SSE is used because inference is a one-way server-to-client stream and is easy to debug with `curl -N`.
- The registry uses `sync.RWMutex` for safe concurrent access.
- Active inference counts use `atomic.Int64` for lock-free counters.
- Each inference request works on a version pointer snapshot, so hot updates do not interrupt active streams.

## Open-Source Readiness

This repository includes:

- English-only docs and code comments/messages.
- AI agent guidance (`CLAUDE.md`, `AGENTS.md`, `.ai/`).
- Basic open-source hygiene (`.gitignore`, `LICENSE`).

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

By participating, you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## Security

Please report vulnerabilities responsibly as described in [SECURITY.md](SECURITY.md).
