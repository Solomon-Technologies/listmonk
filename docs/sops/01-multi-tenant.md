# 01 — Multi-Tenant Model

## What "tenant" means here

A **tenant** = a company that owns its own data inside this single
Listmonk instance. Tenants are stored in the `companies` table, each row
has an `id`, a human `name`, and a URL-safe `slug`.

Every tenant-scoped record (list, subscriber, campaign, template,
warming sender, etc.) carries a `company_id` foreign key to
`companies.id`. Queries filter by it. Cross-tenant access returns
404.

Today's tenants:

```
id=1  Solomon Technologies      slug=solomon
id=2  Rule27 Design              slug=rule27
```

Solomon currently consolidates AnilTX, Auldrom, Solomon Tech, Byte Arch,
and Skulptor under one tenant. If/when those need full split, add new
rows to `companies` (UI: Sidebar → Users → Companies → New) and
manually migrate the data via SQL. Schema supports it; no code changes.

---

## Role tiers

Three tiers exist. The boundary between platform-admin and
tenant-admin is the most important one.

### Tier 1 — Platform Admin (god-tier)

- **Who:** users with `user_role_id = 1` (the legacy Listmonk
  SuperAdminRoleID short-circuit). Today: `alch3my@solomontech.co` (id=1)
  and `gemineye-api` (id=2).
- **Sees:** **everything across all tenants** when
  `app.enforce_company_isolation` is OFF, OR everything in their own
  tenant when ON. (Current prod: ON, so even alch3my sees only Solomon's
  data through normal handlers — to see all tenants, use SQL or flip
  the flag off.)
- **Can:** manage tenants (Companies CRUD), manage users (any tenant),
  manage roles, change global settings (SMTP, CAPTCHA, site config),
  trigger maintenance ops.
- **Specific permissions:** `users:*`, `roles:*`, `settings:*` plus the
  legacy super-admin short-circuit which short-circuits any other check.

### Tier 2 — Tenant Super Admin

- **Who:** users with a role named "Super Admin" but `user_role_id != 1`
  (i.e. company_id=2 Rule27 Super Admin = role id=3).
- **Sees:** all of their tenant's data — lists, subscribers, campaigns,
  templates, drips, automations, segments, scoring, A/B tests, webhooks,
  warming, deals, activities, media, bounces, dashboard.
- **Can:** create/edit/delete any of the above within their tenant.
  Send campaigns. Manage warming. Use the SQL query box for subscriber
  segmentation.
- **Cannot:** invite users, modify roles, change platform-level
  settings (SMTP, CAPTCHA, etc.). Permissions don't include
  `users:*` / `roles:*` / `settings:*`.

### Tier 3 — Tenant Operational Admin

- **Who:** users with role named "Operational Admin"
  (Rule27 Operational Admin = role id=4).
- **Sees:** same as Tenant Super Admin — full tenant data visibility.
- **Can:** day-to-day operations — list mgmt, sub mgmt, campaign
  creation/sending, template editing, drip authoring, warming
  scheduling.
- **Cannot:** subscriber SQL query (no `subscribers:sql_query`), write
  bulk-export, or anything in the platform-admin denied set.
- Designed for staff who run campaigns under direction of the tenant's
  super admin without admin-of-admins reach.

---

## How isolation is enforced

Three layers, in order:

1. **DB schema** — `company_id INT NOT NULL` with FK + index on every
   tenant-scoped table. Per-company unique constraints replace global
   ones (e.g. `subscribers (LOWER(email), company_id)` instead of
   `subscribers (email)`).
2. **Handler queries** — every read query has
   `AND ($N::INT = 0 OR company_id = $N::INT)` filter. Created via the
   `App.tenantFilter(c)` helper. Internal flows (campaign send pipeline,
   drip processor, warming worker) pass `0` (no filter — they operate
   on records they already fetched).
3. **Feature flag** — `app.enforce_company_isolation` (default `false`,
   set via `LISTMONK_app__enforce_company_isolation` env var). When
   `false`, all queries see everything (dual-mode for safe rollout).
   When `true`, tenant filtering is enforced. **Current prod state:
   true.**

There is **no** explicit `CompanyScopeMiddleware`. The SQL filter is
the defense. URLs with `:id` for cross-tenant resources resolve to
404 because the filtered SELECT returns no rows.

---

## What's NOT yet tenant-scoped

| Surface | Behavior | Why |
|---|---|---|
| Materialized views (`mat_dashboard_counts`/`mat_dashboard_charts`/`mat_list_subscriber_stats`) | Global aggregates | The dashboard endpoints bypass these and run live company-scoped queries instead. The mat views are still refreshed by Listmonk's internal cron but only used for legacy non-dashboard paths. |
| `links` table | URL-keyed dedup | URLs are intrinsically global (one URL = one row) — click events through `link_clicks` are scoped via campaign FK. |
| `sessions` table | Per-user, no tenant | Cookie store. Each user belongs to one tenant via `users.company_id`; sessions inherit. |
| Built-in `roles` 1 (Super Admin) and 2 (Admin) | `company_id=1` (Solomon) | Legacy listmonk roles. Don't reuse them for new tenants — create new roles per tenant via Roles UI. |
| Default settings JSON (SMTP servers, etc.) | Global config | Platform-tier config. Per-tenant SMTP routing handled at messenger-name level (e.g. `email-resend-rule27`). |

---

## Adding a new tenant

UI workflow:

1. As platform admin: Sidebar → Users → Companies → **+ New tenant**.
2. Enter Name and Slug. Slug must be lowercase letters/digits/hyphens
   (e.g. `acme-co`).
3. Save. The tenant is created with no roles, no users, no data.
4. Sidebar → Users → User Roles → **+ New role**. Create at minimum a
   "Super Admin" role under the new company. Pick the perms (see Tier 2
   list above for the standard set).
5. Sidebar → Users → **+ New user**. Pick the new company in the Company
   dropdown → role dropdown filters to that company's roles → pick the
   newly-created Super Admin → set password.
6. (Optional) configure a per-tenant messenger by editing
   `config.toml`'s SMTP section, naming it e.g. `email-resend-acme` —
   the picker filter (slug substring match) will route it to that
   tenant.

API workflow: `POST /api/companies` with `{name, slug}` →
`POST /api/roles/users` with `{company_id, name, permissions}` →
`POST /api/users` with `{company_id, user_role_id, ...}`. See
[09-api.md](09-api.md).

---

## Removing a tenant

A `companies` row can only be deleted when **zero** rows in any
tenant-scoped table reference it. The schema enforces this with
`ON DELETE RESTRICT` on every FK. The Companies admin UI shows live
row counts and disables the delete button when any are non-zero.

To clear data first: delete all of the tenant's lists, subscribers,
campaigns, templates, etc. (via UI or `DELETE` SQL with
`WHERE company_id = N`), then delete the users + roles, then the
company. The Roles delete UI is the last step before the final company
delete.

If you just want to suspend rather than delete, set the tenant's users
to `status='disabled'` — they can't log in but data stays.

---

## Quick mental model

> **The schema is the boundary.** Every request goes through a query
> that filters by the requester's `company_id`. There is no other
> layer. If a future query forgets the filter, it leaks data — that's
> the only failure mode. Audit by grepping for query strings and
> checking they pass `companyID` through.

---

## See also

- [02-ui-navigation.md](02-ui-navigation.md) — what each menu item does
- [09-api.md](09-api.md) — programmatic tenant management
- [docs/sessions/05-03-26-multitenant-fork.md](../sessions/05-03-26-multitenant-fork.md) — full implementation history
- [security.md](../../security.md) — threat model + isolation guarantees
