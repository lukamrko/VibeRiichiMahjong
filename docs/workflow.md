# Workflow

## Planning on GitHub

- **Issues** are the unit of work. Every change starts from an issue.
- **Labels**
  - area: `area:backend`, `area:frontend`, `area:vision`, `area:infra`, `area:docs`
  - type: `type:feature`, `type:decision`, `type:research`, `type:bug`, `type:chore`
- **Milestones**: `M1 Backend`, `M2 Frontend`, `M3 Vision`, `M4 Deploy`.
- **Project board**: [VibeRiichiMahjong](https://github.com/users/lukamrko/projects/1) (Todo → In progress → Done). Add every new issue to it.
- `type:decision` issues end in an ADR under `docs/adr/`.

## Branches and PRs

1. Pick an issue. Create branch `<issue#>-<short-slug>` from `main`.
2. Write failing tests first, then implement.
3. Open a PR with `Closes #<issue#>` in the body; describe what changed and how it was tested.
4. CI must be green. Squash merge. Delete the branch.

`main` is protected: no direct pushes.

## Testing expectations

- Test-first (red → green → refactor) for all code.
- Unit tests for every module; integration tests for every boundary (HTTP API, backend ↔ vision, frontend ↔ backend).
- Scoring: every case in `fixtures/hands/` must pass. New yaku/fu rules add fixtures in the same PR.
- An oracle library may cross-check results in tests, never in production code.

## ADRs

`docs/adr/NNNN-title.md`, format: Context / Decision / Consequences / Status.
To change a decision, add a new ADR that supersedes the old one.
