# 06 — Email Warming (Solomon Fork)

Warming = systematically sending small, conversational emails between
your sender addresses BEFORE you send real campaigns. Builds the
sender's reputation with mailbox providers (Gmail, Outlook, Yahoo) so
future bulk sends don't land in spam.

This is a **Solomon fork-only** feature, not in upstream Listmonk.

---

## Why warming matters

A brand-new sending domain (or one that hasn't sent in months) has
zero reputation. Mailbox providers throttle, soft-bounce, or
spam-folder its emails by default. A "send 10k blast on day 1" plan
gets ~70%+ filtered out of the inbox.

Warming sends 5–50 conversational messages per day from your sender
TO a controlled list of recipients (your own addresses, or
opt-in warming partners). Recipients reply (or the script simulates
opens). Mailbox providers see consistent low-volume good-engagement
sends → reputation builds → real campaigns deliver well.

Typical warming runway: 2–6 weeks before a domain is "fully warm" for
high-volume sends.

---

## Concepts

| Object | Purpose |
|---|---|
| **Warming Sender** | An email address you send FROM during warming. e.g. `solomon.c@solomontech.co` |
| **Warming Address** | An email address you send TO during warming. e.g. friendly inboxes you control. |
| **Warming Template** | Short conversational text body — not branded, looks like a personal message. |
| **Warming Campaign** | The schedule + ramp config that uses the above three to send. |
| **Warming Send Log** | Audit trail of every individual warming send. |

All four objects are **tenant-scoped** (have `company_id`). Rule27
admins only see Rule27's warming setup.

---

## Quick-start: warm a new sender

Goal: take `noreply@rule27design.com` from cold to warm in 4 weeks.

1. **Add the sender:** Sidebar → Email Warming → Senders → **+ New**.
   - Email: `noreply@rule27design.com`
   - Name: `Rule27 No-Reply`
   - Brand: `Rule27 Design`
   - Brand URL + color (used in some warming templates).
   - Active: yes.

2. **Add recipients:** Sidebar → Email Warming → Addresses → **+ New**.
   - Add 10–20 friendly inboxes (your own across providers — Gmail,
     Outlook, ProtonMail, Yahoo). Variety matters.
   - Active: yes.

3. **Templates:** out-of-the-box, the v7.8.0 migration seeded ~15
   conversational templates. They're generic ("Quick sync — {{date}}",
   "Just touching base", etc.). Sidebar → Templates lets you edit or
   add tenant-specific ones.

4. **Create the warming campaign:** Sidebar → Email Warming → Campaigns
   → **+ New**.
   - Name: `Rule27 — noreply Warmup`
   - Brand: `Rule27 Design`
   - Sender: pick the one you added in step 1
   - Sender domains: `rule27design.com`
   - **Schedule:**
     - sends_per_run: 3 (start small)
     - runs_per_day: 4
     - schedule_times: `["10:00","14:00","18:00","21:00"]`
     - random_delay_min_s: 30, random_delay_max_s: 120
   - **Ramp:**
     - warmup_start_date: today
     - daily_limits: e.g. `[10, 20, 35, 60, 100, 150, 250, 400, 600, 1000]`
       (day 1 through day 10, doubles roughly each day)
     - hourly_cap: 50
     - business_hours_only: true
   - Recipient IDs: pick which of your warming addresses receive (leave
     empty = all active).
   - Messenger: pick the SMTP messenger for this brand (e.g.
     `email-resend-rule27`).
   - Status: `active`.

5. **Watch it run:** Sidebar → Email Warming → Send Log shows every
   send as it happens (`sender → recipient`, status `sent`/`failed`).

6. **Adjust as you go:**
   - Failures (bounces, rejects) > 5% → pause campaign, fix DKIM/SPF
     setup, then resume at lower daily_limit.
   - All-sends-clean for a week → bump daily_limit faster.
   - After 4–6 weeks of clean sending, the domain is warm — you can
     start running real campaigns from it. Either pause the warming
     campaign or leave it on at low daily_limit indefinitely as a
     reputation maintenance regimen.

---

## Schema (for SQL queries)

```
warming_addresses    (id, email, name, is_active, company_id)
warming_senders      (id, email, name, brand, brand_url, brand_color,
                      is_active, company_id)
warming_templates    (id, subject, body, is_active, company_id)
warming_campaigns    (id, name, brand, sender_id, sender_domains[],
                      status, sends_per_run, runs_per_day,
                      schedule_times[], random_delay_min_s,
                      random_delay_max_s, warmup_start_date,
                      daily_limits[], hourly_cap, business_hours_only,
                      messenger, recipient_ids[], company_id)
warming_send_log     (id, campaign_id, sender_email, recipient_email,
                      template_id, subject, status, error_message,
                      sent_at)
warming_config       (singleton, id=1; legacy global config)
```

`warming_send_log.campaign_id` chains back to `warming_campaigns.id`
which has `company_id` — so log entries are tenant-scoped via
JOIN.

---

## Tenant separation

After v7.17.0:

- **Solomon's** warming is in company_id=1: 9 senders (Solomon Tech,
  AnilTX, etc.), 7 active warming campaigns.
- **Rule27's** warming is in company_id=2: 2 senders (`no-reply@rule27.com`
  and `noreply@rule27.com`), 2 active warming campaigns (id 11 + 12).

Robert (Rule27 Operational Admin) sees ONLY Rule27's senders, addresses,
templates, campaigns, and send log entries. Cross-tenant URL access
returns 404.

The warming worker (server-side cron) processes BOTH tenants' active
campaigns. It uses the campaign's `messenger` field (e.g.
`email-resend-rule27`) to route through the correct SMTP, and the
campaign's `sender_id` for the from-address. No cross-tenant routing.

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| Send Log empty after activating | Worker hasn't ticked yet (interval = 1 min) | Wait 1 min, refresh |
| All sends `failed` with auth error | Messenger SMTP creds wrong/expired | Settings → SMTP, fix server, test connection |
| Sends OK but recipient inbox empty | Going to spam | Check sender's DKIM/SPF; reduce daily_limit; warm longer |
| Daily count exceeds your daily_limits[N] | Worker bug or campaign duplicated | Check `warming_campaigns` table for duplicates; pause one |
| `business_hours_only=true` but sends at 2am | Server timezone wrong | Check `TZ` env var on container; should match your business tz |
| Status `active` but no sends | `recipient_ids[]` empty AND no `warming_addresses` rows where `is_active=true` | Add active warming addresses |

---

## When to NOT use warming

- Your domain has been sending high-volume for years already → reputation
  is built, warming is overhead.
- You're sending one-off transactional (password resets, receipts) — no
  reputation issue because volume is low and engagement is high.
- Recipient list is opted-in, freshly engaged, and spam-folder placement
  isn't a problem.

Warming is for **bulk marketing sends from new/dormant domains**. If
that's not your use case, skip it.

---

## Permissions

Warming is gated by SOLOMON-fork permissions (not in upstream Listmonk
permissions.json):

| Perm | Lets user… |
|---|---|
| `warming:get` | View warming UI + read all warming records |
| `warming:manage` | Create/edit/delete warming senders/addresses/templates/campaigns |

Currently both Tenant Super Admin and Operational Admin have these
perms. Platform admin obviously has them too.
