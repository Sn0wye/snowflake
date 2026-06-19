-- Composite indexes serve the directional transaction listing
-- (WHERE sender/receiver = ? ORDER BY created_at DESC) as index-ordered scans.
CREATE INDEX idx_transactions_sender_created ON transactions(sender_account_id, created_at DESC);
CREATE INDEX idx_transactions_receiver_created ON transactions(receiver_account_id, created_at DESC);

-- Redundant: leading column of the composites covers bare sender/receiver lookups.
DROP INDEX idx_transactions_sender_account_id;
DROP INDEX idx_transactions_receiver_account_id;
