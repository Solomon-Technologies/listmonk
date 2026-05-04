-- automations

-- name: create-automation
-- $6 = company_id (v7.17.0); 0 falls back to Solomon=1.
INSERT INTO automations (uuid, name, description, status, canvas, company_id)
    VALUES($1, $2, $3, $4, $5, COALESCE(NULLIF($6::INT, 0), 1)) RETURNING id;

-- name: query-automations
SELECT COUNT(*) OVER () AS total, automations.* FROM automations WHERE
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

-- name: get-automation
SELECT * FROM automations WHERE
    CASE WHEN $1 > 0 THEN id = $1 ELSE uuid = $2::UUID END
    AND ($3::INT = 0 OR company_id = $3::INT);

-- name: update-automation
UPDATE automations SET
    name=(CASE WHEN $2 != '' THEN $2 ELSE name END),
    description=$3,
    status=(CASE WHEN $4 != '' THEN $4 ELSE status END),
    canvas=$5,
    updated_at=NOW()
WHERE id = $1 RETURNING id;

-- name: update-automation-status
UPDATE automations SET status=$2, updated_at=NOW() WHERE id = $1;

-- name: delete-automation
DELETE FROM automations WHERE id = $1;

-- name: create-automation-node
INSERT INTO automation_nodes (uuid, automation_id, node_type, config, position_x, position_y)
    VALUES($1, $2, $3, $4, $5, $6) RETURNING id;

-- name: get-automation-nodes
SELECT * FROM automation_nodes WHERE automation_id = $1 ORDER BY id ASC;

-- name: get-automation-node
SELECT * FROM automation_nodes WHERE id = $1;

-- name: update-automation-node
UPDATE automation_nodes SET
    node_type=$2, config=$3, position_x=$4, position_y=$5
WHERE id = $1 RETURNING id;

-- name: delete-automation-node
DELETE FROM automation_nodes WHERE id = $1;

-- name: create-automation-edge
INSERT INTO automation_edges (automation_id, from_node_id, to_node_id, label)
    VALUES($1, $2, $3, $4) ON CONFLICT DO NOTHING RETURNING id;

-- name: get-automation-edges
SELECT * FROM automation_edges WHERE automation_id = $1;

-- name: delete-automation-edge
DELETE FROM automation_edges WHERE id = $1;

-- name: delete-automation-edges-by-automation
DELETE FROM automation_edges WHERE automation_id = $1;

-- name: get-pending-automation-enrollments
SELECT e.id AS enrollment_id, e.automation_id, e.subscriber_id, e.current_node_id,
    s.email AS subscriber_email, s.name AS subscriber_name, s.uuid AS subscriber_uuid,
    s.attribs AS subscriber_attribs
FROM automation_enrollments e
JOIN subscribers s ON s.id = e.subscriber_id
WHERE e.status = 'active'
    AND (e.wait_until IS NULL OR e.wait_until <= NOW())
    AND s.status != 'blocklisted'
LIMIT $1;

-- name: enroll-in-automation
INSERT INTO automation_enrollments (automation_id, subscriber_id, current_node_id, status)
    VALUES($1, $2, $3, 'active')
    ON CONFLICT (automation_id, subscriber_id) DO NOTHING
    RETURNING id;

-- name: update-automation-enrollment
UPDATE automation_enrollments SET
    current_node_id=$2, status=$3, wait_until=$4,
    completed_at=(CASE WHEN $3 = 'completed' THEN NOW() ELSE NULL END)
WHERE id = $1;
