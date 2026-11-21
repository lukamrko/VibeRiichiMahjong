# backend

Go scoring engine and public HTTP API (standard library only, ADR-0005). See [AGENTS.md](../AGENTS.md) for
commands and [docs/architecture.md](../docs/architecture.md) for the `POST /score` contract.

```bash
go test ./...                     # unit tests + every case in fixtures/hands/ through the HTTP API
go run ./cmd/server -addr :8080   # then POST a fixture's "input" to http://localhost:8080/score
```

| Package | Role |
|---|---|
| `internal/tile` | Tile kinds, red fives, MPSZ parsing |
| `internal/hand` | 14-tile hand validation, decompositions, wait types |
| `internal/score` | `WinContext` validation, yaku, fu, han, limit hands, payments |
| `internal/api` | JSON request/response mapping and error codes |
| `cmd/server` | HTTP server entrypoint |
