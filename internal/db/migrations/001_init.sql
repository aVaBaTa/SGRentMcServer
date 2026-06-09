CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    discord_id  TEXT UNIQUE NOT NULL,
    username    TEXT NOT NULL,
    avatar      TEXT,
    email       TEXT,
    plan        TEXT NOT NULL DEFAULT 'free',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS game_servers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    subdomain   TEXT UNIQUE NOT NULL,
    game        TEXT NOT NULL DEFAULT 'minecraft',
    plan        TEXT NOT NULL DEFAULT 'free',
    version     TEXT NOT NULL DEFAULT 'LATEST',
    node        TEXT NOT NULL,
    container_id TEXT,
    status      TEXT NOT NULL DEFAULT 'stopped',
    ram_mb      INT NOT NULL DEFAULT 1024,
    cpu_cores   NUMERIC(4,2) NOT NULL DEFAULT 1.0,
    port        INT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_game_servers_user_id ON game_servers(user_id);
CREATE INDEX idx_game_servers_status ON game_servers(status);
