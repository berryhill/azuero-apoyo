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
    pot: svg('<path d="M4 10h16v6a4 4 0 0 1-4 4H8a4 4 0 0 1-4-4v-6zM2 10h2M20 10h2M9 6c0-1 1-1 1-2M14 6c0-1 1-1 1-2"/>'),
    bottle: svg('<path d="M10 2h4M10 5h4l1 3v12a2 2 0 0 1-2 2h-2a2 2 0 0 1-2-2V8l1-3zM9 12h6"/>'),
    fuel: svg('<path d="M4 21V5a2 2 0 0 1 2-2h6a2 2 0 0 1 2 2v16M3 21h12M14 9h2a2 2 0 0 1 2 2v6a1.5 1.5 0 0 0 3 0V8l-3-3M7 7h4"/>'),
    other: svg('<circle cx="12" cy="12" r="9"/><path d="M12 8v8M8 12h8"/>')
  };
  const style = `.pf-section{border:1px solid #e1e9ea;border-radius:12px;padding:10px 14px 14px;margin:0 0 12px;background:#fbfdfd;min-width:0}
.pf-section legend{display:flex;align-items:center;gap:7px;font-weight:800;font-size:15px;color:#173c46;padding:0 6px}
.pf-section legend svg{color:#2d7f86}.pf-step{display:inline-grid;place-items:center;width:22px;height:22px;border-radius:50%;background:#1d6f7a;color:#fff;font-size:12px;font-weight:800}.pf-admin{background:#fffaf4;border-color:#ecd9c8}.pf-admin legend svg{color:#b5653f}
.pf-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(130px,1fr));gap:10px}.pf-grid .wide{grid-column:1/-1}
.pf label{display:block;margin:0;font-size:12px;font-weight:700;color:#36505a}.pf .pf-ico{display:flex;align-items:center;gap:5px}.pf .pf-ico svg{color:#2d7f86;flex:none}
.pf input,.pf textarea,.pf select{display:block;width:100%;max-width:none;min-height:0;margin-top:4px;padding:9px 10px;border:1px solid #d5e0e2;border-radius:8px;background:#fff;font:inherit;font-weight:400;color:#152b3a}
.pf textarea{min-height:64px;resize:vertical}.pf input:focus,.pf textarea:focus,.pf select:focus{outline:2px solid #2d7f8655;border-color:#2d7f86}
.pf-hint{font-size:12px;color:#5b6f76;margin:2px 0 10px;line-height:1.4}.pf-hint b{color:#9a4c2f}.pf-sub{font-size:12px;font-weight:800;color:#2d7f86;text-transform:uppercase;letter-spacing:.05em;margin:14px 0 8px}
.pf .pf-gps{display:flex;align-items:center;justify-content:center;gap:8px;width:100%;margin:2px 0 0;padding:13px;border:0;border-radius:10px;background:#1d6f7a;color:#fff;font-size:15px;font-weight:800}
.pf .pf-gps:hover{background:#175d66}.pf .pf-gps:disabled{opacity:.6;cursor:wait}.pf-gpsmsg{font-size:12px;color:#36505a;margin:6px 0 0}.pf-gpsmsg:empty{display:none}
.pf-member{display:grid;grid-template-columns:2fr 1.4fr .7fr 1.3fr;gap:8px;padding:10px 0 2px;border-top:1px dashed #dbe5e6}
.pf-mhead{grid-column:1/-1;display:flex;justify-content:space-between;align-items:center;font-size:12px;font-weight:800;color:#2d7f86;text-transform:uppercase;letter-spacing:.05em}
.pf .pf-remove{display:inline-flex;margin:0;padding:3px;border:0;border-radius:6px;background:none;color:#8a9ca1}.pf .pf-remove:hover{background:#f2e7e3;color:#9a4536}
.pf .pf-add{display:flex;align-items:center;justify-content:center;gap:6px;width:100%;margin:12px 0 0;padding:9px;border:1px dashed #2d7f86;border-radius:9px;background:none;color:#1d6f7a;font-weight:750}.pf .pf-add:hover{background:#eef7f6}
.pf-needs{display:grid;grid-template-columns:repeat(auto-fill,minmax(170px,1fr));gap:8px}
.pf label.pf-need{position:relative;margin:0;font-weight:650}.pf .pf-need input{position:absolute;opacity:0;width:1px;height:1px;margin:0}
.pf-need span{display:flex;align-items:center;gap:8px;min-height:44px;padding:8px 10px;border:1px solid #d5e0e2;border-radius:10px;background:#fff;color:#36505a;font-size:13px;cursor:pointer}
.pf-need span svg{color:#2d7f86;flex:none}.pf-need span:after{content:"";margin-left:auto;width:16px;height:16px;border:1.5px solid #b7c7ca;border-radius:4px;flex:none}
.pf-need input:checked+span{border-color:#2d7f86;background:#eaf6f3;color:#173c46}.pf-need input:checked+span:after{background:#2d7f86 url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='white' stroke-width='3'%3E%3Cpath d='M5 12l5 5 9-10'/%3E%3C/svg%3E") center/12px no-repeat;border-color:#2d7f86}
.pf-need input:focus-visible+span{outline:2px solid #2d7f86}.pf-otros[hidden]{display:none}.pf label.pf-otros{margin-top:8px}
@media(max-width:560px){.pf-member{grid-template-columns:1fr 1fr}.pf-needs{grid-template-columns:1fr 1fr}}`;
  if (!document.getElementById('pointform-style')) document.head.insertAdjacentHTML('beforeend', `<style id="pointform-style">${style}</style>`);

  // filter: «count» y «age» solo dígitos (0–999); «phone» dígitos, espacios, +, - y paréntesis.
  const input = (name, label, opt = {}) => `<label class="${opt.wide ? 'wide' : ''}"><span class="${opt.icon ? 'pf-ico' : ''}">${opt.icon ? icons[opt.icon] : ''}${esc(label)}</span>${opt.area
    ? `<textarea name="${name}" placeholder="${esc(opt.ph || '')}"></textarea>`
    : `<input name="${name}"${filterAttrs(opt.filter)} placeholder="${esc(opt.ph || '')}" autocomplete="off">`}</label>`;
  const filterAttrs = f => f === 'count' || f === 'age' ? ` data-filter="${f}" inputmode="numeric" maxlength="3"` : f === 'phone' ? ' data-filter="phone" type="tel" inputmode="tel" maxlength="40"' : '';
  // Los tres pasos llevan número (1, 2, 3); la sección de administración lleva candado.
  const section = (icon, title, body, cls = '') => `<fieldset class="pf-section ${cls}"><legend>${typeof icon === 'number' ? `<span class="pf-step">${icon}</span>` : icons[icon]}${esc(title)}</legend>${body}</fieldset>`;
  const NEEDS = [['Agua', 'drop'], ['Alimentos', 'food'], ['Refugio (carpa / lona / toldo)', 'tent'], ['Kit de cocina', 'pot'], ['Kit de bebé', 'bottle'], ['Gasolina para generador', 'fuel'], ['Diésel para generador', 'fuel']];
  const COUNTS = [['people', 'Personas', 'people'], ['adults', 'Adultos', 'adult'], ['children', 'Niños', 'child'], ['babies', 'Bebés', 'baby'], ['olderAdults', 'Adultos mayores', 'elder'], ['disabled', 'Con discapacidad', 'disabled']];

  function html({admin = false, edit = false} = {}) {
    const where = section(1, '¿Dónde estás?', `<button type="button" class="pf-gps">${icons.pin}Enviar mi ubicación GPS</button><p class="pf-gpsmsg" role="status"></p><div class="pf-grid" style="margin-top:12px">${input('gps', 'o pega un enlace de Google Maps', {wide: true, ph: 'https://maps.google.com/…'})}${input('area', 'Zona / área', {ph: 'Ej. Cañas'})}${input('place', 'Ubicación / referencia', {ph: 'Ej. Escuela de la comunidad'})}${input('directions', 'Indicaciones de llegada', {wide: true, area: true, ph: 'Cómo llegar, puntos de referencia…'})}</div>`);
    const who = section(2, '¿Con quién estás?', `<div class="pf-grid">${input('coordinator', 'Coordinador', {icon: 'coordinator'})}${input('coordinatorPhone', 'Teléfono del coordinador', {filter: 'phone', ph: 'Ej. 6000-0000'})}</div><p class="pf-sub">¿Cuántas personas?</p><div class="pf-grid">${COUNTS.map(([k, l, i]) => input(k, l, {icon: i, filter: 'count', ph: '0'})).join('')}</div><p class="pf-sub">Personas del grupo</p><p class="pf-hint"><b>Agrega al menos un teléfono</b> (del coordinador o de alguna persona). Nombre, cédula y edad son opcionales y solo los ve el equipo de curación.</p><div class="pf-members"></div><button type="button" class="pf-add">${icons.plus}Agrega a alguien más</button>`);
    const what = section(3, '¿Qué necesitas?', `<div class="pf-needs">${NEEDS.map(([l, i]) => `<label class="pf-need"><input type="checkbox" data-need="${esc(l)}"><span>${icons[i]}${esc(l)}</span></label>`).join('')}<label class="pf-need"><input type="checkbox" class="pf-otros-check"><span>${icons.other}Otros</span></label></div><label class="pf-otros" hidden>Otros: ¿qué más necesitan?<textarea class="pf-otros-text" placeholder="Medicinas, ropa, colchones…"></textarea></label><div class="pf-grid" style="margin-top:12px">${input('extra', 'Notas adicionales', {wide: true, area: true})}</div>`);
    const adminOnly = admin ? section('lock', 'Solo administración', `<div class="pf-grid">${input('code', 'Código', {ph: 'Se asigna uno si lo dejas vacío'})}${input('row', 'N.º fila')}${input('ordinal', 'Orden')}<label>Servicio reportado<select name="service"><option value="">Sin confirmar</option><option value="YES">Sí (reportado)</option><option value="MAYBE">Tal vez</option><option value="NO">No</option></select></label>${input('serviceNotes', 'Notas de servicio', {wide: true, area: true})}${input('notes', 'Notas de grupo', {wide: true, area: true})}${edit ? input('note', 'Nota privada del campamento (opcional)', {wide: true, area: true}) : ''}</div>`, 'pf-admin') : '';
    return where + who + what + adminOnly;
  }
  const memberFields = [['name', 'Nombre', ''], ['phone', 'Teléfono', 'phone'], ['age', 'Edad', 'age'], ['cedula', 'Cédula', '']];
  const memberRow = (m, i) => `<div class="pf-member"><div class="pf-mhead">Persona ${i + 1}<button type="button" class="pf-remove" aria-label="Quitar persona ${i + 1}" title="Quitar">${icons.x}</button></div>${memberFields.map(([k, l, f]) => `<label>${l}<input data-k="${k}" value="${esc(m[k])}" autocomplete="off"${filterAttrs(f)}></label>`).join('')}</div>`;
  const filters = {count: v => v.replace(/\D/g, '').slice(0, 3), age: v => v.replace(/\D/g, '').slice(0, 3), phone: v => v.replace(/[^\d+\-\s()]/g, '')};
  // Puntos anteriores guardaban hasta tres contactos sueltos; se muestran como personas para editarlos.
  function membersOf(p = {}) {
    if (Array.isArray(p.members) && p.members.length) return p.members;
    return [[p.contact, p.phone], [p.secondContact, p.secondPhone], [p.thirdContact, p.thirdPhone]].filter(([n, t]) => n || t).map(([name, phone]) => ({name: name || '', phone: phone || ''}));
  }
  // «Necesidades» se guarda como texto: opciones marcadas y lo escrito en «Otros», separados por comas.
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
  function readMembers(root) {
    return [...root.querySelectorAll('.pf-member')].map(row => Object.fromEntries(memberFields.map(([k]) => [k, row.querySelector(`[data-k="${k}"]`).value.trim()])));
  }
  function drawMembers(root, list) {
    while (list.length < 3) list.push({});
    root.querySelector('.pf-members').innerHTML = list.map(memberRow).join('');
  }
  function locate(root) {
    const btn = root.querySelector('.pf-gps'), msg = root.querySelector('.pf-gpsmsg'), gps = root.querySelector('[name="gps"]');
    if (!navigator.geolocation) { msg.textContent = 'Este dispositivo no permite obtener la ubicación. Pega un enlace de mapa.'; return; }
    btn.disabled = true; msg.textContent = 'Buscando tu ubicación…';
    navigator.geolocation.getCurrentPosition(pos => {
      const lat = pos.coords.latitude, lng = pos.coords.longitude;
      gps.value = `https://maps.google.com/?q=${lat.toFixed(6)},${lng.toFixed(6)}`;
      const outside = lat < 6 || lat > 10 || lng < -83 || lng > -78;
      msg.textContent = outside ? 'Tu ubicación queda fuera de Panamá; revisa el enlace antes de enviar.' : `✓ Ubicación GPS agregada (precisión aprox. ±${Math.round(pos.coords.accuracy || 0)} m).`;
      btn.disabled = false;
    }, err => {
      msg.textContent = err.code === 1 ? 'Permiso de ubicación denegado. Actívalo en el navegador o pega un enlace de mapa.' : 'No se pudo obtener tu ubicación' + (window.isSecureContext ? '. Intenta de nuevo o pega un enlace de mapa.' : ' (requiere HTTPS).');
      btn.disabled = false;
    }, {enableHighAccuracy: true, timeout: 20000, maximumAge: 0});
  }
  // Dibuja el formulario dentro de `root` (un contenedor dentro del <form>) y carga los datos de `point`.
  function render(root, {admin = false, edit = false, point = {}} = {}) {
    root.classList.add('pf');
    root.innerHTML = html({admin, edit});
    for (const el of root.querySelectorAll('[name]')) el.value = point[el.name] ?? '';
    const needs = parseNeeds(point.needs), otros = root.querySelector('.pf-otros'), otrosCheck = root.querySelector('.pf-otros-check'), otrosText = root.querySelector('.pf-otros-text');
    for (const box of root.querySelectorAll('[data-need]')) box.checked = needs.checked.includes(box.dataset.need);
    otrosText.value = needs.other;
    otrosCheck.checked = !!needs.other;
    otros.hidden = !needs.other;
    otrosCheck.onchange = () => { otros.hidden = !otrosCheck.checked; if (otrosCheck.checked) otrosText.focus(); };
    drawMembers(root, membersOf(point).map(m => ({...m})));
    root.querySelector('.pf-gps').onclick = () => locate(root);
    root.oninput = e => { const f = filters[e.target.dataset?.filter]; if (f) { const v = f(e.target.value); if (v !== e.target.value) e.target.value = v; } };
    root.querySelector('.pf-add').onclick = () => {
      const list = readMembers(root); list.push({}); drawMembers(root, list);
      root.querySelector('.pf-member:last-child input').focus();
    };
    root.querySelector('.pf-members').onclick = e => {
      const b = e.target.closest('.pf-remove'); if (!b) return;
      const list = readMembers(root); list.splice([...root.querySelectorAll('.pf-remove')].indexOf(b), 1); drawMembers(root, list);
    };
  }
  // Datos para la API: campos de texto, «needs» armado desde las casillas y `members` sin filas vacías.
  function values(root) {
    const data = {};
    for (const el of root.querySelectorAll('[name]')) data[el.name] = el.value.trim();
    const checked = [...root.querySelectorAll('[data-need]:checked')].map(b => b.dataset.need);
    data.needs = needsText(checked, root.querySelector('.pf-otros-check').checked ? root.querySelector('.pf-otros-text').value : '');
    data.members = readMembers(root).filter(m => Object.values(m).some(Boolean));
    return data;
  }
  const validPhone = s => (String(s ?? '').match(/\d/g) || []).length >= 7;
  const hasPhone = data => validPhone(data.coordinatorPhone) || (data.members || []).some(m => validPhone(m.phone));
  window.PointForm = {render, values, html, membersOf, parseNeeds, needsText, filters, validPhone, hasPhone};
})();
