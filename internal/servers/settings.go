package servers

import (
	"fmt"
	"regexp"
	"strconv"
)

// SettingDef : un réglage de jeu éditable par le propriétaire du serveur
// (carte, nom de session, mot de passe, difficulté, multiplicateurs…). La
// valeur est stockée en DB (game_servers.settings, JSONB) et appliquée en
// SURCHARGE d'env à la (re)création du container : buildSpec l'ajoute APRÈS
// l'env de base du jeu, et Docker garde la dernière occurrence — pas besoin
// de toucher aux env builders. Valeur vide = retour au défaut du jeu.
//
// Deux modes de projection vers l'env :
//   - Env seul   : `ENV=valeur` (ex. SERVER_MAP).
//   - Env + Query : le réglage est un paramètre de la query-string UE
//     (`?Query=valeur`) ; tous les réglages Query d'un même Env sont composés
//     dans UNE variable, à partir de la base du jeu (cf. gameQueryEnv), ex.
//     ark → `EXTRA_SETTINGS=?listen?OverrideOfficialDifficulty=5?DifficultyOffset=1`.
type SettingDef struct {
	Key     string   `json:"key"`               // identifiant stable (libellé i18n côté frontend)
	Env     string   `json:"-"`                 // variable d'env surchargée
	Query   string   `json:"-"`                 // paramètre ?Query=… composé dans Env (optionnel)
	Extra   string   `json:"-"`                 // suffixe fixe ajouté à la query quand la valeur est posée
	Type    string   `json:"type"`              // "text" | "password" | "select" | "number"
	Options []string `json:"options,omitempty"` // valeurs permises (select)
	Default string   `json:"default"`           // défaut effectif (informatif, côté UI)
	MaxLen  int      `json:"max_len"`
	Min     float64  `json:"min,omitempty"`  // bornes (number)
	Max     float64  `json:"max,omitempty"`  //
	Step    float64  `json:"step,omitempty"` // pas suggéré à l'UI (number)
	Pattern string   `json:"-"`              // regex de validation (voir patterns ci-dessous)
}

// Patterns de validation. Les valeurs finissent dans l'env Docker (sans shell,
// donc sans risque d'injection côté hôte), mais certains jeux les repassent
// dans leur propre parsing (ARK : query-string UE « ?SessionName=… ») → on
// reste conservateur : jamais de retour à la ligne, de « ? » ni de « = ».
const (
	patToken = `^[A-Za-z0-9._-]*$`       // identifiant sans espaces (nom de monde, session ARK)
	patText  = `^[A-Za-z0-9 ,.!'’._-]*$` // texte court affiché (nom de serveur, MOTD)
	patPass  = `^[A-Za-z0-9@#%+.!?_-]*$` // mot de passe (pas d'espace ni de quotes)
	patNum   = `^\d{1,4}(\.\d{1,2})?$`   // nombre décimal positif (multiplicateurs)
)

// gameQueryEnv : pour les jeux à réglages « Query », la variable qui reçoit la
// query-string composée et sa base fixe (toujours présente, même sans réglage).
// Doit rester cohérent avec l'env de base du jeu (games.go).
var gameQueryEnv = map[string]struct{ Env, Base string }{
	"ark": {Env: "EXTRA_SETTINGS", Base: arkExtraSettingsBase},
}

// Multiplicateur ARK standard : 0.1 → 100, défaut 1.
func arkMult(key, query string) SettingDef {
	return SettingDef{Key: key, Env: "EXTRA_SETTINGS", Query: query, Type: "number",
		Default: "1", Min: 0.1, Max: 100, Step: 0.5, Pattern: patNum}
}

// gameSettings : réglages exposés par jeu. Un jeu absent n'expose rien.
// Les jeux configurés par fichiers (7DTD, Eco) ou au flux spécial (Hytale,
// Satisfactory, Calradia) ne sont pas couverts ici.
var gameSettings = map[string][]SettingDef{
	"ark": {
		// Cartes officielles ASA (suffixe _WP = UE5).
		{Key: "map", Env: "SERVER_MAP", Type: "select", Default: "TheIsland_WP",
			Options: []string{"TheIsland_WP", "ScorchedEarth_WP", "TheCenter_WP", "Aberration_WP", "Extinction_WP", "Ragnarok_WP"}},
		// Sans espaces : le nom passe dans la query-string UE (troncature observée).
		{Key: "session_name", Env: "SESSION_NAME", Type: "text", Default: "Playrena-ARK", MaxLen: 32, Pattern: patToken},
		{Key: "password", Env: "SERVER_PASSWORD", Type: "password", MaxLen: 32, Pattern: patPass},
		{Key: "admin_password", Env: "SERVER_ADMIN_PASSWORD", Type: "password", Default: "playrena-admin", MaxLen: 32, Pattern: patPass},
		// Gameplay (paramètres ?… de la commande ASA). Difficulté : niveau max des
		// dinos = valeur × 30 (5 = 150 comme l'officiel) ; DifficultyOffset=1 est
		// requis pour que l'override s'applique.
		{Key: "difficulty", Env: "EXTRA_SETTINGS", Query: "OverrideOfficialDifficulty", Extra: "?DifficultyOffset=1",
			Type: "select", Default: "5", Options: []string{"1", "2", "3", "4", "5", "6", "7", "8", "10"}},
		arkMult("xp", "XPMultiplier"),
		arkMult("taming", "TamingSpeedMultiplier"),
		arkMult("harvest", "HarvestAmountMultiplier"),
		arkMult("baby_mature", "BabyMatureSpeedMultiplier"),
		arkMult("mating_interval", "MatingIntervalMultiplier"),
		{Key: "pve", Env: "EXTRA_SETTINGS", Query: "ServerPVE", Type: "select", Default: "False", Options: []string{"True", "False"}},
		{Key: "show_map_location", Env: "EXTRA_SETTINGS", Query: "ShowMapPlayerLocation", Type: "select", Default: "False", Options: []string{"True", "False"}},
	},
	"valheim": {
		{Key: "server_name", Env: "SERVER_NAME", Type: "text", Default: "Playrena Valheim", MaxLen: 32, Pattern: patText},
		// Changer de nom de monde = nouveau monde (l'ancien reste dans le volume).
		{Key: "world_name", Env: "WORLD_NAME", Type: "text", Default: "Playrena", MaxLen: 24, Pattern: patToken},
		// Valheim exige ≥ 5 caractères, différent du nom — validé au POST.
		{Key: "password", Env: "SERVER_PASS", Type: "password", Default: "playrena", MaxLen: 32, Pattern: patPass},
	},
	"palworld": {
		{Key: "server_name", Env: "SERVER_NAME", Type: "text", Default: "Playrena Palworld", MaxLen: 32, Pattern: patText},
	},
	"project-zomboid": {
		{Key: "server_name", Env: "SERVER_NAME", Type: "text", Default: "Playrena Zomboid", MaxLen: 32, Pattern: patText},
		{Key: "admin_password", Env: "ADMIN_PASSWORD", Type: "password", Default: "playrena-admin", MaxLen: 32, Pattern: patPass},
	},
	"minecraft": {
		{Key: "motd", Env: "MOTD", Type: "text", MaxLen: 59, Pattern: patText},
	},
}

// SettingsFor retourne les réglages éditables d'un jeu (nil si aucun).
func SettingsFor(game string) []SettingDef {
	return gameSettings[game]
}

// ValidateSettings vérifie des valeurs soumises contre les défs du jeu :
// clé inconnue refusée, select hors options refusé, number hors bornes
// refusé, longueur et charset contrôlés. Une valeur vide est permise
// (= retour au défaut) et sera simplement omise de la surcharge d'env.
func ValidateSettings(game string, values map[string]string) error {
	defs := gameSettings[game]
	byKey := make(map[string]SettingDef, len(defs))
	for _, d := range defs {
		byKey[d.Key] = d
	}
	for k, v := range values {
		d, ok := byKey[k]
		if !ok {
			return fmt.Errorf("réglage inconnu %q pour %s", k, game)
		}
		if v == "" {
			continue
		}
		if d.MaxLen > 0 && len(v) > d.MaxLen {
			return fmt.Errorf("%s : %d caractères max", k, d.MaxLen)
		}
		if d.Type == "select" {
			found := false
			for _, o := range d.Options {
				found = found || o == v
			}
			if !found {
				return fmt.Errorf("%s : valeur hors liste", k)
			}
		}
		if d.Pattern != "" && !regexp.MustCompile(d.Pattern).MatchString(v) {
			return fmt.Errorf("%s : caractères non permis", k)
		}
		if d.Type == "number" {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil || f < d.Min || f > d.Max {
				return fmt.Errorf("%s : nombre entre %g et %g attendu", k, d.Min, d.Max)
			}
		}
	}
	// Règle propre à Valheim : mot de passe ≥ 5 caractères (exigence du jeu —
	// un mot de passe trop court empêche le serveur de démarrer).
	if game == "valheim" {
		if v := values["password"]; v != "" && len(v) < 5 {
			return fmt.Errorf("password : Valheim exige au moins 5 caractères")
		}
	}
	return nil
}

// SettingsEnv construit la surcharge d'env depuis les valeurs stockées (les
// valeurs vides ou non définies laissent le défaut du jeu). Les réglages
// « Query » d'un même Env sont composés en une seule variable, sur la base
// fixe du jeu (gameQueryEnv).
func SettingsEnv(game string, values map[string]string) []string {
	var env []string
	queries := map[string]string{} // Env → query-string composée
	for _, d := range gameSettings[game] {
		v := values[d.Key]
		if v == "" {
			continue
		}
		if d.Query == "" {
			env = append(env, d.Env+"="+v)
			continue
		}
		if _, ok := queries[d.Env]; !ok {
			queries[d.Env] = gameQueryEnv[game].Base
		}
		queries[d.Env] += "?" + d.Query + "=" + v + d.Extra
	}
	for _, d := range gameSettings[game] { // ordre stable des défs
		if q, ok := queries[d.Env]; ok && d.Query != "" {
			env = append(env, d.Env+"="+q)
			delete(queries, d.Env)
		}
	}
	return env
}
