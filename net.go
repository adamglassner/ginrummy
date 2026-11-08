// Online play, the same way Hearts, Euchre and Go Fish do it. One player hosts:
// their game runs the table and sends the guest a snapshot of what the guest
// can see (their own hand face up, the other hand face down, cards known only
// by a token that changes every hand) plus the bubbles, banners and sounds,
// worded for their seat. The guest only sends back draws and discards.
// Messages are JSON, one per line, over TCP.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	netPort  = 47250
	protoVer = 1
)

type Msg struct {
	T     string      `json:"t"`
	Name  string      `json:"name,omitempty"`
	Face  string      `json:"face,omitempty"`
	Ver   int         `json:"ver,omitempty"`
	Seat  int         `json:"seat,omitempty"`
	Why   string      `json:"why,omitempty"`
	Lobby []lobbySeat `json:"lobby,omitempty"`
	Snap  *snap       `json:"snap,omitempty"`
	Ev    *event      `json:"ev,omitempty"`
	Res   *wireRes    `json:"res,omitempty"`
	Ask   *wireAsk    `json:"ask,omitempty"`
	Act   *wireAct    `json:"act,omitempty"`
}

type lobbySeat struct {
	Name string `json:"name"`
	Face string `json:"face"`
	Kind Kind   `json:"kind"`
}

// wireAsk is a decision the host needs from a guest.
type wireAsk struct {
	K     string `json:"k"`               // "upcard", "draw" or "discard"
	Top   int    `json:"top"`             // the discard pile's top card
	Stock int    `json:"stock"`           // the stock's top card
	Toks  []int  `json:"toks,omitempty"`  // cards they may discard
	Knock []int  `json:"knock,omitempty"` // discards that let them knock
	Gin   []int  `json:"gin,omitempty"`   // discards that make gin
	Big   bool   `json:"big,omitempty"`   // all eleven cards meld
}

type wireAct struct {
	Pile  string `json:"pile,omitempty"` // "stock", "disc", "pass"
	Tok   int    `json:"tok"`
	Knock bool   `json:"knock,omitempty"`
	Big   bool   `json:"big,omitempty"`
	Drop  bool   `json:"-"` // the guest left
}

type remoteWait struct {
	seat int
	ask  *wireAsk
	ch   chan wireAct
}

type snapCard struct {
	T  int    `json:"t"`
	Z  zone   `json:"z"`
	Sp int    `json:"sp,omitempty"`
	ID string `json:"id,omitempty"`
	W  bool   `json:"w,omitempty"`
	F  bool   `json:"f,omitempty"`
}

type snapPlayer struct {
	Name, Full, Face string
	Kind             Kind
	Turn             bool
}

type snap struct {
	Gen, Hand, HandNo int
	Viewer, Dealer    int
	Score, Boxes      [2]int
	Target            int
	Show              *showT
	P                 [2]snapPlayer
	C                 []snapCard
	Fresh, Finished   bool
}

type wireRes struct {
	Kick, Title, Sub string
	Rows             []resRow
	Won, Shutout     bool
	Score            int // the winning margin
	Tally            tally
}

func encode(m Msg) []byte {
	b, err := json.Marshal(m)
	if err != nil {
		log.Print("net: ", err)
		return nil
	}
	return append(b, '\n')
}

// conn is one end of a connection with a buffered writer goroutine.
type conn struct {
	c      net.Conn
	out    chan []byte
	closed bool
}

func newConn(c net.Conn) *conn {
	k := &conn{c: c, out: make(chan []byte, 1024)}
	go func() {
		for b := range k.out {
			_ = c.SetWriteDeadline(time.Now().Add(15 * time.Second))
			if _, err := c.Write(b); err != nil {
				_ = c.Close()
				for range k.out { // drain
				}
				return
			}
		}
		_ = c.Close()
	}()
	return k
}

func (k *conn) send(m Msg) {
	if k == nil || k.closed {
		return
	}
	b := encode(m)
	select {
	case k.out <- b:
	default: // hopelessly behind: give up on this connection
		k.close()
	}
}

func (k *conn) close() {
	if !k.closed {
		k.closed = true
		close(k.out)
	}
}

/* ---------------- hosting ---------------- */

type peer struct {
	*conn
	seat int
	name string
}

type server struct {
	ln    net.Listener
	peers []*peer
	addrs []string
	open  bool // taking new players (not mid-game)
}

// hostGame starts listening and opens the lobby. You sit at seat 0; the
// other seat is a passenger until a friend takes it.
func (g *game) hostGame() {
	g.leaveOnline()
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", netPort))
	if err != nil {
		g.showMsg("Couldn't open a game on port " + strconv.Itoa(netPort) + ". Is another copy of Gin Rummy already hosting?\n\n" + err.Error())
		return
	}
	g.closePanels()
	g.server = &server{ln: ln, addrs: localAddrs(), open: true}
	g.mode, g.viewer, g.hidden = ModeHost, 0, false
	g.resetDeck(false)
	g.started, g.finished = false, false
	for i := range g.players {
		g.players[i] = nil
	}
	g.setPlayer(0, KindLocal, g.setup.face, cleanName(g.setup.name, "Host"))
	g.fillFolks()
	g.showLobby = true
	g.sendLobby()
	srv := g.server
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go g.handshake(srv, c)
		}
	}()
}

func localAddrs() []string {
	var out []string
	ifs, _ := net.InterfaceAddrs()
	for _, a := range ifs {
		if ip, ok := a.(*net.IPNet); ok && ip.IP.To4() != nil && !ip.IP.IsLoopback() && !ip.IP.IsLinkLocalUnicast() {
			out = append(out, ip.IP.String())
		}
	}
	if len(out) > 3 {
		out = out[:3]
	}
	return out
}

func cleanName(s, def string) string {
	s = strings.TrimSpace(s)
	if len([]rune(s)) > 14 {
		s = string([]rune(s)[:14])
	}
	if s == "" {
		return def
	}
	return s
}

func (g *game) handshake(srv *server, c net.Conn) {
	_ = c.SetReadDeadline(time.Now().Add(10 * time.Second))
	dec := json.NewDecoder(bufio.NewReader(c))
	var m Msg
	k := newConn(c)
	if err := dec.Decode(&m); err != nil || m.T != "hello" {
		k.close()
		return
	}
	_ = c.SetReadDeadline(time.Time{})
	g.mu.Lock()
	if g.server != srv {
		g.mu.Unlock()
		k.close()
		return
	}
	reject, seat := "", -1
	switch {
	case m.Ver != protoVer:
		reject = "That game is running a different version of Gin Rummy. Update both copies to the same version and try again."
	case !srv.open:
		reject = "That game is in the middle of a hand. Ask the host to let you in after this game."
	default:
		if g.players[1].kind == KindAI {
			seat = 1
		} else {
			reject = "That table is full: someone already took the other seat."
		}
	}
	if reject != "" {
		k.send(Msg{T: "reject", Why: reject})
		k.close()
		g.mu.Unlock()
		return
	}
	name := cleanName(m.Name, "Friend")
	for _, p := range g.players { // two people with the same name get told apart
		if p.kind != KindAI && strings.EqualFold(p.name, name) {
			name = cleanName(name, "Friend") + " 2"
		}
	}
	f := m.Face
	if !faceOK(f) {
		f = pickerFaces[g.rand(len(pickerFaces))]
	}
	pr := &peer{conn: k, seat: seat, name: name}
	srv.peers = append(srv.peers, pr)
	g.setPlayer(seat, KindRemote, f, name).peer = pr
	g.fillFolks()
	k.send(Msg{T: "welcome", Seat: seat})
	g.sendLobby()
	if g.started && g.finished { // between games: show them the table as it is
		pr.send(Msg{T: "snap", Snap: g.snapFor(seat, false)})
	}
	g.mu.Unlock()

	for {
		var m Msg
		if err := dec.Decode(&m); err != nil {
			break
		}
		if m.T == "bye" {
			break
		}
		g.mu.Lock()
		if m.T == "act" && m.Act != nil && g.players[pr.seat].peer == pr {
			if w := g.remote[pr.seat]; w != nil && validAct(w.ask, m.Act) {
				delete(g.remote, pr.seat)
				w.ch <- *m.Act
			}
		}
		g.mu.Unlock()
	}
	g.mu.Lock()
	g.dropPeer(pr)
	g.mu.Unlock()
}

func validAct(a *wireAsk, r *wireAct) bool {
	in := func(t int, l []int) bool {
		for _, x := range l {
			if x == t {
				return true
			}
		}
		return false
	}
	switch a.K {
	case "upcard":
		return r.Pile == "disc" || r.Pile == "pass"
	case "draw":
		return r.Pile == "disc" || r.Pile == "stock"
	}
	if r.Big {
		return a.Big
	}
	if r.Knock {
		return in(r.Tok, a.Knock)
	}
	return in(r.Tok, a.Toks)
}

// dropPeer hands a departed guest's seat to a computer player.
func (g *game) dropPeer(pr *peer) {
	pr.close()
	srv := g.server
	if srv == nil {
		return
	}
	keep := srv.peers[:0]
	for _, q := range srv.peers {
		if q != pr {
			keep = append(keep, q)
		}
	}
	srv.peers = keep
	p := g.players[pr.seat]
	if p == nil || p.peer != pr {
		return
	}
	if g.started && !g.finished {
		// a computer player takes over their cards, under the same name
		p.peer, p.kind = nil, KindAI
		f := randomFolks(g.rng, 1, g.facesInUse(), g.namesInUse())
		if len(f) == 1 {
			p.face = f[0].face
		}
		g.emitAll(event{K: "banner", Text: p.name + " left", Sub: "A computer player takes over their cards.", Blue: true})
		if w := g.remote[pr.seat]; w != nil {
			delete(g.remote, pr.seat)
			w.ch <- wireAct{Drop: true}
		}
		g.sync()
	} else {
		g.players[pr.seat] = &player{i: pr.seat, kind: KindAI}
		g.fillFolks()
	}
	g.sendLobby()
}

func (g *game) facesInUse() map[string]bool {
	m := map[string]bool{}
	for _, p := range g.players {
		if p != nil {
			m[p.face] = true
		}
	}
	return m
}

func (g *game) namesInUse() map[string]bool {
	m := map[string]bool{}
	for _, p := range g.players {
		if p != nil {
			m[p.name] = true
		}
	}
	return m
}

func (g *game) sendLobby() {
	if g.server == nil {
		return
	}
	var seats []lobbySeat
	for _, p := range g.players {
		seats = append(seats, lobbySeat{Name: p.full, Face: p.face, Kind: p.kind})
	}
	g.lobby = seats
	for _, pr := range g.server.peers {
		pr.send(Msg{T: "lobby", Lobby: seats})
	}
}

// snapFor builds what the person in seat v gets to see.
func (g *game) snapFor(v int, fresh bool) *snap {
	s := &snap{Gen: g.gameID, Hand: g.handsTot, HandNo: g.handNo, Viewer: v, Dealer: g.dealer, Score: g.score, Boxes: g.boxes,
		Target: g.target, Fresh: fresh, Finished: g.finished}
	if g.show != nil {
		sh := *g.show
		s.Show = &sh
	}
	for i, p := range g.players {
		k := p.kind
		if i == v {
			k = KindLocal
		} else if k == KindLocal {
			k = KindRemote // the host, as a guest sees them
		}
		s.P[i] = snapPlayer{Name: p.name, Full: p.full, Face: p.face, Kind: k, Turn: p.isTurn}
	}
	// the other hand goes in token order, with no meld groups: it says nothing about the cards
	other := g.handSlots(1 - v)
	sort.Slice(other, func(a, b int) bool { return other[a].tok < other[b].tok })
	pos := map[*cardV]int{}
	for i, c := range other {
		pos[c] = i
	}
	for _, c := range g.deck {
		if c.zn.K == zNone {
			continue
		}
		z := c.zn
		mine := z.K == zHand && z.P == v
		if z.K == zHand && !mine {
			z.N, z.M = pos[c], 0
		}
		sc := snapCard{T: c.tok, Z: z, Sp: c.speed, W: c.win, F: c.fresh && mine}
		if mine || z.K == zShow || z.K == zDisc && z.M == 0 {
			sc.ID = c.id
		}
		s.C = append(s.C, sc)
	}
	return s
}

func (srv *server) sendSnaps(g *game, fresh bool) {
	for _, pr := range srv.peers {
		pr.send(Msg{T: "snap", Snap: g.snapFor(pr.seat, fresh)})
	}
}

func (srv *server) sendResults(g *game) {
	srv.open = true
	for _, pr := range srv.peers {
		r := g.resultFor(pr.seat)
		pr.send(Msg{T: "res", Res: &wireRes{Kick: r.kick, Title: r.title, Sub: r.sub, Rows: r.rows, Won: r.won,
			Score: g.final.Margin, Shutout: g.final.Shutout, Tally: g.tallies[pr.seat]}})
	}
	g.sendLobby()
}

// startHosted deals the first (or next) game for everyone at the table.
func (g *game) startHosted() {
	if g.server == nil {
		return
	}
	g.closePanels()
	g.newGame()
}

// remoteAsk sends a guest a decision and waits for their answer. It reports
// false if they left (a computer player decides instead).
func (g *game) remoteAsk(i int, a *wireAsk) (wireAct, bool) {
	if !g.remoteSend(i, a) {
		return wireAct{}, false
	}
	return g.remoteWait(i)
}

// remoteSend asks a guest something without waiting.
func (g *game) remoteSend(i int, a *wireAsk) bool {
	p := g.players[i]
	if p.peer == nil || p.peer.closed {
		return false
	}
	g.remote[i] = &remoteWait{seat: i, ask: a, ch: make(chan wireAct, 1)}
	p.peer.send(Msg{T: "ask", Ask: a})
	return true
}

func (g *game) remoteWait(i int) (wireAct, bool) {
	w := g.remote[i]
	if w == nil {
		return wireAct{}, false
	}
	r := waitOn(g, w.ch)
	return r, !r.Drop
}

/* ---------------- joining ---------------- */

type client struct {
	*conn
	addr string
}

// joinGame connects to a host ("192.168.1.20" or "host:port").
func (g *game) joinGame(addr string) {
	g.leaveOnline()
	addr = strings.TrimSpace(addr)
	if addr == "" {
		g.showMsg("Type the host's address first (it's shown on their screen).")
		return
	}
	if _, _, err := net.SplitHostPort(addr); err != nil {
		addr = net.JoinHostPort(addr, strconv.Itoa(netPort))
	}
	cl := &client{addr: addr}
	g.client = cl
	g.showMsg("Heading over to " + addr + " …")
	name, f := cleanName(g.setup.name, "Friend"), g.setup.face
	go func() {
		c, err := net.DialTimeout("tcp", addr, 8*time.Second)
		g.mu.Lock()
		if g.client != cl {
			g.mu.Unlock()
			if err == nil {
				_ = c.Close()
			}
			return
		}
		if err != nil {
			g.client = nil
			g.showMsg("Couldn't reach a game at " + addr + ".\n\nCheck the address, and that the host's firewall lets Gin Rummy through.")
			g.mu.Unlock()
			return
		}
		cl.conn = newConn(c)
		cl.send(Msg{T: "hello", Name: name, Face: f, Ver: protoVer})
		g.mu.Unlock()
		dec := json.NewDecoder(bufio.NewReader(c))
		for {
			var m Msg
			if err := dec.Decode(&m); err != nil {
				break
			}
			g.mu.Lock()
			if g.client == cl {
				g.clientMsg(&m)
			}
			g.mu.Unlock()
		}
		g.mu.Lock()
		if g.client == cl {
			g.client = nil
			g.leaveOnline()
			g.toMenu()
			g.showMsg("Lost the connection to the host.")
		}
		g.mu.Unlock()
	}()
}

func (cl *client) send(m Msg) {
	if cl != nil && cl.conn != nil {
		cl.conn.send(m)
	}
}

// becomeGuest sets this computer up to show someone else's table.
func (g *game) becomeGuest(seat int) {
	g.gameID++
	if g.cancel != nil {
		close(g.cancel)
	}
	g.cancel = make(chan struct{})
	g.dropClientAsk()
	g.mode, g.viewer, g.hidden = ModeClient, seat, false
	g.started, g.finished = false, false
	g.resetDeck(true)
	g.score, g.boxes, g.show = [2]int{}, [2]int{}, nil
	for i := range g.players {
		g.players[i] = &player{i: i, kind: KindAI}
	}
	g.closePanels()
	g.showTitle, g.showResult = false, false
	g.showLobby = true
}

func (g *game) clientMsg(m *Msg) {
	switch m.T {
	case "welcome":
		if g.mode != ModeClient {
			g.becomeGuest(m.Seat)
		} else {
			g.viewer = m.Seat
		}
	case "lobby":
		g.lobby = m.Lobby
	case "reject":
		g.dropClient()
		g.showMsg(m.Why)
	case "closed":
		g.dropClient()
		g.leaveOnline()
		g.toMenu()
		g.showMsg("The host closed the game.")
	case "snap":
		if m.Snap != nil {
			g.applySnap(m.Snap)
		}
	case "ev":
		if m.Ev != nil {
			g.apply(m.Ev)
		}
	case "ask":
		if m.Ask != nil {
			g.clientAsk(m.Ask)
		}
	case "res":
		if r := m.Res; r != nil {
			g.dropClientAsk()
			g.finished = true
			st := &g.save.Stats
			st.Hands += r.Tally.Hands
			st.HandsWon += r.Tally.HandsWon
			st.Gins += r.Tally.Gins
			st.Knocks += r.Tally.Knocks
			st.Undercuts += r.Tally.Undercuts
			g.recordGame(r.Won, r.Score, r.Shutout)
			g.result = resultT{kick: r.Kick, title: r.Title, sub: r.Sub, rows: r.Rows, won: r.Won, rec: g.recordText()}
			g.showResult, g.resultAt = true, time.Now()
		}
	}
}

func (g *game) dropClient() {
	cl := g.client
	g.client = nil
	if cl != nil && cl.conn != nil {
		cl.close()
	}
	g.dropClientAsk()
}

// dropClientAsk forgets a question the host asked us.
func (g *game) dropClientAsk() {
	if g.cliEnd != nil {
		close(g.cliEnd)
		g.cliEnd = nil
	}
	g.turn = nil
	for _, c := range g.deck {
		c.pick, c.dim = false, false
	}
	g.setPrompt("")
}

func (g *game) applySnap(s *snap) {
	if s.Gen != g.gameID || s.Fresh {
		g.gameID = s.Gen
		g.dropClientAsk()
		g.closePanels()
		g.showTitle, g.showResult = false, false
		g.parts, g.ban = nil, bannerT{}
	}
	if s.Hand != g.handsTot { // new tokens every hand
		g.handsTot = s.Hand
		for _, c := range g.deck {
			c.id = ""
		}
	}
	g.started, g.finished = true, s.Finished
	g.viewer, g.dealer, g.score, g.boxes, g.target = s.Viewer&1, s.Dealer&1, s.Score, s.Boxes, s.Target
	g.handNo, g.show = s.HandNo, s.Show
	if g.show != nil {
		g.show.Knocker &= 1
	}
	for i, sp := range s.P {
		p := g.players[i]
		p.i, p.name, p.full, p.face, p.kind = i, sp.Name, sp.Full, sp.Face, sp.Kind
		if !faceOK(p.face) {
			p.face = ""
		}
		if sp.Turn && !p.isTurn {
			p.turnAt = time.Now()
		}
		p.isTurn = sp.Turn
	}
	seen := map[int]bool{}
	for _, sc := range s.C {
		if sc.T < 0 || sc.T >= 52 {
			continue
		}
		c := g.byTok[sc.T]
		seen[sc.T] = true
		c.zn, c.speed, c.win, c.fresh = sc.Z, sc.Sp, sc.W, sc.F
		c.zn.P &= 1
		if c.zn.M < 0 || c.zn.M > deadGroup {
			c.zn.M = 0
		}
		if sc.ID != "" {
			if cc, ok := cardFromID(sc.ID); ok {
				c.id, c.c = sc.ID, cc
			}
		}
	}
	for t, c := range g.byTok {
		if !seen[t] {
			c.zn = zone{}
		}
	}
	g.handsFromZones()
	g.layoutCards(false)
}

// clientAsk shows the host's question here and sends the answer back.
func (g *game) clientAsk(a *wireAsk) {
	g.dropClientAsk()
	end := make(chan struct{})
	g.cliEnd = end
	cl := g.client
	tok := func(t int) *cardV {
		if t >= 0 && t < 52 {
			return g.byTok[t]
		}
		return nil
	}
	list := func(ts []int) []*cardV {
		var out []*cardV
		for _, t := range ts {
			if c := tok(t); c != nil {
				out = append(out, c)
			}
		}
		return out
	}
	t := &turnT{kind: a.K, top: tok(a.Top), stockTop: tok(a.Stock), ok: list(a.Toks), knock: list(a.Knock), gin: list(a.Gin), big: a.Big, at: time.Now(), ch: make(chan turnRes, 1)}
	g.turn = t
	g.markPicks()
	go func() {
		select {
		case r := <-t.ch:
			act := &wireAct{Pile: r.pile, Tok: -1, Knock: r.knock, Big: r.big}
			if r.card != nil {
				act.Tok = r.card.tok
			}
			cl.send(Msg{T: "act", Act: act})
			g.mu.Lock()
			if g.turn == t {
				g.turn = nil
				g.markPicks()
				g.setPrompt("")
			}
			g.mu.Unlock()
		case <-end:
		}
	}()
	if autopilot {
		go func() {
			time.Sleep(30 * time.Millisecond)
			g.mu.Lock()
			defer g.mu.Unlock()
			if g.turn != t {
				return
			}
			g.turn = nil
			switch t.kind {
			case "upcard":
				t.ch <- turnRes{pile: []string{"disc", "pass"}[g.rng.IntN(2)]}
			case "draw":
				t.ch <- turnRes{pile: []string{"disc", "stock"}[g.rng.IntN(2)]}
			default:
				switch {
				case t.big:
					t.ch <- turnRes{big: true}
				case len(t.gin) > 0:
					t.ch <- turnRes{card: t.gin[0], knock: true}
				case len(t.knock) > 0 && g.rng.IntN(2) == 0:
					t.ch <- turnRes{card: t.knock[0], knock: true}
				default:
					t.ch <- turnRes{card: t.ok[g.rng.IntN(len(t.ok))]}
				}
			}
		}()
	}
}

// leaveOnline closes any online game we're hosting or in.
func (g *game) leaveOnline() {
	if srv := g.server; srv != nil {
		g.server = nil
		_ = srv.ln.Close()
		for _, pr := range srv.peers {
			pr.send(Msg{T: "closed"})
			pr.close()
		}
		for _, p := range g.players {
			if p != nil && p.kind == KindRemote {
				p.kind, p.peer = KindAI, nil
			}
		}
	}
	if g.mode == ModeHost {
		g.mode = ModeSolo
	}
	if cl := g.client; cl != nil {
		g.client = nil
		if cl.conn != nil {
			cl.send(Msg{T: "bye"})
			cl.close()
		}
	}
	if g.mode == ModeClient { // back to a table of our own
		g.gameID++
		if g.cancel != nil {
			close(g.cancel)
		}
		g.cancel = make(chan struct{})
		g.dropClientAsk()
		g.resetDeck(false)
		g.soloTable()
		g.started, g.finished = false, false
		g.score, g.boxes, g.show = [2]int{}, [2]int{}, nil
	}
	g.lobby = nil
	g.showLobby = false
}

// lobbyRows is the lobby list as this computer should show it (seat order).
func (g *game) lobbyRows() []lobbySeat {
	if g.mode == ModeHost || len(g.lobby) != 2 {
		var out []lobbySeat
		for _, p := range g.players {
			out = append(out, lobbySeat{Name: p.full, Face: p.face, Kind: p.kind})
		}
		return out
	}
	return g.lobby
}
