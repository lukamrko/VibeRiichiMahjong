# AGENTS.md

Instructions for AI coding agents (Claude Code, Codex, Copilot, …) working in this repo.
This file is the single source of truth; `CLAUDE.md` just imports it.

## What this project is

**VibeRiichiMahjong** — photograph a Japanese Riichi Mahjong hand, recognize the tiles, and compute the score.

It is a **personal research project**: the main goal is to study how AI-assisted development flows work
end to end (planning → issues → implementation → review). It is not commercial. Prefer clarity,
traceability and good tests over cleverness or speed.

## Repository layout (monorepo)

| Path | What | Status |
|---|---|---|
| `backend/` | Scoring engine + public HTTP API. **Go**, standard library first (ADR-0005). | in progress |
| `frontend/` | React + TypeScript + Vite SPA: upload/camera, tile correction, game-context form, score breakdown. | not started |
| `vision/` | Python service: image in → detected tiles out (fine-tuned YOLO). | not started |
| `fixtures/hands/` | Language-agnostic test fixtures: hands with expected yaku/han/fu/points. Shared by all components. | active |
| `tools/oracle/` | Python script that cross-checks `fixtures/hands/` against the `mahjong` library (test oracle only). | active |
| `docs/` | Architecture, domain glossary, ADRs, workflow. | active |

Build order: **backend (scoring) → frontend → vision → deployment (Docker)**.

## Read before working

- [docs/architecture.md](docs/architecture.md) — components, data flow, API contracts.
- [docs/domain/glossary.md](docs/domain/glossary.md) — Riichi terms and **tile notation**. Use these terms exactly in code.
- [docs/domain/rules.md](docs/domain/rules.md) — the ruleset we implement and v1 scope.
- [docs/adr/](docs/adr/) — decisions already made. Do not re-litigate them; propose a new ADR that supersedes one instead.
- [docs/workflow.md](docs/workflow.md) — branching, PRs, issues, testing expectations.

## Non-negotiable rules

1. **Tile notation is MPSZ** everywhere (APIs, fixtures, logs, UI state): `1m`…`9m`, `1p`…`9p`, `1s`…`9s`,
   `1z`…`7z` (E S W N Haku Hatsu Chun), `0m`/`0p`/`0s` = red five. Hands as compact strings, e.g. `123m406p789s11z`.
2. **Ruleset is Tenhou / M-League style** (see `docs/domain/rules.md`). Don't silently implement other variants.
3. **v1 = closed hands only.** No pon/chi/open kan in v1. Closed kan is also out of scope for v1 unless an issue says otherwise.
4. **Test first.** Write a failing test before implementation (red → green → refactor). Every feature needs
   unit tests; every boundary (HTTP API, service-to-service) needs integration tests. Scoring changes must pass all of `fixtures/hands/`.
5. **We write the scoring engine ourselves.** Existing libraries (e.g. Python `mahjong`) may be used **only as a test oracle**, never as the implementation.
6. **One issue → one branch → one PR.** Branch `<issue#>-<short-slug>`, PR body contains `Closes #<issue#>`. Squash merge. Never commit directly to `main`.
7. Keep changes scoped to the issue. If you discover other work, note it (new issue) rather than doing it.
8. Update docs (glossary, architecture, ADRs) in the same PR when you change something they describe.

## Conventions

- Commit messages: Conventional Commits (`feat(backend): …`, `test(fixtures): …`, `docs: …`, `chore(ci): …`).
- Domain names in code follow the glossary (`Hand`, `Tile`, `Meld`, `Yaku`, `Han`, `Fu`, `WinContext`, …).
- Per-component build/test commands will be added here as each component is scaffolded.

### Go (backend)

- **Standard library first** (ADR-0005): `net/http` (no router framework), `testing` with table-driven tests
  (no testify or other assertion library), `encoding/json`, `log/slog`. Adding any third-party module needs a short
  justification in the PR description. Don't reach for gin/echo/chi/testify out of habit.
- Code must pass `gofmt`, `go vet` and `golangci-lint`. Tests run with `go test ./...`.
- Go has no sum types or real enums. Model domain kinds as typed constants with validating constructors, and make
  `switch` statements over them handle every case explicitly.
- Go tests read `fixtures/hands/` directly and never call Python. The Python `mahjong` oracle lives only in
  `tools/oracle/` as a separate cross-check.

## Commands

_Add `build`, `test`, `lint`, `run` commands per component here as each one is scaffolded._

- **Backend** (run from `backend/`; module `github.com/lukamrko/VibeRiichiMahjong/backend`, Go 1.25):
  - test: `go test ./...` (also reads `fixtures/hands/`)
  - format / vet: `gofmt -l .` (must print nothing), `go vet ./...`
  - lint: `golangci-lint run`
  - packages: `internal/tile` (Tile, Kind, Suit, MPSZ parsing), `internal/hand` (Hand validation,
    decompositions, wait types). Scoring (#4) builds on top of these.

- **Fixture oracle check** (run after any change to `fixtures/hands/`; see its README):
  `python -m venv tools/oracle/.venv && tools/oracle/.venv/Scripts/python -m pip install -r tools/oracle/requirements.txt`,
  then `tools/oracle/.venv/Scripts/python tools/oracle/check_fixtures.py` (use `bin/` instead of `Scripts/` on Linux/macOS).
