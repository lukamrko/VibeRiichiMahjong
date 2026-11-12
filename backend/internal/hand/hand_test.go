package hand

import (
	"errors"
	"slices"
	"testing"

	"github.com/lukamrko/VibeRiichiMahjong/backend/internal/tile"
)

func TestParseValid(t *testing.T) {
	h, err := Parse("123m406p789s11122z", "0p")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if got := tile.FormatTiles(h.Tiles()); got != "123m406p789s11122z" {
		t.Errorf("Tiles() = %q", got)
	}
	if w := h.WinningTile(); w.String() != "0p" {
		t.Errorf("WinningTile() = %s, want 0p", w)
	}
	five, _ := tile.KindOf(tile.Pin, 5)
	if c := h.Counts()[five]; c != 1 {
		t.Errorf("Counts()[5p] = %d, want 1 (red five counts as a five)", c)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name, hand, winning string
		want                error
	}{
		{"bad hand syntax", "123x", "1m", tile.ErrSyntax},
		{"bad winning tile", "123m456p789s11122z", "8z", tile.ErrInvalidTile},
		{"winning tile is several tiles", "123m456p789s11122z", "11z", tile.ErrSyntax},
		{"13 tiles", "123m456p789s1112z", "1m", ErrTileCount},
		{"15 tiles", "123m456p789s111222z", "1m", ErrTileCount},
		{"five copies", "11111m456p789s111z", "1m", ErrTooManyCopies},
		{"five copies incl. red", "55550m456p789s111z", "5m", ErrTooManyCopies},
		{"two red fives in a suit", "00m55m456p789s1112z", "0m", ErrDuplicateRedFive},
		{"winning tile missing", "123m456p789s11122z", "9m", ErrWinningTileNotInHand},
		{"red winning tile but only plain five", "123m456p789s11122z", "0p", ErrWinningTileNotInHand},
		{"plain winning tile but only red five", "123m406p789s11122z", "5p", ErrWinningTileNotInHand},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Parse(tt.hand, tt.winning); !errors.Is(err, tt.want) {
				t.Errorf("Parse(%q, %q) error = %v, want %v", tt.hand, tt.winning, err, tt.want)
			}
		})
	}
}

func TestDecompositions(t *testing.T) {
	tests := []struct {
		name, hand, winning string
		want                []string
	}{
		{
			name: "ryanmen (78s waiting 6s)", hand: "123m45699p345678s", winning: "6s",
			want: []string{"standard[123m 456p 345s 678s 99p] win=678s ryanmen"},
		},
		{
			name: "ryanmen (78s waiting 9s)", hand: "123m45699p345789s", winning: "9s",
			want: []string{"standard[123m 456p 345s 789s 99p] win=789s ryanmen"},
		},
		{
			name: "ryanmen (23m waiting 1m)", hand: "123m456m789p789s11z", winning: "1m",
			want: []string{"standard[123m 456m 789p 789s 11z] win=123m ryanmen"},
		},
		{
			name: "penchan (12m waiting 3m)", hand: "123m456m789p789s11z", winning: "3m",
			want: []string{"standard[123m 456m 789p 789s 11z] win=123m penchan"},
		},
		{
			name: "penchan (89p waiting 7p)", hand: "123m456m789p789s11z", winning: "7p",
			want: []string{"standard[123m 456m 789p 789s 11z] win=789p penchan"},
		},
		{
			name: "kanchan (13m waiting 2m)", hand: "123m456m789p789s11z", winning: "2m",
			want: []string{"standard[123m 456m 789p 789s 11z] win=123m kanchan"},
		},
		{
			name: "shanpon (11z + 55z waiting 5z)", hand: "123m456p789s11555z", winning: "5z",
			want: []string{"standard[123m 456p 789s 555z 11z] win=555z shanpon"},
		},
		{
			name: "tanki (waiting 1z)", hand: "123m456p789s11555z", winning: "1z",
			want: []string{"standard[123m 456p 789s 555z 11z] win=11z tanki"},
		},
		{
			name: "red five as winning tile", hand: "123m406p789s11122z", winning: "0p",
			want: []string{"standard[123m 456p 789s 111z 22z] win=456p kanchan"},
		},
		{
			name: "triplets or sequences (111222333m)", hand: "111222333m456p77s", winning: "3m",
			want: []string{
				"standard[111m 222m 333m 456p 77s] win=333m shanpon",
				"standard[123m 123m 123m 456p 77s] win=123m penchan",
			},
		},
		{
			name: "winning tile fits two groups (12m + 33m won on 3m)", hand: "12333m456p789s555z", winning: "3m",
			want: []string{
				"standard[123m 456p 789s 555z 33m] win=123m penchan",
				"standard[123m 456p 789s 555z 33m] win=33m tanki",
			},
		},
		{
			name: "identical groups are deduplicated (iipeikou)", hand: "112233m456p789s55z", winning: "2m",
			want: []string{"standard[123m 123m 456p 789s 55z] win=123m kanchan"},
		},
		{
			name: "ryanpeikou is also chiitoitsu", hand: "112233m445566p77s", winning: "7s",
			want: []string{
				"chiitoitsu[11m 22m 33m 44p 55p 66p 77s] win=77s tanki",
				"standard[123m 123m 456p 456p 77s] win=77s tanki",
			},
		},
		{
			name: "ryanpeikou/chiitoitsu won inside a sequence", hand: "112233m445566p77s", winning: "2m",
			want: []string{
				"chiitoitsu[11m 22m 33m 44p 55p 66p 77s] win=22m tanki",
				"standard[123m 123m 456p 456p 77s] win=123m kanchan",
			},
		},
		{
			name: "four of a kind is not chiitoitsu", hand: "11112233m445566p", winning: "1m",
			want: []string{
				"standard[123m 123m 456p 456p 11m] win=11m tanki",
				"standard[123m 123m 456p 456p 11m] win=123m ryanmen",
			},
		},
		{
			name: "chiitoitsu only", hand: "1199m2288p3377s11z", winning: "1z",
			want: []string{"chiitoitsu[11m 99m 22p 88p 33s 77s 11z] win=11z tanki"},
		},
		{
			name: "kokushi, single wait", hand: "119m19p19s1234567z", winning: "9m",
			want: []string{"kokushi[11m] win=- tanki"},
		},
		{
			name: "kokushi, 13-sided wait", hand: "119m19p19s1234567z", winning: "1m",
			want: []string{"kokushi[11m] win=11m juusanmen"},
		},
		{
			name: "not a winning hand", hand: "123m456p789s12344z", winning: "1z",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := Parse(tt.hand, tt.winning)
			if err != nil {
				t.Fatalf("Parse(%q, %q) error: %v", tt.hand, tt.winning, err)
			}
			var got []string
			for _, d := range h.Decompositions() {
				got = append(got, d.String())
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("Decompositions() =\n  %q\nwant\n  %q", got, tt.want)
			}
		})
	}
}

func TestDecompositionWinningGroup(t *testing.T) {
	h, err := Parse("123m45699p345678s", "6s")
	if err != nil {
		t.Fatal(err)
	}
	ds := h.Decompositions()
	if len(ds) != 1 {
		t.Fatalf("got %d decompositions, want 1", len(ds))
	}
	d := ds[0]
	if d.Form != Standard || len(d.Groups) != 5 {
		t.Fatalf("got form %v with %d groups", d.Form, len(d.Groups))
	}
	if d.Groups[4].Kind != Pair {
		t.Errorf("last group should be the pair, got %v", d.Groups[4])
	}
	w := d.Groups[d.WinningGroup]
	if w.String() != "678s" || !w.Contains(h.WinningTile().Kind) {
		t.Errorf("winning group = %v", w)
	}
}
