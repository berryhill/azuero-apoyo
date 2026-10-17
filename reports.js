// Reportes (Supply «🖨️ Informe PDF» y la página «📰 Reportes»): elegir formato, período (atajos o de un día a otro) y
// alcance (👥 mi equipo / 🌐 todos los equipos); el informe se abre en otra pestaña listo para imprimir o «Guardar como
// PDF». Datos: GET /api/full/reports/supply (reports.go), sin teléfonos ni contactos. Formatos:
//   - simple: tablas para imprimir.
//   - boletín: el mismo informe con el estilo del boletín del equipo (STYLES; hoy «unidos», el de Unidos por Azuero).
(function () {
  const esc = v => String(v ?? '').replace(/[&<>"']/g, c => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
  const fold = v => String(v ?? '').normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase().trim();
  const num = v => { const n = parseInt(String(v ?? '').replace(/\D/g, ''), 10); return isNaN(n) ? 0 : n; };
  const fmt = n => String(n).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
  const day = v => { const d = new Date(v); return isNaN(d) ? '' : d.toLocaleDateString('es-PA', {day: 'numeric', month: 'short', year: 'numeric'}); };
  const dayShort = v => { const d = new Date(v); return isNaN(d) ? '' : d.toLocaleDateString('es-PA', {day: 'numeric', month: 'short'}); };
  const when = v => { const d = new Date(v); return isNaN(d) ? '' : d.toLocaleString('es-PA', {day: 'numeric', month: 'short', hour: 'numeric', minute: '2-digit'}); };
  const ymd = d => `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
  const midnight = (s, add = 0) => { const [y, m, d] = s.split('-').map(Number); return new Date(y, m - 1, d + add); };

  // Atajos de período: [clave, texto, () => [desde, hasta]]; «Fechas» deja elegir de un día a otro.
  const PRESETS = [
    ['today', 'Hoy', () => { const a = new Date(); a.setHours(0, 0, 0, 0); return [a, new Date()]; }],
    ['yesterday', 'Ayer', () => { const b = new Date(); b.setHours(0, 0, 0, 0); return [new Date(b.getTime() - 864e5), b]; }],
    ['1', 'Últimas 24 horas', () => [new Date(Date.now() - 864e5), new Date()]],
    ['7', 'Últimos 7 días', () => [new Date(Date.now() - 7 * 864e5), new Date()]],
    ['30', 'Últimos 30 días', () => [new Date(Date.now() - 30 * 864e5), new Date()]],
    ['90', 'Últimos 3 meses', () => [new Date(Date.now() - 90 * 864e5), new Date()]],
  ];
  const periodText = (from, to, label) => {
    const a = new Date(from), b = new Date(to), sameDay = ymd(a) === ymd(new Date(b.getTime() - 1));
    const range = sameDay ? `${day(a)}` : `Del ${dayShort(a)} al ${day(new Date(b.getTime() - 1))}`;
    return label ? `${range} (${label.toLowerCase()})` : range;
  };

  // Cifras comunes a los dos formatos.
  function crunch(d) {
    const needs = d.needs || [], way = d.enroute || [], done = d.delivered || [];
    const urgent = needs.filter(n => n.urgency === 'urgente');
    const camps = new Map(); needs.forEach(n => camps.set(n.camp.id, n.camp));
    const people = [...camps.values()].reduce((t, c) => t + num(c.people), 0);
    const items = new Map();
    needs.forEach(n => new Set(String(n.need || '').split(/[,\n;]+/).map(x => x.trim()).filter(Boolean).map(x => x.replace(/\s*[x×]\s*\d+$/i, ''))).forEach(x => {
      const k = fold(x), g = items.get(k) || {name: x, n: 0, u: 0, people: 0, camps: new Set()};
      g.n++; if (n.urgency === 'urgente') g.u++;
      if (!g.camps.has(n.camp.id)) { g.camps.add(n.camp.id); g.people += num(n.camp.people); }
      items.set(k, g);
    }));
    const zones = new Map();
    needs.forEach(n => { const name = (n.camp.area || '').trim() || 'Sin zona', k = fold(name), z = zones.get(k) || {name, pts: new Set(), n: 0, u: 0}; z.pts.add(n.camp.id); z.n++; if (n.urgency === 'urgente') z.u++; zones.set(k, z); });
    const byRunner = new Map(), byDisp = new Map(), delivered = new Map();
    let units = 0;
    done.forEach(n => {
      const r = n.runnerName || 'Runner', w = n.dispatchedByName || 'Sin registrar';
      byRunner.set(r, (byRunner.get(r) || 0) + 1); byDisp.set(w, (byDisp.get(w) || 0) + 1);
      (n.packing || []).forEach(l => { const k = fold(l.item), g = delivered.get(k) || {name: l.item, q: 0}; g.q += num(l.quantity); units += num(l.quantity); delivered.set(k, g); });
    });
    const has = (n, st) => (n.media || []).some(m => m.stage === st);
    return {
      needs, way, done, urgent, camps, people, units, has,
      items: [...items.values()].sort((a, b) => b.n - a.n).slice(0, 15),
      zones: [...zones.values()].sort((a, b) => b.u - a.u || b.n - a.n),
      byRunner: [...byRunner].sort((a, b) => b[1] - a[1]), byDisp: [...byDisp].sort((a, b) => b[1] - a[1]),
      delivered: [...delivered.values()].sort((a, b) => b.q - a.q).slice(0, 20),
      sorted: [...needs].sort((a, b) => (b.urgency === 'urgente') - (a.urgency === 'urgente') || String(a.createdAt || '').localeCompare(String(b.createdAt || ''))),
      withPick: done.filter(n => has(n, 'pickup')).length, withDeliv: done.filter(n => has(n, 'delivery')).length,
    };
  }
  const place = n => n.camp.place || n.camp.area || 'Punto', zone = n => (n.camp.area || '').trim() || 'Sin zona';
  const scopeText = d => d.scope === 'team' ? '👥 Solo mi equipo' : '🌐 Todos los equipos';
  const table = (head, rows, empty, cls = '') => rows.length ? `<table class="${cls}"><thead><tr>${head.map(h => `<th>${h}</th>`).join('')}</tr></thead><tbody>${rows.join('')}</tbody></table>` : `<p class="empty">${empty}</p>`;
  const ck = on => on ? '<span class="ck on">☑</span>' : '<span class="ck">☐</span>';
  const needRows = c => c.sorted.slice(0, 300).map(n => `<tr><td>${n.urgency === 'urgente' ? '<span class="u">Urgente</span>' : ''}</td><td class="code">${esc(n.camp.code)}</td><td>${esc(place(n))}</td><td>${esc(zone(n))}</td><td>${esc(n.need)}</td><td>${esc(day(n.revisedAt || n.createdAt))}</td></tr>`);
  const wayRows = c => c.way.map(n => `<tr><td class="code">${esc(n.camp.code)}</td><td>${esc(place(n))}</td><td>${esc(zone(n))}</td><td>${esc(n.runnerName || 'Runner')}</td><td>${esc(n.dispatchedByName || '')}</td><td>${esc(when(n.dispatchedAt))}</td><td>${esc(when(n.enRouteAt))}</td><td class="c">${ck(c.has(n, 'pickup'))}</td></tr>`);
  const doneRows = c => c.done.slice(0, 300).map(n => `<tr><td class="code">${esc(n.camp.code)}</td><td>${esc(place(n))}</td><td>${esc(zone(n))}</td><td>${esc(n.runnerName || 'Runner')}</td><td>${esc(n.dispatchedByName || '')}</td><td>${esc(n.supplyCenter || '')}</td><td>${esc(when(n.dispatchedAt))}</td><td>${esc(when(n.completedAt))}</td><td class="c">${ck(c.has(n, 'pickup'))}</td><td class="c">${ck(c.has(n, 'delivery'))}</td></tr>`);
  const WAY_HEAD = ['Código', 'Punto', 'Zona', 'Runner', 'Despachó', 'Despachada', 'En camino desde', '📷 Recogida'];
  const DONE_HEAD = ['Código', 'Punto', 'Zona', 'Runner', 'Despachó', 'Centro', 'Despachada', 'Entregada', '📷 Recogida', '📷 Entrega'];
  const NEED_HEAD = ['', 'Código', 'Punto', 'Zona', 'Necesita', 'Desde'];
  const toolbar = '<div class="noprint"><button onclick="print()">🖨️ Imprimir / Guardar PDF</button><button class="ghost" onclick="close()">Cerrar</button></div>';
  const NOTE = 'Cifras reunidas por voluntarios: úsalas como guía, no como datos oficiales. El informe no incluye teléfonos ni contactos.';

  // 🧾 Sencillo: tablas.
  function simple(d, meta) {
    const c = crunch(d), maxN = Math.max(1, ...c.items.map(g => g.n)), period = periodText(d.from, d.to, meta.label);
    const box = (n, l, cls = '') => `<div class="kpi ${cls}"><b>${fmt(n)}</b><span>${l}</span></div>`;
    return `<!doctype html><html lang="es"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Informe de Supply · ${esc(d.team)} · ${esc(day(Date.now()))}</title><style>
*{box-sizing:border-box}body{margin:0;padding:24px;font:13px/1.45 Inter,system-ui,-apple-system,Segoe UI,sans-serif;color:#152b3a;background:#fff}h1{margin:0;font-size:22px}h2{margin:22px 0 8px;font-size:15px;padding-bottom:4px;border-bottom:2px solid #173c46;break-after:avoid}
.meta{display:grid;grid-template-columns:repeat(auto-fit,minmax(170px,1fr));gap:6px 16px;margin:10px 0 4px;padding:10px 12px;border:1px solid #cfdcdc;border-radius:10px;background:#f5f8f7}.meta small{display:block;color:#5d7680;font-size:11px;text-transform:uppercase;letter-spacing:.04em}.meta b{font-size:14px}
.kpis{display:grid;grid-template-columns:repeat(auto-fit,minmax(120px,1fr));gap:8px}.kpi{border:1px solid #cfdcdc;border-radius:10px;padding:10px;text-align:center}.kpi b{display:block;font-size:24px;line-height:1.1}.kpi span{font-size:11.5px;color:#4b6670;font-weight:700}.kpi.urg{border-color:#e57373}.kpi.urg b{color:#c62828}.kpi.go b{color:#19745f}
table{width:100%;border-collapse:collapse;font-size:12px}th,td{padding:5px 6px;border-bottom:1px solid #e1e9ea;text-align:left;vertical-align:top}th{background:#eef3f3;font-size:11px;text-transform:uppercase;letter-spacing:.03em;color:#36505a}tr{break-inside:avoid}td.n{white-space:nowrap;font-weight:800}td.code{font-weight:800;white-space:nowrap}td.c{text-align:center}.ck{font-size:15px;color:#9aa8ad}.ck.on{color:#19745f}
.bar{height:8px;border-radius:4px;background:#e6eeed;min-width:80px}.bar i{display:block;height:100%;border-radius:4px;background:#ed7654}.u{color:#c62828;font-weight:800}.empty{color:#667982}.note{margin-top:22px;font-size:11px;color:#667982}.sub{margin:10px 0 6px;color:#36505a}.three{display:grid;grid-template-columns:1fr 1fr 1fr;gap:16px}@media(max-width:640px){.three{grid-template-columns:1fr}body{padding:14px}}
.noprint{margin:0 0 14px;display:flex;gap:8px}.noprint button{border:0;border-radius:9px;padding:10px 14px;font:800 14px Inter,system-ui,sans-serif;background:#173c46;color:#fff;cursor:pointer}.noprint .ghost{background:#fff;color:#173c46;border:1px solid #cad9d9}@media print{.noprint{display:none}body{padding:0;font-size:11px}table{font-size:10.5px}@page{size:A4 landscape;margin:10mm}}
</style></head><body>${toolbar}
<h1>📦 Informe de Supply</h1>
<div class="meta"><div><small>Equipo</small><b>👥 ${esc(d.team)}</b></div><div><small>Centro de acopio</small><b>🏬 ${esc(d.center || '—')}</b></div><div><small>Alcance</small><b>${scopeText(d)}</b></div><div><small>Período (entregas)</small><b>${esc(period)}</b></div><div><small>Generado</small><b>${esc(when(Date.now()))}${d.by ? ' · ' + esc(d.by) : ''}</b></div></div>
<h2>Cifras</h2><div class="kpis">${box(c.needs.length, 'Necesidades')}${box(c.urgent.length, 'Urgentes', 'urg')}${box(c.camps.size, 'Puntos con necesidades')}${c.people ? box(c.people, 'Personas en esos puntos') : ''}${box(c.way.length, 'En camino')}${box(c.done.length, 'Entregadas en el período', 'go')}${c.units ? box(c.units, 'Artículos entregados', 'go') : ''}</div>
<h2>Qué se necesita</h2>${table(['Necesidad', 'Puntos', 'Urgentes', ''], c.items.map(g => `<tr><td>${esc(g.name)}</td><td class="n">${fmt(g.n)}</td><td class="n${g.u ? ' u' : ''}">${g.u ? fmt(g.u) : '—'}</td><td style="width:35%"><div class="bar"><i style="width:${(g.n / maxN * 100).toFixed(0)}%"></i></div></td></tr>`), 'No hay necesidades abiertas.')}
<h2>Por zona</h2>${table(['Zona', 'Puntos', 'Necesidades', 'Urgentes'], c.zones.map(z => `<tr><td>${esc(z.name)}</td><td class="n">${fmt(z.pts.size)}</td><td class="n">${fmt(z.n)}</td><td class="n${z.u ? ' u' : ''}">${z.u ? fmt(z.u) : '—'}</td></tr>`), 'Ninguna zona con necesidades abiertas.')}
<h2>Necesidades (${fmt(c.needs.length)})</h2>${table(NEED_HEAD, needRows(c), 'No hay necesidades abiertas.')}${c.sorted.length > 300 ? `<p class="empty">Y ${fmt(c.sorted.length - 300)} más.</p>` : ''}
<h2>En camino (${fmt(c.way.length)})</h2>${table(WAY_HEAD, wayRows(c), 'Nada en camino.')}
<h2>Entregas · ${esc(period)} (${fmt(c.done.length)})</h2><div class="three"><div>${table(['Runner', 'Entregas'], c.byRunner.map(([k, v]) => `<tr><td>${esc(k)}</td><td class="n">${fmt(v)}</td></tr>`), 'Sin entregas en el período.')}</div><div>${table(['Despachó', 'Entregas'], c.byDisp.map(([k, v]) => `<tr><td>${esc(k)}</td><td class="n">${fmt(v)}</td></tr>`), '—')}</div><div>${table(['Artículo', 'Cantidad'], c.delivered.map(g => `<tr><td>${esc(g.name)}</td><td class="n">${fmt(g.q)}</td></tr>`), 'Sin listas de empaque en el período.')}</div></div>
${c.done.length ? `<p class="sub">📷 Con fotos de recogida: <b>${fmt(c.withPick)}</b> de ${fmt(c.done.length)} · con fotos de entrega: <b>${fmt(c.withDeliv)}</b> de ${fmt(c.done.length)}</p>` : ''}${table(DONE_HEAD, doneRows(c), '')}
<p class="note">${NOTE}</p></body></html>`;
  }

  // 📰 Boletín: el mismo informe con el estilo del boletín del equipo. Cada estilo: nombre, organización, logo y colores.
  const STYLES = {
    unidos: {org: 'Unidos por Azuero', tagline: 'Registro ciudadano construido y alimentado por voluntarios.', logo: '/report/logo.png', navy: '#02377B', bright: '#048EFF', mid: '#1A85DF', coral: '#E0533D'},
  };
  function boletin(d, meta) {
    const st = STYLES[d.template] || STYLES.unidos, c = crunch(d), period = periodText(d.from, d.to, meta.label), origin = meta.origin || '';
    const maxN = Math.max(1, ...c.items.map(g => g.n)), maxP = Math.max(1, ...c.items.map(g => g.people)), maxZ = Math.max(1, ...c.zones.map(z => z.n));
    const lead = `<b>${fmt(c.needs.length)} ${c.needs.length === 1 ? 'necesidad abierta' : 'necesidades abiertas'}</b> en ${fmt(c.camps.size)} ${c.camps.size === 1 ? 'punto' : 'puntos'} de ${fmt(c.zones.length)} ${c.zones.length === 1 ? 'zona' : 'zonas'}${c.urgent.length ? `, <b>${fmt(c.urgent.length)} ${c.urgent.length === 1 ? 'urgente' : 'urgentes'}</b>` : ''}${c.people ? ` · ${fmt(c.people)} personas en esos puntos` : ''}. En el período se entregaron <b>${fmt(c.done.length)}</b>${c.units ? ` (${fmt(c.units)} artículos)` : ''} y van en camino <b>${fmt(c.way.length)}</b>.`;
    const card = (n, l, sub, cls = '') => `<div class="card ${cls}"><b>${fmt(n)}</b><span>${l}</span>${sub ? `<small>${sub}</small>` : ''}</div>`;
    const ticks = n => `<span class="ticks">${Array.from({length: 10}, (_, i) => `<i class="${i < Math.round(n / maxN * 10) ? 'on' : ''}"></i>`).join('')}</span>`;
    const head = (small) => `<header class="head${small ? ' small' : ''}"><div><div class="kicker">Informe de Supply · ${esc(st.org)} · ${esc(d.team)}</div><h1>Necesidades y entregas${d.center ? ` · ${esc(d.center)}` : ''}</h1>${small ? '' : `<div class="subtitle">${scopeText(d)} · ${esc(period)}</div>`}<div class="cut">Corte: ${esc(when(Date.now()))}${d.by ? ` · ${esc(d.by)}` : ''}${small ? '' : `<span class="pill">Período: ${esc(period)}</span>`}</div></div><div class="logo"><img src="${esc(origin + st.logo)}" alt="${esc(st.org)}"></div></header>`;
    const foot = `<footer class="pfoot"><div><b>${esc(st.org)}.</b> <span>${esc(st.tagline)}</span></div><span>${esc(NOTE)}</span></footer>`;
    return `<!doctype html><html lang="es"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Informe de Supply · ${esc(d.team)} · ${esc(day(Date.now()))}</title><style>
:root{--navy:${st.navy};--bright:${st.bright};--mid:${st.mid};--coral:${st.coral};--gray:#F2F4F6;--ink:#1d2b3a;--muted:#6b7785}
*{box-sizing:border-box;-webkit-print-color-adjust:exact;print-color-adjust:exact}body{margin:0;background:#c9d2db;color:var(--ink);font:13px/1.4 Poppins,"Segoe UI",system-ui,-apple-system,sans-serif}
.noprint{position:sticky;top:0;z-index:5;display:flex;gap:10px;justify-content:center;padding:10px;background:#10263a}.noprint button{border:0;border-radius:20px;padding:8px 14px;font:700 13px system-ui,sans-serif;background:var(--bright);color:#fff;cursor:pointer}.noprint .ghost{background:#fff;color:#10263a}
.page{width:8.5in;max-width:100%;min-height:11in;margin:18px auto;background:#fff;display:flex;flex-direction:column;box-shadow:0 6px 30px #0003;overflow:hidden}
.head{background:var(--navy);color:#fff;padding:22px 30pt 18px;display:flex;justify-content:space-between;gap:20px;border-bottom:6px solid var(--bright)}.kicker{font-size:10px;font-weight:700;letter-spacing:.08em;color:#a9c6ee;text-transform:uppercase}.head h1{font-size:26px;line-height:1.15;margin:6px 0 4px}.subtitle{font-size:13px;color:#d3e3f7}.cut{margin-top:12px;font-size:12.5px;display:flex;gap:12px;align-items:center;flex-wrap:wrap}.pill{background:var(--bright);border-radius:20px;padding:4px 12px;font-size:10px;font-weight:700;text-transform:uppercase;letter-spacing:.03em}.logo{background:#fff;border-radius:14px;padding:8px;width:118px;height:118px;flex:none;display:grid;place-items:center}.logo img{width:100%;height:auto}.head.small{padding:14px 30pt 12px}.head.small .logo{width:76px;height:76px}.head.small h1{font-size:20px}
.body{padding:16px 30pt 0;flex:1}h2{font-size:11px;letter-spacing:.06em;text-transform:uppercase;color:var(--navy);margin:16px 0 9px}.lead{font-size:13.5px;margin:0 0 14px}
.cards{display:grid;grid-template-columns:repeat(5,1fr);gap:9px}.card{background:var(--gray);border-radius:9px;padding:10px 12px;border-left:4px solid var(--navy)}.card b{display:block;font-size:28px;line-height:1.1;color:var(--navy)}.card span{display:block;font-size:11px}.card small{display:block;font-size:9.5px;color:var(--muted);margin-top:4px}.card.mid{border-color:var(--mid)}.card.mid b{color:var(--mid)}.card.coral{border-color:var(--coral)}.card.coral b{color:var(--coral)}.card.bright{border-color:var(--bright)}.card.bright b{color:var(--bright)}
.grid2{display:grid;grid-template-columns:1fr 1fr;gap:22px}.rows{display:grid;gap:5px}.row{display:grid;grid-template-columns:150px 66px 24px 1fr;gap:8px;align-items:center;font-size:12px}.row .nm{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.ticks{display:flex;gap:2px}.ticks i{width:4px;height:11px;border-radius:1px;background:#e2e6ea}.ticks i.on{background:var(--coral)}.row .ct{color:var(--coral);font-weight:800;text-align:right}.track{position:relative;height:13px;border-radius:7px;background:var(--gray)}.track i{display:block;height:100%;border-radius:7px;background:var(--mid)}.track{background:none}.track i{background:var(--mid)}.track em{position:absolute;top:-1px;font-size:10.5px;font-style:normal;font-weight:800;color:var(--ink);white-space:nowrap}
.zrow{display:grid;grid-template-columns:130px 1fr 54px;gap:8px;align-items:center;font-size:12px}.ztrack{display:flex;height:16px;border-radius:4px;overflow:hidden;background:var(--gray)}.ztrack i{background:var(--mid)}.ztrack i.u{background:var(--coral)}.zrow b{font-size:12px}.legend{display:flex;gap:16px;margin-top:8px;font-size:11px}.legend i{display:inline-block;width:11px;height:11px;border-radius:2px;margin-right:5px;vertical-align:-1px;background:var(--mid)}.legend i.u{background:var(--coral)}
.prio{margin:18px 0 0;padding:12px 16px;border-radius:10px;background:#fdeceb;border-left:6px solid var(--coral)}.prio h2{color:var(--coral);margin-top:2px}.pr{display:grid;grid-template-columns:62px 110px 1fr 90px;gap:10px;align-items:start;padding:6px 0;font-size:12px}.pr .code{display:block;text-align:center;border-radius:6px;padding:3px 0;background:var(--coral);color:#fff;font-weight:800}.pr .tag{text-align:right;font-weight:800;color:var(--coral);font-size:11px}.pr .tag small{display:block;color:var(--muted);font-weight:600}
table{width:100%;border-collapse:collapse;font-size:11px}th,td{padding:5px 6px;border-bottom:1px solid #e6eaee;text-align:left;vertical-align:top}th{color:var(--navy);font-size:9.5px;text-transform:uppercase;letter-spacing:.05em}td.code{font-weight:800;color:var(--navy);white-space:nowrap}td.c{text-align:center}td.n{font-weight:800}.ck{font-size:14px;color:#a3adb6}.ck.on{color:var(--mid)}.u{color:var(--coral);font-weight:800}.empty{color:var(--muted);font-size:12px}.three{display:grid;grid-template-columns:1fr 1fr 1fr;gap:16px}
.pfoot{display:flex;justify-content:space-between;gap:16px;padding:10px 30pt 14px;margin-top:18px;border-top:1px solid #e3e7eb;font-size:10px;color:var(--muted)}.pfoot b{color:var(--navy)}.page::after{content:"";display:block;height:6px;background:linear-gradient(90deg,var(--navy) 50%,var(--bright) 50%)}
@media(max-width:820px){.cards{grid-template-columns:repeat(2,1fr)}.grid2,.three{grid-template-columns:1fr}.row{grid-template-columns:110px 60px 22px 1fr}.pr{grid-template-columns:56px 1fr}.pr .tag{text-align:left}}
@media print{body{background:#fff}.noprint{display:none}.page{margin:0;box-shadow:none;width:auto;min-height:0;break-after:page}@page{size:letter;margin:0}}
</style></head><body>${toolbar}
<section class="page">${head(false)}<div class="body"><p class="lead">${lead}</p>
<div class="cards">${card(c.needs.length, 'necesidades', `en ${fmt(c.camps.size)} puntos`)}${card(c.urgent.length, 'urgentes', 'por llevar', 'coral')}${card(c.people, 'personas', 'en esos puntos', 'mid')}${card(c.way.length, 'en camino', 'despachadas')}${card(c.done.length, 'entregadas', c.units ? `${fmt(c.units)} artículos` : 'en el período', 'bright')}</div>
<div class="grid2"><div><h2>Qué se necesita · puntos y personas</h2>${c.items.length ? `<div class="rows">${c.items.map(g => `<div class="row"><span class="nm">${esc(g.name)}</span>${ticks(g.n)}<span class="ct">${fmt(g.n)}</span><span class="track"><i style="width:${(Math.max(4, g.people / maxP * 50)).toFixed(0)}%"></i>${g.people ? `<em style="left:calc(${(Math.max(4, g.people / maxP * 50)).toFixed(0)}% + 6px)">${fmt(g.people)} personas</em>` : ''}</span></div>`).join('')}</div>` : '<p class="empty">No hay necesidades abiertas.</p>'}</div>
<div><h2>Por zona · necesidades y urgentes</h2>${c.zones.length ? `<div class="rows">${c.zones.slice(0, 14).map(z => `<div class="zrow"><span class="nm">${esc(z.name)}</span><span class="ztrack"><i class="u" style="width:${(z.u / maxZ * 100).toFixed(0)}%"></i><i style="width:${((z.n - z.u) / maxZ * 100).toFixed(0)}%"></i></span><b>${fmt(z.n)}</b></div>`).join('')}</div><div class="legend"><span><i class="u"></i>Urgentes</span><span><i></i>Por llevar</span></div>` : '<p class="empty">Ninguna zona con necesidades abiertas.</p>'}</div></div>
${c.urgent.length ? `<div class="prio"><h2>Prioridad: urgentes</h2>${c.urgent.slice(0, 10).map(n => `<div class="pr"><span class="code">${esc(n.camp.code || 's/c')}</span><b>${esc(place(n))}<br><small style="font-weight:600;color:var(--muted)">${esc(zone(n))}</small></b><span>${esc(n.need)}</span><span class="tag">URGENTE<small>desde ${esc(dayShort(n.revisedAt || n.createdAt))}</small></span></div>`).join('')}${c.urgent.length > 10 ? `<p class="empty">Y ${fmt(c.urgent.length - 10)} urgentes más (lista completa en la página siguiente).</p>` : ''}</div>` : ''}
</div>${foot}</section>
<section class="page">${head(true)}<div class="body">
<h2>Necesidades (${fmt(c.needs.length)})</h2>${table(NEED_HEAD, needRows(c), 'No hay necesidades abiertas.')}
<h2>En camino (${fmt(c.way.length)})</h2>${table(WAY_HEAD, wayRows(c), 'Nada en camino.')}
<h2>Entregas · ${esc(period)} (${fmt(c.done.length)})</h2><div class="three"><div>${table(['Runner', 'Entregas'], c.byRunner.map(([k, v]) => `<tr><td>${esc(k)}</td><td class="n">${fmt(v)}</td></tr>`), 'Sin entregas en el período.')}</div><div>${table(['Despachó', 'Entregas'], c.byDisp.map(([k, v]) => `<tr><td>${esc(k)}</td><td class="n">${fmt(v)}</td></tr>`), '—')}</div><div>${table(['Artículo', 'Cantidad'], c.delivered.map(g => `<tr><td>${esc(g.name)}</td><td class="n">${fmt(g.q)}</td></tr>`), 'Sin listas de empaque.')}</div></div>
${c.done.length ? `<p class="lead" style="margin-top:12px">📷 Con fotos de recogida: <b>${fmt(c.withPick)}</b> de ${fmt(c.done.length)} · con fotos de entrega: <b>${fmt(c.withDeliv)}</b> de ${fmt(c.done.length)}</p>` : ''}${table(DONE_HEAD, doneRows(c), '')}
</div>${foot}</section></body></html>`;
  }

  // El formulario (en un cuadro desde Supply, o dentro de la página «📰 Reportes»).
  const FORM_CSS = '.rpf{display:grid;gap:14px;font-family:Inter,system-ui,sans-serif;color:#152b3a}.rpf h3{margin:0 0 6px;font-size:12px;letter-spacing:.05em;text-transform:uppercase;color:#5d7680}.rpfmt{display:grid;grid-template-columns:1fr 1fr;gap:8px}.rpfmt button{display:grid;gap:3px;text-align:left;padding:12px;border:2px solid #dfe8e7;border-radius:12px;background:#fff;cursor:pointer;font:inherit}.rpfmt button b{font-size:15px}.rpfmt button small{color:#5d7680;font-size:12px;line-height:1.3}.rpfmt button[aria-pressed="true"]{border-color:#19745f;background:#eef6f2}.rpchips{display:flex;flex-wrap:wrap;gap:6px}.rpchips button{border:1px solid #cad9d9;border-radius:999px;background:#fff;padding:7px 12px;font:700 13px Inter,system-ui,sans-serif;color:#36505a;cursor:pointer}.rpchips button[aria-pressed="true"]{background:#173c46;border-color:#173c46;color:#fff}.rpdates{display:flex;gap:8px;flex-wrap:wrap;margin-top:8px}.rpdates label{display:grid;gap:3px;font-size:12px;font-weight:800;color:#5d7680;flex:1;min-width:130px}.rpdates input{padding:9px;border:1px solid #cad9d9;border-radius:9px;font:inherit;font-size:15px}.rpsum{margin:0;padding:9px 11px;border-radius:10px;background:#f5f8f7;font-size:13px;color:#36505a}.rpgo{border:0;border-radius:12px;padding:13px;background:#19745f;color:#fff;font:800 15px Inter,system-ui,sans-serif;cursor:pointer}.rpmsg{margin:0;color:#9a4536;font-size:13px;font-weight:700}.rpmsg:empty{display:none}.rpdlg{border:0;border-radius:16px;padding:18px;width:min(520px,calc(94vw / var(--text-zoom,1)));max-height:calc(92dvh / var(--text-zoom,1));overflow:auto}.rpdlg::backdrop{background:#10263ab8}.rptop{display:flex;align-items:center;gap:8px;margin-bottom:10px}.rptop b{flex:1;font-size:18px}.rptop button{border:0;background:none;font-size:20px;cursor:pointer;color:#4b6670}@media(max-width:480px){.rpfmt{grid-template-columns:1fr}}';
  function form(mount, opts = {}) {
    if (!document.getElementById('rpCss')) { const s = document.createElement('style'); s.id = 'rpCss'; s.textContent = FORM_CSS; document.head.appendChild(s); }
    const st = {fmt: opts.format || 'simple', preset: String(opts.preset || '7'), scope: opts.teamOnly ? 'team' : 'all', from: '', to: ''};
    const setPreset = k => { const p = PRESETS.find(x => x[0] === k); if (!p) return; st.preset = k; const [a, b] = p[1] ? p[2]() : []; st.range = [a, b]; st.from = ymd(a); st.to = ymd(new Date(b.getTime() - 1)); };
    setPreset(st.preset);
    // onlySimple (Supply): solo «🧾 Sencillo», sin elegir formato.
    if (opts.onlySimple) st.fmt = 'simple';
    mount.innerHTML = `<div class="rpf"><div${opts.onlySimple ? ' hidden' : ''}><h3>Formato</h3><div class="rpfmt" role="group" aria-label="Formato"><button type="button" data-rp-fmt="simple"><b>🧾 Sencillo</b><small>Tablas claras para imprimir.</small></button><button type="button" data-rp-fmt="boletin"><b>📰 Boletín</b><small>Con el estilo del boletín de tu equipo: tarjetas, barras y prioridades.</small></button></div></div>
<div><h3>Período de las entregas</h3><div class="rpchips" role="group" aria-label="Período">${PRESETS.map(([k, l]) => `<button type="button" data-rp-preset="${k}">${l}</button>`).join('')}<button type="button" data-rp-preset="custom">📅 Fechas</button></div><div class="rpdates"><label>Desde<input type="date" data-rp-from></label><label>Hasta<input type="date" data-rp-to></label></div></div>
<div><h3>Alcance</h3><div class="rpchips" role="group" aria-label="Alcance"><button type="button" data-rp-scope="team">👥 Solo mi equipo</button><button type="button" data-rp-scope="all">🌐 Todos los equipos</button></div></div>
<p class="rpsum" data-rp-sum></p><p class="rpmsg" data-rp-msg role="status"></p><button type="button" class="rpgo" data-rp-go>🖨️ Ver informe para imprimir o guardar en PDF</button></div>`;
    const $ = s => mount.querySelector(s), today = ymd(new Date());
    $('[data-rp-from]').max = $('[data-rp-to]').max = today;
    const label = () => st.preset === 'custom' ? '' : (PRESETS.find(p => p[0] === st.preset) || [, ''])[1];
    const range = () => st.preset === 'custom' || !st.range ? [midnight(st.from), midnight(st.to, 1)] : st.range;
    const sync = () => {
      mount.querySelectorAll('[data-rp-fmt]').forEach(b => b.setAttribute('aria-pressed', String(b.dataset.rpFmt === st.fmt)));
      mount.querySelectorAll('[data-rp-preset]').forEach(b => b.setAttribute('aria-pressed', String(b.dataset.rpPreset === st.preset)));
      mount.querySelectorAll('[data-rp-scope]').forEach(b => b.setAttribute('aria-pressed', String(b.dataset.rpScope === st.scope)));
      $('[data-rp-from]').value = st.from; $('[data-rp-to]').value = st.to;
      const [a, b] = range();
      $('[data-rp-sum]').textContent = b > a ? `${opts.onlySimple ? '' : (st.fmt === 'boletin' ? '📰 Boletín' : '🧾 Sencillo') + ' · '}${periodText(a, b, label())} · ${st.scope === 'team' ? '👥 Solo mi equipo' : '🌐 Todos los equipos'}` : 'Elige «Desde» antes de «Hasta».';
    };
    mount.addEventListener('click', e => {
      const f = e.target.closest('[data-rp-fmt]'), p = e.target.closest('[data-rp-preset]'), s = e.target.closest('[data-rp-scope]');
      if (f) st.fmt = f.dataset.rpFmt;
      if (p) { if (p.dataset.rpPreset === 'custom') { st.preset = 'custom'; $('[data-rp-from]').focus(); } else setPreset(p.dataset.rpPreset); }
      if (s) st.scope = s.dataset.rpScope;
      if (f || p || s) sync();
      if (e.target.closest('[data-rp-go]')) go();
    });
    mount.addEventListener('change', e => { if (e.target.matches('[data-rp-from],[data-rp-to]')) { st.preset = 'custom'; st.from = $('[data-rp-from]').value; st.to = $('[data-rp-to]').value || st.from; if (st.to < st.from) st.to = st.from; sync(); } });
    async function go() {
      const [a, b] = range(), msg = $('[data-rp-msg]');
      if (!(b > a)) { msg.textContent = 'Revisa las fechas.'; return; }
      const w = window.open('', '_blank');
      if (!w) { msg.textContent = 'Permite ventanas emergentes para ver el informe.'; return; }
      w.document.write('<p style="font-family:sans-serif;padding:20px">Preparando el informe…</p>');
      msg.textContent = '';
      try {
        const q = `from=${encodeURIComponent(a.toISOString())}&to=${encodeURIComponent(b.toISOString())}&scope=${st.scope}`;
        const r = await fetch('/api/full/reports/supply?' + q, {credentials: 'same-origin', cache: 'no-store'});
        const d = await r.json().catch(() => ({}));
        if (!r.ok) throw Error(d.error || 'No se pudo preparar el informe.');
        const html = (st.fmt === 'boletin' ? boletin : simple)(d, {label: label(), origin: location.origin});
        w.document.open(); w.document.write(html); w.document.close();
        setTimeout(() => { try { w.focus(); w.print(); } catch {} }, 700);
        opts.onDone?.();
      } catch (err) { w.close(); msg.textContent = err.message; }
    }
    sync();
  }
  function dialog(opts = {}) {
    let d = document.getElementById('rpDlg');
    if (!d) { document.body.insertAdjacentHTML('beforeend', '<dialog id="rpDlg" class="rpdlg" aria-labelledby="rpTitle"><div class="rptop"><b id="rpTitle">🖨️ Informe</b><button type="button" data-rp-close aria-label="Cerrar">✕</button></div><div data-rp-mount></div></dialog>'); d = document.getElementById('rpDlg'); d.addEventListener('click', e => { if (e.target === d || e.target.closest('[data-rp-close]')) d.close(); }); }
    form(d.querySelector('[data-rp-mount]'), {...opts, onDone: () => d.close()});
    d.showModal();
  }
  const api = {form, dialog, simple, boletin, periodText, PRESETS};
  if (typeof window !== 'undefined') window.Reports = api;
  if (typeof module !== 'undefined') module.exports = api;
})();
