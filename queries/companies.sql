-- companies (multi-tenant, v7.17.0+)

-- name: get-companies
SELECT * FROM companies ORDER BY id;

-- name: get-company
SELECT * FROM companies WHERE id = $1;

-- name: create-company
INSERT INTO companies (name, slug)
    VALUES($1, $2) RETURNING *;

-- name: update-company
UPDATE companies SET
    name = (CASE WHEN $2 != '' THEN $2 ELSE name END),
    slug = (CASE WHEN $3 != '' THEN $3 ELSE slug END),
    updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: delete-company
-- FK constraints (ON DELETE RESTRICT) prevent deleting a company with
-- live tenant data. The handler explains the error to the user.
DELETE FROM companies WHERE id = $1;

-- name: get-company-stats
-- Used by the admin Companies page to show row counts per tenant before
-- a delete (so admins know what they'd be blocked by).
SELECT
    co.id,
    co.name,
    co.slug,
    (SELECT COUNT(*) FROM users WHERE company_id = co.id) AS users,
    (SELECT COUNT(*) FROM lists WHERE company_id = co.id) AS lists,
    (SELECT COUNT(*) FROM subscribers WHERE company_id = co.id) AS subscribers,
    (SELECT COUNT(*) FROM campaigns WHERE company_id = co.id) AS campaigns,
    (SELECT COUNT(*) FROM templates WHERE company_id = co.id) AS templates,
    (SELECT COUNT(*) FROM warming_senders WHERE company_id = co.id) AS warming_senders,
    (SELECT COUNT(*) FROM warming_campaigns WHERE company_id = co.id) AS warming_campaigns,
    (SELECT COUNT(*) FROM roles WHERE company_id = co.id) AS roles
FROM companies co
ORDER BY co.id;
