-- Loader Minecraft : type de serveur choisi à la création.
--   paper  → serveur à plugins (Bukkit/Spigot), dossier /plugins (défaut, rétro-compatible)
--   fabric → serveur moddé Fabric, dossier /mods
--   forge  → serveur moddé Forge, dossier /mods
-- Les jeux non-Minecraft gardent 'paper' par défaut mais l'ignorent (loader non utilisé).
ALTER TABLE game_servers ADD COLUMN IF NOT EXISTS loader TEXT NOT NULL DEFAULT 'paper';
