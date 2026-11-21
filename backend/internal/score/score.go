// Package score computes yaku, han, fu and points for a closed winning hand
// under the Tenhou / M-League ruleset (see docs/domain/rules.md).
package score

import (
	"fmt"

	"github.com/lukamrko/VibeRiichiMahjong/backend/internal/hand"
	"github.com/lukamrko/VibeRiichiMahjong/backend/internal/tile"
)

// YakuHan is one scored yaku with its han.
type YakuHan struct {
	Yaku Yaku
	Han  int
}

// Result is the score of a hand: the highest-scoring reading.
type Result struct {
	Decomposition hand.Decomposition
	Yaku          []YakuHan
	// Dora counts. They are all 0 for a yakuman hand, where dora do not count.
	Dora, AkaDora, UraDora int
	// Han is yaku han plus dora; for a yakuman hand it is 13 per yakuman.
	Han int
	// Fu is the rounded fu, and FuBreakdown the items before rounding.
	// Both are zero/empty for a yakuman hand.
	Fu          int
	FuBreakdown []FuItem
	Limit       Limit
	Points      Points
}

// IsYakuman reports whether the result is a (possibly stacked) yakuman;
// kazoe yakuman is not.
func (r Result) IsYakuman() bool { return len(r.Yaku) > 0 && r.Yaku[0].Yaku.IsYakuman() }

// Score scores a validated hand in the given context. It returns
// ErrInvalidContext, ErrNotWinningHand or ErrNoYaku (wrapped) on failure.
// When several decompositions are possible it returns the one with the most
// points, then the most han, then the most fu.
func Score(h hand.Hand, ctx WinContext) (Result, error) {
	if err := ctx.Validate(); err != nil {
		return Result{}, err
	}
	if err := checkPhysicalTiles(h, ctx); err != nil {
		return Result{}, err
	}
	ds := h.Decompositions()
	if len(ds) == 0 {
		return Result{}, ErrNotWinningHand
	}
	var (
		best  Result
		found bool
	)
	for _, d := range ds {
		r, ok := scoreReading(h, newReading(h, d, ctx))
		if ok && (!found || better(r, best)) {
			best, found = r, true
		}
	}
	if !found {
		return Result{}, ErrNoYaku
	}
	return best, nil
}

func better(a, b Result) bool {
	if a.Points.Total != b.Points.Total {
		return a.Points.Total > b.Points.Total
	}
	if a.Han != b.Han {
		return a.Han > b.Han
	}
	return a.Fu > b.Fu
}

func scoreReading(h hand.Hand, r reading) (Result, bool) {
	res := Result{Decomposition: r.d}
	if yakuman := r.yakuman(); len(yakuman) > 0 {
		for _, y := range yakuman {
			res.Yaku = append(res.Yaku, YakuHan{y, y.Han()})
			res.Han += y.Han()
		}
		res.Limit = yakumanLimit(res.Han / 13)
		res.Points = payout(yakumanBase*(res.Han/13), r.ctx)
		return res, true
	}

	yaku := r.regularYaku()
	if len(yaku) == 0 {
		return Result{}, false
	}
	pinfu := false
	for _, y := range yaku {
		res.Yaku = append(res.Yaku, YakuHan{y, y.Han()})
		res.Han += y.Han()
		pinfu = pinfu || y == Pinfu
	}
	res.Dora = countDora(h, r.ctx.DoraIndicators)
	res.AkaDora = h.RedFives()
	if r.ctx.Riichi {
		res.UraDora = countDora(h, r.ctx.UraDoraIndicators)
	}
	res.Han += res.Dora + res.AkaDora + res.UraDora
	res.FuBreakdown, res.Fu = r.fu(pinfu)

	base, limit := basePoints(res.Han, res.Fu)
	res.Limit = limit
	res.Points = payout(base, r.ctx)
	return res, true
}

// DoraFor returns the dora kind shown by an indicator: the next tile in its
// suit, wrapping 9→1, North→East and Chun→Haku.
func DoraFor(indicator tile.Kind) tile.Kind {
	n := indicator.Number()
	suit := indicator.Suit()
	var next int
	switch {
	case suit != tile.Honor:
		next = n%9 + 1
	case n <= 4: // winds
		next = n%4 + 1
	default: // dragons 5–7
		next = (n-5+1)%3 + 5
	}
	k, err := tile.KindOf(suit, next)
	if err != nil {
		panic(err)
	}
	return k
}

func countDora(h hand.Hand, indicators []tile.Tile) int {
	counts := h.Counts()
	n := 0
	for _, ind := range indicators {
		n += counts[DoraFor(ind.Kind())]
	}
	return n
}

// checkPhysicalTiles rejects indicators that would need a fifth copy of a
// tile or a second red five in a suit, counting the hand and all indicators.
func checkPhysicalTiles(h hand.Hand, ctx WinContext) error {
	all := append(h.Tiles(), ctx.DoraIndicators...)
	all = append(all, ctx.UraDoraIndicators...)
	if err := hand.CheckCopies(all); err != nil {
		return fmt.Errorf("%w: hand and indicators: %w", ErrInvalidContext, err)
	}
	return nil
}
