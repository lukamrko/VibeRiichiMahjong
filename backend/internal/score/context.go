package score

import (
	"errors"
	"fmt"

	"github.com/lukamrko/VibeRiichiMahjong/backend/internal/tile"
)

// Errors returned by Score and the context helpers. Use errors.Is to test for them.
var (
	// ErrInvalidContext means the WinContext is malformed or contradicts itself
	// (or the hand, e.g. a dora indicator that would be a fifth copy of a tile).
	ErrInvalidContext = errors.New("invalid win context")
	// ErrNotWinningHand means the 14 tiles do not form any winning shape.
	ErrNotWinningHand = errors.New("not a winning hand")
	// ErrNoYaku means the hand is complete but no reading of it has a yaku.
	ErrNoYaku = errors.New("hand has no yaku")
)

// MaxDoraIndicators is the most dora indicators a round can show (1 + 4 kan dora).
const MaxDoraIndicators = 5

// WinType says how the hand was won.
type WinType uint8

// Win types.
const (
	Ron   WinType = iota // on another player's discard
	Tsumo                // by self-draw
)

func (w WinType) String() string {
	switch w {
	case Ron:
		return "ron"
	case Tsumo:
		return "tsumo"
	default:
		return fmt.Sprintf("WinType(%d)", uint8(w))
	}
}

// ParseWinType parses "ron" or "tsumo".
func ParseWinType(s string) (WinType, error) {
	switch s {
	case "ron":
		return Ron, nil
	case "tsumo":
		return Tsumo, nil
	default:
		return 0, fmt.Errorf("%w: win type %q (want ron or tsumo)", ErrInvalidContext, s)
	}
}

// Wind is a round or seat wind. The player whose seat wind is East is the dealer.
type Wind uint8

// The four winds, in seating order (1z–4z).
const (
	East Wind = iota
	South
	West
	North
)

// Kind returns the wind's honor tile kind, 1z (East) to 4z (North).
func (w Wind) Kind() tile.Kind { return honorKind(int(w) + 1) }

// String returns the wind's MPSZ id, e.g. "1z".
func (w Wind) String() string {
	switch w {
	case East, South, West, North:
		return w.Kind().String()
	default:
		return fmt.Sprintf("Wind(%d)", uint8(w))
	}
}

// valid reports whether w is one of the four winds.
func (w Wind) valid() bool {
	switch w {
	case East, South, West, North:
		return true
	default:
		return false
	}
}

// honorKind returns the honor kind for 1–7 (1z–7z).
func honorKind(n int) tile.Kind {
	k, err := tile.KindOf(tile.Honor, n)
	if err != nil {
		panic(err)
	}
	return k
}

func isWind(k tile.Kind) bool { return k.IsHonor() && k.Number() <= 4 }

// ParseWind parses a wind tile id, "1z" (East) to "4z" (North).
func ParseWind(s string) (Wind, error) {
	t, err := tile.Parse(s)
	if err != nil || t.IsRed() || !isWind(t.Kind()) {
		return 0, fmt.Errorf("%w: wind %q (want 1z–4z)", ErrInvalidContext, s)
	}
	return Wind(t.Kind().Number() - 1), nil
}

// WinContext holds every non-tile input to scoring (see docs/domain/glossary.md).
type WinContext struct {
	RoundWind Wind
	SeatWind  Wind
	WinType   WinType

	Riichi       bool
	DoubleRiichi bool // requires Riichi
	Ippatsu      bool // requires Riichi

	DoraIndicators    []tile.Tile // 1–5 indicator tiles
	UraDoraIndicators []tile.Tile // only with Riichi; at most one per dora indicator

	Honba int

	Haitei  bool // tsumo on the last wall tile
	Houtei  bool // ron on the last discard
	Rinshan bool // always rejected in v1: it needs your own kan (docs/domain/rules.md)
	Chankan bool // ron by robbing a kan
	Tenhou  bool // dealer wins on the initial deal
	Chiihou bool // non-dealer wins on the first draw
}

// IsDealer reports whether the winner is the dealer (seat wind East).
func (c WinContext) IsDealer() bool { return c.SeatWind == East }

// isValueWind reports whether k is the seat or round wind.
func (c WinContext) isValueWind(k tile.Kind) bool {
	return k == c.SeatWind.Kind() || k == c.RoundWind.Kind()
}

// Validate reports the first contradiction in the context, wrapped in ErrInvalidContext.
func (c WinContext) Validate() error {
	invalid := func(format string, args ...any) error {
		return fmt.Errorf("%w: "+format, append([]any{ErrInvalidContext}, args...)...)
	}
	tsumo := c.WinType == Tsumo
	switch c.WinType {
	case Ron, Tsumo:
	default:
		return invalid("unknown win type %d", c.WinType)
	}
	switch {
	case !c.RoundWind.valid():
		return invalid("round wind %s is not a wind", c.RoundWind)
	case !c.SeatWind.valid():
		return invalid("seat wind %s is not a wind", c.SeatWind)
	case c.DoubleRiichi && !c.Riichi:
		return invalid("doubleRiichi requires riichi")
	case c.Ippatsu && !c.Riichi:
		return invalid("ippatsu requires riichi")
	case len(c.DoraIndicators) == 0:
		return invalid("at least one dora indicator is required")
	case len(c.DoraIndicators) > MaxDoraIndicators:
		return invalid("%d dora indicators (max %d)", len(c.DoraIndicators), MaxDoraIndicators)
	case len(c.UraDoraIndicators) > 0 && !c.Riichi:
		return invalid("ura dora indicators require riichi")
	case len(c.UraDoraIndicators) > len(c.DoraIndicators):
		return invalid("%d ura dora indicators for %d dora indicators", len(c.UraDoraIndicators), len(c.DoraIndicators))
	case c.Honba < 0:
		return invalid("negative honba %d", c.Honba)
	case c.Haitei && !tsumo:
		return invalid("haitei requires tsumo")
	case c.Houtei && tsumo:
		return invalid("houtei requires ron")
	case c.Chankan && tsumo:
		return invalid("chankan requires ron")
	case c.Chankan && c.Houtei:
		return invalid("chankan and houtei cannot both apply")
	case c.Rinshan:
		return invalid("rinshan needs a kan, which is out of scope in v1")
	case c.Tenhou && (!c.IsDealer() || !tsumo || c.Riichi):
		return invalid("tenhou requires a dealer tsumo without riichi")
	case c.Chiihou && (c.IsDealer() || !tsumo || c.Riichi):
		return invalid("chiihou requires a non-dealer tsumo without riichi")
	}
	return nil
}
