# SGRentMcServer — Plan de projet

> Dernière mise à jour : 2026-06-09

---

## Marque

- **Nom public : Playrena** (provisoire — l'utilisateur n'est pas 100% convaincu, à réévaluer)
- Slogan : **"Free Game Hosting — Hébergement de jeux gratuit"**
- `SG` / `SGRentMc` = nom **interne** uniquement (repo, containers `sgrent-*`, labels Docker, code). Ne jamais l'exposer côté public.
- Noms écartés car déjà pris : HostArena (hostarena.org), Hostera (hostera.com), Hostari/Hostero (game hosts existants). Le préfixe "Host…" est saturé.
- Vérifier la dispo d'un nom avant de l'adopter (recherche Google + domaine).

## Vision

Panel web d'hébergement de serveurs de jeux, démarrant avec Minecraft, extensible à d'autres jeux (Satisfactory, Rust, ARK). Modèle freemium avec paiement PayPal, Visa/Mastercard et crypto (Bitcoin, Ethereum).

---

## Modèle d'affaires

### Plans Minecraft

| Tier | RAM | Cores | Slots | Mods | Prix/mois |
|------|-----|-------|-------|------|-----------|
| **Gratuit** | 1 GB | 1 | 5 joueurs | Liste curatée seulement | 0$ |
| **Starter** | 2 GB | 1 | 20 joueurs | Upload custom | ~3$ |
| **Standard** | 4 GB | 2 | 50 joueurs | Upload custom | ~7$ |
| **Pro** | 8 GB | 4 | 100 joueurs | Upload custom | ~14$ |
| **Extreme** | 16 GB | 6 | illimité | Upload custom | ~25$ |

- Plan gratuit : 1 serveur par compte, disponible dès la pre-release pour fidéliser les early users
- Plans payants : ressources dédiées garanties dans le container

### Méthodes de paiement
- **Carte (Visa/Mastercard)** — Stripe
- **PayPal** — Stripe ou PayPal SDK direct
- **Crypto (BTC + ETH)** — BTCPay Server (self-hosted, 0% frais tiers)

---

## Infrastructure physique

### Nodes actuels
| Node | RAM | Cores | Rôle |
|------|-----|-------|------|
| Node 1 (Dell Xeon) | 32 GB | 24 | Serveurs joueurs + services |
| Node 2 (Dell Xeon) | 32 GB | 24 | Serveurs joueurs (overflow) |

### Capacité estimée par node
- Marge OS/services : ~4 GB réservés
- RAM utilisable : ~28 GB
- Serveurs gratuits 1 GB : ~28 par node → **~56 total**
- Mix payant typique : ~10-15 serveurs de tailles variées par node

### Réseau & Domaines
- Domaine principal : `mcserver.vbt-prog.com` (panel web)
- API : `api.vbt-prog.com`
- Admin : `admin.vbt-prog.com`
- Serveurs joueurs : `[username].servers.vbt-prog.com`
  - Fallback si conflit : `[username]-[id].servers.vbt-prog.com`
- Cloudflare devant (Full Strict), certificat Origin

---

## Architecture technique

```
Clients
   │
Cloudflare (Full Strict SSL)
   │
Traefik (reverse proxy global — remplace nginx portfolio comme entrypoint)
   ├── vbt-prog.com            → SGPortfolio (container existant)
   ├── mcserver.vbt-prog.com   → Frontend Next.js
   ├── api.vbt-prog.com        → Backend Go
   ├── admin.vbt-prog.com      → Panel admin (Go + HTMX ou Next.js)
   ├── pay.vbt-prog.com        → BTCPay Server
   └── *.servers.vbt-prog.com  → Routing vers containers MC (via Traefik labels)

Backend Go (API principale)
   ├── Auth (Discord OAuth2 + JWT / Auth custom futur)
   ├── Billing (Stripe + BTCPay webhooks)
   ├── Orchestrator (Docker remote API multi-node)
   │    ├── Node 1 — Docker socket (SSH tunnel sécurisé)
   │    └── Node 2 — Docker socket (SSH tunnel sécurisé)
   ├── Load Balancer (sélection du node avec le plus de RAM libre)
   └── Monitor (métriques CPU/RAM/réseau par container → Redis)

PostgreSQL
   ├── users, sessions, oauth_tokens
   ├── servers (config, état, node assigné)
   ├── billing (abonnements, transactions, invoices)
   └── audit_logs

Redis
   ├── métriques temps réel (TTL court)
   ├── sessions JWT
   └── queues tâches (start/stop/restart serveurs)

Containers Game Servers (1 par serveur)
   └── Image modulaire par jeu (tag Docker)
       ├── minecraft:paper      ← actif
       ├── minecraft:pumpkin    ← futur (impl. Rust custom)
       ├── satisfactory:latest  ← placeholder
       ├── rust-game:latest     ← placeholder
       └── ark:latest           ← placeholder
```

---

## Stack technologique

| Composant | Technologie | Raison |
|-----------|-------------|--------|
| Backend | **Go** | Performant, concurrence native, idéal pour orchestrer Docker et streamer logs/métriques |
| Frontend | **Next.js (TypeScript)** | SSR/SSG pour SEO, App Router, bon ecosystem |
| Base de données | **PostgreSQL** | Robuste, relationnel, parfait pour billing |
| Cache / Queues | **Redis** | Métriques temps réel, sessions, jobs async |
| Reverse proxy | **Traefik v3** | Labels Docker natifs, wildcard TLS, multi-node |
| Auth | **Discord OAuth2** (phase 1) → Auth custom email+password (phase 2) | Discord naturel pour audience gaming |
| Paiement carte | **Stripe** | API solide, PayPal via Stripe ou SDK séparé |
| Paiement crypto | **BTCPay Server** | Self-hosted, 0% frais, BTC + ETH |
| Containers | **Docker** (remote API via SSH) | 1 container = 1 serveur de jeu |
| Monitoring admin | **Métriques custom** via Docker stats API | RAM/CPU/réseau par container en temps réel |

---

## Modules du projet

```
SGRentMcServer/
├── frontend/               # Next.js — panel utilisateur + vitrine
│   ├── app/
│   │   ├── (marketing)/    # Page d'accueil, plans, pricing
│   │   ├── dashboard/      # Panel utilisateur (ses serveurs)
│   │   │   ├── servers/    # Liste + détail serveur
│   │   │   ├── mods/       # Gestion mods du serveur
│   │   │   ├── files/      # File manager (worlds, configs)
│   │   │   ├── console/    # Console temps réel (WebSocket)
│   │   │   ├── players/    # Whitelist, banlist, ops
│   │   │   └── billing/    # Abonnement, factures
│   │   └── auth/           # Discord OAuth callback
│   └── ...
│
├── backend/                # Go — API REST + WebSocket
│   ├── cmd/api/
│   ├── internal/
│   │   ├── auth/           # Discord OAuth2, JWT
│   │   ├── billing/        # Stripe, BTCPay
│   │   ├── servers/        # CRUD serveurs, lifecycle
│   │   ├── orchestrator/   # Docker multi-node
│   │   ├── monitor/        # Métriques containers
│   │   └── loadbalancer/   # Sélection node optimal
│   └── ...
│
├── admin/                  # Panel admin (Go + HTMX ou Next.js séparé)
│   ├── nodes/              # Vue des nodes physiques (RAM/CPU global)
│   ├── servers/            # Tous les serveurs, filtres, actions
│   ├── users/              # Gestion comptes
│   └── billing/            # Revenus, transactions
│
├── docker/
│   ├── images/
│   │   ├── minecraft/      # Dockerfile Paper + Fabric/Forge support
│   │   ├── satisfactory/   # placeholder
│   │   ├── rust-game/      # placeholder
│   │   └── ark/            # placeholder
│   ├── traefik/            # traefik.yml + dynamic configs
│   └── btcpay/             # BTCPay Server config
│
└── infra/
    ├── scripts/            # Deploy, node setup, backup
    └── docker-compose.yml  # Stack principale (services, pas les MC)
```

---

## Panel utilisateur — fonctionnalités détaillées

### Dashboard serveur
- **Console** : terminal temps réel via WebSocket (commandes RCON)
- **Start / Stop / Restart** : contrôle du container
- **Métriques** : RAM utilisée, CPU, joueurs connectés, uptime
- **Logs** : logs serveur en temps réel + historique

### Gestion mods (plans payants + liste curatée gratuit)
- Upload de fichiers `.jar` (Fabric/Forge mods)
- Sélection depuis une liste curatée (plans gratuits)
- Gestion des modpacks (import CurseForge/Modrinth futur)
- Version Java / serveur configurable (1.20.x, 1.21.x, etc.)

### Gestion fichiers
- File browser : `world/`, `plugins/`, `config/`, etc.
- Upload/download de fichiers
- Import/export de mondes (`.zip`)

### Règles & joueurs
- Whitelist on/off + gestion liste
- Banlist (IP + UUID)
- Ops / permissions
- Règles serveur (`server.properties` simplifié via UI)

### Billing
- Voir plan actuel, date renouvellement
- Upgrade/downgrade
- Historique transactions
- Annuler abonnement

---

## Panel admin — fonctionnalités détaillées

### Vue nodes physiques
- RAM totale / utilisée / libre par node
- CPU total / utilisé par node
- Nombre de containers actifs / en pause / arrêtés
- Alertes si node > 85% RAM

### Vue serveurs
- Liste tous les serveurs (actifs, inactifs, suspendus)
- Filtres : par node, par plan, par jeu, par utilisateur
- Actions : force stop, migrate (déplacer vers autre node), supprimer
- Voir consommation individuelle en temps réel

### Load balancing
- Algorithme : assignation au node avec le plus de RAM libre au moment de la création
- Vue graphique de la répartition
- Seuil configurable pour refuser nouvelles créations si node surchargé

### Gestion utilisateurs
- Liste comptes, date inscription, plan, serveurs
- Suspendre / bannir compte
- Voir historique billing

---

## Serveur de jeu — Minecraft

### Phase 1 : Paper (Java)
- Image Docker basée sur `openjdk:21-slim`
- Paper comme serveur de base (performances, API plugins)
- Support Fabric et Forge via sélection à la création
- Flags JVM optimisés (Aikar's flags)
- RCON activé pour les commandes console depuis le panel

### Phase 2 : Pumpkin (Rust custom)
- Implémentation Rust du protocole Java Edition (unobfuscated Mojang mappings)
- Objectif : meilleures performances, moins de RAM par instance
- Même interface Docker que Paper (drop-in replacement)
- À développer en parallèle dans un repo séparé

### Isolation & sécurité
- Chaque container : réseau isolé, volume dédié
- Pas d'accès inter-containers
- Limites Docker : `--memory`, `--cpus` selon le plan
- Ports MC : assignés dynamiquement depuis un pool (25565–26565)

---

## Jeux supplémentaires — Placeholders

Ces jeux sont prévus dans l'architecture mais **pas en construction** pour l'instant.

| Jeu | Notes |
|-----|-------|
| Satisfactory | Dédicated server Linux disponible |
| Rust | Dedicated server Linux disponible |
| ARK: Survival | RAM-heavy (~8GB min), plans premium seulement |

L'architecture modulaire (image Docker par jeu, même orchestrateur) permet d'ajouter un jeu en ajoutant une image et un template de config.

---

## Phases de développement

### Phase 1 — Fondations (priorité)
- [ ] Migration Traefik (remplacer nginx comme entrypoint global)
- [x] Backend Go : structure API, routing, middleware JWT ✓
- [x] Auth Discord OAuth2 — code complet ✓
- [x] ✅ **Auth Discord — CONFIGURÉE & DÉPLOYÉE** (2026-06-09)
  - App Discord CLIENT_ID `1513764241784705034`, redirect `https://mcserver.vbt-prog.com/auth/discord/callback`
  - Secrets dans `.env` (gitignored). Migration appliquée automatiquement par compose.
- [x] Base de données : schéma PostgreSQL initial — `internal/db/migrations/001_init.sql` ✓ (tables `users` + `game_servers`, pas encore appliquée en prod)
- [x] Orchestrateur Docker multi-node ✓
  - `internal/orchestrator/node.go` — connexion Docker daemon, stats RAM/CPU par node
  - `internal/orchestrator/container.go` — create/start/stop/restart/remove container avec limites RAM/CPU
  - `internal/orchestrator/orchestrator.go` — gestion N nodes, `BestNode()` = load balancing par RAM libre
  - Config multi-node via `.env` : `DOCKER_NODES=node1=unix:///var/run/docker.sock,node2=tcp://IP:2376`
- [x] Handlers serveurs : create/list/get/delete/start/stop/restart ✓
  - `internal/servers/repo.go` — CRUD game_servers en DB
  - `internal/servers/plans.go` — définition des plans (free/starter/standard/pro/extreme)
  - `internal/servers/ports.go` — allocation de port dynamique (25565–26565)
  - `internal/server/handlers_servers.go` — flow complet : plan → BestNode → port → DB → container Docker → start
- [ ] ⚠️ Collision de sous-domaines : si deux users ont le même username Discord, la contrainte UNIQUE sur `subdomain` va crasher. Prévoir fallback `username-shortid` dans `sanitizeSubdomain()` (`internal/server/handlers_servers.go`)
- [x] Frontend Next.js — structure de base ✓
  - `frontend/app/page.tsx` — vitrine : hero, features, grille de plans, CTA Discord
  - `frontend/app/dashboard/page.tsx` — liste des serveurs, création, start/stop/restart inline
  - `frontend/app/dashboard/servers/[id]/page.tsx` — détail serveur, actions, sections placeholder (console, mods, fichiers, joueurs)
  - Variable d'env : `NEXT_PUBLIC_API_URL` (copier `.env.local.example` → `.env.local`)

### Phase 2 — Panel utilisateur
- [x] Création de serveur asynchrone (réponse immédiate `creating`, provisioning en arrière-plan, polling statut) ✓
- [x] Bouton **Upgrade / changement de plan** ✓ — `POST /api/v1/servers/{id}/upgrade`, applique les ressources à chaud via `ContainerUpdate` (facturation = Phase 3)
- [x] **Choix & changement de version Minecraft** ✓ — colonne `version` en DB, env `VERSION` (image itzg), sélecteur à la création + section "Version Minecraft" sur la page détail. `POST /api/v1/servers/{id}/version` recrée le container (volume/monde conservé). Versions : LATEST, 1.21.x, 1.20.x, 1.19.4, 1.18.2, 1.16.5, 1.12.2, 1.8.9
- [x] Contrôles serveur (start/stop/restart/delete) sur la page détail ✓
- [x] **Sous-domaines uniques par serveur** ✓ (username, username-<nom>, username-N) — débloque plusieurs serveurs par user. Testé : 4 serveurs simultanés OK.
- [x] **Badges de statut clairs** (En ligne / Initialisation… / Arrêté / Erreur) sur dashboard + page détail ✓
- [x] **Upgrade = abonnement** ✓ — plans payants affichent "abonnement requis", bouton "S'abonner — bientôt" (gate billing Phase 3). Plan gratuit applicable directement.
- [x] Fix bug JSON tags (`GameServer` renvoyait du PascalCase au lieu de snake_case) ✓
- [ ] Dashboard serveur (console, file manager, mods — menu complet)
- [x] **Connexion par username (`username.servers.vbt-prog.com`)** — mc-router DÉPLOYÉ & vérifié ✓ (2026-06-09) :
  - ✓ Orchestrateur ajoute labels `mc-router.host` + connecte les serveurs au réseau `mc-net`
  - ✓ Port 25565 réservé à mc-router ; serveurs sur 25566+ (accès direct IP:port aussi possible)
  - ✓ Container `mc-router` (itzg/mc-router) déployé : `-p 25565:25565`, socket Docker, réseau mc-net, `IN_DOCKER=true`, API sur :26666. Auto-découverte des serveurs via labels confirmée (route `avabata.servers.vbt-prog.com → container:25565`).
  - Redéploiement : `docker run -d --name mc-router --restart unless-stopped -p 25565:25565 -v /var/run/docker.sock:/var/run/docker.sock --network mc-net -e IN_DOCKER=true -e API_BINDING=:26666 itzg/mc-router`
  - [ ] ⚠️ ACTION UTILISATEUR : DNS Cloudflare `*.servers.vbt-prog.com` en **DNS-only (nuage gris)** → IP publique
  - [ ] ⚠️ ACTION UTILISATEUR : ouvrir/forwarder le port TCP **25565** vers xe80dell (Cloudflare ne proxy pas le TCP de jeu)
  - Note : le vieux serveur "Test" a été supprimé/recréé en "Mon Serveur" (avabata) pour libérer le port 25565.
- [ ] Gestion mods et fichiers
- [ ] Gestion règles et joueurs
- [ ] WebSocket console temps réel

### Phase 3 — Billing
- [x] **PayPal — ACTIF & testé bout-en-bout** ✓ (2026-06-09) — paiement sandbox complété, plan appliqué automatiquement (standard 7$ → 4GB/2cpu)
  - 🐛 Fix : `UpdateResources` doit set `MemorySwap = Memory` sinon Docker refuse silencieusement une AUGMENTATION de RAM (swap < memory). Corrigé dans create + update.
  - ⚠️ Connu : label `sgrent.ram_mb` n'est pas mis à jour après upgrade (immuable sans recréer) → le monitor admin sous-estime la RAM utilisée. À corriger (recréer au lieu d'updater, ou stocker l'alloc ailleurs).
  - ⚠️ Robustesse : le handler capture ignore l'erreur de `UpdateResources` (à logger/vérifier).
  - `internal/billing/paypal.go` (Orders API v2 : create + capture), `internal/billing/repo.go` (table `payments`)
  - Plans avec prix (`PriceCents`) : starter 3$, standard 7$, pro 14$, extreme 25$/mo
  - Routes : `POST /api/v1/servers/{id}/checkout/paypal` (crée commande), `.../capture` (capture → applique le plan), `GET /api/v1/billing/config` (client-id public pour le SDK)
  - Frontend : bouton PayPal (SDK JS) sur la section upgrade ; au paiement réussi → upgrade auto
  - Mode `PAYPAL_ENV=sandbox` par défaut (basculer `live` en prod). Creds dans `.env` (gitignored).
  - Migration `002_payments.sql` appliquée.
  - [ ] ⚠️ ACTION UTILISATEUR : fournir `PAYPAL_CLIENT_ID` + `PAYPAL_SECRET` (app Sandbox sur developer.paypal.com)
  - [ ] Webhooks PayPal (vérif paiement côté serveur, robustesse) + abonnements récurrents (actuellement paiement unique = 30j implicite)
- [ ] Stripe (Visa/Mastercard)
- [ ] BTCPay Server (BTC + ETH)
- [ ] Gestion abonnements, upgrades, factures, reçus

### Phase 4 — Panel admin & multi-node
- [x] **Panel admin monitoring — EN LIGNE** ✓ (2026-06-09) → `https://mcserver.vbt-prog.com/admin/` (Basic Auth, user `admin`)
  - Service Go autonome `cmd/monitor` (pas de dépendance DB), image `mcserver-monitor`, container `mcserver-monitor` sur `sgportfolio_portfolio-net`
  - `internal/monitor/collector.go` — métriques par container (CPU %, RAM usage/limite, disque rw, réseau, uptime) + hôte (RAM, CPU/load, disque via statfs, nb containers)
  - Lit le socket Docker (`/var/run/docker.sock:ro`) + FS hôte (`/:/host:ro` pour disque & hostname)
  - UI dark auto-rafraîchie (2s) embarquée (`cmd/monitor/static/index.html`, Tailwind CDN)
  - Protégé par nginx Basic Auth : htpasswd dans `SGPortfolio/docker/nginx/certs/.htpasswd` (gitignored). ⚠️ dossier certs passé en 0755 (traversable par worker nginx uid 101) ; `origin.key` reste 0600.
  - Build/redeploy : `docker build -f Dockerfile.monitor -t mcserver-monitor:latest . && docker rm -f mcserver-monitor && docker run -d --name mcserver-monitor --restart unless-stopped --network sgportfolio_portfolio-net -v /var/run/docker.sock:/var/run/docker.sock:ro -v /:/host:ro mcserver-monitor:latest`
- [x] **Actions admin sur les serveurs** ✓ (2026-06-09)
  - Boutons "Forcer l'arrêt" + "Supprimer" dans l'UI du monitor (avec confirmation)
  - Endpoints monitor : `POST /admin/api/containers/{id}/stop|delete` (agissent direct sur Docker, derrière Basic Auth)
  - ⚠️ Limite v1 : actions directes Docker — pour un serveur SGRent géré, la ligne DB `game_servers` n'est pas synchronisée (orphelin possible). Réconciliation à prévoir.
  - Affichage des **cœurs CPU alloués** par container (`cpu_limit` depuis `HostConfig.NanoCPUs`)
- [ ] **Monitoring multi-node** : actuellement le monitor lit UNIQUEMENT le socket Docker local (xe80dell). Le 2e Dell n'apparaîtra PAS tant qu'on n'étend pas le collecteur pour se connecter à plusieurs daemons Docker (Docker API distante via TCP+TLS, ou un agent monitor sur chaque node qui remonte ses métriques). Réutiliser le concept `DOCKER_NODES` de l'orchestrateur.
- [x] **Monitoring trafic réseau** ✓ (2026-06-09) — débit ↓↑ MB/s temps réel + totaux GB, agrégé sur tous les containers (le `/proc/net/dev` de l'hôte n'est pas lisible depuis le container — isolation net namespace — donc on agrège le trafic des containers, ce qui est plus pertinent). Carte "Trafic réseau (serveurs)" dans /admin.
- [ ] Améliorations panel admin : historique/graphiques, taille des volumes (worlds), filtre serveurs-gérés-seulement
- [ ] Migrer le panel vers `admin.vbt-prog.com` (sous-domaine dédié) quand souhaité
- [x] **Load balancer multi-node — FONCTIONNEL** ✓ (2026-06-09)
  - Node 2 = `xe90bequiet` (10.0.0.110), Docker 27.5.1, user `simon` dans groupe docker
  - Accès via transport **ssh://** (connhelper) — pas de daemon TCP/TLS, pas de sudo. Clé dédiée `~/.ssh/orchestrator_node2` autorisée sur node2.
  - Container API : `openssh-client` + entrypoint qui copie `/etc/sgrent-ssh` → `/root/.ssh` en root (ssh exige l'ownership root du config). Secrets montés via `./secrets/ssh` (gitignored).
  - `DOCKER_NODES=node1=unix://...,node2=ssh://simon@10.0.0.110`. `BestNode` choisit le node avec le plus de RAM libre.
  - ✅ Testé : création/start/delete d'un serveur sur node2 OK. Répartition auto confirmée.
  - Projets copiés sur node2 (`~/Shared_Projects/2026/SGPortfolio` + `SGRentMcServer`). `mc-net` créé sur node2.
  - ✅ **Routing cross-host par username IMPLÉMENTÉ** : mc-router géré via API REST (plus d'auto-découverte Docker). L'orchestrateur enregistre `hostname → backend` à la création/recréation, désenregistre à la suppression, réconcilie au démarrage (`ReconcileRoutes`). Backend = nom de container sur mc-net (node local) OU `IP_LAN:port` (node distant). `internal/mcrouter/client.go`, config `MC_ROUTER_API` + `NODE_ADDRS`. mc-router sur portfolio-net + mc-net.
  - ✅ Vérifié : routing node1 (`avabata` → Paper) OK.
  - ⚠️ **BLOQUEUR node2 : FIREWALL**. node2 (xe90bequiet) a un firewall qui ne laisse passer que SSH. Les ports de jeu 25566-26565 sont bloqués → mc-router ne peut pas joindre les serveurs node2. **ACTION UTILISATEUR (sudo sur node2)** : `sudo ufw allow 25566:26565/tcp` (+ ouvrir 25565 si accès direct). Sans ça, les serveurs node2 sont créés/gérés mais injoignables.
  - ⚠️ Note ops : `mc-router` doit être connecté à `mc-net` ET `sgportfolio_portfolio-net` (`docker network connect mc-net mc-router`) — à scripter (non persistant si recréé).
- [ ] Alertes & monitoring (seuils RAM/CPU)

### Phase 5 — Pumpkin & polish
- [ ] Intégration serveur Rust (Pumpkin) comme option
- [ ] Modrinth/CurseForge modpack import
- [ ] Auth custom email+password (en plus de Discord)
- [ ] Autres jeux (Satisfactory, Rust, ARK)

### Monétisation publicitaire (OPTIONNEL — note)
> Revenu principal = plans payants. La pub est secondaire et à faire avec prudence.
- **NE PAS** mettre de pub sur l'app/dashboard/accueil (nuit aux conversions, viole souvent les règles AdSense, looks "cheap" pour un service payant).
- **AdSense uniquement sur le blog** (contenu = autorisé/pertinent). Prérequis : compte AdSense approuvé (`ca-pub-XXXX`), politique de confidentialité, contenu original + un peu de trafic (sites neufs souvent refusés au début).
- **Alternative à faible trafic : affiliation** sur le blog (matériel gaming, clés de jeux) — rapporte plus qu'AdSense à bas volume, moins intrusif.
- Défauts à prévoir quand on l'active : ID éditeur `ca-pub-XXXX` en variable d'env/config, composant `<AdSlot>` blog-only, script AdSense chargé conditionnellement (pages blog seulement), `app/confidentialite` (politique requise).
- Statut : non implémenté, à activer quand AdSense approuvé.

### Phase 6 — Marketing, SEO & acquisition
- [x] Bases SEO techniques ✓ (2026-06-09) : `app/robots.ts` (/robots.txt), `app/sitemap.ts` (/sitemap.xml), metadata enrichies (title/description mots-clés, OpenGraph, canonical, keywords FR Québec)
- [ ] ⚠️ ACTION UTILISATEUR : soumettre `https://mcserver.vbt-prog.com/sitemap.xml` dans Google Search Console + "Demander l'indexation" (la page n'est pas encore indexée — normal, site récent)
- [x] Page indexée sur Google ✓ (confirmé via Search Console 2026-06-09)
- [x] **Données structurées Schema.org** ✓ — Product/AggregateOffer (prix 0-25$) sur la home, Article sur les posts de blog
- [x] **Blog SEO** ✓ — `app/blog/` : 3 guides FR (créer serveur gratuit, meilleurs mods, Paper vs Vanilla), métadonnées + canonical + OG par article, ajoutés au sitemap, lien Blog dans la nav
- [ ] Plan backlinks → voir `BACKLINKS.md` (annuaires MC, Reddit, Bing Webmaster). ⚠️ ACTION UTILISATEUR.
- [ ] **SEO** : continuer (nouveaux articles 1-2/mois, suivi Search Console)
  - Métadonnées par page (title/description optimisés mots-clés : "serveur minecraft gratuit", "hébergement minecraft québec/canada", etc.)
  - Open Graph + Twitter cards (partage réseaux sociaux)
  - `sitemap.xml` + `robots.txt` (Next.js : `app/sitemap.ts` + `app/robots.ts`)
  - Données structurées Schema.org (Product/Offer pour les plans → rich snippets prix)
  - Inscription à **Google Search Console** + soumission du sitemap
  - Performance/Core Web Vitals (Next.js déjà SSR, viser bons scores Lighthouse)
  - Contenu : blog/guides ("comment créer un serveur Minecraft", tutoriels mods) pour le SEO long-tail
- [ ] **Publicité payante** : évaluer les canaux
  - Google Ads (Search) sur mots-clés d'hébergement Minecraft
  - Réseaux sociaux gaming : Reddit (r/admincraft, r/Minecraft), TikTok, YouTube
  - Partenariats influenceurs/YouTubeurs Minecraft (codes promo)
  - Serveurs Discord communautaires Minecraft
- [ ] **Acquisition organique** : listings sur annuaires de serveurs (minecraft-mp.com, etc.), programme de parrainage

---

## Environnement de développement / déploiement

- ⚠️ **CORRECTION ARCHITECTURE** : la machine de dev **est** un des nodes Dell — `xe80dell` (hostname). C'est CETTE machine qui héberge le portfolio (containers `portfolio-nginx` sur 80/443 + `portfolio` sur 3000). Le second Dell n'est pas encore configuré (TODO). Donc portfolio et game-node = même machine pour l'instant.
- Outils installés sur xe80dell : Go 1.24.4 (`~/go-sdk/go/bin`), Node 20.18 (`~/node20/bin`), Docker 28.3.3. PATH ajouté à `~/.bashrc`.
- IP publique : `24.157.140.226` (via Cloudflare). IPs internes : 10.0.0.2 / 10.0.0.245.
- Réseaux Docker préexistants intrigants à investiguer : `docker_game_net`, `docker_auth_net`, `traefik-public` (scaffolding d'une session précédente ?).
- **Traefik** : migration reportée — nginx du portfolio reste l'entrypoint. Routing par ajout de server blocks dans `SGPortfolio/docker/nginx/nginx.conf`.

### ⚠️ Piège bind-mount nginx (IMPORTANT)
`portfolio-nginx` monte `nginx.conf` comme **fichier unique** (pas un dossier). Éditer ce fichier via un outil qui réécrit l'inode **casse le mount** : le container continue de voir l'ancien fichier. Un `nginx -s reload` ne suffit PAS → il faut **`docker restart portfolio-nginx`** (coupe le portfolio ~1-2s). Vérifier avec `docker exec portfolio-nginx grep server_name /etc/nginx/nginx.conf`.

---

## Déploiement vitrine — FAIT (2026-06-09)

Vitrine SGRentMc en ligne côté serveur (mode "Bientôt disponible", login désactivé via flag `COMING_SOON` dans `frontend/app/page.tsx`).

- `frontend/Dockerfile` — build Next.js standalone (`output: "standalone"` dans `next.config.ts`)
- Image `mcserver-frontend:latest`, container `mcserver-frontend` sur réseau `sgportfolio_portfolio-net`, `--restart unless-stopped`
- Server block `mcserver.vbt-prog.com` ajouté dans `SGPortfolio/docker/nginx/nginx.conf` → proxy vers `http://mcserver-frontend:3000`
- Cert Origin couvre `*.vbt-prog.com` ✓ (vérifié via SAN)
- Vérifié : `mcserver.vbt-prog.com` → page SGRentMc ✓ / `vbt-prog.com` → portfolio intact ✓

### Statut : EN LIGNE PUBLIQUEMENT ✓ (2026-06-09)
- [x] **DNS Cloudflare** : record `mcserver` → 172.64.80.1 (proxied) ✓ — vérifié via edge Cloudflare, sert bien la page SGRentMc
- [x] Lien header portfolio pointe déjà vers `https://mcserver.vbt-prog.com` (`SGPortfolio/src/components/Header.tsx`) — garde un badge "bientôt" (à retirer au lancement)
- [x] Commit dans SGPortfolio : `b65eef9` sur branche `Claude/feature/v0.0.1-setup` (pas encore push)
- [ ] Quand backend prêt : passer `COMING_SOON = false` (`frontend/app/page.tsx`) + `docker build` + `docker rm -f mcserver-frontend && docker run ...`
- [ ] Note : badge "bientôt" du header à retirer au vrai lancement

---

## Décisions d'architecture

| Sujet | Décision |
|-------|----------|
| **BTCPay Server** | Sur le VPS principal (vbt-prog.com) avec Traefik — à ajouter en Phase 3 |
| **Backups automatiques** | Option payante uniquement (pas inclus dans le plan gratuit) |
| **DDoS protection ports MC** | À planifier en Phase 5 (ports 25565+ exposés directement, Cloudflare ne couvre pas le TCP/UDP jeux) |
| **Conformité données** | Loi 25 Québec pour le lancement — RGPD européen en phase avancée si expansion internationale |
