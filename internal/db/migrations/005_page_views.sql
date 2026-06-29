-- Tracking des pages visitées (analytics /admin) → savoir sur quelles pages les
-- utilisateurs étaient. Beacon front (POST /api/v1/track) à chaque changement de route.
-- user_id/username sont remplis quand le visiteur est connecté (cookie JWT), sinon NULL
-- (visiteur anonyme). Agrégats dans /admin (exclut le compte proprio, cf. METRICS_EXCLUDE_USERS).
CREATE TABLE IF NOT EXISTS page_views (
    id         BIGSERIAL PRIMARY KEY,
    user_id    UUID,           -- NULL si visiteur anonyme
    username   TEXT,           -- snapshot du pseudo Discord au moment de la visite (NULL si anonyme)
    path       TEXT NOT NULL,  -- chemin visité (sans query string), ex. /games/minecraft
    referer    TEXT,           -- page précédente (optionnel)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_page_views_created ON page_views(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_page_views_user ON page_views(user_id) WHERE user_id IS NOT NULL;
