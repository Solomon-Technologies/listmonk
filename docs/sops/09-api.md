# 09 — API Usage

Full API reference is generated from
[docs/swagger/collections.yaml](../swagger/collections.yaml). This
doc covers Solomon-fork additions, auth, and the most-used endpoints
with working examples.

Base URL: `https://mail.eyeingem.com/api`.

---

## Authentication

Two methods:

### 1. Session cookie (browser)

When you log in via the admin UI, your browser stores a session
cookie. All `/api/*` calls from the same session use it automatically.

### 2. API user (server-to-server)

Use this for scripts and integrations.

1. Sidebar → Users → + New → Type: API.
2. Pick the company (the API user is tenant-scoped just like a human
   user).
3. On save, a one-time token displays. **Copy it now** — it's only
   shown once.
4. Use as HTTP Basic Auth:
   ```
   curl -u 'username:token' https://mail.eyeingem.com/api/lists
   ```

Today's API users:
- `gemineye-api` (company_id=1, role=Super Admin id=1) — Solomon's
  god-tier API user. Used by the GemineYe backend for cross-system
  integration.

To create a tenant-scoped API user (e.g. for Rule27 to script their
own data): same flow, pick Rule27 as the company.

### Token storage

Don't commit tokens. Use env vars:
```
LISTMONK_API_USER=gemineye-api
LISTMONK_API_TOKEN=<paste here>
```

Then in scripts:
```bash
curl -u "$LISTMONK_API_USER:$LISTMONK_API_TOKEN" "$LISTMONK_URL/api/lists"
```

---

## Response shape

All `/api/*` JSON responses wrap the payload in `data`:

```json
{
  "data": {
    "results": [...],
    "total": 25,
    "page": 1,
    "per_page": 20
  }
}
```

For single-resource endpoints, `data` is the resource directly:
```json
{ "data": { "id": 1, "name": "Newsletter", ... } }
```

Errors:
```json
{ "message": "List not found" }
```

with HTTP 4xx/5xx status.

---

## Pagination

Most list endpoints accept:
- `?page=1` (1-indexed)
- `?per_page=20` (default 20, max varies)
- `?order=asc|desc`
- `?order_by=field`
- `?query=substring` (free-text search where supported)

The response includes `total` for client-side pagination.

---

## Most-used endpoints

### Subscribers

```
GET    /api/subscribers?per_page=100
GET    /api/subscribers/:id
GET    /api/subscribers/by-email/:email
POST   /api/subscribers                                    # body: {email,name,attribs,lists,status}
PUT    /api/subscribers/:id
DELETE /api/subscribers/:id
POST   /api/subscribers/import
POST   /api/subscribers/:id/blocklist
```

### Lists

```
GET    /api/lists
GET    /api/lists/:id
POST   /api/lists                                          # body: {name,type,optin,tags,description}
PUT    /api/lists/:id
DELETE /api/lists/:id
```

### Campaigns

```
GET    /api/campaigns
GET    /api/campaigns/:id
POST   /api/campaigns                                      # body: see schema
PUT    /api/campaigns/:id
PUT    /api/campaigns/:id/status                           # {status:"running"|"paused"|"cancelled"}
DELETE /api/campaigns/:id
GET    /api/campaigns/:id/preview
POST   /api/campaigns/:id/test                             # send test to specified emails
```

### Templates

```
GET    /api/templates?no_body=true
GET    /api/templates/:id
POST   /api/templates
PUT    /api/templates/:id
PUT    /api/templates/:id/default
DELETE /api/templates/:id
GET    /api/templates/:id/preview
```

### Transactional

```
POST   /api/tx                                             # send a transactional email
```

Body shape:
```json
{
  "subscriber_email": "alice@example.com",
  "template_id": 14,
  "from_email": "support@solomontech.co",
  "messenger": "email-resend-solomontech",
  "data": { "name": "Alice", "reset_link": "..." }
}
```

### Companies (Solomon fork)

```
GET    /api/companies                                      # list, filterable view
GET    /api/companies/stats                                # list + per-tenant row counts (settings:manage)
GET    /api/companies/:id
POST   /api/companies                                      # platform admin only
PUT    /api/companies/:id                                  # platform admin only
DELETE /api/companies/:id                                  # platform admin only; FK-blocked if data exists
```

Body for create/update:
```json
{ "name": "Acme Co", "slug": "acme-co" }
```

### Drips (Solomon fork)

```
GET    /api/drips
GET    /api/drips/:id
POST   /api/drips
PUT    /api/drips/:id
PUT    /api/drips/:id/status
DELETE /api/drips/:id

# Steps (within a drip):
GET    /api/drips/:id/steps
POST   /api/drips/:id/steps
PUT    /api/drips/:drip_id/steps/:id
DELETE /api/drips/:drip_id/steps/:id

# Enrollments:
GET    /api/drips/:id/enrollments
POST   /api/drips/:id/enroll                               # body: {subscriber_ids:[...]}
```

### Warming (Solomon fork)

```
GET    /api/warming/addresses
POST   /api/warming/addresses
PUT    /api/warming/addresses/:id
DELETE /api/warming/addresses/:id

GET    /api/warming/senders
POST   /api/warming/senders
...

GET    /api/warming/templates
GET    /api/warming/campaigns
POST   /api/warming/campaigns
GET    /api/warming/send-log
GET    /api/warming/sends-today                            # quick today-count
GET    /api/warming/config                                 # singleton config (legacy global)
```

### Dashboard

```
GET    /api/dashboard/counts                               # subs/lists/campaigns/messages, scoped to your tenant
GET    /api/dashboard/charts                               # 30-day clicks + views chart
GET    /api/dashboard/features                             # drips/automations/segments/scoring/deals/webhooks/warming
```

### Server config (frontend boot)

```
GET    /api/config                                         # site name, lang, messengers (filtered to tenant)
GET    /api/health                                         # liveness — always returns {"data": true}
```

---

## Working examples

### Add a subscriber to a list

```bash
curl -u "$USER:$TOKEN" -X POST "$URL/api/subscribers" \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "alice@example.com",
    "name": "Alice",
    "attribs": {"plan": "trial"},
    "lists": [3],
    "status": "enabled"
  }'
```

### Bulk import via CSV

```bash
curl -u "$USER:$TOKEN" -X POST "$URL/api/subscribers/import" \
  -F 'params={"mode":"subscribe","subscription_status":"confirmed","lists":[3]}' \
  -F 'file=@subscribers.csv'
```

### Trigger a transactional email

```bash
curl -u "$USER:$TOKEN" -X POST "$URL/api/tx" \
  -H 'Content-Type: application/json' \
  -d '{
    "subscriber_email": "alice@example.com",
    "template_id": 14,
    "messenger": "email-resend-solomontech",
    "data": {"reset_link": "https://app.example.com/reset/abc123"}
  }'
```

### List all of your tenant's campaigns

```bash
curl -u "$USER:$TOKEN" "$URL/api/campaigns?per_page=100" | jq '.data.results[] | {id, name, status, messenger}'
```

### Create a new tenant (platform admin only)

```bash
curl -u "alch3my:..." -X POST "$URL/api/companies" \
  -H 'Content-Type: application/json' \
  -d '{"name": "AnilTX Standalone", "slug": "aniltx"}'
```

Then create a role under it, then a user.

### Enroll a subscriber in a drip

```bash
curl -u "$USER:$TOKEN" -X POST "$URL/api/drips/5/enroll" \
  -H 'Content-Type: application/json' \
  -d '{"subscriber_ids": [12345]}'
```

### Send a test email of a campaign

```bash
curl -u "$USER:$TOKEN" -X POST "$URL/api/campaigns/123/test" \
  -H 'Content-Type: application/json' \
  -d '{"emails": ["alch3my@solomontech.co"]}'
```

---

## Bounce webhooks (incoming)

If you configure bounce processing via webhook (instead of mailbox
scanning), other systems POST bounce events to:

```
POST /webhooks/bounce?service=ses        # AWS SES
POST /webhooks/bounce?service=sendgrid   # SendGrid
POST /webhooks/bounce?service=postmark   # Postmark
POST /webhooks/bounce?service=forwardemail
POST /webhooks/bounce?service=resend     # Solomon fork
```

The `resend` endpoint expects Svix-style signed headers
(`svix-id`, `svix-timestamp`, `svix-signature`). Sign in the Resend
dashboard with the secret configured in
`bounce.resend.signing_key` (config.toml or
`LISTMONK_bounce__resend__signing_key` env var).

---

## Outbound webhooks

Listmonk can fire webhooks on certain events. Sidebar → Webhooks → + New:

- Events: `subscriber.created`, `subscriber.updated`, `campaign.sent`,
  `campaign.failed`, `bounce.recorded`, `score.changed`, etc.
- URL: where to POST.
- Secret: HMAC-SHA256 secret; signature is in `X-Listmonk-Signature`.
- Max retries / timeout configurable.

Tenant-scoped — Rule27's webhooks fire on Rule27 events; Solomon's
on Solomon's. Cross-tenant events not propagated.

---

## Rate limiting

No global rate limiting at the API layer. Be reasonable. The
`/api/subscribers/import` endpoint blocks while the CSV is being
processed; large imports can take minutes — set client timeouts
generously.

---

## SDK / language bindings

None official. The API is straightforward REST. For Go projects, the
Solomon org may publish a thin client wrapper later — see
[https://github.com/Solomon-Technologies](https://github.com/Solomon-Technologies)
for any updates.
