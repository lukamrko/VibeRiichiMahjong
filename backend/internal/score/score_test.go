package score

import (
	"errors"
	"slices"
	"testing"

	"github.com/lukamrko/VibeRiichiMahjong/backend/internal/hand"
	"github.com/lukamrko/VibeRiichiMahjong/backend/internal/tile"
)

func mustKind(t *testing.T, s string) tile.Kind {
	t.Helper()
	tl, err := tile.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return tl.Kind
}

func mustTiles(t *testing.T, ids ...string) []tile.Tile {
	t.Helper()
	out := make([]tile.Tile, len(ids))
	for i, id := range ids {
		tl, err := tile.Parse(id)
		if err != nil {
			t.Fatal(err)
		}
		out[i] = tl
	}
	return out
}

// baseContext is a non-dealer (South seat, East round) ron with a
// dora indicator that makes no tile in the test hands dora unless stated.
func baseContext(t *testing.T) WinContext {
	t.Helper()
	return WinContext{
		RoundWind:      mustKind(t, "1z"),
		SeatWind:       mustKind(t, "2z"),
		WinType:        Ron,
		DoraIndicators: mustTiles(t, "3z"), // dora 4z
	}
}

func mustScore(t *testing.T, handStr, win string, ctx WinContext) Result {
	t.Helper()
	h, err := hand.Parse(handStr, win)
	if err != nil {
		t.Fatalf("hand.Parse(%q, %q): %v", handStr, win, err)
	}
	r, err := Score(h, ctx)
	if err != nil {
		t.Fatalf("Score(%q, %q): %v", handStr, win, err)
	}
	return r
}

func yakuNames(r Result) []string {
	out := make([]string, len(r.Yaku))
	for i, y := range r.Yaku {
		out[i] = y.Yaku.String()
	}
	return out
}

func TestScoreYaku(t *testing.T) {
	tsumo := func(c *WinContext) { c.WinType = Tsumo }
	dealerTsumo := func(c *WinContext) { c.WinType = Tsumo; c.SeatWind = c.RoundWind }
	tests := []struct {
		name      string
		hand, win string
		mutate    func(*WinContext)
		wantYaku  []string
		wantHan   int
		wantFu    int
		wantLimit Limit
		wantTotal int
	}{
		{
			name: "pinfu with a non-value wind pair",
			hand: "123m456p234789s33z", win: "6p",
			wantYaku: []string{"pinfu"}, wantHan: 1, wantFu: 30, wantTotal: 1000,
		},
		{
			name: "triplet reading beats pinfu+iipeikou",
			hand: "111222333m789p55s", win: "1m", mutate: tsumo,
			wantYaku: []string{"menzen_tsumo", "sanankou"}, wantHan: 3, wantFu: 40, wantTotal: 5200,
		},
		{
			name: "honroutou chiitoitsu",
			hand: "1199m1199p11s1122z", win: "2z",
			wantYaku: []string{"honroutou", "chiitoitsu"}, wantHan: 4, wantFu: 25, wantTotal: 6400,
		},
		{
			name: "tsuuiisou chiitoitsu",
			hand: "11223344556677z", win: "7z",
			wantYaku: []string{"tsuuiisou"}, wantHan: 13, wantLimit: Yakuman, wantTotal: 32000,
		},
		{
			name: "ryuuiisou without hatsu",
			hand: "22233344466688s", win: "6s",
			wantYaku: []string{"ryuuiisou"}, wantHan: 13, wantLimit: Yakuman, wantTotal: 32000,
		},
		{
			name: "tenhou stacks with suuankou tanki",
			hand: "22233m444p666888s", win: "3m", mutate: func(c *WinContext) { dealerTsumo(c); c.Tenhou = true },
			wantYaku: []string{"suuankou_tanki", "tenhou"}, wantHan: 39, wantLimit: TripleYakuman, wantTotal: 144000,
		},
		{
			name: "sextuple yakuman",
			hand: "11122233344455z", win: "5z", mutate: func(c *WinContext) { dealerTsumo(c); c.Tenhou = true },
			wantYaku: []string{"suuankou_tanki", "daisuushii", "tsuuiisou", "tenhou"}, wantHan: 78,
			wantLimit: SextupleYakuman, wantTotal: 288000,
		},
		{
			name: "dora do not count on yakuman",
			hand: "22233m444p666888s", win: "8s", mutate: func(c *WinContext) {
				tsumo(c)
				c.DoraIndicators = mustTiles(t, "1m")
			},
			wantYaku: []string{"suuankou"}, wantHan: 13, wantLimit: Yakuman, wantTotal: 32000,
		},
		{
			name: "double wind (East round, East seat) gives two yakuhai entries",
			hand: "234567m345p99s111z", win: "3p", mutate: func(c *WinContext) { c.SeatWind = c.RoundWind },
			wantYaku: []string{"yakuhai_seat_wind", "yakuhai_round_wind"}, wantHan: 2, wantFu: 40, wantTotal: 3900,
		},
		{
			name: "multiple dora indicators and ura dora",
			hand: "123m45699p345678s", win: "6s", mutate: func(c *WinContext) {
				c.Riichi = true
				c.DoraIndicators = mustTiles(t, "8p", "8p") // 9p is dora twice: 2 tiles × 2
				c.UraDoraIndicators = mustTiles(t, "2m")    // 3m
			},
			wantYaku: []string{"riichi", "pinfu"}, wantHan: 7, wantFu: 30, wantLimit: Haneman, wantTotal: 12000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := baseContext(t)
			if tt.mutate != nil {
				tt.mutate(&ctx)
			}
			r := mustScore(t, tt.hand, tt.win, ctx)
			if got := yakuNames(r); !slices.Equal(got, tt.wantYaku) {
				t.Errorf("yaku = %v, want %v", got, tt.wantYaku)
			}
			if r.Han != tt.wantHan || r.Fu != tt.wantFu || r.Limit != tt.wantLimit || r.Points.Total != tt.wantTotal {
				t.Errorf("han/fu/limit/total = %d/%d/%v/%d, want %d/%d/%v/%d",
					r.Han, r.Fu, r.Limit, r.Points.Total, tt.wantHan, tt.wantFu, tt.wantLimit, tt.wantTotal)
			}
		})
	}
}

func TestScoreErrors(t *testing.T) {
	tests := []struct {
		name, hand, win string
		mutate          func(*WinContext)
		want            error
	}{
		{name: "not a winning hand", hand: "123m456p789s11z234z", win: "1m", want: ErrNotWinningHand},
		{name: "no yaku", hand: "123m456p234789s55z", win: "6p", want: ErrNoYaku},
		{name: "dora alone is not a yaku", hand: "123m456p234789s55z", win: "6p",
			mutate: func(c *WinContext) { c.DoraIndicators = mustTiles(t, "7z") }, want: ErrNoYaku},
		{name: "fifth copy via dora indicator", hand: "11112345678999m", win: "5m",
			mutate: func(c *WinContext) { c.DoraIndicators = mustTiles(t, "1m") }, want: ErrInvalidContext},
		{name: "second red five via dora indicator", hand: "123m406p789s11122z", win: "1m",
			mutate: func(c *WinContext) { c.DoraIndicators = mustTiles(t, "0p") }, want: ErrInvalidContext},
		{name: "context is validated", hand: "123m45699p345678s", win: "6s",
			mutate: func(c *WinContext) { c.Ippatsu = true }, want: ErrInvalidContext},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := baseContext(t)
			if tt.mutate != nil {
				tt.mutate(&ctx)
			}
			h, err := hand.Parse(tt.hand, tt.win)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Score(h, ctx); !errors.Is(err, tt.want) {
				t.Errorf("Score error = %v, want %v", err, tt.want)
			}
		})
	}
}
