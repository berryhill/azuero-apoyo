import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
const camp = readFileSync(new URL('./camp.html', import.meta.url), 'utf8');
const template = readFileSync(new URL('./index.html', import.meta.url), 'utf8');
const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
const pointformSrc = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
const PointForm = (() => { const window = {}; const document = {getElementById: () => null, head: {insertAdjacentHTML() {}}}; new Function('window', 'document', pointformSrc)(window, document); return window.PointForm; })();
test('admin panel keeps login but has no master password change form', () => {
  assert.match(admin, /id="login"/);
  assert.doesNotMatch(admin, /id="changePassword"|\/api\/admin\/password|passwordMessage/);
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
  assert.match(template, /\$\{selectedCampStatus&&r\.id\?serviceNeedsHtml\(r\):/);
  const src = template.match(/const activeNeeds=[^\n]*\nfunction loadNeeds[^\n]*\nfunction serviceNeedsHtml[^\n]*/)[0];
  const render = new Function('clean', 'selectedCampStatus', 'fetch', 'renderFeed', `${src};campNeeds.set('c1',{gen:0,list:[{need:'Agua',status:'IN_PROGRESS',urgency:'urgente',createdAt:'2026-10-13T13:00:00Z',enRouteAt:'2026-10-13T14:30:00Z'},{need:'Comida',status:'OPEN',createdAt:'2026-10-13T12:00:00Z'},{need:'Ropa',status:'COMPLETED',createdAt:'2026-10-13T11:00:00Z'}]});campNeeds.set('c2',{gen:0,list:[{need:'Ropa',status:'COMPLETED'}]});return (id,all)=>serviceNeedsHtml({id},...(all?[activeNeeds]:[]))`)(s => String(s), 'IN_PROGRESS', () => { throw Error('should use cache'); }, () => {});
  const html = render('c1');
  assert.match(html, /<b>En ruta · Urgente<\/b><span>Agua<\/span><small>Creada: [^<]+<br>En ruta: [^<]+<\/small>/);
  assert.doesNotMatch(html, /Comida|Sin necesidades registradas/);
  // Sin filtro: el punto muestra sus necesidades abiertas y en ruta (no completadas), más reciente primero.
  assert.match(template, /\$\{r\.id&&activeNeeds\.some\(s=>hasNeed\(r,s\)\)\?serviceNeedsHtml\(r,activeNeeds\):''\}<div class="need">/);
  const unfiltered = render('c1', true);
  assert.match(unfiltered, /<b>En ruta · Urgente<\/b><span>Agua<\/span>[\s\S]*<b>Abierta<\/b><span>Comida<\/span>/);
  assert.doesNotMatch(unfiltered, /Ropa/);
  assert.equal(render('c2', true), '');
});
test('public page never labels points as "Aprobado"', () => {
  assert.doesNotMatch(template, /'Aprobado'|<label>Aprobado<\/label>/);
});
test('public list has a Verifica GPS box that drops a distinct temporary pin from a Maps link', () => {
  assert.match(template, /id="gpsCheck"[^>]*placeholder="Verifica GPS/);
  const src = template.match(/function coordsFromMapLink[^\n]*?:null\}/)[0];
  const coordsFromMapLink = new Function(`${src};return coordsFromMapLink`)();
  assert.deepEqual(coordsFromMapLink('https://maps.google.com/maps?q=7.3845953941345215%2C-80.45873260498047&z=17&hl=en'), [7.3845953941345215, -80.45873260498047]);
  assert.deepEqual(coordsFromMapLink('https://www.google.com/maps/place/X/@7.40,-80.30,17z/data=!8m2!3d7.4012345!4d-80.3123456'), [7.4012345, -80.3123456]);
  assert.equal(coordsFromMapLink('https://maps.app.goo.gl/abc'), null);
  assert.match(template, /className:'verifyicon'/);
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
  const make = (current, available) => new Function('current', 'followup', 'available', 'apple', `${src};return {popupNeed,popupList}`)(current, null, available, false);
  const n = {id: 'n1', need: 'Agua <b>', urgency: 'urgente', camp: {id: 'c1', code: 'C-1', area: 'Cañas', coords: [7.4, -80.3]}};
  const html = make(null, [n]).popupNeed(n, true);
  assert.match(html, /data-take="n1">¡Yo lo llevo!/);
  assert.match(html, /Agua &lt;b&gt;/);
  assert.match(html, /data-back="c1"/);
  assert.doesNotMatch(html, /class="where"/);
  assert.doesNotMatch(make(n, []).popupNeed(n, false), /data-take/);
  assert.match(make(null, [n]).popupList({camp: n.camp, needs: [n]}), /class="popneed" data-need="n1"/);
});
test('runner name gate also requires a phone number and sends it with the session', () => {
  const gate = runner.slice(runner.indexOf('id="gateForm"'), runner.indexOf('</form>', runner.indexOf('id="gateForm"')));
  assert.match(gate, /<input name="name" required/);
  assert.match(gate, /<input name="telephone" type="tel" required/);
  assert.match(runner, /JSON\.stringify\(\{name,telephone\}\)/);
  assert.match(runner, /startSession\(e\.target\.elements\.name\.value,e\.target\.elements\.telephone\.value\)/);
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
  assert.match(pub, /Agrega a alguien más/);
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
  assert.match(template, /PointForm\.render\(\$\('#pointFields'\)\)/);
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
  assert.doesNotMatch(template, /formIntro|Agregar punto/);
  assert.doesNotMatch(template, /Un administrador la revisará/);
  assert.match(modal, /<div class="notice sent" id="sentNotice" role="status" hidden><b>Se ha enviado tu solicitud\.<\/b>/);
  assert.doesNotMatch(template, /Se envía a revisión/);
  assert.match(template, /PointForm\.render\(\$\('#pointFields'\)\);showSent\(true\)\}/);
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
  const html = new Function('clean', 'stamp', 'needTimes', 'needChips', 'needStatsHtml', 'allNeeds', 'runners', 'builders', `${src};const needClass=n=>(needLook[n.status]||['deleted'])[0],needBadge=n=>'';return cardNeeds({id:'c'})`)(String, String, () => '', n => n.need, () => '<i>stats</i>', [
    {campId: 'c', status: 'IN_PROGRESS', urgency: 'urgente', need: 'techo', kind: 'repair', builderId: 'b1', runnerId: 'r1'},
    {campId: 'c', status: 'IN_PROGRESS', urgency: 'urgente', need: 'agua', runnerId: 'r1'},
    {campId: 'c', status: 'IN_PROGRESS', urgency: 'urgente', need: 'pan'},
    {campId: 'c', status: 'OPEN', urgency: 'urgente', need: 'arroz', runnerId: 'r1'},
    {id: 'n5', campId: 'c', status: 'OPEN', urgency: 'normal', need: 'leche'},
    {id: 'n6', campId: 'c', status: 'OPEN', urgency: 'normal', need: 'pared', kind: 'repair'},
    {id: 'n7', campId: 'c', status: 'COMPLETED', urgency: 'normal', need: 'agua vieja'},
  ], [{id: 'r1', name: 'Ana', telephone: '6000-1234'}], [{id: 'b1', name: 'Luis Obra', telephone: '6000-9999'}]);
  assert.match(html, /techo.*Constructor: <b>Luis Obra<\/b> · <a href="tel:60009999">/);
  assert.match(html, /agua.*Corredor: <b>Ana<\/b> · <a href="tel:60001234">6000-1234<\/a>/);
  // Sin asignar: botón «Asignar» del equipo que corresponde; asignada aunque siga abierta: se ve quién la tiene.
  assert.match(html, /pan<button type="button" class="assignbtn" data-assign-need="undefined" aria-haspopup="listbox">\+ Asignar corredor<\/button>/);
  assert.match(html, /leche<button type="button" class="assignbtn" data-assign-need="n5"[^>]*>\+ Asignar corredor<\/button>/);
  assert.match(html, /pared<button type="button" class="assignbtn" data-assign-need="n6"[^>]*>\+ Asignar constructor<\/button>/);
  assert.match(html, /arroz.*Corredor: <b>Ana<\/b>/);
  assert.doesNotMatch(html, /agua vieja<button/, 'no Asignar on completed');
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
  assert.match(builder, /<label>Nombre completo<input name="name" required/);
  assert.match(builder, /fetch\('\/api\/builder\/session'/);
  assert.match(builder, /api\('\/api\/builder\/needs'\);mine=data\.mine\|\|\[\];available=data\.available\|\|\[\];max=data\.max/);
  assert.doesNotMatch(builder, /\/api\/runner\//);
  assert.doesNotMatch(builder, /href="\/#agregar"/);
  assert.match(builder, /data-complete=.*Reparación terminada/);
  assert.match(admin, /<h2>Constructores<\/h2>/);
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
  assert.doesNotMatch(card, /Gestionar punto y necesidades|data-copy-id|Copiar enlace|Abrir ficha pública|data-decision="archived"/);
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
  assert.match(admin, /\$\('#closeCampEditor'\)\.textContent=editing\?'Cancelar':'← Volver';\$\('#closeCampEditor'\)\.classList\.toggle\('quiet-red',editing\);\$\('#adminFoot'\)\.hidden=editing;/);
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
  assert.match(form, /<div><b class="needpick-label">Necesidad<\/b><div id="needPicker" class="pf"><\/div><\/div><div class="need-row"><label>Urgencia<select name="urgency">[\s\S]*?<\/select><\/label><label>Estado<select name="status">[\s\S]*?<\/select><\/label><label class="need-assignee"><span id="needAssignLabel">Corredor \(opcional\)<\/span><select name="assignee" id="needRunner"><\/select><\/label><\/div><p id="needMessage"/);
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
  assert.doesNotMatch(template, /nearby:/, 'public form unchanged');
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
  assert.match(admin, /<b>\$\{needBadge\(n\)\}\$\{urgentBadge\(n\)\}<\/b>\$\{needChips\(n\)\}\$\{needTimes\(n\)\}/);
  assert.match(admin, /<b>\$\{label\}<\/b>\$\{k==='needs'\?needChips\(\{need:p\[k\]\}\):clean\(p\[k\]\)\}/, 'pending report needs as chips');
  assert.match(admin, /\.card-needs li\{padding:9px 0 9px 12px;/);
  assert.doesNotMatch(admin, /\.card-needs li\{padding-left:9px\}/);
});
test('admin need cards show repair as its icon chip only, and Asignar opens a searchable runner/builder list that saves', () => {
  assert.doesNotMatch(admin, /class="need-kind"/, 'no separate Reparación de casa badge');
  assert.match(admin, /needBadge=n=>`<span class="need-state \$\{needClass\(n\)\}">/);
  assert.match(admin, /<div class="picker" id="assignPicker" hidden><input id="assignSearch" type="search" autocomplete="off"><div id="assignList" role="listbox"><\/div><\/div><script>/);
  assert.ok(admin.indexOf('id="assignPicker"') < admin.indexOf("$('#assignSearch').oninput"), 'markup exists before the script uses it');
  assert.match(admin, /opts=crewOf\(n\)\.filter\(r=>!q\|\|fold\(r\.name\+' '\+r\.telephone\)\.includes\(q\)\)/);
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
