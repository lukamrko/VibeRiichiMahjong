# ADR-0003: Ruleset, tile notation, v1 scope

**Status:** Accepted (2026-11-12)

## Decision
- Ruleset: Tenhou / M-League style (aka dora, no kiriage mangan, kazoe yakuman). Details in `docs/domain/rules.md`.
- Tile notation: MPSZ strings across all components; `0` = red five.
- v1: closed hands only. Game context (winds, riichi, tsumo/ron, dora indicators, …) entered via a form, not detected.
- Scoring engine is written by us, test-first, against shared fixtures; an existing library is used only as a test oracle.

## Consequences
Simple, well-defined v1. Open melds and photo-based context detection are explicit later work.
