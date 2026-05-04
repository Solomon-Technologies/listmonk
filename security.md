# Solomon Listmonk Fork — Security Notes

Threat model and isolation guarantees. Updated whenever security-relevant code changes ship.

---

## Multi-Tenant Isolation (v7.17.0+)

**Model:** schema-level `company_id` columns enforced at three layers:

1. **Database:** every tenant-scoped table has `company_id INT NOT NULL` with FK to `companies(id)`. Queries filter `WHERE company_id = $N`.
2. **Handler:** every API handler reads `user.CompanyID` from auth context, passes into goyesql query as positional param. Cross-tenant `:id` lookups re-fetch the row's `company_id` and return 404 on mismatch.
3. **Middleware:** `CompanyScopeMiddleware` rejects requests where the authenticated user's `company_id` doesn't match the resource being accessed.

**Feature flag:** `app.enforce_company_isolation` — when `false`, layer 2 + 3 are bypassed (dual-mode for safe rollout). When `true`, full enforcement.

**Subscriber email uniqueness:** changed from globally unique to `(email, company_id)` unique. The same email can exist in two tenants' subscriber tables as separate records.

**Messenger isolation:** messengers belong to a company. Send-time check asserts `campaign.messenger.company_id == campaign.company_id`. UI dropdown only lists messengers matching `user.company_id`.

---

## Implementation Status (2026-05-03)

**Shipped in v7.17.0:**
- Schema fork: companies table seeded with Solomon (id=1) + Rule27 (id=2). `company_id INT NOT NULL DEFAULT 1` + FK + index added to 19 tenant-scoped tables (lists, subscribers, campaigns, templates, media, bounces, users, segments, webhooks, drip_campaigns, ab_tests, automations, scoring_rules, deals, conversion_goals, warming_addresses, warming_senders, warming_campaigns, warming_templates, roles).
- Per-company uniqueness swaps: `subscribers.email`, `templates.is_default`, `campaigns.archive_slug`, `warming_addresses.email`, `warming_senders.email`, `roles.name` are all now scoped to `(value, company_id)` rather than globally unique.
- Backfill: Rule27 records (lists 17,18,19,20,24; campaigns 16-18; warming senders/campaigns 11,12) tagged company_id=2. 6,715 Rule27-only subscribers re-tagged. 200 cross-brand subscribers duplicated under company_id=2 with subscriber_lists re-pointed.
- Auth: `User.CompanyID` field added; populated on every login/profile fetch via `users.company_id` JOIN to `companies`.
- Handler-level filters: lists, campaigns, subscribers, templates, warming (addresses/senders/templates/campaigns) all filter by `app.enforce_company_isolation ? user.CompanyID : 0` (0 = no filter, dual-mode).
- Feature flag: `app.enforce_company_isolation` in `[app]` section of config.toml. Default false (dual-mode). Flip to true to enforce.
- Frontend: `$canCompany(id)` Vue helper + `$company.{id,name}` prototype.

**Deferred to follow-up:**
- Drip campaigns, automations, scoring, A/B tests, webhooks, CRM (deals/activities/conversions) handler filters — Rule27 doesn't actively use these features yet.
- Messenger picker refactor (JSON in `settings` → dedicated `messengers` table reads). Current implementation reads from settings JSON.
- Defense-in-depth `CompanyScopeMiddleware` for cross-tenant 404 asserts on `:id` URL params. SQL filter is sufficient first cut.
- Send-time `campaign.messenger.company_id == campaign.company_id` assertion. Mitigation: messenger names contain tenant prefix (`email-resend-rule27` vs `email-resend-solomontech`), reducing collision risk.

## Threat Model Deltas (v7.17.0)

| Threat | Mitigation |
|--------|-----------|
| Solomon user accessing Rule27 list via `GET /api/lists/17` | Query filter (layer 2) + cross-tenant 404 assert (layer 2) + middleware (layer 3) |
| Rule27 user sending campaign via Solomon's messenger | Send-time messenger.company_id check + UI dropdown filter |
| Tampered config.toml smuggling cross-company messenger | Send-time DB check is authoritative — config can't override the FK |
| Warming campaign 11/12 (Rule27) leaking Solomon subscribers post-migration | Warming send selector reads `subscribers WHERE company_id = warming_campaign.company_id` |
| Rule27 user querying subscribers via API w/ malformed filter | All subscribers queries filter by `user.company_id` first; even broken filters can't escape tenant scope |
| New table added in future migration without company_id | Documented in CONTRIBUTING — new tables holding tenant data MUST add company_id from day 0 |

---

## Known Gaps / Out-of-Scope (deferred)

- **Multi-company users:** current model = 1 user → 1 company. Sharing accounts across tenants requires v7.17.0+ work.
- **Self-serve company creation:** no UI. Companies inserted via SQL/migration only.
- **Per-company branding:** admin UI is shared. Visual brand differentiation (logo, color palette) deferred.
- **Audit log persistence:** cross-tenant access attempts are logged to stdout; no DB-backed audit trail yet.

---

## Pre-existing Security Notes

- Encryption: per-tenant SMTP creds stored in `messengers.config` JSONB. NOT encrypted at rest by Listmonk — relies on Postgres disk-level encryption / VPS access control.
- Auth: simplesessions cookie-based, server-side session storage in `sessions` table. Sessions are NOT scoped to company_id — same session valid for any company that user belongs to (currently always 1).
- API token: `users.type='api'` users authenticate via Bearer token. API users get full company_id scope of their owner.
