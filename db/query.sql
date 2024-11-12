-- name: GetAccountByEmail :one
SELECT account_id, first_name, last_name, email, created_at, updated_at FROM accounts WHERE email = $1 LIMIT 1;

-- name: CreateAccount :one
INSERT INTO accounts
    (first_name, last_name, email, created_at, updated_at)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING account_id;
