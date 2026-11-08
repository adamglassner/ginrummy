// The passengers who might sit down across from you in the club car. Every face is made up
// fresh from a seed, and its name, title and way of talking always fit the
// face: a lady gets a lady's name, the dining-car chef in his toque is "Chef", a grandmother
// with silver hair is "Grandma". (The drawing itself is in cameo.go.)
package main

import (
	gn "github.com/adamglassner/ginrummy/internal/gin"
	"image/color"
	"math/rand/v2"
	"strconv"
	"strings"
)

// Face is all a seat needs to share over the network: the look, the name and
// the manner are all worked out from it. As text it is "f:<seed>" or "m:<seed>".
type Face struct {
	Female bool
	Seed   uint64
}

func (f Face) String() string {
	if f.Female {
		return "f:" + strconv.FormatUint(f.Seed, 10)
	}
	return "m:" + strconv.FormatUint(f.Seed, 10)
}

func parseFace(s string) (Face, bool) {
	if len(s) < 3 || s[1] != ':' || (s[0] != 'f' && s[0] != 'm') {
		return Face{}, false
	}
	n, err := strconv.ParseUint(s[2:], 10, 64)
	if err != nil {
		return Face{}, false
	}
	return Face{Female: s[0] == 'f', Seed: n}, true
}

func faceOK(s string) bool { _, ok := parseFace(s); return ok }

/* ---------------- the look ---------------- */

type Look struct {
	Lady, Stern, Smile      bool
	Skin, Hair, Brow, Lip   color.RGBA
	HairStyle               string // bob, bun, waves, curls | short, side, bald, curly
	Top                     string // stripes, blouse, turtleneck, suit, chef, cardigan
	Coat, Dark, Accent      color.RGBA
	Stripe                  color.RGBA
	Neck                    string // scarf, pearls, bowtie, tie, kerchief, ""
	NeckColor               color.RGBA
	Hat                     string // beret, cloche, cap, fedora, toque, ""
	HatColor                color.RGBA
	Mustache                string // pencil, handlebar, walrus, ""
	Beard                   bool
	Glasses                 string // round, cateye, ""
	Mole, Earrings, Freckle bool
}

func rgb(s string) color.RGBA {
	s = strings.TrimPrefix(s, "#")
	v := func(i int) uint8 { n, _ := strconv.ParseUint(s[i:i+2], 16, 8); return uint8(n) }
	return color.RGBA{v(0), v(2), v(4), 255}
}

func shade(c color.RGBA, k float64) color.RGBA {
	f := func(v uint8) uint8 { return uint8(min(255, float64(v)*k)) }
	return color.RGBA{f(c.R), f(c.G), f(c.B), 255}
}

func choose[T any](r *rand.Rand, a ...T) T { return a[r.IntN(len(a))] }

var skinTones = []string{"#f6dbc4", "#f2d4b6", "#f1d0b5", "#eec3a2", "#e3b48e", "#d6a27a", "#c98f66", "#a8704a", "#8a5a3a", "#6e452c"}

func (f Face) rng(salt uint64) *rand.Rand { return rand.New(rand.NewPCG(f.Seed, salt)) }

// look builds a regular's appearance from their face. Ladies and gentlemen
// draw from different wardrobes, so a face always matches its name.
func (f Face) look() Look {
	r := f.rng(0xca7e)
	l := Look{Skin: rgb(choose(r, skinTones...)), Stern: r.IntN(10) < 3}
	l.Smile = !l.Stern && r.IntN(10) < 6
	l.Freckle = r.IntN(10) < 2
	grey := r.IntN(10) < 3
	if f.Female {
		l.Lady = true
		if grey {
			l.Hair = rgb(choose(r, "#b8b6bc", "#d6d2cc", "#9c9aa2", "#e8e4de"))
			l.Brow = shade(l.Hair, .7)
		} else {
			l.Hair = rgb(choose(r, "#2a1e18", "#4a2e1c", "#6b3f22", "#8a5a30", "#b5651d", "#d8b46a", "#e6c88a", "#1c1414", "#a8461e"))
		}
		l.HairStyle = choose(r, "bob", "bob", "bun", "waves", "curls")
		l.Lip = rgb(choose(r, "#b3283a", "#c0392b", "#9c2a40", "#c25a5a", "#a0404a"))
		l.Top = choose(r, "stripes", "blouse", "blouse", "turtleneck", "cardigan")
		l.Coat = rgb(choose(r, "#2a3a5c", "#6e1f2a", "#2f5a4a", "#c9a24c", "#4b2a5c", "#1e1e24", "#b3473a", "#3a6a8a", "#e8dcc8", "#7a4a6a"))
		l.Stripe = rgb(choose(r, "#1f2f55", "#1f2f55", "#a3242b"))
		l.Accent = rgb(choose(r, "#f4ecd8", "#e8b4b8", "#d9b44a", "#ffffff", "#b3283a"))
		switch n := r.IntN(10); {
		case n < 3:
			l.Neck = "scarf"
			l.NeckColor = rgb(choose(r, "#b3283a", "#e8b04a", "#3a7a9a", "#7a3a7a", "#2f7a5a", "#e07a5f"))
		case n < 6:
			l.Neck = "pearls"
		}
		switch n := r.IntN(100); {
		case n < 30:
			l.Hat = "beret"
			l.HatColor = rgb(choose(r, "#1e1e24", "#b3283a", "#2a3a5c", "#6e1f2a", "#c9a24c"))
		case n < 52:
			l.Hat = "cloche"
			l.HatColor = rgb(choose(r, "#6e1f2a", "#2a3a5c", "#3a2a2a", "#c9a27a", "#4b2a5c", "#2f5a4a"))
		}
		l.Earrings = r.IntN(10) < 5
		l.Mole = r.IntN(10) < 2
		if r.IntN(100) < 18 {
			l.Glasses = "cateye"
		}
		if grey && r.IntN(3) == 0 {
			l.Glasses = "round"
		}
		if l.Hat != "" && l.HairStyle == "bun" { // a hat would hide the bun: let the hair show
			l.HairStyle = choose(r, "bob", "waves", "curls")
		}
	} else {
		if grey {
			l.Hair = rgb(choose(r, "#e8e4dc", "#c9c4bc", "#a8a4a0"))
			l.Brow = shade(l.Hair, .75)
		} else {
			l.Hair = rgb(choose(r, "#2a1e18", "#4a2e1c", "#6b3f22", "#8a5a30", "#b5651d", "#c8a060", "#1c1414"))
		}
		l.HairStyle = choose(r, "short", "side", "side", "curly")
		if grey && r.IntN(2) == 0 {
			l.HairStyle = "bald"
		}
		l.Top = choose(r, "stripes", "suit", "suit", "turtleneck", "cardigan", "chef")
		l.Coat = rgb(choose(r, "#2a3a5c", "#3b2a1e", "#2f4a3a", "#4a2a2a", "#2a2a3a", "#5a4a3a", "#6b3a28", "#3a3a4a", "#7a6a50"))
		l.Stripe = rgb(choose(r, "#1f2f55", "#1f2f55", "#a3242b"))
		l.Accent = rgb(choose(r, "#f4ecd8", "#ffffff", "#e8e0c8"))
		switch l.Top {
		case "suit":
			l.Neck = choose(r, "tie", "bowtie", "bowtie")
			l.NeckColor = rgb(choose(r, "#b3283a", "#1e1e24", "#2a3a5c", "#c9a24c", "#2f5a4a", "#6e1f2a"))
		case "chef":
			l.Neck = "kerchief"
			l.NeckColor = rgb(choose(r, "#b3283a", "#2a3a5c"))
			l.Hat = "toque"
		case "stripes", "turtleneck":
			if r.IntN(3) == 0 {
				l.Neck = "scarf"
				l.NeckColor = rgb(choose(r, "#b3283a", "#e8b04a", "#3a7a9a", "#2a2a3a"))
			}
		}
		if l.Hat == "" {
			switch n := r.IntN(100); {
			case n < 30:
				l.Hat = "beret"
				l.HatColor = rgb(choose(r, "#1e1e24", "#2a3a5c", "#6e1f2a", "#3a3a3a"))
			case n < 45:
				l.Hat = "cap"
				l.HatColor = rgb(choose(r, "#5a4a3a", "#3a3a4a", "#6b5a3a", "#4a3a2a"))
			case n < 58:
				l.Hat = "fedora"
				l.HatColor = rgb(choose(r, "#3a3a3a", "#5a4a3a", "#2a2a3a", "#7a6a50"))
			}
		}
		switch n := r.IntN(100); {
		case n < 22:
			l.Mustache = "pencil"
		case n < 40:
			l.Mustache = "handlebar"
		case n < 52:
			l.Mustache = "walrus"
		}
		l.Beard = r.IntN(100) < 18
		if r.IntN(100) < 24 {
			l.Glasses = "round"
		}
	}
	l.Dark = shade(l.Coat, .62)
	return l
}

/* ---------------- names and manners ---------------- */

// manner is the kind of passenger a face suggests; it picks the title and the turns of phrase.
func (f Face) manner() string {
	l := f.look()
	grey := l.Brow != (color.RGBA{}) // only silver hair sets the brow colour
	if f.Female {
		switch {
		case grey:
			return "grandma"
		case l.Hat == "beret" || l.Top == "stripes":
			return "starlet"
		case l.Neck == "pearls" || l.Hat == "cloche":
			return "society"
		}
		return "miss"
	}
	switch {
	case l.Hat == "toque":
		return "chef"
	case grey:
		return "grandpa"
	case l.Glasses == "round":
		return "professor"
	case l.Top == "stripes":
		return "captain"
	case l.Hat == "beret":
		return "bandleader"
	}
	return "mister"
}

// Every passenger is called by title and name ("Miss Betty", "Mr. Dawson").
var mannerNames = map[string]struct {
	titles []string
	names  []string
}{
	"grandma":    {[]string{"Grandma", "Aunt"}, []string{"Edna", "Mabel", "Ethel", "Hazel", "Myrtle", "Opal", "Agnes", "Bess"}},
	"starlet":    {[]string{"Miss"}, []string{"Vivian", "Rosalind", "Lorraine", "Dolores", "Arlene", "Maxine", "Gloria", "Veronica"}},
	"society":    {[]string{"Mrs."}, []string{"Whitmore", "Vanderhoff", "Ashby", "Pembrook", "Hollis", "Carrington", "Fairbanks", "Delancey"}},
	"miss":       {[]string{"Miss"}, []string{"Betty", "Dottie", "June", "Peggy", "Ruth", "Lucille", "Irene", "Doris"}},
	"chef":       {[]string{"Chef"}, []string{"Gus", "Sal", "Earl", "Rudy", "Louis", "Otis", "Marty"}},
	"grandpa":    {[]string{"Grandpa", "Pop"}, []string{"Walt", "Hank", "Clem", "Amos", "Floyd", "Virgil", "Orville"}},
	"professor":  {[]string{"Professor", "Doc"}, []string{"Abernathy", "Lindqvist", "Hargrove", "Bellweather", "Kowalski", "Moreau", "Pruitt"}},
	"captain":    {[]string{"Captain"}, []string{"Ned", "Gil", "Ray", "Hal", "Mack", "Lyle"}},
	"bandleader": {[]string{"Maestro"}, []string{"Lou", "Benny", "Vince", "Teddy", "Sid", "Frankie", "Cal"}},
	"mister":     {[]string{"Mr."}, []string{"Dawson", "Kelly", "Brooks", "Fletcher", "Harlan", "Novak", "Reyes", "Sullivan"}},
}

func (f Face) mannerKey() string { return f.manner() }

// nameFor gives the face a name with its title, avoiding names already at the table.
func (f Face) nameFor(avoid map[string]bool) (name, full string, ok bool) {
	mn := mannerNames[f.mannerKey()]
	r := f.rng(0x9a3e5)
	names := append([]string(nil), mn.names...)
	r.Shuffle(len(names), func(i, j int) { names[i], names[j] = names[j], names[i] })
	t := choose(r, mn.titles...)
	for _, n := range names {
		full := t + " " + n
		if !avoid[full] && !avoid[n] {
			return full, full, true
		}
	}
	return "", "", false
}

// style is how this passenger plays: some knock the moment they can, some
// hold out for gin.
func (f Face) style() gn.Style {
	switch f.mannerKey() {
	case "professor", "society":
		return gn.GinHunt
	case "captain", "miss", "bandleader":
		return gn.Eager
	}
	return gn.Style(f.Seed % 3)
}

/* ---------------- table talk ---------------- */

// {p}: whoever the line is about.
var mannerLines = map[string]map[string][]string{
	"grandma": {
		"knock":      {"Knock, knock, dearie.", "I'll just knock, if you don't mind.", "That'll do me. Knock."},
		"gin":        {"Gin! Oh, I do love this game.", "Gin, sweetheart!", "Well, would you look at that. Gin!"},
		"undercut":   {"Oh dear, did I undercut you, {p}?", "Not so fast, honey!", "Grandma still has a trick or two."},
		"undercutBy": {"Oh, fiddlesticks.", "Well, I never!", "Bless your heart, {p}."},
		"take":       {"Oh, thank you, dear.", "I'll take that, thank you.", "Just what I needed."},
		"pass":       {"No thank you, dear.", "Not for me."},
		"caught":     {"Oh, my stars.", "Heavens, look at all that."},
		"dead":       {"Well, that was a long ride."},
		"win":        {"What a lovely game, dear!"}, "lose": {"Well played, sweetheart. Cookie?"}},
	"starlet": {
		"knock":      {"Knock, knock, darling.", "I'll knock. Ready for my close-up.", "And... knock."},
		"gin":        {"Gin! Isn't that divine?", "Gin, darling! Print it!", "That's a wrap. Gin!"},
		"undercut":   {"Oh, did I steal your scene, {p}?", "Undercut! How dramatic.", "Not tonight, darling."},
		"undercutBy": {"Cut! Cut!", "Well, that wasn't in the script.", "Darling, how rude."},
		"take":       {"Thank you, darling.", "Don't mind if I do.", "Oh, I'll take that."},
		"pass":       {"Pass, darling."},
		"caught":     {"Oh, I look dreadful.", "Ugh, not my best angle."},
		"dead":       {"Next take, everyone."},
		"win":        {"Thank you, thank you all!"}, "lose": {"Rematch, darling. I insist."}},
	"society": {
		"knock":      {"I believe I'll knock.", "Knock. Do keep up.", "A modest knock, I think."},
		"gin":        {"Gin. Naturally.", "Gin, if you please.", "Oh, how very gin."},
		"undercut":   {"Undercut, I'm afraid, {p}.", "How unfortunate for you.", "Mm. Not quite, {p}."},
		"undercutBy": {"Well. How vulgar.", "Hmph. Beginner's luck.", "I beg your pardon?"},
		"take":       {"I'll have that, thank you.", "Most kind.", "Thank you, {p}."},
		"pass":       {"I'll pass, thank you."},
		"caught":     {"How tiresome.", "This simply won't do."},
		"dead":       {"How dreary."},
		"win":        {"A pleasant diversion."}, "lose": {"Well played. Tea, anyone?"}},
	"miss": {
		"knock":      {"Knock knock!", "I'm knocking! Ha!", "Knock! Quick, count 'em!"},
		"gin":        {"Gin! Gin! Gin!", "Holy cow, gin!", "Gin! Did you see that?"},
		"undercut":   {"Gotcha, {p}! Undercut!", "Ha! Undercut!", "Too slow, {p}!"},
		"undercutBy": {"Aw, nuts!", "Rats!", "Oh, applesauce."},
		"take":       {"Ooh, thanks!", "Swell, I'll take it!", "Gimme!"},
		"pass":       {"Nah, pass!"},
		"caught":     {"Aw, shucks.", "Oh no, look at all those!"},
		"dead":       {"Aw, nobody wins?"},
		"win":        {"I won! Again, again!"}, "lose": {"Rats! One more?"}},
	"chef": {
		"knock":      {"Knock! Order's up.", "Knock. Soup's on!", "That's done. Knock."},
		"gin":        {"Gin! Cooked to perfection!", "Gin! Chef's special!", "Ding! Gin is served!"},
		"undercut":   {"Undercut! Back in the oven, {p}!", "Ha! Not done yet, {p}!", "Too soon out the oven!"},
		"undercutBy": {"Burnt it!", "Aw, the soufflé fell.", "Overcooked."},
		"take":       {"Thanks, I'll cook with that.", "Ooh, fresh ingredients.", "Into the pot!"},
		"pass":       {"Not on my menu."},
		"caught":     {"What a mess of a kitchen.", "Too many cooks."},
		"dead":       {"Kitchen's closed. Again!"},
		"win":        {"Pie for everybody!"}, "lose": {"Next game, dinner's on me."}},
	"grandpa": {
		"knock":      {"Knock. Done.", "I'll knock, sonny.", "Knock. Count 'em up."},
		"gin":        {"Gin! Still got it!", "Gin, by golly!", "Ha! Gin, like in '29."},
		"undercut":   {"Undercut you, {p}. Heh.", "Not so fast, youngster.", "Heh. Undercut."},
		"undercutBy": {"Dagnabbit!", "Well, I'll be.", "Consarn it."},
		"take":       {"I'll take that.", "Heh. Thank you kindly.", "That'll do nicely."},
		"pass":       {"Pass."},
		"caught":     {"Well, shucks.", "Too much deadwood. Like my woodshed."},
		"dead":       {"Long line, short game."},
		"win":        {"Not bad for an old timer!"}, "lose": {"Good game, kid. Good game."}},
	"professor": {
		"knock":      {"I shall knock.", "Knock. The numbers favour me.", "A calculated knock."},
		"gin":        {"Gin. As the theory predicted.", "Eureka! Gin!", "Gin. Quod erat demonstrandum."},
		"undercut":   {"Undercut, {p}. Elementary.", "Your deadwood exceeds mine.", "A miscalculation, {p}."},
		"undercutBy": {"Fascinating. And dreadful.", "Hm. An anomaly.", "Statistically improbable!"},
		"take":       {"Most useful, thank you.", "Ah, the missing variable.", "Thank you, {p}."},
		"pass":       {"I'll pass."},
		"caught":     {"Hm. Suboptimal.", "A poor sample."},
		"dead":       {"An inconclusive result."},
		"win":        {"A tidy proof."}, "lose": {"Well played. I shall review my notes."}},
	"captain": {
		"knock":      {"Knock! All hands!", "Knocking, matey.", "Knock! Drop anchor."},
		"gin":        {"Gin! Full speed ahead!", "Gin! Hoist the colours!", "Gin! Smooth sailing!"},
		"undercut":   {"Undercut! Man overboard, {p}!", "Ha! Sunk you, {p}!", "Not this voyage!"},
		"undercutBy": {"Blast and barnacles!", "Torpedoed!", "Abandon ship!"},
		"take":       {"Aye, I'll take that.", "Thank ye, {p}.", "Haul it aboard!"},
		"pass":       {"Pass, matey."},
		"caught":     {"Rough seas.", "Taking on water."},
		"dead":       {"Becalmed."},
		"win":        {"Land ho! Victory!"}, "lose": {"We'll sail again, {p}."}},
	"bandleader": {
		"knock":      {"Knock! Hit it!", "Knock, two, three, four!", "Knock. Take it from the top."},
		"gin":        {"Gin! Now that's swing!", "Gin! Ba-dum-tss!", "Gin, baby, gin!"},
		"undercut":   {"Undercut! Off the beat, {p}!", "Ha! You missed your cue!", "Undercut, cat!"},
		"undercutBy": {"Sour note.", "Aw, flat!", "That's a wrong chord."},
		"take":       {"Groovy, thanks.", "That's my jam.", "Dig it. Thanks, {p}."},
		"pass":       {"I'll sit this one out."},
		"caught":     {"Way off key.", "Ouch. Out of tune."},
		"dead":       {"That song went on forever."},
		"win":        {"And that's the big finish!"}, "lose": {"Great set, {p}."}},
	"mister": {
		"knock":      {"I'll knock.", "Knock. Let's see 'em.", "Knock. Count 'em up."},
		"gin":        {"Gin!", "Well, how about that. Gin!", "Gin! Read 'em and weep."},
		"undercut":   {"Undercut, {p}. Tough break.", "Gotcha. Undercut.", "Not quite, pal."},
		"undercutBy": {"Aw, for Pete's sake.", "Well, that's swell. Just swell.", "Nuts."},
		"take":       {"Thanks, pal.", "I'll take that.", "Much obliged, {p}."},
		"pass":       {"Pass."},
		"caught":     {"Yikes.", "Caught with my hand full."},
		"dead":       {"Nobody wins that one."},
		"win":        {"Good game, pal!"}, "lose": {"Nice playing, {p}."}},
}

func (f Face) lines(kind string) []string { return mannerLines[f.mannerKey()][kind] }

/* ---------------- picking passengers ---------------- */

// folk is one computer player picked for a game.
type folk struct{ face, name, full string }

// randomFolks makes up n passengers with different names.
func randomFolks(rng *rand.Rand, n int, avoidFaces, avoidNames map[string]bool) []folk {
	var out []folk
	for tries := 0; len(out) < n && tries < 500; tries++ {
		f := Face{Female: rng.IntN(2) == 0, Seed: rng.Uint64()}
		if avoidFaces[f.String()] {
			continue
		}
		name, full, ok := f.nameFor(avoidNames)
		if !ok {
			continue
		}
		avoidNames[name] = true
		out = append(out, folk{face: f.String(), name: name, full: full})
	}
	return out
}

// pickerFaces are the faces offered for your own player online and on a shared computer.
var pickerFaces = func() []string {
	var out []string
	r := rand.New(rand.NewPCG(2026, 11))
	for len(out) < 10 {
		f := Face{Female: len(out)%2 == 0, Seed: r.Uint64()}
		if f.mannerKey() == "chef" && len(out) > 6 {
			continue
		}
		out = append(out, f.String())
	}
	return out
}()
