package tile

import (
	"errors"
	"testing"
)

func TestKindOfAndProperties(t *testing.T) {
	tests := []struct {
		suit                    Suit
		number                  int
		wantKind                Kind
		wantHonor, wantTerminal bool
		wantTerminalOrHonor     bool
		wantString              string
	}{
		{Man, 1, 0, false, true, true, "1m"},
		{Man, 5, 4, false, false, false, "5m"},
		{Man, 9, 8, false, true, true, "9m"},
		{Pin, 1, 9, false, true, true, "1p"},
		{Sou, 9, 26, false, true, true, "9s"},
		{Honor, 1, 27, true, false, true, "1z"},
		{Honor, 7, 33, true, false, true, "7z"},
	}
	for _, tt := range tests {
		k, err := KindOf(tt.suit, tt.number)
		if err != nil {
			t.Fatalf("KindOf(%v, %d) error: %v", tt.suit, tt.number, err)
		}
		if k != tt.wantKind {
			t.Errorf("KindOf(%v, %d) = %d, want %d", tt.suit, tt.number, k, tt.wantKind)
		}
		if k.Suit() != tt.suit || k.Number() != tt.number {
			t.Errorf("kind %d: Suit/Number = %v/%d, want %v/%d", k, k.Suit(), k.Number(), tt.suit, tt.number)
		}
		if k.IsHonor() != tt.wantHonor {
			t.Errorf("%s.IsHonor() = %v", k, k.IsHonor())
		}
		if k.IsTerminal() != tt.wantTerminal {
			t.Errorf("%s.IsTerminal() = %v", k, k.IsTerminal())
		}
		if k.IsTerminalOrHonor() != tt.wantTerminalOrHonor {
			t.Errorf("%s.IsTerminalOrHonor() = %v", k, k.IsTerminalOrHonor())
		}
		if k.String() != tt.wantString {
			t.Errorf("String() = %q, want %q", k.String(), tt.wantString)
		}
	}
}

func TestKindOfInvalid(t *testing.T) {
	tests := []struct {
		suit   Suit
		number int
	}{
		{Man, 0}, {Man, 10}, {Honor, 0}, {Honor, 8}, {Suit(9), 1},
	}
	for _, tt := range tests {
		if _, err := KindOf(tt.suit, tt.number); !errors.Is(err, ErrInvalidTile) {
			t.Errorf("KindOf(%v, %d) error = %v, want ErrInvalidTile", tt.suit, tt.number, err)
		}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		in       string
		wantKind string
		wantRed  bool
	}{
		{"1m", "1m", false},
		{"5p", "5p", false},
		{"0p", "5p", true},
		{"0m", "5m", true},
		{"0s", "5s", true},
		{"7z", "7z", false},
	}
	for _, tt := range tests {
		got, err := Parse(tt.in)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", tt.in, err)
		}
		if got.Kind.String() != tt.wantKind || got.Red != tt.wantRed {
			t.Errorf("Parse(%q) = {%s, red=%v}, want {%s, red=%v}", tt.in, got.Kind, got.Red, tt.wantKind, tt.wantRed)
		}
		if got.String() != tt.in {
			t.Errorf("Parse(%q).String() = %q", tt.in, got.String())
		}
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		in   string
		want error
	}{
		{"", ErrEmpty},
		{"1", ErrMissingSuit},
		{"m", ErrMissingDigits},
		{"11m", ErrSyntax},
		{"1x", ErrSyntax},
		{"8z", ErrInvalidTile},
		{"0z", ErrInvalidTile},
	}
	for _, tt := range tests {
		if _, err := Parse(tt.in); !errors.Is(err, tt.want) {
			t.Errorf("Parse(%q) error = %v, want %v", tt.in, err, tt.want)
		}
	}
}

func TestParseTiles(t *testing.T) {
	got, err := ParseTiles("123m406p789s11z")
	if err != nil {
		t.Fatalf("ParseTiles error: %v", err)
	}
	want := []string{"1m", "2m", "3m", "4p", "0p", "6p", "7s", "8s", "9s", "1z", "1z"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].String() != want[i] {
			t.Errorf("tile %d = %s, want %s", i, got[i], want[i])
		}
	}
	if !got[4].Red || got[4].Kind.Number() != 5 {
		t.Errorf("tile 4 should be red 5p, got %+v", got[4])
	}
}

func TestParseTilesErrors(t *testing.T) {
	tests := []struct {
		in   string
		want error
	}{
		{"", ErrEmpty},
		{"123", ErrMissingSuit},
		{"123m45", ErrMissingSuit},
		{"m123", ErrMissingDigits},
		{"12mp", ErrMissingDigits},
		{"12x", ErrSyntax},
		{"12 3m", ErrSyntax},
		{"12M", ErrSyntax},
		{"89z", ErrInvalidTile},
		{"0z", ErrInvalidTile},
	}
	for _, tt := range tests {
		if _, err := ParseTiles(tt.in); !errors.Is(err, tt.want) {
			t.Errorf("ParseTiles(%q) error = %v, want %v", tt.in, err, tt.want)
		}
	}
}

func TestFormatTiles(t *testing.T) {
	tests := []string{
		"123m406p789s11z",
		"19m19p19s1234567z",
		"5m",
		"1m1p1m",
	}
	for _, in := range tests {
		tiles, err := ParseTiles(in)
		if err != nil {
			t.Fatalf("ParseTiles(%q): %v", in, err)
		}
		if got := FormatTiles(tiles); got != in {
			t.Errorf("FormatTiles(ParseTiles(%q)) = %q", in, got)
		}
	}
	if got := FormatTiles(nil); got != "" {
		t.Errorf("FormatTiles(nil) = %q, want empty", got)
	}
}
