-- Sondage / feedback visiteurs (widget « Ton avis ? ») → comprendre l'acquisition,
-- les cas d'usage et les pistes d'amélioration. Aucune donnée perso obligatoire.
CREATE TABLE IF NOT EXISTS feedback (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rating     INT,            -- 1-5 (nullable)
    source     TEXT,           -- comment connu : reddit / discord / ami / google / autre
    usecase    TEXT,           -- pourquoi : amis / communaute / test / autre
    game       TEXT,           -- jeu d'intérêt : minecraft / hytale / satisfactory / autre
    message    TEXT,           -- commentaire libre
    email      TEXT,           -- optionnel (suivi)
    page       TEXT,           -- page d'où le sondage a été soumis
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_feedback_created ON feedback(created_at DESC);
