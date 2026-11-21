package score

import "fmt"

// Limit is the limit-hand level of a result.
type Limit uint8

// Limit levels (closed vocabulary in fixtures/hands/README.md).
const (
	NoLimit Limit = iota
	Mangan
	Haneman
	Baiman
	Sanbaiman
	Yakuman
	KazoeYakuman
	DoubleYakuman
	TripleYakuman
	QuadrupleYakuman
	QuintupleYakuman
	SextupleYakuman
)

// String returns the limit id, or "" for NoLimit.
func (l Limit) String() string {
	switch l {
	case NoLimit:
		return ""
	case Mangan:
		return "mangan"
	case Haneman:
		return "haneman"
	case Baiman:
		return "baiman"
	case Sanbaiman:
		return "sanbaiman"
	case Yakuman:
		return "yakuman"
	case KazoeYakuman:
		return "kazoe_yakuman"
	case DoubleYakuman:
		return "double_yakuman"
	case TripleYakuman:
		return "triple_yakuman"
	case QuadrupleYakuman:
		return "quadruple_yakuman"
	case QuintupleYakuman:
		return "quintuple_yakuman"
	case SextupleYakuman:
		return "sextuple_yakuman"
	default:
		return fmt.Sprintf("Limit(%d)", uint8(l))
	}
}

// yakumanBase is the base points of one yakuman.
const yakumanBase = 8000

// basePoints returns the base points (before the dealer/ron multipliers) and
// the limit for a non-yakuman hand. There is no kiriage mangan, and kazoe
// yakuman is a single yakuman however many han it has.
func basePoints(han, fu int) (int, Limit) {
	switch {
	case han >= 13:
		return yakumanBase, KazoeYakuman
	case han >= 11:
		return 6000, Sanbaiman
	case han >= 8:
		return 4000, Baiman
	case han >= 6:
		return 3000, Haneman
	case han == 5:
		return 2000, Mangan
	}
	base := fu << (han + 2)
	if base > 2000 {
		return 2000, Mangan
	}
	return base, NoLimit
}

// yakumanLimit returns the limit for n stacked yakuman (a double yakuman counts as 2).
func yakumanLimit(n int) Limit {
	switch n {
	case 1:
		return Yakuman
	case 2:
		return DoubleYakuman
	case 3:
		return TripleYakuman
	case 4:
		return QuadrupleYakuman
	case 5:
		return QuintupleYakuman
	case 6:
		return SextupleYakuman
	default:
		panic(fmt.Sprintf("score: impossible yakuman multiple %d", n))
	}
}

// Points are the payments for a win, each including honba. Which fields are
// set depends on the win: ron → FromDiscarder; non-dealer tsumo → FromDealer
// and FromNonDealer (each non-dealer pays it); dealer tsumo → FromEachNonDealer.
type Points struct {
	Total             int
	FromDiscarder     int
	FromDealer        int
	FromNonDealer     int
	FromEachNonDealer int
}

func roundUp100(n int) int { return (n + 99) / 100 * 100 }

// payout turns base points into payments for the winner described by ctx.
func payout(base int, ctx WinContext) Points {
	dealer, honba := ctx.IsDealer(), ctx.Honba
	switch ctx.WinType {
	case Ron:
		mult := 4
		if dealer {
			mult = 6
		}
		p := roundUp100(base*mult) + 300*honba
		return Points{Total: p, FromDiscarder: p}
	case Tsumo:
		if dealer {
			each := roundUp100(base*2) + 100*honba
			return Points{Total: 3 * each, FromEachNonDealer: each}
		}
		fromDealer := roundUp100(base*2) + 100*honba
		fromNonDealer := roundUp100(base) + 100*honba
		return Points{Total: fromDealer + 2*fromNonDealer, FromDealer: fromDealer, FromNonDealer: fromNonDealer}
	default:
		panic(fmt.Sprintf("score: unknown win type %v", ctx.WinType))
	}
}
