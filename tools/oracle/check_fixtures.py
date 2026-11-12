"""Cross-check fixtures/hands/*.json against the Python `mahjong` library (test oracle only).

Usage:
    python tools/oracle/check_fixtures.py [--verbose]

Two kinds of checks run on every case:

1. Schema / self-consistency: required fields, MPSZ syntax, yaku ids and fu reasons
   from the closed vocabularies in fixtures/hands/README.md, han = sum(yaku) + dora,
   fu = rounded sum(fuBreakdown), points fields match the win type.
2. Oracle: han, fu, yaku set, dora counts, limit and points are compared with
   mahjong.HandCalculator configured for our ruleset (docs/domain/rules.md).
   The fu breakdown is not compared (the library names fu items differently;
   the total fu already covers it).

Exits non-zero if any case fails.
"""

from __future__ import annotations

import argparse
import json
import math
import re
import sys
from pathlib import Path

from mahjong.constants import EAST, NORTH, SOUTH, WEST
from mahjong.hand_calculating.hand import HandCalculator
from mahjong.hand_calculating.hand_config import HandConfig, HandConstants, OptionalRules
from mahjong.tile import TilesConverter

REPO_ROOT = Path(__file__).resolve().parents[2]
FIXTURES_DIR = REPO_ROOT / "fixtures" / "hands"

# Our yaku id -> library yaku name. Wind yakuhai are matched by prefix.
YAKU_TO_ORACLE = {
    "riichi": "Riichi",
    "double_riichi": "Double Riichi",
    "ippatsu": "Ippatsu",
    "menzen_tsumo": "Menzen Tsumo",
    "pinfu": "Pinfu",
    "tanyao": "Tanyao",
    "iipeikou": "Iipeiko",
    "yakuhai_haku": "Yakuhai (haku)",
    "yakuhai_hatsu": "Yakuhai (hatsu)",
    "yakuhai_chun": "Yakuhai (chun)",
    "yakuhai_seat_wind": "Yakuhai (seat wind",
    "yakuhai_round_wind": "Yakuhai (round wind",
    "haitei": "Haitei Raoyue",
    "houtei": "Houtei Raoyui",
    "chankan": "Chankan",
    "sanshoku_doujun": "Sanshoku Doujun",
    "ittsu": "Ittsu",
    "chanta": "Chantai",
    "junchan": "Junchan",
    "ryanpeikou": "Ryanpeikou",
    "sanankou": "San Ankou",
    "sanshoku_doukou": "Sanshoku Doukou",
    "toitoi": "Toitoi",
    "shousangen": "Shou Sangen",
    "honroutou": "Honroutou",
    "honitsu": "Honitsu",
    "chinitsu": "Chinitsu",
    "chiitoitsu": "Chiitoitsu",
    "kokushi_musou": "Kokushi Musou",
    "kokushi_musou_juusanmen": "Kokushi Musou Juusanmen Matchi",
    "suuankou": "Suu Ankou",
    "suuankou_tanki": "Suu Ankou Tanki",
    "daisangen": "Daisangen",
    "shousuushii": "Shousuushii",
    "daisuushii": "Dai Suushii",
    "tsuuiisou": "Tsuu Iisou",
    "chinroutou": "Chinroutou",
    "ryuuiisou": "Ryuuiisou",
    "chuuren_poutou": "Chuuren Poutou",
    "junsei_chuuren_poutou": "Daburu Chuuren Poutou",
    "tenhou": "Tenhou",
    "chiihou": "Chiihou",
}
DORA_NAMES = {"Dora": "dora", "Aka Dora": "akaDora", "Ura Dora": "uraDora"}

FU_REASONS = {
    "base",
    "chiitoitsu",
    "closed_ron",
    "tsumo",
    "wait_kanchan",
    "wait_penchan",
    "wait_tanki",
    "pair_yakuhai",
    "closed_triplet_simple",
    "closed_triplet_terminal_honor",
    "open_triplet_simple",
    "open_triplet_terminal_honor",
}

LIMIT_FROM_ORACLE = {
    "": None,
    "mangan": "mangan",
    "haneman": "haneman",
    "baiman": "baiman",
    "sanbaiman": "sanbaiman",
    "yakuman": "yakuman",
    "kazoe yakuman": "kazoe_yakuman",
    "2x yakuman": "double_yakuman",
    "3x yakuman": "triple_yakuman",
}
YAKUMAN_LIMITS = {"yakuman", "kazoe_yakuman", "double_yakuman", "triple_yakuman"}

CONTEXT_FIELDS = {
    "roundWind": str,
    "seatWind": str,
    "winType": str,
    "riichi": bool,
    "doubleRiichi": bool,
    "ippatsu": bool,
    "doraIndicators": list,
    "uraDoraIndicators": list,
    "honba": int,
    "haitei": bool,
    "houtei": bool,
    "rinshan": bool,
    "chankan": bool,
    "tenhou": bool,
    "chiihou": bool,
}
WINDS = {"1z": EAST, "2z": SOUTH, "3z": WEST, "4z": NORTH}
TILE_RE = re.compile(r"^[0-9][mps]$|^[1-7]z$")
HAND_RE = re.compile(r"^(?:[0-9]+[mps]|[1-7]+z)+$")

RULES = OptionalRules(
    has_open_tanyao=True,
    has_aka_dora=True,
    has_double_yakuman=True,
    kazoe_limit=HandConstants.KAZOE_LIMITED,
    kiriage=False,
    fu_for_pinfu_tsumo=False,
)


def parse_hand(hand: str) -> list[str]:
    """'123m406p' -> ['1m','2m','3m','4p','0p','6p']"""
    if not HAND_RE.match(hand):
        raise ValueError(f"invalid MPSZ hand string {hand!r}")
    tiles = []
    for digits, suit in re.findall(r"([0-9]+)([mpsz])", hand):
        for d in digits:
            if suit == "z" and d == "0":
                raise ValueError(f"invalid honor tile 0z in {hand!r}")
            tiles.append(d + suit)
    return tiles


def to_136(tiles: list[str]) -> list[int]:
    by_suit = {"m": "", "p": "", "s": "", "z": ""}
    for t in tiles:
        by_suit[t[1]] += t[0]
    return TilesConverter.string_to_136_array(
        man=by_suit["m"], pin=by_suit["p"], sou=by_suit["s"], honors=by_suit["z"], has_aka_dora=True
    )


def is_red(tile136: int) -> bool:
    return tile136 in (16, 52, 88)


def win_tile_136(hand136: list[int], winning: str) -> int:
    (target,) = to_136([winning])
    want_red = winning[0] == "0"
    for t in hand136:
        if t // 4 == target // 4 and is_red(t) == want_red:
            return t
    raise ValueError(f"winning tile {winning} not in hand")


def validate(case: dict) -> list[str]:
    """Schema and self-consistency checks. Returns a list of problems."""
    errs: list[str] = []
    for key in ("id", "description", "input", "expected"):
        if key not in case:
            errs.append(f"missing {key}")
    if errs:
        return errs
    inp, exp = case["input"], case["expected"]

    try:
        tiles = parse_hand(inp["hand"])
        if len(tiles) != 14:
            errs.append(f"hand has {len(tiles)} tiles, want 14")
        if inp["winningTile"] not in tiles:
            errs.append(f"winningTile {inp['winningTile']} not in hand")
        counts: dict[str, int] = {}
        for t in tiles:
            k = t.replace("0", "5")
            counts[k] = counts.get(k, 0) + 1
        if any(c > 4 for c in counts.values()):
            errs.append("more than 4 copies of a tile")
    except (ValueError, KeyError) as e:
        errs.append(str(e))

    ctx = inp.get("context", {})
    if set(ctx) != set(CONTEXT_FIELDS):
        errs.append(f"context fields differ: missing {set(CONTEXT_FIELDS) - set(ctx)}, extra {set(ctx) - set(CONTEXT_FIELDS)}")
    for k, typ in CONTEXT_FIELDS.items():
        if k in ctx and not isinstance(ctx[k], typ):
            errs.append(f"context.{k} should be {typ.__name__}")
    if ctx.get("roundWind") not in WINDS or ctx.get("seatWind") not in WINDS:
        errs.append("roundWind/seatWind must be 1z..4z")
    if ctx.get("winType") not in ("ron", "tsumo"):
        errs.append("winType must be ron or tsumo")
    for t in ctx.get("doraIndicators", []) + ctx.get("uraDoraIndicators", []):
        if not TILE_RE.match(t):
            errs.append(f"bad indicator {t}")
    if ctx.get("uraDoraIndicators") and not (ctx.get("riichi") or ctx.get("doubleRiichi")):
        errs.append("uraDoraIndicators require riichi")
    if (ctx.get("doubleRiichi") or ctx.get("ippatsu")) and not ctx.get("riichi"):
        errs.append("doubleRiichi and ippatsu require riichi: true")
    if ctx.get("rinshan"):
        errs.append("rinshan is impossible in v1 (no kan)")

    yaku_han = 0
    for y in exp.get("yaku", []):
        if y.get("name") not in YAKU_TO_ORACLE:
            errs.append(f"unknown yaku id {y.get('name')!r}")
        yaku_han += y.get("han", 0)
    limit = exp.get("limit")
    if limit is not None and limit not in LIMIT_FROM_ORACLE.values():
        errs.append(f"unknown limit {limit!r}")
    is_yakuman = limit in YAKUMAN_LIMITS and limit != "kazoe_yakuman"
    dora_total = exp.get("dora", 0) + exp.get("akaDora", 0) + exp.get("uraDora", 0)
    want_han = yaku_han if is_yakuman else yaku_han + dora_total
    if exp.get("han") != want_han:
        errs.append(f"han {exp.get('han')} != sum of yaku{'' if is_yakuman else ' + dora'} ({want_han})")

    if is_yakuman:
        if exp.get("fu") is not None or exp.get("fuBreakdown"):
            errs.append("yakuman cases use fu: null and an empty fuBreakdown")
    else:
        fb = exp.get("fuBreakdown", [])
        for item in fb:
            if item.get("reason") not in FU_REASONS:
                errs.append(f"unknown fu reason {item.get('reason')!r}")
        raw = sum(item.get("fu", 0) for item in fb)
        want_fu = raw if raw == 25 else int(math.ceil(raw / 10) * 10)
        if exp.get("fu") != want_fu:
            errs.append(f"fu {exp.get('fu')} != rounded fuBreakdown sum {raw} -> {want_fu}")

    pts = exp.get("points", {})
    dealer = ctx.get("seatWind") == "1z"
    if ctx.get("winType") == "ron":
        want_keys = {"total", "fromDiscarder"}
    elif dealer:
        want_keys = {"total", "fromEachNonDealer"}
    else:
        want_keys = {"total", "fromDealer", "fromNonDealer"}
    if set(pts) != want_keys:
        errs.append(f"points keys {sorted(pts)} != {sorted(want_keys)}")
    else:
        parts = (
            pts.get("fromDiscarder", 0)
            + 3 * pts.get("fromEachNonDealer", 0)
            + pts.get("fromDealer", 0)
            + 2 * pts.get("fromNonDealer", 0)
        )
        if parts != pts["total"]:
            errs.append(f"points total {pts['total']} != sum of payments {parts}")
    return errs


def oracle(case: dict) -> dict:
    """Score the case with the library and return it in our expected-shape."""
    inp = case["input"]
    ctx = inp["context"]
    tiles = parse_hand(inp["hand"])
    hand136 = to_136(tiles)
    config = HandConfig(
        is_tsumo=ctx["winType"] == "tsumo",
        is_riichi=ctx["riichi"],
        is_daburu_riichi=ctx["doubleRiichi"],
        is_ippatsu=ctx["ippatsu"],
        is_rinshan=ctx["rinshan"],
        is_chankan=ctx["chankan"],
        is_haitei=ctx["haitei"],
        is_houtei=ctx["houtei"],
        is_tenhou=ctx["tenhou"],
        is_chiihou=ctx["chiihou"],
        player_wind=WINDS[ctx["seatWind"]],
        round_wind=WINDS[ctx["roundWind"]],
        tsumi_number=ctx["honba"],
        options=RULES,
    )
    res = HandCalculator.estimate_hand_value(
        hand136,
        win_tile_136(hand136, inp["winningTile"]),
        dora_indicators=to_136(ctx["doraIndicators"]) or None,
        ura_dora_indicators=to_136(ctx["uraDoraIndicators"]) or None,
        config=config,
    )
    if res.error:
        return {"error": res.error}

    yaku: list[str] = []
    dora = {"dora": 0, "akaDora": 0, "uraDora": 0}
    reverse = {v: k for k, v in YAKU_TO_ORACLE.items()}
    for y in res.yaku:
        if y.name in DORA_NAMES:
            dora[DORA_NAMES[y.name]] = y.han_closed
            continue
        if y.name.startswith("Yakuhai (seat wind"):
            yaku.append("yakuhai_seat_wind")
        elif y.name.startswith("Yakuhai (round wind"):
            yaku.append("yakuhai_round_wind")
        else:
            yaku.append(reverse.get(y.name, f"<unmapped {y.name}>"))

    cost = res.cost
    limit = LIMIT_FROM_ORACLE.get(cost["yaku_level"], f"<unmapped {cost['yaku_level']}>")
    # The library labels a capped kazoe hand plain "yakuman"; tell it apart by the
    # absence of any yakuman-level yaku.
    if limit == "yakuman" and not any(y.is_yakuman for y in res.yaku):
        limit = "kazoe_yakuman"
    dealer = config.is_dealer
    if not config.is_tsumo:
        points = {"total": cost["total"], "fromDiscarder": cost["main"] + cost["main_bonus"]}
    elif dealer:
        points = {"total": cost["total"], "fromEachNonDealer": cost["main"] + cost["main_bonus"]}
    else:
        points = {
            "total": cost["total"],
            "fromDealer": cost["main"] + cost["main_bonus"],
            "fromNonDealer": cost["additional"] + cost["additional_bonus"],
        }
    return {
        "yaku": sorted(yaku),
        "han": res.han,
        "fu": res.fu,
        "limit": limit,
        "points": points,
        **dora,
    }


def compare(case: dict, got: dict) -> list[str]:
    if "error" in got:
        return [f"oracle error: {got['error']}"]
    exp = case["expected"]
    errs = []
    want_yaku = sorted(y["name"] for y in exp["yaku"])
    if want_yaku != got["yaku"]:
        errs.append(f"yaku {want_yaku} != oracle {got['yaku']}")
    is_yakuman = exp["limit"] in YAKUMAN_LIMITS and exp["limit"] != "kazoe_yakuman"
    checks = ["han", "limit", "points"]
    if not is_yakuman:
        checks += ["fu", "dora", "akaDora", "uraDora"]
    for k in checks:
        if exp[k] != got[k]:
            errs.append(f"{k} {exp[k]!r} != oracle {got[k]!r}")
    return errs


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--verbose", "-v", action="store_true", help="print the oracle result for every case")
    args = ap.parse_args()

    files = sorted(FIXTURES_DIR.glob("*.json"))
    if not files:
        print(f"no fixtures in {FIXTURES_DIR}", file=sys.stderr)
        return 1
    seen: set[str] = set()
    failed = total = 0
    for f in files:
        cases = json.loads(f.read_text(encoding="utf-8"))
        for case in cases:
            total += 1
            cid = case.get("id", "<no id>")
            errs = []
            if cid in seen:
                errs.append("duplicate id")
            seen.add(cid)
            errs += validate(case)
            got = None
            if not errs:
                got = oracle(case)
                errs += compare(case, got)
            status = "FAIL" if errs else "ok"
            print(f"{status:4} {f.name}:{cid}")
            for e in errs:
                print(f"       - {e}")
            if args.verbose and got:
                print(f"       oracle: {json.dumps(got)}")
            failed += bool(errs)
    print(f"\n{total - failed}/{total} cases passed")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
