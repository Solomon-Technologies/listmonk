-- drip campaigns

-- name: create-drip-campaign
-- $9 = company_id (v7.17.0); 0 falls back to Solomon=1.
INSERT INTO drip_campaigns (uuid, name, description, status, trigger_type, trigger_config, segment_id, from_email, company_id)
    VALUES($1, $2, $3, $4, $5::drip_trigger_type, $6, $7, $8, COALESCE(NULLIF($9::INT, 0), 1)) RETURNING id;

-- name: query-drip-campaigns
SELECT COUNT(*) OVER () AS total, drip_campaigns.* FROM drip_campaigns WHERE
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

-- name: get-drip-campaign
SELECT * FROM drip_campaigns WHERE
    CASE WHEN $1 > 0 THEN id = $1 ELSE uuid = $2::UUID END
    AND ($3::INT = 0 OR company_id = $3::INT);

-- name: update-drip-campaign
UPDATE drip_campaigns SET
    name=(CASE WHEN $2 != '' THEN $2 ELSE name END),
    description=$3,
    status=(CASE WHEN $4 != '' THEN $4::drip_status ELSE status END),
    trigger_type=(CASE WHEN $5 != '' THEN $5::drip_trigger_type ELSE trigger_type END),
    trigger_config=$6,
    segment_id=$7,
    from_email=$8,
    max_send_per_day=$9,
    updated_at=NOW()
WHERE id = $1 RETURNING id;

-- name: update-drip-campaign-status
UPDATE drip_campaigns SET status=$2::drip_status, updated_at=NOW() WHERE id = $1;

-- name: delete-drip-campaign
DELETE FROM drip_campaigns WHERE id = $1;

-- name: update-drip-campaign-counts
UPDATE drip_campaigns SET
    total_entered=$2, total_completed=$3, total_exited=$4,
    updated_at=NOW()
WHERE id = $1;

-- name: create-drip-step
INSERT INTO drip_steps (uuid, drip_campaign_id, sequence_order, delay_value, delay_unit,
    name, subject, from_email, body, alt_body, content_type, template_id, messenger, headers, send_conditions)
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id;

-- name: get-drip-steps
SELECT * FROM drip_steps WHERE drip_campaign_id = $1 ORDER BY sequence_order ASC;

-- name: get-drip-step
SELECT * FROM drip_steps WHERE id = $1;

-- name: update-drip-step
UPDATE drip_steps SET
    sequence_order=$2, delay_value=$3, delay_unit=$4,
    name=$5, subject=$6, from_email=$7, body=$8, alt_body=$9,
    content_type=$10, template_id=$11, messenger=$12, headers=$13, send_conditions=$14,
    updated_at=NOW()
WHERE id = $1 RETURNING id;

-- name: delete-drip-step
DELETE FROM drip_steps WHERE id = $1;

-- name: update-drip-step-counts
UPDATE drip_steps SET sent=$2, opened=$3, clicked=$4, updated_at=NOW() WHERE id = $1;

-- name: enroll-subscriber-in-drip
INSERT INTO drip_enrollments (drip_campaign_id, subscriber_id, status, current_step_id, next_send_at)
    VALUES($1, $2, 'active', $3, $4)
    ON CONFLICT (drip_campaign_id, subscriber_id) DO NOTHING
    RETURNING id;

-- name: get-pending-drip-sends
SELECT e.id AS enrollment_id, e.drip_campaign_id, e.subscriber_id, e.current_step_id,
    s.email AS subscriber_email, s.name AS subscriber_name, s.uuid AS subscriber_uuid,
    s.attribs AS subscriber_attribs, s.status AS subscriber_status,
    ds.subject, ds.body, ds.alt_body, ds.from_email AS step_from_email,
    ds.content_type, ds.template_id, ds.messenger, ds.headers,
    ds.uuid AS step_uuid,
    dc.from_email AS campaign_from_email, dc.name AS campaign_name,
    dc.uuid AS campaign_uuid, dc.max_send_per_day,
    COALESCE(t.body, '') AS template_body
FROM drip_enrollments e
JOIN subscribers s ON s.id = e.subscriber_id
JOIN drip_steps ds ON ds.id = e.current_step_id
JOIN drip_campaigns dc ON dc.id = e.drip_campaign_id
LEFT JOIN templates t ON t.id = ds.template_id
WHERE e.status = 'active'
    AND e.next_send_at <= NOW()
    AND s.status != 'blocklisted'
    AND dc.status = 'active'
ORDER BY e.next_send_at ASC
LIMIT $1;

-- name: advance-drip-enrollment
-- After sending the current step, advance to the next step or mark complete.
-- $1=enrollment_id, $2=next_step_id (NULL if complete), $3=next_send_at (NULL if complete), $4=status ('active' or 'completed')
UPDATE drip_enrollments SET
    current_step_id=$2,
    next_send_at=$3,
    status=$4,
    completed_at=(CASE WHEN $4 = 'completed' THEN NOW() ELSE NULL END)
WHERE id = $1;

-- name: exit-drip-enrollment
UPDATE drip_enrollments SET status='exited', completed_at=NOW() WHERE id = $1;

-- name: get-drip-enrollments
SELECT COUNT(*) OVER () AS total, e.*, s.email AS subscriber_email, s.name AS subscriber_name
FROM drip_enrollments e
JOIN subscribers s ON s.id = e.subscriber_id
WHERE e.drip_campaign_id = $1
ORDER BY e.entered_at DESC
OFFSET $2 LIMIT (CASE WHEN $3 < 1 THEN NULL ELSE $3 END);

-- name: get-drip-enrollment-count
SELECT COUNT(*) FROM drip_enrollments WHERE drip_campaign_id = $1 AND status = $2;

-- name: insert-drip-send-log
INSERT INTO drip_send_log (drip_campaign_id, drip_step_id, subscriber_id, status, error_message)
    VALUES($1, $2, $3, $4, $5);

-- name: get-active-drips-by-trigger
SELECT * FROM drip_campaigns WHERE status = 'active' AND trigger_type = $1::drip_trigger_type;

-- name: get-drip-sends-today
SELECT COUNT(*) FROM drip_send_log WHERE drip_campaign_id = $1 AND sent_at >= CURRENT_DATE AND status = 'sent';

-- name: update-drip-step-sent
UPDATE drip_steps SET sent = sent + 1, updated_at = NOW() WHERE id = $1;

-- name: update-drip-step-opened
UPDATE drip_steps SET opened = opened + 1, updated_at = NOW() WHERE id = $1;

-- name: update-drip-step-clicked
UPDATE drip_steps SET clicked = clicked + 1, updated_at = NOW() WHERE id = $1;

-- name: update-drip-campaign-entered
UPDATE drip_campaigns SET total_entered = total_entered + 1, updated_at = NOW() WHERE id = $1;

-- name: update-drip-campaign-completed
UPDATE drip_campaigns SET total_completed = total_completed + 1, updated_at = NOW() WHERE id = $1;

-- name: bulk-enroll-in-drip
INSERT INTO drip_enrollments (drip_campaign_id, subscriber_id, status, current_step_id, next_send_at)
    SELECT $1, unnest($4::INT[]), 'active', $2, $3
    ON CONFLICT (drip_campaign_id, subscriber_id) DO NOTHING;

-- name: get-drip-campaign-by-uuid
SELECT * FROM drip_campaigns WHERE uuid = $1::UUID;

-- name: get-drip-step-by-uuid
SELECT * FROM drip_steps WHERE uuid = $1::UUID;
