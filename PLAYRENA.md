# Playrena — État du projet (contexte pour nouvelles conversations)

> Fichier de reprise rapide. Mentionne **« Playrena »** dans une nouvelle conversation et
> pointe vers ce fichier pour recharger tout le contexte.
> Dernière mise à jour : 2026-06-10.

## C'est quoi

**Playrena** = panel web pour **louer et gérer des serveurs de jeux** (modèle type
Shockbyte), hébergé sur **https://mcserver.vbt-prog.com**. Propriétaire : Simon Gaudet
(`aVaBaTa` sur GitHub). Marque affichée : « Playrena ». Création de serveur en quelques
secondes via login Discord OAuth.

## Les 3 repos GitHub (privés, propriétaire `aVaBaTa`)

| Repo | Chemin local | Branche courante | Rôle |
|---|---|---|---|
| **SGRentMcServer** | `~/Shared_Projects/2026/SGRentMcServer` | `Claude/feature/playrena-mvp` | Le panel : backend Go + frontend Next.js |
| **SGPortfolio** | `~/Shared_Projects/2026/SGPortfolio` | `Claude/feature/v0.0.1-setup` | Portfolio + **nginx** (reverse proxy global) + certs Origin partagés |
| **SGMail** | `~/Shared_Projects/2026/SGMail` | `Claude/feature/mailserver` | Serveur mail docker-mailserver (`mail.vbt-prog.com`) |

Remotes : `git@github.com:aVaBaTa/<repo>.git`. **Aucun push automatique** — l'utilisateur
décide quand pusher. Convention de branche imposée : `Claude/feature/<desc>` (skill
`git-feature-branch`). Trailer de commit : `Co-Authored-By: Claude Opus 4.8`.

## Infra (tout sur une seule machine : `xe80dell`)

- IP publique **24.157.140.226**, domaine **vbt-prog.com**, Cloudflare **Full Strict**.
  Cert Origin couvre `*.vbt-prog.com` (dans `SGPortfolio/docker/nginx/certs/`).
- **La machine de dev EST le serveur de prod** (`xe80dell`). Le 2ᵉ Dell (node2,
  `10.0.0.110`) n'est pas pleinement configuré. ⚠️ **Ne jamais valider en localhost** —
  le site public existe déjà ; valider en `curl`ant `https://mcserver.vbt-prog.com`.
- **Routing nginx** (`SGPortfolio/docker/nginx/nginx.conf`, entrypoint réel = nginx, pas
  Traefik) pour `mcserver.vbt-prog.com` :
  - `/admin/` → `mcserver-monitor:8090` (Basic Auth, `.htpasswd`)
  - `/auth/` + `/api/` → `mcserver-api:8080`
  - `/` → `mcserver-frontend:3000`
  - ⚠️ **Piège bind-mount** : éditer `nginx.conf` casse l'inode → `docker restart portfolio-nginx` (pas juste reload).
- **Containers prod** : `portfolio-nginx` (80/443), `portfolio`, `mcserver-frontend`,
  `mcserver-api`, `mcserver-monitor`, `mcserver-postgres`, `mcserver-redis`,
  `mc-router` (25565), `mailserver`.
- ⚠️ **Cloudflare ne proxifie PAS le trafic de jeu** (TCP/UDP brut) → les joueurs se
  connectent à l'IP d'origine. Le **pare-feu** (`ufw`) et le NAT du routeur doivent ouvrir
  les ports de jeu.

## Architecture du panel (SGRentMcServer)

- **Backend Go** (`internal/`) : API chi sur `:8080`, Postgres (`internal/db/migrations/`,
  appliquées via `docker-entrypoint-initdb.d` → migrations manuelles en prod), Redis.
  Orchestration Docker via `internal/orchestrator` (socket local node1 + SSH node2).
- **Registre de jeux** = cœur du multi-jeux : `internal/servers/games.go`. Une `GameDef`
  par jeu pilote image Docker, ports (`PortMapping` host↔interne+proto), volume (`DataPath`),
  routing (`UsesMCRouter`), planchers de ressources (`MinRAMMb`/`MinCPUCores`), `NeedsAuth`,
  et un builder d'env. **Ajouter un jeu = ajouter une entrée ici** (+ frontend).
- **Ports** : `internal/servers/ports.go` alloue un « port de base » dans **25566–26065**
  (25565 réservé à mc-router). La moitié haute **26066–26565** est réservée aux ports dérivés
  (ex. messaging Satisfactory = base+500). Toute la plage `25566:26565` doit être ouverte au
  pare-feu (TCP **et** UDP).
- **Frontend Next.js** (custom, voir `frontend/AGENTS.md`), build **standalone**, i18n FR/EN
  dans `frontend/lib/i18n.tsx` (dict `fr` typé, `en: Dict`). Plans tarifaires partagés dans
  `lib/i18n.tsx` (`PLANS`). Liste des jeux dans `frontend/lib/games.ts`.
- **Mail/support** : `POST /api/v1/support` → `internal/mailer` → SMTP `mailserver:587`.

## État des jeux

| Jeu | Statut | Image | Port(s) | Volume | Notes |
|---|---|---|---|---|---|
| **Minecraft** | ✅ live | `itzg/minecraft-server` | 25565/tcp (mc-router par hostname) | `/data` | versions PAPER, RCON (console/joueurs), gratuit dès 1 Go |
| **Satisfactory** | ✅ live | `wolveix/satisfactory-server` | base/udp (jeu) + base+500/tcp (messaging), identité | `/config` | plancher **4 Go gratuit** (promo), IP:port direct, pas de RCON |
| **Hytale** | ✅ live | `ghcr.io/terkea/hytale-server` | base/udp (QUIC, `SERVER_PORT`) | `/data` | plancher **4 Go gratuit**, **OAuth device-code au 1er boot** (statut `auth_required`, panel affiche URL+code), IP:port direct |
| Rust, ARK | ⏳ soon | — | — | — | marqués « Bientôt » dans `games.ts` |

**Flux d'auth Hytale** (`internal/server/handlers_auth.go`) : à la 1ʳᵉ création, un watcher
détaché lit les logs, détecte le prompt OAuth Hytale (`oauth.accounts.hytale.com`), passe le
statut à `auth_required`, expose `GET /servers/{id}/auth` (URL+code), puis `running` une fois
autorisé. Creds persistés dans le volume → redémarrages auto.

**Promo prix** : Satisfactory & Hytale sont **offerts gratuitement pour l'instant** (plancher
4 Go appliqué quel que soit le plan choisi, facturation = 0 $), pour attirer des clients.

## Déploiement (sur xe80dell)

```bash
# Backend
cd ~/Shared_Projects/2026/SGRentMcServer && docker compose up -d --build api
# Frontend (container standalone, PAS dans compose)
cd ~/Shared_Projects/2026/SGRentMcServer/frontend && docker build -t mcserver-frontend:latest .
docker rm -f mcserver-frontend && docker run -d --name mcserver-frontend \
  --restart unless-stopped --network sgportfolio_portfolio-net mcserver-frontend:latest
# Pare-feu jeux (node1 + node2) — requis pour Satisfactory/Hytale
sudo ufw allow 25566:26565/udp && sudo ufw allow 25566:26565/tcp
```

## Pièges / règles

- **Jamais valider en localhost** ; toujours via `https://mcserver.vbt-prog.com`.
- Secrets **jamais commités** : `.env`, `CREDENTIALS.md.save`, `.mailpw*`, tokens Discord.
- Frontend = Next.js modifié → lire `frontend/AGENTS.md` avant d'y coder.
- `NEXT_PUBLIC_SERVER_HOST` (frontend) = host affiché pour la connexion directe (défaut
  `24.157.140.226`).

## Backlog en cours (demandé le 2026-06-10, à traiter)

1. **Dashboard par jeu** : un dashboard similaire mais adapté par jeu.
2. Mettre à jour les chiffres sur `/games/satisfactory`.
3. Rust & ARK : activer « être prévenu » + **ping serveur** sur la page avant achat.
4. Ce fichier de résumé ✅ + (proposé) skill « playrena » d'auto-contexte.
5. Adapter `/admin` (mcserver-monitor) aux dernières fonctionnalités + en proposer.
6. Section site pour les **mods Hytale** (en cours de création par Simon).
7. Baisser les prix Minecraft (catchy) — ⚠️ `PLANS` est partagé entre tous les jeux.
