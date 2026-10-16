# Azuero Apoyo — Manual

How the app works, who uses each page, and how points, needs, runs, supply centers and teams relate.
The UI is in Spanish; on-screen labels are quoted here exactly as they appear (e.g. «¡Yo lo llevo!»).
This manual is updated with every new feature or behavior change (see "Keeping this manual current" at the end).

---

## What's new

**2026-10-16**
- One «Buscar» for every page (per tab), with a floating «🔎 Filtrado por «…» ✕» while it's active; phones list
  «🔒 Administración» in the page menu for admins (§2).
- Supply «¡A empacar!» card: 🙌 Necesita (pink) → 📦 Para cargar (share next to «Despachada») → 📍 Destino (with «🧭
  Agregado por») → 🏃 Runner (§8).
- User notes («Notas»): admin edits them in Usuarios and the team manager in «Mi equipo» (also open to the Admin role,
  read-only); imports link «Agregado por» (with its phone) to existing users or create them (§12, §13).
- Checkup on phones: filters collapse into a «Filtros» sheet from the bottom, «Orden» stays visible as a dropdown;
  cards show the last 3 checkups, tap → map, small «Detalles», «Actualizar» on each need, the checkup answer
  «Nadie contestó, pero no necesita nada», and every answer is logged (§10).
- **The «Command» page is now «📞 Checkup»** (same address `/command.html`, same Agente role; the page menu shows 📞
  Checkup; new change-history entries say «Checkup»). Its point cards: the orange «Checkup» button is the main action
  in the top-right corner, contacts show names only (no phones), the need counters are gone, and «🧭 Creado por *name*»
  carries the «Verificado» medal (§10).
- The «Checkup» call is a phone-friendly screen with the point's contacts («📞 Llamar» / «💬 WhatsApp») and the answers
  at the bottom; «Ya no necesita nada» looks like a button (§10).
- Supply: empty search results say what was searched with «✕ Borrar búsqueda y ver todas (N)»; an active search
  highlights the 🔍 button and the box; «¡A empacar!» crosses out needs already packed and turns green when all are (§2, §8).
- Need chips with many materials or brands (e.g. a home repair listing ten materials) wrap inside the card instead of
  spilling over it (Command, Supply, Runners).
- Packing list window: needs on top that grey out as items cover them, items for what's still needed first, categories to tap into with a «‹ Todos ›» breadcrumb, «Agua» and «Lonas» as starting
  items, «Cancelar» goes back to the items (§6).
- «Verificado por equipo» is now just the gold medal icon everywhere; team members tap it to see which team, who and
  when; the public and people without a team get no details (§6, §15).
- Command: home repairs show the builder's work («🔨 Obra»: progress, what's missing, materials, evaluations, photos)
  and repair wording («🔨 En obra», «Asignar constructor») (§10).
- Supply «Buscar» also finds who added the need, who reported it and the point's contacts (§8).
- Supply «Por surtir»: most recently updated need first, with «🕓 Actualizada hace …» on each card; buttons and share
  on one row; no «Reportado por» label; «🧭 Agregado por» with the medal and call/WhatsApp for team members (§8).
- Scout: «🔍 Corroborar» on points «Por revisar» and «Ver más» on published ones open the point as a full screen
  (`/scout.html?punto=<id>`, «‹ Volver», Back/forward work); agents get «✏️ Editar» (Command edit mode, which returns
  to Scout); team members see «Contactos» with «📞 Llamar / 💬 WhatsApp» (§14).
- Back links: the public point page's «←» says «← Volver a *page*» and goes back there (§2).
- Phone header the same everywhere: page links always folded into «current page ▾», also on `signup.html` and
  `camp.html`; «Buscar» has a ✕ and keeps what you typed when you come back to the page; «!código» in the home
  search no longer signs in (§2, §14).
- Signing out is global (ends every session in the browser, admin included) and goes to the home page; Scout, Cocina, Mi equipo and the point page get «Salir» in ☰ (§4).
- Manual audited against the code and corrected throughout (sign-in, roles, home page, points, privacy, Runners,
  Supply, Constructores, Cocina, Command, Donations, teams, admin, technical summary).
- New role and page **🍳 Cocina** for prepared meals: team members ask «🍲 Pedir comida preparada» (people + desayuno /
  almuerzo / cena); one cook takes it, lists dishes, adds a pickup photo and asks runners (open or assigned, who
  accept) to pick up at a time; each runner has their own run; the last delivery completes it (§9b).
- Cocina laid out like Supply: map left, list right; «🍳 Preparando / 🍲 Pedidos» plus 🕓 history; a kitchen picker
  («Elige tu cocina», shared list or a new name) and «Solo mi equipo»; a regular need asking for «Comida preparada» can
  be taken from «Pedidos» (§9b).
- **Scout now lists every published point** (plus your and your team's points still «Por revisar») in cards like
  Runners, each with its active needs; people **on a team** get **«📝 Actualizar necesidades»** for any point (report
  a new need, change one, remove an open one), applied directly and logged. A dropdown picks «Todos los puntos / Creados por mí / Creados por mi equipo». Laid out like
  Supply: no title, map left, list right («Lista / Mapa» on phones), **Buscar** in the header (§14).
- Runners cards: the «Reportado por *phone*» line is gone; the reporter's name opens «📞 Llamar / 💬 WhatsApp»; no
  «sin GPS» text in the location row (the crossed pin at the top right stays) (§8).
- Services log «🧾 Servicios recibidos» on the point page (meals logged automatically; Command «🧾 Registrar servicio»
  for fumigation, doctor visits…; optional «válido por N días» or next date) (§6, §10, §14).
- Runners «Runs» are now newest first (§7).

**2026-10-15**
- Admin Usuarios: «📊 En vivo» box with online, users, teams and their members, en ruta, waiting for a driver and GPS active (§13).
- Admin Usuarios: «Asignar en lote» removed from Equipos; «todos» is a header checkbox; «Seleccionados» starts collapsed with 2 or more (§13).
- Admin «Importar reportes»: drop the .csv/.json file onto the box, or tap to choose it (§5).
- New role **Admin** (only the admin panel can give it; team managers can't, and it stays when they change other roles); it shows «🔒 Administración» in the nav. Admin panel header uses the shared page switcher with only Inicio and Administración; «Administración ▾» opens the admin tools (with icons, grouped by lines) instead of the far-right ☰ (§2, §3).
- Nav: «🔒 Administración» link for browsers with the admin session open, set apart and asking before opening (§2).
- Home Resumen: no «Resumen» title or «Al …» timestamp above the figures (§14).
- Home: «¡Ayuda!» is now «Reportar punto»; the header sign-in button reads «Inicia sesión». Tabs on
  every page share one design: Runners/Supply pills for views, Supply underlined tabs with icons above names on
  phones for sections (§2, §14).
- Manual checked against the change history, team-data guards, activity log and Scout: pages table, top bar, roles
  (no-role Scout, global agents), teams (trusted teammates) (§2, §3, §12).
- Point page «🕓 Actividad del punto»: public activity log since the point was reported (who as a role only; full names
  for your teammates). New «🧭 Scout» page for anyone signed in: your and your team's points, list and map (§2, §14).
- Change history: every change to points, needs, runs, packing, photos and supply is recorded with who, team, page and
  when; «🕓 Historial de cambios» on each point in admin and Command. Global agents (no team) and other teams can't
  overwrite a team's data: point edits are change requests too, and they can't delete/move others' photos, merge into
  or «Publicar ahora» another team's point (§6, §10).
- Home: signed-out visitors see «🙌 ¿Quieres ayudar?» above the Resumen (it can't be closed), with one «¡Empieza!» button that opens the sign-in window (§14).
- Change requests: an agent outside the team of whoever last updated a need can't overwrite it; the change goes to
  that person as «📝 Cambios propuestos» (Runners, Command) to accept or reject. Applies to every Command need change,
  including «Checkup» and «Ya no necesita nada» (§3, §6, §7, §10).
- Manual checked against how runs work now: Command «¿quién la suplió?» and the need editor, admin, Runners «No
  puedo continuar», privacy and technical notes (§6, §7, §10, §13, §15, §16).
- Several runners on one need never overwrite each other: «🚚 Agregar otro conductor» (home) and choosing another
  runner in admin/Command give the new runner their own linked run; edits by others to a runner's run become a new
  open need (§6).
- Home: runner actions on a point (assign, add another or remove «conductor», «📦 Agregar a lista de empaque», «📷 Foto de
  partida») are only for the Runner role. On phones, without that role there is no «Lista» and tapping a point on the
  map opens nothing (§14).
- «Comando» is now **«Command»** with a 🧑‍💻 icon (the HQ where data is checked and managed) (§2, §10).
- Admin Registros en vivo: a simple list; tapping a person opens «Asignar a un equipo» (1 Equipo, 2 Roles, 3 Guardar).
- Admin Usuarios: access codes shown instead of IDs, team buttons with initials, numbered steps 1-2-3, and a
  «🔗 Invitaciones» tab to create invites and monitor them (pending, used, expired, revoked) with «Anular» (§13).
- Admin: «📡 Registros en vivo» is its own tab with a filter; Equipos has a search; «🔗 Invitar a un equipo» makes a
  one-use, 24 h link only that user ID can accept (§13).
- Admin Usuarios: tabs «👤 Usuarios» / «👥 Equipos», live refresh every 10 s, «🆕 Recién registrados sin equipo» to
  assign team and roles in one step, team cards with Team manager and add-member search (§13).
- «Coordinador» is now **«Reportado por»** (megaphone) on every screen; the data field is unchanged. Need requesters on
  cards show as «Pidió» (§5).
- Add point: no «Personas del grupo» title, shorter required message, **«＋ Agregar otro contacto»**, and for teams
  **«Nuestro equipo lo abastece»**: hidden from the public and other teams for 3 hours, then published (§5, §10, §15).
- Sign-up accepts a phone already used (same person who lost their code); teams confirm by user ID (§4).
- Supply: on phones the 🕓 history shows only the history, without the other tabs (§8).
- Supply: «🙋 ¡Yo abastezco!» replaces «Pedir runner»: several centers can offer the same need and the runner picks
  where to pick up in «Pedidos»; «Asignar a runner»; «Despachadas» moved into 🕓 Historial; rename/archive centers;
  team switch on the right; cards like Runners (map top right, people breakdown, contacts collapse after 2) (§7, §8).
- Command: «+ Agregar punto» opens the form inside Command; «Creado por» is a full-name badge; person fields in the
  point form have icons (§5, §10).
- Runners cards: map/Maps buttons top right, people breakdown before «Necesita», no tap-to-select highlight (§7).
- Add point: team members are the coordinator by default (card with «Cambiar»); new «Familias» count (§5).
- Command: «Agregado por» filter (Todos / Míos / Equipo), «Filtros ▾» on phones, clearer rows with icons (§10).
- **Live location is strictly team-only** on every page: runners without a code (no team) and other teams no longer
  appear on anyone's map, and Supply sees a runner's position only if they are on its team (§7, §8, §15).
- Supply: «📍 Centro» is a button with a typeahead list (no always-open text box), «Solo mi equipo» switch on the same
  line, Buscar in the header, and the map shows your team's live runners (§8).
- Supply: tabs renamed and reordered (¡A empacar!, Despachadas, ¡Abastéceme!, Por surtir last), icons on phones, and a
  🕓 history of delivered runs next to «Lista / Mapa» (§8).
- Phone header: the menu button shows your initials when signed in; page links that don't fit collapse into a
  «current page ▾» dropdown; the header search box now also shows on desktop (§2).
- Command: Buscar in the header, «+ Agregar punto» always visible, totals with icons in the «Puntos» tab (§10).
- Runners: header search (🔍 on phones) filters the open tab and the map (§7).
- Needs: options as tap buttons with icons for Agua, Alimentos, generator fuel, prepaid data card and cooking gas,
  with «Otra» to type another; the separate gasolina/diésel boxes became one fuel need; «Saneamiento / baño» removed
  (§6).
- House repair is in every need form with optional materials; it always goes to Builders as a private repair need,
  and builders see the requested materials (§6, §9).
- The same need checklist and need chips everywhere, including the home «Reportar necesidad» window and the public
  camp page (§6).
- «Verificado por equipo» gold badge on points and needs created or updated by someone in a team (§6, §15).
- Command: «Ya no necesita nada» asks who fulfilled it and can complete the need for that runner (§10).
- Mi equipo: role icons and a quick member search (§12).
- Home on phones opens on «📊 Resumen» (§14).
- Development environments (localhost, local network) show a «DEV» marker on every page (§16).

---

## 1. What it is

Azuero Apoyo coordinates relief after a disaster in Azuero (Panama):

- **Points** (*puntos*): places with people who need help (shelters, communities, houses). Each has a location,
  contacts, head counts and what they need.
- **Needs and runs**: what has to be taken to a point. When someone takes a need to deliver it, it becomes a **run**
  handled by a **Runner**.
- **Supply** (*acopios*): collection centers that pack the load, document it and dispatch it with a Runner.
- **Builders** (*constructores*): repair damaged houses.
- **Cocina**: kitchens that cook prepared meals for points; runners carry them.
- **Teams**: groups of volunteers who trust each other's data; people outside a team can't overwrite it.
- **Donations**: donation tickets from offer to delivery.
- **Admin and agents**: review reports, maintain the data, grant access.

Everything the community reports is *reported* data, not verified, and the app presents it that way.

---

## 2. Pages

| Page | Who | What it does |
| --- | --- | --- |
| `/` (home) | Public and everyone | Map plus «📊 Resumen» figures (the side panel on a computer, the first view on phones; the point list is only on phones, for the Runner role), «Reportar punto» (signed in: «+ Agregar punto») to report a point, «Reportar necesidad» on a selected point; signed out, **«Inicia sesión»** in the header (and the «🙌 ¿Quieres ayudar?» card's «¡Empieza!») opens the sign-in window. |
| `/signup.html` | Anyone (link to share) | Sign in with a code or create an account. When signed in, shows which pages the person can open. Also accepts team invites (`?invite=…`). |
| `/camp.html?id=…` | Public | Public page of an approved point (no private data; header «← Inicio» only. When opened from another page of the site, its «←» links at the top and bottom say «← Volver a *page*» (Scout, Runners, Checkup, Supply…) and go back there through the browser history; opened directly, they go to the home page): need chips with their options, «Verificado por un equipo» badge, «Necesidades de servicio» (state, chips, requester), link to Google/Apple Maps (without GPS, an approximate place-name search), **«🧾 Servicios recibidos»** and **«🕓 Actividad del punto»** (§14). |
| `/scout.html` («🧭 Scout») | Anyone signed in with a code (no role needed) | Every published point with its active needs, in cards and on a map; update a point's needs (§14). |
| `/runner.html` («Runners») | Runner role (and agents, for packing lists only; agents without Runner see it as «Empaque») | Take runs, route, packing list, photos, complete; answer «📝 Cambios propuestos». |
| `/builder.html` («Constructores») | Builder role | Repair jobs: assess, take, progress, materials, photos; answer «📝 Cambios propuestos». |
| `/supplier.html` («Supply») | Supply role | Collection center: request runners, assign in person, answer runner requests, pack, dispatch. |
| `/cook.html` («🍳 Cocina») | Cocina role | Prepared meals: take a request, list the dishes, pickup photo, ask runners to pick up at a time (§9b). |
| `/command.html` («Checkup») | Agent role | Call center: points, needs, checkups, report needs; change requests to other teams' data and «🕓 Historial de cambios» per point. |
| `/donations.html` («Donaciones») | Donations role and admin | Donation tickets: offered → confirmed → at center → on the way → delivered. |
| `/team.html` («Mi equipo») | Team manager role | Invite people to the team, see members, change their roles, remove them. Lives in the ☰ menu. |
| `/admin.html` | Admin (master password) | Everything: review reports, points, needs, users, teams, backups, newsletter; «🕓 Historial de cambios» per point. Its header has the same page switcher as the other pages but only with «🏠 Inicio» and «🔒 Administración ▾»; **«Administración» opens the admin tools menu** (there is no separate ☰): 👥 Usuarios, 🎁 Donaciones │ 📥 Importar reportes (por revisar)… │ 📰 Generar boletín, 🔗 Ver último boletín ↗ │ ✅ Requerir revisión de puntos, 🔑 Cambiar contraseña maestra │ 💾 Descargar respaldo, 📂 Importar respaldo… (lines between the groups). |

The top bar links to the pages the person can use when signed in: «Inicio», «🧭 Scout», then one per role (signed out there is no page switcher). **«🔒 Administración»** appears for users with the **Admin** role, or in a browser that already has the admin session open (master password): last in the list on a computer, and on phones both last in the «página ▾» list and at the bottom of the ☰ menu, small, dashed and faded, and it asks «¿Abrir Administración?» before going there, so it isn't opened by mistake. On phones rarely used ones (Mi equipo) are in the menu. **On phones the
links are always collapsed into one «current page ▾» button** that opens the list as a dropdown, on every page alike,
even when there are only two (with a single page it shows just the page name). Every page uses this same header,
including `signup.html` and the public point page `camp.html` (whose «← Volver…» link is now just under the header),
so the page action (＋ «Agregar punto»), 🔍 and the menu button always stay visible on the right.
**«Buscar…»** in the header (Inicio, Scout, Runners, Supply, Checkup, admin) has a **✕** inside once something is typed,
which clears it. There is **one search for every page** (per browser tab): type «erasmo» in Supply, open Checkup or Scout, and the
same search is there and applied. While a search is active a **floating pill** at the bottom says **«🔎 Filtrado por
«erasmo» ✕»**, so it's clear why a list is filtered even if the search was typed on another page or the search bar is
hidden: tapping the text opens the search to change it, **✕** clears it on every page. It's forgotten when the tab
is closed. **While something is typed it stands out**, so empty results aren't a mystery: on phones the 🔍 button turns
orange with a yellow dot (even with the search bar hidden; its label says what is being searched), and on a computer the
search box gets an orange outline. When signed in, the
menu button shows the person's **initials** in a gold circle (e.g. «MS»; tooltip with the full name) instead of ☰. The
menu shows the signed-in name and **roles**, «Copiar mi ID», «Salir» and text size (90 %–150 %).

**Tabs look the same on every page.** View switchers («Lista / Mapa», «Lista / Mapa / Resumen» on the home page,
«📍 Puntos / 🗺️ Mapa» in Checkup on phones, «Lista / Mapa» in Scout and Cocina) are rounded buttons on a light grey bar;
the open one is dark with a shadow. Section tabs (Runners, Supply, Constructores, admin Usuarios) are underlined: the
open one has an orange line and its count in an orange pill; on phones each shows its icon above the name and the
count below.

---

## 3. Roles and access

| Role (internal name) | Label | Opens | Notes |
| --- | --- | --- | --- |
| `runner` | Runner | Runners | Takes and delivers runs. |
| `builder` | Constructor | Constructores | Repair jobs. |
| `agent` | Agente | Checkup (and Runners for packing lists only) | Sees private point data. Admin-only to grant. Changes to a need someone outside their team last updated go to that person as a request (§6). |
| `donations` | Donaciones | Donaciones | Sees donor contacts. Admin-only to grant. |
| `supplier` | Supply | Supply | A collection center. |
| `cook` | Cocina | Cocina | Cooks prepared meals for points; a team manager can grant it. |
| `teammanager` | Team manager | Mi equipo | **Must belong to a team.** |
| `admin` | Admin | Shows «🔒 Administración» in the nav | **Only the admin panel can give or remove it** (Usuarios: user form and «Asignar a un equipo»); never team managers, agents or sign-up. The admin panel itself still asks for the master password. |
| — | Admin | admin.html | Master password; not a user role. |

- A person can hold several roles.
- Admin grants roles in **Usuarios** (checkboxes). A team manager can grant or remove Runner, Builder, Supply, Cocina and Team
  manager **within their own team** (see §12).
- **Live access**: every signed-in page re-checks the current roles every 45 s (and when the tab regains focus). A new
  role shows up in the menu; a removed role ends its sessions immediately and sends the person home with the notice
  «Tu acceso cambió…».
- Sessions live in server memory: after a server restart, anyone with a remembered code signs back in automatically.
- **Scout is not a role**: every signed-in code user sees it (§14). But to sign in at all a code needs **at least one
  page role**: with no roles, sign-in says «Tu cuenta está pendiente de aprobación. Administración te dará acceso
  pronto.»; with only Admin, or Team manager without a team, it says «Tu código no tiene acceso a ningún panel.». So the
  Admin role only adds the nav link for someone who also has a page role.
- **Global agent**: an agent with no team. Agents help verify data, but an agent outside the team of whoever updated a
  need or point last (a global agent, or one from another team) sends a change request instead of overwriting it (§6).

---

## 4. Signing in and creating an account

The home sign-in window (opened by «Inicia sesión» or «¡Empieza!»), Runners, Constructores and `signup.html` show two
groups side by side, with no hidden fields (Checkup and Supply show only «Ya tengo código»; Cocina and Donaciones have a
single code box with «Entrar»; «Mi equipo» sends you to `signup.html`):

- **«Ya tengo código»** (I have a code): access code + «Recordar mi código de acceso» + «Entrar».
- **«Crear cuenta nueva»** (new account): full name, phone, email (optional) + «Crear mi código».

Details:

- **Access code**: generated by the server when the account is created (word + digits + part of the name, e.g.
  `foca72nora6`). It may be typed with or without a leading «!»; case and spaces don't matter. Codes admin assigned
  earlier keep working.
- **Approval**: new accounts are approved immediately as Runner (as Builder when created from Constructores) and are
  signed in right away. Admin can change the role or deactivate them.
- **Remember**: after a successful sign-in the code is saved on the phone; next time it is prefilled and «Entrar» is
  enough. Unchecking «Recordar» forgets it. «Salir» ends the session but keeps the remembered code. **Signing out
  is global**: from any page it ends every session this browser has (the general code session, Runners, Constructores,
  Checkup, Supply, Cocina, Donaciones and the admin master-password session) and takes you to the home page, so going
  back to any page asks you to sign in again. Pages without their own «Salir» button
  (Scout, Cocina, Mi equipo, the public point page) offer it in the ☰ menu when signed in; it asks «¿Cerrar sesión?»
  and ends the code sessions like the home page's ✕.
- **Duplicate names**: if someone already has that full name, the new account becomes e.g. «Juan Pérez 2».
- **A phone can be used again** when creating an account: it may be the same person who lost their code. They get a new
  account and a new code (never the old one); the same full name gets a number («Juan Pérez 2»). When putting people on
  a team, ask them for their user ID («Copiar mi ID» in the menu) to know which account is which. People who
  registered before codes existed (name + phone only) sign in with name and phone on Runners/Constructores.
- **Email**: optional, for code recovery: admin looks the person up by phone or email in Usuarios and shows or
  changes their code. No emails are sent.
- **Checks when creating**: full name needs first and last name, the phone 7+ digits, the email must be valid. After
  creating: «¡Listo, ya entraste!» with the code and «Copiar código». «¿Olvidaste tu código?» says to ask admin or your
  coordinator. A code pasted into «Nombre completo» moves to the code box; typing «!» fills the remembered code.
- **Limits**: 10 wrong codes in 10 minutes from one connection block every code sign-in from it (even the right code)
  for 10 minutes («Demasiados códigos incorrectos desde esta conexión. Espera N min…»). Account creation is limited to
  10 per 10 minutes per connection.
- **Session length**: code sessions last 12 hours; then the remembered code signs you back in.
- **Home page**: «Salir» is the ✕ next to your name and asks «¿Cerrar sesión?»; tapping your name copies your user ID.

---

## 5. Points

**The form** has three steps: **«1 ¿Dónde estás?»** («Ubicar en mapa», or «Enviar mi ubicación GPS» without a map; «o
pega un enlace de Google Maps»; «Zona / área», «Ubicación / referencia», «Indicaciones de llegada»), **«2 ¿Con quién
estás?»** («Reportado por», «Teléfono de quien reporta», «¿Cuántas personas?» with Personas, Adultos, Niños, Bebés,
Mayores, Discapacitados, Familias as numbers 0–999, and contacts) and **«3 ¿Qué necesitas?»** (checklist, «Otros»,
«Notas adicionales»); admin adds «Solo administración». It starts reading GPS as soon as it opens («Buscando señal
GPS…», «GPS ±N m ✓ ubicación agregada») and fills the location until you type or pick one. «Ubicar en mapa» opens a
full-screen map with a fixed crosshair, other points (green approved, orange pending), «Mi GPS» and «Usar esta
ubicación». On a computer, clicking the home map while the form is open sets the GPS. The server refuses coordinates
outside lat 6–10 / lon −83…−78, keeps unrecognized map links for review without coordinates, needs numeric counts and
at least one phone with 7+ digits; with no zone it becomes «Sin ubicación»; with no code the server assigns «P-…» and
cards show the zone instead. **Without connection** the point is saved on the phone, copied to the clipboard and sent
automatically later («Sin conexión: el punto quedó guardado en este teléfono…»); resends never duplicate it. If a
code user's session can't be renewed, the point is sent as a public report («Por revisar») instead.

A point has zone/area, place, directions, GPS (Maps link or coordinates), **«Reportado por»** (who reported it and knows
most about it, with a megaphone icon; stored as the `coordinator` field, older screens called it «Coordinador») and
contacts, counts (people,
adults, children, babies, elderly, disability, **families**), reported needs and notes.

**«Reportado por» when adding a point**: if the person adding it is signed in and **on a team**, they are «Reportado
por» by default: a card «Reportado por · Ada Agente · Tú · Equipo · phone» with **«Cambiar»** instead of fields. «Cambiar»
opens empty fields to write someone else; «↺ Soy yo» puts them back. Without a team, the usual fields (there is no
«Yo» button anymore). Contacts: «Al menos un contacto con teléfono es obligatorio»; each row has icons for person (name), phone, cake
(age) and ID card (cédula); **«＋ Agregar otro contacto»** adds a row (a new one also appears when you type in the last).

**«Nuestro equipo lo abastece»** (only for someone on a team adding a point): a switch at the end of the form. When on,
the point is **not published for 3 hours**: the public map, the camp page, the Resumen, the newsletter and other
teams' Runners and Supply don't see it (nor its needs); the team does, marked «🔒 Solo tu equipo». After 3 hours it
publishes by itself. In Checkup, an agent of the holding team can publish it sooner with «🌐 Publicar ahora» (agents of
other teams can't; the admin panel has no button for it).

**States**

- `pending` («Por revisar»): public reports, imports, unreviewed reports. Not public. Unless admin turns off
  «Requerir revisión de puntos» (admin menu): then public reports publish immediately.
- `approved`: published. Approving creates its open needs.
- `archived` («Archivados»): archived points and rejected reports («Rechazar»); hidden from the public and runners;
  «Recuperar punto» publishes it again.
- `rejected` (tab «Duplicados»): merged into another existing point with «Duplicado…».
- Signed-in code users can add points that publish immediately (with the gold medal badge, «Verificado por equipo», if they
  are in a team, see §6).

**GPS**

- Without GPS a point doesn't appear on the map; it shows with the crossed-out pin icon in the phone list (Runners
  role), Scout, Checkup, admin and Runners.
- The Maps button is only enabled with a real location (coordinates or a Maps link); otherwise it is disabled next to
  the crossed-out pin (home, Runner and Supply cards don't fall back to a place-name search; the home detail says «Sin
  GPS: confirmar ubicación antes de salir.»). Exception: the public point page searches Maps for the place and zone,
  labeled approximate («Buscar referencia en Google Maps ↗», «Sin GPS: la búsqueda es aproximada. Confirma la ubicación
  antes de salir.»).

**Point admin** (admin.html)

- «Por revisar»: Approve, Edit, Reject, «Duplicado…» (merge into an approved point).
- Select reports with checkboxes and **⬇ Exportar CSV** (all fields, including private ones).
- «🔒 Administración ▾» menu → **📥 Importar reportes (por revisar)…**: an exported CSV or JSON, **dropped onto the dashed box** («Suelta aquí el archivo», which lights up green while dragging; other file types are refused with «Solo archivos .csv o .json.») or chosen by tapping it; the file name shows under it; always imported as «Por revisar»; **«Agregado por»** (who reported it) and an optional **«Teléfono de quien agregó»**
column (JSON: `addedByName`, `addedByPhone`) link each report to a user: the one with the same phone (last 8 digits),
or with no phone the only user with that name; if there's none, the person is **created as a user** with no access
code, roles or team («Pendiente de aprobación», with a notice that the import created them), so admin can later give
them access; the preview shows «🧭 *name*» per row; `addedBy` (an existing user ID) still works; rows **without any phone** (no reporter phone, no contacts) import fine: the «at
least one phone» rule only applies to new points from the public form, Runners, Checkup and Scout, not to imports,
approvals or edits, so such a point can be approved and edited (also from Checkup), get checkups and needs, and shows
«Sin teléfono registrado» / «Sin contactos registrados» where phones would go;
  re-importing the same file doesn't create duplicates.

---

## 6. Needs and runs

A need belongs to a point and says what is missing (item text: «Agua, Arroz, Carpas»).

**States**

| State | Meaning |
| --- | --- |
| `OPEN` («Abierta» / «Necesita») | Nobody is taking it yet. |
| `IN_PROGRESS` («En camino») | A runner took it: it is a **run**. |
| ↳ **Despachada** (dispatched) | Sub-step of «En camino»: the supply center marked it as left with its load. It stays on the way until the runner taps «Completé». |
| `COMPLETED` («Completada») | Delivered. |
| `DELETED` | Removed (not shown). |

- **Need checklist**: the same one everywhere a need is entered (home «Reportar punto» and «Reportar necesidad», Checkup,
  admin point and need editors, runner follow-up after a run). Grouped as «Agua y comida», «Refugio y cuidado»,
  «Energía y comunicación», «Casa» and «¿Algo más?» (Otros), two per row. Checking a need that has options makes it
  full width and shows its options as buttons to tap (one or several):
  - **Agua**: glass icon «Para beber», pot icon «Para cocinar».
  - **Alimentos**: hot-meal icon «Comida preparada», pan icon «Ingredientes para cocinar».
  - **Combustible para generador**: Gasolina (selected automatically when checked), Diésel, «Otro».
  - **Tarjeta de datos prepago**: Tigo, +Móvil (the only active mobile networks in Panama; Claro became +Móvil in
    2023, Digicel left in 2024; typing «Claro» shows the +Móvil badge), «Otra».
  - **Gas de cocina**: Tropigas, Panagas, «Otra».
  - **Reparación de casa** (in every form): optional materials: hojas de zinc, carriola, varilla, tubo de hierro,
    bolsas de cemento, bloques, tornillos, clavos, soldadura, «Otro».
  «Otra/Otro» opens a short field to type the name. Choices are saved in parentheses, e.g. «Agua (Para beber /
  Para cocinar)», «Gas de cocina (Gas Sol)», and need chips on every page (including the public camp page) show
  them as small colored badges (not official logos). The Resumen counts all options of one need in one row. There is
  no bathroom box: the portable toilet is a recurring service; older «Saneamiento / baño» needs show as text.
- **House repair**: a separate kind of need for Builders, never public. Whoever reports it (public, runners, agents,
  admin), the server stores «Reparación de casa (materials)» as its own repair need, apart from any supplies. Builders see the requested materials on the
  card as «PIDEN hojas de zinc · varilla …». When editing a repair need, its materials can be changed.
- **Urgency**: normal or urgent (shows «Urgente»).
- **«Verificado por equipo»** (a gold medal **icon only**, no text, anywhere): a point or need created or updated by someone in a team is marked
  as checked more thoroughly. Stamped by the server when:
  - a code user in a team adds a point (home or panels);
  - an agent in a team creates a point, edits a point, or creates/updates a need in Checkup;
  - a runner in a team records a new need after completing a run.
  An update by someone without a team removes the mark (the data changed without a team check); edits by admin with
  the master password keep it. **Who sees what**: people **on a team** (and Checkup/admin) can **tap the icon** to see
  a small box «Verificado por el equipo Norte · Ana · 15 oct»; the server sends the team and person only to them.
  The public, and signed-in people without a team, see only the icon (labeled «Verificado»), with no team, person or
  date, and nothing to tap.
- **Recurring services** (garbage, portable toilet, fumigation, medical round): repeat every N days; once completed
  they are scheduled and don't count as open until due.
- **Services log («🧾 Servicios recibidos»)**: every time a service is given at a point it is logged with when, how much
  and an optional detail (e.g. «Comidas preparadas · 40 · almuerzo», «Fumigación de mosquitos · zona del río»), and
  optionally **«válido por N días»** or a **next date**, which schedules it again for then (not required). Logged
  automatically when a prepared-meal delivery is completed (§9b), or by hand from Checkup «🧾 Registrar servicio»
  (service list plus «Otro…», «Cantidad», «Detalle», «Válido por (días)» or «o próxima fecha»); admin has the same API
  (`/api/admin/camps/{id}/services`) but no screen yet. The point's public page («🧾 Servicios recibidos») lists the
  latest 20, newest first: what, how much, when, and «próxima vez» if scheduled; completions of recurring services
  appear too. Never who. Hidden while a team holds the point.
- **Prepared meals («Comida preparada»)**: a request for hot meals at a point, with the **number of people** and
  **desayuno / almuerzo / cena** (both required) and an optional note. **Only people on a team** can ask, from the home
  point detail («🍲 Pedir comida preparada»). It is cooked by a Cocina and carried by runners (§9b); it never shows in
  Runners «Runs» or Supply. In the point's team needs it reads «🍳 Esperando una cocina» or «Una cocina la está
  preparando», without the runner buttons.
- **Several runners on the same need**: each runner always has a run of their own (a linked copy, `sameAs`) with its
  own runner, supply center, packing list, photos, dispatch and «Completé», so every shipment is documented and no
  one's run is overwritten. Cards show «👥 También va: …». A second run is created when:
  - a supply center sends «otro runner» to a need that is already on the way;
  - on the home page a runner taps **«🚚 Agregar otro conductor»** on a need someone already carries (picking a runner
    who already carries it says «Ese runner ya lleva esta necesidad.»);
  - admin or Checkup choose a different runner for a need that already has one (dispatch, the need editor, or «¿quién
    la suplió?» after «Ya no necesita nada»). If that runner already has an open or on-the-way run of this need, the
    change goes to their run (for example it is marked completed) instead of creating another.
  The first runner's run is not changed. To take a runner off, use «Quitar conductor» (home) or «Sin asignar»
  (admin/Checkup); that run goes back to «Abierta» as before. For an agent outside the team of whoever updated the
  need last, any of these Checkup changes is first a change request and only happens once that person accepts it
  (see «Change requests» below).
- **Editing someone else's run**: when admin or an agent who is not the runner changes what an on-the-way or completed
  run asks for, the run keeps its text (what the runner is carrying). What was added becomes a **new open need** of the
  same point, linked to the run and noted with who added it (private: `createdByName`, never public). Urgency and
  state still change on the run. An edit that only removes things from a runner's run is refused with «Lo que lleva un
  runner no se cambia desde aquí: solo se pueden agregar cosas, que quedan como otra necesidad.» (409). Open needs
  without a runner are edited in place as before. An agent outside the last updater's team goes through a change
  request first; these rules apply when it is accepted.
- **Adding to someone else's run**: packing lines and departure photos are always added (never replace earlier ones)
  and each one keeps who added it.
- **Teams and global agents**: people on the same team work together and trust each other's data; their edits apply
  directly. A **global agent** is an agent with no team: they help follow up and verify data (calls, checkups) but
  can't overwrite, lose or delete a team's or a runner's work. Agents from another team are treated the same way.
- **Change requests («📝 Cambios propuestos»)**: the person who was at the point is the better source of truth, so an
  agent from outside can suggest changes but can't impose them.
  - **Who decides** (private, never public):
    - for a **need**: who updated it last (`updatedBy`: the runner or builder who took, completed or dropped it, the
      runner who recorded it after a run, a code user who assigned a runner from the home page, the agent who edited
      it, or whoever accepted a request); if that person can't answer, the run's runner, then its builder;
    - for a **point's data**: who edited it last from Checkup (or accepted a request), otherwise who added it with
      their code.
    Only people who can answer on a page count (Runner, Constructor or agent, not disabled).
  - When an agent changes a need (editor, «Despachar», «Completada», «Ya no necesita nada», after a «Checkup») or a
    point's data («Editar datos») and that person is **outside the agent's team** (or the agent has no team), nothing
    changes yet: the change is saved as a request to them. Checkup shows «📝 Cambio enviado a … para que lo acepte o lo
    rechace: actualizó esta necesidad / este punto por última vez». One pending request per agent and record; a new
    one replaces it.
  - The change applies directly when the agent is that same person or on their **same team**, for admin with the
    master password, for edits that change nothing, and when nobody can answer (a point reported by the public).
  - Adding packing lines or photos, creating a new need, and logging the checkup itself never need a request:
    nothing is overwritten.
  - The person sees it at the top of **Runners** and **Constructores** (above the tabs) or **Checkup** (under the
    filters): who asks (name and team, or «sin equipo»), the point, the need («Datos del punto» for point edits), and
    each change as «Label: before → after», with **«✓ Aceptar»** or **«✗ Rechazar»**. Accepting applies the change as
    the agent's edit would (including the run rules above, so items added to a run still become a new open need) and
    the record stays theirs; rejecting changes nothing. The list refreshes every 30 s.
  - In Checkup, **«📤 Cambios que pediste»** lists your requests of the last 7 days: «⏳ Esperando a», «✓ Lo aceptó» or
    «✗ Lo rechazó».
- **Never by an outside agent** (refused with «Esto es de …: solo esa persona, su equipo o administración lo pueden
  cambiar.», 403): deleting or moving (pickup ↔ delivery) a photo someone else uploaded (allowed for your own, your
  team's and admin's photos), merging a duplicate report into a point that belongs to another person or team, and
  «🌐 Publicar ahora» on a point another team is holding.
- **Change history («🕓 Historial de cambios»)**: every change is recorded so the trail of any data can be followed:
  points (every field, including people, contacts, notes, GPS, hold and checkups), needs and runs (request, state,
  urgency, runner/builder, packing lines, photos, supply center and offers, «Despachada», «Pidió a acopios», repairs),
  and supply centers (each person's center, rename, archive, restore). Each entry has who (name, team or «sin equipo»,
  «Administración», «Público (sin sesión)» or «Sistema»), from which page (Admin, Checkup, Runners, Constructores,
  Supply, Cocina, Donaciones, Scout, Inicio; «Checkup · aceptado por Fufo» for an accepted request), when, and each field as before →
  after (names instead of IDs). Admin and Checkup show it per point, newest first, in a collapsed «🕓 Historial de
  cambios» (the point and all its needs). Admin can also read the supply-center trail (`/api/admin/centers/trail`; no
  screen yet). Recording started on 2026-10-15; earlier changes are not in it. Private: not on public pages.
- **Packing list**: what is loaded for that run (item × quantity). Private. The «Agregar a lista de empaque» window
  (Runners, Supply, Checkup, admin) shows **«Lo que necesitan»** at the top: the need's items as chips that turn grey
  with ✓ once the saved list or what you're adding covers them (e.g. «Agua» covers Agua, «Bolsas alimentación» covers
  Alimentos, «Carpas / toldas» or «Camping» cover Refugio; otherwise by matching words). With nothing typed, the
  list starts with **«🙌 Para lo que necesitan»**: the items that cover needs not crossed out yet, in the needs' order;
  then the **categories** as buttons with their count (💧 Agua y comida, ⛺ Refugio: carpas, lonas, camping, hamacas,
  mosquiteros, colchonetas, sillas; 🍳 Cocina, 🧼 Higiene y limpieza, 🍼 Bebé y salud, ☔ Lluvia, 🔦 Luz y energía, 📦
  Otros). Tapping one shows only its items under «‹ Todos › ⛺ Refugio» («‹ Todos» goes back). Typing searches every
  item, those covering what's missing first. Catalogs without categories (builders' materials, dishes) stay a list. Starting
  items now include **Agua** and **Lonas**. Choosing the quantity shows only «Agregar a la lista» and «Cancelar» (back
  to the items); in the review, «Cancelar» also goes back to the items; on the item list it closes (asks if anything
  isn't saved).
- **Photos**: departure photos (when loading) and delivery photos (when completing). Private.

---

## 7. Runners (`runner.html`)

Tabs: **Yo** (my runs on the way), **Runs** (open needs, **newest first**: new or updated most recently, which is the
data we know is still valid; same time → urgent first), **📦 Pedidos** (supply centers asking for a runner) and
**🍲 Comidas** (a Cocina asking for drivers for prepared meals, §9b). The page opens on «Yo» if you have runs,
otherwise «Runs», remembers the tab for the session and refreshes every 30 s. On phones, «Lista / Mapa» switch between
the list and the map. A link `/runner.html?punto=<id>` (and `&necesidad=<id>`) opens the panel on that point or run.

- **Agents** who open Runners see a single **«Listas de empaque»** list instead of the tabs (open and on-the-way needs,
  except other teams' runs) to add packing lines; they can't take runs or meals («Como agente puedes armar listas de
  empaque, pero no tomar carreras.»).
- **«Todas / Solo mi equipo»** (with a code): limits «Runs» to points your team added or has a run on; your own runs
  always show.
- Needs that a supply center offered to supply leave «Runs» and appear only in «📦 Pedidos». Recurring services
  appear when they are due.
- **«+ Agregar punto»** in the header opens the point form in place; with a code the point is published at once.

- **Buscar** (search): in the header, like the home page; on phones it opens from the 🔍 button. It filters the open
  tab (Yo, Runs, Pedidos) and the map by point code, zone, place, need, coordinator, supply center or runner,
  ignoring accents and case; tab counters follow it. In «Yo», non-matching runs are hidden without changing stop
  numbers or order.

- **«¡Yo lo llevo!»** asks for confirmation (point and what to bring); accepting moves it to «Yo» with the run
  highlighted.
- **Picking from the list** (the 📍 «Ver en el mapa» button): the map centers on the point and opens a short popup with
  «¡Yo lo llevo!» and «Ver en la lista».
- **Card**: point name with the 📍 «Ver en el mapa» and Maps buttons at the **top right**; then where it is; then how many
  people with the breakdown and icons («40 personas · 20 adultos · 15 niños · 9 familias», zeros left out) and how
  long ago it was reported; then «Necesita» (what the point asked for) separate from «📦 Lo que llevas» (packing list),
  contacts, and at the bottom share + the action. Tapping the card background does nothing (no highlight); a card you
  jump to from the map flashes briefly. The location row always uses the plain pin and reads «Sin indicaciones
  escritas» when there are no directions; a point without GPS is marked only by the crossed pin at the top right. The
  reporter's name (where the text already names them, or «📣 *name*» added at the end; long names end in «…») is
  tappable and opens **«📞 Llamar»** and **«💬 WhatsApp»** with their phone; there's no separate «Reportado por *phone*»
  line. With a supply center: «📦 Recoger en acopio: Villa Marina · Julie» plus phone.
- **Packing list**, **📷 Foto de partida**, **✓ Completé** (the confirmation shows what is being delivered and allows
  delivery photos) and «No puedo continuar» (the need goes back to open; its packing list and photos stay on it, each
  with who added it, and the next runner who takes it sees them). A prepared-meal run that is dropped is removed
  instead of reopened; the Cocina asks for another driver.
- **«📝 Cambios propuestos»** (above the tabs): changes an agent from another team (or with no team) asks to make to
  a need or point you updated last; «✓ Aceptar» or «✗ Rechazar» (§6). Builders see the same box on Constructores.
- **Only open needs can be taken** in «Runs»: a need another runner already carries can't be taken here. A second
  runner on the same need is added from the home page («🚚 Agregar otro conductor»), admin, Checkup or Supply, and gets
  a run of their own (§6).
- **After completing**: «📝 Actualizar necesidad» (opens the same need checklist to record what is still missing,
  including a house repair, which goes to Builders) or «✓ No necesitan nada más».
- **Gold badge**: a card shows the gold «Equipo …» badge next to the point name when a team verified the point or that need.
- **«📦 Pedidos»**: open needs that one or more supply centers offered to supply. The card header reads «Un acopio la
  abastece» or «N acopios la abastecen: elige el más cercano» and lists every offering center (center, person, phone)
  with its own **«¡Yo lo llevo! · recojo aquí»**: pick the closest (the confirmation shows «Recoges en 📦 …»); the run
  is linked to that center and shows «📦 Recoger en acopio: …».
- **«📣 Pedir a acopios»** (on a run with no center; not on meal runs): while waiting it shows «⏳ Pidiendo a acopios …
  espera a que respondan» and «Cancelar pedido»; centers answer whether they can supply it; the card lists those that
  can («✓ 2 acopios pueden surtirla: elige el más cercano») and «Elegir» links the run to the chosen center (the other
  centers stop seeing it).
- **«🍲 Comidas»**: prepared-meal requests where a Cocina asked for drivers or invited you. Invitations to you come
  first, highlighted «🍳 La cocina te lo pidió a ti»; then by earliest pickup. Each card: «🕐 Recoger a las … en la
  cocina de …» with the cook's phone, the dishes and «👥 Ya van: …». **«✓ Acepto llevarla»** (anyone, when open to all)
  and «No puedo» (only if invited). Accepting gives you your own run in «Yo» (it counts toward the 10-run limit), shown
  with «🍳 Recoger a las … en la cocina de …» and the cook's phone, no supply-center actions. Search doesn't filter
  this tab.
- **Live location**: turned on with **«📍 Mi ubicación en vivo»** (offered on entry: «¿Activar tu ubicación en vivo?»
  «📍 Activar» / «Ahora no», which lasts until the tab closes). The first time, the runner must accept the location
  notice. Tapping it again («📍 En vivo · compartida con el equipo») turns it off. It is shared **only with the
  runner's own team** (strict rule, on every page: other teams, the public, people without a team and other supply
  centers never see it). A position is sent every 15 s or after moving 50 m and kept alive while standing still; a
  position not updated for 2 minutes disappears. Losing signal doesn't stop it; denying browser permission, turning
  it off or logging out does. Agents never share their location. The map follows the runner until they pick a run or
  move the map; «🎯 Ir a mi ubicación» resumes following.
- **Route on the map**: with live location on, the map draws the street route to stop 1 (orange) and dotted lines to
  the next stops (refreshed when stop 1 changes or after 150 m, at most every 15 s).
- **Team on the map**: with a code, your team's runners sharing their location show as orange trucks with name and
  «hace …», refreshed every 15 s.
- **Offline**: a screen still shows the route; photos and «Completé» stay on the phone and are sent when the
  connection returns.
- **Reorder** the route by dragging ☰ or with ▲ ▼. A runner can have up to 10 runs on the way at once.

---

## 8. Supply (`supplier.html`)

Top line: the **📍 Centro** button shows the chosen collection center («📍 Villa Marina ▾»; orange «Elige tu centro»
when none). Tapping it opens the shared list of centers with a field to filter or type a new one («＋ Usar «…» (centro
nuevo)»); picking saves it right away. At the bottom of that list, **«⚙ Administrar centros…»** (rarely needed) opens a
window to **rename** a center (for everyone on it, history included) or **archive** it: it disappears from the list
to choose from, but nothing is deleted (people keep it, runs and history stay intact) and **«↺ Recuperar»** brings it
back; both ask for confirmation. The **Solo mi equipo** switch sits at the right of the same line. **Buscar** (runner,
point, need, who added it, who reported it or the point's contacts, by name or phone) is in the header like other pages
(🔍 on phones). When a tab has nothing matching the search, it says «Ninguna necesidad coincide con la búsqueda «…»» with what you
typed and a **«✕ Borrar búsqueda y ver todas (N)»** button (N = what the tab has without the search, respecting «Solo mi
equipo»).

**Map**: the points of the open tab plus **your team's runners sharing live location** (orange trucks), even when they
are not on a run of this center; with no points it centers on your team. Runners of other teams never appear.

Tabs:

| Tab | Content |
| --- | --- |
| 📦 ¡A empacar! | Runs linked to this center, on the way, not dispatched yet: pack them and dispatch. |
| 🏃 ¡Abastéceme! | Runners with a run asking for a supply center: «✓ Puedo surtirlo» / «✗ No puedo» / «Retirar». |
| 🙌 Por surtir (last) | Open needs: **«🙋 ¡Yo abastezco!»** offers to supply it (several centers can offer the same need; the runner chooses where to pick up, the closest) or «Retirar oferta»; **«🏃 Asignar a runner»** links it to this center with a runner who is here. The card says how many centers offer it. Below, «🚚 Ya en camino»: other people's runs, with «🙋 Abastezco otro envío» / «🏃 Asignar otro runner». |
| 🕓 Historial | On phones the other tabs hide while you are in the history. First **«🚚 Despachadas · en camino»** (dispatched, still on the way, with «Deshacer»), then **«✓ Entregadas»** of the last 30 days (up to 60, newest first): point, items, runner, when it left and when it was delivered, packing list and photos. On phones it is the clock button next to «Lista / Mapa»; on desktop, the clock at the end of the tabs. There is no separate «Despachadas» tab. |

On phones each tab shows its icon above the name and the count below.

**Cards** look like Runners cards: the 📍 map and Maps buttons at the top right, the people count with its breakdown
before «Necesita», and contacts show the first 2 with «y N contactos más» to expand the rest (same on Runners).

**«¡A empacar!» card, in this order** (runs of this center)

1. **🙌 Necesita** (light pink): what the point needs, with «Urgente». Each need already covered by the packing list
   shows crossed out and grey with ✓ (same matching as the packing window); when all are covered this section and «Para
   cargar» both turn the same light green.
2. **📦 Para cargar**: packing list, departure photos, «🚚 Despachada» (with confirmation) with the share-link button
   next to it, or «Salió del acopio» with share and «Deshacer» (undo).
3. **📍 Destino**: point, directions, people, contacts, «🧭 Agregado por» (team members, as in «Por surtir»), show on
   map / Maps.
4. **🏃 Runner**: name, phone, «La tomó hace…» (taken … ago), live («📍 Ver runner en el mapa») if sharing location and
   on your team, who dispatched it, «👥 También va».

«¡Abastéceme!» cards keep Runner, Destino (with «Necesita» and share) and «¿Puedes surtirlo?».

**Each supply center only sees and dispatches the runs linked to it.** A run becomes linked to a center when:

- the center offered it («¡Yo abastezco!») and a runner chose this center in «Pedidos»;
- the center assigned it to a runner who was there («Asignar a runner»);
- a runner asked for a center and chose this one;
- the center sent «otro runner» (a new linked run).

**Solo mi equipo** (my team only, remembered on the phone): «Por surtir» shows only points my team added; the other
tabs show my team's points or runs by my team's runners; the runner picker shows only my team's runners.

- **Order**: «Por surtir» shows the **most recently updated need first** (created, edited, urgency changed, released…), and each card says when: «🕓 Actualizada hace 3 min» (exact date and time on hover; the people row has no separate «hace …» there). Its buttons are on one row: «🙋 ¡Yo abastezco!», «🏃 Asignar a runner» and the share-link button. On every Supply card the reporter shows as just name and phone (no «Reportado por» label); for people **on a team**, «Por surtir» cards also show **«🧭 Agregado por *name* 🏅 *phone*»** (the «Verificado» medal sits by that name instead of the point's title) (who added that need, from the change history, or else who added the point); tapping the phone opens «📞 Llamar / 💬 WhatsApp». People without a team don't get this line (the server doesn't send it); «¡A empacar!» shows
  the most recently taken run first.
- **«✗ No puedo»** in «¡Abastéceme!»: this center stops seeing that ask; others can still answer. The tab counter turns
  orange while there are asks you haven't answered.
- **Without a center**: «¡Yo abastezco!» or «✓ Puedo surtirlo» asks «¿Pedir/Ofrecer sin nombre de centro?», because
  runners then see only your name and phone.
- **Runner picker** («Asignar a runner»): search by name or phone; shows each runner's phone and «N en camino»; a
  runner who already has 10 on the way is refused.
- **Runner section of the card**: «📡 En vivo · hace N s» or «Sin ubicación en vivo» and «Despachó …». Map pins are
  orange (to dispatch) or green (dispatched). Tapping a card background centers the map (unlike Runners).

### Supply ↔ Runner flows

```
A) Centers offer to supply it
   Supply «🙌 Por surtir» → «🙋 ¡Yo abastezco!» (one or more centers) ─▶ runners see it in «📦 Pedidos» with every
   offering center (who, center, phone) ─▶ the runner taps «¡Yo lo llevo! · recojo aquí» on the closest
   ─▶ run in the runner's «Yo» and in that center's «¡A empacar!»
   ─▶ center packs + departure photo ─▶ «🚚 Despachada» (now in 🕓 Historial) ─▶ runner «✓ Completé»

B) Runner who is at the center
   Supply «🏃 Asignar a runner» (pick the runner) ─▶ same as A from «¡A empacar!»

C) Runner asks for a center
   Runner (in «Yo») «📣 Pedir a acopios» ─▶ centers see it in «🏃 ¡Abastéceme!»
   ─▶ «✓ Puedo surtirlo» (one or more) ─▶ runner «Elegir» the closest
   ─▶ run in that center's «¡A empacar!» ─▶ same as A

D) Another shipment of the same need
   Supply «🙌 Por surtir» → «🚚 Ya en camino» ─▶ «🙋 Abastezco otro envío» ─▶ new open shipment offered by this center,
   in runners' «📦 Pedidos» (then as A); «🏃 Asignar otro runner» ─▶ new run for that runner, straight into «¡A empacar!»
```

If a runner drops a run that came from a center's offer, it goes back to «Pedidos» with that center still offering. If they drop one assigned in person, it
becomes a normal open need. Dispatching doesn't change the state: the runner completes it.

---

## 9. Builders (`builder.html`)

Tabs «🔨 Mis obras» (my jobs) and «🚩 Obras» (jobs). A builder can have up to **20 jobs** in progress («Tomar obra» is
disabled at the limit). «📝 Cambios propuestos» sits above the tabs (§6). Map pins: blue for your jobs, green for
available ones.

- **Job card**: «Avance» is a slider in 5 % steps, saved when released; «Qué falta para terminar» saves when you leave
  the box; «Materiales necesarios» with «Usar materiales de la evaluación» (copies the latest assessment's materials)
  and «+ Agregar materiales» (same items add up; with none: «Agrégalos para poder pedir donaciones»); photo rows
  «Antes / Avance / Final»; «No puedo continuar» drops the job; **«✓ Obra terminada»** asks for a final photo first
  when photos are available, then confirms and sets it to 100 %.
- **Assess** a job without taking it: **«📋 Evaluar»** (becomes «📋 Mi evaluación» once you have one: one assessment
  per builder, editable only by its author; up to 20 per job; completed jobs can't be assessed). Work items «Techo /
  zinc», «Paredes / bloque», «Piso / losa», «Estructura / columnas», «Puertas y ventanas», «Electricidad», «Plomería /
  agua», «Baño / letrina», «Soldadura», «Repello / pintura», «Limpieza / escombros» plus free text; materials, days and
  people (0–365), notes and photos. Shown to all builders as «Evaluaciones (N)», newest first. Repairs are never public; their materials are totaled in the Resumen.
Repair needs can come from any form (public, runners, Checkup, admin); when the reporter picked materials, the card
shows them under the title as «PIDEN hojas de zinc · varilla …».

---

## 9b. Cocina (`cook.html`, prepared meals)

For people with the **Cocina** role (given by admin or by their team manager). Signs in with the access code like the
other pages; «🍳 Cocina» appears in the top bar.

Laid out like Supply: no title; on a computer the **map** (requests with GPS) is on the left and the list on the right;
on phones «Lista / Mapa / 🕓» switch between them. At the top of the list:
- **Kitchen picker «🍳 *name* ▾»** (orange «Elige tu cocina» until set): search the shared list of kitchens or type a
  new one and tap «＋ Usar «*name*» (cocina nueva)» (Enter picks the first line). It's saved on the person («Cocina
  guardada: los runners la ven al recoger.») and runners see it on their run: «🍳 Recoger a las … en *kitchen* · *cook*».
- **«Solo mi equipo»** switch (remembered on this device): only requests from your team or for points your team added.
- Tabs **🍳 Preparando** (requests this cook took) and **🍲 Pedidos** (meal requests waiting for a cook, plus regular
  needs that ask for «Comida preparada»), and at the end a **🕓** tab: «🕓 Historial · comidas que completaste
  (7 días)». The page refreshes every 20 s.
- Taking a **regular need** from «Pedidos» asks for the number of people and the meal (desayuno / almuerzo / cena); it
  becomes a meal request this cook is already preparing, linked to that need.
Each card: the point (area · place, Maps ↗), «🍲 40 personas · Almuerzo», who asked and when, the point's people count.

1. **«🍳 Tomar este pedido»**: only **one cook at a time**; another cook gets «Esta comida ya la tomó otra cocina.» and
   sees who has it. «Soltar pedido» gives it back; it disappears once any runner has accepted, even if they already
   delivered («Ya hay runners con esta comida: no se puede soltar.»).
2. **🍽️ Platos** (optional packing list): the dishes (e.g. «Arroz con pollo × 40»), with suggestions; runners see them.
   Dishes added after runners accepted are copied to their on-the-way runs.
3. **📷 Foto de partida** from the kitchen, visible only to the cook (runners add theirs on their own run).
4. **🚚 Conductores**: pick the **pickup time** («Recoger a las»: a time of day for today, Panama time, defaulting to one
   hour from now) and
   - **«📣 Pedir conductores»**: open to every runner in Runners «🍲 Comidas» («🔔 Pidiendo conductores»; «Dejar de
     pedir» closes it), or
   - **«Asignar»** a specific runner: they see «🍳 La cocina te lo pidió a ti» and must tap **«✓ Acepto llevarla»** (or
     «No puedo», which takes them off the cook's list so the cook can invite someone else); until then the cook sees
     «⏳ Esperando que acepte». The «Asignar» list shows active runners with their team, no phones; inviting someone who
     already carries it: «Ese runner ya lleva esta comida.» «Dejar de pedir» also cancels pending invitations.
   Several runners can carry the same request; each one who accepts gets **their own run** («Yo» in Runners) with the
   pickup time, the cook's name and phone, the dishes, photos and «✓ Completé».
5. When **the last run on the way is completed**, the request is completed and the point logs the service
   «🍲 Comidas preparadas · 40 almuerzos» with the date and time. With no runs on the way, **«✓ Terminar»** closes it
   (and logs it) by hand; while a run is on the way it is refused («Hay runners en camino: se termina sola cuando
   entreguen.»). Agents can't accept meals. After completing, the runner can update the point's needs as after any run.

Privacy: the public point shows the request («Comida preparada · 40 personas · almuerzo») and the service log, never
who cooks or carries it.

---

## 10. Checkup (`command.html`, agents)

**Buscar** is in the header (on phones, the 🔍 button), like Runners, and «+ Agregar punto» is always at the top right:
it opens the same point form **right in Checkup** (no trip to the home page; a team agent is the coordinator by
default) and publishes the point on «Guardar», opening it in the list.
Access: its own code gate («Ya tengo código», «Recordar mi código de acceso», «Entrar»); only Agente users get in.
Signing in with a code opens the sessions of all the user's panels at once; tapping your name copies your user ID.

The totals are shown with icons: 📍 points (or «4/54» while searching or filtering) and 🕓 «por revisar»; on phones they
are inside the «📍 Puntos» tab (the other tab is «🗺️ Mapa»). List of points with filters (state, area, need,
urgent/open/on the way), a map and two monitors: **«En camino»** (needs on the way, longest first; orange from 6 h, red
from 24 h; tap to open the point) and **«Sin GPS»** (points not on the map, those with open or on-the-way needs first,
then pending ones; tap to open in edit mode and add the location). Pin colours: red for an urgent unfinished need,
orange for open needs, green when nothing is pending, grey for «por revisar». Sort: «Orden: checkup más antiguo» (default),
«checkup más reciente», «prioridad», «más recientes», «código», «área» (a sort saved before this change switches once to
the new default); need chips are buttons with counts; filters are remembered in
this browser. Opening a point gives the link `/command.html?punto=<id>`; on phones the open point is shown alone with
«← Puntos». Each card:
point name (the «Verificado» medal goes by «🧭 Creado por *name*» when the creator is known), place and reporter on
one line, the people count on its own line, **«Necesita reparación estructural»** (repair icon) when the point has a pending home repair (repairs
are not listed with the card's needs, aren't shown on the checkup screen, and a checkup never removes them; they stay in
the opened point and in Constructores), the group's names (up to three, «y N más»; no phones), the needs as they are (no counters), the orange
**«Checkup»** button (log a call; top right, the card's main action; **green** if the last checkup was in the last
24 h, **yellow** between 24 and 48 h, **red** if older or never, with «Último checkup hace …» on hover; no «Editar punto» on the card: open the
point and use «Editar datos»). Each need has **«Actualizar»** on its top line, next to «Necesita», aligned right (tapping the need's
background does nothing); «+ Reportar
necesidad» adds a new one. A small **«Detalles»** button under «Checkup» opens the point (full data, edit, needs). At the bottom, approved cards show **«Últimos checkups»** (up to 3): «📞 *who* · hace 3 min · ✓ Están bien»
(short answers: «✓ Están bien», «🙌 Necesitan cosas», «📵 No contestó · nada»; who and answer when recorded, the exact
date on hover). The card's top line says **«Checkup hace 56 min»** with **📞 *who did it*** (or «Creado / Suplido hace …»);
the creator shows only in «🧭 Creado por», where tapping the name opens «📞 Llamar / 💬 WhatsApp / 📋 Copiar» when
their phone is available (runners and builders; agents' phones are never shown to other agents). Below «Detalles», **«Ver en mapa»** (same pin as Runners) shows the point on
the map, or a red crossed-out pin when it has no GPS; the card itself isn't clickable or selectable. Phone numbers
anywhere (runner, contacts, requester) open «📞 Llamar / 💬 WhatsApp / 📋 Copiar» when tapped, the same on every page.
The phone tab «📊 Overview» is now **«🗺️ Mapa»**, and the default order is «checkup más antiguo» (also «checkup más
reciente»). Opening a point shows full data, contacts, changelog and its needs (update, report a new one, mark
services done). A **pending** point shows only «Editar datos» and **«Aprobar»** (no Reject or «Duplicado…»: those stay
in admin); approved points show «Editar datos», «Checkup», «🧾 Registrar servicio», «Página pública ↗» and the Maps
link (without GPS: «Búsqueda aproximada: no hay GPS registrado.» or «Sin ubicación registrada.»).

**Needs**: each has «Actualizar» (on its top line, aligned right), «Lista de empaque», «Completar» and **«Enlace para runner»** (copies or shares
`/runner.html?punto=…&necesidad=…`, which opens that run in Runners). States «Necesita / En camino / Completada».
«+ Reportar necesidad» has a step 2 **«¿La suplimos ahora?»**: optional packing list, dispatch to a runner and departure
photos («📷 Tomar foto», «🎥 Grabar video», «🖼️ Galería»); «Más tarde» leaves it open. The runner picker has «Solo mi
equipo», on by default for agents on a team.

**«Checkup»** opens a **«📞 Checkup» screen** (full screen on phones, with «‹ Atrás»; the phone's Back closes it): the
point's contacts, who reported it (marked with 📣) and each person of the group without repeated numbers, each as a card
with name and number on one line and two big buttons **«📞 Llamar»** and **«💬 WhatsApp»**; below them **«🙌 Lo que reportaron que necesitan (N)»**, the
point's pending needs (state «Necesita» / «🏃 En camino» / «🔨 En obra», «Urgente», chips, how long ago), to confirm on
the call; at the bottom, fixed, «¿Lograste comunicarte con
alguien?» with the answers (and how many needs «Sí, están bien» would remove): «No» does nothing; **«Nadie contestó, pero no
necesita nada»** logs the checkup and removes the needs still to be delivered, like «Sí, están bien»;
«Sí, están bien» logs the checkup and removes the needs still to be delivered; «Sí, necesitan cosas» logs it and opens
the needs to update. The checkup itself is always saved (it never becomes a request). Each answer is kept in the point's private checkup log (when, the answer, who), shown in the opened point under
«Último checkup» («Oct 16, 2026 3:10pm · Nadie contestó · no necesita nada · Ana»); never public. Removing or updating a need that
someone outside your team updated last becomes a change request to that person (below), so after «Sí, están bien»
those needs stay until they accept.

In «Actualizar necesidad», **«Ya no necesita nada»** asks **«¿Alguien la suplió?»**:

- «No, ya no la necesitan» removes the need (`DELETED`) and counts as a checkup on the point.
- «Sí, la suplió alguien» asks for the Runner (or Constructor for a repair), required, and saves the need as
  `COMPLETED` under that person with the original items, like a finished run. A new name creates the runner with
  their phone. If another runner already has the need on the way, that run stays as it is and the person chosen gets
  their own completed run, linked to it (§6).
- Both answers are change requests when the need was last updated by someone outside your team.

**🧾 Registrar servicio** (on an opened approved point): «Servicio» («Comidas preparadas», «Recolección de basura»,
«Cambio de baño portátil», «Fumigación de mosquitos», «Ronda médica» or «Otro…», which asks «¿Cuál?»), optional
«Cantidad», «Detalle» and «Válido por (días)» or «o próxima fecha»; it shows in the point's «🧾 Servicios recibidos»
(§6).

**Changes to someone else's data**: if the need or point was last updated by someone outside your team (or you have
no team: a «global» agent), any change to it (editor, «Despachar», «Completada», «Ya no necesita nada», the need
removals of a checkup, «Editar datos») is sent to them as a request
instead («📝 Cambio enviado a …»); they accept or reject it on their page, and you follow it in
«📤 Cambios que pediste». Requests sent to you (as the last one who updated a need) show as «📝 Cambios propuestos»
under the filters (§6). Who decides: for a need, the last person who updated it (else its runner or constructor); for a
point, the last one who edited it in Checkup (else whoever added it with a code). They must be an active Runner,
Constructor or Agente; otherwise the change applies directly. One pending request per agent and record (a newer one
replaces it); «📤 Cambios que pediste» keeps the last 7 days. You can't delete or move another team's photos, merge into another team's point, or
«Publicar ahora» another team's held point. An opened point has **«🕓 Historial de cambios»** (who changed what and
when, §6).

Once a change applies (directly, or when accepted), in the need editor (and «Despachar»), choosing a **different runner** for a need that already has one adds a run
for that runner instead of replacing the first one; «Sin asignar» takes the runner off. Changing what an on-the-way or
completed run asks for (when you are not its runner) keeps the run as it is and saves what you added as a new open
need of the point, marked «Agregada por» you; removing items from a runner's run is refused (§6).

**Filters**: status (Todos / Por revisar / Aprobados) and **«Agregado por»**: Todos, «Míos» (added by me) and «Equipo»
(added by anyone on my team; only for agents on a team), each with its count; area, need type, «Sin GPS», need chips and
the **«Orden: …»** dropdown (checkup más antiguo — the default —, checkup más reciente, prioridad, más recientes,
código, área). **On phones** only two controls
show: **«Filtros ▾»** (with how many are active) and the «Orden: …» dropdown. «Filtros» opens a **full-screen page**
(big buttons, sections «Buscar», «Estado», «Agregado por», «Área y tipo», «Necesidades»; at the top «Limpiar
todo» and «Listo», at the bottom «Limpiar todo» and «Ver N puntos» with how many will show; «Limpiar todo» is greyed
out when nothing is set) with **«Buscar»** (the same search as the header: typing in either updates the other; it counts as an active filter
and «Limpiar» clears it) and all the filters (state, «Agregado por», area, type, «Sin GPS», «Limpiar», need chips); «Listo», tapping outside or
Escape closes it.

**Rows**: a short line «Revisado hace 1 h» (or «Creado / Suplido hace …»; exact date on hover) with a **badge of who
added it** (full name; green with «(tú)» if it was you), the
point name, a line with small icons for place, people and coordinator, then the counters, «Checkup» and «Editar».
Point rows and need cards show the gold medal icon; tapping it shows «Verificado por el equipo X · who · date» (it doesn't open the row). A point a team is holding
shows «🔒 Solo equipo X · se publica a las …» with **«🌐 Publicar ahora»** (asks first). Editing a repair need shows
its materials, which can be changed. **Home repairs** («Reparación de casa») read as a job, not a delivery: the card
says «🔨 En obra» instead of «En camino» (and «🔨 Avance N%» in the row once the builder reports progress); opened, it
shows a **«🔨 Obra»** box with what the builder documented in Constructores: progress bar, «Qué falta», materials,
«Evaluaciones (N)» (who, when, work, materials, days, people, notes) and photos (Evaluación, Antes, Avance, Final), or
«El constructor todavía no documentó la obra…». After reporting one, the follow-up asks «¿Quién hace la obra?» with
«Asignar a un constructor» / «Asignar constructor» instead of «Despachar». Works on phones (single column).

---

## 11. Donations (`donations.html`)

Donation tickets: **Por confirmar → Confirmada → En acopio → En camino → Entregada** (or Cancelada). Each ticket:
donor (private), items, collection center and arrival date, transport, destination point with its GPS (optional
«GPS exacto de entrega (enlace de Maps…)»), departure date, notes and history («Anotar en el historial»).
Access: Donations role (single code box + «Entrar») or admin (menu «🎁 Donaciones»).

- **«+ Registrar donación»**: saved as «Guardar por confirmar» or «✓ Guardar confirmada». Item types «💧 Agua»,
  «🍚 Comida», «⛺ Carpas», «🚽 Baños o letrinas», «🍼 Pañales», «⛽ Combustible», 🧱 repair materials and «Otra cosa…».
- **Steps**: «✓ Confirmar donación»; «📦 Llegó al acopio» (optional) or «🚚 Salió al punto»; «✅ Entregada»; «Cancelar
  donación» (asks first); «Reabrir» for cancelled tickets. Contact buttons «📞 Llamar» and «💬 WhatsApp».
- **Filters**: «Activas» (default), each state, «Todas», each with a count, and a search.
- **Top cards**: «Necesario» (the home-page calculation for 5 days), «Comprometido» (confirmed, at the center or on the
  way), «Por confirmar», «Entregado», with «Faltan N» or «✓ Cubierto».
- A rate-limited public offer API exists (`POST /api/donations`, 10 per 10 min per connection) but no page uses it.

---

## 12. Teams and Team manager

- **Teams**: admin creates them in Usuarios → 👥 Equipos (or from «👥 Mover a equipo ▾») and assigns users (one team
  per user; code users only). Teams can be created and renamed, not deleted. Teams drive «Solo mi equipo» (Runners,
  Supply, Checkup's runner picker) and who sees the team's live locations.
- **Teammates trust each other's data**: their edits to each other's points, needs and runs apply directly. People
  outside the team (global agents, other teams) can only propose changes (§6), and every change by anyone is recorded
  in the change history. Teammates see each other's full names in a point's public activity log, and Scout can filter
  the team's points (§14).
- **User notes** («Notas»): free text about a person (skills, vehicle, schedule, languages…). Admin writes them in
  Usuarios (user form → «Notas»; shown in the user's view); the person's **team manager** writes them in «Mi equipo»
  («📝 Notas» on each member, also their own). Never public; other teams don't see them.
- A **Team manager** (must be in a team) opens **Mi equipo** from the ☰ menu (also shown to users with the **Admin**
  role, read-only there; other team members don't get the page):
  - **Notes**: each member's notes under their roles; «📝 Notas» opens a box to write them («Guardar notas»).
  - **Invite**: «Crear enlace» (expires in 24 h, 3 days (default) or 7 days; **single use**). Copy, share, or send via
    WhatsApp. The last 20 invites: «Pendiente · vence …», «Usada por …», «Vencida», «Anulada», «Creada … por …», with
    «Anular».
  - **Members**: a quick search above the list filters as you type by name, phone (3+ digits) or role, ignoring
    accents and case (the counter shows e.g. «1 de 3»). Each member shows an online dot, «Tú» for yourself, phone, role
    chips with the page icons (🏃 Runner, 🔨 Constructor, 📦 Supply, 🍳 Cocina, 👥 Team manager, 🧑‍💻 Agente, 🎁 Donaciones,
    🔒 Admin), «Activo …»; «Roles» can grant or remove Runner, Constructor, Supply, Cocina and Team manager (Agente,
    Donaciones and Admin show disabled with «(solo administración)»; at least one role must stay: «Deja al menos un
    rol.»; losing a role closes that person's panel sessions) and «Quitar del equipo» (remove from team). They can't
    change themselves. Removing a team manager from the team also removes that role. Someone without the role sees
    «Esta página es para el team manager de un equipo…».
- **Opening an invite** (`/signup.html?invite=…`) requires a code session. If the person is already signed in they
  join the team at once; otherwise they sign in or create an account and then join («👥 Te invitaron a un equipo. Entra
  o crea tu cuenta para unirte.», then «✓ Te uniste al equipo …»; already a member: «✓ Ya estás en el equipo X.»).
  Accepting moves the person out of any previous team (their roles, including Team manager, stay as they were). A
  used, expired or revoked link doesn't work. Only a hash of the link token is stored.

---

## 13. Admin (`admin.html`)

- **Points**: cards with counters, filters, «Agregado por» (added by), map on the right, point editor with its needs
  (list above the form) and services. Cards and needs show the gold medal icon (tap: «Verificado por el
  equipo X · who · date», for team members only);
  editing a repair need shows its materials. As in Checkup, choosing a different runner for a need that already has
  one adds a run for that runner, and adding items to a runner's run saves them as a new open need (§6). Admin with
  the master password is not an agent: its edits apply directly, without change requests. The point page ends with
  **«🕓 Historial de cambios»** (§6).
- **Usuarios** («🔒 Administración ▾» menu) has four tabs, **«👤 Usuarios»**, **«👥 Equipos»**, **«📡 Registros en vivo»** and
  **«🔗 Invitaciones»** (the tab is kept in the link: `&tab=equipos` / `registros` / `invitaciones`; the live tab shows
  how many are waiting, Invitaciones how many are pending). Registros, Equipos and the searches show each person's
  **access code** («🔑 LUCIA777», loaded one by one like the 🔑 button) instead of the user ID, and searches match it.
  **Choosing a team**: with fewer than 9 teams, one button per team with its **initials in a colored circle** («Or»
  Oro, «UA» Unidos por Azuero), one tap; with more, a dropdown. The page refreshes by itself **every 10 seconds** («● En vivo · hace N s»), except
  while you are typing or editing a user; the browser tab title shows how many new people are waiting, e.g. «(2)».
  - **«📡 Registros en vivo»**: a plain list of everyone who signed up in the last 3 days (e.g. with the sign-up link),
    newest first, with its own filter (name, phone, code or team). Each person shows how long ago, phone, **access
    code** (confirm it with the person) and their team, or **«Sin equipo»** in orange; someone who appears while you
    watch is marked «NUEVO». The tab badge counts those without a team. **Tap a person** to open the **«Asignar a un
    equipo»** window: **1 Equipo** (team buttons; it remembers the last one you used), **2 Roles** (🏃 Runner,
    🔨 Constructor, 📦 Supply, 🍳 Cocina, 👥 Team manager, 🧑‍💻 Agente, 🎁 Donaciones, 🔒 Admin), **3 «✓ Guardar»**. In step 1, **«⊘ Sin equipo»** takes the person out of their team (asks first; it also removes
    Team manager, which needs a team). From there, «Abrir ficha completa»
    opens the user, and «🔗 Mandarle una invitación en vez de asignar» opens Invitaciones with that person and team
    ready. The list doesn't refresh while the window is open.
  - **«🔗 Invitaciones»**: at the top, **«Nueva invitación a un equipo»** in three steps: **1 Persona** (by name, phone or
    code), **2 Equipo**, **3 «Crear enlace»**: a **one-use link valid 24 hours that only that user ID can use** (the person opens it signed in
    with their own code and joins the team; anyone else gets «Esta invitación es para otra persona» and the link is
    not spent). «Copiar enlace» or «WhatsApp» (to their number). Team managers' own invites from «Mi equipo» stay open
    to anyone. Below, **every team invite** (admin's and Team managers') with filters ⏳ Pendientes (default), ✓ Usadas,
    ⌛ Vencidas, ✕ Anuladas, Todas and a search: team (initials), for whom (or «cualquiera»), who created it and when,
    «vence en 5 h» or who used it and when. **«Anular»** stops a pending link (asks first). It refreshes with the page
    every 10 seconds. A link can only be copied right after creating it (it is not stored).
  - **«📊 En vivo»** box (right column on wide screens, at the top on phones; on every Usuarios tab, refreshed with the
    list): **En línea** (of the active users), **Usuarios** (and how many have a code), **Equipos** (and people with a
    code but no team), **En ruta** (users with a run on the way, and how many runs), **Esperando conductor** (open
    supply needs with no runner yet, orange when there are any), **GPS activo** (how many people are sharing their live
    location right now; the tile only counts, to see where someone has been open the user: «📍 Recorrido GPS»), and **Personas por equipo** (each team's initials, name, how
    many are online and how many members).
  - The full list: search, «Solo con código», «Ver desactivados». Click headers to sort (click again to reverse; ▲▼
    shows direction): Usuario, Roles, Actividad, Creado (default: newest first). «Roles» filters with checkboxes
    (several at once; «Sin acceso» = no roles); «Equipo» filters by team. Empty fields are left blank. Bulk select to
    move to a team, deactivate or activate: a checkbox per user, and a **checkbox in the column header** selects every
    result («Todos» next to it on phones). Selected users move to **«☑ Seleccionados (N)»** above the results: open
    with one, **collapsed with 2 or more** (tap its title or the **▾** on the right to open or fold it; that choice is
    kept until the selection is empty). The checked box at the left of its title clears the whole selection, even
    while it's folded. View/copy each user's code (🔑), email, roles, team. The batch bar shows «N seleccionados»,
    «👥 Mover a equipo ▾» (search «Buscar o crear equipo…» with each team's member count, and «＋ Crear equipo «X» y
    mover aquí»; only users with a code move, the rest stay selected), «Quitar del equipo», «Desactivar» (asks first),
    «Activar» and «Limpiar selección». The list search matches name, phone, email or ID. A user with no roles shows a
    «Pendiente» chip (give them a role to activate them); rows show «N en camino» when the user has runs on the way. The
    phone copies on a computer and opens WhatsApp on a phone.
  - **A user** («+ Nuevo», or «Abrir» on a row; links `?view=usuarios&user=<id>`, `&edit=1`, `&user=new`): the view has
    «Editar» and «Activar/Desactivar», the user's runs on the way (tap one to open the point) and **«📍 Recorrido GPS»**:
    their saved path (period «Últimas 24 horas», «Hoy», «Ayer», the last 7 days by name, «Últimos 7 días», «Últimos 30
    días»; start green, end red; tap a dot for its time; ▶ replays it with a 🚚 at «10×», «1 min/s», «5 min/s», «15
    min/s», «1 h/s», with «Saltar pausas»; gaps over 10 min are dotted). A position is saved after 10 m of movement or
    2 min; history is kept 30 days; admin only. The form: «Nombre y apellido», «Teléfono», «Correo (para recuperar el
    código)», «Accesos» checkboxes («👥 Team manager (necesita equipo)», «🔒 Admin (solo lo da administración)»),
    «Código de acceso» with «Generar» (at least 4 characters; case and spaces ignored) and «Equipo» («Sin equipo»; only
    users with a code can be on a team).
  - **«👥 Equipos»**: create a team at the top; a **search** («🔍 Buscar equipo o persona (nombre, teléfono o código)…»)
    finds teams by name, or by a person (name, phone, code or ID),
    showing only their teams with just those people («1 de 5 equipos»); one **card per team** with its **★ Team manager** (red when the team
    has none yet), then its members with their roles and access code («🔑 …», or «sin código»). **«☆ Hacer team manager»** gives that role in one tap
    (that person then manages the team from «Mi equipo»: invites, roles); **«Quitar»** takes someone out of the team
    (and out of Team manager), after confirming. **«＋ Agregar miembro»** searches people with a code by name, phone,
    code or ID and adds them (asks before moving someone from another team). «Editar nombre» renames. (To move many people at
    once, select them in «👤 Usuarios» and use «👥 Mover a equipo».)
  - Role names in admin always carry their page icon: 🏃 Runner, 🔨 Constructor, 🧑‍💻 Agente, 🎁 Donaciones, 📦 Supply,
    🍳 Cocina, 👥 Team manager, 🔒 Admin.
- **«🔒 Administración ▾» menu**: «👥 Usuarios», «🎁 Donaciones» │ «📥 Importar reportes (por revisar)…» │ «📰 Generar
  boletín», «🔗 Ver último boletín ↗» │ «✅ Requerir revisión de puntos: Sí/No», «🔑 Cambiar contraseña maestra» │
  «💾 Descargar respaldo», «📂 Importar respaldo…».
  - **Requerir revisión de puntos**: «Sí» sends points added by the public to «Puntos por revisar»; «No» publishes them
    immediately, like a code user's, and their needs open for runners (turning it off asks first).
  - **Boletín**: every newsletter action asks for a separate newsletter access code («Código de acceso»: «Los boletines
    piden el código de acceso cada vez»).
  - **Cambiar contraseña maestra**: 12–72 characters, typed twice; saving closes every admin session.
  - **Importar respaldo…** appears only when the server runs with `ALLOW_IMPORT=1` (test servers); it replaces all data
    except the master password.

---

## 14. Public summary (home)

«📊 Resumen» is the home sidebar (on phones, the «Resumen» view, which is where the home page opens unless the link
points to a specific point). On phones the views are switched with «Lista / Mapa / Resumen» (icon and name; «Lista» only
for Runners). Signed out, the header has «Reportar punto», **«Inicia sesión»** and ☰ (on phones too); both «Inicia
sesión» and «¡Empieza!» below open the sign-in window (code or new account). «Inicia sesión» hides once signed in. The summary has no title or timestamp: it starts with the card (signed
out), then the note «Cifras reunidas por voluntarios: úsalas como guía, no como datos oficiales.» and the figures.

- **«🙌 ¿Quieres ayudar?»** (only when not signed in): a card at the top inviting visitors to join: «Crea tu cuenta
  con tu nombre y teléfono. Te sumamos a un equipo y empiezas a ayudar: identifica y reporta necesidades, entrega
  donaciones o ayuda a actualizar datos por teléfono.» One button, **«¡Empieza!»**, opens «Login», where they enter
  their access code or create an account. New accounts appear to admin in «📡 Registros en vivo» to be given a team
  and roles (§13). It can't be closed; it disappears once signed in.

- **Reach**: people (children, babies, elderly), points (how many mapped), zones.
- **Estimated consumption\***: liters of water per day (babies 0.75 L, children 1.5 L, adults 3 L, elderly 2.5 L, for
  the whole registered population; also shown as 5-gal botellones) and «Bocas que alimentar» (people at points asking
  for food, × 3 meals a day).
- **Needs now**: urgent, to deliver (new in 24 h), on the way, completed.
- **What's being asked for** (max 8 rows) and **Zones with needs**.

**Header**: signed out on a computer it also shows «Registro ciudadano · corte descargado 12 oct 2026»; signed in, the
logo gets a gold ring. **Map tools**: «Mostrar mi ubicación» (blue pin with accuracy, only on this device); for code
users, «📍 Compartir mi ubicación» / «📍 Compartiendo · Detener» (needs the location notice accepted first; sent every
15 s or 50 m; only your team sees it) and your team's runners as orange trucks («🚚 name · hace N min», refreshed every
15 s). Legend: Punto (green), Necesidad (blue), Urgente (red). Data refreshes every 30 s.

**Search** («Buscar…»): filters by code, zone, place, need or reporter. A Google Maps link or «lat, lon» drops a
temporary purple pin «Ubicación buscada» (only in the browser) with «+ Registrar punto aquí» (opens the form with that
GPS) and an open-in-Maps link; coordinates outside Panama, or a short goo.gl link without coordinates, get a message.
Access codes are no longer typed here (the old hidden «!código» + Enter entry is gone): sign in with «Inicia sesión».

**Point detail** (tap a point; the card can be dragged by its top bar): status badges, «Urgente», «Última revisión»,
«+ Reportar necesidad», «🍲 Pedir comida preparada» (only for people on a team, any role: people, «🍳 Desayuno / 🍛
Almuerzo / 🍲 Cena», «Nota», «Pedir»; §6), «Personas», «Necesidades», notes, «Registrado», «Reportado por»,
«Familias», «Generador». Footer: «🗺️ Google Maps» or «🗺️ Apple Maps» (only with GPS), «🔗 Compartir» (the phone's
share menu, or copies the link on a computer; the link is `/?punto=<id>`; an unpublished point says «Ese punto no está
publicado o ya no existe.») and «➕ Reportar».

**«Reportar necesidad»** window: «¿Qué necesitan?» checklist, «Urgencia» (Normal / Urgente), «Tu nombre» and «Tu
teléfono» (required, remembered on the device), «Enviar necesidad». The need opens immediately with no review
(«Necesidad reportada. Ya está abierta para los corredores.»), and the name and phone show publicly with it.

**Phone list (Runner role)**: a filter line «Con necesidades (N)» opens counts «Puntos / En mapa / Sin GPS», Necesidad
chips «Todas, Urgentes, Normales, Abiertas, En camino, Atendidos, Completadas», when signed in «Reportado por»
«Elegir…» (search, «Sin quién reporta») and «Solo nuestro equipo» (points added by your team or with a need assigned to
your team), «Orden» («Atención» by default, «Más recientes», «Más antiguos», «Área», «Reportado por») and «Listo».
Cards show zone, badges, «🔒 Solo tu equipo», people 👥, the crossed pin, the place, each need line («Necesita / En
camino / Despachada / Completada» + how long ago) and «<name> sabe más».

**Who can do what on the home page**

- **Runner role (signed in with a code)**: the full home page. In a point's detail, «Necesidades por atender» shows the
  team's open/en-camino needs with **«🚚 Asignar conductor / Agregar otro conductor»** (a second runner gets their own run, §6), «Quitar conductor», **«📦 Agregar a
  lista de empaque»** and **«📷 Foto de partida»** (team rules still apply: never another team's run).
- **Other code users (agent, Supply, Constructor, Donaciones, Team manager without Runner)**: on a computer they see the
  same needs list, read-only (no buttons). The server refuses those actions with «Solo quien tiene el rol Runner puede
  asignar conductor, agregar empaque o fotos de partida.» (403).
- **On phones, without the Runner role (including the public)**: only «📊 Resumen» and the map. There is no «Lista»
  view, and tapping a point on the map opens nothing. A shared link to a specific point (`/?punto=…`) still opens it,
  without the needs list. The view updates by itself when a role is added or removed (checked every 45 s).

**«🧾 Servicios recibidos»** (on the public point page): services given at the point (prepared meals, fumigation,
medical round…) with how much, the date and time and the next date if any; no names (§6).

**«🕓 Actividad del punto»** (on the public point page `/camp.html`): everything that happened at the point since it
was reported, newest first, each with date and time, who, and what changed (before → after). It comes from the change
history (§6) but shows **only public data**: the point's code, zone, place, needs, head counts, GPS/location
(«actualizada», without numbers), public notes and directions, approval and «Checkup registrado»; for needs, the
request, state, urgency, «Despachada» and the runner only as «asignado / sin asignar». Never contacts, phones, the
people list, private notes, runner or supply names, packing lists, photos or repairs; another team's reserved run shows
as «Necesidad de un equipo».
- **Who, for the public**: only the role, never a name or phone: «Un runner», «Un agente», «Un acopio», «Un
  constructor», «Un voluntario» (a code user on the home page, Cocina or other pages), «Alguien del público», «Administración», «Sistema», plus
  «de un equipo» when that person is on a team.
- **Who, when signed in with a code**: the full name of people on **your own team** (in green) and of
  yourself; everyone else as for the public. (`/api/full/camps/{id}/log`, falling back to `/api/camps/{id}/log`.)
- The first 15 entries show; «Ver toda la actividad (N)» shows the rest.
- History is recorded from 2026-10-15. For older points and needs, the log is rebuilt from their saved dates
  (reported, approved, checkups, en camino, despachada, completada), with who only when known.

**«🧭 Scout»** (`/scout.html`): for anyone signed in with a code, no new role; it's in the top bar after «Inicio». It
lists **every published point** (and the points you or your team added that are still «Por revisar», or held for your
team), newest first, so volunteers in the field can keep needs current. Laid out like Supply, with no title: on a
computer the map is on the left and the list on the right; on phones «Lista / Mapa» switch between them.
- **Buscar** is in the header after the page links (on phones, the 🔍 button): code, zone, place, directions, needs,
  «Reportado por», who added it. **«+ Agregar punto»** is at the top right and opens the home page's point form
  (`/?agregar=1`).
- A **dropdown** at the top of the list: «Todos los puntos» (default), «Creados por mí», «Creados por mi equipo»
  (remembered on this device), with the number of points shown.
- **Cards like Runners**: zone and code, «Por revisar» / «🔒 Solo tu equipo hasta …», «Urgente», and at the top right
  📍 (show on the map; the crossed pin when it has no GPS) and the Maps button (Google Maps, Apple Maps on iPhone/iPad;
  greyed out with no GPS). Then the place and directions with «📣 *Reportado por*», people with the breakdown, and
  **«Necesita»**: each active need with its chips, «Necesita» or «🏃 En camino · *runner*», «Urgente» and how long ago
  («Sin necesidades abiertas» when there are none). At the bottom «Lo agregaste tú» / «Agregado por *name* (tu
  equipo)», when it was reported and the last checkup, **«📝 Actualizar necesidades»**, **«Ver más»** (the point in view
  mode, below) and «Ver ficha y actividad» (the public point page). Points «Por revisar» show the needs from the report and, instead of the need actions,
  **«🔍 Corroborar»** (see below).
- **«Ver más»** (published points, including ones «🔒 Solo tu equipo»): opens the same view-mode screen as «Corroborar»,
  titled «📋 Ver punto», with the point's data, «Última revisión», and «Necesita» (its active needs, as on the card)
  instead of the reported needs, plus «📝 Actualizar necesidades» (team members) and «Ver ficha y actividad». For
  **agents** it also has **«✏️ Editar»**, which opens the point in edit mode in Checkup
  (`/command.html?punto=<id>&editar=1`, «Editar datos» with the point form), with Checkup's usual rules: an agent of
  another team, or one with no team, sends a change request instead of overwriting. Opened from Scout, Checkup's
  «Cancelar», «Guardar» and the phone's «← Scout» (instead of «← Puntos») go **back to the Scout point screen**; after
  saving, Scout shows the result there («Punto actualizado.» or the change-request message). Once you open other points
  inside Checkup, its back buttons work as usual. People without the agent role don't
  see it; Scout itself never edits point data.
- «Ver ficha y actividad» opens the public point page in the same tab; its «←» links (top and bottom) then say
  «← Volver a Scout» and go back to Scout with the browser history, keeping the list where it was.
- **«🔍 Corroborar»** (points «Por revisar» only, i.e. not yet approved, so only yours or your team's): opens the point
  **in view mode**, like opening a point in Checkup. It is a **full Scout screen, not a pop-up**, with its own address
  (`/scout.html?punto=<id>`) and a «‹ Volver» bar at the top: «‹ Volver», the phone's back button and the browser's back
  and forward all move between the list and the point, a reload keeps the point open, and going back returns the list
  where it was. «📍 Ver en el mapa» goes back and centres the map on it. A point that is no longer in Scout shows «Punto
  no disponible». It is there so someone in the field can check the report against what they
  see. A note says it isn't public yet and is published when approved in Checkup or admin. It shows Código (unless
  automatic), Área, Ubicación, Dirección, Reportado por, Personas and the breakdown, Servicio, Notas del grupo, GPS,
  when it was reported and by whom (you / your teammate), and «Necesidades reportadas». Buttons: the Maps app («Abrir
  en Google Maps» / «Abrir en Apple Maps»), or with no GPS «Buscar en Google Maps» by place and zone, labeled
  «Búsqueda aproximada: no hay GPS registrado.»; and «📍 Ver en el mapa» when it has GPS. Nothing can be edited or
  approved there; contacts show only as on the cards (below). **Agents** (Checkup role) also get **«✏️ Editar»**, as in «Ver
  más» below.
- **«Contactos»** (only for people **on a team**): the card lists the point's contacts, names and phones only (no
  cédula or age): «📣 *name* reporta» (who reported it, when they left a phone) and «👤 *name*» for each person of the
  group; three on the card and the rest under «y N contactos más»; «Ver más» / «Corroborar» show them all. **Tapping a
  number** opens a small box with the name, the number, **«📞 Llamar»** and **«💬 WhatsApp»** (as «Reportado por» in
  Runners). The server only sends contacts to people on a team: anyone can create an account and open Scout, so
  without a team no names or phones reach the browser.
- **«📝 Actualizar necesidades»** is only for people **on a team**; they can update **any** point's needs, directly (no
  change request, even if another team updated the need last). Without a team the button isn't shown and a note at the
  top says «Para actualizar necesidades tienes que ser parte de un equipo. Pide a un team manager que te invite.» The
  window lists the point's active needs:
  - **«Cambiar»** (the same need checkboxes as the point form, plus «Urgencia» Normal/Urgente; «Guardar cambios»).
    Needs on the way («🏃 En camino») can be changed too («Va en camino: se puede cambiar, no quitar.»); what's added to
    a run goes in a separate open need, as in Checkup. When the runner is on another team the window only offers the
    urgency («La lleva un runner de otro equipo: solo puedes cambiar la urgencia.»).
  - **«Ya no lo necesitan»** (open needs only; asks first): it's removed, «Quitada (ya no necesita nada)» in the history.
  - **«＋ Reportar otra necesidad»**: checkboxes + urgency, «Reportar necesidad»; it's open for runners right away
    («Necesidad reportada: ya la ven los runners.»). «Reparación de casa» goes to Constructores, as from any form.
  - Repair, prepared-meal and service needs aren't changed here.
  Every change is recorded in the change history with the page **«Scout»** and the person's name, and counts as a
  checkup of the point.
- **Map**: published points with GPS: red urgent, orange with open needs, green without; the popup has «Ver en la
  lista», which scrolls to the card and flashes it. Points without GPS stay in the list only.
- Signed out, it asks to sign in from the home page. Signed in **without a team**, no contacts or phones are shown
  (only the public «Reportado por» name and the assigned runner's name); **with a team**, «Contactos» shows names and
  phones (see above).

---

## 15. Privacy

- **Public**: approved points (no phones, contacts, private notes or group members), needs (no runner, supply
  center, packing list or photos; «Despachada» yes, but not by whom), Resumen totals. Public point fields include the
  «Reportado por» name (never its phone), «Notas de grupo», «Notas de servicio», «Otras notas» and «Indicaciones»; only
  the private camp note is hidden. Exception: the name and phone of whoever reported a need from the home page
  («Reportar necesidad») are shown publicly with that need.
- **Private**: contacts and phones, notes, ID numbers, who carries each run, supply centers, packing lists, photos,
  donors, live locations (**own team only**, everywhere; admin can also see each user's saved route in Usuarios «📍
  Recorrido GPS», kept 30 days), repairs.
- **Scout contacts**: a point's contact names and phones (not cédula or age) are shown in Scout only to signed-in people
  **on a team**; signed-in people without a team (anyone can create an account) don't receive them.
- **«Verificado por equipo»**: the public and people without a team see only the medal icon; which team, who and
  when reach only people on a team (Runners, Supply) and Checkup/admin, who tap the icon to see them.
- Repair needs and their materials are never public.
- A point held with «Nuestro equipo lo abastece» is hidden from the public and from other teams for 3 hours.
- Anyone signed in with a code sees, on the home page, the runner's name, packing list and «Foto de partida» photos of
  every need except another team's reserved runs.
- The point activity log on `/camp.html` is public but only with public fields and no names or phones; teammates see
  each other's full names (§14).
- Who updated a need or point last (`updatedBy`) and change requests are private: only the people involved see a
  request. The change history is only in admin and Checkup.
- Who added a need by changing someone's run is private: admin and Checkup show «Agregada por …» on that need; the
  public, Runners and Supply don't.
- When several runners carry the same need, the public sees each run as its own «En camino» need, without names.

---

## 16. Technical summary

- Go server (`main.go` plus per-topic files) with a bbolt database on a persistent volume. HTML pages with no build
  step (`nav.js` shared for the bar, ☰ menu, sign-in and live access; `pointform.js` for forms, packing lists and
  photos).
- APIs: `/api/…` public, `/api/session/…` (code sign-in and sign-up), `/api/admin/…` (master password),
  `/api/runner/…`, `/api/builder/…`, `/api/agent/…`, `/api/supplier/…`, `/api/cook/…`, `/api/desk/…` (donations),
  `/api/full/…` (code session: full view, Mi equipo, invites). Each has its own path-scoped, SameSite=Strict cookie
  (`relief_session`, `runner_session`, `builder_session`, `agent_session`, `supplier_session`, `desk_session`,
  `cook_session`, `viewer_session`); code sessions last 12 h and live in server memory. Writes carry a CSRF token,
  except `/api/full/…`, which relies on the SameSite=Strict cookie. `/api/agent/<x>` is a proxy to `/api/admin/<x>`
  restricted to the `agentAllowed` list (`agent.go`); users, teams (except listing), password, backups, settings,
  newsletters and GPS tracks stay admin-only.
- Other files: `settings.go` (meta `require_reviews`, `GET/POST /api/admin/settings`), `track.go` (bucket
  `location_history`, `GET /api/admin/users/{id}/track`, admin only), `teams.go` / `teaminvites.go` (buckets `teams`,
  `team_invites`, which store only a SHA-256 of the token), `donations.go`, `backup.go` (`GET /api/admin/backup`,
  `POST /api/admin/restore` with `ALLOW_IMPORT=1`).
- Change requests (`changes.go`, bucket `change_requests`): `/api/runner/…`, `/api/builder/…` and
  `/api/agent/change-requests` (list; `POST …/{id}/accept|reject`). An agent's `PATCH /api/agent/needs/{id}` or
  `PATCH /api/agent/camps/{id}` answers 202 with the request when it needs the owner's OK.
- Prepared meals and services (`meals.go`): needs of kind `meal` (the request) and `mealrun` (one runner's delivery,
  `sameAs` = request); `/api/full/camps/{id}/meals` (team members), `/api/cook/…` (Cocina session `cook_session`),
  `/api/runner/meals` (+ `/accept`, `/decline`), `GET /api/camps/{id}/services` (public) and
  `POST /api/admin/camps/{id}/services` (admin and agents). Each service need keeps `serviceLog` entries.
- Point activity log (`camplog.go`): `GET /api/camps/{id}/log` (public), `GET /api/full/camps/{id}/log`. Scout
  (`scout.go`, code session): `GET /api/full/scout` (published points + mine/team's pending, with `activeNeeds`, and `contacts` — names and
  phones only — when the viewer is on a team),
  `POST /api/full/scout/camps/{id}/needs`, `PATCH /api/full/scout/needs/{id}` (need, urgency, or status `DELETED`
  for open needs; open or on-the-way needs without a kind). Writes need a team; no change requests. New pages must be added to the static allowlist in `main.go` (which also
  serves `report/boletin_general.html` and `report/logo.png`) and to the `Dockerfile` COPY line (`scout.html` and
  `cook.html` are).
- Change history (`audit.go`, bucket `audit`, keyed by point and time): every need and point write goes through
  `putNeed` / `putPoint`, which compare with what was stored; `a.write(r, …)` records who made the request.
  `GET /api/admin/camps/{id}/trail` (also for agents) and `GET /api/admin/centers/trail`.
- Several runners on one need (`runs.go`): each runner's run is its own record linked by `sameAs`; the server never
  puts a different runner on a record that already has one.
- Tests: `go test ./...`, `go vet ./...`, `go build ./...`, `node --test feature_ui_test.mjs`,
  `python3 -m unittest ci_index_test.py`. CI («CI and deploy») runs test/build/node/python plus `helm lint` /
  `helm template`, then publishes the image and deploys.
- **Development marker**: when the site runs on localhost, a local-network address (10.x, 192.168.x, 172.16–31.x) or a
  `*.local` / `*.test` name, every page shows a thin amber-and-navy stripe at the top, a small «DEV» tag in the
  top-left corner (tooltip: the host) and «DEV ·» before the tab title. It never shows on the public domain.
  (`nav.js`, loaded by every page including `admin.html`.)

---

## Keeping this manual current

Every new feature or behavior change is recorded here in the same task: what it does, who uses it, how it relates to
the rest (states, roles, privacy), plus a dated line in «What's new». If something stops existing, remove it here.
