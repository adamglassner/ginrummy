package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"strings"
	"sync"
	"time"

	gn "github.com/adamglassner/ginrummy/internal/gin"
)

/* ---------------- who's at the table ---------------- */

// Kind says who plays a seat.
type Kind int

const (
	KindAI     Kind = iota // a computer player
	KindLocal              // someone at this computer
	KindRemote             // someone playing over the network
)

// Mode is how this game is being played.
type Mode int

const (
	ModeSolo    Mode = iota // you and a fellow passenger
	ModeHotseat             // two people sharing this computer
	ModeHost                // hosting an online game
	ModeClient              // playing in someone else's online game
)

// Seats as drawn: the viewer always sits at the bottom (S), the other player across (N).
const (
	seatS = 'S'
	seatN = 'N'
)

type player struct {
	i          int
	kind       Kind
	face       string // a Face as text, "" for none
	name, full string // both "Miss Betty"
	hand       []*cardV
	say        string
	sayAt      time.Time
	sayEnd     time.Time
	turnAt     time.Time
	isTurn     bool
	peer       *peer
	style      gn.Style
}

/* ---------------- where a card is ---------------- */

type zoneK int

const (
	zNone  zoneK = iota
	zStock       // the stock, N from the bottom
	zDisc        // the discard pile, N from the bottom (M = 1: face down, the knock card)
	zHand        // in player P's hand, position N, in meld group M
	zShow        // laid down at the end of a hand: P's row, group M (9 = deadwood), position N
)

const deadGroup = 9

type zone struct {
	K zoneK   `json:"k"`
	P int     `json:"p,omitempty"`
	N int     `json:"n,omitempty"`
	M int     `json:"m,omitempty"`
	J float64 `json:"j,omitempty"` // a little random rotation
}

type cardV struct {
	c   gn.Card
	id  string // "H11"; "" when this computer doesn't know the face
	idx int
	tok int // what guests call this card (shuffled every hand)
	zn  zone
	// position animation (table coordinates, card centre)
	fx, fy, fr, fs, tx, ty, tr, ts float64
	t0                             time.Time
	dur                            float64
	placed                         bool
	// flip animation: 0 = face up, 180 = face down
	fa0, fa1 float64
	ft0      time.Time
	fdur     float64
	z        int
	speed    int // 0 normal, 1 fast, 2 instant
	pick     bool
	dim      bool
	win      bool // laid off (or the card that made gin)
	fresh    bool // just drawn
	lift     float64
	x, y     float64
}

func (c *cardV) pos(now time.Time) (x, y, r, s float64) {
	if c.dur <= 0 {
		return c.tx, c.ty, c.tr, c.ts
	}
	t := easeCard(now.Sub(c.t0).Seconds() / c.dur)
	l := func(a, b float64) float64 { return a + (b-a)*t }
	return l(c.fx, c.tx), l(c.fy, c.ty), l(c.fr, c.tr), l(c.fs, c.ts)
}

func (c *cardV) angle(now time.Time) float64 {
	if c.fdur <= 0 {
		return c.fa1
	}
	t := easeInOut(now.Sub(c.ft0).Seconds() / c.fdur)
	return c.fa0 + (c.fa1-c.fa0)*t
}

/* ---------------- game state ---------------- */

type askBtn struct {
	Label string `json:"label"`
	Cls   string `json:"cls,omitempty"` // "", "ghost"
}

// turnT is a decision someone at this computer is making: take the upcard,
// draw, or discard (and maybe knock).
type turnT struct {
	kind     string // "upcard", "draw", "discard", "next"
	top      *cardV // the discard pile's top card
	stockTop *cardV
	ok       []*cardV // cards that may be discarded
	knock    []*cardV // discards that let you knock
	gin      []*cardV // discards that make gin
	big      bool     // all eleven cards meld
	knocking bool     // picking the card to knock with
	label    string   // the "next" button
	at       time.Time
	ch       chan turnRes
}

type turnRes struct {
	pile  string // "stock", "disc", "pass" (draws and upcards)
	card  *cardV
	knock bool
	big   bool
}

type bannerT struct {
	title, small string
	blue         bool
	at, until    time.Time
}

type particle struct {
	k                        int // 0 star, 1 spark, 2 confetti
	x, y, vx, vy, life, d, s float64
	rot, vr                  float64
	col                      int
}

type resRow struct {
	Name   string `json:"n"`
	Face   string `json:"f,omitempty"`
	Score  int    `json:"s"`
	Detail string `json:"d,omitempty"`
	Place  int    `json:"p"`
	Me     bool   `json:"m,omitempty"`
}

type resultT struct {
	kick, title, sub string
	rows             []resRow
	rec              string
	won              bool
}

// showT is the end of a hand, laid out on the table.
type showT struct {
	Knocker int    `json:"k"`
	Kind    string `json:"kind"` // "knock", "gin", "big"
	KDW     int    `json:"kdw"`
	DDW     int    `json:"ddw"`
	Done    bool   `json:"done,omitempty"` // the defender has laid down
	Winner  int    `json:"w"`
	Points  int    `json:"pts"`
	Under   bool   `json:"under,omitempty"`
	Lays    int    `json:"lays,omitempty"`
}

// tally is one seat's haul this game, for guests' records.
type tally struct{ Hands, HandsWon, Gins, Knocks, Undercuts int }

type game struct {
	mu sync.Mutex

	save    saveData
	mode    Mode
	viewer  int  // whose seat the table is drawn from
	hidden  bool // shared computer: hands face down between turns
	players [2]*player
	deck    []*cardV
	byCard  map[gn.Card]*cardV
	byTok   [52]*cardV

	stock    []*cardV
	disc     []*cardV
	dealer   int
	handNo   int // hands dealt this game (the first is 0)
	handsTot int // hands dealt since the program started (for guests' snapshots)
	score    [2]int
	boxes    [2]int
	target   int
	mem      [2]gn.Memory
	turns    [2]int
	show     *showT
	final    gn.Final

	gameID   int
	cancel   chan struct{}
	finished bool
	started  bool
	turn     *turnT
	remote   map[int]*remoteWait
	prompt   string
	promptAt time.Time
	ban      bannerT
	pace     float64
	parts    []particle

	tallies [2]tally

	// layout, in logical pixels (table space)
	Wd, Ht, cw, ch float64

	// overlays
	showTitle, showRules, showCredits, showResult bool
	showTogether, showHotseat, showOnline         bool
	showLobby                                     bool
	passTo                                        int
	passCh                                        chan struct{}
	message                                       string
	result                                        resultT
	resultAt                                      time.Time
	ui                                            uiState

	setup  setupState
	focus  string // text field being typed in
	server *server
	client *client
	lobby  []lobbySeat
	cliEnd chan struct{} // closed to drop a guest's pending question

	nextFolk []folk // who sits down for the next single-player game

	rng *rand.Rand
}

var errCancel = fmt.Errorf("cancelled")

func newGameState() *game {
	seed := uint64(time.Now().UnixNano())
	g := &game{rng: rand.New(rand.NewPCG(seed, seed>>7^0x5eed)), pace: 1, showTitle: true, remote: map[int]*remoteWait{}}
	g.save = loadSave()
	g.setup = newSetup(&g.save)
	g.target = g.save.Settings.Target
	g.resetDeck(false)
	g.soloTable()
	g.nextFolk = []folk{{face: g.players[1].face, name: g.players[1].name, full: g.players[1].full}}
	g.Wd, g.Ht = 1400, 820
	g.measure()
	return g
}

// resetDeck makes the 52 cards. A guest's cards start out unknown.
func (g *game) resetDeck(guest bool) {
	g.deck = nil
	g.byCard = map[gn.Card]*cardV{}
	for i, c := range gn.Deck() {
		v := &cardV{c: c, id: c.ID(), idx: i, tok: i, ts: 1, fs: 1, fa1: 180}
		if guest {
			v.id = ""
		}
		g.deck = append(g.deck, v)
		g.byCard[c] = v
		g.byTok[i] = v
	}
	g.stock, g.disc = nil, nil
}

func (g *game) rand(n int) int      { return g.rng.IntN(n) }
func pickS[T any](g *game, a []T) T { return a[g.rand(len(a))] }

/* pacing */
var paceOf = map[string]float64{"relaxed": 1.35, "normal": 1, "brisk": 0.6}

func (g *game) T(k string) float64 {
	v := map[string]float64{"deal": 55, "move": 420, "play": 420, "think": 800, "quip": 1900, "banner": 1700, "show": 650}[k]
	v *= g.pace
	if testFast {
		v *= 0.04
	}
	return v
}

// sleep releases the table while waiting; a new game cancels the old one.
func (g *game) sleep(ms float64) {
	id, cancel := g.gameID, g.cancel
	g.mu.Unlock()
	select {
	case <-time.After(time.Duration(ms * float64(time.Millisecond))):
	case <-cancel:
	}
	g.mu.Lock()
	if id != g.gameID {
		panic(errCancel)
	}
}

// waitOn releases the table until ch delivers (or the game is cancelled).
func waitOn[T any](g *game, ch chan T) T {
	id, cancel := g.gameID, g.cancel
	g.mu.Unlock()
	var v T
	select {
	case v = <-ch:
	case <-cancel:
	}
	g.mu.Lock()
	if id != g.gameID {
		panic(errCancel)
	}
	return v
}

/* ---------------- cards ---------------- */

func cardsOf(h []*cardV) []gn.Card {
	out := make([]gn.Card, len(h))
	for i, c := range h {
		out[i] = c.c
	}
	return out
}

// cardFromID reads "H11" (aces are 14 in the art) back into a card.
func cardFromID(id string) (gn.Card, bool) {
	if len(id) < 2 {
		return gn.Card{}, false
	}
	r := 0
	for _, ch := range id[1:] {
		if ch < '0' || ch > '9' {
			return gn.Card{}, false
		}
		r = r*10 + int(ch-'0')
	}
	if r == 14 {
		r = 1
	}
	if r < 1 || r > 13 || !strings.ContainsRune("SHDC", rune(id[0])) {
		return gn.Card{}, false
	}
	return gn.Card{S: gn.Suit(id[0]), R: r}, true
}

func without(h []*cardV, c *cardV) []*cardV {
	out := make([]*cardV, 0, len(h))
	for _, x := range h {
		if x != c {
			out = append(out, x)
		}
	}
	return out
}

func inList(c *cardV, l []*cardV) bool {
	for _, x := range l {
		if x == c {
			return true
		}
	}
	return false
}

func (g *game) topDisc() *cardV {
	if len(g.disc) == 0 {
		return nil
	}
	return g.disc[len(g.disc)-1]
}

func (g *game) topStock() *cardV {
	if len(g.stock) == 0 {
		return nil
	}
	return g.stock[len(g.stock)-1]
}

/* ---------------- names, as each seat sees them ---------------- */

// who is how player i is named to the person in seat v ("you" or a name).
func (g *game) who(i, v int) string {
	if i == v && g.mode != ModeHotseat {
		return "you"
	}
	return g.players[i].name
}

func (g *game) Who(i, v int) string {
	s := g.who(i, v)
	if s == "you" {
		return "You"
	}
	return s
}

// verb picks "knock"/"knocks" for the person named.
func (g *game) verb(i, v int, you, they string) string {
	if g.who(i, v) == "you" {
		return you
	}
	return they
}

/* ---------------- layout ---------------- */

func (g *game) measure() {
	g.ch = math.Max(92, math.Min(math.Min(g.Ht*0.205, g.Wd*0.12), 200))
	g.cw = g.ch * cardW / cardH
}

func (g *game) cy() float64 { return g.Ht * 0.47 }

func (g *game) seatOf(i int) byte {
	if i == g.viewer {
		return seatS
	}
	return seatN
}

func (g *game) anchor(seat byte) (float64, float64) {
	if seat == seatS {
		return g.Wd / 2, g.Ht - g.ch*0.58 - 4
	}
	return g.Wd / 2, g.ch*0.42 + 4
}

// piles: the stock on the left of the felt, the discards on the right; at
// the end of a hand they move aside to make room for the cards laid down.
func (g *game) stockXY() (float64, float64, float64) {
	if g.show != nil {
		return g.cw*0.75 + 6, g.cy() - g.ch*0.32, 0.72
	}
	return g.Wd/2 - g.cw*0.78, g.cy(), 1
}

func (g *game) discXY() (float64, float64, float64) {
	if g.show != nil {
		return g.cw*0.75 + 6, g.cy() + g.ch*0.5, 0.72
	}
	return g.Wd/2 + g.cw*0.78, g.cy(), 1
}

// handSlots lists player P's hand cards in order (by N).
func (g *game) handSlots(P int) []*cardV {
	var out []*cardV
	for _, c := range g.deck {
		if c.zn.K == zHand && c.zn.P == P {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].zn.N < out[b].zn.N })
	return out
}

// faceUpHand: the viewer's own hand, unless a shared computer is between turns.
func (g *game) faceUpHand(P int) bool { return P == g.viewer && !g.hidden }

// handSpread is the step between your cards: as wide as fits between your
// name box on the left and the score pad on the right (with room for the
// gaps between meld groups), but never so tight the corners are hidden.
func (g *game) handSpread() float64 {
	base := math.Min(g.cw*0.6, g.Wd*0.6/11)
	avail := g.Wd - 2*(math.Min(150, g.ch*0.9)+18+200) // the coffee cup or score pad, then a name box
	fit := (avail - g.cw) / (10 + 1.5)
	return math.Max(g.cw*0.34, math.Min(base, fit))
}

// myHandPos lays your hand out in its meld groups with a little gap between them.
func (g *game) myHandPos(c *cardV) (x, y, r, s float64) {
	cs := g.handSlots(c.zn.P)
	sp := g.handSpread()
	gap := sp * 0.5
	idx, grp, groups := 0, 0, 0
	last := -1
	for i, x := range cs {
		if x.zn.M != last {
			if i > 0 {
				groups++
			}
			last = x.zn.M
		}
		if x == c {
			idx, grp = i, groups
		}
	}
	total := float64(len(cs)-1)*sp + float64(groups)*gap
	ax, ay := g.anchor(seatS)
	x = ax - total/2 + float64(idx)*sp + float64(grp)*gap
	k := (x - ax) / sp
	return x, ay + k*k*0.9, k * 1.6, 1
}

// otherHandPos fans a face-down hand. The order is by guest token, which
// says nothing about the cards.
func (g *game) otherHandPos(c *cardV) (x, y, r, s float64) {
	cs := g.handSlots(c.zn.P)
	sort.Slice(cs, func(a, b int) bool { return cs[a].tok < cs[b].tok })
	idx := 0
	for i, x := range cs {
		if x == c {
			idx = i
		}
	}
	n := len(cs)
	seat := g.seatOf(c.zn.P)
	ax, ay := g.anchor(seat)
	k := float64(idx) - float64(n-1)/2
	if seat == seatS {
		sp := g.handSpread()
		return ax + k*sp, ay + k*k*0.9, k * 1.6, 1
	}
	sp := math.Min(g.cw*0.34, g.Wd*0.36/11)
	return ax - k*sp, ay - k*k*0.5, 180 - k*1.6, 0.82
}

const showScale = 0.8

// showRow lays out the cards player P put down at the end of the hand.
func (g *game) showRow(P int) (y float64) {
	if g.seatOf(P) == seatS {
		return g.cy() + g.ch*0.52
	}
	return g.cy() - g.ch*0.5
}

func (g *game) showPos(c *cardV) (x, y, r, s float64) {
	var cs []*cardV
	for _, x := range g.deck {
		if x.zn.K == zShow && x.zn.P == c.zn.P {
			cs = append(cs, x)
		}
	}
	sort.Slice(cs, func(a, b int) bool {
		if cs[a].zn.M != cs[b].zn.M {
			return cs[a].zn.M < cs[b].zn.M
		}
		return cs[a].zn.N < cs[b].zn.N
	})
	s = showScale
	sp := g.cw * s * 0.36
	gap := g.cw * s * 0.55
	idx, grp, groups, last := 0, 0, 0, -1
	for i, x := range cs {
		if x.zn.M != last {
			if i > 0 {
				groups++
			}
			last = x.zn.M
		}
		if x == c {
			idx, grp = i, groups
		}
	}
	total := float64(len(cs)-1)*sp + float64(groups)*gap
	left := g.cw*1.5 + 12
	mid := left + (g.Wd-left-12)/2
	x = mid - total/2 + float64(idx)*sp + float64(grp)*gap
	return x, g.showRow(c.zn.P), c.zn.J * 0.4, s
}

// zonePos is where a card in a zone sits, as the viewer sees the table.
func (g *game) zonePos(c *cardV) (x, y, r, s float64, up bool, z int) {
	zn := c.zn
	N := float64(zn.N)
	switch zn.K {
	case zStock:
		sx, sy, ss := g.stockXY()
		return sx - N*0.12, sy - N*0.22, zn.J, ss, false, 20 + zn.N
	case zDisc:
		dx, dy, ds := g.discXY()
		return dx + zn.J*1.2, dy - N*0.22, zn.J, ds, zn.M == 0, 100 + zn.N
	case zHand:
		var x, y, r, s float64
		if g.faceUpHand(zn.P) {
			x, y, r, s = g.myHandPos(c)
		} else {
			x, y, r, s = g.otherHandPos(c)
		}
		return x, y, r, s, g.faceUpHand(zn.P), 200 + zn.N
	case zShow:
		x, y, r, s := g.showPos(c)
		return x, y, r, s, true, 300 + zn.M*16 + zn.N
	}
	return c.tx, c.ty, c.tr, c.ts, false, c.z
}

// layoutCards moves every card to where its zone says (animated unless instant).
func (g *game) layoutCards(instant bool) {
	for _, c := range g.deck {
		if c.zn.K == zNone {
			continue
		}
		x, y, r, s, up, z := g.zonePos(c)
		sp := c.speed
		if instant {
			c.speed = 2
		}
		if !c.placed || x != c.tx || y != c.ty || r != c.tr || s != c.ts || z != c.z {
			g.place(c, x, y, r, z, s)
		}
		g.flip(c, up && c.id != "")
		c.speed = sp
	}
}

func (g *game) place(c *cardV, x, y, r float64, z int, scale float64) {
	now := time.Now()
	c.fx, c.fy, c.fr, c.fs = c.pos(now)
	c.tx, c.ty, c.tr, c.ts = x, y, r, scale
	c.x, c.y = x, y
	c.t0 = now
	switch c.speed {
	case 0:
		c.dur = 0.42
	case 1:
		c.dur = 0.24
	default:
		c.dur = 0
	}
	if !c.placed { // first placement: appear in place rather than fly in from the corner
		c.dur, c.placed = 0, true
	}
	if z >= 0 {
		c.z = z
	}
}

func (g *game) flip(c *cardV, up bool) {
	now := time.Now()
	target := 180.0
	if up {
		target = 0
	}
	if c.fa1 == target {
		return
	}
	c.fa0, c.fa1, c.ft0 = c.angle(now), target, now
	c.fdur = 0.36
	if c.speed == 2 {
		c.fdur = 0
	}
}

// arrange sorts p's hand into its melds (then the deadwood, low to high)
// and gives every card its zone.
func (g *game) arrange(p *player) {
	a := gn.Best(cardsOf(p.hand))
	var out []*cardV
	grp := map[*cardV]int{}
	for k, m := range a.Melds {
		for _, c := range m.Cards {
			v := g.byCard[c]
			out = append(out, v)
			grp[v] = k
		}
	}
	for _, c := range a.Dead {
		v := g.byCard[c]
		out = append(out, v)
		grp[v] = deadGroup
	}
	p.hand = out
	for i, c := range p.hand {
		c.zn = zone{K: zHand, P: p.i, N: i, M: grp[c]}
	}
}

func (g *game) setPiles() {
	for i, c := range g.stock {
		c.zn = zone{K: zStock, N: i, J: c.zn.J}
	}
	for i, c := range g.disc {
		m := 0
		if c.zn.K == zDisc {
			m = c.zn.M
		}
		c.zn = zone{K: zDisc, N: i, M: m, J: c.zn.J}
	}
}

// sync lays the table out here and sends every guest their view of it.
func (g *game) sync() {
	g.layoutCards(false)
	if g.server != nil {
		g.server.sendSnaps(g, false)
	}
}

func (g *game) setTurn(i int) {
	for _, p := range g.players {
		on := p.i == i
		if on && !p.isTurn {
			p.turnAt = time.Now()
		}
		p.isTurn = on
	}
}

func (g *game) setPrompt(t string) {
	if t != g.prompt {
		g.promptAt = time.Now()
	}
	g.prompt = t
}

func (g *game) banner(title, small string, blue bool, ms float64) {
	if ms == 0 {
		ms = 1700 * g.pace
		if testFast {
			ms *= 0.2
		}
	}
	now := time.Now()
	g.ban = bannerT{title, small, blue, now, now.Add(time.Duration(ms) * time.Millisecond)}
}

/* ---------------- events everyone sees (each worded for their seat) ---------------- */

type event struct {
	K    string  `json:"k"`
	P    int     `json:"p,omitempty"`
	N    int     `json:"n,omitempty"`
	Text string  `json:"text,omitempty"`
	Sub  string  `json:"sub,omitempty"`
	Blue bool    `json:"blue,omitempty"`
	Ms   float64 `json:"ms,omitempty"`
}

// emit shows an event here and sends it to every guest, worded for their seat.
func (g *game) emit(f func(v int) *event) {
	if e := f(g.viewer); e != nil {
		g.apply(e)
	}
	if g.server != nil {
		for _, pr := range g.server.peers {
			if e := f(pr.seat); e != nil {
				pr.send(Msg{T: "ev", Ev: e})
			}
		}
	}
}

func (g *game) emitAll(e event) { g.emit(func(int) *event { c := e; return &c }) }

func (g *game) sound(name string) { g.emitAll(event{K: "snd", Text: name}) }

func (g *game) apply(e *event) {
	switch e.K {
	case "snd":
		snd.play(e.Text)
	case "say":
		if e.P < 0 || e.P > 1 || g.players[e.P] == nil {
			return
		}
		p := g.players[e.P]
		now := time.Now().Add(time.Duration(e.Ms) * time.Millisecond)
		p.say, p.sayAt, p.sayEnd = e.Text, now, now.Add(time.Duration(g.T("quip")+500)*time.Millisecond)
	case "banner":
		g.banner(e.Text, e.Sub, e.Blue, e.Ms)
	case "spark":
		x, y, _ := g.discXY()
		g.burst(x, y, 1, 12)
	case "stars":
		g.burst(g.Wd/2, g.cy(), 0, e.N)
	case "confetti":
		g.burst(0, 0, 2, e.N)
	case "prompt":
		g.setPrompt(e.Text)
	}
}

// say has a computer player say one of their lines; {p} becomes about's name.
func (g *game) say(p *player, kind string, chance float64, about int, delay float64) {
	f, ok := parseFace(p.face)
	if p.kind != KindAI || !ok || g.rng.Float64() > chance {
		return
	}
	ls := f.lines(kind)
	if len(ls) == 0 {
		return
	}
	line := pickS(g, ls)
	if about >= 0 && about < 2 {
		n := g.players[about].name
		if n == "You" || n == "" {
			line = strings.ReplaceAll(line, ", {p}", "")
		}
		line = strings.ReplaceAll(line, "{p}", n)
	}
	g.emitAll(event{K: "say", P: p.i, Text: line, Ms: delay})
}

/* ---------------- effects ---------------- */

func (g *game) burst(x, y float64, kind int, n int) {
	R := g.rng.Float64
	for i := 0; i < n; i++ {
		switch kind {
		case 0:
			a, v := -math.Pi/2+(R()-.5)*2.4, 3+R()*6
			g.parts = append(g.parts, particle{k: 0, x: x, y: y, vx: math.Cos(a) * v, vy: math.Sin(a) * v, life: 1, d: 0.009, s: 8 + R()*9, rot: (R() - .5) * 0.8, vr: (R() - .5) * .08})
		case 1:
			a, v := R()*math.Pi*2, 1.5+R()*4
			g.parts = append(g.parts, particle{k: 1, x: x, y: y, vx: math.Cos(a) * v, vy: math.Sin(a)*v - 1, life: 1, d: 0.02 + R()*0.02, s: 2 + R()*3})
		case 2:
			g.parts = append(g.parts, particle{k: 2, x: R() * g.Wd, y: -20 - R()*g.Ht*0.5, vx: (R() - .5) * 2, vy: 2 + R()*2.5, life: 1, d: 0.003, s: 6 + R()*6, rot: R() * 6, vr: (R() - .5) * .25, col: g.rand(6)})
		}
	}
}

/* ---------------- decisions: here, over the network, or the computer ---------------- */

// localTurn is true when someone at this computer decides for seat i.
func (g *game) localTurn(i int) bool { return g.players[i].kind == KindLocal && !autopilot }

// ensureView turns a shared computer over to the player in seat i: a
// "pass the computer" screen while everyone's cards are face down, then the
// table turns to their seat.
func (g *game) ensureView(i int) {
	if g.mode != ModeHotseat || (g.viewer == i && !g.hidden) {
		return
	}
	if g.viewer != i {
		g.hidden = true
		g.sync()
		g.passTo = i
		ch := make(chan struct{}, 1)
		g.passCh = ch
		func() {
			defer func() { g.passCh = nil }()
			waitOn(g, ch)
		}()
		g.viewer = i
	}
	g.hidden = false
	g.sync()
	g.sleep(g.T("move"))
}

func (g *game) hideAfter() {
	if g.mode == ModeHotseat {
		g.hidden = true
		g.sync()
	}
}

// askLocal shows a decision here and waits for it.
func (g *game) askLocal(t *turnT) turnRes {
	t.at, t.ch = time.Now(), make(chan turnRes, 1)
	g.turn = t
	g.markPicks()
	defer func() {
		g.turn = nil
		g.markPicks()
		g.setPrompt("")
	}()
	return waitOn(g, t.ch)
}

// markPicks lights up the cards the current decision can use.
func (g *game) markPicks() {
	for _, c := range g.deck {
		c.pick, c.dim = false, false
	}
	t := g.turn
	if t == nil {
		return
	}
	switch t.kind {
	case "upcard":
		if t.top != nil {
			t.top.pick = true
		}
		g.setPrompt("Take the " + label(t.top) + "? Or pass.")
	case "draw":
		if t.top != nil {
			t.top.pick = true
		}
		if t.stockTop != nil {
			t.stockTop.pick = true
		}
		g.setPrompt("Your turn: draw from the stock, or take the " + label(t.top))
	case "discard":
		ok := t.ok
		if t.knocking {
			ok = t.knock
			if len(t.gin) > 0 {
				ok = t.gin
			}
		}
		me := g.players[g.viewer]
		for _, c := range me.hand {
			c.pick, c.dim = inList(c, ok), !inList(c, ok)
		}
		switch {
		case t.knocking && len(t.gin) > 0:
			g.setPrompt("Gin! Pick the card to discard")
		case t.knocking:
			g.setPrompt("Pick the card to knock with (it goes face down)")
		default:
			g.setPrompt("Discard a card")
		}
	case "next":
		g.setPrompt("")
	}
}

func label(c *cardV) string {
	if c == nil || c.id == "" {
		return "card"
	}
	cc, _ := cardFromID(c.id)
	return cc.Label()
}

// decideUpcard: does seat i take the first upcard?
func (g *game) decideUpcard(i int) bool {
	p, top := g.players[i], g.topDisc()
	switch {
	case g.localTurn(i):
		g.ensureView(i)
		r := g.askLocal(&turnT{kind: "upcard", top: top})
		return r.pile == "disc"
	case p.kind == KindRemote:
		if r, ok := g.remoteAsk(i, &wireAsk{K: "upcard", Top: top.tok}); ok {
			return r.Pile == "disc"
		}
	}
	g.sleep(g.T("think") * (0.7 + g.rng.Float64()*0.5))
	return gn.WantsDiscard(cardsOf(p.hand), top.c)
}

// decideDraw: does seat i take the discard (true) or draw from the stock?
func (g *game) decideDraw(i int) bool {
	p, top := g.players[i], g.topDisc()
	if top == nil {
		return false
	}
	switch {
	case g.localTurn(i):
		g.ensureView(i)
		r := g.askLocal(&turnT{kind: "draw", top: top, stockTop: g.topStock()})
		return r.pile == "disc"
	case p.kind == KindRemote:
		if r, ok := g.remoteAsk(i, &wireAsk{K: "draw", Top: top.tok, Stock: g.topStock().tok}); ok {
			return r.Pile == "disc"
		}
	}
	g.sleep(g.T("think") * (0.6 + g.rng.Float64()*0.5))
	return gn.WantsDiscard(cardsOf(p.hand), top.c)
}

// discardOptions works out what seat i may do with their eleven cards.
func (g *game) discardOptions(p *player, forbid *cardV) (ok, knock, gin []*cardV, big bool) {
	var fc *gn.Card
	if forbid != nil {
		fc = &forbid.c
	}
	k, gg := gn.CanKnockWith(cardsOf(p.hand), fc)
	for _, c := range p.hand {
		if c != forbid {
			ok = append(ok, c)
		}
	}
	for _, c := range k {
		knock = append(knock, g.byCard[c])
	}
	for _, c := range gg {
		gin = append(gin, g.byCard[c])
	}
	return ok, knock, gin, gn.IsBigGin(cardsOf(p.hand))
}

func toks(cs []*cardV) []int {
	var out []int
	for _, c := range cs {
		out = append(out, c.tok)
	}
	return out
}

// decideDiscard: which card seat i throws, and whether they knock (or go big gin).
func (g *game) decideDiscard(i int, forbid *cardV) (*cardV, bool, bool) {
	p := g.players[i]
	ok, knock, gin, big := g.discardOptions(p, forbid)
	valid := func(c *cardV, kn, bg bool) bool {
		if bg {
			return big
		}
		if !inList(c, ok) {
			return false
		}
		return !kn || inList(c, knock)
	}
	switch {
	case g.localTurn(i):
		g.ensureView(i)
		r := g.askLocal(&turnT{kind: "discard", top: g.topDisc(), ok: ok, knock: knock, gin: gin, big: big})
		if valid(r.card, r.knock, r.big) {
			g.hideAfter()
			return r.card, r.knock, r.big
		}
	case p.kind == KindRemote:
		a := &wireAsk{K: "discard", Toks: toks(ok), Knock: toks(knock), Gin: toks(gin), Big: big}
		if r, okk := g.remoteAsk(i, a); okk {
			var c *cardV
			if r.Tok >= 0 && r.Tok < 52 {
				c = g.byTok[r.Tok]
			}
			if valid(c, r.Knock, r.Big) {
				return c, r.Knock, r.Big
			}
		}
	}
	g.sleep(g.T("think") * (0.7 + g.rng.Float64()*0.6))
	if big {
		return nil, true, true
	}
	var fc *gn.Card
	if forbid != nil {
		fc = &forbid.c
	}
	c, kn := gn.ChooseTurn(cardsOf(p.hand), fc, &g.mem[i], len(g.stock), g.turns[i], p.style)
	return g.byCard[c], kn, false
}

// askNext waits for "Next Hand" (or "Final Score") from whoever runs the table.
func (g *game) askNext(lbl string) {
	if autopilot || g.mode == ModeClient {
		g.sleep(g.T("banner") * 1.2)
		return
	}
	g.askLocal(&turnT{kind: "next", label: lbl})
}

/* ---------------- the passengers ---------------- */

// soloTable seats you with the passenger the title card shows and picks who comes next time.
func (g *game) soloTable() {
	g.mode, g.viewer, g.hidden = ModeSolo, 0, false
	if len(g.nextFolk) != 1 {
		g.rollNext()
	}
	g.players[0] = &player{i: 0, kind: KindLocal, name: "You", full: "You"}
	f := g.nextFolk[0]
	g.players[1] = g.aiPlayer(1, f)
	g.rollNext()
}

func (g *game) aiPlayer(i int, f folk) *player {
	p := &player{i: i, kind: KindAI, face: f.face, name: f.name, full: f.full}
	if fc, ok := parseFace(f.face); ok {
		p.style = fc.style()
	}
	return p
}

func (g *game) rollNext() {
	avoid := map[string]bool{"You": true}
	if g.players[1] != nil {
		avoid[g.players[1].name] = true
	}
	g.nextFolk = randomFolks(g.rng, 1, map[string]bool{}, avoid)
}

// fillFolks gives every empty or computer seat a random passenger.
func (g *game) fillFolks() {
	faces, names := map[string]bool{}, map[string]bool{}
	for _, p := range g.players {
		if p != nil && p.kind != KindAI {
			faces[p.face], names[p.name] = true, true
		}
	}
	var open []int
	for i, p := range g.players {
		if p == nil || p.kind == KindAI {
			open = append(open, i)
		}
	}
	fs := randomFolks(g.rng, len(open), faces, names)
	for k, i := range open {
		g.players[i] = g.aiPlayer(i, fs[k%len(fs)])
	}
}

func (g *game) setPlayer(i int, kind Kind, face, name string) *player {
	p := &player{i: i, kind: kind, face: face, name: name, full: name}
	g.players[i] = p
	return p
}

/* ---------------- a game ---------------- */

// newGame is called with the lock held.
func (g *game) newGame() {
	g.gameID++
	if g.cancel != nil {
		close(g.cancel)
	}
	g.cancel = make(chan struct{})
	g.finished, g.started = false, true
	g.pace = paceOf[g.save.Settings.Speed]
	if g.pace == 0 {
		g.pace = 1
	}
	if g.mode == ModeSolo { // a new passenger every game
		g.soloTable()
	}
	g.target = g.save.Settings.Target
	g.showResult, g.turn, g.passCh = false, nil, nil
	g.remote = map[int]*remoteWait{}
	g.prompt = ""
	g.ban = bannerT{}
	g.parts = nil
	g.show = nil
	for _, p := range g.players {
		p.hand, p.isTurn, p.say = nil, false, ""
	}
	if g.mode == ModeHotseat {
		g.viewer, g.hidden = 0, true
	}
	for _, c := range g.deck {
		c.speed, c.pick, c.dim, c.win, c.fresh = 2, false, false, false, false
		c.zn = zone{}
	}
	g.stock, g.disc = nil, nil
	g.score, g.boxes = [2]int{}, [2]int{}
	g.tallies = [2]tally{}
	g.handNo, g.dealer = 0, g.rand(2)
	if g.server != nil {
		g.server.open = false
		g.server.sendSnaps(g, true)
	}
	go g.run(g.gameID)
}

func (g *game) run(id int) {
	g.mu.Lock()
	defer func() {
		if r := recover(); r != nil && r != errCancel {
			g.mu.Unlock()
			panic(r)
		}
		g.mu.Unlock()
	}()
	if id != g.gameID {
		return
	}
	for !gn.GameOver(g.score, g.target) {
		dead := g.playHand()
		g.handNo++
		if !dead {
			g.dealer = 1 - g.dealer
		}
	}
	g.endGame()
}

// playHand deals and plays one hand; it reports a dead hand (the stock ran out).
func (g *game) playHand() (dead bool) {
	g.show = nil
	g.mem, g.turns = [2]gn.Memory{}, [2]int{}
	g.handsTot++
	for _, p := range g.players {
		p.hand = nil
	}
	g.setTurn(-1)
	// guests know a card only by a token that changes every hand
	for i, t := range g.rng.Perm(52) {
		g.deck[i].tok = t
		g.byTok[t] = g.deck[i]
	}
	deck := append([]*cardV(nil), g.deck...)
	g.rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	for _, c := range deck {
		c.speed, c.win, c.pick, c.dim, c.fresh = 1, false, false, false, false
		c.zn = zone{K: zStock, J: (g.rng.Float64() - .5) * 3}
	}
	g.stock, g.disc = deck, nil
	g.setPiles()
	g.sync()
	g.sound("shuffle")
	d := g.dealer
	g.emit(func(v int) *event {
		if g.who(d, v) == "you" {
			return &event{K: "prompt", Text: "You deal…"}
		}
		return &event{K: "prompt", Text: g.players[d].name + " deals…"}
	})
	g.sleep(g.T("move") * 1.6)
	nd := 1 - g.dealer
	for k := 0; k < 20; k++ {
		p := g.players[(nd+k)%2]
		c := g.stock[len(g.stock)-1]
		g.stock = g.stock[:len(g.stock)-1]
		p.hand = append(p.hand, c)
		g.arrange(p)
		g.setPiles()
		g.sync()
		g.sound("deal")
		g.sleep(g.T("deal") * 1.6)
	}
	// the upcard
	up := g.stock[len(g.stock)-1]
	g.stock = g.stock[:len(g.stock)-1]
	up.speed = 0
	g.disc = []*cardV{up}
	up.zn = zone{K: zDisc, J: (g.rng.Float64() - .5) * 6}
	g.setPiles()
	for _, c := range deck {
		c.speed = 0
	}
	g.sync()
	g.sound("draw")
	g.emitAll(event{K: "prompt"})
	g.sleep(g.T("move") * 1.4)

	// the first upcard goes to the non-dealer first, then the dealer; if
	// both pass, the non-dealer draws from the stock.
	cur := nd
	g.setTurn(nd)
	g.sync()
	if g.decideUpcard(nd) {
		top := g.takeDiscard(nd)
		if g.finishTurn(nd, top) {
			return false
		}
		cur = g.dealer
	} else {
		g.passUp(nd)
		g.setTurn(g.dealer)
		g.sync()
		if g.decideUpcard(g.dealer) {
			top := g.takeDiscard(g.dealer)
			if g.finishTurn(g.dealer, top) {
				return false
			}
			cur = nd
		} else {
			g.passUp(g.dealer)
			g.setTurn(nd)
			g.sync()
			g.drawStock(nd)
			if g.finishTurn(nd, nil) {
				return false
			}
			cur = g.dealer
		}
	}
	for {
		if len(g.stock) <= gn.StockFloor {
			g.deadHand()
			return true
		}
		g.setTurn(cur)
		g.sync()
		if g.decideDraw(cur) {
			top := g.takeDiscard(cur)
			if g.finishTurn(cur, top) {
				return false
			}
		} else {
			g.drawStock(cur)
			if g.finishTurn(cur, nil) {
				return false
			}
		}
		cur = 1 - cur
	}
}

func (g *game) passUp(i int) {
	g.hideAfter()
	g.emit(func(v int) *event {
		return &event{K: "prompt", Text: g.Who(i, v) + " " + g.verb(i, v, "pass", "passes") + " on the " + label(g.topDisc())}
	})
	g.say(g.players[i], "pass", 0.3, -1, 0)
	g.sleep(g.T("move") * 1.3)
	g.emitAll(event{K: "prompt"})
}

func (g *game) clearFresh() {
	for _, c := range g.deck {
		c.fresh = false
	}
}

// takeDiscard moves the top discard into seat i's hand.
func (g *game) takeDiscard(i int) *cardV {
	p := g.players[i]
	c := g.topDisc()
	g.disc = g.disc[:len(g.disc)-1]
	g.clearFresh()
	c.fresh = true
	p.hand = append(p.hand, c)
	g.arrange(p)
	g.setPiles()
	g.mem[1-i].OppTook = append(g.mem[1-i].OppTook, c.c)
	g.sync()
	g.sound("take")
	if g.players[1-i].kind != KindAI {
		g.say(p, "take", 0.3, 1-i, 200)
	}
	g.sleep(g.T("move"))
	return c
}

func (g *game) drawStock(i int) {
	p := g.players[i]
	c := g.topStock()
	g.stock = g.stock[:len(g.stock)-1]
	g.clearFresh()
	c.fresh = true
	p.hand = append(p.hand, c)
	g.arrange(p)
	g.setPiles()
	g.sync()
	g.sound("draw")
	g.sleep(g.T("move"))
}

// finishTurn has seat i discard (and maybe knock); it reports whether the hand is over.
func (g *game) finishTurn(i int, forbid *cardV) bool {
	p := g.players[i]
	g.turns[i]++
	c, knock, big := g.decideDiscard(i, forbid)
	g.clearFresh()
	if big {
		g.showdown(i, "big")
		return true
	}
	p.hand = without(p.hand, c)
	c.zn = zone{K: zDisc, J: (g.rng.Float64() - .5) * 8}
	if knock {
		c.zn.M = 1
	}
	g.disc = append(g.disc, c)
	g.arrange(p)
	g.setPiles()
	if !knock {
		g.mem[1-i].OppDiscarded = append(g.mem[1-i].OppDiscarded, c.c)
	}
	g.sync()
	g.sound("discard")
	g.sleep(g.T("play"))
	if knock {
		kind := "knock"
		if gn.Deadwood(cardsOf(p.hand)) == 0 {
			kind = "gin"
		}
		g.showdown(i, kind)
		return true
	}
	return false
}

func (g *game) deadHand() {
	g.setTurn(-1)
	g.sound("dead")
	g.emitAll(event{K: "banner", Text: "Dead Hand", Sub: "Down to the last two cards. The same dealer deals again.", Blue: true, Ms: g.T("banner") * 1.4})
	for _, p := range g.players {
		g.say(p, "dead", 0.6, -1, 300)
	}
	g.sleep(g.T("banner") * 1.6)
	g.gather()
}

// gather sweeps every card back to the stock.
func (g *game) gather() {
	for _, p := range g.players {
		p.hand = nil
	}
	g.show = nil
	g.disc = nil
	g.stock = append([]*cardV(nil), g.deck...)
	for _, c := range g.stock {
		c.win, c.fresh, c.pick, c.dim = false, false, false, false
		c.zn = zone{K: zStock, J: c.zn.J}
	}
	g.setPiles()
	g.sync()
	g.sound("sweep")
	g.sleep(g.T("move") * 1.4)
}

// showdown: seat k knocks (or goes gin). Both hands are laid down, the
// defender lays off what they can, and the hand is scored.
func (g *game) showdown(k int, kind string) {
	g.setTurn(-1)
	kp, dp := g.players[k], g.players[1-k]
	d := 1 - k
	ka := gn.Best(cardsOf(kp.hand))
	g.show = &showT{Knocker: k, Kind: kind, KDW: ka.Deadwood, Winner: -1}
	switch kind {
	case "big":
		g.sound("biggin")
		g.emit(func(v int) *event {
			return &event{K: "banner", Text: "Big Gin!", Sub: g.Who(k, v) + " " + g.verb(k, v, "meld", "melds") + " all eleven cards.", Blue: k != v && g.mode != ModeHotseat, Ms: g.T("banner") * 1.4}
		})
		g.say(kp, "gin", 1, -1, 0)
	case "gin":
		g.sound("gin")
		g.emit(func(v int) *event {
			return &event{K: "banner", Text: "Gin!", Sub: g.Who(k, v) + " " + g.verb(k, v, "go", "goes") + " gin: not a card of deadwood.", Blue: k != v && g.mode != ModeHotseat, Ms: g.T("banner") * 1.3}
		})
		g.say(kp, "gin", 1, -1, 0)
	default:
		g.sound("knock")
		g.emit(func(v int) *event {
			return &event{K: "banner", Text: "Knock, Knock!", Sub: fmt.Sprintf("%s %s with %d.", g.Who(k, v), g.verb(k, v, "knock", "knocks"), ka.Deadwood), Blue: k != v && g.mode != ModeHotseat}
		})
		g.say(kp, "knock", 0.9, -1, 0)
	}
	// the knocker lays down
	for m, ml := range ka.Melds {
		for n, c := range ml.Cards {
			g.byCard[c].zn = zone{K: zShow, P: k, M: m, N: n, J: (g.rng.Float64() - .5) * 3}
		}
	}
	for n, c := range ka.Dead {
		g.byCard[c].zn = zone{K: zShow, P: k, M: deadGroup, N: n, J: (g.rng.Float64() - .5) * 3}
	}
	kp.hand = nil
	g.sync()
	g.sound("sweep")
	g.sleep(g.T("show") * 2)

	// the defender lays down
	var da gn.Arrangement
	var lays []gn.Layoff
	if kind == "knock" {
		da, lays = gn.Defend(cardsOf(dp.hand), ka.Melds)
	} else {
		da = gn.Best(cardsOf(dp.hand))
	}
	for m, ml := range da.Melds {
		for n, c := range ml.Cards {
			g.byCard[c].zn = zone{K: zShow, P: d, M: m, N: n, J: (g.rng.Float64() - .5) * 3}
		}
	}
	dead := append([]gn.Card(nil), da.Dead...)
	for _, l := range lays {
		dead = append(dead, l.Card)
	}
	for n, c := range dead {
		g.byCard[c].zn = zone{K: zShow, P: d, M: deadGroup, N: n, J: (g.rng.Float64() - .5) * 3}
	}
	dp.hand = nil
	g.show.Done = true
	g.show.DDW = gn.Value(dead)
	g.sync()
	g.sound("sweep")
	g.sleep(g.T("show") * 2)

	// lay off, one card at a time
	melds := make([]gn.Meld, len(ka.Melds))
	copy(melds, ka.Melds)
	for _, l := range lays {
		ml := &melds[l.Meld]
		ml.Cards = append(append([]gn.Card(nil), ml.Cards...), l.Card)
		sort.Slice(ml.Cards, func(a, b int) bool {
			return ml.Cards[a].R < ml.Cards[b].R || ml.Cards[a].R == ml.Cards[b].R && ml.Cards[a].S < ml.Cards[b].S
		})
		for n, c := range ml.Cards {
			g.byCard[c].zn = zone{K: zShow, P: k, M: l.Meld, N: n, J: g.byCard[c].zn.J}
		}
		v := g.byCard[l.Card]
		v.win = true
		rest := 0
		for _, c := range g.deck {
			if c.zn.K == zShow && c.zn.P == d && c.zn.M == deadGroup {
				c.zn.N = rest
				rest++
			}
		}
		g.show.Lays++
		g.show.DDW -= l.Card.Value()
		g.sync()
		g.sound("slide")
		g.sleep(g.T("show"))
	}

	// score it
	ok := gn.Knocked
	switch kind {
	case "gin":
		ok = gn.Gin
	case "big":
		ok = gn.BigGin
	}
	res := gn.ScoreKnock(k, ka.Deadwood, da.Deadwood, ok)
	w := res.Winner
	g.show.DDW, g.show.Winner, g.show.Points, g.show.Under = da.Deadwood, w, res.Points, res.Kind == gn.Undercut
	g.score[w] += res.Points
	g.boxes[w]++
	for i := range g.tallies {
		g.tallies[i].Hands++
	}
	g.tallies[w].HandsWon++
	switch res.Kind {
	case gn.Gin, gn.BigGin:
		g.tallies[k].Gins++
	case gn.Knocked:
		g.tallies[k].Knocks++
	case gn.Undercut:
		g.tallies[d].Undercuts++
	}
	if me := g.statSeat(); me >= 0 {
		st := &g.save.Stats
		st.Hands++
		if w == me {
			st.HandsWon++
		}
		switch {
		case (res.Kind == gn.Gin || res.Kind == gn.BigGin) && k == me:
			st.Gins++
		case res.Kind == gn.Knocked && k == me:
			st.Knocks++
		case res.Kind == gn.Undercut && d == me:
			st.Undercuts++
		}
	}
	if res.Kind == gn.Undercut {
		g.sound("undercut")
		g.emit(func(v int) *event {
			return &event{K: "banner", Text: "Undercut!", Sub: fmt.Sprintf("%s %s %s: +%d", g.Who(d, v), g.verb(d, v, "undercut", "undercuts"), g.who(k, v), res.Points), Blue: d != v && g.mode != ModeHotseat, Ms: g.T("banner") * 1.5}
		})
		g.say(dp, "undercut", 1, k, 0)
		g.say(kp, "undercutBy", 0.9, d, 1100)
	} else {
		g.sound("point")
		g.emit(func(v int) *event {
			t := "Hand " + itoa(g.handNo+1)
			if w == v && g.mode != ModeHotseat {
				t = "Your Hand!"
			}
			return &event{K: "banner", Text: t, Sub: fmt.Sprintf("%s %s %d points.", g.Who(w, v), g.verb(w, v, "score", "scores"), res.Points), Blue: w != v && g.mode != ModeHotseat, Ms: g.T("banner") * 1.5}
		})
		if res.Kind == gn.Knocked && da.Deadwood >= 30 {
			g.say(dp, "caught", 0.8, k, 900)
		}
	}
	g.emit(func(v int) *event {
		if (w == v || g.mode == ModeHotseat) && (res.Kind != gn.Knocked) {
			return &event{K: "stars", N: 50}
		}
		if w == v || g.mode == ModeHotseat {
			return &event{K: "spark"}
		}
		return nil
	})
	g.sync()
	g.persist()
	g.sleep(g.T("banner") * 1.2)
	lbl := "Next Hand"
	if gn.GameOver(g.score, g.target) {
		lbl = "Final Score"
	}
	g.askNext(lbl)
	g.gather()
}

// statSeat is the seat whose record this computer keeps (none on a shared computer).
func (g *game) statSeat() int {
	if g.mode == ModeHotseat {
		return -1
	}
	return 0
}

// resultFor is the end-of-game card for seat v.
func (g *game) resultFor(v int) resultT {
	f := g.final
	var r resultT
	for k := 0; k < 2; k++ {
		i := (v + k) % 2
		nm := g.players[i].full
		if g.who(i, v) == "you" {
			nm = "You"
		}
		det := fmt.Sprintf("%d points + %d %s", g.score[i], g.boxes[i], plural(g.boxes[i], "box", "boxes"))
		if f.GameBonuses[i] > 0 {
			det += " + game"
		}
		if f.Shutout && i == f.Winner {
			det += ", doubled"
		}
		pl := 2
		if i == f.Winner {
			pl = 1
		}
		r.rows = append(r.rows, resRow{Name: nm, Face: g.players[i].face, Score: f.Totals[i], Detail: det, Place: pl, Me: i == v && g.mode != ModeHotseat})
	}
	sort.SliceStable(r.rows, func(a, b int) bool { return r.rows[a].Place < r.rows[b].Place })
	w := f.Winner
	r.sub = fmt.Sprintf("%d to %d, by %d.", f.Totals[w], f.Totals[1-w], f.Margin)
	switch {
	case g.mode == ModeHotseat:
		r.kick, r.title, r.won = "End of the line", g.players[w].name+" Wins!", true
	case w == v:
		r.kick, r.title, r.won = "Next stop: victory", "You Win!", true
	default:
		r.kick, r.title = "All change", g.players[w].name+" Wins"
	}
	if f.Shutout {
		r.kick = "A shutout! Double score"
	}
	return r
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func (g *game) endGame() {
	g.finished = true
	g.setTurn(-1)
	g.emitAll(event{K: "prompt"})
	g.final = gn.FinalScore(g.score, g.boxes)
	w := g.final.Winner
	if me := g.statSeat(); me >= 0 {
		g.recordGame(w == me, g.final.Margin, g.final.Shutout)
	}
	g.emit(func(v int) *event {
		if w == v || g.mode == ModeHotseat {
			return &event{K: "snd", Text: "win"}
		}
		return &event{K: "snd", Text: "lose"}
	})
	g.emit(func(v int) *event {
		if w == v || g.mode == ModeHotseat {
			return &event{K: "confetti", N: 160}
		}
		return nil
	})
	for _, p := range g.players {
		if p.kind == KindAI {
			if p.i == w {
				g.say(p, "win", 1, -1, 0)
			} else {
				g.say(p, "lose", 1, -1, 0)
			}
		}
	}
	ms := 1400.0
	if testFast {
		ms = 140
	}
	g.sleep(ms)
	g.result = g.resultFor(g.viewer)
	g.result.rec = g.recordText()
	g.showResult = true
	g.resultAt = time.Now()
	if g.server != nil {
		g.server.sendResults(g)
	}
}

func (g *game) recordGame(won bool, margin int, shutout bool) {
	st := &g.save.Stats
	st.Played++
	if won {
		st.Won++
		st.Streak++
		st.BestStreak = max(st.BestStreak, st.Streak)
		st.Best = max(st.Best, margin)
		if shutout {
			st.Shutouts++
		}
	} else {
		st.Lost++
		st.Streak = 0
	}
	g.persist()
}

func (g *game) recordText() string {
	s := g.save.Stats
	if s.Played == 0 {
		return "No games played yet."
	}
	best := ""
	if s.Won > 0 {
		best = fmt.Sprintf(" · Best win: by %d", s.Best)
	}
	return fmt.Sprintf("Games: %d · Won: %d · Gins: %d · Undercuts: %d%s", s.Played, s.Won, s.Gins, s.Undercuts, best)
}

// handsFromZones rebuilds each player's hand and the piles from the card zones (guests).
func (g *game) handsFromZones() {
	for _, p := range g.players {
		p.hand = p.hand[:0]
	}
	g.stock, g.disc = g.stock[:0], g.disc[:0]
	for _, c := range g.deck {
		switch c.zn.K {
		case zHand:
			if c.zn.P >= 0 && c.zn.P < 2 {
				g.players[c.zn.P].hand = append(g.players[c.zn.P].hand, c)
			}
		case zStock:
			g.stock = append(g.stock, c)
		case zDisc:
			g.disc = append(g.disc, c)
		}
	}
	for _, p := range g.players {
		sort.Slice(p.hand, func(a, b int) bool { return p.hand[a].zn.N < p.hand[b].zn.N })
	}
	sort.Slice(g.stock, func(a, b int) bool { return g.stock[a].zn.N < g.stock[b].zn.N })
	sort.Slice(g.disc, func(a, b int) bool { return g.disc[a].zn.N < g.disc[b].zn.N })
}

// myDeadwood is the deadwood in the viewer's own hand (-1 if it isn't showing).
func (g *game) myDeadwood() (int, int) {
	p := g.players[g.viewer]
	if p == nil || len(p.hand) == 0 || !g.faceUpHand(g.viewer) {
		return -1, 0
	}
	var cs []gn.Card
	for _, c := range p.hand {
		cc, ok := cardFromID(c.id)
		if !ok {
			return -1, 0
		}
		cs = append(cs, cc)
	}
	if len(cs) == 11 { // after drawing: what's left after the best discard
		var forbid *gn.Card
		if t := g.turn; t != nil && t.kind == "discard" {
			for _, c := range p.hand {
				if !inList(c, t.ok) {
					cc, _ := cardFromID(c.id)
					forbid = &cc
				}
			}
		}
		_, d := gn.BestDiscard(cs, forbid)
		return d, 11
	}
	return gn.Deadwood(cs), len(cs)
}
