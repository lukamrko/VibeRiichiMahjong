// Package hand models a closed 14-tile winning hand and enumerates its
// decompositions and wait types (see docs/domain/glossary.md).
package hand

import (
	"errors"
	"fmt"
	"slices"

	"github.com/lukamrko/VibeRiichiMahjong/backend/internal/tile"
)

// Size is the number of tiles in a closed winning hand (v1: no kan).
const Size = 14

// Validation errors returned by New and Parse. Use errors.Is to test for them.
var (
	ErrTileCount            = errors.New("hand must have exactly 14 tiles")
	ErrTooManyCopies        = errors.New("more than 4 copies of a tile")
	ErrDuplicateRedFive     = errors.New("more than one red five in a suit")
	ErrWinningTileNotInHand = errors.New("winning tile is not in the hand")
)

// Hand is a validated closed hand of 14 tiles, including the winning tile.
type Hand struct {
	tiles   []tile.Tile
	winning tile.Tile
	counts  [tile.NumKinds]int
}

// New validates the tiles and returns a Hand. The tiles include the winning tile.
// A red winning tile must match a red five in the hand, and a plain one a plain tile.
func New(tiles []tile.Tile, winning tile.Tile) (Hand, error) {
	if len(tiles) != Size {
		return Hand{}, fmt.Errorf("%w: got %d", ErrTileCount, len(tiles))
	}
	var (
		counts [tile.NumKinds]int
		reds   [3]int // per suited Suit
	)
	for _, t := range tiles {
		counts[t.Kind]++
		if t.Red {
			reds[t.Kind.Suit()]++
		}
	}
	for k, c := range counts {
		if c > 4 {
			return Hand{}, fmt.Errorf("%w: %d × %s", ErrTooManyCopies, c, tile.Kind(k))
		}
	}
	for s, r := range reds {
		if r > 1 {
			return Hand{}, fmt.Errorf("%w: %d × 0%s", ErrDuplicateRedFive, r, tile.Suit(s))
		}
	}
	if !slices.Contains(tiles, winning) {
		return Hand{}, fmt.Errorf("%w: %s", ErrWinningTileNotInHand, winning)
	}
	return Hand{tiles: slices.Clone(tiles), winning: winning, counts: counts}, nil
}

// Parse parses a compact MPSZ hand (14 tiles, including the winning tile) and
// the winning tile id, then validates them like New.
func Parse(hand, winning string) (Hand, error) {
	tiles, err := tile.ParseTiles(hand)
	if err != nil {
		return Hand{}, fmt.Errorf("hand: %w", err)
	}
	w, err := tile.Parse(winning)
	if err != nil {
		return Hand{}, fmt.Errorf("winning tile: %w", err)
	}
	return New(tiles, w)
}

// Tiles returns a copy of the hand's tiles in their original order.
func (h Hand) Tiles() []tile.Tile { return slices.Clone(h.tiles) }

// WinningTile returns the tile that completed the hand.
func (h Hand) WinningTile() tile.Tile { return h.winning }

// Counts returns how many tiles of each kind the hand holds (red fives count as fives).
func (h Hand) Counts() [tile.NumKinds]int { return h.counts }
