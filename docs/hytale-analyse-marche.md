# Hytale — Analyse de marché & stratégie (réflexion)
*Playrena · 2026-06-12 · sources web + données infra réelles. Synthèse des échanges de session.*

> **TL;DR** — Louer des serveurs Hytale **payants n'est pas rentable** (produit gratuit en
> self-host, marché saturé de partenaires officiels, plancher RAM 4 Go non-surbookable). MAIS la
> **demande existe** (gros lancement, modding fort) → garder Hytale comme **vitrine SEO + coup
> marketing**. La stratégie viable = **events sponsorisés** (25/50/100 j) sur **cloud loué à
> l'heure (~2-10 €/event)**, pas sur la machine maison. Concentrer le vrai business sur **Minecraft**.

---

## 1. Le produit Hytale est gratuit (le point qui change tout)
- Hytale **encourage le self-hosting gratuit** : `HytaleServer.jar` + `Assets.zip`, Java 25,
  ouvrir **UDP 5520**, forward routeur. Documenté par Hypixel Studios.
- Hytale **n'héberge aucun serveur** lui-même (ni gratuit ni payant) — il fournit le binaire.
- **Aucun cap joueurs** côté licence : `MAX_PLAYERS` défaut 100, cap pratique ~200. La limite
  de licence = **500 serveurs concurrents** par compte (pas des joueurs).
- ⇒ Ton client paierait pour ce qu'il peut faire **gratuitement**. Le seul vrai coût = le matériel.

## 2. Le marché est saturé (et il y a du gratuit)
- **Partenaires officiels** : Apex (« launch partner »), Shockbyte (« Launch Verified »),
  Pingperfect (« official partner »). Plus BisectHosting (21 datacenters), G-Portal, GTXGaming,
  PebbleHost (1 $/Go).
- **Offres gratuites** : ScalaCube (free tier), Gaming4Free (4 Go / 20 slots gratuits).
- ⇒ Prix plancher ≈ **0**. Personne ne donne un 50 j gratuit (≈ 50 Go RAM = argent réel) ; le
  gratuit plafonne à ~20 j / 4 Go.

## 3. L'auth : tu n'es PAS en retard
Deux modèles sur le marché :
- **Modèle B (standard, la majorité)** : le **client** authentifie avec **son** compte Hytale via
  la console (`auth login device` → code → `accounts.hytale.com/device` → autoriser). C'est
  **exactement ce que Playrena fait déjà** (watcher device-code, panel URL+code). Hyper Layer le
  qualifie de « temporary manual process » → **tout le monde subit la même jank**.
- **Modèle A (premium)** : le **provider** pré-authentifie avec **sa** licence → injecte
  `HYTALE_SERVER_SESSION_TOKEN` / `IDENTITY_TOKEN` (0-click, jusqu'à 500 serveurs, entitlement
  `unlimited_servers` sur demande). Plus de travail, marge nulle → pas justifié pour Playrena.
- ⇒ L'auth n'est **pas** le facteur décisif : on est au niveau du marché. Le problème est l'infra + l'éco.

## 4. L'économie de l'hébergement (pourquoi 119 €/mois peut être rentable… pour Minecraft)
Les hôtes sont rentables par **densité + surbooking**, pas au serveur :
- Minecraft se vend **1-3 $/Go/mois**. Un dédié 128 Go (~119 €/mo) découpé en **30-60 locataires**
  → **250-400 $/mo** de revenu. Marge.
- **Surbooking** : vendre plus de Go que la RAM physique (swap/paging), en pariant sur des pics
  non simultanés. Les budget hosts entassent 20-50 instances là où 5 tiennent proprement.
- **Amortissement** : serveur acheté ~1500 $ / 4-5 ans ≈ **30 $/mo** → moins cher que louer.
- **Coût marginal ≈ 0** : panel self-service, volume de petits clients, upsells (backups, slots, IP).

**Le levier = nombre de locataires par machine.** Et c'est là que Hytale meurt :

| Jeu | RAM/serveur | Locataires sur 128 Go | Surbookable ? | Business |
|---|---|---|---|---|
| **Minecraft** | 1-2 Go starter | **60-100** | oui | ✅ excellent |
| **Hytale** | **4 Go plancher + ~1 Go/joueur** | **~8-15 max** | non (plancher dur) | ❌ mauvais |

## 5. Les coûts RAM Hytale (pourquoi gros = cher)
Formule officielle : **4 Go (4 j) + ~1 Go par joueur supplémentaire**.
| Joueurs | RAM requise | Réalité |
|---|---|---|
| 25  | ~25 Go  | les hôtes vendent ça sur 8 Go en survendant (view distance limitée, joueurs groupés) |
| 50  | ~50 Go  | « 20-50 j » vendu 25-50 $/mo sur 8-16 Go (oversell) |
| 100 | ~100 Go | machine quasi dédiée |
View distance (cap 12 chunks) = principal driver ; exploration dispersée = bien plus de RAM.

## 6. La contrainte matérielle dure (xe80dell)
- **31 Go RAM total (~21 Go libres), 48 cœurs, 8 Go swap.** Possédé → coût marginal ≈ électricité.
- ⇒ **Impossible** d'héberger 25 j (~25 Go), encore moins 50/100 j Hytale sur la machine actuelle.
- Minecraft sur xe80dell : ~7×4 Go ou ~15×2 Go = **45-90 $/mo** sur du matériel déjà payé. **C'est ça le business.**

## 7. Signaux de marché Hytale (mitigés)
**✅ Positifs (demande réelle)**
- Lancement 13 jan 2026 énorme : **Twitch 420 k+ peak**, 25 000+ créateurs, files d'attente de plusieurs heures.
- **Modding fort** : 1 M de téléchargements de mods en 48 h, 500+ mods → serveurs gourmands = plus de valeur.
- **Hypixel financé ~2 ans**, plan support 10 ans.
- Culture des **events 100 joueurs** (énorme sur MC) qui se transfère.

**⚠️ Négatifs (l'éco ne suit pas)**
- **MAU déjà en baisse : ~740 k, -13,6 % m/m.** Hype qui refroidit (le « 2,8 M concurrent » était un bug/fake).
- Prix plancher ≈ 0 ; marché saturé de partenaires officiels ; latence mono-région défavorable.

## 8. La stratégie viable : events sponsorisés sur cloud à l'heure
Le coût serveur d'un event est **dérisoire** (cloud Hetzner, à l'heure, plafonné au mois) :

| Joueurs | Type | RAM | €/h | Event 5 h | Week-end |
|---|---|---|---|---|---|
| 25  | ccx33 | 32 Go  | 0,10 € | ~0,50 € | ~5 €  |
| 50  | ccx43 | 64 Go  | 0,20 € | ~1 €    | ~10 € |
| 100 | ccx53 | 128 Go | 0,40 € | ~2 €    | ~20 € |

- **Cloud = règle aussi le NAT** (IP publique, pas de forward) **et l'anti-DDoS** (inclus Hetzner/OVH).
- Dédié mensuel **déconseillé** (MAU en baisse → serveurs vides) : AX102 128 Go ~119 €/mo,
  auction 50-90 €/mo, OVH Scale dès ~513 $/mo. Latence : **OVH Beauharnois (Montréal)** pour le QC.
- **Le vrai livrable = automatiser** le spin-up/tear-down (cf. `scripts/event-server/`).
- **Démarchage** : viser des créateurs **mid-tier (50k-1M)** qui font *vraiment* des events,
  offrir un serveur d'event gratuit contre mention/lien (cf. `PARTENAIRES.md` §1.bis).
  **Ne rien promettre avant** d'avoir validé le POC sur un vrai token cloud.

## 9. Anti-DDoS (risque à mitiger — #T)
- Aujourd'hui le trafic de jeu tape l'**IP résidentielle** (CF ne proxifie pas le jeu) → une
  attaque peut **couper Internet de la maison**.
- **Minecraft (TCP)** : **TCPShield** (gratuit) masque l'IP + absorbe → à faire en 1er.
- **UDP (Satisfactory/Hytale)** : pas de proxy gratuit → **anti-DDoS hébergeur** sur cloud
  (OVH « Game » inclus, Hetzner de base). Option payante : **Cloudflare Spectrum**.
- Court terme : **ne jamais exposer l'IP résidentielle** pour des events publics → passer par un VPS relais.

## 10. Recommandations (priorisées)
1. **Minecraft = priorité business** (densité, marge, demande mûre). Maximiser les locataires sur xe80dell.
2. **Hytale = vitrine SEO + marketing**, pas une ligne payante. Garder petit (4-8 Go) sur la maison.
3. **Events sponsorisés** = la seule façon « rentable » de toucher l'audience Hytale → cloud à
   l'heure, automatisé, créateurs mid-tier. Fenêtre courte (MAU baisse) → bouger vite.
4. **Anti-DDoS** (#T) avant toute exposition publique sérieuse (TCPShield d'abord).
5. **Mining distribué (#R)** : ne pas coder ; cloud à l'heure + nodes partenaires de confiance suffisent.

## Liens internes
- Backlog & décisions : `PLAYRENA.md` (#G Hytale validé, #R mining, #S events, #T anti-DDoS).
- Cibles de démarchage : `PARTENAIRES.md` §1.bis.
- POC provisioning : `scripts/event-server/`.
- Faisabilité mining : `docs/mining-feasibility.md`.

## Sources (web, juin 2026)
- Hytale Server Manual / Server Provider Authentication Guide (support.hytale.com)
- Self-host guide (vpntierlists.com) ; providers & pricing (hostinger.com) ; free tier (gaml4free.net)
- Player count fév. 2026 ~740k −13,6 % (hytaletop100.com) ; launch ~3M debunk (thegamer.com) ; EA >1M (notebookcheck.net)
- RAM : nodecraft.com, g-portal wiki, ghostcap calculator ; tier populaire 8 Go (winternode.com)
- Cloud pricing : Hetzner Cloud CCX (costgoat.com) ; AX102 (hetzner.com) ; OVH bare metal (ovhcloud.com)
- Auth modèles : xgamingserver.com (pre-auth), docs.hyperlayer.net (client), eastgate.host, pingplayers.com
- Économie hosting / oversell : gameteam.io, akliz.net, biomehosting.com
