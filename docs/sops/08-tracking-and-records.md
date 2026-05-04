# 08 — Tracking & Records

What gets recorded where, and how to look it up.

---

## What's recorded for every send

When a campaign sends ONE email to ONE subscriber, these rows are
written:

```
campaign_send_log     (campaign_id, subscriber_id, email,
                       sent_at, messenger, status, error_message)
   └─ "We attempted to send X to Y at time T via messenger M.
       Status='sent' or 'failed' with error_message."
   └─ Solomon fork (v7.13.0). The "Send Log" UI tab reads from this.

campaign_views        (campaign_id, subscriber_id, created_at,
                       user_agent, device_type, email_client, country)
   └─ Written when the tracking pixel loads (recipient opened the email).
   └─ One row per open — duplicates possible (one user opens 3x = 3 rows).

link_clicks           (campaign_id, subscriber_id, link_id, created_at,
                       user_agent, device_type, email_client, country)
   └─ Written when a tracked link is clicked.
   └─ One row per click; same recipient clicking the same link twice
      = 2 rows.

bounces               (subscriber_id, campaign_id, type, source, meta)
   └─ Written by the bounce processor (mailbox scanner OR webhook from
      Postmark/Sendgrid/Resend/etc.).
   └─ type: soft / hard / complaint
   └─ When bounces.count >= settings threshold, subscriber is auto-
      blocklisted (configurable).
```

For DRIPS:
```
drip_send_log         (drip_campaign_id, drip_step_id, subscriber_id,
                       status, error_message, sent_at)
   └─ Per-step send record. Doesn't fire campaign_send_log
      separately — drip is its own pipeline.
```

For WARMING:
```
warming_send_log      (campaign_id, sender_email, recipient_email,
                       template_id, subject, status, sent_at)
   └─ Every warming send.
```

For AUTOMATIONS:
```
automation_enrollments (automation_id, subscriber_id, current_node_id,
                        status, wait_until, completed_at)
   └─ Tracks where each subscriber is in the flow.
   └─ Doesn't log individual sends — those flow through the regular
      campaign pipeline if the automation node is "Send Email".
```

---

## "Did this person get my email?" lookup

UI:

1. Sidebar → Subscribers → search for them by email → click in.
2. "Activity" section shows recent campaigns/drips they've received.
3. For one specific campaign: Sidebar → Campaigns → click the campaign
   → "Send Log" tab → search by email.

SQL (Solomon admin or via DB):
```sql
SELECT cs.sent_at, cs.status, cs.error_message, cs.messenger
FROM campaign_send_log cs
JOIN subscribers s ON s.id = cs.subscriber_id
WHERE LOWER(s.email) = LOWER('alice@example.com')
  AND s.company_id = 1   -- your tenant
ORDER BY cs.sent_at DESC LIMIT 20;
```

---

## "Who has the highest engagement?" lookup

**Contact scoring** (Solomon fork, v7.4.0). Subscribers accumulate
points based on rules you define.

Example rules (Sidebar → Scoring → + New):
- "email.opened" → +5 points
- "email.clicked" → +10 points
- "email.complained" → -50 points
- Custom event via API → arbitrary delta

The processor applies these on each event, writing to
`subscribers.score` and logging in `score_log`.

To see the leaderboard:
```sql
SELECT email, name, score
FROM subscribers
WHERE company_id = 1
ORDER BY score DESC LIMIT 50;
```

---

## Aggregate analytics

| Where | What |
|---|---|
| Dashboard (`/dashboard`) | Tenant-scoped counts (subs, lists, campaigns), 30-day click/view chart, feature counts (drips/warming/etc.) |
| Campaign view (`/campaigns/:id`) | Per-campaign opens, clicks, bounces, distinct vs total |
| Campaign Analytics (`/campaigns/analytics`) | Cross-campaign chart over time |
| Drip view (`/drips/:id`) | Per-step sent + open + click counts |
| Warming Send Log (`/warming/send-log`) | Last N warming sends with filter |

Numbers update in near-real-time. Dashboard counts are computed live
(post-v7.17.0); previously they came from materialized views with a 5-min
refresh.

---

## Audit log: who did what

There's no built-in admin audit log of UI actions today. To know who
changed a campaign:

1. **For subscriber changes**: `contact_activities` records
   create/update/delete events with the acting `users.id` in
   `created_by`. Sidebar → Subscribers → click sub → Activity tab.

2. **For everything else** (campaigns, lists, templates, drips):
   `created_at` / `updated_at` timestamps only. No actor recorded.
   To investigate who touched a campaign, cross-reference with
   server logs:
   ```
   docker logs gemineye-listmonk-1 2>&1 | grep -E "PUT /api/campaigns/123"
   ```
   Listmonk logs the username in request lines.

3. **For destructive ops** (delete user, delete company): `patchlog.md`
   in repo root is where Diablo session logs incidents. Add an entry
   manually when you do something irreversible.

For richer audit, future work: enable Listmonk's debug log level +
forward to a SIEM. Or add a generic `audit_log` table writeable from
each PUT/DELETE handler. Not implemented in v7.17.0.

---

## Bounce log

Sidebar → Subscribers → Bounces, or:

```sql
SELECT b.created_at, s.email, b.type, b.source, c.name AS campaign_name
FROM bounces b
JOIN subscribers s ON s.id = b.subscriber_id
LEFT JOIN campaigns c ON c.id = b.campaign_id
WHERE b.company_id = 1
ORDER BY b.created_at DESC LIMIT 100;
```

`type`: `soft` (mailbox full / temporarily unavailable), `hard`
(permanent failure), `complaint` (recipient marked spam).

`source`: `mailbox-scanner` / `sendgrid` / `postmark` / `resend` /
`forwardemail` / etc.

---

## Tracking opt-out

A subscriber can opt out of tracking via the preferences page. When
opted out, their campaign_views and link_clicks rows are still written
(per Listmonk's design) but anonymized — `subscriber_id=NULL`.

To honor a "no tracking" toggle globally per tenant: Settings →
Privacy → Disable tracking. (Platform-admin-only.)

---

## Records retention

There's no automatic cleanup. `campaign_send_log` for a campaign that
sent 8k emails = 8k rows. After a year of campaigns: ~tens of millions
of rows.

For maintenance: Sidebar → Settings → Maintenance → "Clean up old
campaign_send_log entries" (planned button — not yet implemented).
Until then:

```sql
DELETE FROM campaign_send_log
WHERE company_id_of_campaign = 1   -- via JOIN
  AND sent_at < NOW() - INTERVAL '1 year';
```

(Done as platform admin via direct SQL.)

`campaign_views` and `link_clicks` similarly. Be aware these are
also what powers the dashboard charts — deleting old data shrinks
the chart window.

---

## Solomon-fork specifics

**campaign_send_log** is a Solomon-only addition (v7.13.0). Upstream
Listmonk doesn't track per-recipient sends explicitly — it just
increments `campaigns.sent`. Send Log gives us the granularity to
say "was this email actually sent to alice@example.com on March 10?".

**Evergreen rescans** also use `campaign_send_log` — the
`NOT EXISTS (SELECT 1 FROM campaign_send_log ...)` dedup ensures a
sub doesn't get the same evergreen campaign twice across rescans.
