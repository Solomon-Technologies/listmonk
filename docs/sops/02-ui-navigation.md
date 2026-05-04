# 02 — UI Navigation

The admin UI is at <https://mail.eyeingem.com>. Layout: left sidebar
(menu groups), main pane (current view), top-right (user menu, logout).

What you see in the sidebar depends on your **role's permissions**.
A Tenant Operational Admin won't see Users, Roles, Companies, Settings
because those require platform-tier perms.

---

## Sidebar map (full menu, what each does)

```
Dashboard                              # /dashboard
                                       # Overview: counts, charts (last 30 days),
                                       # feature counts (drips, automations,
                                       # webhooks, deals, warming).
                                       # ✅ Tenant-scoped.

Lists                                  # Submenu group
├ All lists                            # /lists — every list with subscriber count
└ Forms                                # /forms — public opt-in form embed snippets

Subscribers                            # Submenu group
├ All subscribers                      # /subscribers — searchable table
├ Import                               # /subscribers/import — CSV upload
└ Bounces                              # /subscribers/bounces — bounce log

Campaigns                              # Submenu group
├ All campaigns                        # /campaigns — list view
├ New campaign                         # /campaigns/new — create form
├ Media                                # /media — uploaded images for campaign body
├ Templates                            # /templates — campaign + tx templates
├ Analytics                            # /campaigns/analytics — opens/clicks aggregates
└ A/B Tests                            # /ab-tests — split-test management

Segments                               # /segments — saved subscriber filters

Drip Campaigns                         # /drips — multi-step nurture sequences
                                       # See 03-campaigns-vs-drips.md and
                                       # 04-running-campaigns.md.

Automations                            # /automations — visual canvas builder
                                       # for branching/conditional flows.
                                       # Heavier than drips. Use when drip's
                                       # linear sequence isn't enough.

Scoring                                # /scoring — engagement-based subscriber
                                       # scores. Define rules ("opened email"
                                       # → +5 points, etc.).

CRM                                    # /crm — deals + contact activities

Email Warming                          # Submenu group
├ Campaigns                            # /warming — warming campaign list
├ Senders                              # /warming/senders — addresses we send FROM
├ Addresses                            # /warming/addresses — addresses we send TO
├ Templates                            # /warming/templates — short conversational text
└ Send Log                             # /warming/send-log — audit of every warming send

Webhooks                               # /webhooks — outbound event subscriptions
                                       # (campaign events → external HTTP endpoints)

Users                                  # Submenu group, platform-admin only by default
├ Users                                # /users — accounts (incl. API tokens)
├ User Roles                           # /users/roles/users — role definitions
├ List Roles                           # /users/roles/lists — per-list permission roles
└ Companies                            # /companies — tenant management (Solomon-fork)

Settings                               # Submenu group, platform-admin only
├ General                              # /settings — site name, timezone, lang
├ SMTP                                 # /settings#smtp — outbound mail servers
├ Bounces                              # /settings#bounces — bounce processing
├ Privacy                              # /settings#privacy — tracking/exports
├ Security                             # /settings#security — CAPTCHA, OIDC
├ Appearance                           # /settings#appearance — admin/public CSS-JS
├ Maintenance                          # /settings/maintenance — DB cleanup
└ Logs                                 # /settings/logs — recent server log lines
```

What a **Tenant Operational Admin** (e.g. Robert) sees:
all of the above EXCEPT Users group, Settings group. They see Companies
(if their role has `users:get` — currently it doesn't, so they don't).

What a **Tenant Super Admin** (e.g. info@) sees: same as Operational
Admin. They have a few extra perms (subscriber SQL query, bounce
management) but the menu structure is identical.

What **Platform admin** (alch3my) sees: everything.

---

## Common actions and where to do them

| To do this | Go here |
|---|---|
| Add a new email contact | Subscribers → All subscribers → + New |
| Bulk import contacts from CSV | Subscribers → Import |
| Build a one-off announcement | Campaigns → New campaign |
| Build a 3-step welcome sequence | Drip Campaigns → + New |
| Define "subscribers who opened in last 7 days" | Segments → + New |
| Save a reusable email layout | Templates → + New |
| Upload images for emails | Media → + Upload |
| Watch deliverability of a sender | Email Warming → Senders → click sender → Send Log |
| Add a new tenant (platform admin) | Users → Companies → + New tenant |
| Add a user under existing tenant | Users → Users → + New (pick company first) |
| Change SMTP server (platform admin) | Settings → SMTP |
| Audit who sent what when | Tracking — see [08-tracking-and-records.md](08-tracking-and-records.md) |

---

## Profile menu (top-right)

- **Profile** — change own password, name, email, avatar.
- **Logout** — kills the session cookie + DB row.

The header badge (Solomon-fork addition) shows the user's company name
once Vue's `$company` prototype is in place — see
[frontend/src/main.js](../../frontend/src/main.js) for the
`Vue.prototype.$company` definition. Wire it into the navbar in a
follow-up if a visual tenant indicator is wanted.

---

## Keyboard shortcuts

Listmonk doesn't have a global keyboard shortcut system. Standard
browser shortcuts apply — Cmd-K isn't bound; use the URL bar or the
menu.

---

## What's gated behind which permission

| UI element | Required permission |
|---|---|
| Lists submenu | `lists:get` or `lists:get_all` |
| Subscribers submenu | `subscribers:*` |
| Campaigns submenu | `campaigns:*` |
| Subscriber SQL query box | `subscribers:sql_query` |
| Templates → New | `templates:manage` |
| Drips submenu | `drips:get` |
| Webhooks submenu | `webhooks:get` |
| Users submenu | `users:*` or `roles:*` |
| Companies | `users:get` (read), `settings:manage` (write) |
| Settings submenu | `settings:get` |
| SMTP test button | `settings:manage` |
| Maintenance | `settings:maintain` |

A user lacking the perm doesn't see the menu item at all — Vue uses
`v-if="$can('domain:action')"` everywhere.
