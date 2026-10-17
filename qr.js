// QR para enlaces (invitaciones de «Mi equipo»): modo byte, corrección M, versiones 1–15 (hasta ~410 bytes), sin
// dependencias. AppQR.matrix(texto[, máscara]) da la matriz (true = oscuro); AppQR.svg(texto) un <svg> con margen.
(function () {
  // Por versión: [corrección por bloque, [bloques, datos], [bloques, datos]] (nivel M).
  const BLOCKS = [null,
    [10, [1, 16]], [16, [1, 28]], [26, [1, 44]], [18, [2, 32]], [24, [2, 43]],
    [16, [4, 27]], [18, [4, 31]], [22, [2, 38], [2, 39]], [22, [3, 36], [2, 37]], [26, [4, 43], [1, 44]],
    [30, [1, 50], [4, 51]], [22, [6, 36], [2, 37]], [22, [8, 37], [1, 38]], [24, [4, 40], [5, 41]], [24, [5, 41], [5, 42]]];
  const ALIGN = [null, [], [6, 18], [6, 22], [6, 26], [6, 30], [6, 34], [6, 22, 38], [6, 24, 42], [6, 26, 46], [6, 28, 50],
    [6, 30, 54], [6, 32, 58], [6, 34, 62], [6, 26, 46, 66], [6, 26, 48, 70]];
  const EXP = new Array(512), LOG = new Array(256);
  for (let i = 0, x = 1; i < 255; i++) { EXP[i] = x; LOG[x] = i; x <<= 1; if (x & 256) x ^= 0x11d; }
  for (let i = 255; i < 512; i++) EXP[i] = EXP[i - 255];
  const mul = (a, b) => a && b ? EXP[LOG[a] + LOG[b]] : 0;
  function ecc(data, n) {
    let gen = [1];
    for (let i = 0; i < n; i++) { const g = new Array(gen.length + 1).fill(0); gen.forEach((c, j) => { g[j] ^= c; g[j + 1] ^= mul(c, EXP[i]); }); gen = g; }
    const r = data.concat(new Array(n).fill(0));
    for (let i = 0; i < data.length; i++) { const f = r[i]; if (f) for (let j = 0; j < gen.length; j++) r[i + j] ^= mul(gen[j], f); }
    return r.slice(data.length);
  }
  const bch = (data, gen, bits) => { let d = data << bits; const top = g => 31 - Math.clz32(g); while (d && top(d) >= top(gen)) d ^= gen << (top(d) - top(gen)); return (data << bits) | d; };
  const MASKS = [(i, j) => (i + j) % 2 === 0, i => i % 2 === 0, (i, j) => j % 3 === 0, (i, j) => (i + j) % 3 === 0,
    (i, j) => (Math.floor(i / 2) + Math.floor(j / 3)) % 2 === 0, (i, j) => (i * j) % 2 + (i * j) % 3 === 0,
    (i, j) => ((i * j) % 2 + (i * j) % 3) % 2 === 0, (i, j) => ((i * j) % 3 + (i + j) % 2) % 2 === 0];
  function codewords(bytes) {
    for (let v = 1; v < BLOCKS.length; v++) {
      const [ecn, ...groups] = BLOCKS[v], total = groups.reduce((t, [n, d]) => t + n * d, 0), lenBits = v < 10 ? 8 : 16;
      if (4 + lenBits + bytes.length * 8 > total * 8) continue;
      const bits = [], put = (val, n) => { for (let i = n - 1; i >= 0; i--) bits.push((val >>> i) & 1); };
      put(4, 4); put(bytes.length, lenBits); bytes.forEach(b => put(b, 8));
      if (bits.length + 4 <= total * 8) put(0, 4);
      while (bits.length % 8) bits.push(0);
      const data = [];
      for (let i = 0; i < bits.length; i += 8) data.push(bits.slice(i, i + 8).reduce((a, b) => a * 2 + b, 0));
      for (let k = 0; data.length < total; k++) data.push(k % 2 ? 0x11 : 0xec);
      const blocks = [];
      let at = 0;
      groups.forEach(([n, d]) => { for (let b = 0; b < n; b++) { const dc = data.slice(at, at + d); at += d; blocks.push([dc, ecc(dc, ecn)]); } });
      const out = [], maxD = Math.max(...blocks.map(b => b[0].length));
      for (let i = 0; i < maxD; i++) blocks.forEach(b => { if (i < b[0].length) out.push(b[0][i]); });
      for (let i = 0; i < ecn; i++) blocks.forEach(b => out.push(b[1][i]));
      return {version: v, data: out};
    }
    throw new Error('Texto demasiado largo para el QR');
  }
  function build(version, data, mask) {
    const n = version * 4 + 17, m = Array.from({length: n}, () => new Array(n).fill(null));
    const probe = (r0, c0) => { for (let r = -1; r <= 7; r++) for (let c = -1; c <= 7; c++) { const R = r0 + r, C = c0 + c; if (R < 0 || C < 0 || R >= n || C >= n) continue; m[R][C] = (r >= 0 && r <= 6 && (c === 0 || c === 6)) || (c >= 0 && c <= 6 && (r === 0 || r === 6)) || (r >= 2 && r <= 4 && c >= 2 && c <= 4); } };
    probe(0, 0); probe(n - 7, 0); probe(0, n - 7);
    const pos = ALIGN[version];
    for (const r of pos) for (const c of pos) { if (m[r][c] !== null) continue; for (let y = -2; y <= 2; y++) for (let x = -2; x <= 2; x++) m[r + y][c + x] = y === -2 || y === 2 || x === -2 || x === 2 || (y === 0 && x === 0); }
    for (let i = 8; i < n - 8; i++) { if (m[i][6] === null) m[i][6] = i % 2 === 0; if (m[6][i] === null) m[6][i] = i % 2 === 0; }
    const fmt = bch(mask, 0x537, 10) ^ 0x5412; // nivel M = 00
    for (let i = 0; i < 15; i++) {
      const on = ((fmt >> i) & 1) === 1;
      if (i < 6) m[i][8] = on; else if (i < 8) m[i + 1][8] = on; else m[n - 15 + i][8] = on;
      if (i < 8) m[8][n - i - 1] = on; else if (i < 9) m[8][15 - i] = on; else m[8][14 - i] = on;
    }
    m[n - 8][8] = true;
    if (version >= 7) {
      const vb = bch(version, 0x1f25, 12);
      for (let i = 0; i < 18; i++) { const on = ((vb >> i) & 1) === 1; m[Math.floor(i / 3)][i % 3 + n - 11] = on; m[i % 3 + n - 11][Math.floor(i / 3)] = on; }
    }
    const fn = MASKS[mask];
    let inc = -1, row = n - 1, bit = 7, byte = 0;
    for (let c0 = n - 1; c0 > 0; c0 -= 2) {
      const col = c0 <= 6 ? c0 - 1 : c0;
      for (;;) {
        for (const c of [col, col - 1]) {
          if (m[row][c] !== null) continue;
          let dark = byte < data.length && ((data[byte] >> bit) & 1) === 1;
          if (fn(row, c)) dark = !dark;
          m[row][c] = dark;
          if (--bit < 0) { byte++; bit = 7; }
        }
        row += inc;
        if (row < 0 || row >= n) { row -= inc; inc = -inc; break; }
      }
    }
    return m;
  }
  // Puntaje de penalización estándar (las 4 reglas) para elegir la máscara más legible.
  function penalty(m) {
    const n = m.length;
    let p = 0, dark = 0;
    const runs = line => { let s = 0, k = 1; for (let i = 1; i <= line.length; i++) { if (i < line.length && line[i] === line[i - 1]) k++; else { if (k >= 5) s += k - 2; k = 1; } } return s; };
    const pat = line => { let s = 0; const a = [1, 0, 1, 1, 1, 0, 1]; for (let i = 0; i + 7 <= line.length; i++) { if (a.every((v, k) => !!line[i + k] === !!v)) { const before = line.slice(Math.max(0, i - 4), i), after = line.slice(i + 7, i + 11); if ((before.length === 4 && before.every(x => !x)) || (after.length === 4 && after.every(x => !x))) s += 40; } } return s; };
    for (let i = 0; i < n; i++) { const row = m[i], col = m.map(r => r[i]); p += runs(row) + runs(col) + pat(row) + pat(col); row.forEach(x => { if (x) dark++; }); }
    for (let i = 0; i < n - 1; i++) for (let j = 0; j < n - 1; j++) { const v = m[i][j]; if (m[i][j + 1] === v && m[i + 1][j] === v && m[i + 1][j + 1] === v) p += 3; }
    p += Math.floor(Math.abs(dark * 100 / (n * n) - 50) / 5) * 10;
    return p;
  }
  function matrix(text, mask) {
    const {version, data} = codewords([...new TextEncoder().encode(String(text))]);
    if (mask != null) return build(version, data, mask);
    let best = null, score = Infinity;
    for (let k = 0; k < 8; k++) { const mm = build(version, data, k), s = penalty(mm); if (s < score) { best = mm; score = s; } }
    return best;
  }
  function svg(text, {margin = 4, title = ''} = {}) {
    const m = matrix(text), n = m.length + margin * 2;
    let d = '';
    m.forEach((row, r) => row.forEach((on, c) => { if (on) d += `M${c + margin} ${r + margin}h1v1h-1z`; }));
    const t = String(title).replace(/[&<>"]/g, ch => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;'}[ch]));
    return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${n} ${n}" shape-rendering="crispEdges" role="img" aria-label="${t || 'Código QR'}"><rect width="${n}" height="${n}" fill="#fff"/><path d="${d}" fill="#000"/></svg>`;
  }
  const api = {matrix, svg};
  if (typeof window !== 'undefined') window.AppQR = api;
  if (typeof module !== 'undefined') module.exports = api;
})();
