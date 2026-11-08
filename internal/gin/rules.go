// Package gin holds the rules of Gin Rummy and the computer players.
// It has no graphics or sound, so it can be tested on its own.
package gin

import "sort"

// Suit is one of 'S', 'H', 'D', 'C'.
type Suit byte

const (
	Spades   Suit = 'S'
	Hearts   Suit = 'H'
	Diamonds Suit = 'D'
	Clubs    Suit = 'C'
)

// Suits in the order a hand is sorted.
var Suits = []Suit{Spades, Hearts, Clubs, Diamonds}

func (s Suit) Name() string {
	switch s {
	case Spades:
		return "Spades"
	case Hearts:
		return "Hearts"
	case Diamonds:
		return "Diamonds"
	case Clubs:
		return "Clubs"
	}
	return ""
}

func (s Suit) Symbol() string {
	switch s {
	case Spades:
		return "♠"
	case Hearts:
		return "♥"
	case Diamonds:
		return "♦"
	case Clubs:
		return "♣"
	}
	return ""
}

func suitIdx(s Suit) int {
	for i, x := range Suits {
		if x == s {
			return i
		}
	}
	return 0
}

// Card ranks run 1..13 (A=1, J=11, Q=12, K=13). Aces are always low in gin.
type Card struct {
	S Suit
	R int
}

// ID is the card's art name ("H11"); the art uses 14 for aces.
func (c Card) ID() string {
	r := c.R
	if r == 1 {
		r = 14
	}
	return string(c.S) + itoa(r)
}

func RankName(r int) string {
	switch r {
	case 1:
		return "A"
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	}
	return itoa(r)
}

func (c Card) Label() string { return RankName(c.R) + c.S.Symbol() }

// Value is what the card counts as deadwood: aces 1, faces 10.
func (c Card) Value() int {
	if c.R >= 10 {
		return 10
	}
	return c.R
}

func itoa(n int) string {
	if n < 0 {
		return "-" + itoa(-n)
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}

// Deck returns the 52 cards in a fixed order.
func Deck() []Card {
	var d []Card
	for _, s := range Suits {
		for r := 1; r <= 13; r++ {
			d = append(d, Card{s, r})
		}
	}
	return d
}

// SortHand sorts by suit, then rank.
func SortHand(h []Card) {
	sort.Slice(h, func(i, j int) bool {
		if h[i].S != h[j].S {
			return suitIdx(h[i].S) < suitIdx(h[j].S)
		}
		return h[i].R < h[j].R
	})
}

/* ---------------- melds ---------------- */

// Meld is a set (three or four of a rank) or a run (three or more in a row
// of one suit). Runs are kept in rank order.
type Meld struct {
	Cards []Card
	Run   bool
}

func (m Meld) contains(c Card) bool {
	for _, x := range m.Cards {
		if x == c {
			return true
		}
	}
	return false
}

func Value(cs []Card) int {
	t := 0
	for _, c := range cs {
		t += c.Value()
	}
	return t
}

// allMelds lists every meld that can be made from the hand, including the
// three-card subsets of a four-of-a-kind and every stretch of a long run.
func allMelds(h []Card) []Meld {
	var out []Meld
	byRank := map[int][]Card{}
	have := map[Card]bool{}
	for _, c := range h {
		byRank[c.R] = append(byRank[c.R], c)
		have[c] = true
	}
	for r := 1; r <= 13; r++ {
		g := byRank[r]
		if len(g) >= 3 {
			if len(g) == 4 {
				out = append(out, Meld{Cards: append([]Card(nil), g...)})
				for skip := 0; skip < 4; skip++ {
					var m []Card
					for i, c := range g {
						if i != skip {
							m = append(m, c)
						}
					}
					out = append(out, Meld{Cards: m})
				}
			} else {
				out = append(out, Meld{Cards: append([]Card(nil), g...)})
			}
		}
	}
	for _, s := range Suits {
		for lo := 1; lo <= 11; lo++ {
			if !have[Card{s, lo}] {
				continue
			}
			var run []Card
			for r := lo; r <= 13 && have[Card{s, r}]; r++ {
				run = append(run, Card{s, r})
				if len(run) >= 3 {
					out = append(out, Meld{Cards: append([]Card(nil), run...), Run: true})
				}
			}
		}
	}
	return out
}

// Arrangement is a hand split into melds and deadwood.
type Arrangement struct {
	Melds    []Meld
	Dead     []Card
	Deadwood int
}

// Best finds the arrangement with the least deadwood. Ties go to the one
// with fewer, longer melds (and then the lowest leftover cards).
func Best(h []Card) Arrangement {
	ms := allMelds(h)
	best := Arrangement{Dead: append([]Card(nil), h...), Deadwood: Value(h)}
	var cur []Meld
	used := map[Card]bool{}
	var rec func(start int)
	rec = func(start int) {
		dw := 0
		for _, c := range h {
			if !used[c] {
				dw += c.Value()
			}
		}
		if dw < best.Deadwood || dw == best.Deadwood && len(cur) < len(best.Melds) && len(cur) > 0 {
			best.Deadwood = dw
			best.Melds = append([]Meld(nil), cur...)
		}
		for i := start; i < len(ms); i++ {
			ok := true
			for _, c := range ms[i].Cards {
				if used[c] {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			for _, c := range ms[i].Cards {
				used[c] = true
			}
			cur = append(cur, ms[i])
			rec(i + 1)
			cur = cur[:len(cur)-1]
			for _, c := range ms[i].Cards {
				used[c] = false
			}
		}
	}
	rec(0)
	best.Dead = nil
	in := map[Card]bool{}
	for _, m := range best.Melds {
		for _, c := range m.Cards {
			in[c] = true
		}
	}
	for _, c := range h {
		if !in[c] {
			best.Dead = append(best.Dead, c)
		}
	}
	SortHand(best.Dead)
	sortMelds(best.Melds)
	return best
}

func sortMelds(ms []Meld) {
	for _, m := range ms {
		sortMeld(m)
	}
	sort.SliceStable(ms, func(i, j int) bool {
		a, b := ms[i], ms[j]
		if a.Run != b.Run {
			return a.Run // runs first
		}
		if a.Cards[0].S != b.Cards[0].S && a.Run {
			return suitIdx(a.Cards[0].S) < suitIdx(b.Cards[0].S)
		}
		return a.Cards[0].R < b.Cards[0].R
	})
}

func sortMeld(m Meld) {
	sort.Slice(m.Cards, func(i, j int) bool {
		if m.Cards[i].R != m.Cards[j].R {
			return m.Cards[i].R < m.Cards[j].R
		}
		return suitIdx(m.Cards[i].S) < suitIdx(m.Cards[j].S)
	})
}

// Deadwood is the hand's deadwood count when arranged as well as possible.
func Deadwood(h []Card) int { return Best(h).Deadwood }

// KnockLimit: you may knock with this much deadwood or less.
const KnockLimit = 10

// BestDiscard finds the card to throw from an 11-card hand that leaves the
// least deadwood (never the card in forbid, the one just taken from the
// discard pile). It returns that card and the deadwood left.
func BestDiscard(h []Card, forbid *Card) (Card, int) {
	bc, bd := Card{}, 1<<30
	for i, c := range h {
		if forbid != nil && c == *forbid {
			continue
		}
		rest := make([]Card, 0, len(h)-1)
		rest = append(rest, h[:i]...)
		rest = append(rest, h[i+1:]...)
		d := Deadwood(rest)
		if d < bd || d == bd && c.Value() > bc.Value() {
			bc, bd = c, d
		}
	}
	return bc, bd
}

// Without returns h minus c.
func Without(h []Card, c Card) []Card {
	out := make([]Card, 0, len(h))
	for _, x := range h {
		if x != c {
			out = append(out, x)
		}
	}
	return out
}

// CanKnockWith says which discards let you knock (deadwood <= 10 after it).
func CanKnockWith(h []Card, forbid *Card) (knock, gin []Card) {
	for _, c := range h {
		if forbid != nil && c == *forbid {
			continue
		}
		d := Deadwood(Without(h, c))
		if d <= KnockLimit {
			knock = append(knock, c)
		}
		if d == 0 {
			gin = append(gin, c)
		}
	}
	return
}

/* ---------------- laying off ---------------- */

// canLayOff says whether c extends meld m (a set's fourth card, or either end of a run).
func canLayOff(m Meld, c Card) bool {
	if m.contains(c) {
		return false
	}
	if !m.Run {
		return len(m.Cards) < 4 && c.R == m.Cards[0].R
	}
	lo, hi := m.Cards[0], m.Cards[len(m.Cards)-1]
	return c.S == lo.S && (c.R == lo.R-1 || c.R == hi.R+1)
}

// Layoff is a card the defender adds to one of the knocker's melds.
type Layoff struct {
	Card Card
	Meld int
}

// Defend works out the defender's best answer to a knock: their own melds,
// the cards they lay off on the knocker's melds, and what's left.
func Defend(h []Card, knocker []Meld) (Arrangement, []Layoff) {
	ms := allMelds(h)
	bestA := Arrangement{Deadwood: 1 << 30}
	var bestL []Layoff
	var cur []Meld
	used := map[Card]bool{}
	try := func() {
		var left []Card
		for _, c := range h {
			if !used[c] {
				left = append(left, c)
			}
		}
		lays, rest := layOff(left, knocker)
		dw := Value(rest)
		if dw < bestA.Deadwood || dw == bestA.Deadwood && len(lays) > len(bestL) {
			bestA = Arrangement{Melds: append([]Meld(nil), cur...), Dead: rest, Deadwood: dw}
			bestL = lays
		}
	}
	var rec func(start int)
	rec = func(start int) {
		try()
		for i := start; i < len(ms); i++ {
			ok := true
			for _, c := range ms[i].Cards {
				if used[c] {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			for _, c := range ms[i].Cards {
				used[c] = true
			}
			cur = append(cur, ms[i])
			rec(i + 1)
			cur = cur[:len(cur)-1]
			for _, c := range ms[i].Cards {
				used[c] = false
			}
		}
	}
	rec(0)
	for i := range bestA.Melds {
		bestA.Melds[i] = Meld{Cards: append([]Card(nil), bestA.Melds[i].Cards...), Run: bestA.Melds[i].Run}
	}
	sortMelds(bestA.Melds)
	SortHand(bestA.Dead)
	return bestA, bestL
}

// layOff lays off as many of left as possible (runs can grow card by card),
// trying the highest-value cards first. It returns the layoffs and what's left.
func layOff(left []Card, knocker []Meld) ([]Layoff, []Card) {
	ms := make([]Meld, len(knocker))
	for i, m := range knocker {
		ms[i] = Meld{Cards: append([]Card(nil), m.Cards...), Run: m.Run}
	}
	rest := append([]Card(nil), left...)
	sort.Slice(rest, func(i, j int) bool { return rest[i].Value() > rest[j].Value() })
	var lays []Layoff
	for changed := true; changed; {
		changed = false
		for k := 0; k < len(rest); k++ {
			c := rest[k]
			for i := range ms {
				if canLayOff(ms[i], c) {
					ms[i].Cards = append(ms[i].Cards, c)
					sortMeld(ms[i])
					lays = append(lays, Layoff{Card: c, Meld: i})
					rest = append(rest[:k], rest[k+1:]...)
					k--
					changed = true
					break
				}
			}
		}
	}
	return lays, rest
}

/* ---------------- scoring ---------------- */

const (
	GinBonus      = 25
	BigGinBonus   = 31
	UndercutBonus = 25
	GameBonus     = 100
	BoxBonus      = 25 // per hand won, added at the end of the game
)

// Outcome is what a hand was worth.
type Outcome int

const (
	Knocked Outcome = iota
	Undercut
	Gin
	BigGin
	DeadHand
)

// Result of a hand: who scored how much and why.
type Result struct {
	Kind              Outcome
	Winner            int // seat that scores (-1 for a dead hand)
	Points            int
	KnockDW, DefendDW int
}

// ScoreKnock scores a knock by seat k with knocker deadwood kd against
// defender deadwood dd (after laying off).
func ScoreKnock(k, kd, dd int, kind Outcome) Result {
	d := 1 - k
	switch kind {
	case Gin:
		return Result{Kind: Gin, Winner: k, Points: dd + GinBonus, KnockDW: 0, DefendDW: dd}
	case BigGin:
		return Result{Kind: BigGin, Winner: k, Points: dd + BigGinBonus, KnockDW: 0, DefendDW: dd}
	}
	if dd <= kd {
		return Result{Kind: Undercut, Winner: d, Points: kd - dd + UndercutBonus, KnockDW: kd, DefendDW: dd}
	}
	return Result{Kind: Knocked, Winner: k, Points: dd - kd, KnockDW: kd, DefendDW: dd}
}

// Final works out the end-of-game totals: the game bonus for the winner, a
// box bonus for every hand each player won, and everything doubled for a
// shutout (the loser never scored).
type Final struct {
	Winner      int
	Totals      [2]int
	Shutout     bool
	Margin      int
	GameBonuses [2]int
	BoxBonuses  [2]int
}

func FinalScore(score [2]int, boxes [2]int) Final {
	w := 0
	if score[1] > score[0] {
		w = 1
	}
	l := 1 - w
	f := Final{Winner: w}
	f.GameBonuses[w] = GameBonus
	for i := 0; i < 2; i++ {
		f.BoxBonuses[i] = boxes[i] * BoxBonus
		f.Totals[i] = score[i] + f.BoxBonuses[i] + f.GameBonuses[i]
	}
	if score[l] == 0 {
		f.Shutout = true
		f.Totals[w] *= 2
	}
	f.Margin = f.Totals[w] - f.Totals[l]
	return f
}

// GameOver says whether someone has reached the target.
func GameOver(score [2]int, target int) bool { return score[0] >= target || score[1] >= target }

// StockFloor: when this many cards are left in the stock and nobody has
// knocked, the hand is dead and dealt again.
const StockFloor = 2
