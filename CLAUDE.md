# Claude Code — Azuero Apoyo

Read [AGENTS.md](AGENTS.md) first. This repo is an active emergency-relief tool, not a prototype or a generic camp-management SaaS. Your job is to help field workers submit imperfect reports quickly, help admins turn them into trustworthy public camp data and track service needs, and keep operators able to log in. Work from the user's **current** request; do not invent a roadmap.

## First inspection

1. Run `git status -sb` and check the current branch. The production branch may auto-deploy on push; do not commit/push unless explicitly asked. Preserve concurrent edits.
2. Read the exact user path in code: public map/list and intake in the tracked `index.html`; public camp review in `camp.html`; admin review, runners, and per-camp needs at `/admin.html?camp=<id>` in `admin.html`; storage/API/auth in `main.go`. Edit `index.html` directly for public UI changes; there is no generation step.
3. Trace the relevant tests before editing. Match existing Spanish labels and project conventions rather than adding a framework or dependencies for a small change.

## Definition of done for common requests

- Submission fix: a sparse report actually saves to the pending queue, gives a useful error when malformed, and does not appear publicly before approval. Verify with a disposable DB, not a fake live camp.
- Camp/need UI: the admin can open the **right** camp, see its saved needs above the form, create one, see it in the refreshed list, and reload the deep link without losing that list. Pending reports still show full review data. The public detail is separate and shows no private contacts or runner phone numbers.
- Map/location: unmapped camps remain in the list; the camp detail links to the external device-appropriate Maps app without embedding another map. A place-name search with no GPS is labeled approximate.
- Release: only when asked, verify Go and UI tests, CI test/publish/deploy conclusions, and the exact live click path. A successful push is not evidence of a working production page.

Run `go test ./...`, `go vet ./...`, `go build ./...`, `node --test feature_ui_test.mjs`, and `python3 -m unittest ci_index_test.py`. Use a throwaway local bbolt file for write-path browser checks; never write test needs to production. Do not fabricate snapshot records. Do not print or change credentials, run the password-rotation workflow, or add IP-wide login restrictions. `ADMIN_BOOTSTRAP_PASSWORD` initializes a new DB; it does not reset the existing hash.

End with a short factual handoff: changed files, exercised path, test result, live vs local status, and what's blocked. If a deployment or credential operation was not explicitly authorized, stop before it.
