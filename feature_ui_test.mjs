import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
const camp = readFileSync(new URL('./camp.html', import.meta.url), 'utf8');
const template = readFileSync(new URL('./index.html', import.meta.url), 'utf8');
const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
const pointformSrc = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
const PointForm = (() => { const window = {}; const document = {getElementById: () => null, head: {insertAdjacentHTML() {}}}; new Function('window', 'document', pointformSrc)(window, document); return window.PointForm; })();
test('admin ☰ menu can change the master password (typed twice), then signs in again', () => {
  assert.match(admin, /id="login"/);
  assert.match(admin, /<button type="button" role="menuitem" id="passwordToggle">Cambiar contraseña maestra<\/button><a role="menuitem" id="backupDownload"[^\n]*?<\/div>/);
  assert.match(admin, /<dialog id="changePassword"[^>]*><form id="changePasswordForm">[^]*?name="password" minlength="12" maxlength="72" autocomplete="new-password" required[^]*?name="confirm"/);
  assert.match(admin, /if\(pw!==f\.elements\.confirm\.value\)\{\$\('#passwordMessage'\)\.textContent='Las dos contraseñas no coinciden\.'/);
  assert.match(admin, /<label>Contraseña actual<input type="password" name="current" autocomplete="current-password" required><\/label><label>Nueva contraseña/);
  assert.match(admin, /await adminRequest\('\/api\/admin\/password','POST',\{current:f\.elements\.current\.value,password:pw\}\);[^]*?location\.reload\(\)/);
  assert.match(admin, /closeEditor\(\);if\(\$\('#dashboard'\)\.classList\.contains\('users-mode'\)\)showUsers\(false\);panel\.hidden=elsewhere\?false:!panel\.hidden;/, '+ Agregar punto works from Usuarios and from a point view');
});
test('admin camp detail is adjacent to the camp list rather than after runner management', () => {
  assert.ok(admin.indexOf('id="campEditor"') < admin.indexOf('id="items"'));
});
test('admin camp page lists saved needs before creation and keeps edits collapsed', () => {
  const start = admin.indexOf('id="campEditor"');
  const end = admin.indexOf('id="items"', start);
  const detail = admin.slice(start, end);
  assert.ok(detail.indexOf('id="needList"') < detail.indexOf('id="needForm"'));
  assert.ok(detail.indexOf('id="needForm"') < detail.indexOf('id="campEditDetails"'));
  assert.match(admin, /admin\.html\?camp=/);
  assert.match(admin, /camp-mode/);
});
test('public camp detail has a rendered service need list target', () => {
  assert.match(camp, /id="serviceNeeds"/);
});
const needStatusOptions = [['', 'Sin filtro de necesidades'], ['OPEN', 'Abierto'], ['IN_PROGRESS', 'En curso'], ['COMPLETED', 'Completado']];
test('admin filters camps by need status only through the stats counters, without a dropdown', () => {
  assert.doesNotMatch(admin, /id="needFilter"|Sin filtro de necesidades/);
  assert.match(admin, /needStatus=needFilter;/);
  assert.match(admin, /needFilter=needFilter===b\.dataset\.stat\?'':b\.dataset\.stat;view='approved'\}render\(\)/);
  assert.match(admin, /\/api\/admin\/needs/);
});
test('public stats are static counts; only need-status and coordinator stat buttons filter, with no dropdowns or chips', () => {
  assert.doesNotMatch(template, /mapCampStatus|listCampStatus|Sin filtro de necesidades|id="filters"|Por confirmar|data-filter="(all|mapped|unmapped)"/);
  for (const [id, label] of [['total', 'TOTAL PUNTOS'], ['mapped', 'EN MAPA'], ['unmapped', 'SIN GPS']]) assert.match(template, new RegExp(`<div class="stat"><b id="${id}">—</b><span>${label}</span></div>`), label);
  assert.match(template, /\$\('#total'\)\.textContent=everyRow\.length/);
  for (const [value, label] of needStatusOptions.slice(1)) assert.match(template, new RegExp(`<button type="button" class="stat fstat" data-need="${value}"[^>]*><b id="need${value}">—</b><span>${label.toUpperCase()}</span>`), label);
  assert.doesNotMatch(template, /CampUrgency|Todas las urgencias/);
  assert.match(template, /needStatuses/);
});
test('public map/list default to «Con necesidades»; pins are light green (camp), light red (need) or strong red (urgent need)', () => {
  assert.match(template, /<div class="stats" id="needStats"><button type="button" class="stat fstat active" data-need="ACTIVE" aria-pressed="true"><b id="needACTIVE">—<\/b><span>CON NECESIDADES<\/span><\/button><button type="button" class="stat fstat" data-need="OPEN"/);
  assert.match(template, /let selectedCoordinator='',selectedCampStatus='ACTIVE',teamOnly=false;/, 'on by default');
  assert.match(template, /selectedCampStatus==='ACTIVE'\?hasActiveNeed\(r\):hasNeed\(r,selectedCampStatus\)/);
  assert.match(template, /\$\('#needACTIVE'\)\.textContent=needBase\.filter\(hasActiveNeed\)\.length/);
  assert.match(template, /Toca «Con necesidades» para ver todos los puntos\./, 'an empty list says how to see every point');
  const src = template.match(/const hasNeed=[^;]*;/)[0] + template.match(/const pinColors=[^\n]*/)[0];
  const {pinColor, pinColors} = new Function(`${src};return {pinColor, pinColors}`)();
  assert.deepEqual(pinColors, {camp: '#7fcf9a', need: '#f28b82', urgent: '#c62828'});
  assert.equal(pinColor({needStatuses: ['COMPLETED']}), pinColors.camp, 'only completed needs = camp');
  assert.equal(pinColor({}), pinColors.camp);
  assert.equal(pinColor({needStatuses: ['IN_PROGRESS'], highestUrgency: 'normal'}), pinColors.need);
  assert.equal(pinColor({needStatuses: ['OPEN'], highestUrgency: 'urgente'}), pinColors.urgent);
  assert.match(template, /<div class="legend"><span><i style="background:#7fcf9a"><\/i>Punto<\/span><span><i style="background:#f28b82"><\/i>Necesidad<\/span><span><i style="background:#c62828"><\/i>Urgente<\/span><\/div>/, 'the map explains the colors');
});
test('public coordinator filter is a sidebar stat button with a searchable picker, not a map dropdown', () => {
  assert.doesNotMatch(template, /mapCoordinator|listCoordinator|<select[^>]*coordinador/i);
  assert.ok(template.indexOf('id="needStats"') < template.indexOf('id="coordOpen"'));
  assert.match(template, /<button type="button" class="stat fstat coordbtn" id="coordOpen" aria-haspopup="listbox"[^>]*><b id="coordCount">—<\/b><span id="coordLabel">COORDINADOR: TODOS<\/span><\/button><div class="picker" id="coordPicker" hidden><input id="coordSearch"/);
  const src = template.match(/const coordNames=[^\n]*\nfunction drawCoordList[^\n]*/)[0];
  const els = {'#coordSearch': {value: 'mar'}, '#coordList': {innerHTML: ''}, '#coordLabel': {}, '#coordOpen': {classList: {toggle() {}}}};
  const api = new Function('$', 'all', 'clean', 'selectedCoordinator', `${src};return {drawCoordList, coordLabel}`)(s => els[s], () => [{coordinator: 'Marta Ruiz'}, {coordinator: 'Luis'}, {coordinator: ' Marta Ruiz '}, {}], s => String(s), 'marta ruiz');
  api.drawCoordList();
  assert.equal((els['#coordList'].innerHTML.match(/data-coord=/g) || []).length, 1);
  assert.match(els['#coordList'].innerHTML, /data-coord="marta ruiz" class="hl sel" aria-selected="true">Marta Ruiz \(2\)</);
  assert.equal(api.coordLabel(), 'Marta Ruiz');
});
test('public need filter shows the matching saved needs with Spanish created / en ruta times', () => {
  assert.match(template, /\$\{selectedCampStatus&&selectedCampStatus!=='ACTIVE'&&r\.id\?serviceNeedsHtml\(r\):/);
  const src = template.match(/const activeNeeds=[^\n]*\nfunction loadNeeds[^\n]*\nfunction serviceNeedsHtml[^\n]*/)[0];
  const render = new Function('clean', 'selectedCampStatus', 'fetch', 'renderFeed', `${src};campNeeds.set('c1',{gen:0,list:[{need:'Agua',status:'IN_PROGRESS',urgency:'urgente',createdAt:'2026-10-13T13:00:00Z',enRouteAt:'2026-10-13T14:30:00Z'},{need:'Comida',status:'OPEN',createdAt:'2026-10-13T12:00:00Z'},{need:'Ropa',status:'COMPLETED',createdAt:'2026-10-13T11:00:00Z'}]});campNeeds.set('c2',{gen:0,list:[{need:'Ropa',status:'COMPLETED'}]});return (id,all)=>serviceNeedsHtml({id},...(all?[activeNeeds]:[]))`)(s => String(s), 'IN_PROGRESS', () => { throw Error('should use cache'); }, () => {});
  const html = render('c1');
  assert.match(html, /<b>En ruta <span class="badge urgent">Urgente<\/span><\/b><div class="chips"><span class="chip text">Agua<\/span><\/div><small>Creada: [^<]+<br>En ruta: [^<]+<\/small>/);
  assert.doesNotMatch(html, /Comida|Sin necesidades registradas/);
  // Sin filtro: el punto muestra sus necesidades abiertas y en ruta (no completadas), más reciente primero.
  assert.match(template, /serviceNeedsHtml\(r\):\(r\.id&&activeNeeds\.some\(s=>hasNeed\(r,s\)\)\?serviceNeedsHtml\(r,activeNeeds\):r\.needs\?needChips\(r\.needs\):'<div class="need">Sin necesidades registradas<\/div>'\)/, 'saved needs replace the reported list in the card, so items are not shown twice');
  const unfiltered = render('c1', true);
  assert.match(unfiltered, /<b>En ruta <span class="badge urgent">Urgente<\/span><\/b><div class="chips"><span class="chip text">Agua<\/span>[\s\S]*<b>Abierta <\/b><div class="chips"><span class="chip text">Comida<\/span>/, 'normal priority shows no label');
  assert.doesNotMatch(unfiltered, /Ropa/);
  assert.equal(render('c2', true), '');
});
test('public page never labels points as "Aprobado"', () => {
  assert.doesNotMatch(template, /'Aprobado'|<label>Aprobado<\/label>/);
});
test('public Buscar also takes Google Maps links or coordinates and drops a temporary pin to register or open in Maps', () => {
  assert.doesNotMatch(template, /id="gpsCheck"/, 'no separate Verifica GPS box');
  assert.match(template, /<input class="topsearch" id="search" type="search" placeholder="Buscar…"/);
  const grab = re => template.match(re)[0];
  const src = [grab(/function coordsFromMapLink[^\n]*?:null\}/), grab(/function isMapLink[^\n]*/)].join('\n');
  const {coordsFromMapLink, isMapLink} = new Function(`${src};return {coordsFromMapLink,isMapLink}`)();
  assert.deepEqual(coordsFromMapLink('https://maps.google.com/maps?q=7.3845953941345215%2C-80.45873260498047&z=17&hl=en'), [7.3845953941345215, -80.45873260498047]);
  assert.deepEqual(coordsFromMapLink('https://www.google.com/maps/place/X/@7.40,-80.30,17z/data=!8m2!3d7.4012345!4d-80.3123456'), [7.4012345, -80.3123456]);
  assert.deepEqual(coordsFromMapLink('7.40, -80.30'), [7.40, -80.30], 'plain lat, lon');
  assert.equal(coordsFromMapLink('https://maps.app.goo.gl/abc'), null);
  assert.ok(isMapLink('https://maps.app.goo.gl/abc') && !isMapLink('Venao'));
  // A location or a !code never filters the list.
  assert.match(template, /function searchQuery\(\)\{const v=\$\('#search'\)\.value\.trim\(\);return v\.startsWith\('!'\)\|\|coordsFromMapLink\(v\)\|\|isMapLink\(v\)\?'':v\.toLocaleLowerCase\(\)\}/);
  assert.match(template, /const q=searchQuery\(\);/);
  assert.match(template, /No es un punto registrado\. ¿Quieres registrarlo\?<\/p><div class="gpsactions"><button type="button" class="gpsadd" data-register="/);
  assert.match(template, /\+ Registrar punto aquí<\/button><a class="gpsopen" target="_blank"/);
  assert.match(template, /gps\.value='https:\/\/maps\.google\.com\/\?q='\+b\.dataset\.register/, 'Registrar opens the form with the GPS filled in');
  assert.match(template, /className:'verifyicon'/);
});
test('public Buscar: «!code» + Enter signs in with an access code, with no hint in the interface', () => {
  assert.match(template, /if\(e\.key!=='Enter'\|\|!v\.startsWith\('!'\)\)return;/);
  assert.match(template, /fetch\('\/api\/session\/code',\{method:'POST'/);
  assert.match(template, /\$\('#search'\)\.value='';/, 'the code is cleared from the box');
  assert.doesNotMatch(template, /placeholder="[^"]*(código|!)/i, 'the placeholder does not reveal it');
  assert.match(template, /<span class="codeuser" id="codeUser" hidden><\/span>/);
  const show = template.match(/function showCodeUser\(\)[^\n]*/)[0];
  assert.match(show, /<button type="button" id="copyMyId" title="Copiar mi ID de usuario">\$\{clean\(u\.name\)\}<\/button><span class="idtip" id="idTip" role="status" hidden><\/span>/, 'just the name, no link; tapping it copies the user ID');
  assert.doesNotMatch(show, /<a href=/);
  assert.doesNotMatch(show, /p\.label/, 'no role or panel names in the header');
});
test('public map has an opt-in "my location" control with a distinct person pin kept in the browser', () => {
  assert.match(template, /id='locateBtn'|b\.id='locateBtn'/);
  assert.match(template, /L\.DomEvent\.on\(b,'click',locateUser\)/);
  assert.match(template, /className:'usericon'/);
  const src = template.match(/function locateUser\(\)[^\n]*/)[0];
  assert.match(src, /navigator\.geolocation\.watchPosition/);
  assert.doesNotMatch(src, /fetch\(/);
});
test('runner top nav links to the public Agregar punto form, which opens from #agregar', () => {
  assert.match(runner, /<header class="top">[\s\S]*?<a class="add" href="\/#agregar">\+ Agregar punto<\/a>[\s\S]*?<\/header>/);
  assert.match(template, /location\.hash==='#agregar'[^\n]*\$\('#addBtn'\)\.click\(\)/);
});
test('runner map pin popup shows the full need card with ¡Yo lo llevo! like the list', () => {
  const grab = re => runner.match(re)[0];
  const src = [grab(/const clean=[^\n]*/), grab(/const urgencyLabel=[^\n]*/), grab(/function locationLink[^\n]*/), grab(/const foldNeed=[^\n]*/), grab(/function needChips[^\n]*/), grab(/function needDetails[^\n]*/), grab(/function popupList[^\n]*/), grab(/function popupNeed[^\n]*/)].join('\n');
  // mine: carreras en curso; con 10 (el máximo) ya no se ofrece tomar otra.
  const make = (mine, available, agentMode = false) => new Function('mine', 'max', 'followup', 'available', 'apple', 'agentMode', `${src};return {popupNeed,popupList}`)(mine, 10, null, available, false, agentMode);
  const n = {id: 'n1', need: 'Agua <b>', urgency: 'urgente', camp: {id: 'c1', code: 'C-1', area: 'Cañas', coords: [7.4, -80.3]}};
  const html = make([], [n]).popupNeed(n, true);
  assert.match(html, /data-take="n1">¡Yo lo llevo!/);
  assert.doesNotMatch(make([], [n], true).popupNeed(n, true), /data-take/, 'agents never get ¡Yo lo llevo!');
  assert.match(html, /Agua &lt;b&gt;/);
  assert.match(html, /data-back="c1"/);
  assert.doesNotMatch(html, /class="where"/);
  assert.doesNotMatch(make([n], []).popupNeed(n, false), /data-take/, 'own run');
  assert.match(make([{id: 'x'}], [n]).popupNeed(n, false), /data-take="n1"/, 'one run in progress still allows taking another');
  assert.doesNotMatch(make(Array.from({length: 10}, (_, i) => ({id: 'r' + i})), [n]).popupNeed(n, false), /data-take/, 'at the limit');
  assert.match(make([], [n]).popupList({camp: n.camp, needs: [n]}), /class="popneed" data-need="n1"/);
});
test('runner name gate also requires a phone number and sends it with the session', () => {
  const gate = runner.slice(runner.indexOf('id="gateForm"'), runner.indexOf('</form>', runner.indexOf('id="gateForm"')));
  assert.match(gate, /<input name="name" maxlength="120"/);
  assert.match(gate, /<input name="telephone" type="tel"/);
  // Without a code, name alone is not enough: both name and phone are needed.
  assert.match(runner, /if\(!accessCode&&!\(name&&telephone\)\)/);
  assert.match(runner, /body:JSON\.stringify\(creds\)/);
});
test('selected point offers Reportar necesidad with required requester name and phone', () => {
  assert.match(template, /\$\{r\.id\?'<button type="button" class="add reportneed" id="reportNeed">\+ Reportar necesidad<\/button>':''\}/);
  const form = template.slice(template.indexOf('id="needForm"'), template.indexOf('</form>', template.indexOf('id="needForm"')));
  assert.match(form, /<textarea name="need" required/);
  assert.match(form, /<input name="requesterName" required/);
  assert.match(form, /<input name="requesterPhone" type="tel"[^>]* required/);
  assert.match(template, /fetch\('\/api\/camps\/'\+encodeURIComponent\(needCamp\.id\)\+'\/needs'/);
  assert.match(template, /!validPhone\(v\('requesterPhone'\)\)/);
  assert.match(camp, /Reportado por \$\{clean\(n\.requesterName\)\}/);
  assert.match(admin, /Reportado por \$\{clean\(n\.requesterName\)\}/);
  assert.match(runner, /Reportado por:<\/b> \$\{clean\(n\.requesterName\)\}/);
});
test('admin stats bar: need counters filter approved points; Puntos por revisar and Archivados toggle the view', () => {
  const dash = admin.slice(admin.indexOf('id="dashboard"'));
  assert.ok(dash.indexOf('id="adminStats"') < dash.indexOf('id="items"'));
  const src = admin.match(/function renderStats\(\)[^\n]*/)[0];
  const draw = (needFilter, view, coordFilter = '') => { const els = {'#adminStats': {}, '#coordPicker': {hidden: true}}; new Function('$', 'needFilter', 'view', 'coordFilter', 'coordLabel', 'clean', 'allNeeds', 'points', `${src};renderStats()`)(s => els[s], needFilter, view, coordFilter, () => coordFilter ? 'Marta' : 'Todos', s => String(s), [{campId: 'a', status: 'OPEN'}, {campId: 'a', status: 'OPEN'}, {campId: 'a', status: 'IN_PROGRESS'}, {campId: 'a', status: 'COMPLETED'}, {campId: 'z', status: 'OPEN'}], [{id: 'p', status: 'pending'}, {id: 'a', status: 'approved'}, {id: 'z', status: 'archived'}, {id: 'y', status: 'archived'}, {id: 'r', status: 'rejected'}]); return els['#adminStats'].innerHTML; };
  let html = draw('OPEN', 'approved');
  assert.match(html, /class="open active" data-stat="OPEN" aria-pressed="true"><b>2<\/b><span>Necesidades abiertas/);
  assert.match(html, /data-stat="IN_PROGRESS" aria-pressed="false"><b>1<\/b><span>Necesidades en curso/);
  assert.match(html, /data-stat="COMPLETED" aria-pressed="false"><b>1<\/b>/);
  assert.match(html, /^<button type="button" class="review urgent " data-view="pending" aria-pressed="false"><b>1<\/b><span>⚠ Puntos por revisar<\/span><\/button><button type="button" class="all /);
  assert.match(html, /class="archived " data-view="archived" aria-pressed="false"><b>2<\/b><span>Archivados/);
  assert.match(html, /class="rejected " data-view="rejected" aria-pressed="false"><b>1<\/b><span>Duplicados/);
  assert.match(admin, /<button class="reject" data-merge="\$\{clean\(p\.id\)\}">Duplicado…<\/button>/);
  assert.match(admin, /id="mergeNoCopy">Marcar duplicado sin copiar datos<\/button>/);
  assert.match(admin, /\$\{p\.status==='rejected'\?'duplicado':clean\(p\.status\)\}<\/span>/);
  assert.match(html, /<button type="button" class="all " data-view="approved" aria-pressed="false"><b>1<\/b><span>Puntos aprobados/);
  assert.match(html, /class="coord " data-coord-open aria-haspopup="listbox" aria-expanded="false"><b>Todos<\/b><span>Coordinador ▾/);
  html = draw('', 'archived', 'marta');
  assert.match(html, /class="archived active" data-view="archived" aria-pressed="true"/);
  assert.match(html, /class="coord active"[^>]*><b>Marta<\/b>/);
  assert.match(draw('', 'approved'), /class="all active" data-view="approved" aria-pressed="true"/);
  assert.match(admin, /view='approved',needFilter=''/);
  const click = admin.match(/if\(b\.dataset\.view\)\{(.*?)\}else/)[1];
  const press = (view, next, needFilter = '') => new Function('b', 'view', 'needFilter', 'coordFilter', `${click};return view`)({dataset: {view: next}}, view, needFilter, '');
  assert.equal(press('approved', 'pending'), 'pending');
  assert.equal(press('pending', 'pending'), 'approved');
  assert.equal(press('approved', 'approved'), 'pending');
  assert.equal(press('approved', 'approved', 'OPEN'), 'approved');
  assert.equal(press('archived', 'archived'), 'approved');
  assert.equal(press('approved', 'rejected'), 'rejected');
  assert.equal(press('rejected', 'rejected'), 'approved');
  assert.doesNotMatch(admin, /id="(pendingTab|allTab|archivedTab|needsTab)"|Necesidades de servicio<\/button>|renderAllNeeds/);
  assert.match(admin, /view==='archived'\?p\.status==='archived':view==='rejected'\?p\.status==='rejected':p\.status==='approved'\)/);
  assert.match(admin, /p\.status==='rejected'\?\(p\.mergedInto\?mergedLink\(p\):`<button data-id="\$\{clean\(p\.id\)\}" data-decision="approved">Aprobar<\/button>`\)/);
  assert.match(admin, /data-decision="approved">Recuperar punto<\/button>/);
});
test('camp needs show their creation time, or «Actualizado» once revised, in Spanish on admin, runner and public pages', () => {
  assert.match(admin, /n\.revisedAt\?'Actualizado: '\+stamp\(n\.revisedAt\):n\.createdAt\?'Creada: '\+stamp\(n\.createdAt\)/);
  assert.match(template, /n\.revisedAt\?'Actualizado: '\+needStamp\(n\.revisedAt\):n\.createdAt\?'Creada: '/);
  assert.match(camp, /n\.revisedAt\?'Actualizado: '\+clean\(stamp\(n\.revisedAt\)\):'Creada: '\+\(n\.createdAt\?clean\(stamp\(n\.createdAt\)\)/);
  const grab = re => runner.match(re)[0];
  const src = [grab(/const clean=[^\n]*/), grab(/const stamp=[^\n]*/), grab(/const urgencyLabel=[^\n]*/), grab(/function locationLink[^\n]*/), grab(/const foldNeed=[^\n]*/), grab(/function needChips[^\n]*/), grab(/function needDetails[^\n]*/)].join('\n');
  const needDetails = new Function('apple', `${src};return needDetails`)(false);
  const camp1 = {id: 'c', code: 'C-1'};
  assert.match(needDetails({id: 'n', need: 'Agua', urgency: 'urgente', createdAt: '2026-10-01T12:00:00Z', camp: camp1}), /<b>Creada:<\/b> /);
  assert.doesNotMatch(needDetails({id: 'n', need: 'Agua', urgency: 'urgente', createdAt: '2026-10-01T12:00:00Z', revisedAt: '2026-10-02T12:00:00Z', camp: camp1}), /Creada/);
  assert.match(needDetails({id: 'n', need: 'Agua', urgency: 'urgente', revisedAt: '2026-10-02T12:00:00Z', camp: camp1}), /<b>Actualizado:<\/b> /);
  assert.match(needDetails({id: 'n', need: 'Agua', urgency: 'urgente', camp: camp1}), /<b>Creada:<\/b> sin fecha registrada/);
});
test('runner need card lists the camp contact phones as tap-to-call links', () => {
  const grab = re => runner.match(re)[0];
  const src = [grab(/const clean=[^\n]*/), grab(/const stamp=[^\n]*/), grab(/const urgencyLabel=[^\n]*/), grab(/function locationLink[^\n]*/), grab(/const foldNeed=[^\n]*/), grab(/function needChips[^\n]*/), grab(/function needDetails[^\n]*/)].join('\n');
  const needDetails = new Function('apple', `${src};return needDetails`)(false);
  const html = needDetails({id: 'n', need: 'Agua', urgency: 'urgente', camp: {id: 'c', code: 'C-1', contacts: [{name: 'Ana', phone: '6555-1234'}, {phone: '+507 6000 0000'}]}});
  assert.match(html, /Contactos del punto:<\/b> Ana · <a href="tel:65551234">6555-1234<\/a> · <a href="tel:\+50760000000">\+507 6000 0000<\/a>/);
  assert.match(needDetails({id: 'n', need: 'Agua', urgency: 'urgente', camp: {id: 'c', code: 'C-1'}}), /sin teléfono registrado/);
});
test('admin header holds Agregar punto and boletín actions, shown only when logged in', () => {
  const header = admin.slice(admin.indexOf('<header>'), admin.indexOf('</header>'));
  for (const re of [/id="topActions" hidden/, /id="addPoint">\+ Agregar punto</, /id="boletinToggle">Generar boletín</, /href="\/report\/boletin_general\.html"/]) assert.match(header, re);
  const dash = admin.slice(admin.indexOf('id="dashboard"'));
  assert.doesNotMatch(dash, /id="addPoint"|id="boletinToggle"/);
  assert.match(admin, /\$\('#dashboard'\)\.hidden=\$\('#topActions'\)\.hidden=\$\('#topSearch'\)\.hidden=\$\('#adminFoot'\)\.hidden=false/);
});
test('every boletín link and button in admin asks for the access code, checked by the server', () => {
  const anchors = admin.match(/<a [^>]*href="(?:[^"]*boletin_general\.html[^"]*|\$\{url\})"[^>]*>/g) || [];
  assert.ok(anchors.length >= 2, 'top link and post-generation link');
  for (const a of anchors) assert.match(a, /class="[^"]*boletinlink/, a);
  assert.match(admin, /<dialog id="boletinAccess"/);
  assert.match(admin, /adminRequest\('\/api\/admin\/boletin-access','POST',\{code\}\)/);
  assert.match(admin, /\$\('#boletinToggle'\)\.onclick=\(\)=>\{[^\n]*askBoletinCode\(/);
  assert.match(admin, /accessCode:code/);
  assert.doesNotMatch(admin, /PRINT5/i, 'code is validated server-side, not shipped in the page');
});
test('point form groups fields into Dónde / Con quién / Qué necesitan, with admin-only fields only for admins', () => {
  const pub = PointForm.html(), adm = PointForm.html({admin: true}), edit = PointForm.html({admin: true, edit: true});
  for (const title of ['¿Dónde estás?', '¿Con quién estás?', '¿Qué necesitas?']) { assert.ok(pub.includes(title), title); assert.ok(adm.includes(title), title); }
  for (const name of ['area', 'place', 'directions', 'gps', 'people', 'adults', 'children', 'babies', 'olderAdults', 'disabled', 'coordinator', 'coordinatorPhone', 'extra']) assert.match(pub, new RegExp(`name="${name}"`), name);
  for (const name of ['code', 'row', 'ordinal', 'service', 'serviceNotes', 'notes']) { assert.doesNotMatch(pub, new RegExp(`name="${name}"`), name); assert.match(adm, new RegExp(`name="${name}"`), name); }
  assert.doesNotMatch(pub, /Solo administración|name="note"/);
  assert.match(edit, /name="note"/);
  assert.doesNotMatch(pub, /name="(contact|phone|secondContact|secondPhone|thirdContact|thirdPhone)"/);
  assert.match(pub, /class="pf-gps"[^>]*>.*Enviar mi ubicación GPS<\/button>/);
  const where = pub.slice(pub.indexOf('¿Dónde estás?'), pub.indexOf('¿Con quién estás?'));
  assert.ok(where.indexOf('class="pf-gps"') < where.indexOf('name="gps"') && where.indexOf('name="gps"') < where.indexOf('name="area"'), 'GPS button first, then link, then zona');
  const who = pub.slice(pub.indexOf('¿Con quién estás?'), pub.indexOf('¿Qué necesitas?'));
  assert.ok(who.indexOf('name="coordinator"') < who.indexOf('name="coordinatorPhone"') && who.indexOf('name="coordinatorPhone"') < who.indexOf('name="people"'), 'coordinator first');
  for (const name of ['people', 'adults', 'children', 'babies', 'olderAdults', 'disabled']) assert.match(who, new RegExp(`name="${name}" data-filter="count" inputmode="numeric" maxlength="3"`), name);
  assert.equal((who.match(/class="pf-ico"><svg/g) || []).length, 7, 'icons on coordinator and the six counts');
  assert.doesNotMatch(pub, /name="needs"|name="generator"/);
  assert.doesNotMatch(pub, /Agrega a alguien más/, 'a new empty row appears by itself');
});
test('point form people: one compact numbered row each, always one empty row ready to fill', () => {
  assert.doesNotMatch(pointformSrc, /Persona \$\{i \+ 1\}<button/, 'no Persona N heading');
  assert.match(pointformSrc, /<div class="pf-member"><span class="pf-mnum" aria-hidden="true">\$\{i \+ 1\}<\/span>\$\{memberFields\.map\(\(\[k, l, f\]\) => `<input class="pf-m-\$\{k\}" data-k="\$\{k\}" value="\$\{esc\(m\[k\]\)\}" placeholder="\$\{l\}"/);
  assert.match(pointformSrc, /list = list\.filter\(memberFilled\);\n    list\.push\(\{\}\);/, 'saved people plus one empty row');
  assert.match(pointformSrc, /if \(last && \[\.\.\.last\.querySelectorAll\('input'\)\]\.some\(el => el\.value\.trim\(\)\)\) box\.insertAdjacentHTML\('beforeend', memberRow\(\{\}, rows\.length\)\);/, 'typing in the last row adds the next one');
  assert.match(pointformSrc, /grid-template-areas:'num name phone age ced del'/, 'one line per person on desktop');
});
test('point form people: legacy contacts become people, phone rule accepts any person', () => {
  assert.deepEqual(PointForm.membersOf({contact: 'Ana', phone: '6555-1234', thirdPhone: '6000-0000'}), [{name: 'Ana', phone: '6555-1234'}, {name: '', phone: '6000-0000'}]);
  assert.deepEqual(PointForm.membersOf({members: [{name: 'Eva', cedula: '8-1-1'}], contact: 'Old'}), [{name: 'Eva', cedula: '8-1-1'}]);
  assert.equal(PointForm.hasPhone({members: [{name: 'Sin tel'}, {phone: '+507 6555 1234'}]}), true);
  assert.equal(PointForm.hasPhone({members: [{phone: '12-34'}]}), false);
  assert.equal(PointForm.hasPhone({members: []}), false);
});
test('public and admin point forms use the shared PointForm with GPS link and dynamic people', () => {
  assert.match(template, /<script src="\.\/pointform\.js"><\/script>/);
  assert.match(template, /<form id="form"><div class="formtop">[^]*?<\/div><div id="pointFields"><\/div>/);
  assert.match(template, /PointForm\.render\(\$\('#pointFields'\),\{nearby:publicNearby,me:codeMe\(\)\}\)/);
  assert.match(template, /if\(!PointForm\.hasPhone\(record\)\)/);
  assert.match(admin, /<script src="\.\/pointform\.js"><\/script>/);
  assert.match(admin, /PointForm\.render\(\$\('#pointFormFields'\),\{admin:true,nearby:addPointNearby\}\)/);
  assert.match(admin, /function drawCampFields\(p\)\{PointForm\.render\(\$\('#campFields'\),\{admin:true,edit:true,point:p\}\)\}/);
  assert.match(admin, /'PATCH',PointForm\.values\(\$\('#campFields'\)\)/);
  // Igual que «Editar datos del punto»: Cancelar a la izquierda y Guardar a la derecha, arriba y abajo.
  assert.match(admin, /<section id="newPoint" hidden><div class="camp-head"><button type="button" class="quiet-red" id="pointCancelTop">Cancelar<\/button><h2>Agregar punto aprobado<\/h2><button type="submit" form="pointForm">Guardar y publicar<\/button><\/div>/);
  assert.match(admin, /<div id="pointFormFields"><\/div><div class="form-actions"><button type="button" class="quiet-red" id="pointCancel">Cancelar<\/button><span class="spacer"><\/span><button type="submit">Guardar y publicar<\/button><\/div><\/form>/);
  assert.match(admin, /\$\('#pointCancel'\)\.onclick=\$\('#pointCancelTop'\)\.onclick=/);
  assert.match(admin, /buttons=\[\.\.\.document\.querySelectorAll\('#newPoint button\[type="submit"\]'\)\];buttons\.forEach\(b=>b\.disabled=true\)/);
  assert.match(pointformSrc, /const mapLink = \(\[lat, lng\]\) => `https:\/\/maps\.google\.com\/\?q=\$\{lat\.toFixed\(6\)\},\$\{lng\.toFixed\(6\)\}`;[\s\S]*gps\.value = mapLink\(\[lat, lng\]\);/);
});
test('Qué necesitas checkboxes round-trip through the Necesidades text, with Otros free text', () => {
  const pub = PointForm.html();
  const what = pub.slice(pub.indexOf('¿Qué necesitas?'));
  for (const label of ['Agua', 'Alimentos', 'Refugio (carpa / lona / toldo)', 'Kit de cocina', 'Kit de bebé', 'Gasolina para generador', 'Diésel para generador']) assert.ok(what.includes(`data-need="${label}"`), label);
  // Reparación de casa es solo para administración (constructores), nunca en el formulario público.
  assert.ok(!what.includes('Reparación de casa'));
  assert.ok(PointForm.html({admin: true}).includes('data-need="Reparación de casa"'));
  assert.match(what, /class="pf-otros-check"><span><svg[^]*?Otros<\/span>/);
  assert.equal(PointForm.needsText(['Agua', 'Kit de bebé'], ' Medicinas, ropa '), 'Agua, Kit de bebé, Medicinas, ropa');
  assert.deepEqual(PointForm.parseNeeds('agua, Kit de bebe, Medicinas, ropa'), {checked: ['Agua', 'Kit de bebé'], other: 'Medicinas, ropa'});
  assert.deepEqual(PointForm.parseNeeds('Agua y comida para 20'), {checked: [], other: 'Agua y comida para 20'});
  assert.deepEqual(PointForm.parseNeeds(''), {checked: [], other: ''});
  assert.deepEqual(PointForm.parseNeeds('Reparacion de casa, Agua'), {checked: ['Reparación de casa', 'Agua'], other: ''});
});
test('counts and edad accept only digits up to 999; teléfonos only digits and + - ( )', () => {
  assert.equal(PointForm.filters.count('12a3x4'), '123');
  assert.equal(PointForm.filters.age('dos 7'), '7');
  assert.equal(PointForm.filters.phone('Tel: +507 (6555)-1234 ext'), ' +507 (6555)-1234 ');
  const pub = PointForm.html();
  assert.match(pub, /name="coordinatorPhone" data-filter="phone" type="tel"/);
  assert.equal(PointForm.hasPhone({coordinatorPhone: '6555-1234', members: []}), true);
});
test('point form: numbered 3 steps, persona fields ordered nombre, teléfono, edad, cédula', () => {
  const pub = PointForm.html();
  assert.match(pub, /<legend><span class="pf-step">1<\/span>¿Dónde estás\?<\/legend>/);
  assert.match(pub, /<legend><span class="pf-step">2<\/span>¿Con quién estás\?<\/legend>/);
  assert.match(pub, /<legend><span class="pf-step">3<\/span>¿Qué necesitas\?<\/legend>/);
  assert.match(pointformSrc, /const memberFields = \[\['name', 'Nombre', ''\], \['phone', 'Teléfono', 'phone'\], \['age', 'Edad', 'age'\], \['cedula', 'Cédula', ''\]\];/);
});
test('¡Ayuda! form has no intro and only confirms "Se ha enviado tu solicitud" after sending', () => {
  const modal = template.slice(template.indexOf('id="modal"'), template.indexOf('<form id="form">'));
  assert.match(modal, /<h2>¡Ayuda!<\/h2><div class="notice sent"/);
  assert.match(template, /id="addBtn">¡Ayuda!<\/button>/);
  assert.doesNotMatch(template, /formIntro/);
  // Signed in with a code, the button reads «+ Agregar punto» like the admin and runner pages; otherwise «¡Ayuda!».
  assert.match(template, /\$\('#addBtn'\)\.textContent=\$\('#addBtnBar'\)\.textContent=u\?'\+ Agregar punto':'¡Ayuda!';/);
  assert.doesNotMatch(template, /Un administrador la revisará/);
  assert.match(modal, /<div class="notice sent" id="sentNotice" role="status" hidden><b id="sentTitle">Se ha enviado tu solicitud\.<\/b>/);
  assert.doesNotMatch(template, /Se envía a revisión/);
  assert.match(template, /PointForm\.render\(\$\('#pointFields'\),\{nearby:publicNearby,me:codeMe\(\)\}\);\$\('#sentTitle'\)\.textContent=[^;]*;showSent\(true\);/);
  assert.match(template, /\$\('#addBtn'\)\.onclick=\(\)=>\{showSent\(false\);/);
});
test('admin header: AZUERO / ADMIN with Buscar… right after it; Salir in a small footer; coordinator typeahead', () => {
  const header = admin.slice(admin.indexOf('<header>'), admin.indexOf('</header>'));
  assert.match(header, /^<header><span>AZUERO \/ ADMIN<\/span><div class="topsearch" id="topSearch" hidden><input id="search" type="search" placeholder="Buscar…"/);
  assert.match(admin, /<\/main><footer class="adminfoot" id="adminFoot" hidden><button type="button" id="logout">Salir<\/button><\/footer>/);
  assert.doesNotMatch(admin, /class="toolbar"|id="coordinatorFilter"|CURACIÓN/);
  const src = ['coordNames', 'coordLabel', 'drawCoordList'].map(n => admin.match(new RegExp(`(const ${n}=|function ${n}\\()[^\\n]*`))[0]).join('\n');
  const els = {'#coordSearch': {value: 'mar'}, '#coordList': {}};
  const api = new Function('$', 'points', 'clean', 'coordFilter', `${src};return {drawCoordList, coordLabel}`)(s => els[s], [{coordinator: 'Marta Ruiz'}, {coordinator: 'Luis'}, {coordinator: ' Marta Ruiz '}, {}], s => String(s), 'marta ruiz');
  api.drawCoordList();
  assert.match(els['#coordList'].innerHTML, /^<button type="button" role="option" data-coord="marta ruiz" class="hl sel" aria-selected="true">Marta Ruiz<\/button>$/);
  els['#coordSearch'].value = ''; api.drawCoordList();
  assert.equal((els['#coordList'].innerHTML.match(/role="option"/g) || []).length, 4, 'Todos, Luis, Marta Ruiz, Sin coordinador');
  assert.equal(api.coordLabel(), 'Marta Ruiz');
});
test('admin map pins the current view, including por revisar and archivados, with only a color legend (no location counts)', () => {
  const src = admin.match(/function renderMap\(rows\)\{[^\n]*/)[0];
  assert.match(src, /const located=rows\.filter\(p=>Array\.isArray\(p\.coords\)&&p\.coords\.length===2\)/);
  assert.doesNotMatch(src, /status!=='archived'/);
  assert.match(admin, /const pinColor=\{approved:'#277a66',pending:'#e57450',archived:'#8a8f91',rejected:'#9a4536'\}/);
  const els = {'#mapCount': {}};
  new Function('$', 'L', 'pinColor', `${src};renderMap([{status: 'archived', coords: [7.4, -80.3]}, {status: 'pending'}])`)(s => els[s], undefined, {});
  assert.match(els['#mapCount'].textContent, /^Verde: aprobado · naranja: por revisar · gris: archivado · rojo: duplicado/);
  assert.doesNotMatch(admin, /ubicaciones'\} en mapa|sin GPS \(solo en la lista\)/);
});
test('admin: no pending points means a plain Puntos por revisar box', () => {
  const src = admin.match(/function renderStats\(\)[^\n]*/)[0];
  const els = {'#adminStats': {}, '#coordPicker': {hidden: true}};
  new Function('$', 'needFilter', 'view', 'coordFilter', 'coordLabel', 'clean', 'allNeeds', 'points', `${src};renderStats()`)(s => els[s], '', 'approved', '', () => 'Todos', String, [], [{id: 'a', status: 'approved'}]);
  assert.match(els['#adminStats'].innerHTML, /^<button type="button" class="review  " data-view="pending" aria-pressed="false"><b>0<\/b><span>Puntos por revisar<\/span>/);
});
test('¡Ayuda! form has a Guardar button at the top too, and both disable while sending', () => {
  assert.match(template, /<form id="form"><div class="formtop"><button class="add" type="submit">Guardar punto<\/button><\/div><div id="pointFields"><\/div>/);
  assert.equal((template.slice(template.indexOf('<form id="form">'), template.indexOf('</form>', template.indexOf('<form id="form">'))).match(/type="submit">Guardar punto/g) || []).length, 2);
  assert.match(template, /const subs=\[\.\.\.e\.target\.querySelectorAll\('\[type=submit\]'\)\];subs\.forEach\(x=>x\.disabled=true\)/);
});
test('admin por revisar: list items and map pins share a number, and selecting either highlights both', () => {
  assert.match(admin, /<span class="pinnum"[^`]*>\$\{i\+1\}<\/span>/);
  assert.match(admin, /num=p\.status==='pending'\?rows\.indexOf\(p\)\+1:0/);
  assert.match(admin, /if\(point\?\.status==='pending'\)\{selectPending\(point\.id,true\);return\}/);
  assert.match(admin, /else\{if\(num\)selectPending\(p\.id\);/);
  assert.match(admin, /el\.classList\.toggle\('selected',el\.dataset\.pointId===id\)/);
});
test('admin camp list: en ruta needs show the runner name and tap-to-call phone', () => {
  const pick = re => admin.match(re)[0];
  const src = [/const urgentBadge=[^;]*;/, /const needLook=[^\n]*/, /const needActivity=[^\n]*/, /const crewOf=[^\n]*/, /const runnerLine=[^\n]*/, /const requester=[^\n]*/, /const cardNeeds=[^\n]*/, /const assignLine=[^\n]*/].map(pick).join('\n');
  const build = (doneShown = new Set()) => new Function('clean', 'stamp', 'needTimes', 'needChips', 'needStatsHtml', 'allNeeds', 'runners', 'builders', 'doneShown', `const users=[...runners,...builders];${src};const needClass=n=>(needLook[n.status]||['deleted'])[0],needBadge=n=>'';return cardNeeds({id:'c'})`)(String, String, () => '', n => n.need, () => '<i>stats</i>', [
    {campId: 'c', status: 'IN_PROGRESS', urgency: 'urgente', need: 'techo', kind: 'repair', builderId: 'b1', runnerId: 'r1'},
    {campId: 'c', status: 'IN_PROGRESS', urgency: 'urgente', need: 'agua', runnerId: 'r1'},
    {campId: 'c', status: 'IN_PROGRESS', urgency: 'urgente', need: 'pan'},
    {campId: 'c', status: 'OPEN', urgency: 'urgente', need: 'arroz', runnerId: 'r1'},
    {id: 'n5', campId: 'c', status: 'OPEN', urgency: 'normal', need: 'leche'},
    {id: 'n6', campId: 'c', status: 'OPEN', urgency: 'normal', need: 'pared', kind: 'repair'},
    {id: 'n7', campId: 'c', status: 'COMPLETED', urgency: 'normal', need: 'agua vieja'},
  ], [{id: 'r1', name: 'Ana', telephone: '6000-1234'}], [{id: 'b1', name: 'Luis Obra', telephone: '6000-9999'}], doneShown);
  const html = build();
  assert.match(html, /techo.*Constructor: <b>Luis Obra<\/b> · <a href="tel:60009999">/);
  assert.match(html, /agua.*Corredor: <b>Ana<\/b> · <a href="tel:60001234">6000-1234<\/a>/);
  // Sin asignar: botón «Asignar» del equipo que corresponde; asignada aunque siga abierta: se ve quién la tiene.
  assert.match(html, /pan<button type="button" class="assignbtn" data-assign-need="undefined" aria-haspopup="listbox">\+ Asignar corredor<\/button>/);
  assert.match(html, /leche<button type="button" class="assignbtn" data-assign-need="n5"[^>]*>\+ Asignar corredor<\/button>/);
  assert.match(html, /pared<button type="button" class="assignbtn" data-assign-need="n6"[^>]*>\+ Asignar constructor<\/button>/);
  assert.match(html, /arroz.*Corredor: <b>Ana<\/b>/);
  assert.doesNotMatch(html, /agua vieja/, 'completed needs are hidden by default');
  assert.match(build(new Set(['c'])), /agua vieja/, 'shown after tapping the Completadas box');
  assert.doesNotMatch(build(new Set(['c'])), /agua vieja<button/, 'no Asignar on completed');
  // Cajas de conteo en lugar del título.
  assert.match(html, /^<aside class="card-needs"><div class="needstats"><i>stats<\/i><\/div><ul>/);
  assert.doesNotMatch(admin, /Necesidades de servicio \(/);
});
test('admin camp list: needs sort by latest activity first', () => {
  const src = admin.match(/const needActivity=[^\n]*/)[0];
  const needActivity = new Function(`${src};return needActivity`)();
  const rows = [{id: 'old', createdAt: '2026-10-01T10:00:00Z'}, {id: 'taken', createdAt: '2026-09-01T10:00:00Z', updatedAt: '2026-10-12T10:00:00Z'}, {id: 'done', createdAt: '2026-09-01T10:00:00Z', completedAt: '2026-10-05T10:00:00Z'}];
  assert.deepEqual(rows.sort((a, b) => needActivity(b) - needActivity(a)).map(n => n.id), ['taken', 'done', 'old']);
  assert.match(admin, /\.sort\(\(a,b\)=>needActivity\(b\)-needActivity\(a\)\)/);
});
test('builder page: own login with full name, repairs API only, several jobs at once and no public add link', () => {
  const builder = readFileSync(new URL('./builder.html', import.meta.url), 'utf8');
  assert.match(builder, /<title>Constructores · Azuero<\/title>/);
  assert.match(builder, /<label>Nombre completo<input name="name" maxlength="120"/);
  assert.match(builder, /fetch\('\/api\/builder\/session'/);
  assert.match(builder, /api\('\/api\/builder\/needs'\);mine=data\.mine\|\|\[\];available=data\.available\|\|\[\];max=data\.max/);
  assert.doesNotMatch(builder, /\/api\/runner\//);
  assert.doesNotMatch(builder, /href="\/#agregar"/);
  assert.match(builder, /data-complete=.*Reparación terminada/);
  assert.match(admin, /name="roles" value="builder"> Constructor</, 'builders are managed as users with the Constructor role');
  assert.match(admin, /await create\(\{need:repairNeedLabel,kind:'repair',/);
});
test('runner need cards show needs as icon chips, with free text as plain chips', () => {
  const grab = re => runner.match(re)[0];
  const src = [grab(/const clean=[^\n]*/), grab(/const foldNeed=[^\n]*/), grab(/function needChips[^\n]*/)].join('\n');
  const needChips = new Function('PointForm', `${src};return needChips`)(PointForm);
  const html = needChips('agua, Alimentos, colchones');
  assert.match(html, /<span class="chip"><svg[^]*?<\/svg>Agua<\/span><span class="chip"><svg[^]*?<\/svg>Alimentos<\/span><span class="chip text"><svg[^]*?<\/svg>colchones<\/span>/);
  assert.match(runner, /<script src="\.\/pointform\.js"><\/script>/);
});
test('need priority is only Normal / Urgente on every page', () => {
  const builder = readFileSync(new URL('./builder.html', import.meta.url), 'utf8');
  for (const page of [template, runner, admin, camp, builder]) assert.doesNotMatch(page, /Today|Tomorrow|Eventually|Eventualmente|value="Urgent"/);
  for (const page of [template, runner, admin]) assert.match(page, /<select name="urgency"><option value="normal"( selected)?>Normal<\/option><option value="urgente">Urgente<\/option><\/select>/);
});
test('admin approved card lists people with phones (no cédula) and opens a read-only camp view', () => {
  const card = admin.match(/function render\(\)[^\n]*/)[0];
  assert.doesNotMatch(card, /Gestionar punto y necesidades|data-copy-id|Copiar enlace|Abrir ficha pública/);
  assert.match(card, /p\.status==='approved'\?peopleLine\(p\)/);
  const src = ['clean', 'labels', 'contactKeys', 'memberList', 'peopleOf', 'peopleLine', 'campViewHtml'].map(k => admin.match(new RegExp(`const ${k}=[^\\n]*`))[0]).join('\n');
  const {peopleLine, campViewHtml} = new Function(`${src};return {peopleLine,campViewHtml}`)();
  const p = {code: 'C-1', area: 'Venao', place: 'Escuela', notes: 'Techo roto', coords: [7.4, -80.1], members: [{name: 'Ana', cedula: '8-123-456', phone: '6555-1234'}, {name: 'Luis'}]};
  const line = peopleLine(p);
  assert.match(line, /<li>Ana · <a href="tel:65551234">6555-1234<\/a><\/li><li>Luis<\/li>/);
  assert.doesNotMatch(line, /8-123-456|Cédula/);
  assert.match(peopleLine({contact: 'Eva', phone: '6000-0000'}), /Eva · <a href="tel:60000000">/);
  const view = campViewHtml(p);
  for (const text of ['Escuela', 'Techo roto', '7.4, -80.1', 'Ana', 'Luis']) assert.ok(view.includes(text), text);
  assert.match(view, /<a class="mapsbtn" href="https:\/\/www\.google\.com\/maps\/search\/\?api=1&query=7\.4%2C-80\.1" target="_blank" rel="noopener noreferrer">Abrir en Google Maps ↗<\/a>/);
  assert.doesNotMatch(campViewHtml({code: 'C-2', place: 'Sin GPS'}), /mapsbtn/);
  const editor = admin.slice(admin.indexOf('id="campEditor"'), admin.indexOf('</section>', admin.indexOf('id="campEditor"')));
  assert.ok(editor.indexOf('id="campView"') < editor.indexOf('id="needList"'));
  assert.match(editor, /id="archiveCamp">Archivar punto<\/button>/);
  assert.match(editor, /^id="campEditor" hidden><div class="camp-head"><button type="button" id="closeCampEditor">← Volver<\/button><h2 id="campTitle">[^<]*<\/h2><a id="campPublicLink" class="iconlink"[^>]*aria-label="Abrir ficha pública">/);
  assert.match(editor, /<article class="camp-card"><div class="camp-body"><div><h3>Datos del punto<\/h3><div id="campView"><\/div><\/div><div class="camp-map" hidden><div id="campMap"[^>]*><\/div>[\s\S]*?<\/div><\/div><footer><button type="button" id="editCamp" class="editbtn"><svg[\s\S]*?<\/svg>Editar<\/button><details id="campCheckup"><\/details><button type="button" id="markCheckup" class="editbtn">/);
  assert.doesNotMatch(admin, /Gestión del punto/);
  assert.match(admin, /matchMedia\('\(min-width:900px\)'\)/);
  assert.match(admin, /fillColor:'#1f6fd1'/);
  assert.match(admin, /history\.pushState\(null,'','\/admin\.html\?camp='\+encodeURIComponent\(id\)\+\(edit\?'&edit=1':needId/);
  assert.match(admin, /\$\('#editCamp'\)\.onclick=\(\)=>openEditor\(activeCamp,true,true\)/);
  assert.match(editor, /<button type="submit" form="campEditForm" class="edit-only">Guardar punto<\/button>/);
  assert.match(editor, /<div class="form-actions"><button type="button" class="quiet-red" id="archiveCamp">Archivar punto<\/button><span class="spacer"><\/span><button type="button" class="quiet-red" id="cancelCampEdit">Cancelar<\/button><button type="submit">Guardar punto<\/button><\/div><\/form>/);
  assert.match(admin, /\$\('#closeCampEditor'\)\.textContent=editing\?'Cancelar':campReturn==='usuarios'\?'← Usuarios':'← Volver';\$\('#closeCampEditor'\)\.classList\.toggle\('quiet-red',editing\);\$\('#adminFoot'\)\.hidden=editing;/);
  assert.match(admin, /#campEditor\.edit-mode>:not\(\.camp-head\):not\(#campEditDetails\)\{display:none!important\}/);
  assert.match(admin, /L\.map\('campMap',\{dragging:false,zoomControl:false,scrollWheelZoom:false,doubleClickZoom:false,touchZoom:false,boxZoom:false,keyboard:false\}\)/);
});
test('admin camp needs: header stats on one line and Editar opens a need-only page with the other needs as reference', () => {
  const editor = admin.slice(admin.indexOf('id="campEditor"'), admin.indexOf('</section>', admin.indexOf('id="campEditor"')));
  assert.match(editor, /<div class="needs-head"><div id="needStats" class="needstats"[^>]*><\/div><button[^]*?Nueva<\/button><\/div><p id="needNotice" class="message" role="status"><\/p><h3 class="need-only" id="otherNeedsTitle">Otras necesidades del punto<\/h3><div id="needList">/);
  assert.match(editor, /<button type="submit" form="needForm" class="need-only" id="saveNeedTop">Guardar cambios<\/button>/);
  assert.doesNotMatch(editor.slice(editor.indexOf('<form id="needForm">')), /id="cancelNeed"|id="saveNeed"|class="form-actions"><button type="button" class="quiet-red" id="cancelNeed/);
  const needStatsHtml = new Function(`${admin.match(/const statIcon=[^\n]*/)[0]}\n${admin.match(/function needStatsHtml[^\n]*/)[0]};return needStatsHtml`)();
  const html = needStatsHtml([{status: 'OPEN'}, {status: 'OPEN'}, {status: 'IN_PROGRESS'}, {status: 'COMPLETED'}, {status: 'DELETED'}]);
  // Total con texto; Abiertas (bandera roja), En ruta (corriendo) y Completadas (casilla) con icono y el nombre como título.
  assert.match(html, /^<div class="all"><b>4<\/b><span>Necesidades<\/span><\/div>/);
  assert.match(html, /<div class="open ico" title="Abiertas" role="img" aria-label="Abiertas: 2"><svg class="stat-ico"[^>]*><path d="M4 22V3"\/><path d="M4 4h13l-2\.5 4\.5L17 13H4" fill="currentColor"\/><\/svg><b>2<\/b><\/div>/);
  assert.match(html, /<div class="route ico" title="En ruta" role="img" aria-label="En ruta: 1"><svg class="stat-ico"[^>]*><circle cx="14" cy="4" r="2"\/>.*?<\/svg><b>1<\/b><\/div>/);
  assert.match(html, /<div class="done ico" title="Completadas" role="img" aria-label="Completadas: 1"><svg class="stat-ico"[^>]*><rect[^>]*\/><path d="m8 12 3 3 5-6"\/><\/svg><b>1<\/b><\/div>$/);
  assert.doesNotMatch(html, /<span>(Abiertas|En ruta|Completadas)<\/span>/);
  assert.match(admin, /class="editbtn" data-edit-need="\$\{clean\(n\.id\)\}"><svg[^`]*<\/svg>Editar<\/button>/);
  assert.doesNotMatch(admin, /Abrir \/ editar/);
  assert.match(admin, /edit\?'&edit=1':needId\?'&need='\+encodeURIComponent\(needId\):''/);
  assert.match(admin, /const shown=focusNeed\?campNeeds\.filter\(n=>n\.id!==focusNeed\):campNeeds;/);
  assert.match(admin, /#campEditor\.need-mode>:not\(\.camp-head\):not\(#needForm\):not\(#otherNeedsTitle\):not\(#needList\)\{display:none!important\}/);
});
test('admin need form uses the same icon checkboxes as Agregar punto and keeps options without a box in Otros', () => {
  const editor = admin.slice(admin.indexOf('id="needForm"'), admin.indexOf('</form>', admin.indexOf('id="needForm"')));
  assert.match(editor, /<div id="needPicker" class="pf"><\/div>/);
  assert.doesNotMatch(editor, /name="need"/);
  assert.match(admin, /\$\('#needPicker'\)\.innerHTML=PointForm\.needsHtml\(\);/);
  assert.match(admin, /need:PointForm\.readNeeds\(\$\('#needPicker'\)\)/);
  assert.equal(PointForm.needsHtml(), PointForm.html().match(/<div class="pf-needs">[\s\S]*?<\/textarea><\/label>/)[0]);
  // Raíz falsa con las casillas públicas (sin «Reparación de casa»).
  const boxes = PointForm.NEEDS.filter(n => n[3] !== 'admin').map(([l]) => ({dataset: {need: l}, checked: false}));
  const otros = {hidden: true}, otrosCheck = {checked: false}, otrosText = {value: ''};
  const root = {querySelectorAll: s => s === '[data-need]' ? boxes : boxes.filter(b => b.checked), querySelector: s => ({'.pf-otros': otros, '.pf-otros-check': otrosCheck, '.pf-otros-text': otrosText})[s]};
  PointForm.fillNeeds(root, 'Agua, Reparación de casa, Medicinas');
  assert.deepEqual(boxes.filter(b => b.checked).map(b => b.dataset.need), ['Agua']);
  assert.equal(otrosText.value, 'Reparación de casa, Medicinas');
  assert.equal(otrosCheck.checked && !otros.hidden, true);
  assert.equal(PointForm.readNeeds(root), 'Agua, Reparación de casa, Medicinas');
  PointForm.fillNeeds(root, '');
  assert.equal(PointForm.readNeeds(root), '');
  assert.equal(otros.hidden, true);
});
test('admin need form: Necesidad first, then Urgencia/Estado/assignee on one line, no Tipo; a repair need shows only a fixed Reparación de casa', () => {
  const form = admin.slice(admin.indexOf('<form id="needForm">'), admin.indexOf('</form>', admin.indexOf('<form id="needForm">')));
  assert.match(form, /<div><b class="needpick-label">Necesidad<\/b><div id="needPicker" class="pf"><\/div><\/div><div class="need-packing" id="needPacking">[\s\S]*?<\/div><div class="need-row"><label>Urgencia<select name="urgency">[\s\S]*?<\/select><\/label><label>Estado<select name="status">[\s\S]*?<\/select><\/label><label class="need-assignee"><span id="needAssignLabel">Corredor \(opcional\)<\/span><select name="assignee" id="needRunner"><\/select><\/label><\/div><p id="needMessage"/);
  assert.doesNotMatch(form, /name="kind"|Tipo/);
  assert.match(admin, /#campEditor\.need-mode #needTitle\{display:none\}/);
  assert.match(admin, /if\(n\.kind==='repair'\)\$\('#needPicker'\)\.innerHTML=`<div class="pf-needs"><span class="pf-need pf-repair pf-fixed"><span>\$\{PointForm\.icons\.repair\}/);
  assert.match(admin, /'Constructor \(opcional\)':'Corredor \(opcional\)'/);
  assert.match(admin, /function needKind\(\)\{if\(editingNeed\)return campNeeds\.find\(n=>n\.id===editingNeed\)\?\.kind\|\|'';/);
});
test('admin point view shows each need as non-clickable icon chips instead of comma text', () => {
  assert.match(admin, /<b>\$\{needBadge\(n\)\}\$\{urgentBadge\(n\)\}<\/b>\$\{needChips\(n\)\}/);
  const needChips = new Function('PointForm', 'clean', `${admin.match(/function needChips[^\n]*/)[0]};return needChips`)(PointForm, String);
  const html = needChips({need: 'Agua, Alimentos, Medicinas, ropa'});
  assert.match(html, /^<div class="needchips"><span class="needchip ">.*<\/svg>Agua<\/span><span class="needchip ">.*<\/svg>Alimentos<\/span><span class="needchip ">.*<\/svg>Medicinas<\/span><span class="needchip ">.*<\/svg>ropa<\/span><\/div>$/);
  assert.doesNotMatch(html, /<button|<input|<a /);
  assert.match(needChips({need: 'Reparación de casa'}), /<span class="needchip repair">/);
});
test('admin shows priority only when urgente, as a red badge', () => {
  const urgentBadge = new Function(`${admin.match(/const urgentBadge=[^;]*;/)[0]};return urgentBadge`)();
  assert.equal(urgentBadge({urgency: 'urgente'}), '<span class="need-urgent">Urgente</span>');
  assert.equal(urgentBadge({urgency: 'normal'}), '');
  assert.match(admin, /\.need-urgent\{background:#c62828\}/);
  assert.doesNotMatch(admin, /urgencyLabel/);
});
test('admin Marcar completada is a small icon button like Editar', () => {
  assert.match(admin, /<button type="button" class="editbtn" data-complete-need="\$\{clean\(n\.id\)\}"><svg[^>]*><path d="M5 12l5 5 9-10"\/><\/svg>Marcar completada<\/button>/);
});
test('admin Nueva in the needs header opens the need page with an empty form and the other needs listed below', () => {
  const editor = admin.slice(admin.indexOf('id="campEditor"'), admin.indexOf('</section>', admin.indexOf('id="campEditor"')));
  assert.match(editor, /<div class="needs-head"><div id="needStats"[^>]*><\/div><button type="button" class="editbtn" id="newNeed"><svg[^>]*><path d="M12 5v14M5 12h14"\/><\/svg>Nueva<\/button><\/div><p id="needNotice"/);
  assert.match(admin, /\$\('#newNeed'\)\.onclick=\(\)=>openEditor\(activeCamp,true,false,'new'\)/);
  assert.match(admin, /#campEditor:not\(\.need-mode\) #needForm\{display:none\}/);
  assert.match(admin, /#campEditor:not\(\.edit-mode\) #campEditDetails\{display:none\}/);
  assert.match(admin, /needId==='new'\?'Nueva necesidad · '/);
  assert.match(admin, /#campEditor\.need-mode #campPublicLink\{display:none\}/);
  assert.match(admin, /\$\('#saveNeedTop'\)\.textContent=needId==='new'\?'Crear necesidad':'Guardar cambios';/);
  assert.match(admin, /await openEditor\(activeCamp\);\$\('#needNotice'\)\.textContent='Necesidad guardada\.'/);
});
test('every page formats timestamps as «Oct 13, 2026 7:33pm» in local time', () => {
  const builder = readFileSync(new URL('./builder.html', import.meta.url), 'utf8');
  const pages = {admin: [admin, 'stamp'], runner: [runner, 'stamp'], builder: [builder, 'stamp'], camp: [camp, 'stamp'], index: [template, 'needStamp']};
  for (const [name, [src, fn]] of Object.entries(pages)) {
    assert.doesNotMatch(src, /toLocaleString/, name);
    const f = new Function(`${src.match(new RegExp(`const ${fn}=[^\\n]*`))[0]};return ${fn}`)();
    assert.equal(f(new Date(2026, 9, 13, 19, 33).toISOString()), 'Oct 13, 2026 7:33pm', name);
    assert.equal(f(new Date(2026, 0, 2, 0, 5).toISOString()), 'Ene 2, 2026 12:05am', name);
    assert.equal(f(new Date(2026, 7, 9, 12, 0).toISOString()), 'Ago 9, 2026 12:00pm', name);
    assert.equal(f(''), '', name);
  }
});
test('admin Checkup button (phone icon) asks for confirmation and the footer reads Último checkup', () => {
  assert.match(admin, /<button type="button" id="markCheckup" class="editbtn"><svg[^>]*><path d="M22 16\.9[^"]*"\/><\/svg>Checkup<\/button>/);
  assert.match(admin, /const b=e\.currentTarget;if\(!activeCamp\|\|!await askConfirm\('¿Lograste comunicarte con alguien en este punto y has actualizado\/corroborado los datos\?'/);
  assert.match(admin, /<dialog id="confirmDialog" aria-labelledby="confirmText"><form method="dialog"><div id="confirmIcon" class="confirm-icon" hidden><\/div><p id="confirmText"><\/p><div id="confirmDetail" hidden><\/div><div class="confirm-actions"><button value="no" class="quiet-red" id="confirmNo">No<\/button><button value="alt" class="confirm-alt" id="confirmAlt" hidden><\/button><button value="yes" id="confirmYes">Sí<\/button><\/div><\/form><\/dialog>/);
  assert.match(admin, /function askConfirm\(text,opts=\{\}\)\{return askChoice\(text,\{\.\.\.opts,alt:''\}\)\.then\(v=>v==='yes'\)\}/);
  const draw = new Function('clean', 'stamp', `${admin.match(/function drawCheckup\(p\)\{[^\n]*/)[0]};return drawCheckup`)(s => String(s), v => v);
  const el = {innerHTML: ''};
  globalThis.$ = () => el;
  draw({lastCheckup: 'T3', checkups: ['T1', 'T3']});
  assert.equal(el.innerHTML, '<summary>Último checkup: T3</summary><ul><li>T3</li><li>T1</li></ul>');
  draw({});
  assert.equal(el.innerHTML, '<summary>Último checkup: sin registrar</summary><p class="muted">Sin checkups registrados.</p>');
  delete globalThis.$;
  assert.match(admin, /\$\('#campCheckup'\)\.open=false;/);
  // Al expandir el registro, Editar y Checkup quedan arriba.
  assert.match(admin, /\.camp-card footer\{[^}]*align-items:flex-start/);
  assert.doesNotMatch(admin, /Marcar revisado|Última revisión|'Revisado: '/);
});
test('count fields have − / + steppers, short labels, and keep the stored field names', () => {
  const who = PointForm.html();
  assert.equal((who.match(/class="pf-numbtn pf-dec"/g) || []).length, 6);
  assert.equal((who.match(/class="pf-numbtn pf-inc"/g) || []).length, 6);
  // El campo va antes de los botones para que el <label> apunte al campo.
  assert.match(who, /<span class="pf-num"><input name="people"[^>]*><button type="button" class="pf-numbtn pf-dec"/);
  assert.match(who, /Mayores<\/span><span class="pf-num"><input name="olderAdults"/);
  assert.match(who, /Discapacitados<\/span><span class="pf-num"><input name="disabled"/);
  assert.doesNotMatch(who, /Adultos mayores|Con discapacidad/);
  assert.match(pointformSrc, /Math\.min\(999, Math\.max\(0,/);
});
test('admin punto view ends with a collapsed Changelog card of Personas and the people list', () => {
  const src = admin.match(/const changelogHtml=(p=>\{[^\n]*\});\n/)[1];
  const clean = s => String(s ?? '').replace(/[&<>"']/g, c => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
  const stamp = v => v;
  const changelogHtml = new Function('clean', 'stamp', `return ${src}`)(clean, stamp);
  assert.match(changelogHtml({}), /^<summary>Changelog<\/summary><p class="muted">Sin cambios registrados/);
  const html = changelogHtml({decided: 'D0', history: [{field: 'people', people: '10', until: 'T1'}, {field: 'members', members: [{name: 'Ana', phone: '6000-1111'}], until: 'T2'}, {field: 'people', people: '12', until: 'T3'}]});
  assert.match(html, /^<summary>Changelog \(3 cambios\)<\/summary><ul>/);
  assert.ok(html.indexOf('12') < html.indexOf('Ana') && html.indexOf('Ana') < html.indexOf('Personas:</b> 10'), 'newest first');
  assert.match(html, /Personas:<\/b> 12<small>Desde T1 hasta T3/);
  assert.match(html, /Personas:<\/b> 10<small>Desde D0 hasta T1/);
  assert.match(html, /Personas del grupo \(1\):<\/b> Ana · 6000-1111<small>Desde D0 hasta T2/);
  // Último bloque de la página del punto, fuera de «Datos del punto», y se abre plegado.
  assert.match(admin, /<\/form><\/details><details id="campChangelog"><\/details><\/section>/);
  assert.doesNotMatch(admin, /historyHtml|Historial de personas/);
  assert.match(admin, /\$\('#campChangelog'\)\.innerHTML=changelogHtml\(p\)/);
  assert.match(admin, /\$\('#campChangelog'\)\.open=false;/);
});
test('admin Guardar punto returns to the updated punto view instead of staying on the edit page', () => {
  const submit = admin.match(/\$\('#campEditForm'\)\.onsubmit=[^\n]*/)[0];
  assert.match(submit, /'PATCH',PointForm\.values\(\$\('#campFields'\)\)\);await load\(\);const id=activeCamp;history\.replaceState\(null,'','\/admin\.html\?camp='\+encodeURIComponent\(id\)\);await openEditor\(id,false\);\$\('#needNotice'\)\.textContent='Punto actualizado\.'/);
  assert.match(submit, /catch\(err\)\{\$\('#campEditMessage'\)\.textContent='No se pudo guardar: '/);
});
test('Agregar punto: GPS opens a full-screen map centered on you, or on the last added punto without GPS', () => {
  assert.doesNotMatch(PointForm.html({admin: true}), /pf-gpsmap/, 'no inline map in the form');
  assert.match(admin, /const addPointNearby=\(\)=>points\.filter\(p=>p\.status==='approved'\|\|p\.status==='pending'\);/);
  assert.equal((admin.match(/nearby:addPointNearby/g) || []).length, 2, 'open and after save');
  assert.equal((template.match(/nearby:publicNearby/g) || []).length, 2, 'public ¡Ayuda! form gets the same map picker (open and after send)');
  assert.match(template, /const publicNearby=\(\)=>approved;/);
  const list = [{code: 'LEJOS', coords: [7.6, -80.3], created: '2026-10-01T00:00:00Z'}, {code: 'SIN-GPS'}, {code: 'CERCA', coords: [7.4501, -80.3501], created: '2026-10-12T00:00:00Z'}];
  assert.deepEqual(PointForm.nearest([7.45, -80.35], list).map(n => n.p.code), ['CERCA', 'LEJOS']);
  assert.equal(PointForm.lastAdded(list).code, 'CERCA');
  assert.equal(PointForm.lastAdded([]), null);
  const m = PointForm.meters([7.45, -80.35], [7.46, -80.35]);
  assert.ok(m > 1100 && m < 1120, `1 km-ish: ${m}`);
  // Se abre al pulsar el botón, antes de pedir el GPS, y es solo el mapa con botón de cerrar.
  assert.match(pointformSrc, /const view = openGpsMap\(root\);\n    if \(!navigator\.geolocation\) \{[^\n]*view\?\.status\(msg\.textContent\); return; \}/);
  assert.match(pointformSrc, /<dialog class="pf-mapdialog"[^`]*><div class="pf-mapbox"><\/div><div class="pf-crosshair" aria-hidden="true">\$\{crosshair\}<\/div><button type="button" class="pf-mapclose"[^`]*<div class="pf-mappanel"><p class="pf-mapstatus" role="status"><\/p><p class="pf-mappick"><\/p><button type="button" class="pf-mapuse">Usar esta ubicación<\/button><\/div><\/dialog>/);
  assert.match(pointformSrc, /d\.showModal\(\);/);
  assert.match(pointformSrc, /if \(nearby\) root\.querySelector\('\.pf-gps'\)\.innerHTML = `\$\{icons\.pin\}Ubicar en mapa`;/);
  assert.match(pointformSrc, /map\.setView\(last \? last\.coords : \[7\.46, -80\.33\], last \? 15 : 11\);/);
  assert.match(pointformSrc, /if \(!moved\) map\.setView\(here, 16\);/, 'GPS recenters only if the map was not moved yet');
  // Cruz: arrastrar o tocar mueve el punto; «Usar esta ubicación» lo pasa al enlace GPS y cierra.
  assert.match(pointformSrc, /map\.on\('click', e => \{ moved = true; map\.panTo\(e\.latlng\); \}\);/);
  assert.match(pointformSrc, /root\.querySelector\('\[name="gps"\]'\)\.value = mapLink\(here\);[^\n]*\n[^\n]*\n      d\.close\(\);/);
  assert.match(pointformSrc, /use\.disabled = !ok;/);
  // Al volver a abrir, el mapa parte de la ubicación ya elegida y el GPS no la reemplaza.
  assert.deepEqual(PointForm.linkCoords('https://maps.google.com/?q=7.450120,-80.350110'), [7.45012, -80.35011]);
  assert.equal(PointForm.linkCoords(''), null);
  assert.match(pointformSrc, /const prev = linkCoords\(root\.querySelector\('\[name="gps"\]'\)\.value\), keep = !!prev && inPanama\(prev\);\n    let moved = keep,/);
  assert.match(pointformSrc, /if \(keep\) map\.setView\(prev, 17\);/);
  // «Mi GPS» vuelve a centrar la cruz en tu GPS (y lo pide si aún no llegó) sin tocar el formulario.
  assert.match(pointformSrc, /<button type="button" class="pf-mapgps" title="Mover la cruz a tu ubicación GPS">\$\{icons\.locate\}Mi GPS<\/button>/);
  const gpsClick = pointformSrc.match(/gpsBtn\.onclick = \(\) => \{[\s\S]*?\n    \};/)[0];
  assert.match(gpsClick, /if \(gpsHere\) \{ map\.setView\(gpsHere, Math\.max\(map\.getZoom\(\), 16\)\); return; \}/);
  assert.match(gpsClick, /navigator\.geolocation\.getCurrentPosition/);
  assert.doesNotMatch(gpsClick, /name="gps"|gpsmsg/, 'Mi GPS does not change the form');
  assert.match(pointformSrc, /if \(view\?\.keep\) \{[^\n]*return; \}\n      gps\.value = mapLink\(\[lat, lng\]\);/);
  assert.match(pointformSrc, /if \(view\) outside \? view\.status\(msg\.textContent\) : view\.me\(\[lat, lng\], pos\.coords\.accuracy\);/);
  assert.match(pointformSrc, /view\?\.status\(msg\.textContent\);/);
  assert.match(pointformSrc, /L\.divIcon\(\{className: '', html: '<div class="pf-me"><\/div>'/);
});
test('admin punto list shows needs as icon chips (not comma text) with room after the colored line', () => {
  assert.match(admin, /<b>\$\{needBadge\(n\)\}\$\{urgentBadge\(n\)\}<\/b>\$\{needChips\(n\)\}\$\{typeof PointForm!=='undefined'\?PointForm\.packingHtml\(n\.packing\):''\}\$\{needTimes\(n\)\}/);
  assert.match(admin, /<b>\$\{label\}<\/b>\$\{k==='needs'\?needChips\(\{need:p\[k\]\}\):clean\(p\[k\]\)\}/, 'pending report needs as chips');
  assert.match(admin, /\.card-needs li\{padding:8px 10px 8px 12px;border:1px solid #dce5e4;border-radius:9px;/, 'each need is its own bordered tile');
  assert.doesNotMatch(admin, /\.card-needs li\{padding-left:9px\}/);
});
test('admin need cards show repair as its icon chip only, and Asignar opens a searchable runner/builder list that saves', () => {
  assert.doesNotMatch(admin, /class="need-kind"/, 'no separate Reparación de casa badge');
  assert.match(admin, /needBadge=n=>`<span class="need-state \$\{needClass\(n\)\}">/);
  assert.match(admin, /<div class="picker" id="assignPicker" hidden><input id="assignSearch" type="search" autocomplete="off"><div id="assignList" role="listbox"><\/div><\/div><script>/);
  assert.ok(admin.indexOf('id="assignPicker"') < admin.indexOf("$('#assignSearch').oninput"), 'markup exists before the script uses it');
  assert.match(admin, /opts=crewOf\(n\)\.filter\(r=>!r\.disabled&&\(!q\|\|fold\(r\.name\+' '\+r\.telephone\)\.includes\(q\)\)\)/);
  assert.match(admin, /adminRequest\('\/api\/admin\/needs\/'\+encodeURIComponent\(n\.id\),'PATCH',n\.kind==='repair'\?\{builderId:id\}:\{runnerId:id\}\)/);
  assert.match(admin, /if\(e\.key==='Enter'\)\{e\.preventDefault\(\);const first=\$\('#assignList \[data-assign\]'\)/);
});
test('admin dashboard: no page title or intro, small sticky stats, fixed map with Mover and Maximizar', () => {
  assert.doesNotMatch(admin, /Panel de Administración|id="pageTitle"|Datos completos del registro/);
  assert.match(admin, /<section id="dashboard" hidden><div class="statswrap">/);
  assert.match(admin, /\.statswrap\{position:sticky;top:0;z-index:1001;/, 'stays above the map panes while scrolling');
  assert.match(admin, /\.adminstats b\{display:block;font-size:17px;/);
  assert.match(admin, /<div class="mapwrap" id="mapWrap"><div id="adminMap"[^>]*><\/div><div class="mapctl"><button type="button" id="mapMove" aria-pressed="false">Mover<\/button><button type="button" id="mapMax" aria-pressed="false">Maximizar<\/button><\/div><\/div>/);
  assert.match(admin, /adminMap=L\.map\('adminMap'\)\.setView\(\[7\.46,-80\.33\],11\);setMapMove\(false\);/, 'map starts fixed');
  assert.match(admin, /const mapHandlers=\['dragging','touchZoom','scrollWheelZoom','doubleClickZoom','boxZoom','keyboard'\];/);
  assert.match(admin, /\.mapwrap\.max\{position:fixed;inset:0;z-index:2000;/);
  assert.match(admin, /if\(e\.key==='Escape'\)setMapMax\(false\)/);
});
test('admin punto card: clicks in the necesidades column do not open the punto', () => {
  assert.match(admin, /if\(article&&!e\.target\.closest\('button,a,input,select,textarea,\.card-needs'\)\)\{/);
});
test('admin Marcar completada asks first: Cancelar, Sí, completar, and the default Completar y crear otra, showing the need icons', () => {
  const handler = admin.match(/\$\('#needList'\)\.onclick=[^\n]*/)[0];
  assert.match(handler, /choice=await askChoice\('¿Marcar esta necesidad como completada\?',\{yes:'Completar y crear otra',alt:'Sí, completar',no:'Cancelar',icon:complete\.querySelector\('svg'\)\.outerHTML,detail:n\?needChips\(n\):'',focusYes:true\}\);if\(choice==='no'\)return;complete\.disabled=true;try\{await adminRequest\('\/api\/admin\/needs\/'\+encodeURIComponent\(complete\.dataset\.completeNeed\),'PATCH',\{status:'COMPLETED'\}\)/, 'asks before saving');
  assert.match(handler, /if\(choice==='yes'\)await openEditor\(activeCamp,true,false,'new'\)/);
  assert.match(admin, /<p id="confirmText"><\/p><div id="confirmDetail" hidden><\/div><div class="confirm-actions">/);
  assert.doesNotMatch(handler, /[^a-zA-Z]confirm\(/);
  // askChoice: 'yes' / 'alt' / 'no' (Esc o cerrar = 'no').
  const d = {returnValue: '', showModal() {}, addEventListener(_, f) { this.close = f; }};
  let focused = null;
  const els = {'#confirmDialog': d, '#confirmText': {}, '#confirmDetail': {}, '#confirmYes': {focus() { focused = 'yes'; }}, '#confirmNo': {}, '#confirmAlt': {}, '#confirmIcon': {}};
  const askChoice = new Function('$', `${admin.match(/function askChoice[^\n]*/)[0]};return askChoice`)(s => els[s]);
  const p = askChoice('¿?', {yes: 'Sí', alt: 'Otra', detail: '<i>chips</i>', focusYes: true});
  assert.equal(els['#confirmAlt'].hidden, false);
  assert.equal(els['#confirmDetail'].innerHTML, '<i>chips</i>');
  assert.equal(focused, 'yes');
  d.returnValue = 'alt'; d.close();
  return p.then(v => {
    assert.equal(v, 'alt');
    const q = askChoice('¿?');
    assert.equal(els['#confirmAlt'].hidden, true, 'two-button prompts hide the third');
    assert.equal(els['#confirmDetail'].hidden, true);
    d.returnValue = ''; d.close();
    return q.then(v2 => assert.equal(v2, 'no'));
  });
});
test('admin user management lists, adds, edits and disables users with roles and an access code', () => {
  const ops = admin.slice(admin.indexOf('id="operations"'));
  assert.match(ops, /<h2 id="usersTitle">Usuarios<\/h2>/);
  assert.match(ops, /id="userList"/);
  assert.match(ops, /id="userForm"/);
  for (const role of ['runner', 'builder', 'agent']) assert.match(ops, new RegExp(`name="roles" value="${role}"`));
  assert.match(ops, /name="accessCode"/);
  assert.doesNotMatch(admin, /id="runnerForm"|id="builderForm"/);
  assert.match(admin, /adminRequest\('\/api\/admin\/users'\)/);
  assert.match(admin, /'\/api\/admin\/users\/'\+encodeURIComponent\(u\.id\),'PATCH',\{disabled:!u\.disabled\}/);
  // Editing sends the access code only when the admin typed a new one.
  assert.match(admin, /if\(code\)data\.accessCode=code/);
  // Disabled users stay out of the assignment picker and need form.
  assert.match(admin, /crewOf\(n\)\.filter\(r=>!r\.disabled&&/);
  assert.match(admin, /filter\(r=>!r\.disabled\|\|r\.id===value\)/);
});
test('Usuarios is its own admin page opened from the header and deep-linkable', () => {
  assert.match(admin, /<button type="button" role="menuitem" id="usersToggle">Usuarios<\/button>/);
  assert.match(admin, /id="usersBack">← Volver al panel<\/button><h2 id="usersTitle">Usuarios<\/h2><button type="button" id="userNew">\+ Nuevo<\/button><button type="button" id="userEdit">Editar<\/button><button type="button" id="userToggle">Desactivar<\/button><button type="submit" form="userForm" id="userSaveTop">Guardar<\/button>/);
  assert.match(admin, /#dashboard\.users-mode>:not\(#operations\)\{display:none!important\}#dashboard:not\(\.users-mode\)>#operations\{display:none!important\}/);
  assert.match(admin, /history\.pushState\(null,'',on\?'\/admin\.html\?view=usuarios'\+\(viewing\?'&user='\+encodeURIComponent\(viewingUser\):''\):'\/admin\.html'\)/, 'returning from a point keeps the open user');
  assert.match(admin, /else if\(q\.get\('view'\)==='usuarios'\)\{showUsers\(true,false\);/, 'reloading the deep link reopens Usuarios');
  assert.match(admin, /showUsers\(q\.get\('view'\)==='usuarios',false\)/, 'back/forward keeps the page in sync');
  // Need cards resolve assignees from every user, so migrated or placeholder users still show a name.
  assert.match(admin, /const r=userById\(assigneeOf\(n\)\);/);
  assert.match(admin, /userById\(n\.releasedBy\)\?\.name/);
});
test('runner and builder sign-in show only name and phone; «!código» in Nombre signs in with a code, no phone', () => {
  for (const [page, who] of [[runner, 'runner'], [readFileSync(new URL('./builder.html', import.meta.url), 'utf8'), 'builder']]) {
    const gate = page.match(/<form id="gateForm"[^]*?<\/form>/)[0];
    assert.match(gate, /<h1>¿Cómo te llamas\?<\/h1>/);
    assert.doesNotMatch(gate, /name="accessCode"|código de acceso|gate-or/i, 'nothing hints at access codes');
    assert.match(gate, /<input name="name" maxlength="120"/);
    assert.match(page, /accessCode=raw\.startsWith\('!'\)\?raw\.slice\(1\)\.trim\(\):'',name=accessCode\?'':raw/);
    assert.match(page, /if\(accessCode\)f\.name\.value='';/, 'the code is cleared from the box');
    assert.match(page, /if\(!accessCode&&!\(name&&telephone\)\)\{\$\('#gateMessage'\)\.textContent='Escribe tu nombre y teléfono\.'/);
    assert.match(page, /await startSession\(accessCode\?\{accessCode\}:\{name,telephone\}\)/);
    assert.match(page, new RegExp(`localStorage\\.removeItem\\('${who}Code'\\)`), 'logout forgets the saved code');
  }
});
test('public pages show priority only as an Urgente badge, never Normal, and nothing about reserved needs', () => {
  const camp = readFileSync(new URL('./camp.html', import.meta.url), 'utf8');
  for (const page of [template, camp]) {
    assert.match(page, /const urgentBadge=u=>u==='urgente'\?'<span class="badge urgent">Urgente<\/span>':'';/);
    assert.match(page, /\.badge\.urgent\{background:#c62828;color:#fff\}/);
    assert.doesNotMatch(page, /urgencyEs|urgencyLabel|restricted|Atención reservada/);
    const badge = new Function(`${page.match(/const urgentBadge=[^\n]*/)[0]};return urgentBadge`)();
    assert.equal(badge('normal'), '');
    assert.equal(badge(''), '');
    assert.equal(badge('urgente'), '<span class="badge urgent">Urgente</span>');
  }
  assert.match(template, /<b>\$\{clean\(needStatusEs\[n\.status\]\|\|n\.status\)\} \$\{urgentBadge\(n\.urgency\)\}<\/b>/);
  assert.match(camp, /\$\{urgentBadge\(p\.highestUrgency\)\}/);
  // A need without text (reserved for an access-code runner) shows only its status: no empty text or time line.
  assert.match(template, /\$\{n\.need\?`\$\{needChips\(n\.need\)\}<small>/);
  assert.match(camp, /\$\{n\.need\?`<p class="bodytext">\$\{clean\(n\.need\)\}<\/p>/);
});
test('public root page shows needs as icon chips, like the runner page', () => {
  const grab = re => template.match(re)[0];
  const line = grab(/const urgentBadge=[^\n]*/);
  const needChips = new Function('PointForm', 'clean', `${line};return needChips`)(PointForm, s => String(s));
  assert.match(needChips('agua, Alimentos, colchones'), /<span class="chip"><svg[^]*?<\/svg>Agua<\/span><span class="chip"><svg[^]*?<\/svg>Alimentos<\/span><span class="chip text"><svg[^]*?<\/svg>colchones<\/span>/);
  assert.match(template, /:r\.needs\?needChips\(r\.needs\):'<div class="need">Sin necesidades registradas<\/div>'\)/);
  assert.match(template, /\$\{r\.needs\?needChips\(r\.needs\):'<div class="value">Sin datos<\/div>'\}/);
  assert.doesNotMatch(template, /clean\(r\.needs/);
});
test('packing list: every starting item has its own icon and matches the server catalog', () => {
  const goSrc = readFileSync(new URL('./packing.go', import.meta.url), 'utf8');
  const goNames = [...goSrc.match(/var defaultPackingItems = \[\]string\{([^}]*)\}/)[1].matchAll(/"([^"]+)"/g)].map(m => m[1]);
  assert.deepEqual(PointForm.PACKING.map(p => p[0]), goNames);
  assert.equal(goNames.length, 19);
  const icons = new Set();
  for (const [name, key] of PointForm.PACKING) {
    assert.ok(PointForm.icons[key], `${name} icon ${key}`);
    assert.notEqual(key, 'box', `${name} needs a specific icon`);
    assert.equal(PointForm.packIcon(name.toUpperCase()), PointForm.icons[key], 'icons match without case');
    icons.add(key);
  }
  assert.equal(icons.size, PointForm.PACKING.length, 'no two starting items share an icon');
  assert.equal(PointForm.packIcon('Agua embotellada'), PointForm.icons.box, 'new items get the generic box');
  const html = PointForm.packingHtml([{item: 'Linternas', quantity: 5}, {item: 'Agua embotellada', quantity: 3}, {item: 'linternas', quantity: 2}]);
  assert.match(html, /Lista de empaque/);
  assert.match(html, /<svg[^]*?<\/svg>Linternas <b>× 7<\/b>/);
  assert.match(html, /Agua embotellada <b>× 3<\/b>/);
  assert.equal(PointForm.packingHtml([]), '');
});
test('packing dialog is shared: type-ahead, quantity, review and save on the runner page and in the admin need form; agents only pack', () => {
  const pf = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
  assert.match(pf, /<dialog id="pfPack"/);
  assert.match(pf, /<input id="pfPackSearch"[^>]*role="combobox"/);
  assert.match(pf, /Agregar «\$\{esc\(raw\)\}» como artículo nuevo/);
  assert.match(pf, /<input id="pfPackQty" type="number" min="1"/);
  assert.match(pf, /data-act="review" disabled>Revisar<\/button><button type="button" class="pfp-btn" data-act="save" hidden>Guardar lista<\/button>/);
  assert.match(pf, /openPacking, prepareMedia, uploadMedia, mediaThumbs, setMe\};/);
  // Runner page: catalog + save through the runner API.
  assert.doesNotMatch(runner, /id="packDialog"/, 'no second copy of the dialog');
  assert.match(runner, /api\('\/api\/runner\/packing-items'\)/);
  assert.match(runner, /api\('\/api\/runner\/needs\/'\+encodeURIComponent\(n\.id\)\+'\/packing','POST',\{items:lines\}\)/);
  assert.match(runner, /\$\{needDetails\(n\)\}<button class="btn take" type="button" data-take="\$\{clean\(n\.id\)\}"\$\{full\?[^}]*\}>¡Yo lo llevo!<\/button>\$\{n\.assigned\?packButton\(n\):''\}<\/article>/, 'open needs get the packing button only when assigned to this runner');
  assert.match(runner, /\$\{needDetails\(n\)\}\$\{mediaStrip\(n\)\}\$\{packButton\(n\)\}/);
  const agent = runner.match(/if\(agentMode\)\{[^\n]*/)[0];
  assert.match(agent, /packButton\(n\)/);
  assert.doesNotMatch(agent, /data-take/);
  assert.match(runner, /agentMode=!!body\.agent/);
  // Admin need form: saved at once when editing, or after creating a new need.
  assert.match(admin, /<button type="button" class="editbtn" id="needPackAdd">\+ Agregar a lista de empaque<\/button>/);
  assert.match(admin, /adminRequest\('\/api\/admin\/packing-items'\)/);
  assert.match(admin, /if\(n\)\{await adminRequest\('\/api\/admin\/needs\/'\+encodeURIComponent\(n\.id\)\+'\/packing','POST',\{items:lines\}\)/);
  assert.match(admin, /if\(supply&&newPacking\.length\)await adminRequest\('\/api\/admin\/needs\/'\+encodeURIComponent\(supply\.id\)\+'\/packing','POST',\{items:newPacking\}\)/);
  assert.match(admin, /\$\('#needPacking'\)\.hidden=needKind\(\)==='repair'/, 'repairs have no packing list');
  assert.match(admin, /PointForm\.packingHtml\(n\.packing\)/);
});
test('runner holds several runs, reorders them by drag or arrows, and sees them numbered on a live map', () => {
  assert.match(runner, /<div id="runList">\$\{mine\.map\(runCard\)\.join\(''\)\}<\/div>/);
  const card = runner.match(/function runCard[^\n]*/)[0];
  assert.match(card, /<button type="button" class="drag" aria-label="Arrastrar para cambiar el orden"/);
  assert.match(card, /data-move="-1"[^>]*>▲<\/button><button type="button" class="move" data-move="1"/);
  assert.match(card, /data-complete="\$\{id\}">✓ Completé esta carrera<\/button><button class="btn quit" type="button" data-release="\$\{id\}">/);
  // Dragging works with mouse and touch (pointer events) and saves the order on the server.
  assert.match(runner, /\.drag\{touch-action:none;/);
  assert.match(runner, /addEventListener\('pointerdown'[^\n]*setPointerCapture/);
  assert.match(runner, /api\('\/api\/runner\/order','POST',\{ids:mine\.map\(n=>n\.id\)\}\)/);
  assert.match(runner, /mine=data\.mine\|\|\[\];max=data\.max\|\|max;/);
  // Map: numbered stops, dashed route, live position that the map follows.
  assert.match(runner, /<div class="runpin\$\{focused\?' sel':''\}">\$\{Math\.min\(\.\.\.stops\)\+1\}<\/div>/);
  assert.match(runner, /L\.polyline\(pts,\{color:'#1f63c6',weight:3,dashArray:'6 8'/);
  assert.match(runner, /<button type="button" class="livebtn" id="liveLoc" aria-pressed="false">📍 Mi ubicación en vivo<\/button>/);
  assert.match(runner, /navigator\.geolocation\.watchPosition\(/);
  assert.match(runner, /icon:L\.divIcon\(\{className:'',html:truckIcon,/, 'the runner shows as a pickup truck');
  assert.match(runner, /const truckIcon='<div class="truckpin"><svg viewBox="0 0 32 20"/);
  assert.doesNotMatch(runner, /mepin/);
  // Street route: truck → stop 1 in orange through our server; later stops stay dashed.
  assert.match(runner, /api\('\/api\/runner\/route\?from='\+me\.join\(\)\+'&to='\+to\)/);
  assert.match(runner, /streetLine=L\.polyline\(street\.coords,\{color:'#ed7654',weight:6/);
  assert.match(runner, /if\(to===streetTo&&\(since<15000\|\|street&&moved<150\)\)return;/, 'route requests are throttled');
  assert.match(runner, /<div class="routeinfo" id="routeInfo" role="status" hidden><\/div>/);
  assert.doesNotMatch(runner, /router\.project-osrm\.org/, 'phones never call the routing service directly');
  assert.match(runner, /navigator\.geolocation\.clearWatch\(watchId\)/);
  assert.match(runner, /if\(!focusCamp&&located\.length&&sig!==mapSig\)/, 'the map only re-fits when stops change');
});
test('after completing a run, the follow-up form uses the icon need picker instead of a text box', () => {
  const form = runner.match(/<form class="followup" id="followupForm">[^`]*/)[0];
  assert.match(form, /<div class="pf" id="followupNeeds">\$\{PointForm\.needsHtml\(\)\}<\/div>/);
  assert.doesNotMatch(form, /<textarea name="need"/);
  assert.match(runner, /PointForm\.fillNeeds\(\$\('#followupNeeds'\),''\);renderMap\(\[followup\]\)/, 'Otros toggle is wired');
  assert.match(runner, /const need=PointForm\.readNeeds\(\$\('#followupNeeds'\)\);if\(!need\)throw Error\('Marca al menos una necesidad o escribe en «Otros»\.'\)/);
  assert.doesNotMatch(PointForm.needsHtml(), /Reparación de casa/, 'runners do not get the admin-only repair option');
});
test('admin point card: needs use the full card width as a grid of tiles, stats stay top right', () => {
  assert.match(admin, /#items article\.with-needs\{display:grid;grid-template-columns:minmax\(0,1fr\) auto;grid-template-areas:'main stats' 'needs needs';/);
  assert.match(admin, /\.with-needs>\.card-needs\{display:contents\}/);
  assert.match(admin, /\.with-needs>\.card-needs>\.needstats\{grid-area:stats;justify-self:end;/);
  assert.match(admin, /\.with-needs \.card-needs ul\{display:grid;grid-template-columns:repeat\(auto-fill,minmax\(250px,1fr\)\);/);
  assert.match(admin, /@media\(max-width:600px\)\{#items article\.with-needs\{grid-template-columns:1fr;grid-template-areas:'main' 'stats' 'needs'\}/);
});
test('admin point card: the Completadas box toggles completed needs for that point', () => {
  const needStatsHtml = new Function('clean', `${admin.match(/const statIcon=[^\n]*/)[0]}\n${admin.match(/function needStatsHtml[^\n]*/)[0]};return needStatsHtml`)(String);
  const list = [{status: 'OPEN'}, {status: 'COMPLETED'}, {status: 'COMPLETED'}];
  assert.match(needStatsHtml(list, {id: 'p1', on: false}), /<button type="button" class="done ico" data-show-done="p1" aria-pressed="false" title="Mostrar completadas"[^>]*><svg[^]*<b>2<\/b><\/button>$/);
  assert.match(needStatsHtml(list, {id: 'p1', on: true}), /aria-pressed="true" title="Ocultar completadas"/);
  assert.doesNotMatch(needStatsHtml(list), /<button/, 'the camp view keeps plain boxes');
  assert.match(admin, /if\(!doneShown\.delete\(id\)\)doneShown\.add\(id\);render\(\)/);
});
test('runner confirms completing or leaving a run in a Sí / No modal, not a browser popup', () => {
  assert.match(runner, /<dialog id="confirmDialog" class="confirmdlg"[^>]*><form method="dialog" class="confirmbox"><h2 id="confirmText"><\/h2><div id="confirmNeed"><\/div><p id="confirmDetail"[^>]*><\/p><div id="confirmMedia"[^]*?<div class="confirmactions"><button class="btn plain" value="no" id="confirmNo">No<\/button><button class="btn" value="yes" id="confirmYes">Sí<\/button>/);
  assert.match(runner, /askConfirm\('¿Estás seguro de que quieres completar esta carrera\?',\{yes:'Sí, completar',need:done,media:mediaOn\}\)/);
  assert.match(runner, /askConfirm\('¿Estás seguro de que no puedes continuar esta carrera\?',\{yes:'Sí, dejarla',danger:true,/);
  assert.match(runner, /resolve\(d\.returnValue==='yes'\)/, 'Esc or No means no');
  assert.doesNotMatch(runner, /[^a-zA-Z]confirm\('¿Dejar esta carrera/);
});
test('runner panel shows priority only as an Urgente badge, never Normal', () => {
  assert.match(runner, /\$\{n\.urgency==='urgente'\?'<span class="badge urgente">Urgente<\/span>':''\}<\/div>/);
  assert.match(runner, /data-need="\$\{clean\(n\.id\)\}">\$\{n\.urgency==='urgente'\?'Urgente: ':''\}/);
  assert.doesNotMatch(runner, /urgencyLabel\[n\.urgency\]/);
});
test('admin point view: each supply need has a Lista de empaque button next to Editar', () => {
  assert.match(admin, /Editar<\/button>\$\{n\.kind!=='repair'&&n\.status!=='DELETED'\?`<button type="button" class="editbtn" data-pack-need="\$\{clean\(n\.id\)\}">\$\{PointForm\.icons\.box\}Agregar a lista de empaque<\/button>`:''\}/);
  assert.match(admin, /<div class="need-actions"><button type="button" class="editbtn" data-edit-need=/);
  assert.match(admin, /\.need-actions \[data-complete-need\]\{margin-left:auto\}/, 'Marcar completada sits apart on the right');
  assert.match(admin, /pack=e\.target\.closest\('\[data-pack-need\]'\);if\(await mediaClick\(e\)\)return;if\(pack\)\{const n=campNeeds\.find\(x=>x\.id===pack\.dataset\.packNeed\);if\(n\)packNeed\(n\);return\}/);
  assert.match(admin, /async function packNeed\(n\)\{[^\n]*adminRequest\('\/api\/admin\/needs\/'\+encodeURIComponent\(n\.id\)\+'\/packing','POST',\{items:lines\}\);await loadNeeds\(\)/);
});
test('runner documents runs: Foto de partida on each run, delivery media when completing, smaller No puedo continuar', () => {
  const card = runner.match(/function runCard[^\n]*/)[0];
  assert.match(card, /\$\{mediaStrip\(n\)\}\$\{packButton\(n\)\}\$\{!mediaOn\?'<small class="nomedia">📷 Fotos no disponibles<\/small>':\(n\.media\|\|\[\]\)\.some\(m=>m\.stage==='pickup'\)\?'':`<button class="btn plain media" type="button" data-media="\$\{id\}">📷 Foto de partida<\/button>`\}/, 'the big button only until the first pickup photo, and only with media storage');
  assert.match(runner, /\$\{st==='pickup'&&mediaOn\?`<button type="button" class="thumbadd" data-media="\$\{clean\(n\.id\)\}"[^>]*>\+<\/button>`:''\}/, 'then a + tile after the last pickup photo');
  assert.ok(card.indexOf('Foto de partida') < card.indexOf('Completé esta carrera'), 'Foto de partida comes before completing');
  // Foto de partida opens a chooser: Android's photo picker has no camera unless the input asks for it.
  assert.match(runner, /<dialog id="pickDialog"[^>]*><form method="dialog" class="confirmbox"><h2 id="pickTitle">Foto de partida<\/h2><div class="mediabtns pickbtns"><button type="button" class="btn plain" data-pick="pickPhoto">📷 Tomar foto/);
  assert.match(runner, /if\(media\)\{mediaTarget=media\.dataset\.media;\$\('#pickDialog'\)\.showModal\(\);return\}/);
  assert.match(runner, /files\.then\(f=>\{if\(target&&f\.length\)uploadRunMedia\(target,'pickup',f\)\}\)/);
  assert.ok(runner.indexOf('id="pickPhoto"') < runner.indexOf('<dialog id="confirmDialog"'), 'file inputs live outside the dialogs');  // Entrega: camera buttons on mobile, several items, thumbnails before completing.
  assert.match(runner, /<input type="file" id="pickPhoto" accept="image\/\*" capture="environment" hidden><input type="file" id="pickVideo" accept="video\/\*" capture="environment" hidden><input type="file" id="pickGallery" accept="image\/\*,video\/\*" multiple hidden>/);
  assert.match(runner, /data-pick="pickPhoto">📷 Tomar foto<\/button><button type="button" class="btn plain" data-pick="pickVideo">🎥 Grabar video<\/button><button type="button" class="btn plain" data-pick="pickGallery">🖼️ Galería<\/button>/);
  assert.match(runner, /files\.then\(f=>\{confirmFiles\.push\(\.\.\.f\);drawConfirmFiles\(\)\}\)/, 'each pick adds to the list');
  assert.match(runner, /URL\.createObjectURL\(f\)/, 'thumbnails preview before upload');
  assert.match(runner, /askConfirm\('¿Estás seguro de que quieres completar esta carrera\?',\{yes:'Sí, completar',need:done,media:mediaOn\}\)/);
  assert.match(runner, /if\(confirmFiles\.length&&!await uploadRunMedia\(done\.id,'delivery',confirmFiles\)\)\{complete\.disabled=false;return\}/, 'delivery media uploads before the run is completed');
  assert.match(runner, /PointForm\.uploadMedia\('\/api\/runner\/needs\/'\+encodeURIComponent\(id\)\+'\/media\?stage='\+stage,file,\{csrf,/);
  assert.match(runner, /\.btn\.quit\{display:block;width:auto;margin:12px auto 0;padding:6px 12px;font-size:12px;/);
});
test('admin curates run media: thumbnails per stage, add, move between partida and entrega, delete', () => {
  assert.match(admin, /\$\{needMedia\(n\)\}\$\{needTimes\(n\)\}/);
  const fn = admin.match(/function needMedia[^\n]*/)[0];
  assert.match(fn, /'\/api\/admin\/media\/'\+encodeURIComponent\(n\.id\)/);
  assert.match(fn, /data-media-move=/);
  assert.match(fn, /data-media-del=/);
  assert.match(fn, /data-media-add=/);
  assert.match(admin, /adminRequest\('\/api\/admin\/needs\/'\+encodeURIComponent\(id\)\+'\/media\/'\+encodeURIComponent\(mid\),'PATCH',\{stage\}\)/);
  assert.match(admin, /adminRequest\('\/api\/admin\/needs\/'\+encodeURIComponent\(id\)\+'\/media\/'\+encodeURIComponent\(mid\),'DELETE',\{\}\)/, 'DELETE sends the CSRF header');
  assert.match(admin, /askConfirm\('¿Eliminar esta foto o video\?'/);
  // Shared helpers: photos shrink before upload; SVG-free thumbnails.
  assert.equal(typeof PointForm.uploadMedia, 'function');
  assert.equal(typeof PointForm.prepareMedia, 'function');
  assert.match(PointForm.mediaThumbs([{id: 'm', type: 'video/mp4'}], () => '/x'), /pf-thumb-icon">▶/);
  assert.match(PointForm.mediaThumbs([{id: 'm', type: 'image/jpeg'}], () => '/x'), /<img src="\/x" alt="" loading="lazy"/);
});
test('public list shows reserved run details to a code-signed-in user, and falls back to the public view', () => {
  assert.match(template, /const r=await dataFetch\('\/points'\);/);
  assert.match(template, /dataFetch\('\/camps\/'\+encodeURIComponent\(id\)\+'\/needs'\)/);
  const fn = template.match(/async function dataFetch[^\n]*/)[0];
  assert.match(fn, /fetch\('\/api\/full'\+path/);
  assert.match(fn, /if\(r\.status===401&&await reloginWithCode\(\)\)r=await get\(\);if\(r\.status!==401\)return r;forgetCodeUser\(\)\}return fetch\('\/api'\+path/);
  assert.match(template, /showCodeUser\(\);refresh\(\);loadRunnersOnMap\(\)\}/, 'signing in reloads the list with details');
  assert.match(template, /'\/api\/full\/logout'/);
});
test('public coordinator filter only appears when signed in with an access code', () => {
  assert.match(template, /<span class="needlabel coordlabel">Filtrar por coordinador<\/span><div class="coordwrap">/);
  assert.match(template, /body:not\(\.signed-in\) \.coordlabel,body:not\(\.signed-in\) \.coordwrap,[^{]*\{display:none!important\}/);
  assert.match(template, /document\.body\.classList\.toggle\('signed-in',!!u\);[^\n]*?if\(!u&&\(selectedCoordinator\|\|teamOnly\)\)\{selectedCoordinator='';/);
});
test('chosen location shows a small fixed map preview instead of «Ubicación elegida» text; Ubicar en mapa still moves it', () => {
  assert.match(PointForm.html(), /<button type="button" class="pf-gps">[^]*?<\/button><div class="pf-gpsthumb" hidden title="Toca para mover la ubicación"/);
  assert.doesNotMatch(pointformSrc, /Ubicación elegida en el mapa/);
  const fn = pointformSrc.match(/function drawGpsThumb[^]*?\n  \}/)[0];
  assert.match(fn, /dragging: false, scrollWheelZoom: false, doubleClickZoom: false/, 'the preview is fixed');
  assert.match(fn, /L\.marker\(c, \{icon: L\.divIcon\(\{className: '', html: '<div class="pf-thumbpin"><\/div>'/);
  assert.match(pointformSrc, /root\.querySelector\('\.pf-gpsthumb'\)\.onclick = \(\) => locate\(root\);/, 'tapping the preview reopens the map');
  assert.match(pointformSrc, /root\.querySelector\('\[name="gps"\]'\)\.addEventListener\('input', \(\) => drawGpsThumb\(root\)\);/, 'pasted links preview too');
  assert.match(pointformSrc, /d\.close\(\);\n      drawGpsThumb\(root\);/);
});
test('signed-in users get «Yo» and ✕ next to Coordinador to fill or clear their name and phone', () => {
  const fn = pointformSrc.match(/function setMe\(root, me\) \{[^]*?\n  \}/)[0];
  assert.match(fn, /if \(!me \|\| !me\.name\) return;/, 'only when the user is known');
  assert.match(fn, /<button type="button" class="pf-yo" title="Usar mi nombre y teléfono">Yo<\/button><button type="button" class="pf-clearcoord"/);
  assert.match(fn, /set\(me\.name, me\.telephone \|\| ''\)/);
  assert.match(fn, /set\('', ''\)/);
  assert.match(pointformSrc, /setMe\(root, me\);/);
  assert.equal((template.match(/me:codeMe\(\)/g) || []).length, 2);
  assert.match(template, /\$\('#addBtn'\)\.onclick=\(\)=>\{showSent\(false\);PointForm\.setMe\(\$\('#pointFields'\),codeMe\(\)\);/);
  assert.match(template, /telephone:body\.telephone,roles:body\.roles,panels:body\.panels/);
  assert.match(runner, /localStorage\.setItem\('codeUser',JSON\.stringify\(\{id:body\.id,name:body\.name,telephone:body\.telephone\|\|'',roles:body\.roles\|\|\[\],panels:\[\{label:'Corredores',url:'\/runner\.html'\}\]\}\)\)/, 'signing in with a code on the runner page counts too');
});
test('a point added by a code-signed-in user is published at once; otherwise it goes to review', () => {
  const fn = template.match(/async function submitPoint[^\n]*/)[0];
  assert.match(fn, /send\('\/api\/full\/points'\)/);
  assert.match(fn, /if\(res\.status===401&&await reloginWithCode\(\)\)res=await send\('\/api\/full\/points'\);if\(res\.status!==401\)return res;forgetCodeUser\(\)\}return send\('\/api\/submissions'\)/);
  assert.match(template, /saved\.status==='approved'\?'Punto publicado\. Ya aparece en el mapa y en la lista\.':'Se ha enviado tu solicitud\.'/);
});
test('Usuarios: 🔑 left of the name copies the access code and shows it in a tooltip for 5 seconds', () => {
  assert.match(admin, /role="img" aria-label="\$\{u\.online\?'En línea':'Desconectado'\}"><\/span>\$\{keyBtn\(u\)\}<b>\$\{clean\(u\.name\)\}<\/b>/);
  const src = admin.match(/const keyBtn=[^\n]*/)[0];
  const keyBtn = new Function('clean', `${src};return keyBtn`)(String);
  assert.match(keyBtn({id: 'u1', name: 'Ana', hasAccessCode: true}), /<span class="keytip"><button type="button" class="keybtn" data-copy-code="u1" title="Copiar código de acceso"[^>]*>🔑<\/button><span class="tip" role="status" hidden><\/span><\/span>/);
  assert.equal(keyBtn({id: 'u2', name: 'Beto'}), '<span class="keyslot"></span>', 'no key without a code; names stay aligned');
  assert.match(admin, /adminRequest\('\/api\/admin\/users\/'\+encodeURIComponent\(id\)\+'\/access-code'\)/, 'fetched only when copying');
  assert.match(admin, /navigator\.clipboard\.writeText\(c\)/);
  assert.match(admin, /'Copiado: '\+c:c\+' \(cópialo a mano\)'/, 'the code is still shown if the clipboard is unavailable');
  assert.match(admin, /keyTips\[key\]=setTimeout\(\(\)=>\{tip\.hidden=true;delete keyTips\[key\]\},5000\)/);
  assert.doesNotMatch(admin, /data-reveal|eyebtn|••••••/, 'no more 👁 masked code');
});
test('Usuarios: tapping the phone copies it on a computer and opens WhatsApp on a phone', () => {
  const ops = admin.slice(admin.indexOf('// Usuarios: corredores'));
  assert.doesNotMatch(ops.slice(0, ops.indexOf('// Usuarios es su propia página')), /href="tel:/, 'the list phone is not a link');
  assert.match(admin, /\$\{field\('Teléfono',phoneBtn\(u\)\)\}/, 'nor in the view');
  const src = ['waNumber', 'phoneBtn'].map(n => admin.match(new RegExp(`const ${n}=[^\\n]*`))[0]).join('\n');
  const {waNumber, phoneBtn} = new Function('clean', 'mobile', `${src};return {waNumber,phoneBtn}`)(String, false);
  assert.equal(waNumber('6999-1111'), '50769991111', 'local numbers get Panamá 507');
  assert.equal(waNumber('+1 (305) 555-0100'), '13055550100');
  assert.equal(waNumber('507 6999 1111'), '50769991111');
  assert.match(phoneBtn({id: 'u1', telephone: '6999-1111'}), /<span class="keytip"><button type="button" class="phonebtn" data-phone="u1" title="Copiar teléfono">6999-1111<\/button><span class="tip" role="status" hidden><\/span><\/span>/);
  assert.equal(phoneBtn({id: 'u2', telephone: ''}), '<small>Sin teléfono</small>');
  const fn = admin.match(/function phoneClick\(b\)[^\n]*/)[0];
  assert.match(fn, /if\(mobile\)\{window\.open\('https:\/\/wa\.me\/'\+waNumber\(u\.telephone\),'_blank','noopener'\);return\}/);
  assert.match(fn, /navigator\.clipboard\.writeText\(u\.telephone\)/);
  assert.match(fn, /showTip\(b,'tel:'\+u\.id,'Copiado: '\+u\.telephone\)/);
  assert.match(admin, /const phone=e\.target\.closest\('\[data-phone\]'\);if\(phone\)\{phoneClick\(phone\);return\}/);
});
test('admin header: Usuarios and Generar boletín live in a ☰ menu that opens and closes', () => {
  assert.match(admin, /<button type="button" class="menubtn" id="menuToggle" aria-expanded="false" aria-controls="adminMenu"[^>]*>☰<\/button><div class="adminmenu" id="adminMenu" role="menu" hidden><button type="button" role="menuitem" id="usersToggle">Usuarios<\/button><button type="button" role="menuitem" id="boletinToggle">Generar boletín<\/button><a role="menuitem" class="boletinlink" href="\/report\/boletin_general\.html"[^>]*>Ver último boletín ↗<\/a><button type="button" role="menuitemcheckbox" id="reviewToggle"[^>]*>[^\n]*?<\/button><button type="button" role="menuitem" id="passwordToggle">Cambiar contraseña maestra<\/button><a role="menuitem" id="backupDownload"[^\n]*?<\/div>/);
  assert.match(admin, /\$\('#menuToggle'\)\.onclick=e=>\{e\.stopPropagation\(\);setMenu\(\$\('#adminMenu'\)\.hidden\)\};/);
  assert.match(admin, /\$\('#adminMenu'\)\.addEventListener\('click',\(\)=>setMenu\(false\)\);/, 'choosing an option closes it');
  assert.match(admin, /if\(!e\.target\.closest\('\.menuwrap'\)\)setMenu\(false\)/, 'clicking outside closes it');
});
test('admin shows «Agregado por» for points published by access-code users', () => {
  assert.match(admin, /p\.addedByName&&'Agregado por '\+p\.addedByName\]\.filter\(Boolean\)\.join\(' · '\)/);
  const src = ['clean', 'labels', 'contactKeys', 'memberList', 'peopleOf', 'peopleLine', 'campViewHtml'].map(k => admin.match(new RegExp(`const ${k}=[^\\n]*`))[0]).join('\n');
  const {campViewHtml} = new Function(`${src};return {campViewHtml}`)();
  assert.match(campViewHtml({code: 'C', area: 'A', addedByName: 'Fufo <b>'}), /<div class="field"><b>Agregado por<\/b>Fufo &lt;b&gt;<\/div>/);
  assert.doesNotMatch(campViewHtml({code: 'C', area: 'A'}), /Agregado por/);
});
test('Usuarios can filter to users with an access code only', () => {
  assert.match(admin, /<label class="check"><input type="checkbox" id="userOnlyCode"> Solo con código<\/label>/);
  assert.match(admin, /rows=users\.filter\(u=>\(showOff\|\|!u\.disabled\)&&\(!onlyCode\|\|u\.hasAccessCode\)&&/);
  assert.match(admin, /\$\('#userOnlyCode'\)\.onchange=drawUsers;/);
});
test('public header shows «Mis carreras» only to a signed-in user with the runner role', () => {
  assert.match(template, /<a class="navbtn" id="runnerNav" href="\/runner\.html" hidden>🚚 Mis carreras<\/a>/);
  assert.match(template, /\$\('#runnerNav'\)\.hidden=\$\('#runnerNavBar'\)\.hidden=!\(u&&\(\(u\.roles\|\|\[\]\)\.includes\('runner'\)/);
  assert.match(template, /roles:body\.roles,panels:body\.panels/);
  assert.match(runner, /roles:body\.roles\|\|\[\],panels:\[\{label:'Corredores'/);
});
test('public header hides the «Registro ciudadano · corte…» note when signed in', () => {
  assert.match(template, /<span class="source">Registro ciudadano/);
  assert.match(template, /body\.signed-in \.topright \.source\{display:none\}/);
});
test('runner header: ✕ to sign out (no «Cambiar nombre») and «Mapa público» link for code users', () => {
  assert.match(runner, /<button type="button" id="logout" title="Salir" aria-label="Salir">✕<\/button>/);
  assert.doesNotMatch(runner, /Cambiar nombre/);
  assert.match(runner, /<a class="publink" id="publicNav" href="\/" hidden>← Mapa público<\/a>/);
  assert.match(runner, /function syncPublicNav\(\)\{\$\('#publicNav'\)\.hidden=!localStorage\.getItem\('runnerCode'\);/);
  assert.equal((runner.match(/syncPublicNav\(\)/g) || []).length, 4, 'defined, after sign-in or session check, and after sign-out');
});
test('Usuarios can sort by name or by creation date, newest or oldest first', () => {
  assert.match(admin, /<select id="userSort" aria-label="Ordenar usuarios"><option value="name">Nombre \(A–Z\)<\/option><option value="newest">Más recientes primero<\/option><option value="oldest">Más antiguos primero<\/option><\/select>/);
  const sortSrc = admin.match(/if\(sort!=='name'\)rows\.sort\([^;]*;return sort==='newest'\?-d:d\}\);/)[0];
  const order = (sort, list) => { const rows = [...list]; new Function('rows', 'sort', sortSrc)(rows, sort); return rows.map(u => u.name); };
  const list = [{name: 'B', createdAt: '2026-10-02T00:00:00Z'}, {name: 'Old'}, {name: 'A', createdAt: '2026-10-05T00:00:00Z'}];
  assert.deepEqual(order('newest', list), ['A', 'B', 'Old'], 'users without a date (migrated) sort as oldest');
  assert.deepEqual(order('oldest', list), ['Old', 'B', 'A']);
  assert.match(admin, /data-label=\"Creado\">\$\{u\.createdAt\?`<small class=\"user-created\" title=\"Creado\">\$\{clean\(stamp\(u\.createdAt\)\)\}<\/small>`/);
});
test('Usuarios: «+ Nuevo» top right opens the user form as its own page (like camp needs), no intro text', () => {
  const ops = admin.slice(admin.indexOf('id="operations"'));
  assert.doesNotMatch(ops.slice(0, ops.indexOf('id="userForm"')), /Corredores y constructores también pueden registrarse|runner\.html ↗/, 'no intro text or links');
  assert.match(admin, /#operations\.user-form-mode \.userbar,#operations\.user-form-mode #userList,[^{]*#operations:not\(\.user-form-mode\) #userForm,#operations:not\(\.user-form-mode\) #userSaveTop\{display:none!important\}/);
  assert.match(admin, /\$\('#userNew'\)\.onclick=\(\)=>openUserForm\('new'\);/);
  assert.match(admin, /history\.pushState\(null,'','\/admin\.html\?view=usuarios&user='\+encodeURIComponent\(id\)\+\(u\?'&edit=1':''\)\)/);
  assert.match(admin, /const userFromUrl=q=>\{const id=q\.get\('user'\);if\(id==='new'\|\|q\.get\('edit'\)==='1'\)openUserForm\(id,false\);else openUserView\(id,false\)\};/);
  assert.match(admin, /if\(q\.get\('user'\)\)userFromUrl\(q\)\}\};/, 'reloading the deep link reopens the view or the form');
  assert.match(admin, /if\(q\.get\('view'\)==='usuarios'&&q\.get\('user'\)\)userFromUrl\(q\);else closeUserPage\(false\)/, 'back/forward too');
  assert.match(admin, /\$\('#usersBack'\)\.onclick=\(\)=>userFormOpen\(\)\?cancelUserForm\(\):userViewOpen\(\)\?closeUserPage\(\):showUsers\(false\);/);
  assert.match(admin, /function cancelUserForm\(\)\{const id=editingUser;if\(id\)openUserView\(id\);else closeUserPage\(\)\}/, 'Cancelar goes back to the user view, or to the list for a new one');
  assert.match(admin, /await loadRunners\(\);render\(\);openUserView\(saved\.id\);\$\('#usersNotice'\)\.textContent='Usuario guardado\.'/, 'saving shows the saved user');
});
test('Usuarios list: «Abrir» opens the user in view mode with Editar and Activar/Desactivar', () => {
  assert.match(admin, /<button type="button" class="uc-open" data-user-open="\$\{clean\(u\.id\)\}">Abrir<\/button>/);
  assert.doesNotMatch(admin, /data-user-edit|data-user-toggle/, 'edit and disable live in the view, not the list');
  assert.match(admin, /const open=e\.target\.closest\('\[data-user-open\]'\);if\(open\)openUserView\(open\.dataset\.userOpen\)/);
  assert.match(admin, /history\.pushState\(null,'','\/admin\.html\?view=usuarios&user='\+encodeURIComponent\(id\)\)\}/, 'deep link to the view');
  assert.match(admin, /#operations\.user-view-mode \.userbar,#operations\.user-view-mode #userList,#operations\.user-view-mode #userNew,#operations:not\(\.user-view-mode\) #userView,#operations:not\(\.user-view-mode\) #userEdit,#operations:not\(\.user-view-mode\) #userToggle\{display:none!important\}/);
  assert.match(admin, /\$\('#userEdit'\)\.onclick=\(\)=>openUserForm\(viewingUser\);/);
  assert.match(admin, /t\.textContent=u\.disabled\?'Activar':'Desactivar';t\.classList\.toggle\('quiet-red',!u\.disabled\)/);
  assert.match(admin, /\$\('#userToggle'\)\.onclick=async e=>\{[^\n]*askConfirm\(`¿Desactivar a \$\{u\.name\}\?`[^\n]*'PATCH',\{disabled:!u\.disabled\}\);await loadRunners\(\)/);
  assert.match(admin, /drawUsers\(\);if\(userViewOpen\(\)\)drawUserView\(\);runnerOptions\(\)\}/, 'the view refreshes after a change');
  assert.match(admin, /\$\('#userView'\)\.onclick=\$\('#userList'\)\.onclick;/, '🔑 and en ruta runs work in the view too');
});
test('without media storage, upload buttons are hidden and a short notice shows (runner and admin)', () => {
  assert.match(runner, /fetch\('\/api\/media\/status',\{cache:'no-store'\}\)\.then\(r=>r\.json\(\)\)\.then\(s=>\{mediaOn=!!s\.enabled;/);
  assert.match(admin, /fetch\('\/api\/media\/status',\{cache:'no-store'\}\)\.then\(r=>r\.json\(\)\)\.then\(s=>\{mediaOn=!!s\.enabled;/);
  assert.match(admin, /\$\{mediaOn\?`<button type="button" class="thumbadd" data-media-add=/);
  assert.match(admin, /\$\{mediaOn\?'':'<small class="nomedia">📷 Fotos no disponibles<\/small>'\}/);
});
test('runner footer shows the user ID with a copy button; Usuarios shows and searches the same ID', () => {
  assert.match(runner, /<footer class="sidefoot" id="idFoot" hidden>ID de usuario: <code id="userIdShort" title=""><\/code><button type="button" id="copyId"[^>]*>📋<\/button>/);
  assert.match(runner, /showUserId\(body\.id\)/);
  assert.match(runner, /showUserId\(s\.id\)/);
  assert.match(runner, /navigator\.clipboard\.writeText\(userId\)/);
  assert.match(admin, /<small class="user-id" title="ID de usuario: \$\{clean\(u\.id\)\}">ID \$\{clean\(u\.id\.slice\(0,8\)\)\}<\/small>/);
  assert.match(admin, /fold\(u\.name\+' '\+u\.telephone\+' '\+u\.id\)\.includes\(q\)/, 'search by ID');
});
test('Usuarios shows a green/grey dot before each name and the last active time', () => {
  assert.match(admin, /<span class="presence\$\{u\.online\?' on':''\}" title="\$\{u\.online\?'En línea':'Desconectado'\}"[^>]*><\/span>\$\{keyBtn\(u\)\}<b>\$\{clean\(u\.name\)\}<\/b>/);
  assert.match(admin, /\$\{u\.online\?'En línea':u\.lastActiveAt\?'Activo '\+clean\(stamp\(u\.lastActiveAt\)\):'Sin actividad'\}/);
  assert.match(admin, /\.presence\{[^}]*background:#b5c0c3[^}]*\}\.presence\.on\{background:#23a35a/);
  assert.match(admin, /setInterval\(\(\)=>\{if\(\$\('#dashboard'\)\.classList\.contains\('users-mode'\)&&!userFormOpen\(\)&&!renamingTeam&&!\$\('#teamsBox'\)\.contains\(document\.activeElement\)&&!\$\('#userBatch'\)\.contains\(document\.activeElement\)&&!Object\.keys\(keyTips\)\.length\)Promise\.all\(\[loadRunners\(\),adminRequest\('\/api\/admin\/needs'\)/, 'refreshes users and their runs');
});
test('Usuarios lists each runner\'s en ruta runs (point and needs), with a count, and opens the point on click', () => {
  const src = [admin.match(/const userRuns=[^\n]*/)[0], admin.match(/const userRunsHtml=[^\n]*/)[0]].join('\n');
  const allNeeds = [
    {id: 'n1', campId: 'c1', runnerId: 'r1', status: 'IN_PROGRESS', need: 'Agua', enRouteAt: '2026-10-14T10:00:00Z'},
    {id: 'n2', campId: 'c1', runnerId: 'r1', status: 'OPEN', need: 'Comida'},
    {id: 'n3', campId: 'c2', runnerId: 'r1', status: 'IN_PROGRESS', need: 'Techo', kind: 'repair'},
    {id: 'n4', campId: 'c2', runnerId: 'r2', status: 'IN_PROGRESS', need: 'Pan'},
  ];
  const points = [{id: 'c1', code: 'C-1', area: 'Venao'}, {id: 'c2', code: 'C-2', area: 'Cañas'}];
  const {userRuns, userRunsHtml} = new Function('allNeeds', 'points', 'clean', 'stamp', 'needChips', `${src};return {userRuns,userRunsHtml}`)(allNeeds, points, String, () => 'Oct 14', n => `<i>${n.need}</i>`);
  const runs = userRuns({id: 'r1', roles: ['runner']});
  assert.deepEqual(runs.map(n => n.id), ['n1'], 'only their en ruta supply runs');
  assert.deepEqual(userRuns({id: 'r1', roles: ['builder']}), [], 'only for runners');
  const html = userRunsHtml(runs);
  assert.match(html, /<button type="button" class="user-run" data-open-camp="c1" title="Abrir el punto">📍 <b>C-1<\/b> · Venao <small>· en ruta desde Oct 14<\/small><i>Agua<\/i><\/button>/);
  assert.equal(userRunsHtml([]), '');
  assert.match(admin, /<span class="enruta-box" title="Carreras en ruta"><b>\$\{runs\.length\}<\/b> en ruta<\/span>/);
  assert.match(admin, /const run=e\.target\.closest\('\[data-open-camp\]'\);if\(run\)\{campReturn='usuarios';openEditor\(run\.dataset\.openCamp\)/);
  assert.match(admin, /\$\{online\} en línea · 🚚 \$\{enRoute\}/);
});
test('a point opened from Usuarios goes back to Usuarios («← Usuarios»), otherwise to the panel', () => {
  assert.match(admin, /if\(run\)\{campReturn='usuarios';openEditor\(run\.dataset\.openCamp\)/);
  assert.match(admin, /\$\('#closeCampEditor'\)\.textContent=editing\?'Cancelar':campReturn==='usuarios'\?'← Usuarios':'← Volver';/);
  assert.match(admin, /else if\(campReturn==='usuarios'\)\{closeEditor\(false\);showUsers\(true\)\}else closeEditor\(\)/);
  assert.match(admin, /function closeEditor\(navigate=true\)\{campReturn='';/, 'leaving the point forgets where it came from');
  assert.ok(admin.indexOf("let campReturn=''") < admin.indexOf('async function openEditor('), 'declared before openEditor');
});
test('runner page asks the runner to turn on live location after signing in («Ahora no» lasts until the tab closes)', () => {
  const fn = runner.match(/async function offerLive\(\)[^\n]*/)[0];
  assert.match(fn, /if\(!map\|\|agentMode\|\|!csrf\|\|watchId!==null\|\|liveOffered\|\|sessionStorage\.getItem\('runnerLiveAsked'\)\)return;liveOffered=true;/, 'not for agents, not when already live, once per load');
  assert.match(fn, /askConfirm\('¿Activar tu ubicación en vivo\?',\{yes:'📍 Activar',no:'Ahora no',/);
  assert.match(fn, /\)\)\{if\(watchId===null\)startLive\(\)\}else sessionStorage\.setItem\('runnerLiveAsked','1'\)\}$/);
  assert.match(fn, /p\.state==='denied'/, 'explains a browser block');
  assert.match(runner, /syncPublicNav\(\);await refresh\(\);loadTeam\(\);return offerLive\(\)\}if\(await relogin\(\)\)\{await refresh\(\);loadTeam\(\);return offerLive\(\)\}showGate\(\)/, 'on page load with a session');
  assert.match(runner, /await startSession\(accessCode\?\{accessCode\}:\{name,telephone\}\);await refresh\(\);loadTeam\(\);offerLive\(\)\}/, 'after signing in at the gate');
  assert.match(runner, /sessionStorage\.removeItem\('runnerLiveAsked'\);liveOffered=false;/, 'the next person on the device is asked too');
});
test('runner page: code users see the other live runners as orange trucks, like the public map', () => {
  const fn = runner.match(/async function loadTeam\(\)[^\n]*/)[0];
  assert.match(fn, /if\(csrf&&localStorage\.getItem\('runnerCode'\)\)\{try\{const r=await fetch\('\/api\/runner\/team'/, 'only for code sign-ins; otherwise the trucks are removed');
  assert.match(fn, /html:runnerTruck,/);
  assert.match(fn, /for\(const \[id,pin\] of teamPins\)if\(!seen\.has\(id\)\)\{map\.removeLayer\(pin\);teamPins\.delete\(id\)\}/);
  assert.equal(runner.match(/const runnerTruck='[^']*'/)[0], template.match(/const runnerTruck='[^']*'/)[0], 'same truck as the public map');
  assert.match(runner, /setInterval\(loadTeam,15000\);/);
  assert.match(runner, /await refresh\(\);loadTeam\(\);return offerLive\(\)\}if\(await relogin\(\)\)\{await refresh\(\);loadTeam\(\);return offerLive\(\)\}/);
  assert.match(runner, /render\(\);loadTeam\(\);\$\('#gateForm'\)\.reset\(\);showGate\(\)\}/, 'signing out clears them');
});
test('runner live location is shared with the team; code users see other runners as small orange trucks on the public map', () => {
  assert.match(runner, /function shareLocation\(p\)\{lastFix=p;if\(agentMode\|\|!csrf\)return;[^\n]*Date\.now\(\)-sharedAt<15000&&\(!map\|\|map\.distance\(sharedPos,here\)<50\)\)return;[^\n]*api\('\/api\/runner\/location','POST'/, 'every 15 s or after 50 m');
  assert.match(runner, /api\('\/api\/runner\/location\/stop','POST',\{\}\)/, 'turning it off stops sharing');
  assert.match(runner, /'📍 En vivo · compartida con el equipo'/, 'the runner is told it is shared');
  assert.match(template, /fetch\('\/api\/full\/runners',\{cache:'no-store',credentials:'same-origin'\}\)/);
  assert.match(template, /const runnerTruck='<div class="runnertruck"><svg viewBox="0 0 32 20" width="28" height="18"[^']*fill="#ed7654"/, 'smaller orange truck, distinct from your blue pin');
  assert.match(template, /setInterval\(loadRunnersOnMap,15000\);/);
  assert.match(template, /if\(localStorage\.getItem\('codeUser'\)\)\{try\{const r=await fetch\('\/api\/full\/runners'/, 'only for code users; otherwise the trucks are removed');
});
test('public map: code users turn location sharing on and off with a map button; it never asks on its own', () => {
  assert.match(template, /L\.DomUtil\.create\('button','sharebtn'\);b\.id='shareBtn';b\.type='button';b\.hidden=true;[^\n]*L\.DomEvent\.on\(b,'click',\(\)=>liveShare\?stopShare\(\):startShare\(\)\)/);
  assert.match(template, /b\.hidden=!localStorage\.getItem\('codeUser'\);b\.classList\.toggle\('on',liveShare\);b\.textContent=liveShare\?'📍 Compartiendo · Detener':'📍 Compartir mi ubicación'/, 'only for code users, with a clear on/off label');
  assert.match(template, /fullPost\('\/api\/full\/location',\{lat:here\[0\],lon:here\[1\]/);
  assert.match(template, /fullPost\('\/api\/full\/location\/stop',\{\}\)/);
  assert.match(template, /Date\.now\(\)-sharedAt<15000&&map\.distance\(sharedPos,here\)<50\)return/, 'every 15 s or 50 m');
  assert.match(template, /lastFix=p;shareFromMap\(p\);/, 'the blue pin feeds sharing');
  assert.match(template, /setInterval\(\(\)=>\{if\(liveShare&&lastFix\)shareFromMap\(lastFix\)\},20000\)/, 'a phone standing still keeps re-sending its last position');
  assert.match(runner, /setInterval\(\(\)=>\{if\(watchId!==null&&lastFix\)shareLocation\(lastFix\)\},20000\)/, 'a runner standing still keeps re-sending too');
  assert.match(template, /'Compartida con quienes entraron con código\.':'Solo visible en este dispositivo\.'/);
  assert.match(template, /document\.body\.classList\.toggle\('signed-in',!!u\);if\(!u\)\{stopShare\(false\);[^}]*\}setShareButton\(\);/, 'signing out stops sharing');
  assert.match(template, /p\.state==='granted'\?startShare\(\):localStorage\.removeItem\('mapLive'\)/, 'a reload resumes only without a permission prompt');
  const calls = template.match(/startShare\(\)/g).length;
  assert.equal(calls, 3, 'started only by the button, the reload resume, or its own definition');
  assert.doesNotMatch(template, /getCurrentPosition/, 'no automatic location request');
});
test('share a point link: public ?punto= opens its detail, runner ?punto= focuses it after sign-in', () => {
  assert.match(template, /<button type="button" class="footbtn" id="shareCamp" title="Compartir enlace del punto">🔗 Compartir<\/button>/);
  assert.match(template, /const campLink=r=>location\.origin\+'\/\?punto='\+encodeURIComponent\(r\.id\|\|r\.code\);/);
  assert.match(template, /renderMarkers\(\);renderFeed\(\);openLinkedCamp\(\)\}/, 'opened after the data loads');
  assert.match(template, /const want=linkedCamp,r=all\(\)\.find\(x=>x\.id===want\|\|x\.code===want\);linkedCamp=null;[^\n]*if\(r\)show\(r\);else \$\('#gpsMessage'\)\.textContent='Ese punto no está publicado o ya no existe\.'/, 'unpublished points are not shown');
  assert.match(runner, /data-share-camp="\$\{clean\(c\.id\)\}"[^>]*>🔗 Compartir<\/button>/);
  assert.match(runner, /shareLink\(location\.origin\+'\/runner\.html\?punto='\+encodeURIComponent\(b\.dataset\.shareCamp\)/);
  assert.match(runner, /available=data\.available\|\|\[\];render\(\);openLinkedCamp\(\)\}/, 'applied after sign-in loads the needs');
  assert.match(runner, /if\(hasCoords\(n\)\)\{focusCamp=want;pendingFocus=true;[^\n]*else goToNeed\(n\.id\)\}/);
  for (const page of [template, runner]) {
    assert.match(page, /async function shareLink\(url,title,btn\)\{if\(mobileDevice&&navigator\.share\)\{try\{await navigator\.share\(\{title,url\}\)\}catch\{\}return\}[^\n]*navigator\.clipboard\.writeText\(url\);btn\.textContent='✓ Enlace copiado'/, 'share sheet on a phone, copy on a computer');
  }
});
test('narrow screens: ¡Ayuda! / + Agregar punto and Mis carreras get their own full-width row under the header', () => {
  assert.match(template, /<\/header>\n?<div class="addbar"><a class="navbtn" id="runnerNavBar" href="\/runner\.html" hidden>🚚 Mis carreras<\/a><button type="button" class="add" id="addBtnBar">¡Ayuda!<\/button><\/div><nav class="mobile-nav"/);
  assert.match(template, /\.addbar\{display:none\}/, 'desktop keeps it in the header');
  assert.match(template, /@media\(max-width:760px\)\{body\{overflow:hidden\}\.addbar\{display:flex;gap:8px;height:54px;[^}]*\}\.addbar \.add\{flex:1;[^}]*\}[^@]*\.topright #addBtn,\.topright #runnerNav\{display:none\}/);
  assert.match(template, /\.app\{height:calc\(100dvh - 166px\);/, 'header 62 + add row 54 + tabs 50');
  assert.match(template, /\$\('#runnerNav'\)\.hidden=\$\('#runnerNavBar'\)\.hidden=/, 'Mis carreras shows for the same runners');
  assert.match(template, /\$\('#addBtnBar'\)\.onclick=\(\)=>\$\('#addBtn'\)\.click\(\);/, 'same action as the header button');
});
test('public header: a code sign-in puts a gold glowing ring around the logo', () => {
  assert.match(runner, /body\.code-user \.logo\{box-shadow:0 0 0 2px #10263a,0 0 0 4px #ffd166,0 0 16px 5px #ffd16680\}/, 'same ring on the runner page');
  assert.match(runner, /document\.body\.classList\.toggle\('code-user',!!localStorage\.getItem\('runnerCode'\)\)/);
  assert.match(template, /body\.signed-in \.logo\{box-shadow:0 0 0 2px #10263a,0 0 0 4px #ffd166,0 0 16px 5px #ffd16680\}/);
  assert.match(template, /document\.body\.classList\.toggle\('signed-in',!!u\)/, 'only while signed in with a code');
});
test('signed-in headers show just the name, without «Hola,»', () => {
  const builder = readFileSync(new URL('./builder.html', import.meta.url), 'utf8');
  for (const page of [template, runner, builder]) assert.doesNotMatch(page, /Hola, /);
});
test('admin list: filter by Agregado por, sort by recent/código/área/coordinador, and Archivar on approved cards', () => {
  assert.match(admin, /<div class="listtools"><label>Agregado por<select id="addedByFilter"><\/select><\/label><label>Ordenar<select id="sortBy"><option value="recent">Más recientes<\/option><option value="code">Código<\/option><option value="area">Área<\/option><option value="coord">Coordinador<\/option><\/select><\/label><\/div><p id="count"/);
  assert.match(admin, /&&\(!addedByFilter\|\|\(addedByFilter==='__none__'\?!p\.addedBy:p\.addedBy===addedByFilter\)\)&&/);
  const src = ['byText', 'newest', 'pointSort'].map(k => admin.match(new RegExp(`const ${k}=[^\\n]*`))[0]).join('\n') + '\n' + admin.match(/function pointOrder\(\)[^\n]*/)[0];
  const order = (sortBy, view, rows) => rows.slice().sort(new Function('sortBy', 'view', `${src};return pointOrder()`)(sortBy, view)).map(p => p.code);
  const rows = [{code: 'B', area: 'Venao', coordinator: 'Zoe', created: '2026-10-01T00:00:00Z'}, {code: 'A', area: 'Ávila', coordinator: '', created: '2026-10-03T00:00:00Z'}, {code: 'C', area: 'Pedasí', coordinator: 'Ana', created: '2026-10-02T00:00:00Z'}];
  assert.deepEqual(order('', 'approved', rows), ['A', 'B', 'C'], 'approved default: código');
  assert.deepEqual(order('', 'pending', rows), ['A', 'C', 'B'], 'por revisar default: most recent first');
  assert.deepEqual(order('recent', 'approved', rows), ['A', 'C', 'B']);
  assert.deepEqual(order('area', 'approved', rows), ['A', 'C', 'B'], 'Ávila, Pedasí, Venao');
  assert.deepEqual(order('coord', 'approved', rows), ['C', 'B', 'A'], 'no coordinator last');
  const card = admin.match(/function render\(\)[^\n]*/)[0];
  assert.match(card, /p\.status==='approved'\?`<div><button class="reject" data-id="\$\{clean\(p\.id\)\}" data-decision="archived">Archivar<\/button><\/div>`/);
  assert.match(admin, /if\(b\.dataset\.decision==='archived'&&!confirm\(archiveQuestion\)\)return;/, 'asks before archiving; Recuperar punto undoes it');
  assert.match(admin, /\$\('#archiveCamp'\)\.onclick=async e=>\{if\(!activeCamp\|\|!confirm\(archiveQuestion\)\)return;/);
});
test('public list: Ordenar by Más recientes (default), Más antiguos, Área or Coordinador, for everyone', () => {
  assert.match(template, /<label class="sortrow"><span class="needlabel">Ordenar<\/span><select id="sortList"><option value="recent">Más recientes<\/option><option value="oldest">Más antiguos<\/option><option value="area">Área<\/option><option value="coord">Coordinador<\/option><\/select><\/label>/);
  assert.doesNotMatch(template, /body:not\(\.signed-in\) [^{]*\.sortrow/, 'not limited to signed-in users');
  assert.match(template, /\.includes\(q\)\)\.sort\(listOrder\(\)\);/);
  const src = template.match(/const addedAt=[^\n]*/)[0] + '\n' + template.match(/function listOrder\(\)[^\n]*/)[0];
  const order = (value, rows) => rows.slice().sort(new Function('$', `${src};return listOrder()`)(() => ({value}))).map(r => r.code);
  const rows = [{code: 'B', area: 'Venao', coordinator: 'Zoe', created: '2026-10-01T00:00:00Z'}, {code: 'A', area: 'Ávila', created: '2026-10-03T00:00:00Z'}, {code: 'C', area: 'Pedasí', coordinator: 'Ana', created: '2026-10-02T00:00:00Z'}];
  assert.deepEqual(order('recent', rows), ['A', 'C', 'B']);
  assert.deepEqual(order('oldest', rows), ['B', 'C', 'A']);
  assert.deepEqual(order('area', rows), ['A', 'C', 'B']);
  assert.deepEqual(order('coord', rows), ['C', 'B', 'A'], 'no coordinator last');
});
test('admin ☰ menu: «Requerir revisión de puntos» toggles whether public points wait for review', () => {
  assert.match(admin, /<div class="adminmenu" id="adminMenu" role="menu" hidden>[^\n]*<button type="button" role="menuitemcheckbox" id="reviewToggle" aria-checked="true">Requerir revisión de puntos: <b>Sí<\/b><\/button>/);
  assert.match(admin, /const got=await adminRequest\('\/api\/admin\/settings'\);requireReviews=got\.requireReviews!==false/, 'read on load; Sí unless the server says otherwise');
  assert.match(admin, /if\(!next&&!confirm\('¿Publicar al momento los puntos que agregue el público\?/, 'turning it off asks first');
  assert.match(admin, /adminRequest\('\/api\/admin\/settings','POST',\{requireReviews:next\}\)/);
  assert.match(template, /saved\.status==='approved'\?'Punto publicado\. Ya aparece en el mapa y en la lista\.':'Se ha enviado tu solicitud\.'/, 'the public form already says which happened');
});
test('admin ☰ menu: «Descargar respaldo» downloads the whole DB; «Importar respaldo…» only on a local test server', () => {
  assert.match(admin, /<a role="menuitem" id="backupDownload" href="\/api\/admin\/backup" download>Descargar respaldo<\/a><button type="button" role="menuitem" id="backupImport" hidden>Importar respaldo…<\/button><input type="file" id="backupFile" accept="\.gz,application\/gzip" hidden>/);
  assert.match(admin, /\$\('#backupImport'\)\.hidden=!got\.importEnabled/, 'import shows only when the server allows it');
  assert.match(admin, /if\(!confirm\(`¿Importar «\$\{file\.name\}»\? Se reemplazarán TODOS los datos/, 'importing asks first');
  assert.match(admin, /fetch\('\/api\/admin\/restore',\{method:'POST',credentials:'same-origin',headers:\{'Content-Type':'application\/gzip','X-CSRF-Token':csrf\},body:file\}\)/);
});
test('«Solo Nuestro Equipo»: code users can limit the public list/map and runner needs to points added by a code user', () => {
  assert.match(template, /<button type="button" class="stat fstat teambtn" id="teamOnly" aria-pressed="false"><b id="teamCount">—<\/b><span>SOLO NUESTRO EQUIPO<\/span><\/button>/);
  assert.match(template, /body:not\(\.signed-in\) \.coordwrap,body:not\(\.signed-in\) \.teambtn\{display:none!important\}/, 'only with a code sign-in');
  assert.match(template, /const teamPoint=r=>!!r\.team;/, 'team comes from runners with a code (server-side)');
  assert.match(template, /const coordRows=\(\)=>all\(\)\.filter\(r=>\(!teamOnly\|\|teamPoint\(r\)\)&&/, 'filters the list, the map and the counts');
  assert.match(template, /if\(!u&&\(selectedCoordinator\|\|teamOnly\)\)\{selectedCoordinator='';teamOnly=false;/, 'signing out clears it');
  assert.match(runner, /<button type="button" class="teamonly" id="teamOnly" aria-pressed="false">👥 Solo Nuestro Equipo<\/button>/);
  assert.match(runner, /body:not\(\.code-user\) \.teamonly\{display:none\}/);
  assert.match(runner, /const listed=\(\)=>teamOnly\?available\.filter\(n=>n\.camp\.team\):available;/);
  assert.match(runner, /renderMap\(\[\.\.\.mine,\.\.\.shown\]\)/, 'your own runs always stay');
  assert.match(runner, /if\(!localStorage\.getItem\('runnerCode'\)\)teamOnly=false;/);
});
test('public narrow view: 📊 tab holds the totals, Lista keeps only filters; tapping your name copies your user ID', () => {
  assert.match(template, /<button type="button" data-view="map">◉ Mapa<\/button><button type="button" data-view="stats" class="iconview" title="Estadísticas" aria-label="Estadísticas">📊<\/button><\/nav>/, 'icon only');
  assert.match(template, /<div class="sidehead"><div class="stats pointstats">/);
  assert.match(template, /\.mobile-list \.pointstats\{display:none\}/, 'no totals in Lista on a phone');
  assert.match(template, /\.mobile-stats \.mapwrap\{display:none\}\.mobile-stats \.feed,\.mobile-stats \.sortrow,[^{]*\{display:none!important\}/, '📊 shows only totals and need counts');
  assert.match(template, /document\.body\.classList\.toggle\('mobile-stats',view==='stats'\)/);
  assert.match(template, /if\(document\.body\.classList\.contains\('mobile-stats'\)\)mobileView\('list'\);/, 'a need tapped in 📊 filters the list');
  assert.match(template, /if\(e\.target\.closest\('#copyMyId'\)\)\{copyMyId\(\);return\}/);
  const fn = template.match(/async function copyMyId\(\)[^\n]*/)[0];
  assert.match(fn, /let id=u\?\.id;if\(!id\)\{try\{const r=await fetch\('\/api\/full\/me'/, 'older saved sessions fetch it once');
  assert.match(fn, /navigator\.clipboard\.writeText\(id\);text='ID copiado: '\+id\.slice\(0,8\)\+'…'\}catch\{text='Tu ID: '\+id\}/);
  assert.match(fn, /setTimeout\(\(\)=>tip\.hidden=true,4000\)/);
  assert.match(template, /localStorage\.setItem\('codeUser',JSON\.stringify\(\{id:body\.id,name:body\.name/);
});
test('public detail for code users: each open/en ruta need can get a conductor and packing, like admin', () => {
  assert.match(template, /\$\{r\.id&&codeMe\(\)\?'<div id="teamNeeds"><\/div>':''\}/, 'only with a code sign-in');
  const draw = template.match(/function drawTeamNeeds\(\)[^\n]*/)[0];
  assert.match(draw, /filter\(n=>n\.status==='OPEN'\|\|n\.status==='IN_PROGRESS'\)/);
  assert.match(draw, /Conductor: <b>\$\{n\.runnerName\?clean\(n\.runnerName\):'sin asignar'\}<\/b>/);
  assert.match(draw, /PointForm\.packingHtml\(n\.packing\)/);
  assert.match(draw, /data-assign-need="\$\{clean\(n\.id\)\}" aria-haspopup="listbox">🚚 \$\{n\.runnerName\?'Cambiar conductor':'Asignar conductor'\}<\/button><button type="button" data-pack-need="\$\{clean\(n\.id\)\}">📦 Agregar a lista de empaque<\/button>/);
  assert.match(template, /<div class="picker assignpick" id="assignPicker" hidden><input id="assignSearch" type="search" placeholder="Buscar conductor…"/);
  assert.ok(template.indexOf('id="assignPicker"') < template.indexOf('<script>'), 'markup exists before the script binds it');
  assert.match(template, /crewCache=\{at:Date\.now\(\),list:await fullGet\('\/api\/full\/crew'\)\}/, 'the server lists code users first');
  assert.match(template, /c\.team\?'<small>Equipo<\/small>':''/);
  assert.match(template, /fullJSON\('\/api\/full\/needs\/'\+encodeURIComponent\(need\)\+'\/runner',\{runnerId:id\}\)/);
  assert.match(template, /PointForm\.openPacking\(\{subtitle:[^\n]*fullJSON\('\/api\/full\/needs\/'\+encodeURIComponent\(id\)\+'\/packing',\{items:lines\}\)/, 'same packing dialog as admin and runners');
  assert.match(template, /\.length;drawTeamNeeds\(\)\}/, 'refreshed with the list and after each save');
});
test('public detail: with needs to fulfil, the big buttons are Asignar conductor and packing; Reportar becomes a small link at the bottom', () => {
  assert.match(template, /<button type="button" class="footbtn reportsmall" data-report-need title="Reportar otra necesidad">➕ Reportar<\/button>/);
  assert.match(template, /\$\('#detailBody'\)\.classList\.toggle\('has-team-needs',list\.length>0\)/);
  assert.match(template, /\.has-team-needs \.reportneed\{display:none\}\.has-team-needs \.reportsmall\{display:inline-block\}/);
  assert.match(template, /\.teamacts button\{[^}]*padding:11px 12px;font:800 14px[^}]*background:#ed7654\}/);
  assert.match(template, /if\(!e\.target\.closest\('#reportNeed,\[data-report-need\]'\)\)return;/, 'both open the same form');
});
test('public detail: code users can remove the conductor (the need reopens)', () => {
  assert.match(template, /\$\{n\.runnerName\?` · <button type="button" class="unassign" data-unassign-need="\$\{clean\(n\.id\)\}">Quitar conductor<\/button>`:''\}/);
  assert.match(template, /async function unassign\(need\)\{if\(!confirm\('¿Quitar el conductor de esta necesidad\? Volverá a quedar abierta\.'\)\)return;try\{await fullJSON\('\/api\/full\/needs\/'\+encodeURIComponent\(need\)\+'\/runner',\{runnerId:''\}\)/);
});
test('public detail: Maps, Compartir and Reportar are small footer buttons; code users add a pickup photo; nothing of the team shows publicly', () => {
  assert.match(template, /d\.innerHTML\+=`<div class="detailfoot">\$\{maps\?`<a class="footbtn" target="_blank" rel="noopener noreferrer" href="\$\{clean\(maps\.url\)\}"/, 'the external Maps app, no embedded map');
  assert.doesNotMatch(template, /class="maplink"/);
  assert.match(template, /\$\{r\.coords\?'':'<p>Sin GPS: confirmar ubicación antes de salir\.<\/p>'\}/);
  assert.match(template, /\$\{mediaOn\?`<button type="button" data-photo-need="\$\{clean\(n\.id\)\}">📷 Foto de partida<\/button>`:''\}/);
  assert.match(template, /<dialog id="pickDialog" class="pickdlg"[^>]*><form method="dialog"><h2 id="pickTitle">Foto de partida<\/h2><div class="pickbtns"><button type="button" data-pick="pickPhoto">📷 Tomar foto<\/button><button type="button" data-pick="pickVideo">🎥 Grabar video<\/button><button type="button" data-pick="pickGallery">🖼️ Galería<\/button>/);
  assert.match(template, /<input type="file" id="pickPhoto" accept="image\/\*" capture="environment" hidden>/);
  assert.match(template, /url='\/api\/full\/needs\/'\+encodeURIComponent\(id\)\+'\/media\?stage=pickup'/);
  assert.match(template, /PointForm\.mediaThumbs\(n\.media,m=>'\/api\/full\/media\/'/);
  // Team-only: the section exists only with a code sign-in, and signing out wipes what was loaded with it.
  assert.match(template, /\$\{r\.id&&codeMe\(\)\?'<div id="teamNeeds"><\/div>':''\}/);
  assert.match(template, /if\(!u\)\{stopShare\(false\);[^}]*campNeeds\.clear\(\);\$\('#teamNeeds'\)\?\.remove\(\);/);
  const draw = template.match(/function drawTeamNeeds\(\)[^\n]*/)[0];
  assert.match(draw, /^function drawTeamNeeds\(\)\{const box=\$\('#teamNeeds'\);if\(!box\)\{/, 'draws nothing without the team section');
  assert.doesNotMatch(template.match(/function serviceNeedsHtml[^\n]*/)[0], /runnerName|packing|media/, 'the public need list never shows them');
});

test('Usuarios › Equipos: create/rename teams; a two-column table assigns each user with a team type-ahead, select all, and batch actions', () => {
  assert.match(admin, /<details id="teamsBox" class="teams-box"><summary>👥 Equipos <small id="teamsCount"><\/small><\/summary>/);
  assert.match(admin, /<form id="teamNewForm" class="team-new"><input name="name" maxlength="80" required/);
  assert.match(admin, /#operations\.user-form-mode #teamsBox,#operations\.user-view-mode #teamsBox,#operations\.user-form-mode #userBatch,#operations\.user-view-mode #userBatch\{display:none!important\}/, 'only on the list');
  assert.match(admin, /<label>Equipo<select name="teamId"><option value="">Sin equipo<\/option><\/select><\/label>/);
  assert.match(admin, /const data=\{[^}]*roles,teamId:form\.elements\.teamId\.value\};/, 'saved with the user');
  assert.match(admin, /<table class="team-users" id="teamUsers"><thead><tr><th><input type="checkbox" id="teamSelAll" aria-label="Seleccionar todos"><\/th><th>Usuario<\/th><th>Equipo<\/th><\/tr><\/thead><tbody><\/tbody><\/table>/);
  assert.match(admin, /<div class="team-batch" id="teamBatch" hidden><b id="teamSelCount"><\/b><input id="teamBatchInput" list="teamNames"[^>]*><button type="button" id="teamBatchAssign">Asignar equipo<\/button><button type="button" class="quiet-red" id="teamBatchRemove">Quitar del equipo<\/button><\/div>/);
  assert.match(admin, /!\$\('#teamsBox'\)\.contains\(document\.activeElement\)/, 'the 30 s refresh never interrupts typing a team');
  const fold = s => String(s).normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase();
  const clean = s => String(s ?? '').replace(/[&<>"']/g, c => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
  const src = ['drawTeams', 'drawTeamUsers', 'syncTeamSel'].map(f => admin.match(new RegExp(`function ${f}\\(\\)?[^\\n]*`))[0]).concat([admin.match(/const teamSel=[^\n]*/)[0], admin.match(/const teamRows=[^\n]*/)[0], admin.match(/const teamByName=[^\n]*/)[0]]).join('\n');
  const els = {}; const el = k => els[k] ??= {value: ''};
  const teams = [{id: 't1', name: 'Norte <b>'}, {id: 't2', name: 'Peñas'}];
  const users = [{id: 'u1', name: 'Fufo', teamId: 't1', hasAccessCode: true, roles: ['runner']}, {id: 'u2', name: 'Bea', hasAccessCode: true, roles: ['agent']}, {id: 'u3', name: 'Ana', hasAccessCode: false, roles: ['runner']}];
  const api = new Function('$', 'teams', 'users', 'renamingTeam', 'clean', 'fold', 'teamName', 'roleNames', `${src};return {drawTeams, teamSel, teamByName}`)(el, teams, users, null, clean, fold, id => teams.find(t => t.id === id)?.name || '', {runner: 'Corredor', agent: 'Agente'});
  api.drawTeams();
  const body = el('#teamUsers tbody').innerHTML;
  assert.match(el('#teamNames').innerHTML, /<option value="Norte &lt;b&gt;"><\/option><option value="Peñas">/, 'type-ahead list of teams');
  assert.match(el('#teamList').innerHTML, /<b>👥 Norte &lt;b&gt;<\/b><small>1 miembro<\/small>/);
  assert.equal((body.match(/<tr/g) || []).length, 3, 'every user is listed');
  assert.match(body, /data-team-sel="u1"[^>]*><\/td><td><b>Fufo<\/b>[^]*?<input class="team-pick" list="teamNames" data-team-user="u1" value="Norte &lt;b&gt;" placeholder="Sin equipo"/);
  assert.match(body, /<td><\/td><td><b>Ana<\/b>[^]*?Sin código: no puede estar en un equipo/, 'no checkbox or team picker without a code');
  assert.equal(el('#teamBatch').hidden, true, 'batch bar only with a selection');
  api.teamSel.add('u1'); api.teamSel.add('u2'); api.drawTeams();
  assert.equal(el('#teamSelAll').checked, true, 'all selectable users selected');
  assert.equal(el('#teamBatch').hidden, false);
  assert.equal(el('#teamSelCount').textContent, '2 seleccionados');
  api.teamSel.delete('u2'); api.drawTeams();
  assert.equal(el('#teamSelAll').indeterminate, true);
  assert.equal(api.teamByName(' penas '), 't2', 'typed names match without case or accents');
  assert.equal(api.teamByName(''), '', 'empty = no team');
  assert.equal(api.teamByName('Nada'), null);
  el('#teamUserSearch').value = 'peñas'; api.drawTeams();
  assert.match(el('#teamUsers tbody').innerHTML, /Sin coincidencias/);
});
test('public assign picker tags teammates with «Equipo»; the server lists only own team and runners without a code', () => {
  assert.match(template, /c\.team\?'<small>Equipo<\/small>':''/);
});

test('Usuarios: checkboxes per user; selected users that no longer match the search stay below the results; batch move/deactivate', () => {
  assert.doesNotMatch(admin, /id="userSelAll"/, 'select-all is a button over the results, not a filter checkbox');
  assert.match(admin, /<div class="user-batch" id="userBatch" hidden><b id="userSelCount"><\/b><span class="teampop-wrap"><button type="button" id="userBatchMove" aria-haspopup="listbox" aria-expanded="false">👥 Mover a equipo ▾<\/button><div class="teampop" id="teamPop" hidden><input id="teamPopSearch" type="search" placeholder="Buscar o crear equipo…"[^>]*role="combobox"[^>]*><div id="teamPopList" role="listbox"[^>]*><\/div><\/div><\/span><button type="button" class="quiet" id="userBatchUnteam">Quitar del equipo<\/button><button type="button" class="quiet-red" id="userBatchOff">Desactivar<\/button><button type="button" class="quiet" id="userBatchOn">Activar<\/button><button type="button" class="linkbtn" id="userBatchClear">Limpiar selección<\/button><\/div>/);
  assert.match(admin, /\.user-batch\{position:sticky;top:0/, 'the batch bar stays in view while scrolling');
  const src = [admin.match(/const userSel=[^\n]*/)[0], admin.match(/const userHead=[^\n]*/)[0], admin.match(/const userCard=[^\n]*/)[0], ...['userResults', 'drawUsers', 'syncUserSel', 'batchUsers'].map(f => admin.match(new RegExp(`(async )?function ${f}\\([^\\n]*`))[0])].join('\n');
  const els = {}; const el = k => els[k] ??= {value: '', checked: false, querySelectorAll: () => []};
  el('#userRole').value = ''; el('#userTeam').value = ''; el('#userSort').value = 'name';
  const users = [{id: 'u1', name: 'Fufo', roles: ['runner'], hasAccessCode: true}, {id: 'u2', name: 'Bea', roles: ['agent'], hasAccessCode: true, teamId: 't1'}, {id: 'u3', name: 'Ana', roles: ['runner'], hasAccessCode: false}];
  const patched = [];
  const fold = s => String(s).toLowerCase();
  const api = new Function('$', 'users', 'fold', 'clean', 'userRuns', 'userRunsHtml', 'keyBtn', 'phoneBtn', 'teamChip', 'roleNames', 'stamp', 'patchUsers', `${src};return {userSel, drawUsers, batchUsers}`)(el, users, fold, s => String(s ?? ''), () => [], () => '', () => '', () => '', () => '', {}, s => s, async (ids, body) => { patched.push([ids, body]); return ids.includes('u3') ? [{id: 'u3', text: 'Ana: no'}] : []; });
  api.drawUsers();
  assert.equal(el('#userBatch').hidden, true);
  assert.doesNotMatch(el('#userList').innerHTML, /users-chosen/, 'no Seleccionados group without a selection');
  assert.match(el('#userList').innerHTML, /<input type="checkbox" class="user-sel" data-user-sel="u1" aria-label="Seleccionar a Fufo">/);
  api.userSel.add('u1'); api.userSel.add('u3');
  el('#userSearch').value = 'bea'; api.drawUsers();
  let html = el('#userList').innerHTML;
  const chosen = html.indexOf('<details class="users-group users-chosen" open><summary>☑ Seleccionados (2)<button type="button" class="sel-all" data-unsel-all>☐ Deseleccionar todos</button></summary>'), results = html.indexOf('<div class="users-group users-results"><h3>Resultados (1)');
  assert.ok(chosen > 0 && results > chosen, 'collapsible Seleccionados on top, Resultados (not collapsible) below');
  assert.ok(html.indexOf('Seleccionar a Fufo') < results && html.indexOf('Seleccionar a Ana') < results, 'every selected user is grouped on top, even outside the search');
  assert.ok(html.indexOf('Seleccionar a Bea') > results);
  assert.match(html, /user-card sel"><input type="checkbox" class="user-sel" data-user-sel="u1" aria-label="Seleccionar a Fufo" checked>/);
  assert.match(html, /<button type="button" class="sel-all" data-sel-results>☑ Seleccionar el resultado<\/button>/);
  assert.equal(el('#userBatch').hidden, false);
  assert.equal(el('#userSelCount').textContent, '2 seleccionados');
  el('#userSearch').value = ''; api.drawUsers();
  html = el('#userList').innerHTML;
  assert.match(html, /<h3>Resultados \(1\) <small>· 2 ya seleccionados arriba<\/small>/, 'selected users are not repeated in the results');
  assert.equal((html.match(/data-user-sel="u1"/g) || []).length, 1);
  api.userSel.add('u2'); api.drawUsers();
  assert.match(el('#userList').innerHTML, /Resultados \(0\)[^]*Todos los resultados ya están seleccionados\./);
  api.userSel.delete('u2');
  assert.match(admin, /if\(!e\.target\.closest\('\[data-sel-results\]'\)\)return;for\(const u of userResults\(\)\)userSel\.add\(u\.id\);drawUsers\(\)/, 'adds to the selection; others stay');
  assert.match(admin, /if\(!e\.target\.closest\('\[data-unsel-all\]'\)\)return;e\.preventDefault\(\);userSel\.clear\(\);drawUsers\(\)/, 'un-select all, without folding the group');
  assert.match(admin, /addEventListener\('toggle',e=>\{if\(e\.target\.matches\('\.users-chosen'\)\)selOpen=e\.target\.open\},true\)/, 'collapsed state survives redraws');
  return api.batchUsers(u => u.hasAccessCode, {teamId: 't1'}, ['movido a Norte', 'movidos a Norte'], 'tiene código').then(() => {
    assert.deepEqual(patched, [[['u1'], {teamId: 't1'}]], 'users without a code are skipped');
    assert.deepEqual([...api.userSel], ['u3'], 'changed users leave the selection; skipped ones stay');
    assert.match(el('#usersNotice').textContent, /^1 usuario movido a Norte\. 1 no cumplía la condición \(tiene código\) y sigue marcado\.$/);
  });
});

test('batch «Mover a equipo» is a team search dropdown that can also create the team and move the selection there', async () => {
  const src = ['let popHl=0;', admin.match(/function teamPopOptions\(\)[^\n]*/)[0], admin.match(/function drawTeamPop\(\)[^\n]*/)[0], admin.match(/async function pickTeamPop\([^\n]*/)[0]].join('\n');
  const els = {}; const el = k => els[k] ??= {value: '', textContent: '', focus() {}, setAttribute() {}, hidden: false};
  el('#teamPopList .hl').scrollIntoView = () => {};
  const fold = s => String(s).normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase();
  const clean = s => String(s ?? '').replace(/[&<>"']/g, c => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
  let teams = [{id: 't1', name: 'Norte'}, {id: 't2', name: 'Peñas'}];
  const users = [{id: 'u1', name: 'Fufo', teamId: 't1', hasAccessCode: true}, {id: 'u2', name: 'Bea', hasAccessCode: true}];
  const calls = [];
  const adminRequest = async (url, method, body) => { calls.push([url, method, body]); return {id: 't3', name: body.name}; };
  const loadRunners = async () => { teams = [...teams, {id: 't3', name: 'Sur'}]; };
  const batchUsers = async (want, body, done) => { calls.push(['batch', body, done[1]]); el('#usersNotice').textContent = '1 usuario movido a Sur.'; };
  const api = new Function('$', 'getTeams', 'users', 'userSel', 'fold', 'clean', 'adminRequest', 'loadRunners', 'batchUsers', 'openTeamPop', `const teamName=id=>getTeams().find(t=>t.id===id)?.name||'';${src.replace(/(?<!\/)\bteams\b/g, 'getTeams()')};return {drawTeamPop, teamPopOptions, pickTeamPop}`)(el, () => teams, users, new Set(['u2']), fold, clean, adminRequest, loadRunners, batchUsers, () => {});
  api.drawTeamPop();
  assert.match(el('#teamPopList').innerHTML, /<button type="button" role="option" data-pop="0" aria-selected="true" class="hl">👥 Norte <small>1 miembro<\/small><\/button><button[^>]*data-pop="1"[^>]*>👥 Peñas <small>0 miembros<\/small>/, 'teams with member counts');
  el('#teamPopSearch').value = 'penas';
  assert.deepEqual(api.teamPopOptions().map(o => o.name || o.create), ['Peñas'], 'matches without accents; no create when the name exists');
  el('#teamPopSearch').value = '  Sur  ';
  api.drawTeamPop();
  assert.match(el('#teamPopList').innerHTML, /data-pop="0" aria-selected="true" class="hl create">＋ Crear equipo «Sur» y mover aquí<\/button>/);
  await api.pickTeamPop(0);
  assert.deepEqual(calls[0], ['/api/admin/teams', 'POST', {name: 'Sur'}], 'creates the team');
  assert.deepEqual(calls[1], ['batch', {teamId: 't3'}, 'movidos a Sur'], 'then moves the selection into it');
  assert.equal(el('#usersNotice').textContent, 'Equipo «Sur» creado. 1 usuario movido a Sur.');
  assert.match(admin, /if\(e\.key==='ArrowDown'\|\|e\.key==='ArrowUp'\)[^\n]*else if\(e\.key==='Enter'\)\{e\.preventDefault\(\);pickTeamPop\(popHl\)\}else if\(e\.key==='Escape'\)/, 'keyboard: arrows, Enter, Esc');
});

test('Usuarios: every user is a row with the same columns (Usuario, Accesos, Equipo, Actividad, Creado) under a header; labeled lines on phones', () => {
  assert.match(admin, /const userHead='<div class="user-head" aria-hidden="true"><span><\/span><span>Usuario<\/span><span>Accesos<\/span><span>Equipo<\/span><span>Actividad<\/span><span>Creado<\/span><span><\/span><\/div>';/);
  assert.match(admin, /\.user-card,\.user-head\{display:grid;grid-template-columns:22px minmax\(190px,2\.2fr\)[^}]*\}/, 'rows and header share one column grid');
  assert.match(admin, /@media\(max-width:720px\)\{\.user-head\{display:none\}[^}]*\}[^@]*\.uc-col:before\{content:attr\(data-label\)/, 'phones: one labeled line per column');
  assert.match(admin, /\$\{userHead\}\$\{chosen\.map\(userCard\)/);
  assert.match(admin, /rest\.length\?userHead\+rest\.map\(userCard\)/);
  const card = new Function('userRuns', 'userSel', 'clean', 'keyBtn', 'phoneBtn', 'roleNames', 'teamChip', 'stamp', 'userRunsHtml', `${admin.match(/const userCard=[^\n]*/)[0]};return userCard`)(() => [], new Set(), s => String(s ?? ''), () => '', () => '', {runner: 'Corredor'}, u => u.teamId ? '<span class="team-chip">👥 Norte</span>' : '', s => s, () => '');
  const a = card({id: 'u1abcdefgh', name: 'Fufo', roles: ['runner'], teamId: 't1', online: true, createdAt: '1 oct'});
  const b = card({id: 'u2abcdefgh', name: 'Ana', roles: [], disabled: true});
  const cols = h => [...h.matchAll(/<div class="uc-col" data-label="([^"]+)">/g)].map(m => m[1]);
  assert.deepEqual(cols(a), ['Accesos', 'Equipo', 'Actividad', 'Creado']);
  assert.deepEqual(cols(b), cols(a), 'same columns for every user, so they line up');
  assert.match(a, /data-label="Equipo"><span class="team-chip">👥 Norte<\/span>/);
  assert.match(b, /data-label="Accesos"><span class="uc-none">—<\/span>[^]*data-label="Equipo"><span class="uc-none">Sin equipo<\/span>[^]*Sin actividad[^]*data-label="Creado"><span class="uc-none">—<\/span>/, 'empty values are shown, not skipped');
  assert.match(b, /<b>Ana<\/b><span class="role-chip off">Desactivado<\/span>/);
});
