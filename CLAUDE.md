# Claude Code handoff

Read [AGENTS.md](AGENTS.md) before making changes. Its operational boundaries apply here too; this file is the short entry point for Claude Code.

- Work from the user's latest request, not an inferred roadmap. Inspect `git status -sb`, relevant source, and tests; leave unrelated or concurrent edits alone. Prefer a small working change over an architecture rewrite.
- Main code: `main.go` (Go/bbolt API), `template.html` (public map snapshot template), `camp.html` (public detail), `admin.html` (admin dashboard and `/admin.html?camp=<id>` needs view). The generated `index.html` and `seed.json` are ignored. `build_data.py` needs a local CSV that may not exist in a new checkout.
- Run `go test ./...`, `go vet ./...`, `go build ./...`, and `node --test feature_ui_test.mjs`; verify the actual UI path in a browser against a disposable local DB when behavior changes. Do not use production to create test records.
- Keep private contacts and runner phone numbers off public pages. Do not change or rotate the live password, modify login restrictions, or deploy unrelated work. Do not commit or push without authorization: pushing `main` may deploy to production.
- If blocked by unavailable source data, credentials, or deployment access, say exactly what is missing instead of guessing. Report whether work is only local or verified live.
