-- name: get-dashboard-charts
-- Multi-tenant (v7.17.0+): $1 = company_id; 0 disables filter (platform admin sees global).
-- Replaced the mat_dashboard_charts read with an on-the-fly aggregate so
-- per-tenant scoping works. Filtered via JOIN to campaigns.company_id.
WITH clicks AS (
    SELECT JSON_AGG(ROW_TO_JSON(row))
    FROM (
        WITH viewDates AS (
          SELECT lc.created_at::DATE AS to_date,
                 lc.created_at::DATE - INTERVAL '30 DAY' AS from_date
                 FROM link_clicks lc
                 LEFT JOIN campaigns c ON c.id = lc.campaign_id
                 WHERE ($1::INT = 0 OR c.company_id = $1::INT)
                 ORDER BY lc.id DESC LIMIT 1
        )
        SELECT COUNT(*) AS count, lc.created_at::DATE as date FROM link_clicks lc
          LEFT JOIN campaigns c ON c.id = lc.campaign_id
          WHERE lc.created_at >= (SELECT from_date FROM viewDates)
            AND lc.created_at < (SELECT to_date FROM viewDates) + INTERVAL '1 day'
            AND ($1::INT = 0 OR c.company_id = $1::INT)
          GROUP by date ORDER BY date
    ) row
),
views AS (
    SELECT JSON_AGG(ROW_TO_JSON(row))
    FROM (
        WITH viewDates AS (
          SELECT cv.created_at::DATE AS to_date,
                 cv.created_at::DATE - INTERVAL '30 DAY' AS from_date
                 FROM campaign_views cv
                 LEFT JOIN campaigns c ON c.id = cv.campaign_id
                 WHERE ($1::INT = 0 OR c.company_id = $1::INT)
                 ORDER BY cv.id DESC LIMIT 1
        )
        SELECT COUNT(*) AS count, cv.created_at::DATE as date FROM campaign_views cv
          LEFT JOIN campaigns c ON c.id = cv.campaign_id
          WHERE cv.created_at >= (SELECT from_date FROM viewDates)
            AND cv.created_at < (SELECT to_date FROM viewDates) + INTERVAL '1 day'
            AND ($1::INT = 0 OR c.company_id = $1::INT)
          GROUP by date ORDER BY date
    ) row
)
SELECT JSON_BUILD_OBJECT('link_clicks', COALESCE((SELECT * FROM clicks), '[]'),
                        'campaign_views', COALESCE((SELECT * FROM views), '[]')
                       ) AS data;

-- name: get-dashboard-counts
-- Multi-tenant (v7.17.0+): $1 = company_id; 0 disables filter.
WITH subs AS (
    SELECT COUNT(*) AS num, status FROM subscribers
    WHERE ($1::INT = 0 OR company_id = $1::INT)
    GROUP BY status
)
SELECT JSON_BUILD_OBJECT(
    'subscribers', JSON_BUILD_OBJECT(
        'total', COALESCE((SELECT SUM(num) FROM subs), 0),
        'blocklisted', COALESCE((SELECT num FROM subs WHERE status='blocklisted'), 0),
        'orphans', (
            SELECT COUNT(s.id) FROM subscribers s
            LEFT JOIN subscriber_lists ON (s.id = subscriber_lists.subscriber_id)
            WHERE subscriber_lists.subscriber_id IS NULL
              AND ($1::INT = 0 OR s.company_id = $1::INT)
        )
    ),
    'lists', JSON_BUILD_OBJECT(
        'total', (SELECT COUNT(*) FROM lists WHERE ($1::INT = 0 OR company_id = $1::INT)),
        'private', (SELECT COUNT(*) FROM lists WHERE type='private' AND ($1::INT = 0 OR company_id = $1::INT)),
        'public', (SELECT COUNT(*) FROM lists WHERE type='public' AND ($1::INT = 0 OR company_id = $1::INT)),
        'optin_single', (SELECT COUNT(*) FROM lists WHERE optin='single' AND ($1::INT = 0 OR company_id = $1::INT)),
        'optin_double', (SELECT COUNT(*) FROM lists WHERE optin='double' AND ($1::INT = 0 OR company_id = $1::INT))
    ),
    'campaigns', JSON_BUILD_OBJECT(
        'total', (SELECT COUNT(*) FROM campaigns WHERE ($1::INT = 0 OR company_id = $1::INT)),
        'by_status', COALESCE((
            SELECT JSON_OBJECT_AGG (status, num) FROM
            (SELECT status, COUNT(*) AS num FROM campaigns WHERE ($1::INT = 0 OR company_id = $1::INT) GROUP BY status) r
        ), '{}'::JSON)
    ),
    'messages', COALESCE((SELECT SUM(sent) FROM campaigns WHERE ($1::INT = 0 OR company_id = $1::INT)), 0)
) AS data;

-- name: get-settings
SELECT JSON_OBJECT_AGG(key, value) AS settings FROM (SELECT * FROM settings ORDER BY key) t;

-- name: update-settings
UPDATE settings AS s SET value = c.value
    -- For each key in the incoming JSON map, update the row with the key and its value.
    FROM(SELECT * FROM JSONB_EACH($1)) AS c(key, value) WHERE s.key = c.key;

-- name: update-settings-by-key
UPDATE settings SET value = $2, updated_at = NOW() WHERE key = $1;

-- name: get-dashboard-feature-counts
-- Multi-tenant (v7.17.0+): $1 = company_id; 0 disables filter.
-- warming_send_log inherits company_id via warming_campaigns.id FK.
SELECT JSON_BUILD_OBJECT(
    'drips', JSON_BUILD_OBJECT(
        'total', (SELECT COUNT(*) FROM drip_campaigns WHERE ($1::INT = 0 OR company_id = $1::INT)),
        'active', (SELECT COUNT(*) FROM drip_campaigns WHERE status = 'active' AND ($1::INT = 0 OR company_id = $1::INT))
    ),
    'automations', JSON_BUILD_OBJECT(
        'total', (SELECT COUNT(*) FROM automations WHERE ($1::INT = 0 OR company_id = $1::INT)),
        'active', (SELECT COUNT(*) FROM automations WHERE status = 'active' AND ($1::INT = 0 OR company_id = $1::INT))
    ),
    'segments', (SELECT COUNT(*) FROM segments WHERE ($1::INT = 0 OR company_id = $1::INT)),
    'scoring_rules', (SELECT COUNT(*) FROM scoring_rules WHERE ($1::INT = 0 OR company_id = $1::INT)),
    'deals', JSON_BUILD_OBJECT(
        'total', (SELECT COUNT(*) FROM deals WHERE ($1::INT = 0 OR company_id = $1::INT)),
        'open', (SELECT COUNT(*) FROM deals WHERE status = 'open' AND ($1::INT = 0 OR company_id = $1::INT))
    ),
    'webhooks', JSON_BUILD_OBJECT(
        'total', (SELECT COUNT(*) FROM webhooks WHERE ($1::INT = 0 OR company_id = $1::INT)),
        'active', (SELECT COUNT(*) FROM webhooks WHERE enabled = true AND ($1::INT = 0 OR company_id = $1::INT))
    ),
    'warming', JSON_BUILD_OBJECT(
        'campaigns', (SELECT COUNT(*) FROM warming_campaigns WHERE ($1::INT = 0 OR company_id = $1::INT)),
        'active', (SELECT COUNT(*) FROM warming_campaigns WHERE status = 'active' AND ($1::INT = 0 OR company_id = $1::INT)),
        'sent_today', (
            SELECT COUNT(*) FROM warming_send_log wsl
            LEFT JOIN warming_campaigns wc ON wc.id = wsl.campaign_id
            WHERE wsl.sent_at >= CURRENT_DATE AND wsl.status = 'sent'
              AND ($1::INT = 0 OR wc.company_id = $1::INT)
        ),
        'total_sent', (
            SELECT COUNT(*) FROM warming_send_log wsl
            LEFT JOIN warming_campaigns wc ON wc.id = wsl.campaign_id
            WHERE wsl.status = 'sent'
              AND ($1::INT = 0 OR wc.company_id = $1::INT)
        ),
        'total_errors', (
            SELECT COUNT(*) FROM warming_send_log wsl
            LEFT JOIN warming_campaigns wc ON wc.id = wsl.campaign_id
            WHERE wsl.status = 'failed'
              AND ($1::INT = 0 OR wc.company_id = $1::INT)
        )
    )
) AS data;

-- name: get-db-info
SELECT JSON_BUILD_OBJECT('version', (SELECT VERSION()),
                        'size_mb', (SELECT ROUND(pg_database_size((SELECT CURRENT_DATABASE()))/(1024^2)))) AS info;
