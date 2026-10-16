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
.telnum{border:0;background:none;padding:0;color:#126e7c;font:inherit;font-weight:750;cursor:pointer;text-decoration:underline dotted;text-underline-offset:3px;white-space:nowrap}.telnum:hover{color:#10263a}
.navtelpop{position:absolute;z-index:4000;min-width:210px;background:#fff;border:1px solid #cad9d9;border-radius:12px;box-shadow:0 12px 30px #17354233;padding:10px 12px;font-family:Inter,system-ui,sans-serif;color:#173c46}.ntp-h{font-weight:800;font-size:14px}.ntp-num{font-size:17px;font-weight:800;margin:4px 0 8px;letter-spacing:.02em}.ntp-acts{display:flex;gap:8px}.ntp-acts a{flex:1;display:inline-flex;align-items:center;justify-content:center;gap:5px;padding:8px 10px;border-radius:9px;font-weight:800;font-size:13px;text-decoration:none}.ntp-call{background:#19745f;color:#fff}.ntp-wa{background:#25d366;color:#08331b}.ntp-copy{display:block;width:100%;margin-top:6px;padding:6px;border:1px solid #cad9d9;border-radius:9px;background:#fff;color:#36505a;font:inherit;font-size:12px;font-weight:700;cursor:pointer}
.navtip{position:fixed;z-index:4100;background:#fff;color:#173542;border:1px solid #cad9d9;border-radius:14px;box-shadow:0 14px 34px #17354240;padding:12px 14px;font-family:Inter,system-ui,sans-serif;font-size:14px;line-height:1.45;text-align:left}.navtip-t{display:block;font-size:15px;margin-bottom:4px}.navtip-b p{margin:0 0 6px}.navtip-b ol{margin:4px 0 6px;padding-left:20px}.navtip-b li{margin:2px 0}.navtip-a{display:flex;gap:8px;justify-content:flex-end;margin-top:8px}.navtip-a button{border:1px solid #cad9d9;border-radius:9px;background:#fff;color:#35505e;font:inherit;font-size:13px;font-weight:700;padding:7px 10px;cursor:pointer}.navtip-a .ok{background:#126e7c;border-color:#126e7c;color:#fff}
.swipe-l{animation:navSwipeL .22s ease-out}.swipe-r{animation:navSwipeR .22s ease-out}@keyframes navSwipeL{from{transform:translateX(28px);opacity:.35}}@keyframes navSwipeR{from{transform:translateX(-28px);opacity:.35}}@media (prefers-reduced-motion:reduce){.swipe-l,.swipe-r{animation:none}}
.appconfirm{border:0;border-radius:16px;padding:0;width:min(380px,calc(92vw / var(--text-zoom,1)));box-shadow:0 20px 60px #10263a66;font-family:Inter,system-ui,sans-serif}.appconfirm::backdrop{background:#10263ab8}.appconfirm form{margin:0;padding:22px}.appconfirm h2{margin:0;font-size:19px;line-height:1.3;color:#152b3a}.appconfirm p{margin:8px 0 0;color:#667982;font-size:13px;line-height:1.45}.appconfirm p:empty{display:none}
.textsize{position:relative;margin-left:8px;flex:none}.textsize>button{height:28px;padding:0 8px;border:1px solid #47657a;border-radius:8px;background:none;color:#c8d6db;font:800 13px Inter,system-ui,sans-serif;cursor:pointer}.textsize>button:hover,.textsize>button[aria-expanded="true"]{background:#ffffff1a;color:#fff}
.textsize-pop{position:absolute;top:calc(100% + 8px);left:0;z-index:3100;display:flex;align-items:center;gap:6px;background:#fff;border-radius:12px;padding:7px;box-shadow:0 10px 26px #0005}.textsize-pop[hidden]{display:none}.textsize-pop button{width:40px;height:36px;border:1px solid #cad9d9;border-radius:9px;background:#fff;color:#173c46;font:800 15px Inter,system-ui,sans-serif;cursor:pointer}.textsize-pop button:disabled{opacity:.4;cursor:default}.textsize-pop span{min-width:46px;text-align:center;font:750 13px Inter,system-ui,sans-serif;color:#36505a}
html{zoom:var(--text-zoom,1)}.leaflet-container{zoom:calc(1 / var(--text-zoom,1))}
.appnav-list .appnav-admin{margin-left:10px;border:1px dashed #ffb4a366;font-size:12px!important;opacity:.7;color:#ffb4a3!important}.appnav-list .appnav-admin:hover{opacity:1}
.madmin{display:flex;align-items:center;justify-content:center;gap:6px;margin:14px auto 0;padding:6px 12px;width:max-content;border:1px dashed #ffb4a366;border-radius:8px;color:#ffb4a3;font:700 12px Inter,system-ui,sans-serif;text-decoration:none;opacity:.75}
.tswrap{position:relative;display:block}.tswrap:focus-within{border-color:#ec7651;background:#203c54}.tswrap .tsinput{display:block;width:100%;margin:0;padding:0 26px 0 0;border:0;background:none;color:inherit;font:inherit;outline:none}.tswrap .tsinput::placeholder{color:#9fb6c2}.tsbox{position:relative}.tsbox input{padding-right:34px!important}.tsinput::-webkit-search-cancel-button,.tsbox input::-webkit-search-cancel-button{display:none}
.tswrap.tson,.tsbox.tson input{border-color:#ed7654!important;box-shadow:0 0 0 2px #ed765466}.msearchbtn.tson{position:relative;background:#ed7654;border-color:#ed7654;color:#fff}.msearchbtn.tson:after{content:'';position:absolute;top:5px;right:5px;width:9px;height:9px;border-radius:50%;background:#ffd166;box-shadow:0 0 0 2px #10263a}
.searchpill{position:fixed;left:50%;bottom:max(14px,env(safe-area-inset-bottom));transform:translateX(-50%);z-index:3300;display:flex;align-items:center;max-width:calc(calc(100vw / var(--text-zoom,1)) - 24px);border-radius:999px;background:#10263a;color:#fff;box-shadow:0 8px 24px #10263a59;font:700 13px Inter,system-ui,sans-serif}.searchpill[hidden]{display:none}.searchpill button{border:0;background:none;color:inherit;font:inherit;cursor:pointer}.sp-what{min-width:0;padding:9px 6px 9px 14px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.sp-what b{color:#ffd166}.sp-x{flex:none;width:34px;height:34px;margin-right:2px;border-radius:50%!important;background:#ffffff1f!important}
.gsdlg{border:0;border-radius:16px;padding:0;width:min(640px,calc(94vw / var(--text-zoom,1)));max-height:calc(92dvh / var(--text-zoom,1));overflow:auto;font-family:Inter,system-ui,sans-serif;color:#173c46;background:#f5f8f7}.gsdlg::backdrop{background:#10263ab8}.gs-top{position:sticky;top:0;z-index:2;display:flex;align-items:center;gap:10px;padding:10px 14px;padding-top:max(10px,env(safe-area-inset-top));background:#10263a;color:#fff}.gs-top b{flex:1;font-size:16px}.gs-back,.gs-x{border:1px solid #47657a;border-radius:9px;background:none;color:#fff;font:800 14px Inter,system-ui,sans-serif;padding:7px 11px;cursor:pointer}.gs-bar{position:sticky;top:52px;z-index:1;padding:10px 14px;background:#f5f8f7;border-bottom:1px solid #e1e9ea}.gs-bar input{width:100%;box-sizing:border-box;padding:11px 12px;border:1px solid #cad9d9;border-radius:10px;font:inherit;font-size:16px}.gs-count{margin:10px 14px;color:#5d7680;font-size:13px}.gs-list{display:grid;gap:8px;padding:0 14px 18px}.gs-card{padding:12px 14px;border:1px solid #dce5e4;border-radius:12px;background:#fff}.gs-when{color:#667982;font-size:12px}.gs-card h3{margin:4px 0;font-size:16px;line-height:1.3}.gs-badge{display:inline-block;vertical-align:middle;border-radius:20px;padding:2px 8px;background:#eef3f2;color:#566a71;font-size:11px;font-weight:800}.gs-badge.urg{background:#fde4dc;color:#a83c1f}.gs-badge.pend{background:#fff0e8;color:#ae4f32}.gs-needs{list-style:none;margin:8px 0 0;padding:0;display:grid;gap:6px}.gs-needs li{padding:8px 10px;border:1px solid #e4ecec;border-radius:10px;background:#fafcfc}.gs-st{display:inline-block;margin-right:6px;border-radius:20px;padding:2px 9px;background:#173c46;color:#fff;font-size:11px;font-weight:800}.gs-st.route{background:#e8f0fb;color:#1f4f9a}.gs-code{margin-left:6px;color:#7a8d94;font-weight:700}.gs-card p{margin:5px 0 0;font-size:13px;color:#36505a;overflow-wrap:anywhere}.gs-meta{color:#7a8d94!important;font-size:12px!important}.gs-acts{display:flex;flex-wrap:wrap;gap:8px;margin-top:9px}.gs-btn{display:inline-flex;align-items:center;padding:7px 12px;border:1px solid #cad9d9;border-radius:9px;background:#fff;color:#173c46;font-weight:800;font-size:13px;text-decoration:none}
@media(max-width:760px){.gsdlg{width:calc(100vw / var(--text-zoom,1));max-width:calc(100vw / var(--text-zoom,1));height:calc(100dvh / var(--text-zoom,1));max-height:calc(100dvh / var(--text-zoom,1));margin:0;border-radius:0}}
.tsclear{position:absolute;right:6px;top:50%;transform:translateY(-50%);display:grid;place-items:center;width:26px;height:26px;padding:0;border:0;border-radius:50%;background:#ffffff26;color:#fff;font:800 12px Inter,system-ui,sans-serif;cursor:pointer}.tsclear:hover{background:#ffffff40}.tsclear[hidden]{display:none}
.mbar,.msheet{display:none}
@media(max-width:760px){.top:has(>.mbar)>.appnav,.top:has(>.mbar) .textsize,.top:has(>.mbar)>.topright{display:none!important}.top:has(>.mbar) .brand{display:none!important}.top:has(>.mbar){gap:10px}
.mbar{display:flex;align-items:center;gap:8px;flex:1;min-width:0}.mlinks{flex:1;min-width:0;display:flex;gap:3px;overflow:hidden;background:#ffffff12;border-radius:11px;padding:3px}.mlinks a{flex:none;display:inline-flex;align-items:center;gap:6px;padding:8px 10px;border-radius:9px;color:#c8d6db;text-decoration:none;font:800 14px Inter,system-ui,sans-serif;white-space:nowrap}.mlinks a[aria-current="page"]{background:#fff;color:#10263a}.mlinks .mi{display:inline-flex}.mbar .mtitle{display:none}.mbar.tight .mlinks{display:none}.mbar.tight .mtitle{display:flex}.mtitle{flex:1;min-width:0;align-items:center;gap:7px;margin:0;padding:8px 10px;border:0;border-radius:11px;background:none;color:#fff;font:800 17px Inter,system-ui,sans-serif;text-align:left;white-space:nowrap;overflow:hidden}.mtitle .mtl{min-width:0;overflow:hidden;text-overflow:ellipsis}.mtitle:not(:disabled){background:#ffffff12;cursor:pointer}.mtitle:disabled{opacity:1;cursor:default}.mtitle .chev{margin-left:auto;padding-left:4px;font-size:12px;opacity:.85}.mtitle[aria-expanded="true"]{background:#fff;color:#10263a}
.mdrop{position:absolute;top:100%;left:8px;right:8px;z-index:3001;display:flex;flex-direction:column;gap:2px;padding:6px;background:#10263a;border-radius:0 0 14px 14px;box-shadow:0 16px 34px #0008}.mdrop[hidden]{display:none}.mdrop a{display:flex;align-items:center;gap:10px;padding:13px 12px;border-radius:10px;color:#e3ecef;text-decoration:none;font:800 16px Inter,system-ui,sans-serif}.mdrop a[aria-current="page"]{background:#fff;color:#10263a}.mdrop .ml{display:inline!important}.mdrop .mdrop-admin{margin-top:6px;border:1px dashed #ffb4a366;color:#ffb4a3;font-size:14px;opacity:.85}
.mmenu.me{border-radius:50%;background:#ffd166;border-color:#ffd166;color:#10263a;box-shadow:0 0 0 2px #10263a,0 0 0 3px #ffd16699}.minit{font:900 15px/1 Inter,system-ui,sans-serif;letter-spacing:.02em}.mmenu.me[aria-expanded="true"]{background:#fff;border-color:#fff}}
.madd,.mmenu,.msearchbtn{flex:none;display:grid;place-items:center;width:42px;height:42px;border-radius:12px;border:0;cursor:pointer;text-decoration:none}.madd{background:#ed7654;color:#fff}.madd[hidden]{display:none}.mmenu,.msearchbtn{background:#ffffff14;border:1px solid #47657a;color:#fff}.msearchbtn[aria-expanded="true"]{background:#fff;color:#10263a}.madd.text{width:auto;padding:0 14px;font:800 15px Inter,system-ui,sans-serif;white-space:nowrap}.top:has(>.mbar){flex-wrap:wrap;height:auto;min-height:62px;align-content:center}@media(max-width:760px){.top:has(>.mbar) .topsearch{display:none!important}}.top.msearch .topsearch{display:block!important;flex:1 0 100%;order:9;max-width:none;margin:0 0 10px}body:has(.top>.mbar) .addbar{display:none!important}.msuser .msoft{color:#c8d6db}.mlogin{display:none}.mbar.public .mlinks,.mbar.public .mtitle,.mbar.public .msearchbtn{display:none!important}.mbar.public .madd{order:-1}.mbar.public .mlogin{display:inline-flex;align-items:center;height:42px;margin-left:auto;padding:0 14px;border:1px solid #ffffff55;border-radius:12px;background:none;color:#fff;font:800 15px Inter,system-ui,sans-serif;cursor:pointer}.top.msearch:has(.mbar.public) .topsearch{display:none!important}.msuser .mgo{background:#ed7654;border-color:#ed7654;color:#fff}.mmenu[aria-expanded="true"]{background:#fff;color:#10263a}
.msheet{position:absolute;top:100%;left:0;right:0;z-index:3000;background:#10263a;padding:6px 14px 16px;box-shadow:0 16px 34px #0008;border-radius:0 0 16px 16px;font-family:Inter,system-ui,sans-serif}.msheet:not([hidden]){display:block}
.msuser{display:flex;align-items:center;gap:10px;padding:12px 4px;border-bottom:1px solid #ffffff1f;color:#fff}.msuser b{flex:1;min-width:0;font-size:15px;overflow-wrap:anywhere}.msuser button{border:1px solid #ffffff40;background:none;color:#ffd0c2;border-radius:9px;padding:8px 12px;font:800 13px Inter,system-ui,sans-serif;cursor:pointer}
.msroles{display:flex;flex-wrap:wrap;gap:4px;margin-top:6px}.msrole{font-size:11px;font-weight:800;padding:2px 8px;border-radius:999px;background:#ffffff1f;color:#d7e6ea}.mpages{display:flex;flex-direction:column;gap:2px;padding:8px 0;border-bottom:1px solid #ffffff1f}.mpages a{display:flex;align-items:center;gap:10px;padding:12px 10px;border-radius:10px;color:#e3ecef;text-decoration:none;font-size:15px;font-weight:750}.mpages a[aria-current="page"]{background:#fff;color:#10263a}.mpages a:active{background:#ffffff26}
.msize{display:flex;align-items:center;gap:8px;padding:12px 4px 2px;color:#c8d6db;font-size:14px;font-weight:700}.msize .lbl{flex:1}.msize button{width:44px;height:38px;border:1px solid #47657a;border-radius:9px;background:none;color:#fff;font:800 15px Inter,system-ui,sans-serif;cursor:pointer}.msize button:disabled{opacity:.4}.msize output{min-width:46px;text-align:center;color:#fff}}
.gate:has(form[data-signin]){overflow-y:auto;place-items:start center}.gate form[data-signin]{margin:auto 0}.accesstoast{position:fixed;z-index:4000;left:50%;top:76px;transform:translateX(-50%);width:min(440px,calc(calc(100vw / var(--text-zoom,1)) - 24px));padding:12px 16px;border-radius:12px;background:#10263a;color:#fff;font:700 14px Inter,system-ui,sans-serif;box-shadow:0 8px 24px #0006;text-align:center}.sg{border:1px solid #dbe5e5;border-radius:12px;margin:12px 0 0;padding:4px 14px 14px;min-width:0}.sg legend{padding:0 6px;font:800 14px Inter,system-ui,sans-serif;color:#173c46}.sg label{display:block}.sg .sgbtn{display:block;width:100%;margin-top:12px;border:0;border-radius:10px;padding:12px 16px;font:800 15px Inter,system-ui,sans-serif;cursor:pointer;color:#fff;background:#19745f}.sg-new .sgbtn{background:#ed7654}.sg small{font-weight:600;color:#7a8d94}.sgor{margin:10px 0 0;text-align:center;font-size:12px;font-weight:700;color:#8a9aa0}.remember{display:flex!important;align-items:center;gap:8px;margin:10px 0 2px;font-size:13px;font-weight:650;color:#36505a;cursor:pointer}.remember[hidden]{display:none!important}.remember input{width:20px;height:20px;margin:0;accent-color:#19745f;flex:none}
.appconfirm .acts{display:flex;gap:10px;margin-top:18px}.appconfirm button{flex:1;border:0;border-radius:10px;padding:13px 16px;font:inherit;font-size:15px;font-weight:800;cursor:pointer}.appconfirm .no{background:#eef3f3;color:#173c46}.appconfirm .yes{background:#9a4536;color:#fff}.appchoose .acts{flex-direction:column}.appconfirm .yes.ok{background:#19745f}.appconfirm .yes.go{background:#1f63c6}`;
  const style = document.createElement('style');
  style.textContent = css;
  document.head.appendChild(style);
  const panelUrl = {runner: '/runner.html', builder: '/builder.html', agent: '/command.html', donations: '/donations.html', supplier: '/supplier.html', cook: '/cook.html', teammanager: '/team.html'};
  const has = (u, role) => (u.roles || []).includes(role) || (u.panels || []).some(p => p.url === panelUrl[role]);
  // Agentes sin rol de runner entran al panel de runners solo para las listas de empaque.
  const pages = [
    ['/', '🏠', () => 'Inicio', () => true],
    // Scout: los puntos que agregaste tú o tu equipo; cualquiera que entró con código (no es un rol).
    ['/scout.html', '🧭', () => 'Scout', u => !!u],
    ['/runner.html', '<svg class="walkico" style="color:#5b9be6" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="14" cy="4" r="2"/><path d="M12.5 7.5 10 13"/><path d="m12 8 4 2.5 3-1"/><path d="m12 8-4 1.5-2 2.5"/><path d="m10 13 3.5 3-1 5"/><path d="m10 13-2.5 4H4"/></svg>', u => !u.roles || u.roles.includes('runner') ? 'Runners' : 'Empaque', u => has(u, 'runner') || has(u, 'agent')],
    ['/builder.html', '🔨', () => 'Constructores', u => has(u, 'builder')],
    ['/command.html', '📞', () => 'Checkup', u => has(u, 'agent')],
    ['/supplier.html', '📦', () => 'Supply', u => has(u, 'supplier')],
    ['/cook.html', '🍳', () => 'Cocina', u => has(u, 'cook')],
    ['/donations.html', '🎁', () => 'Donaciones', u => has(u, 'donations')],
    ['/team.html', '👥', () => 'Mi equipo', u => has(u, 'teammanager') || has(u, 'admin'), 'menu'],
  ];
  function user() {
    try { return JSON.parse(localStorage.getItem('codeUser') || 'null'); } catch { return null; }
  }
  const esc = s => String(s).replace(/[&<>"']/g, c => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
  function draw() {
    const el = document.getElementById('appNav');
    if (!el) return;
    const u = user(), here = location.pathname === '/index.html' ? '/' : location.pathname;
    let list = u ? pages.filter(([url, , , ok]) => url === here || ok(u)) : [];
    // En el panel de administración solo Inicio y «🔒 Administración» (la actual); sus herramientas están en su ☰.
    if (here === '/admin.html') list = [pages[0], ['/admin.html', '🔒', () => 'Administración', () => true]];
    const adminMenu = here === '/admin.html' ? ' data-admin-menu aria-haspopup="menu" aria-controls="adminMenu"' : '';
    const current = list.find(([url]) => url === here) || list[0] || pages[0];
    // Administración: con el rol Admin o si este navegador tiene abierta la sesión de la contraseña maestra; va al final,
    // apartada y pequeña, y pide confirmar para no entrar por error.
    const admin = (adminOK || (u?.roles || []).includes('admin')) && here !== '/admin.html' ? `<a href="/admin.html" class="appnav-admin" data-admin-link><span aria-hidden="true">🔒</span><span class="appnav-label">Administración</span></a>` : '';
    el.hidden = list.length < 2 && !admin;
    el.classList.remove('open');
    el.setAttribute('aria-label', 'Páginas');
    el.innerHTML = list.length < 2 && !admin ? '' : `<button type="button" class="appnav-toggle" aria-expanded="false" aria-controls="appNavList" aria-label="Cambiar de página"><span aria-hidden="true">${current[1]}</span>${esc(current[2](u))}<span class="chev" aria-hidden="true">▼</span></button><div class="appnav-list" id="appNavList">${list.map(([url, icon, label]) => `<a href="${url}"${url === here ? ' aria-current="page"' + adminMenu : ''}><span aria-hidden="true">${icon}</span><span class="appnav-label">${esc(label(u))}</span>${url === here && adminMenu ? '<span class="chev" aria-hidden="true">▾</span>' : ''}</a>`).join('')}${admin}</div>`;
  }
  // ¿Este navegador tiene abierta la sesión de administración? Se pregunta una vez por página (no en admin.html).
  let adminOK = false;
  if (location.pathname !== '/admin.html') fetch('/api/admin/session', {credentials: 'same-origin', cache: 'no-store'}).then(r => { if (r.ok) { adminOK = true; draw(); drawTitle(); } }).catch(() => {});
  document.addEventListener('click', async e => {
    const a = e.target.closest('[data-admin-link]');
    if (!a) return;
    e.preventDefault();
    if (await confirm('¿Abrir Administración?', {yes: 'Abrir', no: 'Cancelar', detail: 'Panel con la contraseña maestra: los cambios ahí afectan a todos.'})) location.href = a.href;
  });
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
  // Teléfono en cualquier página, igual en todas: tocar el número abre un cuadrito con el número, «📞 Llamar»,
  // «💬 WhatsApp» y «📋 Copiar» (como «Reportado por» en Runners y los contactos de Scout y Supply).
  // Sin código de país se asume Panamá (507), como en administración.
  const waNumber = t => { const d = String(t).replace(/\D/g, ''); return String(t).trim().startsWith('+') || d.length > 8 ? d : '507' + d; };
  function phone(raw, name = '') {
    const p = String(raw || '').trim(), dial = p.replace(/[^\d+]/g, '');
    if (!dial) return esc(p);
    return `<button type="button" class="telnum" data-tel-pop="${esc(p)}" data-tel-name="${esc(name)}" title="Llamar o escribir por WhatsApp">${esc(p)}</button>`;
  }
  const closeTelPop = () => document.querySelectorAll('.navtelpop').forEach(x => x.remove());
  document.addEventListener('click', async e => {
    const cp = e.target.closest('[data-copy-tel]');
    if (cp) {
      e.preventDefault();
      let text = '✓ Copiado';
      try { await navigator.clipboard.writeText(cp.dataset.copyTel); } catch { text = 'Selecciona el número y cópialo'; }
      cp.textContent = text;
      setTimeout(closeTelPop, 900);
      return;
    }
    const b = e.target.closest('[data-tel-pop]');
    if (!b) { if (!e.target.closest('.navtelpop')) closeTelPop(); return; }
    e.preventDefault(); e.stopPropagation();
    closeTelPop();
    const p = b.dataset.telPop, d = p.replace(/[^\d+]/g, ''), name = b.dataset.telName || '';
    const pop = document.createElement('div');
    pop.className = 'navtelpop';
    pop.setAttribute('role', 'dialog');
    pop.innerHTML = `${name ? `<div class="ntp-h">${esc(name)}</div>` : ''}<div class="ntp-num">${esc(p)}</div><div class="ntp-acts"><a class="ntp-call" href="tel:${esc(d)}">📞 Llamar</a><a class="ntp-wa" href="https://wa.me/${waNumber(p)}" target="_blank" rel="noopener noreferrer">💬 WhatsApp</a></div><button type="button" class="ntp-copy" data-copy-tel="${esc(p)}">📋 Copiar</button>`;
    (b.closest('dialog[open]') || document.body).appendChild(pop);
    const r = b.getBoundingClientRect(), host = pop.parentElement === document.body ? {x: scrollX, y: scrollY} : {x: -pop.parentElement.getBoundingClientRect().left + pop.parentElement.scrollLeft, y: -pop.parentElement.getBoundingClientRect().top + pop.parentElement.scrollTop};
    pop.style.top = (r.bottom + host.y + 6) + 'px';
    pop.style.left = Math.max(8, Math.min(r.left + host.x, (pop.parentElement === document.body ? document.documentElement.clientWidth : pop.parentElement.clientWidth) - pop.offsetWidth - 8)) + 'px';
  }, true);
  document.addEventListener('scroll', e => { if (!e.target.closest?.('.navtelpop')) closeTelPop(); }, true);
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
    bar.innerHTML = `<nav class="mlinks" id="mLinks" aria-label="Páginas"></nav><button type="button" class="mtitle" id="mTitle" aria-haspopup="true" aria-expanded="false" disabled></button>${top.querySelector('.topsearch') ? `<button type="button" class="msearchbtn" id="mSearch" aria-expanded="false" aria-label="Buscar">${svg('<circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/>')}</button>` : ''}<button type="button" class="madd" id="mAdd" hidden></button>${document.getElementById('loginBtn') ? '<button type="button" class="mlogin" id="mLogin">Inicia sesión</button>' : ''}<button type="button" class="mmenu" id="mMenu" aria-expanded="false" aria-controls="mSheet" aria-label="Menú">${svg('<path d="M4 7h16M4 12h16M4 17h16"/>')}</button>`;
    top.appendChild(bar);
    const sheet = document.createElement('div');
    sheet.id = 'mSheet';
    sheet.className = 'msheet';
    sheet.hidden = true;
    top.appendChild(sheet);
    const drop = document.createElement('nav');
    drop.id = 'mDrop';
    drop.className = 'mdrop';
    drop.setAttribute('aria-label', 'Páginas');
    drop.hidden = true;
    top.appendChild(drop);
    drawMobileBar();
  }
  function drawMobileBar() {
    const title = document.getElementById('mTitle');
    if (!title) return;
    const u = user() || {}, here = location.pathname;
    // Páginas de poco uso (Mi equipo) van solo en el menú ☰, salvo que sea la página actual.
    const list = user() ? pages.filter(([url, , , ok, where]) => url === here || ok(u) && where !== 'menu') : [], links = document.getElementById('mLinks');
    links.innerHTML = list.length > 1 ? list.map(([url, icon, label]) => `<a href="${url}"${url === here ? ' aria-current="page"' : ''} title="${esc(label(u))}"><span class="mi" aria-hidden="true">${icon}</span><span class="ml">${esc(label(u))}</span></a>`).join('') : '';
    // Público sin sesión: «¡Ayuda!» primero, luego Login; sin título ni búsqueda.
    document.getElementById('mBar').classList.toggle('public', !user() && !!document.getElementById('loginBtn'));
    fitMobileBar();
    drawMenuButton();
    const add = document.querySelector('.top .topright .add'), madd = document.getElementById('mAdd');
    madd.hidden = !add;
    if (add) { const raw = add.textContent.trim(), plus = raw.startsWith('+'), label = raw.replace(/^\+\s*/, ''); madd.classList.toggle('text', !plus); madd.innerHTML = plus ? svg('<path d="M12 5v14M5 12h14"/>') : esc(raw); madd.title = label; madd.setAttribute('aria-label', label); }
  }
  // En el teléfono las páginas van siempre plegadas, en todas las páginas por igual (aunque sean pocas): un botón con
  // la página actual y ▾ que abre la lista; así ＋ y el menú siempre quedan a la vista a la derecha.
  function fitMobileBar() {
    const bar = document.getElementById('mBar'), links = document.getElementById('mLinks');
    if (!bar || !links) return;
    bar.classList.add('tight');
    drawTitle();
  }
  function drawTitle() {
    const title = document.getElementById('mTitle'), drop = document.getElementById('mDrop'), links = document.getElementById('mLinks'), bar = document.getElementById('mBar');
    if (!title || !drop || !links) return;
    const u = user() || {}, here = location.pathname, page = pages.find(([url]) => url === here), label = page ? page[2](u) : document.title.replace(/^DEV · /, '').split('·')[0].trim();
    // Con el rol Admin (o la sesión de administración abierta) la lista de páginas termina con «🔒 Administración», como en
    // la computadora (pide confirmar antes de ir).
    const admin = (adminOK || (u.roles || []).includes('admin')) && here !== '/admin.html';
    const pick = bar.classList.contains('tight') && (links.children.length > 1 || admin && !!user());
    title.disabled = !pick;
    title.innerHTML = `${page ? `<span class="mi" aria-hidden="true">${page[1]}</span>` : ''}<span class="mtl">${esc(label)}</span>${pick ? '<span class="chev" aria-hidden="true">▾</span>' : ''}`;
    title.setAttribute('aria-label', pick ? `${label}: cambiar de página` : label);
    drop.innerHTML = pick ? links.innerHTML + (admin ? '<a href="/admin.html" class="mdrop-admin" data-admin-link><span class="mi" aria-hidden="true">🔒</span><span class="ml">Administración</span></a>' : '') : '';
    if (!pick) showDrop(false);
  }
  function showDrop(open) {
    const drop = document.getElementById('mDrop'), title = document.getElementById('mTitle');
    if (!drop || !title) return;
    drop.hidden = !open;
    title.setAttribute('aria-expanded', String(open));
  }
  // Con sesión, el botón del menú muestra las iniciales de quien entró (MS, AO): sirve también para saber quién es.
  const initialsOf = name => String(name || '').split(/\s+/).filter(w => /^\p{L}/u.test(w)).slice(0, 2).map(w => w[0].toLocaleUpperCase('es')).join('');
  // Páginas sin botón de salir propio (Scout, Cocina, Mi equipo, ficha del punto): con sesión de código, el menú ☰
  // ofrece «Salir» igual (sharedLogout).
  const sharedOut = () => !document.getElementById('logout') && !document.getElementById('codeLogout') && !!user();
  // Salir es global: en cualquier página cierra TODAS las sesiones de este navegador (cada panel tiene la suya: si
  // quedara una, al volver al inicio la sesión general te haría entrar de nuevo), también la de administración, y
  // termina en el inicio.
  const CODE_LOGOUTS = ['/api/full/logout', '/api/runner/logout', '/api/builder/logout', '/api/agent/logout', '/api/supplier/logout', '/api/cook/logout', '/api/desk/logout'];
  async function endSessions() {
    // La sesión de administración (si este navegador la tiene abierta) también: pide su token y la cierra.
    const adminOut = fetch('/api/admin/session', {credentials: 'same-origin', cache: 'no-store'}).then(r => r.ok ? r.json() : null).then(b => b?.csrf && fetch('/api/admin/logout', {method: 'POST', credentials: 'same-origin', headers: {'Content-Type': 'application/json', 'X-CSRF-Token': b.csrf}, body: '{}'})).catch(() => {});
    await Promise.all([adminOut, ...CODE_LOGOUTS.map(u => fetch(u, {method: 'POST', credentials: 'same-origin', headers: {'Content-Type': 'application/json'}, body: '{}'}).catch(() => {}))]);
    adminOK = false;
    try { ['codeUser', 'runnerCode', 'runnerName', 'runnerTelephone', 'builderCode', 'builderName', 'builderTelephone', 'agentCode'].forEach(k => localStorage.removeItem(k)); } catch {}
  }
  async function logoutAll() { await endSessions(); location.href = '/'; }
  async function sharedLogout() {
    if (!await confirm('¿Cerrar sesión?', {yes: 'Sí, salir', detail: 'Tendrás que volver a entrar con tu código de acceso.'})) return;
    await logoutAll();
  }
  function signedName() {
    const out = document.getElementById('logout') || document.getElementById('codeLogout');
    if (sharedOut()) return user()?.name || '';
    if (!out || out.closest('[hidden]')) return '';
    return document.getElementById('whoName')?.title || document.getElementById('whoName')?.textContent?.trim() || user()?.name || '';
  }
  function drawMenuButton() {
    const btn = document.getElementById('mMenu');
    if (!btn) return;
    const name = signedName(), ini = initialsOf(name);
    btn.classList.toggle('me', !!ini);
    btn.innerHTML = ini ? `<span class="minit">${esc(ini)}</span>` : svg('<path d="M4 7h16M4 12h16M4 17h16"/>');
    btn.setAttribute('aria-label', ini ? `Menú de ${name}` : 'Menú');
    btn.title = ini ? name : 'Menú';
  }
  addEventListener('resize', () => fitMobileBar());
  function showSheet(open) {
    const sheet = document.getElementById('mSheet'), btn = document.getElementById('mMenu');
    if (!sheet) return;
    if (open) {
      const u = user(), here = location.pathname === '/index.html' ? '/' : location.pathname, out = document.getElementById('logout') || document.getElementById('codeLogout');
      const name = document.getElementById('whoName')?.title || u?.name || '';
      const signedIn = !!out && !out.closest('[hidden]') || sharedOut(), login = !signedIn && document.getElementById('loginBtn'), copyId = signedIn && document.getElementById('copyMyId');
      const list = u ? pages.filter(([url, , , ok]) => url === here || ok(u)) : [];
      const outLabel = out && out.textContent.trim() !== '✕' ? out.textContent.trim() : 'Salir';
      // Roles de quien entró (chips bajo su nombre) y las páginas de poco uso (Mi equipo), solo en este menú.
      const roleNames = {runner: 'Runner', builder: 'Constructor', agent: 'Agente', donations: 'Donaciones', supplier: 'Supply', cook: 'Cocina', teammanager: 'Team manager', admin: 'Admin'};
      const roleChips = signedIn && u?.roles?.length ? `<span class="msroles">${u.roles.map(r => `<span class="msrole">${esc(roleNames[r] || r)}</span>`).join('')}</span>` : '';
      const menuOnly = signedIn && u ? pages.filter(([url, , , ok, where]) => where === 'menu' && url !== here && ok(u)).map(([url, icon, label]) => `<nav class="mpages" aria-label="Más"><a href="${url}"><span aria-hidden="true">${icon}</span>${esc(label(u))}</a></nav>`).join('') : '';
      sheet.innerHTML = `${signedIn ? `<div class="msuser"><b>${esc(name)}${roleChips}</b>${copyId ? '<button type="button" data-m="copyid" class="msoft">Copiar mi ID</button>' : ''}<button type="button" data-m="logout">${esc(outLabel)}</button></div>` : ''}${menuOnly}${list.length > 1 && document.getElementById('mBar')?.classList.contains('tight') ? `<nav class="mpages" aria-label="Páginas">${list.map(([url, icon, label]) => `<a href="${url}"${url === here ? ' aria-current="page"' : ''}><span aria-hidden="true">${icon}</span>${esc(label(u))}</a>`).join('')}</nav>` : ''}<div class="msize"><span class="lbl">Tamaño del texto</span><button type="button" data-msize="-1" aria-label="Texto más pequeño">A−</button><output>${Math.round(zoom * 100)}%</output><button type="button" data-msize="1" aria-label="Texto más grande">A+</button></div>${(adminOK || (u?.roles || []).includes('admin')) && here !== '/admin.html' ? '<a href="/admin.html" class="madmin" data-admin-link><span aria-hidden="true">🔒</span> Administración</a>' : ''}`;
      sheet.querySelector('[data-msize="-1"]').disabled = zoom === sizes[0];
      sheet.querySelector('[data-msize="1"]').disabled = zoom === sizes[sizes.length - 1];
    }
    sheet.hidden = !open;
    btn.setAttribute('aria-expanded', String(open));
  }
  document.addEventListener('click', e => {
    const sheet = document.getElementById('mSheet');
    if (!sheet) return;
    if (e.target.closest('#mTitle')) { const d = document.getElementById('mDrop'); showSheet(false); showDrop(d.hidden); return; }
    if (!e.target.closest('#mDrop')) showDrop(false);
    if (e.target.closest('#mMenu')) { showSheet(sheet.hidden); return; }
    const size = e.target.closest('[data-msize]');
    if (size) { zoom = sizes[Math.min(sizes.length - 1, Math.max(0, sizes.indexOf(zoom) + Number(size.dataset.msize)))]; try { localStorage.setItem('textZoom', String(zoom)); } catch {} applySize(); showSheet(true); return; }
    if (e.target.closest('[data-m="logout"]')) { showSheet(false); const out = document.getElementById('logout') || document.getElementById('codeLogout'); if (out) out.click(); else sharedLogout(); return; }
    if (e.target.closest('#mLogin')) { showSheet(false); document.getElementById('loginBtn')?.click(); return; }
    if (e.target.closest('[data-m="login"]')) { showSheet(false); document.getElementById('loginBtn')?.click(); return; }
    const cid = e.target.closest('[data-m="copyid"]');
    if (cid) { document.getElementById('copyMyId')?.click(); cid.textContent = 'ID copiado'; return; }
    if (e.target.closest('#mAdd')) { showSheet(false); document.querySelector('.top .topright .add')?.click(); return; }
    const sb = e.target.closest('#mSearch');
    if (sb && document.querySelector('header input[type=search][data-global]')) { showSheet(false); openGlobalSearch(document.querySelector('header input[type=search]').value); return; }
    if (sb) { const top = document.querySelector('header.top'), open = !top.classList.contains('msearch'); top.classList.toggle('msearch', open); sb.setAttribute('aria-expanded', String(open)); showSheet(false); if (open) top.querySelector('input[type=search]')?.focus(); return; }
    if (!sheet.hidden && (!sheet.contains(e.target) || e.target.closest('a'))) showSheet(false);
  });
  document.addEventListener('keydown', e => { if (e.key === 'Escape') { showSheet(false); showDrop(false); } });

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
  // Acceso al día: cada página con sesión consulta la suya cada 45 s (y al volver a la pestaña). Los roles nuevos
  // aparecen en el menú; si la persona perdió el acceso a esta página (o la desactivaron), vuelve al inicio con un aviso.
  // Solo actúa después de que la página tuvo sesión (no saca a quien está en la entrada escribiendo su código).
  function saveAccess(b) {
    try { const u = JSON.parse(localStorage.getItem('codeUser') || 'null') || {}; localStorage.setItem('codeUser', JSON.stringify({...u, id: b.id || u.id, name: b.name || u.name, roles: b.roles || u.roles || [], panels: b.panels})); } catch {}
    drawAll();
    dispatchEvent(new Event('accesschange'));
  }
  function watchAccess(url, page) {
    let armed = false, busy = false;
    const leave = () => { location.href = '/?acceso=0'; };
    async function check() {
      if (document.hidden || busy) return;
      busy = true;
      try {
        const r = await fetch(url, {credentials: 'same-origin', cache: 'no-store'});
        if (r.ok) {
          const b = await r.json().catch(() => null);
          if (!b) return;
          armed = true;
          if (Array.isArray(b.panels)) { saveAccess(b); if (page && !b.panels.some(p => p.url === page)) leave(); }
          return;
        }
        if (r.status !== 401 || !armed || !page) return;
        // La sesión de esta página terminó (se quitó el rol, se desactivó al usuario o venció): con el código guardado
        // se vuelve a entrar; si esta página ya no es suya, al inicio.
        let code = rememberedCode();
        try { code = code || codeKeys.map(k => localStorage.getItem(k)).find(Boolean) || ''; } catch {}
        if (code) {
          const s = await fetch('/api/session/code', {method: 'POST', credentials: 'same-origin', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({accessCode: code})});
          const b = await s.json().catch(() => ({}));
          if (s.ok && Array.isArray(b.panels)) { saveAccess(b); if (!b.panels.some(p => p.url === page)) leave(); return; }
        }
        try { localStorage.removeItem('codeUser'); codeKeys.forEach(k => localStorage.removeItem(k)); } catch {}
        leave();
      } catch {
        // Sin conexión: se vuelve a revisar en la próxima vuelta.
      } finally { busy = false; }
    }
    setInterval(check, 45000);
    document.addEventListener('visibilitychange', () => { if (!document.hidden) check(); });
    setTimeout(check, 1500);
  }
  // Al llegar al inicio porque cambió el acceso: aviso breve y se limpia la dirección.
  if (new URLSearchParams(location.search).get('acceso') === '0') {
    const show = () => {
      const t = document.createElement('div');
      t.className = 'accesstoast'; t.setAttribute('role', 'status');
      t.textContent = 'Tu acceso cambió: ya no puedes usar esa página. Tus páginas están en el menú.';
      document.body.appendChild(t);
      setTimeout(() => t.remove(), 7000);
      window.history.replaceState(null, '', location.pathname + location.hash);
    };
    if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', show); else show();
  }
  // Nombre corto para encabezados: nombre + primeras letras del apellido («Arianne Ori…»).
  const shortName = n => { const w = String(n || '').trim().split(/\s+/).filter(Boolean); return w.length < 2 ? w.join('') : w[0] + ' ' + (w[1].length > 3 ? w[1].slice(0, 3) + '…' : w[1]); };
  const drawAll = () => { draw(); drawMobileBar(); };
  // Entorno de desarrollo (localhost, red local, *.local, túnel ngrok a la máquina local): franja ámbar arriba, «DEV» en la esquina y «DEV ·» en el
  // título de la pestaña, para no confundirlo con el sitio público. admin.html también carga nav.js.
  const DEV_HOST=/^(localhost|127\.|0\.0\.0\.0$|::1$|\[::1\]$|10\.|192\.168\.|172\.(1[6-9]|2\d|3[01])\.)|\.(local|localhost|test)$|\.ngrok(-free)?\.(app|dev|io)$/;function devMark(){if(!DEV_HOST.test(location.hostname)||document.querySelector('.devstripe'))return;document.head.insertAdjacentHTML('beforeend','<style>.devstripe{position:fixed;top:0;left:0;right:0;height:4px;z-index:99990;pointer-events:none;background:repeating-linear-gradient(-45deg,#ffb300 0 8px,#10263a 8px 16px)}.devpill{position:fixed;top:4px;left:4px;z-index:99991;padding:1px 6px;border-radius:0 0 6px 6px;background:#ffb300;color:#10263a;font:900 10px/1.4 system-ui,sans-serif;letter-spacing:.08em;pointer-events:auto;cursor:help;box-shadow:0 1px 3px #0005}</style>');document.body.insertAdjacentHTML('afterbegin','<div class="devstripe" aria-hidden="true"></div><span class="devpill" title="Entorno de desarrollo: '+location.host+'. No es el sitio público.">DEV</span>');if(!document.title.startsWith('DEV · '))document.title='DEV · '+document.title}
  // «Buscar» del encabezado: una ✕ adentro borra lo escrito, y lo escrito se guarda por pestaña y por página, así que al
  // volver (Atrás, o desde el menú de páginas) la búsqueda sigue puesta. Al restaurar, en el teléfono se deja abierta la
  // barra de búsqueda.
  function mountSearch() {
    const input = document.querySelector('header input[type=search]');
    if (!input || input.dataset.tsReady) return;
    input.dataset.tsReady = '1';
    let box = input.parentElement;
    if (input.classList.contains('topsearch')) {
      box = document.createElement('span');
      box.className = 'topsearch tswrap';
      input.classList.remove('topsearch');
      input.classList.add('tsinput');
      input.before(box);
      box.appendChild(input);
    } else box.classList.add('tsbox');
    const x = document.createElement('button');
    x.type = 'button';
    x.className = 'tsclear';
    x.setAttribute('aria-label', 'Borrar búsqueda');
    x.title = 'Borrar búsqueda';
    x.textContent = '✕';
    box.appendChild(x);
    // Búsqueda global (Runners, Supply: data-global): Enter abre «Búsqueda global» encima de la página; no filtra la
    // lista ni se guarda (cada página tiene su «Buscar en …» por pestaña).
    if (input.dataset.global) {
      x.hidden = true;
      input.addEventListener('input', () => { x.hidden = !input.value; });
      input.addEventListener('keydown', e => { if (e.key === 'Enter') { e.preventDefault(); openGlobalSearch(input.value); } });
      x.onclick = () => { input.value = ''; x.hidden = true; input.focus(); };
      return;
    }
    // Una sola búsqueda para todas las páginas (por pestaña): lo que se escribe en una sigue puesto al ir a otra.
    const key = 'search:all';
    // Aviso flotante abajo mientras hay búsqueda («🔎 Filtrado por «…» ✕»): así se entiende por qué la lista está
    // filtrada aunque la búsqueda se haya escrito en otra página o la barra esté oculta. ✕ la borra en todas.
    let pill = document.getElementById('searchPill');
    if (!pill) {
      pill = document.createElement('div');
      pill.id = 'searchPill';
      pill.className = 'searchpill';
      pill.hidden = true;
      pill.setAttribute('role', 'status');
      pill.innerHTML = '<button type="button" class="sp-what" title="Ver o cambiar la búsqueda">🔎 Filtrado por «<b></b>»</button><button type="button" class="sp-x" aria-label="Borrar búsqueda (en todas las páginas)" title="Borrar búsqueda">✕</button>';
      document.body.appendChild(pill);
      pill.querySelector('.sp-x').onclick = () => { input.value = ''; input.dispatchEvent(new Event('input', {bubbles: true})); };
      pill.querySelector('.sp-what').onclick = () => { const top = input.closest('header.top'), sb = document.getElementById('mSearch'); if (top && sb && getComputedStyle(sb).display !== 'none') { top.classList.add('msearch'); sb.setAttribute('aria-expanded', 'true'); } input.focus(); input.select?.(); };
    }
    // Con algo escrito se nota aunque la barra esté oculta: el 🔍 del teléfono en naranja con un punto, y la caja con borde.
    const sync = () => { const on = !!input.value.trim(); x.hidden = !input.value; box.classList.toggle('tson', on); pill.hidden = !on; pill.querySelector('b').textContent = input.value.trim().length > 30 ? input.value.trim().slice(0, 30) + '…' : input.value.trim(); const sb = document.getElementById('mSearch'); if (sb) { sb.classList.toggle('tson', on); sb.setAttribute('aria-label', on ? `Buscar (buscando «${input.value.trim()}»)` : 'Buscar'); sb.title = on ? `Buscando «${input.value.trim()}»` : ''; } try { input.value ? sessionStorage.setItem(key, input.value) : sessionStorage.removeItem(key); } catch {} };
    input.addEventListener('input', sync);
    input._tsSync = sync;
    x.onclick = () => { input.value = ''; input.dispatchEvent(new Event('input', {bubbles: true})); input.focus(); };
    let saved = '';
    try { saved = sessionStorage.getItem(key) || ''; } catch {}
    if (saved && !input.value) {
      input.value = saved;
      // Después de que la página conectó su búsqueda: el aviso «input» vuelve a filtrar con lo restaurado.
      const fire = () => setTimeout(() => input.dispatchEvent(new Event('input', {bubbles: true})));
      if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', fire); else fire();
    }
    sync();
  }
  // «🔎 Búsqueda global»: pantalla encima de la página (en el teléfono ocupa todo) con los puntos que coinciden en
  // código, zona, lugar, indicaciones, quién reporta, necesidades o notas, como en Checkup. «‹ Volver», ✕ o Atrás del
  // teléfono la cierran y se vuelve a lo que se estaba haciendo, sin cambiar nada.
  const gs = {points: null, at: 0, pushed: false};
  const gsFold = t => String(t || '').normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase();
  function gsDialog() {
    let d = document.getElementById('gsDlg');
    if (d) return d;
    document.body.insertAdjacentHTML('beforeend', `<dialog id="gsDlg" class="gsdlg" aria-labelledby="gsTitle"><div class="gs-top"><button type="button" class="gs-back" data-gs-close>‹ Volver</button><b id="gsTitle">🔎 Búsqueda global</b><button type="button" class="gs-x" data-gs-close aria-label="Cerrar">✕</button></div><div class="gs-bar"><input type="search" id="gsInput" placeholder="Código, zona, lugar, necesidad, quién reporta…" autocomplete="off" aria-label="Buscar en todos los puntos"></div><p class="gs-count" id="gsCount" role="status"></p><div class="gs-list" id="gsList"></div></dialog>`);
    d = document.getElementById('gsDlg');
    d.addEventListener('click', e => { if (e.target.closest('[data-gs-close]') || e.target === d) d.close(); });
    d.querySelector('#gsInput').addEventListener('input', gsDraw);
    d.addEventListener('close', () => { if (gs.pushed) { gs.pushed = false; window.history.back(); } });
    addEventListener('popstate', () => { if (d.open) { gs.pushed = false; d.close(); } });
    return d;
  }
  async function gsLoad() {
    if (gs.points && Date.now() - gs.at < 60000) return;
    // Los mismos datos que Scout: puntos publicados (y los por revisar propios o del equipo) con sus necesidades activas;
    // contactos solo para quien es de un equipo.
    try { const r = await fetch('/api/full/scout', {credentials: 'same-origin', cache: 'no-store'}); gs.points = r.ok ? await r.json() : []; gs.at = Date.now(); } catch { gs.points = gs.points || []; }
  }
  const gsAgo = v => { const m = Math.round((Date.now() - Date.parse(v)) / 60000); return !v || isNaN(m) ? '' : m < 60 ? `hace ${Math.max(1, m)} min` : m < 1440 ? `hace ${Math.round(m / 60)} h` : `hace ${Math.round(m / 1440)} d`; };
  const gsChips = t => window.PointForm?.needChipsHtml ? PointForm.needChipsHtml(t) : `<p>${esc(t)}</p>`;
  // Tarjeta como una fila de Checkup: cuándo fue el último checkup, nombre y código, urgente / por revisar, lugar, quién
  // reporta, personas, contactos y las necesidades activas con su estado; «Ficha» y Maps.
  function gsCard(p) {
    const pending = p.pointStatus === 'pending', needs = p.activeNeeds || [], maps = p.coords?.length === 2 ? `https://www.google.com/maps/search/?api=1&query=${p.coords.join(',')}` : p.gps || '';
    const when = p.lastCheckup ? `Checkup ${gsAgo(p.lastCheckup)}` : `Creado ${gsAgo(p.created)}`;
    const contacts = (p.contacts || []).map(c => c.name).filter(Boolean);
    return `<article class="gs-card"><small class="gs-when">${esc(when)}</small><h3>${esc([p.area, p.place].filter(Boolean).join(' · ') || p.code)}${p.code && !/^P-[0-9a-f]{12}$/i.test(p.code) ? ` <small class="gs-code">${esc(p.code)}</small>` : ''}${p.highestUrgency === 'urgente' ? ' <span class="gs-badge urg">Urgente</span>' : ''}${pending ? ' <span class="gs-badge pend">Por revisar</span>' : ''}${p.coords?.length === 2 ? '' : ' <span class="gs-badge">📍 sin GPS</span>'}</h3><p class="gs-meta">${[p.directions && '📍 ' + esc(p.directions), p.coordinator && '📣 ' + esc(p.coordinator), p.people && '👥 ' + esc(p.people) + ' personas'].filter(Boolean).join(' · ')}</p>${contacts.length ? `<p class="gs-meta">👤 ${esc(contacts.slice(0, 3).join(', '))}${contacts.length > 3 ? ` y ${contacts.length - 3} más` : ''}</p>` : ''}${needs.length ? `<ul class="gs-needs">${needs.map(n => `<li><span class="gs-st${n.status === 'IN_PROGRESS' ? ' route' : ''}">${n.status === 'IN_PROGRESS' ? '🏃 En camino' + (n.runnerName ? ' · ' + esc(n.runnerName) : '') : 'Necesita'}</span>${n.urgency === 'urgente' ? '<span class="gs-badge urg">Urgente</span>' : ''}${gsChips(n.need)}</li>`).join('')}</ul>` : pending && p.needs ? `<div class="gs-needs">${gsChips(p.needs)}</div>` : '<p class="gs-meta">Sin necesidades abiertas</p>'}<div class="gs-acts"><a class="gs-btn" href="/camp.html?id=${encodeURIComponent(p.id)}">Ficha ›</a>${maps ? `<a class="gs-btn" target="_blank" rel="noopener noreferrer" href="${esc(maps)}">Abrir en Maps ↗</a>` : ''}</div></article>`;
  }
  function gsDraw() {
    const raw = document.getElementById('gsInput').value.trim(), q = gsFold(raw), list = document.getElementById('gsList'), count = document.getElementById('gsCount');
    if (!gs.points) { count.textContent = 'Cargando…'; list.innerHTML = ''; return; }
    if (!q) { count.textContent = 'Escribe para buscar en todos los puntos: código, zona, lugar, necesidad, quién reporta o un contacto.'; list.innerHTML = ''; return; }
    const hits = gs.points.filter(p => gsFold([p.code, p.area, p.place, p.directions, p.coordinator, p.needs, p.notes, p.extra, p.addedByMate, ...(p.activeNeeds || []).map(n => n.need), ...(p.contacts || []).map(c => c.name)].join(' ')).includes(q));
    hits.sort((a, b) => (b.highestUrgency === 'urgente') - (a.highestUrgency === 'urgente') || (b.activeNeeds || []).length - (a.activeNeeds || []).length);
    count.textContent = hits.length ? `${hits.length} ${hits.length === 1 ? 'punto' : 'puntos'} con «${raw}»` : `Ningún punto coincide con «${raw}».`;
    list.innerHTML = hits.slice(0, 60).map(gsCard).join('') + (hits.length > 60 ? `<p class="gs-count">y ${hits.length - 60} más: escribe algo más preciso.</p>` : '');
  }
  async function openGlobalSearch(q) {
    const d = gsDialog(), input = d.querySelector('#gsInput');
    input.value = String(q || '').trim();
    if (!d.open) { d.showModal(); window.history.pushState({gs: 1}, ''); gs.pushed = true; }
    gsDraw(); await gsLoad(); gsDraw(); input.focus();
  }
  // Ayuda de una pestaña o pantalla: nota junto a lo que se tocó, con «No mostrar de nuevo» (recordado en este
  // teléfono). again=true (tocar la pestaña en la que ya estás, o ❓) la muestra aunque se haya ocultado.
  const tipHidden = key => { try { return localStorage.getItem('tip:' + key) === '1'; } catch { return false; } };
  const closeTip = () => document.querySelectorAll('.navtip').forEach(x => x.remove());
  function tip(anchor, key, title, html, again) {
    closeTip();
    if (!anchor || (!again && tipHidden(key))) return;
    const d = document.createElement('div');
    d.className = 'navtip'; d.setAttribute('role', 'dialog'); d.setAttribute('aria-label', title);
    d.innerHTML = `<b class="navtip-t">${title}</b><div class="navtip-b">${html}</div><div class="navtip-a"><button type="button" data-tip="never">No mostrar de nuevo</button><button type="button" class="ok" data-tip="ok">Entendido</button></div>`;
    (anchor.closest('dialog') || document.body).appendChild(d);
    const vw = document.documentElement.clientWidth, vh = document.documentElement.clientHeight, w = Math.min(340, vw - 24), r = anchor.getBoundingClientRect();
    d.style.width = w + 'px';
    d.style.left = Math.max(12, Math.min(vw - w - 12, r.left + r.width / 2 - w / 2)) + 'px';
    const below = r.bottom + 8, h = d.offsetHeight;
    d.style.top = (below + h > vh - 8 && r.top - h - 8 > 8 ? r.top - h - 8 : Math.max(8, Math.min(below, vh - h - 8))) + 'px';
    d.onclick = e => {
      const b = e.target.closest('[data-tip]'); if (!b) return;
      if (b.dataset.tip === 'never') try { localStorage.setItem('tip:' + key, '1'); } catch {}
      closeTip();
    };
    setTimeout(() => document.addEventListener('pointerdown', function out(e) {
      if (e.target.closest('.navtip')) return;
      document.removeEventListener('pointerdown', out, true); closeTip();
    }, true), 0);
  }
  // Deslizar a los lados sobre la lista (teléfono) pasa a la pestaña siguiente / anterior de tabs (las visibles), como
  // tocarla. No cuenta si el gesto empieza en un campo, un mapa, un diálogo o algo que se desplaza a los lados.
  function swipeTabs(area, tabs) {
    if (!area || !tabs) return;
    let x0 = 0, y0 = 0, t0 = 0, on = false;
    area.addEventListener('touchstart', e => {
      on = false;
      if (e.touches.length !== 1 || !matchMedia('(max-width:760px)').matches) return;
      const t = e.target;
      if (t.closest('input,textarea,select,.leaflet-container,dialog,[data-noswipe]')) return;
      for (let el = t; el && el !== area; el = el.parentElement) if (el.scrollWidth > el.clientWidth + 2 && /auto|scroll/.test(getComputedStyle(el).overflowX)) return;
      x0 = e.touches[0].clientX; y0 = e.touches[0].clientY; t0 = Date.now(); on = true;
    }, {passive: true});
    area.addEventListener('touchend', e => {
      if (!on) return; on = false;
      const c = e.changedTouches[0], dx = c.clientX - x0, dy = c.clientY - y0;
      if (Math.abs(dx) < 60 || Math.abs(dx) < Math.abs(dy) * 1.7 || Date.now() - t0 > 800) return;
      const list = [...tabs.querySelectorAll('[role=tab]')].filter(b => b.offsetParent !== null && !b.disabled);
      const i = list.findIndex(b => b.getAttribute('aria-selected') === 'true');
      const next = list[i + (dx < 0 ? 1 : -1)];
      if (i < 0 || !next) return;
      next.click();
      next.scrollIntoView({block: 'nearest', inline: 'nearest'});
      area.classList.remove('swipe-l', 'swipe-r'); void area.offsetWidth; area.classList.add(dx < 0 ? 'swipe-l' : 'swipe-r');
    }, {passive: true});
  }
  // Toda página con pestañas (#mainTabs) y su lista (#feed o #list): deslizar cambia de pestaña en el teléfono.
  const wireSwipe = () => swipeTabs(document.getElementById('feed') || document.getElementById('list'), document.getElementById('mainTabs'));
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', wireSwipe); else wireSwipe();
  window.AppNav = {tip, swipeTabs, draw: drawAll, openGlobalSearch, endSessions, logoutAll, confirm, choose, phone, shortName, codeIntro, isCode, codeOf, codeSignedIn, prefillCode, signinMode, createAccount, showNewCode, watchAccess};
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', devMark); else devMark();
  mountSearch();
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', () => { draw(); mountTextSize(); mountMobileBar(); mountSearch(); document.querySelector('header input[type=search]')?._tsSync?.(); }); else { draw(); mountTextSize(); mountMobileBar(); document.querySelector('header input[type=search]')?._tsSync?.(); }
  // Entrar o salir en otra pestaña actualiza esta.
  addEventListener('storage', e => { if (e.key === 'codeUser' || e.key === null) draw(); if (e.key === 'textZoom' || e.key === null) { zoom = sizes.includes(savedSize()) ? savedSize() : 1; applySize(); } });
})();
