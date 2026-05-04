-- crm deals and activities

-- name: create-deal
-- $11 = company_id (v7.17.0); 0 falls back to Solomon=1.
INSERT INTO deals (uuid, subscriber_id, name, value, currency, status, stage, expected_close, notes, attribs, company_id)
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, COALESCE(NULLIF($11::INT, 0), 1)) RETURNING id;

-- name: query-deals
SELECT COUNT(*) OVER () AS total, deals.* FROM deals WHERE
    CASE
        WHEN $1 > 0 THEN subscriber_id = $1
        ELSE TRUE
    END
    AND (CASE WHEN $2 != '' THEN status = $2 ELSE TRUE END)
    -- Multi-tenant filter (v7.17.0): $5=0 disables.
    AND ($5::INT = 0 OR company_id = $5::INT)
    ORDER BY created_at DESC
    OFFSET $3 LIMIT (CASE WHEN $4 < 1 THEN NULL ELSE $4 END);

-- name: get-deal
SELECT * FROM deals WHERE
    CASE WHEN $1 > 0 THEN id = $1 ELSE uuid = $2::UUID END
    AND ($3::INT = 0 OR company_id = $3::INT);

-- name: update-deal
UPDATE deals SET
    name=(CASE WHEN $2 != '' THEN $2 ELSE name END),
    value=$3, currency=$4, status=$5, stage=$6,
    expected_close=$7, closed_at=$8, notes=$9, attribs=$10,
    updated_at=NOW()
WHERE id = $1 RETURNING id;

-- name: delete-deal
DELETE FROM deals WHERE id = $1;

-- name: get-deal-pipeline
SELECT status, stage, COUNT(*) AS count, SUM(value) AS total_value
FROM deals WHERE status = 'open'
GROUP BY status, stage ORDER BY stage;

-- name: create-activity
INSERT INTO contact_activities (subscriber_id, activity_type, description, meta, created_by)
    VALUES($1, $2, $3, $4, $5) RETURNING id;

-- name: get-subscriber-activities
SELECT COUNT(*) OVER () AS total, ca.*, u.username AS created_by_name
FROM contact_activities ca
LEFT JOIN users u ON u.id = ca.created_by
WHERE ca.subscriber_id = $1
ORDER BY ca.created_at DESC
OFFSET $2 LIMIT (CASE WHEN $3 < 1 THEN NULL ELSE $3 END);

-- name: delete-activity
DELETE FROM contact_activities WHERE id = $1;
