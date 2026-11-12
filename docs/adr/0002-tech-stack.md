# ADR-0002: Tech stack

**Status:** Partially accepted (2026-11-12)

## Decision
- **Frontend:** React + TypeScript + Vite SPA.
- **Vision:** Python, fine-tuned YOLO object detector, served as its own HTTP service.
- **Backend:** **Undecided.** Go is the leading candidate; to be decided in a dedicated `type:decision` issue, recorded as a new ADR.
- **Deployment:** Docker images per component, `docker compose`; hosting target decided in M4.

## Consequences
Because vision is a separate service, the backend language choice is independent of the ML stack.
If the backend ends up not being Python, the oracle library (Python `mahjong`) is used via fixture generation/cross-check scripts rather than in-process.
