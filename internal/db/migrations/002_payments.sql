CREATE TABLE IF NOT EXISTS payments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    server_id         UUID REFERENCES game_servers(id) ON DELETE SET NULL,
    plan              TEXT NOT NULL,
    amount_cents      INT NOT NULL,
    currency          TEXT NOT NULL DEFAULT 'USD',
    provider          TEXT NOT NULL DEFAULT 'paypal',
    provider_order_id TEXT,
    status            TEXT NOT NULL DEFAULT 'created',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_payments_user ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_order ON payments(provider_order_id);
