package servers

import "fmt"

// satisfactoryMsgOffset : décalage entre le port de base (jeu) et le port
// "reliable messaging" de Satisfactory. La plage des ports de base s'arrête à
// 26065 (voir ports.go) et la moitié haute (26066-26565) est réservée aux ports
// dérivés — ainsi un port messaging ne peut jamais entrer en collision avec le
// port de base d'un autre serveur.
const satisfactoryMsgOffset = 500

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

	// Planchers de ressources : appliqués au provisioning même si le plan
	// choisi est plus petit (ex. Satisfactory tourne en 4 Go même en "free").
	MinRAMMb    int64
	MinCPUCores float64

	// Ports retourne les bindings à publier pour un port de base alloué.
	Ports func(basePort int) []PortMapping
	// Env construit les variables d'environnement du container. basePort permet
	// aux jeux dont les ports sont configurables (Satisfactory) de s'aligner sur
	// le port alloué.
	Env func(plan Plan, version string, basePort int) []string
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
			// Patch 1.1 : deux ports distincts requis, redirection non supportée
			// pour le port de jeu → on configure le container pour écouter
			// exactement sur les ports alloués (hôte == interne).
			return []PortMapping{
				{HostPort: base, Internal: base, Proto: "udp"},                                                 // jeu (-Port)
				{HostPort: base + satisfactoryMsgOffset, Internal: base + satisfactoryMsgOffset, Proto: "tcp"}, // reliable messaging (-ReliablePort)
			}
		},
		Env: satisfactoryEnv,
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

// minecraftEnv : config pour l'image itzg/minecraft-server.
// TYPE=PAPER → performant + support plugins. Moddable via Fabric/Forge plus tard.
// version : "LATEST", "1.21.4", etc. basePort n'est pas utilisé (port interne figé).
func minecraftEnv(plan Plan, version string, _ int) []string {
	if version == "" {
		version = "LATEST"
	}
	// MEMORY = tas JVM = la RAM annoncée du plan. La limite mémoire du container
	// (HostConfig.Memory) est volontairement plus haute pour laisser de la marge
	// au non-heap (metaspace, threads, buffers directs, GC) — voir container.go.
	env := []string{
		"EULA=TRUE",
		"TYPE=PAPER",
		"VERSION=" + version,
		fmt.Sprintf("MEMORY=%dM", plan.RAMMb),
		"USE_AIKAR_FLAGS=true",
	}
	if plan.MaxSlots > 0 {
		env = append(env, fmt.Sprintf("MAX_PLAYERS=%d", plan.MaxSlots))
	}
	return env
}

// satisfactoryEnv : config pour l'image wolveix/satisfactory-server.
// Satisfactory n'a pas de notion de "version" exposée (version ignorée). Les
// ports sont alignés sur le port de base alloué (le messaging = base + offset).
func satisfactoryEnv(plan Plan, _ string, base int) []string {
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
