# Azuero Apoyo — Manual

How the app works, who uses each page, and how points, needs, runs, supply centers and teams relate.
The UI is in Spanish; on-screen labels are quoted here exactly as they appear (e.g. «¡Yo lo llevo!»).
This manual is updated with every new feature or behavior change (see "Keeping this manual current" at the end).

---

## What's new

**2026-10-15**
- Home: signed-out visitors see «🙌 ¿Quieres ayudar?» above the Resumen, with «Crear mi cuenta» and «Ya tengo código» (§14).
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
- **Admin and agents**: review reports, maintain the data, grant access.

Everything the community reports is *reported* data, not verified, and the app presents it that way.

---

## 2. Pages

| Page | Who | What it does |
| --- | --- | --- |
| `/` (home) | Public and everyone | Map and list of points, «📊 Resumen» figures (the first view on phones), «¡Ayuda!» to report a point, «Reportar necesidad» on a selected point, «Login». |
| `/signup.html` | Anyone (link to share) | Sign in with a code or create an account. When signed in, shows which pages the person can open. Also accepts team invites (`?invite=…`). |
| `/camp.html?id=…` | Public | Public page of an approved point (no private data): need chips with their options, «Verificado por un equipo» badge, link to Google/Apple Maps. |
| `/runner.html` («Runners») | Runner role (and agents, for packing lists only) | Take runs, route, packing list, photos, complete. |
| `/builder.html` («Constructores») | Builder role | Repair jobs: assess, take, progress, materials, photos. |
| `/supplier.html` («Supply») | Supply role | Collection center: request runners, assign in person, answer runner requests, pack, dispatch. |
| `/command.html` («Command») | Agent role | Call center: points, needs, checkups, report needs. |
| `/donations.html` («Donaciones») | Donations role and admin | Donation tickets: offered → confirmed → at center → on the way → delivered. |
| `/team.html` («Mi equipo») | Team manager role | Invite people to the team, see members, change their roles, remove them. Lives in the ☰ menu. |
| `/admin.html` | Admin (master password) | Everything: review reports, points, needs, users, teams, backups, newsletter. |

On phones the top bar links to the pages the person can use; rarely used ones (Mi equipo) are in the menu. If the links
don't fit, they shrink to icons and then collapse into one «current page ▾» button that opens the list as a dropdown,
so the page action (＋ «Agregar punto»), 🔍 and the menu button always stay visible on the right. When signed in, the
menu button shows the person's **initials** in a gold circle (e.g. «MS»; tooltip with the full name) instead of ☰. The
menu shows the signed-in name and **roles**, «Copiar mi ID», «Salir» and text size (90 %–150 %).

---

## 3. Roles and access

| Role (internal name) | Label | Opens | Notes |
| --- | --- | --- | --- |
| `runner` | Runner | Runners | Takes and delivers runs. |
| `builder` | Constructor | Constructores | Repair jobs. |
| `agent` | Agente | Command (and Runners for packing lists only) | Sees private point data. Admin-only to grant. Changes to a need someone outside their team last updated go to that person as a request (§6). |
| `donations` | Donaciones | Donaciones | Sees donor contacts. Admin-only to grant. |
| `supplier` | Supply | Supply | A collection center. |
| `teammanager` | Team manager | Mi equipo | **Must belong to a team.** |
| — | Admin | admin.html | Master password; not a user role. |

- A person can hold several roles.
- Admin grants roles in **Usuarios** (checkboxes). A team manager can grant or remove Runner, Builder, Supply and Team
  manager **within their own team** (see §12).
- **Live access**: every signed-in page re-checks the current roles every 45 s (and when the tab regains focus). A new
  role shows up in the menu; a removed role ends its sessions immediately and sends the person home with the notice
  «Tu acceso cambió…».
- Sessions live in server memory: after a server restart, anyone with a remembered code signs back in automatically.

---

## 4. Signing in and creating an account

Every sign-in (home «Login», Runners, Constructores, Command, Supply, `signup.html`) shows two groups side by side,
with no hidden fields:

- **«Ya tengo código»** (I have a code): access code + «Recordar mi código de acceso» + «Entrar».
- **«Crear cuenta nueva»** (new account): full name, phone, email (optional) + «Crear mi código».

Details:

- **Access code**: generated by the server when the account is created (word + digits + part of the name, e.g.
  `foca72nora6`). It may be typed with or without a leading «!»; case and spaces don't matter. Codes admin assigned
  earlier keep working.
- **Approval**: new accounts are approved immediately as Runner (as Builder when created from Constructores) and are
  signed in right away. Admin can change the role or deactivate them.
- **Remember**: after a successful sign-in the code is saved on the phone; next time it is prefilled and «Entrar» is
  enough. Unchecking «Recordar» forgets it. «Salir» ends the session but keeps the remembered code.
- **Duplicate names**: if someone already has that full name, the new account becomes e.g. «Juan Pérez 2».
- **A phone can be used again** when creating an account: it may be the same person who lost their code. They get a new
  account and a new code (never the old one); the same full name gets a number («Juan Pérez 2»). When putting people on
  a team, ask them for their user ID («Copiar mi ID» in the menu) to know which account is which. People who
  registered before codes existed (name + phone only) sign in with name and phone on Runners/Constructores.
- **Email**: optional, for code recovery: admin looks the person up by phone or email in Usuarios and shows or
  changes their code. No emails are sent.

---

## 5. Points

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
publishes by itself; Command and admin can publish it sooner with «🌐 Publicar ahora».

**States**

- `pending` («Por revisar»): public reports, imports, unreviewed reports. Not public.
- `approved`: published. Approving creates its open needs.
- `archived`: hidden from the public and runners; can be restored.
- `rejected` («Duplicado»): merged into another existing point.
- Signed-in code users can add points that publish immediately (with the gold «Verificado por equipo» badge if they
  are in a team, see §6).

**GPS**

- Without GPS a point doesn't appear on the map, but it stays in the list with the crossed-out pin icon.
- The Maps button is only enabled with a real location (coordinates or a Maps link); otherwise it is disabled next to
  the crossed-out pin (Runner and Supply cards don't fall back to a place-name search).

**Point admin** (admin.html)

- «Por revisar»: Approve, Edit, Reject, «Duplicado…» (merge into an approved point).
- Select reports with checkboxes and **⬇ Exportar CSV** (all fields, including private ones).
- ☰ menu → **Importar reportes (por revisar)…**: an exported CSV or JSON; always imported as «Por revisar»;
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

- **Need checklist**: the same one everywhere a need is entered (home «¡Ayuda!» and «Reportar necesidad», Command,
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
- **«Verificado por equipo»** (gold award badge): a point or need created or updated by someone in a team is marked
  as checked more thoroughly. Stamped by the server when:
  - a code user in a team adds a point (home or panels);
  - an agent in a team creates a point, edits a point, or creates/updates a need in Command;
  - a runner in a team records a new need after completing a run.
  An update by someone without a team removes the mark (the data changed without a team check); edits by admin with
  the master password keep it. Panels with a session show the team, person and date («Equipo Norte»); public pages
  show only «Verificado por un equipo», without names.
- **Recurring services** (garbage, portable toilet, fumigation, medical round): repeat every N days; once completed
  they are scheduled and don't count as open until due.
- **Several runners on the same need**: each runner always has a run of their own (a linked copy, `sameAs`) with its
  own runner, supply center, packing list, photos, dispatch and «Completé», so every shipment is documented and no
  one's run is overwritten. Cards show «👥 También va: …». A second run is created when:
  - a supply center sends «otro runner» to a need that is already on the way;
  - on the home page a runner taps **«🚚 Agregar otro conductor»** on a need someone already carries (picking a runner
    who already carries it says «Ese runner ya lleva esta necesidad.»);
  - admin or Command choose a different runner for a need that already has one (dispatch, the need editor, or «¿quién
    la suplió?» after «Ya no necesita nada»). If that runner already has an open or on-the-way run of this need, the
    change goes to their run (for example it is marked completed) instead of creating another.
  The first runner's run is not changed. To take a runner off, use «Quitar conductor» (home) or «Sin asignar»
  (admin/Command); that run goes back to «Abierta» as before. For an agent outside the team of whoever updated the
  need last, any of these Command changes is first a change request and only happens once that person accepts it
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
- **Change requests («📝 Cambios propuestos»)**: the person who was at the point is the better source of truth, so an
  agent can help verify data but can't override them. Each need remembers **who updated it last** (`updatedBy`,
  private): the runner who took, completed or dropped it, the runner who recorded it after a run, a code user who
  assigned a runner from the home page, or the agent who edited it in Command. Older needs fall back to their runner
  (or builder).
  - When an agent changes a need (after a «Checkup» call or any edit in Command) and the person who updated it last is
    someone else **outside the agent's team** (the agent has no team, or a different one), nothing changes yet: the
    change is saved as a request to that person. Command shows «📝 Cambio enviado a … para que lo acepte o lo rechace».
    One pending request per agent and need; a new one replaces it.
  - Agents on the **same team** as that person edit directly, as before. So do admin with the master password, and
    edits that change nothing.
  - Requests only go to people who can answer them (a Runner or an agent); otherwise the edit applies directly.
  - Adding packing lines or photos, creating a new need, and logging the checkup itself never need a request:
    nothing is overwritten.
  - The person sees it at the top of **Runners** (above the tabs) or **Command** (under the filters): who asks (name and
    team, or «sin equipo»), the point, the need, and each change as «Pedido / Estado / Urgencia / Runner: before →
    after», with **«✓ Aceptar»** or **«✗ Rechazar»**. Accepting applies the change exactly as the agent's edit would
    (including the run rules above, so added items still become a new open need) and the need stays theirs; rejecting
    changes nothing. The list refreshes every 30 s.
  - In Command, **«📤 Cambios que pediste»** lists your requests of the last 7 days: «⏳ Esperando a», «✓ Lo aceptó» or
    «✗ Lo rechazó».
- **Packing list**: what is loaded for that run (item × quantity). Private.
- **Photos**: departure photos (when loading) and delivery photos (when completing). Private.

---

## 7. Runners (`runner.html`)

Tabs: **Yo** (my runs on the way), **Runs** (open needs) and **📦 Pedidos** (supply centers asking for a runner).
On phones, «Lista / Mapa» switch between the list and the map.

- **Buscar** (search): in the header, like the home page; on phones it opens from the 🔍 button. It filters the open
  tab (Yo, Runs, Pedidos) and the map by point code, zone, place, need, coordinator, supply center or runner,
  ignoring accents and case; tab counters follow it. In «Yo», non-matching runs are hidden without changing stop
  numbers or order.

- **«¡Yo lo llevo!»** asks for confirmation (point and what to bring); accepting moves it to «Yo» with the run
  highlighted.
- **Picking from the list** (card or «Ver en el mapa»): the map centers on the point and opens a short popup with
  «¡Yo lo llevo!» and «Ver en la lista».
- **Card**: point name with the 📍 «Ver en el mapa» and Maps buttons at the **top right**; then where it is; then how many
  people with the breakdown and icons («40 personas · 20 adultos · 15 niños · 9 familias», zeros left out) and how
  long ago it was reported; then «Necesita» (what the point asked for) separate from «📦 Lo que llevas» (packing list),
  contacts, and at the bottom share + the action. Tapping the card background does nothing (no highlight); a card you
  jump to from the map flashes briefly. With a supply center: «📦 Recoger en acopio: Villa Marina · Julie» plus phone.
- **Packing list**, **📷 Foto de partida**, **✓ Completé** (the confirmation shows what is being delivered and allows
  delivery photos) and «No puedo continuar» (the need goes back to open; its packing list and photos stay on it, each
  with who added it, and the next runner who takes it sees them).
- **«📝 Cambios propuestos»** (above the tabs): changes an agent from another team (or with no team) asks to make to
  a need you updated last; «✓ Aceptar» or «✗ Rechazar» (§6).
- **Only open needs can be taken** in «Runs»: a need another runner already carries can't be taken here. A second
  runner on the same need is added from the home page («🚚 Agregar otro conductor»), admin, Command or Supply, and gets
  a run of their own (§6).
- **After completing**: «📝 Actualizar necesidad» (opens the same need checklist to record what is still missing,
  including a house repair, which goes to Builders) or «✓ No necesitan nada más».
- **Gold badge**: a card shows the gold «Equipo …» badge next to the point name when a team verified the point or that need.
- **«📦 Pedidos»**: open needs that one or more supply centers offered to supply. The card lists every offering center
  (center, person, phone) with its own **«¡Yo lo llevo! · recojo aquí»**: pick the closest; the run is linked to that
  center and shows «📦 Recoger en acopio: …».
- **«Pedir a acopios»** (ask supply centers, on a run with no center): centers answer whether they can supply it; the
  card lists those that can («✓ 2 acopios pueden surtirla: elige el más cercano») and «Elegir» links the run to the
  chosen center.
- **Live location**: once permission is granted it is shared **only with the runner's own team** (strict rule, on
  every page: other teams, the public, people without a team and other supply centers never see it). The map follows the runner until they pick a run or move the map; «🎯 Ir a mi ubicación» resumes
  following. Losing signal doesn't stop it; only denying permission does.
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
point or phone) is in the header like other pages (🔍 on phones).

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

**Card in three parts** (runs of this center)

1. **Runner**: name, phone, «La tomó hace…» (taken … ago), live («📍 Ver runner en el mapa») if sharing location and on
   your team, who
   dispatched it, «👥 También va».
2. **Destination**: point, directions, «Necesita», place contacts, show on map / Maps.
3. **To load**: packing list, departure photos, «🚚 Despachada» (with confirmation) or «Salió del acopio» with
   «Deshacer» (undo).

**Each supply center only sees and dispatches the runs linked to it.** A run becomes linked to a center when:

- the center offered it («¡Yo abastezco!») and a runner chose this center in «Pedidos»;
- the center assigned it to a runner who was there («Asignar a runner»);
- a runner asked for a center and chose this one;
- the center sent «otro runner» (a new linked run).

**Solo mi equipo** (my team only): «Por surtir» shows only points my team added; the other tabs show my team's
points or runs by my team's runners; the runner picker shows only my team's runners.

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
   Supply «🙌 Por surtir» → «🚚 Ya en camino» ─▶ «🙋 Abastezco otro envío» / «🏃 Asignar otro runner» ─▶ new run
```

If a runner drops a run that came from a center's offer, it goes back to «Pedidos» with that center still offering. If they drop one assigned in person, it
becomes a normal open need. Dispatching doesn't change the state: the runner completes it.

---

## 9. Builders (`builder.html`)

Tabs «🔨 Mis obras» (my jobs) and «🚩 Obras» (jobs). Each repair job has progress 0–100, «Qué falta para terminar»
(what's left), required materials, Before / Progress / Final photos and «✓ Obra terminada» (needs a final photo).
**Assess** a job without taking it: work items (roof, walls, floor…), materials, days, people, notes and photos;
assessments are visible to all builders. Repairs are never public; their materials are totaled in the Resumen.
Repair needs can come from any form (public, runners, Command, admin); when the reporter picked materials, the card
shows them under the title as «PIDEN hojas de zinc · varilla …».

---

## 10. Command (`command.html`, agents)

**Buscar** is in the header (on phones, the 🔍 button), like Runners, and «+ Agregar punto» is always at the top right:
it opens the same point form **right in Command** (no trip to the home page; a team agent is the coordinator by
default) and publishes the point on «Guardar», opening it in the list.
The totals are shown with icons: 📍 points (or «4/54» while searching or filtering) and 🕓 «por revisar»; on phones they
are inside the «Puntos» tab. List of points with filters (state, area, need, urgent/open/on the way), a map, and an
"en ruta" monitor. Each card:
point name, people, coordinator, counters «🚩 N por llevar» / «🏃 N en camino», needs, «Checkup» (log a call) and
«Editar punto». Opening a point shows full data, contacts, changelog and its needs (update, report a new one, mark
services done).

**«Checkup»** (after calling someone at the point) asks «¿Lograste comunicarte con alguien en este punto?»: «No» does nothing;
«Sí, están bien» logs the checkup and removes the needs still to be delivered; «Sí, necesitan cosas» logs it and opens
the needs to update. The checkup itself is always saved. Removing or updating a need that someone outside your team
updated last becomes a change request to that person (below), so after «Sí, están bien» those needs stay until they
accept.

In «Actualizar necesidad», **«Ya no necesita nada»** asks **«¿Alguien la suplió?»**:

- «No, ya no la necesitan» removes the need (`DELETED`) and counts as a checkup on the point.
- «Sí, la suplió alguien» asks for the Runner (or Constructor for a repair), required, and saves the need as
  `COMPLETED` under that person with the original items, like a finished run. A new name creates the runner with
  their phone. If another runner already has the need on the way, that run stays as it is and the person chosen gets
  their own completed run, linked to it (§6).
- Both answers are change requests when the need was last updated by someone outside your team.

**Changes to someone else's data**: if the need was last updated by someone outside your team (or you have no team),
any change to it (editor, «Despachar», «Completada», «Ya no necesita nada», «Checkup») is sent to them as a request
instead («📝 Cambio enviado a …»); they accept or reject it on their page, and you follow it in
«📤 Cambios que pediste». Requests sent to you (as the last one who updated a need) show as «📝 Cambios propuestos»
under the filters (§6).

Once a change applies (directly, or when accepted), in the need editor (and «Despachar»), choosing a **different runner** for a need that already has one adds a run
for that runner instead of replacing the first one; «Sin asignar» takes the runner off. Changing what an on-the-way or
completed run asks for (when you are not its runner) keeps the run as it is and saves what you added as a new open
need of the point, marked «Agregada por» you; removing items from a runner's run is refused (§6).

**Filters**: status (Todos / Por revisar / Aprobados) and **«Agregado por»**: Todos, «Míos» (added by me) and «Equipo»
(added by anyone on my team; only for agents on a team), each with its count. On phones the rest (area, need type,
order, «Sin GPS», need chips) fold behind **«Filtros ▾»**, which shows how many are active; they wrap to fit.

**Rows**: a short line «Revisado hace 1 h» (or «Creado / Suplido hace …»; exact date on hover) with a **badge of who
added it** (full name; green with «(tú)» if it was you), the
point name, a line with small icons for place, people and coordinator, then the counters, «Checkup» and «Editar».
Point rows and need cards show the gold «Verificado por equipo» badge with the team name. A point a team is holding
shows «🔒 Solo equipo X · se publica a las …» with **«🌐 Publicar ahora»** (asks first). Editing a repair need shows
its materials, which can be changed. Works on phones (single column).

---

## 11. Donations (`donations.html`)

Donation tickets: **Por confirmar → Confirmada → En acopio → En camino → Entregada** (or Cancelada). Each ticket:
donor (private), items, collection center and arrival date, transport, destination point with its GPS, departure
date, notes and history. At the top: needed vs committed vs unconfirmed per item (unconfirmed doesn't count).
Access: Donations role or admin.

---

## 12. Teams and Team manager

- **Teams**: admin creates them in Usuarios → 👥 Equipos and assigns users (one team per user; code users only).
  Teams drive «Solo mi equipo» (Runners, Supply) and who sees the team's live locations.
- A **Team manager** (must be in a team) opens **Mi equipo** from the ☰ menu:
  - **Invite**: «Crear enlace» (expires in 24 h, 3 days or 7 days; **single use**). Copy, share, or send via
    WhatsApp. Invite list: pending, used by…, expired, revoked («Anular»).
  - **Members**: a quick search above the list filters as you type by name, phone (3+ digits) or role, ignoring
    accents and case (the counter shows e.g. «1 de 3»). Each member shows phone, role chips with the page icons
    (🏃 Runner, 🔨 Constructor, 📦 Supply, 👥 Team manager, 🎧 Agente, 🎁 Donaciones), activity; «Roles» (can grant or remove Runner, Builder, Supply, Team manager; Agent
    and Donations are locked) and «Quitar del equipo» (remove from team). They can't change themselves. Removing a
    team manager from the team also removes that role.
- **Opening an invite** (`/signup.html?invite=…`) requires a code session. If the person is already signed in they
  join the team at once; otherwise they sign in or create an account and then join («✓ Te uniste al equipo …»).
  A used, expired or revoked link doesn't work. Only a hash of the link token is stored.

---

## 13. Admin (`admin.html`)

- **Points**: cards with counters, filters, «Agregado por» (added by), map on the right, point editor with its needs
  (list above the form) and services. Cards and needs show the gold «Verificado por equipo» badge with the team name;
  editing a repair need shows its materials. As in Command, choosing a different runner for a need that already has
  one adds a run for that runner, and adding items to a runner's run saves them as a new open need (§6). Admin with
  the master password is not an agent: its edits apply directly, without change requests.
- **Usuarios** (☰ menu) has four tabs, **«👤 Usuarios»**, **«👥 Equipos»**, **«📡 Registros en vivo»** and
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
    equipo»** window: **1 Equipo** (team buttons; it remembers the last one you used), **2 Roles** (Runner,
    Constructor, Supply, Team manager, Agente, Donaciones), **3 «✓ Guardar»**. In step 1, **«⊘ Sin equipo»** takes the person out of their team (asks first; it also removes
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
  - The full list: search, «Solo con código», «Ver desactivados». Click headers to sort (click again to reverse; ▲▼
    shows direction): Usuario, Roles, Actividad, Creado (default: newest first). «Roles» filters with checkboxes
    (several at once; «Sin acceso» = no roles); «Equipo» filters by team. Empty fields are left blank. Bulk select to
    move to a team, deactivate or activate. View/copy each user's code (🔑), email, roles, team.
  - **«👥 Equipos»**: create a team at the top; a **search** finds teams by name, or by a person (name, phone or ID),
    showing only their teams with just those people («1 de 5 equipos»); one **card per team** with its **★ Team manager** (red when the team
    has none yet), then its members with their roles and ID. **«☆ Hacer team manager»** gives that role in one tap
    (that person then manages the team from «Mi equipo»: invites, roles); **«Quitar»** takes someone out of the team
    (and out of Team manager), after confirming. **«＋ Agregar miembro»** searches people with a code by name, phone or
    ID and adds them (asks before moving someone from another team). «Editar nombre» renames. Below, «Asignar en
    lote» keeps the table to assign many users at once.
  - Role names in admin always carry their page icon: 🏃 Runner, 🔨 Constructor, 🧑‍💻 Agente, 🎁 Donaciones, 📦 Supply,
    👥 Team manager.
- **Teams**, **newsletter** (boletín), **backup** (download; import only on test servers), **master password**.

---

## 14. Public summary (home)

«📊 Resumen» is the home sidebar (on phones, the 📊 view, which is where the home page opens unless the link points to a
specific point):

- **«🙌 ¿Quieres ayudar?»** (only when not signed in): a card at the top inviting visitors to sign up: «Crea tu cuenta
  con tu nombre y teléfono. Te sumamos a un equipo y empiezas a ayudar…». **«Crear mi cuenta»** opens «Login» with the
  cursor in «Crear cuenta nueva»; **«Ya tengo código»** opens it on the code field. New accounts appear to admin in
  «📡 Registros en vivo» to be given a team and roles (§13). «×» hides the card for this visit; it disappears once
  signed in.

- **Reach**: people (children, babies, elderly), points (how many mapped), zones.
- **Estimated consumption\***: liters of water per day (babies 0.75 L, children 1.5 L, adults 3 L, elderly 2.5 L, for
  the whole registered population) and mouths to feed (3 meals a day).
- **Needs now**: urgent, to deliver (new in 24 h), on the way, completed.
- **What's being asked for** (max 8 rows) and **Zones with needs**.

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

---

## 15. Privacy

- **Public**: approved points (no phones, contacts, private notes or group members), needs (no runner, supply
  center, packing list or photos; «Despachada» yes, but not by whom), Resumen totals.
- **Private**: contacts and phones, notes, ID numbers, who carries each run, supply centers, packing lists, photos,
  donors, live locations (**own team only**, everywhere), repairs.
- **«Verificado por equipo»**: public pages show only that a team verified it («Verificado por un equipo»); which team,
  who and when are visible only in panels with a session (Command, admin, Runners, Supply).
- Repair needs and their materials are never public.
- A point held with «Nuestro equipo lo abastece» is hidden from the public and from other teams for 3 hours.
- Code users see more (e.g. the runner's name on their team's needs).
- Who updated a need last (`updatedBy`) and change requests are private: only the people involved see a request.
- Who added a need by changing someone's run is private: admin and Command show «Agregada por …» on that need; the
  public, Runners and Supply don't.
- When several runners carry the same need, the public sees each run as its own «En camino» need, without names.

---

## 16. Technical summary

- Go server (`main.go` plus per-topic files) with a bbolt database on a persistent volume. HTML pages with no build
  step (`nav.js` shared for the bar, ☰ menu, sign-in and live access; `pointform.js` for forms, packing lists and
  photos).
- APIs: `/api/…` public, `/api/admin/…` (master password), `/api/runner/…`, `/api/builder/…`, `/api/agent/…`,
  `/api/supplier/…`, `/api/desk/…` (donations), `/api/full/…` (code session: full view, Mi equipo, invites). Each has
  its own session (path-scoped cookie) and a CSRF token on writes.
- Change requests (`changes.go`, bucket `change_requests`): `/api/runner/change-requests` and
  `/api/agent/change-requests` (list; `POST …/{id}/accept|reject`). An agent's `PATCH /api/agent/needs/{id}` answers
  202 with the request when it needs the owner's OK.
- Several runners on one need (`runs.go`): each runner's run is its own record linked by `sameAs`; the server never
  puts a different runner on a record that already has one.
- Tests: `go test ./...`, `node --test feature_ui_test.mjs`, `python3 -m unittest ci_index_test.py`.
- **Development marker**: when the site runs on localhost, a local-network address (10.x, 192.168.x, 172.16–31.x) or a
  `*.local` / `*.test` name, every page shows a thin amber-and-navy stripe at the top, a small «DEV» tag in the
  top-left corner (tooltip: the host) and «DEV ·» before the tab title. It never shows on the public domain.
  (`nav.js`; `admin.html` has the same copy because it doesn't load `nav.js`.)

---

## Keeping this manual current

Every new feature or behavior change is recorded here in the same task: what it does, who uses it, how it relates to
the rest (states, roles, privacy), plus a dated line in «What's new». If something stops existing, remove it here.
