// Navegación superior entre páginas para quien entró con código: Inicio (el mapa público) y los paneles de su usuario, según el
// registro «codeUser» de este navegador (lo guardan la página pública y cada panel al entrar). La página actual queda
// marcada. Sin código no se muestra nada: el público y quien entró con nombre y teléfono siguen igual.
// Cada página pone <nav class="appnav" id="appNav" hidden></nav> en su encabezado y llama AppNav.draw() al entrar o salir.
// En pantalla angosta (≤760px) se pliega: un botón con la página actual abre la lista completa bajo el encabezado.
(() => {
  const css = `.appnav{position:relative;display:flex;align-items:center;min-width:0}.appnav[hidden]{display:none}
.appnav-list{display:flex;align-items:center;gap:2px;background:#ffffff12;border-radius:10px;padding:3px}
.appnav-list a{display:inline-flex;align-items:center;gap:6px;color:#c8d6db;text-decoration:none;font-size:13px;font-weight:750;padding:7px 11px;border-radius:8px;white-space:nowrap}.appnav-list a:hover{background:#ffffff1a;color:#fff}
.appnav-list a[aria-current="page"]{background:#fff;color:#10263a}.top>.appnav{margin:0 auto 0 6px}
.appnav-toggle{display:none}
@media(max-width:760px){.top{position:relative}.appnav{position:static}
.appnav-toggle{display:inline-flex;align-items:center;gap:7px;border:1px solid #47657a;background:#ffffff12;color:#fff;border-radius:9px;padding:7px 10px;font-size:13px;font-weight:800;white-space:nowrap;cursor:pointer}
.appnav-toggle .chev{font-size:10px;transition:transform .15s}.appnav.open .appnav-toggle{background:#fff;color:#10263a;border-color:#fff}.appnav.open .chev{transform:rotate(180deg)}
.appnav-list{display:none;position:absolute;top:100%;left:0;right:0;z-index:3000;flex-direction:column;align-items:stretch;gap:2px;background:#10263a;border-radius:0 0 14px 14px;padding:6px 10px 12px;box-shadow:0 14px 30px #0007}
.appnav.open .appnav-list{display:flex}.appnav-list a{padding:13px 12px;font-size:15px;border-radius:10px;color:#e3ecef}.appnav-list a:active{background:#ffffff26}}
html:has(dialog[open]),html:has(.modalbg.open){overflow:hidden}html:has(dialog[open]) body,html:has(.modalbg.open) body{overflow:hidden}dialog[open],.modalbg.open,.modalbg.open .modal{overscroll-behavior:contain}
.telx{position:relative;display:inline-flex;align-items:center;gap:4px;vertical-align:middle;white-space:nowrap}.telnum{border:0;background:none;padding:0;color:#1d6f7a;font:inherit;font-weight:700;cursor:copy;text-decoration:underline dotted;text-underline-offset:3px}.telnum:hover{color:#10263a}
.telico{display:inline-grid;place-items:center;width:26px;height:26px;border-radius:50%;border:1px solid #cad9d9;background:#fff;color:#173c46;text-decoration:none}.telico svg{width:13px;height:13px}.telico:hover{filter:brightness(.95)}
.teltip{position:absolute;left:0;bottom:calc(100% + 6px);z-index:3200;background:#10263a;color:#fff;border-radius:6px;padding:4px 8px;font:700 11px Inter,system-ui,sans-serif;white-space:nowrap;pointer-events:none}
.appconfirm{border:0;border-radius:16px;padding:0;width:min(380px,calc(92vw / var(--text-zoom,1)));box-shadow:0 20px 60px #10263a66;font-family:Inter,system-ui,sans-serif}.appconfirm::backdrop{background:#10263ab8}.appconfirm form{margin:0;padding:22px}.appconfirm h2{margin:0;font-size:19px;line-height:1.3;color:#152b3a}.appconfirm p{margin:8px 0 0;color:#667982;font-size:13px;line-height:1.45}.appconfirm p:empty{display:none}
.textsize{position:relative;margin-left:8px;flex:none}.textsize>button{height:28px;padding:0 8px;border:1px solid #47657a;border-radius:8px;background:none;color:#c8d6db;font:800 13px Inter,system-ui,sans-serif;cursor:pointer}.textsize>button:hover,.textsize>button[aria-expanded="true"]{background:#ffffff1a;color:#fff}
.textsize-pop{position:absolute;top:calc(100% + 8px);left:0;z-index:3100;display:flex;align-items:center;gap:6px;background:#fff;border-radius:12px;padding:7px;box-shadow:0 10px 26px #0005}.textsize-pop[hidden]{display:none}.textsize-pop button{width:40px;height:36px;border:1px solid #cad9d9;border-radius:9px;background:#fff;color:#173c46;font:800 15px Inter,system-ui,sans-serif;cursor:pointer}.textsize-pop button:disabled{opacity:.4;cursor:default}.textsize-pop span{min-width:46px;text-align:center;font:750 13px Inter,system-ui,sans-serif;color:#36505a}
html{zoom:var(--text-zoom,1)}.leaflet-container{zoom:calc(1 / var(--text-zoom,1))}
.mbar,.msheet{display:none}
@media(max-width:760px){.top:has(>.mbar)>.appnav,.top:has(>.mbar) .textsize,.top:has(>.mbar)>.topright{display:none!important}.top:has(>.mbar) .brand{display:none!important}.top:has(>.mbar){gap:10px}
.mbar{display:flex;align-items:center;gap:8px;flex:1;min-width:0}.mlinks{flex:1;min-width:0;display:flex;gap:3px;overflow:hidden;background:#ffffff12;border-radius:11px;padding:3px}.mlinks a{flex:none;display:inline-flex;align-items:center;gap:6px;padding:8px 10px;border-radius:9px;color:#c8d6db;text-decoration:none;font:800 14px Inter,system-ui,sans-serif;white-space:nowrap}.mlinks a[aria-current="page"]{background:#fff;color:#10263a}.mlinks .mi{display:inline-flex}.mbar .mtitle{display:none}.mbar.compact .mlinks a:not([aria-current]) .ml{display:none}.mbar.tight .mlinks{display:none}.mbar.tight .mtitle{display:block}.mtitle{flex:1;min-width:0;color:#fff;font:800 17px Inter,system-ui,sans-serif;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.madd,.mmenu,.msearchbtn{flex:none;display:grid;place-items:center;width:42px;height:42px;border-radius:12px;border:0;cursor:pointer;text-decoration:none}.madd{background:#ed7654;color:#fff}.madd[hidden]{display:none}.mmenu,.msearchbtn{background:#ffffff14;border:1px solid #47657a;color:#fff}.msearchbtn[aria-expanded="true"]{background:#fff;color:#10263a}.madd.text{width:auto;padding:0 14px;font:800 15px Inter,system-ui,sans-serif;white-space:nowrap}.top:has(>.mbar){flex-wrap:wrap;height:auto;min-height:62px;align-content:center}.top:has(>.mbar) .topsearch{display:none!important}.top.msearch .topsearch{display:block!important;flex:1 0 100%;order:9;max-width:none;margin:0 0 10px}body:has(.top>.mbar) .addbar{display:none!important}.msuser .msoft{color:#c8d6db}.mlogin{display:none}.mbar.public .mlinks,.mbar.public .mtitle,.mbar.public .msearchbtn{display:none!important}.mbar.public .madd{order:-1}.mbar.public .mlogin{display:inline-flex;align-items:center;height:42px;margin-left:auto;padding:0 14px;border:1px solid #ffffff55;border-radius:12px;background:none;color:#fff;font:800 15px Inter,system-ui,sans-serif;cursor:pointer}.top.msearch:has(.mbar.public) .topsearch{display:none!important}.msuser .mgo{background:#ed7654;border-color:#ed7654;color:#fff}.mmenu[aria-expanded="true"]{background:#fff;color:#10263a}
.msheet{position:absolute;top:100%;left:0;right:0;z-index:3000;background:#10263a;padding:6px 14px 16px;box-shadow:0 16px 34px #0008;border-radius:0 0 16px 16px;font-family:Inter,system-ui,sans-serif}.msheet:not([hidden]){display:block}
.msuser{display:flex;align-items:center;gap:10px;padding:12px 4px;border-bottom:1px solid #ffffff1f;color:#fff}.msuser b{flex:1;min-width:0;font-size:15px;overflow-wrap:anywhere}.msuser button{border:1px solid #ffffff40;background:none;color:#ffd0c2;border-radius:9px;padding:8px 12px;font:800 13px Inter,system-ui,sans-serif;cursor:pointer}
.mpages{display:flex;flex-direction:column;gap:2px;padding:8px 0;border-bottom:1px solid #ffffff1f}.mpages a{display:flex;align-items:center;gap:10px;padding:12px 10px;border-radius:10px;color:#e3ecef;text-decoration:none;font-size:15px;font-weight:750}.mpages a[aria-current="page"]{background:#fff;color:#10263a}.mpages a:active{background:#ffffff26}
.msize{display:flex;align-items:center;gap:8px;padding:12px 4px 2px;color:#c8d6db;font-size:14px;font-weight:700}.msize .lbl{flex:1}.msize button{width:44px;height:38px;border:1px solid #47657a;border-radius:9px;background:none;color:#fff;font:800 15px Inter,system-ui,sans-serif;cursor:pointer}.msize button:disabled{opacity:.4}.msize output{min-width:46px;text-align:center;color:#fff}}
.gate:has(form[data-signin]){overflow-y:auto;place-items:start center}.gate form[data-signin]{margin:auto 0}.sg{border:1px solid #dbe5e5;border-radius:12px;margin:12px 0 0;padding:4px 14px 14px;min-width:0}.sg legend{padding:0 6px;font:800 14px Inter,system-ui,sans-serif;color:#173c46}.sg label{display:block}.sg .sgbtn{display:block;width:100%;margin-top:12px;border:0;border-radius:10px;padding:12px 16px;font:800 15px Inter,system-ui,sans-serif;cursor:pointer;color:#fff;background:#19745f}.sg-new .sgbtn{background:#ed7654}.sg small{font-weight:600;color:#7a8d94}.sgor{margin:10px 0 0;text-align:center;font-size:12px;font-weight:700;color:#8a9aa0}.remember{display:flex!important;align-items:center;gap:8px;margin:10px 0 2px;font-size:13px;font-weight:650;color:#36505a;cursor:pointer}.remember[hidden]{display:none!important}.remember input{width:20px;height:20px;margin:0;accent-color:#19745f;flex:none}
.appconfirm .acts{display:flex;gap:10px;margin-top:18px}.appconfirm button{flex:1;border:0;border-radius:10px;padding:13px 16px;font:inherit;font-size:15px;font-weight:800;cursor:pointer}.appconfirm .no{background:#eef3f3;color:#173c46}.appconfirm .yes{background:#9a4536;color:#fff}.appchoose .acts{flex-direction:column}.appconfirm .yes.ok{background:#19745f}.appconfirm .yes.go{background:#1f63c6}`;
  const style = document.createElement('style');
  style.textContent = css;
  document.head.appendChild(style);
  const panelUrl = {runner: '/runner.html', builder: '/builder.html', agent: '/command.html', donations: '/donations.html', supplier: '/supplier.html'};
  const has = (u, role) => (u.roles || []).includes(role) || (u.panels || []).some(p => p.url === panelUrl[role]);
  // Agentes sin rol de runner entran al panel de runners solo para las listas de empaque.
  const pages = [
    ['/', '🏠', () => 'Inicio', () => true],
    ['/runner.html', '<svg class="walkico" style="color:#5b9be6" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="14" cy="4" r="2"/><path d="M12.5 7.5 10 13"/><path d="m12 8 4 2.5 3-1"/><path d="m12 8-4 1.5-2 2.5"/><path d="m10 13 3.5 3-1 5"/><path d="m10 13-2.5 4H4"/></svg>', u => !u.roles || u.roles.includes('runner') ? 'Runners' : 'Empaque', u => has(u, 'runner') || has(u, 'agent')],
    ['/builder.html', '🔨', () => 'Constructores', u => has(u, 'builder')],
    ['/command.html', '🎧', () => 'Comando', u => has(u, 'agent')],
    ['/supplier.html', '📦', () => 'Supply', u => has(u, 'supplier')],
    ['/donations.html', '🎁', () => 'Donaciones', u => has(u, 'donations')],
  ];
  function user() {
    try { return JSON.parse(localStorage.getItem('codeUser') || 'null'); } catch { return null; }
  }
  const esc = s => String(s).replace(/[&<>"']/g, c => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
  function draw() {
    const el = document.getElementById('appNav');
    if (!el) return;
    const u = user(), here = location.pathname === '/index.html' ? '/' : location.pathname;
    const list = u ? pages.filter(([url, , , ok]) => url === here || ok(u)) : [];
    const current = list.find(([url]) => url === here) || list[0];
    el.hidden = list.length < 2;
    el.classList.remove('open');
    el.setAttribute('aria-label', 'Páginas');
    el.innerHTML = list.length < 2 ? '' : `<button type="button" class="appnav-toggle" aria-expanded="false" aria-controls="appNavList" aria-label="Cambiar de página"><span aria-hidden="true">${current[1]}</span>${esc(current[2](u))}<span class="chev" aria-hidden="true">▼</span></button><div class="appnav-list" id="appNavList">${list.map(([url, icon, label]) => `<a href="${url}"${url === here ? ' aria-current="page"' : ''}><span aria-hidden="true">${icon}</span><span class="appnav-label">${esc(label(u))}</span></a>`).join('')}</div>`;
  }
  function setOpen(open) {
    const el = document.getElementById('appNav'), b = el?.querySelector('.appnav-toggle');
    if (!b) return;
    el.classList.toggle('open', open);
    b.setAttribute('aria-expanded', String(open));
  }
  // Abrir y cerrar el menú plegado: el botón lo alterna; tocar fuera, Escape o elegir una página lo cierra.
  document.addEventListener('click', e => {
    const el = document.getElementById('appNav');
    if (!el) return;
    if (e.target.closest('.appnav-toggle')) setOpen(!el.classList.contains('open'));
    else if (!el.contains(e.target) || e.target.closest('.appnav-list a')) setOpen(false);
  });
  document.addEventListener('keydown', e => { if (e.key === 'Escape') setOpen(false); });
  matchMedia('(max-width:760px)').addEventListener('change', () => setOpen(false));
  // Confirmación Sí / No con el estilo de las páginas (p. ej. antes de salir): resuelve true solo con «Sí»; «No» o Esc dan false.
  function confirm(text, {yes = 'Sí', no = 'No', detail = ''} = {}) {
    let d = document.getElementById('appConfirm');
    if (!d) {
      d = document.createElement('dialog');
      d.id = 'appConfirm';
      d.className = 'appconfirm';
      d.innerHTML = '<form method="dialog"><h2></h2><p></p><div class="acts"><button class="no" value="no">No</button><button class="yes" value="yes">Sí</button></div></form>';
      document.body.appendChild(d);
    }
    d.querySelector('h2').textContent = text;
    d.querySelector('p').textContent = detail;
    d.querySelector('.yes').textContent = yes;
    d.querySelector('.no').textContent = no;
    d.returnValue = '';
    d.showModal();
    d.querySelector('.no').focus();
    return new Promise(resolve => d.addEventListener('close', () => resolve(d.returnValue === 'yes'), {once: true}));
  }
  // Tamaño del texto (accesibilidad): «Aa» junto al logo abre A− / A+. Escala toda la página menos el mapa (que se
  // compensa para que siga respondiendo bien al tocarlo). La elección queda en este navegador.
  const sizes = [0.9, 1, 1.15, 1.3, 1.5];
  const savedSize = () => { try { return Number(localStorage.getItem('textZoom')) || 1; } catch { return 1; } };
  let zoom = sizes.includes(savedSize()) ? savedSize() : 1;
  function applySize() {
    document.documentElement.style.setProperty('--text-zoom', String(zoom));
    const pop = document.getElementById('textSizePop');
    if (!pop) return;
    pop.querySelector('span').textContent = Math.round(zoom * 100) + '%';
    pop.querySelector('[data-size="-1"]').disabled = zoom === sizes[0];
    pop.querySelector('[data-size="1"]').disabled = zoom === sizes[sizes.length - 1];
    fitMobileBar();
  }
  function mountTextSize() {
    const brand = document.querySelector('.top .brand');
    if (brand && !document.getElementById('textSizeBtn')) {
      const box = document.createElement('div');
      box.className = 'textsize';
      box.innerHTML = '<button type="button" id="textSizeBtn" aria-expanded="false" aria-controls="textSizePop" title="Tamaño del texto" aria-label="Tamaño del texto">Aa</button><div class="textsize-pop" id="textSizePop" hidden><button type="button" data-size="-1" aria-label="Texto más pequeño">A−</button><span role="status"></span><button type="button" data-size="1" aria-label="Texto más grande">A+</button></div>';
      brand.appendChild(box);
      const btn = box.querySelector('#textSizeBtn'), pop = box.querySelector('#textSizePop');
      const show = open => { pop.hidden = !open; btn.setAttribute('aria-expanded', String(open)); };
      btn.onclick = () => show(pop.hidden);
      pop.onclick = e => {
        const b = e.target.closest('[data-size]');
        if (!b) return;
        zoom = sizes[Math.min(sizes.length - 1, Math.max(0, sizes.indexOf(zoom) + Number(b.dataset.size)))];
        try { localStorage.setItem('textZoom', String(zoom)); } catch {}
        applySize();
      };
      document.addEventListener('click', e => { if (!box.contains(e.target)) show(false); });
      document.addEventListener('keydown', e => { if (e.key === 'Escape') show(false); });
    }
    applySize();
  }
  applySize();
  // Teléfono en cualquier página: tocar el número lo copia (aviso «Copiado»); al lado, un botón con el ícono de teléfono que abre WhatsApp
  // para escribirle (no hace llamada).
  // Sin código de país se asume Panamá (507), como en administración.
  const waNumber = t => { const d = String(t).replace(/\D/g, ''); return String(t).trim().startsWith('+') || d.length > 8 ? d : '507' + d; };
  const callIco = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M22 16.9v3a2 2 0 0 1-2.2 2 19.8 19.8 0 0 1-8.6-3.1 19.5 19.5 0 0 1-6-6A19.8 19.8 0 0 1 2.1 4.2 2 2 0 0 1 4.1 2h3a2 2 0 0 1 2 1.7c.1.9.4 1.8.7 2.8a2 2 0 0 1-.5 2.1L8.1 9.9a16 16 0 0 0 6 6l1.3-1.3a2 2 0 0 1 2.1-.4c.9.3 1.8.6 2.8.7a2 2 0 0 1 1.7 2z"/></svg>';
  function phone(raw) {
    const p = String(raw || '').trim(), dial = p.replace(/[^\d+]/g, '');
    if (!dial) return esc(p);
    return `<span class="telx"><button type="button" class="telnum" data-copy-tel="${esc(p)}" title="Copiar número">${esc(p)}</button><a class="telico call" href="https://wa.me/${waNumber(p)}" target="_blank" rel="noopener noreferrer" title="Escribir por WhatsApp" aria-label="Escribir por WhatsApp al ${esc(p)}">${callIco}</a></span>`;
  }
  document.addEventListener('click', async e => {
    const b = e.target.closest('[data-copy-tel]');
    if (!b) return;
    e.preventDefault();
    let text = 'Copiado';
    try { await navigator.clipboard.writeText(b.dataset.copyTel); } catch { getSelection().selectAllChildren(b); text = 'Selecciónalo y cópialo'; }
    const box = b.closest('.telx');
    box.querySelector('.teltip')?.remove();
    const tip = document.createElement('span');
    tip.className = 'teltip';
    tip.setAttribute('role', 'status');
    tip.textContent = text;
    box.appendChild(tip);
    setTimeout(() => tip.remove(), 1500);
  });
  // Con una ventana abierta (dialog o .modalbg.open) la página de atrás no se mueve: en iPhone overflow:hidden no basta,
  // así que el body queda fijo en su posición y al cerrar vuelve exactamente al mismo lugar.
  let locked = null;
  function syncScrollLock() {
    const open = !!document.querySelector('dialog[open],.modalbg.open'), b = document.body;
    if (open && !locked) {
      locked = {y: scrollY};
      Object.assign(b.style, {position: 'fixed', top: `-${locked.y}px`, left: '0', right: '0', width: '100%'});
    } else if (!open && locked) {
      const y = locked.y;
      locked = null;
      Object.assign(b.style, {position: '', top: '', left: '', right: '', width: ''});
      scrollTo(0, y);
    }
  }
  new MutationObserver(syncScrollLock).observe(document.documentElement, {subtree: true, attributes: true, attributeFilter: ['open', 'class']});
  // Varias respuestas con el mismo estilo: options = [[valor, texto, 'ok'|'go'|''], …], en columna; resuelve el valor
  // elegido, o null con Escape.
  function choose(text, {detail = '', options = []} = {}) {
    let d = document.getElementById('appChoose');
    if (!d) {
      d = document.createElement('dialog');
      d.id = 'appChoose';
      d.className = 'appconfirm appchoose';
      document.body.appendChild(d);
    }
    d.innerHTML = `<form method="dialog"><h2></h2><p></p><div class="acts">${options.map(([v, label, kind]) => `<button value="${esc(v)}" class="${kind === 'ok' ? 'yes ok' : kind === 'go' ? 'yes go' : 'no'}">${esc(label)}</button>`).join('')}</div></form>`;
    d.querySelector('h2').textContent = text;
    d.querySelector('p').textContent = detail;
    d.returnValue = '';
    d.showModal();
    return new Promise(resolve => d.addEventListener('close', () => resolve(d.returnValue || null), {once: true}));
  }

  // Teléfono (≤760px) en los paneles: un encabezado de app — logo, nombre de la página, la acción principal de la página
  // («+», p. ej. Agregar punto) y ☰, que abre desde arriba tu nombre con «Salir», las páginas y el tamaño del texto.
  // Reutiliza los botones de la página (salir, agregar), que en el teléfono quedan ocultos. La página pública tiene su propia barra.
  const svg = d => `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${d}</svg>`;
  function mountMobileBar() {
    const top = document.querySelector('header.top');
    if (!top || !document.getElementById('appNav') || document.getElementById('mBar')) return;
    const bar = document.createElement('div');
    bar.id = 'mBar';
    bar.className = 'mbar';
    bar.innerHTML = `<nav class="mlinks" id="mLinks" aria-label="Páginas"></nav><span class="mtitle" id="mTitle"></span>${top.querySelector('.topsearch') ? `<button type="button" class="msearchbtn" id="mSearch" aria-expanded="false" aria-label="Buscar">${svg('<circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/>')}</button>` : ''}<button type="button" class="madd" id="mAdd" hidden></button>${document.getElementById('loginBtn') ? '<button type="button" class="mlogin" id="mLogin">Login</button>' : ''}<button type="button" class="mmenu" id="mMenu" aria-expanded="false" aria-controls="mSheet" aria-label="Menú">${svg('<path d="M4 7h16M4 12h16M4 17h16"/>')}</button>`;
    top.appendChild(bar);
    const sheet = document.createElement('div');
    sheet.id = 'mSheet';
    sheet.className = 'msheet';
    sheet.hidden = true;
    top.appendChild(sheet);
    drawMobileBar();
  }
  function drawMobileBar() {
    const title = document.getElementById('mTitle');
    if (!title) return;
    const u = user() || {}, here = location.pathname, page = pages.find(([url]) => url === here);
    title.textContent = page ? page[2](u) : document.title.split('·')[0].trim();
    const list = user() ? pages.filter(([url, , , ok]) => url === here || ok(u)) : [], links = document.getElementById('mLinks');
    links.innerHTML = list.length > 1 ? list.map(([url, icon, label]) => `<a href="${url}"${url === here ? ' aria-current="page"' : ''} title="${esc(label(u))}"><span class="mi" aria-hidden="true">${icon}</span><span class="ml">${esc(label(u))}</span></a>`).join('') : '';
    // Público sin sesión: «¡Ayuda!» primero, luego Login; sin título ni búsqueda.
    document.getElementById('mBar').classList.toggle('public', !user() && !!document.getElementById('loginBtn'));
    fitMobileBar();
    const add = document.querySelector('.top .topright .add'), madd = document.getElementById('mAdd');
    madd.hidden = !add;
    if (add) { const raw = add.textContent.trim(), plus = raw.startsWith('+'), label = raw.replace(/^\+\s*/, ''); madd.classList.toggle('text', !plus); madd.innerHTML = plus ? svg('<path d="M12 5v14M5 12h14"/>') : esc(raw); madd.title = label; madd.setAttribute('aria-label', label); }
  }
  // Las páginas van en la barra si caben; si no, primero solo íconos (menos la actual) y, si aún no caben, al menú ☰.
  function fitMobileBar() {
    const bar = document.getElementById('mBar'), links = document.getElementById('mLinks');
    if (!bar || !links) return;
    bar.classList.remove('compact', 'tight');
    if (!links.children.length) { bar.classList.add('tight'); return; }
    if (getComputedStyle(bar).display === 'none') return;
    const over = () => links.scrollWidth > links.clientWidth + 1;
    if (over()) bar.classList.add('compact');
    if (over()) { bar.classList.remove('compact'); bar.classList.add('tight'); }
  }
  addEventListener('resize', () => fitMobileBar());
  function showSheet(open) {
    const sheet = document.getElementById('mSheet'), btn = document.getElementById('mMenu');
    if (!sheet) return;
    if (open) {
      const u = user(), here = location.pathname === '/index.html' ? '/' : location.pathname, out = document.getElementById('logout') || document.getElementById('codeLogout');
      const name = document.getElementById('whoName')?.title || u?.name || '';
      const signedIn = !!out && !out.closest('[hidden]'), login = !signedIn && document.getElementById('loginBtn'), copyId = signedIn && document.getElementById('copyMyId');
      const list = u ? pages.filter(([url, , , ok]) => url === here || ok(u)) : [];
      const outLabel = out && out.textContent.trim() !== '✕' ? out.textContent.trim() : 'Salir';
      sheet.innerHTML = `${signedIn ? `<div class="msuser"><b>${esc(name)}</b>${copyId ? '<button type="button" data-m="copyid" class="msoft">Copiar mi ID</button>' : ''}<button type="button" data-m="logout">${esc(outLabel)}</button></div>` : ''}${list.length > 1 && document.getElementById('mBar')?.classList.contains('tight') ? `<nav class="mpages" aria-label="Páginas">${list.map(([url, icon, label]) => `<a href="${url}"${url === here ? ' aria-current="page"' : ''}><span aria-hidden="true">${icon}</span>${esc(label(u))}</a>`).join('')}</nav>` : ''}<div class="msize"><span class="lbl">Tamaño del texto</span><button type="button" data-msize="-1" aria-label="Texto más pequeño">A−</button><output>${Math.round(zoom * 100)}%</output><button type="button" data-msize="1" aria-label="Texto más grande">A+</button></div>`;
      sheet.querySelector('[data-msize="-1"]').disabled = zoom === sizes[0];
      sheet.querySelector('[data-msize="1"]').disabled = zoom === sizes[sizes.length - 1];
    }
    sheet.hidden = !open;
    btn.setAttribute('aria-expanded', String(open));
  }
  document.addEventListener('click', e => {
    const sheet = document.getElementById('mSheet');
    if (!sheet) return;
    if (e.target.closest('#mMenu')) { showSheet(sheet.hidden); return; }
    const size = e.target.closest('[data-msize]');
    if (size) { zoom = sizes[Math.min(sizes.length - 1, Math.max(0, sizes.indexOf(zoom) + Number(size.dataset.msize)))]; try { localStorage.setItem('textZoom', String(zoom)); } catch {} applySize(); showSheet(true); return; }
    if (e.target.closest('[data-m="logout"]')) { showSheet(false); (document.getElementById('logout') || document.getElementById('codeLogout'))?.click(); return; }
    if (e.target.closest('#mLogin')) { showSheet(false); document.getElementById('loginBtn')?.click(); return; }
    if (e.target.closest('[data-m="login"]')) { showSheet(false); document.getElementById('loginBtn')?.click(); return; }
    const cid = e.target.closest('[data-m="copyid"]');
    if (cid) { document.getElementById('copyMyId')?.click(); cid.textContent = 'ID copiado'; return; }
    if (e.target.closest('#mAdd')) { showSheet(false); document.querySelector('.top .topright .add')?.click(); return; }
    const sb = e.target.closest('#mSearch');
    if (sb) { const top = document.querySelector('header.top'), open = !top.classList.contains('msearch'); top.classList.toggle('msearch', open); sb.setAttribute('aria-expanded', String(open)); showSheet(false); if (open) top.querySelector('.topsearch')?.focus(); return; }
    if (!sheet.hidden && (!sheet.contains(e.target) || e.target.closest('a'))) showSheet(false);
  });
  document.addEventListener('keydown', e => { if (e.key === 'Escape') showSheet(false); });

  // Entradas (Login del mapa, Runners, Constructores, Comando): dos grupos siempre a la vista, sin ocultar campos.
  // «Ya tengo código» (accessCode + Recordar + Entrar) y «Crear cuenta nueva» (nombre, teléfono, correo). Los formularios
  // llevan data-signin; cada página hace el envío y pregunta a signinMode qué grupo se envió.
  const codeIntro = 'Escribe tu código de acceso para retomar tus últimas carreras.';
  // ¿Es un código de acceso? Empieza con «!», o es una sola palabra con letras y números (como los que se generan al
  // registrarse, p. ej. loro47juan3). Los nombres no llevan números.
  const isCode = v => { v = String(v || '').trim(); return /^!\S/.test(v) || /^(?=.*\d)(?=.*[a-zñ])[a-z0-9ñ]{6,24}$/i.test(v); };
  const codeOf = v => String(v || '').trim().replace(/^!/, '').replace(/\s+/g, '');
  const signinForms = 'form[data-signin]';
  // Qué grupo se envió: el botón tocado o, con Enter, el grupo donde está el cursor (con Enter el navegador reporta el
  // primer botón del formulario y le hace un clic sin puntero (detail 0) con el cursor en otro campo: ese no cuenta).
  document.addEventListener('click', e => { const b = e.target.closest?.('form[data-signin] .sgbtn'); if (b?.form && (e.detail > 0 || document.activeElement === b)) { b.form.dataset.clicked = b.value; setTimeout(() => delete b.form.dataset.clicked); } }, true);
  function signinMode(e) {
    const clicked = e.target?.dataset?.clicked;
    if (clicked) return clicked === 'new' ? 'new' : 'code';
    return document.activeElement?.closest?.('.sg-new') ? 'new' : 'code';
  }
  const codeKeys = ['runnerCode', 'builderCode', 'agentCode'];
  const rememberedCode = () => { try { return localStorage.getItem('rememberCode') !== '0' && localStorage.getItem('rememberedCode') || ''; } catch { return ''; } };
  // El código recordado ya escrito (al abrir la página, al abrir «Login» y después de «Salir»): solo falta «Entrar».
  function prefillCode(form) {
    const box = form?.elements?.accessCode, saved = rememberedCode();
    if (!box || !saved || box.value.trim()) return;
    box.value = saved;
    form.dataset.filled = '1';
  }
  const prefillAll = () => document.querySelectorAll(signinForms).forEach(f => { prefillCode(f); const r = f.querySelector('.remember input'); if (r) { try { r.checked = localStorage.getItem('rememberCode') !== '0'; } catch {} } });
  document.addEventListener('reset', e => { if (e.target.matches?.(signinForms)) setTimeout(() => prefillAll()); });
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', prefillAll); else setTimeout(prefillAll);
  document.addEventListener('input', e => {
    const f = e.target.closest?.(signinForms);
    if (!f) return;
    // Un código pegado en «Nombre completo» pasa a su campo.
    if (e.target.name === 'name' && f.elements.accessCode && isCode(e.target.value)) { f.elements.accessCode.value = codeOf(e.target.value); e.target.value = ''; f.elements.accessCode.focus(); return; }
    // Escribir solo «!» en el código completa el recordado (una vez por intento).
    const saved = rememberedCode();
    if (e.target.name === 'accessCode' && e.target.value.trim() === '!' && saved && !f.dataset.filled && e.inputType !== 'deleteContentBackward') { f.dataset.filled = '1'; e.target.value = saved; }
    if (e.target.name === 'accessCode' && !e.target.value.trim()) delete f.dataset.filled;
  });
  // Al enviar el código se anota la preferencia «Recordar»; el código se guarda solo cuando la página confirma que entró.
  document.addEventListener('submit', e => {
    const box = e.target.querySelector?.('.remember input');
    if (!e.target.matches?.(signinForms) || !box || signinMode(e) !== 'code') return;
    try { localStorage.setItem('rememberCode', box.checked ? '1' : '0'); } catch {}
  }, true);
  // Entró con este código: se recuerda en el teléfono (por defecto) para todas las páginas; si desmarcó «Recordar», se olvida.
  function codeSignedIn(code) {
    code = codeOf(code);
    if (!code) return;
    try {
      if (localStorage.getItem('rememberCode') !== '0') { localStorage.setItem('rememberCode', '1'); localStorage.setItem('rememberedCode', code); codeKeys.forEach(k => localStorage.setItem(k, code)); }
      else { localStorage.removeItem('rememberedCode'); setTimeout(() => codeKeys.forEach(k => localStorage.removeItem(k)), 4000); }
    } catch {}
  }
  // Cuenta nueva desde una entrada: crea el usuario (aprobado al instante) y devuelve su código; muestra el código para copiarlo.
  async function createAccount({name, telephone, email, panel}) {
    name = String(name || '').trim().replace(/\s+/g, ' ');
    if (name.split(' ').length < 2) throw Object.assign(Error('Escribe tu nombre y apellido.'), {field: 'name'});
    if (String(telephone || '').replace(/\D/g, '').length < 7) throw Object.assign(Error('Escribe un teléfono válido (por ejemplo 6000-0000).'), {field: 'telephone'});
    if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(email)) throw Object.assign(Error('Revisa el correo (por ejemplo nombre@gmail.com) o déjalo vacío.'), {field: 'email'});
    const r = await fetch('/api/session/signup', {method: 'POST', credentials: 'same-origin', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({name, telephone, email, ...(panel ? {panel} : {})})});
    const body = await r.json().catch(() => ({}));
    if (!r.ok) throw Object.assign(Error(body.error || 'No se pudo crear la cuenta.'), {field: body.field});
    try { localStorage.setItem('rememberCode', '1'); } catch {}
    // Si otra persona ya tenía ese nombre, el servidor le agrega un número («Juan Pérez 2»).
    return {code: String(body.accessCode).toLowerCase(), name: body.name};
  }
  async function showNewCode(code, name = '') {
    const choice = await choose('Tu código de acceso: ' + code, {detail: (name ? `Tu cuenta: ${name}. ` : '') + 'Ya quedó guardado en este teléfono: la próxima vez aparece escrito y solo tocas «Entrar». Cópialo para entrar en otro teléfono.', options: [['ok', 'Listo', 'go'], ['copy', 'Copiar código', 'ok']]});
    if (choice === 'copy') { try { await navigator.clipboard.writeText(code); } catch {} }
  }
  // Nombre corto para encabezados: nombre + primeras letras del apellido («Arianne Ori…»).
  const shortName = n => { const w = String(n || '').trim().split(/\s+/).filter(Boolean); return w.length < 2 ? w.join('') : w[0] + ' ' + (w[1].length > 3 ? w[1].slice(0, 3) + '…' : w[1]); };
  const drawAll = () => { draw(); drawMobileBar(); };
  window.AppNav = {draw: drawAll, confirm, choose, phone, shortName, codeIntro, isCode, codeOf, codeSignedIn, prefillCode, signinMode, createAccount, showNewCode};
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', () => { draw(); mountTextSize(); mountMobileBar(); }); else { draw(); mountTextSize(); mountMobileBar(); }
  // Entrar o salir en otra pestaña actualiza esta.
  addEventListener('storage', e => { if (e.key === 'codeUser' || e.key === null) draw(); if (e.key === 'textZoom' || e.key === null) { zoom = sizes.includes(savedSize()) ? savedSize() : 1; applySize(); } });
})();
