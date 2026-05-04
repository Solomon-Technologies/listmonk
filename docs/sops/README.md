# Solomon Listmonk — Operating Manual

Operational SOPs for the Solomon fork of Listmonk. Day-to-day usage,
multi-tenant boundaries, what gets recorded where, when to use which
feature, and how to script anything via the API.

These docs are **fork-specific**. For upstream Listmonk product docs
(installation, configuration, messengers, templating syntax, etc.) see
[docs/docs/content/](../docs/content/) or <https://listmonk.app/docs/>.

---

## Index

| # | File | Topic |
|---|------|-------|
| 01 | [multi-tenant.md](01-multi-tenant.md) | Tenant model, role tiers, platform-vs-tenant power |
| 02 | [ui-navigation.md](02-ui-navigation.md) | Where everything lives in the admin UI |
| 03 | [campaigns-vs-drips.md](03-campaigns-vs-drips.md) | When to use which (and why) |
| 04 | [running-campaigns.md](04-running-campaigns.md) | Step-by-step: regular campaigns, drips, A/B tests |
| 05 | [subscribers-lists-segments.md](05-subscribers-lists-segments.md) | Lists vs segments, importing, opt-in flows |
| 06 | [warming.md](06-warming.md) | Solomon warming SOPs (sender warm-up before sends) |
| 07 | [templates.md](07-templates.md) | Building, sharing, and applying templates |
| 08 | [tracking-and-records.md](08-tracking-and-records.md) | What gets logged where; how to audit a send |
| 09 | [api.md](09-api.md) | Auth, common endpoints, working examples |
| 10 | [troubleshooting.md](10-troubleshooting.md) | Common issues + fixes |

---

## Tenancy at a glance

This Listmonk runs **two tenants** today:

| Company ID | Name | Slug | Users |
|---|---|---|---|
| 1 | Solomon Technologies | `solomon` | `alch3my@solomontech.co`, `gemineye-api` |
| 2 | Rule27 Design | `rule27` | `info@rule27design.com`, `robert@rule27design.com` |

When `app.enforce_company_isolation` is on (current prod state), a tenant
user only sees their own company's data — lists, subscribers, campaigns,
templates, drips, automations, segments, scoring, A/B tests, webhooks,
warming, deals, activities, media, dashboard counts, charts.

Three role tiers (see [01-multi-tenant.md](01-multi-tenant.md)):

1. **Platform admin** (god-tier) — only `alch3my` and `gemineye-api`. Can
   manage tenants, users, roles, global settings.
2. **Tenant Super Admin** — full tenant data access, **cannot** invite
   users / change roles / change global settings.
3. **Tenant Operational Admin** — same as above minus a few perms (no
   audit log, no destructive bulk operations).

---

## Reading order for new operators

1. [01-multi-tenant.md](01-multi-tenant.md) — understand the boundaries first
2. [02-ui-navigation.md](02-ui-navigation.md) — find your way around the UI
3. [03-campaigns-vs-drips.md](03-campaigns-vs-drips.md) — pick the right tool
4. [04-running-campaigns.md](04-running-campaigns.md) — start sending
5. [08-tracking-and-records.md](08-tracking-and-records.md) — verify it
   landed; audit later if needed

The remaining files are reference for specific tasks.

---

## Conventions

- **Routes** are written `GET /api/path/:id` style.
- **Permissions** are written `domain:action` (e.g. `campaigns:manage`).
- **DB tables** are written in `lowercase_snake`.
- **Code paths** use repo-relative paths like
  [internal/core/lists.go](../../internal/core/lists.go).
- **Tenant scoping** is denoted "✅ scoped" / "❌ global" / "⚠️ partial".
