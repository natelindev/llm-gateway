# Core Rules

These rules are mandatory for AI agents working in this repository.

1. Keep all code comments, logs, docs, and user-facing messages in English.
2. Preserve API compatibility unless explicitly asked to change behavior.
3. Do not break SSE contract on `/infer` (incremental `data:` events + terminal `done=true`).
4. Maintain hot-update isolation (new requests get new config; active streams continue).
5. Keep per-version concurrency protection intact (`429` on saturation).
6. Do not rename or remove existing Prometheus metric names/labels.
7. Avoid introducing secrets into code, docs, or examples.
8. Prefer small, scoped changes with clear rationale.
