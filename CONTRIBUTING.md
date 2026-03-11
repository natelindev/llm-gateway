## Contributing

Thanks for your interest in contributing to LLM Gateway.

### Development Setup

1. Fork and clone the repository.
2. Install Go `1.26+`.
3. Install dependencies and run tests:

```bash
go mod tidy
go test ./...
```

4. Run the service locally:

```bash
go run main.go
```

### Pull Request Guidelines

- Keep changes focused and minimal.
- Preserve existing API behavior unless discussed in an issue.
- Keep user-facing docs, comments, and messages in English.
- Add or update tests for behavior changes when possible.
- Ensure `go test ./...` passes before opening a PR.

### Commit Messages

Use concise messages that explain why the change exists.

Example:

```text
fix infer handler trace_id propagation for 429 responses
```

### Local Verification Checklist

Before submitting a PR, run:

```bash
go test ./...
go run main.go
```

Then verify:

- `GET /health`
- `GET /metrics`
- One end-to-end SSE call via `POST /infer`

### Reporting Issues

If you find a bug or have a feature request, open an issue with:

- Clear reproduction steps
- Expected vs actual behavior
- Environment details
