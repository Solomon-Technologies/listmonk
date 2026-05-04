-- webhooks

-- name: create-webhook
-- $9 = company_id (v7.17.0); 0 falls back to Solomon=1.
INSERT INTO webhooks (uuid, name, url, secret, enabled, events, max_retries, timeout_seconds, company_id)
    VALUES($1, $2, $3, $4, $5, $6::TEXT[], $7, $8, COALESCE(NULLIF($9::INT, 0), 1)) RETURNING id;

-- name: query-webhooks
SELECT COUNT(*) OVER () AS total, webhooks.* FROM webhooks WHERE
    CASE
        WHEN $1 > 0 THEN id = $1
        WHEN $2 != '' THEN uuid = $2::UUID
        WHEN $3 != '' THEN (name ILIKE $3)
        ELSE TRUE
    END
    -- Multi-tenant filter (v7.17.0): $6=0 disables.
    AND ($6::INT = 0 OR company_id = $6::INT)
    ORDER BY %order%
    OFFSET $4 LIMIT (CASE WHEN $5 < 1 THEN NULL ELSE $5 END);

-- name: get-webhook
SELECT * FROM webhooks WHERE
    CASE WHEN $1 > 0 THEN id = $1 ELSE uuid = $2::UUID END
    AND ($3::INT = 0 OR company_id = $3::INT);

-- name: update-webhook
UPDATE webhooks SET
    name=(CASE WHEN $2 != '' THEN $2 ELSE name END),
    url=(CASE WHEN $3 != '' THEN $3 ELSE url END),
    secret=$4,
    enabled=$5,
    events=$6::TEXT[],
    max_retries=$7,
    timeout_seconds=$8,
    updated_at=NOW()
WHERE id = $1 RETURNING id;

-- name: delete-webhook
DELETE FROM webhooks WHERE id = $1;

-- name: get-webhooks-by-event
SELECT * FROM webhooks WHERE enabled = true AND $1 = ANY(events);

-- name: insert-webhook-log
INSERT INTO webhook_log (webhook_id, event, payload, response_code, response_body, error, attempt)
    VALUES($1, $2, $3, $4, $5, $6, $7);

-- name: query-webhook-log
SELECT COUNT(*) OVER () AS total, webhook_log.* FROM webhook_log WHERE
    CASE
        WHEN $1 > 0 THEN webhook_id = $1
        ELSE TRUE
    END
    ORDER BY created_at DESC
    OFFSET $2 LIMIT (CASE WHEN $3 < 1 THEN NULL ELSE $3 END);
