import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
const camp = readFileSync(new URL('./camp.html', import.meta.url), 'utf8');
const template = readFileSync(new URL('./index.html', import.meta.url), 'utf8');
const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
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
  const toolbar = admin.slice(admin.indexOf('class="toolbar"'), admin.indexOf('id="mapCount"'));
  assert.doesNotMatch(admin, /id="needFilter"|Sin filtro de necesidades/);
  assert.doesNotMatch(toolbar, /Today|Hoy/);
  assert.match(admin, /needStatus=needFilter;/);
  assert.match(admin, /needFilter=needFilter===b\.dataset\.stat\?'':b\.dataset\.stat;render\(\)/);
  assert.match(admin, /\/api\/admin\/needs/);
});
test('public map and list filter camps by service need status without an urgency filter', () => {
  for (const id of ['mapCampStatus', 'listCampStatus']) {
    const start = template.indexOf(`id="${id}"`);
    const select = template.slice(start, template.indexOf('</select>', start));
    for (const [value, label] of needStatusOptions) assert.ok(select.includes(`<option value="${value}">${label}</option>`), `${id} ${label}`);
  }
  assert.doesNotMatch(template, /CampUrgency|Todas las urgencias/);
  assert.match(template, /needStatuses/);
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
test('Agregar punto blocks submit until any one of the three contact phones has 7+ digits', () => {
  const validPhone = new Function(`${template.match(/const validPhone=[^\n]*/)[0]};return validPhone`)();
  assert.equal(validPhone('6555-1234'), true);
  assert.equal(validPhone('+507 6555 1234'), true);
  assert.equal(validPhone('12-34'), false);
  assert.equal(validPhone(''), false);
  assert.match(template, /\['phone','secondPhone','thirdPhone'\]\.some\(n=>validPhone\(v\(n\)\)\)/);
  assert.match(template, /id="phoneHint"/);
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
test('admin dashboard shows need counts by state at the top and clicking filters the list', () => {
  const dash = admin.slice(admin.indexOf('id="dashboard"'));
  assert.ok(dash.indexOf('id="adminStats"') < dash.indexOf('id="items"'));
  const src = admin.match(/function renderStats\(\)[^\n]*/)[0];
  const els = {'#adminStats': {}};
  new Function('$', 'needFilter', 'allNeeds', 'points', `${src};renderStats()`)(s => els[s], 'OPEN', [{campId: 'a', status: 'OPEN'}, {campId: 'a', status: 'OPEN'}, {campId: 'a', status: 'IN_PROGRESS'}, {campId: 'a', status: 'COMPLETED'}, {campId: 'z', status: 'OPEN'}], [{id: 'p', status: 'pending'}, {id: 'a', status: 'approved'}, {id: 'z', status: 'archived'}]);
  const html = els['#adminStats'].innerHTML;
  assert.match(html, /class="open active" data-stat="OPEN" aria-pressed="true"><b>2<\/b><span>Necesidades abiertas/);
  assert.match(html, /data-stat="IN_PROGRESS" aria-pressed="false"><b>1<\/b><span>Necesidades en curso/);
  assert.match(html, /data-stat="COMPLETED" aria-pressed="false"><b>1<\/b>/);
  assert.match(html, /data-stat="pending"><b>1<\/b><span>Puntos por revisar/);
  assert.match(admin, /\$\('#pendingTab'\)\.click\(\)/);
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
test('admin has an Archivados tab with Recuperar punto; Todos hides archived points', () => {
  assert.match(admin, /<button id="archivedTab">Archivados<\/button>/);
  assert.match(admin, /view==='archived'\?p\.status==='archived':p\.status!=='rejected'&&p\.status!=='archived'/);
  assert.match(admin, /data-decision="approved">Recuperar punto<\/button>/);
});
test('admin header holds Agregar punto and boletín actions, shown only when logged in', () => {
  const header = admin.slice(admin.indexOf('<header>'), admin.indexOf('</header>'));
  for (const re of [/id="topActions" hidden/, /id="addPoint">\+ Agregar punto</, /id="boletinToggle">Generar boletín</, /href="\/report\/boletin_general\.html"/]) assert.match(header, re);
  const dash = admin.slice(admin.indexOf('id="dashboard"'));
  assert.doesNotMatch(dash, /id="addPoint"|id="boletinToggle"/);
  assert.match(admin, /\$\('#dashboard'\)\.hidden=\$\('#topActions'\)\.hidden=false/);
});
test('admin Agregar punto form keeps a Cancelar button that hides and clears it', () => {
  const tpl = template.slice(template.indexOf('id="form"'), template.indexOf('</form>', template.indexOf('id="form"')));
  assert.match(tpl, /<button class="secondary" type="button" id="cancel">Cancelar<\/button>/);
  assert.match(admin, /const cancel=form\.querySelector\('#cancel'\);cancel\.id='pointCancel';[^\n]*cancel\.onclick=\(\)=>\{form\.reset\(\);\$\('#pointMessage'\)\.textContent='';panel\.hidden=true\}/);
  assert.match(admin, /form\.querySelector\('#phoneHint'\)\?\.remove\(\)/);
});
