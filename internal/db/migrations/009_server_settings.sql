-- Réglages de jeu par serveur (carte, nom de session, mots de passe…).
-- Clés et validation définies par GameDef.Settings (internal/servers/settings.go) ;
-- appliqués en surcharge d'env à la (re)création du container.
ALTER TABLE game_servers
    ADD COLUMN IF NOT EXISTS settings JSONB NOT NULL DEFAULT '{}'::jsonb;
