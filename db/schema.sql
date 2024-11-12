CREATE TABLE accounts (
    account_id BIGSERIAL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX accounts_email_idx ON accounts(email);

CREATE TYPE TX_OPERATION_TYPE AS ENUM('debit', 'credit');

CREATE TABLE transactions (
    transaction_id BIGSERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL REFERENCES accounts(account_id),
    operation TX_OPERATION_TYPE NOT NULL,
    amount DECIMAL(16, 2) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'MXN',
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);