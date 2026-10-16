import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
const camp = readFileSync(new URL('./camp.html', import.meta.url), 'utf8');
const template = readFileSync(new URL('./index.html', import.meta.url), 'utf8');
const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
const agent = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
const pointformSrc = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
const PointForm = (() => { const window = {}; const document = {getElementById: () => null, head: {insertAdjacentHTML() {}}}; new Function('window', 'document', pointformSrc)(window, document); return window.PointForm; })();
test('admin ☰ menu can change the master password (typed twice), then signs in again', () => {
  assert.match(admin, /id="login"/);
  assert.match(admin, /<button type="button" role="menuitem" id="passwordToggle"><span class="mi" aria-hidden="true">[^<]*<\/span>Cambiar contraseña maestra<\/button>(?:<hr class="msep" role="separator">)?<a role="menuitem" id="backupDownload"[^\n]*?<\/div>/);
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
const needStatusOptions = [['', 'Sin filtro de necesidades'], ['OPEN', 'Abierto'], ['IN_PROGRESS', 'En camino'], ['COMPLETED', 'Completado']];
test('admin filters camps by need status only through the stats counters, without a dropdown', () => {
  assert.doesNotMatch(admin, /id="needFilter"|Sin filtro de necesidades/);
  assert.match(admin, /needStatus=needFilter;/);
  assert.match(admin, /needFilter=needFilter===b\.dataset\.stat\?'':b\.dataset\.stat;view='approved'\}render\(\)/);
  assert.match(admin, /\/api\/admin\/needs/);
});
test('public stats are static counts; only need-status and coordinator stat buttons filter, with no dropdowns or chips', () => {
  assert.doesNotMatch(template, /mapCampStatus|listCampStatus|Sin filtro de necesidades|id="filters"|Por confirmar|data-filter="(all|mapped|unmapped)"/);
  for (const [id, label] of [['total', 'Puntos'], ['mapped', 'En mapa'], ['unmapped', 'Sin GPS']]) assert.match(template, new RegExp(`<div class="stat"><b id="${id}">—</b><span>${label}</span></div>`), label);
  assert.match(template, /\$\('#total'\)\.textContent=everyRow\.length/);
  for (const [value, label] of [['OPEN', 'Abiertas'], ['IN_PROGRESS', '<svg class="fico"[^>]*>.*?</svg>En camino'], ['COMPLETED', 'Completadas']]) assert.match(template, new RegExp(`<button type="button" class="stat fstat" data-need="${value}"[^>]*><b id="need${value}">—</b><span>${label}</span>`), label);
  assert.doesNotMatch(template, /CampUrgency|Todas las urgencias/);
  assert.match(template, /needStatuses/);
});
test('public map/list default to «Con necesidades»; pins are light green (camp), blue (need) or light red (urgent need)', () => {
  assert.match(template, /<div class="stats" id="needStats"><button type="button" class="stat fstat active" data-need="ACTIVE" aria-pressed="true"><b id="needACTIVE">—<\/b><span>Todas<\/span><\/button><button type="button" class="stat fstat urgentf" data-need="URGENT" aria-pressed="false"><b id="needURGENT">—<\/b><span>Urgentes<\/span><\/button><button type="button" class="stat fstat" data-need="NORMAL" aria-pressed="false"><b id="needNORMAL">—<\/b><span>Normales<\/span><\/button><button type="button" class="stat fstat" data-need="OPEN"/);
  assert.match(template, /\|\|selectedCoordinator\):'Elegir…';/, 'no coordinator filter reads as a placeholder, not «Todos»');
  assert.match(template, /\.filterpanel \.coordbtn:not\(\.active\) b\{display:none\}/);
  assert.match(template, /isNormal=r=>hasActiveNeed\(r\)&&r\.highestUrgency!=='urgente'/);
  assert.match(template, /body:not\(\.mobile-stats\) \.filterpanel \.fstat:is\(\[data-need=ACTIVE\],\[data-need=OPEN\],\[data-need=COMPLETED\]\),body:not\(\.mobile-stats\) \.filterpanel \.pointstats\{display:none\}/, 'filter panel: just Urgentes / Normales / En camino, no totals; all counts stay in 📊');
  assert.match(template, /selectedCampStatus=selectedCampStatus===b\.dataset\.need\?'ACTIVE':b\.dataset\.need;/, 'tapping the chosen pill again goes back to every point with needs');
  assert.match(template, /isUrgent=r=>hasActiveNeed\(r\)&&r\.highestUrgency==='urgente'/, '«Urgentes»: an open or en-route need marked urgent');
  assert.match(template, /selectedCampStatus==='URGENT'\?isUrgent\(r\)/);
  assert.match(template, /let selectedCoordinator='',selectedCampStatus='ACTIVE',teamOnly=false;/, 'on by default');
  assert.match(template, /selectedCampStatus==='ACTIVE'\?hasActiveNeed\(r\):selectedCampStatus==='URGENT'\?isUrgent\(r\):selectedCampStatus==='NORMAL'\?isNormal\(r\):selectedCampStatus==='SERVED'\?isServed\(r\):hasNeed\(r,selectedCampStatus\)/);
  assert.match(template, /\$\('#needACTIVE'\)\.textContent=needBase\.filter\(hasActiveNeed\)\.length/);
  assert.match(template, /'Ningún punto con necesidades abiertas o en camino\.'/, 'an empty list says why');
  const src = template.match(/const hasNeed=[^;]*;/)[0] + template.match(/const pinColors=[^\n]*/)[0];
  const {pinColor, pinColors} = new Function(`${src};return {pinColor, pinColors}`)();
  assert.deepEqual(pinColors, {camp: '#7fcf9a', need: '#5b9be6', urgent: '#f28b82'}, 'normal needs medium blue, urgent light red: not a wall of dark red');
  assert.equal(pinColor({needStatuses: ['COMPLETED']}), pinColors.camp, 'only completed needs = camp');
  assert.equal(pinColor({}), pinColors.camp);
  assert.equal(pinColor({needStatuses: ['IN_PROGRESS'], highestUrgency: 'normal'}), pinColors.need);
  assert.equal(pinColor({needStatuses: ['OPEN'], highestUrgency: 'urgente'}), pinColors.urgent);
  assert.match(template, /<div class="legend"><span><i style="background:#7fcf9a"><\/i>Punto<\/span><span><i style="background:#5b9be6"><\/i>Necesidad<\/span><span><i style="background:#f28b82"><\/i>Urgente<\/span><\/div>/, 'the map explains the colors');
});
test('public coordinator filter is a sidebar stat button with a searchable picker, not a map dropdown', () => {
  assert.doesNotMatch(template, /mapCoordinator|listCoordinator|<select[^>]*coordinador/i);
  assert.ok(template.indexOf('id="needStats"') < template.indexOf('id="coordOpen"'));
  assert.match(template, /<button type="button" class="stat fstat coordbtn" id="coordOpen" aria-haspopup="listbox"[^>]*><b id="coordCount">—<\/b><span id="coordLabel">Elegir…<\/span><\/button><div class="picker" id="coordPicker" hidden><input id="coordSearch"/);
  const src = template.match(/const coordNames=[^\n]*\nfunction drawCoordList[^\n]*/)[0];
  const els = {'#coordSearch': {value: 'mar'}, '#coordList': {innerHTML: ''}, '#coordLabel': {}, '#coordOpen': {classList: {toggle() {}}}};
  const api = new Function('$', 'all', 'clean', 'selectedCoordinator', `${src};return {drawCoordList, coordLabel}`)(s => els[s], () => [{coordinator: 'Marta Ruiz'}, {coordinator: 'Luis'}, {coordinator: ' Marta Ruiz '}, {}], s => String(s), 'marta ruiz');
  api.drawCoordList();
  assert.equal((els['#coordList'].innerHTML.match(/data-coord=/g) || []).length, 1);
  assert.match(els['#coordList'].innerHTML, /data-coord="marta ruiz" class="hl sel" aria-selected="true">Marta Ruiz \(2\)</);
  assert.equal(api.coordLabel(), 'Marta Ruiz');
});
test('public cards: each pending need on its own line, urgent and open first, oldest waiting first, with relative times', () => {
  assert.match(template, /const st=\['OPEN','IN_PROGRESS','COMPLETED'\]\.includes\(selectedCampStatus\)\?\[selectedCampStatus\]:selectedCampStatus==='SERVED'\?\['COMPLETED'\]:activeNeeds;/);
  assert.match(template, /\$\{r\.id&&st\.some\(s=>hasNeed\(r,s\)\)\?serviceNeedsHtml\(r,st\):r\.needs\?needChips\(r\.needs\):'<div class="need">Sin necesidades pendientes<\/div>'\}/, 'saved needs replace the reported list in the card, so items are not shown twice');
  const src = template.match(/const activeNeeds=[^\n]*\nfunction loadNeeds[^\n]*\nfunction serviceNeedsHtml[^\n]*\n[^\n]*\n[^\n]*\nconst agoEs=[^\n]*\nfunction needLine[^\n]*/)[0];
  const now = Date.now(), iso = h => new Date(now - h * 36e5).toISOString();
  const render = new Function('clean', 'selectedCampStatus', 'fetch', 'renderFeed', `${src};campNeeds.set('c1',{gen:0,list:[{need:'Agua',status:'IN_PROGRESS',urgency:'urgente',createdAt:'${iso(5)}',enRouteAt:'${iso(2)}'},{need:'Comida',status:'OPEN',createdAt:'${iso(30)}'},{need:'Techo',status:'OPEN',urgency:'urgente',createdAt:'${iso(3)}'},{need:'Ropa',status:'COMPLETED',createdAt:'${iso(50)}'}]});campNeeds.set('c2',{gen:0,list:[{need:'Ropa',status:'COMPLETED'}]});return (id,all)=>serviceNeedsHtml({id},...(all?[activeNeeds]:[]))`)(s => String(s), 'IN_PROGRESS', () => { throw Error('should use cache'); }, () => {});
  assert.match(render('c1'), /^<div class="nlines"><div class="nline in_progress"><div class="nhead"><span class="ntag"><svg[^]*?<\/svg>En camino<\/span><span class="badge urgent">Urgente<\/span><small title="[^"]+">hace 2 h<\/small><\/div><div class="chips"><span class="chip text">Agua<\/span><\/div><\/div><\/div>$/);
  const unfiltered = render('c1', true);
  assert.deepEqual([...unfiltered.matchAll(/<span class="chip text">(\w+)<\/span>/g)].map(m => m[1]), ['Techo', 'Agua', 'Comida'], 'urgent open, urgent en route, then normal');
  assert.match(unfiltered, /<span class="ntag">Necesita<\/span><span class="badge urgent">Urgente<\/span><small title="[^"]+">hace 3 h<\/small>/);
  assert.match(unfiltered, /<span class="ntag">Necesita<\/span><small title="[^"]+">hace 1 d<\/small>/, 'normal priority shows no label');
  assert.doesNotMatch(unfiltered, /Ropa/);
  assert.equal(render('c2', true), '');
  assert.doesNotMatch(template.match(/function needLine[^\n]*/)[0], /runnerName|packing|media/, 'never runner names or private data in public');
  // Card: border in the pin color, no legacy «Servicio reportado» badge, crossed pin when there is no GPS.
  assert.match(template, /<button class="entry \$\{!hasActiveNeed\(r\)\?\(hasNeed\(r,'COMPLETED'\)\?'served':'calm'\):r\.highestUrgency==='urgente'\?'urgent':'need'\}/);
  assert.match(template, /\.entry\{border-left:3px solid transparent\}\.entry\.urgent\{border-left-color:#f6b3ac\}/, 'only urgent cards get a soft red edge');
  assert.match(template, /\.ntag\{[^}]*background:#e8eef0;color:#36505a\}/, '«Necesita» is a light tag, not a dark block');
  assert.doesNotMatch(template.match(/function renderFeed\(\)[^\n]*/)[0], /status\(r\)|campStatusLabel|GPS disponible/);
});
test('public cards: area then people count on top; no generated ID or «Sin nombre»; who knows more below', () => {
  const autoCode = new Function(`${template.match(/const autoCode=[^\n]*/)[0]};return autoCode`)();
  assert.equal(autoCode('P-1a2b3c4d5e6f'), true);
  assert.equal(autoCode('AZ-12'), false);
  assert.match(template, /<span class="code">\$\{autoCode\(r\.code\)\?'':clean\(r\.code\)\+' '\}<span class="area">\$\{clean\(r\.area\)\}<\/span>\$\{vb\(r\)\}\$\{r\.heldUntil[^]*?\}\$\{r\.people\?` <span class="ppl" title="\$\{clean\(r\.people\)\} personas">\$\{clean\(r\.people\)\} 👥<\/span>`:''\}<\/span>/);
  assert.doesNotMatch(template, /Sin nombre/);
  assert.match(template, /\$\{r\.coordinator\?`<div class="meta">\$\{clean\(r\.coordinator\)\} sabe más<\/div>`:''\}/);
});
test('public point card on the map sits above the map controls and can be dragged by its top bar', () => {
  assert.doesNotMatch(template.match(/function renderMarkers[^\n]*|const m=L\.circleMarker[^\n]*/)[0], /bindPopup/, 'no small pop-up on the pin: the card is the detail');
  assert.match(template, /markers\.forEach\(\(m,code\)=>\{m\.setRadius\(code===r\.code\?13:9\)/, 'the chosen pin grows instead');
  assert.match(template, /<article class="detail" id="detail"><div class="detailgrip" id="detailGrip" title="Arrastrar para mover"><\/div><button class="close" id="closeDetail"/);
  assert.match(template, /\.detail\{position:absolute;z-index:1100;/, 'above Leaflet controls (1000), below dialogs (1200)');
  assert.match(template, /grip\.addEventListener\('pointerdown',e=>\{drag=\{x:e\.clientX-d\.offsetLeft,y:e\.clientY-d\.offsetTop\};grip\.setPointerCapture\(e\.pointerId\)/);
  assert.match(template, /\.detailgrip\{[^}]*touch-action:none/, 'drags with a finger without scrolling the page');
});
test('public filters: Atendidos shows places already served (completed needs, nothing pending) with their completed needs', () => {
  assert.match(template, /<button type="button" class="stat fstat servedf" data-need="SERVED" aria-pressed="false"><b id="needSERVED">—<\/b><span>Atendidos<\/span><\/button>/);
  assert.match(template, /isServed=r=>hasNeed\(r,'COMPLETED'\)&&!hasActiveNeed\(r\)/);
  assert.match(template, /selectedCampStatus==='SERVED'\?isServed\(r\)/);
  assert.match(template, /'Ningún punto atendido todavía\.'/);
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
  assert.match(template, /function searchQuery\(\)\{const v=\$\('#search'\)\.value\.trim\(\);return coordsFromMapLink\(v\)\|\|isMapLink\(v\)\?'':v\.toLocaleLowerCase\(\)\}/);
  assert.match(template, /const q=searchQuery\(\);/);
  assert.match(template, /No es un punto registrado\. ¿Quieres registrarlo\?<\/p><div class="gpsactions"><button type="button" class="gpsadd" data-register="/);
  assert.match(template, /\+ Registrar punto aquí<\/button><a class="gpsopen" target="_blank"/);
  assert.match(template, /gps\.value='https:\/\/maps\.google\.com\/\?q='\+b\.dataset\.register/, 'Registrar opens the form with the GPS filled in');
  assert.match(template, /className:'verifyicon'/);
});
test('public Buscar is only a search (no «!code» sign-in); signed-in name in the header', () => {
  assert.doesNotMatch(template, /v\.startsWith\('!'\)/, 'access codes are not typed into Buscar anymore');
  assert.match(template, /fetch\('\/api\/session\/code',\{method:'POST'/);
  assert.doesNotMatch(template.match(/<input[^>]*id="search"[^>]*>/)[0], /placeholder="[^"]*(código|!)/i);
  assert.match(template, /<span class="codeuser" id="codeUser" hidden><\/span>/);
  const show = template.match(/function showCodeUser\(\)[^\n]*/)[0];
  assert.match(show, /<button type="button" id="copyMyId" title="\$\{clean\(u\.name\)\} · copiar mi ID de usuario">\$\{clean\(window\.AppNav\?AppNav\.shortName\(u\.name\):u\.name\)\}<\/button><span class="idtip" id="idTip" role="status" hidden><\/span>/, 'short name («Ana Gon…»), no link; tapping it copies the user ID');
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
  const src = [grab(/const clean=[^\n]*/), grab(/const stamp=[^\n]*/), grab(/const urgencyLabel=[^\n]*/), grab(/function locationLink[^\n]*/), grab(/const foldNeed=[^\n]*/), grab(/function needChips[^\n]*/), grab(/const svgI=[^\n]*/), grab(/const noGpsIco=[^\n]*/), grab(/function needTitle[^\n]*/), grab(/function reporterIn[^\n]*/), grab(/function cardMeta[^\n]*/), grab(/const PPL=[^\n]*/), grab(/function peopleLine[^\n]*/), grab(/function locButtons[^\n]*/), grab(/const isCoord=[^\n]*/), grab(/function needDetails[^\n]*/), grab(/function popupList[^\n]*/), grab(/function popupNeed[^\n]*/)].join('\n');
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
  assert.match(form, /<span class="needpick-label">¿Qué necesitan\?<\/span><div class="pf" id="needPick"><\/div>/, 'same need checklist as everywhere else');
  assert.match(template, /\$\('#needPick'\)\.innerHTML=PointForm\.needsHtml\(\);PointForm\.fillNeeds\(\$\('#needPick'\),''\)/);
  assert.match(template, /const need=PointForm\.readNeeds\(\$\('#needPick'\)\);/);
  assert.match(form, /<input name="requesterName" required/);
  assert.match(form, /<input name="requesterPhone" type="tel"[^>]* required/);
  assert.match(template, /fetch\('\/api\/camps\/'\+encodeURIComponent\(needCamp\.id\)\+'\/needs'/);
  assert.match(template, /!validPhone\(v\('requesterPhone'\)\)/);
  assert.match(camp, /Reportado por \$\{clean\(n\.requesterName\)\}/);
  assert.match(admin, /Reportado por \$\{clean\(n\.requesterName\)\}/);
  assert.match(runner, /<small class="role">Pidió<\/small> \$\{clean\(n\.requesterName\)\}/);
});
test('admin stats bar: need counters filter approved points; Puntos por revisar and Archivados toggle the view', () => {
  const dash = admin.slice(admin.indexOf('id="dashboard"'));
  assert.ok(dash.indexOf('id="adminStats"') < dash.indexOf('id="items"'));
  const src = admin.match(/function renderStats\(\)[^\n]*/)[0];
  const draw = (needFilter, view, coordFilter = '') => { const els = {'#adminStats': {}, '#coordPicker': {hidden: true}}; new Function('$', 'needFilter', 'view', 'coordFilter', 'coordLabel', 'clean', 'allNeeds', 'points', `${src};renderStats()`)(s => els[s], needFilter, view, coordFilter, () => coordFilter ? 'Marta' : 'Todos', s => String(s), [{campId: 'a', status: 'OPEN'}, {campId: 'a', status: 'OPEN'}, {campId: 'a', status: 'IN_PROGRESS'}, {campId: 'a', status: 'COMPLETED'}, {campId: 'z', status: 'OPEN'}], [{id: 'p', status: 'pending'}, {id: 'a', status: 'approved'}, {id: 'z', status: 'archived'}, {id: 'y', status: 'archived'}, {id: 'r', status: 'rejected'}]); return els['#adminStats'].innerHTML; };
  let html = draw('OPEN', 'approved');
  assert.match(html, /class="open active" data-stat="OPEN" aria-pressed="true"><b>2<\/b><span>Abiertas<\/span>/);
  assert.match(html, /data-stat="IN_PROGRESS" aria-pressed="false"><b>1<\/b><span>En camino<\/span>/);
  assert.match(html, /data-stat="COMPLETED" aria-pressed="false"><b>1<\/b>/);
  assert.match(html, /^<button type="button" class="review urgent " data-view="pending" aria-pressed="false"><b>1<\/b><span>⚠ Por revisar<\/span><\/button><button type="button" class="all /);
  assert.match(html, /<span>Duplicados<\/span><\/button><i class="statsep" aria-hidden="true"><\/i><span class="statgrp">Necesidades<\/span><button type="button" class="open /, 'Puntos | Necesidades | Coordinador, in one compact row');
  assert.match(html, /class="archived " data-view="archived" aria-pressed="false"><b>2<\/b><span>Archivados/);
  assert.match(html, /class="rejected " data-view="rejected" aria-pressed="false"><b>1<\/b><span>Duplicados/);
  assert.match(admin, /<button class="merge" data-merge="\$\{clean\(p\.id\)\}">Duplicado…<\/button>/);
  assert.match(admin, /id="mergeNoCopy">Marcar duplicado sin copiar datos<\/button>/);
  assert.match(admin, /\$\{p\.status==='rejected'\?'duplicado':p\.status==='pending'\?'Esperando aprobación':clean\(p\.status\)\}<\/span>/);
  assert.match(html, /<button type="button" class="all " data-view="approved" aria-pressed="false"><b>1<\/b><span>Aprobados/);
  assert.match(html, /class="coord " data-coord-open aria-haspopup="listbox" aria-expanded="false"><span>Reportado por<\/span><b>Todos<\/b><span>▾<\/span>/);
  html = draw('', 'archived', 'marta');
  assert.match(html, /class="archived active" data-view="archived" aria-pressed="true"/);
  assert.match(html, /class="coord active"[^>]*><span>Reportado por<\/span><b>Marta<\/b>/);
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
  assert.match(template, /title="\$\{clean\(needStamp\(at\)\)\}">\$\{clean\(agoEs\(at\)\)\}<\/small>/, 'just «hace 1 d»; the exact date on hover');
  assert.match(camp, /n\.revisedAt\?'Actualizado: '\+clean\(stamp\(n\.revisedAt\)\):'Creada: '\+\(n\.createdAt\?clean\(stamp\(n\.createdAt\)\)/);
  const grab = re => runner.match(re)[0];
  const src = [grab(/const clean=[^\n]*/), grab(/const stamp=[^\n]*/), grab(/const urgencyLabel=[^\n]*/), grab(/function locationLink[^\n]*/), grab(/const foldNeed=[^\n]*/), grab(/function needChips[^\n]*/), grab(/const svgI=[^\n]*/), grab(/const noGpsIco=[^\n]*/), grab(/function needTitle[^\n]*/), grab(/function reporterIn[^\n]*/), grab(/function cardMeta[^\n]*/), grab(/const PPL=[^\n]*/), grab(/function peopleLine[^\n]*/), grab(/function locButtons[^\n]*/), grab(/const isCoord=[^\n]*/), grab(/function needDetails[^\n]*/)].join('\n');
  const needDetails = new Function('apple', 'AppNav', `${src};return needDetails`)(false, {phone: p => `[${p}]`});
  const camp1 = {id: 'c', code: 'C-1'};
  assert.match(needDetails({id: 'n', need: 'Agua', urgency: 'urgente', createdAt: '2026-10-01T12:00:00Z', camp: camp1}), /<small class="when" title="Reportada: Oct 1, 2026 [^"]+">hace \d+ d<\/small>/, 'just «hace 1 d» in the footer; what and the exact date on hover');
  assert.doesNotMatch(needDetails({id: 'n', need: 'Agua', urgency: 'urgente', createdAt: '2026-10-01T12:00:00Z', revisedAt: '2026-10-02T12:00:00Z', camp: camp1}), /Reportada/);
  assert.match(needDetails({id: 'n', need: 'Agua', urgency: 'urgente', revisedAt: '2026-10-02T12:00:00Z', camp: camp1}), /class="when" title="Actualizada: [^"]+">hace/);
  assert.doesNotMatch(needDetails({id: 'n', need: 'Agua', urgency: 'urgente', camp: camp1}), /class="when"|sin fecha/, 'no date: nothing, not «sin fecha registrada»');
});
test('runner need card lists the camp contact phones with copy / call / WhatsApp (shared nav.js phone)', () => {
  const grab = re => runner.match(re)[0];
  const src = [grab(/const clean=[^\n]*/), grab(/const stamp=[^\n]*/), grab(/const urgencyLabel=[^\n]*/), grab(/function locationLink[^\n]*/), grab(/const foldNeed=[^\n]*/), grab(/function needChips[^\n]*/), grab(/const svgI=[^\n]*/), grab(/const noGpsIco=[^\n]*/), grab(/function needTitle[^\n]*/), grab(/function reporterIn[^\n]*/), grab(/function cardMeta[^\n]*/), grab(/const PPL=[^\n]*/), grab(/function peopleLine[^\n]*/), grab(/function locButtons[^\n]*/), grab(/const isCoord=[^\n]*/), grab(/function needDetails[^\n]*/)].join('\n');
  const needDetails = new Function('apple', 'AppNav', `${src};return needDetails`)(false, {phone: p => `[${p}]`});
  const html = needDetails({id: 'n', need: 'Agua', urgency: 'urgente', camp: {id: 'c', code: 'C-1', contacts: [{name: 'Ana', phone: '6555-1234'}, {phone: '+507 6000 0000'}]}});
  assert.match(html, /<li><svg[^]*?<\/svg><span>Ana \[6555-1234\]<\/span><\/li><li><svg[^]*?<\/svg><span>\[\+507 6000 0000\]<\/span><\/li>/, 'one icon line per contact');
  assert.match(needDetails({id: 'n', need: 'Agua', urgency: 'urgente', camp: {id: 'c', code: 'C-1'}}), /Sin teléfono registrado/);
  // The reporter is not a contact line: their name is a tappable «📣 name» in the place line that opens their phone.
  const co = needDetails({id: 'n', need: 'Agua', urgency: 'urgente', camp: {id: 'c', code: 'C-1', coordinator: 'Sarah Spalding', contacts: [{name: 'Sarah Spalding (coordinador)', phone: '6629-6224'}]}});
  assert.match(co, /<button type="button" class="reporter" data-reporter="6629-6224" data-rname="Sarah Spalding" title="Reportado por Sarah Spalding: ver teléfono">📣 Sarah Spalding<\/button>/);
  assert.doesNotMatch(co, /<li>[^]*Sarah Spalding[^]*<\/li>|Sin teléfono registrado/, 'no contact line for the reporter');
  // Same phone as a contact: the reporter is named once, without repeating the number.
  assert.equal(needDetails({id: 'n', need: 'Agua', requesterName: 'Javier', requesterPhone: '6895-5161', camp: {id: 'c', code: 'C-1', contacts: [{name: 'Javier', phone: '6895 5161'}]}}).match(/6895/g).length, 1);
});
test('admin header holds Agregar punto and boletín actions, shown only when logged in', () => {
  const header = admin.slice(admin.indexOf('<header>'), admin.indexOf('</header>'));
  for (const re of [/id="topActions" hidden/, /id="addPoint">\+ Agregar punto</, /id="boletinToggle"><span class="mi" aria-hidden="true">[^<]*<\/span>Generar boletín</, /href="\/report\/boletin_general\.html"/]) assert.match(header, re);
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
  assert.equal((who.match(/class="pf-ico"><svg/g) || []).length, 8, 'icons on coordinator and the seven counts (with Familias)');
  assert.doesNotMatch(pub, /name="needs"|name="generator"/);
  assert.doesNotMatch(pub, /Agrega a alguien más/, 'a new empty row appears by itself');
});
test('point form people: one compact numbered row each, always one empty row ready to fill', () => {
  assert.doesNotMatch(pointformSrc, /Persona \$\{i \+ 1\}<button/, 'no Persona N heading');
  assert.match(pointformSrc, /<div class="pf-member"><span class="pf-mnum" aria-hidden="true">\$\{i \+ 1\}<\/span>\$\{memberFields\.map\(\(\[k, l, f\]\) => `<span class="pf-mf pf-m-\$\{k\}"><span class="pf-mi" aria-hidden="true">\$\{icons\[memberIcons\[k\]\]\}<\/span><input data-k="\$\{k\}" value="\$\{esc\(m\[k\]\)\}" placeholder="\$\{l\}"/, 'person, phone, age and ID icons in the fields');
  assert.match(pointformSrc, /const memberIcons = \{name: 'user', phone: 'phone', age: 'cake', cedula: 'idcard'\};/);
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
  assert.match(template, /<form id="form"><div id="pointFields"><\/div>/);
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
  for (const label of ['Agua', 'Alimentos', 'Refugio (carpa / lona / toldo)', 'Kit de cocina', 'Kit de bebé', 'Combustible para generador', 'Tarjeta de datos prepago', 'Gas de cocina']) assert.ok(what.includes(`data-need="${label}"`), label);
  // Reparación de casa sale en todos los formularios (con materiales); el servidor la separa para constructores y nunca es pública.
  assert.ok(what.includes('data-need="Reparación de casa"'));
  assert.match(what, /class="pf-otros-check"><span><svg[^]*?Otros<\/span>/);
  assert.equal(PointForm.needsText(['Agua', 'Kit de bebé'], ' Medicinas, ropa '), 'Agua, Kit de bebé, Medicinas, ropa');
  assert.deepEqual(PointForm.parseNeeds('agua, Kit de bebe, Medicinas, ropa'), {checked: ['Agua', 'Kit de bebé'], other: 'Medicinas, ropa', brands: {}, net: ''});
  assert.deepEqual(PointForm.parseNeeds('Agua y comida para 20'), {checked: [], other: 'Agua y comida para 20', brands: {}, net: ''});
  assert.deepEqual(PointForm.parseNeeds(''), {checked: [], other: '', brands: {}, net: ''});
  assert.deepEqual(PointForm.parseNeeds('Reparacion de casa, Agua'), {checked: ['Reparación de casa', 'Agua'], other: '', brands: {}, net: ''});
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
  assert.match(modal, /<h2 id="addTitle">Reportar punto<\/h2><button type="submit" form="form" class="addsave">Guardar<\/button><\/div><div class="addbody"><div class="notice sent"/);
  assert.match(template, /id="addBtn">Reportar punto<\/button>/);
  assert.doesNotMatch(template, /formIntro/);
  // Signed in with a code, the button reads «+ Agregar punto» like the admin and runner pages; otherwise «¡Ayuda!».
  assert.match(template, /\$\('#addBtn'\)\.textContent=\$\('#addBtnBar'\)\.textContent=u\?'\+ Agregar punto':'Reportar punto';/);
  assert.doesNotMatch(template, /Un administrador la revisará/);
  assert.match(modal, /<div class="notice sent" id="sentNotice" role="status" hidden><b id="sentTitle">Se ha enviado tu solicitud\.<\/b>/);
  assert.doesNotMatch(template, /Se envía a revisión/);
  assert.match(template, /PointForm\.render\(\$\('#pointFields'\),\{nearby:publicNearby,me:codeMe\(\)\}\);\$\('#sentTitle'\)\.textContent=[^;]*;showSent\(true\);/);
  assert.match(template, /\$\('#addBtn'\)\.onclick=\(\)=>\{showSent\(false\);/);
});
test('admin header: just the logo with Buscar… right after it; Salir in a small footer; coordinator typeahead', () => {
  const header = admin.slice(admin.indexOf('<header>'), admin.indexOf('</header>'));
  assert.match(header, /^<header><span class="adminlogo" title="Azuero Apoyo · Administración">✚<\/span><nav class="appnav" id="appNav" hidden><\/nav><div class="topsearch" id="topSearch" hidden><input id="search" type="search" placeholder="Buscar…"/);
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
  assert.match(els['#adminStats'].innerHTML, /^<button type="button" class="review  " data-view="pending" aria-pressed="false"><b>0<\/b><span>Por revisar<\/span>/);
});
test('¡Ayuda! window matches Runners: Cancelar · title · Guardar bar on top, full screen on a phone, both Guardar disable while sending', () => {
  assert.match(template, /<div class="addhead"><button type="button" class="addcancel" id="cancelTop">Cancelar<\/button><h2 id="addTitle">Reportar punto<\/h2><button type="submit" form="form" class="addsave">Guardar<\/button><\/div>/);
  assert.equal((template.slice(template.indexOf('<form id="form">'), template.indexOf('</form>', template.indexOf('<form id="form">'))).match(/type="submit">Guardar punto/g) || []).length, 1, 'one Guardar punto at the bottom; the top one is in the bar');
  assert.match(template, /const subs=\[\.\.\.document\.querySelectorAll\('#form \[type=submit\],\[form="form"\]\[type=submit\]'\)\];subs\.forEach\(x=>x\.disabled=true\)/, 'both disable while sending');
  assert.match(template, /@media\(max-width:760px\)\{#modal\.modalbg\{padding:0;align-items:stretch\}\.addmodal\{width:100%;max-height:none;border-radius:0\}/, 'full screen on a phone, like Runners');
});
test('admin por revisar: list items and map pins share a number, and selecting either highlights both', () => {
  assert.match(admin, /<span class="pinnum"[^`]*>\$\{i\+1\}<\/span>/);
  assert.match(admin, /num=p\.status==='pending'\?rows\.indexOf\(p\)\+1:0/);
  assert.match(admin, /if\(point\?\.status==='pending'\)\{selectPending\(point\.id,true\);return\}/);
  assert.match(admin, /else\{if\(num\)selectPending\(p\.id\);/);
  assert.match(admin, /el\.classList\.toggle\('selected',el\.dataset\.pointId===id\)/);
});
test('admin camp list: en camino needs show the runner name and tap-to-call phone', () => {
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
  assert.match(builder, /<form id="gateForm" novalidate data-signin><h1>Entrar a Constructores<\/h1>/);
  assert.match(builder, /fetch\('\/api\/builder\/session'/);
  assert.match(builder, /api\('\/api\/builder\/needs'\);mine=data\.mine\|\|\[\];available=data\.available\|\|\[\];max=data\.max/);
  assert.doesNotMatch(builder, /\/api\/runner\//);
  assert.doesNotMatch(builder, /href="\/#agregar"/);
  assert.match(builder, /data-complete="\$\{id\}">✓ Obra terminada<\/button>/);
  assert.match(admin, /name="roles" value="builder"> 🔨 Constructor</, 'builders are managed as users with the Constructor role');
  assert.match(admin, /await create\(\{need:parts\.find\(s=>PointForm\.needBase\(s\)===repairNeedLabel\)\|\|repairNeedLabel,kind:'repair',/, 'the repair need keeps its materials');
});
test('runner need cards show needs as icon chips, with free text as plain chips', () => {
  const grab = re => runner.match(re)[0];
  const src = [grab(/const clean=[^\n]*/), grab(/const foldNeed=[^\n]*/), grab(/function needChips[^\n]*/)].join('\n');
  const needChips = new Function('PointForm', `${src};return needChips`)(PointForm);
  const html = needChips('agua, Alimentos, colchones');
  assert.match(html, /<span class="chip"><svg[^]*?<\/svg>Agua<\/span><span class="chip"><svg[^]*?<\/svg>Alimentos<\/span><span class="chip text">colchones<\/span>/);
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
  assert.match(card, /p\.status==='approved'\?`<div class="card-foot">\$\{peopleLine\(p\)\}<button class="minibtn" data-id="\$\{clean\(p\.id\)\}" data-decision="archived">Archivar<\/button><\/div>`/);
  const src = ['clean', 'labels', 'contactKeys', 'memberList', 'peopleOf', 'peopleLine', 'shortName', 'addedByChip', 'noGpsIco', 'campViewHtml'].map(k => admin.match(new RegExp(`const ${k}=[^\\n]*`))[0]).join('\n');
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
  assert.match(admin, /#campEditor\.edit-mode>:not\(\.camp-head\):not\(#campEditDetails\):not\(#reviewNotes\)\{display:none!important\}/);
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
  assert.match(html, /<div class="route ico" title="En camino" role="img" aria-label="En camino: 1"><svg class="stat-ico"[^>]*><circle cx="14" cy="4" r="2"\/>.*?<\/svg><b>1<\/b><\/div>/);
  assert.match(html, /<div class="done ico" title="Completadas" role="img" aria-label="Completadas: 1"><svg class="stat-ico"[^>]*><rect[^>]*\/><path d="m8 12 3 3 5-6"\/><\/svg><b>1<\/b><\/div>$/);
  assert.doesNotMatch(html, /<span>(Abiertas|En camino|Completadas)<\/span>/);
  assert.match(admin, /class="editbtn" data-edit-need="\$\{clean\(n\.id\)\}"><svg[^`]*<\/svg>Editar<\/button>/);
  assert.doesNotMatch(admin, /Abrir \/ editar/);
  assert.match(admin, /edit\?'&edit=1':needId\?'&need='\+encodeURIComponent\(needId\):''/);
  assert.match(admin, /const shown=focusNeed\?regular\.filter\(n=>n\.id!==focusNeed\):regular;/);
  assert.match(admin, /#campEditor\.need-mode>:not\(\.camp-head\):not\(#needForm\):not\(#otherNeedsTitle\):not\(#needList\)\{display:none!important\}/);
});
test('admin need form uses the same icon checkboxes as Agregar punto and keeps options without a box in Otros', () => {
  const editor = admin.slice(admin.indexOf('id="needForm"'), admin.indexOf('</form>', admin.indexOf('id="needForm"')));
  assert.match(editor, /<div id="needPicker" class="pf"><\/div>/);
  assert.doesNotMatch(editor, /name="need"/);
  assert.match(admin, /\$\('#needPicker'\)\.innerHTML=PointForm\.needsHtml\(\);/);
  assert.match(admin, /need:PointForm\.readNeeds\(\$\('#needPicker'\)\)/);
  assert.ok(PointForm.html().includes(PointForm.needsHtml()), 'same needs block as Agregar punto');
  // Raíz falsa con las casillas públicas; una opción sin casilla («Medicinas») queda en Otros.
  const boxes = PointForm.NEEDS.filter(n => n[3] !== 'admin').map(([l]) => ({dataset: {need: l}, checked: false}));
  const otros = {hidden: true}, otrosCheck = {checked: false}, otrosText = {value: ''};
  const root = {querySelectorAll: s => s === '[data-need]' ? boxes : boxes.filter(b => b.checked), querySelector: s => ({'.pf-otros': otros, '.pf-otros-check': otrosCheck, '.pf-otros-text': otrosText})[s]};
  PointForm.fillNeeds(root, 'Agua, Reparación de casa, Medicinas');
  assert.deepEqual(boxes.filter(b => b.checked).map(b => b.dataset.need), ['Agua', 'Reparación de casa']);
  assert.equal(otrosText.value, 'Medicinas');
  assert.equal(otrosCheck.checked && !otros.hidden, true);
  assert.equal(PointForm.readNeeds(root), 'Agua, Reparación de casa, Medicinas');
  PointForm.fillNeeds(root, '');
  assert.equal(PointForm.readNeeds(root), '');
  assert.equal(otros.hidden, true);
  // «Otros» marcado ⇔ hay texto: escribir lo marca, desmarcarlo lo borra y el texto nunca se pierde al guardar.
  otrosText.value = 'Colchones';
  otrosText.oninput();
  assert.equal(otrosCheck.checked && !otros.hidden, true, 'typing ticks Otros');
  otrosCheck.checked = false;
  otrosText.value = 'Colchones';
  assert.equal(PointForm.readNeeds(root), 'Colchones', 'text in Otros is saved even if the box was unticked');
  assert.equal(otrosCheck.checked, true);
  otrosCheck.checked = false;
  otrosCheck.onchange();
  assert.equal(otrosText.value, '', 'unticking Otros clears its text');
  assert.equal(PointForm.readNeeds(root), '');
});
test('admin need form: Necesidad first, then Urgencia/Estado/assignee on one line, no Tipo; a repair need shows only a fixed Reparación de casa', () => {
  const form = admin.slice(admin.indexOf('<form id="needForm">'), admin.indexOf('</form>', admin.indexOf('<form id="needForm">')));
  assert.match(form, /<div><b class="needpick-label">Necesidad<\/b><div id="needPicker" class="pf"><\/div><\/div><div class="need-packing" id="needPacking">[\s\S]*?<\/div><div class="need-row"><label>Urgencia<select name="urgency">[\s\S]*?<\/select><\/label><label>Estado<select name="status">[\s\S]*?<\/select><\/label><label class="need-assignee"><span id="needAssignLabel">Corredor \(opcional\)<\/span><select name="assignee" id="needRunner"><\/select><\/label><\/div><p id="needMessage"/);
  assert.doesNotMatch(form, /name="kind"|Tipo/);
  assert.match(admin, /#campEditor\.need-mode #needTitle\{display:none\}/);
  assert.match(admin, /if\(n\.kind==='repair'\)\{\$\('#needPicker'\)\.innerHTML=PointForm\.needsHtml\(\{only:\[repairNeedLabel\]\}\);PointForm\.fillNeeds\(\$\('#needPicker'\),n\.need\|\|repairNeedLabel\);const box=\$\('#needPicker'\)\.querySelector\('\[data-need\]'\);box\.checked=true;box\.disabled=true;/, 'a repair need: the fixed box and its materials');
  assert.match(admin, /'Constructor \(opcional\)':'Corredor \(opcional\)'/);
  assert.match(admin, /function needKind\(\)\{if\(editingNeed\)return campNeeds\.find\(n=>n\.id===editingNeed\)\?\.kind\|\|'';/);
});
test('admin point view shows each need as non-clickable icon chips instead of comma text', () => {
  assert.match(admin, /<b>\$\{needBadge\(n\)\}\$\{urgentBadge\(n\)\}<\/b>\$\{needChips\(n\)\}/);
  const needChips = new Function('PointForm', 'clean', `${admin.match(/function needChips[^\n]*/)[0]};return needChips`)(PointForm, String);
  const html = needChips({need: 'Agua, Alimentos, Medicinas, ropa'});
  assert.match(html, /^<div class="needchips"><span class="needchip ">.*<\/svg>Agua<\/span><span class="needchip ">.*<\/svg>Alimentos<\/span><span class="needchip  text">Medicinas<\/span><span class="needchip  text">ropa<\/span><\/div>$/);
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
  assert.match(admin, /\$\{edit\?'Editar · ':needId&&needId!=='new'\?'Editar necesidad · ':''\}\$\{p\.code\} · \$\{p\.area\}/, 'new need: the title is just the point (the form says «Necesidad» right below)');
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
  assert.match(admin, /const choice=await askChoice\('¿Lograste comunicarte con alguien en este punto\?',\{yes:'Sí, necesitan cosas',alt:'Sí, están bien',no:'No',/, 'No / Sí, están bien / Sí, necesitan cosas');
  assert.match(admin, /if\(choice==='yes'\)\{const pending=campNeeds\.filter\(n=>n\.status==='OPEN'\|\|n\.status==='IN_PROGRESS'\);if\(pending\.length===1\)await openEditor\(activeCamp,true,false,pending\[0\]\.id\);else if\(!pending\.length\)await openEditor\(activeCamp,true,false,'new'\);/, '«necesitan cosas» opens the need form after the checkup');
  assert.match(admin, /if\(choice==='alt'&&openNeeds\.length\)\{for\(const n of openNeeds\)await adminRequest\('\/api\/admin\/needs\/'\+encodeURIComponent\(n\.id\),'PATCH',\{status:'DELETED'\}\);/, '«Sí, están bien» removes the needs still to deliver (en camino stay)');
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
  assert.equal((who.match(/class="pf-numbtn pf-dec"/g) || []).length, 7);
  assert.equal((who.match(/class="pf-numbtn pf-inc"/g) || []).length, 7);
  assert.match(who, /Familias<\/span><span class="pf-num"><input name="families"/);
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
  assert.match(admin, /<\/form><\/details><details id="campChangelog"><\/details><details id="campTrail" data-trail="">/);
  assert.doesNotMatch(admin, /historyHtml|Historial de personas/);
  assert.match(admin, /\$\('#campChangelog'\)\.innerHTML=changelogHtml\(p\)/);
  assert.match(admin, /\$\('#campChangelog'\)\.open=false;/);
});
test('admin Guardar punto returns to the updated punto view instead of staying on the edit page', () => {
  const submit = admin.match(/\$\('#campEditForm'\)\.onsubmit=[^\n]*/)[0];
  assert.match(submit, /'PATCH',PointForm\.values\(\$\('#campFields'\)\)\);const pending=editingPending\(\);await load\(\);if\(pending\)\{[^}]*\}const id=activeCamp;history\.replaceState\(null,'','\/admin\.html\?camp='\+encodeURIComponent\(id\)\);await openEditor\(id,false\);\$\('#needNotice'\)\.textContent='Punto actualizado\.'/);
  assert.match(submit, /catch\(err\)\{\$\('#campEditMessage'\)\.textContent='No se pudo guardar: '/);
});
test('Agregar punto: GPS opens a full-screen map centered on you, or on the last added punto without GPS', () => {
  assert.doesNotMatch(PointForm.html({admin: true}), /pf-gpsmap/, 'no inline map in the form');
  assert.match(admin, /const addPointNearby=\(\)=>points\.filter\(p=>p\.status==='approved'\|\|p\.status==='pending'\);/);
  assert.equal((admin.match(/nearby:addPointNearby/g) || []).length, 2, 'open and after save');
  assert.equal((template.match(/nearby:publicNearby/g) || []).length, 3, 'public ¡Ayuda! form gets the same map picker (open, after send, after saving offline)');
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
  assert.match(admin, /\$\{p\.needs\?`<div class="rv-sec"><span class="rv-label">Necesidades<\/span>\$\{needChips\(\{need:p\.needs\}\)\}<\/div>`:''\}/, 'pending report needs as chips');
  assert.match(admin, /\.card-needs li\{padding:8px 10px 8px 12px;border:1px solid #dce5e4;border-radius:9px;/, 'each need is its own bordered tile');
  assert.doesNotMatch(admin, /\.card-needs li\{padding-left:9px\}/);
});
test('admin need cards show repair as its icon chip only, and Asignar opens a searchable runner/builder list that saves', () => {
  assert.doesNotMatch(admin, /class="need-kind"/, 'no separate Reparación de casa badge');
  assert.match(admin, /needBadge=n=>`<span class="need-state \$\{needClass\(n\)\}">/);
  assert.match(admin, /<div class="picker" id="assignPicker" hidden><input id="assignSearch" type="search" autocomplete="off"><div id="assignList" role="listbox"><\/div><\/div><script src="\.\/nav\.js"><\/script><script>/);
  assert.ok(admin.indexOf('id="assignPicker"') < admin.indexOf("$('#assignSearch').oninput"), 'markup exists before the script uses it');
  assert.match(admin, /opts=crewOf\(n\)\.filter\(r=>!r\.disabled&&\(!q\|\|fold\(r\.name\+' '\+r\.telephone\)\.includes\(q\)\)\)/);
  assert.match(admin, /adminRequest\('\/api\/admin\/needs\/'\+encodeURIComponent\(n\.id\),'PATCH',n\.kind==='repair'\?\{builderId:id\}:\{runnerId:id\}\)/);
  assert.match(admin, /if\(e\.key==='Enter'\)\{e\.preventDefault\(\);const first=\$\('#assignList \[data-assign\]'\)/);
});
test('admin dashboard: no page title or intro, small sticky stats, fixed map with Mover and Maximizar', () => {
  assert.doesNotMatch(admin, /Panel de Administración|id="pageTitle"|Datos completos del registro/);
  assert.match(admin, /<section id="dashboard" hidden><div class="statswrap">/);
  assert.match(admin, /\.statswrap\{position:sticky;top:0;z-index:1001;/, 'stays above the map panes while scrolling');
  assert.match(admin, /\.adminstats b\{display:inline;font-size:15px;/);
  assert.match(admin, /\.adminstats button\{flex:none;display:inline-flex;align-items:baseline;gap:6px;/, 'number and label on one line');
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
  assert.match(admin, /<button type="button" role="menuitem" id="usersToggle"><span class="mi" aria-hidden="true">👥<\/span>Usuarios<\/button>/);
  assert.match(admin, /id="usersBack">← Volver al panel<\/button><h2 id="usersTitle">Usuarios<\/h2><button type="button" id="userNew">\+ Nuevo<\/button><button type="button" id="userEdit">Editar<\/button><button type="button" id="userToggle">Desactivar<\/button><button type="submit" form="userForm" id="userSaveTop">Guardar<\/button>/);
  assert.match(admin, /#dashboard\.users-mode>:not\(#operations\)\{display:none!important\}#dashboard:not\(\.users-mode\)>#operations\{display:none!important\}/);
  assert.match(admin, /history\.pushState\(null,'',on\?'\/admin\.html\?view=usuarios'\+\(viewing\?'&user='\+encodeURIComponent\(viewingUser\):''\):'\/admin\.html'\)/, 'returning from a point keeps the open user');
  assert.match(admin, /else if\(q\.get\('view'\)==='usuarios'\)\{showUsers\(true,false\);/, 'reloading the deep link reopens Usuarios');
  assert.match(admin, /showUsers\(q\.get\('view'\)==='usuarios',false\)/, 'back/forward keeps the page in sync');
  // Need cards resolve assignees from every user, so migrated or placeholder users still show a name.
  assert.match(admin, /const r=userById\(assigneeOf\(n\)\);/);
  assert.match(admin, /userById\(n\.releasedBy\)\?\.name/);
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
  assert.match(template, /<b>\$\{clean\(n\.status==='IN_PROGRESS'&&n\.dispatchedAt\?'Despachada':needStatusEs\[n\.status\]\|\|n\.status\)\} \$\{urgentBadge\(n\.urgency\)\}\$\{vb\(n\)\}<\/b>/);
  assert.match(camp, /\$\{urgentBadge\(p\.highestUrgency\)\}/);
  // A need without text (reserved for an access-code runner) shows only its status: no empty text or time line.
  assert.match(template, /\$\{n\.need\?needChips\(n\.need\):''\}/);
  assert.match(camp, /\$\{n\.need\?`\$\{needChips\(n\.need\)\}/, 'same need chips as the other pages');
  assert.match(camp, /<script src="\.\/pointform\.js"><\/script>/);
  assert.match(PointForm.needChipsHtml('Agua (Para beber), colchones'), /^<div class="pf-chips"><span class="pf-chip"><svg[^]*?<\/svg>Agua<span class="pf-netbadge pf-net-beber"[^>]*>para beber<\/span><\/span><span class="pf-chip text">colchones<\/span><\/div>$/);
});
test('public root page shows needs as icon chips, like the runner page', () => {
  const grab = re => template.match(re)[0];
  const line = grab(/const urgentBadge=[^\n]*/);
  const needChips = new Function('PointForm', 'clean', `${line};return needChips`)(PointForm, s => String(s));
  assert.match(needChips('agua, Alimentos, colchones'), /<span class="chip"><svg[^]*?<\/svg>Agua<\/span><span class="chip"><svg[^]*?<\/svg>Alimentos<\/span><span class="chip text">colchones<\/span>/, '«Otros» text without the other icon');
  assert.match(template, /:r\.needs\?needChips\(r\.needs\):'<div class="need">Sin necesidades pendientes<\/div>'\}/);
  assert.match(template, /\$\{r\.needs\?needChips\(r\.needs\):'<div class="value">Sin datos<\/div>'\}/);
  assert.doesNotMatch(template, /clean\(r\.needs/);
});
test('packing list: every starting item has its own icon and matches the server catalog', () => {
  const goSrc = readFileSync(new URL('./packing.go', import.meta.url), 'utf8');
  const goNames = [...goSrc.match(/var defaultPackingItems = \[\]string\{([^}]*)\}/)[1].matchAll(/"([^"]+)"/g)].map(m => m[1]);
  assert.deepEqual(PointForm.PACKING.map(p => p[0]), goNames);
  assert.equal(goNames.length, 21);
  assert.equal(goNames[0], 'Agua', 'water is a starting item');
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
  assert.match(pf, /openPacking, prepareMedia, uploadMedia, mediaThumbs, setMe, locationNotice, sendOrQueue, setQueue, queuePoint, flushQueue, enterOffline, SERVICES, serviceIcon, serviceStatus, serviceScheduled\};/);
  // Runner page: catalog + save through the runner API.
  assert.doesNotMatch(runner, /id="packDialog"/, 'no second copy of the dialog');
  assert.match(runner, /api\('\/api\/runner\/packing-items'\)/);
  assert.match(runner, /api\('\/api\/runner\/needs\/'\+encodeURIComponent\(n\.id\)\+'\/packing','POST',\{items:lines\}\)/);
  assert.match(runner, /<button class="btn take" type="button" data-take="\$\{clean\(n\.id\)\}"\$\{full\?[^}]*\}>¡Yo lo llevo!<\/button><\/div>\$\{n\.assigned\?packButton\(n\):''\}<\/article>/, 'open needs get the packing button only when assigned to this runner');
  assert.match(runner, /\$\{needDetails\(n,false\)\}<div class="actrow">\$\{needButtons\(n\)\}<button class="btn take"/, 'map, Maps, share and ¡Yo lo llevo! in one row');
  assert.match(runner, /\$\{c\.coords\?`<button type="button" class="iconbtn gps" data-where="\$\{clean\(n\.id\)\}" title="Ver en el mapa"[^`]*`:`<span class="iconbtn gps off"[^`]*\$\{noGpsIco\}<\/span>`\}/, 'look at it on the map before taking it; crossed pin when there is no GPS');
  assert.match(runner, /\$\{needDetails\(n,true,false\)\}\$\{mediaStrip\(n\)\}<div class="runacts">\$\{packButton\(n\)\}/);
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
  assert.match(card, /data-complete="\$\{id\}" title="Completé esta carrera">✓ Completé<\/button><\/div><div class="runfoot">[^]*?<button class="btn quit" type="button" data-release="\$\{id\}">/);
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
  assert.match(PointForm.needsHtml(), /data-need="Reparación de casa"/, 'runners can report a house repair (goes to builders, never public)');
});
test('admin point card: needs use the full card width as a grid of tiles, stats stay top right', () => {
  assert.match(admin, /#items article\.with-needs\{display:grid;grid-template-columns:minmax\(0,1fr\) auto;grid-template-areas:'main stats' 'needs needs';/);
  assert.match(admin, /\.with-needs>\.card-needs\{display:contents\}/);
  assert.match(admin, /\.with-needs>\.card-needs>\.needstats\{grid-area:stats;justify-self:end;/);
  assert.match(admin, /\.with-needs \.card-needs ul\{display:grid;grid-template-columns:1fr;gap:6px;/, 'one full-width row per need');
  assert.match(admin, /\.card-needs li\{display:grid;grid-template-columns:minmax\(150px,auto\) minmax\(0,1fr\) auto;grid-template-areas:'badge chips act' 'badge meta act';/, 'status · items · assign on one line');
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
  assert.match(runner, /askConfirm\('¿Estás seguro de que quieres completar esta carrera\?',\{yes:'Sí, completar',need:done,media:mediaOn,packing:true\}\)/);
  assert.match(runner, /askConfirm\('¿Estás seguro de que no puedes continuar esta carrera\?',\{yes:'Sí, dejarla',danger:true,/);
  assert.match(runner, /resolve\(d\.returnValue==='yes'\)/, 'Esc or No means no');
  assert.doesNotMatch(runner, /[^a-zA-Z]confirm\('¿Dejar esta carrera/);
});
test('runner panel shows priority only as an Urgente badge, never Normal', () => {
  assert.match(runner, /\$\{n\.urgency==='urgente'\?'<span class="badge urgente">Urgente<\/span>':''\}\$\{locButtons\(n\)\}<\/div>/);
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
  assert.match(card, /\$\{mediaStrip\(n\)\}<div class="runacts">\$\{packButton\(n\)\}\$\{!mediaOn\?'<small class="nomedia">📷 Fotos no disponibles<\/small>':\(n\.media\|\|\[\]\)\.some\(m=>m\.stage==='pickup'\)\?'':`<button class="btn plain media" type="button" data-media="\$\{id\}">📷 Foto de partida<\/button>`\}/, 'the big button only until the first pickup photo, and only with media storage');
  assert.match(runner, /\$\{st==='pickup'&&mediaOn\?`<button type="button" class="thumbadd" data-media="\$\{clean\(n\.id\)\}"[^>]*>\+<\/button>`:''\}/, 'then a + tile after the last pickup photo');
  assert.ok(card.indexOf('Foto de partida') < card.indexOf('Completé esta carrera'), 'Foto de partida comes before completing');
  // Foto de partida opens a chooser: Android's photo picker has no camera unless the input asks for it.
  assert.match(runner, /<dialog id="pickDialog"[^>]*><form method="dialog" class="confirmbox"><h2 id="pickTitle">Foto de partida<\/h2><div class="mediabtns pickbtns"><button type="button" class="btn plain" data-pick="pickPhoto">📷 Tomar foto/);
  assert.match(runner, /if\(media\)\{mediaTarget=media\.dataset\.media;\$\('#pickDialog'\)\.showModal\(\);return\}/);
  assert.match(runner, /files\.then\(f=>\{if\(target&&f\.length\)\(roOn\?roStoreMedia\(target,'pickup',f\):uploadRunMedia\(target,'pickup',f\)\)\}\)/);
  assert.ok(runner.indexOf('id="pickPhoto"') < runner.indexOf('<dialog id="confirmDialog"'), 'file inputs live outside the dialogs');  // Entrega: camera buttons on mobile, several items, thumbnails before completing.
  assert.match(runner, /<input type="file" id="pickPhoto" accept="image\/\*" capture="environment" hidden><input type="file" id="pickVideo" accept="video\/\*" capture="environment" hidden><input type="file" id="pickGallery" accept="image\/\*,video\/\*" multiple hidden>/);
  assert.match(runner, /data-pick="pickPhoto">📷 Tomar foto<\/button><button type="button" class="btn plain" data-pick="pickVideo">🎥 Grabar video<\/button><button type="button" class="btn plain" data-pick="pickGallery">🖼️ Galería<\/button>/);
  assert.match(runner, /files\.then\(f=>\{confirmFiles\.push\(\.\.\.f\);drawConfirmFiles\(\)\}\)/, 'each pick adds to the list');
  assert.match(runner, /URL\.createObjectURL\(f\)/, 'thumbnails preview before upload');
  assert.match(runner, /askConfirm\('¿Estás seguro de que quieres completar esta carrera\?',\{yes:'Sí, completar',need:done,media:mediaOn,packing:true\}\)/);
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
  assert.match(template, /AppNav\.endSessions\(\)/, 'signing out ends the viewer session too (global logout in nav.js)');
});
test('public coordinator filter only appears when signed in with an access code', () => {
  assert.match(template, /<span class="needlabel coordlabel">Reportado por<\/span><div class="coordwrap">/);
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
test('team members are the coordinator by default (a card with «Cambiar», no fields); «↺ Soy yo» puts them back; no «Yo» button', () => {
  const fn = pointformSrc.match(/function setMe\(root, me\) \{[^]*?\n  \}/)[0];
  assert.match(fn, /if \(!me \|\| !me\.name \|\| !me\.team\) return;/, 'only for a signed-in team member');
  assert.doesNotMatch(pointformSrc, /class="pf-yo"/);
  assert.match(fn, /<div class="pf-coordme"><span class="pf-cm-ico">\$\{icons\.megaphone\}<\/span><div class="pf-cm-txt"><small>Reportado por<\/small><b>\$\{esc\(me\.name\)\}<\/b><span class="pf-cm-sub">Tú · \$\{esc\(me\.team\)\}/);
  assert.match(fn, /set\(me\.name, me\.telephone \|\| ''\);\s*labels\.forEach\(l => \{ l\.hidden = true; \}\);/, 'fields filled and hidden');
  assert.match(fn, /<button type="button" class="pf-cm-change">Cambiar<\/button>/);
  assert.match(fn, /set\('', ''\); backLink\(\); coordEl\.focus\(\);/, '«Cambiar»: empty fields and «↺ Soy yo»');
  assert.match(fn, /if \(coordEl\.value\.trim\(\) && fold\(coordEl\.value\) !== fold\(me\.name\)\) \{ backLink\(\); return; \}/, 'someone else already typed: kept');
  assert.match(template, /team:u\.team\|\|''\}/);
  assert.match(template, /PointForm\.setMe\(\$\('#pointFields'\),codeMe\(\)\);refreshMeTeam\(\);/);
  assert.match(pointformSrc, /setMe\(root, me\);/);
  assert.equal((template.match(/me:codeMe\(\)/g) || []).length, 3, 'open, after send, after saving offline');
  assert.match(template, /\$\('#addBtn'\)\.onclick=\(\)=>\{showSent\(false\);\$\('#addTitle'\)\.textContent=[^;]*;PointForm\.setMe\(\$\('#pointFields'\),codeMe\(\)\);/);
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
  assert.match(admin, /<button type="button" class="menubtn" id="menuToggle" aria-expanded="false" aria-controls="adminMenu"[^>]*hidden>☰<\/button><div class="adminmenu" id="adminMenu" role="menu" hidden><button type="button" role="menuitem" id="usersToggle"><span class="mi" aria-hidden="true">[^<]*<\/span>Usuarios<\/button><a role="menuitem" href="\/donations\.html"><span class="mi" aria-hidden="true">[^<]*<\/span>Donaciones<\/a>(?:<hr class="msep" role="separator">)?<button type="button" role="menuitem" id="importPendingToggle"><span class="mi" aria-hidden="true">[^<]*<\/span>Importar reportes \(por revisar\)…<\/button>(?:<hr class="msep" role="separator">)?<button type="button" role="menuitem" id="boletinToggle"><span class="mi" aria-hidden="true">[^<]*<\/span>Generar boletín<\/button><a role="menuitem" class="boletinlink" href="\/report\/boletin_general\.html"[^>]*><span class="mi" aria-hidden="true">[^<]*<\/span>Ver último boletín ↗<\/a>(?:<hr class="msep" role="separator">)?<button type="button" role="menuitemcheckbox" id="reviewToggle"[^>]*>[^\n]*?<\/button><button type="button" role="menuitem" id="passwordToggle"><span class="mi" aria-hidden="true">[^<]*<\/span>Cambiar contraseña maestra<\/button>(?:<hr class="msep" role="separator">)?<a role="menuitem" id="backupDownload"[^\n]*?<\/div>/);
  assert.match(admin, /\$\('#menuToggle'\)\.onclick=e=>\{e\.stopPropagation\(\);setMenu\(\$\('#adminMenu'\)\.hidden\)\};/);
  assert.match(admin, /\$\('#adminMenu'\)\.addEventListener\('click',\(\)=>setMenu\(false\)\);/, 'choosing an option closes it');
  assert.match(admin, /if\(!e\.target\.closest\('\.menuwrap'\)\)setMenu\(false\)/, 'clicking outside closes it');
});
test('admin shows «Agregado por» for points published by access-code users', () => {
  assert.match(admin, /<small class="cardtop">\$\{p\.source==='sheet'\?'Registro importado':clean\(stamp\(p\.created\)\)\}\$\{p\.addedByName\?addedByChip\(p\):''\}<\/small>/, 'list card: chip next to the date');
  assert.doesNotMatch(admin, /'Agregado por '\+p\.addedByName/, 'not repeated in the summary line');
  const src = ['clean', 'labels', 'contactKeys', 'memberList', 'peopleOf', 'peopleLine', 'shortName', 'addedByChip', 'noGpsIco', 'campViewHtml'].map(k => admin.match(new RegExp(`const ${k}=[^\\n]*`))[0]).join('\n');
  const {campViewHtml} = new Function(`${src};return {campViewHtml}`)();
  assert.match(campViewHtml({code: 'C', area: 'A', addedByName: 'Fufo <b>'}), /^<div class="addedrow"><span class="addedby" title="Quién agregó este punto"><svg[^]*<\/svg>Agregado por <b>Fufo &lt;b&gt;<\/b><\/span><\/div><div class="grid">/, 'point file: chip at the top');
  assert.match(campViewHtml({code: 'C', area: 'A', source: 'submission'}), /<span class="addedby muted"[^>]*>[^]*Agregado por <b>Formulario público<\/b>/, 'no user: where it came from, muted');
  assert.match(campViewHtml({code: 'C', area: 'A', source: 'admin'}), /Agregado por <b>Administración<\/b>/);
  assert.doesNotMatch(campViewHtml({code: 'C', area: 'A'}), /Agregado por/);
  const chip = campViewHtml({code: 'C', area: 'A', addedByName: 'Arianne Orillac'});
  assert.match(chip, /Agregado por <b>Arianne Ori…<\/b>/, 'first name + first letters of the last name');
  assert.doesNotMatch(chip, /Orillac/, 'never the full name');
  assert.match(campViewHtml({code: 'C', area: 'A', addedByName: 'Ana Paz'}), /<b>Ana Paz<\/b>/, 'short last names stay whole');
  assert.match(campViewHtml({code: 'C', area: 'A', addedByName: 'Luis'}), /<b>Luis<\/b>/);
});

test('admin list on a computer: the map sits in a sticky column on the right, like the command center', () => {
  assert.match(admin, /<div class="mapcorner" id="mapCorner"><p id="mapCount" class="muted"><\/p><div class="mapwrap" id="mapWrap">[^]*?<\/div><\/div><\/div><div class="listtools">/);
  assert.match(admin, /@media\(min-width:1100px\)\{main:has\(#dashboard:not\(\[hidden\]\):not\(\.camp-mode\):not\(\.users-mode\)\)\{max-width:1440px\}#dashboard:not\(\.camp-mode\):not\(\.users-mode\)\{display:grid;grid-template-columns:minmax\(0,1fr\) 380px;/);
  assert.match(admin, />\.mapcorner\{grid-column:2;grid-row:1\/span 40;position:sticky;top:14px\}/);
});
test('Usuarios can filter to users with an access code only', () => {
  assert.match(admin, /<label class="check"><input type="checkbox" id="userOnlyCode"> Solo con código<\/label>/);
  assert.match(admin, /rows=users\.filter\(u=>\(showOff\|\|!u\.disabled\)&&\(!onlyCode\|\|u\.hasAccessCode\)&&/);
  assert.match(admin, /\$\('#userOnlyCode'\)\.onchange=drawUsers;/);
});
test('public header: top navigation for signed-in code users; «Mis carreras» lives in that menu', () => {
  assert.match(template, /<div class="topright"><nav class="appnav" id="appNav" hidden><\/nav>/);
  assert.doesNotMatch(template, /id="runnerNav"/, 'the top nav replaces the header «Mis carreras» button');
  assert.doesNotMatch(template, /id="runnerNavBar"|Mis carreras/, 'no separate «Mis carreras» button: it is in the top menu');
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
  assert.match(runner, /<\/div><nav class="appnav" id="appNav" hidden><\/nav><input class="topsearch" id="search" type="search" data-global="1" placeholder="Buscar en todo…" aria-label="Búsqueda global: todos los puntos \(Enter para buscar\)" autocomplete="off"><div class="topright">/, 'top navigation replaces «← Mapa público»');
  assert.match(runner, /function syncPublicNav\(\)\{window\.AppNav\?\.draw\(\);/);
  assert.equal((runner.match(/syncPublicNav\(\)/g) || []).length, 4, 'defined, after sign-in or session check, and after sign-out');
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
  assert.match(admin, /\$\('#userView'\)\.onclick=\$\('#userList'\)\.onclick;/, '🔑 and en camino runs work in the view too');
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
  assert.match(admin, /fold\(u\.name\+' '\+u\.telephone\+' '\+\(u\.email\|\|''\)\+' '\+u\.id\)\.includes\(q\)/, 'search by ID');
});
test('Usuarios shows a green/grey dot before each name and the last active time', () => {
  assert.match(admin, /<span class="presence\$\{u\.online\?' on':''\}" title="\$\{u\.online\?'En línea':'Desconectado'\}"[^>]*><\/span>\$\{keyBtn\(u\)\}<b>\$\{clean\(u\.name\)\}<\/b>/);
  assert.match(admin, /\$\{u\.online\?'En línea':u\.lastActiveAt\?'Activo '\+clean\(stamp\(u\.lastActiveAt\)\):''\}/);
  assert.match(admin, /\.presence\{[^}]*background:#b5c0c3[^}]*\}\.presence\.on\{background:#23a35a/);
  assert.match(admin, /setInterval\(\(\)=>\{if\(\$\('#dashboard'\)\.classList\.contains\('users-mode'\)&&!userFormOpen\(\)&&!renamingTeam&&!\$\('#teamsBox'\)\.contains\(document\.activeElement\)&&!\$\('#userBatch'\)\.contains\(document\.activeElement\)&&!Object\.keys\(keyTips\)\.length\)Promise\.all\(\[loadRunners\(\),adminRequest\('\/api\/admin\/needs'\)/, 'refreshes users and their runs');
});
test('Usuarios lists each runner\'s en camino runs (point and needs), with a count, and opens the point on click', () => {
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
  assert.deepEqual(runs.map(n => n.id), ['n1'], 'only their en camino supply runs');
  assert.deepEqual(userRuns({id: 'r1', roles: ['builder']}), [], 'only for runners');
  const html = userRunsHtml(runs);
  assert.match(html, /<button type="button" class="user-run" data-open-camp="c1" title="Abrir el punto">📍 <b>C-1<\/b> · Venao <small>· en camino desde Oct 14<\/small><i>Agua<\/i><\/button>/);
  assert.equal(userRunsHtml([]), '');
  assert.match(admin, /<span class="enruta-box" title="Carreras en camino"><b>\$\{runs\.length\}<\/b> en camino<\/span>/);
  assert.match(admin, /const run=e\.target\.closest\('\[data-open-camp\]'\);if\(run\)\{campReturn='usuarios';openEditor\(run\.dataset\.openCamp\)/);
  assert.match(admin, /tile\('🚚',riding\.length,'En ruta',`\$\{runs\} \$\{runs===1\?'carrera':'carreras'\}`\)/, 'runs en camino are counted in the «📊 En vivo» box');
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
  assert.match(runner, /await startSession\(\{accessCode\}\);await refresh\(\);loadTeam\(\);offerLive\(\);\}/, 'after signing in at the gate');
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
  assert.match(runner, /render\(\);loadTeam\(\);\$\('#gateForm'\)\.reset\(\);showGate\(\);await \(window\.AppNav\?AppNav\.logoutAll\(\):location\.assign\('\/'\)\)\}/, 'signing out clears them');
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
  assert.match(template, /'Compartida solo con tu equipo\.':'Solo visible en este dispositivo\.'/);
  assert.match(template, /document\.body\.classList\.toggle\('signed-in',!!u\);if\(!u\)\{stopShare\(false\);[^}]*\}setShareButton\(\);/, 'signing out stops sharing');
  assert.match(template, /p\.state==='granted'\?startShare\(true\):localStorage\.removeItem\('mapLive'\)/, 'a reload resumes only without a permission prompt (or notice)');
  const calls = template.match(/startShare\(/g).length;
  assert.equal(calls, 3, 'started only by the button, the reload resume, or its own definition');
  assert.doesNotMatch(template, /getCurrentPosition/, 'no automatic location request');
});
test('share a point link: public ?punto= opens its detail, runner ?punto= focuses it after sign-in', () => {
  assert.match(template, /<button type="button" class="footbtn" id="shareCamp" title="Compartir enlace del punto">🔗 Compartir<\/button>/);
  assert.match(template, /const campLink=r=>location\.origin\+'\/\?punto='\+encodeURIComponent\(r\.id\|\|r\.code\);/);
  assert.match(template, /renderMarkers\(\);renderFeed\(\);openLinkedCamp\(\)[;}]/, 'opened after the data loads');
  assert.match(template, /const want=linkedCamp,r=all\(\)\.find\(x=>x\.id===want\|\|x\.code===want\);linkedCamp=null;[^\n]*if\(r\)show\(r\);else \$\('#gpsMessage'\)\.textContent='Ese punto no está publicado o ya no existe\.'/, 'unpublished points are not shown');
  assert.match(runner, /data-share-camp="\$\{clean\(c\.id\)\}"[^>]*title="Compartir enlace" aria-label="Compartir enlace">\$\{shareIco\}<\/button>/, 'icon button');
  assert.match(runner, /shareLink\(location\.origin\+'\/runner\.html\?punto='\+encodeURIComponent\(b\.dataset\.shareCamp\)/);
  assert.match(runner, /available=data\.available\|\|\[\];render\(\);openLinkedCamp\(\)\}/, 'applied after sign-in loads the needs');
  assert.match(runner, /if\(!hasCoords\(n\)\)\{goToNeed\(n\.id\);return\}highlightNeed=n\.id;focusCamp=want;pendingFocus=true;/, 'flag on the map and the card highlighted');
  assert.match(runner, /n=all\.find\(x=>x\.id===linkedNeed&&x\.camp\.id===want\)\|\|all\.find\(x=>x\.camp\.id===want\)/, '&necesidad= picks that run; otherwise the point');
  assert.match(runner, /'\/runner\.html\?punto='\+encodeURIComponent\(b\.dataset\.shareCamp\)\+\(b\.dataset\.shareNeed\?'&necesidad='\+encodeURIComponent\(b\.dataset\.shareNeed\):''\)/, 'Compartir sends the run link');
  for (const page of [template, runner]) {
    assert.match(page, /async function shareLink\(url,title,btn\)\{if\(mobileDevice&&navigator\.share\)\{try\{await navigator\.share\(\{title,url\}\)\}catch\{\}return\}[^\n]*navigator\.clipboard\.writeText\(url\);(btn\.textContent='✓ Enlace copiado'|btn\.innerHTML='✓';btn\.title='Enlace copiado')/, 'share sheet on a phone, copy on a computer');
  }
});
test('narrow screens: ¡Ayuda! / + Agregar punto gets its own full-width row under the header', () => {
  assert.match(template, /function placeNav\(\)\{const nav=\$\('#appNav'\);if\(narrowNav\.matches\)\$\('\.addbar'\)\.prepend\(nav\);else \$\('\.top \.brand'\)\.after\(nav\)\}/, 'phone: page menu in the second row (hidden by the app bar); computer: next to the logo');
  assert.match(template, /<\/header>\n?<div class="addbar"><button type="button" class="add" id="addBtnBar">Reportar punto<\/button><\/div><nav class="mobile-nav"/);
  assert.match(template, /\.addbar\{display:none\}/, 'desktop keeps it in the header');
  assert.match(template, /@media\(max-width:760px\)\{body\{overflow:hidden\}\.addbar\{display:flex;gap:8px;height:54px;[^}]*\}\.addbar \.add\{flex:1;[^}]*\}[^@]*\.topright #addBtn,\.topright #runnerNav,\.topright #loginBtn\{display:none\}/);
  assert.match(template, /\.app\{height:calc\(calc\(100dvh \/ var\(--text-zoom,1\)\) - 166px\);/, 'header 62 + add row 54 + tabs 50 (screen height divided by the text zoom)');
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
  assert.match(admin, /<div class="listtools"><p id="count" class="muted"><\/p><label>Agregado por<select id="addedByFilter"><\/select><\/label><label>Ordenar<select id="sortBy"><option value="recent">Más recientes<\/option><option value="code">Código<\/option><option value="area">Área<\/option><option value="coord">Reportado por<\/option><\/select><\/label><div class="pendbulk" id="pendBulk" hidden><\/div><\/div>/);
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
  assert.match(card, /data-decision="archived">Archivar<\/button><\/div>`/);
  assert.match(admin, /if\(b\.dataset\.decision==='archived'&&!confirm\(b\.dataset\.reject\?rejectQuestion:archiveQuestion\)\)return;/, 'asks before archiving; Recuperar punto undoes it');
  assert.match(admin, /\$\('#archiveCamp'\)\.onclick=async e=>\{if\(!activeCamp\|\|!confirm\(archiveQuestion\)\)return;/);
});
test('public list: Ordenar by Más recientes (default), Más antiguos, Área or Coordinador, for everyone', () => {
  assert.match(template, /<label class="sortrow"><span class="needlabel">Orden<\/span><select id="sortList" autocomplete="off"><option value="attention">Atención<\/option><option value="recent">Más recientes<\/option><option value="oldest">Más antiguos<\/option><option value="area">Área<\/option><option value="coord">Reportado por<\/option><\/select><\/label>/);
  assert.doesNotMatch(template, /body:not\(\.signed-in\) [^{]*\.sortrow/, 'not limited to signed-in users');
  assert.match(template, /\.filterpanel \.sortrow\{display:none!important\}/, 'hidden for now; the list stays newest first');
  assert.match(template, /<button type="button" class="filterdone" id="filterDone">Listo<\/button><\/div><p id="gpsMessage"/, '«Listo» at the bottom of the panel closes it');
  assert.match(template, /\$\('#filterDone'\)\.onclick=\(\)=>\{setFilters\(false\);/);
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
  assert.match(admin, /<div class="adminmenu" id="adminMenu" role="menu" hidden>[^\n]*<button type="button" role="menuitemcheckbox" id="reviewToggle" aria-checked="true"><span class="mi" aria-hidden="true">[^<]*<\/span>Requerir revisión de puntos: <b>Sí<\/b><\/button>/);
  assert.match(admin, /const got=await adminRequest\('\/api\/admin\/settings'\);requireReviews=got\.requireReviews!==false/, 'read on load; Sí unless the server says otherwise');
  assert.match(admin, /if\(!next&&!confirm\('¿Publicar al momento los puntos que agregue el público\?/, 'turning it off asks first');
  assert.match(admin, /adminRequest\('\/api\/admin\/settings','POST',\{requireReviews:next\}\)/);
  assert.match(template, /saved\.status==='approved'\?'Punto publicado\. Ya aparece en el mapa y en la lista\.':'Se ha enviado tu solicitud\.'/, 'the public form already says which happened');
});
test('admin ☰ menu: «Descargar respaldo» downloads the whole DB; «Importar respaldo…» only on a local test server', () => {
  assert.match(admin, /<a role="menuitem" id="backupDownload" href="\/api\/admin\/backup" download><span class="mi" aria-hidden="true">[^<]*<\/span>Descargar respaldo<\/a><button type="button" role="menuitem" id="backupImport" hidden><span class="mi" aria-hidden="true">[^<]*<\/span>Importar respaldo…<\/button><input type="file" id="backupFile" accept="\.gz,application\/gzip" hidden>/);
  assert.match(admin, /\$\('#backupImport'\)\.hidden=!got\.importEnabled/, 'import shows only when the server allows it');
  assert.match(admin, /if\(!confirm\(`¿Importar «\$\{file\.name\}»\? Se reemplazarán TODOS los datos/, 'importing asks first');
  assert.match(admin, /fetch\('\/api\/admin\/restore',\{method:'POST',credentials:'same-origin',headers:\{'Content-Type':'application\/gzip','X-CSRF-Token':csrf\},body:file\}\)/);
});
test('«Solo Nuestro Equipo»: code users can limit the public list/map and runner needs to points added by a code user', () => {
  assert.match(template, /<button type="button" class="stat fstat teambtn" id="teamOnly" aria-pressed="false"><b id="teamCount">—<\/b><span>Solo nuestro equipo<\/span><\/button>/);
  assert.match(template, /body:not\(\.signed-in\) \.coordwrap,body:not\(\.signed-in\) \.teambtn\{display:none!important\}/, 'only with a code sign-in');
  assert.match(template, /const teamPoint=r=>!!r\.team;/, 'team comes from runners with a code (server-side)');
  assert.match(template, /const coordRows=\(\)=>all\(\)\.filter\(r=>\(!teamOnly\|\|teamPoint\(r\)\)&&/, 'filters the list, the map and the counts');
  assert.match(template, /if\(!u&&\(selectedCoordinator\|\|teamOnly\)\)\{selectedCoordinator='';teamOnly=false;/, 'signing out clears it');
  assert.match(runner, /body:not\(\.code-user\) \.rfilter\{display:none\}/, 'only for code users');
  assert.match(runner, /<nav class="mobile-nav" aria-label="Vista móvil"><button type="button" data-view="list" class="active" aria-pressed="true"><svg[^]*?<\/svg><span>Lista<\/span><\/button><button type="button" data-view="map" aria-pressed="false"><svg[^]*?<\/svg><span>Mapa<\/span><\/button><\/nav>/, 'phone tabs: icon + word, selected one announced');
  assert.match(runner, /\.mobile-nav\{display:flex;height:50px;gap:6px;padding:5px 10px;background:#e9f0f0;/, 'a segmented control on a soft track, same 50px height');
  assert.match(runner, /\.mobile-nav button\.active\{background:#173c46;border-color:#173c46;color:#fff;box-shadow:/, 'selected = dark pill');
  assert.match(runner, /b\.classList\.toggle\('active',on\);b\.setAttribute\('aria-pressed',on\)/);
  assert.doesNotMatch(runner, /en camino · \$\{shown\.length\} por llevar|Tus carreras <small>/, 'no 2/10 count line');
  assert.match(runner, /const listed=\(\)=>\(teamOnly\?available\.filter\(n=>n\.camp\.team\):available\)\.filter\(n=>\(agentMode\|\|!n\.requestedAt\)&&matches\(n\)\)/);
  assert.match(runner, /renderMap\(\[\.\.\.mine,\.\.\.shown\]\)/, 'your own runs always stay');
  assert.match(runner, /if\(!localStorage\.getItem\('runnerCode'\)\)teamOnly=false;/);
});
test('public narrow view: 📊 tab holds the totals, Lista keeps only filters; tapping your name copies your user ID', () => {
  assert.match(template, /<button type="button" data-view="map"><svg[^]*?<span>Mapa<\/span><\/button><button type="button" data-view="stats"><svg[^]*?<span>Resumen<\/span><\/button><\/nav>/, 'Lista, Mapa and Resumen: icon and name, like Runners');
  assert.match(template, /<div class="sidehead"><button type="button" class="filtertoggle" id="filterToggle" aria-expanded="false" aria-controls="filterPanel" aria-label="Filtros">[^]*?<\/button><div class="filterpanel" id="filterPanel"><div class="stats pointstats">/, 'filters fold into one summary line; the list gets the space');
  assert.match(template, /\.filterpanel\{display:none;padding-top:12px\}\.filterpanel\.open,\.mobile-stats \.filterpanel\{display:block\}/, 'open on tap; always open in the stats view');
  assert.match(template, /<\/select><\/label><button type="button" class="filterdone" id="filterDone">Listo<\/button><\/div><p id="gpsMessage"/, 'messages stay visible outside the folded panel');
  assert.match(template, /selectedCoordinator\?'Reportado por '\+coordLabel\(\):''/, 'the summary names the coordinator only when one is picked');
  assert.match(template, /\.filter\(Boolean\)\.join\(' · '\)\+` \(\$\{n\}\)`\}/, 'summary ends with the count in parentheses, no «Filtros» or «puntos» words');
  assert.match(template, /aria-controls="filterPanel" aria-label="Filtros"><svg class="ftitle"/);
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
test('public detail for code users: each open/en camino need can get a conductor and packing, like admin', () => {
  assert.match(template, /\$\{r\.id&&codeMe\(\)&&!phoneLocked\(\)\?'<div id="teamNeeds"><\/div>':''\}/, 'only with a code sign-in');
  const draw = template.match(/function drawTeamNeeds\(\)[^\n]*/)[0];
  assert.match(draw, /filter\(n=>n\.status==='OPEN'\|\|n\.status==='IN_PROGRESS'\)/);
  assert.match(draw, /Conductor: <b>\$\{n\.runnerName\?clean\(n\.runnerName\):'sin asignar'\}<\/b>/);
  assert.match(draw, /PointForm\.packingHtml\(n\.packing\)/);
  assert.match(draw, /data-assign-need="\$\{clean\(n\.id\)\}" aria-haspopup="listbox">🚚 \$\{n\.runnerName\?'Agregar otro conductor':'Asignar conductor'\}<\/button><button type="button" data-pack-need="\$\{clean\(n\.id\)\}">📦 Agregar a lista de empaque<\/button>/);
  assert.match(template, /<div class="picker assignpick" id="assignPicker" hidden><input id="assignSearch" type="search" placeholder="Buscar conductor…"/);
  assert.ok(template.indexOf('id="assignPicker"') < template.indexOf('<script>'), 'markup exists before the script binds it');
  assert.match(template, /crewCache=\{at:Date\.now\(\),list:await fullGet\('\/api\/full\/crew'\)\}/, 'the server lists code users first');
  assert.match(template, /c\.team\?'<small>Equipo<\/small>':''/);
  assert.match(template, /fullJSON\('\/api\/full\/needs\/'\+encodeURIComponent\(need\)\+'\/runner',\{runnerId:id\}\)/);
  assert.match(template, /PointForm\.openPacking\(\{subtitle:[^\n]*fullJSON\('\/api\/full\/needs\/'\+encodeURIComponent\(id\)\+'\/packing',\{items:lines\}\)/, 'same packing dialog as admin and runners');
  assert.match(template, /\.length;drawFilterSummary\(rows\.length\);drawTeamNeeds\(\)\}/, 'refreshed with the list and after each save');
});
test('public detail: with needs to fulfil, the big buttons are Asignar conductor and packing; Reportar becomes a small link at the bottom', () => {
  assert.match(template, /<button type="button" class="footbtn reportsmall" data-report-need title="Reportar otra necesidad">➕ Reportar<\/button>/);
  assert.match(template, /\$\('#detailBody'\)\.classList\.toggle\('has-team-needs',list\.length>0\)/);
  assert.match(template, /\.has-team-needs \.reportneed\{display:none\}\.has-team-needs \.reportsmall\{display:inline-block\}/);
  assert.match(template, /\.teamacts button\{[^}]*padding:11px 12px;font:800 14px[^}]*background:#ed7654\}/);
  assert.match(template, /if\(!e\.target\.closest\('#reportNeed,\[data-report-need\]'\)\)return;/, 'both open the same form');
});
test('public detail: code users can remove the conductor (the need reopens)', () => {
  assert.match(template, /\$\{n\.runnerName&&canRun\(\)\?` · <button type="button" class="unassign" data-unassign-need="\$\{clean\(n\.id\)\}">Quitar conductor<\/button>`:''\}/);
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
  assert.match(template, /\$\{r\.id&&codeMe\(\)&&!phoneLocked\(\)\?'<div id="teamNeeds"><\/div>':''\}/);
  assert.match(template, /if\(!u\)\{stopShare\(false\);[^}]*campNeeds\.clear\(\);\$\('#teamNeeds'\)\?\.remove\(\);/);
  const draw = template.match(/function drawTeamNeeds\(\)[^\n]*/)[0];
  assert.match(draw, /^function drawTeamNeeds\(\)\{const box=\$\('#teamNeeds'\);if\(!box\)\{/, 'draws nothing without the team section');
  assert.doesNotMatch(template.match(/function serviceNeedsHtml[^\n]*/)[0], /runnerName|packing|media/, 'the public need list never shows them');
});

test('Usuarios › Equipos: create/rename teams; a two-column table assigns each user with a team type-ahead, select all, and batch actions', () => {
  assert.match(admin, /<details id="teamsBox" class="teams-box" open><summary>👥 Equipos <small id="teamsCount"><\/small><\/summary>/, 'shown as the «👥 Equipos» tab');
  assert.match(admin, /<form id="teamNewForm" class="team-new"><input name="name" maxlength="80" required/);
  assert.match(admin, /#operations\.user-form-mode #teamsBox,#operations\.user-view-mode #teamsBox,#operations\.user-form-mode #userBatch,#operations\.user-view-mode #userBatch\{display:none!important\}/, 'only on the list');
  assert.match(admin, /<label>Equipo<select name="teamId"><option value="">Sin equipo<\/option><\/select><\/label>/);
  assert.match(admin, /const data=\{[^}]*roles,teamId:form\.elements\.teamId\.value\};/, 'saved with the user');
  assert.doesNotMatch(admin, /team-bulk|Asignar en lote|id="teamUsers"|id="teamBatch"/, 'no bulk «Asignar en lote» table: teams are managed from each team card');
  assert.match(admin, /!\$\('#teamsBox'\)\.contains\(document\.activeElement\)/, 'the 30 s refresh never interrupts typing a team');
  const fold = s => String(s).normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase();
  const clean = s => String(s ?? '').replace(/[&<>"']/g, c => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
  const src = ['drawTeams', 'teamMatches'].map(f => admin.match(new RegExp(`function ${f}\\(\\)?[^\\n]*`))[0]).concat([admin.match(/const tcAdd=[^\n]*/)[0], admin.match(/const codeCache=[^\n]*/)[0], admin.match(/const codeTag=[^\n]*/)[0], admin.match(/function teamCard\(t,only=null\)\{[^]*?\n(?=\/\/ Buscar en «Equipos»)/)[0], admin.match(/const teamByName=[^\n]*/)[0]]).join('\n');
  const els = {}; const el = k => els[k] ??= {value: ''};
  const teams = [{id: 't1', name: 'Norte <b>'}, {id: 't2', name: 'Peñas'}];
  const users = [{id: 'u1', name: 'Fufo', teamId: 't1', hasAccessCode: true, roles: ['runner']}, {id: 'u2', name: 'Bea', hasAccessCode: true, roles: ['agent']}, {id: 'u3', name: 'Ana', hasAccessCode: false, roles: ['runner']}];
  const api = new Function('$', 'teams', 'users', 'renamingTeam', 'clean', 'fold', 'teamName', 'roleNames', 'document', `const roleLabel=r=>roleNames[r]||r;${src};return {drawTeams, teamByName}`)(el, teams, users, null, clean, fold, id => teams.find(t => t.id === id)?.name || '', {runner: 'Runner', agent: 'Agente'}, {activeElement: null});
  api.drawTeams();
  assert.match(el('#teamNames').innerHTML, /<option value="Norte &lt;b&gt;"><\/option><option value="Peñas">/, 'type-ahead list of teams');
  assert.match(el('#teamList').innerHTML, /<b>👥 Norte &lt;b&gt;<\/b><small>1 miembro<\/small>/);
  assert.equal(api.teamByName(' penas '), 't2', 'typed names match without case or accents');
  assert.equal(api.teamByName(''), '', 'empty = no team');
  assert.equal(api.teamByName('Nada'), null);
});
test('public assign picker tags teammates with «Equipo»; the server lists only own team and runners without a code', () => {
  assert.match(template, /c\.team\?'<small>Equipo<\/small>':''/);
});

test('Usuarios: checkboxes per user; selected users that no longer match the search stay below the results; batch move/deactivate', () => {
  assert.doesNotMatch(admin, /id="userSelAll"/, 'select-all is a button over the results, not a filter checkbox');
  assert.match(admin, /<div class="user-batch" id="userBatch" hidden><b id="userSelCount"><\/b><span class="teampop-wrap"><button type="button" id="userBatchMove" aria-haspopup="listbox" aria-expanded="false">👥 Mover a equipo ▾<\/button><div class="teampop" id="teamPop" hidden><input id="teamPopSearch" type="search" placeholder="Buscar o crear equipo…"[^>]*role="combobox"[^>]*><div id="teamPopList" role="listbox"[^>]*><\/div><\/div><\/span><button type="button" class="quiet" id="userBatchUnteam">Quitar del equipo<\/button><button type="button" class="quiet-red" id="userBatchOff">Desactivar<\/button><button type="button" class="quiet" id="userBatchOn">Activar<\/button><button type="button" class="linkbtn" id="userBatchClear">Limpiar selección<\/button><\/div>/);
  assert.match(admin, /\.user-batch\{position:sticky;top:0/, 'the batch bar stays in view while scrolling');
  const src = [admin.match(/const userSel=[^\n]*/)[0], "let userTeamFilter='';const teams=[];", admin.match(/let userSortKey=[^\n]*/)[0], admin.match(/const ROLE_FILTERS=[^\n]*/)[0], admin.match(/const sortBtn=[^\n]*/)[0], admin.match(/const roleHead=[^\n]*/)[0], admin.match(/const userHead=[^\n]*/)[0], admin.match(/const userCard=[^\n]*/)[0], ...['userResults', 'drawUsers', 'syncUserSel', 'batchUsers'].map(f => admin.match(new RegExp(`(async )?function ${f}\\([^\\n]*`))[0])].join('\n');
  const els = {}; const el = k => els[k] ??= {value: '', checked: false, querySelectorAll: () => []};
  el('#userRole').value = ''; el('#userTeam').value = ''; el('#userSort').value = 'name';
  const users = [{id: 'u1', name: 'Fufo', roles: ['runner'], hasAccessCode: true}, {id: 'u2', name: 'Bea', roles: ['agent'], hasAccessCode: true, teamId: 't1'}, {id: 'u3', name: 'Ana', roles: ['runner'], hasAccessCode: false}];
  const patched = [];
  const fold = s => String(s).toLowerCase();
  const api = new Function('$', 'users', 'fold', 'clean', 'userRuns', 'userRunsHtml', 'keyBtn', 'phoneBtn', 'teamChip', 'roleNames', 'stamp', 'patchUsers', `const roleLabel=r=>roleNames[r]||r;const drawUserStats=()=>{};${src};return {userSel, drawUsers, batchUsers}`)(el, users, fold, s => String(s ?? ''), () => [], () => '', () => '', () => '', () => '', {}, s => s, async (ids, body) => { patched.push([ids, body]); return ids.includes('u3') ? [{id: 'u3', text: 'Ana: no'}] : []; });
  api.drawUsers();
  assert.equal(el('#userBatch').hidden, true);
  assert.doesNotMatch(el('#userList').innerHTML, /users-chosen/, 'no Seleccionados group without a selection');
  assert.match(el('#userList').innerHTML, /<input type="checkbox" class="user-sel" data-user-sel="u1" aria-label="Seleccionar a Fufo">/);
  api.userSel.add('u1'); api.userSel.add('u3');
  el('#userSearch').value = 'bea'; api.drawUsers();
  let html = el('#userList').innerHTML;
  const chosen = html.indexOf('<details class="users-group users-chosen"><summary><input type="checkbox" class="selsum-box" checked data-unsel-all aria-label="Deseleccionar todos" title="Deseleccionar todos"><span class="selsum-t">Seleccionados (2)</span><span class="selsum-chev" aria-hidden="true">▾</span></summary>'), results = html.indexOf('<div class="users-group users-results"><h3>Resultados (1)');
  assert.ok(chosen >= 0 && results > chosen, 'collapsible Seleccionados on top, Resultados (not collapsible) below');
  assert.ok(html.indexOf('Seleccionar a Fufo') < results && html.indexOf('Seleccionar a Ana') < results, 'every selected user is grouped on top, even outside the search');
  assert.ok(html.indexOf('Seleccionar a Bea') > results);
  assert.match(html, /user-card sel"><input type="checkbox" class="user-sel" data-user-sel="u1" aria-label="Seleccionar a Fufo" checked>/);
  assert.match(html, /<span class="headsel"><label class="hsbox"><input type="checkbox" data-sel-results aria-label="Seleccionar el resultado"[^>]*><span class="hsl">Todos<\/span><\/label><\/span>/, 'select-all is a checkbox in the column header');
  assert.match(html, /<details class="users-group users-chosen"><summary><input type="checkbox" class="selsum-box" checked data-unsel-all[^>]*><span class="selsum-t">Seleccionados \(2\)<\/span><span class="selsum-chev" aria-hidden="true">▾<\/span><\/summary>/, '2 or more selected: collapsed; the checked box in the title clears them; arrow on the right');
  assert.doesNotMatch(html, /class="sel-all"/, 'no select-all buttons');
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
  assert.match(admin, /addEventListener\('click',e=>\{const sm=e\.target\.closest\('\.users-chosen>summary'\);if\(sm&&!e\.target\.closest\('input,button'\)\)selOpen=!sm\.parentElement\.open\},true\)/, 'collapsed state survives redraws');
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
  const head = new Function('clean', `let userTeamFilter='t1';const teams=[{id:'t1',name:'Norte'}];${admin.match(/let userSortKey=[^\n]*/)[0]};${admin.match(/const ROLE_FILTERS=[^\n]*/)[0]};${admin.match(/const sortBtn=[^\n]*/)[0]};${admin.match(/const roleHead=[^\n]*/)[0]};${admin.match(/const userHead=[^\n]*/)[0]};return userHead`)(s => String(s));
  const plain = head().replace(/<button[^>]*data-usort="(\w+)"[^>]*>([^<]*)<span class="arrow"[^>]*>([^<]*)<\/span><\/button>/g, '[$1:$2$3]');
  assert.equal(plain, '<div class="user-head"><span class="headsel"></span><span>[name:Usuario]</span><span class="rolehead">[roles:Roles]</span><span>Equipo</span><span>[activity:Actividad]</span><span>[created:Creado▼]</span><span></span></div>', 'sortable headers; newest first by default; Seleccionados has no filters');
  assert.match(head(true), /^<div class="user-head with-filters">[^]*<span><select id="userTeam" class="head-filter on" aria-label="Filtrar por equipo"><option value="">Equipo: todos<\/option><option value="t1" selected>👥 Norte<\/option><option value="__none__">Sin equipo<\/option><\/select><\/span><span><button type="button" class="sorth" data-usort="activity"/, 'Resultados: team filter in the Equipo column, highlighted when on');
  assert.doesNotMatch(admin, /<div class="userbar">[^]*?id="userTeam"[^]*?<div id="userList">/, 'no longer in the top filter bar');
  assert.match(admin, /\.user-head\.with-filters\{display:flex!important;flex-wrap:wrap;gap:6px 12px;padding:4px 0\}/, 'phones keep the sort buttons and filters');
  assert.match(head(true), /<button type="button" class="head-filter rolefilter" data-rolefilter aria-expanded="false" aria-label="Filtrar por rol">▾<\/button><div class="rolepop" id="rolePop" hidden><label class="check"><input type="checkbox" data-rolecheck="runner"> 🏃 Runner<\/label>/, 'Roles: checkbox filter, several at once');
  assert.doesNotMatch(admin, /id="userRole"|id="userSort"/, 'no more top filters that repeat the headers');
  assert.match(admin, /\.user-card,\.user-head\{display:grid;grid-template-columns:22px minmax\(190px,2\.2fr\)[^}]*\}/, 'rows and header share one column grid');
  assert.match(admin, /@media\(max-width:720px\)\{\.user-head\{display:none\}[^}]*\}[^@]*\.uc-col:before\{content:attr\(data-label\)/, 'phones: one labeled line per column');
  assert.match(admin, /<\/summary>\$\{userHead\(\)\}\$\{chosen\.map\(userCard\)/);
  assert.match(admin, /\$\{users\.length\?userHead\(true,rest\.length\?`[^`]*`:''\):''\}\$\{rest\.length\?rest\.map\(userCard\)/, 'the filter header stays even with no results, so it can be cleared');
  const card = new Function('userRuns', 'userSel', 'clean', 'keyBtn', 'phoneBtn', 'roleNames', 'teamChip', 'stamp', 'userRunsHtml', `const roleLabel=r=>roleNames[r]||r;${admin.match(/const userCard=[^\n]*/)[0]};return userCard`)(() => [], new Set(), s => String(s ?? ''), () => '', () => '', {runner: 'Corredor'}, u => u.teamId ? '<span class="team-chip">👥 Norte</span>' : '', s => s, () => '');
  const a = card({id: 'u1abcdefgh', name: 'Fufo', roles: ['runner'], teamId: 't1', online: true, createdAt: '1 oct'});
  const b = card({id: 'u2abcdefgh', name: 'Ana', roles: [], disabled: true});
  const cols = h => [...h.matchAll(/<div class="uc-col" data-label="([^"]+)">/g)].map(m => m[1]);
  assert.deepEqual(cols(a), ['Roles', 'Equipo', 'Actividad', 'Creado']);
  assert.deepEqual(cols(b), cols(a), 'same columns for every user, so they line up');
  assert.match(a, /data-label="Equipo"><span class="team-chip">👥 Norte<\/span>/);
  assert.match(b, /data-label="Roles"><span class="role-chip pending"[^>]*>Pendiente<\/span>/);
  assert.match(b, /data-label="Equipo"><\/div><div class="uc-col" data-label="Actividad"><small class="user-active"><\/small><\/div><div class="uc-col" data-label="Creado"><\/div>/, 'empty values stay blank, no «Sin …» text');
  assert.match(b, /<b>Ana<\/b><span class="role-chip off">Desactivado<\/span>/);
});
test('user view: «Recorrido GPS» map with period picker and a timelapse player (play, drag, speed, skip pauses)', () => {
  assert.match(admin, /<section id="userTrack" class="need-card track"><div class="track-head"><h3>📍 Recorrido GPS<\/h3><select id="trackRange"/);
  assert.match(admin, /<button type="button" id="trackToggle" aria-label="Reproducir">▶<\/button><input type="range" id="trackSlider" min="0" max="1000"[^>]*><select id="trackSpeed"[^>]*><option value="10">10×<\/option><option value="60" selected>1 min\/s<\/option>/);
  assert.match(admin, /#operations:not\(\.user-view-mode\) #userTrack\{display:none!important\}/, 'only in the user view');
  assert.match(admin, /userPage\(u\.name,'← Usuarios'\);drawUserView\(\);loadTrack\(u\.id\);/, 'loads when a user is opened');
  assert.match(admin, /function closeUserPage\(navigate=true\)\{stopTrack\(\);/, 'stops playback when leaving');
  assert.match(admin, /\/api\/admin\/users\/\$\{encodeURIComponent\(id\)\}\/track\?from=/);
  const src = [admin.match(/const TRACK_GAP=[^\n]*/)[0], admin.match(/function trackTimeline\(\)[^\n]*/)[0], admin.match(/function trackAt\(t\)[^\n]*/)[0], admin.match(/function trackSpan\(v\)[^\n]*/)[0]].join('\n');
  const skip = {checked: true};
  const api = new Function('$', `${src};return {set:p=>{trackPts=p},timeline:()=>{trackTimeline();return trackPlayT},trackAt,trackSpan}`)(() => skip);
  const t0 = Date.parse('2026-10-14T12:00:00Z');
  api.set([{lat: 7, lon: -80, ms: t0}, {lat: 7.01, lon: -80, ms: t0 + 30e3}, {lat: 7.01, lon: -80, ms: t0 + 3600e3}, {lat: 7.02, lon: -80, ms: t0 + 3630e3}]);
  assert.deepEqual(api.timeline(), [0, 30e3, 90e3, 120e3], 'skipping pauses caps idle time at 1 min');
  let at = api.trackAt(15e3);
  assert.equal(at.i, 0); assert.ok(Math.abs(at.lat - 7.005) < 1e-9, 'interpolates between points'); assert.equal(at.ms, t0 + 15e3);
  at = api.trackAt(60e3);
  assert.equal(at.lat, 7.01, 'no interpolation across a gap of more than 10 min');
  assert.equal(api.trackAt(10e6).lat, 7.02, 'clamps to the end');
  skip.checked = false;
  assert.deepEqual(api.timeline(), [0, 30e3, 3600e3, 3630e3], 'real time without skipping');
  const [from, to] = api.trackSpan('d1');
  assert.equal(to - from, 24 * 3600e3, '«Ayer» is a full local day');
  assert.equal(from.getHours(), 0);
  const clock = new Function(`${admin.match(/const stamp=[^\n]*/)[0]};${admin.match(/const trackClockText=[^\n]*/)[0]};return trackClockText`)();
  assert.equal(clock(new Date(2026, 9, 13, 19, 33, 7).getTime()), 'Oct 13, 2026 7:33:07pm', 'player clock uses the app format, with seconds');
});

test('location sharing asks for the location notice (with a required «Acepto» checkbox) before it starts, on both pages', () => {
  const pf = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
  assert.match(pf, /<h2 id="pfLocNoticeTitle">📍 Tu ubicación, por tu seguridad<\/h2><p class="pf-ln-text">Estás en una zona de desastre\. Compartir tu ubicación en vivo puede salvarte[^<]*<\/p><label class="pf-ln-ok"><input type="checkbox" name="ok" required> Por mi seguridad, acepto que mi ubicación se use y se comparta mientras esté activa\.<\/label><div class="pf-ln-actions">/, 'safety wording, one checkbox');
  assert.match(readFileSync(new URL('./location.go', import.meta.url), 'utf8'), /const locationNoticeVersion = "2026-10-15"/, 'new text → everyone accepts it again');
  assert.match(pf, /<button type="submit" value="yes" disabled>Aceptar y activar<\/button>/, 'disabled until ticked');
  assert.match(pf, /resolve\(d\.returnValue === 'yes' && ok\.checked\)/);
  // Runner panel: notice → saved on the server → then the GPS starts; 428 turns it off.
  assert.match(runner, /async function startLive\(\)\{if\(!map\)return false;if\(!agentMode&&!locConsent\)\{if\(!await PointForm\.locationNotice\(\)\)return false;try\{await api\('\/api\/runner\/location\/consent','POST',\{accept:true\}\);locConsent=true\}/);
  assert.match(runner, /if\(err\.status===428\)\{locConsent=false;stopLive\(\);/);
  assert.match(runner, /function resumeLive\(\)\{[^}]*if\(locConsent\|\|agentMode\)startLiveNow\(\);else localStorage\.removeItem\('runnerLive'\)\}\}/, 'never resumes without the notice');
  assert.match(runner, /if\(!locConsent&&!denied\)\{if\(!await startLive\(\)\)sessionStorage\.setItem\('runnerLiveAsked','1'\);return\}/, 'the login offer goes straight to the notice');
  // Public map (code users).
  assert.match(template, /async function locationAccepted\(quiet\)\{try\{if\(\(await fullGet\('\/api\/full\/me'\)\)\.locationConsent\)return true\}catch\{return false\}if\(quiet\|\|!await PointForm\.locationNotice\(\)\)return false;try\{await fullJSON\('\/api\/full\/location\/consent',\{accept:true\}\)/);
  assert.match(template, /async function startShare\(quiet=false\)\{[^}]*if\(!await locationAccepted\(quiet\)\)/);
  assert.match(template, /else if\(r\.status===428\)\{stopShare\(false\);/);
  assert.match(admin, /field\('Aviso de ubicación',u\.locationConsentAt\?'✅ Aceptado '\+clean\(stamp\(u\.locationConsentAt\)\):'No aceptado: no puede compartir su ubicación en vivo'\)/, 'admin sees whether and when it was accepted');
});
test('shared phone widget (nav.js): tapping a number opens «📞 Llamar / 💬 WhatsApp / 📋 Copiar», the same on every page', () => {
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  const grab = re => nav.match(re)[0];
  const phone = new Function(`${grab(/const esc = [^\n]*/)}\n${grab(/const waNumber = [^\n]*/)}\n${grab(/function phone\(raw, name = ''\) \{[^]*?\n  \}/)};return phone`)();
  assert.equal(phone('6555-1234', 'Ana'), '<button type="button" class="telnum" data-tel-pop="6555-1234" data-tel-name="Ana" title="Llamar o escribir por WhatsApp">6555-1234</button>');
  assert.equal(phone(''), '');
  assert.match(nav, /<a class="ntp-call" href="tel:\$\{esc\(d\)\}">📞 Llamar<\/a><a class="ntp-wa" href="https:\/\/wa\.me\/\$\{waNumber\(p\)\}"[^>]*>💬 WhatsApp<\/a><\/div><button type="button" class="ntp-copy" data-copy-tel="\$\{esc\(p\)\}">📋 Copiar<\/button>/);
  assert.match(nav, /const waNumber = t => \{ const d = String\(t\)\.replace\(\/\\D\/g, ''\); return String\(t\)\.trim\(\)\.startsWith\('\+'\) \|\| d\.length > 8 \? d : '507' \+ d; \};/, 'Panama code when missing');
  assert.match(nav, /navigator\.clipboard\.writeText\(cp\.dataset\.copyTel\)/);
  assert.match(nav, /html:has\(dialog\[open\]\),html:has\(\.modalbg\.open\)\{overflow:hidden\}/, 'the page behind an open pop-up does not scroll');
  assert.match(nav, /Object\.assign\(b\.style, \{position: 'fixed', top: `-\$\{locked\.y\}px`, left: '0', right: '0', width: '100%'\}\);/, 'iPhone: the body is pinned, then restored to the same spot');
  for (const page of ['runner.html', 'builder.html', 'camp.html', 'command.html']) assert.doesNotMatch(readFileSync(new URL('./' + page, import.meta.url), 'utf8'), /<a href="tel:/, page + ' uses the shared widget');
  assert.match(readFileSync(new URL('./camp.html', import.meta.url), 'utf8'), /<script src="\.\/nav\.js"><\/script><script>/);
});
test('shared top navigation (nav.js): Inicio and the panels of the signed-in code user, current page marked', () => {
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  for (const page of [template, runner, agent]) assert.match(page, /<script src="\.\/nav\.js"><\/script>/);
  assert.match(readFileSync(new URL('./builder.html', import.meta.url), 'utf8'), /<script src="\.\/nav\.js"><\/script>/);
  assert.match(nav, /\['\/', '🏠', \(\) => 'Inicio', \(\) => true\]/, 'the public map page is «Inicio»');
  assert.match(nav, /\['\/runner\.html', '<svg class="walkico"[^']*<\/svg>', u => !u\.roles \|\| u\.roles\.includes\('runner'\) \? 'Runners' : 'Empaque'/, 'the runner page is «Runners» with the blue en-camino walking icon');
  assert.match(nav, /\['\/command\.html', '📞', \(\) => 'Checkup', u => has\(u, 'agent'\)\]/, 'Command: the HQ where data is checked and managed at a computer');
  assert.match(nav, /let list = u \? pages\.filter\(\(\[url, , , ok\]\) => url === here \|\| ok\(u\)\) : \[\];\n[^]*?\n    el\.hidden = list\.length < 2 && !admin;/, 'nothing for the public (unless this browser has an admin session)');
  // Narrow screens: one toggle with the current page opens the full list under the header; outside tap / Escape closes it.
  assert.match(nav, /<button type="button" class="appnav-toggle" aria-expanded="false" aria-controls="appNavList" aria-label="Cambiar de página">/);
  assert.match(nav, /@media\(max-width:760px\)\{\.top\{position:relative\}\.appnav\{position:static\}\n\.appnav-toggle\{display:inline-flex/);
  assert.match(nav, /\.appnav-list\{display:none;position:absolute;top:100%;left:0;right:0;z-index:3000;/);
  assert.match(nav, /else if \(!el\.contains\(e\.target\) \|\| e\.target\.closest\('\.appnav-list a'\)\) setOpen\(false\);/);
  assert.match(nav, /if \(e\.key === 'Escape'\) setOpen\(false\);/);
  // Text size (accessibility): «Aa» next to the logo opens A− / A+; the map is compensated so it keeps working.
  assert.match(nav, /<button type="button" id="textSizeBtn" aria-expanded="false" aria-controls="textSizePop" title="Tamaño del texto"/);
  assert.match(nav, /const sizes = \[0\.9, 1, 1\.15, 1\.3, 1\.5\];/);
  assert.match(nav, /html\{zoom:var\(--text-zoom,1\)\}\.leaflet-container\{zoom:calc\(1 \/ var\(--text-zoom,1\)\)\}/);
  assert.doesNotMatch(nav, /data-theme|themeBtn|invert\(/, 'no dark mode');
  // ✕ next to the name asks first, in the pages' own Sí / No window.
  assert.match(nav, /function confirm\(text, \{yes = 'Sí', no = 'No', detail = ''\} = \{\}\)/);
  assert.match(nav, /resolve\(d\.returnValue === 'yes'\)/);
  const ask = /if\(window\.AppNav&&!await AppNav\.confirm\('¿Cerrar sesión\?',\{yes:'Sí, salir'/;
  for (const page of [template, runner, agent]) assert.match(page, ask);
  assert.match(nav, /url === here \? ' aria-current="page"' : ''/);
});
test('command center (/command.html): admin-style rows, corner map, en-ruta monitor, view mode inside the row', () => {
  // Headers: just the logo, no «AZUERO / …» titles, on every page.
  for (const page of [template, runner, agent, readFileSync(new URL('./builder.html', import.meta.url), 'utf8'), admin]) assert.doesNotMatch(page.match(/<header[^]*?<\/header>/)[0], /AZUERO|topsmall/);
  assert.match(agent, /<header class="top"><div class="brand"><span class="logo" title="Azuero Apoyo">✚<\/span><\/div><nav class="appnav" id="appNav" hidden><\/nav><input class="topsearch" id="search"[^>]*>/);
  assert.match(agent, /<main class="board"><div class="boardmain">[^]*?<div class="feed" id="feed">[^]*?<aside class="mapcorner"><section class="mapwrap"><div id="map"/, 'rows first, small map in the top corner');
  assert.match(agent, /\.mapcorner\{position:sticky;top:84px\}/);
  assert.match(agent, /\.top\{position:sticky;top:0;z-index:1100\}/, 'header stays on top while scrolling (desktop)');
  assert.match(agent, /\.row,\[data-campform\]\{scroll-margin-top:84px\}/, 'an opened row stops below the header, not under it');
  assert.match(agent, /@media\(max-width:760px\)\{\.top\{position:static\}/, 'phone: the header scrolls with the page');
  assert.doesNotMatch(agent, /hdr-hidden/);
  assert.doesNotMatch(agent, /class="mobile-nav"/, 'no list/map toggle: the map sits above the rows on phones');
  assert.match(agent, /<form id="gateForm" novalidate data-signin><h1>📞 Checkup · centro de datos<\/h1><fieldset class="sg sg-code">/);
  assert.doesNotMatch(agent.match(/<form id="gateForm"[^]*?<\/form>/)[0], /sg-new/, 'agents are created by admin: code only');
  assert.match(agent, /fetch\('\/api\/session\/code',\{method:'POST'[^\n]*?body:JSON\.stringify\(\{accessCode\}\)[^\n]*?if\(!\(body\.roles\|\|\[\]\)\.includes\('agent'\)\)throw Error\('Tu código no tiene acceso al Centro de comando\.'\)/, 'one code login opens every panel of the user');
  assert.match(agent, /localStorage\.getItem\('agentCode'\)\|\|localStorage\.getItem\('runnerCode'\)\|\|localStorage\.getItem\('builderCode'\)/, 're-entry with a code saved by any page');
  // Header: shared nav, name copies the user ID with a tooltip, no ID footer.
  assert.match(agent, /<input class="topsearch" id="search"[^>]*><div class="topright"><div class="who" id="who" hidden><button type="button" id="whoName" title="Copiar mi ID de usuario"><\/button><span class="idtip" id="idTip" role="status" hidden><\/span>/);
  assert.match(agent, /await navigator\.clipboard\.writeText\(userId\);text='ID copiado: '\+userId\.slice\(0,8\)\+'…'/);
  assert.doesNotMatch(agent, /id="idFoot"|← Mapa público/);
  assert.match(agent, /api\('\/api\/agent\/points'\),api\('\/api\/agent\/needs'\)/);
  assert.doesNotMatch(agent, /\/api\/(admin|runner)\//, 'only the agent API');
  assert.match(agent, /points\.filter\(p=>\['approved','pending'\]\.includes\(p\.status\)/, 'approved and pending points, no archived or duplicates');
  assert.match(agent, /<article class="row point\$\{pending\?' pending':''\}\$\{open\?' open':''\}\$\{open&&mode==='camp'\?' editing':''\}\$\{p\.id===focusId\?' focus':''\}" data-id="\$\{clean\(p\.id\)\}"/);
  assert.match(agent, /\$\{pending\?' <span class="status pending">Por revisar<\/span>':''\}\$\{p\.toLocate\?' <span class="status tolocate"[^`]*?📍 Por ubicar<\/span>':''\}<\/h2>/, 'only «Por revisar» is tagged; approved rows carry no badge');
  assert.doesNotMatch(agent, />Aprobado<\/span>/);
  assert.match(agent, /\$\{open\?pointView\(p\):active\.length\?`<ul class="needlist rowneeds">/, 'open/en-route needs on the row, the view mode when expanded');
  assert.match(agent, /pending&&!open\?`<div class="fields">/, 'pending points show their full data in the row');
  // Abiertas and En ruta stand out with flag / walking icons; total and completed are small.
  // Collapsed cards: each pending need has «Actualizar», and «+ Reportar necesidad» sits under the needs.
  assert.match(agent, /\$\{actions\?`<button type="button" class="act nupd" data-act="editneed" data-need="\$\{clean\(n\.id\)\}">\$\{icoEdit\}Actualizar<\/button>`:''\}<\/div>/, '«Actualizar» on the need\'s top line, aligned right');
  assert.match(agent, /if\(e\.target\.closest\('\[data-onmap\]'\)\)\{showOnMap\(id\);return\}\};/, 'the card itself does nothing; «📍 Ver en mapa» shows it on the map');
  assert.match(agent, /\$\{hasCoords\(p\)\?`<button type="button" class="act mapbtn" data-onmap>\$\{typeof PointForm!=='undefined'\?PointForm\.icons\.pin:''\}Ver en mapa<\/button>`:`<span class="nogpsmark" title="Sin GPS: no aparece en el mapa"/, 'no GPS: red crossed-out pin');
  assert.match(agent, /\.row\{cursor:default;user-select:none;/);
  assert.match(agent, /function showOnMap\(id\)\{const p=points\.find\(x=>x\.id===id\);focusPoint\(id\);if\(narrowCmd\.matches&&hasCoords\(p\)\)showCView\('overview'\)\}/);
  assert.match(agent, /<button type="button" class="act detbtn" data-details>Detalles<\/button>/, '«Detalles» opens it');
  assert.match(agent, /if\(seen&&seen>\(last\|\|''\)\)return rel\('Checkup',seen\)\+checkedBy\(p\);return last\?rel\('Suplido',last\)/, 'top line: «Checkup hace …» with who did it; the creator only in «Creado por»');
  assert.match(agent, /repair=list\.some\(n=>n\.kind==='repair'&&n\.status!=='COMPLETED'\),active=list\.filter\(n=>n\.status!=='COMPLETED'&&n\.kind!=='repair'\)/, 'home repairs are not in the card\'s needs');
  assert.match(agent, /Necesita reparación estructural<\/span><\/p>/, 'shown as a detail with the repair icon');
  assert.match(agent, /const openNeeds=p\?needsOf\(p\)\.filter\(n=>n\.status==='OPEN'&&n\.kind!=='repair'\):\[\];/, 'a checkup never removes a repair');
  assert.match(agent, /function checkupLines\(p\)\{const notes=p\.checkupLog\|\|\[\],last=\(p\.checkups\|\|\[\]\)\.slice\(-3\)\.reverse\(\);/, 'the last 3 checkups on the card');
  assert.match(agent, /\$\{!open&&!pending\?checkupLines\(p\):''\}<\/article>`\}/, 'checkup log at the bottom of the card');
  assert.match(agent, /const CK_SHORT=\{'Sí, están bien':'✓ Están bien','Sí, necesitan cosas':'🙌 Necesitan cosas','Nadie contestó · no necesita nada':'📵 No contestó · nada'\};/, 'what they reported, short');
  assert.match(agent, /active\.map\(n=>needItem\(n,pending\?false:'compact'\)\)\.join\(''\)\}<\/ul>\$\{pending\?'':`<div class="rowreport">\$\{reportBtn\}<\/div>`\}/);
  const stats = new Function(`const icoFlag='F',icoWalk='W';${agent.match(/function needStats\(list\)[^\n]*/)[0]};return needStats`)();
  assert.equal(stats([]), '', 'nothing pending, nothing completed: no counters at all');
  assert.equal(stats([{status: 'COMPLETED'}]), '<small class="minorstats">1 completada</small>');
  assert.doesNotMatch(stats([{status: 'OPEN'}, {status: 'OPEN'}]), /class="route"|minorstats/, 'only what there is');
  assert.match(stats([{status: 'OPEN'}, {status: 'OPEN'}, {status: 'IN_PROGRESS'}]), /title="Por llevar"[^>]*>F<b>2<\/b><span>por llevar<\/span><\/div><div class="route" title="En camino"[^>]*>W<b>1<\/b><span>en camino<\/span>/);
  assert.match(agent, /`<p class="summary">\$\{\[p\.place&&`<span class="si" title="Lugar">\$\{icoPinS\}\$\{clean\(p\.place\)\}<\/span>`,p\.coordinator&&`<span class="si" title="Reportado por">\$\{icoCoordS\}\$\{clean\(p\.coordinator\)\}<\/span>`\]\.filter\(Boolean\)\.join\(''\)\}<\/p>\$\{p\.people\?`<p class="summary"><span class="si"/, 'place and reporter on one line, people on its own');
  assert.match(agent, /names\.length>3\?` <span class="more">y \$\{names\.length-3\} más<\/span>`/, 'up to three names, «y N más»');
  assert.match(agent, /role=n\.kind==='repair'\?'Constructor':'Runner'/, '«Runner», the term used in the operation');
  // Runner box on each need: name, phone opens WhatsApp, call button, time en route.
  assert.match(agent, /\$\{phone\?tel\(phone,u\?\.name\|\|n\.runnerName\|\|''\):'<small>sin teléfono<\/small>'\}/, 'runner phone uses the shared copy / call / WhatsApp widget');
  assert.match(agent, /const tel=\(phone,name=''\)=>window\.AppNav\?AppNav\.phone\(phone,name\):clean\(phone\);/);
  assert.doesNotMatch(agent, /closest\('\[data-needcard\]'\)/, 'tapping a need\'s background does nothing; «Actualizar» edits it');
  assert.match(agent, /\$\('#needDialog'\)\.addEventListener\('cancel',e=>e\.preventDefault\(\)\);/, 'Escape does not close it');
  // Phone: the need window is a history step, so the phone's Back closes it (and ‹ Atrás / saving go back that step).
  assert.match(agent, /function needDlgHistory\(kind\)\{if\(!narrowCmd\.matches\)return;const url=new URL\(location\.href\);url\.searchParams\.set\('necesidad',kind\);if\(dlgPushed\)history\.replaceState\(\{dlg:1\},'',url\);else\{history\.pushState\(\{dlg:1\},'',url\);dlgPushed=true\}\}/);
  assert.match(agent, /addEventListener\('popstate',\(\)=>\{if\(skipPop\)\{skipPop=false;return\}const ck=\$\('#checkupDlg'\);if\(ck\?\.open\)\{ckPushed=false;ck\.close\(\);return\}const d=\$\('#needDialog'\);if\(d\.open\)\{dlgPushed=false;d\.close\(\);return\}/);
  assert.doesNotMatch(agent, /\$\('#needDialog'\)\.close\(\)/, 'every close goes through closeNeedDlg');
  assert.doesNotMatch(agent, /if\(e\.target===d\)\{d\.close\(\)/, 'tapping the backdrop does not close it');
  assert.doesNotMatch(agent.match(/function openSupplyStep[^\n]*/)[0], /\.focus\(\)/, 'no auto-focus (no surprise keyboard)');
  assert.match(agent, /<span class="state">\$\{n\.status==='OPEN'\?'Necesita':n\.kind==='repair'&&n\.status==='IN_PROGRESS'\?'🔨 En obra':n\.status==='IN_PROGRESS'&&n\.dispatchedAt\?'Despachada':needLook\[n\.status\]\[1\]\}<\/span>[^\n]*?\$\{needChips\(n\.need\)\}[^\n]*?\$\{runnerBlock\(n,actions===true\)\}/, 'open needs read «Necesita» right before the chips');
  assert.match(agent, /\.needlist \.open \.state\{background:#173c46;color:#fff\}/, 'not red: red is for Urgente');
  assert.doesNotMatch(agent, /Corredor/);
  assert.match(agent, /const needOrder=\(a,b\)=>Object\.keys\(needLook\)\.indexOf\(a\.status\)-Object\.keys\(needLook\)\.indexOf\(b\.status\)\|\|/, 'abiertas first, then en camino, then completed');
  assert.match(agent, /const rank=p=>needTests\.urgent\[1\]\(p\)\?0:p\.status==='pending'\?1:needTests\.OPEN\[1\]\(p\)\?2:3;/);
  // Row background → map; «Ver detalles» → view mode in the row (edit, checkup, changelog, needs) through /api/agent.
  assert.match(agent, /if\(e\.target\.closest\('\[data-details\]'\)\)\{toggleDetails\(id\);return\}/);
  assert.match(agent, /if\(e\.target\.closest\('\[data-onmap\]'\)\)\{showOnMap\(id\);return\}\};/, 'clicking the row toggles its details (one open at a time) and moves the map');
  assert.match(agent, /function afterClick\(e,run\)\{clearTimeout\(clickTimer\);if\(e\.detail>1\|\|String\(getSelection\(\)\|\|''\)\.trim\(\)\)return;clickTimer=setTimeout\(\(\)=>\{if\(!String\(getSelection\(\)\|\|''\)\.trim\(\)\)run\(\)\},230\)\}/, 'selecting text (drag or double-click) never toggles');
  assert.match(agent, /function toggleDetails\(id\)\{const open=openId!==id;openId=open\?id:null;[^\n]*?if\(open\)\{[^\n]*?focusPoint\(id\)\}\}/);
  assert.match(agent, /if\(map&&hasCoords\(p\)\)\{if\(map\.getContainer\(\)\.offsetWidth\)map\.flyTo\(p\.coords/);
  assert.match(agent, /\.on\('click',\(\)=>\{if\(narrowCmd\.matches\)showCView\('points'\);focusPoint\(p\.id,true\)\}\)/, 'a pin scrolls to its row (switching to Puntos on a phone)');
  // Phone: no side column on the points view; «Overview» tab holds the map and monitors.
  assert.match(agent, /<nav class="cmdtabs" aria-label="Vista"><button type="button" data-cview="points" aria-pressed="true"><span aria-hidden="true">📍<\/span> Puntos<\/button><button type="button" data-cview="overview" aria-pressed="false"><span aria-hidden="true">🗺️<\/span> Mapa<\/button><\/nav>/);
  assert.match(agent, /body:not\(\.cview-overview\) \.mapcorner\{display:none\}body\.cview-overview \.boardmain\{display:none\}/);
  // Phone: the need windows are full-screen with a fixed bar: «‹ Atrás» left, title, main action right.
  assert.match(agent, /<div class="dlghead"><button type="button" class="dlgback" data-act="cancel">‹ Atrás<\/button><div class="dlgtitle">[^`]*<\/div><button type="submit" class="dlgsave">\$\{n\?'Guardar':'Reportar'\}<\/button>/);
  assert.match(agent, /<button type="button" class="dlgback" data-act="cancel">‹ Más tarde<\/button>[^`]*<button type="submit" class="dlgsave">\$\{repair\?'Asignar':'Despachar'\}<\/button>/);
  assert.match(agent, /@media\(max-width:760px\)\{\.needdlg\{width:calc\(100vw \/ var\(--text-zoom,1\)\);max-width:calc\(100vw \/ var\(--text-zoom,1\)\);height:calc\(100dvh \/ var\(--text-zoom,1\)\);[^}]*\}[^@]*\.needdlg \.dlghead\{position:sticky;top:0;/);
  assert.match(agent, /\.needdlg \.dlgx,\.needdlg \.formacts\{display:none\}/);
  // Like the admin ficha: an opened point stands alone with «← Puntos» + title; editing is its own history step with
  // Cancelar / «Editar · punto» / Guardar on top, and saving or cancelling steps back to the point.
  assert.match(agent, /data-ck="no">No<\/button><button type="button" class="ckbtn ok" data-ck="fine">Sí, están bien<\/button><button type="button" class="ckbtn go" data-ck="needs">Sí, necesitan cosas<\/button>/, 'checkup answers: No / Sí, están bien / Sí, necesitan cosas');
  assert.match(agent, /if\(choice==='fine'\|\|choice==='none'\)for\(const n of openNeeds\)await send\('\/api\/agent\/needs\/'\+encodeURIComponent\(n\.id\),'PATCH',\{status:'DELETED'\}\);[^\n]*?if\(choice==='needs'\)needsAfterCheckup\(id\)/, '«Sí, necesitan cosas» records the checkup and opens the need window');
  assert.match(agent, /if\(pending\.length===1\)\{openNeedDialog\(pending\[0\],p\);return\}if\(!pending\.length\)\{openNeedDialog\(null,p\);return\}/);
  assert.match(agent, /function openEdit\(id\)\{if\(openId!==id\)toggleDetails\(id\);history\.pushState\(\{edit:1\},'','\?punto='\+encodeURIComponent\(id\)\+'&editar=1'\);mode='camp';/);
  assert.match(agent, /function closeEdit\(\)\{mode='';if\(backToScout\(\)\)return;if\(history\.state\?\.edit\)\{skipPop=true;history\.back\(\)\}else history\.replaceState\(null,'','\?punto='\+encodeURIComponent\(openId\)\);/);
  assert.match(agent, /`<button type="button" class="sbback sbcancel" data-solo="cancel">Cancelar<\/button><span class="sbtitle">Editar · \$\{title\}<\/span><button type="submit" form="campForm" class="sbsave">Guardar<\/button>`:`<button type="button" class="sbback" data-solo="back">\$\{fromScout&&!history\.state\?\.pt\?'← Scout':'← Puntos'\}<\/button><span class="sbtitle">\$\{title\}<\/span>`/);
  assert.match(agent, /mode=openId&&q\.get\('editar'\)==='1'\?'camp':''/, 'Back / Forward restore edit mode from the address');
  assert.match(agent, /closeEdit\(\);await after\('Punto actualizado\.'\)/);
  assert.match(agent, /\.row\.editing>:not\(\.rowdetail\)\{display:none\}\.row\.editing \.rowdetail>:not\(\.campedit\)\{display:none\}/, 'edit mode shows only the form');
  assert.match(agent, /body\.cmd-solo \.boardmain>\.sidehead,body\.cmd-solo \.row:not\(\.open\)\{display:none\}/, 'desktop too: the opened point stands alone');
  // Phone: an opened point shows on its own, with «← Puntos» back to the list.
  assert.match(agent, /function render\(\)\{document\.body\.classList\.toggle\('cmd-solo',!!openId\);/);
  assert.match(agent, /data-act="edit"[^`]*Editar datos<\/button><button type="button" class="act" data-act="checkup">/);
  assert.match(agent, /<div class="rowside">\$\{pending\|\|open\?'':`<button type="button" class="act checkpoint \$\{checkupAge\(p\)\}" data-act="checkup"/, 'card: only «Checkup»; edit from the opened point');
  assert.match(agent, /data-act="edit"\$\{mode==='camp'\?' disabled':''\}>\$\{icoEdit\}Editar datos<\/button><button type="button" class="act go" data-act="approve">/, 'pending points can be corrected before approving');
  assert.match(agent, /<details class="dlog">\$\{checkupHtml\(p\)\}<\/details><details class="dlog">\$\{changelogHtml\(p\)\}<\/details>/);
  assert.match(agent, /PointForm\.render\(cf\.querySelector\('\.campfields'\),\{admin:true,edit:true,point:p\}\)/);
  assert.match(agent, /send\('\/api\/agent\/camps\/'\+encodeURIComponent\(id\),'PATCH',PointForm\.values/);
  assert.match(agent, /send\('\/api\/agent\/camps\/'\+encodeURIComponent\(id\)\+'\/checkup','POST',\{result:choice\}\)/);
  assert.match(agent, /send\('\/api\/agent\/needs\/'\+encodeURIComponent\(n\.id\),'PATCH',\{status:'COMPLETED'\}\)/);
  assert.match(agent, /send\('\/api\/agent\/decisions','POST',\{id,decision:'approved'\}\)/);
  assert.match(agent, /setInterval\(\(\)=>\{if\(csrf&&!document\.hidden&&!mode&&!\$\('#needDialog'\)\.open\)refresh\(\)\},30000\)/, 'no auto-refresh while a form or the need window is open');
  // Row date, short and relative (exact date on hover): «Suplido hace 2 d» from the latest completed need, «Revisado hace
  // 1 h» after a newer checkup or «Ya no necesita nada», else «Creado hace …»; then who added it («por Ada (tú)»).
  assert.match(agent, /`<span class="chip text">\$\{clean\(part\)\}<\/span>`/, '«Otros» text without the other icon');
  assert.match(agent, /n\.createdAt&&'Solicitado: '\+stamp\(n\.createdAt\)/);
  // Tapping a need opens «Actualizar necesidad»: urgency, needs, «Ya no necesita nada» → quitar (DELETED).
  assert.match(agent, /<dialog class="needdlg" id="needDialog" aria-label="Actualizar necesidad"><\/dialog>/);
  assert.match(agent, /<div class="urgseg" role="radiogroup" aria-label="Urgencia"><label><input type="radio" name="urgency" value="normal" checked><span>Normal<\/span><\/label><label class="urg"><input type="radio" name="urgency" value="urgente"><span>Urgente<\/span><\/label><\/div>/);
  assert.match(agent, /<button type="button" class="noneed" data-act="noneed"><span aria-hidden="true">✕<\/span>Ya no necesita nada<\/button>/);
  assert.match(readFileSync(new URL('./command.html', import.meta.url), 'utf8'), /\.noneed\{display:inline-flex;[^}]*border-radius:10px;padding:8px 13px;[^}]*cursor:pointer;/, '«Ya no necesita nada» looks like a button');
  // No pending needs: «+ Reportar necesidad» on the row (approved points) and in the details opens the same window.
  assert.match(agent, /const reportBtn='<button type="button" class="act report" data-act="newneed">\+ Reportar necesidad<\/button>';/);
  assert.match(agent, /\(pending\?'':`<div class="rownone">\$\{reportBtn\}<\/div>`\)/, 'no «sin necesidades» text: just the Reportar button');
  assert.doesNotMatch(agent, /Sin necesidades abiertas ni en camino|Todavía no hay necesidades de servicio/);
  // Need window (Actualizar and Reportar share it): state buttons, required who with typeahead (new runner allowed), packing list.
  assert.match(agent, /<div class="urgseg stateseg" role="radiogroup" aria-label="Estado"><label><input type="radio" name="status" value="OPEN" checked><span>Necesita<\/span><\/label><label class="route"><input type="radio" name="status" value="IN_PROGRESS"><span>\$\{repair\?'🔨 En obra':icoWalk\+'En camino'\}<\/span><\/label><label class="done"><input type="radio" name="status" value="COMPLETED"><span>\$\{icoDone\}Completada<\/span><\/label><\/div>/);
  assert.doesNotMatch(agent.match(/function needForm\(n,p\)[^]*?<\/form>`\}/)[0], /DELETED|Eliminada/, '«Eliminada» is «Ya no necesita nada», not an option');
  assert.match(agent, /<input id="whoInput" name="who" autocomplete="off" spellcheck="false" placeholder="Escribe o elige un nombre…" role="combobox"[^>]*><button type="button" class="tacaret" data-act="whoall" aria-label="Ver la lista"/, 'typeahead with a ▾ to pick from the whole list');
  assert.match(agent, /\$\{myTeam\(\)\?'<label class="teamtoggle"><input type="checkbox" data-teamonly checked> Solo mi equipo<\/label>':''\}/, 'agents on a team get «Solo mi equipo», on by default');
  assert.match(agent, /&&\(!f\.querySelector\('\[data-teamonly\]'\)\?\.checked\|\|u\.teamId===myTeam\(\)\)\);/);
  assert.match(agent, /<button type="button" class="bigact" data-act="needpack">📦 Preparar lista de empaque<\/button>/, 'packing is a big next-action button');
  assert.match(agent, /if\(status!=='OPEN'&&!assignee\)\{[^\n]*?send\('\/api\/agent\/runners','POST',\{name:typed,telephone:phone\}\)[^\n]*?throw Error\(filled\?'Elige quién la suplió, o marca «No, ya no la necesitan»\.':`Para «\$\{status==='COMPLETED'\?'Completada':'En camino'\}» elige quién la lleva\.`\)/, 'En camino and Completada require a runner; a new name creates one with its phone');
  // «Ya no necesita nada» asks who fulfilled it: «Sí» saves it Completada for that runner (a finished run), «No» removes it.
  assert.match(agent, /<div class="dlglabel">¿Alguien la suplió\?<\/div><div class="urgseg noneby" role="radiogroup" aria-label="¿Alguien la suplió\?"><label><input type="radio" name="noneby" value="" checked><span>No, ya no la necesitan<\/span><\/label><label class="done"><input type="radio" name="noneby" value="done">/);
  assert.match(agent, /filled=filledBy\(f\),status=filled\?'COMPLETED':f\.elements\.status\?\.value\|\|'OPEN',text=filled\?editing\.need:needText\(f\)/, 'fulfilled keeps the need text and completes it');
  assert.match(agent, /<div class="dlgsec"><div class="packbox"><\/div><button type="button" class="bigact" data-act="needpack">/, 'Actualizar has the packing list');
  // Reportar is two steps: 1) urgency + needs only; 2) «¿La suplimos ahora?» packing, dispatch (runner typeahead), pickup photos — all optional.
  assert.match(agent, /\$\{n\?`<div class="dlgsec statesec"><div class="dlglabel">Estado<\/div>/, 'step 1 has no state, runner or packing');
  assert.match(agent, /if\(!editing&&supply\)\{notice='Necesidad reportada\.';await refresh\(\);openSupplyStep\(/);
  const step2 = agent.match(/function supplyForm\(n,p\)[^\n]*(\n[^\n]*){5}/)[0];
  assert.match(step2, /<p class="supplyq">\$\{repair\?'¿Quién hace la obra\?':'¿La suplimos ahora\?'\}<\/p>/);
  assert.match(step2, /<div class="dlgsec"><div class="packbox"><\/div><button type="button" class="bigact" data-act="needpack">📦 Preparar lista de empaque<\/button>/);
  assert.match(step2, /<div class="dlglabel">\$\{repair\?'Asignar a un constructor':'Despachar a un '\+who\}<\/div><div class="whobox">\$\{whoInputHtml\(\)\}/);
  assert.match(step2, /<div class="dlgsec"><div class="dlglabel">Documentar partida<\/div>\$\{mediaOn\?'<div class="pthumbs pickupthumbs"><\/div>/, '«Documentar partida» is the section title');
  assert.match(agent, /<button type="button" class="pthumb add\$\{list\.length\?'':' empty'\}" data-act="pickphotos" aria-label="Agregar foto o video">\$\{list\.length\?'\+':'＋ Agregar foto o video'\}<\/button>/, '«+» tile adds another');
  assert.match(agent, /<dialog id="mediaViewer" class="mviewer" aria-label="Foto de partida"><div class="mvbody" id="mvBody"><\/div><div class="mvacts"><button type="button" class="mvdel" id="mvDelete">Eliminar<\/button><button type="button" class="mvclose" id="mvClose">Cerrar<\/button><\/div><\/dialog>/);
  assert.match(agent, /send\('\/api\/agent\/needs\/'\+encodeURIComponent\(v\.needId\)\+'\/media\/'\+encodeURIComponent\(v\.mediaId\),'DELETE',\{\}\)/, 'delete a mistaken photo');
  assert.match(agent, /<dialog id="pickDialog" class="confirmdlg" aria-labelledby="pickTitle"><form method="dialog" class="confirmbox"><h2 id="pickTitle">Documentar partida<\/h2><div class="mediabtns pickbtns"><button type="button" class="btn plain" data-pick="pickPhoto">📷 Tomar foto<\/button><button type="button" class="btn plain" data-pick="pickVideo">🎥 Grabar video<\/button><button type="button" class="btn plain" data-pick="pickGallery">🖼️ Galería<\/button>/, 'same camera / video / gallery picker as the runner page');
  assert.match(agent, /<input type="file" id="pickPhoto" accept="image\/\*" capture="environment" hidden>/, 'Tomar foto opens the camera');
  assert.match(step2, /<button type="button" class="act" data-act="cancel">Más tarde<\/button><button type="submit" class="act go">\$\{repair\?'Asignar constructor':'Despachar'\}<\/button>/);
  assert.match(agent, /PointForm\.uploadMedia\('\/api\/agent\/needs\/'\+encodeURIComponent\(id\)\+'\/media\?stage=pickup',file,\{csrf,/);
  assert.match(agent, /'PATCH',repair\?\{builderId:assignee,status:'IN_PROGRESS'\}:\{runnerId:assignee,status:'IN_PROGRESS'\}\)/, 'dispatch sends it en camino with the runner');
  assert.match(agent, /if\(act==='editneed'\|\|act==='newneed'\)\{openNeedDialog\(act==='editneed'\?n:null,p\);return\}/);
  assert.match(agent, /if\(act==='noneed'\)\{const pick=f\.querySelector\('\.needpick'\);if\(pick\.querySelector\('\.pf-otros-text'\)\)PointForm\.fillNeeds\(pick,''\);f\.dataset\.none='1'/);
  assert.match(agent, /if\(editing&&!text\)\{await send\('\/api\/agent\/needs\/'\+encodeURIComponent\(editing\.id\),'PATCH',\{status:'DELETED'\}\)/);
  // «En ruta» monitor under the map: longest en route first, opens the point.
  assert.match(agent, /<\/section><p class="mapnote" id="maphint">[^\n]*<section class="routemon" aria-label="Monitor en camino">/);
  assert.match(agent, /needs\.filter\(n=>n\.status==='IN_PROGRESS'&&live\.has\(n\.campId\)\)\.map\(n=>\(\{n,t:n\.enRouteAt\|\|n\.updatedAt\|\|n\.createdAt\|\|''\}\)\)\.sort\(\(a,b\)=>String\(a\.t\)\.localeCompare\(String\(b\.t\)\)\)/);
  assert.match(agent, /class="monitem \$\{h>=24\?'late':h>=6\?'slow':''\}"/);
  // Compact filters: status toggle, one need chip with counts, area / need type / sort selects, clear.
  assert.match(agent, /<div class="seg" role="group" aria-label="Estado del punto"><button type="button" data-status="" aria-pressed="true">Todos<\/button><button type="button" data-status="pending"[^>]*>Por revisar<\/button><button type="button" data-status="approved"[^>]*>Aprobados<\/button><button type="button" data-status="tolocate"[^>]*>📍 Por ubicar<\/button><\/div>/);
  assert.match(agent, /<select id="areaFilter" aria-label="Área"><\/select><select id="typeFilter" aria-label="Tipo de necesidad"><\/select>[^]*?<select id="sortBy" aria-label="Ordenar"><option value="checkup">Orden: checkup más antiguo<\/option><option value="checkupNew">Orden: checkup más reciente<\/option>|<select id="sortBy" aria-label="Ordenar"><option value="checkup">Orden: checkup más antiguo<\/option><option value="checkupNew">Orden: checkup más reciente<\/option>[^]*?<select id="areaFilter" aria-label="Área"><\/select><select id="typeFilter" aria-label="Tipo de necesidad"><\/select>/);
  assert.match(agent, /const needTests=\{urgent:\['Urgentes'[^\n]*OPEN:\['Abiertas'[^\n]*IN_PROGRESS:\['En camino'[^\n]*repair:\['Reparaciones'[^\n]*none:\['Sin pendientes'/);
  assert.match(agent, /const listed=\(\)=>filtered\(\)\.sort\(sorters\[filt\.sort\]\|\|sorters\.checkup\)/, 'the map and the list share the filtered set');
  assert.match(agent, /try\{localStorage\.setItem\('commandFilters'/, 'filters remembered per browser, guarded');
  assert.match(agent, /'Ningún punto con esta búsqueda y filtros\.'/);
  // Missing GPS: crossed-pin icon on the row and a «Sin GPS» filter with its count.
  assert.match(agent, /<h2>\$\{hasCoords\(p\)\?'':`<span class="nogpstag" title="Sin GPS: no aparece en el mapa" role="img" aria-label="Sin GPS">\$\{icoNoGps\}<\/span>`\}/);
  assert.match(agent, /<button type="button" class="pill gpsf" id="gpsFilter" aria-pressed="false"><\/button>/);
  assert.match(agent, /&&\(skip==='gps'\|\|!filt\.gps\|\|!hasCoords\(p\)\)/);
  // «Sin GPS» side widget under the en-ruta monitor: points with pending needs first; tapping opens the point.
  assert.match(agent, /<\/ol><\/section><section class="routemon gpsmon" aria-label="Puntos sin GPS">[^\n]*<ol id="gpsList">/);
  assert.match(agent, /points\.filter\(p=>\['approved','pending'\]\.includes\(p\.status\)&&!hasCoords\(p\)\)\.sort\(\(a,b\)=>\(pendingNeeds\(a\)\.length\?0:1\)-\(pendingNeeds\(b\)\.length\?0:1\)/);
  assert.match(agent, /\$\('#gpsList'\)\.onclick=e=>\{[^\n]*openEdit\(b\.dataset\.mon\)\};/);
  assert.match(agent, /history\.pushState\(\{pt:1\},'',open\?'\?punto='\+encodeURIComponent\(id\):location\.pathname\)/, 'deep link survives reload');
  assert.match(agent, /Búsqueda aproximada: no hay GPS registrado\./);
});

test('public map, desktop: the side panel is «📊 Resumen» only (no Lista tab); the summary shows totals from /api/stats without filters', () => {
  const html = readFileSync(new URL('./index.html', import.meta.url), 'utf8');
  assert.match(html, /<aside class="side"><div class="sidehead">/, 'no tabs: the side panel is the summary');
  assert.doesNotMatch(html, /class="sidetabs"|data-side="list"/, 'no «Lista» tab on a computer');
  assert.match(html, /let sideTab='stats',summaryStats=null,summaryLoading=false;/, 'always starts on Resumen, not a remembered tab');
  assert.match(html, /<section class="summary" id="summary" role="tabpanel" aria-label="Resumen" hidden><\/section><\/aside>/);
  assert.match(html, /fetch\('\/api\/stats',\{cache:'no-store'\}\)/);
  const draw = html.slice(html.indexOf('function drawSummary()'), html.indexOf("// Computadora: el panel lateral muestra solo el Resumen"));
  assert.match(draw, /const rows=all\(\)/, 'every published point, not the filtered list');
  assert.doesNotMatch(draw, /campRows\(\)|coordRows\(\)|selectedCampStatus/, 'no filters in the summary');
  for (const label of ['Personas', 'Puntos', 'Zonas', 'Urgentes', 'Por llevar', 'En camino', 'Completadas', 'Qué se pide', 'Zonas con necesidades']) assert.ok(draw.includes(label), label);
  for (const label of ['L de agua al día', 'Bocas que alimentar', 'Consumo estimado*']) assert.ok(draw.includes(label), label);
  assert.doesNotMatch(draw, /'Carpas'|combustible/i, 'tents and fuel hidden for now');
  const WATER_L = new Function(html.match(/const WATER_L=[^\n]*/)[0] + ';return WATER_L')();
  assert.deepEqual(WATER_L, {babies: .75, children: 1.5, adults: 3, older: 2.5}, 'drinking water by age group');
  assert.match(draw, /<p class="sumstar">\* Agua de beber al día/, 'assumptions stated');
  assert.doesNotMatch(draw, /Estado de los puntos/, 'dropped: it repeated the need counts and the map colors');
  assert.match(draw, /if\(items\.length>8\)\{const rest=items\.slice\(7\);/, 'at most 8 rows: hand-written items by name, the tail folded into one row');
  const aliases = new Function(html.match(/const SUM_ALIASES=[^\n]*/)[0] + ';return SUM_ALIASES')();
  const hit = t => aliases.find(([re]) => re.test(t))?.[1];
  assert.equal(hit('comida'), 'Alimentos');
  assert.equal(hit('panales'), 'Kit de bebé');
  assert.equal(hit('lonas'), 'Refugio (carpa / lona / toldo)');
  assert.equal(hit('colchones'), undefined);
  assert.match(draw, /\.filter\(z=>z\.needing\)/, 'only zones that need something');
  assert.doesNotMatch(html, /data-sumpt|Sin GPS <small>/, 'no «Sin GPS» list in the side panel');
  assert.match(html, /@media\(max-width:760px\)\{\.sidetabs,\.summary\{display:none!important\}[^\n]*\}\n/, 'tabs on a computer; on a phone the 📊 view shows the summary');
  assert.match(html, /openLinkedCamp\(\);if\(sideTab==='stats'\|\|document\.body\.classList\.contains\('mobile-stats'\)\)loadSummary\(\)\}/, 'refreshes with the map data');
  assert.match(html, /body\.mobile-stats \.sidehead\{display:none!important\}body\.mobile-stats \.summary:not\(\[hidden\]\)\{display:block!important\}/, 'phone 📊 shows the same Resumen, not the filter panel');
  assert.match(html, /if\(view==='stats'\)\{drawSummary\(\);loadSummary\(\)\}/);
});

test('admin: pending point card has «Rechazar» next to «Aprobar»; it asks, then archives the report unpublished', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  assert.match(admin, /p\.status==='pending'\?`<div class="card-actions"><button data-id="\$\{clean\(p\.id\)\}" data-decision="approved">Aprobar<\/button><button class="editpending" data-edit-pending="\$\{clean\(p\.id\)\}">Editar<\/button><button class="reject" data-id="\$\{clean\(p\.id\)\}" data-decision="archived" data-reject="1">Rechazar<\/button><button class="merge" data-merge=/);
  assert.match(admin, /if\(b\.dataset\.decision==='archived'&&!confirm\(b\.dataset\.reject\?rejectQuestion:archiveQuestion\)\)return;/);
  assert.match(admin, /const rejectQuestion='¿Rechazar este reporte\? Se moverá a «Archivados» sin publicarse\./);
});

test('admin map: clicking a pin opens the point (or selects the pending card) without a pop-up bubble', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  const map = admin.slice(admin.indexOf('function renderMap'), admin.indexOf('const addPointNearby'));
  assert.doesNotMatch(map, /bindPopup|bindTooltip/);
  assert.match(map, /pin\.on\('click',\(\)=>\{if\(p\.status==='approved'\)openEditor\(p\.id\)/);
  assert.match(admin, /if\(fromList&&pin&&adminMap\)adminMap\.panTo\(pin\.getLatLng\(\)\)\}/);
});

test('admin review: «Editar» on a pending report opens only the point form; save or cancel returns to the list, still pending', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  assert.match(admin, /data-decision="approved">Aprobar<\/button><button class="editpending" data-edit-pending="\$\{clean\(p\.id\)\}">Editar<\/button><button class="reject"/);
  assert.match(admin, /const p=points\.find\(row=>row\.id===id&&\(row\.status==='approved'\|\|edit&&row\.status==='pending'\)\);if\(!p\)return;activeCamp=id;\$\('#archiveCamp'\)\.hidden=p\.status==='pending';/, 'pending opens only in edit mode; no Archivar there');
  assert.match(admin, /if\(b\)openEditor\(b\.dataset\.editPending,true,true\)/);
  assert.match(admin, /const pending=editingPending\(\);await load\(\);if\(pending\)\{backToReview\('Reporte actualizado\. Revísalo y apruébalo cuando esté listo\.'\);return\}/);
  assert.match(admin, /\$\('#closeCampEditor'\)\.onclick=\(\)=>\{if\(editingPending\(\)\)backToReview\(\);/);
  assert.match(admin, /\$\('#cancelCampEdit'\)\.onclick=\(\)=>editingPending\(\)\?backToReview\(\):/);
});

test('admin cards: people counts as icon tiles (like «Agregar punto»), review card in clear blocks, notes beside the form while editing a report', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  const src = ['clean', 'labels', 'contactKeys', 'countKeys', 'countTiles'].map(k => admin.match(new RegExp(`const ${k}=[^\\n]*`))[0]).join('\n');
  const PointForm = {icons: {people: '<svg id=pp/>', child: '<svg id=ch/>'}};
  const countTiles = new Function('PointForm', `${src};return countTiles`)(PointForm);
  assert.equal(countTiles({people: '30', children: '2', adults: ''}), '<div class="counts"><div class="count total"><svg id=pp/><b>30</b><span>Personas</span></div><div class="count"><svg id=ch/><b>2</b><span>Niños</span></div></div>');
  assert.equal(countTiles({}), '');
  assert.match(admin, /\$\{p\.status==='pending'\|\|p\.status==='rejected'\?reviewBody\(p\):`<p class="summary">/);
  assert.match(admin, /reviewNoteKeys=\['extra','note','notes','serviceNotes'\]/, 'the private import note is shown too');
  assert.match(admin, /noteText=\(t,k=''\)=>clean\(t\)\.split\('\\n'\)\.map\(\(l,i\)=>\/\^Dato dudoso\/i\.test\(l\.trim\(\)\)\?`<mark class="doubt">/, '«Dato dudoso» lines highlighted');
  const noteText = new Function(`const clean=s=>String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));const reviewNoteKeys=[],noteLabels={},${admin.match(/noteText=\(t,k=''\)=>[^\n]*?\.join\('\\n'\)/)[0]};return noteText`)();
  assert.equal(noteText('3 familias.\nDato dudoso: people «10».'), '3 familias.\n<mark class="doubt">Dato dudoso: people «10».</mark>', 'card: highlighted, no ×');
  assert.match(noteText('3 familias.\nDato dudoso: people «10».', 'extra'), /<mark class="doubt">Dato dudoso: people «10»\.<button type="button" class="doubtx" data-doubt-field="extra" data-doubt-line="1" title="Quitar esta duda \(ya resuelta\)" aria-label="Quitar esta duda">×<\/button><\/mark>$/, 'sidebar: × removes that line of that field');
  assert.match(admin, /lines\.splice\(\+b\.dataset\.doubtLine,1\);f\.value=lines\.join\('\\n'\);f\.dispatchEvent\(new Event\('input',\{bubbles:true\}\)\);drawReviewNotes\(\)/);
  assert.match(admin, /drawCampFields\(p\);drawReviewNotes\(\);/, 'sidebar reads the form, so saving keeps the removal');
  assert.match(admin, /<aside id="reviewNotes" class="review-notes" hidden><\/aside><details id="campEditDetails">/);
  assert.match(admin, /const rv=p\.status==='pending'&&edit;\$\('#campEditor'\)\.classList\.toggle\('review-edit',rv\);\$\('#reviewNotes'\)\.hidden=!rv;/);
  assert.match(admin, /#campEditor\.review-edit>#reviewNotes\{grid-column:2;grid-row:2\/span 20;position:sticky;/, 'right column on a computer');
  assert.match(admin, /async function loadNeeds\(\)\{if\(!activeCamp\|\|editingPending\(\)\)return;/, 'a pending report has no needs to load');
});

test('admin: points without GPS use the crossed-pin icon (card title, review «Dónde», point file), like the public map', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  const index = readFileSync(new URL('./index.html', import.meta.url), 'utf8');
  assert.equal(admin.match(/const noGpsIco='([^']*)'/)[1], index.match(/const noGpsIco='([^']*)'/)[1], 'same icon as the public map');
  const src = admin.match(/const noGpsIco=[^\n]*/)[0];
  const {noGpsTag} = new Function(`${src};return {noGpsTag}`)();
  assert.match(noGpsTag({}), /^<span class="nogpstag" title="Sin GPS: no aparece en el mapa" role="img" aria-label="Sin GPS"><svg[^]*M3 3l18 18/);
  assert.equal(noGpsTag({coords: [7.4, -80.4]}), '');
  assert.match(admin, /\$\{clean\(p\.code\)\} · \$\{clean\(p\.area\)\} \$\{PointForm\.verifiedBadge\(p,\{text:true\}\)\}\$\{noGpsTag\(p\)\}<span class="status/);
  assert.match(admin, /\$\{hasGps\(p\)\?icons\.pin\|\|'':noGpsIco\}Dónde\$\{hasGps\(p\)\?'':' · sin GPS'\}/);
  assert.match(admin, /<span class="nogpsline">\$\{noGpsIco\}Sin GPS: no aparece en el mapa<\/span>/);
});

test('runner: Ver en el mapa centers the point with a flag, opens its card, and highlights the list card', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  assert.match(runner, /function applyFocus\(\)\{[^\n]*map\.closePopup\(\);map\.setView\(pin\.getLatLng\(\),16\);/);
  assert.match(runner.match(/function applyFocus\(\)\{[^\n]*/)[0], /if\(n&&n\.camp\.id===focusCamp\)\{[^}]*pin\.openPopup\(\)\}/, 'chosen in the list: its card opens on the pin');
  assert.match(runner, /focused\?L\.marker\(camp\.coords,\{icon:L\.divIcon\(\{className:'',html:focusFlag,/, 'a flag instead of a bigger circle');
  assert.match(runner, /\$\('#feed'\)\.onclick=async e=>\{[^\n]*?const where=e\.target\.closest\('\[data-where\]'\);/, 'only the 📍 button goes to the map: tapping the card background does nothing (no stuck highlight)');
  assert.doesNotMatch(runner, /const bg=!e\.target\.closest/);
  assert.match(runner, /highlightNeed=n\.id;if\(!hasCoords\(n\)\)\{focusCamp=null;render\(\);fitAllPins\(\);return\}focusCamp=n\.camp\.id;pendingFocus=true;/, 'picking a card without GPS removes the old flag');
  assert.match(runner, /function goToNeed\(id\)\{tabFor\(id\);highlightNeed=id;if\(!hasCoords\(findNeed\(id\)\|\|\{camp:\{\}\}\)\)\{focusCamp=null;requestAnimationFrame\(fitAllPins\)\}/);
  assert.match(runner, /function fitAllPins\(\)\{if\(map&&pins\.length\)\{map\.closePopup\(\);map\.fitBounds\(pins\.map\(p=>p\.getLatLng\(\)\),/, 'no GPS: zoom out to all points');
});

test('run link: «Enlace para runner» in the command center opens runner.html on that run (?punto=&necesidad=)', () => {
  const command = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  assert.match(command, /data-act="sharerun" data-need="\$\{clean\(n\.id\)\}"[^>]*>\$\{icoShare\}Enlace para runner<\/button>/);
  assert.match(command, /const url=location\.origin\+'\/runner\.html\?punto='\+encodeURIComponent\(n\.campId\)\+'&necesidad='\+encodeURIComponent\(n\.id\)/);
});

test('runner: «Mi recorrido» (only your runs and their route) and «Se necesita» (what else is open, team filter inside)', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  assert.match(runner, /<nav class="maintabs" id="mainTabs" role="tablist" aria-label="Vista"><button type="button" role="tab" data-main="mine" aria-selected="false" title="Mis carreras en camino"><svg class="walkico"[^]*?<\/svg> Yo <b id="cntMine"><\/b><\/button><button type="button" role="tab" data-main="needs" aria-selected="false" title="Necesidades por llevar"><span aria-hidden="true">🚩<\/span> Runs <b id="cntNeeds"><\/b><\/button><button type="button" role="tab" data-main="req"[^>]*><span aria-hidden="true">📦<\/span> Pedidos <b id="cntReq"><\/b><\/button><button type="button" role="tab" data-main="meals"[^>]*><span aria-hidden="true">🍲<\/span> Comidas <b id="cntMeals"><\/b><\/button><\/nav><div class="tabfind"><input type="search" id="tabFind" autocomplete="off" aria-label="Buscar en esta pestaña"><\/div><div class="rfilter" id="teamFilter"><select id="teamSelect" class="rselect" aria-label="Filtro"><option value="0">Todas<\/option><option value="1">Solo mi equipo<\/option><\/select><\/div>/, 'a plain drop menu: Todas \/ Solo mi equipo');
  const tabNow = new Function('mine', 'mainTab', `return (${runner.match(/const tabNow=([^\n;]*);/)[1]})()`);
  assert.equal(tabNow([{id: 'a'}], ''), 'mine', 'opens on your route when you have one');
  assert.equal(tabNow([], ''), 'needs');
  assert.equal(tabNow([{id: 'a'}], 'needs'), 'needs', 'keeps the tab you picked');
  assert.match(runner, /if\(tab==='mine'\)\{const hits=mine\.filter\(matches\);feed\.innerHTML=mine\.length\?`<div id="runList">\$\{mine\.map\(runCard\)\.join\(''\)\}<\/div>\$\{hits\.length\?'':noHits\(\)\}`:'<p class="empty">No tienes carreras en camino\. En «Runs» puedes tomar una\.<\/p>';[^\n]*?renderMap\(\[\.\.\.hits,\.\.\.shown\]\);return\}/, '«Yo» lists only your runs; the map still shows the other open runs around you');
  assert.match(runner, /renderMap\(\[\.\.\.mine,\.\.\.shown\]\)\}/, 'Se necesita keeps your stops on the map');
  assert.match(runner, /body\.tab-mine \.rfilter/, 'Todas \/ Mi equipo only in Se necesita');
  assert.match(runner, /function goToNeed\(id\)\{tabFor\(id\);/, 'links and map pins switch to the right tab');
});

test('runner: with a single run there is nothing to reorder (no ☰ or ▲ ▼)', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  const card = runner.match(/function runCard[^\n]*/)[0];
  assert.match(card, /<div class="runhead">\$\{mine\.length>1\?'<button type="button" class="drag"/);
  assert.match(card, /\$\{mine\.length>1\?`<button type="button" class="move" data-move="-1"/);
});

test('packing list dialog is full screen on phones (not centered under the keyboard)', () => {
  const pf = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
  assert.match(pf, /@media\(max-width:760px\)\{\.pfp\{inset:0;width:auto;height:auto;max-width:none;max-height:none;margin:0;border-radius:0\}/, 'full screen on a phone, sized from the screen edges so the text-size zoom does not overflow it');
  assert.doesNotMatch(pf, /\.pfp\{[^}]*vw/, 'no viewport-width units');
});

test('runner run card: «hace 2 d» at the bottom next to «No puedo continuar»; people stay on top', () => {
  const card = readFileSync(new URL('./runner.html', import.meta.url), 'utf8').match(/function runCard[^\n]*/)[0];
  assert.match(card, /<span class="spacer"><\/span>\$\{locButtons\(n\)\}/, 'map and Maps buttons top right');
  assert.match(card, /<div class="runfoot">\$\{n\.revisedAt\|\|n\.createdAt\?`<small class="when"[^`]*`:'<span><\/span>'\}<button class="btn quit" type="button" data-release="\$\{id\}">No puedo continuar<\/button><\/div>/);
});

test('runner «Se necesita» filter: a plain select Todas / Solo mi equipo', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  assert.match(runner, /function drawTeamTabs\(\)\{\$\('#teamSelect'\)\.value=teamOnly\?'1':'0'\}/);
  assert.match(runner, /\$\('#teamSelect'\)\.onchange=e=>\{teamOnly=e\.target\.value==='1';render\(\)\};/);
});

test('Command: home repairs show the builder\'s work (progress, what\'s missing, materials, evaluations, photos) and repair wording', () => {
  const cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  assert.match(cmd, /function needMedia\(n\)\{if\(n\.kind==='repair'\)return repairInfo\(n\);/);
  const rep = cmd.match(/function repairInfo\(n\)\{[^\n]*/)[0];
  for (const bit of ["'/api/agent/media/'", 'Avance ${clean(n.progress||0)}%', 'Qué falta:', 'Materiales:', 'Evaluaciones (${ev.length})', "[['inspect','Evaluación'],['before','Antes'],['progress','Avance'],['after','Final']]"]) assert.ok(rep.includes(bit), bit);
  assert.match(cmd, /El constructor todavía no documentó la obra/);
  assert.match(cmd, /\$\{repair\?'¿Quién hace la obra\?':'¿La suplimos ahora\?'\}/);
  assert.match(cmd, /\$\{repair\?'🔨 En obra':icoWalk\+'En camino'\}/);
});
test('packing list window: the needs on top grey out as items cover them; grouped catalog; Cancelar goes back', () => {
  const pf = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
  assert.match(pf, /function openPacking\(\{subtitle = '', items = \[\], save, need = '', packed = \[\]\}\)/);
  assert.match(pf, /<span class="pfp-need\$\{done \? ' done' : ''\}">/);
  assert.match(pf, /const NEED_PACK = \{'Agua': \['Agua'\], 'Alimentos': \['Bolsas alimentación'\]/);
  assert.match(pf, /\['⛺ Refugio', \['Carpas \/ toldas', 'Lonas', 'Camping', 'Hamacas'/);
  assert.match(pf, /data-act="other">Cancelar<\/button>/, 'choosing a quantity: «Cancelar» goes back to the items');
  assert.match(pf, /head\('🙌 Para lo que necesitan'\) \+ top\.map\(btn\)\.join\(''\) \+ head\('Categorías'\)/, 'items for needs not crossed out yet come first');
  assert.match(pf, /<button type="button" class="pfp-cat" data-cat="\$\{g\}">/, 'categories you tap into');
  assert.match(pf, /<div class="pfp-crumbs"><button type="button" class="pfp-crumb" data-crumb>‹ Todos<\/button>/, 'breadcrumb back to all');
  assert.match(pf, /pq\('\[data-act=cancel\]'\)\.hidden = step === 'qty';/);
  assert.match(pf, /pq\('\[data-act=review\]'\)\.hidden = step === 'review' \|\| step === 'qty';/);
  assert.match(pf, /else if \(act === 'cancel'\) \{ if \(!pq\('\.pfp-review'\)\.hidden\) packStep\('pick'\); else if \(packDiscard\(\)\) d\.close\(\); \}/);
  for (const page of ['supplier.html', 'runner.html', 'command.html', 'admin.html']) assert.match(readFileSync(new URL('./' + page, import.meta.url), 'utf8'), /PointForm\.openPacking\(\{need:/, page + ' passes the need');
  const sup = readFileSync(new URL('./supplier.html', import.meta.url), 'utf8');
  assert.match(sup, /function needSection\(n\)\{return `<section class="sec sec-need\$\{allPacked\(n\)\?' packed':''\}">\$\{secHead\('h-need','🙌','Necesita'/);
  assert.match(sup, /\.sec-need\{background:#fdeef0;/, 'Necesita section is light pink');
  assert.match(sup, /function noMatch\(total\)\{[^\n]*Ninguna necesidad coincide con la búsqueda «[^\n]*data-clearfind>✕ Borrar búsqueda y ver todas \(\$\{total\}\)/, 'no results: says what was searched and offers to clear it');
  for (const call of ['noMatch(todo().filter(teamOk).length)', 'noMatch(asks.filter(teamOk).length)', 'noMatch(open.filter(teamOk).length)', 'noMatch(all.filter(teamOk).length)']) assert.ok(sup.includes(call), call);
  assert.match(sup, /\$\{needChips\(n\.need,n\.packing\|\|\[\]\)\}/, 'needs already in the packing list are crossed out');
  assert.match(sup, /packed&&pf\.covers&&packed\.some\(l=>pf\.covers\(part,l\.item\)\)/);
  assert.match(sup, /<section class="sec sec-need\$\{allPacked\(n\)\?' packed':''\}">/, 'all covered: light green');
  assert.match(sup, /\.sec-need\.packed,\.sec-load\.packed\{background:#e8f6ee;/, 'Necesita and Para cargar turn the same green');
  assert.match(sup, /<section class="sec sec-load\$\{allPacked\(n\)\?' packed':''\}">/);
});
test('Command («Checkup») point card: checkup is the main action top right; names without phones; no need counters; «Creado por» with the medal', () => {
  const cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  assert.match(cmd, /\.act\.checkpoint\{order:-1;background:#ed7654;/);
  assert.match(cmd, /function checkupAge\(p\)\{const h=\(Date\.now\(\)-Date\.parse\(p\.lastCheckup\|\|''\)\)\/36e5;return h<24\?'ck-fresh':h<48\?'ck-warn':'ck-old'\}/, 'green < 24 h, yellow < 48 h, red older or never');
  assert.match(cmd, /\.row \.act\.checkpoint\{position:absolute;top:10px;right:10px;/, 'phones: top-right corner of the card');
  assert.match(cmd, /<p class="peopleline">\$\{icoTeamS\}\$\{names\.slice\(0,3\)\.map\(clean\)\.join\(', '\)\}/, 'contact names only');
  assert.doesNotMatch(cmd.match(/function pointRow\(p\)\{[^\n]*/)[0], /needStats\(|tel\(c\.phone\)/);
  assert.match(cmd, /<p class="byline">🧭 Creado por \$\{byName\(p\)\}\$\{vb\(p,\{text:true\}\)\}<\/p>/);
  assert.match(cmd, /const byName=p=>\{const t=\(users\.find\(u=>u\.id===p\.addedBy\)\|\|\{\}\)\.telephone;return t\?`<button type="button" class="telnum byname" data-tel-pop=/, 'tap the creator\'s name: call / WhatsApp');
});
test('«📍 Por ubicar»: banner on Runners, Supply and Checkup; locate as a new place or into an existing point', () => {
  const pf = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
  assert.match(pf, /fetch\('\/api\/full\/tolocate'/);
  assert.match(pf, /fetch\('\/api\/full\/camps\/' \+ encodeURIComponent\(id\) \+ '\/locate'/);
  assert.match(pf, /data-tl-new="[^"]*">📍 Lugar nuevo<\/button><button type="button" class="pf-tl-btn" data-tl-pick="[^"]*">🔎 Es un punto que ya existe<\/button>/);
  assert.match(pf, /📍 Usar mi ubicación/);
  for (const page of ['runner.html', 'supplier.html', 'command.html']) {
    const src = readFileSync(new URL('./' + page, import.meta.url), 'utf8');
    assert.match(src, /<div class="feed" id="feed"><p class="empty">Cargando…<\/p><\/div><div class="tlmount" id="toLocate" hidden><\/div>/, page);
    assert.match(src, /PointForm\.toLocate\(\$\('#toLocate'\),\{onDone:\(\)=>refresh\(\),bottomOf:\$\('#feed'\)\}\)/, page + ': «Por ubicar» at the bottom of the list');
    assert.match(src, /function checkToLocate\(\)\{if\(typeof PointForm!=='undefined'\)PointForm\.toLocate\(\$\('#toLocate'\)/, page);
  }
  const cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  assert.match(cmd, /data-status="tolocate" aria-pressed="false">📍 Por ubicar<\/button>/);
});
test('Runners and Supply: header «Buscar» is a global search page over the current one; a search inside each tab', () => {
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  assert.match(nav, /if \(input\.dataset\.global\) \{[^]*?openGlobalSearch\(input\.value\)/, 'Enter opens the global search');
  assert.match(nav, /<b id="gsTitle">🔎 Búsqueda global<\/b>/);
  assert.match(nav, /fetch\('\/api\/full\/scout'/, 'same data as Scout: active needs, contacts for teams');
  assert.match(nav, /function gsCard\(p\)\{|function gsCard\(p\) \{/, 'Checkup-style result cards');
  assert.match(nav, /window\.history\.pushState\(\{gs: 1\}, ''\)/, 'Back closes it');
  assert.doesNotMatch(nav, /[^.\w]history\.(pushState|back|replaceState)/, 'nav.js uses window.history: pages may have their own «history» variable (Supply does)');
  for (const [page, id] of [['runner.html', 'search'], ['supplier.html', 'find']]) {
    const src = readFileSync(new URL('./' + page, import.meta.url), 'utf8');
    assert.match(src, new RegExp(`<input class="topsearch" id="${id}" type="search" data-global="1"`), page);
    assert.match(src, /<div class="tabfind"><input type="search" id="tabFind"/, page + ': search inside the tab');
    assert.match(src, /function syncTabFind\(\)\{/, page);
  }
});
test('Supply «+ Nuevo despacho»: 5 steps (destination new or existing, runner new or existing, packing, photos, review)', () => {
  const sup = readFileSync(new URL('./supplier.html', import.meta.url), 'utf8');
  assert.match(sup, /<a class="add" href="#" id="newDispatch">\+ Nuevo despacho<\/a>/, 'header action (the «+» on phones)');
  assert.match(sup, /const DSP_STEPS=\['Destino','Runner','Lista de empaque','Fotos de partida','Revisar'\];/);
  assert.match(sup, /data-dsp="newdest">＋ Crear punto nuevo/);
  assert.match(sup, /data-dsp="newrunner">＋ Runner nuevo/);
  assert.match(sup, /dsp\.step\+\+;dspDraw\(\);if\(dsp\.step===3&&!dsp\.items\.length\)dspOpenItems\(\);/, 'step 3 opens the packing list by itself');
  assert.match(sup, /dsp\.step>=3\?dsp\.items\.length>0:true/, 'no dispatch without packed items');
  assert.match(sup, /function histCard[^\n]*🚚 Despachada \$\{clean\(agoEs\(n\.dispatchedAt\)\)\}/, 'history cards say when they were dispatched');
  assert.match(sup, /let mainTab=kept\('supplyTab',\['todo','asks','req','hist'\],'todo'\),histDays=\+kept\('supplyHistDays',\['7','30','0'\],'7'\);/, 'Supply keeps its tab and history period across a reload (default: last 7 days)');
  assert.match(sup, /sessionStorage\.setItem\('supplyTab',mainTab\)/);
  assert.equal((sup.match(/<\/svg><b class="cnthist"><\/b><\/button>/g) || []).length, 2, 'both history buttons carry the count');
  assert.match(sup, /data-hdays="\$\{d\}"[^`]*`\)\.join\(''\)\}<\/div>`/, '7 días · 30 días · Todo');
  assert.match(sup, /<p class="hmeta">\$\{n\.supplyCenter\?`🏬 \$\{clean\(n\.supplyCenter\)\} · `:''\}🏃/, 'history cards say which supply center it left from');
  assert.match(sup, /const again=mainTab===b\.dataset\.main;mainTab=b\.dataset\.main;[^\n]*tabTip\(b,again\)/, 'tapping a tab explains it; tapping the current tab again shows it even if hidden');
  for (const k of ['todo', 'asks', 'req', 'hist']) assert.match(sup, new RegExp(`\\n${k}:\\['`), `tip for ${k}`);
  assert.match(sup, /data-dsp="help" aria-label="Cómo funciona">❓<\/button>/);
  assert.match(sup, /dsp\.pushed=true;setTimeout\(\(\)=>dspHelp\(false\),200\);/, 'wizard explains itself the first time');
  assert.match(sup, /bottomOf:\$\('#feed'\)/, '«Por ubicar» stays at the bottom of the list');
  const navSrc = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  assert.match(navSrc, /if \(!anchor \|\| \(!again && tipHidden\(key\)\)\) return;\n    tipsSeen\.add\(key\); try \{ localStorage\.setItem\('tip:' \+ key, '1'\); \} catch \{\}/, 'a tip shows by itself only once');
  assert.doesNotMatch(navSrc, /No mostrar de nuevo/);
  assert.match(navSrc, /const wireSwipe = \(\) => swipeTabs\(document\.getElementById\('feed'\) \|\| document\.getElementById\('list'\), document\.getElementById\('mainTabs'\)\);/, 'side swipe switches tabs on every tabbed page');
  assert.match(sup, /data-dsp="newzone">🗺️ Destino es el área general de/, 'generic destination: only the zone');
  assert.match(sup, /function dspDestPreview\(p\)\{[^]*?<div class="dsp-mini" id="dspMini"><\/div>[^]*?📍 Sin GPS/, 'picked destination: mini map or «sin GPS»');
  assert.match(sup, /function dspMiniMap\(\)\{dsp\.mini\?\.remove\(\);/);
  assert.match(sup, /data-dsp="destsearch">‹ Volver a los destinos<\/button>/, 'from «Crear destino nuevo» back to the list (footer)');
  assert.match(sup, /class="dsp-back" data-dsp="destsearch">‹ Volver a la lista de destinos<\/button>/);
  assert.match(sup, /toLocate:true\}:\{area:dsp\.newDest\.area\.trim\(\)/, 'sent as «por ubicar»');
  assert.match(sup, /await api\('\/api\/supplier\/dispatches','POST',body\)/);
  assert.match(sup, /PointForm\.uploadMedia\('\/api\/supplier\/needs\/'\+encodeURIComponent\(res\.id\)\+'\/media\?stage=pickup'/, 'departure photos after the dispatch');
  assert.match(sup, /window\.history\.pushState\(\{dsp:1\},''\)/, 'Back goes one step back');
});
test('Command checkup is a phone-friendly screen: contacts with Llamar / WhatsApp, answers at the bottom', () => {
  const cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  assert.match(cmd, /const choice=p\?await checkupScreen\(p,/);
  assert.match(cmd, /<a class="ckcall" href="tel:\$\{clean\(r\.d\)\}">📞 Llamar<\/a><a class="ckwa" href="https:\/\/wa\.me\//);
  assert.match(cmd, /add\(p\.coordinator,p\.coordinatorPhone,'Reporta'\);peopleOf\(p\)\.forEach/, 'reporter and the group\'s people');
  assert.match(cmd, /<div class="ckacts">[^]*?data-ck="no">No<\/button>[^]*?data-ck="fine">Sí, están bien<\/button>[^]*?data-ck="needs">Sí, necesitan cosas<\/button>/, 'answers at the bottom');
  assert.match(cmd, /@media\(max-width:760px\)\{\.ckdlg\{width:calc\(100vw/, 'full screen on phones');
  assert.match(cmd, /data-ck="none">Nadie contestó, pero no necesita nada<\/button>/);
  assert.match(cmd, /\/checkup','POST',\{result:choice\}\);if\(choice==='fine'\|\|choice==='none'\)for\(/, 'the answer is logged; «no necesita nada» removes the open needs like «están bien»');
  assert.match(cmd, /const checkupHtml=p=>\{const notes=p\.checkupLog\|\|\[\]/, 'checkup history shows the answers');
  assert.match(cmd, /\$\{checkupPeople\(p\)\}\$\{checkupNeeds\(p\)\}/, 'the pending needs show below the contacts, to confirm on the call');
  assert.match(cmd, /function checkupNeeds\(p\)\{const list=pendingNeeds\(p\)/);
});
test('Supply «Puntos con necesidades» cards show when the need was last updated («🕓 Actualizada hace …»)', () => {
  const sup = readFileSync(new URL('./supplier.html', import.meta.url), 'utf8');
  assert.match(sup, /function reqCard\(n\)\{[^\n]*?\$\{n\.activityAt\?`<p class="reqwhen"[^`]*>🕓 Actualizada \$\{clean\(ago\(n\.activityAt\)\)\}<\/p>`:''\}/);
  assert.match(sup, /const ago=v=>\{[^\n]*hace \$\{Math\.max\(1,m\)\} min/);
});
test('phone header (nav.js, panels): logo · page name · «+» page action · ☰ menu with name + Salir, pages and text size', () => {
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  assert.match(nav, /bar\.innerHTML = `<nav class="mlinks" id="mLinks" aria-label="Páginas"><\/nav><button type="button" class="mtitle" id="mTitle" aria-haspopup="true" aria-expanded="false" disabled><\/button>\$\{top\.querySelector\('\.topsearch'\) \? `<button type="button" class="msearchbtn" id="mSearch"[^`]*` : ''\}<button type="button" class="madd" id="mAdd" hidden><\/button>\$\{document\.getElementById\('loginBtn'\) \? '<button type="button" class="mlogin" id="mLogin">Inicia sesión<\/button>' : ''\}<button type="button" class="mmenu" id="mMenu" aria-expanded="false" aria-controls="mSheet" aria-label="Menú">/, 'pages · 🔍 (public page) · action · ☰');
  assert.match(nav, /if \(!top \|\| !document\.getElementById\('appNav'\) \|\| document\.getElementById\('mBar'\)\) return;/, 'panels and the public map share the same phone header');
  assert.match(nav, /function fitMobileBar\(\) \{[^}]*?bar\.classList\.add\('tight'\);\s*drawTitle\(\);/, 'on phones the pages are always folded into «current page ▾», on every page, even when few');
  assert.doesNotMatch(nav, /over\(\)|\.mbar\.compact/, 'no inline pill row of pages on phones');
  assert.match(nav, /drop\.innerHTML = pick \? links\.innerHTML \+ \(admin \? '<a href="\/admin\.html" class="mdrop-admin" data-admin-link>/, 'phones: «🔒 Administración» at the end of the page list, as on a computer');
  assert.match(nav, /const admin = \(adminOK \|\| \(u\.roles \|\| \[\]\)\.includes\('admin'\)\) && here !== '\/admin\.html';/);
  assert.match(nav, /async function sharedLogout\(\) \{[^]*?await logoutAll\(\);/, 'pages without their own «Salir» get one in ☰');
  assert.match(nav, /async function logoutAll\(\) \{ await endSessions\(\); location\.href = '\/'; \}/, 'logout ends on the home page');
  for (const u of ['full', 'runner', 'builder', 'agent', 'supplier', 'cook', 'desk']) assert.match(nav, new RegExp(`'/api/${u}/logout'`), 'global logout ends the ' + u + ' session');
  assert.match(nav, /fetch\('\/api\/admin\/logout', \{method: 'POST'[^}]*'X-CSRF-Token': b\.csrf\}/, 'and the admin session');
  assert.match(template, /await \(window\.AppNav\?AppNav\.endSessions\(\):Promise\.resolve\(\)\);forgetCodeUser\(\);/, 'home ✕ logout is global too');
  for (const page of ['admin.html', 'builder.html', 'command.html', 'donations.html', 'runner.html', 'signup.html', 'supplier.html']) assert.match(readFileSync(new URL('./' + page, import.meta.url), 'utf8'), /\$\('#logout'\)\.onclick=async\(\)=>\{[^\n]*?await \(window\.AppNav\?AppNav\.logoutAll\(\):location\.assign\('\/'\)\)\}/, page + ': logging out is global and goes to the home page');
  assert.match(nav, /x\.setAttribute\('aria-label', 'Borrar búsqueda'\);[^]*?x\.onclick = \(\) => \{ input\.value = ''; input\.dispatchEvent\(new Event\('input', \{bubbles: true\}\)\);/, 'Buscar has a ✕ that clears it');
  assert.match(nav, /box\.classList\.toggle\('tson', on\);[^\n]*sb\.classList\.toggle\('tson', on\)/, 'a search in progress shows on the 🔍 button and the box');
  assert.match(nav, /\.msearchbtn\.tson\{position:relative;background:#ed7654;/);
  assert.match(nav, /const key = 'search:all';[^]*?sessionStorage\.setItem\(key, input\.value\)[^]*?saved = sessionStorage\.getItem\(key\)/, 'one search shared by every page (per tab)');
  assert.match(nav, /pill\.innerHTML = '<button type="button" class="sp-what" title="Ver o cambiar la búsqueda">🔎 Filtrado por «<b><\/b>»<\/button><button type="button" class="sp-x"/, 'a floating «Filtrado por «…» ✕» makes the active search obvious');
  assert.match(nav, /pill\.hidden = !on;/);
  for (const page of ['index.html', 'scout.html', 'runner.html', 'builder.html', 'supplier.html', 'command.html', 'cook.html', 'donations.html', 'team.html', 'signup.html', 'camp.html']) assert.match(readFileSync(new URL('./' + page, import.meta.url), 'utf8'), /<header class="top[^"]*"[^>]*>[^]*?<nav class="appnav" id="appNav" hidden><\/nav>/, page + ' uses the shared header (same phone bar)');
  assert.match(nav, /const pick = bar\.classList\.contains\('tight'\) && \(links\.children\.length > 1 \|\| admin && !!user\(\)\);/, 'when they do not fit: «current page ▾» opens the list');
  assert.match(nav, /\.mbar\.tight \.mtitle\{display:flex\}/);
  assert.doesNotMatch(nav, /¿Tienes cuenta\?/, 'Login lives in the bar only, not repeated in the menu');
  assert.match(nav, /body:has\(\.top>\.mbar\) \.addbar\{display:none!important\}/, 'the public page drops its second row');
  assert.match(nav, /\.top:has\(>\.mbar\)>\.appnav,\.top:has\(>\.mbar\) \.textsize,\.top:has\(>\.mbar\)>\.topright\{display:none!important\}/, 'on a phone the separate pieces fold into the menu');
  assert.match(nav, /<div class="msuser"><b>\$\{esc\(name\)\}\$\{roleChips\}<\/b>\$\{copyId \? '<button type="button" data-m="copyid" class="msoft">Copiar mi ID<\/button>' : ''\}<button type="button" data-m="logout">/);
  assert.match(nav, /if \(e\.target\.closest\('\[data-m="logout"\]'\)\) \{ showSheet\(false\); const out = document\.getElementById\('logout'\) \|\| document\.getElementById\('codeLogout'\); if \(out\) out\.click\(\); else sharedLogout\(\); return; \}/, 'Salir uses the page\'s own sign-out (with its confirmation)');
  assert.match(nav, /<div class="msize"><span class="lbl">Tamaño del texto<\/span>/);
  assert.match(nav, /\.mbar,\.msheet\{display:none\}/, 'computer header unchanged');
});

test('packing list dialog opens straight to the search: no visible title or point line (the card already says where)', () => {
  const pf = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
  assert.match(pf, /#pfPackTitle,\.pfp-pick>label\{position:absolute;width:1px;height:1px;[^}]*clip:rect\(0 0 0 0\)[^}]*\}\.pfp-for\{display:none\}/, 'title kept for screen readers only');
  assert.match(pf, /<dialog id="pfPack" class="pfp" aria-labelledby="pfPackTitle">/);
});

test('public «Resumen»: disclaimer that the numbers are community volunteer estimates', () => {
  const html = readFileSync(new URL('./index.html', import.meta.url), 'utf8');
  assert.match(html, /<p class="sumnote" role="note"><svg[^]*?<\/svg><span>Cifras reunidas por voluntarios: úsalas como guía, no como datos oficiales\.<\/span><\/p>/);
});

test('runner «+ Agregar punto» opens the point form here (full screen on a phone, Cancelar top left), not the public page', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  assert.match(runner, /<dialog id="addDialog" class="adddlg" aria-labelledby="addTitle"><form id="addForm" novalidate><div class="addhead"><button type="button" class="addcancel" id="addCancel">Cancelar<\/button><h2 id="addTitle">Agregar punto<\/h2><button type="submit" class="addsave">Guardar<\/button><\/div>/);
  assert.match(runner, /document\.addEventListener\('click',e=>\{const a=e\.target\.closest\('a\[href="\/#agregar"\]'\);if\(!a\)return;e\.preventDefault\(\);openAdd\(\)\},true\);/, 'header link and the phone «+» both open it');
  assert.match(runner, /if\(code\)\{let r=await post\('\/api\/full\/points'\);if\(r\.status===401\)/, 'code users publish; the map session is renewed with the code');
  assert.match(runner, /return post\('\/api\/submissions'\)\}/, 'without a code it goes to review');
  assert.match(runner, /@media\(max-width:760px\)\{\.adddlg\{inset:0;width:auto;height:auto;max-width:none;max-height:none;margin:0;border-radius:0\}/, 'full screen on a phone, without vw units that the text-size zoom would inflate');
});

test('runner card icons: megaphone for who reported it, saluting figure for the coordinator', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  assert.match(runner, /flagIco=svgI\('<path d="m3 11 18-5v12L3 14v-3z"\/>/, 'megaphone');
  assert.match(runner, /saluteIco=svgI\(/);
  assert.doesNotMatch(runner, /ico\('coordinator'\)/, 'coordinator lines use the salute');
});

test('public page on a phone, signed out: «Reportar punto», «Inicia sesión», then ☰; no title or search', () => {
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  assert.match(nav, /classList\.toggle\('public', !user\(\) && !!document\.getElementById\('loginBtn'\)\)/);
  assert.match(nav, /\.mbar\.public \.mlinks,\.mbar\.public \.mtitle,\.mbar\.public \.msearchbtn\{display:none!important\}\.mbar\.public \.madd\{order:-1\}/);
  assert.match(nav, /\$\{document\.getElementById\('loginBtn'\) \? '<button type="button" class="mlogin" id="mLogin">Inicia sesión<\/button>' : ''\}/);
  assert.match(nav, /if \(e\.target\.closest\('#mLogin'\)\) \{ showSheet\(false\); document\.getElementById\('loginBtn'\)\?\.click\(\); return; \}/);
  assert.match(template, /id="loginBtn">Inicia sesión<\/button>/);
  assert.match(template, /\$\('#loginBtn'\)\.hidden=!!u;/, 'hidden once signed in');
});

test('header color system: orange only for help/add, teal confirms sign-in, ghost buttons for secondary header controls', () => {
  const index = readFileSync(new URL('./index.html', import.meta.url), 'utf8');
  assert.match(index, /\.navbtn\{background:none;color:#fff;[^}]*border:1px solid #ffffff55;/, 'Login is a ghost button, not the browser grey');
  assert.match(index, /#loginSubmit\{background:#19745f\}/, '«Continuar» matches the gate «Entrar»');
  assert.match(index, /else \$\('\.top \.brand'\)\.after\(nav\)\}/, 'desktop page links next to the logo, like the other pages');
});

test('phone: no flash of the old header while a page loads (header stays empty until the app bar is mounted)', () => {
  for (const page of ['index.html', 'runner.html', 'builder.html', 'command.html']) {
    const html = readFileSync(new URL('./' + page, import.meta.url), 'utf8');
    assert.ok(html.includes('@media(max-width:760px){.top:not(:has(>.mbar))>*{visibility:hidden}}'), page);
    assert.ok(html.indexOf('.top:not(:has(>.mbar))') < html.indexOf('</style>'), page + ': in the page CSS, before any script');
  }
  assert.match(readFileSync(new URL('./index.html', import.meta.url), 'utf8'), /\.addbar\{display:none!important\}\.app\{height:calc\(calc\(100dvh \/ var\(--text-zoom,1\)\) - 112px\)!important\}/, 'the old second row never shows on a phone');
});

test('offline queue: points saved without internet wait in the browser and send themselves later, once', () => {
  const pf = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
  const index = readFileSync(new URL('./index.html', import.meta.url), 'utf8');
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  assert.match(pf, /record\.clientRef \|\|= newRef\(\);\s*if \(navigator\.onLine === false\) \{ queuePoint\(record\); enterOffline\(true\); return \{queued: true\}; \}/, 'offline: straight to the queue and into offline mode');
  assert.match(pf, /try \{ r = await send\(record\); \} catch \{ queuePoint\(record\); enterOffline\(true\); return \{queued: true\}; \}/, 'network error: queued');
  assert.match(pf, /navigator\.clipboard\?\.writeText\(JSON\.stringify\(record, null, 2\)\)/, 'its data is copied to the clipboard');
  assert.match(pf, /if \(r\.ok\) \{ qSave\(qLoad\(\)\.filter\(x => x\.ref !== item\.ref\)\); sent\+\+; continue; \}/, 'removed only once the server confirms');
  assert.match(pf, /x\.error = body\.error \|\| 'Error HTTP ' \+ r\.status;/, 'refused points stay with the reason');
  assert.match(pf, /addEventListener\('online', \(\) => \{ if \(offOn\) checkOnline\(\); else flushQueue\(\); \}\);/, 'sends when the connection comes back');
  assert.match(index, /const out=await PointForm\.sendOrQueue\(record,submitPoint\);if\(out\.queued\)\{/);
  assert.match(index, /PointForm\.setQueue\(\{send:submitPoint,sent:\(\)=>refresh\(\)\}\);/);
  assert.match(runner, /const out=await PointForm\.sendOrQueue\(record,sendPoint\);if\(out\.queued\)\{/);
  assert.match(runner, /PointForm\.setQueue\(\{send:p=>sendPoint\(p\),sent:\(\)=>refresh\(\)\}\);/);
});

test('offline mode: a full-screen lock with only the saved points (edit, add), a badge, and «Reintentar envío»', () => {
  const pf = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
  assert.match(pf, /<div id="pfOffline" class="pfo" role="dialog" aria-modal="true" aria-labelledby="pfoTitle" hidden><div class="pfo-head">[^']*<h2 id="pfoTitle">Sin conexión<\/h2><span class="pfo-badge" id="pfoBadge"/);
  assert.match(pf, /function lockPage\(on\) \{ for \(const el of document\.body\.children\) if \(el\.id !== 'pfOffline'[^}]*\) el\.inert = on; \}/, 'nothing behind it can be used');
  assert.match(pf, /addEventListener\('offline', \(\) => \{ if \(addFormOpen\(\)\) enterOffline\(\); \}\);/, 'losing the connection locks it only while adding a point');
  assert.match(pf, /if \(navigator\.onLine !== false\) flushQueue\(\);/, 'opening a page offline does not lock it');
  assert.match(pf, /<button type="button" class="pfo-go" data-o="check">Reintentar envío<\/button>/);
  assert.match(pf, /if \(!l\.length\) \{ leaveOffline\(/, 'only leaves once every point was sent');
  assert.match(pf, /render\(document\.getElementById\('pfoForm'\), \{point: existing \? existing\.record : offEdit\.draft \|\| \{\}\}\);/, 'tap a point to edit it');
  assert.match(pf, /if \(x\) \{ record\.clientRef = x\.ref; x\.record = record; x\.error = ''; qSave\(l\); \}/, 'edits keep the same clientRef (no duplicate)');
  assert.match(pf, /if \(!saved && Object\.entries\(draft\)/, 'an open form is carried over as a draft (unless it was just queued)');
});

test('GPS without a map: live accuracy lock instead of the map; GPS starts warming up as soon as the form opens', () => {
  const pf = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
  assert.match(pf, /const noMap = \(\) => navigator\.onLine === false \|\| !window\.L/, 'offline or no map library');
  assert.match(pf, /if \(noMap\(\)\) \{ gpsLock\(root\); return; \}/);
  assert.match(pf, /acc <= 15 \? '#23834f' : acc <= 50 \? '#d69a1b' : '#c62828'/, 'green / amber / red accuracy');
  assert.match(pf, /if \(!best \|\| pos\.coords\.accuracy <= best\.coords\.accuracy\) best = pos;/, 'uses the most precise reading');
  assert.match(pf, /box\.hidden = !\(c && inPanama\(c\) && window\.L\) \|\| noMap\(\);/, 'no map thumbnail offline');
  assert.match(pf, /setMe\(root, me\);\s*drawGpsThumb\(root\);\s*if \(!admin\) \{ warmGps\(\); gpsLive\(root\); \}/, 'GPS starts when the form opens, with progress shown in the form');
  assert.match(pf, /<div class="pf-gpslive" role="status"><span class="pf-gl-dot"><\/span><span class="pf-gl-t">Buscando señal GPS…<\/span>/, 'live progress under the GPS button, without blocking');
  assert.match(pf, /if \(inPanama\(c\) && \(gpsEl\.value === '' \|\| gpsEl\.value === root\._autoGps\) && acc <= shown - 5\) \{/, 'the location fills itself with better readings, unless you chose one');
  assert.match(pf, /gpsEl\.addEventListener\('input', \(\) => \{ root\._autoGps = null; \}\);/, 'typing a link stops the automatic fill');
  assert.match(pf, /let best = freshFix\(\), n = best \? 1 : 0;/, 'the lock screen starts from the warm fix');
  assert.match(pf, /timeout: 20000, maximumAge: 15000\}\);/, 'the map uses a fix from the last 15 s right away');
});

test('«Ubicar en mapa» fills the screen at any text size (sized from the screen edges, no vw/dvh units)', () => {
  const pf = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
  assert.match(pf, /dialog\.pf-mapdialog\{inset:0;width:auto;height:auto;max-width:none;max-height:none;margin:0;/);
  assert.doesNotMatch(pf, /dialog\.pf-mapdialog\{[^}]*(vw|dvh)/);
});

test('runner without signal on a run: locked «Sin señal · Tu recorrido» screen, photos and «Completé» kept on the phone and sent in order', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  assert.match(runner, /<div id="runOffline" class="ro" role="dialog" aria-modal="true" aria-labelledby="roTitle" hidden><div class="ro-head">[^']*<h2 id="roTitle">Sin señal · Tu recorrido<\/h2><span class="ro-badge" id="roBadge"/);
  assert.match(runner, /function enterRunOffline\(\)\{if\(agentMode\|\|!mine\.length\|\|roOn\|\|document\.querySelector\('#pfOffline:not\(\[hidden\]\)'\)\)return;/, 'only with runs in progress (and not over the add-point offline screen)');
  assert.match(runner, /addEventListener\('offline',\(\)=>enterRunOffline\(\)\);/);
  assert.match(runner, /catch\{if\(typeof enterRunOffline==='function'\)enterRunOffline\(\);throw Error\('Sin conexión\.'\)\}/, 'a failed request also locks it');
  assert.match(runner, /indexedDB\.open\('runnerOffline',1\)/, 'photos are kept in IndexedDB (too big for localStorage)');
  assert.match(runner, /await roAdd\(\{kind:'complete',needId:n\.id,at:Date\.now\(\)\}\)/, '«Completé» is queued after its delivery photos');
  assert.match(runner, /for\(const x of \(await roAll\(\)\)\.sort\(\(a,b\)=>a\.key-b\.key\)\)/, 'sent in the order they were made');
  assert.match(runner, /if\(!err\.status\|\|err\.status>=500\)break;/, 'no signal or server down: wait and retry');
  assert.match(runner, /if\(!left\.length\)\{leaveRunOffline\(/, 'only unlocks once everything was sent');
  assert.match(runner, /<a class="ro-call" href="tel:/, 'plain phone calls work without data');
});

test('builders: like Runners but for house repairs — tabs, progress %, what is missing, materials, before/progress/final photos', () => {
  const builder = readFileSync(new URL('./builder.html', import.meta.url), 'utf8');
  assert.match(builder, /<nav class="maintabs" id="mainTabs"[^>]*><button[^>]*data-main="mine"[^>]*><span aria-hidden="true">🔨<\/span> Mis obras <b id="cntMine"><\/b><\/button><button[^>]*data-main="open"[^>]*><span aria-hidden="true">🚩<\/span> Obras <b id="cntOpen"><\/b><\/button><\/nav>/);
  assert.match(builder, /<input type="range" class="prog" min="0" max="100" step="5" value="\$\{p\}" data-progress="\$\{id\}"/, 'progress slider');
  assert.match(builder, /<textarea class="scope" data-scope="\$\{id\}"/, 'what is still needed');
  assert.match(builder, /PointForm\.openPacking\(\{subtitle:'Materiales',items,save:/, 'materials with the packing-list picker');
  assert.match(builder, /const stages=\[\['before','Antes'\],\['progress','Avance'\],\['after','Final'\]\]/);
  assert.match(builder, /'\/api\/builder\/needs\/'\+encodeURIComponent\(id\)\+'\/media\?stage='\+stage/);
  assert.match(builder, /if\(mediaOn&&!skipPhoto&&!\(n\.media\|\|\[\]\)\.some\(m=>m\.stage==='after'\)\)\{/, 'finishing asks for a final photo');
  assert.match(builder, /saveProject\(r\.dataset\.progress,\{progress:Number\(r\.value\)\},'ps'\)/);
});

test('screen-size units follow the text zoom on every page (no window or panel wider/taller than the phone at 130 %)', () => {
  for (const f of ['index.html', 'runner.html', 'builder.html', 'command.html', 'admin.html', 'nav.js', 'pointform.js']) {
    const src = readFileSync(new URL('./' + f, import.meta.url), 'utf8');
    assert.doesNotMatch(src.replace(/calc\(\d+(\.\d+)?(dvh|svh|lvh|vh|vw) \/ var\(--text-zoom,1\)\)/g, ''), /(?<![\w.-])\d+(\.\d+)?(dvh|svh|lvh|vh|vw)\b/, f + ': every vw/vh unit is divided by --text-zoom');
  }
});

test('admin on a phone: logo, round «+» and ☰ on top; search on its own full-width row', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  assert.match(admin, /@media\(max-width:760px\)\{header\{padding:10px 12px;gap:8px;min-height:0\}header \.topactions\{margin-left:auto;[^}]*\}header \.topsearch\{order:5;flex:1 0 100%;/);
  assert.match(admin, /#addPoint\{width:42px;height:42px;padding:0!important;font-size:0;/);
  assert.match(admin, /#addPoint::before\{content:"\+";/);
});
test('command center: focusing a point with the map hidden does not animate it (Leaflet would fail with size 0)', () => {
  const command = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  assert.match(command, /if\(map&&hasCoords\(p\)\)\{if\(map\.getContainer\(\)\.offsetWidth\)map\.flyTo\(p\.coords,Math\.max\(map\.getZoom\(\),15\),\{duration:\.6\}\);else mapFocus=p\.coords\}/);
  assert.match(command, /if\(mapFocus\)\{map\.setView\(mapFocus,/, 'centered when the map tab is shown');
});

test('recurring camp services: admin section with countdown, done/edit/remove; same status helper everywhere; no bathroom need box', () => {
  const pf = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  assert.doesNotMatch(pf.match(/const NEEDS = [^\n]*/)[0], /Saneamiento|baño/, 'no bathroom need box (the portable toilet is a service)');
  assert.match(pf, /const SERVICES = \[\['Recolección de basura', 'trash', 7\], \['Cambio de baño portátil', 'toilet', 7\], \['Fumigación de mosquitos', 'spray', 14\], \['Ronda médica', 'medicine', 30\]\];/);
  const status = new Function(`${pf.match(/  function serviceStatus\(n, now = Date\.now\(\)\) \{[^]*?\n  \}/)[0]};return serviceStatus`)();
  const now = Date.parse('2026-10-15T12:00:00Z'), day = 864e5;
  assert.deepEqual(status({status: 'OPEN', nextDueAt: new Date(now + 3 * day).toISOString()}, now), {cls: 'ok', text: 'En 3 días'});
  assert.deepEqual(status({status: 'OPEN', nextDueAt: new Date(now + 20 * 3600e3).toISOString()}, now), {cls: 'soon', text: 'Mañana'});
  assert.deepEqual(status({status: 'OPEN', nextDueAt: new Date(now - 3600e3).toISOString()}, now), {cls: 'due', text: 'Toca hoy'});
  assert.deepEqual(status({status: 'OPEN', nextDueAt: new Date(now - 2.5 * day).toISOString()}, now), {cls: 'late', text: 'Atrasado 2 días'});
  assert.deepEqual(status({status: 'IN_PROGRESS'}, now), {cls: 'route', text: 'En camino'});
  assert.match(admin, /<div class="svcbox" id="serviceBox"><div class="svchead"><h3>Servicios del punto<\/h3>/);
  assert.match(admin, /adminRequest\('\/api\/admin\/camps\/'\+encodeURIComponent\(activeCamp\)\+'\/needs','POST',\{\.\.\.body,urgency:'normal',kind:'service'\}\)/);
  assert.match(admin, /await adminRequest\('\/api\/admin\/needs\/'\+encodeURIComponent\(id\),'PATCH',\{status:'COMPLETED'\}\);/, '«✓ Hecho» completes it; the server schedules the next one');
});



test('builders: «Evaluar» a job before (or without) taking it — work checkboxes + custom, materials, days, people, notes, photos', () => {
  const builder = readFileSync(new URL('./builder.html', import.meta.url), 'utf8');
  assert.match(builder, /const WORK=\['Techo \/ zinc','Paredes \/ bloque','Piso \/ losa','Estructura \/ columnas','Puertas y ventanas','Electricidad','Plomería \/ agua','Baño \/ letrina','Soldadura','Repello \/ pintura','Limpieza \/ escombros'\];/);
  assert.match(builder, /data-assess="\$\{clean\(n\.id\)\}">📋 \$\{[^}]*\?'Mi evaluación':'Evaluar'\}<\/button><button class="btn take"/, 'Evaluar next to Tomar obra; they are separate');
  assert.match(builder, /<label class="aother">Otro trabajo<input name="other"/, 'custom work besides the checkboxes');
  assert.match(builder, /'\/api\/builder\/needs\/'\+encodeURIComponent\(n\.id\)\+'\/assessments'\+\(assessMine\?'\/'\+encodeURIComponent\(assessMine\.id\):''\)/, 'new or edit your own');
  assert.match(builder, /'\/media\?stage=inspect&ref='\+encodeURIComponent\(x\.id\)/, 'photos tied to the assessment');
  assert.match(builder, /data-usemats="\$\{id\}">Usar materiales de la evaluación<\/button>/, 'whoever takes it can reuse the materials');
});

test('home page on main has no «¿Qué donar?» tab (donations live on donations.html and the donations-wip branch)', () => {
  const html = readFileSync(new URL('./index.html', import.meta.url), 'utf8');
  assert.doesNotMatch(html, /data-side="donate"|data-view="donate"|id="donate"|drawDonate|¿Qué donar/);
});

test('Donaciones page: only admin or the «Donaciones» role (code); tickets with status steps, logistics and Maps link', () => {
  const page = readFileSync(new URL('./donations.html', import.meta.url), 'utf8');
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  assert.match(page, /fetch\('\/api\/desk\/session',\{cache:'no-store'/, 'role session first');
  assert.match(page, /fetch\('\/api\/admin\/session'/, 'admin session also opens it');
  assert.doesNotMatch(page, /\/api\/agent/, 'agents need the role, not their agent session');
  assert.match(page, /necesita el rol «Donaciones»/);
  assert.match(admin, /<input type="checkbox" name="roles" value="donations"> 🎁 Donaciones<\/label>/);
  assert.match(nav, /\['\/donations\.html', '🎁', \(\) => 'Donaciones', u => has\(u, 'donations'\)\]/);
  for (const s of ['Por confirmar', 'Confirmada', 'En acopio', 'En camino', 'Entregada', 'Cancelada']) assert.ok(page.includes(s), s);
  for (const f of ['dropPlace', 'dropAt', 'transporter', 'transporterPhone', 'campId', 'dispatchAt', 'destination']) assert.match(page, new RegExp(`name="${f}"`), f);
  assert.match(page, /https:\/\/www\.google\.com\/maps\/search\/\?api=1&query=/, 'device-appropriate Maps link, no embedded map');
  assert.match(page, /const left=Math\.max\(0,a\.need-a\.committed\)|left=Math\.max\(0,a\.need-a\.committed\)/, 'only committed subtracts');
});

test('admin «Por revisar»: check reports (one by one or all in view) and export them to CSV with every field', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  assert.match(admin, /<option value="coord">Reportado por<\/option><\/select><\/label><div class="pendbulk" id="pendBulk" hidden><\/div><\/div>/);
  assert.match(admin, /<input type="checkbox" class="pendcheck" data-pendsel="\$\{clean\(p\.id\)\}"/, 'a checkbox on each pending card');
  assert.match(admin, /bar\.hidden=view!=='pending'\|\|!pendVisible\.length/, 'only in «Por revisar»');
  assert.match(admin, /for\(const id of pendVisible\)e\.target\.checked\?pendSel\.add\(id\):pendSel\.delete\(id\)/, '«Seleccionar» = the reports in the current view (after search and filters)');
  assert.match(admin, /if\(!points\.some\(p=>p\.id===id&&p\.status==='pending'\)\)pendSel\.delete\(id\)/, 'approved or rejected reports drop out of the selection');
  const src = admin.match(/const PEND_COLS=[^\n]*/)[0] + '\n' + admin.match(/function pendingCsv\(list\)[^\n]*/)[0];
  const csv = new Function(`${src};return pendingCsv`)()([{id: 'x1', code: 'P-1', area: 'Tonosí', place: 'Escuela, "la vieja"', coords: [7.4, -80.4], phone: '6000-0000', needs: 'Agua, Alimentos', extra: 'línea 1\nlínea 2', members: [{name: 'Ana', age: '7', cedula: '8-1-1'}]}]);
  assert.ok(csv.startsWith('﻿ID,Código,Recibido'), 'BOM so Excel reads the accents');
  const head = csv.slice(1).split('\r\n')[0], row = csv.slice(1).split('\r\n').slice(1).join('\r\n');
  assert.equal(head.split(',').length, 29, 'every field, including Familias');
  assert.ok(row.startsWith('x1,P-1,,,Tonosí,"Escuela, ""la vieja""",,,7.4,-80.4,'), row);
  assert.ok(row.includes('"Agua, Alimentos"') && row.includes('"línea 1\nlínea 2"') && row.includes('Ana, 7 años, cédula 8-1-1'), row);
  assert.match(admin, /a\.download=`por-revisar-\$\{new Date\(\)\.toISOString\(\)\.slice\(0,10\)\}-\$\{list\.length\}\.csv`/);
});

test('admin «Importar reportes (por revisar)»: CSV from the export or a JSON list; always ?pending=1, never published', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  assert.match(admin, /<button type="button" role="menuitem" id="importPendingToggle"><span class="mi" aria-hidden="true">[^<]*<\/span>Importar reportes \(por revisar\)…<\/button>/);
  assert.match(admin, /adminRequest\('\/api\/admin\/points\?pending=1','POST',p\)/, 'the pending route only');
  const imp = admin.slice(admin.indexOf('// Importar reportes:'), admin.indexOf('function render(){'));
  assert.doesNotMatch(imp, /'\/api\/admin\/points','POST'|\/api\/admin\/points'?,'POST'/, 'never the publishing route');
  assert.match(admin, /clientRef:await rowRef\(r\.p\)/, 're-importing the same rows does not duplicate');
  const grab = re => admin.match(re)[0];
  const src = [/const fold=[^;]*;/, /const PEND_COLS=[^\n]*/, /function pendingCsv\(list\)[^\n]*/, /const IMPORT_FIELDS=[^\n]*/, /function parseCsv\(text\)[^\n]*/].map(grab).join('\n') + '\n' + admin.slice(admin.indexOf('function importRows('), admin.indexOf('let importList=[];'));
  const {pendingCsv, importRows} = new Function('points', `${src};return {pendingCsv, importRows}`)([{id: 'old', code: 'P-taken'}]);
  const src1 = {id: 'x1', code: 'P-taken', created: '2026-10-15', status: 'pending', area: 'Tonosí', place: 'Escuela, "la vieja"', coords: [7.4, -80.4], phone: '6000-0000', people: '12', needs: 'Agua, Alimentos', extra: 'línea 1\nlínea 2', note: 'privada', addedByName: 'Arianne', members: [{name: 'Ana', age: '7'}]};
  const [row] = importRows('x.csv', pendingCsv([src1]));
  assert.equal(row.id, 'x1');
  assert.deepEqual(row.p.coords, [7.4, -80.4]);
  assert.equal(row.p.place, 'Escuela, "la vieja"');
  assert.equal(row.p.extra, 'línea 1\nlínea 2');
  assert.equal(row.p.needs, 'Agua, Alimentos');
  assert.equal(row.p.code, undefined, 'a code already in use is dropped');
  assert.ok(!('status' in row.p) && !('created' in row.p) && !('id' in row.p), 'no status, dates or id sent');
  assert.match(row.p.note, /^privada\nPersonas del grupo \(del archivo\): Ana, 7 años$/);
  const [j] = importRows('c.json', JSON.stringify([{area: 'Flores', gps: 'https://maps.app.goo.gl/x', coords: [7, -80], members: [{name: 'B'}], status: 'approved', addedBy: 'u1'}]));
  assert.deepEqual(j.p, {area: 'Flores', gps: 'https://maps.app.goo.gl/x', members: [{name: 'B'}], addedBy: 'u1'}, 'JSON: link wins over coords; status ignored');
});

test('«Salir» on any page forgets every page copy of the code (no silent sign-in again); the remembered code stays for «!»', () => {
  for (const f of ['runner.html', 'builder.html']) {
    const page = readFileSync(new URL('./' + f, import.meta.url), 'utf8');
    const out = page.slice(page.indexOf("$('#logout').onclick"), page.indexOf("$('#logout').onclick") + 900);
    assert.match(out, /\['runnerCode','builderCode','agentCode'\]\.forEach\(k=>localStorage\.removeItem\(k\)\);/, f);
    assert.doesNotMatch(out, /rememberedCode/, f + ' keeps the remembered code');
  }
  const cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  assert.match(cmd, /\['agentCode','codeUser','runnerCode','runnerName','builderCode','builderName'\]/);
});

test('admin Usuarios: email (for code recovery) can be edited and searched', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  assert.match(admin, /<label>Correo <small>\(para recuperar el código\)<\/small><input name="email" type="email" maxlength="160"/);
  assert.match(admin, /f\.elements\.email\.value=u\.email\|\|'';/);
  assert.match(admin, /email:form\.elements\.email\.value\.trim\(\),/);
  assert.match(admin, /fold\(u\.name\+' '\+u\.telephone\+' '\+\(u\.email\|\|''\)\+' '\+u\.id\)/, 'find a person by email to give them a new code');
});


test('sign-in everywhere: «Ya tengo código» and «Crear cuenta nueva» side by side, nothing hidden; code remembered and prefilled', () => {
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  const pages = {index: template, runner: readFileSync(new URL('./runner.html', import.meta.url), 'utf8'), builder: readFileSync(new URL('./builder.html', import.meta.url), 'utf8'), command: readFileSync(new URL('./command.html', import.meta.url), 'utf8')};
  const codeGroup = /<fieldset class="sg sg-code"><legend>Ya tengo código<\/legend><label>Código de acceso<input name="accessCode"[^>]*><\/label><label class="remember"><input type="checkbox" name="remember" checked> Recordar mi código de acceso<\/label><button[^>]*value="code"[^>]*>Entrar<\/button>/;
  const newGroup = /<fieldset class="sg sg-new"><legend>Crear cuenta nueva<\/legend><label>Nombre completo<input name="name"[^>]*><\/label><label>Teléfono<input name="telephone"[^>]*><\/label><label>Correo <small>\(opcional, para recuperar tu código\)<\/small><input name="email" type="email"[^>]*><\/label><button[^>]*value="new">Crear mi código<\/button><\/fieldset>/;
  for (const [k, page] of Object.entries(pages)) {
    const form = page.match(/<form id="(loginForm|gateForm)" novalidate data-signin>[^]*?<\/form>/)[0];
    assert.match(form, codeGroup, k);
    if (k !== 'command') assert.match(form, newGroup, k);
    assert.doesNotMatch(form, / hidden[ >]/, k + ': no hidden fields');
  }
  for (const page of [template, pages.runner, pages.builder]) assert.doesNotMatch(page, /loginPhoneRow|loginEmailRow|syncCodeLogin|syncLoginMode|phone\.style\.display/, 'nothing toggles fields any more');
  assert.doesNotMatch(nav, /phone\.style\.display|syncCodeLogin/);
  // Which group was sent: the button, or with Enter the group holding the cursor.
  assert.match(nav, /function signinMode\(e\) \{\n    const clicked = e\.target\?\.dataset\?\.clicked;\n    if \(clicked\) return clicked === 'new' \? 'new' : 'code';\n    return document\.activeElement\?\.closest\?\.\('\.sg-new'\) \? 'new' : 'code';/, 'Enter in «Crear cuenta nueva» creates the account, not the code sign-in');
  assert.match(nav, /if \(b\?\.form && \(e\.detail > 0 \|\| document\.activeElement === b\)\)/, 'the browser click that Enter fires on the first button does not count');
  assert.match(template, /mode=window\.AppNav\?AppNav\.signinMode\(e\):/);
  for (const k of ['runner', 'builder']) assert.match(pages[k], /const f=e\.target\.elements,mode=AppNav\.signinMode\(e\)/, k);
  // New account: approved at once, signs in with the new code; Runners/Constructores pass their panel.
  assert.match(nav, /fetch\('\/api\/session\/signup', \{method: 'POST'[^\n]*body: JSON\.stringify\(\{name, telephone, email, \.\.\.\(panel \? \{panel\} : \{\}\)\}\)/);
  assert.match(pages.runner, /made=await AppNav\.createAccount\(\{name,telephone,email:f\.email\.value\.trim\(\),panel:'runner'\}\);code=made\.code/);
  assert.match(pages.builder, /made=await AppNav\.createAccount\(\{name,telephone,email:f\.email\.value\.trim\(\),panel:'builder'\}\);code=made\.code/);
  assert.match(pages.runner, /if\(err\.field!=='telephone'\)throw err;await startSession\(\{name,telephone\}\)/, 'people registered before codes still get in with name and phone');
  assert.match(template, /const \{code,name\}=await AppNav\.createAccount\(\{name:f\.elements\.name\.value,telephone:f\.elements\.telephone\.value,email:f\.elements\.email\.value\.trim\(\)\}\);\$\('#newCode'\)\.textContent=code;\$\('#newName'\)\.textContent=name;try\{await signInCode\(code\)\}catch\{\}loginStep\('done'\)/);
  assert.match(template, /<div id="loginDone" hidden><h2>¡Listo, ya entraste!<\/h2><p>Tu cuenta: <b id="newName"><\/b><\/p>/, 'shows the saved name (with a number if someone had the same name)');
  // Remembered after a successful sign-in, prefilled next time (Login, gates, after «Salir»).
  assert.match(template, /if\(!r\.ok\)throw Error\(body\.error\|\|'No se pudo entrar\.'\);window\.AppNav\?\.codeSignedIn\(code\);/);
  assert.match(nav, /if \(localStorage\.getItem\('rememberCode'\) !== '0'\) \{ localStorage\.setItem\('rememberCode', '1'\); localStorage\.setItem\('rememberedCode', code\);/);
  assert.match(nav, /else \{ localStorage\.removeItem\('rememberedCode'\); setTimeout\(\(\) => codeKeys\.forEach\(k => localStorage\.removeItem\(k\)\), 4000\); \}/, 'unchecked forgets it');
  assert.match(nav, /function prefillCode\(form\) \{\n    const box = form\?\.elements\?\.accessCode, saved = rememberedCode\(\);\n    if \(!box \|\| !saved \|\| box\.value\.trim\(\)\) return;/, 'never overwrites what someone typed');
  assert.match(template, /function openLogin\(\)\{\$\('#loginForm'\)\.reset\(\);window\.AppNav\?\.prefillCode\(\$\('#loginForm'\)\);/);
  for (const k of ['runner', 'builder', 'command']) assert.match(pages[k], /window\.AppNav\?\.prefillCode\(/, k);
  assert.match(nav, /document\.addEventListener\('reset', e => \{ if \(e\.target\.matches\?\.\(signinForms\)\) setTimeout\(\(\) => prefillAll\(\)\); \}\);/, 'after «Salir» the code comes back');
  // A code pasted into «Nombre completo» moves to its own field.
  assert.match(nav, /if \(e\.target\.name === 'name' && f\.elements\.accessCode && isCode\(e\.target\.value\)\) \{ f\.elements\.accessCode\.value = codeOf\(e\.target\.value\);/);
  const helpers = nav.match(/const isCode = [^\n]*/)[0] + '\n' + nav.match(/const codeOf = [^\n]*/)[0];
  const {isCode, codeOf} = new Function(helpers + ';return {isCode, codeOf}')();
  for (const v of ['!loro47juan3', 'loro47juan3', '!jadersurf3', '!PRUEBACORRE']) assert.equal(isCode(v), true, v);
  for (const v of ['Juan Pérez', 'María', '6000-0000', '!']) assert.equal(isCode(v), false, v);
  assert.equal(codeOf(' !jadersurf3 '), 'jadersurf3', 'existing codes with «!» reach the server exactly as before');
});

test('runners map: choosing a run keeps the map on it even with live GPS; «🎯 Ir a mi ubicación» follows again; weak signal does not turn live off', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  assert.match(runner, /function applyFocus\(\)\{const pin=pinsByCamp\.get\(focusCamp\);if\(!map\|\|!pin\|\|!pendingFocus\)return;pendingFocus=false;setFollow\(false\);/, 'choosing a run stops following the person');
  assert.match(runner, /syncFollowBtn\(\);if\(!follow\)\{first=false;return\}/, 'GPS updates never pull the map away from the chosen run (not even the first fix)');
  assert.match(runner, /map\.on\('dragstart',\(\)=>\{lastPan=Date\.now\(\);setFollow\(false\)\}\)/, 'moving the map by hand also stops following');
  assert.match(runner, /<button type="button" class="followbtn" id="followMe" hidden>🎯 Ir a mi ubicación<\/button>/);
  assert.match(runner, /\$\('#followMe'\)\.onclick=\(\)=>\{setFollow\(true\);if\(map&&me\)map\.setView\(me,Math\.max\(map\.getZoom\(\),15\)\)\};/);
  assert.match(runner, /let first=true;setFollow\(true\);watchId=navigator\.geolocation\.watchPosition\(/, 'turning live on follows the person');
  assert.match(runner, /if\(err\.code!==1\)return;stopLive\(\);flash\('Permite el acceso a tu ubicación para verla en el mapa\.'\)/, 'only a denied permission turns live location off');
});

test('runners: «¡Yo lo llevo!» switches to «Yo» with the taken run highlighted and in view (from the map too)', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  const take = runner.slice(runner.indexOf('async function takeNeed('), runner.indexOf('function flash('));
  assert.match(take, /highlightNeed=id;setMainTab\('mine'\);await refresh\(\);/);
  assert.match(take, /if\(window\.matchMedia\('\(max-width:760px\)'\)\.matches&&document\.body\.classList\.contains\('mobile-map'\)\)mobileView\('list'\);/, 'from a map pin on a phone, back to the list');
  assert.match(take, /scrollIntoView\(\{behavior:'smooth',block:'center'\}\)/);
});

test('runners: a run chosen in the list opens on the map with its card (¡Yo lo llevo!) already open; other map moves do not', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  assert.match(runner, /focusCamp=n\.camp\.id;pendingFocus=true;popupNeedId=n\.id;render\(\);/, 'only the list sets it');
  assert.equal((runner.match(/popupNeedId=n\.id/g) || []).length, 1);
  assert.match(runner, /const n=popupNeedId&&findNeed\(popupNeedId\);popupNeedId=null;if\(n&&n\.camp\.id===focusCamp\)\{pin\.setPopupContent\(popupShort\(n\)\);pin\.openPopup\(\)\}/);
  const short = runner.match(/function popupShort\(n\)\{[^\n]*/)[0];
  assert.match(short, /<button class="btn take" type="button" data-take=/, '«¡Yo lo llevo!» is the one main button');
  assert.match(short, /class="poplink" data-list=/, 'Ver en la lista as a small link');
  assert.doesNotMatch(short, /needDetails|data-back/, 'no needs list, contacts or extra links');
});

test('runners: «¡Yo lo llevo!» asks first (point + what to carry); «Cancelar» takes nothing and keeps the map card open', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  const take = runner.slice(runner.indexOf('async function takeNeed('), runner.indexOf('function flash('));
  assert.match(take, /if\(await AppNav\.choose\('¿Llevas esta carrera\?',\{detail:off\?[^,]*:detail,options:\[\['take','✓ Sí, la llevo','ok'\],\['no','Cancelar',''\]\]\}\)!=='take'\)return false\}button\.disabled=true;/, 'confirm before the request');
  assert.ok(take.indexOf("AppNav.choose(") < take.indexOf("/take'"), 'nothing is sent before confirming');
  assert.match(runner, /takeNeed\(b\.dataset\.take,b\)\.then\(ok=>\{if\(!ok\)return;/, 'cancel in the map card does not close it');
});

test('Acopio (supplier.html): runs in progress with the runner, packing, departure photos and «Despachada» (a step inside «En camino»)', () => {
  const page = readFileSync(new URL('./supplier.html', import.meta.url), 'utf8');
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  assert.match(page, /<title>Acopio · Azuero<\/title>/);
  assert.match(page, /<button type="button" role="tab" data-main="todo"[^>]*><span aria-hidden="true">📦<\/span> Empacando <b id="cntTodo"><\/b><\/button><button type="button" role="tab" data-main="asks"/, 'no «Despachadas» tab: dispatched runs are in the history');
  assert.match(page, /if\(mainTab==='hist'\)\{const way=done\(\)\.filter\(match\),all=histRecent\(\),list=all\.filter\(match\);[^\n]*?feed\.innerHTML=\(way\.length\?`<h3 class="listhead">🚚 Despachadas · en camino<\/h3>\$\{way\.map\(runCard\)\.join\(''\)\}`:''\)/);
  assert.match(page, /<input class="topsearch" id="find" type="search" data-global="1" placeholder="Buscar en todo…"/, 'find the runner who just arrived (header search, like other pages)');
  assert.match(page, /<form id="gateForm" novalidate data-signin><h1>Entrar a Supply<\/h1><fieldset class="sg sg-code">/);
  assert.doesNotMatch(page.match(/<form id="gateForm"[^]*?<\/form>/)[0], /sg-new/, 'admin gives the role; no self sign-up here');
  for (const api of ['/api/supplier/runs', '/api/supplier/packing-items', "/api/supplier/needs/'+encodeURIComponent(n.id)+'/packing", "/api/supplier/needs/'+encodeURIComponent(id)+'/media?stage=pickup", "/dispatch','POST',{undo:true}"]) assert.ok(page.includes(api), api);
  assert.doesNotMatch(page, /data-complete|data-take|data-release|\/api\/runner\//, 'no Completé, no taking or leaving runs, no runner endpoints');
  assert.match(page, /AppNav\.choose\('¿Sale del acopio este run\?'/, 'confirm before dispatching');
  assert.match(page, /<button class="btn dispatch" type="button" data-dispatch="\$\{id\}">🚚 Despachada<\/button>/);
  // Card in three parts: runner (who, phone, when, live position) · destination (point, contacts, needs, map) · what to load (packing, photos, dispatch).
  assert.match(page, /function runCard\(n\)\{return `<article class="scard split[^`]*\$\{needSection\(n\)\}\$\{loadSection\(n\)\}\$\{destSection\(n,false,true\)\}\$\{runnerSection\(n\)\}<\/article>`\}/);
  assert.match(page, /secHead\('h-runner','🏃','Runner'/);
  assert.match(page, /<b class="rn">\$\{clean\(n\.runnerName\|\|'Runner'\)\}<\/b>\$\{n\.runnerPhone\?/, 'who carries it, with call buttons');
  assert.match(page, /La tomó \$\{clean\(agoEs\(n\.enRouteAt\)\)\}/);
  assert.match(page, /data-runnerlive="\$\{clean\(n\.id\)\}">📍 Ver runner en el mapa<\/button>/, 'live position on the map when shared');
  assert.match(page, /secHead\('h-dest','📍','Destino'[^\n]*\$\{needDetails\(\{\.\.\.n,packing:null,noWhat:split\},true,false\)\}\$\{addedByLine\(n\)\}\$\{share\?`<div class="actrow">\$\{needButtons\(n\)\}/, 'point, contacts and needs (no packing here), see on map');
  assert.match(page, /secHead\('h-load','📦','Para cargar'[^\n]*PointForm\.packingHtml\(n\.packing\)[^\n]*\$\{mediaStrip\(n\)\}/, 'packing and departure photos together');
  assert.match(admin, /<input type="checkbox" name="roles" value="supplier"> 📦 Supply<\/label>/);
  assert.match(admin, /n\.status==='IN_PROGRESS'&&n\.dispatchedAt\?'Despachada · en camino'/);
  assert.match(nav, /\['\/supplier\.html', '📦', \(\) => 'Supply', u => has\(u, 'supplier'\)\]/);
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  assert.match(runner, /🚚 Despachada desde acopio/, 'the runner sees it on the card');
});

test('Supply ↔ Runners: Supply asks for a runner («📣 Pedidos» for runners, with who/where/phone) or assigns one in person; only linked runs to dispatch', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  const page = readFileSync(new URL('./supplier.html', import.meta.url), 'utf8');
  // Runners: requests are their own tab, not in Runs; accepting is «¡Yo lo llevo!»; «Yo» shows where to pick up.
  assert.match(runner, /requests=\(\)=>available\.filter\(n=>n\.requestedAt&&n\.status==='OPEN'&&matches\(n\)\)/);
  assert.match(runner, /if\(tab==='req'\)\{const req=requests\(\);feed\.innerHTML=req\.length\?req\.map\(n=>`<article class="card reqcard[^`]*\$\{offersBox\(n,full\)\}/, 'each offering center with its own «¡Yo lo llevo! · recojo aquí»');
  assert.match(runner, /data-take="\$\{id\}" data-from="\$\{clean\(o\.supplierId\)\}"/);
  assert.match(runner, /'\/take'\+\(from\?'\?supplier='\+encodeURIComponent\(from\):''\)/);
  assert.match(runner, /function supplierLine\(n,asking\)\{const x=n\.supplier;[^\n]*\$\{asking\?'Pide':'Recoger en acopio'\}: <b>\$\{clean\(x\.center\|\|'Centro de acopio'\)\}<\/b> · \$\{clean\(x\.name\)\}/);
  assert.match(runner, /\$\{n\.kind==='mealrun'\?mealLine\(n\):supplierLine\(n,false\)\}\$\{alsoLine\(n\)\}\$\{supplyAskBox\(n\)\}\$\{n\.dispatchedAt\?/, 'pickup line on the «Yo» card');
  const tabNow = new Function('mine', 'mainTab', `return (${runner.match(/const tabNow=([^\n;]*);/)[1]})()`);
  assert.equal(tabNow([], 'req'), 'req');
  // Supply: three tabs, center name, request/cancel, in-person assign with a runner picker.
  assert.match(page, /data-main="req"[^>]*><span aria-hidden="true">🙌<\/span> Puntos con necesidades <b id="cntReq"><\/b>/);
  assert.match(page, /<button type="button" class="cbtn" id="centerBtn" aria-haspopup="listbox"[^>]*><\/button><div class="cpop" id="centerPop" hidden><input id="centerInput"[^>]*role="combobox"[^>]*><div class="clist" id="centerList" role="listbox"><\/div><button type="button" class="cmanage" id="centerManage">⚙ Administrar centros…<\/button><\/div>/, 'center is a button («📍 Villa Marina ▾»); the typeahead list opens only to change it');
  assert.match(page, /\(q&&!exact\?`<button type="button" role="option" class="cnew" data-center="\$\{clean\(typed\)\}">＋ Usar «\$\{clean\(typed\)\}» \(centro nuevo\)<\/button>`/, 'type a new center');
  assert.match(page, /\$\('#centerList'\)\.onclick=e=>\{const b=e\.target\.closest\('\[data-center\]'\);if\(b\)saveCenter\(b\.dataset\.center\)\};/);
  assert.doesNotMatch(page, /id="heading"|id="sub"/, 'no heading or intro text taking space');
  for (const api of ["/request','POST',{})", "/request','POST',{cancel:true})", "(assignAnother?'/another':'/assign'),'POST',{runnerId:r.id})", "/api/supplier/center','POST'"]) assert.ok(page.includes(api), api);
  assert.match(page, /<dialog id="assignDlg"[^>]*>[^]*?<input class="sfind" id="assignFind"[^>]*>[^]*?<div class="runnerpick" id="assignList"/);
  assert.match(page, /\$\{mine\?`<button class="btn assign" type="button" data-cancelreq="\$\{id\}">Retirar oferta<\/button>`:`<button class="btn request" type="button" data-request="\$\{id\}">🙋 ¡Yo abastezco!<\/button>`\}<button class="btn assign" type="button" data-assign="\$\{id\}">🏃 Asignar a runner<\/button>/, 'several centers can offer the same need; the runner chooses');
});

test('runners: the «Completé» confirmation shows the packing list being delivered', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  assert.match(runner, /\$\{packing\?`<div class="confirmpack"><b>📦 Lo que entregas<\/b>\$\{\(need\.packing\|\|\[\]\)\.length\?PointForm\.packingHtml\(need\.packing\):'<p class="muted">Sin lista de empaque\.<\/p>'\}<\/div>`:''\}/);
  assert.match(runner, /askConfirm\('¿Estás seguro de que quieres completar esta carrera\?',\{yes:'Sí, completar',need:done,media:mediaOn,packing:true\}/);
  assert.match(runner, /askConfirm\('¿Completaste esta carrera\?',\{yes:'Sí, completar',need:n,media:mediaOn,packing:true,/, 'offline too');
  assert.doesNotMatch(runner.match(/askConfirm\('¿Estás seguro de que no puedes continuar[^)]*\)/)[0], /packing/, 'not when leaving the run');
});

test('runner cards: «Necesita» (what the point needs) apart from «📦 Lo que llevas» (the runner packing list)', () => {
  for (const f of ['runner.html', 'supplier.html']) {
    const page = readFileSync(new URL('./' + f, import.meta.url), 'utf8');
    assert.match(page, /(?:\$\{n\.noWhat\?'':`)?<div class="grp g-what"><span class="needlbl">Necesita<\/span>\$\{needChips\(n\.need\)\}<\/div>(?:`\})?\$\{\(n\.packing\|\|\[\]\)\.length&&typeof PointForm!=='undefined'\?`<div class="grp g-carry"><span class="needlbl">📦 Lo que llevas<\/span>\$\{PointForm\.packingHtml\(n\.packing\)\}<\/div>`:''\}/, f);
  }
});

test('runners: «Carrera completada» offers «📝 Actualizar necesidad» (opens the form) or «✓ No necesitan nada más»', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  assert.match(runner, /<div class="fuchoice" id="fuChoice"><button class="btn update" type="button" id="openFollowup">📝 Actualizar necesidad<\/button><button class="btn ok" type="button" id="skipFollowup">✓ No necesitan nada más<\/button><\/div><div id="fuForm" hidden>/);
  assert.match(runner, /<button class="btn" type="submit">Guardar necesidad<\/button><button class="btn plain" type="button" id="backFollowup">Volver<\/button><\/div>/);
  assert.match(runner, /if\(e\.target\.id==='openFollowup'\)\{\$\('#fuChoice'\)\.hidden=true;\$\('#fuForm'\)\.hidden=false;/);
  assert.match(runner, /if\(e\.target\.id==='skipFollowup'\)\{followup=null;await refresh\(\)\}/);
});

test('Runners ask supply centers; Supply answers «Puedo surtirla» / «No puedo»; the runner picks the closest', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  const page = readFileSync(new URL('./supplier.html', import.meta.url), 'utf8');
  const box = runner.match(/function supplyAskBox\(n\)\{[^\n]*/)[0];
  assert.match(box, /if\(n\.supplier\|\|n\.kind==='mealrun'\)return '';/, 'only runs without a supply center');
  assert.match(box, /data-supplyask="\$\{id\}">📣 Pedir a acopios<\/button>/);
  assert.match(box, /acopios pueden'\} surtirla: elige el más cercano/);
  assert.match(box, /data-choose="\$\{id\}" data-supplier="\$\{clean\(o\.supplierId\)\}">Elegir<\/button>/);
  assert.match(box, /data-supplycancel="\$\{id\}">Cancelar pedido<\/button>/);
  for (const api of ["/supply-ask','POST',{})", "/supply-ask','POST',{cancel:true})", "/supply-choose','POST',{supplierId:o.supplierId})"]) assert.ok(runner.includes(api), api);
  assert.match(page, /data-main="asks"[^>]*><span aria-hidden="true">🏃<\/span> Solicitudes <b id="cntAsk"><\/b>/);
  const card = page.match(/function askCard\(n\)\{[^\n]*/)[0];
  assert.match(card, /\$\{runnerSection\(n\)\}\$\{destSection\(n\)\}/);
  assert.match(card, /data-offer="\$\{id\}">✓ Puedo surtirlo<\/button><button class="btn assign" type="button" data-decline="\$\{id\}">✗ No puedo<\/button>/);
  assert.match(card, /data-withdraw="\$\{id\}">Retirar<\/button>/);
  for (const api of ["/offer','POST',{can:true})", "/offer','POST',{can:false})", "/offer','POST',{withdraw:true})"]) assert.ok(page.includes(api), api);
});

test('runner and Supply cards: no real location → the Maps button is disabled next to the crossed pin; Supply says Runner/run', () => {
  for (const f of ['runner.html', 'supplier.html']) {
    const page = readFileSync(new URL('./' + f, import.meta.url), 'utf8');
    assert.match(page, /\$\{link&&!link\.approximate\?`<a class="iconbtn" target="_blank"[^`]*`:`<span class="iconbtn off" title="Sin ubicación: no se puede abrir en Maps" role="img" aria-label="Sin ubicación para Maps">\$\{navIco\}<\/span>`\}/, f);
  }
  const page = readFileSync(new URL('./supplier.html', import.meta.url), 'utf8');
  const code = page.replace(/\/\*[^]*?\*\//g, '').split('\n').map(l => l.replace(/(^|[;}])\s*\/\/.*$/, '$1')).join('\n');
  assert.doesNotMatch(code, /[Cc]orredor|[Cc]arrera|surtirla/, 'Supply uses Runner / run');
});

test('access stays current: pages check their session every 45 s; new roles show in the menu, a lost page sends you home', () => {
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  const w = nav.slice(nav.indexOf('function watchAccess('), nav.indexOf('// Al llegar al inicio porque cambió el acceso'));
  assert.match(w, /setInterval\(check, 45000\);/);
  assert.match(w, /document\.addEventListener\('visibilitychange', \(\) => \{ if \(!document\.hidden\) check\(\); \}\);/, 'and when coming back to the tab');
  assert.match(w, /if \(Array\.isArray\(b\.panels\)\) \{ saveAccess\(b\); if \(page && !b\.panels\.some\(p => p\.url === page\)\) leave\(\); \}/);
  assert.match(w, /if \(r\.status !== 401 \|\| !armed \|\| !page\) return;/, 'only after the page had a session');
  assert.match(w, /fetch\('\/api\/session\/code'/, 'with the saved code it signs in again and keeps the pages it still has');
  assert.match(w, /const leave = \(\) => \{ location\.href = '\/\?acceso=0'; \};/);
  assert.match(nav, /Tu acceso cambió: ya no puedes usar esa página\. Tus páginas están en el menú\./);
  for (const [f, url, page] of [['runner.html', '/api/runner/session', '/runner.html'], ['builder.html', '/api/builder/session', '/builder.html'], ['command.html', '/api/agent/session', '/command.html'], ['supplier.html', '/api/supplier/session', '/supplier.html'], ['donations.html', '/api/desk/session', '/donations.html'], ['index.html', '/api/full/me', null]]) {
    const html = readFileSync(new URL('./' + f, import.meta.url), 'utf8');
    assert.ok(html.includes(`window.AppNav?.watchAccess('${url}',${page ? `'${page}'` : 'null'});`), f);
  }
});

test('several runners on the same need: Supply sends «otro runner» (request or in person) from «Ya en camino»; cards show who else goes; «Solo mi equipo» on Supply', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  const page = readFileSync(new URL('./supplier.html', import.meta.url), 'utf8');
  for (const [f, html] of [['runner', runner], ['supplier', page]]) assert.match(html, /function alsoLine\(n\)\{const a=n\.also\|\|\[\];if\(!a\.length\)return '';return `<div class="alsoline">👥 También va:/, f);
  assert.match(runner, /reqcard[^`]*\$\{alsoLine\(n\)\}\$\{needDetails\(n,false\)\}\$\{offersBox\(n,full\)\}/, 'also on «Pedidos» cards');
  assert.match(page, /<h3 class="listhead">🚚 Ya en camino · si hace falta, manda otro runner<\/h3>\$\{way\.map\(enrouteCard\)\.join\(''\)\}/);
  assert.match(page, /data-another="\$\{id\}">🙋 Abastezco otro envío<\/button><button class="btn assign" type="button" data-anotherassign="\$\{id\}">🏃 Asignar otro runner<\/button>/);
  assert.match(page, /\(assignAnother\?'\/another':'\/assign'\),'POST',\{runnerId:r\.id\}\)/);
  assert.match(page, /<label class="tswitch" title="Solo puntos y runners de mi equipo"><input type="checkbox" id="teamOnly"><span class="sw" aria-hidden="true"><\/span><span>Solo mi equipo<\/span><\/label>/, 'a switch on the same line as the center');
  assert.match(page, /const teamOk=n=>!teamOnly\|\|\(mainTab==='req'\?!!n\.camp\.team:!!\(n\.camp\.team\|\|n\.runnerTeam\)\);/, 'my team points (and my team runners on runs)');
  assert.match(page, /list=runners\.filter\(r=>\(!teamOnly\|\|r\.team\)&&/, 'only my team runners in the picker');
});

test('command center on phones: opened point stacks full width; title is the point name; counters say what they count; edit form uses the full width', () => {
  const cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  assert.match(cmd, /\.row\{grid-template-columns:1fr;grid-template-areas:'main' 'stats' 'needs' 'report' 'detail';padding:12px 14px\}/, 'the opened detail is part of the one-column phone layout');
  assert.match(cmd, /\.row\.editing\{padding:0;border:0;background:none;box-shadow:none\}/);
  assert.match(cmd, /\$\{clean\(pointName\(p\)\)\}\$\{p\.addedByName\?'':vb\(p,\{text:true\}\)\}\$\{p\.code&&!autoCode\(p\.code\)\?` <small class="pcode">/);
  const {autoCode, pointName} = new Function(cmd.match(/const autoCode=[^\n]*/)[0] + ';return {autoCode, pointName}')();
  assert.equal(pointName({area: 'la Colorada', code: 'P-f1bc49236515'}), 'la Colorada');
  assert.equal(pointName({area: 'Sin ubicación', place: 'Escuela'}), 'Escuela');
  assert.ok(autoCode('P-f1bc49236515') && !autoCode('AB-6'));
});

test('Usuarios: click a header to sort (again to reverse); newest first by default; Roles filter checks several roles', () => {
  const src = [admin.match(/let userSortKey=[^\n]*/)[0], 'const userSel=new Set();let userTeamFilter="";', admin.match(/function userResults\(\)\{[^\n]*/)[0]].join('\n');
  const users = [{id: 'a', name: 'Bea', roles: ['runner'], createdAt: '2026-10-02'}, {id: 'b', name: 'Ana', roles: [], createdAt: '2026-10-05', lastActiveAt: '2026-10-14'}, {id: 'c', name: 'Ceci', roles: ['supplier', 'agent'], createdAt: '2026-10-01', online: true}];
  const el = {'#userSearch': {value: ''}, '#userShowDisabled': {checked: false}, '#userOnlyCode': {checked: false}};
  const run = (setup) => new Function('$', 'users', 'fold', 'roleNames', `const roleLabel=r=>roleNames[r]||r;${src};${setup};return userResults().map(u=>u.name)`)(k => el[k], users, x => String(x).toLowerCase(), {runner: 'Corredor', agent: 'Agente', supplier: 'Acopio'});
  assert.deepEqual(run(''), ['Ana', 'Bea', 'Ceci'], 'created, newest first');
  assert.deepEqual(run("userSortDir='asc'"), ['Ceci', 'Bea', 'Ana']);
  assert.deepEqual(run("userSortKey='name';userSortDir='asc'"), ['Ana', 'Bea', 'Ceci']);
  assert.deepEqual(run("userSortKey='activity';userSortDir='desc'"), ['Ceci', 'Ana', 'Bea'], 'online first, then last activity');
  assert.deepEqual(run("userRoleFilter.add('runner');userRoleFilter.add('__none__')"), ['Ana', 'Bea'], 'any of the checked roles, «Sin acceso» = no roles');
  assert.match(admin, /if\(userSortKey===k\)userSortDir=userSortDir==='asc'\?'desc':'asc';else\{userSortKey=k;/);
});

test('signup.html: a link to send people; signed out → code or new account; signed in → their pages', () => {
  const page = readFileSync(new URL('./signup.html', import.meta.url), 'utf8');
  assert.match(page, /<form id="signinForm" novalidate data-signin><fieldset class="sg sg-code"><legend>Ya tengo código<\/legend>[^]*<fieldset class="sg sg-new"><legend>Crear cuenta nueva<\/legend>/);
  assert.match(page, /fetch\('\/api\/session\/code',\{method:'POST'/, 'signs in on every page at once');
  assert.match(page, /const \{code\}=await AppNav\.createAccount\([^\n]*?\);showDone\(await signIn\(code\),code\)/, 'new account: signed in, code shown');
  assert.match(page, /fetch\('\/api\/full\/me'/, 'already signed in: shows where to go');
  assert.match(page, /p=>`<a class="btn go" href="\$\{clean\(p\.url\)\}">/);
  const main = readFileSync(new URL('./main.go', import.meta.url), 'utf8');
  assert.match(main, /"\/signup\.html"/, 'served');
  assert.match(readFileSync(new URL('./Dockerfile', import.meta.url), 'utf8'), /signup\.html/, 'shipped');
});

test('team manager: «Mi equipo» only in the ☰ menu on phones (desktop: page menu); ☰ shows the user roles', () => {
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  assert.match(nav, /\['\/team\.html', '👥', \(\) => 'Mi equipo', u => has\(u, 'teammanager'\) \|\| has\(u, 'admin'\), 'menu'\]/);
  assert.match(nav, /pages\.filter\(\(\[url, , , ok, where\]\) => url === here \|\| ok\(u\) && where !== 'menu'\)/, 'not in the phone page bar');
  assert.match(nav, /const menuOnly = signedIn && u \? pages\.filter\(\(\[url, , , ok, where\]\) => where === 'menu' && url !== here && ok\(u\)\)/);
  assert.match(nav, /const roleChips = signedIn && u\?\.roles\?\.length \? `<span class="msroles">/);
});

test('«Mi equipo» (team.html): invite links (one use, expiring), members with roles, change roles, remove; invite opens signup.html', () => {
  const page = readFileSync(new URL('./team.html', import.meta.url), 'utf8');
  const signup = readFileSync(new URL('./signup.html', import.meta.url), 'utf8');
  assert.match(page, /<select id="hours" aria-label="Vence en"><option value="24">Vence en 24 horas<\/option><option value="72" selected>Vence en 3 días<\/option><option value="168">Vence en 7 días<\/option><\/select>/);
  assert.match(page, /const url=location\.origin\+'\/signup\.html\?invite='\+encodeURIComponent\(r\.token\)/);
  for (const api of ["'/api/full/team/invites','POST'", "/revoke','POST'", "/remove','POST'", "'PATCH',{roles}"]) assert.ok(page.includes(api), api);
  assert.match(page, /\$\{!data\.manager\?'':`<div class="acts"><button type="button" class="btn small ghost" data-notes="\$\{clean\(m\.id\)\}">📝 Notas<\/button>\$\{self\?'':`<button type="button" class="btn small ghost" data-roles=/, 'not on yourself');
  assert.match(page, /\$\{can\?'':' disabled'\}/, 'admin-only roles cannot be changed here');
  assert.match(signup, /let invite=new URLSearchParams\(location\.search\)\.get\('invite'\)\|\|'';/);
  assert.match(signup, /fetch\('\/api\/full\/invites\/accept',\{method:'POST'/);
  assert.match(signup, /showDone\(user\);await acceptInvite\(\);return\}showEnter\(\)/, 'already signed in: joins right away');
  assert.match(signup, /showDone\(await signIn\(code\),code\);await acceptInvite\(\)/, 'new account: joins after creating it');
  assert.match(page, /const ROLE_ICONS=\{runner:'🏃',builder:'🔨',agent:'🧑‍💻',donations:'🎁',supplier:'📦',teammanager:'👥',cook:'🍳',admin:'🔒'\}/, 'role chips carry the page icons');
  assert.ok((page.match(/clean\(roleLabel\(r\)\)/g) || []).length === 2, 'icons on member chips and in the roles dialog');
  assert.match(page, /<input type="search" class="msearch" id="memberSearch"/);
  assert.match(page, /\$\('#memberSearch'\)\.oninput=renderMembers;/, 'filters as you type');
  assert.match(page, /\$\{ms\.length\} de \$\{all\.length\}/, 'count shows matches of total');
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  assert.match(admin, /<input type="checkbox" name="roles" value="teammanager"> 👥 Team manager <small>\(necesita equipo\)<\/small><\/label>/);
});

test('Runners: header «Buscar» like the home page (🔍 pop-up on phones) filters the open tab and the map', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  assert.match(nav, /\$\{top\.querySelector\('\.topsearch'\) \? `<button type="button" class="msearchbtn" id="mSearch"/, 'nav.js adds the 🔍 button wherever the header has .topsearch');
  assert.match(runner, /<input class="topsearch" id="search" type="search"/);
  assert.match(runner, /foldNeed\(\[c\.code,c\.area,c\.place,c\.coordinator,n\.need,x\.center,x\.name,n\.runnerName,/, 'code, zone, place, need, coordinator, supply center, runner');
  assert.match(runner, /el\.hidden=!hits\.some\(n=>n\.id===el\.dataset\.run\)/, '«Yo» hides non-matching runs without renumbering stops');
  assert.match(runner, /\$\('#tabFind'\)\.addEventListener\('input',render\);/);
  assert.equal((runner.match(/searchQ\(\)\?noHits\(\)/g) || []).length, 3, '«Nada coincide» on Runs, Pedidos and agent lists');
});

test('needs with a brand: «Tarjeta de datos prepago» (Tigo / +Móvil) and «Gas de cocina» (Tropigas / Panagas), pick or type another', () => {
  for (const label of ['Tarjeta de datos prepago', 'Gas de cocina']) assert.ok(PointForm.needsHtml().includes(`data-need="${label}"`), label);
  assert.deepEqual(PointForm.BRANDED['Tarjeta de datos prepago'].brands.map(b => b[0]), ['Tigo', '+Móvil'], 'active networks in Panamá');
  assert.deepEqual(PointForm.BRANDED['Gas de cocina'].brands.map(b => b[0]), ['Tropigas', 'Panagas']);
  assert.match(PointForm.needsHtml(), /<div class="pf-datanet" data-brandfor="Gas de cocina" hidden><span class="pf-datanet-label">Empresa <small>toca una o varias<\/small><\/span><div class="pf-pills"[^>]*><button type="button" class="pf-pill" data-pill="Tropigas" aria-pressed="false"><i class="pf-swatch pf-net-tropigas"><\/i><b>Tropigas<\/b><\/button>[^]*?<button type="button" class="pf-pill pf-pill-other" data-pill-other aria-pressed="false"><svg[^]*?<\/svg><b>Otra<\/b><\/button><\/div><input class="pf-other-input" maxlength="40" placeholder="Escribe la empresa"[^>]*hidden><input type="hidden" class="pf-net-input" data-brand="Gas de cocina">/, 'pills like Agua, plus «Otra» with a short field');
  const p = PointForm.parseNeeds('Agua, Tarjeta de datos prepago (Tigo), Gas de cocina (Otra Gas), tarjeta de datos prepago');
  assert.deepEqual([...p.checked], ['Agua', 'Tarjeta de datos prepago', 'Gas de cocina']);
  assert.deepEqual({...p.brands}, {'Tarjeta de datos prepago': 'Tigo', 'Gas de cocina': 'Otra Gas'});
  assert.match(PointForm.needChip('Tarjeta de datos prepago (claro)'), /pf-net-masmovil" title="\+Móvil">\+móvil</, 'Claro is now +Móvil');
  assert.match(PointForm.needChip('Gas de cocina (panagas)'), /<span class="chip"><svg[^]*?<\/svg>Gas de cocina<span class="pf-netbadge pf-net-panagas" title="Panagas">Panagas<\/span><\/span>/);
  assert.match(PointForm.needChip('Gas de cocina (Gas Sur)'), /<span class="pf-netbadge" title="Gas Sur">Gas Sur<\/span>/, 'a typed company keeps its name');
  assert.equal(PointForm.needChip('Arroz'), '');
  // Agua and Alimentos: tap one or both options (no typing): «Agua (Para beber / Para cocinar)».
  assert.deepEqual(PointForm.BRANDED.Agua.brands.map(b => b[0]), ['Para beber', 'Para cocinar']);
  assert.deepEqual(PointForm.BRANDED.Alimentos.brands.map(b => b[0]), ['Comida preparada', 'Ingredientes para cocinar']);
  assert.match(PointForm.needsHtml(), /<div class="pf-datanet" data-brandfor="Agua" hidden>[^]*?<button type="button" class="pf-pill" data-pill="Para beber" aria-pressed="false"><svg[^]*?<\/svg><b>Para beber<\/b>/, 'option icons instead of descriptions');
  assert.deepEqual(PointForm.BRANDED.Alimentos.brands.map(b => b[4]), ['meal', 'pan'], 'hot meal and pan icons');
  assert.deepEqual(PointForm.BRANDED['Reparación de casa'].brands.map(b => b[0]), ['Hojas de zinc', 'Carriola', 'Varilla', 'Tubo de hierro', 'Bolsas de cemento', 'Bloques', 'Tornillos', 'Clavos', 'Soldadura']);
  assert.ok(PointForm.needsHtml().includes('data-brandfor="Reparación de casa"'), 'repair materials in every form');
  assert.match(PointForm.needChip('Agua (Para beber / Para cocinar)'), /Agua<span class="pf-netbadge pf-net-beber" title="Para beber">para beber<\/span><span class="pf-netbadge pf-net-cocinar" title="Para cocinar">para cocinar<\/span>/);
  assert.deepEqual({...PointForm.parseNeeds('Alimentos (Comida preparada), Agua').brands}, {Alimentos: 'Comida preparada'});
  // Grouped list, two columns; a checked need with options opens them right under it, full width.
  const groups = [...PointForm.needsHtml({admin: true}).matchAll(/<p class="pf-needgroup">([^<]+)<\/p>/g)].map(m => m[1]);
  assert.deepEqual(groups, ['Agua y comida', 'Refugio y cuidado', 'Energía y comunicación', 'Casa', '¿Algo más?']);
  assert.match(pointformSrc, /\.pf-needitem\.has-opts:has\(input\[data-need\]:checked\)\{grid-column:1\/-1/);
  assert.equal(PointForm.needBase('Gas de cocina (Tropigas)'), 'Gas de cocina');
  // Generator fuel: one need, Gasolina by default, Diésel in the list, or type another; old two-box texts still read.
  assert.deepEqual(PointForm.BRANDED['Combustible para generador'].brands.map(b => b[0]), ['Gasolina', 'Diésel']);
  assert.equal(PointForm.BRANDED['Combustible para generador'].def, 'Gasolina');
  assert.ok(!PointForm.NEEDS.some(n => /Gasolina para generador|Diésel para generador/.test(n[0])), 'no separate boxes');
  assert.deepEqual({...PointForm.parseNeeds('Gasolina para generador, Diesel para generador').brands}, {'Combustible para generador': 'Gasolina / Diésel'});
  assert.match(PointForm.needChip('Diésel para generador'), /Combustible para generador<span class="pf-netbadge pf-net-diesel" title="Diésel">Diésel</);
  assert.match(pointformSrc, /if \(check\.checked && b\.def && !input\.value\) \{ on\(pills\.find\(p => p\.dataset\.pill === b\.def\), true\)/, 'checking fuel picks Gasolina');
  assert.match(pointformSrc, /if \(p === otherPill\) \{ otherIn\.hidden = !isOn\(p\); if \(isOn\(p\)\) otherIn\.focus\(\); \}/, '«Otra» opens the short field');
  assert.equal(PointForm.BRANDED['Gas de cocina'].other, 'Otra');
  assert.match(pointformSrc, /\.pf-other-input\[hidden\]\{display:none\}/, 'the short field stays hidden until «Otra»');
  assert.equal(PointForm.BRANDED['Tarjeta de datos prepago'].other, 'Otra');
  // readNeeds writes the brand in parentheses, normalized to the listed name.
  const box = (l, c) => ({dataset: {need: l}, checked: c});
  const boxes = [box('Agua', true), box('Tarjeta de datos prepago', true), box('Gas de cocina', true)];
  const inputs = {'Tarjeta de datos prepago': {value: 'tigo'}, 'Gas de cocina': {value: '  Gas (Sur), Norte '}};
  const root = {querySelectorAll: () => boxes.filter(b => b.checked), querySelector: s => { const m = /data-brand="([^"]+)"/.exec(s); return m ? inputs[m[1]] : s === '.pf-otros-check' ? {checked: false} : s === '.pf-otros-text' ? {value: ''} : null; }};
  assert.equal(PointForm.readNeeds(root), 'Agua, Tarjeta de datos prepago (Tigo), Gas de cocina (Gas Sur Norte)');
  for (const page of ['runner.html', 'supplier.html', 'command.html', 'index.html']) assert.match(readFileSync(new URL('./' + page, import.meta.url), 'utf8'), /(?:\.map\(part=>\{|function chipOf\(part,pf,types,icons\)\{)const dc=pf\.needChip\?pf\.needChip\(part\):'';if\(dc\)return dc;/, page + ' shows the brand on need chips');
  assert.match(admin, /PointForm\.BRANDED\[l\]\?PointForm\.brandBadge\(l,brands\[l\]\):''/);
});

test('home on phones opens on «📊 Resumen» (a link to a point still opens that point)', () => {
  assert.match(template, /if\(window\.matchMedia\('\(max-width:760px\)'\)\.matches&&!new URLSearchParams\(location\.search\)\.get\('punto'\)\)mobileView\('stats'\);/);
});

test('«Verificado por equipo»: gold badge on points and needs across pages; public pages without names', () => {
  const pfSrc = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
  const ctx = PointForm;
  assert.equal(ctx.verifiedBadge({}), '');
  assert.equal(ctx.verifiedBadge({verifiedAt: '2026-10-15T12:00:00Z'}, {text: true}).replace(/<svg[^]*<\/svg>/, ''), '<span class="pf-verified" title="Verificado" role="img" aria-label="Verificado"></span>', 'without team data (public, no team): just the icon, no details');
  assert.match(ctx.verifiedBadge({verifiedAt: '2026-10-15T12:00:00Z', verifiedTeam: 'Equipo Norte', verifiedBy: 'Ana'}, {text: true}), /^<button type="button" class="pf-verified tap" data-verified="Verificado por el equipo Norte · Ana · [^"]+" aria-expanded="false" aria-label="Verificado: ver qué equipo"><svg[^]*<\/svg><\/button>$/, 'team members: icon only; tapping it shows the team');
  assert.match(pfSrc, /pop\.textContent = b\.dataset\.verified;/);
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8'), supplier = readFileSync(new URL('./supplier.html', import.meta.url), 'utf8'), cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8'), camp = readFileSync(new URL('./camp.html', import.meta.url), 'utf8');
  assert.match(runner, /v=\(n\.verifiedAt\|\|''\)>\(c\.verifiedAt\|\|''\)\?n:c;[^\n]*PointForm\.verifiedBadge\(v,\{text:true\}\)/);
  assert.match(supplier, /const verifiedOf=n=>\{const c=n\.camp,v=\(n\.verifiedAt\|\|''\)>\(c\.verifiedAt\|\|''\)\?n:c;return [^\n]*PointForm\.verifiedBadge\(v\)/);
  assert.match(supplier, /<b>\$\{clean\(c\.area\|\|c\.code\)\}<\/b>\$\{n\.addedByName\?'':verifiedOf\(n\)\}/, 'Supply: badge in the title unless «Agregado por» shows it');
  assert.match(cmd, /\$\{vb\(n,\{text:true\}\)\}\$\{actions\?`<button[^`]*`:''\}<\/div>\$\{needChips\(n\.need\)\}/);
  assert.match(admin, /\$\{PointForm\.verifiedBadge\(n,\{text:true\}\)\}`;/);
  assert.match(template, /\$\{urgentBadge\(n\.urgency\)\}\$\{vb\(n\)\}\$\{when\}/);
  assert.match(camp, /<h1>\$\{clean\(p\.code\)\}\$\{verified\(p\)\}<\/h1>/);
  assert.doesNotMatch(camp, /Verificado por un equipo/, 'public point page: icon only');
  assert.match(camp, /title="Verificado" role="img" aria-label="Verificado"/);
});

test('builders see the materials requested with a house repair', () => {
  const builder = readFileSync(new URL('./builder.html', import.meta.url), 'utf8');
  assert.match(builder, /\$\{cardMeta\(n\)\}<\/div>\$\{reqMaterials\(n\)\}\$\{whereGroup\(n\)\}/);
  const b = PointForm.brandOf('Reparación de casa (Varilla / Otro tubo)');
  assert.match(PointForm.brandBadge(b.label, b.brand), /title="Varilla">varilla<\/span><span class="pf-netbadge" title="Otro tubo">Otro tubo<\/span>$/);
});

test('development environment (localhost / local network) is marked on every page: amber stripe, «DEV» pill, tab title', () => {
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  const fn = src => src.match(/const DEV_HOST=[^\n]*/)[0];
  assert.match(admin, /<script src="\.\/nav\.js"><\/script>/, 'admin.html now loads nav.js, which marks it');
  const DEV = new Function(`${fn(nav)};return DEV_HOST`)();
  for (const h of ['localhost', '127.0.0.1', '192.168.1.20', '10.0.0.5', '172.20.1.1', '::1', 'mi-pc.local', 'ab12-190.ngrok-free.app', 'x.ngrok.io', 'x.ngrok-free.dev']) assert.ok(DEV.test(h), h);
  for (const h of ['azueroapoyo.org', 'app.unidosporazuero.com', 'ngrok.com.evil.org', 'apoyo.example.com', '172.32.0.1', '8.8.8.8']) assert.ok(!DEV.test(h), h);
  assert.match(nav, /<div class=\\?"devstripe\\?" aria-hidden=\\?"true\\?"><\/div><span class=\\?"devpill\\?"/);
  assert.match(nav, /document\.title='DEV · '\+document\.title/);
  assert.match(nav, /if \(document\.readyState === 'loading'\) document\.addEventListener\('DOMContentLoaded', devMark\); else devMark\(\);/);
  assert.doesNotMatch(admin, /devMark\(\);/, 'admin.html relies on nav.js for the DEV mark');
});

test('phone menu button shows the signed-in person\'s initials; Comando always has «+ Agregar punto» at the top right', () => {
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8'), cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  const initialsOf = new Function(`${nav.match(/const initialsOf = [^\n]*/)[0]};return initialsOf`)();
  assert.equal(initialsOf('Marco Silva'), 'MS');
  assert.equal(initialsOf('ana ortega ruiz'), 'AO');
  assert.equal(initialsOf('Juan Pérez 2'), 'JP', 'duplicate-name number is not an initial');
  assert.equal(initialsOf('Ñandú'), 'Ñ');
  assert.equal(initialsOf(''), '');
  assert.match(nav, /btn\.innerHTML = ini \? `<span class="minit">\$\{esc\(ini\)\}<\/span>` : svg\(/);
  assert.match(nav, /btn\.setAttribute\('aria-label', ini \? `Menú de \$\{name\}` : 'Menú'\);/);
  assert.match(cmd, /<\/div><button type="button" class="add" id="addPointBtn">\+ Agregar punto<\/button><\/div><\/header>/, 'opens the form right in Comando (no trip to the home page)');
  assert.match(cmd, /const p=await send\('\/api\/agent\/points','POST',data\);/);
  assert.match(cmd, /PointForm\.render\(\$\('#addFields'\),\{nearby:[^}]*,me:await myMe\(\)\}\)/, 'team agents are the coordinator by default');
  const nav2 = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  assert.match(nav2, /@media\(max-width:760px\)\{\.top:has\(>\.mbar\) \.topsearch\{display:none!important\}\}/, 'the header search box hides only on phones (🔍 opens it); desktop keeps it');
});

test('Comando: Buscar in the header (🔍 on phones, like Runners); totals with icons in the «Puntos» tab', () => {
  const cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  assert.match(cmd, /<header class="top"><div class="brand">[^]*?<\/div><nav class="appnav" id="appNav" hidden><\/nav><input class="topsearch" id="search" type="search" placeholder="Buscar…"/);
  assert.doesNotMatch(cmd, /<input class="search" id="search"/, 'no second search box in the list');
  assert.match(cmd, /const counts=`<span class="tcnt" title="[^\n]*?">\$\{icoPinS\}/);
  assert.match(cmd, /\$\{pending\?`<span class="tcnt pend" title="\$\{pending\} por revisar">\$\{icoReviewS\}\$\{pending\}<\/span>`:''\}/);
  assert.match(cmd, /if\(ptab\)ptab\.innerHTML=`<span>Puntos<\/span>\$\{counts\}`;/);
  assert.match(cmd, /@media\(max-width:760px\)\{\.board \.sidehead \.sub\{display:none\}/, 'on phones the totals live in the tab');
});

test('Supply tabs: Empacando, Despachadas, Solicitudes, Puntos con necesidades (last); 🕓 history next to Lista/Mapa on phones', () => {
  const page = readFileSync(new URL('./supplier.html', import.meta.url), 'utf8');
  const order = [...page.match(/<nav class="maintabs" id="mainTabs"[^]*?<\/nav>/)[0].matchAll(/data-main="(\w+)"/g)].map(m => m[1]);
  assert.deepEqual(order, ['todo', 'asks', 'req', 'hist']);
  assert.match(page, /<button type="button" data-view="hist" class="mhist" aria-pressed="false" title="Historial" aria-label="Historial"><svg[^>]*><path d="M3 12a9 9 0 1 0 3-6\.7"\/>/, 'clock turning back');
  assert.match(page, /function mobileView\(view\)\{if\(view==='hist'\)\{mainTab='hist';highlightNeed=null;view='list'\}else if\(view==='list'&&mainTab==='hist'\)mainTab='todo';/);
  assert.match(page, /if\(mainTab==='hist'\)\{const way=done\(\)\.filter\(match\),all=histRecent\(\),list=all\.filter\(match\);[^\n]*?[^\n]*list\.map\(histCard\)/);
  assert.match(page, /<span class="hdone" title="[^"]*">✓ Entregada \$\{clean\(agoEs\(n\.completedAt\)\)\}<\/span>/);
  assert.match(page, /@media\(max-width:760px\)\{\.maintabs \.histtab\{display:none\}body\.in-hist #mainTabs\{display:none\}\}/, 'on phones history lives next to Lista/Mapa and shows only the history, no other tabs');
  assert.match(page, /markNav\(\);document\.body\.classList\.toggle\('in-hist',mainTab==='hist'\);/);
});

test('Supply map: team runners in live (only my team, strict rule) even without a run of this center', () => {
  const page = readFileSync(new URL('./supplier.html', import.meta.url), 'utf8');
  assert.match(page, /teamLive=data\.live\|\|\[\];/);
  assert.match(page, /for\(const r of teamLive\.filter\(r=>!shown\.has\(r\.id\)\)\)/, 'no duplicate truck for a runner already on a run');
  assert.match(page, /const fit=campPins\.length\?campPins:teamPins/, 'with no points, the map centers on the team');
});

test('runner cards: map pin and Maps top right; people count with breakdown before «Necesita»; no persistent selected color', () => {
  const grab = re => runner.match(re)[0];
  assert.match(grab(/function locButtons[^\n]*/), /<span class="cardloc">\$\{c\.coords\?`<button type="button" class="iconbtn gps" data-where=/);
  assert.doesNotMatch(grab(/function needButtons[^\n]*/), /data-where|iconbtn gps/, 'bottom row: share only');
  const peopleLine = new Function('clean', 'ico', 'stamp', 'agoEs', 'PointForm', `${grab(/const PPL=[^\n]*/)};${grab(/function peopleLine[^\n]*/)};return peopleLine`)(s => String(s), () => '<svg/>', () => 'x', () => 'hace 1 h', {icons: {adult: '<i>a</i>', child: '<i>c</i>', family: '<i>f</i>'}});
  const html = peopleLine({createdAt: '2026-10-15T10:00:00Z', camp: {people: '60', adults: '30', children: '25', babies: '0', families: '12'}});
  assert.match(html, /<span class="pc total"><svg\/><b>60<\/b> personas<\/span><span class="pc" title="30 adultos"><i>a<\/i><b>30<\/b> adultos<\/span><span class="pc" title="25 niños"><i>c<\/i><b>25<\/b> niños<\/span><span class="pc" title="12 familias"><i>f<\/i><b>12<\/b> familias<\/span>/, 'zeros left out');
  assert.match(runner, /\$\{peopleLine\(n\)\}<div class="grp g-what"><span class="needlbl">Necesita<\/span>/);
  assert.doesNotMatch(runner, /\.card\.selected\{background/);
});

test('Comando: «Agregado por» Todos / Míos / Mi equipo; on phones the other filters fold behind «Filtros ▾ (n)»', () => {
  const cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  assert.match(cmd, /<div class="filters" id="filters"><button type="button" class="morefilt" id="moreFilt" aria-expanded="false" aria-controls="fSheet">Filtros<b id="moreCount" hidden><\/b><span aria-hidden="true">▾<\/span><\/button><select id="sortBy" aria-label="Ordenar">[^]*?<div class="fsheet" id="fSheet" role="dialog" aria-label="Filtros"><div class="fsheethead"><b>Filtros<\/b><span class="fheadacts"><button type="button" class="fclear" data-fclear>Limpiar todo<\/button><button type="button" class="fdone" data-fdone>Listo<\/button><\/span><\/div>[^]*?<div class="seg whoseg" id="whoSeg"/, 'phones: «Filtros» opens a sheet; the sort stays visible as a dropdown');
  assert.match(agent, /\.filters\.open \.fsheet\{display:flex;flex-direction:column;gap:14px;position:fixed;inset:0;[^}]*border-radius:0;/, 'phones: filters take the whole screen');
  assert.match(agent, /<div class="fdone2"><button type="button" class="fclear" data-fclear>Limpiar todo<\/button><button type="button" class="fgo" data-fdone>Ver resultados<\/button><\/div>/, '«Limpiar todo» next to «Ver resultados» too');
  assert.match(agent, /go\.textContent=`Ver \$\{n\} \$\{n===1\?'punto':'puntos'\}`/, 'the bottom button says how many points');
  assert.match(agent, /else if\(e\.target\.closest\('#clearFilters,\[data-fclear\]'\)\)/);
  assert.match(cmd, /const whoTests=\{me:p=>!!userId&&p\.addedBy===userId,team:p=>!!myTeam\(\)&&!!p\.addedBy&&users\.find\(u=>u\.id===p\.addedBy\)\?\.teamId===myTeam\(\)\};/);
  assert.match(cmd, /\.\.\.\(myTeam\(\)\?\[\['team','Equipo',icoTeamS\]\]:\[\]\)/, '«Equipo» (my team) only for agents on a team');
  assert.match(cmd, /\(skip==='who'\|\|!filt\.who\|\|whoTests\[filt\.who\]\?\.\(p\)\)/);
  assert.match(cmd, /const active=\[filt\.need,filt\.area,filt\.type,filt\.gps,\$\('#search'\)\.value\.trim\(\)\]\.filter\(Boolean\)\.length;/);
  assert.match(agent, /<input class="fsearch" id="sheetSearch" type="search" placeholder="Buscar…"/, 'Filtros sheet includes the search');
  assert.match(agent, /\$\('#sheetSearch'\)\.oninput=e=>\{const top=\$\('#search'\);top\.value=e\.target\.value;top\.dispatchEvent\(new Event\('input',\{bubbles:true\}\)\)\};/, 'the badge counts active filters');
  assert.match(cmd, /\.fsheet\{display:none\}\.filters\.open::before\{content:'';position:fixed;inset:0;/, 'filters collapsed behind «Filtros» on phones');
  assert.match(cmd, /\.morebox \.pills\{flex-wrap:wrap;overflow:visible\}/, 'need chips wrap instead of scrolling off the edge');
});

test('Supply: «⚙ Administrar centros…» renames or archives (nothing is deleted, can be restored); team switch on the right', () => {
  const page = readFileSync(new URL('./supplier.html', import.meta.url), 'utf8');
  assert.match(page, /<dialog id="centersDlg" class="confirmdlg"[^>]*>[^]*?Archivar solo lo quita de la lista: no se borra nada/);
  assert.match(page, /AppNav\.confirm\(`¿Archivar «\$\{d\.carchive\}»\?`/, 'archive asks first');
  assert.match(page, /AppNav\.confirm\(`¿Cambiar «\$\{d\.crenameSave\}» por «\$\{to\}»\?`/, 'rename asks first');
  assert.match(page, /data-crestore="\$\{clean\(c\)\}">↺ Recuperar<\/button>/);
  assert.match(page, /api\('\/api\/supplier\/centers\/manage','POST',body\)/);
  assert.match(page, /\.tswitch\{margin-left:auto;/);
});

test('«Reportado por» (megaphone) replaces «Coordinador» in labels; the stored field stays coordinator', () => {
  const pub = PointForm.html();
  assert.match(pub, /<input name="coordinator"[^>]*placeholder="Quien más sabe de este punto"/);
  assert.match(pointformSrc, /input\('coordinator', 'Reportado por', \{icon: 'megaphone'/);
  assert.match(pointformSrc, /input\('coordinatorPhone', 'Teléfono de quien reporta'/);
  assert.doesNotMatch(pub, /Personas del grupo/, 'no redundant title');
  assert.match(pub, /<p class="pf-hint"><b>Al menos un contacto con teléfono es obligatorio\.<\/b><\/p><div class="pf-members"><\/div><button type="button" class="pf-addmember">＋ Agregar otro contacto<\/button>/);
  assert.match(readFileSync(new URL('./builder.html', import.meta.url), 'utf8'), /<small class="role">Reportado por<\/small>/);
  const sup = readFileSync(new URL('./supplier.html', import.meta.url), 'utf8');
  assert.doesNotMatch(sup, /Reportado por/, 'Supply cards: the reporter is just the name and phone');
  assert.match(sup, /when=n\.activityAt\?'':n\.revisedAt\|\|n\.createdAt;/, 'Puntos con necesidades cards: no «hace …» in the people row (they have «Actualizada hace …»)');
  assert.match(sup, /🏃 Asignar a runner<\/button>\$\{needButtons\(n\)\}<\/div><\/article>`\}/, 'share button in the same row as ¡Yo abastezco! / Asignar a runner');
  assert.match(sup, /<div class="dispatchrow"><button class="btn dispatch" type="button" data-dispatch="\$\{id\}">🚚 Despachada<\/button>\$\{needButtons\(n\)\}<\/div>/, 'Empacando: share next to Despachada');
  assert.match(sup, /\$\{needDetails\(\{\.\.\.n,packing:null,noWhat:split\},true,false\)\}\$\{addedByLine\(n\)\}/, '«Agregado por» on Empacando too');
  assert.match(sup, /match=n=>teamOk\(n\)&&\(!q\|\|foldNeed\(\[n\.runnerName,n\.runnerPhone,n\.addedByName,/, 'search finds who added it');
  assert.match(sup, /const addedByLine=n=>n\.addedByName\?`<p class="reqby">🧭 Agregado por <b>\$\{clean\(n\.addedByName\)\}<\/b>\$\{verifiedOf\(n\)\}\$\{n\.addedByPhone\?` <button type="button" class="telbtn" data-tel=/, '«Agregado por» with a tappable phone (sent only to team members)');
  assert.match(sup, /href="tel:\$\{clean\(d\)\}">📞 Llamar<\/a><a class="btn small wa"[^>]*>💬 WhatsApp<\/a>/);
  assert.match(readFileSync(new URL('./runner.html', import.meta.url), 'utf8'), /title="Reportado por \$\{n\}: ver teléfono"/, 'Runners: the reporter is the tappable name');
  assert.match(runner, /<small class="role">Pidió<\/small> \$\{clean\(n\.requesterName\)\}/, 'whoever asked for a need');
  assert.doesNotMatch(admin, /coordinator:'Coordinador'/);
  assert.match(admin, /\['Coordinador','coordinator'\]/, 'the CSV column keeps its name, so old exports still import');
});

test('«Nuestro equipo lo abastece»: team members can hold a new point from the public for 3 hours; Comando can publish it now', () => {
  const fn = pointformSrc.match(/function setMe\(root, me\) \{[^]*?\n  \}/)[0];
  assert.match(fn, /if \(me && me\.team\) root\.insertAdjacentHTML\('beforeend', `<label class="pf-hold"><input type="checkbox" name="holdTeam">/);
  assert.match(pointformSrc, /const HOLD_HOURS = 3;/);
  assert.match(pointformSrc, /data\[el\.name\] = el\.type === 'checkbox' \? el\.checked : el\.value\.trim\(\);/, 'sent as true/false');
  const cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  assert.match(cmd, /const heldLine=p=>p\.heldBy&&p\.heldUntil&&Date\.parse\(p\.heldUntil\)>Date\.now\(\)\?`<p class="heldline">🔒 Solo equipo/);
  assert.match(cmd, /send\('\/api\/agent\/camps\/'\+encodeURIComponent\(p\.id\)\+'\/publish','POST',\{\}\)/);
  assert.match(template, /r\.heldUntil&&Date\.parse\(r\.heldUntil\)>Date\.now\(\)\?`<span class="heldtag"/, 'the team sees it marked in the list');
});

test('admin Usuarios: tabs Usuarios / Equipos, live refresh every 10 s, and «🆕 Recién registrados» to assign team and roles in one step', () => {
  assert.match(admin, /<nav class="utabs" id="userTabs" role="tablist"[^>]*><button type="button" role="tab" data-utab="users" aria-selected="true"><span aria-hidden="true">👤<\/span> Usuarios<\/button><button type="button" role="tab" data-utab="teams" aria-selected="false"><span aria-hidden="true">👥<\/span> Equipos/);
  assert.match(admin, /<span class="livetag" id="liveTag"[^>]*>● En vivo/);
  assert.match(admin, /\},10000\);/, 'refresh every 10 s');
  assert.match(admin, /const newcomers=\(\)=>\{const now=Date\.now\(\);return users\.filter\(u=>!u\.disabled&&now-Date\.parse\(u\.createdAt\|\|0\)<NEW_WINDOW\)/, 'everyone from the last 3 days, assigned or not');
  assert.match(admin, /const failed=await patchUsers\(\[u\.id\],\{teamId:team,roles\}\);/, 'team and roles in one save');
  assert.match(admin, /localStorage\.setItem\('adminLastTeam',team\)/, 'remembers the last team');
  assert.match(admin, /<button type="button" class="nc-item\$\{freshIds\.has\(u\.id\)\?' fresh':''\}\$\{u\.teamId\?'':' noteam'\}" data-nc-pick=/, 'a plain list: tap someone to open the assign window');
  assert.match(admin, /<dialog id="adDlg" class="addlg"[^>]*><div class="adbox"><h2 id="adTitle">Asignar a un equipo<\/h2>/);
  assert.match(admin, /const codeTag=u=>u\.hasAccessCode\?`<span class="ucode" data-code-for="\$\{clean\(u\.id\)\}">🔑 <code>/);
  assert.match(admin, /adminRequest\('\/api\/admin\/users\/'\+encodeURIComponent\(id\)\+'\/access-code'\);codeCache\.set\(id,/, 'loaded one by one like 🔑, cached');
  assert.match(admin, /<div class="adstep"><span class="stepn">1<\/span><span class="stepl">Equipo<\/span><div id="adTeamWrap"><\/div><\/div><div class="adstep"><span class="stepn">2<\/span><span class="stepl">Roles<\/span>[^]*?<span class="stepn">3<\/span><span class="stepl">Guardar<\/span>/, 'steps 1, 2, 3 in the window');
  assert.match(admin, /busy=userFormOpen\(\)\|\|renamingTeam\|\|\$\('#adDlg'\)\.open\|\|/, 'no refresh while the window is open');
  assert.match(admin, /if\(seenUsers\)for\(const u of users\)if\(!seenUsers\.has\(u\.id\)\)freshIds\.set\(u\.id,now\)/, 'newcomers since the page opened are highlighted');
  assert.match(admin, /const roleNames=\{runner:'Runner',builder:'Constructor',agent:'Agente',donations:'Donaciones',supplier:'Supply',teammanager:'Team manager',cook:'Cocina',admin:'Admin'\};/);
});

test('admin Equipos tab: a card per team with its Team managers, members (☆ Hacer team manager / Quitar) and an add-member search', () => {
  assert.match(admin, /<p class="tc-mgr\$\{mgrs\.length\?'':' none'\}"><span class="tc-lbl">★ Team manager<\/span>/);
  assert.match(admin, /data-tm-toggle="\$\{clean\(u\.id\)\}" aria-pressed="\$\{mgr\}">\$\{mgr\?'★ Team manager':'☆ Hacer team manager'\}<\/button>/);
  assert.match(admin, /const failed=await patchUsers\(\[u\.id\],\{teamId:'',roles:u\.roles\.filter\(r=>r!=='teammanager'\)\}\);/, 'removing from the team also removes Team manager');
  assert.match(admin, /placeholder="＋ Agregar miembro: nombre, teléfono o código…"/);
  assert.match(admin, /userTab\(\{equipos:'teams',registros:'live',invitaciones:'invites'\}\[q\.get\('tab'\)\]\|\|'users',false\)/, 'the tab is in the link');
});

test('admin Equipos: quick search by team name or by person (name, phone, ID) filters the cards', () => {
  assert.match(admin, /<div class="team-find"><input id="teamFind" type="search" placeholder="🔍 Buscar equipo o persona \(nombre, teléfono o código\)…"/);
  const fold = s => String(s).normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase();
  const teams = [{id: 't1', name: 'Oro'}, {id: 't2', name: 'Plata'}];
  const users = [{id: 'aa11', name: 'Lucía', teamId: 't1', telephone: '6611-2233'}, {id: 'bb22', name: 'Pablo', teamId: 't2'}, {id: 'cc33', name: 'Ana', teamId: 't2'}];
  const find = q => new Function('$', 'teams', 'users', 'fold', 'codeCache', `${admin.match(/function teamMatches\(\)\{[^\n]*/)[0]};return teamMatches()`)(s => s === '#teamFind' ? {value: q} : null, teams, users, fold, new Map([['bb22', 'pablo77']]));
  assert.equal(find('').length, 2);
  assert.deepEqual(find('pla').map(([t, only]) => [t.id, only]), [['t2', null]], 'by team name: the whole card');
  assert.deepEqual(find('lucia').map(([t, only]) => [t.id, [...only]]), [['t1', ['aa11']]], 'by person (accents ignored): only them');
  assert.deepEqual(find('2233').map(([t]) => t.id), ['t1'], 'by phone');
  assert.deepEqual(find('cc3').map(([t, only]) => [t.id, [...only]]), [['t2', ['cc33']]], 'by ID');
  assert.deepEqual(find('pablo77').map(([t, only]) => [t.id, [...only]]), [['t2', ['bb22']]], 'by access code once loaded');
});

test('admin «📡 Registros en vivo» tab (with its own filter) and «🔗 Invitar a un equipo»: one-use 24 h link only for that user ID', () => {
  assert.match(admin, /<button type="button" role="tab" data-utab="live" aria-selected="false"><span aria-hidden="true">📡<\/span> Registros en vivo <b class="livecount" id="liveTabCount" hidden><\/b><\/button>/);
  assert.match(admin, /#operations:not\(\.tab-live\) #newcomers\{display:none!important\}/, 'the sign-ups list is its own tab');
  assert.match(admin, /<input id="ncFind" type="search" value="\$\{clean\(ncQuery\)\}" placeholder="🔍 Filtrar: nombre, teléfono, código o equipo…"/);
  assert.match(admin, /nq\?all\.filter\(u=>fold\(u\.name\+' '\+\(u\.telephone\|\|''\)\+' '\+u\.id\+' '\+\(codeCache\.get\(u\.id\)\|\|''\)\+' '\+teamName\(u\.teamId\)\)\.includes\(nq\)\):all/);
  assert.match(admin, /<details class="invitebox" id="inviteBox" open><summary>🔗 Nueva invitación a un equipo <small>enlace de un solo uso, 24 h, solo para esa persona<\/small><\/summary>/);
  assert.match(admin, /adminRequest\('\/api\/admin\/teams\/'\+encodeURIComponent\(team\)\+'\/invites','POST',\{userId:u\.id\}\)/);
  assert.match(admin, /const url=location\.origin\+'\/signup\.html\?invite='\+encodeURIComponent\(r\.token\)/);
  assert.match(admin, /data-ad-invite>🔗 Mandarle una invitación en vez de asignar<\/button>/, 'from the assign window, prefilled');
});

test('admin search results (invite person, add member) are readable: dark text on white, not the white admin button text', () => {
  assert.match(admin, /\.inv-hits button\{[^}]*background:#fff;color:#173c46;/);
  assert.match(admin, /\.tc-hits button\{[^}]*background:#fff;color:#173c46;/);
});

test('admin team picker: under 9 teams, one tap button per team with its initials in a colored circle; otherwise a dropdown', () => {
  const src = [admin.match(/const TEAM_COLORS=[^\n]*/)[0], admin.match(/const teamColor=[^\n]*/)[0], admin.match(/const teamInitials=[^\n]*/)[0], admin.match(/function teamPicker[^\n]*/)[0]].join('\n');
  const clean = s => String(s ?? '');
  const mk = teams => new Function('teams', 'clean', `${src};return {teamPicker, teamInitials}`)(teams, clean);
  const {teamPicker, teamInitials} = mk([{id: 'a', name: 'Oro'}, {id: 'b', name: 'Unidos por Azuero'}]);
  assert.equal(teamInitials('Oro'), 'Or');
  assert.equal(teamInitials('Unidos por Azuero'), 'UA');
  assert.equal(teamInitials('Equipo principal'), 'Pr');
  const html = teamPicker('data-nc-team', 'b');
  assert.match(html, /<input type="hidden" data-nc-team value="b">/);
  assert.match(html, /<button type="button" class="tbadge on" data-tpick="b" role="radio" aria-checked="true" title="Unidos por Azuero"><span class="ti" style="background:#19745f">UA<\/span>/);
  const nine = mk(Array.from({length: 9}, (_, i) => ({id: 't' + i, name: 'Equipo ' + i}))).teamPicker('data-nc-team', '');
  assert.match(nine, /^<select data-nc-team/, '9 or more: dropdown');
});

test('admin «🔗 Invitaciones» tab: create invites there and monitor all of them (pending, used, expired, revoked) with «Anular»', () => {
  assert.match(admin, /data-utab="invites" aria-selected="false"><span aria-hidden="true">🔗<\/span> Invitaciones <small id="invTabCount"><\/small><\/button>/);
  assert.match(admin, /#operations:not\(\.tab-invites\) #inviteBox,#operations:not\(\.tab-invites\) #invitesPane\{display:none!important\}/);
  assert.match(admin, /invites=await adminRequest\('\/api\/admin\/invites'\)/);
  assert.match(admin, /const INV_STATES=\[\['pending','⏳ Pendientes'\],\['used','✓ Usadas'\],\['expired','⌛ Vencidas'\],\['revoked','✕ Anuladas'\],\['','Todas'\]\];/);
  assert.match(admin, /adminRequest\('\/api\/admin\/invites\/'\+encodeURIComponent\(i\.id\)\+'\/revoke','POST',\{\}\)/);
  assert.match(admin, /if\(\$\('#operations'\)\.classList\.contains\('tab-invites'\)\)loadInvites\(\)/, 'refreshes with the 10 s live update');
});

test('admin: roles carry their page icon everywhere; the assign window offers «⊘ Sin equipo» to take someone out of a team', () => {
  assert.match(admin, /const ROLE_ICONS=\{runner:'🏃',builder:'🔨',agent:'🧑‍💻',donations:'🎁',supplier:'📦',teammanager:'👥',cook:'🍳',admin:'🔒'\},roleLabel=r=>`\$\{ROLE_ICONS\[r\]\?ROLE_ICONS\[r\]\+' ':''\}\$\{roleNames\[r\]\|\|r\}`;/);
  const roleLabel = new Function(`const roleNames={runner:'Runner'};${admin.match(/const ROLE_ICONS=[^\n]*/)[0]};return roleLabel`)();
  assert.equal(roleLabel('runner'), '🏃 Runner', 'no recursion');
  assert.doesNotMatch(admin, /roleNames\[r\]\|\|r\)\.join/, 'role lists use roleLabel');
  assert.match(admin, /data-tpick="__none__" role="radio"[^>]*title="Sacarlo de su equipo"><span class="ti">⊘<\/span><small>Sin equipo<\/small>/);
  assert.match(admin, /teamPicker\('id="adTeam"',u\.teamId\|\|lastTeam\(\),true\)/);
  assert.match(admin, /const lostTM=noTeam&&roles\.includes\('teammanager'\);if\(lostTM\)roles=roles\.filter\(r=>r!=='teammanager'\);/, 'no team also drops Team manager');
});

test('home: runner actions (conductor, empaque, foto de partida) only for the Runner role; on phones, non-runners get no list and map taps open nothing', () => {
  assert.match(template, /function canRun\(\)\{try\{return\(JSON\.parse\(localStorage\.getItem\('codeUser'\)\|\|'null'\)\?\.roles\|\|\[\]\)\.includes\('runner'\)/);
  assert.match(template, /\$\{canRun\(\)&&n\.kind!=='meal'\?`<div class="teamacts"><button type="button" data-assign-need=/);
  assert.match(template, /n\.runnerName&&canRun\(\)\?` · <button type="button" class="unassign"/);
  assert.match(template, /m\.on\('click',\(\)=>\{if\(!phoneLocked\(\)\)show\(r\)\}\)/);
  assert.match(template, /function mobileView\(view\)\{if\(view==='list'&&phoneLocked\(\)\)view='stats';/);
  assert.match(template, /body\.no-runner \.mobile-nav button\[data-view=list\]\{display:none\}/);
  assert.match(template, /codeMe\(\)&&!phoneLocked\(\)\?'<div id="teamNeeds">/);
  assert.match(template, /addEventListener\('accesschange'/);
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  assert.match(nav, /dispatchEvent\(new Event\('accesschange'\)\)/);
  const go = readFileSync(new URL('./fullneeds.go', import.meta.url), 'utf8');
  for (const route of ['runner', 'media', 'packing']) assert.match(go, new RegExp(`POST /api/full/needs/\\{id\\}/${route}", func\\(w http.ResponseWriter, r \\*http.Request\\) \\{\\n\\t\\t(me|u), ok := a\\.runnerViewer`));
});

test('admin and Command show «Agregada por» on a need added by editing someone else\'s run', () => {
  const cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  const adm = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  assert.match(cmd, /n\.createdByName&&'Agregada por '\+n\.createdByName/);
  assert.match(adm, /n\.createdByName\?'Agregada por '\+n\.createdByName:''/);
});

test('change requests: agents from another team send «Cambios propuestos» to whoever last updated the need (Runners and Command)', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  const cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  const pf = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
  assert.match(pf, /function changeRequests\(box, \{load, decide, after, showSent\} = \{\}\)/);
  assert.match(pf, /📝 Cambios propuestos/);
  assert.match(pf, /✓ Aceptar<\/button><button type="button" class="crq-no" data-crq-no>✗ Rechazar/);
  assert.match(pf, /📤 Cambios que pediste/);
  assert.match(runner, /<div class="crqbox" id="changeReqs" hidden><\/div>/);
  assert.match(runner, /fetch\('\/api\/runner\/change-requests'/);
  assert.match(cmd, /<div class="crqbox" id="changeReqs" hidden><\/div>/);
  assert.match(cmd, /showSent:true,load:async\(\)=>\{const r=await fetch\('\/api\/agent\/change-requests'/);
  assert.match(cmd, /if\(r\.status===202&&body\.request\)askedOwner=body\.request/);
  assert.match(cmd, /📝 Cambio enviado a \$\{askedOwner\.toName\}/);
  assert.match(cmd, /askedOwner\.kind==='punto'\?'este punto':'esta necesidad'/);
});

test('home: signed-out visitors see «🙌 ¿Quieres ayudar?» above the Resumen; «¡Empieza!» opens Login', () => {
  assert.match(template, /box\.innerHTML=joinCard\(\)\+`<p class="sumnote" role="note">/, 'no «Resumen» title or timestamp above the summary');
  assert.doesNotMatch(template, /<h2>Resumen<\/h2>/);
  assert.match(template, /hide=!!localStorage\.getItem\('codeUser'\)\}catch\{\}return hide/);
  assert.doesNotMatch(template, /joinHidden|data-join="hide"/);
  assert.match(template, /<b>🙌 ¿Quieres ayudar\?<\/b>/);
  assert.match(template, /<div class="joinacts"><button type="button" class="add" data-join="start">¡Empieza!<\/button><\/div>/);
  assert.match(template, /identifica y reporta necesidades, entrega donaciones o ayuda a actualizar datos por teléfono/);
  assert.doesNotMatch(template, /Crear mi cuenta<\/button>|data-join="code"/);
});

test('«🕓 Historial de cambios» on a point in admin and Command; builders answer change requests too', () => {
  const cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  const adm = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  const bld = readFileSync(new URL('./builder.html', import.meta.url), 'utf8');
  const pf = readFileSync(new URL('./pointform.js', import.meta.url), 'utf8');
  assert.match(pf, /function trailHtml\(list\)/);
  assert.match(pf, /function watchTrails\(urlFor\)/);
  assert.match(pf, /details\[data-trail\]/);
  assert.match(cmd, /<details class="dlog" data-trail="\$\{clean\(p\.id\)\}"><summary>🕓 Historial de cambios<\/summary><div class="trail"><\/div><\/details>/);
  assert.match(cmd, /PointForm\.watchTrails\(id=>'\/api\/agent\/camps\/'\+encodeURIComponent\(id\)\+'\/trail'\)/);
  assert.match(adm, /<details id="campTrail" data-trail=""><summary>🕓 Historial de cambios<\/summary><div class="trail"><\/div><\/details>/);
  assert.match(adm, /\$\('#campTrail'\)\.dataset\.trail=p\.id;/);
  assert.match(adm, /PointForm\.watchTrails\(id=>'\/api\/admin\/camps\/'\+encodeURIComponent\(id\)\+'\/trail'\)/);
  assert.match(bld, /<div class="crqbox" id="changeReqs" hidden><\/div>/);
  assert.match(bld, /fetch\('\/api\/builder\/change-requests'/);
});

test('camp page «🕓 Actividad del punto» (public log, teammates by name) and the «🧭 Scout» page (my and my team\'s points, list and map)', () => {
  const camp = readFileSync(new URL('./camp.html', import.meta.url), 'utf8');
  const scout = readFileSync(new URL('./scout.html', import.meta.url), 'utf8');
  const cmdPage = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  const go = readFileSync(new URL('./main.go', import.meta.url), 'utf8');
  const docker = readFileSync(new URL('./Dockerfile', import.meta.url), 'utf8');
  assert.match(camp, /<section id="campLog"><h2>🕓 Actividad del punto<\/h2>/);
  assert.match(camp, /get\('\/api\/full\/camps\/'\+encodeURIComponent\(id\)\+'\/log'\)/);
  assert.match(camp, /if\(!r\|\|!r\.ok\)r=await get\('\/api\/camps\/'\+encodeURIComponent\(id\)\+'\/log'\)/);
  assert.doesNotMatch(camp, /Por privacidad no se muestran/);
  assert.match(scout, /fetch\('\/api\/full\/scout'/);
  assert.match(scout, /<nav class="appnav" id="appNav" hidden><\/nav><input class="topsearch" id="find" type="search" placeholder="Buscar…"/, 'search in the header after the nav, like Runners');
  assert.match(scout, /<main class="app" id="app" hidden><section class="mapwrap"><div id="map"[^>]*><\/div>[^]*?<section class="side">/, 'map left, list right, like Supply');
  assert.match(scout, /<button type="button" data-view="list" class="active">[^]*?<span>Lista<\/span><\/button><button type="button" data-view="map">/);
  assert.doesNotMatch(scout.slice(scout.indexOf('<main class="app"')), /<h1>/, 'no page title (only on the sign-in card)');
  assert.match(scout, /<select class="whosel" id="who" aria-label="Qué puntos ver"><option value="all">Todos los puntos<\/option><option value="mine">Creados por mí<\/option><option value="team">Creados por mi equipo<\/option><\/select>/, 'filter is a dropdown');
  assert.match(scout, /data-update="\$\{clean\(p\.id\)\}">📝 Actualizar necesidades<\/button>/, 'the card action is «Actualizar necesidades»');
  assert.match(scout, /\$\{myTeam\?`<button type="button" class="btn small" data-update=/, 'only people on a team update needs');
  assert.match(scout, /Va en camino: se puede cambiar, no quitar\./);
  assert.match(scout, /data-check="\$\{clean\(p\.id\)\}">🔍 Corroborar<\/button>/, 'points «Por revisar» get «Corroborar»');
  assert.match(scout, /function renderView\(id\)[^]*?🔍 Corroborar punto[^]*?Necesidades reportadas/, '«Corroborar» opens the point in view mode');
  assert.doesNotMatch(scout, /id="checkDialog"/, 'the point view is a screen, not a modal');
  assert.match(scout, /<section class="pview" id="pview" hidden><\/section>/);
  assert.match(scout, /history\.pushState\(\{pv:1\},'','\?punto='\+encodeURIComponent\(id\)\)/, 'the view has its own address');
  assert.match(scout, /function leaveView\(\)\{if\(history\.state\?\.pv\)history\.back\(\)/, '«‹ Volver» goes back in history');
  assert.match(scout, /addEventListener\('popstate',route\)/);
  assert.match(scout, /\$\{peopleLine\(p\)\}\$\{contactsHtml\(p\)\}/, 'cards list contacts (sent only to team members)');
  assert.match(scout, /<a class="btn small" href="tel:\$\{clean\(d\)\}">📞 Llamar<\/a><a class="btn small wa"[^>]*href="https:\/\/wa\.me\/\$\{clean\(wa\)\}">💬 WhatsApp<\/a>/, 'tapping a number offers Llamar / WhatsApp');
  assert.match(cmdPage, /function backToScout\(\)\{if\(!fromScout\|\|history\.state\?\.edit\|\|history\.state\?\.pt/, 'Command opened from Scout goes back to Scout');
  assert.match(cmdPage, /function closeEdit\(\)\{mode='';if\(backToScout\(\)\)return;/);
  assert.match(cmdPage, /sessionStorage\.setItem\('scoutNotice'/);
  assert.match(scout, /Búsqueda aproximada: no hay GPS registrado\./, 'a place-name search with no GPS is labeled approximate');
  assert.match(scout, /<button type="button" class="btn small ghost" data-check="\$\{clean\(p\.id\)\}">Ver más<\/button>/, 'published points get «Ver más» (view mode)');
  assert.match(scout, /\$\{pending\?'🔍 Corroborar punto':'📋 Ver punto'\}/);
  assert.match(camp, /<main><a class="backlink" href="\/" data-back>← Inicio<\/a>/, 'camp page back links know where you came from');
  assert.match(camp, /if\(a&&cameFrom&&history\.length>1[^]*?history\.back\(\)/, 'back goes through history, keeping the list');
  assert.match(camp, /'\/scout\.html':'Scout'/);
  assert.match(scout, /\$\{canEdit\?`<a class="btn small" href="\/command\.html\?punto=\$\{encodeURIComponent\(p\.id\)\}&editar=1">✏️ Editar<\/a>`:''\}/, '«Editar» opens the point in edit mode in Command, agents only');
  assert.match(scout, /<ul class="nl">\$\{needs\.map\(needRow\)\.join\(''\)\}<\/ul>/, 'each card lists its active needs like Runners');
  assert.match(scout, /send\('\/api\/full\/scout\/needs\/'\+encodeURIComponent\(rm\.dataset\.remove\),'PATCH',\{status:'DELETED'\}\)/);
  assert.match(scout, /send\('\/api\/full\/scout\/camps\/'\+encodeURIComponent\(editing\)\+'\/needs','POST',\{need,urgency\}\)/);
  assert.match(scout, /<a class="add" href="\/\?agregar=1">\+ Agregar punto<\/a>/);
  assert.match(template, /get\('agregar'\)\)\{history\.replaceState\(null,'',location\.pathname\);setTimeout\(\(\)=>\$\('#addBtn'\)\.click\(\),300\)\}/);
  assert.match(nav, /\['\/scout\.html', '🧭', \(\) => 'Scout', u => !!u\]/);
  assert.match(go, /"\/camp\.html", "\/scout\.html",/);
  assert.match(docker, /camp\.html scout\.html /);
});

test('tabs look the same everywhere: view switchers as Runners/Supply pills, section tabs as Supply (icon above the name on phones)', () => {
  const pill = /background:#e9f0f0;[^}]*\}[^]*?\.active\{background:#173c46;border-color:#173c46;color:#fff;box-shadow:0 2px 8px #17354240\}/;
  for (const f of ['index.html', 'builder.html', 'runner.html', 'supplier.html']) assert.match(readFileSync(new URL('./' + f, import.meta.url), 'utf8'), pill, f);
  assert.match(readFileSync(new URL('./command.html', import.meta.url), 'utf8'), /\.cmdtabs button\[aria-pressed=true\]\{background:#173c46;border-color:#173c46;color:#fff;box-shadow:0 2px 8px #17354240\}/);
  assert.match(readFileSync(new URL('./scout.html', import.meta.url), 'utf8'), /\.mobile-nav button\.active\{background:#173c46;border-color:#173c46;color:#fff;box-shadow:0 2px 8px #17354240\}/);
  for (const f of ['runner.html', 'builder.html', 'supplier.html']) assert.match(readFileSync(new URL('./' + f, import.meta.url), 'utf8'), /@media\(max-width:760px\)\{#mainTabs button:not\(\.histtab\)\{display:flex;flex-direction:column;align-items:center;gap:2px;line-height:1\.15\}/, f);
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  assert.match(admin, /\.utabs \[data-utab\]\[aria-selected="true"\] small:not\(:empty\)\{background:#ed7654;color:#fff\}/);
  assert.match(admin, /@media\(max-width:760px\)\{\.utabs \[data-utab\]\{flex:1 1 0;display:flex;flex-direction:column/);
});

test('nav: «🔒 Administración» only when this browser has an admin session; last, small, dashed, and asks before opening', () => {
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  assert.match(nav, /fetch\('\/api\/admin\/session', \{credentials: 'same-origin', cache: 'no-store'\}\)\.then\(r => \{ if \(r\.ok\) \{ adminOK = true; draw\(\); drawTitle\(\); \} \}\)/);
  assert.match(nav, /\$\{admin\}<\/div>`;/, 'desktop: after the pages');
  assert.match(nav, /class="madmin" data-admin-link><span aria-hidden="true">🔒<\/span> Administración<\/a>/, 'phone menu: at the very bottom');
  assert.match(nav, /confirm\('¿Abrir Administración\?', \{yes: 'Abrir', no: 'Cancelar'/);
  assert.match(nav, /\.appnav-list \.appnav-admin\{[^}]*border:1px dashed/);
});

test('role «🔒 Admin»: only the admin panel offers it (user form and assign window); team page and nav show it', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  const go = readFileSync(new URL('./users.go', import.meta.url), 'utf8');
  assert.match(admin, /<input type="checkbox" name="roles" value="admin"> 🔒 Admin <small>\(solo lo da administración\)<\/small>/);
  assert.match(admin, /\['admin','🔒 Admin'\]\];/);
  assert.match(go, /roleAdmin = "admin"/);
  assert.match(go, /roleTeamManager, roleAdmin\}/);
  assert.match(nav, /const admin = \(adminOK \|\| \(u\?\.roles \|\| \[\]\)\.includes\('admin'\)\) && here !== '\/admin\.html'/);
});

test('admin panel header: shared nav.js page switcher with only «🏠 Inicio» and «🔒 Administración», which opens the admin tools menu (no separate ☰)', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  assert.match(admin, /<header><span class="adminlogo"[^>]*>✚<\/span><nav class="appnav" id="appNav" hidden><\/nav>/);
  assert.match(admin, /<script src="\.\/nav\.js"><\/script>/);
  assert.doesNotMatch(admin, /function devMark/, 'the DEV stripe now comes from nav.js');
  assert.match(admin, /\}header\{position:relative;z-index:2000;/, 'the dropdown opens under the header, above the map');
  assert.match(nav, /if \(here === '\/admin\.html'\) list = \[pages\[0\], \['\/admin\.html', '🔒', \(\) => 'Administración', \(\) => true\]\];/);
  assert.match(admin, /function setMenu\(open\)\{document\.querySelector\('#appNav \[data-admin-menu\]'\)\?\.setAttribute\('aria-expanded',String\(!!open\)\);if\(open\)\{document\.getElementById\('appNav'\)\?\.classList\.remove\('open'\)/, 'only one menu open at a time');
});

test('admin tools menu: opens from «🔒 Administración ▾», every item has an icon, groups split by lines', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  assert.match(nav, /const adminMenu = here === '\/admin\.html' \? ' data-admin-menu aria-haspopup="menu" aria-controls="adminMenu"' : '';/);
  assert.match(admin, /const t=e\.target\.closest\('\[data-admin-menu\]'\);if\(t\)\{e\.preventDefault\(\);/);
  const menu = admin.match(/<div class="adminmenu" id="adminMenu"[^]*?<input type="file"/)[0];
  assert.equal((menu.match(/class="mi"/g) || []).length, 9, 'an icon on each of the 9 items');
  assert.equal((menu.match(/<hr class="msep" role="separator">/g) || []).length, 4, 'people · import · boletín · settings · backup');
  assert.match(admin, /reviewToggle'\)\.innerHTML=`<span class="mi" aria-hidden="true">✅<\/span>Requerir revisión/, 'the icon survives the Sí/No update');
});

test('admin «Importar reportes»: drop the .csv/.json onto the box (or tap to choose); same reader for both', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  assert.match(admin, /<label class="ipdrop" id="importPendingDrop"><span class="ipdropico" aria-hidden="true">📄<\/span><b>Suelta aquí el archivo<\/b>/);
  assert.match(admin, /\$\('#importPendingFile'\)\.onchange=e=>readImportFile\(e\.target\.files\[0\]\);/);
  assert.match(admin, /\$\('#importPending'\)\.addEventListener\('drop',e=>\{e\.preventDefault\(\);[^\n]*if\(!\/\\\.\(csv\|json\|gz\)\$\/i\.test\(f\.name\)\)\{\$\('#importPendingMsg'\)\.textContent='Solo archivos \.csv, \.json o \.json\.gz\.';return\}readImportFile\(f\)\}\);/);
  assert.match(admin, /async function readImportFile\(f\)\{/);
  assert.match(admin, /async function readBundle\(f\)\{const text=await new Response\(f\.stream\(\)\.pipeThrough\(new DecompressionStream\('gzip'\)\)\)\.text\(\)/, 'reads .json.gz import bundles');
  assert.match(admin, /importRef:`\$\{ref\}:\$\{j\+1\}`/, 'same delivery refs as import.py, so re-importing creates nothing twice');
  assert.match(admin, /\$\('#importPendingGo'\)\.onclick=async\(\)=>\{if\(importBundle\)return runBundle\(\);/);
  assert.match(admin, /\.ipdrop:hover,\.ipdrop\.over\{border-color:#19745f;background:#e9f5f1\}/);
});

test('Usuarios «Seleccionados»: open with 1, collapsed with 2 or more unless toggled by hand (remembered until the selection is empty)', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  assert.match(admin, /const userSel=new Set\(\);let selOpen=null;/);
  assert.match(admin, /users-chosen"\$\{\(selOpen\?\?chosen\.length<2\)\?' open':''\}>/);
  assert.match(admin, /if\(!userSel\.size\)selOpen=null;/);
});

test('«Seleccionados» title: checkbox on the left clears the selection even when collapsed; ▾ on the right turns when open', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  assert.match(admin, /\.users-chosen\[open\]>summary \.selsum-chev\{transform:rotate\(180deg\)\}/);
  assert.match(admin, /\$\('#userList'\)\.addEventListener\('click',e=>\{if\(!e\.target\.closest\('\[data-unsel-all\]'\)\)return;e\.preventDefault\(\);userSel\.clear\(\);drawUsers\(\)\}\);/, 'the click on the box does not fold the group');
});

test('Usuarios «📊 En vivo» box (right column on wide screens, top on phones): online, users, teams with members, en ruta, waiting for a driver, GPS', () => {
  const admin = readFileSync(new URL('./admin.html', import.meta.url), 'utf8');
  const users = readFileSync(new URL('./users.go', import.meta.url), 'utf8');
  assert.match(admin, /<section class="operations" id="operations"><aside class="ustats" id="userStats" aria-label="Usuarios en vivo"><\/aside>/);
  for (const label of ["'En línea'", "'Usuarios'", "'Equipos'", "'En ruta'", "'Esperando conductor'", "'GPS activo'", 'Personas por equipo']) assert.ok(admin.includes(label), label);
  assert.match(admin, /waiting=allNeeds\.filter\(n=>n\.status==='OPEN'&&!n\.runnerId&&!n\.kind\)\.length,gps=act\.filter\(u=>u\.sharing\)\.length/);
  assert.match(admin, /#dashboard\.users-mode #operations>\.ustats\{grid-column:2;grid-row:1\/span 60;position:sticky/);
  assert.match(users, /u\.Sharing = sharing && time\.Since\(loc\.At\) < locationTTL/, 'only yes/no, never the position');
});

test('prepared meals: «🍲 Pedir comida preparada» (team only) on the home point; Cocina page (take, platos, foto, conductores); Runners «🍲 Comidas»; services on the point page and in Command', () => {
  const cook = readFileSync(new URL('./cook.html', import.meta.url), 'utf8');
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  const camp = readFileSync(new URL('./camp.html', import.meta.url), 'utf8');
  const cmd = readFileSync(new URL('./command.html', import.meta.url), 'utf8');
  const nav = readFileSync(new URL('./nav.js', import.meta.url), 'utf8');
  const go = readFileSync(new URL('./main.go', import.meta.url), 'utf8');
  const docker = readFileSync(new URL('./Dockerfile', import.meta.url), 'utf8');
  assert.match(template, /\$\{r\.id&&codeMe\(\)\?\.team\?'<button type="button" class="add mealreq" data-meal-req>🍲 Pedir comida preparada<\/button>':''\}/, 'only for people on a team');
  assert.match(template, /<input name="people" type="number" inputmode="numeric" min="1" max="5000" required/);
  assert.match(template, /\[\['desayuno','🍳 Desayuno'\],\['almuerzo','🍛 Almuerzo'\],\['cena','🍲 Cena'\]\]/);
  for (const re of [/\/api\/cook\/tickets'/, /data-take="\$\{clean\(t\.id\)\}">🍳 Tomar este pedido/, /data-ask="\$\{id\}">📣 Pedir conductores/, /data-invite="\$\{id\}">Asignar/, /🍳 La tiene <b>/, /type="time" data-time=/, /PointForm\.openPacking\(\{subtitle:/, /\/media\?stage=pickup'/]) assert.match(cook, re);
  assert.match(runner, /data-main="meals"[^>]*><span aria-hidden="true">🍲<\/span> Comidas <b id="cntMeals"><\/b>/);
  assert.match(runner, /data-meal-accept="\$\{id\}">✓ Acepto llevarla/);
  assert.match(runner, /🍳 La cocina te lo pidió a ti/);
  assert.match(camp, /<h2>🧾 Servicios recibidos<\/h2>/);
  assert.match(cmd, /data-act="service">🧾 Registrar servicio<\/button>/);
  assert.match(nav, /\['\/cook\.html', '🍳', \(\) => 'Cocina', u => has\(u, 'cook'\)\]/);
  assert.match(go, /"\/scout\.html", "\/cook\.html",/);
  assert.match(docker, /scout\.html cook\.html /);
  assert.match(cook, /\['mine','🍳','Preparando'\],\['open','🍲','Pedidos'\]\]/, 'section tabs like Supply');
  assert.match(cook, /data-view="hist" title="Historial"/);
  assert.match(cook, /Historial · comidas que completaste \(7 días\)/);
  assert.doesNotMatch(cook, /Otras cocinas/);
  assert.match(cook, /api\('\/api\/cook\/kitchen','POST',\{kitchen:b\.dataset\.kitchen\}\)/, 'kitchen picker: pick from the shared list or type a new one');
  assert.match(cook, /if\(e\.key!=='Enter'\)return;e\.preventDefault\(\);const b=\$\('#kitchenList \[data-kitchen\]'\);if\(b\)b\.click\(\)/, 'Enter picks or creates');
  assert.match(cook, /\/api\/cook\/needs\/'\+encodeURIComponent\(id\)\+'\/take','POST',people\)/, 'a regular «Comida preparada» need becomes my meal ticket');
  assert.match(cook, /localStorage\.setItem\('cookTeamOnly'/);
});

test('runner cards: no «Reportado por <phone>» line; the reporter name (in the text or «📣 name») opens their phone with «📞 Llamar» / «💬 WhatsApp»; no «sin GPS» text and a plain pin label (the crossed pin is only the top-right button)', () => {
  const runner = readFileSync(new URL('./runner.html', import.meta.url), 'utf8');
  assert.match(runner, /function reporterIn\(html,name,phone\)\{/);
  assert.match(runner, /data-reporter="\$\{clean\(phone\)\}" data-rname="\$\{n\}"/);
  assert.match(runner, /href="tel:\$\{clean\(d\)\}">📞 Llamar<\/a><a class="btn small wa" target="_blank" rel="noopener noreferrer" href="https:\/\/wa\.me\/\$\{clean\(wa\)\}">💬 WhatsApp/);
  assert.match(runner, /<div class="grow">\$\{ico\('pin'\)\}<span class="gtext">\$\{place\|\|'<span class="muted">Sin indicaciones escritas<\/span>'\}<\/span><\/div>/);
  assert.doesNotMatch(runner, /nogpstxt|Sin GPS ni indicaciones/);
  assert.match(runner, /\.reporter\{display:inline-block;max-width:13em;overflow:hidden;text-overflow:ellipsis;white-space:nowrap/, 'long names end in «…»');
  const cook = readFileSync(new URL('./cook.html', import.meta.url), 'utf8');
  assert.match(cook, /if\(await relogin\(\)\)\{showApp\(\);load\(\);return\}showGate\(\)/, 'Cocina signs back in with the remembered code (role added after signing in)');
});
