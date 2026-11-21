package score

import "testing"

func TestBasePoints(t *testing.T) {
	tests := []struct {
		han, fu   int
		wantBase  int
		wantLimit Limit
	}{
		{1, 30, 240, NoLimit},
		{2, 20, 320, NoLimit},
		{2, 25, 400, NoLimit},
		{3, 60, 1920, NoLimit}, // no kiriage mangan
		{4, 30, 1920, NoLimit}, // no kiriage mangan
		{3, 70, 2000, Mangan},
		{4, 40, 2000, Mangan},
		{5, 20, 2000, Mangan},
		{6, 30, 3000, Haneman},
		{7, 30, 3000, Haneman},
		{8, 30, 4000, Baiman},
		{10, 30, 4000, Baiman},
		{11, 30, 6000, Sanbaiman},
		{12, 30, 6000, Sanbaiman},
		{13, 30, 8000, KazoeYakuman},
		{20, 30, 8000, KazoeYakuman}, // kazoe is capped at a single yakuman
	}
	for _, tt := range tests {
		base, limit := basePoints(tt.han, tt.fu)
		if base != tt.wantBase || limit != tt.wantLimit {
			t.Errorf("basePoints(%d, %d) = %d, %v; want %d, %v", tt.han, tt.fu, base, limit, tt.wantBase, tt.wantLimit)
		}
	}
}

func TestYakumanLimit(t *testing.T) {
	want := []Limit{Yakuman, DoubleYakuman, TripleYakuman, QuadrupleYakuman, QuintupleYakuman, SextupleYakuman}
	for i, w := range want {
		if got := yakumanLimit(i + 1); got != w {
			t.Errorf("yakumanLimit(%d) = %v, want %v", i+1, got, w)
		}
	}
}

func TestPayout(t *testing.T) {
	tests := []struct {
		name   string
		base   int
		dealer bool
		win    WinType
		honba  int
		want   Points
	}{
		{"non-dealer ron", 240, false, Ron, 0, Points{Total: 1000, FromDiscarder: 1000}},
		{"dealer ron", 240, true, Ron, 0, Points{Total: 1500, FromDiscarder: 1500}},
		{"non-dealer tsumo", 240, false, Tsumo, 0, Points{Total: 1100, FromDealer: 500, FromNonDealer: 300}},
		{"dealer tsumo", 240, true, Tsumo, 0, Points{Total: 1500, FromEachNonDealer: 500}},
		{"ron honba", 2000, false, Ron, 2, Points{Total: 8600, FromDiscarder: 8600}},
		{"non-dealer tsumo honba", 2000, false, Tsumo, 1, Points{Total: 8300, FromDealer: 4100, FromNonDealer: 2100}},
		{"dealer tsumo honba", 2000, true, Tsumo, 1, Points{Total: 12300, FromEachNonDealer: 4100}},
		{"dealer ron mangan", 2000, true, Ron, 0, Points{Total: 12000, FromDiscarder: 12000}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WinContext{SeatWind: South, WinType: tt.win, Honba: tt.honba}
			if tt.dealer {
				ctx.SeatWind = East
			}
			if got := payout(tt.base, ctx); got != tt.want {
				t.Errorf("payout = %+v, want %+v", got, tt.want)
			}
		})
	}
}
