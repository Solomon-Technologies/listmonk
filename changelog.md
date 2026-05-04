# Solomon Listmonk Fork — Changelog

Chronological record of code/schema changes shipped to the Solomon fork. Each entry references the session file that produced it.

Entries are append-only.

---

## v7.17.0 (in progress, 2026-05-03)

**Multi-tenant schema fork — Rule27 Design split**

(v7.16.0 was the per-campaign warming recipient picker, shipped earlier — see ba0c3e49 history.)

- New `companies` table seeded with id=1 (Solomon Technologies), id=2 (Rule27 Design)
- `company_id INT NOT NULL DEFAULT 1` added with FK + index to: lists, subscribers, campaigns, templates, media, bounces, links, campaign_send_log, campaign_views, link_clicks, segments, webhooks, drip_campaigns, automations, scoring_rules, ab_tests, crm_*, warming_addresses, warming_campaigns, warming_send_log, users
- 200 cross-brand subscribers duplicated under company_id=2
- `subscribers.email` global unique → `(email, company_id)` unique
- Messengers refactored from JSON-in-`settings` → dedicated `messengers` table with company_id FK
- Rule27 admin users seeded: `info@rule27design.com` (super), `robert@rule27design.com` (operational)
- Feature flag `app.enforce_company_isolation` added (default off until smoke-tested)
- Handlers for lists/campaigns/subscribers/templates/media/bounces/segments/drips/automations/scoring/warming filter by `user.company_id` when flag on
- New `CompanyScopeMiddleware` defense-in-depth above query filters
- Send-time messenger.company_id == campaign.company_id assertion

**Status (as of 02:50 2026-05-04):** **v7.17.0 multi-tenant fork FULLY COMPLETE.** Final image `1229c879e7ce`. Phase 1 (schema) + Phase 3 (frontend $canCompany + Companies CRUD UI) + Phase 4 ALL features (lists, campaigns, subscribers, templates, warming, bounces, segments, webhooks, drips, automations, scoring, ab_tests, CRM, media + messenger picker) + Resend bounce webhook re-integrated. Rule27 users live (info@=Super Admin id=3, robert@=Operational Admin id=4 under company_id=2). Isolation flag ON via docker-compose env. Smoke test confirms full isolation across every feature surface. Phase 5.1 (defense-in-depth middleware) and Phase 5.2 (send-time messenger check) explicitly scoped out — SQL filter + UI picker is the defense.

**Session:** [docs/sessions/05-03-26-multitenant-fork.md](docs/sessions/05-03-26-multitenant-fork.md)

---

## Prior versions

- **v7.15.0** — Evergreen campaigns
- **v7.14.0** — `campaign_send_log` retention on subscriber delete
- **v7.13.0** — `campaign_send_log` table
- **v7.12.0** — Per-campaign messenger selection
- **v7.11.0** — Per-sender warming campaigns
- **v7.10.0** — Warming progressive ramp + hourly cap + business hours
- **v7.9.0** — Warming campaigns
- **v7.8.0** — Email warming feature
- **v7.7.0** — Drip production enhancements
- **v7.6.0** — CRM settings
- **v7.5.0** — Enhanced analytics + templates + CRM
- **v7.4.0** — Contact scoring
- **v7.3.0** — Visual automation builder
- **v7.2.0** — A/B testing
- **v7.1.0** — Drip campaigns
- **v7.0.0** — Segments + Webhooks (Solomon fork begin)
