// Package tile models Riichi Mahjong tiles and their MPSZ notation
// (see docs/domain/glossary.md).
package tile

import (
	"errors"
	"fmt"
	"strings"
)

// Suit is one of the four tile families.
type Suit uint8

// The four suits, in MPSZ order.
const (
	Man   Suit = iota // m: manzu (characters)
	Pin               // p: pinzu (circles)
	Sou               // s: souzu (bamboo)
	Honor             // z: winds 1–4 (E S W N), dragons 5–7 (Haku Hatsu Chun)
)

// Letter returns the MPSZ suit letter.
func (s Suit) Letter() byte {
	switch s {
	case Man:
		return 'm'
	case Pin:
		return 'p'
	case Sou:
		return 's'
	case Honor:
		return 'z'
	default:
		return '?'
	}
}

func (s Suit) String() string { return string(s.Letter()) }

func suitFromLetter(c byte) (Suit, bool) {
	switch c {
	case 'm':
		return Man, true
	case 'p':
		return Pin, true
	case 's':
		return Sou, true
	case 'z':
		return Honor, true
	default:
		return 0, false
	}
}

// Kind is one of the 34 distinct tile kinds, ignoring red fives.
// Kinds 0–8 are 1m–9m, 9–17 are 1p–9p, 18–26 are 1s–9s and 27–33 are 1z–7z.
type Kind uint8

// NumKinds is the number of distinct tile kinds.
const NumKinds = 34

// Errors returned by the constructors and parsers. Use errors.Is to test for them.
var (
	ErrEmpty         = errors.New("empty tile string")
	ErrSyntax        = errors.New("invalid MPSZ syntax")
	ErrMissingSuit   = errors.New("digits without a suit letter")
	ErrMissingDigits = errors.New("suit letter without digits")
	ErrInvalidTile   = errors.New("invalid tile")
)

// KindOf returns the kind for a suit and number (1–9 for suits, 1–7 for honors).
func KindOf(s Suit, number int) (Kind, error) {
	switch s {
	case Man, Pin, Sou:
		if number < 1 || number > 9 {
			return 0, fmt.Errorf("%w: %d%s", ErrInvalidTile, number, s)
		}
		return Kind(int(s)*9 + number - 1), nil
	case Honor:
		if number < 1 || number > 7 {
			return 0, fmt.Errorf("%w: %d%s", ErrInvalidTile, number, s)
		}
		return Kind(27 + number - 1), nil
	default:
		return 0, fmt.Errorf("%w: unknown suit %d", ErrInvalidTile, s)
	}
}

// Suit returns the kind's suit.
func (k Kind) Suit() Suit {
	if k >= 27 {
		return Honor
	}
	return Suit(k / 9)
}

// Number returns 1–9 for suited kinds and 1–7 for honors.
func (k Kind) Number() int {
	if k >= 27 {
		return int(k) - 27 + 1
	}
	return int(k)%9 + 1
}

// IsHonor reports whether the kind is a wind or dragon.
func (k Kind) IsHonor() bool { return k.Suit() == Honor }

// IsTerminal reports whether the kind is a suited 1 or 9.
func (k Kind) IsTerminal() bool {
	return !k.IsHonor() && (k.Number() == 1 || k.Number() == 9)
}

// IsTerminalOrHonor reports whether the kind is a terminal or an honor (yaochuuhai).
func (k Kind) IsTerminalOrHonor() bool { return k.IsHonor() || k.IsTerminal() }

// IsSimple reports whether the kind is a suited 2–8 (chunchanpai).
func (k Kind) IsSimple() bool { return !k.IsTerminalOrHonor() }

// String returns the MPSZ id, e.g. "5p".
func (k Kind) String() string {
	return fmt.Sprintf("%d%s", k.Number(), k.Suit())
}

// Tile is a single tile: its kind plus whether it is a red five (aka dora).
// Hand-shape logic only looks at Kind, so a red five and a normal five match.
type Tile struct {
	Kind Kind
	Red  bool
}

// New returns a tile. Red is only allowed on suited fives.
func New(s Suit, number int, red bool) (Tile, error) {
	k, err := KindOf(s, number)
	if err != nil {
		return Tile{}, err
	}
	if red && (s == Honor || number != 5) {
		return Tile{}, fmt.Errorf("%w: only a suited five can be red", ErrInvalidTile)
	}
	return Tile{Kind: k, Red: red}, nil
}

// String returns the MPSZ id; red fives use 0, e.g. "0p".
func (t Tile) String() string {
	if t.Red {
		return "0" + t.Kind.Suit().String()
	}
	return t.Kind.String()
}

func (t Tile) digit() byte {
	if t.Red {
		return '0'
	}
	return byte('0' + t.Kind.Number())
}

// Parse parses a single MPSZ tile id such as "1m", "0p" or "7z".
func Parse(s string) (Tile, error) {
	tiles, err := ParseTiles(s)
	if err != nil {
		return Tile{}, err
	}
	if len(tiles) != 1 {
		return Tile{}, fmt.Errorf("%w: %q is not a single tile", ErrSyntax, s)
	}
	return tiles[0], nil
}

// ParseTiles parses a compact MPSZ string such as "123m406p789s11z" and keeps
// the tiles in the order written. It does not check counts; see package hand.
func ParseTiles(s string) ([]Tile, error) {
	if s == "" {
		return nil, ErrEmpty
	}
	var (
		tiles   []Tile
		pending []byte // digits waiting for their suit letter
	)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			pending = append(pending, c)
			continue
		}
		suit, ok := suitFromLetter(c)
		if !ok {
			return nil, fmt.Errorf("%w: unexpected %q at position %d", ErrSyntax, c, i)
		}
		if len(pending) == 0 {
			return nil, fmt.Errorf("%w: %q at position %d", ErrMissingDigits, c, i)
		}
		for _, d := range pending {
			n := int(d - '0')
			red := n == 0
			if red {
				n = 5
				if suit == Honor {
					return nil, fmt.Errorf("%w: 0z", ErrInvalidTile)
				}
			}
			t, err := New(suit, n, red)
			if err != nil {
				return nil, err
			}
			tiles = append(tiles, t)
		}
		pending = pending[:0]
	}
	if len(pending) > 0 {
		return nil, fmt.Errorf("%w: trailing %q", ErrMissingSuit, pending)
	}
	return tiles, nil
}

// FormatTiles writes tiles in compact MPSZ, keeping their order and grouping
// consecutive tiles of the same suit, e.g. "123m406p789s11z".
func FormatTiles(tiles []Tile) string {
	var b strings.Builder
	for i, t := range tiles {
		b.WriteByte(t.digit())
		if i == len(tiles)-1 || tiles[i+1].Kind.Suit() != t.Kind.Suit() {
			b.WriteByte(t.Kind.Suit().Letter())
		}
	}
	return b.String()
}
