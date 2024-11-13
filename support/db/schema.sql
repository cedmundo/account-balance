CREATE TABLE accounts (
    account_id BIGSERIAL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL,
    locale TEXT NOT NULL DEFAULT 'es-MX',
    total_balance DECIMAL(16, 2),
    avg_debit_amount DECIMAL(16, 2),
    avg_credit_amount DECIMAL(16, 2),
    last_balance_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX accounts_email_idx ON accounts(email);

CREATE TYPE TX_OPERATION_TYPE AS ENUM('debit', 'credit');

CREATE TABLE transactions (
    transaction_id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(account_id),
    operation TX_OPERATION_TYPE NOT NULL,
    amount DECIMAL(16, 2) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'MXN',
    performed_at DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);
