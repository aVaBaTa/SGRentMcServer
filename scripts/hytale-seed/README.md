# Hytale — fichiers de jeu pré-téléchargés (seed)

But : **télécharger Hytale une seule fois** sur la machine et **réutiliser** ces fichiers à
chaque création de serveur. Bénéfices :
- ⚡ Création quasi instantanée (pas de re-téléchargement de plusieurs centaines de Mo).
- 🔑 **Plus d'OAuth "downloader"** (step 1) à faire à chaque fois : le client n'a **que l'auth
  SERVEUR** (step 2, device-code) à compléter dans le panel — c'est lui le responsable.

## Comment ça marche
1. `scripts/hytale-seed/download.sh` lance un container temporaire qui télécharge Hytale dans
   `seeds/_hytale-dl/`. **Tu autorises l'OAuth downloader une seule fois** (compte Hytale licencié).
2. Les fichiers `HytaleServer.jar` + `Assets.zip` sont copiés dans `seeds/hytale/`.
3. Le backend (`SEED_DIR=/seeds`, monté en lecture seule dans `mcserver-api`) détecte ces
   fichiers (`GameDef.SeedFiles`) et, à chaque création d'un serveur Hytale :
   - crée le container avec `AUTO_DOWNLOAD=false` / `SKIP_DOWNLOAD=true`,
   - **seede le volume** (`CopyFileToContainer`) avec les fichiers AVANT le `start`,
   - le serveur démarre direct sur l'**auth serveur** → le panel affiche l'URL+code au client.

## Mise en place (une fois)
```bash
cd ~/Shared_Projects/2026/SGRentMcServer
./scripts/hytale-seed/download.sh          # suivre les instructions OAuth
docker compose up -d --build api           # l'API monte ./seeds et active le seeding
```

## Détails / pièges
- `seeds/` est **gitignored** (fichiers lourds, pas de secret mais inutile en git).
- Si `seeds/hytale/` est vide → **repli automatique** : le container télécharge lui-même
  (comportement historique, avec OAuth downloader). Donc rien ne casse si le seed n'est pas fait.
- **Plancher RAM Hytale = 10 Go** (`MinRAMMb` dans `games.go`) : réaliste/stable (4 Go crashe
  sous charge). Sur xe80dell (31 Go) → ~1-2 serveurs Hytale max. Assumé non rentable.
- **Mises à jour Hytale** : pour rafraîchir les fichiers, relancer `download.sh` (écrase le seed).
- **node2 (SSH)** : `CopyFileToContainer` stream les fichiers via l'API Docker → fonctionne aussi
  sur le node distant (les fichiers sont lus depuis le container API qui monte `./seeds`).
- ⚠️ Le seed ne contient **pas** de credentials : l'auth serveur reste à la charge du client.
