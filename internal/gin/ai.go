package gin

// The computer players. They keep track of what their opponent picked up
// and threw away, take a discard only when it really helps, throw the card
// that leaves the least deadwood (keeping pairs and near-runs, and not
// feeding the other player), and knock early, or hold out for gin.

// Memory is what a computer player remembers about this hand.
type Memory struct {
	OppTook      []Card // cards the opponent took from the discard pile
	OppDiscarded []Card // cards the opponent threw away
}

// Style is how a computer player likes to play.
type Style int

const (
	Steady  Style = iota // knocks when it's sensible
	Eager                // knocks as soon as it can
	GinHunt              // holds out for gin unless time runs short
)

func inMeld(a Arrangement, c Card) bool {
	for _, m := range a.Melds {
		if m.contains(c) {
			return true
		}
	}
	return false
}

// WantsDiscard says whether to take top from the discard pile (instead of
// drawing blind from the stock) with the 10-card hand h.
func WantsDiscard(h []Card, top Card) bool {
	now := Deadwood(h)
	with := append(append([]Card(nil), h...), top)
	_, after := BestDiscard(with, &top)
	if after >= now {
		return false
	}
	if inMeld(Best(with), top) {
		return true
	}
	// a low card that cuts deadwood a lot (handy when close to knocking)
	return after <= KnockLimit && now-after >= 4
}

// potential counts the deadwood cards that are one card from a meld: pairs
// and suited cards one or two apart.
func potential(dead []Card) int {
	n := 0
	for i, a := range dead {
		for j, b := range dead {
			if i >= j {
				continue
			}
			if a.R == b.R {
				n += 2
			} else if a.S == b.S && abs(a.R-b.R) <= 2 {
				n += 1 + boolInt(abs(a.R-b.R) == 1)
			}
		}
	}
	return n
}

// danger is how likely c is to help the opponent, from what they took and threw.
func danger(c Card, m *Memory) int {
	if m == nil {
		return 0
	}
	d := 0
	for _, t := range m.OppTook {
		if t.R == c.R {
			d += 3
		}
		if t.S == c.S && abs(t.R-c.R) <= 2 {
			d += 2
		}
	}
	for _, t := range m.OppDiscarded {
		if t.R == c.R {
			d -= 2
		}
		if t.S == c.S && abs(t.R-c.R) == 1 {
			d--
		}
	}
	return d
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ChooseDiscard picks the card to throw from an 11-card hand.
func ChooseDiscard(h []Card, forbid *Card, m *Memory) Card {
	best, bestCost := Card{}, 1<<30
	for _, c := range h {
		if forbid != nil && c == *forbid {
			continue
		}
		a := Best(Without(h, c))
		cost := a.Deadwood*4 - potential(a.Dead)*3 + danger(c, m)*5
		if cost < bestCost || cost == bestCost && c.Value() > best.Value() {
			best, bestCost = c, cost
		}
	}
	return best
}

// ChooseTurn is the end of a computer player's turn: what to throw and
// whether to knock with it. stock is the cards left to draw; turns is how
// many turns this player has had this hand.
func ChooseTurn(h []Card, forbid *Card, m *Memory, stock, turns int, st Style) (Card, bool) {
	knock, gin := CanKnockWith(h, forbid)
	if len(gin) > 0 {
		return pickLowest(h, gin, forbid), true
	}
	if len(knock) > 0 {
		c := pickLowest(h, knock, forbid)
		dw := Deadwood(Without(h, c))
		limit := 10
		switch {
		case st == Eager:
			limit = 10
		case st == GinHunt && stock > 10:
			limit = 1
		case st == GinHunt:
			limit = 5
		case turns <= 5:
			limit = 10
		case stock > 12:
			limit = 7
		default:
			limit = 4
		}
		if stock <= 6 {
			limit = 10 // knock before the hand runs out
		}
		if dw <= limit {
			return c, true
		}
	}
	return ChooseDiscard(h, forbid, m), false
}

// pickLowest is the allowed discard (from among cs) that leaves the least deadwood.
func pickLowest(h, cs []Card, forbid *Card) Card {
	best, bd := cs[0], 1<<30
	for _, c := range cs {
		if forbid != nil && c == *forbid {
			continue
		}
		d := Deadwood(Without(h, c))
		if d < bd || d == bd && c.Value() > best.Value() {
			best, bd = c, d
		}
	}
	return best
}

// IsBigGin says whether all 11 cards meld (gin without a discard).
func IsBigGin(h []Card) bool { return len(h) == 11 && Deadwood(h) == 0 }
