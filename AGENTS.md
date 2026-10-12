# Agent guide — Azuero relief

This is a live earthquake-relief coordination app. Optimize for the exact requested operator path, not adjacent features. Make the smallest change, exercise it against a disposable local database, and report what is and is not live. Never treat a passing CI run alone as proof that the user-facing click works.

## Start here

- Check `git status -sb`, the current branch, and the latest commit before editing. Do not overwrite another agent's uncommitted work. Do not commit or push unless the user asks; a push to `main` can trigger a production deploy.
- Read the relevant code and tests first. `main.go` owns the Go HTTP API, bbolt storage, admin sessions, public projection, and static routes. `template.html` generates the public map/list snapshot; `camp.html` is a public camp detail. `admin.html` contains the admin dashboard and the deep-linkable camp-management view at `/admin.html?camp=<id>`. `build_data.py` generates ignored `index.html` and `seed.json` from a local CSV at its configured absolute path; a fresh checkout may not have that source.
- `README.md`, `.github/workflows/ci.yml`, `Dockerfile`, and `helm/` describe build and deployment. Check the implementation when documentation conflicts with it.

## Non-negotiable operational boundaries

- Do not change, rotate, reset, print, request in chat, or commit the live admin password. Do not introduce a shared-IP lockout or change authentication policy as part of a UI/data fix. Do not invoke `.github/workflows/rotate-admin.yml` without a separate explicit request. Existing credentials and the database must survive deployments.
- Keep contact names, telephone numbers, private camp notes, and runner phone numbers out of public APIs, HTML, and generated snapshots. The public detail is for coordinator review of nonprivate report fields, not the entire admin record. `publicPoint` in `main.go` is the API privacy boundary; check free-text fields for misfiled contacts too. Never open or print the ignored `password`, `.env`, kubeconfig, or production data files.
- Public submissions are pending until an admin approves them; admin-created points publish immediately. Archived points must stay out of the public map/list, including the bundled snapshot fallback. The bbolt file is a single-writer persistent store: one server replica, persistent `/data` volume.
- Public camp details should link out to Apple Maps on iPhone and Google Maps on Android/desktop; do not embed another map on the detail page. A camp without GPS may offer a clearly labeled approximate place search, never an implied precise pin. Admin camp-management links must not expose admin data publicly.
- Do not create fake records or test needs in production. Use a disposable local bbolt file outside the repo for write-path tests. Do not invent data to hide a blocked test.

## Practical workflow

1. Reproduce the exact user path. For admin needs: dashboard card/map pin → `/admin.html?camp=<id>` → existing needs list → create/edit/complete → refreshed list. The admin detail view is session-gated by its API, and its public-detail link is separate.
2. Change only relevant files. Preserve the pending-review fields and the privacy split. For mobile layout, verify actual scrolling and no horizontal overflow in a browser.
3. Run `go test ./...`, `go vet ./...`, `go build ./...`, and `node --test feature_ui_test.mjs`. If changing Helm, run `helm lint ./helm --set image.tag=ci` and `helm template azuero-apoyo ./helm --set image.tag=ci --set ingress.enabled=true`.
4. For an HTTP smoke test, use `DB_PATH` under a disposable scratch directory, `STATIC_DIR` pointing at this checkout, and a throwaway bootstrap password for that new database only. Test the real save→list→reopen flow locally; stop the preview server and discard its test DB afterward.
5. If and only if a release is requested, check the exact target and CI status first. The public `index.html` is generated and git-ignored: CI supplies it through `INDEX_HTML_B64`. Do not regenerate/upload it without the source CSV and a privacy review. Pushing `main` with `DEPLOY_ENABLED=true` publishes and deploys; check test, publish, and deploy jobs, then read back the live page/API. Never modify production credentials as a release prerequisite.

Production URLs: `https://azuero-apoyo.berryhill.dev/` and `/admin.html`. Keep handoff reports short: changed files, tests/browser path exercised, deployment state, and remaining blocker. Do not claim an unpushed change is live.
