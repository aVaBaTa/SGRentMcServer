# Jeux populaires à ajouter au catalogue Playrena

> Backlog catalogue — candidats à ajouter au registre de jeux.
> Chaque jeu = une `GameDef` dans `internal/servers/games.go` (+ `frontend/lib/games.ts`,
> i18n/PLANS, page `/games/<jeu>`). Cf. `PLAYRENA.md` § « Registre de jeux ».
> Dernière mise à jour : 2026-06-12.

## ⚠️ Contraintes à intégrer AVANT d'ajouter un jeu

1. **RAM de xe80dell** : **31 Go total (~21 Go libres)**, 48 cœurs. Un jeu « lourd »
   (Palworld, ARK) = **~1 serveur à la fois** (comme Hytale). Les jeux légers (Terraria,
   Stardew) sont surbookables → vrais funnels gratuits/rentables.
2. **Plage de ports — allocation par BLOC** (refonte 2026-06-12) : `ports.go` attribue à chaque
   serveur un **bloc contigu** de `portBlockSize` (=16) ports `[base, base+16)` dans `25566–26565`.
   Un `GameDef.Ports(base)` doit dériver **TOUS** ses ports depuis `base` avec un **offset < 16**
   (ils restent dans le bloc du serveur → aucune collision possible, même avec des ports dérivés
   consécutifs `base+1`, `base+2`…). **Jamais de port fixe** type `27015`/`8766` en dur : on
   **configure le jeu** pour écouter sur `base+offset` via son `Env` (`-port`, `?QueryPort=`,
   `PublicPort`, `serverSteamPort`…) — à vérifier que l'image expose bien ce paramètre. Cas rare
   d'un port **non configurable** : soit remapping Docker (`HostPort=base+offset → Internal=fixe`,
   si le jeu n'annonce pas ce port), soit plafonner le jeu à 1 instance (réservation globale).
   Cœur d'allocation : `NextAvailableBase` (+ `firstFreeBlockBase` pur, testé dans `ports_test.go`).
3. **Pare-feu/NAT** : plage `25566:26565` TCP **et** UDP (cf. blocage UDP connu). Tout nouveau
   jeu UDP hérite du même prérequis routeur.
4. **mc-router** : réservé à Minecraft (routing par hostname). Tous les autres = **IP:port
   direct** (`UsesMCRouter:false`, `NEXT_PUBLIC_SERVER_HOST` affiché).
5. **Serveurs Windows-only via Wine/Proton** (The Forest, ARK: Survival Ascended) : images plus
   lourdes, démarrage plus lent, plus fragiles. Effort d'intégration **élevé**.
6. **`NeedsAuth`** : aucun de ces jeux n'utilise l'OAuth interactif d'Hytale ; la plupart se
   protègent par **mot de passe serveur** (champ env), pas par device-code.

## Tableau récap (priorité produit)

| Prio | Jeu | Effort | RAM plancher | Proto / ports (depuis `base`) | Image candidate | Capacité xe80dell |
|---|---|---|---|---|---|---|
| ⭐1 | **Valheim** | Faible | 4 Go | UDP `base`, `base+1` (+`base+2` crossplay) | `lloesche/valheim-server` | Plusieurs |
| ⭐2 | **Terraria** | Faible | 1–2 Go | TCP `base` (+UDP) | `ryshe/terraria` / `beardedio/terraria` (TShock) | Beaucoup (léger) |
| 3 | **7 Days to Die** | Moyen | 8 Go | TCP `base` + UDP `base..base+3` | `vinanrra/7dtd-server` | 1–2 |
| 4 | **Palworld** | Moyen | 8–16 Go | UDP `base` (game) + `base+1` query | `thijsvanloef/palworld-server-docker` | ~1 |
| 5 | **ARK: Survival Evolved** | Moyen-élevé | 8–16 Go | UDP `base`, `base+1`, `base+2` (query) | `hermsi/ark-server` | ~1 |
| 6 | **The Forest** | Élevé (Wine) | 4 Go | UDP `base`, `base+1`, `base+2` (Steam fixes ⚠️) | `jammsen/docker-the-forest-dedicated-server` | 1–2 |
| 7 | **ARK: Survival Ascended** | Élevé (Wine/UE5) | 16 Go+ | UDP `base` (RCON opt.) | `acekorneya/asa_server` / `mschnitzer/...` | ~0–1 ⚠️ cloud event |
| 8 | **Stardew Valley** | Élevé (modèle réseau) | 1–2 Go | ⚠️ réseau Steam/Galaxy P2P (pas IP:port) | `stardew-valley-dedicated-server/server` | Beaucoup (mais niche) |

---

## Détails par jeu

### ⭐1. Valheim — *meilleur prochain ajout*
- **Image** : `lloesche/valheim-server` (la plus mature : auto-update, backups, BepInEx/mods).
- **Ports** : `2456-2457/udp` (jeu) + `2458/udp` si crossplay → 2–3 ports UDP **consécutifs**
  → mappe proprement sur `base`, `base+1`, `base+2`.
- **RAM** : 4 Go min jouable, 8 Go confortable → `MinRAMMb: 4096`.
- **Volume** : `/config` (+ `/opt/valheim`). **Auth** : mot de passe serveur (env `SERVER_PASS`,
  min 5 car.), pas d'OAuth.
- **Pourquoi #1** : très populaire et durable, image excellente, empreinte raisonnable
  (surbookable modérément), s'intègre directement à la logique `PortMapping` UDP existante
  (calqué sur Satisfactory). Effort faible.

### ⭐2. Terraria — *funnel gratuit léger*
- **Image** : `ryshe/terraria` (vanilla) ou `beardedio/terraria` (TShock = permissions,
  anti-cheat, plugins, console d'admin).
- **Ports** : `7777` TCP (et souvent UDP) → un seul port = `base`.
- **RAM** : 1–1,5 Go pour ≤8 joueurs → `MinRAMMb: 1024` (2048 si moddé). **Très léger**.
- **Volume** : dossier `Worlds`. **Auth** : mot de passe optionnel.
- **Atout** : TCP (proxyfiable type TCPShield, cf. #T) + RAM minime = **rentable même gratuit**,
  beaucoup d'instances par machine. Idéal pour grossir la base d'utilisateurs.

### 3. 7 Days to Die
- **Image** : `vinanrra/7dtd-server` (populaire, mods/allocs) ou `didstopia/7dtd-server`.
- **Ports** : `26900/tcp` + `26900-26903/udp` (jeu) + `8080/tcp` (webpanel) + `8081` (telnet)
  → `base` (tcp) et `base..base+3` (udp) ; webpanel/telnet sur ports dérivés internes.
- **RAM** : 6–8 Go → `MinRAMMb: 8192`. **Volume** : `/home/sdtdserver`. **Auth** : mot de passe.
- **Note** : utilise déjà une **plage de ports dérivés** → cadre bien avec `base..base+3`.

### 4. Palworld
- **Image** : `thijsvanloef/palworld-server-docker` (la plus complète : RCON, backups, multi-env).
- **Ports** : `8211/udp` (jeu, requis) + `27015/udp` (query, optionnel, **remappable** via
  `PublicPort`) + `25575/tcp` (RCON, optionnel) → `base` (udp) + `base+1` (query) + RCON interne.
- **RAM** : 8 Go min, **16 Go recommandé** → `MinRAMMb: 8192` (idéalement 12288). **Lourd** →
  ~1 serveur sur xe80dell. **Volume** : `/palworld`. **Auth** : mot de passe serveur.
- **Note** : forte demande (hype 2024-25), mais gourmand → cadrer comme Hytale (1 à la fois)
  ou réserver au cloud à l'heure pour les gros (cf. #S).

### 5. ARK: Survival Evolved
- **Image** : `hermsi/ark-server` ou `didstopia/ark-server` (Linux natif, SteamCMD).
- **Ports** : `7777/udp` (jeu) + `7778/udp` (raw) + `27015/udp` (query) → `base`, `base+1`,
  `base+2`. **RAM** : 6–8 Go vanilla, **16 Go+ moddé** → `MinRAMMb: 8192`. **Lourd** → ~1 serveur.
- **Volume** : données ARK. **Auth** : mot de passe. Déjà marqué « Bientôt » dans `games.ts`.

### 6. The Forest — *Wine, effort élevé*
- **Image** : `jammsen/docker-the-forest-dedicated-server` (serveur **Windows-only via Wine**).
- **Ports** : `8766/udp` (Steam), `27015/udp` (jeu), `27016/udp` (query) → `base`, `base+1`,
  `base+2`. ⚠️ **Ports Steam parfois imposés en dur** par l'image → vérifier la remappabilité
  avant multi-instances. **RAM** : ~3–4 Go → `MinRAMMb: 4096`. **Auth** : mot de passe.
- **Note** : Wine = démarrage plus lent et plus fragile. À faire **après** les natifs Linux.
  (`jammsen` maintient aussi **Sons of the Forest** — candidat jumeau si demande.)

### 7. ARK: Survival Ascended — *très lourd, plutôt cloud*
- **Image** : `acekorneya/asa_server`, `mschnitzer/ark-survival-ascended-linux-container-image`
  ou `sknnr/ark-ascended-server` (**binaires Windows via Wine/Proton GE**).
- **Ports** : seul `SERVER_PORT` (`7777/udp`) doit être exposé (RCON optionnel, **pas de query
  port** en ASA) → `base`. **RAM** : **16 Go+** (UE5, très gourmand) → `MinRAMMb: 16384`.
- **Verdict** : dépasse la capacité confortable de xe80dell (≈1 et au prix de tout le reste) →
  **réserver au cloud à l'heure** (cf. #S, Hetzner CCX). Effort + coût RAM élevés.

### 8. Stardew Valley — *modèle réseau incompatible, niche*
- **Image** : `stardew-valley-dedicated-server/server` (ou `junimoserver`) = **SMAPI + mod
  « Always On Server »** (funny-snek). Pas de serveur dédié officiel.
- **Réseau** : ⚠️ Stardew utilise le **réseau Steam/GOG Galaxy (P2P)** → connexion par **code
  d'invitation**, **pas par IP:port** classique → ne colle **pas** au modèle du panel
  (affichage IP:port, allocation de ports). Intégration = cas particulier important à concevoir.
- **RAM** : léger (~1–2 Go). **Note** : valeur de **niche**, effort/risque d'intégration élevé
  à cause du modèle réseau. À garder pour plus tard, en bas de liste.

---

## Autres jeux populaires à considérer (hors demande initiale)

Candidats fréquents en hosting, par profil — à évaluer selon la demande :

- **Rust** (déjà « Bientôt » dans `games.ts`) — `didstopia/rust-server`, UDP `28015` + RCON,
  **lourd** (8 Go+, mappe lourd). Très demandé.
- **Project Zomboid** — `cyrale/project-zomboid` ou steamcmd, UDP `16261-16262`, 4–8 Go. Populaire,
  natif Linux, bon candidat moyen.
- **V Rising** — Wine (`gabekangas/vrising` / `trueosiris/...`), UDP `9876-9877`, 4–6 Go.
- **Factorio** — `factoriotools/factorio` (officiel-ish), UDP `34197`, **très léger** (~1–2 Go),
  natif Linux → excellent candidat léger/rentable.
- **Enshrouded** — Wine, UDP `15636-15637`, 8 Go+. Survie co-op récente, demande en hausse.
- **Core Keeper** — `escaping/core-keeper-dedicated`, léger, co-op.
- **Don't Starve Together** — `jamesits/dst-server` ou klei, léger, co-op.
- **Counter-Strike 2 / Garry's Mod / TF2** (Source) — SteamCMD/SRCDS, UDP `27015`, modéré ;
  écosystème pub/communautés important.
- **Sons of the Forest** — `jammsen/docker-sons-of-the-forest-dedicated-server` (Wine), jumeau de
  The Forest.

**Recommandation de séquençage** : commencer par les **natifs Linux légers/moyens** à forte
demande et faible empreinte — **Valheim → Terraria → 7 Days to Die → Factorio/Project Zomboid** —
avant les **lourds** (Palworld, ARK SE) et les **Wine** (The Forest, ARK SA, V Rising). Garder
les **lourds/gros lobbies** pour le **cloud à l'heure** (cf. backlog #S).

## Sources (vérifiées 2026-06-12)
- Palworld : [thijsvanloef/palworld-server-docker](https://github.com/thijsvanloef/palworld-server-docker), [docs officielles](https://docs.palworldgame.com/getting-started/requirements/)
- ARK: Survival Ascended : [azixus ASA Docker](https://azixus.github.io/ARK_Ascended_Docker/), [mschnitzer/ark-survival-ascended-linux-container-image](https://github.com/mschnitzer/ark-survival-ascended-linux-container-image), [acekorneya/asa_server](https://hub.docker.com/r/acekorneya/asa_server)
- Valheim : [lloesche/valheim-server](https://github.com/lloesche/valheim-server-docker)
- 7 Days to Die : [Steam guide Docker](https://steamcommunity.com/sharedfiles/filedetails/?id=2833928667)
- The Forest : [jammsen/docker-the-forest-dedicated-server](https://github.com/jammsen/docker-the-forest-dedicated-server)
- Terraria : [ryshe/terraria](https://hub.docker.com/r/ryshe/terraria), [beardedio/terraria](https://hub.docker.com/r/beardedio/terraria)
- Stardew Valley : [stardew-valley-dedicated-server/server](https://github.com/stardew-valley-dedicated-server/server), [Always On Server mod](https://www.nexusmods.com/stardewvalley/mods/2677)
