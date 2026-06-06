CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    balance BIGINT DEFAULT 0 NOT NULL,
    last_reconciled_at TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'active' NOT NULL,
    reconciliation_status VARCHAR(20) DEFAULT 'ok' NOT NULL,
    discrepancy_detected_at TIMESTAMPTZ,
    discrepancy_amount BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_accounts_user_id ON accounts(user_id);

CREATE TABLE flakes (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL,
    key_type VARCHAR(20) NOT NULL,
    key_value VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'active' NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_flakes_key_value ON flakes(key_value);
CREATE INDEX idx_flakes_account_id ON flakes(account_id);

CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    amount BIGINT NOT NULL,
    sender_account_id UUID,
    receiver_account_id UUID,
    flake_key_used VARCHAR(255),
    description TEXT,
    idempotency_key UUID NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_transactions_idempotency_key ON transactions(idempotency_key);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_sender_account_id ON transactions(sender_account_id);
CREATE INDEX idx_transactions_receiver_account_id ON transactions(receiver_account_id);

CREATE TABLE transaction_histories (
    id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL,
    account_id UUID NOT NULL,
    entry_type VARCHAR(10) NOT NULL,
    amount BIGINT NOT NULL,
    balance_before BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transaction_histories_transaction_id ON transaction_histories(transaction_id);
CREATE INDEX idx_account_created ON transaction_histories(account_id, created_at);
