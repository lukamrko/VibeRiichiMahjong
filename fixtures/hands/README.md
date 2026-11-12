# Hand fixtures

Language-agnostic scoring test cases shared by all components. Every scoring implementation (the Go backend,
the frontend's mocks, the oracle cross-check) reads these files directly.

## Files

One JSON file per category. Each file is an array of cases; `id` is unique across **all** files.

| File | Covers |
|---|---|
| `yaku.json` | Regular yaku, fu (waits, triplets, yakuhai pairs), kiriage-free boundaries, honba |
| `special-shapes.json` | Chiitoitsu (kokushi lives in `limits.json`) |
| `situational.json` | Double riichi, ippatsu, haitei, houtei, chankan |
| `dora.json` | Aka dora, including a red five as the winning tile |
| `limits.json` | Kazoe yakuman and every v1 yakuman, including double yakuman |

## Case format

```json
{
  "id": "riichi-pinfu-ron",
  "description": "Riichi + pinfu, ryanmen ron on 6s (78s wait). 2 han 30 fu.",
  "input": {
    "hand": "123m45699p345678s",
    "winningTile": "6s",
    "context": {
      "roundWind": "1z", "seatWind": "2z",
      "winType": "ron",
      "riichi": true, "doubleRiichi": false, "ippatsu": false,
      "doraIndicators": ["1z"], "uraDoraIndicators": ["7z"],
      "honba": 0,
      "haitei": false, "houtei": false, "rinshan": false, "chankan": false, "tenhou": false, "chiihou": false
    }
  },
  "expected": {
    "yaku": [{ "name": "riichi", "han": 1 }, { "name": "pinfu", "han": 1 }],
    "dora": 0, "akaDora": 0, "uraDora": 0,
    "han": 2, "fu": 30,
    "fuBreakdown": [{ "reason": "base", "fu": 20 }, { "reason": "closed_ron", "fu": 10 }],
    "limit": null,
    "points": { "total": 2000, "fromDiscarder": 2000 }
  }
}
```

`input` is exactly the score request and `expected` is exactly the score response from
[docs/architecture.md](../../docs/architecture.md), so a case can be POSTed to the API as is.

### `input`

- `hand`: all 14 tiles **including** the winning tile, compact MPSZ (`0` = red five).
- `winningTile`: MPSZ id of the tile that completed the hand. If it is a red five, write `0m`/`0p`/`0s`.
- `context`: **every field is always present**; there are no defaults.
  - `roundWind`, `seatWind`: `1z`–`4z`. `seatWind: "1z"` means the winner is the dealer.
  - `winType`: `ron` or `tsumo`.
  - `doubleRiichi` and `ippatsu` require `riichi: true`. `uraDoraIndicators` is non-empty only with riichi.
  - `doraIndicators` / `uraDoraIndicators`: indicator tiles (the dora is the *next* tile).
  - `rinshan` is always `false` in v1: it needs your own kan, and kan is out of scope.

### `expected`

- `yaku`: list of `{name, han}` using the ids below. Han is the closed value. Dora are **not** yaku entries.
- `dora`, `akaDora`, `uraDora`: number of dora of each kind.
- `han`: sum of yaku han plus all dora. For yakuman hands it is the yakuman han only (13 per yakuman, 26 per double).
- `fu`: rounded fu (up to the next 10; chiitoitsu is a fixed 25). `null` for yakuman hands.
- `fuBreakdown`: fu items **before** rounding, using the reasons below. Empty for yakuman hands.
  Pinfu tsumo is just `base` (20). Chiitoitsu is just `chiitoitsu` (25).
- `limit`: `null`, `mangan`, `haneman`, `baiman`, `sanbaiman`, `yakuman`, `kazoe_yakuman`, `double_yakuman`,
  `triple_yakuman`.
- `points`: all values include honba (+300 ron, +100 per payer on tsumo). Riichi sticks on the table are not
  included. The keys depend on the win:

| Win | Keys |
|---|---|
| Ron | `total`, `fromDiscarder` |
| Non-dealer tsumo | `total`, `fromDealer`, `fromNonDealer` (each of the two non-dealers pays this) |
| Dealer tsumo | `total`, `fromEachNonDealer` |

### Yaku ids

Romaji, snake_case. This is a closed list; add new ids here first.

| Id | Han | Id | Han |
|---|---|---|---|
| `riichi` | 1 | `sanankou` | 2 |
| `double_riichi` | 2 (replaces `riichi`) | `sanshoku_doukou` | 2 |
| `ippatsu` | 1 | `toitoi` | 2 |
| `menzen_tsumo` | 1 | `shousangen` | 2 |
| `pinfu` | 1 | `honroutou` | 2 |
| `tanyao` | 1 | `honitsu` | 3 |
| `iipeikou` | 1 | `chinitsu` | 6 |
| `yakuhai_haku` / `yakuhai_hatsu` / `yakuhai_chun` | 1 | `chiitoitsu` | 2 |
| `yakuhai_seat_wind` | 1 | `kokushi_musou` | 13 |
| `yakuhai_round_wind` | 1 (a double wind gives both entries) | `kokushi_musou_juusanmen` | 26 |
| `haitei` | 1 | `suuankou` | 13 |
| `houtei` | 1 | `suuankou_tanki` | 26 |
| `chankan` | 1 | `daisangen` | 13 |
| `sanshoku_doujun` | 2 | `shousuushii` | 13 |
| `ittsu` | 2 | `daisuushii` | 26 |
| `chanta` | 2 | `tsuuiisou` | 13 |
| `junchan` | 3 | `chinroutou` | 13 |
| `ryanpeikou` | 3 (replaces `iipeikou`) | `ryuuiisou` | 13 |
| | | `chuuren_poutou` | 13 |
| | | `junsei_chuuren_poutou` | 26 |
| | | `tenhou` / `chiihou` | 13 |

Out of scope in v1, so not in the list: rinshan kaihou, sankantsu and suukantsu (need kan), renhou (not in the
ruleset), and all open-hand variants.

### Fu reasons

| Reason | Fu |
|---|---|
| `base` | 20 |
| `chiitoitsu` | 25 (the only item for chiitoitsu) |
| `closed_ron` | 10 |
| `tsumo` | 2 (not for pinfu tsumo) |
| `wait_kanchan` / `wait_penchan` / `wait_tanki` | 2 |
| `pair_yakuhai` | 2 per yakuhai pair; a double-wind pair is a single item worth 4 |
| `closed_triplet_simple` | 4 |
| `closed_triplet_terminal_honor` | 8 |
| `open_triplet_simple` | 2 (a triplet completed by ron on a shanpon wait) |
| `open_triplet_terminal_honor` | 4 |

When a hand has several decompositions, `expected` is the highest-scoring one (see architecture.md).

## Oracle cross-check

`tools/oracle/check_fixtures.py` validates every case's format (vocabularies, han and fu sums, point keys) and
compares han, fu, yaku, dora counts, limit and points with the Python [`mahjong`](https://pypi.org/project/mahjong/)
library, configured for our ruleset. It does not compare the fu breakdown, because the library names fu items
differently and the total fu already covers it. The library is a test oracle only; see AGENTS.md rule 5.

```bash
python -m venv tools/oracle/.venv
tools/oracle/.venv/Scripts/python -m pip install -r tools/oracle/requirements.txt   # bin/python on Linux/macOS
tools/oracle/.venv/Scripts/python tools/oracle/check_fixtures.py                     # -v prints the oracle result
```

Add a case: write it by hand (compute the expected value yourself), run the check, and fix whichever side is wrong.
