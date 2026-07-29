-- Candidatures d'accès anticipé au mod Calradia-Coop (avant la release publique).
-- Gérées depuis le panel /admin (onglet Calradia) ; le gate de téléchargement
-- (GET /api/v1/calradia/download-auth) autorise admin + candidats « approved ».
CREATE TABLE IF NOT EXISTS calradia_early_access (
    id               BIGSERIAL PRIMARY KEY,
    discord_username TEXT NOT NULL,
    message          TEXT NOT NULL DEFAULT '',
    status           TEXT NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    decided_at       TIMESTAMPTZ
);

-- Une seule candidature par pseudo Discord (insensible à la casse).
CREATE UNIQUE INDEX IF NOT EXISTS calradia_early_access_username_key
    ON calradia_early_access (lower(discord_username));
