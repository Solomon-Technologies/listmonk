# 04 — Running Campaigns (Step-by-Step)

Concrete workflows. Read [03-campaigns-vs-drips.md](03-campaigns-vs-drips.md)
first to know which to use.

---

## A. Regular Campaign — quick send

1. **Subscribers loaded?** Confirm the list you'll target has
   subscribers: Sidebar → Lists → click your list → see "Subscribers"
   count. If empty, import or add first.

2. **Template ready?** Sidebar → Campaigns → Templates. Pick or build
   a template (see [07-templates.md](07-templates.md)). At minimum a
   default exists.

3. **Sender SMTP exists?** Sidebar → Settings → SMTP (platform admin
   only) — at least one server must be enabled. Tenant admins skip this
   — your tenant's messenger is already configured.

4. **Create the campaign:** Sidebar → Campaigns → **+ New campaign**.
   Fill in:
   - **Name** (internal): e.g. `Q2 Launch — Solomon`
   - **Subject** (recipient sees): e.g. `New tools for your team`
   - **From email**: e.g. `Solomon <hello@solomontech.co>` (must match
     a configured messenger sender)
   - **Lists**: pick the target list(s)
   - **Type**: leave `regular` for normal sends
   - **Messenger**: pick the messenger (the dropdown only shows your
     tenant's messengers)
   - **Tags**: e.g. `q2,launch`
   - **Template**: pick one
   - **Body**: write your email (Markdown/HTML/Visual)
   - **Send at**: leave blank for immediate; set future timestamp to
     schedule

5. **Preview:** click "Preview" — opens a rendered preview using a
   sample subscriber's data merged into the template.

6. **Test send:** there's a "Test" tab on the campaign edit page
   (after first save). Enter your own email → "Send Test". Verify
   format/links before mass-sending.

7. **Send / schedule:** save → click **Start campaign** (or Schedule).
   Status moves draft → running.

8. **Monitor:** Campaigns list → your campaign row shows live
   `sent / to_send` counts. Click in for opens/clicks/bounces.

9. **Stop early?** Pause or Cancel from the campaign view. Pause is
   reversible (resume); cancel is final but doesn't roll back already-
   sent messages.

### Evergreen variant

Toggle the **Evergreen** switch on the campaign edit page (Solomon
fork addition). Behavior change:

- The campaign stays in `running` status after the initial drain.
- A scanner re-runs at intervals and re-queues the campaign to send to
  subscribers who joined the target list AFTER the initial send.
- `campaign_send_log` dedup ensures no one gets it twice.
- Use for "always-on" welcomes that don't need their own drip.

To force a re-scan now (e.g. after editing the body): click **Rewind**
on the campaign edit page. Resets the cursor; the next manager tick
will pick it up.

---

## B. Drip Campaign — multi-step nurture

Example: 3-step welcome series for new AnilTX trial signups.

1. **Build your trigger:**
   - If trigger is "joined a list" → make sure that list exists.
   - If trigger is "joined a segment" → build the segment (Sidebar →
     Segments → + New, define WHERE conditions).

2. **Create the drip:** Sidebar → Drip Campaigns → **+ New**. Fill:
   - **Name**: `AniltX — Trial Signup → 3-step Onboarding`
   - **Trigger type**: `subscription`
   - **Trigger config**: pick the trial-signup list ID
   - **From email**: tenant's send-from
   - **Status**: `draft` initially

3. **Add steps:** click **+ Add step** for each email in the sequence.
   Each step:
   - **Sequence order**: 1, 2, 3, ...
   - **Delay**: e.g. `0 hours` (immediate) for step 1, `2 days` for
     step 2, `5 days` for step 3
   - **Subject + body** + Template + Messenger
   - Optional **send conditions** — JSON expressions that skip the
     step if not met (e.g. skip if subscriber already has tag X)

4. **Activate:** flip Status to `active`. From this moment, every
   subscriber matching the trigger gets enrolled in the sequence.
   Existing subs who already match get enrolled too (one-time backfill).

5. **Monitor:** Drip Campaigns → click in → "Enrollments" tab shows
   active/completed/exited counts. The "Send Log" shows individual
   sends across all enrolled subscribers.

6. **Edit a running drip?** You CAN edit step bodies/subjects/delays
   while active. New enrollments use the new content; in-flight
   subscribers continue with whatever was their next step (no re-send).

7. **Pause / Archive:** flip Status. Paused drips stop sending but
   don't drop enrollments — resuming picks up where they left off.
   Archived = read-only history.

### Drip vs evergreen campaign — when to choose

- **Drip**: per-subscriber timing, multiple emails over days/weeks,
  each subscriber has their own "Day 1, Day 3, Day 7"
- **Evergreen**: single email, sent to subs whenever they qualify, no
  step structure

If your "welcome" is one email, use evergreen. If it's a sequence,
use a drip.

---

## C. Automation — branching flow

Reach for this only when drips can't model the logic.

1. Sidebar → Automations → **+ New**.
2. Name + initial canvas.
3. Drag nodes from the palette: Trigger, Wait, Send Email, Condition,
   Add Tag, Webhook, Exit.
4. Connect with edges. Each edge can be labeled (e.g. "if opened",
   "if clicked", "else").
5. Save → Activate.

**Authoring tips:**
- Always include exit nodes for every branch — otherwise enrollments
  pile up in `wait_until` indefinitely.
- Test with a fake subscriber first by manually enrolling them
  (admin-only API or button).
- Webhooks are powerful but make sure the receiver is idempotent —
  the automation will retry on transient failure.

---

## D. A/B Test on a Campaign

A/B = send two (or more) variants of one campaign to a small percentage
of the list, pick the winner, then send the winner to the rest.

1. Create a regular campaign as in section A. Save (status=draft).
2. On the campaign edit page → "A/B test" tab → **+ Enable**.
3. Configure:
   - **Test type**: `subject` (test subject lines), `body` (test bodies),
     `from_email`, `template`
   - **Test percentage**: e.g. `20` — 20% of the list gets one variant or
     the other, evenly split.
   - **Winner metric**: `open_rate` or `click_rate` (default `open_rate`)
   - **Winner wait hours**: how long to wait before deciding (default 4)
4. Add 2+ variants (label A/B/C/...). Each gets the field-under-test
   value.
5. Save → start the campaign as normal.
6. Listmonk holds back 80% of the list. After 4 hours (or your wait
   time) it picks the winner by metric and sends to remaining 80%.
7. Final stats per variant + the winning variant flag are visible on
   the campaign page.

**Records produced:**
- `ab_tests` row per test
- `ab_test_variants` rows per variant
- `ab_test_assignments` rows per subscriber → variant

---

## E. Test sending without affecting subscribers

Three options:

1. **Test send** on the campaign edit page — sends to YOUR address only,
   no `campaign_send_log` row, doesn't decrement `to_send`.

2. **Test list** — create a private list of internal addresses
   (e.g. `Test — Internal`) and target campaigns to that list first.
   Adds rows to all logs (looks like a real send). Use for end-to-end
   validation including tracking pixel/click rewrite.

3. **Disabled subscribers** — set sub `status='disabled'` to keep them
   in the list but skip sending to them. Useful for staging accounts
   that should remain in the list but never receive.

---

## F. Common mistakes

| Symptom | Cause | Fix |
|---|---|---|
| Campaign stuck in `running` with `sent < to_send` | Worker hit rate limit, paused, or messenger error | Check logs (`docker logs gemineye-listmonk-1`). Often messenger creds expired. |
| Drip subscribers never get step 2 | Step 2 has a `send_conditions` JSON they don't match | Edit step or add tag they DO have |
| Test send works but real campaign doesn't | Real campaign uses different messenger than test | Verify Messenger field on the campaign matches a working SMTP server |
| "Unknown messenger" error | Campaign saved with messenger that's no longer configured | Edit campaign, pick a current messenger from the dropdown |
| Subscribers receive email twice | Subscriber is on two of the targeted lists, no dedup | Use a segment that ANDs the lists, or a single list |
| Evergreen campaign stuck (no new sends) | `last_subscriber_id` advanced past current max | Click Rewind button to reset cursor |
