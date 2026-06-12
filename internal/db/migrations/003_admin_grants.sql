-- #M / droits admin : droit de création de serveur "illimité" (sans paiement,
-- tout plan) par utilisateur, activable/désactivable depuis /admin.
ALTER TABLE users ADD COLUMN IF NOT EXISTS unlimited_create BOOLEAN NOT NULL DEFAULT FALSE;

-- Le propriétaire (Simon) a le droit par défaut. Reste désactivable via /admin.
UPDATE users SET unlimited_create = TRUE WHERE email = 'simong2004@hotmail.com';
