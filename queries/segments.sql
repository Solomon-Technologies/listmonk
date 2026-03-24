-- segments

-- name: create-segment
INSERT INTO segments (uuid, name, description, match_type, conditions, tags)
    VALUES($1, $2, $3, $4, $5, $6) RETURNING id;

-- name: query-segments
SELECT COUNT(*) OVER () AS total, segments.* FROM segments WHERE
    CASE
        WHEN $1 > 0 THEN id = $1
        WHEN $2 != '' THEN uuid = $2::UUID
        WHEN $3 != '' THEN (TO_TSVECTOR(name) @@ TO_TSQUERY ($3) OR name ILIKE $3)
        ELSE TRUE
    END
    AND (CARDINALITY($4::VARCHAR(100)[]) = 0 OR $4 <@ tags)
    ORDER BY %order%
    OFFSET $5 LIMIT (CASE WHEN $6 < 1 THEN NULL ELSE $6 END);

-- name: get-segment
SELECT * FROM segments WHERE
    CASE WHEN $1 > 0 THEN id = $1 ELSE uuid = $2::UUID END;

-- name: update-segment
UPDATE segments SET
    name=(CASE WHEN $2 != '' THEN $2 ELSE name END),
    description=$3,
    match_type=(CASE WHEN $4 != '' THEN $4::segment_match ELSE match_type END),
    conditions=$5,
    tags=$6::VARCHAR(100)[],
    updated_at=NOW()
WHERE id = $1 RETURNING id;

-- name: update-segment-count
UPDATE segments SET subscriber_count=$2, updated_at=NOW() WHERE id = $1;

-- name: delete-segment
DELETE FROM segments WHERE id = $1;
