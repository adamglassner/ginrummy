//go:build headless

// A test build with no window or sound: plays whole games with the computer
// at every seat, including online games between a host and guests over
// real TCP on this computer. Run: go run -tags headless .
package main

import (
	"fmt"
	"os"
	"time"

	gn "github.com/adamglassner/ginrummy/internal/gin"
)

type uiState struct{}

func check(ok bool, f string, a ...any) {
	if !ok {
		fmt.Printf("FAIL: "+f+"\n", a...)
		os.Exit(1)
	}
}

func waitFor(what string, d time.Duration, f func() bool) {
	end := time.Now().Add(d)
	for time.Now().Before(end) {
		if f() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	check(false, "timed out waiting for %s", what)
}

func locked[T any](g *game, f func() T) T {
	g.mu.Lock()
	defer g.mu.Unlock()
	return f()
}

// fitsFace checks a regular's name was made for their face.
func fitsFace(face, name string) bool {
	f, ok := parseFace(face)
	if !ok {
		return false
	}
	mn := mannerNames[f.mannerKey()]
	for _, t := range mn.titles {
		for _, n := range mn.names {
			if t+" "+n == name {
				return true
			}
		}
	}
	return false
}

func main() {
	dir := os.TempDir() + "/ginrummy-headless-test"
	os.Setenv("GINRUMMY_DATA", dir)
	os.RemoveAll(dir)
	autopilot, testFast = true, true

	// 0. faces: names always fit, ladies are ladies
	for i := 0; i < 2000; i++ {
		f := Face{Female: i%2 == 0, Seed: uint64(i)*7919 + 3}
		n, full, ok := f.nameFor(map[string]bool{})
		check(ok && fitsFace(f.String(), n), "face %v got %q", f, full)
		l := f.look()
		check(l.Lady == f.Female, "a lady's face must look like a lady")
		if !f.Female {
			check(l.Lip.A == 0 && l.HairStyle != "bob", "a gentleman's face got a lady's look")
		}
		g2, ok := parseFace(f.String())
		check(ok && g2 == f, "face round trip")
	}

	// sane checks every card is in exactly one place and hands are the right size.
	sane := func(g *game, where string) {
		n := len(g.stock) + len(g.disc)
		for _, p := range g.players {
			n += len(p.hand)
			if len(p.hand) > 0 {
				check(len(p.hand) == 10 || len(p.hand) == 11 || g.show == nil && len(p.hand) < 10, "%s: %s holds %d cards", where, p.name, len(p.hand))
			}
		}
		shown := 0
		for _, c := range g.deck {
			if c.zn.K == zShow {
				shown++
			}
		}
		check(n+shown == 52 || n+shown < 52 && g.show == nil, "%s: %d cards accounted for", where, n+shown)
	}

	// 1. single player: a random passenger whose name fits their face, several games
	names := map[string]bool{}
	for game := 0; game < 4; game++ {
		g := newGameState()
		g.mu.Lock()
		g.goPlay()
		p := g.players[1]
		check(p.kind == KindAI && fitsFace(p.face, p.name), "%q doesn't fit %s", p.name, p.face)
		names[p.name] = true
		fmt.Printf("solo game %d vs %s (%s)\n", game+1, p.full, p.face)
		g.mu.Unlock()
		waitFor("solo game", 5*time.Minute, func() bool {
			return locked(g, func() bool {
				if !g.showResult {
					sane(g, "solo")
				}
				return g.showResult
			})
		})
		g.mu.Lock()
		check(gn.GameOver(g.score, g.target), "game ended early: %v", g.score)
		f := g.final
		check(f.Totals[f.Winner] >= f.Totals[1-f.Winner], "final totals %+v", f)
		check(g.save.Stats.Played == game+1 && g.save.Stats.Hands > 0, "record %+v after %d hands", g.save.Stats, g.handsTot)
		fmt.Printf("  %v after %d hands (boxes %v): %q / %s\n", g.score, g.handNo, g.boxes, g.result.title, g.result.sub)
		g.mu.Unlock()
	}
	check(len(names) > 1, "every game had the same passenger")

	// 2. online: a host and a guest
	host := newGameState()
	host.setup.name, host.setup.face = "Hosty", pickerFaces[1]
	host.mu.Lock()
	host.hostGame()
	check(host.server != nil, "host should be listening")
	host.mu.Unlock()
	c := newGameState()
	c.save.Stats = defaultSave().Stats // both share one save folder here
	c.setup.name, c.setup.face = "Ann", pickerFaces[2]
	c.mu.Lock()
	c.joinGame("127.0.0.1")
	c.mu.Unlock()
	waitFor("Ann to join", 10*time.Second, func() bool {
		return locked(host, func() bool { return len(host.server.peers) == 1 }) && locked(c, func() bool { return c.mode == ModeClient })
	})
	host.mu.Lock()
	check(host.players[1].name == "Ann" && host.players[1].kind == KindRemote, "Ann should sit across")
	host.mu.Unlock()
	time.Sleep(100 * time.Millisecond)

	for round := 1; round <= 2; round++ {
		host.mu.Lock()
		if round == 1 {
			host.startHosted()
		} else {
			host.newDeal()
		}
		host.mu.Unlock()
		done, asks := false, 0
		var lastAsk *turnT
		for !done {
			time.Sleep(5 * time.Millisecond)
			c.mu.Lock()
			me := c.viewer
			for _, card := range c.deck {
				if card.zn.K == zHand && card.zn.P == me {
					check(card.id != "", "guest doesn't know a card in their own hand")
				}
				if card.zn.K == zStock {
					check(card.id == "" || card.fa1 == 180, "guest can see into the stock")
				}
			}
			for i, p := range c.players {
				if i != me && p != nil {
					for _, card := range p.hand {
						check(card.id == "" || card.fa1 == 180, "guest can see the host's card")
					}
				}
			}
			if c.turn != nil && c.turn != lastAsk {
				asks++
				lastAsk = c.turn
			}
			c.mu.Unlock()
			done = locked(host, func() bool { return host.showResult })
		}
		check(asks > 0, "the guest was never asked to play")
		waitFor("guest's result", 10*time.Second, func() bool { return locked(c, func() bool { return c.showResult }) })
		hs := host.score
		c.mu.Lock()
		check(c.score == hs, "guest scores %v vs host %v", c.score, hs)
		check(len(c.result.rows) == 2, "guest result rows")
		check(c.save.Stats.Played == round, "guest record: %d games", c.save.Stats.Played)
		fmt.Printf("online round %d: Ann sees %q (%s) %v, %d asks\n", round, c.result.title, c.result.sub, c.score, asks)
		c.mu.Unlock()
	}

	// 3. the guest leaves mid-game: a computer player takes over and the game finishes
	host.mu.Lock()
	host.newDeal()
	host.mu.Unlock()
	time.Sleep(400 * time.Millisecond)
	c.mu.Lock()
	c.leaveOnline()
	c.mu.Unlock()
	waitFor("game after the guest left", 5*time.Minute, func() bool { return locked(host, func() bool { return host.showResult }) })
	host.mu.Lock()
	check(host.players[1].kind == KindAI, "Ann's seat should be a computer now")
	check(len(host.server.peers) == 0, "no guests left")
	fmt.Printf("after Ann left: %s finished it, %v\n", host.players[1].full, host.score)
	// a new guest can join again and the host can close the table
	host.mu.Unlock()
	c2 := newGameState()
	c2.save.Stats = defaultSave().Stats
	c2.setup.name, c2.setup.face = "Bo", pickerFaces[3]
	c2.mu.Lock()
	c2.joinGame("127.0.0.1")
	c2.mu.Unlock()
	waitFor("Bo to join", 10*time.Second, func() bool { return locked(c2, func() bool { return c2.mode == ModeClient }) })
	host.mu.Lock()
	host.leaveOnline()
	host.mu.Unlock()
	waitFor("Bo to hear the host closed", 10*time.Second, func() bool {
		return locked(c2, func() bool { return c2.mode == ModeSolo && c2.message != "" })
	})

	// 4. same computer: two people, answered through the real pass screen and turn boxes
	autopilot = false
	hs := newGameState()
	hs.mu.Lock()
	hs.save.Settings.Target = 100
	hs.setup.hotNames = [2]string{"Al", "Bea"}
	hs.startHotseat()
	check(hs.players[0].name == "Al" && hs.players[1].name == "Bea" && hs.players[1].kind == KindLocal, "hot-seat seating")
	hs.mu.Unlock()
	passScreens, turns, knocks := 0, map[int]int{}, 0
	var answered *turnT
	for !locked(hs, func() bool { return hs.showResult }) {
		time.Sleep(2 * time.Millisecond)
		hs.mu.Lock()
		switch t := hs.turn; {
		case hs.passCh != nil:
			for _, c := range hs.deck {
				check(c.zn.K != zHand || c.fa1 == 180, "a hand is showing behind the pass screen")
			}
			passScreens++
			hs.passCh <- struct{}{}
			hs.passCh = nil
		case t != nil && t != answered:
			answered = t
			if t.kind != "next" {
				check(!hs.hidden && hs.players[hs.viewer].kind == KindLocal, "a turn box with the cards hidden")
				turns[hs.viewer]++
			}
			switch t.kind {
			case "upcard":
				t.ch <- turnRes{pile: []string{"disc", "pass"}[hs.rng.IntN(2)]}
			case "draw":
				t.ch <- turnRes{pile: []string{"disc", "stock", "stock"}[hs.rng.IntN(3)]}
			case "discard":
				for _, c := range t.ok {
					check(c.zn.P == hs.viewer, "discarding from someone else's hand")
				}
				switch {
				case t.big:
					t.ch <- turnRes{big: true, knock: true}
				case len(t.gin) > 0:
					t.ch <- turnRes{card: t.gin[0], knock: true}
					knocks++
				case len(t.knock) > 0:
					t.ch <- turnRes{card: t.knock[0], knock: true}
					knocks++
				default:
					// throw the card a sensible player would
					var fc *gn.Card
					for _, c := range hs.players[hs.viewer].hand {
						if !inList(c, t.ok) {
							fc = &c.c
						}
					}
					best, _ := gn.BestDiscard(cardsOf(hs.players[hs.viewer].hand), fc)
					t.ch <- turnRes{card: hs.byCard[best]}
				}
			default:
				t.ch <- turnRes{}
			}
		}
		if !hs.showResult {
			sane(hs, "same computer")
		}
		hs.mu.Unlock()
	}
	check(passScreens > 3 && turns[0] > 0 && turns[1] > 0 && knocks > 0, "hot-seat turns %v knocks %d screens %d", turns, knocks, passScreens)
	fmt.Printf("same computer: %d pass screens, turns %v, %d knocks, %q %s\n", passScreens, turns, knocks, hs.result.title, hs.result.sub)
	fmt.Println("PASS")
}
