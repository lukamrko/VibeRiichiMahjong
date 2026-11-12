package hand

import (
	"fmt"
	"slices"
	"strings"

	"github.com/lukamrko/VibeRiichiMahjong/backend/internal/tile"
)

// GroupKind is the shape of a group of tiles.
type GroupKind uint8

// Group kinds. The order is used when sorting groups.
const (
	Sequence GroupKind = iota // shuntsu, e.g. 345p
	Triplet                   // koutsu, e.g. 777s
	Pair                      // jantou or one of the chiitoitsu pairs
)

func (g GroupKind) String() string {
	switch g {
	case Sequence:
		return "sequence"
	case Triplet:
		return "triplet"
	case Pair:
		return "pair"
	default:
		return fmt.Sprintf("GroupKind(%d)", uint8(g))
	}
}

// Group is a sequence, triplet or pair. First is the lowest tile kind of a
// sequence, or the tile kind of a triplet or pair.
type Group struct {
	Kind  GroupKind
	First tile.Kind
}

// Contains reports whether the group holds a tile of kind k.
func (g Group) Contains(k tile.Kind) bool {
	switch g.Kind {
	case Sequence:
		return k >= g.First && k <= g.First+2
	case Triplet, Pair:
		return k == g.First
	default:
		return false
	}
}

// String returns the group in compact MPSZ, e.g. "123m", "777s", "55p".
func (g Group) String() string {
	n := g.First.Number()
	suit := g.First.Suit()
	switch g.Kind {
	case Sequence:
		return fmt.Sprintf("%d%d%d%s", n, n+1, n+2, suit)
	case Triplet:
		return fmt.Sprintf("%d%d%d%s", n, n, n, suit)
	case Pair:
		return fmt.Sprintf("%d%d%s", n, n, suit)
	default:
		return fmt.Sprintf("Group(%d,%s)", uint8(g.Kind), g.First)
	}
}

func compareGroups(a, b Group) int {
	if a.Kind == Pair && b.Kind != Pair {
		return 1
	}
	if b.Kind == Pair && a.Kind != Pair {
		return -1
	}
	if a.First != b.First {
		return int(a.First) - int(b.First)
	}
	return int(a.Kind) - int(b.Kind)
}

// Form is the overall shape of a winning hand.
type Form uint8

// Winning forms in v1 (closed hands).
const (
	Standard   Form = iota // 4 sets + 1 pair
	Chiitoitsu             // 7 distinct pairs
	Kokushi                // kokushi musou: 13 terminals/honors plus one duplicate
)

func (f Form) String() string {
	switch f {
	case Standard:
		return "standard"
	case Chiitoitsu:
		return "chiitoitsu"
	case Kokushi:
		return "kokushi"
	default:
		return fmt.Sprintf("Form(%d)", uint8(f))
	}
}

// Wait is the shape (machi) that the winning tile completed.
type Wait uint8

// Wait types.
const (
	Ryanmen   Wait = iota // two-sided sequence wait, e.g. 45 waiting 3/6
	Kanchan               // closed (middle) wait, e.g. 46 waiting 5
	Penchan               // edge wait: 12 waiting 3, or 89 waiting 7
	Shanpon               // two pairs, one becomes a triplet
	Tanki                 // single-tile wait on the pair; also chiitoitsu and single-wait kokushi
	Juusanmen             // kokushi musou 13-sided wait (the winning tile completed the pair)
)

func (w Wait) String() string {
	switch w {
	case Ryanmen:
		return "ryanmen"
	case Kanchan:
		return "kanchan"
	case Penchan:
		return "penchan"
	case Shanpon:
		return "shanpon"
	case Tanki:
		return "tanki"
	case Juusanmen:
		return "juusanmen"
	default:
		return fmt.Sprintf("Wait(%d)", uint8(w))
	}
}

// Decomposition is one reading of a winning hand: how its tiles split into
// groups, and which group the winning tile completed with which wait.
//
//   - Standard: Groups holds the 4 sets (sorted) followed by the pair.
//   - Chiitoitsu: Groups holds the 7 pairs (sorted).
//   - Kokushi: Groups holds only the duplicated tile as a pair. WinningGroup is 0
//     when the winning tile completed that pair (Juusanmen), otherwise -1 (Tanki).
//
// The same split of tiles can appear several times with a different
// WinningGroup when the winning tile fits more than one group; scoring keeps
// the highest-scoring reading.
type Decomposition struct {
	Form         Form
	Groups       []Group
	WinningGroup int
	Wait         Wait
}

// String returns a stable, human-readable form used in tests and logs, e.g.
// "standard[123m 456p 345s 678s 99p] win=678s ryanmen".
func (d Decomposition) String() string {
	parts := make([]string, len(d.Groups))
	for i, g := range d.Groups {
		parts[i] = g.String()
	}
	win := "-"
	if d.WinningGroup >= 0 && d.WinningGroup < len(d.Groups) {
		win = d.Groups[d.WinningGroup].String()
	}
	return fmt.Sprintf("%s[%s] win=%s %s", d.Form, strings.Join(parts, " "), win, d.Wait)
}

// Decompositions returns every reading of the hand as a winning hand, sorted by
// String and without duplicates. It returns nil if the hand is not complete.
func (h Hand) Decompositions() []Decomposition {
	win := h.winning.Kind
	var out []Decomposition
	for _, groups := range standardSplits(h.counts) {
		for i, g := range groups {
			if g.Contains(win) {
				out = append(out, Decomposition{Form: Standard, Groups: groups, WinningGroup: i, Wait: waitOf(g, win)})
			}
		}
	}
	if pairs, ok := chiitoitsuPairs(h.counts); ok {
		i := slices.IndexFunc(pairs, func(g Group) bool { return g.First == win })
		out = append(out, Decomposition{Form: Chiitoitsu, Groups: pairs, WinningGroup: i, Wait: Tanki})
	}
	if pair, ok := kokushiPair(h.counts); ok {
		d := Decomposition{Form: Kokushi, Groups: []Group{pair}, WinningGroup: -1, Wait: Tanki}
		if pair.First == win {
			d.WinningGroup, d.Wait = 0, Juusanmen
		}
		out = append(out, d)
	}
	slices.SortFunc(out, func(a, b Decomposition) int { return strings.Compare(a.String(), b.String()) })
	return slices.CompactFunc(out, func(a, b Decomposition) bool { return a.String() == b.String() })
}

// waitOf classifies the wait for the winning tile kind inside group g.
func waitOf(g Group, win tile.Kind) Wait {
	switch g.Kind {
	case Pair:
		return Tanki
	case Triplet:
		return Shanpon
	case Sequence:
		switch {
		case win == g.First+1:
			return Kanchan
		case win == g.First && g.First.Number() == 7: // 89 waiting 7
			return Penchan
		case win == g.First+2 && g.First.Number() == 1: // 12 waiting 3
			return Penchan
		default:
			return Ryanmen
		}
	default:
		panic(fmt.Sprintf("hand: unknown group kind %d", g.Kind))
	}
}

// standardSplits returns every distinct split of counts into 4 sets + 1 pair.
// Each result has its sets sorted, followed by the pair.
func standardSplits(counts [tile.NumKinds]int) [][]Group {
	var out [][]Group
	for k := range tile.Kind(tile.NumKinds) {
		if counts[k] < 2 {
			continue
		}
		c := counts
		c[k] -= 2
		splitSets(&c, 0, nil, func(sets []Group) {
			groups := append(slices.Clone(sets), Group{Kind: Pair, First: k})
			out = append(out, groups)
		})
	}
	return out
}

// splitSets removes sets from counts, always taking the lowest remaining kind
// first, and calls emit for every way to use up all tiles. Sets come out sorted.
func splitSets(c *[tile.NumKinds]int, from tile.Kind, sets []Group, emit func([]Group)) {
	k := from
	for k < tile.NumKinds && c[k] == 0 {
		k++
	}
	if k == tile.NumKinds {
		emit(sets)
		return
	}
	if c[k] >= 3 {
		c[k] -= 3
		splitSets(c, k, append(sets, Group{Kind: Triplet, First: k}), emit)
		c[k] += 3
	}
	if !k.IsHonor() && k.Number() <= 7 && c[k+1] > 0 && c[k+2] > 0 {
		c[k]--
		c[k+1]--
		c[k+2]--
		splitSets(c, k, append(sets, Group{Kind: Sequence, First: k}), emit)
		c[k]++
		c[k+1]++
		c[k+2]++
	}
}

// chiitoitsuPairs reports whether counts are exactly 7 distinct pairs
// (four of a kind is not two pairs under Tenhou rules).
func chiitoitsuPairs(counts [tile.NumKinds]int) ([]Group, bool) {
	var pairs []Group
	for k, c := range counts {
		switch c {
		case 0:
		case 2:
			pairs = append(pairs, Group{Kind: Pair, First: tile.Kind(k)})
		default:
			return nil, false
		}
	}
	return pairs, len(pairs) == 7
}

// kokushiPair reports whether counts form kokushi musou and returns the
// duplicated terminal/honor as a pair.
func kokushiPair(counts [tile.NumKinds]int) (Group, bool) {
	var pair Group
	pairs := 0
	for k, c := range counts {
		kind := tile.Kind(k)
		switch {
		case !kind.IsTerminalOrHonor():
			if c != 0 {
				return Group{}, false
			}
		case c == 0:
			return Group{}, false
		case c == 1:
		case c == 2:
			pairs++
			pair = Group{Kind: Pair, First: kind}
		default:
			return Group{}, false
		}
	}
	return pair, pairs == 1
}
