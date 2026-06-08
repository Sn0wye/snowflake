CREATE TABLE scores (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    income NUMERIC(38,2),
    debt NUMERIC(38,2),
    assets_value NUMERIC(38,2),
    credit_score INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE INDEX idx_scores_user_id ON scores (user_id);
