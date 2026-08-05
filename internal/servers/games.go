package servers

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// satisfactoryMsgOffset : décalage, DANS LE BLOC du serveur, entre le port de
// base (jeu) et le port "reliable messaging" de Satisfactory. Chaque serveur
// possède un bloc de ports contigus (voir portBlockSize dans ports.go), donc un
// simple +1 suffit : il reste dans le bloc et ne peut pas entrer en collision
// avec un autre serveur. (Doit rester < portBlockSize.)
const satisfactoryMsgOffset = 1

// PortMapping décrit un port à publier : port hôte → port interne du container,
// pour un protocole donné. Pour Minecraft, l'interne est figé (25565) et l'hôte
// est dynamique (remapping). Pour Satisfactory, hôte == interne (la redirection
// de port n'est pas supportée pour le port de jeu).
type PortMapping struct {
	HostPort int
	Internal int
	Proto    string // "tcp" | "udp"
}

// GameDef décrit tout ce qui est spécifique à un jeu : image Docker, ports,
// variables d'environnement, volume de données, routing et planchers de
// ressources. Ajouter un jeu = ajouter une entrée dans `games`.
type GameDef struct {
	ID              string
	Image           string
	DataPath        string // point de montage du volume nommé dans le container
	UsesMCRouter    bool   // routing par hostname via mc-router (Minecraft)
	RouterPort      int    // port interne ciblé par mc-router (si UsesMCRouter)
	SupportsVersion bool

	// NeedsAuth : le serveur exige une authentification interactive (OAuth
	// device-code) au premier démarrage (ex. Hytale). Déclenche le watcher
	// d'auth dans provisionServer et le statut "auth_required".
	NeedsAuth bool

	// Planchers de ressources : appliqués au provisioning même si le plan
	// choisi est plus petit (ex. Satisfactory tourne en 4 Go même en "free").
	MinRAMMb    int64
	MinCPUCores float64

	// SeedFiles : fichiers de jeu pré-téléchargés à injecter dans le volume avant
	// le 1er démarrage (depuis <SeedDir>/<ID>/<fichier> vers <DataPath>/<fichier>).
	// Permet de réutiliser un téléchargement local (Hytale) au lieu de re-télécharger
	// — et donc d'éviter l'OAuth "downloader" à chaque création. Le client n'a plus
	// que l'auth SERVEUR à faire. Si les fichiers sont absents → repli sur download.
	SeedFiles []string

	// ConsoleStdin : le serveur lit ses commandes sur STDIN (console interactive,
	// ex. Hytale → `discovery link <token>`). Le container doit être créé avec
	// OpenStdin ; les commandes sont envoyées via Docker attach (cf. SendStdin).
	// Les jeux à RCON (Minecraft) gardent ConsoleStdin=false → exec rcon-cli.
	ConsoleStdin bool

	// Ports retourne les bindings à publier pour un port de base alloué.
	Ports func(basePort int) []PortMapping
	// Env construit les variables d'environnement du container. basePort permet
	// aux jeux dont les ports sont configurables (Satisfactory) de s'aligner sur
	// le port alloué. loader = type de serveur (Minecraft : paper|fabric|forge),
	// modpack = réf de pack ("ftb:..."/"modrinth:..." ou "" = aucun). Tous deux ignorés
	// par les jeux qui n'en ont pas.
	Env func(plan Plan, version, loader, modpack string, basePort int) []string
}

// MinecraftLoaders : types de serveur Minecraft sélectionnables à la création.
//   paper  → plugins Bukkit/Spigot (dossier /plugins)
//   fabric → mods Fabric (dossier /mods)
//   forge  → mods Forge (dossier /mods)
var MinecraftLoaders = []string{"paper", "fabric", "forge"}

// NormalizeLoader valide/normalise le loader. Minecraft → paper|fabric|forge
// (défaut paper si invalide). Autres jeux → "paper" (non utilisé).
func NormalizeLoader(game, loader string) string {
	if game != "minecraft" {
		return "paper"
	}
	switch loader {
	case "fabric", "forge":
		return loader
	default:
		return "paper"
	}
}

// IsModded : true si le loader charge des mods (dossier /mods) au lieu de plugins.
func IsModded(loader string) bool {
	return loader == "fabric" || loader == "forge"
}

// MinCPUForModded : plancher CPU des serveurs moddés/modpacks. La génération de chunks
// (C2ME) profite de plusieurs cœurs ; NanoCPUs étant un plafond (pas une réservation),
// un serveur idle ne consomme pas ces cœurs → généreux sans coût réel.
const MinCPUForModded = 4.0

// perfProjects : mods de performance Modrinth injectés (via itzg MODRINTH_PROJECTS, qui
// résout la version compatible) pour un serveur moddé DIRECT (hors modpack). C2ME
// parallélise la génération de chunks (le gros gain « envoi des chunks »). Vide si loader
// inconnu → évite un mod incompatible qui empêcherait le démarrage. NE PAS utiliser pour
// un modpack : le pack inclut souvent déjà Lithium/FerriteCore → doublon = crash.
func perfProjects(loader string) string {
	// Slugs Modrinth exacts (vérifiés) — un slug erroné/incompatible fait planter itzg
	// au démarrage, donc on n'inclut que des mods très largement disponibles.
	switch loader {
	case "fabric":
		return "lithium,ferrite-core,c2me-fabric,krypton"
	case "forge":
		return "ferrite-core"
	}
	return ""
}

// javaTagForVersion : le tag d'image itzg (= version de Java) DÉPEND de la version de
// Minecraft, pas du loader. MC récent (schéma CalVer 26.x / LATEST) est compilé pour
// Java 25 et ne tourne PAS sous Java 21 ; à l'inverse, beaucoup de mods/packs 1.20.x–
// 1.21.x exigent Java 21 et refusent Java 25. On mappe donc :
//   LATEST / 26.x+  → `latest` (Java le plus récent)
//   1.x (1.18–1.21) → `java21`  (couvre l'immense majorité des mods/packs)
func javaTagForVersion(version string) string {
	if version == "" || strings.EqualFold(version, "LATEST") {
		return "latest"
	}
	if parts := strings.SplitN(version, ".", 2); len(parts) > 0 {
		if maj, err := strconv.Atoi(parts[0]); err == nil && maj >= 2 {
			return "latest" // nouveau schéma (26.x, 27.x…)
		}
	}
	return "java21"
}

// MinecraftImage : image itzg avec la bonne version de Java pour la version de MC.
func MinecraftImage(version string) string {
	return "itzg/minecraft-server:" + javaTagForVersion(version)
}

// ApplyFloor relève la RAM/CPU au plancher du jeu si nécessaire (sans jamais
// les abaisser sous ce que le plan offre).
func (g GameDef) ApplyFloor(ramMb int64, cpu float64) (int64, float64) {
	if g.MinRAMMb > ramMb {
		ramMb = g.MinRAMMb
	}
	if g.MinCPUCores > cpu {
		cpu = g.MinCPUCores
	}
	return ramMb, cpu
}

var games = map[string]GameDef{
	"minecraft": {
		ID:              "minecraft",
		Image:           "itzg/minecraft-server:latest",
		DataPath:        "/data",
		UsesMCRouter:    true,
		RouterPort:      25565,
		SupportsVersion: true,
		Ports: func(base int) []PortMapping {
			// MC écoute toujours sur 25565 dans le container ; le port hôte
			// dynamique y est mappé. mc-router route par hostname.
			return []PortMapping{{HostPort: base, Internal: 25565, Proto: "tcp"}}
		},
		Env: minecraftEnv,
	},
	"satisfactory": {
		ID:           "satisfactory",
		Image:        "wolveix/satisfactory-server:latest",
		DataPath:     "/config",
		UsesMCRouter: false, // protocole UDP, pas de routing par hostname → IP:port direct
		MinRAMMb:     4096,  // minimum jouable (offert gratuitement pour l'instant)
		MinCPUCores:  2.0,
		Ports: func(base int) []PortMapping {
			// Patch 1.1 : le port de jeu (base) écoute en UDP (trafic de jeu) ET
			// TCP (API/HTTPS), la redirection n'est pas supportée → hôte == interne.
			// Le reliable messaging est sur un 2e port TCP distinct (base+offset).
			return []PortMapping{
				{HostPort: base, Internal: base, Proto: "udp"},                                                 // jeu (-Port)
				{HostPort: base, Internal: base, Proto: "tcp"},                                                 // API/HTTPS du jeu
				{HostPort: base + satisfactoryMsgOffset, Internal: base + satisfactoryMsgOffset, Proto: "tcp"}, // reliable messaging (-ReliablePort)
			}
		},
		Env: satisfactoryEnv,
	},
	"hytale": {
		ID:           "hytale",
		Image:        "ghcr.io/terkea/hytale-server:latest",
		DataPath:     "/data",
		UsesMCRouter: false, // QUIC/UDP, pas de routing par hostname → IP:port direct
		NeedsAuth:    true,  // auth SERVEUR interactive faite par le CLIENT (device-code)
		MinRAMMb:     10240, // 10 Go : plancher réaliste/stable (4 Go crashe sous charge)
		MinCPUCores:  4.0,   // serveur Java lourd : 2 cœurs throttlent (lag aux actions), CPU abondant
		// Fichiers de jeu pré-téléchargés sur l'hôte (cf. scripts/hytale-seed) →
		// réutilisés à chaque création : pas de re-download ni d'OAuth downloader.
		SeedFiles: []string{"HytaleServer.jar", "Assets.zip"},
		// Console interactive via stdin (pas de RCON) → `discovery link <token>`, etc.
		ConsoleStdin: true,
		Ports: func(base int) []PortMapping {
			// Un seul port UDP (QUIC). On configure le serveur pour écouter sur
			// le port alloué (SERVER_PORT), mapping identité hôte == interne.
			return []PortMapping{{HostPort: base, Internal: base, Proto: "udp"}}
		},
		Env: hytaleEnv,
	},
	"valheim": {
		ID:           "valheim",
		Image:        "lloesche/valheim-server:latest",
		DataPath:     "/config",
		UsesMCRouter: false, // UDP, pas de routing par hostname → IP:port direct
		MinRAMMb:     4096,  // plancher jouable (4 Go min, 8 Go confortable)
		MinCPUCores:  2.0,
		Ports: func(base int) []PortMapping {
			// Valheim écoute en UDP sur SERVER_PORT et SERVER_PORT+1 (deux ports
			// consécutifs, dans le bloc du serveur). Identité hôte == interne car le
			// jeu annonce son port via Steam (pas de remapping possible).
			return []PortMapping{
				{HostPort: base, Internal: base, Proto: "udp"},         // jeu
				{HostPort: base + 1, Internal: base + 1, Proto: "udp"}, // +1 requis par Valheim
			}
		},
		Env: valheimEnv,
	},
	"7dtd": {
		ID: "7dtd",
		// LinuxGSM officiel : un seul volume /data (fichiers serveur + saves),
		// téléchargés au 1er démarrage via steamcmd.
		Image:        "gameservermanagers/gameserver:sdtd",
		DataPath:     "/data",
		UsesMCRouter: false, // UDP brut → IP:port direct
		MinRAMMb:     8192,  // monde vaste + IA des hordes : 8 Go minimum stable
		MinCPUCores:  2.0,
		Ports: func(base int) []PortMapping {
			// L'image n'expose pas le port en env → le serveur écoute sur ses
			// défauts (26900-26902) et Docker remappe vers le bloc alloué. La
			// connexion directe IP:port marche remappée ; seul l'annuaire Steam
			// annoncerait le port interne (sans importance : accès direct).
			return []PortMapping{
				{HostPort: base, Internal: 26900, Proto: "tcp"},     // login/query
				{HostPort: base, Internal: 26900, Proto: "udp"},     // jeu
				{HostPort: base + 1, Internal: 26901, Proto: "udp"}, // Steam networking
				{HostPort: base + 2, Internal: 26902, Proto: "udp"}, // Steam networking
			}
		},
		Env: sdtdEnv,
	},
	"project-zomboid": {
		ID:           "project-zomboid",
		Image:        "renegademaster/zomboid-dedicated-server:latest",
		DataPath:     "/home/steam/Zomboid", // config + saves (les fichiers du jeu se re-téléchargent)
		UsesMCRouter: false,                 // UDP brut → IP:port direct
		MinRAMMb:     4096,
		MinCPUCores:  2.0,
		Ports: func(base int) []PortMapping {
			// Ports configurés via env (DEFAULT_PORT/UDP_PORT) → hôte == interne,
			// alignés sur le bloc alloué.
			return []PortMapping{
				{HostPort: base, Internal: base, Proto: "udp"},         // jeu
				{HostPort: base + 1, Internal: base + 1, Proto: "udp"}, // connexions clients
			}
		},
		Env: zomboidEnv,
	},
	"palworld": {
		ID:           "palworld",
		Image:        "thijsvanloef/palworld-server-docker:latest",
		DataPath:     "/palworld",
		UsesMCRouter: false, // UDP brut → IP:port direct
		MinRAMMb:     8192,  // le serveur mange la RAM avec le temps (16 Go confortable)
		MinCPUCores:  4.0,
		Ports: func(base int) []PortMapping {
			// PORT (env) aligne le port de jeu sur le bloc alloué → hôte == interne.
			return []PortMapping{{HostPort: base, Internal: base, Proto: "udp"}}
		},
		Env: palworldEnv,
	},
	"eco": {
		ID: "eco",
		// Image officielle Strange Loop Games ; « default » = tag roulant.
		Image:        "strangeloopgames/eco-game-server:default",
		DataPath:     "/app/Storage", // monde + saves (les Configs par défaut restent dans l'image)
		UsesMCRouter: false,          // UDP brut → IP:port direct
		MinRAMMb:     4096,
		MinCPUCores:  2.0,
		Ports: func(base int) []PortMapping {
			// Pas de config de port en env → défauts du serveur (3000/udp jeu,
			// 3001/tcp interface web), remappés vers le bloc alloué.
			return []PortMapping{
				{HostPort: base, Internal: 3000, Proto: "udp"},     // jeu
				{HostPort: base + 1, Internal: 3001, Proto: "tcp"}, // interface web
			}
		},
		Env: ecoEnv,
	},
	"calradia-coop": {
		ID: "calradia-coop",
		// Image buildée LOCALEMENT sur xe80dell depuis le repo privé Calradia-Coop
		// (server/Dockerfile) — jamais poussée sur un registry (serveur confidentiel,
		// offre de location uniquement). pullImage a un repli image-locale.
		Image:        "calradia-server:latest",
		DataPath:     "/data", // CALRADIA_DATA_DIR : monde persisté + logs
		UsesMCRouter: false,   // TCP+UDP bruts → IP:port direct
		MinRAMMb:     1024,    // serveur Rust std-only, très léger
		MinCPUCores:  1.0,
		Ports: func(base int) []PortMapping {
			// Le monde écoute TCP (frames) ET UDP (canal positions) sur le MÊME
			// port, configuré via CALRADIA_ADDR → hôte == interne, aligné sur base.
			return []PortMapping{
				{HostPort: base, Internal: base, Proto: "tcp"}, // monde (login, save-sync, relay)
				{HostPort: base, Internal: base, Proto: "udp"}, // positions haute fréquence
			}
		},
		Env: calradiaEnv,
	},
}

// GetGame retourne la définition d'un jeu, ou une erreur si l'id est inconnu.
func GetGame(id string) (GameDef, error) {
	g, ok := games[id]
	if !ok {
		return GameDef{}, fmt.Errorf("unknown game %q", id)
	}
	return g, nil
}

// GameIDs retourne les identifiants de jeux disponibles (ordre stable).
func GameIDs() []string {
	ids := make([]string, 0, len(games))
	for id := range games {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// minecraftEnv : config pour l'image itzg/minecraft-server.
// Si modpack est défini → l'image installe le pack complet (TYPE=FTBA / MODRINTH), qui
// fixe lui-même loader+version+mods. Sinon : loader → TYPE paper/fabric/forge + VERSION.
// modpack : "ftb:<packId>:<verId>" ou "modrinth:<projId>:<verId>" (verId optionnel).
func minecraftEnv(plan Plan, version, loader, modpack string, _ int) []string {
	// MEMORY = tas JVM = la RAM annoncée du plan. La limite mémoire du container
	// (HostConfig.Memory) est volontairement plus haute pour laisser de la marge
	// au non-heap (metaspace, threads, buffers directs, GC) — voir container.go.
	env := []string{
		"EULA=TRUE",
		fmt.Sprintf("MEMORY=%dM", plan.RAMMb),
		"USE_AIKAR_FLAGS=true",
	}
	if mp := strings.SplitN(modpack, ":", 3); modpack != "" && len(mp) >= 2 {
		id, ver := mp[1], ""
		if len(mp) > 2 {
			ver = mp[2]
		}
		switch mp[0] {
		case "ftb":
			env = append(env, "TYPE=FTBA", "FTB_MODPACK_ID="+id)
			if ver != "" {
				env = append(env, "FTB_MODPACK_VERSION_ID="+ver)
			}
		case "modrinth":
			env = append(env, "TYPE=MODRINTH", "MODRINTH_MODPACK="+id)
			if ver != "" {
				env = append(env, "MODRINTH_VERSION="+ver)
			}
		}
	} else {
		if version == "" {
			version = "LATEST"
		}
		typ := "PAPER"
		switch loader {
		case "fabric":
			typ = "FABRIC"
		case "forge":
			typ = "FORGE"
		}
		env = append(env, "TYPE="+typ, "VERSION="+version)
	}
	// Optimisations serveurs moddés/modpacks : view-distance plus basse (moins de chunks
	// à générer/envoyer). Pour les serveurs moddés DIRECTS (hors modpack), on ajoute aussi
	// les mods de performance (C2ME = génération parallèle, le gros gain). Jamais pour un
	// modpack : risque de doublon avec les mods du pack (→ crash au démarrage).
	if modpack != "" || IsModded(loader) {
		env = append(env, "VIEW_DISTANCE=8", "SIMULATION_DISTANCE=6")
		if modpack == "" {
			if p := perfProjects(loader); p != "" {
				// MODRINTH_ALLOWED_VERSION_TYPE=alpha : C2ME ne publie qu'en alpha, et itzg
				// refuse les non-release par défaut → sans ça il planterait au démarrage.
				env = append(env, "MODRINTH_PROJECTS="+p, "MODRINTH_ALLOWED_VERSION_TYPE=alpha")
			}
		}
	}
	if plan.MaxSlots > 0 {
		env = append(env, fmt.Sprintf("MAX_PLAYERS=%d", plan.MaxSlots))
	}
	return env
}

// satisfactoryEnv : config pour l'image wolveix/satisfactory-server.
// Satisfactory n'a pas de notion de "version" exposée (version ignorée). Les
// ports sont alignés sur le port de base alloué (le messaging = base + offset).
func satisfactoryEnv(plan Plan, _, _, _ string, base int) []string {
	env := []string{
		"STEAMBETA=false",
		fmt.Sprintf("SERVERGAMEPORT=%d", base),
		fmt.Sprintf("SERVERMESSAGINGPORT=%d", base+satisfactoryMsgOffset),
	}
	if plan.MaxSlots > 0 {
		env = append(env, fmt.Sprintf("MAXPLAYERS=%d", plan.MaxSlots))
	}
	return env
}

// hytaleEnv : config pour l'image ghcr.io/terkea/hytale-server (serveur Java 25).
// Le port QUIC est aligné sur le port de base alloué. L'image télécharge et
// authentifie le serveur au 1er démarrage (OAuth) — voir le watcher d'auth.
func hytaleEnv(plan Plan, _, _, _ string, base int) []string {
	env := []string{
		fmt.Sprintf("SERVER_PORT=%d", base),
		fmt.Sprintf("MEMORY=%dM", plan.RAMMb),
		"AUTO_DOWNLOAD=true",
		"AUTO_UPDATE=true",
	}
	if plan.MaxSlots > 0 {
		env = append(env, fmt.Sprintf("MAX_PLAYERS=%d", plan.MaxSlots))
	}
	return env
}

// valheimEnv : config pour l'image lloesche/valheim-server. Le port de jeu est
// aligné sur le port de base alloué (Valheim écoute base et base+1 en UDP).
// SERVER_PASS : Valheim exige un mot de passe (≥5 car., différent du nom) — on
// pose un défaut tant que le panel ne le collecte pas à la création (à raffiner).
// Valheim plafonne nativement à 10 joueurs (pas de MAX_PLAYERS configurable).
func valheimEnv(_ Plan, _, _, _ string, base int) []string {
	return []string{
		"SERVER_NAME=Playrena Valheim",
		fmt.Sprintf("SERVER_PORT=%d", base),
		"WORLD_NAME=Playrena",
		"SERVER_PASS=playrena",
		"SERVER_PUBLIC=1",
	}
}

// sdtdEnv : config pour l'image LinuxGSM (gameservermanagers/gameserver:sdtd).
// LinuxGSM se configure surtout par fichiers ; l'env sert au user mapping. Les
// ports restent les défauts du jeu (remappés par Docker, cf. Ports).
func sdtdEnv(_ Plan, _, _, _ string, _ int) []string {
	return []string{
		"UID=1000",
		"GID=1000",
	}
}

// zomboidEnv : config pour l'image renegademaster/zomboid-dedicated-server.
// Ports alignés sur le bloc alloué via env. MAX_RAM = tas JVM : la RAM du plan,
// mais jamais sous 3 Go — le container est de toute façon relevé au plancher du
// jeu (4 Go), et un tas d'1 Go (plan free) part en OutOfMemoryError au premier
// chargement de Knox County. La limite du container garde une marge (container.go).
// Mot de passe admin dérivé plus tard si on expose la console — défaut sûr en attendant.
func zomboidEnv(plan Plan, _, _, _ string, base int) []string {
	heapMb := plan.RAMMb
	if heapMb < 3072 {
		heapMb = 3072
	}
	env := []string{
		fmt.Sprintf("DEFAULT_PORT=%d", base),
		fmt.Sprintf("UDP_PORT=%d", base+1),
		fmt.Sprintf("MAX_RAM=%dm", heapMb),
		"SERVER_NAME=Playrena Zomboid",
		"ADMIN_USERNAME=playrena_admin",
		"ADMIN_PASSWORD=playrena-admin",
	}
	if plan.MaxSlots > 0 {
		env = append(env, fmt.Sprintf("MAX_PLAYERS=%d", plan.MaxSlots))
	}
	return env
}

// palworldEnv : config pour l'image thijsvanloef/palworld-server-docker.
// PORT (env) aligne le port de jeu sur le bloc alloué. Palworld plafonne à
// 32 joueurs. COMMUNITY=false : pas de listing public, on rejoint par IP:port.
func palworldEnv(plan Plan, _, _, _ string, base int) []string {
	players := plan.MaxSlots
	if players <= 0 || players > 32 {
		players = 32
	}
	return []string{
		fmt.Sprintf("PORT=%d", base),
		fmt.Sprintf("PLAYERS=%d", players),
		"SERVER_NAME=Playrena Palworld",
		"COMMUNITY=false",
		"PUID=1000",
		"PGID=1000",
	}
}

// ecoEnv : config pour l'image officielle strangeloopgames/eco-game-server.
// Eco se configure par fichiers (Configs/*.eco) — pas d'env de port ; les
// défauts (3000/3001) sont remappés par Docker. Rien à injecter pour l'instant.
func ecoEnv(_ Plan, _, _, _ string, _ int) []string {
	return nil
}

// calradiaEnv : config pour l'image locale calradia-server (mod Calradia-Coop,
// Mount & Blade II: Bannerlord). Le serveur écoute TCP+UDP sur CALRADIA_ADDR,
// aligné sur le port de base alloué. UPnP coupé : en bridge Docker il
// annoncerait l'IP 172.x du container au routeur (injoignable) — le NAT passe
// par le publish Docker + la plage de ports Playrena. Joueurs gameplay = slots
// du plan, plafonnés à 8 (limite testée du mod v0.0.1) ; les spectateurs
// gardent le défaut du serveur (32).
func calradiaEnv(plan Plan, _, _, _ string, base int) []string {
	gameplay := plan.MaxSlots
	if gameplay <= 0 || gameplay > 8 {
		gameplay = 8
	}
	return []string{
		fmt.Sprintf("CALRADIA_ADDR=0.0.0.0:%d", base),
		"CALRADIA_DATA_DIR=/data",
		"CALRADIA_UPNP=0",
		fmt.Sprintf("CALRADIA_MAX_GAMEPLAY=%d", gameplay),
	}
}
