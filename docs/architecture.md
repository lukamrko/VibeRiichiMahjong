# Architecture

## Components

```
┌──────────┐   image + context    ┌──────────┐    image     ┌──────────┐
│ frontend │ ───────────────────▶ │ backend  │ ───────────▶ │  vision  │
│ React/TS │ ◀─────────────────── │ Go       │ ◀─────────── │ Python   │
└──────────┘  tiles / score       └──────────┘  detections  └──────────┘
```

- **frontend** talks only to **backend**.
- **backend** (Go, standard library first; ADR-0005) owns the public API and the scoring engine. It calls **vision** internally.
- **vision** is a separate Python HTTP service so the backend language is free to choose (ADR-0002, ADR-0004).
- Each component ships as its own Docker image; local dev and deployment use `docker compose` (later milestone).

## User flow (v1)

1. User takes/uploads a photo of a **closed** 14-tile winning hand.
2. Backend forwards the image to vision → gets detected tiles (MPSZ id, confidence, bounding box).
3. Frontend shows detected tiles; user **corrects** misreads and marks the **winning tile**.
4. User fills the **win context** form: round wind, seat wind, riichi / double riichi / ippatsu, tsumo or ron,
   dora indicators (and ura dora if riichi), honba, plus situational yaku (haitei, houtei, rinshan, chankan, tenhou/chiihou).
5. Backend scores the hand → returns yaku list, han, fu (with breakdown), limit name, and points (dealer/non-dealer, tsumo split).

The scoring endpoint must work **without** vision (tiles typed in directly) — that is how M1 and M2 are built and tested.

## Contracts (draft — finalize in the relevant issues)

### Vision → backend
```json
{
  "tiles": [
    { "tile": "5m", "confidence": 0.97, "bbox": [x, y, w, h] }
  ]
}
```

### Score request (backend public API)
```json
{
  "hand": "123m406p789s11z",
  "winningTile": "6p",
  "context": {
    "roundWind": "1z", "seatWind": "2z",
    "winType": "ron",
    "riichi": true, "doubleRiichi": false, "ippatsu": false,
    "doraIndicators": ["3s"], "uraDoraIndicators": ["9m"],
    "honba": 0,
    "haitei": false, "houtei": false, "rinshan": false, "chankan": false, "tenhou": false, "chiihou": false
  }
}
```
`hand` contains all 14 tiles including the winning tile.

### Score response
```json
{
  "yaku": [{ "name": "riichi", "han": 1 }, { "name": "pinfu", "han": 1 }],
  "dora": 1, "akaDora": 1, "uraDora": 0,
  "han": 4, "fu": 30,
  "fuBreakdown": [{ "reason": "base", "fu": 20 }, { "reason": "closed ron", "fu": 10 }],
  "limit": null,
  "points": { "total": 7700, "fromDiscarder": 7700 }
}
```
When several decompositions of a hand are possible, the engine returns the **highest-scoring** one.

## Testing layers

- **Unit**: tile parsing, hand decomposition, each yaku, fu, point table.
- **Fixtures**: `fixtures/hands/*` — shared, language-agnostic cases (input → expected output). Run by backend tests; also cross-checked against an oracle library.
- **Integration**: backend HTTP API; backend ↔ vision (vision stubbed and real); frontend ↔ backend.
- **Vision**: model evaluation on a held-out set of own photos (precision/recall per tile class).
