-- name: GetAccountByEmail :one
SELECT * FROM accounts WHERE email = $1 LIMIT 1;

-- name: CreateAccount :one
INSERT INTO accounts
    (first_name, last_name, email, created_at, updated_at)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateAccountBalance :one
UPDATE accounts
    SET last_balance_at = $1, total_balance = $2, avg_debit_amount = $3, avg_credit_amount = $4
    WHERE account_id = $5
RETURNING *;

-- name: InsertTransaction :one
INSERT INTO transactions
    (account_id, operation, amount, performed_at, created_at, updated_at)
VALUES
    ($1, $2, $3, $4, $5, $6)
RETURNING transaction_id;

