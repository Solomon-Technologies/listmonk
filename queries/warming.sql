-- warming

-- name: get-warming-addresses
SELECT * FROM warming_addresses ORDER BY created_at DESC;

-- name: create-warming-address
INSERT INTO warming_addresses (email, name) VALUES($1, $2) RETURNING id;

-- name: update-warming-address
UPDATE warming_addresses SET
    email = $2, name = $3, is_active = $4, updated_at = NOW()
WHERE id = $1;

-- name: delete-warming-address
DELETE FROM warming_addresses WHERE id = $1;

-- name: get-active-warming-addresses
SELECT * FROM warming_addresses WHERE is_active = true;

-- name: get-warming-senders
SELECT * FROM warming_senders ORDER BY created_at DESC;

-- name: create-warming-sender
INSERT INTO warming_senders (email, name, brand, brand_url, brand_color)
    VALUES($1, $2, $3, $4, $5) RETURNING id;

-- name: update-warming-sender
UPDATE warming_senders SET
    email = $2, name = $3, brand = $4, brand_url = $5, brand_color = $6,
    is_active = $7, updated_at = NOW()
WHERE id = $1;

-- name: delete-warming-sender
DELETE FROM warming_senders WHERE id = $1;

-- name: get-active-warming-senders
SELECT * FROM warming_senders WHERE is_active = true;

-- name: get-warming-templates
SELECT * FROM warming_templates ORDER BY created_at DESC;

-- name: create-warming-template
INSERT INTO warming_templates (subject, body) VALUES($1, $2) RETURNING id;

-- name: update-warming-template
UPDATE warming_templates SET
    subject = $2, body = $3, is_active = $4, updated_at = NOW()
WHERE id = $1;

-- name: delete-warming-template
DELETE FROM warming_templates WHERE id = $1;

-- name: get-active-warming-templates
SELECT * FROM warming_templates WHERE is_active = true;

-- name: get-warming-config
SELECT * FROM warming_config WHERE id = 1;

-- name: update-warming-config
UPDATE warming_config SET
    sends_per_run = $1, runs_per_day = $2, schedule_times = $3,
    random_delay_min_s = $4, random_delay_max_s = $5, is_active = $6,
    updated_at = NOW()
WHERE id = 1;

-- name: get-warming-sends-today
SELECT COUNT(*) FROM warming_send_log
WHERE sent_at >= CURRENT_DATE AND status = 'sent';

-- name: insert-warming-send-log
INSERT INTO warming_send_log (sender_email, recipient_email, template_id, subject, status, error_message)
    VALUES($1, $2, $3, $4, $5, $6);

-- name: get-warming-send-log
-- raw
SELECT l.*, COALESCE(wc.name, '') AS campaign_name
FROM warming_send_log l
LEFT JOIN warming_campaigns wc ON wc.id = l.campaign_id
ORDER BY l.sent_at DESC LIMIT $1 OFFSET $2;

-- name: get-warming-send-log-count
SELECT COUNT(*) FROM warming_send_log;

-- name: get-warming-send-log-by-campaign
-- raw
SELECT l.*, COALESCE(wc.name, '') AS campaign_name
FROM warming_send_log l
LEFT JOIN warming_campaigns wc ON wc.id = l.campaign_id
WHERE ($3 = 0 OR l.campaign_id = $3)
ORDER BY l.sent_at DESC LIMIT $1 OFFSET $2;

-- name: get-warming-send-log-count-by-campaign
SELECT COUNT(*) FROM warming_send_log WHERE ($1 = 0 OR campaign_id = $1);

-- Warming campaigns

-- name: get-warming-campaigns
SELECT * FROM warming_campaigns ORDER BY created_at DESC;

-- name: create-warming-campaign
INSERT INTO warming_campaigns (name, brand, sender_domains, status, sends_per_run, runs_per_day, schedule_times, random_delay_min_s, random_delay_max_s, warmup_start_date, daily_limits, hourly_cap, business_hours_only, sender_id, messenger, recipient_ids)
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING id;

-- name: update-warming-campaign
UPDATE warming_campaigns SET
    name = $2, brand = $3, sender_domains = $4, status = $5,
    sends_per_run = $6, runs_per_day = $7, schedule_times = $8,
    random_delay_min_s = $9, random_delay_max_s = $10,
    warmup_start_date = $11, daily_limits = $12, hourly_cap = $13,
    business_hours_only = $14, sender_id = $15, messenger = $16,
    recipient_ids = $17, updated_at = NOW()
WHERE id = $1;

-- name: delete-warming-campaign
DELETE FROM warming_campaigns WHERE id = $1;

-- name: get-active-warming-campaigns
SELECT * FROM warming_campaigns WHERE status = 'active';

-- name: get-warming-senders-by-domains
SELECT * FROM warming_senders WHERE is_active = true
    AND SPLIT_PART(email, '@', 2) = ANY($1);

-- name: get-warming-sends-today-by-campaign
SELECT COUNT(*) FROM warming_send_log
WHERE campaign_id = $1 AND sent_at >= CURRENT_DATE AND status = 'sent';

-- name: insert-warming-send-log-campaign
INSERT INTO warming_send_log (campaign_id, sender_email, recipient_email, template_id, subject, status, error_message)
    VALUES($1, $2, $3, $4, $5, $6, $7);

-- name: get-warming-sends-last-hour-by-campaign
SELECT COUNT(*) FROM warming_send_log
WHERE campaign_id = $1 AND sent_at >= NOW() - INTERVAL '1 hour' AND status = 'sent';

-- name: set-warming-campaign-start-date
UPDATE warming_campaigns SET warmup_start_date = CURRENT_DATE WHERE id = $1 AND warmup_start_date IS NULL;

-- name: get-warming-campaign-stats-by-id
SELECT
    COUNT(*) FILTER (WHERE sent_at >= CURRENT_DATE AND status = 'sent') AS sent_today,
    COUNT(*) FILTER (WHERE sent_at >= CURRENT_DATE AND status = 'failed') AS errors_today,
    COUNT(*) FILTER (WHERE status = 'sent') AS total_sent,
    COUNT(*) FILTER (WHERE status = 'failed') AS total_errors
FROM warming_send_log WHERE campaign_id = $1;

-- name: get-warming-sender-by-id
SELECT * FROM warming_senders WHERE id = $1 AND is_active = true;
