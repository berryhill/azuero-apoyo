# Azuero Apoyo — operating guide for coding agents

## Why this repo exists

People in Azuero use this site to find relief camps and delivery points after an earthquake, report shortages, and coordinate service needs. A bad location, hidden request, leaked phone number, or locked-out operator can affect real deliveries. The app is already live at https://azuero-apoyo.berryhill.dev/. Favor a working, legible operator path over abstraction or polish. Build only what the user asked for; never add an adjacent feature or a new restriction because it seems prudent.

The UI is Spanish-first. Treat reports as reported data, not verified deliveries or guaranteed availability. Preserve the distinction between where a camp is, what its group says it needs, and an actionable service need assigned to a runner.

## People and their paths

| Person | Actual path | What must remain true |
| --- | --- | --- |
| Resident or field worker | Public map/list → `+ Agregar punto` → `POST /api/submissions` | Sparse reports are accepted and saved for review; success does **not** imply publication. No code, area, GPS, or numeric counts are required. Explain errors in actionable Spanish. |
| Public viewer or coordinator | Filter map/list by coordinator/status/urgency → approved camp at `/camp.html?id=<id>` | Can review nonprivate report fields and service needs. A camp without a pin still appears in the list; an approximate Maps search must not masquerade as exact coordinates. |
| Admin operator | `/admin.html` → pending review or compact register/map → `/admin.html?camp=<id>` | Full private record, per-camp saved needs list above create form, runners, edit/complete actions. The explicit public-detail link is separate from the private camp-management view. |
| Runner | Assigned by an admin to a service need | Runner name/phone are private admin data, not a new login or public contact directory. |

## Where truth lives

- `main.go`: Go HTTP server, bbolt buckets for points, camp needs, runners, and password hash; session/CSRF checks; public projections. Read the route and its tests before changing a payload. `Point`, `CampNeed`, and `Runner` are distinct records. Public `GET /api/points` includes approved records and archive tombstones; `GET /api/camps/{id}/needs` omits deleted needs and runner assignment. Admin APIs under `/api/admin/` require a session; writes require CSRF.
- `template.html`: public map/list and submission form source. `build_data.py` injects a privacy-filtered historical sheet snapshot into ignored `index.html` and writes ignored `seed.json`. Its source CSV path is local to the original workspace; it may be unavailable to another agent. The browser merges approved API records with the bundled snapshot by code. Do not assume the snapshot is live or that pushing `template.html` refreshes the deployed `index.html`.
- `camp.html`: public camp detail for an approved ID. It shows nonprivate fields and service needs, with a link **out** to Maps—no embedded map. Prefer coordinates, then an allowed original Google Maps URL, then a clearly marked approximate place/area search. Apple Maps on iPhone; Google Maps on Android/desktop where coordinates are available.
- `admin.html`: authenticated dashboard and deep-linkable private camp view (`?camp=<id>`). The register stays compact, but pending submissions must retain all fields needed for review. The saved-needs list is visible before the create form; long camp editing is collapsed. Creating a need must refresh the list from the persisted API, and reloading its URL must show the same camp.
- `leaflet.css` and `leaflet.js` are bundled for the map. `Dockerfile` serves `index.html`, `admin.html`, `camp.html`, and these assets. `helm/` runs one writable process with a persistent `/data` volume.

## Domain and safety invariants

- Point lifecycle: public submissions start `pending`; an admin approves or rejects them. Admin-created points publish immediately. `archived` points disappear from both public API/UI and the historical snapshot fallback; keep their records for admin restoration. Never silently delete a real camp.
- Service needs have their own `OPEN`, `IN_PROGRESS`, `COMPLETED`, and `DELETED` states and `Today`, `Urgent`, `Tomorrow`, `Eventually` urgencies. A camp's aggregate status/urgency is derived from its service needs; do not conflate that with its independently reported `service` flag. `DELETED` needs are excluded from the visible lists. Check `aggregateCamp` before changing status semantics.
- `publicPoint` is the API privacy boundary, not the only leak path. Contact names, phone numbers, private camp notes, runner identifiers/phones, and values misfiled into other fields must not enter public responses, snapshots, or screenshots. Public detail is **not** a dump of the admin object. Admin links to a camp are not permission to expose admin data to a coordinator.
- Keep field intake forgiving: the server can generate a code and neutral area for a sparse report. Preserve the bounded request body, JSON parsing, safe output escaping, and admin authorization/CSRF. Do not reintroduce required optional fields or short text limits on the form that prevent emergency reports from saving.
- Keep map pins honest. Missing GPS belongs in the list, not at guessed coordinates. A non-extractable map URL may be retained for review. An approximate location search must say it is approximate and ask users to confirm before travel.
- Never change, rotate, reset, print, request in chat, or commit a live admin password while doing other work. Never invoke `.github/workflows/rotate-admin.yml` without an explicit, separate request. Do not introduce or tighten shared-IP login throttling; operators may share a network. The existing password hash is stored in bbolt; `ADMIN_BOOTSTRAP_PASSWORD` is only for a **new** database. Do not touch production data, credentials, kubeconfigs, or ignored `password`/`.env` files during ordinary work.
- No fake camps, needs, runner assignments, or test deliveries in production. Tests that write must use a disposable local bbolt database, never the live PVC or primary local DB.

## How to work here

1. Check `git status -sb`, branch, and recent commits. Read the relevant route, UI, and sibling tests; do not overwrite another agent's work. For a bug, reproduce the user's exact click or payload before changing code. Avoid unrelated refactors. Do not commit or push unless asked—`main` pushes can deploy automatically.
2. Prove the behavior end to end. Public report: submit → pending is absent from public API → admin approval → public readback. Admin need: open the camp via card or pin → create → see the new item in that camp's list → reload/reopen → edit or complete → check the public aggregate. Check mobile scrolling and horizontal overflow; do not substitute a static DOM assertion for this path.
3. Run `go test ./...`, `go vet ./...`, `go build ./...`, and `node --test feature_ui_test.mjs`. Tests live in `main_test.go`, `sparse_submission_test.go`, `login_lockout_test.go`, `camp_page_test.go`, `camp_service_test.go`, and `feature_ui_test.mjs`. When Helm changes, also run `helm lint ./helm --set image.tag=ci` and `helm template azuero-apoyo ./helm --set image.tag=ci --set ingress.enabled=true`.
4. For a local browser test, compile/run the app on loopback with `STATIC_DIR` pointed at the checkout, `DB_PATH` at a fresh disposable scratch file, and a throwaway bootstrap password for **that file only**. Exercise HTTP save/readback, stop the preview server, then discard the test DB. A fresh clone may not have generated `index.html` or the private source CSV; say so rather than inventing a snapshot.
5. If the user authorizes release, verify the target branch and prerequisites. CI is `.github/workflows/ci.yml`: tests/build/Helm checks, then gated publish/deploy from `main` when `DEPLOY_ENABLED=true`. The public snapshot comes from the `INDEX_HTML_B64` repository secret; only refresh it from the actual source after privacy review. A main-branch push can redeploy even for documentation changes. Never update a credential to make deployment pass. Read the exact CI job conclusions and the live URL/API after rollout; distinguish local-only from deployed.

Handoff succinctly: what operator path changed, what real tests/browser steps passed, whether the live site was checked, and any remaining blocker. The latest user request is the scope—not this document's list of possible tasks.
