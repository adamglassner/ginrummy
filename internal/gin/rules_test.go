package gin

import (
	"math/rand/v2"
	"testing"
)

func cards(s string) []Card {
	var out []Card
	rank := map[byte]int{'A': 1, 'T': 10, 'J': 11, 'Q': 12, 'K': 13}
	for i := 0; i+1 < len(s); i += 3 {
		r, ok := rank[s[i]]
		if !ok {
			r = int(s[i] - '0')
		}
		out = append(out, Card{Suit(s[i+1]), r})
	}
	return out
}

func TestMeldsAndDeadwood(t *testing.T) {
	cases := []struct {
		hand string
		dw   int
	}{
		{"AS 2S 3S 7H 7D 7C KH QD 4C 9S", 10 + 10 + 4 + 9},
		{"AS 2S 3S 4S 5H 5D 5C 9H 9D 9C", 0},
		// the 7♥ fits the set or the run, not both
		{"5H 6H 7H 7D 7C 8S 9S TS KD KC", 5 + 6 + 20},
		// a four-of-a-kind gives up a card to a run
		{"4H 4D 4C 4S 5S 6S 9H 9D 2C 3D", 18 + 5},
		{"KH KD KC KS QH QD QC JH JD JC", 0},
		{"AH 2D 3C 4S 6H 8D TC QS KH 9C", 1 + 2 + 3 + 4 + 6 + 8 + 10 + 10 + 10 + 9},
	}
	for _, c := range cases {
		if got := Deadwood(cards(c.hand)); got != c.dw {
			t.Errorf("%s: deadwood %d, want %d (%+v)", c.hand, got, c.dw, Best(cards(c.hand)))
		}
	}
	// aces are low only: Q K A is no run
	if d := Deadwood(cards("QS KS AS")); d != 21 {
		t.Errorf("Q-K-A should not be a run, deadwood %d", d)
	}
}

func TestLayoff(t *testing.T) {
	knocker := Best(cards("4S 5S 6S 8H 8D 8C 2C 2D 2H 3H"))
	if knocker.Deadwood != 3 {
		t.Fatalf("knocker deadwood %d", knocker.Deadwood)
	}
	// 3♠ and 7♠ extend the run, 8♠ the set; then 2♠ fits the set of twos... and then the run (7S already)
	def, lays := Defend(cards("3S 7S 8S KD KC KH 9D JC QC 5H"), knocker.Melds)
	if def.Deadwood != 9+10+10+5 && def.Deadwood != 9+5+10+10 {
		t.Errorf("defender deadwood %d, lays %+v, %+v", def.Deadwood, lays, def)
	}
	if len(lays) != 3 {
		t.Errorf("want 3 layoffs, got %+v", lays)
	}
	// a run can grow card by card: 7♠ then 8♠
	k2 := []Meld{{Cards: cards("4S 5S 6S"), Run: true}}
	_, l2 := Defend(cards("7S 8S"), k2)
	if len(l2) != 2 {
		t.Errorf("chain layoff: %+v", l2)
	}
}

func TestScoring(t *testing.T) {
	if r := ScoreKnock(0, 5, 20, Knocked); r.Winner != 0 || r.Points != 15 {
		t.Errorf("knock %+v", r)
	}
	if r := ScoreKnock(0, 5, 5, Knocked); r.Kind != Undercut || r.Winner != 1 || r.Points != 25 {
		t.Errorf("tie is an undercut %+v", r)
	}
	if r := ScoreKnock(1, 7, 3, Knocked); r.Kind != Undercut || r.Winner != 0 || r.Points != 29 {
		t.Errorf("undercut %+v", r)
	}
	if r := ScoreKnock(0, 0, 30, Gin); r.Points != 55 {
		t.Errorf("gin %+v", r)
	}
	f := FinalScore([2]int{104, 0}, [2]int{4, 0})
	if !f.Shutout || f.Totals[0] != (104+100+100)*2 {
		t.Errorf("shutout %+v", f)
	}
	f = FinalScore([2]int{60, 102}, [2]int{3, 4})
	if f.Winner != 1 || f.Totals[1] != 102+100+100 || f.Totals[0] != 60+75 || f.Margin != 302-135 {
		t.Errorf("final %+v", f)
	}
}

// Computer players against each other: every hand ends properly and the
// numbers always add up.
func TestComputerGames(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	gins, knocks, unders, deads := 0, 0, 0, 0
	for game := 0; game < 120; game++ {
		deck := Deck()
		r.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
		hands := [2][]Card{append([]Card(nil), deck[:10]...), append([]Card(nil), deck[10:20]...)}
		stock := append([]Card(nil), deck[21:]...)
		disc := []Card{deck[20]}
		mem := [2]Memory{}
		p := game % 2
		turns := [2]int{}
		done := false
		for !done {
			if len(stock) <= StockFloor {
				deads++
				break
			}
			top := disc[len(disc)-1]
			var forbid *Card
			if WantsDiscard(hands[p], top) {
				hands[p] = append(hands[p], top)
				disc = disc[:len(disc)-1]
				mem[1-p].OppTook = append(mem[1-p].OppTook, top)
				forbid = &top
			} else {
				hands[p] = append(hands[p], stock[0])
				stock = stock[1:]
			}
			if len(hands[p]) != 11 {
				t.Fatalf("hand size %d", len(hands[p]))
			}
			turns[p]++
			c, knock := ChooseTurn(hands[p], forbid, &mem[p], len(stock), turns[p], Style(game%3))
			if forbid != nil && c == *forbid {
				t.Fatalf("threw the card it just took")
			}
			hands[p] = Without(hands[p], c)
			if knock {
				k := Best(hands[p])
				if k.Deadwood > KnockLimit {
					t.Fatalf("knocked with %d", k.Deadwood)
				}
				kind := Knocked
				if k.Deadwood == 0 {
					kind = Gin
					gins++
				} else {
					knocks++
				}
				var dd int
				if kind == Gin {
					dd = Deadwood(hands[1-p])
				} else {
					a, lays := Defend(hands[1-p], k.Melds)
					dd = a.Deadwood
					n := len(lays) + len(a.Dead)
					for _, m := range a.Melds {
						n += len(m.Cards)
					}
					if n != 10 {
						t.Fatalf("defence lost cards: %d", n)
					}
				}
				res := ScoreKnock(p, k.Deadwood, dd, kind)
				if res.Kind == Undercut {
					unders++
				}
				if res.Points < 0 {
					t.Fatalf("negative points %+v", res)
				}
				done = true
			} else {
				disc = append(disc, c)
				mem[1-p].OppDiscarded = append(mem[1-p].OppDiscarded, c)
			}
			p = 1 - p
		}
	}
	t.Logf("gins %d knocks %d undercuts %d dead hands %d", gins, knocks, unders, deads)
	if gins+knocks < 100 {
		t.Errorf("too few hands finished: %d", gins+knocks)
	}
}
