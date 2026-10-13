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
  const render = new Function('clean', 'selectedCampStatus', 'fetch', 'renderFeed', `${src};campNeeds.set('c1',{gen:0,list:[{need:'Agua',status:'IN_PROGRESS',urgency:'Today',createdAt:'2026-10-13T13:00:00Z',enRouteAt:'2026-10-13T14:30:00Z'},{need:'Comida',status:'OPEN',createdAt:'2026-10-13T12:00:00Z'},{need:'Ropa',status:'COMPLETED',createdAt:'2026-10-13T11:00:00Z'}]});campNeeds.set('c2',{gen:0,list:[{need:'Ropa',status:'COMPLETED'}]});return (id,all)=>serviceNeedsHtml({id},...(all?[activeNeeds]:[]))`)(s => String(s), 'IN_PROGRESS', () => { throw Error('should use cache'); }, () => {});
  const html = render('c1');
  assert.match(html, /<b>En ruta · Hoy<\/b><span>Agua<\/span><small>Creada: [^<]+<br>En ruta: [^<]+<\/small>/);
  assert.doesNotMatch(html, /Comida|Sin necesidades registradas/);
  // Sin filtro: el punto muestra sus necesidades abiertas y en ruta (no completadas), más reciente primero.
  assert.match(template, /\$\{r\.id&&activeNeeds\.some\(s=>hasNeed\(r,s\)\)\?serviceNeedsHtml\(r,activeNeeds\):''\}<div class="need">/);
  const unfiltered = render('c1', true);
  assert.match(unfiltered, /<b>En ruta · Hoy<\/b><span>Agua<\/span>[\s\S]*<b>Abierta<\/b><span>Comida<\/span>/);
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
  const src = [grab(/const clean=[^\n]*/), grab(/const urgencyLabel=[^\n]*/), grab(/function locationLink[^\n]*/), grab(/function needDetails[^\n]*/), grab(/function popupList[^\n]*/), grab(/function popupNeed[^\n]*/)].join('\n');
  const make = (current, available) => new Function('current', 'followup', 'available', 'apple', `${src};return {popupNeed,popupList}`)(current, null, available, false);
  const n = {id: 'n1', need: 'Agua <b>', urgency: 'Today', camp: {id: 'c1', code: 'C-1', area: 'Cañas', coords: [7.4, -80.3]}};
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
  assert.match(admin, /data-decision="rejected">Duplicado<\/button>/);
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
  assert.match(admin, /p\.status==='rejected'\?`<button data-id="\$\{clean\(p\.id\)\}" data-decision="approved">Aprobar<\/button>`/);
  assert.match(admin, /data-decision="approved">Recuperar punto<\/button>/);
});
test('camp needs show their creation time in Spanish on admin, runner and public camp pages', () => {
  assert.match(admin, /'Creada: '\+stamp\(n\.createdAt\)/);
  assert.match(runner, /<b>Creada:<\/b> \$\{n\.createdAt\?clean\(stamp\(n\.createdAt\)\):'sin fecha registrada'\}/);
  assert.match(runner, /const stamp=v=>v\?new Date\(v\)\.toLocaleString\('es-PA'/);
  assert.match(camp, /Creada: \$\{n\.createdAt\?clean\(new Date\(n\.createdAt\)\.toLocaleString\('es-PA'/);
});
test('runner need card lists the camp contact phones as tap-to-call links', () => {
  const grab = re => runner.match(re)[0];
  const src = [grab(/const clean=[^\n]*/), grab(/const stamp=[^\n]*/), grab(/const urgencyLabel=[^\n]*/), grab(/function locationLink[^\n]*/), grab(/function needDetails[^\n]*/)].join('\n');
  const needDetails = new Function('apple', `${src};return needDetails`)(false);
  const html = needDetails({id: 'n', need: 'Agua', urgency: 'Today', camp: {id: 'c', code: 'C-1', contacts: [{name: 'Ana', phone: '6555-1234'}, {phone: '+507 6000 0000'}]}});
  assert.match(html, /Contactos del punto:<\/b> Ana · <a href="tel:65551234">6555-1234<\/a> · <a href="tel:\+50760000000">\+507 6000 0000<\/a>/);
  assert.match(needDetails({id: 'n', need: 'Agua', urgency: 'Today', camp: {id: 'c', code: 'C-1'}}), /sin teléfono registrado/);
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
  assert.match(admin, /PointForm\.render\(\$\('#pointFormFields'\),\{admin:true\}\)/);
  assert.match(admin, /function drawCampFields\(p\)\{PointForm\.render\(\$\('#campFields'\),\{admin:true,edit:true,point:p\}\)\}/);
  assert.match(admin, /'PATCH',PointForm\.values\(\$\('#campFields'\)\)/);
  assert.match(admin, /id="pointCancel">Cancelar<\/button>/);
  assert.match(pointformSrc, /gps\.value = `https:\/\/maps\.google\.com\/\?q=\$\{lat\.toFixed\(6\)\},\$\{lng\.toFixed\(6\)\}`/);
});
test('Qué necesitas checkboxes round-trip through the Necesidades text, with Otros free text', () => {
  const pub = PointForm.html();
  const what = pub.slice(pub.indexOf('¿Qué necesitas?'));
  for (const label of ['Agua', 'Alimentos', 'Refugio (carpa / lona / toldo)', 'Kit de cocina', 'Kit de bebé', 'Gasolina para generador', 'Diésel para generador']) assert.ok(what.includes(`data-need="${label}"`), label);
  assert.match(what, /class="pf-otros-check"><span><svg[^]*?Otros<\/span>/);
  assert.equal(PointForm.needsText(['Agua', 'Kit de bebé'], ' Medicinas, ropa '), 'Agua, Kit de bebé, Medicinas, ropa');
  assert.deepEqual(PointForm.parseNeeds('agua, Kit de bebe, Medicinas, ropa'), {checked: ['Agua', 'Kit de bebé'], other: 'Medicinas, ropa'});
  assert.deepEqual(PointForm.parseNeeds('Agua y comida para 20'), {checked: [], other: 'Agua y comida para 20'});
  assert.deepEqual(PointForm.parseNeeds(''), {checked: [], other: ''});
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
test('admin map pins the current view, including por revisar and archivados, and counts points without GPS', () => {
  const src = admin.match(/function renderMap\(rows\)\{[^\n]*/)[0];
  assert.match(src, /const located=rows\.filter\(p=>Array\.isArray\(p\.coords\)&&p\.coords\.length===2\)/);
  assert.doesNotMatch(src, /status!=='archived'/);
  assert.match(admin, /const pinColor=\{approved:'#277a66',pending:'#e57450',archived:'#8a8f91',rejected:'#9a4536'\}/);
  const els = {'#mapCount': {}};
  new Function('$', 'L', 'pinColor', `${src};renderMap([{status: 'archived', coords: [7.4, -80.3]}, {status: 'pending'}])`)(s => els[s], undefined, {});
  assert.match(els['#mapCount'].textContent, /^1 ubicación en mapa · 1 sin GPS \(solo en la lista\)/);
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
  const src = [/const needLook=[^\n]*/, /const needActivity=[^\n]*/, /const runnerLine=[^\n]*/, /const requester=[^\n]*/, /const cardNeeds=[^\n]*/].map(pick).join('\n');
  const html = new Function('clean', 'stamp', 'needTimes', 'allNeeds', 'runners', `${src};const needClass=n=>(needLook[n.status]||['deleted'])[0],needBadge=n=>'';return cardNeeds({id:'c'})`)(String, String, () => '', [
    {campId: 'c', status: 'IN_PROGRESS', urgency: 'Today', need: 'agua', runnerId: 'r1'},
    {campId: 'c', status: 'IN_PROGRESS', urgency: 'Today', need: 'pan'},
    {campId: 'c', status: 'OPEN', urgency: 'Today', need: 'arroz', runnerId: 'r1'},
  ], [{id: 'r1', name: 'Ana', telephone: '6000-1234'}]);
  assert.match(html, /agua.*Corredor: <b>Ana<\/b> · <a href="tel:60001234">6000-1234<\/a>/);
  assert.match(html, /pan.*Corredor: sin asignar/);
  assert.doesNotMatch(html, /arroz[^<]*<small class="need-runner"/);
});
test('admin camp list: needs sort by latest activity first', () => {
  const src = admin.match(/const needActivity=[^\n]*/)[0];
  const needActivity = new Function(`${src};return needActivity`)();
  const rows = [{id: 'old', createdAt: '2026-10-01T10:00:00Z'}, {id: 'taken', createdAt: '2026-09-01T10:00:00Z', updatedAt: '2026-10-12T10:00:00Z'}, {id: 'done', createdAt: '2026-09-01T10:00:00Z', completedAt: '2026-10-05T10:00:00Z'}];
  assert.deepEqual(rows.sort((a, b) => needActivity(b) - needActivity(a)).map(n => n.id), ['taken', 'done', 'old']);
  assert.match(admin, /\.sort\(\(a,b\)=>needActivity\(b\)-needActivity\(a\)\)/);
});
