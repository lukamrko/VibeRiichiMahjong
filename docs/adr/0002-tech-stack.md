# ADR-0002: Tech stack

**Status:** Accepted (2026-11-12). The backend line is superseded by [ADR-0005](0005-backend-language-go.md) (Go).

## Decision
- **Frontend:** React + TypeScript + Vite SPA.
- **Vision:** Python, fine-tuned YOLO object detector, served as its own HTTP service.
- **Backend:** ~~Undecided.~~ **Go**, standard library first. See ADR-0005.
- **Deployment:** Docker images per component, `docker compose`; hosting target decided in M4.

## Consequences
Because vision is a separate service, the backend language choice is independent of the ML stack.
If the backend ends up not being Python, the oracle library (Python `mahjong`) is used via fixture generation/cross-check scripts rather than in-process.
