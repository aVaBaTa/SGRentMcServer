# Playrena — État du projet (contexte pour nouvelles conversations)

> Fichier de reprise rapide. Mentionne **« Playrena »** dans une nouvelle conversation
> (le skill `playrena` charge ce fichier + l'état git) OU dis simplement
> **« va lire PLAYRENA.md »**.
> Dernière mise à jour : 2026-06-11.

## 🔄 Reprise rapide (à exécuter au début d'une nouvelle conversation)

1. Lire ce fichier en entier (archi, état des jeux, **backlog #A→#F**, blocage UDP).
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
    `v*.*.*` n'expose plus de téléchargement. Typecheck OK. ⚠️ pas encore commité.
- 🟡 **#E — Nouveau site « SG Studio » (sites web propulsés par IA)** — MVP CODÉ (2026-06-11) :
  - **Nouveau repo** `~/Shared_Projects/2026/SGWebStudio` (git init + commit initial sur `main`).
    Next.js 15 standalone, Tailwind, marque/contenu centralisés dans `src/lib/site.ts`
    (renommable en 1 fichier). Domaine cible **`studio.vbt-prog.com`**.
  - **Landing** : hero, avantages, services, process, contact. **Formulaire de contact**
    `POST /api/contact` → **nodemailer** vers SMTP `mailserver:587` (calque le mailer Go,
    `tls.rejectUnauthorized:false`, honeypot anti-spam, validation). Env dans `.env.example`.
  - **Lien depuis le portfolio** ajouté (`SGPortfolio/src/components/Header.tsx` → « SG Studio »).
  - Build + tests de fumée OK (homepage, `/api/contact` 503 sans SMTP, 400 si email invalide).
  - **Reste (non bloquant, à déployer)** : créer le repo GitHub `aVaBaTa/SGWebStudio` + push ;
    DNS Cloudflare A `studio` → 24.157.140.226 (proxied) ; bloc nginx `studio.vbt-prog.com`
    → `sgwebstudio:3000` ; `docker build` + `docker run` (réseaux `portfolio-net` +
    `sgmail_default`, `--env-file .env` avec creds SMTP). Plus tard : visuels/réalisations,
    grille de prix, i18n.
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
- **#H — Ajouter un CI/CD** : pipeline de build/test/déploiement (GitHub Actions ?) pour les
  repos (au moins SGRentMcServer : backend Go + frontend Next.js).
- **#I — Faire en sorte que Satisfactory fonctionne** : connexion joueur réelle (dépend du
  pare-feu UDP, cf. blocage connu ci-dessous) + validation bout en bout.
- **#J — S'inspirer de la concurrence** : prendre référence sur
  https://shockbyte.com/games/minecraft-server-hosting **et un autre site** (à choisir) pour
  améliorer offre, UX, pricing, pages jeux.
- **#K — Publicité dans les différents éléments** : insérer de la pub / mises en avant dans
  les composants du site (bannières, cross-sell entre jeux, promos, etc.).
- **#L — Mods et plugins pour Minecraft** : permettre l'ajout de mods/plugins (page `/mods`
  actuellement placeholder « bientôt »).
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
  - [ ] **CRUD serveurs — CREATE/ré-attribution** : créer un serveur au nom d'un user →
        nécessite l'**orchestrateur** (vit dans l'API `internal/orchestrator`, pas le monitor).
        À câbler : soit endpoint admin sur l'API (token interne), soit le monitor importe
        l'orchestrateur. ⚠️ ne pas faire à moitié (sinon ports/volumes/mc-router orphelins).
  - [ ] **Mini-Jira (board tickets)** : colonnes à faire/en cours/fait, persisté en DB
        (nouvelle table `admin_tasks`), pour remplacer ce backlog .md à terme.
  - [ ] **Métriques business avancées** : conversions, **coûts annonces** (Google Ads),
        revenu réel (lié #F/#P), churn, etc.
- **#N — Visuels** : ajouter des **images de Minecraft** (et des autres jeux) sur le site.
- **#O — Serveur communautaire / partenaires** : monter un serveur communautaire et/ou
  **trouver des partenaires** pour héberger des serveurs.
- **#P — Vérifier les paiements PayPal** : s'assurer que le flux de paiement PayPal
  fonctionne de bout en bout.
- **#Q — Plus de méthodes de paiement et de connexion** : ajouter d'autres moyens de
  paiement et d'autres providers d'authentification (au-delà de Discord / PayPal).

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
