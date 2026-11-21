# Ruleset and scope

## Ruleset: Tenhou / M-League style

- Aka dora: one red five per suit (`0m`, `0p`, `0s`), each worth 1 han of dora.
- Kuitan (open tanyao) allowed — irrelevant in v1 (closed only), but don't encode the opposite.
- **No kiriage mangan**: 4 han 30 fu = 7700 (non-dealer), 3 han 60 fu = 7700.
- Kazoe yakuman: 13+ han counts as yakuman.
- Multiple yakuman stack (double yakuman possible), with no cap: up to 6× for a closed hand (ADR-0006).
  Regular yaku and dora are ignored when a hand has a yakuman. For tenhou / chiihou the wait (and so
  single vs double forms) comes from the declared winning tile.
- **Double yakuman hands count double** (decided in #2): suuankou tanki, kokushi musou 13-sided wait
  (juusanmen), junsei chuuren poutou, and daisuushii are each worth 2× yakuman.
- Fu rounded up to the nearest 10 (chiitoitsu fixed at 25 fu). Pinfu tsumo = 20 fu; closed ron +10 fu.
- Honba: +300 per honba (ron), +100 per player (tsumo).

## v1 scope

In:
- Closed hands only (standard 4 sets + pair, chiitoitsu, kokushi musou).
- All closed-hand yaku, dora / aka / ura, riichi-related yaku, situational yaku from the context form
  except **rinshan kaihou**: it needs your own kan, so v1 rejects `rinshan: true` as `invalid_context`
  (the field stays in the API contract for when kan is in scope).
- Points for dealer / non-dealer, ron / tsumo.

Out (later issues):
- Open melds (pon, chi, open kan) and closed kan.
- Detecting dora indicators, riichi sticks, etc. from photos.
- Ruleset configurability.
- Local yaku.
