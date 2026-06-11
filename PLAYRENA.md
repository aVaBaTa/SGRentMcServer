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

- **Jamais valider en localhost** ; toujours via le domaine public **`https://playrena.vbt-prog.com`** (ex-`mcserver.vbt-prog.com`, encore actif en parallèle).
- Secrets **jamais commités** : `.env`, `CREDENTIALS.md.save`, `.mailpw*`, tokens Discord.
- Frontend = Next.js modifié → lire `frontend/AGENTS.md` avant d'y coder.
- `NEXT_PUBLIC_SERVER_HOST` (frontend) = host affiché pour la connexion directe (défaut
  `24.157.140.226`).

## Fait (2026-06-10/11)

- ✅ Dashboard par jeu (hub `/dashboard` + `/dashboard/[game]`, accent par jeu).
- ✅ Chiffres `/games/satisfactory` & `/games/hytale` : plancher 4 Go affiché.
- ✅ Rust & ARK : pages « être prévenu » (`/games/rust`, `/games/ark`) + composant `ComingSoonGame`.
- ✅ Ping de latence sur les pages d'achat (composant `PingBadge`). ✅ #B corrigé **et déployé** : endpoint nginx `= /ping.ico` (réponse instantanée + `immutable` → Cloudflare sert en **HIT**, vérifié), badge pointé dessus en relatif (suit le domaine courant), warm-up + min sur 5 échantillons, préfixe `~`.
- ✅ `PLAYRENA.md` + skill `~/.claude/skills/playrena/`.
- ✅ `/admin` enrichi : vue Serveurs (DB), start/restart, compteurs/filtre par jeu, business. ⚠️ **image build, container à redéployer (cmd dans le résumé de session)**.
- ✅ Page `/mods` placeholder « bientôt » + lien depuis `/games/hytale`.
- ✅ Prix baissés (grille `PLANS` partagée) : Starter 2$/3$, Standard 5$/6$, Pro 9$/12$, Extreme 16$/22$.
- ✅ Robustesse Hytale (parsing OAuth tolérant + logs bruts) ; port TCP Satisfactory ajouté.

## ✅ Rebrand domaine `mcserver` → `playrena` — FAIT (2026-06-11)

Domaine public = **`https://playrena.vbt-prog.com`** (Option 1 : sous-domaine de vbt-prog.com). Bascule sans coupure.

**Fait et déployé :**
- DNS Cloudflare : A `playrena.vbt-prog.com` → 24.157.140.226 (proxied).
- nginx : bloc 443 `playrena.vbt-prog.com` (principal) + bloc `mcserver.vbt-prog.com` → **301** vers playrena (préserve backlinks/SEO).
- Discord : Redirect URI = `https://playrena.vbt-prog.com/auth/discord/callback` (app `1513764241784705034`) ; `DISCORD_REDIRECT_URL` (compose) basculé → `mcserver-api` rebuild.
- Frontend SEO → playrena (`layout.tsx` metadataBase/OG/canonical, `sitemap.ts`, `robots.ts`, `blog`) → `mcserver-frontend` rebuild.
- Lien header portfolio (`SGPortfolio/src/components/Header.tsx`) → playrena → `portfolio` rebuild.
- Vérifié live : 301 mcserver→playrena, login Discord OK, sitemap/robots/portfolio en playrena.

**Reste (manuel, non bloquant) :** Google Search Console (nouvelle propriété + sitemap) ; mettre à jour `BACKLINKS.md` pour les futures inscriptions.

**Gardé tel quel (interne, invisible du public) :** repo `SGRentMcServer`, module Go `github.com/aVaBaTa/SGRentMcServer`, conteneurs/images `mcserver-*`, mail `contact@mcserver.vbt-prog.com`. Domaine propre (`playrena.com`/`.gg`) = plus tard.

## Backlog à traiter (demandé le 2026-06-11)

- **#A — `mcserver-monitor` à REDÉPLOYER** (manuel, sudo non requis mais secret DB) :
  build déjà fait ; recréer le container avec ses env/réseaux (cmd fournie en session).
- ✅ **#B — Latence du ping trop élevée** — CORRIGÉ ET DÉPLOYÉ (2026-06-11) :
  1) **nginx** (`SGPortfolio/docker/nginx/nginx.conf`, bloc playrena) : nouvelle `location =
     /ping.ico` qui renvoie un 200 vide **instantanément** (aucun upstream) avec
     `Cache-Control: public, max-age=31536000, immutable` + CORS + extension `.ico` →
     Cloudflare la met en cache et répond en **HIT** (vérifié : MISS puis HIT sur playrena
     ET mcserver). Plus aucune revalidation vers l'origine.
  2) **frontend** (`frontend/components/ping-badge.tsx`) : cible relative `/ping.ico` (suit le
     domaine courant, même origine), 1 warm-up (DNS+TLS écartés) puis **min sur 5 échantillons**
     à chaud, préfixe `~`. Mesure désormais le **pur RTT navigateur↔edge** (proximité réseau).
  NB : une **vraie** latence ICMP vers le node de jeu reste impossible côté navigateur — le
  seul cert présent est l'Origin CA Cloudflare (non reconnu par les navigateurs), donc un
  endpoint grey-cloud nécessiterait un cert publiquement valide (Let's Encrypt) dédié.
  Leviers de **vraie** latence joueur (hors badge) : SQM/anti-bufferbloat sur le routeur,
  filaire côté serveur, et multi-région / Cloudflare Spectrum pour les joueurs distants.
- **#C — `/admin` : « faire pour les commandes »** : à clarifier — probablement appliquer
  la même robustesse / les actions aux commandes admin.
- **#D — SGPortfolio : nettoyer les releases GitHub** non nommées `v0.*`
  (`gh release list` / `gh release delete` dans le repo SGPortfolio).
- **#E — Nouveau site « services de création de site web propulsés par IA »**, accessible
  depuis le portfolio, **contact par courriel** (réutiliser `internal/mailer` / SGMail).
  Gros chantier : décider repo (SGPortfolio ? nouveau ?), routing nginx, design, formulaire.

## ⚠️ Blocage connu : connectivité des serveurs UDP (Satisfactory/Hytale)

Le container Satisfactory **fonctionne** (jeu démarré, ports bindés sur l'hôte), mais les
joueurs ne peuvent pas se connecter tant que le **pare-feu UDP n'est pas ouvert** (le MC
marche car seul le **TCP** 25565 a été ouvert). Action utilisateur (sudo, node1 **et** node2)
**+ NAT routeur** pour la plage de jeu :

```bash
sudo ufw allow 25566:26565/udp
sudo ufw allow 25566:26565/tcp
```
C'est le prérequis qui débloque réellement Satisfactory et la connexion Hytale.
