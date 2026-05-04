# 05 — Subscribers, Lists, Segments

The three subscriber-organization concepts. Pick the right one or
your audience targeting will be wrong.

---

## Concepts

| Concept | What it is | When you use it |
|---|---|---|
| **Subscriber** | A single contact (email + optional name + attribs JSON) | Everyone you have permission to email |
| **List** | A static collection of subscribers (you add/remove explicitly) | Audience grouping (`Newsletter`, `Customers`, `Q2 Leads`) |
| **Segment** | A SAVED FILTER over subscribers (dynamic — recomputed at use) | "Subscribers who opened in last 7 days", "Score > 50" |

---

## Subscribers

**Schema** (post v7.17.0):
```
subscribers (
  id, uuid, email, name, attribs JSONB, status,
  company, phone, lifecycle_stage,    -- v7.5.0 CRM additions
  score INT,                          -- v7.4.0 contact scoring
  company_id INT NOT NULL,            -- v7.17.0 multi-tenant
  created_at, updated_at
)
UNIQUE(LOWER(email), company_id)      -- email unique per tenant
```

**Statuses:**
- `enabled` — fine to email
- `disabled` — manually paused; campaigns skip them
- `blocklisted` — bounced too many times or marked complaint; never
  email; auto-set by bounce processor

**`attribs` JSONB field:** arbitrary structured data per subscriber.
Examples: `{"city":"Tampa","plan":"trial","trial_started":"2026-04-01"}`.
Used for templating (`{{ .Subscriber.Attribs.city }}`) and segment
conditions (`attribs->>'plan' = 'trial'`).

### Adding subscribers

| Path | Use when |
|---|---|
| Sidebar → Subscribers → + New | Single contact |
| Subscribers → Import → CSV | Bulk (recommended ≤500k per file) |
| Public form (Lists → Forms tab) | Self-service signup with optional double opt-in |
| `POST /api/subscribers` | Programmatic |
| `POST /api/subscribers/import` | Programmatic bulk |

**CSV import format:**
```csv
email,name,attribs,status
alice@example.com,Alice,"{""plan"":""trial""}",enabled
```
- `attribs` must be valid JSON (escape quotes properly).
- During import, choose lists to add them to + the subscription
  status (confirmed / unconfirmed for double-opt-in).

### Cross-brand duplicates (Solomon fork specifics)

When tenants split, an email that existed on both Solomon and Rule27
lists got DUPLICATED — one row per company. So
`alice@example.com` may exist as id=8654 (company_id=1) AND id=74819
(company_id=2). They're independent records. Editing one doesn't
affect the other. Segments and queries scoped by company never
collide.

---

## Lists

**Schema:**
```
lists (
  id, uuid, name, type, optin, status,
  tags VARCHAR(100)[], description,
  company_id INT NOT NULL,            -- v7.17.0
  created_at, updated_at
)
```

**Type:**
- `private` — admin-managed, not shown on public subscription page
- `public` — shown on public sub form (`/subscription/`)
- `temporary` — short-lived, filtered out of most UIs

**Optin:**
- `single` — sub joins immediately
- `double` — Listmonk sends a confirmation email, sub must click to
  finalize. Status=`unconfirmed` until confirmed.

**Subscription status (per subscriber-list link):**
- `unconfirmed` — pending double-opt-in
- `confirmed` — actively subscribed
- `unsubscribed` — opted out (kept for audit; campaigns skip)

### When to use a list vs a segment

- **List**: stable membership you control. "All Q2 trial signups."
  Subscribers don't join/leave automatically.
- **Segment**: dynamic — query that re-evaluates each time you use it.
  "Subscribers with score > 50" auto-includes new high-scorers.

**Targeting a campaign**: campaigns target LISTS, not segments. To
campaign-target a segment, run a "Build list from segment" action
first (planned future feature) — until then, use segments via
Drip Campaigns whose trigger is `segment_entry`.

### Common list patterns

| Pattern | Setup |
|---|---|
| Newsletter | Public, double-opt-in, used as primary subscription target |
| Internal test | Private, single, contains only staff emails |
| Suppression | Private, single, blocklist; cross-reference at campaign send time (manually exclude) |
| Drip enrollment funnel | Private, used as drip trigger source |
| Segment materialization | Private, periodically rebuilt from a segment query |

---

## Segments

A segment is a saved query — `name + match_type + conditions`.
Match type is `all` (AND) or `any` (OR). Conditions are JSON like:

```json
[
  {"field": "subscribers.status", "op": "=", "value": "enabled"},
  {"field": "attribs->>'plan'",   "op": "=", "value": "trial"},
  {"field": "subscribers.score",  "op": ">", "value": 50}
]
```

The query box compiles these into a SQL `WHERE` clause and runs
against the subscribers table (joined to `subscriber_lists` if list
filters are present).

**UI:**
- Sidebar → Segments → + New.
- Add conditions row by row.
- Click "Preview" to see matching subscriber count + sample.
- Save → segment is reusable.

**Used by:**
- Drip campaigns (trigger type `segment_entry` — sub enters when
  they newly match)
- Subscriber search filtering (Sidebar → Subscribers → "Segment"
  dropdown)
- Manual list rebuilding (run the SQL, INSERT INTO subscriber_lists)

**Cron-style segment recompute** is implicit — every read recomputes.
Be aware that complex segments over millions of rows can be slow.
For very large stable groups, materialize into a list periodically.

---

## Cleanup recipes

### Remove inactive subscribers

```
-- Subs who haven't opened or clicked in 90 days, are still 'enabled'.
SELECT s.id, s.email FROM subscribers s
WHERE s.company_id = $1
  AND s.status = 'enabled'
  AND NOT EXISTS (
    SELECT 1 FROM campaign_views v
    WHERE v.subscriber_id = s.id AND v.created_at > NOW() - INTERVAL '90 days'
  )
  AND NOT EXISTS (
    SELECT 1 FROM link_clicks l
    WHERE l.subscriber_id = s.id AND l.created_at > NOW() - INTERVAL '90 days'
  );
```

Then either set `status='disabled'` (soft) or use the Subscribers UI
delete (hard).

### Find duplicates within a tenant

After v7.17.0, the constraint is `(LOWER(email), company_id)`. So
duplicates within one tenant are impossible. If you see one, it's
because they were imported with case-sensitive emails — the
constraint normalizes via `LOWER()`. Run:
```
SELECT LOWER(email), COUNT(*)
FROM subscribers WHERE company_id = $1
GROUP BY LOWER(email) HAVING COUNT(*) > 1;
```

### Mass-suppress (blocklist)

Bounces auto-blocklist after N (configurable in Settings → Bounces).
Manual: Sidebar → Subscribers → select rows → "Blocklist" action.

---

## Permissions to edit

| Action | Required perm |
|---|---|
| View subscribers | `subscribers:get` (per-list) or `subscribers:get_all` |
| Add/edit/delete subscriber | `subscribers:manage` |
| Bulk import | `subscribers:import` |
| Run raw SQL query box | `subscribers:sql_query` (Tenant Operational Admin: ❌, Super Admin: ✅) |
| Edit a list | `lists:manage` (per-list) or `lists:manage_all` |
| Create segments | `segments:manage` |

The `subscribers:sql_query` perm is intentionally separate because
arbitrary SQL is powerful — Operational Admin doesn't get it; Super
Admin does.
