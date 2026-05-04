-- A/B tests

-- name: create-ab-test
-- $8 = company_id (v7.17.0); 0 falls back to Solomon=1.
INSERT INTO ab_tests (uuid, campaign_id, test_type, status, test_percentage, winner_metric, winner_wait_hours, company_id)
    VALUES($1, $2, $3, $4, $5, $6, $7, COALESCE(NULLIF($8::INT, 0), 1)) RETURNING id;

-- name: get-ab-test
SELECT * FROM ab_tests WHERE
    CASE WHEN $1 > 0 THEN id = $1 ELSE uuid = $2::UUID END
    AND ($3::INT = 0 OR company_id = $3::INT);

-- name: get-ab-test-by-campaign
SELECT * FROM ab_tests WHERE campaign_id = $1
    AND ($2::INT = 0 OR company_id = $2::INT);

-- name: update-ab-test
UPDATE ab_tests SET
    test_type=$2, test_percentage=$3, winner_metric=$4, winner_wait_hours=$5,
    updated_at=NOW()
WHERE id = $1 RETURNING id;

-- name: update-ab-test-status
UPDATE ab_tests SET status=$2,
    started_at=(CASE WHEN $2 = 'running' AND started_at IS NULL THEN NOW() ELSE started_at END),
    finished_at=(CASE WHEN $2 = 'finished' THEN NOW() ELSE finished_at END),
    winning_variant_id=(CASE WHEN $3 > 0 THEN $3 ELSE winning_variant_id END),
    updated_at=NOW()
WHERE id = $1;

-- name: delete-ab-test
DELETE FROM ab_tests WHERE id = $1;

-- name: create-ab-variant
INSERT INTO ab_test_variants (ab_test_id, label, subject, body, from_email)
    VALUES($1, $2, $3, $4, $5) RETURNING id;

-- name: get-ab-variants
SELECT * FROM ab_test_variants WHERE ab_test_id = $1 ORDER BY label ASC;

-- name: get-ab-variant
SELECT * FROM ab_test_variants WHERE id = $1;

-- name: update-ab-variant
UPDATE ab_test_variants SET
    label=$2, subject=$3, body=$4, from_email=$5
WHERE id = $1 RETURNING id;

-- name: update-ab-variant-counts
UPDATE ab_test_variants SET sent=$2, opened=$3, clicked=$4, bounced=$5 WHERE id = $1;

-- name: delete-ab-variant
DELETE FROM ab_test_variants WHERE id = $1;

-- name: assign-subscriber-to-variant
INSERT INTO ab_test_assignments (ab_test_id, variant_id, subscriber_id)
    VALUES($1, $2, $3) ON CONFLICT (ab_test_id, subscriber_id) DO NOTHING;

-- name: get-subscriber-variant
SELECT variant_id FROM ab_test_assignments WHERE ab_test_id = $1 AND subscriber_id = $2;

-- name: get-running-ab-tests
SELECT * FROM ab_tests WHERE status = 'running'
    AND started_at + (winner_wait_hours * INTERVAL '1 hour') <= NOW();
