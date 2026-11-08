'use strict';
/* =====================================================================
   GIN RUMMY - the club-car music and sound effects, synthesized with WebAudio.
   Original compositions (c) Ironite:
     "Starlight Swing"      - a swing tune for clarinet, piano, bass and brushes
     "Prairie Moon"         - a harmonica waltz over a guitar, for the night prairie
     "Boogie on the Rails"  - boogie-woogie piano over a train-rhythm shuffle
   bake.html renders each with an OfflineAudioContext into the game's files.
   ===================================================================== */
const Sound = (() => {
  const mtof = m => 440 * Math.pow(2, (m - 69) / 12);
  let noiseBuf = null;
  function noise(ctx) {
    if (noiseBuf && noiseBuf.sampleRate === ctx.sampleRate) return noiseBuf;
    const n = ctx.sampleRate * 2, b = ctx.createBuffer(1, n, ctx.sampleRate), d = b.getChannelData(0);
    let s = 12345; for (let i = 0; i < n; i++) { s = (s * 1103515245 + 12345) & 0x7fffffff; d[i] = s / 0x3fffffff - 1; }
    return (noiseBuf = b);
  }
  function env(g, t, a, d, sus, r, peak, len) {   // ADSR on a gain param
    g.setValueAtTime(0, t); g.linearRampToValueAtTime(peak, t + a);
    g.linearRampToValueAtTime(peak * sus, t + a + d);
    g.setValueAtTime(peak * sus, Math.max(t + a + d, t + len));
    g.linearRampToValueAtTime(0, Math.max(t + a + d, t + len) + r);
  }
  function reverb(ctx, secs = 1.8, wet = 0.22) {
    const len = Math.floor(ctx.sampleRate * secs), ir = ctx.createBuffer(2, len, ctx.sampleRate);
    for (let ch = 0; ch < 2; ch++) { const d = ir.getChannelData(ch); let s = 99 + ch; for (let i = 0; i < len; i++) { s = (s * 1103515245 + 12345) & 0x7fffffff; d[i] = (s / 0x3fffffff - 1) * Math.pow(1 - i / len, 3); } }
    const conv = ctx.createConvolver(); conv.buffer = ir;
    const inp = ctx.createGain(), dry = ctx.createGain(), w = ctx.createGain();
    dry.gain.value = 1; w.gain.value = wet;
    inp.connect(dry); inp.connect(conv); conv.connect(w);
    const out = ctx.createGain(); dry.connect(out); w.connect(out);
    return { inp, out };
  }

  /* ---------------- instruments ---------------- */
  // the musette accordion: three reeds, two of them tuned a little apart so they shimmer
  function accordion(ctx, dest, t, m, dur, vel, bright = 1) {
    const out = ctx.createGain(); env(out.gain, t, 0.025, 0.08, 0.85, 0.07, vel * 0.2, dur);
    const lp = ctx.createBiquadFilter(); lp.type = 'lowpass'; lp.frequency.value = 2400 * bright; lp.Q.value = 0.7;
    const trem = ctx.createGain(); trem.gain.value = 1;
    const lfo = ctx.createOscillator(); lfo.frequency.value = 5.2; const lg = ctx.createGain(); lg.gain.value = 0.08; lfo.connect(lg); lg.connect(trem.gain);
    for (const [det, typ, g] of [[-9, 'sawtooth', 0.5], [9, 'sawtooth', 0.5], [0, 'square', 0.25]]) {
      const o = ctx.createOscillator(); o.type = typ; o.frequency.value = mtof(m); o.detune.value = det;
      const og = ctx.createGain(); og.gain.value = g; o.connect(og); og.connect(lp);
      o.start(t); o.stop(t + dur + 0.2);
    }
    lp.connect(trem); trem.connect(out); out.connect(dest);
    lfo.start(t); lfo.stop(t + dur + 0.2);
  }
  // a plucked string (guitar, double bass): a bright saw that darkens as it rings
  function pluck(ctx, dest, t, m, dur, vel, bright = 3200, decay = 0.6) {
    const o = ctx.createOscillator(); o.type = 'sawtooth'; o.frequency.value = mtof(m);
    const o2 = ctx.createOscillator(); o2.type = 'triangle'; o2.frequency.value = mtof(m);
    const lp = ctx.createBiquadFilter(); lp.type = 'lowpass'; lp.Q.value = 1;
    lp.frequency.setValueAtTime(bright, t); lp.frequency.exponentialRampToValueAtTime(Math.max(200, bright * 0.18), t + decay);
    const g = ctx.createGain(); g.gain.setValueAtTime(0, t); g.gain.linearRampToValueAtTime(vel * 0.22, t + 0.004);
    g.gain.exponentialRampToValueAtTime(0.0005, t + Math.max(dur, 0.05) + decay);
    const g2 = ctx.createGain(); g2.gain.value = 0.6; o2.connect(g2); g2.connect(lp);
    o.connect(lp); lp.connect(g); g.connect(dest);
    o.start(t); o.stop(t + dur + decay + 0.05); o2.start(t); o2.stop(t + dur + decay + 0.05);
  }
  function bass(ctx, dest, t, m, dur, vel) {
    const o = ctx.createOscillator(); o.type = 'triangle'; o.frequency.value = mtof(m);
    const s = ctx.createOscillator(); s.type = 'sine'; s.frequency.value = mtof(m);
    const g = ctx.createGain(); g.gain.setValueAtTime(0, t); g.gain.linearRampToValueAtTime(vel * 0.42, t + 0.01);
    g.gain.exponentialRampToValueAtTime(vel * 0.12, t + dur * 0.6); g.gain.linearRampToValueAtTime(0, t + dur + 0.05);
    const lp = ctx.createBiquadFilter(); lp.type = 'lowpass'; lp.frequency.value = 900;
    o.connect(lp); s.connect(lp); lp.connect(g); g.connect(dest);
    o.start(t); o.stop(t + dur + 0.1); s.start(t); s.stop(t + dur + 0.1);
  }
  // a violin: a bowed saw with vibrato that blooms in
  function violin(ctx, dest, t, m, dur, vel) {
    const o = ctx.createOscillator(); o.type = 'sawtooth'; o.frequency.value = mtof(m);
    const vib = ctx.createOscillator(); vib.frequency.value = 5.6; const vg = ctx.createGain();
    vg.gain.setValueAtTime(0, t); vg.gain.linearRampToValueAtTime(16, t + Math.min(0.35, dur));
    vib.connect(vg); vg.connect(o.detune);
    const bp = ctx.createBiquadFilter(); bp.type = 'lowpass'; bp.frequency.value = 3600; bp.Q.value = 0.8;
    const pk = ctx.createBiquadFilter(); pk.type = 'peaking'; pk.frequency.value = 2600; pk.gain.value = 6; pk.Q.value = 1.4;
    const g = ctx.createGain(); env(g.gain, t, 0.05, 0.1, 0.8, 0.12, vel * 0.16, dur);
    o.connect(bp); bp.connect(pk); pk.connect(g); g.connect(dest);
    o.start(t); o.stop(t + dur + 0.3); vib.start(t); vib.stop(t + dur + 0.3);
  }
  // a piano: a few decaying partials
  function piano(ctx, dest, t, m, dur, vel) {
    const f = mtof(m), ring = Math.min(3.2, 0.6 + dur * 1.2) * (m < 60 ? 1.3 : 1);
    const out = ctx.createGain(); out.gain.value = vel * 0.2; out.connect(dest);
    for (const [k, a, det] of [[1, 0.62, 0], [1, 0.3, 4], [2, 0.22, 0], [3, 0.1, 0], [4, 0.05, 0], [5.02, 0.03, 0]]) {
      const o = ctx.createOscillator(); o.type = 'sine'; o.frequency.value = f * k; o.detune.value = det;
      const g = ctx.createGain(); g.gain.setValueAtTime(0, t); g.gain.linearRampToValueAtTime(a, t + 0.006);
      g.gain.exponentialRampToValueAtTime(0.0008, t + ring / Math.sqrt(k));
      o.connect(g); g.connect(out); o.start(t); o.stop(t + ring + 0.05);
    }
  }
  function pad(ctx, dest, t, m, dur, vel) {
    for (const det of [-7, 7]) {
      const o = ctx.createOscillator(); o.type = 'sawtooth'; o.frequency.value = mtof(m); o.detune.value = det;
      const lp = ctx.createBiquadFilter(); lp.type = 'lowpass'; lp.frequency.value = 900;
      const g = ctx.createGain(); env(g.gain, t, 0.5, 0.3, 0.8, 0.8, vel * 0.04, dur);
      o.connect(lp); lp.connect(g); g.connect(dest); o.start(t); o.stop(t + dur + 1);
    }
  }
  function brush(ctx, dest, t, vel) {
    const s = ctx.createBufferSource(); s.buffer = noise(ctx);
    const hp = ctx.createBiquadFilter(); hp.type = 'bandpass'; hp.frequency.value = 5000; hp.Q.value = 0.6;
    const g = ctx.createGain(); g.gain.setValueAtTime(vel * 0.12, t); g.gain.exponentialRampToValueAtTime(0.001, t + 0.14);
    s.connect(hp); hp.connect(g); g.connect(dest); s.start(t, Math.random() * 1.5); s.stop(t + 0.16);
  }

  // a clarinet: a hollow, square-ish reed with a little breath and vibrato
  function clarinet(ctx, dest, t, m, dur, vel) {
    const o = ctx.createOscillator(); o.type = 'square'; o.frequency.value = mtof(m);
    const o2 = ctx.createOscillator(); o2.type = 'triangle'; o2.frequency.value = mtof(m);
    const vib = ctx.createOscillator(); vib.frequency.value = 5; const vg = ctx.createGain();
    vg.gain.setValueAtTime(0, t); vg.gain.linearRampToValueAtTime(10, t + Math.min(0.4, dur)); vib.connect(vg); vg.connect(o.detune); vg.connect(o2.detune);
    const lp = ctx.createBiquadFilter(); lp.type = 'lowpass'; lp.Q.value = 0.9;
    lp.frequency.setValueAtTime(1400, t); lp.frequency.linearRampToValueAtTime(2600, t + 0.06); lp.frequency.linearRampToValueAtTime(2000, t + dur);
    const g = ctx.createGain(); env(g.gain, t, 0.03, 0.08, 0.82, 0.08, vel * 0.13, dur);
    const g2 = ctx.createGain(); g2.gain.value = 0.8; o2.connect(g2); g2.connect(lp);
    o.connect(lp); lp.connect(g); g.connect(dest);
    for (const x of [o, o2, vib]) { x.start(t); x.stop(t + dur + 0.2); }
  }
  // a harmonica: two slightly beating reeds through a nasal band
  function harmonica(ctx, dest, t, m, dur, vel) {
    const out = ctx.createGain(); env(out.gain, t, 0.05, 0.1, 0.8, 0.1, vel * 0.16, dur);
    const bp = ctx.createBiquadFilter(); bp.type = 'bandpass'; bp.frequency.value = 1500; bp.Q.value = 0.7;
    const pk = ctx.createBiquadFilter(); pk.type = 'peaking'; pk.frequency.value = 2800; pk.gain.value = 5; pk.Q.value = 2;
    const trem = ctx.createGain(); trem.gain.value = 1;
    const lfo = ctx.createOscillator(); lfo.frequency.value = 6.2; const lg = ctx.createGain();
    lg.gain.setValueAtTime(0, t); lg.gain.linearRampToValueAtTime(0.16, t + Math.min(0.5, dur)); lfo.connect(lg); lg.connect(trem.gain);
    for (const [det, typ, gg] of [[-6, 'sawtooth', 0.5], [6, 'square', 0.35]]) {
      const o = ctx.createOscillator(); o.type = typ; o.frequency.value = mtof(m); o.detune.value = det;
      o.detune.setValueAtTime(det - 40, t); o.detune.linearRampToValueAtTime(det, t + 0.06);   // a little bend into the note
      const og = ctx.createGain(); og.gain.value = gg; o.connect(og); og.connect(bp); o.start(t); o.stop(t + dur + 0.2);
    }
    bp.connect(pk); pk.connect(trem); trem.connect(out); out.connect(dest);
    lfo.start(t); lfo.stop(t + dur + 0.2);
  }
  // a hi-hat / ride tick
  function tick(ctx, dest, t, vel) {
    const s = ctx.createBufferSource(); s.buffer = noise(ctx);
    const hp = ctx.createBiquadFilter(); hp.type = 'highpass'; hp.frequency.value = 7000;
    const g = ctx.createGain(); g.gain.setValueAtTime(vel * 0.07, t); g.gain.exponentialRampToValueAtTime(0.001, t + 0.07);
    s.connect(hp); hp.connect(g); g.connect(dest); s.start(t, (t * 3.1) % 1.5); s.stop(t + 0.08);
  }
  // a steam whistle: three breathy sine voices that bend up into a chord
  function whistle(ctx, dest, t, ms, dur, vel = 1) {
    for (const m of ms) {
      const o = ctx.createOscillator(); o.type = 'sine'; o.frequency.setValueAtTime(mtof(m - 1.2), t); o.frequency.exponentialRampToValueAtTime(mtof(m), t + 0.12);
      const o2 = ctx.createOscillator(); o2.type = 'triangle'; o2.frequency.value = mtof(m) * 2.005;
      const g = ctx.createGain(); env(g.gain, t, 0.06, 0.2, 0.75, 0.35, vel * 0.09, dur);
      const g2 = ctx.createGain(); g2.gain.value = 0.18; o2.connect(g2); g2.connect(g);
      o.connect(g); g.connect(dest); o.start(t); o.stop(t + dur + 0.5); o2.start(t); o2.stop(t + dur + 0.5);
    }
    const s = ctx.createBufferSource(); s.buffer = noise(ctx); const bp = ctx.createBiquadFilter(); bp.type = 'bandpass'; bp.frequency.value = 2200; bp.Q.value = 0.8;
    const g = ctx.createGain(); env(g.gain, t, 0.03, 0.1, 0.5, 0.3, vel * 0.06, dur); s.connect(bp); bp.connect(g); g.connect(dest); s.start(t); s.stop(t + dur + 0.4);
  }

  /* ---------------- the tunes ---------------- */
  const CH = {   // chord tones (mid register) and the bass root
    Fmaj7: [[57, 60, 64, 65], 41], F6: [[57, 60, 62, 65], 41], D7: [[54, 57, 60, 62], 50], Gm7: [[55, 58, 62, 65], 43], C7: [[55, 58, 60, 64], 48],
    Am7: [[55, 57, 60, 64], 45], A7: [[55, 57, 61, 64], 45], Dm7: [[57, 60, 62, 65], 50], G7: [[55, 59, 62, 65], 43],
    G: [[55, 59, 62], 43], C: [[55, 60, 64], 48], D: [[54, 57, 62], 50], Em: [[55, 59, 64], 40], F7: [[57, 60, 63, 65], 41]
  };
  const n = (s) => s.split(' ').filter(x => x && x !== '|').map(x => { const [m, d] = x.split(':'); return [+m, +d]; });
  const isMinor = name => /m(?!aj)/.test(name);

  // 1. Starlight Swing: 4/4 swing in F, AABA, clarinet over piano, walking bass and brushes
  const SWING = {
    name: 'Starlight Swing', bpm: 160, beats: 4, swing: true, end: 'F6',
    A: { ch: 'Fmaj7 D7 Gm7 C7 Am7 D7 Gm7 C7'.split(' '),
      mel: n('72:1.5 69:.5 72:1 77:1 | 74:1.5 72:.5 74:1 78:1 | 77:2 74:1 70:1 | 72:1 70:.5 69:.5 67:2 | 69:1.5 72:.5 76:1 79:1 | 78:2 74:1 72:1 | 70:1 74:.5 77:.5 79:1 77:1 | 76:3 0:1') },
    B: { ch: 'A7 A7 Dm7 Dm7 G7 G7 Gm7 C7'.split(' '),
      mel: n('73:1 76:1 79:1 81:1 | 79:2 76:2 | 77:1 74:.5 77:.5 81:2 | 79:1 77:1 74:2 | 71:1 74:1 77:1 79:1 | 77:2 74:2 | 70:1 74:.5 77:.5 79:1 77:1 | 76:2 72:1 70:1') },
    form: ['A', 'A', 'B', 'A']
  };
  // 2. Prairie Moon: a slow waltz in G for harmonica and guitar
  const PRAIRIE = {
    name: 'Prairie Moon', bpm: 96, beats: 3, end: 'G',
    A: { ch: 'G G C G D D7 G G'.split(' '),
      mel: n('71:1 74:1 79:1 | 78:1.5 76:.5 74:1 | 76:2 72:1 | 74:3 | 69:1 72:1 78:1 | 76:1.5 74:.5 72:1 | 71:2 69:1 | 67:3') },
    B: { ch: 'C C G G A7 A7 D7 D7'.split(' '),
      mel: n('72:1 76:1 79:1 | 81:2 79:1 | 79:1 76:1 74:1 | 71:3 | 73:1 76:1 79:1 | 78:1.5 76:.5 73:1 | 74:1 78:1 81:1 | 78:2 0:1') },
    form: ['A', 'B', 'A']
  };
  // 3. Boogie on the Rails: a 12-bar boogie-woogie in C over a chugging shuffle
  const BOOGIE = {
    name: 'Boogie on the Rails', bpm: 150, beats: 4, swing: true, boogie: true, end: 'C7',
    A: { ch: 'C7 C7 C7 C7 F7 F7 C7 C7 G7 F7 C7 G7'.split(' ') },
    form: ['A', 'A', 'A']
  };
  const TUNES = [SWING, PRAIRIE, BOOGIE];

  // buildQueue turns a tune into timed note events.
  function buildQueue(tune) {
    const beat = 60 / tune.bpm, ev = [];
    const sw = x => tune.swing && Math.abs(x % 1 - 0.5) < 1e-6 ? x + 0.16 : x;
    let t = 0.1;
    const roots = [];
    tune.form.forEach(f => tune[f].ch.forEach(c => roots.push(CH[c][1])));
    roots.push(CH[tune.end][1]);
    let barNo = 0;
    for (let pass = 0; pass < tune.form.length; pass++) {
      const sec = tune[tune.form[pass]];
      // melody
      if (sec.mel) {
        let mt = 0;
        for (const [m, d] of sec.mel) {
          if (m > 0) {
            const st = sw(mt), en = sw(mt + d);
            ev.push({ i: tune.swing ? 'clar' : 'harp', t: t + st * beat, m, d: Math.max(0.08, (en - st) * beat * 0.92), v: 0.85 });
          }
          mt += d;
        }
      }
      sec.ch.forEach((name, k) => {
        const [tones, root] = CH[name], b0 = t + k * tune.beats * beat, next = roots[barNo + k + 1];
        if (tune.boogie) {
          // left hand: the walking boogie figure in swung eighths
          const fig = [0, 4, 7, 9, 10, 9, 7, 4];
          const r = root >= 45 ? root - 12 : root;
          fig.forEach((o, j) => ev.push({ i: 'piano', t: b0 + sw(j * 0.5) * beat, m: r + o, d: beat * 0.42, v: j % 2 ? 0.45 : 0.62 }));
          // right hand: stabs in the first chorus, licks after
          const up = tones.map(x => x + 12), shift = { 48: 0, 41: 5, 43: -5 }[root] || 0;
          if (pass === 0 || k % 2 === 1) {
            for (const at of [1.5, 3]) for (const m of up) ev.push({ i: 'piano', t: b0 + sw(at) * beat, m, d: beat * 0.3, v: 0.4 });
          } else {
            const lick = pass === 1 ? [[79, 0], [82, .5], [79, 1], [76, 1.5], [75, 2], [72, 2.5]] : [[84, 0], [87, .5], [84, 1], [82, 1.5], [79, 2], [76, 2.5], [75, 3], [72, 3.5]];
            for (const [m, at] of lick) ev.push({ i: 'piano', t: b0 + sw(at) * beat, m: m + shift, d: beat * 0.4, v: 0.55 });
          }
          // the train: brushes on every eighth, accented on the backbeat
          for (let j = 0; j < 8; j++) ev.push({ i: 'brush', t: b0 + sw(j * 0.5) * beat, v: j === 2 || j === 6 ? 1.1 : 0.55 });
          ev.push({ i: 'bass', t: b0, m: r - 12, d: beat * 1.8, v: 0.6 }, { i: 'bass', t: b0 + 2 * beat, m: r - 12 + 7, d: beat * 1.8, v: 0.5 });
        } else if (tune.swing) {
          // walking bass, piano comping on 2 and 4, brushes and a ride
          const third = isMinor(name) ? 3 : 4;
          const walk = [root, root + third, root + 7, next + (next > root ? -1 : 1)];
          walk.forEach((m, j) => ev.push({ i: 'bass', t: b0 + j * beat, m, d: beat * 0.88, v: 0.95 }));
          for (const bb of [1, 3]) {
            tones.forEach((tn, j) => ev.push({ i: 'piano', t: b0 + bb * beat + j * 0.008, m: tn, d: beat * 0.32, v: 0.3 }));
            ev.push({ i: 'brush', t: b0 + bb * beat, v: 1 });
          }
          for (const at of [0, 1, 1.5, 2, 3, 3.5]) ev.push({ i: 'tick', t: b0 + sw(at) * beat, v: at % 1 ? 0.6 : 1 });
        } else {
          // a fingerpicked guitar: the bass on one, chord tones on two and three
          ev.push({ i: 'gtr', t: b0, m: k % 2 ? root + 7 - 12 + 12 : root, d: beat * 0.9, v: 0.7, br: 1800 });
          for (const bb of [1, 2]) tones.forEach((tn, j) => ev.push({ i: 'gtr', t: b0 + bb * beat + j * 0.03, m: tn, d: beat * 0.6, v: 0.32, br: 2600 }));
          ev.push({ i: 'pad', t: b0, m: tones[1], d: tune.beats * beat, v: 0.7 });
        }
      });
      t += sec.ch.length * tune.beats * beat;
      barNo += sec.ch.length;
    }
    // the last chord
    const [tones, root] = CH[tune.end];
    ev.push({ i: 'bass', t, m: root >= 45 ? root - 12 : root, d: beat * 3, v: 1 });
    for (const tn of tones) ev.push({ i: tune.swing ? 'piano' : 'gtr', t, m: tn, d: beat * 3, v: 0.5, br: 2400 });
    if (tune.boogie) whistleAt(ev, t + beat * 0.5);
    else ev.push({ i: tune.swing ? 'clar' : 'harp', t, m: tones[tones.length - 1] + 12, d: beat * 3, v: 0.7 });
    return { ev, len: t + beat * 3 + 1.4 };
  }
  function whistleAt(ev, t) { ev.push({ i: 'whistle', t, ms: [74, 77, 81], d: 1.1 }); }

  function renderTune(ctx, dest, tune) {
    const rv = reverb(ctx, tune.swing ? 1.4 : 2.4, tune.swing ? 0.18 : 0.28); rv.out.connect(dest);
    const bus = ctx.createGain(); bus.gain.value = 0.85; bus.connect(rv.inp);
    const q = buildQueue(tune);
    for (const e of q.ev) {
      if (e.i === 'clar') clarinet(ctx, bus, e.t, e.m, e.d, e.v);
      else if (e.i === 'harp') harmonica(ctx, bus, e.t, e.m, e.d, e.v);
      else if (e.i === 'bass') bass(ctx, bus, e.t, e.m, e.d, e.v);
      else if (e.i === 'gtr') pluck(ctx, bus, e.t, e.m, e.d, e.v, e.br || 3000, 0.5);
      else if (e.i === 'piano') piano(ctx, bus, e.t, e.m, e.d, e.v);
      else if (e.i === 'pad') pad(ctx, bus, e.t, e.m, e.d, e.v);
      else if (e.i === 'brush') brush(ctx, bus, e.t, e.v);
      else if (e.i === 'tick') tick(ctx, bus, e.t, e.v);
      else if (e.i === 'whistle') whistle(ctx, bus, e.t, e.ms, e.d, 0.8);
    }
    return q.len;
  }

  /* ---------------- sound effects ---------------- */
  function flick(ctx, dest, t, vel = 1, f = 2600, len = 0.06) {   // one card snapping through the air
    const s = ctx.createBufferSource(); s.buffer = noise(ctx);
    const bp = ctx.createBiquadFilter(); bp.type = 'bandpass'; bp.frequency.setValueAtTime(f, t); bp.frequency.exponentialRampToValueAtTime(f * 0.5, t + len); bp.Q.value = 1.2;
    const g = ctx.createGain(); g.gain.setValueAtTime(0, t); g.gain.linearRampToValueAtTime(vel * 0.6, t + 0.004); g.gain.exponentialRampToValueAtTime(0.001, t + len);
    s.connect(bp); bp.connect(g); g.connect(dest); s.start(t, (t * 7.3) % 1.5); s.stop(t + len + 0.02);
  }
  function thump(ctx, dest, t, vel = 1) {   // a card landing on the baize
    flick(ctx, dest, t, vel * 0.6, 1400, 0.05);
    const o = ctx.createOscillator(); o.frequency.setValueAtTime(150, t); o.frequency.exponentialRampToValueAtTime(60, t + 0.06);
    const g = ctx.createGain(); g.gain.setValueAtTime(vel * 0.22, t); g.gain.exponentialRampToValueAtTime(0.001, t + 0.08);
    o.connect(g); g.connect(dest); o.start(t); o.stop(t + 0.1);
  }
  function knuckle(ctx, dest, t, vel = 1) {   // a rap on the walnut table
    const o = ctx.createOscillator(); o.type = 'sine'; o.frequency.setValueAtTime(240, t); o.frequency.exponentialRampToValueAtTime(110, t + 0.05);
    const g = ctx.createGain(); g.gain.setValueAtTime(vel * 0.55, t); g.gain.exponentialRampToValueAtTime(0.001, t + 0.12);
    o.connect(g); g.connect(dest); o.start(t); o.stop(t + 0.14);
    const s = ctx.createBufferSource(); s.buffer = noise(ctx); const bp = ctx.createBiquadFilter(); bp.type = 'bandpass'; bp.frequency.value = 900; bp.Q.value = 1.6;
    const sg = ctx.createGain(); sg.gain.setValueAtTime(vel * 0.5, t); sg.gain.exponentialRampToValueAtTime(0.001, t + 0.05); s.connect(bp); bp.connect(sg); sg.connect(dest); s.start(t, (t * 5.7) % 1.5); s.stop(t + 0.06);
  }
  function bell(ctx, dest, t, m, vel = 1, len = 1.2) {   // a brass bell
    for (const [k, a] of [[1, 1], [2.76, 0.35], [5.4, 0.15]]) {
      const o = ctx.createOscillator(); o.type = 'sine'; o.frequency.value = mtof(m) * k;
      const g = ctx.createGain(); g.gain.setValueAtTime(0, t); g.gain.linearRampToValueAtTime(vel * 0.16 * a, t + 0.004); g.gain.exponentialRampToValueAtTime(0.0005, t + len / k);
      o.connect(g); g.connect(dest); o.start(t); o.stop(t + len);
    }
  }
  const fx = {
    shuffle(ctx, d) { let t = 0.02; for (let i = 0; i < 26; i++) { flick(ctx, d, t, 0.45 + 0.2 * Math.sin(i), 3200 + (i % 5) * 200, 0.035); t += 0.018 + (i % 3) * 0.004; } thump(ctx, d, t + 0.05, 0.8); return t + 0.3; },
    deal(ctx, d) { flick(ctx, d, 0.01, 0.8, 3000, 0.07); return 0.15; },
    draw(ctx, d) { flick(ctx, d, 0.01, 0.55, 2200, 0.08); flick(ctx, d, 0.06, 0.35, 3400, 0.05); return 0.2; },
    take(ctx, d) { flick(ctx, d, 0.01, 0.7, 2800, 0.06); bell(ctx, d, 0.04, 91, 0.25, 0.3); return 0.4; },
    discard(ctx, d) { flick(ctx, d, 0.01, 0.5, 2400, 0.05); thump(ctx, d, 0.05, 0.9); return 0.2; },
    hover(ctx, d) { const o = ctx.createOscillator(); o.frequency.value = 2200; const g = ctx.createGain(); g.gain.setValueAtTime(0.03, 0.005); g.gain.exponentialRampToValueAtTime(0.0005, 0.04); o.connect(g); g.connect(d); o.start(0.005); o.stop(0.05); return 0.06; },
    tick(ctx, d) { flick(ctx, d, 0.005, 0.5, 4200, 0.025); const o = ctx.createOscillator(); o.frequency.value = 1500; const g = ctx.createGain(); g.gain.setValueAtTime(0.08, 0.005); g.gain.exponentialRampToValueAtTime(0.0005, 0.05); o.connect(g); g.connect(d); o.start(0.005); o.stop(0.06); return 0.08; },
    slide(ctx, d) { const s = ctx.createBufferSource(); s.buffer = noise(ctx); const bp = ctx.createBiquadFilter(); bp.type = 'bandpass'; bp.frequency.setValueAtTime(1200, 0.01); bp.frequency.linearRampToValueAtTime(2600, 0.2); bp.Q.value = 0.8; const g = ctx.createGain(); g.gain.setValueAtTime(0, 0.01); g.gain.linearRampToValueAtTime(0.25, 0.06); g.gain.exponentialRampToValueAtTime(0.001, 0.24); s.connect(bp); bp.connect(g); g.connect(d); s.start(0.01); s.stop(0.26); bell(ctx, d, 0.18, 86, 0.3, 0.5); return 0.6; },
    sweep(ctx, d) { const s = ctx.createBufferSource(); s.buffer = noise(ctx); const bp = ctx.createBiquadFilter(); bp.type = 'bandpass'; bp.frequency.setValueAtTime(800, 0.01); bp.frequency.exponentialRampToValueAtTime(3000, 0.28); bp.Q.value = 0.7; const g = ctx.createGain(); g.gain.setValueAtTime(0, 0.01); g.gain.linearRampToValueAtTime(0.3, 0.1); g.gain.exponentialRampToValueAtTime(0.001, 0.34); s.connect(bp); bp.connect(g); g.connect(d); s.start(0.01); s.stop(0.36); return 0.4; },
    knock(ctx, d) { knuckle(ctx, d, 0.02, 1); knuckle(ctx, d, 0.2, 0.9); return 0.5; },
    gin(ctx, d) {      // the train bell rings and the band hits a chord
      for (let i = 0; i < 4; i++) bell(ctx, d, 0.02 + i * 0.22, 79, 0.9, 0.8);
      for (const m of [65, 69, 72, 76]) piano(ctx, d, 0.9, m, 0.9, 0.6);
      clarinet(ctx, d, 0.9, 81, 0.9, 0.8);
      return 2.2;
    },
    biggin(ctx, d) {   // a whistle, bells and a run up to the top
      whistle(ctx, d, 0.02, [74, 77, 81], 0.9, 1);
      [65, 69, 72, 77, 81, 84, 89].forEach((m, i) => bell(ctx, d, 1.0 + i * 0.07, m, 0.8, 1.4));
      for (const m of [65, 69, 72, 76]) piano(ctx, d, 1.5, m, 1.2, 0.6);
      clarinet(ctx, d, 1.5, 84, 1.1, 0.85);
      return 3.2;
    },
    dead(ctx, d) {     // a long, low whistle fading into the night
      whistle(ctx, d, 0.02, [62, 65, 69], 0.5, 0.8); whistle(ctx, d, 0.75, [58, 62, 65], 1.2, 0.7);
      return 2.6;
    },
    undercut(ctx, d) { // a sly clarinet slide and a thump on the table
      [[76, .12], [74, .12], [72, .12], [79, .5]].reduce((t, [m, dur]) => { clarinet(ctx, d, t, m, dur * 0.95, 0.8); return t + dur; }, 0.02);
      knuckle(ctx, d, 0.9, 0.8); for (const m of [57, 60, 64]) piano(ctx, d, 0.9, m, 0.6, 0.5);
      return 1.8;
    },
    point(ctx, d) { bell(ctx, d, 0.01, 79, 0.7, 1); bell(ctx, d, 0.14, 84, 0.7, 1.3); return 1.5; },
    win(ctx, d) {      // a little big-band fanfare, and a toot from the engine
      const seq = [[72, .14], [76, .14], [79, .14], [84, .5], [83, .14], [84, .7]];
      let t = 0.02; for (const [m, dur] of seq) { clarinet(ctx, d, t, m, dur * 0.95, 0.85); t += dur; }
      for (const m of [48, 60, 64, 67, 72]) piano(ctx, d, t - 0.7, m, 0.9, 0.5);
      whistle(ctx, d, t - 0.1, [74, 77, 81], 0.7, 0.7);
      return t + 1.4;
    },
    lose(ctx, d) {
      const seq = [[67, .3], [65, .3], [64, .3], [60, .9]];
      let t = 0.02; for (const [m, dur] of seq) { clarinet(ctx, d, t, m, dur * 0.95, 0.6); t += dur; }
      for (const m of [48, 55, 60]) piano(ctx, d, t - 0.9, m, 0.9, 0.4);
      return t + 0.8;
    }
  };

  return { fx, TUNES, renderTune, buildQueue, reverb };
})();
