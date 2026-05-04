# 10 — Troubleshooting

Common issues and the fastest path to a fix. Grouped by symptom.

---

## Login / access

### "Invalid credentials" but I'm sure the password is right

- Verify Caps Lock and email casing.
- Check user status: `SELECT username, status FROM users WHERE email='...'`. If `disabled`, a platform admin needs to flip it to `enabled`.
- Check `password_login` boolean: if false, password auth is disabled — user must use OIDC (if configured) or have password_login flipped on.

### Logged in but Companies / Users / Settings menus are missing

Expected — your role doesn't have the perms. Companies needs
`users:get` (read) and `settings:manage` (write). Users group needs
`users:*`. Settings needs `settings:get`. Tenant Super Admin and
Operational Admin lack these by design (only platform admin has them).

### I can see Solomon's data as a Rule27 user

Bug. Should be impossible with `app.enforce_company_isolation=true`.
Check:
1. `docker exec gemineye-listmonk-1 env | grep enforce` — must say `true`.
2. Restart the container if env var was added without recreate.
3. Verify your user record's company_id matches the tenant you expect:
   `SELECT username, company_id FROM users WHERE email='...'`.
4. If company_id is wrong, fix it via SQL: `UPDATE users SET company_id=2 WHERE id=...`.

---

## Campaigns

### Campaign stuck in `running` and not progressing

- Check campaign manager logs:
  ```
  docker logs gemineye-listmonk-1 2>&1 | grep -i 'campaign\|messenger\|error' | tail -30
  ```
- Common cause: messenger connection error (SMTP creds wrong).
  Settings → SMTP → test connection.
- Recovery: `PUT /api/campaigns/:id/status {"status":"paused"}`, fix
  messenger, then re-resume to `running`. The manager picks up where
  it left off (uses `last_subscriber_id`).

### "Unknown messenger X on campaign Y" error

Campaign was saved with a messenger name that's no longer registered
(e.g. config was changed, removing it). Edit the campaign and pick a
currently-configured messenger from the dropdown.

### Campaign sent but recipients say they didn't receive

Order of investigation:
1. Verify `campaign_send_log` row exists for the recipient with `status='sent'`.
2. If `status='failed'`, check `error_message`.
3. If `status='sent'`, check the recipient's spam folder.
4. Check authentication: SPF/DKIM records for the from-domain. Use a
   tool like https://mail-tester.com to score deliverability.
5. Check the messenger provider's dashboard (Resend/Sendgrid/etc.) for
   bounce/complaint flags.

### "expected N arguments, got M" SQL error

Schema/Go drift. The SQL query has N positional params but the Go
caller passes M. After v7.17.0 added `company_id`, all queries got
extended — if you see this it means a code path didn't get updated.
Search the binary's commit log; file an issue with the failing query
name.

---

## Subscribers

### Import failed: "duplicate key value violates idx_subs_email_company"

Same email already exists in the same tenant. Either:
- Set import mode to `upsert` (updates existing rows).
- Skip duplicates: `mode=subscribe` and `overwrite=false`.
- Pre-deduplicate the CSV.

### Subscriber shows "blocklisted" and I want to re-enable

```
PUT /api/subscribers/:id  body: {"status": "enabled"}
```

Or in UI: subscribers table → click subscriber → toggle status.

Note: blocklist is set automatically when bounce count exceeds
threshold. Just flipping status doesn't change underlying behavior —
they may bounce again. Investigate WHY they bounced first.

### Cross-tenant subscribers (200 cross-brand from migration)

After v7.17.0, 200 emails exist as separate rows in both Solomon and
Rule27. If you edit one, the other is untouched. To unify them after
the fact: it's not supported — the architecture intentionally treats
them as separate contacts because each tenant owns their own
relationship.

If you really need to: SQL-merge attribs/score across the two rows
manually, but this contradicts the multi-tenant model.

---

## Drips

### Subscribers enrolled but not progressing

- Check the drip processor is running: `docker logs gemineye-listmonk-1 2>&1 | grep drip | tail -20`. You should see log lines every 30 seconds.
- Check the drip status is `active`: `/api/drips/:id` → `status` field.
- Check `drip_enrollments.next_send_at` for affected subscribers — if
  it's far in the future, the step delay is high.
- Check `drip_enrollments.status='active'`. If it's `exited`, the sub
  unsubscribed from a list or was blocklisted.

### Drip step skipped with no log

Check the step's `send_conditions` JSON. If conditions are `[{"field": "X", "op": "=", "value": "Y"}]` and the subscriber doesn't match, the step is silently skipped and the next step is attempted.

---

## Warming

### Send log empty after activating a warming campaign

- Worker tick interval is 1 minute. Wait at least 90 seconds.
- Check `warming_campaigns.status='active'`.
- Check `recipient_ids` either contains valid `warming_addresses.id`s OR is empty (which means "all active warming_addresses for this tenant").
- Verify at least one `warming_addresses` row has `is_active=true` AND `company_id` matches.

### Daily cap reached too early

- Check `warming_campaigns.daily_limits[N]` for the current day-since-warmup-start.
- Check `hourly_cap` — if 50/hr and you have 4 runs at 3 sends each = 12/hr, you'll never hit the hourly cap. But the campaign daily total might still hit `daily_limits[N]`.
- Edit the schedule and increase if you want more.

### Warming sends getting bounced/spam-foldered

This is the point of warming — fix DKIM/SPF/DMARC for the sender domain BEFORE running real campaigns. Check:
- DKIM record published and matches what the SMTP provider signs with
- SPF includes the SMTP provider's IPs
- DMARC policy at least `p=none` for monitoring (not `p=reject` until you're confident)

Run https://mxtoolbox.com/SuperTool.aspx checks on your sender domain.

---

## Database

### "violates foreign key constraint fk_X_company"

You're trying to insert a row with `company_id` that doesn't exist in
`companies`. Either insert the company first, or use an existing company_id.

### Cannot delete company "still referenced from table X"

Expected — `ON DELETE RESTRICT` blocks deletion when ANY row in any
tenant table references the company. The Companies admin UI shows row
counts to help you see which tables to clear first.

To force-delete (destructive): clear all referencing rows manually,
THEN delete the company.

### Migration v7.17.0 won't apply

Already applied. Check `SELECT value FROM settings WHERE key='migrations'`. If `v7.17.0` is in the array, you're already on it.

If the migration partially applied and crashed: it's wrapped in a
single transaction so partial state shouldn't be possible. If you
suspect it: check `psql \d companies` — if the table doesn't exist,
the migration didn't commit. Re-run the upgrade:
`docker compose run --rm listmonk ./listmonk --upgrade --yes`.

---

## Messengers / SMTP

### Test connection in Settings → SMTP returns "auth failed"

- Username/password wrong (most common). Generate a fresh API key from
  the provider (Resend dashboard, etc.).
- TLS settings wrong: try `TLS` (port 465) vs `STARTTLS` (port 587).

### Per-tenant messenger not appearing in campaign dropdown

The messenger filter (Phase 4.9) excludes messengers whose name
contains another tenant's slug.

If your messenger is named `email-resend-rule27` and you're a Solomon
admin: it won't show up. That's intentional.

If your messenger is named `email` (generic) it shows for everyone.

To make it show: rename to include your slug
(e.g. `email-resend-solomontech` for Solomon) or remove all tenant
slug substrings.

---

## Build / deploy

### Build fails: "not enough arguments in call to X"

A Go signature changed but a call site wasn't updated. After v7.17.0 most core methods take `companyID int` last. Check the line cited and add the missing arg (use `0` for internal flows, `a.tenantFilter(c)` for handler reads, `auth.GetUser(c).CompanyID` for handler creates).

### Deploy looks successful but old code still running

You forgot to `docker compose up -d --force-recreate listmonk`. Just
rebuilding the image doesn't redeploy. Always `--force-recreate`.

### Container restart loop

Likely a panic on startup. `docker logs gemineye-listmonk-1` shows the
panic stack. Common causes:
- Migration error (failed to apply, but startup retries it).
- Config error (missing required env var).
- Port conflict (another process on 9000).

To debug interactively: `docker run -it --rm --network gemineye-network solomon-listmonk:latest sh` then run `./listmonk` manually to see the error.

---

## Where to ask for help

- Solomon fork issues: https://github.com/Solomon-Technologies/listmonk/issues
- Upstream Listmonk: https://github.com/knadh/listmonk/issues
- Session logs (this repo): [docs/sessions/](../sessions/)
- Patch log: [patchlog.md](../../patchlog.md)
- Security issues: see [security.md](../../security.md)

For urgent platform-admin help, contact `alch3my@solomontech.co`.
