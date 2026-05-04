-- media
-- name: insert-media
-- $7 = company_id (v7.17.0); 0 falls back to Solomon=1.
INSERT INTO media (uuid, filename, thumb, content_type, provider, meta, company_id, created_at)
    VALUES($1, $2, $3, $4, $5, $6, COALESCE(NULLIF($7::INT, 0), 1), NOW()) RETURNING id;

-- name: query-media
-- $5 = company_id (v7.17.0); 0 disables filter.
SELECT COUNT(*) OVER () AS total, * FROM media
    WHERE ($1 = '' OR filename ILIKE $1) AND provider=$2
    AND ($5::INT = 0 OR company_id = $5::INT)
    ORDER BY created_at DESC OFFSET $3 LIMIT $4;

-- name: get-media
SELECT * FROM media WHERE
    CASE
        WHEN $1 > 0 THEN id = $1
        WHEN $2 != '' THEN uuid = $2::UUID
        WHEN $3 != '' THEN filename = $3
        ELSE false
    END
    AND ($4::INT = 0 OR company_id = $4::INT);

-- name: delete-media
DELETE FROM media WHERE id=$1 RETURNING filename;

