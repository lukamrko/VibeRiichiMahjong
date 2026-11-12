# ADR-0004: Service topology

**Status:** Accepted (2026-11-12)

## Decision
Frontend → backend only. Backend calls the vision service over HTTP (image in → detections out).
The backend scoring API also accepts tiles directly, so it works without vision.

## Consequences
Three containers. Vision can be developed, stubbed and tested independently. One extra network hop for image requests.
