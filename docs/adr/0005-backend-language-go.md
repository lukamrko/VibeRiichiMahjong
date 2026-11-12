# ADR-0005: Backend language — Go, standard library first

**Status:** Accepted (2026-11-12). Supersedes the "Backend: Undecided" line of ADR-0002.

## Context
Issue #1 asked for a choice of backend language for the scoring engine and public API. The candidates were Go, Python/FastAPI, TypeScript and .NET.

Several usual criteria don't apply here:
- **Performance.** Scoring a 14-tile hand takes microseconds in any language, and the image work runs in the vision service.
- **ML ecosystem.** Vision is a separate Python service (ADR-0004), so the backend doesn't need it.
- **Running the oracle in-process.** Fixtures are language-agnostic (`fixtures/hands/`), so the Python `mahjong` oracle can run as a separate script.

What actually decides it are the project's goals. The project studies AI-assisted development *and* is a chance to learn how other languages behave. The author's strongest language is C#, then Python.

## Decision

### Language
**Go**, chosen deliberately because learning Go is an explicit goal of the project.

- **C#** was rejected. It is the author's strongest language and would teach the least.
- **Python** was rejected. It largely duplicates the vision stack and is the second-strongest language, so again little new would be learned.
- **TypeScript** was not pursued. The frontend already covers it.

Version: the latest stable Go release when the backend is scaffolded, pinned in `go.mod`. Go upgrades are made on purpose in their own PRs.

### Dependencies: standard library first
- HTTP: `net/http` (method and path routing, Go 1.22+). No router framework.
- Tests: `testing` with table-driven tests. No assertion library such as testify.
- JSON: `encoding/json`. Logging: `log/slog`.
- Any third-party Go dependency needs a short written justification in the PR that adds it.

### Tooling
`gofmt`, `go vet`, `golangci-lint` (the one deliberate non-stdlib tool) and `go test ./...`.

### Oracle (Python `mahjong`)
- **Fixtures are written by hand** from rules references and worked examples. People are the authority, not the oracle. Fixtures are never generated from the oracle's output.
- Go tests read only the fixture files. `go test ./...` has no Python dependency.
- A **Python cross-check script** in `tools/oracle/` (with its own `requirements.txt`) checks every fixture against the oracle. It runs as a separate CI job and fails on any disagreement, unless the fixture carries a documented note that the oracle disagrees (for example, a ruleset difference). The fixture format and the script itself are built in #2.
- **Differential fuzzing** (random hands, Go engine vs oracle) is a planned follow-up: #10.

## Consequences
- The research gets a real test of agents writing idiomatic stdlib Go. It also records how well the author can review code in a language they are still learning. golangci-lint helps catch non-idiomatic code that might slip through review.
- **Known cost: Go has no sum types or real enums.** Domain types (`Tile`, suits, winds, yaku, meld/wait shapes) use typed constants and validating constructors. Code that switches over all cases needs care in review, since the compiler won't check that every case is covered (the `exhaustive` linter in golangci-lint can help).
- Assertions without testify are more verbose (`if got != want { t.Errorf(...) }`). Table-driven tests keep that manageable.
- Small static binary, so a small Docker image.
- Deferred to the issues that do the work:
  - package layout and the backend commands in `AGENTS.md`: #3 (first backend code)
  - backend CI job and the oracle CI job: #5
  - Docker image (lean: multi-stage build to distroless/static): #8
