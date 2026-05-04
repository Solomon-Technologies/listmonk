-- contact scoring

-- name: create-scoring-rule
-- $6 = company_id (v7.17.0); 0 falls back to Solomon=1.
INSERT INTO scoring_rules (name, enabled, event_type, score_value, conditions, company_id)
    VALUES($1, $2, $3, $4, $5, COALESCE(NULLIF($6::INT, 0), 1)) RETURNING id;

-- name: get-scoring-rules
-- $1 = company_id (v7.17.0); 0 disables filter.
SELECT * FROM scoring_rules
    WHERE ($1::INT = 0 OR company_id = $1::INT)
    ORDER BY event_type, name;

-- name: get-scoring-rule
SELECT * FROM scoring_rules WHERE id = $1
    AND ($2::INT = 0 OR company_id = $2::INT);

-- name: get-scoring-rules-by-event
SELECT * FROM scoring_rules WHERE enabled = true AND event_type = $1;

-- name: update-scoring-rule
UPDATE scoring_rules SET
    name=$2, enabled=$3, event_type=$4, score_value=$5, conditions=$6, updated_at=NOW()
WHERE id = $1 RETURNING id;

-- name: delete-scoring-rule
DELETE FROM scoring_rules WHERE id = $1;

-- name: update-subscriber-score
UPDATE subscribers SET score = GREATEST(0, score + $2) WHERE id = $1 RETURNING score;

-- name: insert-score-log
INSERT INTO score_log (subscriber_id, rule_id, event_type, score_change, score_after, meta)
    VALUES($1, $2, $3, $4, $5, $6);

-- name: get-subscriber-score-log
SELECT * FROM score_log WHERE subscriber_id = $1 ORDER BY created_at DESC
    OFFSET $2 LIMIT (CASE WHEN $3 < 1 THEN NULL ELSE $3 END);

-- name: decay-inactive-scores
WITH inactive AS (
    SELECT s.id FROM subscribers s
    WHERE s.score > 0
    AND s.id NOT IN (
        SELECT DISTINCT subscriber_id FROM score_log
        WHERE created_at > NOW() - INTERVAL '30 days'
        AND event_type IN ('email.opened', 'email.clicked')
    )
)
UPDATE subscribers SET score = GREATEST(0, score - $1)
WHERE id IN (SELECT id FROM inactive)
RETURNING id, score;
