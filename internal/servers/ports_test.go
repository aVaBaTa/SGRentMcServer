package servers

import "testing"

// blockOf : tous les ports d'un serveur (jeu + dérivés) doivent tomber dans le
// même bloc [base, base+portBlockSize).
func inSameBlock(t *testing.T, base int, ports ...int) {
	t.Helper()
	lo, hi := base, base+portBlockSize-1
	for _, p := range ports {
		if p < lo || p > hi {
			t.Fatalf("port %d hors du bloc [%d,%d] (base=%d)", p, lo, hi, base)
		}
	}
}

func TestBlockBaseOf(t *testing.T) {
	cases := map[int]int{
		portRangeStart:                   portRangeStart,
		portRangeStart + 1:               portRangeStart,
		portRangeStart + portBlockSize:   portRangeStart + portBlockSize,
		portRangeStart + portBlockSize*2: portRangeStart + portBlockSize*2,
	}
	for in, want := range cases {
		if got := blockBaseOf(in); got != want {
			t.Errorf("blockBaseOf(%d) = %d, want %d", in, got, want)
		}
	}
}

func TestFirstFreeBlockBase_Empty(t *testing.T) {
	got, err := firstFreeBlockBase(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != portRangeStart {
		t.Errorf("premier bloc = %d, want %d", got, portRangeStart)
	}
}

// Régression : avec l'ancien allocateur (un port à la fois), un 2e serteur
// recevait base+1 et son port de jeu chevauchait le port dérivé du 1er. Le
// schéma par blocs doit garantir que deux bases successives sont espacées d'au
// moins portBlockSize → aucun chevauchement possible.
func TestFirstFreeBlockBase_NoConsecutiveOverlap(t *testing.T) {
	first, err := firstFreeBlockBase(nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := firstFreeBlockBase([]usedPort{{base: first, game: "minecraft"}})
	if err != nil {
		t.Fatal(err)
	}
	if second-first < portBlockSize {
		t.Fatalf("blocs trop proches : first=%d second=%d (écart %d < %d)", first, second, second-first, portBlockSize)
	}
	// Le bloc du 1er serveur [first, first+15] ne doit jamais contenir le 2e base.
	inSameBlock(t, first, first) // sanity
	if second >= first && second <= first+portBlockSize-1 {
		t.Fatalf("base du 2e serveur %d tombe dans le bloc du 1er [%d,%d]", second, first, first+portBlockSize-1)
	}
}

// Les ports réels de Satisfactory (jeu = base, messaging = base+offset) doivent
// rester dans le bloc du serveur — c'est la garantie qui remplace l'ancien
// décalage +500 et la "demi-haute réservée".
func TestSatisfactoryPortsStayInBlock(t *testing.T) {
	base, err := firstFreeBlockBase(nil)
	if err != nil {
		t.Fatal(err)
	}
	g, err := GetGame("satisfactory")
	if err != nil {
		t.Fatal(err)
	}
	for _, pm := range g.Ports(base) {
		inSameBlock(t, base, pm.HostPort, pm.Internal)
	}
}

// Compat : un ancien serveur Satisfactory écoutant encore sur base+500 (hors
// bloc) doit voir ce port réservé → aucun nouveau bloc ne se place dessus.
func TestLegacySatisfactoryMessagingReserved(t *testing.T) {
	legacy := usedPort{base: portRangeStart, game: "satisfactory"}
	got, err := firstFreeBlockBase([]usedPort{legacy})
	if err != nil {
		t.Fatal(err)
	}
	legacyMsgBlock := blockBaseOf(portRangeStart + legacySatisfactoryMsgOffset)
	if got == legacyMsgBlock {
		t.Fatalf("nouveau bloc %d recouvre le port messaging legacy (bloc %d)", got, legacyMsgBlock)
	}
}

// La plage doit offrir un nombre de blocs cohérent et tous valides (dans la
// plage pare-feu 25566-26565).
func TestBlockExhaustion(t *testing.T) {
	var used []usedPort
	count := 0
	for {
		base, err := firstFreeBlockBase(used)
		if err != nil {
			break
		}
		if base+portBlockSize-1 > portRangeEnd {
			t.Fatalf("bloc %d dépasse portRangeEnd %d", base, portRangeEnd)
		}
		used = append(used, usedPort{base: base, game: "minecraft"})
		count++
		if count > 1000 {
			t.Fatal("boucle infinie : firstFreeBlockBase ne s'épuise pas")
		}
	}
	if count == 0 {
		t.Fatal("aucun bloc disponible")
	}
	t.Logf("%d blocs disponibles dans %d-%d", count, portRangeStart, portRangeEnd)
}
