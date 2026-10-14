// Formulario compartido de puntos (público y administración): «¿Dónde estás?», «¿Con quién estás?»,
// «¿Qué necesitas?» y, solo para administración, los campos internos.
(function () {
  const esc = s => String(s ?? '').replace(/[&<>"']/g, c => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
  const svg = path => `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${path}</svg>`;
  const icons = {
    pin: svg('<path d="M12 21s-7-6.2-7-11a7 7 0 0 1 14 0c0 4.8-7 11-7 11z"/><circle cx="12" cy="10" r="2.5"/>'),
    people: svg('<circle cx="9" cy="8" r="3"/><path d="M3 20a6 6 0 0 1 12 0"/><circle cx="17" cy="9" r="2.5"/><path d="M15.5 14.2A5 5 0 0 1 21 19"/>'),
    list: svg('<path d="M9 6h11M9 12h11M9 18h11"/><circle cx="4.5" cy="6" r="1"/><circle cx="4.5" cy="12" r="1"/><circle cx="4.5" cy="18" r="1"/>'),
    lock: svg('<rect x="5" y="11" width="14" height="9" rx="2"/><path d="M8 11V8a4 4 0 0 1 8 0v3"/>'),
    plus: svg('<path d="M12 5v14M5 12h14"/>'),
    x: svg('<path d="M6 6l12 12M18 6L6 18"/>'),
    coordinator: svg('<circle cx="9" cy="8" r="3.5"/><path d="M3 20a6 6 0 0 1 12 0M16 11l2 2 4-4"/>'),
    adult: svg('<circle cx="12" cy="4.5" r="2.5"/><path d="M12 8v7M8 11h8M12 15l-3 6M12 15l3 6"/>'),
    child: svg('<circle cx="12" cy="8" r="2.5"/><path d="M12 11v5M9 13h6M12 16l-2 4M12 16l2 4"/>'),
    baby: svg('<circle cx="12" cy="12" r="8"/><path d="M9.5 11h.01M14.5 11h.01M9.5 14.5a3.5 3.5 0 0 0 5 0"/>'),
    elder: svg('<circle cx="10" cy="4.5" r="2.2"/><path d="M10 7.5l-1 6 3 3v5M9 13.5l-3 7.5M10 9l4 2M16 11v10"/>'),
    disabled: svg('<circle cx="11" cy="4" r="2"/><path d="M11 7v6h5l3 6M11 10h5"/><path d="M8 11.5a5 5 0 1 0 6.5 6.5"/>'),
    drop: svg('<path d="M12 3s6 7 6 11a6 6 0 0 1-12 0c0-4 6-11 6-11z"/>'),
    food: svg('<path d="M7 3v8M5 3v5a2 2 0 0 0 4 0V3M7 11v10M17 3c-2 2-2 6 0 8v10"/>'),
    tent: svg('<path d="M3 20L12 4l9 16H3zM9 20l3-6 3 6"/>'),
    repair: svg('<path d="M2 12l6.5-5.5L15 12"/><path d="M4 10.5V20h9v-9.5"/><path d="M7 20v-3.5h3V20"/><path d="M16 5h6v3h-6z"/><path d="M19 8v13"/>'),
    pot: svg('<path d="M4 10h16v6a4 4 0 0 1-4 4H8a4 4 0 0 1-4-4v-6zM2 10h2M20 10h2M9 6c0-1 1-1 1-2M14 6c0-1 1-1 1-2"/>'),
    bottle: svg('<path d="M10 2h4M10 5h4l1 3v12a2 2 0 0 1-2 2h-2a2 2 0 0 1-2-2V8l1-3zM9 12h6"/>'),
    fuel: svg('<path d="M4 21V5a2 2 0 0 1 2-2h6a2 2 0 0 1 2 2v16M3 21h12M14 9h2a2 2 0 0 1 2 2v6a1.5 1.5 0 0 0 3 0V8l-3-3M7 7h4"/>'),
    locate: svg('<circle cx="12" cy="12" r="4"/><circle cx="12" cy="12" r="8"/><path d="M12 1v3M12 20v3M1 12h3M20 12h3"/>'),
    other: svg('<circle cx="12" cy="12" r="9"/><path d="M12 8v8M8 12h8"/>'),
    // Lista de empaque.
    box: svg('<path d="M3 7l9-4 9 4v10l-9 4-9-4V7z"/><path d="M3 7l9 4 9-4M12 11v10"/>'),
    foodbag: svg('<path d="M5 8h14l-1 13H6L5 8z"/><path d="M9 8V6a3 3 0 0 1 6 0v2"/><path d="M9 13h6M10 16.5h4"/>'),
    hygiene: svg('<rect x="3" y="11" width="10" height="10" rx="2"/><path d="M6 11V8h4v3M8 8V5h3"/><path d="M18 21V10M16 3h4v7h-4z"/>'),
    cleaning: svg('<path d="M15 2l-3 9"/><path d="M7 11h9l2 10H5l2-10z"/><path d="M9 15v6M13 15v6"/>'),
    flashlight: svg('<path d="M7 3h10v4l-2 3v11H9V10L7 7V3z"/><path d="M7 7h10M12 13v3"/>'),
    repellent: svg('<rect x="6" y="9" width="9" height="12" rx="2"/><path d="M8.5 9V6h4v3M12.5 6h2.5"/><path d="M18 4h.01M20.5 6h.01M18 8h.01"/>'),
    chair: svg('<path d="M8 3v9h8V3"/><path d="M6 12h12M7 12l-1 9M17 12l1 9M7.5 16h9"/>'),
    hammock: svg('<path d="M3 4v17M21 4v17"/><path d="M3 8c3 7 15 7 18 0"/><path d="M3 8l3 4M21 8l-3 4"/>'),
    net: svg('<path d="M12 3L4 20h16L12 3z"/><path d="M12 3v17M8.5 20L12 3l3.5 17M7.3 13.5h9.4"/>'),
    plug: svg('<path d="M9 2v5M15 2v5"/><path d="M6 7h12v4a6 6 0 0 1-12 0V7z"/><path d="M12 17v5"/>'),
    camping: svg('<path d="M12 3c2.5 3 4 4.5 4 7.5a4 4 0 0 1-8 0c0-2 1-3 2-4 0 1.8.8 2.8 2 3 0-2-1-4 0-6.5z"/><path d="M4 21l16-4M20 21L4 17"/>'),
    towel: svg('<path d="M4 4h16v4H4z"/><path d="M6 8v13h12V8"/><path d="M6 15h12"/>'),
    poncho: svg('<path d="M12 3a3 3 0 0 0-3 3v1L3 18h18L15 7V6a3 3 0 0 0-3-3z"/><path d="M9 7h6M12 18v3"/>'),
    umbrella: svg('<path d="M3 12a9 9 0 0 1 18 0H3z"/><path d="M12 12v7a2 2 0 0 0 4 0M12 3V2"/>'),
    medicine: svg('<rect x="3" y="7" width="18" height="13" rx="2"/><path d="M9 7V4h6v3M12 10.5v6M9 13.5h6"/>'),
    utensils: svg('<path d="M6 3v6a2 2 0 0 0 4 0V3M8 3v18"/><path d="M17 3c-2 0-3 2-3 5s1 4 3 4 3-1 3-4-1-5-3-5zM17 12v9"/>'),
    disposable: svg('<path d="M6 4h12l-2 17H8L6 4z"/><path d="M5 4h14M6.6 9h10.8"/>'),
    mattress: svg('<rect x="2" y="8" width="20" height="9" rx="3"/><path d="M5 17v2.5M19 17v2.5M6.5 12.5h.01M10 12.5h.01M14 12.5h.01M17.5 12.5h.01"/>')
  };
  const style = `.pf-thumb{position:relative;margin:0;width:76px}.pf-thumb a{display:grid;place-items:center;width:76px;height:76px;border-radius:9px;overflow:hidden;background:#e7eef0;border:1px solid #d5e0e2;text-decoration:none}.pf-thumb img{width:100%;height:100%;object-fit:cover}.pf-thumb-icon{font-size:26px;color:#173c46}
.pfp{border:0;border-radius:14px;padding:0;width:min(460px,94vw);max-height:92dvh;color:#173c46;font:inherit}.pfp::backdrop{background:#10263ab8}.pfp-box{padding:20px;display:flex;flex-direction:column;gap:6px}.pfp h2{margin:0;font-size:20px}.pfp h3{margin:6px 0 0;font-size:16px}.pfp-for,.pfp-hint{margin:0;color:#667982;font-size:13px}
.pfp label{display:block;font-size:12px;font-weight:750;margin:8px 0 0}.pfp input{width:100%;padding:11px;border:1px solid #cad9d9;border-radius:9px;font:inherit;background:#fff;box-sizing:border-box}
.pfp-options{max-height:min(300px,40dvh);overflow:auto;margin-top:6px;border:1px solid #e1e9ea;border-radius:10px}.pfp-options button{display:flex;align-items:center;gap:9px;width:100%;margin:0;padding:10px 12px;border:0;border-radius:0;border-bottom:1px solid #eef3f3;background:#fff;color:#173c46;font:inherit;font-weight:650;text-align:left;cursor:pointer}.pfp-options button:last-child{border-bottom:0}.pfp-options button.hl,.pfp-options button:hover{background:#eef6f4}.pfp-options svg,.pfp-line svg,.pfp-chosen svg{flex:none;color:#2d7f86}.pfp-options .pfp-new{color:#19745f}
.pfp-chosen{display:flex;align-items:center;gap:8px;font-size:16px;margin-top:4px}.pfp-qtyrow{display:flex;gap:6px}.pfp-qtyrow input{text-align:center;font-size:18px;font-weight:800}.pfp-step{flex:none;width:48px;margin:0;border:1px solid #cad9d9;border-radius:9px;background:#fff;font-size:22px;font-weight:800;color:#173c46;cursor:pointer}
.pfp-list{display:flex;flex-direction:column;gap:4px;margin-top:8px}.pfp-line{display:flex;align-items:center;gap:8px;padding:7px 9px;border:1px solid #e1e9ea;border-radius:9px;background:#fbfdfd}.pfp-line span{flex:1;font-weight:650;overflow-wrap:anywhere}.pfp-line input{width:86px;padding:6px;text-align:center}.pfp-del{margin:0;border:0;background:none;color:#9a4536;font-size:15px;padding:4px 6px;cursor:pointer}
.pfp-btn{border:0;border-radius:10px;background:#19745f;color:#fff;font:inherit;font-weight:800;padding:12px 16px;margin:8px 0 0;width:100%;cursor:pointer}.pfp-btn.pfp-plain{background:#eef3f3;color:#173c46}.pfp-btn:disabled{opacity:.5}.pfp-actions{display:flex;gap:8px;flex-wrap:wrap;margin-top:8px}.pfp-actions .pfp-btn{flex:1 1 120px;width:auto;margin:0}.pfp-msg{margin:0;color:#a33527;font-size:13px}
.pf-pack{margin:8px 0}.pf-pack-title{display:flex;align-items:center;gap:6px;font-size:12px;color:#173c46;margin-bottom:5px}.pf-pack-title svg{width:15px;height:15px;color:#2d7f86}.pf-pack-chips{display:flex;flex-wrap:wrap;gap:5px}.pf-pack-chip{display:inline-flex;align-items:center;gap:5px;border:1px solid #c9dde0;border-radius:16px;padding:3px 9px;font-size:12px;font-weight:650;color:#1f3a46;background:#f3f9f9}.pf-pack-chip svg{width:15px;height:15px;color:#2d7f86;flex:none}.pf-pack-chip b{color:#0f6f63}
.pf-section{border:1px solid #e1e9ea;border-radius:12px;padding:10px 14px 14px;margin:0 0 12px;background:#fbfdfd;min-width:0}
.pf-section legend{display:flex;align-items:center;gap:7px;font-weight:800;font-size:15px;color:#173c46;padding:0 6px}
.pf-section legend svg{color:#2d7f86}.pf-step{display:inline-grid;place-items:center;width:22px;height:22px;border-radius:50%;background:#1d6f7a;color:#fff;font-size:12px;font-weight:800}.pf-admin{background:#fffaf4;border-color:#ecd9c8}.pf-admin legend svg{color:#b5653f}
.pf-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(130px,1fr));gap:10px}.pf-grid .wide{grid-column:1/-1}
.pf label{display:block;margin:0;font-size:12px;font-weight:700;color:#36505a}.pf .pf-ico{display:flex;align-items:center;gap:5px}.pf .pf-ico svg{color:#2d7f86;flex:none}
.pf input,.pf textarea,.pf select{display:block;width:100%;max-width:none;min-height:0;margin-top:4px;padding:9px 10px;border:1px solid #d5e0e2;border-radius:8px;background:#fff;font:inherit;font-weight:400;color:#152b3a}
.pf textarea{min-height:64px;resize:vertical}.pf input:focus,.pf textarea:focus,.pf select:focus{outline:2px solid #2d7f8655;border-color:#2d7f86}
.pf-hint{font-size:12px;color:#5b6f76;margin:2px 0 10px;line-height:1.4}.pf-hint b{color:#9a4c2f}.pf-sub{font-size:12px;font-weight:800;color:#2d7f86;text-transform:uppercase;letter-spacing:.05em;margin:14px 0 8px}
.pf .pf-gps{display:flex;align-items:center;justify-content:center;gap:8px;width:100%;margin:2px 0 0;padding:13px;border:0;border-radius:10px;background:#1d6f7a;color:#fff;font-size:15px;font-weight:800}
.pf .pf-gps:hover{background:#175d66}.pf .pf-gps:disabled{opacity:.6;cursor:wait}.pf-gpsthumb{height:140px;margin:8px 0 0;border-radius:10px;overflow:hidden;border:1px solid #cad9d9;cursor:pointer;background:#dfe8e8}.pf-gpsthumb[hidden]{display:none}.pf-gpsthumb .leaflet-control-attribution{font-size:9px}.pf-thumbpin{width:24px;height:24px;background:#c62828;border:3px solid #fff;border-radius:50% 50% 50% 0;transform:rotate(-45deg);box-shadow:0 2px 7px #0006}.pf-gpsmsg{font-size:12px;color:#36505a;margin:6px 0 0}.pf-gpsmsg:empty{display:none}
.pf-me-btns{display:inline-flex;gap:4px;margin-left:auto}.pf .pf-ico:has(.pf-me-btns){width:100%}.pf .pf-yo{margin:0;padding:2px 9px;border:1px solid #2d7f86;border-radius:12px;background:#eaf6f3;color:#1d6f7a;font:inherit;font-size:11px;font-weight:800;cursor:pointer}.pf .pf-clearcoord{display:inline-flex;margin:0;padding:3px;border:0;border-radius:6px;background:none;color:#8a9ca1;cursor:pointer}.pf .pf-clearcoord svg{width:14px;height:14px}.pf .pf-clearcoord:hover{background:#f2e7e3;color:#9a4536}.pf-members{margin-top:8px}.pf-member{display:grid;grid-template-columns:22px 2fr 1.4fr .7fr 1.3fr 26px;grid-template-areas:'num name phone age ced del';gap:6px;align-items:center;padding:4px 0}.pf-mnum{grid-area:num;font-size:13px;font-weight:800;color:#2d7f86;text-align:right}.pf-m-name{grid-area:name}.pf-m-phone{grid-area:phone}.pf-m-age{grid-area:age}.pf-m-cedula{grid-area:ced}.pf-member .pf-remove{grid-area:del}.pf-member:last-child .pf-remove{visibility:hidden}.pf .pf-member input{min-width:0;margin-top:0}
.pf .pf-remove{display:inline-flex;margin:0;padding:3px;border:0;border-radius:6px;background:none;color:#8a9ca1}.pf .pf-remove:hover{background:#f2e7e3;color:#9a4536}
.pf-needs{display:grid;grid-template-columns:repeat(auto-fill,minmax(170px,1fr));gap:8px}
.pf label.pf-need{position:relative;margin:0;font-weight:650}.pf .pf-need input{position:absolute;opacity:0;width:1px;height:1px;margin:0}
.pf-need span{display:flex;align-items:center;gap:8px;min-height:44px;padding:8px 10px;border:1px solid #d5e0e2;border-radius:10px;background:#fff;color:#36505a;font-size:13px;cursor:pointer}
.pf-need span svg{color:#2d7f86;flex:none}.pf-need span:after{content:"";margin-left:auto;width:16px;height:16px;border:1.5px solid #b7c7ca;border-radius:4px;flex:none}
.pf-need input:checked+span{border-color:#2d7f86;background:#eaf6f3;color:#173c46}.pf-need input:checked+span:after{background:#2d7f86 url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='white' stroke-width='3'%3E%3Cpath d='M5 12l5 5 9-10'/%3E%3C/svg%3E") center/12px no-repeat;border-color:#2d7f86}
.pf-need input:focus-visible+span{outline:2px solid #2d7f86}.pf-repair span svg{color:#1f63c6}.pf-repair input:checked+span{border-color:#1f63c6;background:#e9f1fc}.pf-repair input:checked+span:after{background-color:#1f63c6;border-color:#1f63c6}.pf-need em{font-style:normal}.pf-need em small{display:block;font-size:11px;font-weight:700;color:#ae4f32}.pf-need em small.pf-dup{color:#677c83}.pf-otros[hidden]{display:none}.pf label.pf-otros{margin-top:8px}
.pf .pf-num{display:flex;margin-top:4px}.pf .pf-num input{flex:1;width:auto;min-width:0;margin:0;border-radius:0;text-align:center}
.pf .pf-numbtn{flex:none;width:36px;margin:0;padding:0;border:1px solid #d5e0e2;background:#f1f6f6;color:#1d6f7a;font-size:18px;font-weight:800;line-height:1;cursor:pointer}.pf .pf-numbtn:hover{background:#e2efee}
.pf .pf-dec{order:-1;border-radius:8px 0 0 8px;border-right:0}.pf .pf-inc{border-radius:0 8px 8px 0;border-left:0}
dialog.pf-mapdialog{width:100vw;height:100dvh;max-width:none;max-height:none;margin:0;padding:0;border:0;background:#dfe8e8}dialog.pf-mapdialog::backdrop{background:#10263ab8}
.pf-mapdialog .pf-mapbox{width:100%;height:100%}.pf-mapdialog .pf-mapclose{position:absolute;top:12px;right:12px;z-index:1000;display:flex;margin:0;padding:9px;border:0;border-radius:50%;background:#fff;color:#173c46;box-shadow:0 2px 8px #0004;cursor:pointer}
.pf-mapdialog .pf-mapgps{position:absolute;top:60px;right:12px;z-index:1000;display:flex;align-items:center;gap:6px;margin:0;padding:9px 12px;border:0;border-radius:20px;background:#fff;color:#1f6fd1;font:inherit;font-size:13px;font-weight:800;box-shadow:0 2px 8px #0004;cursor:pointer}.pf-mapdialog .pf-mapgps:disabled{opacity:.6;cursor:wait}
.pf-mapdialog .pf-mappanel{position:absolute;left:12px;right:12px;bottom:22px;z-index:1000;margin:0 auto;max-width:520px;padding:10px 12px;border-radius:12px;background:#fffffff2;color:#173c46;font-size:13px;box-shadow:0 2px 10px #0004}
.pf-mapdialog .pf-mapstatus,.pf-mapdialog .pf-mappick{margin:0 0 6px}.pf-mapdialog .pf-mappick{font-weight:700}
.pf-mapdialog .pf-mapuse{display:block;width:100%;margin:0;padding:11px;border:0;border-radius:9px;background:#1d6f7a;color:#fff;font:inherit;font-size:15px;font-weight:800;cursor:pointer}.pf-mapdialog .pf-mapuse:disabled{opacity:.5;cursor:not-allowed}
.pf-mapdialog .pf-crosshair{position:absolute;top:50%;left:50%;z-index:1000;width:44px;height:44px;transform:translate(-50%,-50%);pointer-events:none;color:#c62828;filter:drop-shadow(0 0 1.5px #fff) drop-shadow(0 0 1.5px #fff)}
.pf-me{width:22px;height:22px;border-radius:50%;background:#1f6fd1;border:3px solid #fff;box-shadow:0 0 0 2px #1f6fd1,0 1px 6px #0006}
@media(max-width:560px){.pf-member{grid-template-columns:22px 1.4fr .7fr 1.3fr 26px;grid-template-areas:'num name name name del' '. phone age ced .';row-gap:4px;padding:6px 0;border-top:1px dashed #dbe5e6}.pf-member:first-child{border-top:0}.pf-needs{grid-template-columns:1fr 1fr}}`;
  if (!document.getElementById('pointform-style')) document.head.insertAdjacentHTML('beforeend', `<style id="pointform-style">${style}</style>`);

  // filter: «count» y «age» solo dígitos (0–999); «phone» dígitos, espacios, +, - y paréntesis.
  const input = (name, label, opt = {}) => `<label class="${opt.wide ? 'wide' : ''}"><span class="${opt.icon ? 'pf-ico' : ''}">${opt.icon ? icons[opt.icon] : ''}${esc(label)}</span>${opt.area
    ? `<textarea name="${name}" placeholder="${esc(opt.ph || '')}"></textarea>`
    : opt.filter === 'count' ? countInput(name, label, opt.ph)
    : `<input name="${name}"${filterAttrs(opt.filter)} placeholder="${esc(opt.ph || '')}" autocomplete="off">`}</label>`;
  // Conteo con botones − / +. El campo va primero en el HTML para que el <label> lo enlace a él y no al botón; el CSS pone «−» a la izquierda.
  const countInput = (name, label, ph) => `<span class="pf-num"><input name="${name}"${filterAttrs('count')} placeholder="${esc(ph || '')}" autocomplete="off"><button type="button" class="pf-numbtn pf-dec" data-step="-1" aria-label="Restar 1 a ${esc(label)}">−</button><button type="button" class="pf-numbtn pf-inc" data-step="1" aria-label="Sumar 1 a ${esc(label)}">+</button></span>`;
  const filterAttrs = f => f === 'count' || f === 'age' ? ` data-filter="${f}" inputmode="numeric" maxlength="3"` : f === 'phone' ? ' data-filter="phone" type="tel" inputmode="tel" maxlength="40"' : '';
  // Los tres pasos llevan número (1, 2, 3); la sección de administración lleva candado.
  const section = (icon, title, body, cls = '') => `<fieldset class="pf-section ${cls}"><legend>${typeof icon === 'number' ? `<span class="pf-step">${icon}</span>` : icons[icon]}${esc(title)}</legend>${body}</fieldset>`;
  // «Reparación de casa» es solo para administración: la atienden constructores y no se publica.
  const NEEDS = [['Agua', 'drop'], ['Alimentos', 'food'], ['Refugio (carpa / lona / toldo)', 'tent'], ['Kit de cocina', 'pot'], ['Kit de bebé', 'bottle'], ['Gasolina para generador', 'fuel'], ['Diésel para generador', 'fuel'], ['Reparación de casa', 'repair', 'pf-repair', 'admin']];
  const COUNTS = [['people', 'Personas', 'people'], ['adults', 'Adultos', 'adult'], ['children', 'Niños', 'child'], ['babies', 'Bebés', 'baby'], ['olderAdults', 'Mayores', 'elder'], ['disabled', 'Discapacitados', 'disabled']];

  // Casillas de necesidades con iconos y «Otros»; también las usa el formulario de necesidades del panel.
  const needsHtml = ({admin = false} = {}) => `<div class="pf-needs">${NEEDS.filter(n => admin || n[3] !== 'admin').map(([l, i, c]) => `<label class="pf-need ${c || ''}"><input type="checkbox" data-need="${esc(l)}"><span>${icons[i]}${esc(l)}</span></label>`).join('')}<label class="pf-need"><input type="checkbox" class="pf-otros-check"><span>${icons.other}Otros</span></label></div><label class="pf-otros" hidden>Otros: ¿qué más necesitan?<textarea class="pf-otros-text" placeholder="Medicinas, ropa, colchones…"></textarea></label>`;
  function html({admin = false, edit = false} = {}) {
    const where = section(1, '¿Dónde estás?', `<button type="button" class="pf-gps">${icons.pin}Enviar mi ubicación GPS</button><div class="pf-gpsthumb" hidden title="Toca para mover la ubicación" aria-label="Vista previa de la ubicación elegida"></div><p class="pf-gpsmsg" role="status"></p><div class="pf-grid" style="margin-top:12px">${input('gps', 'o pega un enlace de Google Maps', {wide: true, ph: 'https://maps.google.com/…'})}${input('area', 'Zona / área', {ph: 'Ej. Cañas'})}${input('place', 'Ubicación / referencia', {ph: 'Ej. Escuela de la comunidad'})}${input('directions', 'Indicaciones de llegada', {wide: true, area: true, ph: 'Cómo llegar, puntos de referencia…'})}</div>`);
    const who = section(2, '¿Con quién estás?', `<div class="pf-grid">${input('coordinator', 'Coordinador', {icon: 'coordinator'})}${input('coordinatorPhone', 'Teléfono del coordinador', {filter: 'phone', ph: 'Ej. 6000-0000'})}</div><p class="pf-sub">¿Cuántas personas?</p><div class="pf-grid">${COUNTS.map(([k, l, i]) => input(k, l, {icon: i, filter: 'count', ph: '0'})).join('')}</div><p class="pf-sub">Personas del grupo</p><p class="pf-hint"><b>Agrega al menos un teléfono</b> (del coordinador o de alguna persona). Nombre, cédula y edad son opcionales y solo los ve el equipo de curación.</p><div class="pf-members"></div>`);
    const what = section(3, '¿Qué necesitas?', `${needsHtml({admin})}<div class="pf-grid" style="margin-top:12px">${input('extra', 'Notas adicionales', {wide: true, area: true})}</div>`);
    const adminOnly = admin ? section('lock', 'Solo administración', `<div class="pf-grid">${input('code', 'Código', {ph: 'Se asigna uno si lo dejas vacío'})}${input('row', 'N.º fila')}${input('ordinal', 'Orden')}<label>Servicio reportado<select name="service"><option value="">Sin confirmar</option><option value="YES">Sí (reportado)</option><option value="MAYBE">Tal vez</option><option value="NO">No</option></select></label>${input('serviceNotes', 'Notas de servicio', {wide: true, area: true})}${input('notes', 'Notas de grupo', {wide: true, area: true})}${edit ? input('note', 'Nota privada del campamento (opcional)', {wide: true, area: true}) : ''}</div>`, 'pf-admin') : '';
    return where + who + what + adminOnly;
  }
  const memberFields = [['name', 'Nombre', ''], ['phone', 'Teléfono', 'phone'], ['age', 'Edad', 'age'], ['cedula', 'Cédula', '']];
  // Una fila por persona: número a la izquierda y los campos con su nombre como texto de ayuda (sin título «Persona N»).
  const memberRow = (m, i) => `<div class="pf-member"><span class="pf-mnum" aria-hidden="true">${i + 1}</span>${memberFields.map(([k, l, f]) => `<input class="pf-m-${k}" data-k="${k}" value="${esc(m[k])}" placeholder="${l}" aria-label="${l} de la persona ${i + 1}" autocomplete="off"${filterAttrs(f)}>`).join('')}<button type="button" class="pf-remove" aria-label="Quitar persona ${i + 1}" title="Quitar">${icons.x}</button></div>`;
  const filters = {count: v => v.replace(/\D/g, '').slice(0, 3), age: v => v.replace(/\D/g, '').slice(0, 3), phone: v => v.replace(/[^\d+\-\s()]/g, '')};
  // Puntos anteriores guardaban hasta tres contactos sueltos; se muestran como personas para editarlos.
  function membersOf(p = {}) {
    if (Array.isArray(p.members) && p.members.length) return p.members;
    return [[p.contact, p.phone], [p.secondContact, p.secondPhone], [p.thirdContact, p.thirdPhone]].filter(([n, t]) => n || t).map(([name, phone]) => ({name: name || '', phone: phone || ''}));
  }
  // «Necesidades» se guarda como texto: opciones marcadas y lo escrito en «Otros», separados por comas.
  // Lista de empaque: artículos iniciales con su ícono (los mismos nombres que defaultPackingItems en packing.go).
  const PACKING = [['Bolsas alimentación', 'foodbag'], ['Kit aseo personal', 'hygiene'], ['Kit limpieza', 'cleaning'], ['Kit bebé', 'bottle'], ['Linternas', 'flashlight'], ['Carpas / toldas', 'tent'], ['Repelente', 'repellent'], ['Sillas', 'chair'], ['Hamacas', 'hammock'], ['Mosquiteros', 'net'], ['Extensiones', 'plug'], ['Camping', 'camping'], ['Toallas', 'towel'], ['Capotes / ponchos', 'poncho'], ['Paraguas', 'umbrella'], ['Medicamentos', 'medicine'], ['Utensilios', 'utensils'], ['Desechables', 'disposable'], ['Colchonetas', 'mattress']];
  const packIcon = name => { const hit = PACKING.find(([l]) => fold(l) === fold(String(name || ''))); return icons[hit ? hit[1] : 'box']; };
  // packingTotals suma las cantidades por artículo (sin distinguir mayúsculas ni acentos), en el orden en que se agregaron.
  const packingTotals = lines => { const out = []; (lines || []).forEach(l => { const hit = out.find(x => fold(x.item) === fold(l.item)); if (hit) hit.quantity += l.quantity; else out.push({item: l.item, quantity: l.quantity}); }); return out; };
  const packingHtml = lines => { const totals = packingTotals(lines); return totals.length ? `<div class="pf-pack"><b class="pf-pack-title">${icons.box}Lista de empaque</b><div class="pf-pack-chips">${totals.map(t => `<span class="pf-pack-chip">${packIcon(t.item)}${esc(t.item)} <b>× ${t.quantity}</b></span>`).join('')}</div></div>` : ''; };
  // Diálogo «Agregar a lista de empaque» (corredores, agentes y administración):
  // elegir un artículo o escribir uno nuevo → cantidad → repetir → revisar → guardar.
  // openPacking({subtitle, items, save}) — save(lines) guarda y devuelve una promesa; si falla, el diálogo muestra el error.
  const pk = {items: [], draft: [], pick: '', save: null};
  const pq = sel => document.querySelector('#pfPack ' + sel);
  const packDraftHtml = edit => pk.draft.length ? `<div class="pfp-list">${pk.draft.map((l, i) => `<div class="pfp-line">${packIcon(l.item)}<span>${esc(l.item)}</span>${edit ? `<input type="number" min="1" max="100000" step="1" inputmode="numeric" value="${l.quantity}" data-qty="${i}" aria-label="Cantidad de ${esc(l.item)}">` : `<b>× ${l.quantity}</b>`}<button type="button" class="pfp-del" data-del="${i}" aria-label="Quitar ${esc(l.item)}">✕</button></div>`).join('')}</div>` : '<p class="pfp-hint">Todavía no agregaste artículos.</p>';
  function packDraw(edit) { const n = pk.draft.length; pq('[data-act=review]').disabled = !n; pq('[data-act=review]').textContent = n ? `Revisar (${n})` : 'Revisar'; pq('.pfp-draft').innerHTML = packDraftHtml(edit); }
  function packOptions() {
    const raw = pq('#pfPackSearch').value.trim().replace(/\s+/g, ' '), q = fold(raw), list = pk.items.filter(i => !q || fold(i.name).includes(q)), exact = pk.items.some(i => fold(i.name) === q);
    pq('.pfp-options').innerHTML = list.map(i => `<button type="button" role="option" data-item="${esc(i.name)}">${packIcon(i.name)}${esc(i.name)}</button>`).join('') + (raw && !exact ? `<button type="button" role="option" class="pfp-new" data-item="${esc(raw)}">${icons.box}Agregar «${esc(raw)}» como artículo nuevo</button>` : '') + (!list.length && !raw ? '<p class="pfp-hint">No hay artículos guardados todavía.</p>' : '');
    pq('.pfp-options [data-item]')?.classList.add('hl');
  }
  function packStep(step) {
    pq('.pfp-pick').hidden = step !== 'pick'; pq('.pfp-qty').hidden = step !== 'qty'; pq('.pfp-review').hidden = step !== 'review';
    pq('[data-act=review]').hidden = step === 'review'; pq('[data-act=save]').hidden = pq('[data-act=more]').hidden = step !== 'review';
    pq('.pfp-msg').textContent = ''; packDraw(step === 'review');
    if (step === 'pick') { pq('#pfPackSearch').value = ''; packOptions(); pq('#pfPackSearch').focus(); }
  }
  function packChoose(name) {
    const known = pk.items.find(i => fold(i.name) === fold(name)); pk.pick = known ? known.name : name;
    pq('.pfp-chosen').innerHTML = `${packIcon(pk.pick)}<b>${esc(pk.pick)}</b>${known ? '' : ' <small>(nuevo)</small>'}`;
    pq('#pfPackQty').value = 1; packStep('qty'); pq('#pfPackQty').focus(); pq('#pfPackQty').select();
  }
  function packAdd() {
    const qty = Number(pq('#pfPackQty').value);
    if (!Number.isInteger(qty) || qty < 1 || qty > 100000) { pq('.pfp-msg').textContent = 'Escribe una cantidad entera mayor que cero.'; pq('#pfPackQty').focus(); return; }
    const hit = pk.draft.find(l => fold(l.item) === fold(pk.pick)); if (hit) hit.quantity += qty; else pk.draft.push({item: pk.pick, quantity: qty});
    packStep('pick');
  }
  const packDiscard = () => !pk.draft.length || confirm('¿Descartar los artículos sin guardar?');
  async function packSave() {
    const bad = pk.draft.find(l => !Number.isInteger(l.quantity) || l.quantity < 1 || l.quantity > 100000);
    if (bad) { pq('.pfp-msg').textContent = `Revisa la cantidad de «${bad.item}»: debe ser un número entero mayor que cero.`; return; }
    const b = pq('[data-act=save]'); b.disabled = true;
    try { await pk.save(pk.draft.map(l => ({item: l.item, quantity: l.quantity}))); pk.draft = []; document.getElementById('pfPack').close(); }
    catch (err) { pq('.pfp-msg').textContent = 'No se pudo guardar: ' + err.message; }
    finally { b.disabled = false; }
  }
  function packDialog() {
    let d = document.getElementById('pfPack');
    if (d) return d;
    document.body.insertAdjacentHTML('beforeend', `<dialog id="pfPack" class="pfp" aria-labelledby="pfPackTitle"><div class="pfp-box"><h2 id="pfPackTitle">Agregar a lista de empaque</h2><p class="pfp-for"></p><div class="pfp-pick"><label for="pfPackSearch">Artículo</label><input id="pfPackSearch" autocomplete="off" spellcheck="false" placeholder="Escribe para buscar o agregar uno nuevo…" role="combobox" aria-controls="pfPackOptions" aria-expanded="true"><div id="pfPackOptions" class="pfp-options" role="listbox" aria-label="Artículos"></div></div><div class="pfp-qty" hidden><div class="pfp-chosen"></div><label for="pfPackQty">Cantidad</label><div class="pfp-qtyrow"><button type="button" class="pfp-step" data-step="-1" aria-label="Menos">−</button><input id="pfPackQty" type="number" min="1" max="100000" step="1" inputmode="numeric" value="1"><button type="button" class="pfp-step" data-step="1" aria-label="Más">+</button></div><button type="button" class="pfp-btn" data-act="add">Agregar a la lista</button><button type="button" class="pfp-btn pfp-plain" data-act="other">Elegir otro artículo</button></div><div class="pfp-review" hidden><h3>Revisa antes de guardar</h3><p class="pfp-hint">Ajusta cantidades o quita artículos. Al guardar se suman a la lista de empaque de esta necesidad.</p></div><div class="pfp-draft"></div><p class="pfp-msg" role="status"></p><div class="pfp-actions"><button type="button" class="pfp-btn pfp-plain" data-act="cancel">Cancelar</button><button type="button" class="pfp-btn pfp-plain" data-act="more" hidden>Seguir agregando</button><button type="button" class="pfp-btn" data-act="review" disabled>Revisar</button><button type="button" class="pfp-btn" data-act="save" hidden>Guardar lista</button></div></div></dialog>`);
    d = document.getElementById('pfPack');
    pq('#pfPackSearch').addEventListener('input', packOptions);
    pq('#pfPackSearch').addEventListener('keydown', e => {
      const opts = [...d.querySelectorAll('.pfp-options [data-item]')], at = opts.findIndex(o => o.classList.contains('hl'));
      if (e.key === 'Enter') { e.preventDefault(); if (opts[at]) packChoose(opts[at].dataset.item); }
      else if (e.key === 'ArrowDown' || e.key === 'ArrowUp') { e.preventDefault(); if (!opts.length) return; opts[at]?.classList.remove('hl'); const next = opts[(at + (e.key === 'ArrowDown' ? 1 : -1) + opts.length) % opts.length]; next.classList.add('hl'); next.scrollIntoView({block: 'nearest'}); }
    });
    pq('.pfp-options').addEventListener('click', e => { const b = e.target.closest('[data-item]'); if (b) packChoose(b.dataset.item); });
    pq('.pfp-qty').addEventListener('click', e => { const b = e.target.closest('[data-step]'); if (b) { const q = pq('#pfPackQty'); q.value = Math.max(1, (parseInt(q.value, 10) || 0) + Number(b.dataset.step)); } });
    pq('#pfPackQty').addEventListener('keydown', e => { if (e.key === 'Enter') { e.preventDefault(); packAdd(); } });
    pq('.pfp-draft').addEventListener('click', e => { const b = e.target.closest('[data-del]'); if (!b) return; pk.draft.splice(Number(b.dataset.del), 1); const reviewing = !pq('.pfp-review').hidden; if (reviewing && !pk.draft.length) packStep('pick'); else packDraw(reviewing); });
    pq('.pfp-draft').addEventListener('input', e => { const i = e.target.dataset.qty; if (i !== undefined) pk.draft[Number(i)].quantity = Number(e.target.value); });
    d.addEventListener('click', e => {
      const act = e.target.closest('[data-act]')?.dataset.act;
      if (act === 'add') packAdd(); else if (act === 'other' || act === 'more') packStep('pick');
      else if (act === 'review' && pk.draft.length) packStep('review'); else if (act === 'save') packSave();
      else if (act === 'cancel' && packDiscard()) d.close();
    });
    d.addEventListener('cancel', e => { if (!packDiscard()) e.preventDefault(); });
    return d;
  }
  function openPacking({subtitle = '', items = [], save}) {
    const d = packDialog(); pk.items = items; pk.draft = []; pk.save = save;
    pq('.pfp-for').textContent = subtitle; d.showModal(); packStep('pick');
  }
  // Fotos y videos de una carrera (partida / entrega). Las fotos grandes se reducen a 2000 px en JPEG antes de subir
  // para gastar menos datos; los videos y lo que el navegador no pueda abrir (p. ej. HEIC) se suben tal cual.
  async function prepareMedia(file) {
    if (!/^image\/(jpeg|png|webp)$/.test(file.type) || file.size < 1.5e6 || typeof createImageBitmap !== 'function') return file;
    try {
      const img = await createImageBitmap(file), scale = Math.min(1, 2000 / Math.max(img.width, img.height));
      const canvas = document.createElement('canvas'); canvas.width = Math.round(img.width * scale); canvas.height = Math.round(img.height * scale);
      canvas.getContext('2d').drawImage(img, 0, 0, canvas.width, canvas.height);
      const blob = await new Promise(r => canvas.toBlob(r, 'image/jpeg', 0.85));
      return blob && blob.size < file.size ? new File([blob], file.name.replace(/\.\w+$/, '') + '.jpg', {type: 'image/jpeg'}) : file;
    } catch { return file; }
  }
  // uploadMedia sube el archivo tal cual (sin multipart) y avisa el avance (0–1); devuelve el registro guardado.
  function uploadMedia(url, file, {csrf = '', onProgress = () => {}} = {}) {
    return new Promise((resolve, reject) => {
      const x = new XMLHttpRequest();
      x.open('POST', url); x.withCredentials = true;
      x.setRequestHeader('Content-Type', file.type || 'application/octet-stream');
      x.setRequestHeader('X-File-Name', encodeURIComponent(file.name || ''));
      if (csrf) x.setRequestHeader('X-CSRF-Token', csrf);
      x.upload.onprogress = e => { if (e.lengthComputable) onProgress(e.loaded / e.total); };
      x.onload = () => { let body = {}; try { body = JSON.parse(x.responseText); } catch {} if (x.status >= 200 && x.status < 300) resolve(body); else reject(Object.assign(new Error(body.error || `Error HTTP ${x.status}`), {status: x.status})); };
      x.onerror = () => reject(new Error('Sin conexión. Intenta de nuevo cuando tengas señal.'));
      x.send(file);
    });
  }
  // mediaThumbs dibuja las miniaturas de una etapa; src(item) da la URL privada (panel de corredor o de administración).
  const mediaThumbs = (items, src, extra = () => '') => items.map(m => `<figure class="pf-thumb"><a href="${esc(src(m))}" target="_blank" rel="noopener noreferrer" title="${esc(m.name || '')}">${String(m.type).startsWith('image/') ? `<img src="${esc(src(m))}" alt="" loading="lazy" onerror="this.replaceWith(Object.assign(document.createElement('span'),{className:'pf-thumb-icon',textContent:'📷'}))">` : '<span class="pf-thumb-icon">▶</span>'}</a>${extra(m)}</figure>`).join('');
  const fold = s => s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase().trim();
  function parseNeeds(text) {
    const checked = new Set(), other = [];
    for (const part of String(text || '').split(',').map(s => s.trim()).filter(Boolean)) {
      const hit = NEEDS.find(([l]) => fold(l) === fold(part));
      if (hit) checked.add(hit[0]); else other.push(part);
    }
    return {checked: [...checked], other: other.join(', ')};
  }
  const needsText = (checked, other) => [...checked, String(other || '').trim()].filter(Boolean).join(', ');
  // Marca las casillas según el texto guardado. Una opción sin casilla aquí (p. ej. «Reparación de casa» fuera de
  // administración) pasa a «Otros» para que guardar no la borre.
  function fillNeeds(root, text) {
    const needs = parseNeeds(text), boxes = [...root.querySelectorAll('[data-need]')], shown = boxes.map(b => b.dataset.need);
    const otros = root.querySelector('.pf-otros'), otrosCheck = root.querySelector('.pf-otros-check'), otrosText = root.querySelector('.pf-otros-text');
    for (const box of boxes) box.checked = needs.checked.includes(box.dataset.need);
    otrosText.value = needsText(needs.checked.filter(l => !shown.includes(l)), needs.other);
    otrosCheck.checked = !!otrosText.value;
    otros.hidden = !otrosText.value;
    otrosCheck.onchange = () => { otros.hidden = !otrosCheck.checked; if (otrosCheck.checked) otrosText.focus(); };
  }
  function readNeeds(root) {
    const checked = [...root.querySelectorAll('[data-need]:checked')].map(b => b.dataset.need);
    return needsText(checked, root.querySelector('.pf-otros-check').checked ? root.querySelector('.pf-otros-text').value : '');
  }
  function readMembers(root) {
    return [...root.querySelectorAll('.pf-member')].map(row => Object.fromEntries(memberFields.map(([k]) => [k, row.querySelector(`[data-k="${k}"]`).value.trim()])));
  }
  // Siempre queda una fila vacía al final: al escribir en ella aparece la siguiente, lista para llenar.
  const memberFilled = m => memberFields.some(([k]) => String(m[k] || '').trim());
  function drawMembers(root, list) {
    list = list.filter(memberFilled);
    list.push({});
    root.querySelector('.pf-members').innerHTML = list.map(memberRow).join('');
  }
  // Distancia en metros entre dos [lat, lng].
  function meters(a, b) {
    const r = Math.PI / 180, dLat = (b[0] - a[0]) * r, dLng = (b[1] - a[1]) * r;
    const h = Math.sin(dLat / 2) ** 2 + Math.cos(a[0] * r) * Math.cos(b[0] * r) * Math.sin(dLng / 2) ** 2;
    return 2 * 6371000 * Math.asin(Math.sqrt(h));
  }
  // Puntos con coordenadas ordenados del más cercano al más lejano, con su distancia.
  function nearest(here, list) {
    return (list || []).filter(p => Array.isArray(p.coords) && p.coords.length === 2).map(p => ({p, d: meters(here, p.coords)})).sort((a, b) => a.d - b.d);
  }
  const distance = d => d < 1000 ? `${Math.round(d)} m` : `${(d / 1000).toFixed(1)} km`;
  const PIN_COLORS = {approved: '#277a66', pending: '#e57450'};
  const inPanama = ([lat, lng]) => lat >= 6 && lat <= 10 && lng >= -83 && lng <= -78;
  // Coordenadas de un enlace de mapa ya escrito en el formulario («…?q=7.45,-80.35»), o null.
  function linkCoords(text) {
    const m = String(text || '').match(/(-?\d{1,2}\.\d+)\s*,\s*(-?\d{1,3}\.\d+)/);
    return m ? [Number(m[1]), Number(m[2])] : null;
  }
  const mapLink = ([lat, lng]) => `https://maps.google.com/?q=${lat.toFixed(6)},${lng.toFixed(6)}`;
  const crosshair = svg('<circle cx="12" cy="12" r="6"/><path d="M12 1v6M12 17v6M1 12h6M17 12h6"/>').replace('width="18" height="18"', 'width="44" height="44"');
  // Ventana de pantalla completa con solo el mapa (solo si se pasó `nearby`): los demás puntos en verde (aprobados) o naranja
  // (pendientes). Mientras llega el GPS, o si no se obtiene, queda centrado en el último punto agregado; `me` pone tu
  // ubicación en azul y centra el mapa en ella, salvo que ya hayas movido el mapa. Si el formulario ya tiene una ubicación
  // (elegida antes, GPS o enlace pegado), abre ahí y el GPS no la mueve ni la reemplaza (`keep`).
  // La cruz fija en el centro marca el punto elegido: se mueve arrastrando el mapa o tocando un lugar, y «Usar esta ubicación»
  // lo pasa al enlace GPS del formulario.
  function openGpsMap(root) {
    if (!root._nearby || !window.L) return null;
    let d = document.querySelector('dialog.pf-mapdialog');
    if (!d) {
      document.body.insertAdjacentHTML('beforeend', `<dialog class="pf-mapdialog" aria-label="Mapa de puntos y tu ubicación"><div class="pf-mapbox"></div><div class="pf-crosshair" aria-hidden="true">${crosshair}</div><button type="button" class="pf-mapclose" aria-label="Cerrar mapa" title="Cerrar">${icons.x}</button><button type="button" class="pf-mapgps" title="Mover la cruz a tu ubicación GPS">${icons.locate}Mi GPS</button><div class="pf-mappanel"><p class="pf-mapstatus" role="status"></p><p class="pf-mappick"></p><button type="button" class="pf-mapuse">Usar esta ubicación</button></div></dialog>`);
      d = document.querySelector('dialog.pf-mapdialog');
      d.querySelector('.pf-mapclose').onclick = () => d.close();
    }
    const list = (root._nearby() || []).filter(p => Array.isArray(p.coords) && p.coords.length === 2);
    const last = lastAdded(list), status = d.querySelector('.pf-mapstatus'), pick = d.querySelector('.pf-mappick'), use = d.querySelector('.pf-mapuse');
    if (d._map) d._map.remove();
    d.showModal();
    const map = d._map = L.map(d.querySelector('.pf-mapbox'));
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {maxZoom: 19, attribution: '© OpenStreetMap contributors'}).addTo(map);
    for (const p of list) L.circleMarker(p.coords, {radius: 8, weight: 2, color: '#fff', fillColor: PIN_COLORS[p.status] || PIN_COLORS.approved, fillOpacity: 1}).addTo(map).bindTooltip(`${esc(p.code)} · ${esc(p.area)}${p.status === 'pending' ? ' · pendiente' : ''}${p === last ? ' · último agregado' : ''}`);
    const prev = linkCoords(root.querySelector('[name="gps"]').value), keep = !!prev && inPanama(prev);
    let moved = keep, gpsHere = null, gpsLayers = [];
    const center = () => { const c = map.getCenter(); return [c.lat, c.lng]; };
    const showPick = () => {
      const here = center(), ok = inPanama(here), near = nearest(here, list)[0];
      use.disabled = !ok;
      pick.textContent = ok ? `Cruz: ${here[0].toFixed(5)}, ${here[1].toFixed(5)}${near ? ` · a ${distance(near.d)} de ${near.p.code}` : ''}` : 'La cruz está fuera de Panamá.';
    };
    map.on('move', showPick);
    map.on('dragstart', () => { moved = true; });
    map.on('click', e => { moved = true; map.panTo(e.latlng); });
    if (keep) map.setView(prev, 17);
    else map.setView(last ? last.coords : [7.46, -80.33], last ? 15 : 11);
    setTimeout(() => map.invalidateSize(), 50);
    status.textContent = keep ? 'La cruz está en la ubicación elegida antes. Mueve el mapa o toca un lugar para cambiarla.' : 'Buscando tu ubicación… Mueve el mapa o toca un lugar para poner la cruz en el punto.';
    use.onclick = () => {
      const here = center();
      if (!inPanama(here)) return;
      root.querySelector('[name="gps"]').value = mapLink(here);
      root.querySelector('.pf-gpsmsg').textContent = '';
      d.close();
      drawGpsThumb(root);
    };
    const view = {
      keep,
      me(here, accuracy) {
        gpsHere = here;
        gpsLayers.forEach(l => map.removeLayer(l));
        gpsLayers = [L.marker(here, {icon: L.divIcon({className: '', html: '<div class="pf-me"></div>', iconSize: [22, 22]}), zIndexOffset: 1000, interactive: false}).addTo(map)];
        if (accuracy) gpsLayers.push(L.circle(here, {radius: accuracy, weight: 1, color: '#1f6fd1', fillColor: '#1f6fd1', fillOpacity: .12, interactive: false}).addTo(map));
        if (!moved) map.setView(here, 16);
        status.textContent = `Azul: tu GPS (±${Math.round(accuracy || 0)} m) · verde: aprobados · naranja: pendientes. ${keep ? 'La cruz sigue en la ubicación elegida antes.' : 'Mueve el mapa para ajustar la cruz.'}`;
      },
      status(text) { status.textContent = text + (last && !moved && !keep ? ` Mapa centrado en el último punto agregado: ${last.code} (${last.area}).` : '') + ' Puedes elegir el punto con la cruz.'; }
    };
    // «Mi GPS» lleva la cruz a tu ubicación; si aún no hay GPS lo pide. No cambia el formulario hasta «Usar esta ubicación».
    const gpsBtn = d.querySelector('.pf-mapgps');
    gpsBtn.disabled = false;
    gpsBtn.onclick = () => {
      moved = true;
      if (gpsHere) { map.setView(gpsHere, Math.max(map.getZoom(), 16)); return; }
      if (!navigator.geolocation) { status.textContent = 'Este dispositivo no permite obtener la ubicación.'; return; }
      gpsBtn.disabled = true; status.textContent = 'Buscando tu ubicación…';
      navigator.geolocation.getCurrentPosition(pos => {
        const here = [pos.coords.latitude, pos.coords.longitude];
        gpsBtn.disabled = false;
        if (!inPanama(here)) { status.textContent = 'Tu GPS queda fuera de Panamá.'; return; }
        view.me(here, pos.coords.accuracy);
        map.setView(here, Math.max(map.getZoom(), 16));
      }, err => {
        gpsBtn.disabled = false;
        status.textContent = err.code === 1 ? 'Permiso de ubicación denegado. Actívalo en el navegador.' : 'No se pudo obtener tu GPS. Intenta de nuevo.';
      }, {enableHighAccuracy: true, timeout: 20000, maximumAge: 0});
    };
    return view;
  }
  // Vista previa fija (sin arrastrar ni zoom) de la ubicación del formulario, con su pin; tocarla abre el mapa para moverla.
  // Se redibuja al elegir en el mapa, al llegar el GPS o al pegar un enlace; se ajusta sola cuando el formulario se hace visible.
  function drawGpsThumb(root) {
    const box = root.querySelector('.pf-gpsthumb');
    if (!box) return;
    const c = linkCoords(root.querySelector('[name="gps"]').value);
    if (box._map) { box._map.remove(); box._map = null; }
    box.hidden = !(c && inPanama(c) && window.L);
    if (box.hidden) return;
    const map = box._map = L.map(box, {zoomControl: false, dragging: false, scrollWheelZoom: false, doubleClickZoom: false, boxZoom: false, keyboard: false, touchZoom: false, tap: false}).setView(c, 16);
    map.attributionControl.setPrefix(false);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {maxZoom: 19, attribution: '© OpenStreetMap'}).addTo(map);
    L.marker(c, {icon: L.divIcon({className: '', html: '<div class="pf-thumbpin"></div>', iconSize: [24, 24], iconAnchor: [12, 28]}), interactive: false}).addTo(map);
    if (window.ResizeObserver) { if (!box._watch) { box._watch = new ResizeObserver(() => { if (box._map && box.offsetWidth) { box._map.invalidateSize(); box._map.setView(box._center, 16); } }); box._watch.observe(box); } }
    box._center = c;
    setTimeout(() => { map.invalidateSize(); map.setView(c, 16); }, 50);
  }
  // Punto agregado más recientemente (por fecha de creación).
  const lastAdded = list => list.reduce((a, p) => !a || String(p.created || '') > String(a.created || '') ? p : a, null);
  function locate(root) {
    const btn = root.querySelector('.pf-gps'), msg = root.querySelector('.pf-gpsmsg'), gps = root.querySelector('[name="gps"]');
    const view = openGpsMap(root);
    if (!navigator.geolocation) { msg.textContent = 'Este dispositivo no permite obtener la ubicación. Pega un enlace de mapa.'; view?.status(msg.textContent); return; }
    btn.disabled = true; msg.textContent = 'Buscando tu ubicación…';
    navigator.geolocation.getCurrentPosition(pos => {
      const lat = pos.coords.latitude, lng = pos.coords.longitude, outside = !inPanama([lat, lng]);
      // Con una ubicación ya elegida en el mapa, el GPS solo se muestra; el enlace cambia con «Usar esta ubicación».
      if (view?.keep) { outside ? view.status('Tu GPS queda fuera de Panamá.') : view.me([lat, lng], pos.coords.accuracy); btn.disabled = false; return; }
      gps.value = mapLink([lat, lng]);
      drawGpsThumb(root);
      msg.textContent = outside ? 'Tu ubicación queda fuera de Panamá; revisa el enlace antes de enviar.' : `✓ Ubicación GPS agregada (precisión aprox. ±${Math.round(pos.coords.accuracy || 0)} m).`;
      if (view) outside ? view.status(msg.textContent) : view.me([lat, lng], pos.coords.accuracy);
      btn.disabled = false;
    }, err => {
      if (view?.keep) { view.status(err.code === 1 ? 'Permiso de ubicación denegado.' : 'No se pudo obtener tu GPS.'); btn.disabled = false; return; }
      msg.textContent = err.code === 1 ? 'Permiso de ubicación denegado. Actívalo en el navegador o pega un enlace de mapa.' : 'No se pudo obtener tu ubicación' + (window.isSecureContext ? '. Intenta de nuevo o pega un enlace de mapa.' : ' (requiere HTTPS).');
      view?.status(msg.textContent);
      btn.disabled = false;
    }, {enableHighAccuracy: true, timeout: 20000, maximumAge: 0});
  }
  // Dibuja el formulario dentro de `root` (un contenedor dentro del <form>) y carga los datos de `point`.
  // `nearby`: función que devuelve los puntos del mapa de pantalla completa que se abre con «Enviar mi ubicación GPS».
  // Con usuario conocido (entró con su código), «Yo» junto a Coordinador pone su nombre y teléfono; la ✕ borra ambos.
  // me = {name, telephone} o null para quitar los botones.
  function setMe(root, me) {
    const label = root.querySelector('[name="coordinator"]')?.closest('label');
    if (!label) return;
    label.querySelector('.pf-me-btns')?.remove();
    if (!me || !me.name) return;
    label.querySelector('span').insertAdjacentHTML('beforeend', `<span class="pf-me-btns"><button type="button" class="pf-yo" title="Usar mi nombre y teléfono">Yo</button><button type="button" class="pf-clearcoord" title="Borrar coordinador y teléfono" aria-label="Borrar coordinador y teléfono">${icons.x}</button></span>`);
    const set = (name, phone) => {
      for (const [k, v] of [['coordinator', name], ['coordinatorPhone', phone]]) { const el = root.querySelector(`[name="${k}"]`); el.value = v; el.dispatchEvent(new Event('input', {bubbles: true})); }
    };
    label.querySelector('.pf-yo').onclick = e => { e.preventDefault(); set(me.name, me.telephone || ''); };
    label.querySelector('.pf-clearcoord').onclick = e => { e.preventDefault(); set('', ''); root.querySelector('[name="coordinator"]').focus(); };
  }
  function render(root, {admin = false, edit = false, point = {}, nearby = null, me = null} = {}) {
    root._nearby = nearby;
    root.classList.add('pf');
    root.innerHTML = html({admin, edit});
    for (const el of root.querySelectorAll('[name]')) el.value = point[el.name] ?? '';
    fillNeeds(root, point.needs);
    drawMembers(root, membersOf(point).map(m => ({...m})));
    // Con mapa, el botón se llama «Ubicar en mapa»; sin mapa (formulario público) conserva «Enviar mi ubicación GPS».
    if (nearby) root.querySelector('.pf-gps').innerHTML = `${icons.pin}Ubicar en mapa`;
    root.querySelector('.pf-gps').onclick = () => locate(root);
    root.querySelector('.pf-gpsthumb').onclick = () => locate(root);
    root.querySelector('[name="gps"]').addEventListener('input', () => drawGpsThumb(root));
    setMe(root, me);
    drawGpsThumb(root);
    root.oninput = e => { const f = filters[e.target.dataset?.filter]; if (f) { const v = f(e.target.value); if (v !== e.target.value) e.target.value = v; } };
    root.onclick = e => {
      const b = e.target.closest('.pf-numbtn'); if (!b) return;
      const el = b.parentNode.querySelector('input');
      el.value = Math.min(999, Math.max(0, (parseInt(el.value, 10) || 0) + Number(b.dataset.step)));
      el.dispatchEvent(new Event('input', {bubbles: true}));
    };
    root.querySelector('.pf-members').addEventListener('input', () => {
      const box = root.querySelector('.pf-members'), rows = box.querySelectorAll('.pf-member'), last = rows[rows.length - 1];
      if (last && [...last.querySelectorAll('input')].some(el => el.value.trim())) box.insertAdjacentHTML('beforeend', memberRow({}, rows.length));
    });
    root.querySelector('.pf-members').onclick = e => {
      const b = e.target.closest('.pf-remove'); if (!b) return;
      const list = readMembers(root); list.splice([...root.querySelectorAll('.pf-remove')].indexOf(b), 1); drawMembers(root, list);
    };
  }
  // Datos para la API: campos de texto, «needs» armado desde las casillas y `members` sin filas vacías.
  function values(root) {
    const data = {};
    for (const el of root.querySelectorAll('[name]')) data[el.name] = el.value.trim();
    data.needs = readNeeds(root);
    data.members = readMembers(root).filter(m => Object.values(m).some(Boolean));
    return data;
  }
  const validPhone = s => (String(s ?? '').match(/\d/g) || []).length >= 7;
  const hasPhone = data => validPhone(data.coordinatorPhone) || (data.members || []).some(m => validPhone(m.phone));
  window.PointForm = {render, values, html, nearest, meters, lastAdded, linkCoords, needsHtml, fillNeeds, readNeeds, membersOf, parseNeeds, needsText, filters, validPhone, hasPhone, NEEDS, icons, PACKING, packIcon, packingTotals, packingHtml, openPacking, prepareMedia, uploadMedia, mediaThumbs, setMe};
})();
