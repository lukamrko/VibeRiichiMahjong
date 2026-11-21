package score

import (
	"errors"
	"testing"
)

func TestWinContextValidate(t *testing.T) {
	valid := func(c *WinContext) {}
	tests := []struct {
		name    string
		mutate  func(*WinContext)
		wantErr bool
	}{
		{"base context", valid, false},
		{"riichi with ura", func(c *WinContext) { c.Riichi = true; c.UraDoraIndicators = mustTiles(t, "1m") }, false},
		{"riichi without ura", func(c *WinContext) { c.Riichi = true }, false},
		{"double riichi + ippatsu", func(c *WinContext) { c.Riichi, c.DoubleRiichi, c.Ippatsu = true, true, true }, false},
		{"tenhou", func(c *WinContext) { c.SeatWind, c.WinType, c.Tenhou = c.RoundWind, Tsumo, true }, false},
		{"chiihou", func(c *WinContext) { c.WinType, c.Chiihou = Tsumo, true }, false},
		{"haitei tsumo", func(c *WinContext) { c.WinType, c.Haitei = Tsumo, true }, false},
		{"houtei ron", func(c *WinContext) { c.Houtei = true }, false},
		{"chankan ron", func(c *WinContext) { c.Chankan = true }, false},
		{"five dora indicators", func(c *WinContext) { c.DoraIndicators = mustTiles(t, "1m", "2m", "3m", "4m", "5m") }, false},
		{"honba", func(c *WinContext) { c.Honba = 3 }, false},

		{"round wind is not a wind", func(c *WinContext) { c.RoundWind = mustKind(t, "5z") }, true},
		{"seat wind is not a wind", func(c *WinContext) { c.SeatWind = mustKind(t, "1m") }, true},
		{"unknown win type", func(c *WinContext) { c.WinType = WinType(9) }, true},
		{"double riichi without riichi", func(c *WinContext) { c.DoubleRiichi = true }, true},
		{"ippatsu without riichi", func(c *WinContext) { c.Ippatsu = true }, true},
		{"ura without riichi", func(c *WinContext) { c.UraDoraIndicators = mustTiles(t, "1m") }, true},
		{"more ura than dora indicators", func(c *WinContext) {
			c.Riichi = true
			c.UraDoraIndicators = mustTiles(t, "1m", "2m")
		}, true},
		{"no dora indicator", func(c *WinContext) { c.DoraIndicators = nil }, true},
		{"six dora indicators", func(c *WinContext) { c.DoraIndicators = mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m") }, true},
		{"negative honba", func(c *WinContext) { c.Honba = -1 }, true},
		{"haitei on ron", func(c *WinContext) { c.Haitei = true }, true},
		{"houtei on tsumo", func(c *WinContext) { c.WinType, c.Houtei = Tsumo, true }, true},
		{"chankan on tsumo", func(c *WinContext) { c.WinType, c.Chankan = Tsumo, true }, true},
		{"chankan and houtei", func(c *WinContext) { c.Chankan, c.Houtei = true, true }, true},
		{"rinshan is out of scope", func(c *WinContext) { c.WinType, c.Rinshan = Tsumo, true }, true},
		{"tenhou for a non-dealer", func(c *WinContext) { c.WinType, c.Tenhou = Tsumo, true }, true},
		{"tenhou on ron", func(c *WinContext) { c.SeatWind, c.Tenhou = c.RoundWind, true }, true},
		{"tenhou with riichi", func(c *WinContext) {
			c.SeatWind, c.WinType, c.Tenhou, c.Riichi = c.RoundWind, Tsumo, true, true
		}, true},
		{"chiihou for the dealer", func(c *WinContext) { c.SeatWind, c.WinType, c.Chiihou = c.RoundWind, Tsumo, true }, true},
		{"chiihou on ron", func(c *WinContext) { c.Chiihou = true }, true},
		{"chiihou with riichi", func(c *WinContext) { c.WinType, c.Chiihou, c.Riichi = Tsumo, true, true }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := baseContext(t)
			tt.mutate(&c)
			err := c.Validate()
			if tt.wantErr != (err != nil) {
				t.Fatalf("Validate() = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ErrInvalidContext) {
				t.Errorf("Validate() = %v, want ErrInvalidContext", err)
			}
		})
	}
}

func TestParseWinTypeAndWind(t *testing.T) {
	for _, s := range []string{"ron", "tsumo"} {
		w, err := ParseWinType(s)
		if err != nil || w.String() != s {
			t.Errorf("ParseWinType(%q) = %v, %v", s, w, err)
		}
	}
	if _, err := ParseWinType("Ron"); !errors.Is(err, ErrInvalidContext) {
		t.Errorf("ParseWinType(\"Ron\") error = %v, want ErrInvalidContext", err)
	}
	if k, err := ParseWind("3z"); err != nil || k.String() != "3z" {
		t.Errorf("ParseWind(\"3z\") = %v, %v", k, err)
	}
	for _, s := range []string{"5z", "1m", "0p", "", "1z2z"} {
		if _, err := ParseWind(s); !errors.Is(err, ErrInvalidContext) {
			t.Errorf("ParseWind(%q) error = %v, want ErrInvalidContext", s, err)
		}
	}
}
