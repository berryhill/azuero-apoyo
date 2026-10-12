# Azuero relief app — API handoff

Build locally (with the generated, git-ignored `index.html` present): `docker build -t venao-relief:local .`

Server: port 8080; serves map at `/`, curation at `/admin.html`, health at `/health`.

Environment:
- `DB_PATH=/data/relief.db` — BoltDB file; mount a persistent volume at `/data` (writable by UID 10001).
- `ADMIN_BOOTSTRAP_PASSWORD` — required only to initialize a new database; at least 12 characters. bcrypt hash stored in BoltDB. No default password. Set from a Kubernetes Secret, not the image/manifests.
- `LISTEN_ADDR=:8080`, `STATIC_DIR=/app/static` (Docker defaults).

Use one replica with an RWO volume. bbolt is a single-file embedded DB; do not run multiple pods against the same file. Put TLS at the ingress; admin password and session must not cross plaintext public HTTP. Back up the DB file with a consistent snapshot, not a live file copy.

API:
- `POST /api/submissions` JSON `{code,area,place,needs,people,coords:[lat,lon]}` → 201 `{id,status:"pending"}`. No account required.
- `GET /api/points` → approved submissions and seeded checkpoints (the original sheet snapshot is also bundled into index.html as fallback; the UI deduplicates by code).
- `POST /api/admin/login` JSON `{password}` → HttpOnly session cookie + JSON `{csrf}`.
- `GET /api/admin/submissions` → pending queue; session required.
- `POST /api/admin/seed` + session cookie + `X-CSRF-Token`: JSON array of checkpoint objects (`code,area,place,needs,people,coords,service,...`). Seeds as approved, idempotent by code; returns `{added,skipped}`. Generate the local `seed.json` from the sheet export with `python3 build_data.py` and POST its JSON body after admin login. Do not expose `seed.json` publicly.
- `POST /api/admin/decisions` JSON `{id,decision:"approved"|"rejected"}` + `X-CSRF-Token`; session required.
- `POST /api/admin/logout` + `X-CSRF-Token`.

The sheet snapshot remains static; submission approvals persist in BoltDB and refresh in browsers every 30 seconds. Data is not written back to Google Sheets. No names/phone fields are accepted by the submission API.

GitHub Actions runs Go tests/build and Helm lint/render on every push and pull request. Image publication and deployment are gated by repository variable `DEPLOY_ENABLED=true` on `main`; leave it unset until all prerequisites are ready. The repository is public, but generated `index.html`, `seed.json`, CSVs, and the database are git-ignored. Do not commit them. The production image embeds the snapshot and may expose its contents to anyone with image access. Review the snapshot and GHCR package visibility before enabling publication.

To enable CD, configure production environment secrets `KUBECONFIG_B64` (base64 of a narrowly scoped kubeconfig whose current context points to the intended cluster), `ADMIN_BOOTSTRAP_PASSWORD` (12+ characters), and `GHCR_READ_TOKEN` (durable read:packages token); repository secret `INDEX_HTML_B64` (base64 of an approved generated `index.html`); repository variable `EXPECTED_KUBE_CONTEXT` (exact kubeconfig current-context); then set repository variable `DEPLOY_ENABLED=true`. The workflow verifies the context, provisions namespaced secrets, deploys an immutable image digest into `azuero-relief`, and enables ingress at `azuero-apoyo.berryhill.dev`. Verify DNS and cluster identity before enabling. Environment approval rules are recommended. No kubeconfig or admin password is stored in git.
