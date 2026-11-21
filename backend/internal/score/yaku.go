package score

import (
	"fmt"
	"slices"

	"github.com/lukamrko/VibeRiichiMahjong/backend/internal/hand"
	"github.com/lukamrko/VibeRiichiMahjong/backend/internal/tile"
)

// Yaku is a scoring pattern. The constant order is the order yaku are listed
// in a Result (it matches the yaku table in fixtures/hands/README.md).
type Yaku uint8

// Regular yaku (closed values), then yakuman.
const (
	Riichi Yaku = iota
	DoubleRiichi
	Ippatsu
	MenzenTsumo
	Pinfu
	Tanyao
	Iipeikou
	YakuhaiHaku
	YakuhaiHatsu
	YakuhaiChun
	YakuhaiSeatWind
	YakuhaiRoundWind
	Haitei
	Houtei
	Chankan
	SanshokuDoujun
	Ittsu
	Chanta
	Junchan
	Ryanpeikou
	Sanankou
	SanshokuDoukou
	Toitoi
	Shousangen
	Honroutou
	Honitsu
	Chinitsu
	Chiitoitsu

	KokushiMusou
	KokushiMusouJuusanmen
	Suuankou
	SuuankouTanki
	Daisangen
	Shousuushii
	Daisuushii
	Tsuuiisou
	Chinroutou
	Ryuuiisou
	ChuurenPoutou
	JunseiChuurenPoutou
	Tenhou
	Chiihou

	numYaku
)

var yakuInfo = [numYaku]struct {
	id  string
	han int
}{
	Riichi:                {"riichi", 1},
	DoubleRiichi:          {"double_riichi", 2},
	Ippatsu:               {"ippatsu", 1},
	MenzenTsumo:           {"menzen_tsumo", 1},
	Pinfu:                 {"pinfu", 1},
	Tanyao:                {"tanyao", 1},
	Iipeikou:              {"iipeikou", 1},
	YakuhaiHaku:           {"yakuhai_haku", 1},
	YakuhaiHatsu:          {"yakuhai_hatsu", 1},
	YakuhaiChun:           {"yakuhai_chun", 1},
	YakuhaiSeatWind:       {"yakuhai_seat_wind", 1},
	YakuhaiRoundWind:      {"yakuhai_round_wind", 1},
	Haitei:                {"haitei", 1},
	Houtei:                {"houtei", 1},
	Chankan:               {"chankan", 1},
	SanshokuDoujun:        {"sanshoku_doujun", 2},
	Ittsu:                 {"ittsu", 2},
	Chanta:                {"chanta", 2},
	Junchan:               {"junchan", 3},
	Ryanpeikou:            {"ryanpeikou", 3},
	Sanankou:              {"sanankou", 2},
	SanshokuDoukou:        {"sanshoku_doukou", 2},
	Toitoi:                {"toitoi", 2},
	Shousangen:            {"shousangen", 2},
	Honroutou:             {"honroutou", 2},
	Honitsu:               {"honitsu", 3},
	Chinitsu:              {"chinitsu", 6},
	Chiitoitsu:            {"chiitoitsu", 2},
	KokushiMusou:          {"kokushi_musou", 13},
	KokushiMusouJuusanmen: {"kokushi_musou_juusanmen", 26},
	Suuankou:              {"suuankou", 13},
	SuuankouTanki:         {"suuankou_tanki", 26},
	Daisangen:             {"daisangen", 13},
	Shousuushii:           {"shousuushii", 13},
	Daisuushii:            {"daisuushii", 26},
	Tsuuiisou:             {"tsuuiisou", 13},
	Chinroutou:            {"chinroutou", 13},
	Ryuuiisou:             {"ryuuiisou", 13},
	ChuurenPoutou:         {"chuuren_poutou", 13},
	JunseiChuurenPoutou:   {"junsei_chuuren_poutou", 26},
	Tenhou:                {"tenhou", 13},
	Chiihou:               {"chiihou", 13},
}

// String returns the yaku id from the closed vocabulary in fixtures/hands/README.md.
func (y Yaku) String() string {
	if y >= numYaku {
		return fmt.Sprintf("Yaku(%d)", uint8(y))
	}
	return yakuInfo[y].id
}

// Han returns the closed-hand han value: 13 for a yakuman, 26 for a double yakuman.
func (y Yaku) Han() int {
	if y >= numYaku {
		panic(fmt.Sprintf("score: unknown yaku %d", y))
	}
	return yakuInfo[y].han
}

// IsYakuman reports whether the yaku is a (single or double) yakuman.
func (y Yaku) IsYakuman() bool { return y >= KokushiMusou && y < numYaku }

// Dragon kinds.
var (
	haku  = honorKind(5)
	hatsu = honorKind(6)
	chun  = honorKind(7)
)

func isDragon(k tile.Kind) bool { return k.IsHonor() && k.Number() >= 5 }

// yakuList collects yaku in detection order.
type yakuList []Yaku

// add appends y when ok holds.
func (l *yakuList) add(ok bool, y Yaku) {
	if ok {
		*l = append(*l, y)
	}
}

// reading is one decomposition prepared for yaku and fu detection.
type reading struct {
	d      hand.Decomposition
	counts [tile.NumKinds]int
	win    tile.Kind
	ctx    WinContext

	sets []hand.Group // standard form only: the 4 sets
	pair hand.Group   // standard form only
	// openTriplet is the index in sets of a triplet completed by ron (shanpon),
	// which counts as open for fu and concealed-triplet yaku; -1 if none.
	openTriplet int
}

func newReading(h hand.Hand, d hand.Decomposition, ctx WinContext) reading {
	r := reading{d: d, counts: h.Counts(), win: h.WinningTile().Kind(), ctx: ctx, openTriplet: -1}
	if d.Form == hand.Standard {
		r.sets, r.pair = d.Groups[:4], d.Groups[4]
		if ctx.WinType == Ron && d.WinningGroup < 4 && d.Groups[d.WinningGroup].Shape == hand.Triplet {
			r.openTriplet = d.WinningGroup
		}
	}
	return r
}

// all reports whether every tile in the hand satisfies pred.
func (r reading) all(pred func(tile.Kind) bool) bool {
	for k, c := range r.counts {
		if c > 0 && !pred(tile.Kind(k)) {
			return false
		}
	}
	return true
}

func (r reading) any(pred func(tile.Kind) bool) bool {
	return !r.all(func(k tile.Kind) bool { return !pred(k) })
}

// setsOf returns the sets of this reading that have the given shape.
func (r reading) setsOf(shape hand.Shape) []hand.Group {
	var out []hand.Group
	for _, g := range r.sets {
		if g.Shape == shape {
			out = append(out, g)
		}
	}
	return out
}

func (r reading) triplets() []hand.Group  { return r.setsOf(hand.Triplet) }
func (r reading) sequences() []hand.Group { return r.setsOf(hand.Sequence) }

// countTriplets returns how many triplets have a kind satisfying pred.
func (r reading) countTriplets(pred func(tile.Kind) bool) int {
	n := 0
	for _, g := range r.triplets() {
		if pred(g.First) {
			n++
		}
	}
	return n
}

func (r reading) concealedTriplets() int {
	n := len(r.triplets())
	if r.openTriplet >= 0 {
		n--
	}
	return n
}

func (r reading) isValueHonor(k tile.Kind) bool {
	return isDragon(k) || r.ctx.isValueWind(k)
}

// yakuman returns the yakuman of this reading, in canonical order.
func (r reading) yakuman() []Yaku {
	var out yakuList
	add := out.add
	switch r.d.Form {
	case hand.Kokushi:
		add(r.d.Wait == hand.Juusanmen, KokushiMusouJuusanmen)
		add(r.d.Wait != hand.Juusanmen, KokushiMusou)
	case hand.Standard:
		if r.concealedTriplets() == 4 {
			add(r.d.Wait == hand.Tanki, SuuankouTanki)
			add(r.d.Wait != hand.Tanki, Suuankou)
		}
		dragons, winds := r.countTriplets(isDragon), r.countTriplets(isWind)
		add(dragons == 3, Daisangen)
		add(winds == 3 && isWind(r.pair.First), Shousuushii)
		add(winds == 4, Daisuushii)
		if pure, ok := r.chuuren(); ok {
			add(pure, JunseiChuurenPoutou)
			add(!pure, ChuurenPoutou)
		}
	case hand.Chiitoitsu:
	default:
		panic(fmt.Sprintf("score: unknown form %v", r.d.Form))
	}
	add(r.all(tile.Kind.IsHonor), Tsuuiisou)
	add(r.all(tile.Kind.IsTerminal), Chinroutou)
	add(r.all(isGreen), Ryuuiisou)
	add(r.ctx.Tenhou, Tenhou)
	add(r.ctx.Chiihou, Chiihou)
	slices.Sort(out)
	return out
}

func isGreen(k tile.Kind) bool {
	if k == hatsu {
		return true
	}
	if k.Suit() != tile.Sou {
		return false
	}
	switch k.Number() {
	case 2, 3, 4, 6, 8:
		return true
	default:
		return false
	}
}

// chuuren reports whether the hand is chuuren poutou (1112345678999 + one
// more tile of the suit) and whether it is the pure nine-sided wait.
func (r reading) chuuren() (pure, ok bool) {
	suit := r.win.Suit()
	if suit == tile.Honor || !r.all(func(k tile.Kind) bool { return k.Suit() == suit }) {
		return false, false
	}
	first := tile.Kind(int(suit) * 9)
	base := [9]int{3, 1, 1, 1, 1, 1, 1, 1, 3}
	extra := -1
	for i, need := range base {
		c := r.counts[first+tile.Kind(i)]
		switch {
		case c == need:
		case c == need+1 && extra < 0:
			extra = i
		default:
			return false, false
		}
	}
	return first+tile.Kind(extra) == r.win, true
}

// regularYaku returns the non-yakuman yaku of this reading, in canonical order.
func (r reading) regularYaku() []Yaku {
	var out yakuList
	add := out.add
	c := r.ctx
	tsumo := c.WinType == Tsumo
	add(c.Riichi && !c.DoubleRiichi, Riichi)
	add(c.DoubleRiichi, DoubleRiichi)
	add(c.Ippatsu, Ippatsu)
	add(tsumo, MenzenTsumo)
	add(c.Haitei, Haitei)
	add(c.Houtei, Houtei)
	add(c.Chankan, Chankan)
	add(r.all(tile.Kind.IsSimple), Tanyao)
	add(r.all(tile.Kind.IsTerminalOrHonor), Honroutou)
	if r.singleSuit() {
		hasHonor := r.any(tile.Kind.IsHonor)
		add(hasHonor, Honitsu)
		add(!hasHonor, Chinitsu)
	}

	switch r.d.Form {
	case hand.Chiitoitsu:
		add(true, Chiitoitsu)
	case hand.Standard:
		r.standardYaku(add)
	case hand.Kokushi:
		// Kokushi is always a yakuman; it has no regular yaku of its own.
	default:
		panic(fmt.Sprintf("score: unknown form %v", r.d.Form))
	}
	slices.Sort(out)
	return out
}

// singleSuit reports whether all suited tiles share one suit (honors allowed)
// and the hand has at least one suited tile.
func (r reading) singleSuit() bool {
	suit, seen := tile.Suit(0), false
	for k, n := range r.counts {
		kind := tile.Kind(k)
		if n == 0 || kind.IsHonor() {
			continue
		}
		if seen && kind.Suit() != suit {
			return false
		}
		suit, seen = kind.Suit(), true
	}
	return seen
}

func (r reading) isPinfu() bool {
	return len(r.sequences()) == 4 && r.d.Wait == hand.Ryanmen && !r.isValueHonor(r.pair.First)
}

func (r reading) standardYaku(add func(bool, Yaku)) {
	seqs := r.sequences()
	trips := r.triplets()

	add(r.isPinfu(), Pinfu)

	// Iipeikou / ryanpeikou: count pairs of identical sequences.
	seqsByFirst := map[tile.Kind]int{}
	for _, g := range seqs {
		seqsByFirst[g.First]++
	}
	peikou := 0 // number of pairs of identical sequences
	for _, n := range seqsByFirst {
		peikou += n / 2
	}
	add(peikou == 1, Iipeikou)
	add(peikou == 2, Ryanpeikou)

	for _, g := range trips {
		add(g.First == haku, YakuhaiHaku)
		add(g.First == hatsu, YakuhaiHatsu)
		add(g.First == chun, YakuhaiChun)
		add(g.First == r.ctx.SeatWind.Kind(), YakuhaiSeatWind)
		add(g.First == r.ctx.RoundWind.Kind(), YakuhaiRoundWind)
	}

	add(sameNumberAllSuits(seqs), SanshokuDoujun)
	add(sameNumberAllSuits(trips), SanshokuDoukou)
	add(hasIttsu(seqs), Ittsu)

	if len(seqs) > 0 && r.everyGroup(func(g hand.Group) bool { return groupHas(g, tile.Kind.IsTerminalOrHonor) }) {
		hasHonor := r.any(tile.Kind.IsHonor)
		add(hasHonor, Chanta)
		add(!hasHonor, Junchan)
	}

	add(r.concealedTriplets() == 3, Sanankou)
	add(len(trips) == 4, Toitoi)

	add(r.countTriplets(isDragon) == 2 && isDragon(r.pair.First), Shousangen)
}

func (r reading) everyGroup(pred func(hand.Group) bool) bool {
	return pred(r.pair) && !slices.ContainsFunc(r.sets, func(g hand.Group) bool { return !pred(g) })
}

func groupHas(g hand.Group, pred func(tile.Kind) bool) bool {
	if g.Shape == hand.Sequence {
		return pred(g.First) || pred(g.First+2)
	}
	return pred(g.First)
}

// sameNumberAllSuits reports whether groups (all sequences or all triplets)
// include the same number in man, pin and sou.
func sameNumberAllSuits(groups []hand.Group) bool {
	var seen [10][3]bool
	for _, g := range groups {
		if g.First.IsHonor() {
			continue
		}
		seen[g.First.Number()][g.First.Suit()] = true
	}
	for _, s := range seen {
		if s[0] && s[1] && s[2] {
			return true
		}
	}
	return false
}

func hasIttsu(seqs []hand.Group) bool {
	var seen [3][10]bool
	for _, g := range seqs {
		seen[g.First.Suit()][g.First.Number()] = true
	}
	for _, s := range seen {
		if s[1] && s[4] && s[7] {
			return true
		}
	}
	return false
}
