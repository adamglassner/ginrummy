'use strict';
/* =====================================================================
   GIN RUMMY - A Club Car Card Game (aboard the Starlight Limited)
   All card art and props are drawn here in code (original artwork, (c) Ironite).
   ===================================================================== */
const Art = (() => {
  const CW = 360, CH = 504;              // card canvas size (HD; the game scales it down)
  const INK = '#24170f';
  const RED = '#a01c30', BLACK = '#18141a', GOLD = '#b8913f', GOLD2 = '#e8cf8a';
  const IVORY1 = '#fdf9ee', IVORY2 = '#f1e6cc';
  const SERIF = '"Lora", Georgia, serif';
  const DISP = '"Playfair Display", Georgia, serif';
  const TAU = Math.PI * 2;

  // Suit palettes for the court figures: Pullman red, brass, midnight blue, bottle green.
  const PAL = {
    H: { coat: '#9c1f33', dark: '#5e1220', accent: '#e8c35a', ink: RED, wall: '#f3dcd2' },
    D: { coat: '#c7822a', dark: '#7e4a14', accent: '#f6efe0', ink: RED, wall: '#f4e4c4' },
    S: { coat: '#24385c', dark: '#121e36', accent: '#e8c35a', ink: BLACK, wall: '#dfe2e8' },
    C: { coat: '#2f5a44', dark: '#183426', accent: '#e8c35a', ink: BLACK, wall: '#dfe8dc' }
  };

  /* ---------- small helpers ---------- */
  function rng(seed) { let s = seed >>> 0 || 1; return () => (s = (s * 1664525 + 1013904223) >>> 0) / 4294967296; }
  function rrect(c, x, y, w, h, r) { c.beginPath(); c.moveTo(x + r, y); c.arcTo(x + w, y, x + w, y + h, r); c.arcTo(x + w, y + h, x, y + h, r); c.arcTo(x, y + h, x, y, r); c.arcTo(x, y, x + w, y, r); c.closePath(); }
  function canvas(w, h) { const c = document.createElement('canvas'); c.width = w; c.height = h; return c; }
  const ink = c => { c.strokeStyle = INK; c.lineWidth = 2.6; c.lineJoin = 'round'; c.lineCap = 'round'; };

  /* ---------- suit shapes (drawn, so they look the same on every PC) ---------- */
  function heartPath(c, x, y, s) {
    c.moveTo(x, y + s * 0.45);
    c.bezierCurveTo(x - s * 0.62, y + s * 0.02, x - s * 0.52, y - s * 0.55, x, y - s * 0.2);
    c.bezierCurveTo(x + s * 0.52, y - s * 0.55, x + s * 0.62, y + s * 0.02, x, y + s * 0.45);
  }
  function suitPath(c, s, x, y, size) {
    if (s === 'H') heartPath(c, x, y, size);
    else if (s === 'D') { c.moveTo(x, y - size * 0.5); c.quadraticCurveTo(x + size * 0.18, y - size * 0.18, x + size * 0.4, y); c.quadraticCurveTo(x + size * 0.18, y + size * 0.18, x, y + size * 0.5); c.quadraticCurveTo(x - size * 0.18, y + size * 0.18, x - size * 0.4, y); c.quadraticCurveTo(x - size * 0.18, y - size * 0.18, x, y - size * 0.5); }
    else if (s === 'S') {
      c.save(); c.translate(x, y - size * 0.08); c.rotate(Math.PI); heartPath(c, 0, 0, size * 0.9); c.restore();
      c.moveTo(x, y + size * 0.05); c.quadraticCurveTo(x - size * 0.05, y + size * 0.38, x - size * 0.24, y + size * 0.5); c.lineTo(x + size * 0.24, y + size * 0.5); c.quadraticCurveTo(x + size * 0.05, y + size * 0.38, x, y + size * 0.05);
    } else {
      const r = size * 0.205;
      c.moveTo(x + r, y - size * 0.24); c.arc(x, y - size * 0.24, r, 0, TAU); c.moveTo(x - size * 0.22 + r, y + size * 0.06); c.arc(x - size * 0.22, y + size * 0.06, r, 0, TAU);
      c.moveTo(x + size * 0.22 + r, y + size * 0.06); c.arc(x + size * 0.22, y + size * 0.06, r, 0, TAU);
      c.moveTo(x + size * 0.12, y); c.arc(x, y, size * 0.12, 0, TAU);
      c.moveTo(x, y + size * 0.02); c.quadraticCurveTo(x - size * 0.04, y + size * 0.38, x - size * 0.22, y + size * 0.5); c.lineTo(x + size * 0.22, y + size * 0.5); c.quadraticCurveTo(x + size * 0.04, y + size * 0.38, x, y + size * 0.02);
    }
  }
  function suit(c, s, x, y, size, color, shine = true) {
    c.save(); c.fillStyle = color; c.beginPath(); suitPath(c, s, x, y, size); c.fill();
    if (shine && size > 30) { // a little lacquer highlight
      c.clip(); c.fillStyle = 'rgba(255,255,255,.16)'; c.beginPath(); c.ellipse(x - size * 0.16, y - size * 0.16, size * 0.12, size * 0.2, -0.6, 0, TAU); c.fill();
    }
    c.restore();
  }

  /* ---------- paper ---------- */
  function paper(c, w, h, seed, r = 22) {
    const g = c.createLinearGradient(0, 0, w, h); g.addColorStop(0, IVORY1); g.addColorStop(1, IVORY2);
    rrect(c, 1, 1, w - 2, h - 2, r); c.fillStyle = g; c.fill();
    c.save(); c.clip();
    const R = rng(seed);
    for (let i = 0; i < 900; i++) { c.fillStyle = `rgba(120,90,40,${R() * 0.045})`; c.fillRect(R() * w, R() * h, 1 + R() * 2, 1 + R() * 2); }
    const v = c.createRadialGradient(w / 2, h / 2, h * 0.3, w / 2, h / 2, h * 0.75); v.addColorStop(0, 'rgba(0,0,0,0)'); v.addColorStop(1, 'rgba(120,80,30,.14)');
    c.fillStyle = v; c.fillRect(0, 0, w, h); c.restore();
    rrect(c, 1, 1, w - 2, h - 2, r); c.strokeStyle = 'rgba(70,50,30,.45)'; c.lineWidth = 2; c.stroke();
  }
  // a streamline border: a thin double line with curls and a little star in each corner
  function border(c, w, h, color) {
    c.save(); c.strokeStyle = color; c.globalAlpha = 0.85;
    c.lineWidth = 2; rrect(c, 14, 14, w - 28, h - 28, 14); c.stroke();
    c.lineWidth = 0.8; rrect(c, 19, 19, w - 38, h - 38, 10); c.stroke();
    for (const [x, y, sx, sy] of [[19, 19, 1, 1], [w - 19, 19, -1, 1], [19, h - 19, 1, -1], [w - 19, h - 19, -1, -1]]) {
      c.lineWidth = 1.4;
      c.beginPath(); c.moveTo(x + sx * 4, y + sy * 46); c.bezierCurveTo(x + sx * 4, y + sy * 14, x + sx * 14, y + sy * 4, x + sx * 46, y + sy * 4); c.stroke();
      c.beginPath(); c.moveTo(x + sx * 10, y + sy * 26); c.quadraticCurveTo(x + sx * 10, y + sy * 10, x + sx * 26, y + sy * 10); c.stroke();
      star(c, x + sx * 16, y + sy * 16, 6, color);
    }
    c.restore();
  }
  function index(c, rank, s, color, w, h) {
    const draw = () => {
      c.fillStyle = color; c.textAlign = 'center'; c.textBaseline = 'alphabetic';
      c.font = `700 ${rank === '10' ? 44 : 52}px ${DISP}`;
      c.fillText(rank, 46, 74);
      suit(c, s, 46, 102, 32, color, false);
    };
    draw(); c.save(); c.translate(w, h); c.rotate(Math.PI); draw(); c.restore();
  }

  /* ---------- the club-car people (court figures) ----------
     Drawn around (0,0) = bottom centre of the bust. Head is ~54px wide at scale 1. */
  function shoulderX(y) {
    if (y <= -64) return 22;
    const bx = t => { const u = 1 - t; return [u * u * u * 22 + 3 * u * u * t * 54 + 3 * u * t * t * 78 + t * t * t * 80, u * u * u * -64 + 3 * u * u * t * -60 + 3 * u * t * t * -42]; };
    let lo = 0, hi = 1; for (let i = 0; i < 30; i++) { const m = (lo + hi) / 2; if (bx(m)[1] < y) lo = m; else hi = m; } return bx(lo)[0];
  }
  function shoulders(c) { c.beginPath(); c.moveTo(-80, 0); c.bezierCurveTo(-78, -42, -54, -60, -22, -64); c.lineTo(22, -64); c.bezierCurveTo(54, -60, 78, -42, 80, 0); c.closePath(); }
  function bust(c, o) {
    const skin = o.skin || '#f2d4b6', hair = o.hair || '#4a2e1c', hy = -118;
    c.save(); ink(c);
    c.fillStyle = hair;
    if (o.hairStyle === 'bun') { c.beginPath(); c.arc(0, hy - 36, 15, 0, TAU); c.fill(); c.stroke(); }
    if (o.hairStyle === 'waves') { c.beginPath(); c.moveTo(-30, hy - 10); c.bezierCurveTo(-44, hy + 20, -40, hy + 52, -26, hy + 60); c.lineTo(26, hy + 60); c.bezierCurveTo(40, hy + 52, 44, hy + 20, 30, hy - 10); c.closePath(); c.fill(); c.stroke(); }
    if (o.hairStyle === 'bob') { c.beginPath(); c.moveTo(-31, hy - 8); c.bezierCurveTo(-38, hy + 14, -36, hy + 28, -30, hy + 32); c.lineTo(30, hy + 32); c.bezierCurveTo(36, hy + 28, 38, hy + 14, 31, hy - 8); c.closePath(); c.fill(); c.stroke(); }
    // the top
    const top = (o.top === 'stripes' || o.top === 'chef' || o.top === 'baker') ? '#fbf8f0' : o.coat;
    shoulders(c); c.fillStyle = top; c.fill();
    if (o.top === 'stripes') { c.fillStyle = o.stripe || '#1f2f55'; for (let y = -58; y < 2; y += 13) { const y0 = y, y1 = Math.min(y + 6, 0), x0 = shoulderX(y0), x1 = shoulderX(y1); c.beginPath(); c.moveTo(-x0, y0); c.lineTo(x0, y0); c.lineTo(x1, y1); c.lineTo(-x1, y1); c.closePath(); c.fill(); } }
    if (o.top === 'dress') { c.save(); shoulders(c); c.clip(); c.fillStyle = 'rgba(255,255,255,.12)'; for (let x = -80; x < 80; x += 14) { c.beginPath(); c.moveTo(x, 0); c.quadraticCurveTo(x + 6, -40, x + 2, -70); c.lineTo(x + 6, -70); c.quadraticCurveTo(x + 10, -40, x + 4, 0); c.fill(); } c.restore(); }
    shoulders(c); c.stroke();
    // neck
    c.fillStyle = skin; c.beginPath(); c.moveTo(-11, -90); c.lineTo(-12, -61); c.lineTo(12, -61); c.lineTo(11, -90); c.closePath(); c.fill();
    c.fillStyle = 'rgba(120,60,40,.18)'; c.beginPath(); c.moveTo(-11, -88); c.quadraticCurveTo(0, -80, 11, -88); c.lineTo(11, -84); c.quadraticCurveTo(0, -76, -11, -84); c.closePath(); c.fill();
    if (o.top === 'dress') { // a sweetheart neckline
      c.fillStyle = skin; c.beginPath(); c.moveTo(-26, -64); c.quadraticCurveTo(-14, -42, 0, -50); c.quadraticCurveTo(14, -42, 26, -64); c.closePath(); c.fill();
      c.beginPath(); c.moveTo(-26, -64); c.quadraticCurveTo(-14, -42, 0, -50); c.quadraticCurveTo(14, -42, 26, -64); c.stroke();
    }
    if (o.top === 'stripes') { c.beginPath(); c.moveTo(-23, -64); c.quadraticCurveTo(0, -55, 23, -64); c.stroke(); }
    if (o.top === 'blouse') { c.fillStyle = o.accent; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 2, -62); c.lineTo(sx * 20, -66); c.lineTo(sx * 16, -50); c.closePath(); c.fill(); c.stroke(); } }
    if (o.top === 'turtleneck') { c.fillStyle = o.coat; rrect(c, -15, -82, 30, 22, 6); c.fill(); c.stroke(); c.strokeStyle = o.dark; c.lineWidth = 1.2; for (let x = -10; x <= 10; x += 5) { c.beginPath(); c.moveTo(x, -80); c.lineTo(x, -62); c.stroke(); } ink(c); }
    if (o.top === 'suit' || o.top === 'waistcoat') {
      c.fillStyle = '#fbf8f0'; c.beginPath(); c.moveTo(-18, -64); c.lineTo(0, -14); c.lineTo(18, -64); c.closePath(); c.fill(); c.stroke();
      if (o.top === 'waistcoat') { c.fillStyle = o.vest || '#1e1e24'; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 10, -48); c.lineTo(sx * 2, -8); c.lineTo(sx * 2, 0); c.lineTo(sx * 40, 0); c.lineTo(sx * 34, -50); c.closePath(); c.fill(); c.stroke(); } c.fillStyle = GOLD; for (const y of [-30, -18, -6]) { c.beginPath(); c.arc(6, y, 2.4, 0, TAU); c.fill(); } }
      c.fillStyle = o.dark; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 20, -64); c.lineTo(sx * (o.top === 'waistcoat' ? 34 : 3), -14); c.lineTo(sx * 46, -34); c.lineTo(sx * 30, -54); c.closePath(); c.fill(); c.stroke(); }
      c.fillStyle = '#fbf8f0'; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 1, -62); c.lineTo(sx * 13, -66); c.lineTo(sx * 10, -54); c.closePath(); c.fill(); c.stroke(); }
    }
    if (o.top === 'cardigan') {
      c.fillStyle = o.accent; c.beginPath(); c.moveTo(-18, -64); c.lineTo(0, -26); c.lineTo(18, -64); c.closePath(); c.fill(); c.stroke();
      c.fillStyle = o.dark; for (const y of [-18, -6]) { c.beginPath(); c.arc(0, y, 3, 0, TAU); c.fill(); }
    }
    if (o.top === 'chef' || o.top === 'baker') {
      c.fillStyle = '#d8d4cc'; for (const y of [-46, -30, -14]) for (const x of [-14, 14]) { c.beginPath(); c.arc(x, y, 3.2, 0, TAU); c.fill(); c.stroke(); }
      if (o.top === 'baker') { c.fillStyle = '#e9dcc0'; c.beginPath(); c.moveTo(-34, 0); c.lineTo(-30, -38); c.lineTo(30, -38); c.lineTo(34, 0); c.closePath(); c.fill(); c.stroke(); c.fillStyle = 'rgba(255,255,255,.7)'; c.beginPath(); c.arc(-14, -24, 4, 0, TAU); c.arc(10, -14, 3, 0, TAU); c.fill(); }
    }
    // neckwear
    if (o.neck === 'scarf') {
      c.fillStyle = o.neckColor; c.beginPath(); c.moveTo(-17, -78); c.quadraticCurveTo(0, -66, 17, -78); c.lineTo(19, -66); c.quadraticCurveTo(0, -54, -19, -66); c.closePath(); c.fill(); c.stroke();
      c.beginPath(); c.moveTo(8, -64); c.quadraticCurveTo(20, -46, 14, -26); c.lineTo(24, -30); c.quadraticCurveTo(26, -48, 14, -62); c.closePath(); c.fill(); c.stroke();
      c.beginPath(); c.ellipse(9, -63, 6, 5, 0, 0, TAU); c.fill(); c.stroke();
    }
    if (o.neck === 'pearls') { c.fillStyle = '#fffaf0'; c.lineWidth = 1.2; for (let i = 0; i < 11; i++) { const a = Math.PI * (0.12 + i * 0.076); c.beginPath(); c.arc(Math.cos(a) * 22, -70 + Math.sin(a) * 16, 3, 0, TAU); c.fill(); c.stroke(); } ink(c); }
    if (o.neck === 'bowtie') { c.fillStyle = o.neckColor; c.beginPath(); c.moveTo(0, -60); c.lineTo(-13, -67); c.lineTo(-13, -52); c.closePath(); c.moveTo(0, -60); c.lineTo(13, -67); c.lineTo(13, -52); c.closePath(); c.fill(); c.stroke(); rrect(c, -3.5, -63.5, 7, 7, 2); c.fill(); c.stroke(); }
    if (o.neck === 'kerchief') { c.fillStyle = o.neckColor; c.beginPath(); c.moveTo(-16, -66); c.quadraticCurveTo(0, -58, 16, -66); c.lineTo(0, -44); c.closePath(); c.fill(); c.stroke(); }
    if (o.neck === 'choker') { c.fillStyle = '#18141a'; c.fillRect(-11, -84, 22, 5); c.fillStyle = '#9a7ab8'; c.beginPath(); c.arc(0, -79, 3.5, 0, TAU); c.fill(); c.stroke(); }
    // head
    c.fillStyle = skin; c.beginPath(); c.ellipse(-27, hy + 2, 5, 8, 0, 0, TAU); c.ellipse(27, hy + 2, 5, 8, 0, 0, TAU); c.fill(); c.stroke();
    if (o.earrings) { c.fillStyle = o.earrings; for (const sx of [-1, 1]) { c.beginPath(); c.arc(sx * 28, hy + 13, 3.4, 0, TAU); c.fill(); } }
    c.fillStyle = skin; c.beginPath(); c.ellipse(0, hy, 27, 33, 0, 0, TAU); c.fill(); c.stroke();
    // front hair
    c.fillStyle = hair;
    if (o.hairStyle === 'short') { c.beginPath(); c.ellipse(0, hy - 14, 28.5, 22, 0, Math.PI, 0); c.quadraticCurveTo(14, hy - 22, -4, hy - 18); c.quadraticCurveTo(-18, hy - 14, -28.5, hy - 6); c.closePath(); c.fill(); c.stroke(); }
    if (o.hairStyle === 'side') { c.beginPath(); c.moveTo(-28.5, hy - 2); c.bezierCurveTo(-34, hy - 36, -6, hy - 44, 6, hy - 38); c.bezierCurveTo(24, hy - 40, 34, hy - 24, 28.5, hy - 4); c.quadraticCurveTo(24, hy - 22, 8, hy - 24); c.quadraticCurveTo(-2, hy - 26, -10, hy - 20); c.quadraticCurveTo(-20, hy - 14, -28.5, hy - 2); c.closePath(); c.fill(); c.stroke(); c.strokeStyle = 'rgba(255,255,255,.22)'; c.lineWidth = 1.4; c.beginPath(); c.moveTo(-12, hy - 32); c.quadraticCurveTo(4, hy - 38, 18, hy - 32); c.stroke(); ink(c); }
    if (o.hairStyle === 'bob') { c.beginPath(); c.moveTo(-29, hy + 14); c.bezierCurveTo(-36, hy - 34, 36, hy - 34, 29, hy + 14); c.lineTo(26, hy - 6); c.quadraticCurveTo(18, hy - 20, 4, hy - 16); c.quadraticCurveTo(-10, hy - 20, -20, hy - 14); c.quadraticCurveTo(-26, hy - 8, -29, hy + 14); c.closePath(); c.fill(); c.stroke(); }
    if (o.hairStyle === 'bun' || o.hairStyle === 'waves') { c.beginPath(); c.moveTo(-29, hy + 6); c.bezierCurveTo(-34, hy - 30, -14, hy - 40, 0, hy - 36); c.bezierCurveTo(14, hy - 40, 34, hy - 30, 29, hy + 6); c.bezierCurveTo(22, hy - 10, 14, hy - 24, 0, hy - 22); c.bezierCurveTo(-14, hy - 24, -22, hy - 10, -29, hy + 6); c.closePath(); c.fill(); c.stroke(); }
    if (o.beard) { c.fillStyle = o.beardColor || hair; c.beginPath(); c.moveTo(-25, hy + 4); c.quadraticCurveTo(-24, hy + 42, 0, hy + 48); c.quadraticCurveTo(24, hy + 42, 25, hy + 4); c.quadraticCurveTo(14, hy + 22, 0, hy + 20); c.quadraticCurveTo(-14, hy + 22, -25, hy + 4); c.fill(); c.stroke(); }
    // face
    c.fillStyle = 'rgba(214,110,100,.28)'; c.beginPath(); c.arc(-15, hy + 9, 6.5, 0, TAU); c.arc(15, hy + 9, 6.5, 0, TAU); c.fill();
    c.fillStyle = INK; const ey = o.lady ? 3.4 : 3;
    c.beginPath(); c.ellipse(-10, hy - 3, 2.6, ey, 0, 0, TAU); c.ellipse(10, hy - 3, 2.6, ey, 0, 0, TAU); c.fill();
    if (o.lady) { c.lineWidth = 1.4; for (const sx of [-1, 1]) { c.beginPath(); c.moveTo(sx * 12.5, hy - 5); c.lineTo(sx * 15.5, hy - 7); c.stroke(); } }
    c.lineWidth = 2.2; c.strokeStyle = o.brow || hair; const st = o.stern ? 2 : 0, st1 = o.stern ? 0 : 1;
    c.beginPath(); c.moveTo(-15, hy - 11 + st); c.lineTo(-5, hy - 12 - st1); c.moveTo(15, hy - 11 + st); c.lineTo(5, hy - 12 - st1); c.stroke();
    c.strokeStyle = INK; c.lineWidth = 1.8; c.beginPath(); c.moveTo(1, hy); c.quadraticCurveTo(-3, hy + 9, 1, hy + 11); c.stroke();
    if (o.mole) { c.fillStyle = INK; c.beginPath(); c.arc(-13, hy + 16, 1.5, 0, TAU); c.fill(); }
    if (o.lady) { c.fillStyle = o.lip || '#b3283a'; c.beginPath(); if (o.smile) { c.moveTo(-7, hy + 18); c.quadraticCurveTo(0, hy + 26, 7, hy + 18); c.quadraticCurveTo(0, hy + 21, -7, hy + 18); } else c.ellipse(0, hy + 20, 6, o.stern ? 1.8 : 2.8, 0, 0, TAU); c.fill(); }
    else if (!o.mustache || o.mustache === 'pencil') { c.beginPath(); if (o.smile) { c.moveTo(-7, hy + 18); c.quadraticCurveTo(0, hy + 24, 7, hy + 18); } else { c.moveTo(-6, hy + 20); c.quadraticCurveTo(0, hy + 22, 6, hy + 20); } c.stroke(); }
    const mc = o.mustacheColor || o.brow || hair;
    if (o.mustache === 'pencil') { c.strokeStyle = mc; c.lineWidth = 2.2; c.beginPath(); c.moveTo(-10, hy + 14); c.quadraticCurveTo(0, hy + 11, 10, hy + 14); c.stroke(); }
    if (o.mustache === 'handlebar') { c.fillStyle = mc; c.strokeStyle = INK; c.lineWidth = 1.6; c.beginPath(); c.moveTo(0, hy + 13); c.bezierCurveTo(-8, hy + 10, -18, hy + 18, -24, hy + 10); c.bezierCurveTo(-28, hy + 6, -26, hy + 2, -22, hy + 4); c.bezierCurveTo(-22, hy + 10, -16, hy + 20, 0, hy + 17); c.bezierCurveTo(16, hy + 20, 22, hy + 10, 22, hy + 4); c.bezierCurveTo(26, hy + 2, 28, hy + 6, 24, hy + 10); c.bezierCurveTo(18, hy + 18, 8, hy + 10, 0, hy + 13); c.fill(); c.stroke(); }
    if (o.mustache === 'walrus') { c.fillStyle = mc; c.strokeStyle = INK; c.lineWidth = 1.6; c.beginPath(); c.moveTo(0, hy + 11); c.bezierCurveTo(-10, hy + 8, -18, hy + 12, -18, hy + 22); c.quadraticCurveTo(-8, hy + 18, 0, hy + 20); c.quadraticCurveTo(8, hy + 18, 18, hy + 22); c.bezierCurveTo(18, hy + 12, 10, hy + 8, 0, hy + 11); c.fill(); c.stroke(); }
    ink(c);
    if (o.glasses === 'round') { c.strokeStyle = '#3a2a1e'; c.lineWidth = 2; c.beginPath(); c.arc(-10, hy - 3, 7.5, 0, TAU); c.moveTo(17.5, hy - 3); c.arc(10, hy - 3, 7.5, 0, TAU); c.moveTo(-2.5, hy - 4); c.lineTo(2.5, hy - 4); c.stroke(); ink(c); }
    // hats
    if (o.hat === 'beret') { c.save(); c.translate(4, hy - 30); c.rotate(-0.14); c.fillStyle = o.hatColor; c.beginPath(); c.ellipse(0, 0, 36, 13, 0, 0, TAU); c.fill(); c.stroke(); c.beginPath(); c.moveTo(-2, -12); c.lineTo(0, -20); c.lineTo(3, -12); c.closePath(); c.fill(); c.stroke(); c.restore(); }
    if (o.hat === 'cloche') {
      c.fillStyle = o.hatColor; c.beginPath(); c.moveTo(-33, hy - 4); c.bezierCurveTo(-36, hy - 50, 36, hy - 50, 33, hy - 4); c.quadraticCurveTo(0, hy - 16, -33, hy - 4); c.closePath(); c.fill(); c.stroke();
      c.strokeStyle = o.band || 'rgba(255,255,255,.4)'; c.lineWidth = 4; c.beginPath(); c.moveTo(-31, hy - 14); c.quadraticCurveTo(0, hy - 26, 31, hy - 14); c.stroke(); ink(c);
      if (o.flowers) { for (const [x, y, col] of [[20, hy - 22, '#e8b4b8'], [27, hy - 17, '#f0d36a'], [21, hy - 13, '#e8b4b8']]) { c.fillStyle = col; c.beginPath(); c.arc(x, y, 4.5, 0, TAU); c.fill(); c.stroke(); } }
      if (o.veil) { c.save(); c.globalAlpha = 0.55; c.strokeStyle = '#18141a'; c.lineWidth = 0.9; c.beginPath(); c.moveTo(-31, hy - 10); c.quadraticCurveTo(0, hy + 14, 31, hy - 10); c.stroke(); for (let x = -28; x <= 28; x += 5) { c.beginPath(); c.moveTo(x, hy - 12); c.lineTo(x * 0.8, hy + 6 - Math.abs(x) * 0.3); c.stroke(); } for (let y = hy - 8; y < hy + 6; y += 5) { c.beginPath(); c.moveTo(-28 + (y - hy) * 0.3, y); c.lineTo(28 - (y - hy) * 0.3, y); c.stroke(); } c.restore(); ink(c); }
    }
    if (o.hat === 'cap') { c.fillStyle = o.hatColor; c.beginPath(); c.moveTo(-30, hy - 14); c.bezierCurveTo(-30, hy - 42, 26, hy - 46, 32, hy - 22); c.quadraticCurveTo(4, hy - 26, -30, hy - 14); c.closePath(); c.fill(); c.stroke(); c.beginPath(); c.moveTo(4, hy - 24); c.quadraticCurveTo(34, hy - 28, 40, hy - 16); c.quadraticCurveTo(24, hy - 12, 4, hy - 18); c.closePath(); c.fill(); c.stroke(); }
    if (o.hat === 'toque' || o.hat === 'bakercap') {
      c.fillStyle = '#fdfcf8'; rrect(c, -24, hy - 46, 48, 16, 3); c.fill(); c.stroke();
      const puffs = o.hat === 'toque' ? [[-18, hy - 60, 15], [18, hy - 60, 15], [0, hy - 70, 18]] : [[-10, hy - 50, 12], [10, hy - 50, 12]];
      for (const [x, y, r] of puffs) { c.beginPath(); c.arc(x, y, r, 0, TAU); c.fill(); c.stroke(); }
      c.fillRect(-22, o.hat === 'toque' ? hy - 58 : hy - 50, 44, o.hat === 'toque' ? 14 : 6);
    }
    if (o.hat === 'feathers') {   // a cabaret headband with a plume
      c.strokeStyle = o.band || '#e8c35a'; c.lineWidth = 4; c.beginPath(); c.moveTo(-29, hy - 14); c.quadraticCurveTo(0, hy - 30, 29, hy - 14); c.stroke(); ink(c);
      c.fillStyle = '#e8c35a'; c.beginPath(); c.arc(16, hy - 22, 5, 0, TAU); c.fill(); c.stroke();
      c.fillStyle = o.plume || '#f6efe2'; c.beginPath(); c.moveTo(16, hy - 26); c.bezierCurveTo(24, hy - 70, 52, hy - 74, 48, hy - 50); c.bezierCurveTo(40, hy - 58, 30, hy - 44, 16, hy - 26); c.fill(); c.stroke();
    }
    // railroad hats
    if (o.hat === 'conductor' || o.hat === 'porter') {   // a pillbox cap with a visor and a badge
      const top = o.hat === 'porter' ? hy - 50 : hy - 52;
      c.fillStyle = o.hatColor; c.beginPath(); c.moveTo(-29, hy - 16); c.lineTo(-27, top); c.quadraticCurveTo(0, top - 8, 27, top); c.lineTo(29, hy - 16); c.quadraticCurveTo(0, hy - 22, -29, hy - 16); c.closePath(); c.fill(); c.stroke();
      c.fillStyle = '#18141a'; c.beginPath(); c.moveTo(-30, hy - 16); c.quadraticCurveTo(0, hy - 24, 30, hy - 16); c.quadraticCurveTo(26, hy - 6, 0, hy - 8); c.quadraticCurveTo(-26, hy - 6, -30, hy - 16); c.closePath(); c.fill(); c.stroke();
      c.strokeStyle = GOLD; c.lineWidth = 2; c.beginPath(); c.moveTo(-27, hy - 22); c.quadraticCurveTo(0, hy - 28, 27, hy - 22); c.stroke(); ink(c);
      c.fillStyle = GOLD2; rrect(c, -9, top + 6, 18, 12, 2); c.fill(); c.lineWidth = 1.4; c.stroke();
      c.fillStyle = INK; c.font = `700 7px ${DISP}`; c.textAlign = 'center'; c.fillText(o.hat === 'porter' ? 'SL' : '★', 0, top + 15); ink(c);
    }
    if (o.hat === 'engcap') {   // the engineer's hickory-striped cap
      c.save(); c.fillStyle = '#e8ecf2'; c.beginPath(); c.moveTo(-31, hy - 12); c.bezierCurveTo(-34, hy - 50, 32, hy - 54, 32, hy - 14); c.quadraticCurveTo(0, hy - 22, -31, hy - 12); c.closePath(); c.fill();
      c.clip(); c.strokeStyle = '#2a3a6a'; c.lineWidth = 2.6; for (let x = -34; x < 36; x += 6) { c.beginPath(); c.moveTo(x, hy - 60); c.lineTo(x, hy - 8); c.stroke(); } c.restore();
      ink(c); c.beginPath(); c.moveTo(-31, hy - 12); c.bezierCurveTo(-34, hy - 50, 32, hy - 54, 32, hy - 14); c.quadraticCurveTo(0, hy - 22, -31, hy - 12); c.closePath(); c.stroke();
      c.fillStyle = '#d8dce4'; c.beginPath(); c.moveTo(-6, hy - 18); c.quadraticCurveTo(26, hy - 22, 42, hy - 12); c.quadraticCurveTo(20, hy - 8, -6, hy - 12); c.closePath(); c.fill(); c.stroke();
    }
    if (o.hat === 'stetson') {   // a wide-brimmed cattleman's hat
      c.fillStyle = o.hatColor; c.beginPath(); c.ellipse(0, hy - 22, 50, 10, 0, 0, TAU); c.fill(); c.stroke();
      c.beginPath(); c.moveTo(-24, hy - 24); c.bezierCurveTo(-26, hy - 66, -6, hy - 60, 0, hy - 54); c.bezierCurveTo(6, hy - 60, 26, hy - 66, 24, hy - 24); c.closePath(); c.fill(); c.stroke();
      c.strokeStyle = '#2a1a10'; c.lineWidth = 4; c.beginPath(); c.moveTo(-24, hy - 30); c.quadraticCurveTo(0, hy - 34, 24, hy - 30); c.stroke(); ink(c);
    }
    if (o.hat === 'fedora') {
      c.fillStyle = o.hatColor; c.beginPath(); c.ellipse(0, hy - 24, 40, 9, -0.05, 0, TAU); c.fill(); c.stroke();
      c.beginPath(); c.moveTo(-26, hy - 26); c.bezierCurveTo(-28, hy - 62, 28, hy - 62, 26, hy - 26); c.closePath(); c.fill(); c.stroke();
      c.beginPath(); c.moveTo(-10, hy - 54); c.quadraticCurveTo(0, hy - 46, 10, hy - 54); c.stroke();
      c.strokeStyle = '#18141a'; c.lineWidth = 5; c.beginPath(); c.moveTo(-26, hy - 32); c.quadraticCurveTo(0, hy - 36, 26, hy - 32); c.stroke(); ink(c);
    }
    if (o.neck === 'tie') { c.fillStyle = o.neckColor; c.beginPath(); c.moveTo(-4, -62); c.lineTo(4, -62); c.lineTo(6, -24); c.lineTo(0, -16); c.lineTo(-6, -24); c.closePath(); c.fill(); c.stroke(); }
    if (o.buttons) { c.fillStyle = GOLD2; for (const y of [-48, -32, -16]) for (const x of [-26, 26]) { c.beginPath(); c.arc(x, y, 3, 0, TAU); c.fill(); c.lineWidth = 1.2; c.stroke(); } ink(c); }
    c.restore();
  }
  function hand(c, x, y, skin) { c.save(); ink(c); c.fillStyle = skin || '#f2d4b6'; c.beginPath(); c.ellipse(x, y, 9, 8, 0, 0, TAU); c.fill(); c.stroke(); c.restore(); }

  // who sits on each court card: the crew and passengers of the Starlight Limited
  const COURT = {
    K: {
      H: { name: 'The Dining-Car Chef', o: { top: 'chef', hat: 'toque', neck: 'kerchief', neckColor: '#b3283a', hair: '#8a8a90', mustache: 'handlebar', mustacheColor: '#d8d4cc', brow: '#c8c4c0', skin: '#eec3a2', smile: true }, prop: 'whisk' },
      D: { name: 'The Conductor', o: { top: 'suit', coat: '#1f2a4a', dark: '#121a30', neck: 'tie', neckColor: '#18141a', hat: 'conductor', hatColor: '#1f2a4a', hair: '#c8c4c0', mustache: 'walrus', mustacheColor: '#e0dcd6', brow: '#c8c4c0', skin: '#f1d0b5', buttons: true, stern: true }, prop: 'watch' },
      S: { name: 'The Engineer', o: { top: 'stripes', stripe: '#2a3a6a', hat: 'engcap', neck: 'kerchief', neckColor: '#c0392b', hair: '#5a3a22', mustache: 'handlebar', mustacheColor: '#4a2e1c', skin: '#c98f66', smile: true }, prop: 'oilcan' },
      C: { name: 'The Cattle Baron', o: { top: 'suit', coat: '#6b4a2a', dark: '#3e2a16', neck: 'bowtie', neckColor: '#18141a', hat: 'stetson', hatColor: '#c9a27a', hair: '#8a8a90', beard: true, beardColor: '#d8d4cc', brow: '#c8c4c0', skin: '#e3b48e' }, prop: 'deed' }
    },
    Q: {
      H: { name: 'The Torch Singer', o: { lady: true, top: 'dress', coat: '#b3283a', dark: '#6e1220', hairStyle: 'waves', hair: '#2a1e18', hat: 'feathers', band: '#e8c35a', neck: 'pearls', lip: '#c0182a', skin: '#a8704a', earrings: '#e8c35a', smile: true, mole: true }, prop: 'mic' },
      D: { name: 'The Heiress', o: { lady: true, top: 'blouse', coat: '#d9902e', dark: '#8a5214', accent: '#f6efe0', hairStyle: 'bob', hair: '#d8b46a', hat: 'cloche', hatColor: '#e9d9b0', band: '#c7822a', flowers: true, neck: 'pearls', skin: '#f6dbc4', smile: true, earrings: '#e8c35a' }, prop: 'hatbox' },
      S: { name: 'The Mystery Lady', o: { lady: true, top: 'dress', coat: '#2a1830', dark: '#120a18', hairStyle: 'bob', hair: '#120c10', hat: 'cloche', hatColor: '#18141a', band: '#7a5a9a', veil: true, neck: 'choker', lip: '#6e1230', skin: '#f1d0b5', stern: true, earrings: '#9a7ab8' }, prop: 'fan' },
      C: { name: 'The Schoolteacher', o: { lady: true, top: 'cardigan', coat: '#2f5a44', dark: '#183426', accent: '#fbf8f0', hairStyle: 'bun', hair: '#6b3f22', glasses: 'round', neck: 'scarf', neckColor: '#c7822a', skin: '#d6a27a', smile: true }, prop: 'book' }
    },
    J: {
      H: { name: 'The Porter', o: { top: 'waistcoat', coat: '#fbf8f0', dark: '#d8d2c4', vest: '#fbf8f0', neck: 'bowtie', neckColor: '#9c1f33', hat: 'porter', hatColor: '#9c1f33', hairStyle: 'short', hair: '#1c1414', skin: '#8a5a3a', smile: true }, prop: 'suitcase' },
      D: { name: 'The Newsboy', o: { top: 'waistcoat', coat: '#c7822a', dark: '#7e4a14', vest: '#5a4a3a', neck: 'kerchief', neckColor: '#c7822a', hat: 'cap', hatColor: '#6b5a3a', hair: '#b5651d', hairStyle: 'short', skin: '#f6dbc4', smile: true }, prop: 'paper' },
      S: { name: 'The Brakeman', o: { top: 'turtleneck', coat: '#24385c', dark: '#121e36', hat: 'engcap', neck: 'scarf', neckColor: '#c0392b', hair: '#4a2e1c', hairStyle: 'short', skin: '#e3b48e', mustache: 'pencil' }, prop: 'lantern' },
      C: { name: 'The Bartender', o: { top: 'waistcoat', coat: '#fbf8f0', dark: '#183426', vest: '#2f5a44', neck: 'bowtie', neckColor: '#18141a', hairStyle: 'side', hair: '#2a1e18', mustache: 'pencil', skin: '#f2d4b6', smile: true }, prop: 'shaker' }
    }
  };

  // what each court figure holds
  function courtProp(c, prop, o) {
    c.save(); ink(c); c.lineWidth = 2.4;
    const sk = o.skin;
    switch (prop) {
      case 'whisk':
        c.strokeStyle = '#7a5530'; c.lineWidth = 6; c.beginPath(); c.moveTo(58, -18); c.lineTo(64, -60); c.stroke();
        c.strokeStyle = '#a8b0b8'; c.lineWidth = 2;
        for (const k of [-1, -0.5, 0, 0.5, 1]) { c.beginPath(); c.moveTo(64, -60); c.bezierCurveTo(64 + k * 18, -80, 64 + k * 14, -108, 66, -112); c.stroke(); }
        hand(c, 58, -30, sk); break;
      case 'watch':   // a gold pocket watch on its chain
        c.strokeStyle = GOLD; c.lineWidth = 1.6; c.beginPath(); c.moveTo(20, -30); c.quadraticCurveTo(40, -10, 56, -50); c.stroke(); ink(c);
        c.fillStyle = GOLD; c.beginPath(); c.arc(58, -64, 18, 0, TAU); c.fill(); c.stroke();
        c.fillStyle = '#fbf6e6'; c.beginPath(); c.arc(58, -64, 13, 0, TAU); c.fill(); c.lineWidth = 1.2; c.stroke();
        c.fillStyle = GOLD; rrect(c, 54, -88, 8, 7, 2); c.fill(); c.stroke();
        c.lineWidth = 1.6; c.beginPath(); c.moveTo(58, -64); c.lineTo(58, -73); c.moveTo(58, -64); c.lineTo(65, -61); c.stroke();
        for (let i = 0; i < 12; i++) { const a = i / 12 * TAU; c.fillStyle = INK; c.fillRect(58 + Math.cos(a) * 10.5 - 0.7, -64 + Math.sin(a) * 10.5 - 0.7, 1.4, 1.4); }
        hand(c, 56, -44, sk); break;
      case 'oilcan':
        c.fillStyle = '#9aa2aa'; c.beginPath(); c.moveTo(40, -14); c.lineTo(44, -44); c.lineTo(72, -44); c.lineTo(76, -14); c.closePath(); c.fill(); c.stroke();
        c.beginPath(); c.moveTo(52, -44); c.lineTo(56, -56); c.lineTo(62, -56); c.lineTo(66, -44); c.closePath(); c.fill(); c.stroke();
        c.strokeStyle = INK; c.lineWidth = 4; c.beginPath(); c.moveTo(59, -56); c.lineTo(86, -96); c.stroke(); c.strokeStyle = '#9aa2aa'; c.lineWidth = 2; c.beginPath(); c.moveTo(59, -56); c.lineTo(86, -96); c.stroke();
        c.fillStyle = 'rgba(255,255,255,.3)'; c.fillRect(46, -40, 4, 22);
        hand(c, 50, -24, sk); break;
      case 'deed':   // a rolled-up land deed with a red seal
        c.save(); c.translate(-50, -50); c.rotate(-0.5);
        c.fillStyle = '#f2e6c8'; rrect(c, -10, -36, 20, 72, 9); c.fill(); c.stroke();
        c.strokeStyle = '#b3283a'; c.lineWidth = 3; c.beginPath(); c.moveTo(-10, 0); c.lineTo(10, 0); c.stroke();
        c.fillStyle = '#b3283a'; c.beginPath(); c.arc(0, 4, 6, 0, TAU); c.fill(); c.restore();
        hand(c, -46, -40, sk); break;
      case 'mic':   // a ribbon microphone on a stand
        c.strokeStyle = '#3a3a40'; c.lineWidth = 4; c.beginPath(); c.moveTo(-52, 0); c.lineTo(-52, -70); c.stroke();
        c.fillStyle = '#c8ccd2'; rrect(c, -66, -112, 28, 42, 12); c.fill(); c.stroke();
        c.strokeStyle = '#6a6e76'; c.lineWidth = 1.2; for (let y = -104; y < -76; y += 5) { c.beginPath(); c.moveTo(-62, y); c.lineTo(-42, y); c.stroke(); }
        ink(c); c.fillStyle = '#c0182a'; c.beginPath(); c.arc(54, -70, 9, 0, TAU); c.fill(); c.stroke();   // a rose too
        c.strokeStyle = '#2f6a3a'; c.lineWidth = 3; c.beginPath(); c.moveTo(50, -12); c.quadraticCurveTo(56, -40, 54, -62); c.stroke();
        hand(c, 50, -24, sk); break;
      case 'hatbox':
        c.save(); c.translate(-50, -34);
        c.fillStyle = '#f0c8d0'; c.beginPath(); c.ellipse(0, 10, 30, 10, 0, 0, TAU); c.fill(); c.stroke();
        c.fillRect(-30, -20, 60, 30); c.beginPath(); c.moveTo(-30, -20); c.lineTo(-30, 10); c.moveTo(30, -20); c.lineTo(30, 10); c.stroke();
        c.strokeStyle = 'rgba(160,28,48,.6)'; c.lineWidth = 3; for (let x = -24; x <= 24; x += 12) { c.beginPath(); c.moveTo(x, -18); c.lineTo(x, 16); c.stroke(); } ink(c);
        c.fillStyle = '#e8b0bc'; c.beginPath(); c.ellipse(0, -20, 32, 10, 0, 0, TAU); c.fill(); c.stroke();
        c.strokeStyle = GOLD; c.lineWidth = 2; c.beginPath(); c.moveTo(-14, -24); c.quadraticCurveTo(0, -44, 14, -24); c.stroke();
        c.restore(); hand(c, -40, -14, sk); break;
      case 'fan':
        c.save(); c.translate(44, -40);
        for (let i = -4; i <= 4; i++) { c.save(); c.rotate(i * 0.17 - 0.3); c.fillStyle = i % 2 ? '#18141a' : '#2a1830'; c.beginPath(); c.moveTo(0, 0); c.lineTo(-7, -46); c.quadraticCurveTo(0, -50, 7, -46); c.closePath(); c.fill(); c.strokeStyle = '#7a5a9a'; c.lineWidth = 1.2; c.stroke(); c.restore(); }
        c.strokeStyle = 'rgba(200,180,220,.6)'; c.lineWidth = 1; c.beginPath(); c.arc(0, 0, 36, -2.4, -0.4); c.stroke(); c.beginPath(); c.arc(0, 0, 24, -2.4, -0.4); c.stroke();
        c.restore(); hand(c, 44, -36, sk); break;
      case 'book':
        c.save(); c.translate(-48, -40); c.rotate(-0.15);
        c.fillStyle = '#7a1e2a'; rrect(c, -22, -30, 44, 58, 3); c.fill(); c.stroke();
        c.fillStyle = '#f2ebde'; c.fillRect(18, -27, 4, 52);
        c.strokeStyle = GOLD; c.lineWidth = 1.4; c.strokeRect(-16, -22, 30, 42); c.fillStyle = GOLD2; c.font = `700 8px ${DISP}`; c.textAlign = 'center'; c.fillText('ABC', -1, 2);
        c.restore(); hand(c, -32, -22, sk); break;
      case 'suitcase':
        c.save(); c.translate(-50, -30);
        c.fillStyle = '#7a4a24'; rrect(c, -30, -26, 60, 44, 6); c.fill(); c.stroke();
        c.strokeStyle = '#3a2210'; c.lineWidth = 2; c.beginPath(); c.moveTo(-30, -6); c.lineTo(30, -6); c.stroke(); ink(c);
        c.fillStyle = GOLD; c.fillRect(-20, -10, 6, 8); c.fillRect(14, -10, 6, 8);
        c.beginPath(); c.moveTo(-10, -26); c.lineTo(-10, -36); c.lineTo(10, -36); c.lineTo(10, -26); c.stroke();
        c.fillStyle = '#f2e6c8'; c.save(); c.translate(14, 6); c.rotate(0.2); c.fillRect(-8, -5, 16, 10); c.restore();   // a luggage label
        c.restore(); hand(c, -50, -62, sk); break;
      case 'paper':
        c.save(); c.translate(-46, -40); c.rotate(-0.12);
        c.fillStyle = '#f2ebde'; c.fillRect(-22, -34, 44, 60); c.strokeRect(-22, -34, 44, 60);
        c.fillStyle = INK; c.font = `900 11px ${DISP}`; c.textAlign = 'center'; c.fillText('EXTRA!', 0, -20);
        c.fillStyle = 'rgba(40,30,20,.35)'; c.fillRect(-17, -14, 34, 12); for (let y = 2; y < 22; y += 5) c.fillRect(-17, y, 34, 2);
        c.restore(); hand(c, -32, -20, sk); break;
      case 'lantern':   // a brakeman's signal lantern
        c.strokeStyle = INK; c.lineWidth = 2.4; c.beginPath(); c.arc(58, -96, 12, Math.PI, 0); c.stroke();
        c.fillStyle = '#5a6068'; rrect(c, 44, -94, 28, 8, 2); c.fill(); c.stroke();
        const lg = c.createRadialGradient(58, -70, 2, 58, -70, 22); lg.addColorStop(0, '#fff6c8'); lg.addColorStop(0.5, '#ffb030'); lg.addColorStop(1, '#c0392b');
        c.fillStyle = lg; rrect(c, 46, -86, 24, 30, 6); c.fill(); c.stroke();
        c.strokeStyle = '#5a6068'; c.lineWidth = 2; c.beginPath(); c.moveTo(52, -86); c.lineTo(52, -56); c.moveTo(64, -86); c.lineTo(64, -56); c.stroke(); ink(c);
        c.fillStyle = '#5a6068'; rrect(c, 44, -58, 28, 8, 2); c.fill(); c.stroke();
        hand(c, 58, -104, sk); break;
      case 'shaker':
        c.fillStyle = '#c8ccd2'; c.beginPath(); c.moveTo(48, -20); c.lineTo(44, -70); c.lineTo(72, -70); c.lineTo(68, -20); c.closePath(); c.fill(); c.stroke();
        c.beginPath(); c.moveTo(46, -70); c.lineTo(50, -84); c.lineTo(66, -84); c.lineTo(70, -70); c.closePath(); c.fill(); c.stroke();
        rrect(c, 53, -92, 10, 8, 2); c.fill(); c.stroke();
        c.fillStyle = 'rgba(255,255,255,.45)'; c.fillRect(50, -66, 4, 40);
        hand(c, 58, -40, sk); break;
    }
    c.restore();
  }

  // a train window behind a court figure: the night prairie rolling by
  function windowScene(c, x, y, w, h, s, night) {
    const p = PAL[s];
    c.fillStyle = p.wall; c.fillRect(x, y, w, h);
    c.fillStyle = 'rgba(255,255,255,.28)'; for (let xx = x; xx < x + w; xx += 20) c.fillRect(xx, y, 2, h);   // Deco pinstripe panelling
    const wx = x + 26, wy = y + 30, ww = w - 52, wh = h * 0.5;
    c.save(); rrect(c, wx, wy, ww, wh, 26); c.clip();
    const g = c.createLinearGradient(0, wy, 0, wy + wh);
    g.addColorStop(0, night ? '#120c22' : '#1c2848'); g.addColorStop(0.7, night ? '#3a2448' : '#4a5a84'); g.addColorStop(1, night ? '#6a3a4a' : '#c89a6a');
    c.fillStyle = g; c.fillRect(wx, wy, ww, wh);
    const R = rng(s.charCodeAt(0) * 31 + (night ? 7 : 3));
    for (let i = 0; i < 26; i++) { c.fillStyle = `rgba(255,248,220,${0.3 + R() * 0.6})`; c.beginPath(); c.arc(wx + R() * ww, wy + R() * wh * 0.6, 0.6 + R() * 1.2, 0, TAU); c.fill(); }
    c.fillStyle = '#f6eccc'; c.beginPath(); c.arc(wx + ww * 0.78, wy + wh * 0.26, 11, 0, TAU); c.fill();
    c.fillStyle = night ? '#120c22' : '#1c2848'; c.beginPath(); c.arc(wx + ww * 0.78 + 5, wy + wh * 0.26 - 3, 10, 0, TAU); c.fill();
    c.fillStyle = 'rgba(20,14,24,.85)'; c.beginPath(); c.moveTo(wx, wy + wh);
    for (let i = 0; i <= 12; i++) c.lineTo(wx + ww * i / 12, wy + wh * (0.78 - 0.08 * Math.sin(i * 1.3 + s.charCodeAt(0))));
    c.lineTo(wx + ww, wy + wh); c.closePath(); c.fill();
    c.fillStyle = 'rgba(255,255,255,.08)'; c.beginPath(); c.moveTo(wx, wy); c.lineTo(wx + ww * 0.4, wy); c.lineTo(wx, wy + wh * 0.6); c.fill();   // glass glint
    c.restore();
    c.strokeStyle = GOLD; c.lineWidth = 3; rrect(c, wx, wy, ww, wh, 26); c.stroke();
    c.lineWidth = 1; rrect(c, wx - 5, wy - 5, ww + 10, wh + 10, 30); c.stroke();
  }

  function courtCard(c, rank, s) {
    const color = PAL[s].ink, w = CW, h = CH;
    const fx = 76, fy = 64, fw = w - 152, fh = h - 128;
    const who = COURT[rank][s], dame = rank === 'Q' && s === 'S';
    c.save(); rrect(c, fx, fy, fw, fh, 12); c.clip();
    windowScene(c, fx, fy, fw, fh, s, dame || rank === 'K');
    if (dame) { c.fillStyle = 'rgba(42,24,64,.55)'; c.fillRect(fx, fy, fw, fh); }
    const glow = c.createRadialGradient(fx + fw / 2, fy + fh * 0.62, 10, fx + fw / 2, fy + fh * 0.62, fh * 0.55);
    glow.addColorStop(0, dame ? 'rgba(200,170,230,.3)' : 'rgba(255,240,200,.45)'); glow.addColorStop(1, 'rgba(255,240,200,0)'); c.fillStyle = glow; c.fillRect(fx, fy, fw, fh);
    const fit = 1.48, o = who.o;
    c.save(); c.translate(fx + fw / 2, fy + fh + 2); c.scale(fit, fit);
    bust(c, o); courtProp(c, who.prop, o);
    c.restore();
    // the name plate: a brass Pullman nameplate
    const pg = c.createLinearGradient(0, fy + fh - 30, 0, fy + fh - 8); pg.addColorStop(0, '#f2dc9a'); pg.addColorStop(1, '#c9a24c');
    c.fillStyle = dame ? 'rgba(18,10,24,.88)' : pg; rrect(c, fx + 14, fy + fh - 30, fw - 28, 22, 4); c.fill();
    c.strokeStyle = dame ? '#e8c35a' : '#7a5a1e'; c.lineWidth = 1.2; c.stroke();
    c.fillStyle = dame ? '#f2d68e' : '#2a1a0a'; c.font = `italic 700 12.5px ${DISP}`; c.textAlign = 'center'; c.fillText(who.name, fx + fw / 2, fy + fh - 14.5);
    c.restore();
    c.strokeStyle = dame ? '#d9b45a' : GOLD; c.lineWidth = 3; rrect(c, fx, fy, fw, fh, 12); c.stroke();
    c.lineWidth = 1; rrect(c, fx + 5, fy + 5, fw - 10, fh - 10, 8); c.stroke();
    suit(c, s, fx + 20, fy + 22, 22, dame ? '#f2d68e' : color, false);
    suit(c, s, fx + fw - 20, fy + 22, 22, dame ? '#f2d68e' : color, false);
  }

  const PIPS = {
    2: [[.5, 0], [.5, 1]], 3: [[.5, 0], [.5, .5], [.5, 1]], 4: [[0, 0], [1, 0], [0, 1], [1, 1]],
    5: [[0, 0], [1, 0], [.5, .5], [0, 1], [1, 1]], 6: [[0, 0], [1, 0], [0, .5], [1, .5], [0, 1], [1, 1]],
    7: [[0, 0], [1, 0], [.5, .25], [0, .5], [1, .5], [0, 1], [1, 1]],
    8: [[0, 0], [1, 0], [.5, .25], [0, .5], [1, .5], [.5, .75], [0, 1], [1, 1]],
    9: [[0, 0], [1, 0], [0, 1 / 3], [1, 1 / 3], [.5, .5], [0, 2 / 3], [1, 2 / 3], [0, 1], [1, 1]],
    10: [[0, 0], [1, 0], [.5, 1 / 6], [0, 1 / 3], [1, 1 / 3], [0, 2 / 3], [1, 2 / 3], [.5, 5 / 6], [0, 1], [1, 1]]
  };
  function pipCard(c, n, s) {
    const color = PAL[s].ink, x0 = 114, x1 = CW - 114, y0 = 104, y1 = CH - 104;
    for (const [px, py] of PIPS[n]) {
      const x = x0 + (x1 - x0) * px, y = y0 + (y1 - y0) * py;
      if (py > 0.5) { c.save(); c.translate(x, y); c.rotate(Math.PI); suit(c, s, 0, 0, 56, color); c.restore(); }
      else suit(c, s, x, y, 56, color);
    }
  }
  // an Art Deco sunburst: alternating rays inside a stepped frame
  function sunburst(c, cx, cy, r0, r1, n, col, alpha) {
    c.save(); c.fillStyle = col; c.globalAlpha = alpha;
    for (let i = 0; i < n; i++) {
      const a = i / n * TAU, da = TAU / n * 0.32;
      c.beginPath(); c.moveTo(cx + Math.cos(a - da) * r0, cy + Math.sin(a - da) * r0); c.lineTo(cx + Math.cos(a) * r1, cy + Math.sin(a) * r1); c.lineTo(cx + Math.cos(a + da) * r0, cy + Math.sin(a + da) * r0); c.closePath(); c.fill();
    }
    c.restore();
  }
  function aceCard(c, s) {
    const color = PAL[s].ink, cx = CW / 2, cy = CH / 2;
    c.save(); c.beginPath(); c.ellipse(cx, cy, 110, 140, 0, 0, TAU); c.clip();
    sunburst(c, cx, cy, 70, 200, 28, GOLD, 0.35); c.restore();
    c.save(); c.strokeStyle = GOLD; c.lineWidth = 2.4; c.beginPath(); c.ellipse(cx, cy, 110, 140, 0, 0, TAU); c.stroke();
    c.lineWidth = 1; c.beginPath(); c.ellipse(cx, cy, 116, 146, 0, 0, TAU); c.stroke();
    c.fillStyle = IVORY1; c.beginPath(); c.ellipse(cx, cy, 78, 96, 0, 0, TAU); c.fill(); c.strokeStyle = GOLD; c.lineWidth = 1.5; c.stroke();
    c.restore();
    suit(c, s, cx, cy, 124, color);
    if (s === 'S') {   // the Ace of Spades carries the railroad's name
      c.fillStyle = '#1f2a4a'; rrect(c, cx - 96, cy + 108, 192, 30, 6); c.fill(); c.strokeStyle = GOLD; c.lineWidth = 1.5; c.stroke();
      c.fillStyle = GOLD2; c.font = `italic 700 15px ${DISP}`; c.textAlign = 'center'; c.fillText('The Starlight Limited', cx, cy + 128);
    }
  }

  /* ---------- the card back: midnight blue and brass, a streamliner and a star ---------- */
  function star(c, x, y, r, col) {
    c.save(); c.fillStyle = col; c.beginPath();
    for (let i = 0; i < 10; i++) { const a = -Math.PI / 2 + i * Math.PI / 5, rr = i % 2 ? r * 0.42 : r; c.lineTo(x + Math.cos(a) * rr, y + Math.sin(a) * rr); }
    c.closePath(); c.fill(); c.restore();
  }
  // the streamliner's nose seen head-on: round headlamp, chrome wings, speed lines
  function locomotive(c, cx, cy, k, col) {
    c.save(); c.translate(cx, cy); c.scale(k, k);
    c.fillStyle = col; c.strokeStyle = col;
    c.beginPath(); c.moveTo(-50, 40); c.bezierCurveTo(-58, -20, -34, -56, 0, -58); c.bezierCurveTo(34, -56, 58, -20, 50, 40); c.closePath(); c.globalAlpha = 0.25; c.fill(); c.globalAlpha = 1;
    c.lineWidth = 3; c.stroke();
    c.lineWidth = 2; for (const y of [8, 18, 28]) { c.beginPath(); c.moveTo(-90 + Math.abs(y - 18), y); c.lineTo(-30, y); c.moveTo(30, y); c.lineTo(90 - Math.abs(y - 18), y); c.stroke(); }   // the chrome wings
    c.beginPath(); c.arc(0, -16, 16, 0, TAU); c.fill();
    c.fillStyle = '#1a2240'; c.beginPath(); c.arc(0, -16, 10, 0, TAU); c.fill();
    c.fillStyle = '#fff6d8'; c.beginPath(); c.arc(0, -16, 7, 0, TAU); c.fill();
    c.fillStyle = col; rrect(c, -30, 40, 60, 8, 3); c.fill();
    for (let x = -24; x <= 24; x += 8) c.fillRect(x - 1.5, 14, 3, 22);   // the grille
    c.restore();
  }
  function backCard() {
    const cv = canvas(CW, CH), c = cv.getContext('2d');
    rrect(c, 1, 1, CW - 2, CH - 2, 22); c.fillStyle = '#f6eedb'; c.fill();
    rrect(c, 12, 12, CW - 24, CH - 24, 14); const g = c.createLinearGradient(0, 0, 0, CH); g.addColorStop(0, '#1f2a4a'); g.addColorStop(1, '#0f1630'); c.fillStyle = g; c.fill();
    c.save(); c.clip();
    // Deco chevrons and little stars
    c.strokeStyle = 'rgba(232,195,90,.22)'; c.lineWidth = 1.2;
    for (let y = -20; y < CH + 20; y += 22) { c.beginPath(); for (let x = 0; x <= CW; x += 22) c.lineTo(x, y + ((x / 22) % 2 ? 11 : 0)); c.stroke(); }
    for (let y = 30; y < CH; y += 44) for (let x = 22 + ((y - 30) / 44 % 2) * 22; x < CW; x += 44) star(c, x, y, 3.2, 'rgba(232,195,90,.35)');
    c.restore();
    c.strokeStyle = '#e8c35a'; c.lineWidth = 3; rrect(c, 26, 26, CW - 52, CH - 52, 12); c.stroke();
    c.lineWidth = 1; rrect(c, 32, 32, CW - 64, CH - 64, 9); c.stroke();
    // stepped Deco corners
    for (const [x, y, sx, sy] of [[32, 32, 1, 1], [CW - 32, 32, -1, 1], [32, CH - 32, 1, -1], [CW - 32, CH - 32, -1, -1]]) {
      c.fillStyle = '#e8c35a'; c.beginPath(); c.moveTo(x, y); c.lineTo(x + sx * 34, y); c.lineTo(x + sx * 34, y + sy * 8); c.lineTo(x + sx * 18, y + sy * 8); c.lineTo(x + sx * 18, y + sy * 18); c.lineTo(x + sx * 8, y + sy * 18); c.lineTo(x + sx * 8, y + sy * 34); c.lineTo(x, y + sy * 34); c.closePath(); c.fill();
    }
    const cx = CW / 2, cy = CH / 2;
    c.save(); c.beginPath(); c.ellipse(cx, cy, 104, 138, 0, 0, TAU); c.clip();
    const mg = c.createRadialGradient(cx, cy - 20, 10, cx, cy, 150); mg.addColorStop(0, '#2c3a66'); mg.addColorStop(1, '#121a36'); c.fillStyle = mg; c.fillRect(0, 0, CW, CH);
    sunburst(c, cx, cy - 34, 30, 170, 32, '#e8c35a', 0.2);
    c.restore();
    c.strokeStyle = '#e8c35a'; c.lineWidth = 3; c.beginPath(); c.ellipse(cx, cy, 104, 138, 0, 0, TAU); c.stroke();
    c.lineWidth = 1; c.beginPath(); c.ellipse(cx, cy, 96, 130, 0, 0, TAU); c.stroke();
    star(c, cx, cy - 78, 18, '#f2d68e');
    locomotive(c, cx, cy + 20, 0.95, '#e8c35a');
    c.fillStyle = '#f2d68e'; c.font = `italic 700 21px ${DISP}`; c.textAlign = 'center'; c.fillText('Starlight', cx, 74); c.fillText('Limited', cx, CH - 54);
    rrect(c, 1, 1, CW - 2, CH - 2, 22); c.strokeStyle = 'rgba(70,50,30,.45)'; c.lineWidth = 2; c.stroke();
    return cv;
  }

  /* ---------- public: a card face ---------- */
  const RANK = { 14: 'A', 11: 'J', 12: 'Q', 13: 'K' };
  function face(card) {
    const cv = canvas(CW, CH), c = cv.getContext('2d');
    paper(c, CW, CH, card.r * 7 + 'SHDC'.indexOf(card.s) * 131);
    const color = PAL[card.s].ink;
    border(c, CW, CH, card.s === 'H' || card.s === 'D' ? 'rgba(160,28,48,.75)' : GOLD);
    const label = RANK[card.r] || String(card.r);
    index(c, label, card.s, color, CW, CH);
    if (card.r === 14) aceCard(c, card.s);
    else if (card.r > 10) courtCard(c, label, card.s);
    else pipCard(c, card.r, card.s);
    return cv;
  }

  /* ---------- club-car props, seen from above ---------- */
  function cup(size) {   // dining-car china: a cup of black coffee with the railroad's blue band
    const cv = canvas(size, size), c = cv.getContext('2d'), k = size / 220; c.scale(k, k);
    c.fillStyle = 'rgba(0,0,0,.35)'; c.beginPath(); c.ellipse(114, 120, 92, 88, 0, 0, TAU); c.fill();
    c.strokeStyle = INK; c.lineWidth = 3;
    let g = c.createRadialGradient(100, 96, 20, 110, 110, 92); g.addColorStop(0, '#fffdf8'); g.addColorStop(1, '#e6ded0');
    c.fillStyle = g; c.beginPath(); c.arc(110, 110, 88, 0, TAU); c.fill(); c.stroke();
    c.strokeStyle = '#1f2a4a'; c.lineWidth = 7; c.beginPath(); c.arc(110, 110, 78, 0, TAU); c.stroke();
    c.strokeStyle = '#c9a24c'; c.lineWidth = 1.6; c.beginPath(); c.arc(110, 110, 72, 0, TAU); c.stroke();
    // the railroad's star on the saucer rim
    star(c, 110, 28, 7, '#c9a24c');
    // spoon
    c.save(); c.translate(48, 140); c.rotate(0.75); c.fillStyle = '#c4ccd4'; c.strokeStyle = INK; c.lineWidth = 2;
    c.beginPath(); c.ellipse(0, -30, 9, 14, 0, 0, TAU); c.fill(); c.stroke(); rrect(c, -3, -18, 6, 54, 3); c.fill(); c.stroke(); c.restore();
    // two sugar cubes
    for (const [x, y, r] of [[150, 168, 0.3], [168, 154, -0.2]]) { c.save(); c.translate(x, y); c.rotate(r); c.fillStyle = '#fbfbf8'; c.strokeStyle = INK; c.lineWidth = 2; rrect(c, -9, -9, 18, 18, 3); c.fill(); c.stroke(); c.restore(); }
    // cup and handle
    c.strokeStyle = INK; c.lineWidth = 3; c.fillStyle = '#fbf8f2';
    c.beginPath(); c.ellipse(166, 82, 22, 12, -0.6, 0, TAU); c.fill(); c.stroke();
    c.fillStyle = '#fbf8f2'; c.beginPath(); c.arc(106, 106, 56, 0, TAU); c.fill(); c.stroke();
    c.strokeStyle = '#1f2a4a'; c.lineWidth = 5; c.beginPath(); c.arc(106, 106, 51, 0, TAU); c.stroke();
    c.strokeStyle = INK; c.lineWidth = 2; c.beginPath(); c.arc(106, 106, 46, 0, TAU); c.stroke();
    g = c.createRadialGradient(98, 98, 4, 106, 106, 45); g.addColorStop(0, '#5a3418'); g.addColorStop(0.6, '#2e1a0c'); g.addColorStop(1, '#180c04');
    c.fillStyle = g; c.beginPath(); c.arc(106, 106, 45, 0, TAU); c.fill();
    c.fillStyle = 'rgba(255,240,210,.22)'; c.beginPath(); c.ellipse(90, 88, 16, 6, -0.7, 0, TAU); c.fill();
    c.strokeStyle = 'rgba(255,230,190,.16)'; c.lineWidth = 1.5; c.beginPath(); c.arc(106, 106, 30, 0.4, 2.2); c.stroke();
    return cv;
  }
  function pad(size) {   // the score pad and a pencil
    const cv = canvas(size, size), c = cv.getContext('2d'), k = size / 220; c.scale(k, k);
    c.save(); c.translate(110, 112); c.rotate(-0.12);
    c.fillStyle = 'rgba(0,0,0,.35)'; rrect(c, -70, -84, 148, 182, 6); c.fill();
    c.fillStyle = '#e8e2d2'; rrect(c, -74, -92, 148, 182, 4); c.fill();   // the pages under the top sheet
    c.fillStyle = '#fbf7ea'; rrect(c, -76, -94, 148, 182, 4); c.fill(); c.strokeStyle = 'rgba(60,40,20,.5)'; c.lineWidth = 1.5; c.stroke();
    c.fillStyle = '#1f2a4a'; rrect(c, -76, -94, 148, 26, 4); c.fill();
    c.fillStyle = '#e8c35a'; c.font = `italic 700 15px ${DISP}`; c.textAlign = 'center'; c.fillText('Gin Rummy', -2, -76);
    c.strokeStyle = 'rgba(40,80,160,.35)'; c.lineWidth = 1; for (let y = -50; y < 84; y += 13) { c.beginPath(); c.moveTo(-70, y); c.lineTo(66, y); c.stroke(); }
    c.strokeStyle = 'rgba(180,40,50,.5)'; c.beginPath(); c.moveTo(-2, -66); c.lineTo(-2, 84); c.stroke(); c.beginPath(); c.moveTo(-70, -52); c.lineTo(66, -52); c.stroke();
    c.fillStyle = '#3a3a48'; c.font = `italic 600 11px ${SERIF}`; c.fillText('YOU', -38, -56); c.fillText('THEM', 32, -56);
    c.font = `italic 600 12px ${SERIF}`; c.fillStyle = 'rgba(40,40,60,.8)';
    for (const [y, a, b] of [[-41, '18', ''], [-28, '', '26'], [-15, '43', ''], [-2, '', '']]) { if (a) c.fillText(a, -38, y); if (b) c.fillText(b, 32, y); }
    // little box tallies
    c.strokeStyle = 'rgba(40,40,60,.7)'; c.lineWidth = 1.2; for (const x of [-52, -46]) { c.beginPath(); c.moveTo(x, 66); c.lineTo(x, 78); c.stroke(); } c.beginPath(); c.moveTo(24, 66); c.lineTo(24, 78); c.stroke();
    c.restore();
    // the pencil
    c.save(); c.translate(150, 150); c.rotate(-0.95);
    c.fillStyle = '#e8b830'; c.strokeStyle = INK; c.lineWidth = 2; c.fillRect(-7, -64, 14, 104); c.strokeRect(-7, -64, 14, 104);
    c.fillStyle = '#c8a028'; c.fillRect(-1.5, -64, 3, 104);
    c.fillStyle = '#c8c4bc'; c.fillRect(-7, 40, 14, 10); c.strokeRect(-7, 40, 14, 10); c.fillStyle = '#d0707a'; rrect(c, -7, 50, 14, 12, 4); c.fill(); c.stroke();
    c.fillStyle = '#f0d8b0'; c.beginPath(); c.moveTo(-7, -64); c.lineTo(0, -86); c.lineTo(7, -64); c.closePath(); c.fill(); c.stroke();
    c.fillStyle = '#2a2a2a'; c.beginPath(); c.moveTo(-2.4, -79); c.lineTo(0, -86); c.lineTo(2.4, -79); c.closePath(); c.fill();
    c.restore();
    return cv;
  }

  /* ---------- the tabletop: club-car green baize ---------- */
  function felt(W, H) {
    const cv = canvas(W, H), c = cv.getContext('2d'), img = c.createImageData(W, H), d = img.data;
    const P = new Uint8Array(512); { const R = rng(1937); const p = [...Array(256).keys()]; for (let i = 255; i > 0; i--) { const j = Math.floor(R() * (i + 1)); [p[i], p[j]] = [p[j], p[i]]; } for (let i = 0; i < 512; i++) P[i] = p[i & 255]; }
    const fade = t => t * t * t * (t * (t * 6 - 15) + 10), lerp = (a, b, t) => a + (b - a) * t;
    const grad = (h, x, y) => { const v = h & 3; return (v & 1 ? -x : x) + (v & 2 ? -y : y); };
    function noise(x, y) { const X = Math.floor(x) & 255, Y = Math.floor(y) & 255; x -= Math.floor(x); y -= Math.floor(y); const u = fade(x), v = fade(y); const a = P[X] + Y, b = P[X + 1] + Y; return lerp(lerp(grad(P[a], x, y), grad(P[b], x - 1, y), u), lerp(grad(P[a + 1], x, y - 1), grad(P[b + 1], x - 1, y - 1), u), v); }
    const R = rng(77);
    for (let y = 0; y < H; y++) for (let x = 0; x < W; x++) {
      const cloud = 0.5 + 0.35 * noise(x / 260, y / 260) + 0.15 * noise(x / 60 + 9, y / 60);
      const fib = (R() - 0.5) * 14 + noise(x / 2.2, y / 9) * 8;   // the nap of the cloth
      const i = (y * W + x) * 4;
      d[i] = 18 + 16 * cloud + fib * 0.5; d[i + 1] = 72 + 34 * cloud + fib; d[i + 2] = 46 + 22 * cloud + fib * 0.7; d[i + 3] = 255;
    }
    c.putImageData(img, 0, 0);
    return cv;
  }

  /* ---------- the icon: a run of hearts on midnight blue, under the Starlight star ---------- */
  function icon(size) {
    const cv = canvas(size, size), c = cv.getContext('2d'), s = size / 512;
    c.scale(s, s);
    rrect(c, 16, 16, 480, 480, 96); const g = c.createLinearGradient(0, 16, 0, 496); g.addColorStop(0, '#2c3a66'); g.addColorStop(1, '#0f1630'); c.fillStyle = g; c.fill();
    c.save(); rrect(c, 16, 16, 480, 480, 96); c.clip(); sunburst(c, 256, 130, 40, 520, 30, '#e8c35a', 0.12); c.restore();
    c.strokeStyle = '#e8c35a'; c.lineWidth = 12; rrect(c, 40, 40, 432, 432, 76); c.stroke();
    star(c, 256, 100, 44, '#f2d68e');
    const ranks = ['7', '8', '9'];
    [-1, 0, 1].forEach((k, i) => {   // a fanned run, with the corner indexes showing like a real hand
      c.save(); c.translate(256 + k * 70, 400); c.rotate(k * 0.3); c.translate(0, -110);
      rrect(c, -78, -112, 156, 224, 18); c.fillStyle = '#fffaf0'; c.fill(); c.strokeStyle = '#24170f'; c.lineWidth = 6; c.stroke();
      c.fillStyle = '#c0182a'; c.font = `700 62px ${DISP}`; c.textAlign = 'center'; c.fillText(ranks[i], -46, -48);
      suit(c, 'H', -46, -16, 36, '#c0182a', false);
      if (i === 2) suit(c, 'H', 8, 40, 92, '#c0182a', false);
      c.restore();
    });
    return cv;
  }

  return { face, back: backCard, cup, pad, felt, icon, locomotive, star, sunburst, heartPath, suit, CW, CH, PAL, DISP, SERIF, COURT };
})();
