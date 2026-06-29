-- Modpack Minecraft : un serveur peut être (re)provisionné depuis un modpack complet
-- (loader + version + mods définis par le pack), installé par l'image itzg.
--   ""                       = pas de modpack (loader paper/fabric/forge classique)
--   "ftb:<packId>:<verId>"      → TYPE=FTBA  (Feed-The-Beast)
--   "modrinth:<projId>:<verId>" → TYPE=MODRINTH (verId optionnel)
ALTER TABLE game_servers ADD COLUMN IF NOT EXISTS modpack TEXT NOT NULL DEFAULT '';
