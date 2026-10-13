import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
const camp = readFileSync(new URL('./camp.html', import.meta.url), 'utf8');
const template = readFileSync(new URL('./template.html', import.meta.url), 'utf8');
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
test('admin toolbar filters camps by service need status, not urgency', () => {
  const toolbar = admin.slice(admin.indexOf('class="toolbar"'), admin.indexOf('id="mapCount"'));
  assert.match(toolbar, /id="needFilter"/);
  for (const [value, label] of needStatusOptions) assert.ok(toolbar.includes(`<option value="${value}">${label}</option>`), label);
  assert.doesNotMatch(toolbar, /Today|Hoy/);
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
