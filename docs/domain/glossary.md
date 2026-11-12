# Glossary

Use these terms (and spellings) in code, tests, APIs and docs.

## Tile notation (MPSZ)

| Notation | Meaning |
|---|---|
| `1m`–`9m` | Manzu (characters) |
| `1p`–`9p` | Pinzu (circles) |
| `1s`–`9s` | Souzu (bamboo) |
| `1z`–`4z` | Winds: East, South, West, North |
| `5z`, `6z`, `7z` | Dragons: Haku (white), Hatsu (green), Chun (red) |
| `0m`, `0p`, `0s` | Aka (red) five — same as a 5 for hand shape, +1 dora |

Compact hand strings group digits before their suit letter: `123m406p789s11z` = 1m 2m 3m 4p 0p 6p 7s 8s 9s 1z 1z.
There are 34 distinct tile kinds (37 with red fives as separate detection classes).

## Domain terms

| Term | Meaning |
|---|---|
| **Tile** | One tile, identified by its MPSZ id. |
| **Hand** | The 14 tiles of a winning hand (v1: all concealed). |
| **Winning tile** | The tile that completed the hand (agari-hai). Affects fu and some yaku (pinfu, wait type). |
| **Meld / Set (mentsu)** | Sequence (shuntsu, e.g. `345p`) or triplet (koutsu, e.g. `777s`). Kan = quad. |
| **Pair (jantou)** | Two identical tiles. |
| **Decomposition** | One way of splitting a hand into 4 sets + pair (or 7 pairs / kokushi), plus which group the winning tile completed. A hand can have several; scoring keeps the highest. |
| **Form** | Overall shape of a winning hand: `standard` (4 sets + pair), `chiitoitsu` (7 distinct pairs; four of a kind is not two pairs), `kokushi`. |
| **Wait (machi)** | Shape the winning tile completed: ryanmen, kanchan, penchan, shanpon, tanki. Chiitoitsu is always tanki; kokushi is tanki, or **juusanmen** (13-sided) when the winning tile completed the pair. |
| **Yaku** | A scoring pattern; a hand needs ≥1 yaku (dora don't count) to win. |
| **Han** | Doubling units from yaku and dora. |
| **Fu** | Minipoints from hand composition and wait. |
| **Dora** | Bonus tile: the tile *after* the indicator (9→1, N→E, Chun→Haku). |
| **Ura dora** | Hidden dora under the indicator; only with riichi. |
| **Aka dora** | Red fives. |
| **Limit hand** | Mangan, haneman, baiman, sanbaiman, yakuman (and kazoe yakuman). |
| **Ron / Tsumo** | Win on another player's discard / by self-draw. |
| **Dealer (oya)** | East seat; scores 1.5×. |
| **Round wind / Seat wind** | Prevailing wind / player's own wind; affect yakuhai and fu. |
| **Honba** | Repeat counter; adds 300 points per honba. |
| **WinContext** | All non-tile inputs to scoring (winds, riichi, ron/tsumo, dora indicators, honba, situational flags). |
| **Detection** | Vision output for one tile: MPSZ id, confidence, bounding box. |
