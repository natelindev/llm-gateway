## What changed

-

## Why

-

## How to test

```bash
go test ./...
go run main.go
```

Then verify:

- `GET /health`
- `GET /metrics`
- One end-to-end SSE call via `POST /infer`

## Checklist

- [ ] I kept changes focused and backward-compatible where possible
- [ ] I added or updated tests when behavior changed
- [ ] I did not include secrets or credentials
- [ ] I updated docs if needed
