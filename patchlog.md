# Solomon Listmonk Fork — Patch Log

Bugs found and fixed, with root-cause analysis. Append-only. Distinct from `changelog.md` (which tracks features); this tracks defects and incidents.

---

## 2026-05-03 — Working tree reconciliation before multi-tenant fork

**Attempted:** Begin Phase 1 of v7.17.0 multi-tenant migration. Read schema + existing migrations to understand the codebase pattern.

**What failed:**
1. `cmd/upgrade.go` had unresolved git merge conflict markers (`<<<<<<< Updated upstream` / `>>>>>>> Stashed changes`) at lines 61–65 — code would not compile.
2. `internal/migrations/v7.13.0.go` had two divergent versions interleaved with conflict markers — one was `campaign_send_log` (matches origin), the other was `warming_campaigns.warming_address_ids` (an older obsolete attempt).
3. Local branch was 5 commits behind `origin/master` including a `feat(v7.16.0): per-campaign warming recipient picker` commit — meaning v7.16.0 was already taken upstream.
4. 13 staged-but-uncommitted files modifying bounce/warming/manager/frontend/queries — appeared to be older drafts of work already shipped to origin from Alchemy's laptop.
5. `stash@{0}: On master: !!GitHub_Desktop<master>` contained 354 insertions across 15 files, including the obsolete v7.13.0 variant.

**Root cause:** Alchemy works across two machines; this machine's working tree had stale work from prior `git stash` in GitHub Desktop that was never reconciled. The merge attempt left conflict markers in two files. The newer canonical work was committed on his laptop and pushed to `origin/master`. Prod's `solomon-listmonk:latest` image (built 2026-04-27 22:29 UTC) is built from `origin/master` HEAD — local was simply out of sync.

**Fix applied:**
1. Inspected each of the 14 staged files → confirmed all 14 differ from origin (origin has newer/refined versions).
2. Found 1 file (`internal/bounce/webhooks/resend.go`, 134 lines, Resend webhook bounce parser w/ HMAC) that exists only locally — origin has no Resend webhook implementation. **Archived to `.codex-archive/resend-bounce-webhook-2026-05-03.go.archived`** for later re-integration as separate workstream.
3. `git fetch origin && git reset --hard origin/master` → tree now at `ba0c3e49`.
4. `git stash drop stash@{0}` → obsolete stash gone.
5. Verified prod commit (ba0c3e49) matches local HEAD.

**Impact on multi-tenant fork:**
- Migration version changes: **v7.16.0 → v7.17.0** (v7.16.0 is taken).
- v7.16.0 added `recipient_ids INT[] NOT NULL DEFAULT '{}'` to `warming_campaigns`. Will need to add `company_id` to that same table.
- No other schema collisions.

**Regression prevention:**
- `.codex-archive/` directory created for future "preserve before destruction" patterns.
- Reminder to surface git state at start of every phase (Diablo pre-flight check #1: working tree clean).
- Memory entry queued: prod listmonk container name is `482357e3d7ed_gemineye-listmonk-db-1` not `listmonk-db` as memory said.

---
