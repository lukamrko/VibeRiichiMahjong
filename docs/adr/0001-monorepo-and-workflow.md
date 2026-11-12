# ADR-0001: Monorepo, GitHub issues, PR-per-issue workflow

**Status:** Accepted (2026-11-12)

## Context
Solo research project studying AI-assisted development flows. Needs a traceable record of planning and changes.

## Decision
- Single public GitHub monorepo `lukamrko/VibeRiichiMahjong` with `backend/`, `frontend/`, `vision/`, `fixtures/`, `docs/`.
- All planning in GitHub Issues + labels + milestones + a Projects board.
- One issue → one branch → one squash-merged PR; `main` protected.
- `AGENTS.md` is the canonical agent instruction file; `CLAUDE.md` imports it.
- MIT license.

## Consequences
Every change is traceable from issue to PR, which is the research record. Slightly slower than pushing to main.
