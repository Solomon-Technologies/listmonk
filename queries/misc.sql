-- name: get-dashboard-charts
SELECT data FROM mat_dashboard_charts;

-- name: get-dashboard-counts
SELECT data FROM mat_dashboard_counts;

-- name: get-settings
SELECT JSON_OBJECT_AGG(key, value) AS settings FROM (SELECT * FROM settings ORDER BY key) t;

-- name: update-settings
UPDATE settings AS s SET value = c.value
    -- For each key in the incoming JSON map, update the row with the key and its value.
    FROM(SELECT * FROM JSONB_EACH($1)) AS c(key, value) WHERE s.key = c.key;

-- name: update-settings-by-key
UPDATE settings SET value = $2, updated_at = NOW() WHERE key = $1;

-- name: get-dashboard-feature-counts
SELECT JSON_BUILD_OBJECT(
    'drips', JSON_BUILD_OBJECT(
        'total', (SELECT COUNT(*) FROM drip_campaigns),
        'active', (SELECT COUNT(*) FROM drip_campaigns WHERE status = 'active')
    ),
    'automations', JSON_BUILD_OBJECT(
        'total', (SELECT COUNT(*) FROM automations),
        'active', (SELECT COUNT(*) FROM automations WHERE status = 'active')
    ),
    'segments', (SELECT COUNT(*) FROM segments),
    'scoring_rules', (SELECT COUNT(*) FROM scoring_rules),
    'deals', JSON_BUILD_OBJECT(
        'total', (SELECT COUNT(*) FROM deals),
        'open', (SELECT COUNT(*) FROM deals WHERE status = 'open')
    ),
    'webhooks', JSON_BUILD_OBJECT(
        'total', (SELECT COUNT(*) FROM webhooks),
        'active', (SELECT COUNT(*) FROM webhooks WHERE enabled = true)
    )
) AS data;

-- name: get-db-info
SELECT JSON_BUILD_OBJECT('version', (SELECT VERSION()),
                        'size_mb', (SELECT ROUND(pg_database_size((SELECT CURRENT_DATABASE()))/(1024^2)))) AS info;
