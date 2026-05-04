# 03 — Campaigns vs Drips (and Automations)

These three concepts overlap. Pick the right tool or you'll be
fighting the platform.

---

## Quick decision tree

```
Sending one message to a fixed audience right now (or scheduled once)?
├─ YES → Regular Campaign       (Campaigns → New)
└─ NO ↓

Sending a SEQUENCE of messages over time, triggered by a single event?
├─ YES → Drip Campaign          (Drip Campaigns → New)
└─ NO ↓

Branching/conditional logic (e.g. "if they clicked this, send X; else
wait 3 days and send Y; if they unsubscribe, exit")?
└─ YES → Automation              (Automations → New)
```

If unsure, **start with a regular campaign**. They're the simplest
mental model and you can layer drips on top later.

---

## 1. Regular Campaigns

**What it is:** one email blast to a list of subscribers, sent once
(or scheduled once).

**Use cases:**
- Product launch announcement
- Newsletter (weekly/monthly)
- Promo or sale notification
- One-off event invite
- Re-engagement blast to dormant subscribers

**Key fields when creating one:**
| Field | Meaning |
|---|---|
| Name | Internal label, not sent |
| Subject | What recipients see in inbox |
| From email | Sender address (must match a configured messenger) |
| Lists | Who receives it (one or more lists) |
| Type | `regular` (default) or `optin` (sends double-opt-in confirmations) |
| Send at | Leave blank for "send now"; set a future timestamp to schedule |
| Messenger | Pick the SMTP backend (filtered by tenant — see 02 nav) |
| Template | Wraps the body content |
| Body | The actual email content (Markdown / HTML / Visual editor) |
| Tags | Free-text labels for grouping/filtering campaigns |

**Status lifecycle:**
```
draft → scheduled → running → finished
                         ↓ paused → running (or cancelled)
                  cancelled
```

**Solomon fork additions:**
- `is_evergreen` — when true, the campaign keeps re-sending to newly-
  added subscribers automatically (re-scanning at scheduled intervals).
  Use for "welcome new joiners with X" without rebuilding a drip.
- Per-campaign messenger override — pick a specific tenant's SMTP
  per campaign (Solomon's vs Rule27's).
- "Rewind" button on evergreen campaigns — manually reset the cursor
  so the next tick re-scans the entire list (sends only to subs who
  haven't received it yet, via the campaign_send_log dedup).

**Tracking:** opens, clicks, bounces, individual sends — all per
campaign. See [08-tracking-and-records.md](08-tracking-and-records.md).

**Records produced:**
- `campaigns` row
- `campaign_lists` rows (link to target lists)
- `campaign_send_log` rows (per recipient)
- `campaign_views` rows (per open)
- `link_clicks` rows (per click)
- `bounces` rows (if delivery fails)

---

## 2. Drip Campaigns

**What it is:** a SEQUENCE of N emails ("steps") sent over time,
triggered when a subscriber meets a condition. One subscriber moves
through the sequence linearly.

**Use cases:**
- Welcome sequence after a list signup (Day 0 / +1 / +3 / +7)
- Onboarding reminders ("Have you tried feature X yet?")
- Lead-nurture for sales (educational content over 2 weeks)
- Course delivery (one lesson per day)
- Stale-lead win-back (3 emails over 14 days, exit if they engage)

**Trigger types:**
| Trigger | When subscriber enters the sequence |
|---|---|
| `subscription` | Joined a specific list |
| `segment_entry` | Now matches a saved segment |
| `tag_added` | Tag added to subscriber |
| `date_field` | A relative date in subscriber's `attribs` arrives |
| `manual` | Enrolled by API or admin button |

**Step structure (each step has):**
- Sequence order (1, 2, 3...)
- Delay (e.g. "+1 day after previous step")
- Subject + body (just like a regular campaign)
- From email + messenger + template
- Optional send conditions (skip this step if X)

**Lifecycle of a subscriber in a drip:**
```
enrolled (status=active, current_step_id=step1, next_send_at=NOW + step1.delay)
   ↓ (step1 sends, processor advances)
on step 2 (status=active, next_send_at=step1.sent_at + step2.delay)
   ↓
... last step ...
   ↓
completed (status=completed, completed_at=NOW)

Or at any point: exited (sub unsubscribed / blocklisted / manually exited)
```

**The processor** runs every 30s (Solomon fork): scans
`drip_enrollments` where `status=active AND next_send_at <= NOW()`,
sends the current step, advances or marks complete.

**Records produced:**
- `drip_campaigns` row (the sequence definition)
- `drip_steps` rows (one per email in the sequence)
- `drip_enrollments` rows (one per subscriber in the sequence)
- `drip_send_log` rows (per individual send within the sequence)

**When NOT to use a drip:**
- If you need branching ("if opened email 1, do X else Y") → use
  **Automation** instead.
- If you just need ONE message → use a **regular campaign**.
- If the audience is fixed and you want all of them on the same
  schedule (e.g. "send announcement Monday, follow-up Wednesday to
  the same group") → use two scheduled regular campaigns. Drips are
  for individual-paced sequences.

---

## 3. Automations

**What it is:** a visual flow (drag-and-drop canvas) with branches,
conditions, waits, and actions.

**Use cases:**
- "If subscriber opened welcome email, tag them 'engaged' and send
  upsell. If not, send re-engagement. If still nothing after 3 days,
  unsubscribe."
- "When form X submitted, add to list Y, send confirmation, wait 2
  days, send onboarding."
- Cross-channel orchestration (email + webhook to Slack on conversion).

**Node types** (in the canvas):
- Trigger (entry condition)
- Send email
- Wait (duration)
- Condition / Branch
- Add tag / Remove tag
- Update subscriber field
- Webhook (call external URL)
- Exit

**When you need automations vs drips:** drips are linear, automations
have branches. Drips are simpler to author and debug. Most "send
sequence" use cases fit drips. Reach for automations when you genuinely
need conditionals.

**Solomon fork uses automations sparingly today.** Most flows are
drips. Automations are heavier to set up and harder to audit.

**Records produced:**
- `automations` row (the canvas definition)
- `automation_nodes` and `automation_edges` (the visual graph)
- `automation_enrollments` rows (per subscriber in the flow)

---

## Side-by-side

| Aspect | Regular Campaign | Drip | Automation |
|---|---|---|---|
| Audience | Lists at send time | Per-subscriber based on trigger | Per-subscriber based on trigger |
| Number of emails | 1 | Linear sequence (N steps) | Arbitrary graph |
| Pacing | All at once (or scheduled once) | Per-subscriber timeline | Per-subscriber timeline w/ conditions |
| Re-trigger | One-time (unless evergreen) | Per-subscriber once (deduped via enrollment uniqueness) | Per-subscriber once |
| Branching | No | No | Yes |
| Authoring complexity | Low | Medium | High |
| Tracking | Per-campaign opens/clicks | Per-step + overall enrollment counts | Per-node + overall enrollment counts |
| Stored in | `campaigns` | `drip_campaigns` + `drip_steps` + `drip_enrollments` | `automations` + `automation_*` |

---

## Naming conventions (Solomon style)

- Campaigns: `{Brand} — {Topic} {Date}` (e.g. `AnilTX — Q2 Magnet 2026-04`)
- Drips: `{Brand} — {Trigger} {Outcome}` (e.g. `AniltX — Trial Signup → Onboarding 7d`)
- Automations: `{Brand} — {Goal}` (e.g. `Solomon — Lead Qualification Branch`)

Tags help filter the campaign list view — use them liberally
(`q2`, `nurture`, `magnet`, `winback`).
