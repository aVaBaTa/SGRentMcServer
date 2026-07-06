# Playrena — État du projet (contexte pour nouvelles conversations)

> Fichier de reprise rapide. Mentionne **« Playrena »** dans une nouvelle conversation
> (le skill `playrena` charge ce fichier + l'état git) OU dis simplement
> **« va lire PLAYRENA.md »**.
> Dernière mise à jour : 2026-06-14.

## 🔄 Reprise rapide (à exécuter au début d'une nouvelle conversation)

1. Lire ce fichier en entier (archi, état des jeux, **backlog #A→#Z**, blocage UDP).
2. Récupérer les **commits récents des 3 repos** (SGRentMcServer travaille sur **`live`**) :

```bash
for d in SGRentMcServer SGPortfolio SGMail; do
  r=~/Shared_Projects/2026/$d
  echo "===== $d ($(git -C "$r" branch --show-current)) ====="
  git -C "$r" log --oneline -8
  git -C "$r" status --short
done
```

3. (Si ops/déploiement) `docker ps --format '{{.Names}}\t{{.Status}}' | grep -E 'mcserver|portfolio|mailserver|sgrent-'`.

→ Branches de travail actuelles : **SGRentMcServer = `live`**, SGPortfolio = `live`
(+ `Claude/feature/portfolio-release-filter` pour le filtre releases), SGMail = `Claude/feature/mailserver`.

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
  - `/admin/` → `mcserver-monitor:8090`. **Accès : (a) session Discord du proprio (aVaBaTa) → entre
    SANS mot de passe**, ou (b) **Basic Auth** (`.htpasswd`, user `admin`) en secours. nginx fait
    `satisfy any` + `auth_request /admin-auth` → `GET /api/v1/admin-session` (API valide le cookie JWT,
    200 si username ∈ `ADMIN_USERS`, défaut `avabata`). → **aVaBaTa connecté peut toujours ouvrir /admin**.
    Vérifié : sans cookie=401(Basic), cookie avabata=200, autre user=401. (Ajouter un admin : env `ADMIN_USERS=a,b`.)
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
- **Ports** : `internal/servers/ports.go` attribue à chaque serveur un **bloc contigu** de
  `portBlockSize` (=16) ports `[base, base+16)` dans **25566–26565** (25565 réservé à mc-router).
  **TOUS** les ports d'un jeu (jeu, query, messaging…) se dérivent de `base` **dans son bloc**
  (offset < 16) → deux serveurs ne peuvent jamais se chevaucher, même avec des ports dérivés
  consécutifs. **Jamais de port fixe en dur** (27015, 8766…) : on configure le jeu pour écouter
  sur `base+offset` (cf. `NextAvailableBase` + cœur pur testable `firstFreeBlockBase`, tests dans
  `ports_test.go`). 62 blocs ⇒ 62 serveurs simultanés (très au-delà de la capacité RAM). Toute la
  plage `25566:26565` doit être ouverte au pare-feu (TCP **et** UDP).
- **Frontend Next.js** (custom, voir `frontend/AGENTS.md`), build **standalone**, i18n FR/EN
  dans `frontend/lib/i18n.tsx` (dict `fr` typé, `en: Dict`). Plans tarifaires partagés dans
  `lib/i18n.tsx` (`PLANS`). Liste des jeux dans `frontend/lib/games.ts`.
- **Mail/support** : `POST /api/v1/support` → `internal/mailer` → SMTP `mailserver:587`.

## État des jeux

| Jeu | Statut | Image | Port(s) | Volume | Notes |
|---|---|---|---|---|---|
| **Minecraft** | ✅ live | `itzg/minecraft-server` | 25565/tcp (mc-router par hostname) | `/data` | versions PAPER, RCON (console/joueurs), gratuit dès 1 Go |
| **Satisfactory** | ✅ live | `wolveix/satisfactory-server` | base/udp+tcp (jeu) + base+1/tcp (messaging), identité, **dans le bloc** | `/config` | plancher **4 Go gratuit** (promo), IP:port direct, pas de RCON |
| **Hytale** | ✅ live | `ghcr.io/terkea/hytale-server` | base/udp (QUIC, `SERVER_PORT`) | `/data` | plancher **10 Go** (réaliste, assumé non rentable), **fichiers pré-téléchargés** (seed → pas de re-download) → le **client** ne fait que l'**auth serveur** (device-code, panel affiche URL+code), IP:port direct |
| **Valheim** | ✅ live | `lloesche/valheim-server` | base/udp (jeu) + base+1/udp, identité, **dans le bloc** | `/config` | plancher **4 Go** (≤10 joueurs, cap natif), IP:port direct, `SERVER_PASS` défaut `playrena` (à rendre configurable), produit **payant** (grille complète) |
| **Calradia-Coop (M&B II)** | ✅ live | `calradia-server:latest` (**image LOCALE**, jamais sur registry — 🔒 confidentiel) | base/tcp + base/udp (même port, `CALRADIA_ADDR`) | `/data` | **notre propre mod Bannerlord** (repo privé Calradia-Coop) ; plancher 1 Go/1 cœur, `freeAtFloor` (offert lancement) ; joueurs = slots du plan cap 8 ; UPnP OFF en container ; vitrine `calradiacoop.vbt-prog.com` |
| Rust, ARK | ⏳ soon | — | — | — | marqués « Bientôt » dans `games.ts` |

**✅ Session 2026-06-29 — Valheim ajouté + promo /games + redeploy** : (1) **Valheim live** (1er jeu
du backlog #U) — `GameDef` `valheim` (`internal/servers/games.go`, `valheimEnv`, ports `base`+`base+1`
udp dans le bloc), `games.ts` passé `live` + page `/games/valheim` + bloc i18n `valheim` + thème cyan.
(2) **#3 masquage promo** : `/games/satisfactory` & `/games/hytale` n'affichent plus que le plan gratuit
au plancher (`freeAtFloor` → `visiblePlans`), payants masqués. (3) **Redéployé** : `mcserver-api`
(active l'allocateur de ports **par blocs** + handlers mods/pageviews commités + Valheim) et
`mcserver-frontend`. Migrations 005/006/007 **déjà appliquées**. Validé en public (Valheim 200, masquage
OK, Hytale 10 Go).

**✅ Session 2026-07-06 — Calradia-Coop (Mount & Blade II: Bannerlord) ajouté + écosystème web** :
mod coop maison (repo privé `aVaBaTa/Calradia-Coop`, branche `claude/BLT/v0.0.1`, cloné dans
`~/Shared_Projects/2026/Calradia-Coop`). **🔒 Code source + serveur Rust = CONFIDENTIELS — jamais
publiés, aucun lien GitHub/téléchargement sur les pages publiques.** (1) **Image Docker locale**
`calradia-server:latest` (server/Dockerfile, 76 Mo, monde 7777 tcp+udp + rendezvous 7778 ; cible bin
`calradia-rendezvous` manquante ajoutée) — `pullImage` (orchestrateur) a maintenant un **repli
image-locale**. (2) **GameDef `calradia-coop`** (games.go) : port unique tcp+udp aligné sur le bloc,
1 Go/1 cœur, `CALRADIA_MAX_GAMEPLAY`=slots cap 8, `CALRADIA_UPNP=0` (bridge Docker). (3) **Frontend** :
`games.ts` live `freeAtFloor` (entrée « Bannerlord (Together) » remplacée), page `/games/calradia-coop`
(features du mod, 4 captures `public/games/calradia/`, crédits → portfolio + vitrine), i18n FR/EN,
article blog `calradia-coop-bannerlord-serveur` + catégorie blog « Bannerlord — Calradia-Coop ».
(4) **Vitrine du mod `calradiacoop.vbt-prog.com`** : statique bilingue dans `Calradia-Coop/site/`
(container `calradiacoop-web` nginx:alpine + bloc nginx SGPortfolio) — vitrine + devlog #1, gatekeep
(pas de download). ⚠️ **DNS Cloudflare `calradiacoop` → 24.157.140.226 = action Simon.** (5) **Infra
xe80dell** : containers `calradia-rendezvous` (7778) + `calradia-world` (7777) up ; **NAT routeur posé
par UPnP sans sudo** (`scripts/upnp-map.py` — la box Helix XB7 ne répond pas au SSDP (ufw) → `--desc
http://10.0.0.1:49152/IGDdevicedesc_brlan0.xml`, et n'accepte AddPortMapping que depuis l'IP cible →
bind source 10.0.0.2 ; bail permanent OK) ; cron 6 h `upnp-refresh.sh` (la table UPnP saute au reboot
routeur) ; **7777+7778 vérifiés joignables depuis Internet** (check-host.net). La plage 25566-26565
reste une règle routeur manuelle (table UPnP trop petite pour 1000 ports). (6) **Annonces** : kit prêt
dans `Calradia-Coop/docs/25-annonces-lancement.md`. (7) Portfolio : lien « Calradia-Coop » (Écosystème).
Reste : clé d'accès par instance (protocole v0.0.2), DNS, poster les annonces.

**Catalogue à étendre (backlog #U)** : ~~Valheim~~ (fait), Palworld, Terraria, The Forest, 7 Days to Die,
ARK: Survival Evolved, ARK: Survival Ascended, Stardew Valley + autres candidats populaires.
📄 **Specs techniques (image Docker, ports, RAM, pièges) : `docs/jeux-populaires-a-ajouter.md`.**

**Flux d'auth Hytale** (`internal/server/handlers_auth.go`) : à la 1ʳᵉ création, un watcher
détaché lit les logs, détecte le prompt OAuth Hytale (`oauth.accounts.hytale.com`), passe le
statut à `auth_required`, expose `GET /servers/{id}/auth` (URL+code), puis `running` une fois
autorisé. Creds persistés dans le volume → redémarrages auto.

**Seed Hytale (fichiers pré-téléchargés)** : pour éviter le re-download (et son OAuth downloader)
à chaque création, les fichiers de jeu sont **téléchargés une fois** sur l'hôte
(`scripts/hytale-seed/download.sh`, OAuth downloader fait UNE fois par l'admin) et stockés dans
`seeds/hytale/`. Le backend (`SEED_DIR=/seeds`, `GameDef.SeedFiles`) **seede le volume**
(`CopyFileToContainer`) avec `AUTO_DOWNLOAD=false` AVANT le start → le serveur démarre direct sur
l'**auth serveur**, que le **client** complète via le panel. Repli automatique sur download si le
seed est absent. (cf. `scripts/hytale-seed/README.md`).

**Promo prix** : Satisfactory **offert** (plancher 4 Go, facturation = 0 $). **Hytale** : plancher
relevé à **10 Go** (2026-06-12, réaliste/stable) — décision produit assumée **non rentable** (cf.
`docs/hytale-analyse-marche.md`) ; ~1 serveur à la fois sur xe80dell.

## Déploiement (sur xe80dell)

> ✅ **Le plus simple : `./scripts/deploy.sh`** (tout) ou `./scripts/deploy.sh api|monitor|frontend`.
> Le script encapsule **tous** les env/mounts/réseaux (source `.env`) → évite les oublis qui cassent
> des features (SSR promo `INTERNAL_API_URL`, seed Hytale `SEED_DIR`, admin token, réseaux du monitor).
> Les commandes manuelles ci-dessous restent la référence.

```bash
# Backend
cd ~/Shared_Projects/2026/SGRentMcServer && docker compose up -d --build api
# Frontend (container standalone, PAS dans compose)
cd ~/Shared_Projects/2026/SGRentMcServer/frontend && docker build -t mcserver-frontend:latest .
docker rm -f mcserver-frontend && docker run -d --name mcserver-frontend \
  --restart unless-stopped -e INTERNAL_API_URL="http://mcserver-api:8080" \
  --network sgportfolio_portfolio-net mcserver-frontend:latest
# ⚠️ INTERNAL_API_URL = pour le SSR de la promo (#Y) : le layout fetch /api/v1/promo côté
#   serveur. Sans cet env il retombe sur l'URL publique (timeout 2s → promo SSR = 0 jusqu'à
#   la revalidation ISR ; le client corrige quand même). Toujours le mettre au redéploiement.
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
- ✅ Chiffres `/games/satisfactory` (4 Go) & `/games/hytale` (**10 Go** depuis 2026-06-12) : plancher affiché.
  Source unique = `frontend/lib/games.ts` (`minRamGb` + flag `freeAtFloor` pour les jeux offerts).
- ✅ **Page serveur `/dashboard/servers/[id]` — plans à jour selon le jeu** (2026-06-12) : la section
  « changement de plan » utilisait un tableau `PLANS` **codé en dur en double** (chiffres figés, sans
  plancher) → remplacé par la **grille partagée** (`lib/i18n.tsx` `PLANS` + `priceFor`/`fmtMoney`,
  devise FR/EN). RAM relevée au **plancher du jeu** (`Math.max(plan, minRamGb)`). Jeux `freeAtFloor`
  (Satisfactory, Hytale) → **plans payants masqués**, plan gratuit affiché « offert au plancher »
  (clé i18n `srv.promoFree`). Build Next.js OK. ⚠️ **frontend à rebuild/redeploy** pour prise d'effet.
  Reste possible (non fait, hors scope demandé) : appliquer le même masquage promo sur les **pages
  publiques** `/games/satisfactory` & `/games/hytale` (elles montrent encore la grille payante).
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

## ✅ Session 2026-06-11 — Discord, mail, webmail, support, i18n, branches

**Frontend / UX :**
- Dashboard `/dashboard` : bouton **« Créer un serveur » centré et mis en avant**, **sélecteur de jeu** (Minecraft live ; autres « Bientôt » désactivés, depuis `lib/games.ts`), **switch FR/EN** (composant `LanguageSwitcher` exporté de `site-chrome.tsx`).
- **i18n complet de la page serveur** `/dashboard/servers/[id]` (dict `srv` FR/EN dans `lib/i18n.tsx`) + switch de langue.
- **Selects cohérents cross-navigateur** (fix rendu macOS/Safari) : classe `.select-dark` dans `globals.css` (appearance-none + chevron SVG + color-scheme dark), appliquée à tous les `<select>`.
- **Section FAQ** sur la page principale (`/games/minecraft`) : accordéon bilingue + balisage SEO `FAQPage`.
- **Section Support** (`#support`, même page) : formulaire bilingue (nom/email/objet/message) → `POST /api/v1/support`.

**Support email (backend + infra) :**
- `internal/mailer/mailer.go` (SMTP STARTTLS + auth, verify off car réseau interne) + `internal/server/handlers_support.go` (route **publique** `POST /api/v1/support`, hors middleware auth — groupe séparé dans `server.go`). Config SMTP dans `config.go`.
- `mcserver-api` connecté au réseau **`sgmail_default`** (compose) pour joindre `mailserver:587`. Envoi authentifié comme `contact@mcserver.vbt-prog.com` → livré dans l'INBOX (vérifié). `SMTP_PASS` dans `.env` (gitignored).

**Discord (serveur communautaire) :**
- Serveur **« Playrena »** créé (guild `1514341359794786426`, proprio = l'utilisateur) et **configuré par l'API** (bot token) : rôles Fondateur/Staff/Modérateur/Client, catégories 📢 INFORMATIONS (lecture seule) / 💬 COMMUNAUTÉ / 🛟 SUPPORT (#support-prioritaire réservé Client) / 🔊 VOCAUX, règlement + FAQ bilingues postés, icône = logo.
- **Invitation `https://discord.gg/3knNHXqpNG`** intégrée au site (`site-chrome.tsx`, bandeau + footer). ⚠️ Le bot token a été partagé en clair → **à régénérer**.

**Mail / DNS (token Cloudflare temporaire — à révoquer) :**
- Réparé : `mail.vbt-prog.com` était **proxied (cassait le mail)** → passé **DNS-only** → 24.157.140.226. Ajout SPF + DKIM pour `mcserver.vbt-prog.com`.
- Comptes mail : `contact@vbt-prog.com`, `contact@mcserver.vbt-prog.com`, `reddit@mcserver.vbt-prog.com` (mdp dans `SGMail/.mailpw*.txt`).
- ⚠️ **Ports 25 (in/out) et 993 bloqués** (FAI/NAT) → pas de réception/envoi externe, pas d'IMAP. Seul HTTP/443 marche. Réception Reddit impossible en self-host → utiliser Hotmail (ou relais SMTP plus tard).
- **Webmail Roundcube déployé : `https://webmail.vbt-prog.com`** (container `roundcube` sur `sgmail_default` + `portfolio-net`, route nginx, config dans `SGMail/docker-data/roundcube/`). Seul moyen de consulter les boîtes.

**Repos / branches :**
- **SGMail** → repo GitHub **privé** `aVaBaTa/SGMail` (push). `.gitignore` élargi à `.mailpw*.txt`.
- Branche **`live`** créée + poussée dans **SGRentMcServer** et **SGMail** (= état déployé, sans secret). SGPortfolio resté sur sa branche feature.
- **`PARTENAIRES.md`** créé (cibles de partenariat : créateurs, communautés, écoles QC, etc.).

**À FAIRE (suite demandée, non terminé) :**
- [ ] Merger la branche SGPortfolio `Claude/feature/portfolio-release-filter` dans `live`.
- [ ] Re-vérifier que les courriels (support) fonctionnent.
- [ ] Créer les comptes Reddit (utiliser Hotmail pour la vérif vu le blocage entrant).
- [ ] Révoquer le **token Cloudflare** + régénérer le **bot token Discord**.

## ✅ Session 2026-06-12 — Hytale jouable (seed + 10 Go) + étude de marché

**Hytale rendu fonctionnel selon décision produit de Simon** (« je veux que Hytale fonctionne,
10 Go par serveur même si pas rentable, le client fait lui-même l'auth, mais le serveur déjà
téléchargé sur ma machine ») :
- **Plancher 10 Go** (`internal/servers/games.go`, `MinRAMMb: 10240`) — 4 Go crashe sous charge.
- **Le client fait l'auth SERVEUR** (device-code) via le panel — comportement déjà en place (watcher).
- **Fichiers de jeu pré-téléchargés (seed)** → plus de re-download ni d'OAuth downloader à chaque
  création. Mécanique :
  - `GameDef.SeedFiles` (Hytale = `HytaleServer.jar` + `Assets.zip`) + config `SEED_DIR` (`/seeds`).
  - `Node.CopyFileToContainer` (`internal/orchestrator/container.go`) : injecte les fichiers dans
    le volume **avant le start** (stream tar, marche aussi sur node2 SSH).
  - `provisionServer`/`buildSpec` (`internal/server/handlers_servers.go`) : si seed présent →
    `AUTO_DOWNLOAD=false` + copie → le serveur boote direct sur l'auth serveur. **Repli auto** sur
    download si seed absent (rien ne casse).
  - `docker-compose.yml` : montage `./seeds:/seeds:ro` + `SEED_DIR=/seeds`. `seeds/` gitignored.
  - Script one-time `scripts/hytale-seed/download.sh` (+ README) : télécharge une fois (OAuth
    downloader fait UNE fois par l'admin) → `seeds/hytale/`.
- **Déployé** : `mcserver-api` rebuild (restart=0, `/seeds` monté, `SEED_DIR` OK). Build + tests verts.
- ⚠️ **Capacité xe80dell** : 10 Go + overhead (~+50 % → ~15 Go cgroup) sur 31 Go total (≈10 Go déjà
  pris) ⇒ réaliste **1 serveur Hytale à la fois**. Assumé (non rentable).
- **Correctif UI auth Hytale (2026-06-12, déployé api+frontend)** : « rien ne se passait » car
  (a) container test orphelin (DB `auth_required` mais container supprimé) → nettoyé + endpoint
  `/auth` renvoie maintenant `gone`/`booting`, (b) la **console n'était visible qu'en `running`**
  → désormais **affichée live pendant `auth_required`** (le client voit le téléchargement + le
  device-code). Nouvelle **belle carte d'auth** (`dashboard/servers/[id]/page.tsx`) : étapes 1-2-3,
  bouton « Autoriser sur Hytale », **code copiable**, spinner, états `booting`/`gone`, i18n FR/EN.
  Rappel : l'auth Hytale est **auto** (l'image poll seule) → le client n'a qu'à autoriser le lien.
- **Console Hytale (stdin) + Listing « Discovery » (2026-06-12, déployé)** : Hytale n'a pas de RCON,
  et le **token de découverte** (server browser in-game, Update 5) s'applique par **commande console
  OP** (`discovery link <token>` ; aussi `discovery unlink`) — l'image n'accepte PAS le token en
  env/config (vérifié). Donc :
  - `GameDef.ConsoleStdin` (true pour Hytale) → conteneur créé avec **`OpenStdin`** ; `Node.SendStdin`
    écrit la commande sur le stdin du serveur via Docker attach (pas de `CloseWrite`/EOF →
    `StdinOnce=false`, stdin reste ouvert). `handleServerCommand` route vers stdin pour les jeux
    `ConsoleStdin` (sinon rcon-cli). Endpoint **`POST /servers/{id}/discovery`** (`{token}` /
    `{unlink:true}`, token validé par regex anti-injection).
  - Frontend : **console tapable** pour Hytale en `running` (sortie dans les logs) + **carte « Listing
    public (browser Hytale) »** (coller le token → bouton Lier → `discovery link`; bouton Délier).
  - ⚠️ Le token discovery est **optionnel** (le serveur marche sans) ; ne marche qu'en `running`
    (serveur authentifié). Heartbeat toutes les 2 min. **Non testé live** (besoin d'un serveur Hytale
    authentifié) — `OpenStdin`+`StdinOnce=false` est la config standard (Pterodactyl) pour ça.
- 💡 **Meilleure manière (recommandée, non faite)** : modèle **provider pre-auth (Model A)** —
  Simon s'auth UNE fois → inject `HYTALE_SERVER_SESSION_TOKEN`/`IDENTITY_TOKEN` (comme le seed) →
  **0 clic pour le client** (serveur déjà authentifié, aucune carte d'auth). À faire si on veut
  l'UX la plus lisse (au prix : identité serveur = compte Simon, limite 500 serveurs/compte).
- **TEST RÉEL réussi (2026-06-12)** : serveur Hytale créé **via le panel** (« Test », 10 Go),
  device-code autorisé par le client → **authentifié**, logs `Skipping download` + `Listening on
  0.0.0.0:25582` + **`Universe ready!`**. Le seed + l'auth + le démarrage fonctionnent **end-to-end**.
  Conteneur créé avec `OpenStdin=true StdinOnce=false` (console/discovery prêts).
- **Hytale CPU 2 → 4 cœurs (déployé)** : lag aux actions = **throttle CPU** (mesuré : conteneur
  bridé à 2 cœurs mais voulait **2,77 cœurs** dès qu'on a relâché à 4). RAM idle = **1,3 Go** seulement
  (plancher 10 Go très large → motive #V). `MinCPUCores: 4.0` dans `games.go` ; serveur courant bumpé
  live (`docker update --cpus 4`) + DB synchro. CPU abondant (48 cœurs), donc généreux sans coût.
- **Fix compteur joueurs « 0/0 » (déployé)** : `/players` passe par `rcon-cli` (Minecraft only) →
  faux 0/0 pour Hytale **ET Satisfactory**. Frontend : compteur + fetch joueurs/whitelist **gatés
  sur `minecraft`** (game-agnostique). Les jeux sans RCON n'affichent plus de compteur ; la console
  (logs) reste visible. Input console = Minecraft (RCON) + Hytale (stdin) ; Satisfactory = aucun.
- ⚠️ **Connectivité = NAT, PAS le serveur** : le serveur écoute, mais (a) il a atterri sur **node2
  (10.0.0.110)** — pas pleinement configuré réseau — et (b) le port UDP n'est pas forwardé. Test
  local = `10.0.0.110:25582` (LAN, pas l'IP publique). Externe = forward routeur **UDP → node2**.
  ⇒ **Décision en attente** : **épingler les jeux à IP directe (Hytale/Satisfactory) sur node1**
  (xe80dell) pour un routage/NAT simple et cohérent (sinon forward au cas par cas selon le node).

**Reste à faire (actions de Simon — compte/ routeur) :**
- [x] **Seed FAIT + VALIDÉ (2026-06-12)** : `seeds/hytale/{HytaleServer.jar 118M, Assets.zip 3.2G}`
  présents, vus par l'API en `/seeds/hytale`. **Test bout en bout OK** : un container booté avec
  `AUTO_DOWNLOAD=false` logue `Skipping download` puis `SERVER AUTHENTICATION REQUIRED` (Visit+Code)
  → **aucun re-download, direct à l'auth serveur** (le client autorise). ⚠️ Le jar est extrait dans
  `Server/` avant d'être déplacé à la racine, et `Assets.zip` est gros (3,2G) et long à extraire →
  `download.sh` corrigé (cherche `Server/`, attend la fin via `game.zip` supprimé). Ne PAS copier un
  `Assets.zip` en cours d'écriture (fichier partiel).
- [x] **Test panel FAIT (2026-06-12)** : serveur Hytale créé via le panel → carte d'auth → autorisé
  → `running` + `Universe ready!`. Auth + seed + démarrage OK.
- [x] **Épinglage node fait + déployé (2026-06-12)** : jeux à IP directe (`!UsesMCRouter` → Hytale,
  Satisfactory) **épinglés sur `cfg.PrimaryNode` (=node1, 10.0.0.2)** via `NodeWithCapacity` ;
  Minecraft reste load-balancé (`BestNode`, mc-router route cross-node). But : NAT simple = **une
  plage → une seule IP LAN**. (config `PRIMARY_NODE`, défaut node1.) Ports déjà sans chevauchement
  (blocs globaux uniques). node1 = **10.0.0.2**, node2 = **10.0.0.110**.
- [ ] **Règles routeur (à poser 1×)** : `UDP 25566-26565 → 10.0.0.2`, `TCP 25566-26565 → 10.0.0.2`
  (Satisfactory), `TCP 25565 → 10.0.0.2` (Minecraft, sûrement déjà fait).
- [ ] **Recréer le Hytale** (l'actuel `078abeef` est sur **node2**, créé avant l'épinglage → le
  supprimer + recréer pour qu'il aille sur node1) **après avoir libéré node1** (le MC « Test »
  extreme = 16 Go bouffe la RAM ; ~5 Go dispo seulement → un Hytale 10 Go n'y rentre pas).
- [ ] **Tester le Discovery** : coller le token (« EDIT SERVER LISTING » sur hytale.com) dans la carte
  « Listing public » → `discovery link <token>` via stdin → vérifier dans les logs. (Code prêt+déployé,
  jamais exécuté en live.)
- [x] (Frontend) plancher **10 Go** Hytale reflété (`lib/games.ts` : `minRamGb: 10`).

**Étude de marché Hytale (réflexion consignée) :** `docs/hytale-analyse-marche.md` (verdict : payant
non rentable — produit gratuit en self-host, marché saturé, plancher RAM non-surbookable ; mais
demande réelle → **vitrine + events sponsorisés sur cloud à l'heure** ~2-10 €/event). Voir aussi
backlog **#S** (events + sponsoring YouTubers), **#R** (mining `docs/mining-feasibility.md`),
**#T** (anti-DDoS). Cibles créateurs : `PARTENAIRES.md` §1.bis. POC cloud : `scripts/event-server/`.

**✅ Launch-readiness « premier utilisateur » (vérifié 2026-06-13)** : le funnel **Minecraft gratuit
marche de bout en bout**. Home `/`→`/games/minecraft` (200, hero+CTA), inscription Discord OK,
création serveur gratuit OK, **connexion externe PROUVÉE** (DNS `*.servers.vbt-prog.com`→IP publique ;
port **25565 forwardé & atteint par un vrai client externe** ; routes mc-router OK : node1 par nom de
conteneur, node2 par `10.0.0.110:port`), **support → courriel livré** (HTTP 200). ⚠️ Connexion MC =
**sous-domaine obligatoire** (mc-router route par hostname, pas l'IP brute — le panel affiche déjà la
bonne adresse ; idée : ajouter un *default-server* mc-router pour l'IP brute). Blog **étendu à 9
articles SEO** (Hytale x3, Minecraft x4, Satisfactory, QC) → trafic organique. **Conclusion : le
produit est livrable ; le maillon manquant = l'ACQUISITION** → kit de lancement prêt à coller
(messages Reddit/Discord/FB) dans **`docs/lancement-premier-user.md`**. Actions Simon : poster ces
messages + sécurité (révoquer token Cloudflare / régénérer bot Discord).

## ✅ Session 2026-06-13/14 — Récap (tout déployé sur xe80dell)
*Grosse session produit. Détails complets dans les entrées backlog correspondantes (#G→#Z).*

**Hytale rendu jouable & validé** : booté en réel (OAuth Hytale **vivant**), parser auth corrigé
(URL 1-clic), **seed** = jeu pré-téléchargé une fois (`scripts/hytale-seed/`, FAIT) → plus de
re-download ; **plancher 10 Go / 4 cœurs** (2 cœurs laggait, corrigé) ; **belle carte d'auth** (le
client autorise) + **console live** pendant l'auth ; **console stdin** + **listing Discovery** (token
`discovery link`). Test réel via panel : créé → autorisé → `running` + « Universe ready! ». ⚠️ Le test
a atterri sur **node2** ; connexion joueur = **NAT routeur** (action Simon).

**Placement & ports** : jeux IP-directe (Hytale/Satisfactory) **épinglés node1** (`PrimaryNode`) → NAT
simple (1 plage → 1 IP). Ports déjà en blocs sans chevauchement.

**/admin enrichi (#M/#W)** : ⚙ **éditer RAM/CPU par serveur** (à chaud) ; **configs de base par jeu**
(planchers éditables, Redis) ; **promo globale** (#Y, slider %/durée) ; **métriques business** (revenu,
conversion, par jeu — **exclut le compte proprio `avabata`**) ; **sondage visiteurs** (#Z, agrégats).

**Promo (#Y)** : rabais global %/durée piloté /admin → tous prix payants (affichage **SSR** + prix
**barrés** + facturation PayPal). Stocké Redis. **Actuellement -75 %** (réglé par Simon, ~10 j).

**Sondage (#Z)** : widget « 💬 Ton avis ? » (note, canal d'acquisition, usage, jeu, commentaire) →
table `feedback` → agrégats /admin. Pour savoir **qui accroche / quoi améliorer**.

**Site / acquisition** : compteur joueurs 0/0 corrigé (MC only) ; **header conscient de la session**
(« Tableau de bord » si connecté) + **bouton déconnexion** ; nav blog/jeux depuis le dashboard (#X) ;
**bloc plugins Minecraft** (upload/list/delete `.jar` + restart, #L v1) ; **blog → 9 articles** en
**sections par jeu** ; **kit de lancement** (`docs/lancement-premier-user.md`) + analyse marché
(`docs/hytale-analyse-marche.md`) + faisabilité mining (`docs/mining-feasibility.md`) + POC events cloud
(`scripts/event-server/`).

**Infra** : **`scripts/deploy.sh`** (déploiement reproductible api/monitor/frontend, encapsule env/mounts
/réseaux — utiliser ça). Frontend **doit** être lancé avec `-e INTERNAL_API_URL=http://mcserver-api:8080`
(SSR promo). CI GitHub Actions ajoutée (`.github/workflows/ci.yml`, #H).

**⚠️ RESTE = actions de Simon (hors code)** : (1) **sécurité** — révoquer token Cloudflare + régénérer
bot Discord ; (2) **NAT routeur** UDP+TCP `25566-26565 → 10.0.0.2` (connexion joueurs externes) ;
(3) **poster le kit** de lancement (le maillon manquant pour le 1ᵉʳ user). Stockage promo/gamecfg = Redis
(non persisté si flush Redis — re-régler via /admin si besoin).

## Backlog à traiter (demandé le 2026-06-11)

- ✅ **#A — `mcserver-monitor` REDÉPLOYÉ** (2026-06-11) : image `mcserver-monitor:latest`
  rebuildée + container recréé (env `HOST_ROOT`/`DOCKER_NODES`/`MONITOR_PORT`/`DATABASE_URL`,
  mounts docker.sock + `/`→`/host` + `secrets/ssh`, réseaux `sgportfolio_portfolio-net` +
  `sgrentmcserver_sgrent`). Embarque la refonte #M/#C ci-dessous. Validé live.
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
- **#C — `/admin` : robustesse des actions container** (clarifié 2026-06-11) : appliquer
  partout la même robustesse aux actions start/stop/restart/delete (et delete serveur) :
  état « en cours » (bouton désactivé + spinner), gestion d'erreurs lisible, **toasts** de
  succès/échec au lieu des `alert()`, pas de double-clic. Lié à #M (même page admin).
  - [x] Toasts + états de chargement + désactivation des boutons pendant l'action.
  - [x] Erreurs réseau/HTTP affichées proprement (message serveur remonté).
- ✅ **#D — SGPortfolio : section « Téléchargements » filtrée par tag semver `v*.*.*`** (2026-06-11) :
  - `src/lib/github.ts` : `getReleases()` filtre sur `/^v\d+\.\d+\.\d+/i` (mergé dans `live`,
    commit `4f16e15`) → seules les releases `v<major>.<minor>.<patch>` s'affichent.
  - `src/app/projects/[slug]/page.tsx` : bouton « Télécharger » **masqué** si le projet n'a
    aucune release valide (`hasDownloads = releases.length > 0`) → un projet sans release
    `v*.*.*` n'expose plus de téléchargement. Commité+poussé (`live`).
- ✅ **#E — Nouveau site « SG Studio » (sites web propulsés par IA)** — EN LIGNE (2026-06-12) :
  - **Nouveau repo** `~/Shared_Projects/2026/SGWebStudio` (git, branche `main`, commit initial).
    Next.js 15 standalone, Tailwind, marque/contenu centralisés dans `src/lib/site.ts`
    (renommable en 1 fichier). Domaine cible **`studio.vbt-prog.com`**.
  - **Landing** : hero, avantages, services, process, contact. **Formulaire de contact**
    `POST /api/contact` → **nodemailer** vers SMTP `mailserver:587` (auth `contact@mcserver.vbt-prog.com`,
    leads → `contact@vbt-prog.com`, honeypot anti-spam, validation). `.env` (gitignored) en place.
  - **Container `sgwebstudio`** up (image `sgwebstudio:latest`, réseaux `sgportfolio_portfolio-net`
    + `sgmail_default`, `--restart unless-stopped`).
  - **nginx** : bloc 443 `studio.vbt-prog.com` → `sgwebstudio:3000` + ajout au redirect 80
    (`SGPortfolio/docker/nginx/nginx.conf`, restart `portfolio-nginx`, `nginx -t` OK).
  - **Lien depuis le portfolio** (`Header.tsx` → « SG Studio »).
  - ✅ **Validé en e2e via l'origine** (Host header) : homepage OK + **envoi réel du formulaire
    de contact `{"ok":true}`** (courriel livré). SGPortfolio + nginx commités/poussés sur `live`.
  - ✅ **DNS Cloudflare** A/AAAA `studio` → Cloudflare (proxied) : `https://studio.vbt-prog.com`
    répond **200 en public**, contenu OK.
  - ✅ **Repo GitHub `aVaBaTa/SGWebStudio`** (privé) créé via l'API (token `repo` du portfolio)
    + `git push -u origin main` (scaffold + budget 300–1200 $). Remote `git@github.com:aVaBaTa/SGWebStudio.git`.
  - ✅ Budget formulaire ajusté à **300–1200 $**. Portfolio **rebuildé** (lien « SG Studio » live).
  - **Plus tard (cf. `SGPortfolio/PORTFOLIO.md`)** : **refonte visuelle (plus beau + original)**,
    visuels/réalisations, grille de prix, i18n. Doc dédié = `SGWebStudio` + voir PORTFOLIO.md.
- **#F — Conversion « douce » Google Ads (inscription / 1ʳᵉ création de serveur)** : en plus
  de l'achat payant (`/merci`, `AW-18226964787/7YNXCMudhLwcELPSpfND`), envoyer un événement
  de conversion **secondaire** à la création d'un serveur gratuit (ou au login Discord),
  **valeur fixe 1,0 CAD** (budget assumé par Simon pour des conversions gratuites). But :
  donner du signal à Google Ads (sinon ~0 conversion car tout passe par le gratuit).
  Mettre l'achat en conversion *primaire*, l'inscription en *secondaire*. Code : nouvel
  événement `gtag('event','conversion', {send_to: <2e libellé Ads>, value:1.0, currency:'CAD'})`
  déclenché côté frontend après création serveur (ex. dans `dashboard/[game]` ou au retour
  d'auth Discord). ⚠️ créer d'abord la 2ᵉ action de conversion dans le compte Google Ads.
  Aussi (lié) : la valeur de la conversion d'achat est codée en dur à 1,0 — pourra refléter
  le vrai prix du plan plus tard.
- 🟡 **#G — Vérifier le fonctionnement des serveurs Hytale** : valider de bout en bout
  (création → OAuth device-code → `running` → connexion joueur réelle). Lié au blocage UDP.
  **Validé (2026-06-11) par boot réel du container `ghcr.io/terkea/hytale-server`** :
  - ✅ L'image boote, télécharge, et atteint l'OAuth. **Les serveurs OAuth Hytale sont VIVANTS**
    (`downloader.hytale.com` répond, device-code réel obtenu — Hytale annulé mais l'infra auth tourne).
  - ✅ Le port **UDP se bind** côté hôte (`0.0.0.0:<base>/udp`).
  - ✅ Le parser d'auth (`parseAuthFromLogs`) extrait **URL + code** correctement sur de **vrais logs**
    (tests permanents `internal/server/hytale_realboot_test.go` : step 1 + step 2 + ligne « prêt »).
  - 🐛 **Corrigé** : le parser renvoyait l'URL **nue** (`/device/verify`) au lieu de l'URL **complète**
    (`?user_code=…`, autorisation en 1 clic) — la version nue écrasait la complète ; l'URL de MAJ
    `downloader…zip` matchait aussi le filtre. ⇒ préférence URL-avec-code + exclusion du bruit `.zip`.
    **Déployé** (`mcserver-api` rebuild 2026-06-11, restart=0, `/auth/discord`→307 OK).
  - ⛔ **Non automatisable ici** : l'autorisation du device-code (step 2) exige un **compte Hytale
    licencié** (action humaine), et la **connexion joueur réelle** dépend du **pare-feu UDP** (sudo).
  - ⚠️ Limite connue : si l'auth échoue, l'entrypoint démarre le serveur **non authentifié** et le
    watcher peut le passer `running` (regex « listening ») alors que les joueurs ne peuvent pas se
    connecter. À durcir plus tard si besoin.
- 🟡 **#H — Ajouter un CI/CD** : pipeline de build/test/déploiement (GitHub Actions ?) pour les
  repos (au moins SGRentMcServer : backend Go + frontend Next.js).
  - ✅ **CI ajoutée (2026-06-14)** : `.github/workflows/ci.yml` — job **backend** (`go build`/`vet`/`test`,
    Go 1.25) + job **frontend** (`npm ci` + `tsc --noEmit` + lint non bloquant). Se déclenche sur push
    `live`/`Claude/**` + PR. ⚠️ S'active **au prochain push** (Actions activé par défaut sur GitHub).
  - [ ] **CD** (déploiement auto) non fait : le déploiement reste manuel sur xe80dell (cf. « Déploiement »).
- **#I — Faire en sorte que Satisfactory fonctionne** : connexion joueur réelle (dépend du
  pare-feu UDP, cf. blocage connu ci-dessous) + validation bout en bout.
- **#J — S'inspirer de la concurrence** : prendre référence sur
  https://shockbyte.com/games/minecraft-server-hosting **et un autre site** (à choisir) pour
  améliorer offre, UX, pricing, pages jeux.
- **#K — Publicité dans les différents éléments** : insérer de la pub / mises en avant dans
  les composants du site (bannières, cross-sell entre jeux, promos, etc.).
- 🟡 **#L — Mods et plugins pour Minecraft** : permettre l'ajout de mods/plugins.
  - ✅ **Gestionnaire de plugins v1 (2026-06-14, déployé)** : section **« Plugins & mods (Paper) »**
    sur la page serveur Minecraft (`dashboard/servers/[id]`) — **liste** les `.jar` du dossier
    `/plugins`, **téléverse** un `.jar` (mkdir `/plugins` auto + `POST /files/upload?path=/plugins`),
    **supprime**, et **bouton « Redémarrer pour appliquer »**. Réutilise l'API fichiers existante
    (`/files`, `/files/upload`, DELETE `/files`, `/files/mkdir`). i18n FR/EN. Minecraft only.
  - ✅ **Serveurs MODDÉS Fabric/Forge (2026-06-14, déployé + vérifié live)** : choix du **type de
    serveur** (loader) à la création — **Paper** (plugins, défaut) / **Fabric** / **Forge** (mods).
    - **DB** : colonne `loader` (migration **`006_server_loader.sql`**, défaut `paper`, backfill OK).
      `GameServer.Loader` threadé partout (repo INSERT + 4 SELECT/Scan).
    - **Backend** : `minecraftEnv(plan, version, loader, _)` → `TYPE=PAPER|FABRIC|FORGE` (l'image itzg
      installe Fabric/Forge auto pour la VERSION). Signature `GameDef.Env` élargie (loader), `buildSpec`
      passe `gs.Loader`. Helpers `servers.NormalizeLoader`/`IsModded`/`MinecraftLoaders`. `createServerRequest.Loader`
      validé/normalisé (jeux ≠ minecraft → ignoré).
    - **Frontend** : sélecteur **Paper/Fabric/Forge** sur `dashboard/[game]` (Minecraft) + hint i18n ;
      la section fichiers cible **`/mods`** (Fabric/Forge) ou **`/plugins`** (Paper) selon le loader,
      titre « Mods (Fabric/Forge) » vs « Plugins (Paper) » (i18n `srv.moddedTitle/moddedDesc`).
    - **Vérifié E2E** : serveur **Fabric** créé via l'API → `TYPE=FABRIC`, `Installing Fabric Loader
      0.19.3`, `fabricloader 0.19.3` + `Done`. Dossier `/data/mods` créé. **Mod réel** (fabric-api)
      déposé dans `/mods` + restart → **« Loading 41 mods »** (= fabric-api + sous-modules chargés).
      Artefacts de test supprimés.
  - ✅ **Navigateur de mods/plugins + install 1-clic Modrinth (2026-06-14, déployé + vérifié live)** :
    `internal/server/handlers_mods.go` — `GET /servers/{id}/mods/search?q=` (Modrinth, sans clé ;
    facets selon loader : `project_type:plugin`+`paper` ou `project_type:mod`+`fabric`/`forge`, +
    version si ≠ LATEST) et `POST /servers/{id}/mods/install {project_id}` (résout la version
    compatible → télécharge le jar → `node.WriteFileReader` dans **`/plugins`** (Paper) ou **`/mods`**
    (moddé), `node.Mkdir` best-effort). Frontend : table de résultats (icône/titre/auteur/downloads +
    bouton **Installer**) dans l'onglet Mods & Plugins. **Vérifié E2E** : search « worldedit » → 20
    plugins ; install Chunky → `{"dir":"/plugins","name":"Chunky-Bukkit-1.5.3.jar","status":"ok"}` +
    jar présent dans le container. Artefact nettoyé.
  - ✅ **Page serveur refondue en ONGLETS (2026-06-14, déployé)** : `dashboard/servers/[id]` passe d'un
    long scroll à des onglets cliquables — **Aperçu** (infos+actions, auth Hytale) / **Console**
    (console + joueurs + discovery Hytale) / **Fichiers** / **Mods & Plugins** (Minecraft only : liste +
    upload + navigateur Modrinth) / **Paramètres** (plan, version MC, danger zone). State `tab`, barre
    d'onglets sous la nav, sections gatées par `tab===…`. i18n `srv.tab*` FR/EN.
  - ✅ **Modpacks tout-en-un FTB + Modrinth (2026-06-14, déployé + vérifié live)** : install 1-clic d'un
    pack complet façon launcher FTB. `internal/server/handlers_modpacks.go` — `GET …/modpacks/search?q=`
    (Modrinth `project_type:modpack` + **FTB** `api.modpacks.ch` : populaires si q vide, sinon recherche ;
    top 6 FTB pour borner le N+1) et `POST …/modpacks/install {source,id,version_id}`. Installer un
    modpack = **recréer le serveur** : colonne DB **`modpack`** (migration `007`, `"ftb:<id>:<ver>"` /
    `"modrinth:<id>:<ver>"`), `minecraftEnv` émet `TYPE=FTBA`+`FTB_MODPACK_ID/VERSION_ID` ou
    `TYPE=MODRINTH`+`MODRINTH_MODPACK/VERSION` (sinon TYPE=loader). RAM relevée à **4 Go mini** à
    l'install. `source:"none"` → retire le pack (retour Paper). Réutilise `recreateServer` (volume/monde
    conservé). Frontend : section **« Modpacks (FTB · Modrinth) »** dans l'onglet Mods (recherche +
    populaires auto + badge source + Installer + retrait), i18n `srv.modpack*`. **Vérifié E2E** : search
    → 15 packs ; install « Create+ » → container recréé `TYPE=MODRINTH MODRINTH_MODPACK=B2nZ0LBO`, RAM
    4 Go, logs itzg `Downloading modpack … Fabric 1.20.1` + mods (fabric-api, jei, jade…). Serveur de
    test supprimé. (FTB : même chaîne `TYPE=FTBA`, vérifié par construction — testable en live au besoin.)
  - ✅ **Tuto blog + aide install côté joueur (2026-06-14, déployé)** : (1) nouvel article blog
    **`/blog/ajouter-mods-serveur-minecraft`** (plugins vs mods, ajout 1-clic, modpacks, **comment les
    joueurs installent le pack côté client** via launchers) — `app/blog/posts.ts` + `POST_GAME` (section
    Minecraft, 10 articles MC). (2) **Carte « Pour tes joueurs »** sur la page serveur (onglet Mods,
    visible si moddé/modpack) : **lien vers la page du pack** (Modrinth `modrinth.com/modpack/<id>` ou
    FTB), bouton **« Copier les instructions joueurs »** (texte prêt à partager : lien + adresse de
    connexion) + lien vers le tuto. Rappelle que plugins Paper = rien côté client, mods/modpacks =
    chaque joueur installe le pack. i18n `srv.client*`. Vérifié live (article 200 + listé, carte dans le bundle).
  - 🐛 **FIX image Java = f(version MC) + optimisations auto moddés (2026-06-16, déployé + vérifié)** :
    le bon Java **dépend de la version de Minecraft**, pas du loader. MC récent (CalVer **26.x** / LATEST)
    est compilé **Java 25** et ne tourne PAS sous 21 (`UnsupportedClassVersionError class 69 vs 65`) ;
    les packs 1.20/1.21 exigent **Java 21** et refusent 25 (`requires version 21`). → `servers.MinecraftImage(version)`
    + `javaTagForVersion` : LATEST/26.x+ → `latest`, 1.x → `java21`. `buildSpec` choisit l'image selon la
    **version**. **Modpack install** résout loader **ET version MC réelle** (`resolveModpackMeta`, Modrinth
    `game_versions` / FTB target `game`) → stockés en DB → bonne image Java + bon dossier `/mods` + recherche
    de mods. **Optimisations auto serveurs moddés** : `VIEW_DISTANCE=8`/`SIMULATION_DISTANCE=6` (modded+modpack) ;
    pour les **Fabric/Forge DIRECTS** (hors modpack), `MODRINTH_PROJECTS=lithium,ferrite-core,c2me-fabric,krypton`
    + `MODRINTH_ALLOWED_VERSION_TYPE=alpha` (C2ME = alpha-only ; itzg résout la version) → **C2ME = génération
    de chunks parallèle**. JAMAIS d'injection de mods dans un modpack (doublon = crash). Plancher CPU moddé
    `MinCPUForModded=4` (plafond, pas de coût idle). Front : `isModded`/`modDir` incluent `modpack` ;
    `effectiveLoader` (modpack non résolu → fabric) pour la recherche de mods.
    **Vérifié E2E** : Fabric LATEST → `latest`/java25, `Loading 38 mods`, **Done (6.8s)**, C2ME+Lithium+
    FerriteCore+Krypton auto, 0 restart. Modpack Create+ → loader=fabric, **version 1.21→1.20.1 résolue**,
    image **java21**. ⚠️ Cobblemon : monde corrompu par les anciens crashs → `rm /data/world` + recréé OK.
    Gotchas appris : slug `ferrite-core` (pas `ferritecore`), C2ME alpha-only, MC version → Java.
  - ✅ **Sauvegarde du monde 1-clic (2026-06-16, déployé + vérifié)** : bouton **« Télécharger le monde
    (.tar.gz) »** (onglet Aperçu, Minecraft) → `GET /servers/{id}/world/download`. `Node.DownloadDir`
    (CopyFromContainer = flux tar du dossier) gzip-é à la volée → archive `<subdomain>-world.tar.gz`.
    Marche serveur arrêté aussi. Vérifié sur Cobblemon : `world/` complet (DIM-1/DIM1/advancements…),
    50 Mo gzip. ⚠️ Pour une sauvegarde 100 % cohérente, arrêter le serveur avant (MC sauvegarde en live).
    [ ] Idée : restore (upload d'un `.tar.gz` → /data/world) + snapshots auto.
  - ⚠️ **Limite modpacks (pas notre bug)** : certains packs « single-player » embarquent des mods
    **client-only** (litematica, malilib…) qui crashent un serveur dédié (`NoClassDefFoundError malilib`).
    Les packs conçus serveur (Cobblemon) bootent ; pour les autres, retirer les mods client via l'onglet
    Mods (dossier `/mods`). [ ] Idée : blocklist auto de mods client connus à la création de modpack.
    Très vieux packs (1.16 = Java 8) non couverts (java21 gère 1.18–1.21).
  - [ ] **Reste** : source **CurseForge** (clé API requise), choisir un modpack **à la création**
    (actuellement via recréation d'un serveur existant), redémarrage auto après install de mod isolé.
- **#M — `/admin` enrichi (multi-pages + métriques + mini-Jira)** : plus d'options dans le
  temps + **graphiques**, passer de **une page à plusieurs** (multi-vues navigables), pouvoir
  **supprimer / ajouter des serveurs pour des utilisateurs**, exposer **le plus de métriques
  possible de manière digérable** (business, usage, **conversions, coûts annonces**, etc.).
  Le `/admin` = binaire Go `cmd/monitor` (HTML embarqué `cmd/monitor/static/index.html`,
  pkg `internal/monitor`), service séparé de l'API (Docker + DB **read**, pas d'orchestrateur).
  Sous-tâches (mini-jira en .md tant que le board n'est pas codé) :
  - [x] **Multi-pages / navigation** : onglets Dashboard / Serveurs / Containers / Nodes /
        Logs / Utilisateurs (SPA, une vue affichée à la fois).
  - [x] **Graphiques historiques** : buffer en mémoire dans le monitor (ring buffer ~échantillon
        / 20 s sur 2 h), endpoint `GET /api/history`, courbes (CPU/RAM cluster, joueurs,
        serveurs up). Pas de table DB (suffisant pour tendances court terme).
  - [x] **Logs des différents nodes/containers** : endpoint `GET /api/containers/{id}/logs?tail=`
        (via le client Docker du bon node) + viewer dans l'onglet Logs.
  - [x] **CRUD serveurs — DELETE** : supprimer un serveur (container + ligne DB `game_servers`)
        depuis l'admin, confirmation forte. Endpoint `POST /api/servers/{id}/delete`.
  - [x] **CRUD serveurs — CREATE depuis /admin** (2026-06-12, EN LIGNE) : onglet Utilisateurs →
        bouton **« + Serveur »** (modale jeu/plan/nom) crée un serveur **au nom du user** via
        l'orchestrateur de l'API. Endpoints admin internes API (`/api/v1/admin/catalog`,
        `/servers`, `/users/{id}/unlimited`) protégés par **`X-Admin-Token`** (`ADMIN_TOKEN`,
        partagé monitor↔API) ; le monitor proxifie (`/api/catalog`, `/api/users/{id}/create-server`,
        `/api/users/{id}/unlimited`). Cœur de création refactoré (`createServerForUser`).
        Validé e2e (create 201 + delete 204).
  - [x] **Droit « création illimitée » par user** (toggle dans /admin) : colonne
        `users.unlimited_create` (migration `003_admin_grants.sql`, owner `simong2004@hotmail.com`
        = true par défaut, **désactivable**). Permet de créer sur un plan **payant sans payer** ;
        sinon un plan payant à la création est ramené à `free` (garde dans `createServerForUser`).
  - [ ] **Mini-Jira (board tickets)** : colonnes à faire/en cours/fait, persisté en DB
        (nouvelle table `admin_tasks`), pour remplacer ce backlog .md à terme.
  - 🟡 **Métriques business** : ✅ **FAIT + déployé (2026-06-14)** — section « 💰 Métriques business »
        dans /admin → Dashboard : **revenu encaissé** (total + ce mois, status `completed`), **MRR estimé**
        (serveurs payants actifs × prix plan), **paiements en attente**, **users payants / total +
        taux de conversion**, **serveurs payants/gratuits + par jeu**, **10 derniers paiements** (statut).
        Backend `handleAdminMetrics` (`handlers_metrics.go`, agrège la table `payments` + `game_servers`),
        endpoint `GET /api/v1/admin/metrics`, rafraîchi 30 s. **Exclut les comptes proprio/tests** (config
        `METRICS_EXCLUDE_USERS`, défaut `avabata`) → les serveurs/paiements/users de `aVaBaTa` ne sont PAS
        comptés (métriques = vrais clients seulement). Pour ajuster : env `METRICS_EXCLUDE_USERS=a,b,c`. [ ] **Reste** : **coûts annonces Google Ads**
        (nécessite l'API Google Ads / saisie manuelle), **churn** (besoin d'un suivi abonnements). Lié #F/#P.
- **#N — Visuels** : ajouter des **images de Minecraft** (et des autres jeux) sur le site.
- **#O — Serveur communautaire / partenaires** : monter un serveur communautaire et/ou
  **trouver des partenaires** pour héberger des serveurs.
- **#P — Vérifier les paiements PayPal** : s'assurer que le flux de paiement PayPal
  fonctionne de bout en bout.
- **#Q — Plus de méthodes de paiement et de connexion** : ajouter d'autres moyens de
  paiement et d'autres providers d'authentification (au-delà de Discord / PayPal).
- **#R — Étudier un modèle « mining de serveurs » rémunéré (hébergement distribué)** :
  évaluer la faisabilité d'un réseau où des **particuliers prêtent leur machine** (CPU/RAM/
  bande passante) pour héberger des serveurs de jeu Playrena et sont **rémunérés** en retour
  (modèle type Salad / Honeygain / pool de hosting communautaire). À creuser :
  - **Technique** : agent à installer chez le contributeur (Docker + tunnel sortant pour
    contourner NAT/CGNAT, ex. WireGuard/FRP/Cloudflare Tunnel ; **le jeu UDP ne passe pas par
    CF** → vrai défi réseau), orchestrateur multi-node (l'archi `internal/orchestrator` gère
    déjà node1+node2 via socket/SSH → étendre à N nodes non fiables), placement/healthcheck,
    isolation/sécurité (on exécute du code chez des tiers et inversement).
  - **Économique** : payer les contributeurs (combien/Go/h ?), vs coût d'un VPS ; le gain =
    capacité élastique sans capex (utile justement pour gros pics Hytale, cf. #S).
  - **Confiance/abus** : latence variable, uptime non garanti, triche/vol de données, fiscalité.
  - **Verdict attendu** : probablement un **gros projet R&D** ; commencer par une note de
    faisabilité (1 page) avant tout code. Lié à #O (partenaires) et #S (capacité événementielle).
  - ✅ **Note de faisabilité écrite : `docs/mining-feasibility.md`** (verdict : faisable mais
    gros R&D + économie douteuse pour du temps réel → préférer cloud à l'heure #S + nodes
    partenaires de confiance #O via l'orchestrateur existant ; ne pas coder maintenant).
- **#S — Capturer la clientèle Hytale (loss-leader) + sponsoring YouTubers events** : décision
  produit assumée par Simon. 📄 **Analyse complète : `docs/hytale-analyse-marche.md`** (marché,
  éco hosting, RAM, events, anti-DDoS, reco). **Aller chercher les joueurs Hytale même à perte** (prix très bas
  ou gratuit) pour la croissance/notoriété, puis **démarcher des YouTubers/streamers** qui
  organisent des **events 100 joueurs** en leur offrant la **location de serveurs 25/50/100
  joueurs en échange de sponso/visibilité**. ⚠️ **Contraintes dures à intégrer (cf. mini-étude
  2026-06-11)** :
  - **Matériel** : `xe80dell` = **31 Go RAM total (~21 Go libres)**, 48 cœurs. Hytale =
    plancher 4 Go + **~1 Go/joueur** ⇒ 25 j ≈ 25 Go (déjà > RAM libre), 50 j ≈ 50 Go, 100 j
    ≈ 100 Go. **La machine actuelle ne peut PAS héberger un 25/50/100 joueurs Hytale.**
  - **Voie réaliste pour les events** : **cloud à l'heure** provisionné par event, PAS sur
    xe80dell, PAS de dédié mensuel (MAU Hytale en baisse → serveurs vides entre events).
    **Chiffrage (2026-06-12)** — Hetzner Cloud, facturé à l'heure plafonné au mois :
    CCX33 32 Go **0,10 €/h** (~25 j) · CCX43 64 Go **0,20 €/h** (~50 j) · CCX53 **128 Go
    0,40 €/h** (~100 j) · CCX63 192 Go 0,60 €/h. ⇒ **un event 100 joueurs de 5 h ≈ 2 €**,
    24 h ≈ 9,60 €, week-end ≈ 20 €. **Le serveur n'est pas le coût** — le vrai livrable est
    l'**automatisation** spin-up/tear-down via API (louer → Docker + image Hytale → auth + DNS
    → détruire). Dédié mensuel seulement si demande récurrente (Hetzner AX102 128 Go ~119 €/mo,
    auction 50-90 €/mo, OVH Scale dès ~513 $/mo). **Latence** : OVH **Beauharnois (Montréal)**
    pour audience QC/FR ; Hetzner EU/US-East sinon. Le « mining » #R = autre voie d'élasticité.
  - **Sur xe80dell** : ne viser que des **petits Hytale (4–8 Go, 4–15 joueurs)** comme funnel
    gratuit/à bas prix.
  - **Marché (signaux mitigés)** : ✅ lancement énorme (Twitch 420 k peak, 1 M mods en 48 h,
    Hypixel financé ~2 ans), modding fort = serveurs gourmands = valeur de location ; ⚠️ **MAU
    déjà en baisse** (~740 k, **-13,6 % m/m** — hype qui refroidit), prix plancher ~0 (self-host
    gratuit + tiers gratuits 20 slots), marché saturé de **partenaires officiels** (Apex,
    Shockbyte, Pingperfect), latence mono-région défavorable.
  - **À livrer** : (1) liste de YouTubers/streamers Hytale events FR+EN à démarcher, (2) offre
    sponso type, (3) chiffrage VPS event, (4) page/offre « events » sur le site.
    ✅ (1)+(2) **faits** : section **`PARTENAIRES.md` §1.bis** (cibles Hytale events + offre type).
    ✅ (3) **fait** : chiffrage ci-dessus + **POC de provisioning `scripts/event-server/`**
    (provision.sh/destroy.sh + cloud-init ; ~2-10 €/event ; **non testé live** — besoin d'un
    token Hetzner). [ ] (4) page « events » sur le site = reste à faire.
    **Avant tout démarchage** : valider qu'on peut *techniquement* livrer un 100 joueurs (lancer
    le POC avec un vrai token Hetzner) — ne pas promettre ce qu'on ne peut pas tenir.
- **#T — Anti-DDoS (protection + masquage de l'IP d'origine)** : aujourd'hui le trafic de jeu
  tape l'**IP résidentielle d'origine** (Cloudflare ne proxifie pas l'UDP/TCP de jeu) → une
  attaque peut **couper Internet de toute la maison**. Risque réel à mitiger. Pistes :
  - **Minecraft (TCP)** : **TCPShield** (plan gratuit) — proxy qui absorbe les attaques et
    **cache l'IP d'origine** ; s'intègre via le hostname mc-router. Le plus rentable, à faire en 1er.
  - **Jeux UDP (Satisfactory/Hytale)** : pas de proxy gratuit → s'appuyer sur la **protection
    niveau hébergeur** quand on est sur cloud (**OVH = anti-DDoS Game inclus gratuitement**,
    Hetzner = protection de base incluse). Argument de plus pour faire les events sur OVH/Hetzner
    plutôt qu'à la maison.
  - **Option payante** : **Cloudflare Spectrum** (proxy TCP/UDP brut, masque l'origine) — cher
    (enterprise), à garder pour plus tard / gros clients.
  - **Court terme maison** : au minimum, ne JAMAIS exposer l'IP résidentielle pour des events
    publics ; passer par un VPS relais. Lié à #S (cloud) et #R (nodes distribués).
- **#U — Étendre le catalogue de jeux** (demandé le 2026-06-12) : ajouter de nouveaux jeux au
  registre (`internal/servers/games.go` + `frontend/lib/games.ts` + i18n/PLANS + page `/games/<jeu>`).
  📄 **Specs techniques complètes (image Docker, ports `depuis base`, RAM plancher, volume, auth,
  pièges, capacité xe80dell) : `docs/jeux-populaires-a-ajouter.md`.** Jeux demandés : **Palworld,
  Terraria, The Forest, 7 Days to Die, Valheim, ARK: Survival Evolved, ARK: Survival Ascended,
  Stardew Valley** (+ autres candidats : Rust, Factorio, Project Zomboid, V Rising, Enshrouded…).
  **Séquençage recommandé** (par effort/valeur/empreinte) : **Valheim → Terraria → 7 Days to Die**
  (natifs Linux, légers/moyens, forte demande) ; puis **lourds** (Palworld, ARK SE, ~1 serveur à la
  fois comme Hytale) ; **Wine** (The Forest, ARK SA — effort élevé, ARK SA plutôt **cloud à l'heure**
  cf. #S) ; **Stardew** en dernier (réseau Steam P2P ≠ modèle IP:port → cas particulier). ⚠️ Pas de
  **port fixe** type `27015` en dur (collision multi-instances) → dériver tous les ports depuis `base`.
- **#V — Redistribution dynamique des ressources des serveurs peu utilisés (overcommit/idle)** (demandé
  2026-06-12) : **étudier** un système qui, quand un serveur est **quasi inactif** (0 joueur, peu de
  CPU/RAM), **réduit/redistribue ses ressources de façon transparente** (le client ne le sait pas), et
  les **restaure** dès qu'il redevient actif. Motivation concrète : un Hytale réserve **10 Go** mais
  n'en utilise que **~1,3 Go au repos** (mesuré 2026-06-12) → énorme RAM gelée pour rien. Pistes :
  - **Overcommit RAM contrôlé** : abaisser la limite cgroup (`docker update --memory`) d'un serveur
    idle, la relever à l'activité ; surveiller pour éviter l'OOM (ne jamais descendre sous l'usage réel
    + marge). Permettrait **plus de serveurs concurrents** sur xe80dell (la vraie contrainte = RAM).
  - **CPU** : déjà partagé (NanoCPUs = plafond, pas réservation) — un serveur idle ne consomme pas ses
    cœurs ; surtout pertinent pour la RAM.
  - **Idle suspension/hibernation** : stopper/`pause` un serveur sans joueur depuis X min, le relancer
    à la 1ʳᵉ tentative de connexion (nécessite un proxy qui réveille — type *lazymc* pour Minecraft, à
    inventer pour l'UDP). Le plus gros gain mais le plus complexe (réveil UDP).
  - **Détection de joueurs = signal central** (demandé 2026-06-12) : savoir si un serveur a des
    joueurs connectés pilote **l'attribution réelle des ressources Docker**. Idée : un serveur **sans
    joueur** ne se voit attribuer qu'un **minimum** (RAM/CPU réduits via `docker update`), et reçoit
    sa **pleine allocation dès qu'un joueur se connecte** ⇒ **beaucoup plus de serveurs simultanés sur
    la même machine** (on n'immobilise la RAM que pour les serveurs réellement joués). Sources de
    détection : joueurs (RCON/MC ; logs/console Hytale ; query UDP par jeu), CPU/RAM réels (`docker
    stats`, déjà dans le monitor #M), I/O réseau, connexions au port.
  - ⚠️ **Risques** : OOM si on coupe trop, lag au « réveil », perception client (« mon serveur rame
    après une pause »). **Commencer par une note de faisabilité** + un overcommit RAM **prudent** sur
    les serveurs gratuits idle. Lié à #M (métriques), #W (planchers éditables) et au modèle économique
    (densité = rentabilité).
- **#W — Édition des ressources (CPU/RAM) depuis `/admin`** (demandé 2026-06-12) :
  - ✅ **FAIT + déployé (2026-06-12) — override par serveur** : bouton **⚙** dans la vue Serveurs de
    `/admin` → édite RAM (Go) + CPU (cœurs) d'un serveur précis, **appliqué à chaud** (`docker update`)
    + persisté en DB. Chaîne : monitor `POST /api/servers/{id}/resources` → API `POST
    /api/v1/admin/servers/{id}/resources` (token admin) → orchestrateur `UpdateResources` (existait
    déjà) → cgroup du conteneur (marche cross-node, testé sur node2 : 4→3→4 cœurs OK) + DB. Bornes
    512 Mo–128 Go / 0.5–64 cœurs. Repo : `GetByIDAny` + `UpdateResources`.
  - ✅ **Planchers par JEU éditables depuis /admin (FAIT + déployé 2026-06-14)** : section
    « ⚙️ Configs de base des jeux » dans /admin → Dashboard → édite **RAM (Go) + CPU** par jeu
    (Hytale/Satisfactory/Minecraft), bouton Enregistrer. **Override stocké dans Redis** (`gamecfg:<id>`,
    `internal/server/gamecfg.go`) ; `createServerForUser` + l'upgrade utilisent `effectiveFloor`
    (override sinon valeurs `games.go`). Endpoints `GET /api/v1/admin/games`, `POST
    /api/v1/admin/games/{id}/config` (bornes 256 Mo–128 Go / 0.5–64 cœurs). ⚠️ S'applique aux
    **nouveaux** serveurs du jeu (les existants : bouton ⚙ par serveur). Chaîne testée (write/validation).
- ✅ **#X — FAIT (2026-06-14, déployé)** : header du dashboard enrichi de liens **Jeux** (`/games`),
  **Blog** (`/blog`), **Mods** (`/mods`) via `<Link>` (nav client, cookie de session conservé → retour
  au dashboard toujours connecté). Logo → `/dashboard`. Aussi : métadonnée blog passée en multi-jeux +
  **blog en sections par jeu** (Minecraft/Hytale/Satisfactory/Guides).
  - ✅ **Header public conscient de la session (2026-06-14)** : avant, les pages publiques affichaient
    toujours « Se connecter » même connecté → l'utilisateur *croyait* être déconnecté. Nouveau composant
    **`AuthButton`** (`site-chrome.tsx`) qui vérifie **`GET /api/v1/user/me`** (401/200) → affiche
    **« Tableau de bord »** (→`/dashboard`) si connecté, **« Se connecter »** sinon. Branché dans
    `SiteNav` (landing, `/games/*`, `/mods`) **et** les nav du **blog** (index + `[slug]`). La session
    (cookie JWT) persistait déjà ; c'était l'UI qui ne la reflétait pas.
    - ✅ **Bouton « Déconnexion » sur le header public (2026-06-14)** : connecté, `AuthButton` affiche
      désormais **« Tableau de bord » + « Déconnexion »** (avant : seulement le lien dashboard, aucun
      moyen de se déconnecter hors du dashboard). Logout → `POST /auth/logout` → redirection accueil
      (`window.location`). Réutilise l'endpoint/style du logout du dashboard. Déployé (frontend).
- **#X — Navigation blog/pages publiques depuis le dashboard connecté** (demandé 2026-06-13) : quand
  l'utilisateur est **connecté** et sur son tableau de bord (liste de ses serveurs), il doit pouvoir
  **accéder au blog (et aux autres pages publiques : jeux, plans…) en restant connecté** — sans
  perdre sa session ni devoir se reconnecter. À faire : ajouter des liens (header/nav du dashboard)
  vers `/blog`, `/games/*`, etc., et s'assurer que la session JWT (cookie) persiste sur tout le site
  (le blog est public donc déjà accessible ; le point clé = la nav y mène depuis l'espace connecté et
  le retour au dashboard reste connecté). Cohérence : header commun connecté/déconnecté.
- ✅ **#Y — Promo globale (rabais % + durée) pilotée depuis /admin** (FAIT + déployé 2026-06-14) :
  rabais appliqué à **tous les prix payants** (affichage **et** facturation PayPal). **Stocké dans
  Redis** (`promo:global`, pas de migration). **Backend** (`internal/server/promo.go`) : `GET
  /api/v1/promo` (public, rabais actif), `GET|POST /api/v1/admin/promo` (token admin, `{percent 0-90,
  days}` ; days=0 = permanent, sinon expire dans N jours) ; checkout PayPal applique `applyPromoCents`.
  **Frontend** : `LanguageProvider` fetch `/api/v1/promo` → `setPromoFactor` → `priceFor` multiplie
  tous les prix (donc **toutes les pages** se mettent à jour sans édition) + **bannière promo**
  (`PromoBanner` dans `site-chrome`, `t.promoBanner`). **Admin** : section « Promotion globale »
  (slider 0-90 % + durée jours + bouton -50% + Appliquer). Pour changer/désactiver : /admin → Dashboard
  → Promotion globale (0 % = off).
  - ✅ **SSR (2026-06-14)** : le rabais s'affiche **dès le HTML serveur** (prix réduits + bannière dans
    le HTML brut → bon pour le SEO, zéro flash). `layout.tsx` (async) fetch `/api/v1/promo` côté serveur
    (`INTERNAL_API_URL`, **timeout 2 s** sinon le build hang 60 s/page, ISR `revalidate:30`) → passé en
    `initialPromo` au `LanguageProvider` (init `useState`/`setPromoFactor`) ; re-fetch client pour rester
    frais. ⚠️ Pages statiques : juste après un build, le 1ᵉʳ rendu peut être promo=0 jusqu'à la 1ʳᵉ
    revalidation ISR — le client corrige immédiatement. Vérifié : prix CAD à -75 % dans le HTML brut.
  - ✅ **Prix barrés (2026-06-14)** : `priceFor` renvoie aussi `originalMonthly`/`originalTotal` +
    `promoActive` → quand une promo est active, le **prix régulier s'affiche barré** (`line-through`) à
    côté du prix réduit. Appliqué aux **3 pages jeux** (`/games/*`) + **sélecteur de plan** du dashboard
    (`servers/[id]`). Rendu dès le SSR (barré dans le HTML brut). i18n inchangé.
- ✅ **#Z — Sondage / feedback visiteurs (FAIT + déployé 2026-06-14)** : comprendre **qui accroche
  et quoi améliorer**. **Widget flottant « 💬 Ton avis ? »** (`FeedbackWidget` dans `site-chrome`,
  rendu global via `layout.tsx`) → modale : note (1-5★), **canal d'acquisition** (Reddit/Discord/ami/
  Google/autre), **cas d'usage** (entre amis/communauté/test/autre), **jeu d'intérêt**, commentaire
  libre, email optionnel. Stocké en DB table **`feedback`** (migration `004_feedback.sql`, appliquée
  live). Endpoints : `POST /api/v1/feedback` (public), `GET /api/v1/admin/feedback` (admin, agrégats
  par canal/usage/jeu + note moyenne + 30 derniers commentaires). **/admin → Dashboard → « 💬 Sondage
  visiteurs »** (rafraîchi 30 s). i18n FR/EN (`t.survey`). Testé E2E (POST→agrégats→affichage).
  But : mesurer le **canal qui convertit** (lié #F Google Ads) et prioriser les améliorations.
- ✅ **#AA — Tracking des pages visitées dans /admin** (demandé + **FAIT + déployé + vérifié live 2026-06-14**) :
  savoir **sur quelle page les utilisateurs (sauf aVaBaTa) étaient**. Mécanique :
  - **Beacon front** : composant `frontend/components/page-tracker.tsx` (`PageTracker`, monté dans
    `app/layout.tsx`) → `POST /api/v1/track {path,referer}` à **chaque changement de route**
    (`usePathname`), `credentials:include` + `keepalive`. Si connecté → le backend lit le **cookie JWT**
    et attache pseudo/`user_id` ; sinon visiteur **anonyme**.
  - **Backend** : `internal/server/handlers_pageviews.go` — `POST /api/v1/track` (public, best-effort,
    jamais d'erreur visible) + `GET /api/v1/admin/pageviews` (token admin). Table **`page_views`**
    (migration **`005_page_views.sql`**, user_id/username/path/referer/created_at). Agrégats **excluent
    le proprio** (`METRICS_EXCLUDE_USERS`, défaut `avabata`) : vues 24h/7j, connectés distincts,
    **connectés actifs (30 min) avec leur page actuelle**, **pages les + vues (7 j)**, **flux récent**.
  - **Monitor** : `AdminClient.PageViews` + proxy `GET /api/pageviews` + section **« 🧭 Navigation des
    visiteurs »** dans /admin → Dashboard (`cmd/monitor/static/index.html`, refresh 15 s).
  - **Déployé** : migration `005` appliquée (`psql -U sgrent -d sgrentmc`), rebuild api+monitor+frontend
    via `./scripts/deploy.sh`. **Vérifié live** : `/api/v1/track` → 204, `/api/v1/admin/pageviews` +
    proxy monitor renvoient les agrégats, lignes en DB. **Exclusion aVaBaTa confirmée** : les lignes
    `avabata` sont stockées mais **filtrées de tous les agrégats** admin (recent/by_page/counts).
  - ✅ **« En ligne maintenant » (présence temps réel, 2026-06-14)** : compteur du nombre
    d'onglets/visiteurs ayant la page ouverte. **Heartbeat** : chaque onglet envoie `POST
    /api/v1/presence {vid,path}` toutes les **20 s** (`vid` = id navigateur en localStorage → 1
    personne = 1 vid, multi-onglets dédupliqués). **Redis** : sorted set `presence:online`
    (score = dernier battement) + hash `presence:meta` ; TTL **60 s** (purge des vid sans battement).
    L'agrégat admin renvoie `online_count` + `online[]` (page de chacun), **exclut aVaBaTa**. Section
    /admin : bloc vert **« 🟢 En ligne maintenant »**. Pas de migration (tout en Redis). Vérifié live
    (2 vid → count=2, même vid rebattu → compté 1×). ⚠️ Non persisté si flush Redis (se reremplit seul).
  - **Fuseau horaire** : la DB est en **UTC**. L'API renvoie les horodatages en **ISO 8601 UTC**
    (`…Z`) et le JS de `/admin` (`fmtLocal`) les affiche dans le **fuseau du navigateur** de l'admin
    (= heure locale de qui regarde, sans fuseau codé en dur). Appliqué aux **deux** champs (actifs +
    récent). Redéployé api+monitor 2026-06-14.
  - [ ] **Rétention** : la table grossit à chaque navigation (pas de purge auto — ajouter un TTL/cron
    si le volume devient gros). Anonymes inclus (« anonyme »).

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
