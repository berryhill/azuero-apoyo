# Azuero relief app — API handoff

Build locally (with the tracked `index.html`): `docker build -t venao-relief:local .`

Server: port 8080; serves map at `/`, curation at `/admin.html`, health at `/health`.

Environment:
- `DB_PATH=/data/relief.db` — BoltDB file; mount a persistent volume at `/data` (writable by UID 10001).
- `S3_BUCKET`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY` (plus `S3_ENDPOINT` and `S3_REGION` for S3-compatible providers) — optional object storage for run photos/videos. In the cluster they come from the operator-created Secret `azuero-apoyo-media` (Helm `mediaEnvSecret`, optional; CI never writes it). Without object storage, uploads are turned off (pages hide the buttons and show «Fotos no disponibles»); set `MEDIA_LOCAL_DISK=1` for local development to save media under `media/` next to the database instead. `MEDIA_MAX_MB` caps each upload (default 250).
- `ADMIN_BOOTSTRAP_PASSWORD` — required only to initialize a new database; at least 12 characters. bcrypt hash stored in BoltDB. No default password. Set from a Kubernetes Secret, not the image/manifests.
- `LISTEN_ADDR=:8080`, `STATIC_DIR=/app/static` (Docker defaults).

Use one replica with an RWO volume. bbolt is a single-file embedded DB; do not run multiple pods against the same file. Put TLS at the ingress; admin password and session must not cross plaintext public HTTP. Back up the DB file with a consistent snapshot, not a live file copy.

API:
- `POST /api/submissions` JSON with the sheet fields (`code,area,place,needs,gps,people,adults,children,babies,olderAdults,disabled,notes,directions,coordinator,contact,phone,secondContact,secondPhone,thirdContact,thirdPhone,service,serviceNotes,generator,extra,row,ordinal`) → 201 `{id,status:"pending"}`. `gps` is an HTTPS Google Maps link; coordinates are derived server-side when present. No account required. Contact fields are retained for the admin view but omitted from the public API.
- `GET /api/admin/points` → all stored points with full sheet fields, including contacts; session required. The admin page defaults to this full register and has a pending-only tab.
- `GET /api/points` → approved submissions and seeded checkpoints. The public map and list read only from this API; no camp data is bundled into index.html.
- `POST /api/admin/login` JSON `{password}` → HttpOnly session cookie + JSON `{csrf}`.
- `GET /api/admin/submissions` → pending queue; session required.
- `POST /api/admin/seed` + session cookie + `X-CSRF-Token`: JSON array of checkpoint objects (`code,area,place,needs,people,coords,service,...`). Seeds as approved, idempotent by code; returns `{added,skipped}`. This admin-only API does not run during deployment. Do not expose private records publicly.
- `POST /api/admin/decisions` JSON `{id,decision:"approved"|"rejected"}` + `X-CSRF-Token`; session required.
- `POST /api/admin/password` JSON `{current, password}` (current master password required) + session cookie + `X-CSRF-Token` → bcrypt-hashes the new password in BoltDB and invalidates all sessions. Use the masked form at `/admin.html`; do not send passwords in chat or commit them.
- `POST /api/admin/logout` + `X-CSRF-Token`.

All points, including the original sheet camps (seeded via `POST /api/admin/seed`), persist in BoltDB and refresh in browsers every 30 seconds. Data is not written back to Google Sheets. Contact names and phone numbers are available only in the authenticated admin API, not the public API.

GitHub Actions runs Go tests/build, UI checks, and Helm lint/render on every push and pull request. Image publication and deployment are gated by repository variable `DEPLOY_ENABLED=true` on `main`. The repository tracks and deploys `index.html` directly; edit it for public UI changes. Its bundled historical snapshot is public, so review changes for private data. `seed.json`, CSVs, and the database remain ignored and must not be committed.

To enable CD, configure production environment secrets `KUBECONFIG_B64` (base64 of a narrowly scoped kubeconfig whose current context points to the intended cluster), `ADMIN_BOOTSTRAP_PASSWORD` (12+ characters), and `GHCR_READ_TOKEN` (durable read:packages token); repository variable `EXPECTED_KUBE_CONTEXT` (exact kubeconfig current-context); then set repository variable `DEPLOY_ENABLED=true`. The workflow verifies the context, provisions namespaced secrets, deploys an immutable image digest into `azuero-relief`, and enables ingress at `azuero-apoyo.berryhill.dev`. Verify DNS and cluster identity before enabling. Environment approval rules are recommended. No kubeconfig or admin password is stored in git.
