package score

import (
	"fmt"
	"slices"

	"github.com/lukamrko/VibeRiichiMahjong/backend/internal/hand"
)

// FuReason names one fu item (closed vocabulary in fixtures/hands/README.md).
type FuReason uint8

// Fu reasons.
const (
	FuBase FuReason = iota
	FuChiitoitsu
	FuClosedRon
	FuTsumo
	FuWaitKanchan
	FuWaitPenchan
	FuWaitTanki
	FuPairYakuhai
	FuClosedTripletSimple
	FuClosedTripletTerminalHonor
	FuOpenTripletSimple
	FuOpenTripletTerminalHonor
)

func (f FuReason) String() string {
	switch f {
	case FuBase:
		return "base"
	case FuChiitoitsu:
		return "chiitoitsu"
	case FuClosedRon:
		return "closed_ron"
	case FuTsumo:
		return "tsumo"
	case FuWaitKanchan:
		return "wait_kanchan"
	case FuWaitPenchan:
		return "wait_penchan"
	case FuWaitTanki:
		return "wait_tanki"
	case FuPairYakuhai:
		return "pair_yakuhai"
	case FuClosedTripletSimple:
		return "closed_triplet_simple"
	case FuClosedTripletTerminalHonor:
		return "closed_triplet_terminal_honor"
	case FuOpenTripletSimple:
		return "open_triplet_simple"
	case FuOpenTripletTerminalHonor:
		return "open_triplet_terminal_honor"
	default:
		return fmt.Sprintf("FuReason(%d)", uint8(f))
	}
}

// FuItem is one line of the fu breakdown.
type FuItem struct {
	Reason FuReason
	Fu     int
}

// fu returns the fu breakdown (before rounding) and the rounded fu for a
// non-yakuman reading. Items are ordered: base, ron/tsumo, triplets (highest
// first), wait, pair.
func (r reading) fu(pinfu bool) ([]FuItem, int) {
	switch r.d.Form {
	case hand.Chiitoitsu:
		return []FuItem{{FuChiitoitsu, 25}}, 25
	case hand.Standard:
	case hand.Kokushi:
		panic("score: kokushi has no fu")
	default:
		panic(fmt.Sprintf("score: unknown form %v", r.d.Form))
	}

	items := []FuItem{{FuBase, 20}}
	ron := r.ctx.WinType == Ron
	if ron {
		items = append(items, FuItem{FuClosedRon, 10})
	}
	if pinfu {
		return items, roundFu(items)
	}
	if !ron {
		items = append(items, FuItem{FuTsumo, 2})
	}

	var triplets []FuItem
	for i, g := range r.sets {
		if g.Kind != hand.Triplet {
			continue
		}
		open := i == r.openTriplet
		terminal := g.First.IsTerminalOrHonor()
		switch {
		case !open && terminal:
			triplets = append(triplets, FuItem{FuClosedTripletTerminalHonor, 8})
		case !open:
			triplets = append(triplets, FuItem{FuClosedTripletSimple, 4})
		case terminal:
			triplets = append(triplets, FuItem{FuOpenTripletTerminalHonor, 4})
		default:
			triplets = append(triplets, FuItem{FuOpenTripletSimple, 2})
		}
	}
	slices.SortStableFunc(triplets, func(a, b FuItem) int { return b.Fu - a.Fu })
	items = append(items, triplets...)

	switch r.d.Wait {
	case hand.Kanchan:
		items = append(items, FuItem{FuWaitKanchan, 2})
	case hand.Penchan:
		items = append(items, FuItem{FuWaitPenchan, 2})
	case hand.Tanki:
		items = append(items, FuItem{FuWaitTanki, 2})
	case hand.Ryanmen, hand.Shanpon:
	case hand.Juusanmen:
		panic("score: juusanmen in a standard hand")
	default:
		panic(fmt.Sprintf("score: unknown wait %v", r.d.Wait))
	}

	pair := r.pair.First
	pairFu := 0
	if isDragon(pair) {
		pairFu = 2
	}
	if pair == r.ctx.SeatWind {
		pairFu += 2
	}
	if pair == r.ctx.RoundWind {
		pairFu += 2
	}
	if pairFu > 0 {
		items = append(items, FuItem{FuPairYakuhai, pairFu})
	}
	return items, roundFu(items)
}

func roundFu(items []FuItem) int {
	sum := 0
	for _, it := range items {
		sum += it.Fu
	}
	return (sum + 9) / 10 * 10
}
