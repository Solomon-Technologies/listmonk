# 07 — Templates

A template is a reusable HTML/Markdown wrapper for email content.
Campaigns and drip steps reference a template; the template renders
around the body content at send time.

---

## Template types

| Type | Used for | Required for |
|---|---|---|
| `campaign` | Wraps regular campaign + drip step bodies | Every regular campaign needs one |
| `campaign_visual` | Visual-editor (drag-drop) variant | Visual editor campaigns |
| `tx` | Transactional emails (password resets, receipts) | `POST /api/tx` calls reference these by name |

The `is_default` flag picks one template per (type, company_id) as the
fallback when a campaign doesn't specify one. Per v7.17.0 each
tenant has its own default.

---

## Building a template

1. Sidebar → Campaigns → Templates → **+ New**.
2. Name: `Solomon Standard`, type: `campaign`.
3. Subject: leave blank for `campaign` type (each campaign sets its
   own); fill for `tx` type.
4. Body: HTML with template tags. Listmonk uses Go's
   `html/template` syntax + Sprig functions. The `{{ template "content" . }}`
   placeholder is where the campaign body gets injected.

Minimal example:

```html
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>{{ .Subject }}</title></head>
<body style="font-family: -apple-system, sans-serif; max-width: 600px; margin: auto;">
  <header><img src="https://solomon.tech/logo.png" alt="Solomon" height="40"></header>
  <main>
    {{ template "content" . }}
  </main>
  <footer style="font-size: 12px; color: #888; padding-top: 40px;">
    <p>Solomon Technologies · Tampa, FL</p>
    <p><a href="{{ UnsubscribeURL }}">Unsubscribe</a> · <a href="{{ ManageURL }}">Manage preferences</a></p>
  </footer>
</body>
</html>
```

5. Save. The new template shows up in the campaign editor's Template
   dropdown.

---

## Available template variables

Inside both the template AND the campaign body:

| Variable | Value |
|---|---|
| `.Subscriber.Email` | Recipient email |
| `.Subscriber.Name` | Recipient name |
| `.Subscriber.UUID` | Subscriber UUID (for tracking links) |
| `.Subscriber.Attribs.<key>` | Any field in subscriber's attribs JSON |
| `.Campaign.Name` | Internal campaign name |
| `.Campaign.Subject` | Campaign subject line |
| `.Campaign.UUID` | Campaign UUID |
| `.Subject` | Resolved subject (after merge fields applied) |
| `UnsubscribeURL` | One-click unsubscribe link (RFC 8058) |
| `ManageURL` | Subscriber preferences page |
| `OptinURL` | Double-opt-in confirmation link |
| `MessageURL` | Web view of this campaign |
| `TrackLink "url"` | Wraps a URL for click tracking |
| `Date "layout"` | Current date formatted via Go time layout |

Plus the entire [Sprig function library](http://masterminds.github.io/sprig/)
(`upper`, `lower`, `default`, `trunc`, etc.).

---

## Tenant scoping

After v7.17.0:
- Each template has a `company_id`.
- Listing templates filters by user's tenant.
- Default template is per-tenant — Solomon and Rule27 each have their
  own default.
- Internal flows (campaign send pipeline pulling the template at send
  time, transactional template cache loading on startup) bypass the
  filter — they trust the campaign's reference and load whichever
  template matches by ID.

To copy a template across tenants: there's no UI for it. SQL:

```sql
INSERT INTO templates (name, type, subject, body, body_source, company_id, is_default)
SELECT name || ' (cloned)', type, subject, body, body_source, 2, false
FROM templates WHERE id = 5;  -- the source template ID
```

---

## Transactional templates (`tx` type)

Transactional templates are for `POST /api/tx` calls — single-recipient
emails like password resets, order confirmations, trial-end reminders.

Setup:
1. Templates → + New, type `tx`. Set name, subject, body (with merge
   fields).
2. Send via API:
```http
POST /api/tx
Authorization: Basic <base64 user:token>
Content-Type: application/json

{
  "subscriber_email": "alice@example.com",
  "template_id": 14,
  "data": {
    "first_name": "Alice",
    "reset_link": "https://app.example.com/reset/abc123"
  },
  "messenger": "email-resend-solomontech"
}
```
3. Listmonk renders `template_id=14` with `data` available as
   `.Tx.Data.first_name`, `.Tx.Data.reset_link` etc., and sends via
   the picked messenger.

The `tx` template body has access to `.Tx.Data` (the JSON you POST)
and the standard subscriber fields.

---

## Visual editor templates

If you don't want to write HTML, use type `campaign_visual`. The body
is built via the drag-drop editor and stored as JSON in `body_source`,
with the rendered HTML in `body`.

Limitations:
- More fragile than hand-written HTML
- Doesn't support all Sprig functions in the visual blocks
- Visual templates can't be `is_default = true` for non-visual campaigns

---

## Common patterns

| Goal | How |
|---|---|
| Standard branded email | One `campaign` template with header logo + footer w/ unsub link |
| Plain-text receipt | One `tx` template, body is `<pre>` or just plain HTML |
| Different layout for different list | Multiple `campaign` templates, pick at campaign creation |
| A/B test layouts | Two campaigns referencing different templates, run as A/B |

---

## Permissions

| Action | Perm |
|---|---|
| View templates | `templates:get` |
| Create/edit/delete | `templates:manage` |
| Set default | `templates:manage` |

Both Tenant Super Admin and Operational Admin have full template
management within their tenant.
