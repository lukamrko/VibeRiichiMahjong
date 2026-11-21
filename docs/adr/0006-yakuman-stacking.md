# ADR-0006: Yakuman stacking and double yakuman

**Status:** Accepted (2026-11-21). Settles the "double-yakuman handling" question left open in #4.

## Context
Rulesets disagree on what happens when a hand has more than one yakuman, or one of the "double" yakuman:

- Some cap a hand at one yakuman (or two).
- Some count suuankou tanki, kokushi juusanmen, junsei chuuren poutou and daisuushii as a double yakuman. Others count them as a single one.
- Some also count dora and regular yaku on top of a yakuman.

#2 already decided that the four double yakuman count double (`docs/domain/rules.md`). #4 had to decide stacking, the cap, and what tenhou/chiihou do to the wait.

## Decision
We follow Tenhou / M-League.

- **Yakuman stack with no cap.** Each yakuman adds 13 han: 1×, and 2× for the four double yakuman. The limit is `yakuman`, `double_yakuman`, `triple_yakuman` … `sextuple_yakuman`. Base points are 8000 × the multiple. The highest a closed v1 hand can reach is 6×, for example tenhou + daisuushii + tsuuiisou + suuankou tanki. The fixture vocabulary and the oracle script include every level up to `sextuple_yakuman`.
- **A yakuman hand scores yakuman only.** Regular yaku and dora (dora, aka, ura) are ignored. The API reports `dora`, `akaDora` and `uraDora` as `0`, `fu` as `null` and `fuBreakdown` as empty.
- **Kazoe yakuman** (13+ han from regular yaku and dora) is always a single yakuman. It never stacks with a real yakuman, because a hand with a real yakuman ignores regular yaku.
- **Tenhou / chiihou keep the declared winning tile.** The wait, and so the choice between a single and a double form (kokushi vs juusanmen, suuankou vs suuankou tanki, chuuren vs junsei chuuren), comes from the `winningTile` in the request. Tenhou has no "winning tile" in the game itself, so the client chooses which tile to report. We don't treat every tenhou hand as the double form.

## Consequences
- The `Limit` type has a value for each multiple from 1× to 6×. More than 6× is impossible for a closed v1 hand and panics.
- Open melds and closed kan (out of v1 scope) can't push the maximum above 6×. Revisit this ADR if a new yakuman is added.
- If a fixture's expectation differs from the Python oracle because of these choices, the fixture carries a note, as ADR-0005 requires.
