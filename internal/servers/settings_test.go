package servers

import (
	"strings"
	"testing"
)

// La surcharge d'env des réglages doit s'appuyer sur les défs du jeu : valeur
// posée → KEY=VAL émis, valeur vide/absente → rien (le défaut du jeu reste).
func TestSettingsEnvOverrides(t *testing.T) {
	env := SettingsEnv("ark", map[string]string{
		"map":          "Ragnarok_WP",
		"session_name": "",
	})
	joined := strings.Join(env, "\n")
	if !strings.Contains(joined, "SERVER_MAP=Ragnarok_WP") {
		t.Fatalf("SERVER_MAP manquant : %q", env)
	}
	if strings.Contains(joined, "SESSION_NAME") {
		t.Fatalf("une valeur vide ne doit pas surcharger : %q", env)
	}
	if got := SettingsEnv("satisfactory", map[string]string{"x": "y"}); got != nil {
		t.Fatalf("jeu sans réglages → rien, obtenu %q", got)
	}
}

// Les réglages gameplay ARK sont des paramètres ?… composés dans UNE seule
// variable EXTRA_SETTINGS, à partir de la base fixe (?listen), dans l'ordre
// des défs ; la difficulté entraîne son suffixe DifficultyOffset.
func TestSettingsEnvComposesQuery(t *testing.T) {
	env := SettingsEnv("ark", map[string]string{
		"difficulty": "5",
		"taming":     "3",
		"xp":         "2",
		"pve":        "True",
	})
	want := "EXTRA_SETTINGS=?listen?OverrideOfficialDifficulty=5?DifficultyOffset=1?XPMultiplier=2?TamingSpeedMultiplier=3?ServerPVE=True"
	found := false
	for _, e := range env {
		if e == want {
			found = true
		}
		if strings.HasPrefix(e, "EXTRA_SETTINGS=") && e != want {
			t.Fatalf("composition inattendue : %q", e)
		}
	}
	if !found {
		t.Fatalf("attendu %q dans %q", want, env)
	}
	// Sans réglage Query, EXTRA_SETTINGS n'est pas surchargé (la base de
	// arkEnv reste).
	if env := SettingsEnv("ark", map[string]string{"map": "TheCenter_WP"}); len(env) != 1 || env[0] != "SERVER_MAP=TheCenter_WP" {
		t.Fatalf("surcharge inattendue : %q", env)
	}
}

func TestValidateSettings(t *testing.T) {
	cases := []struct {
		game   string
		values map[string]string
		ok     bool
	}{
		{"ark", map[string]string{"map": "Ragnarok_WP"}, true},
		{"ark", map[string]string{"map": "NotAMap"}, false},                        // hors options
		{"ark", map[string]string{"session_name": "Mon Serveur"}, false},           // espace interdit (query-string UE)
		{"ark", map[string]string{"session_name": "Tribu-de-Simon"}, true},
		{"ark", map[string]string{"inconnu": "x"}, false},                          // clé inconnue
		{"ark", map[string]string{"password": strings.Repeat("a", 40)}, false},     // trop long
		{"valheim", map[string]string{"password": "abc"}, false},                   // Valheim exige ≥ 5
		{"valheim", map[string]string{"password": "abcde"}, true},
		{"valheim", map[string]string{"server_name": "Les Vikings de Sherbrooke"}, true},
		{"minecraft", map[string]string{"motd": "Bienvenue chez Simon !"}, true},
		{"minecraft", map[string]string{"motd": "a\nb"}, false},                    // retour à la ligne
		{"ark", map[string]string{"xp": "2.5"}, true},
		{"ark", map[string]string{"xp": "0"}, false},                               // sous le min
		{"ark", map[string]string{"xp": "1e3"}, false},                             // pas un décimal simple
		{"ark", map[string]string{"difficulty": "12"}, false},                      // hors liste
	}
	for _, c := range cases {
		err := ValidateSettings(c.game, c.values)
		if c.ok && err != nil {
			t.Errorf("%s %v : erreur inattendue %v", c.game, c.values, err)
		}
		if !c.ok && err == nil {
			t.Errorf("%s %v : erreur attendue", c.game, c.values)
		}
	}
}
