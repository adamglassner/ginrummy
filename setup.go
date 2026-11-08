// Play Together: who sits where on a shared computer, and the name and face
// you use online. (The screens for these are in setup_ui.go.)
package main

import (
	"fmt"
	"strings"
)

type setupState struct {
	hotNames   [2]string
	hotFaces   [2]string
	name, face string
	addr       string
}

func newSetup(sv *saveData) setupState {
	s := sv.Settings
	st := setupState{name: s.Name, face: s.Face, addr: s.Host}
	for i := 0; i < 2; i++ {
		st.hotNames[i] = fmt.Sprintf("Player %d", i+1)
		st.hotFaces[i] = pickerFaces[i]
		if i < len(s.HotNames) && strings.TrimSpace(s.HotNames[i]) != "" {
			st.hotNames[i] = s.HotNames[i]
		}
		if i < len(s.HotFaces) && faceOK(s.HotFaces[i]) {
			st.hotFaces[i] = s.HotFaces[i]
		}
	}
	if strings.TrimSpace(st.name) == "" {
		st.name = "Player"
	}
	if !faceOK(st.face) {
		st.face = pickerFaces[0]
	}
	return st
}

// keepSetup saves the setup choices with the settings.
func (g *game) keepSetup() {
	s := &g.save.Settings
	st := &g.setup
	s.Name, s.Face, s.Host = st.name, st.face, st.addr
	s.HotNames, s.HotFaces = st.hotNames[:], st.hotFaces[:]
	g.persist()
}

func (g *game) startHotseat() {
	g.leaveOnline()
	g.keepSetup()
	st := &g.setup
	g.mode, g.viewer = ModeHotseat, 0
	for i := range g.players {
		g.players[i] = nil
	}
	for k := 0; k < 2; k++ {
		name := cleanName(st.hotNames[k], fmt.Sprintf("Player %d", k+1))
		if k == 1 && strings.EqualFold(g.players[0].name, name) { // two people with the same name get told apart
			name = cleanName(name, "Player") + " 2"
		}
		g.setPlayer(k, KindLocal, st.hotFaces[k], name)
	}
	g.closePanels()
	g.showTitle = false
	g.startAudio()
	g.newGame()
}

func (g *game) startAudio() {
	snd.setVolume(float64(g.save.Settings.Vol) / 100)
	snd.setSfx(g.save.Settings.Sfx)
	snd.setMusic(g.save.Settings.Music)
	snd.startMusic()
}

// goPlay is "Deal" on the title card: you and a random passenger.
func (g *game) goPlay() {
	g.leaveOnline()
	g.mode = ModeSolo
	g.closePanels()
	g.showTitle = false
	g.startAudio()
	g.newGame()
}

// newDeal is New Game / F2 / Play Again: the same table, dealt again.
func (g *game) newDeal() {
	switch g.mode {
	case ModeClient:
		g.showMsg("Only the host can deal a new game.")
	case ModeHost:
		g.showTitle = false
		g.startHosted()
	case ModeHotseat:
		g.closePanels()
		g.showTitle, g.showResult = false, false
		g.newGame()
	default:
		g.goPlay()
	}
}

func (g *game) toMenu() {
	g.showResult = false
	g.showTitle = true
}

func (g *game) closePanels() {
	g.showTogether, g.showHotseat, g.showOnline, g.showLobby = false, false, false, false
	g.message, g.focus = "", ""
}

func (g *game) showMsg(s string) { g.message = s }
