-- name: GetAccountByEmail :one
SELECT * FROM accounts WHERE email = $1 LIMIT 1;

-- name: CreateAccount :one
INSERT INTO accounts
    (first_name, last_name, email, created_at, updated_at)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING *;

-- name: InsertTransaction :one
INSERT INTO transactions
    (account_id, operation, amount, performed_at, created_at, updated_at)
VALUES
    ($1, $2, $3, $4, $5, $6)
RETURNING transaction_id;
