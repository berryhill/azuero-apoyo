import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
const camp = readFileSync(new URL('./camp.html', import.meta.url), 'utf8');
const template = readFileSync(new URL('./template.html', import.meta.url), 'utf8');
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
test('admin needs tab lists needs across camps and links back to private camp management', () => {
  assert.match(admin, /id="needsTab"[^>]*>Necesidades de servicio/);
  assert.match(admin, /id="allNeedList"/);
  assert.match(admin, /allNeeds\.filter\(/);
  assert.match(admin, /n\.campId===p\.id/);
  assert.match(admin, /admin\.html\?camp=/);
  assert.match(admin, /clean\(n\.need\)/);
  assert.match(admin, /needStates\[n\.status\]/);
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
