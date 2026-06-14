export type GameStatus = "live" | "soon";

export interface Game {
  id: string;
  name: string;
  href: string; // "#" si pas de page dédiée
  status: GameStatus;
  accent: string; // dégradé Tailwind
  minRamGb?: number; // plancher RAM imposé au provisioning (reflète le backend)
  freeAtFloor?: boolean; // offert (promo) : gratuit à son plancher, plans payants masqués
}

export const GAMES: Game[] = [
  { id: "minecraft", name: "Minecraft", href: "/games/minecraft", status: "live", accent: "from-green-500 to-emerald-600" },
  { id: "hytale", name: "Hytale", href: "/games/hytale", status: "live", accent: "from-blue-500 to-indigo-600", minRamGb: 10, freeAtFloor: true },
  { id: "rust", name: "Rust", href: "/games/rust", status: "soon", accent: "from-red-500 to-orange-600" },
  { id: "satisfactory", name: "Satisfactory", href: "/games/satisfactory", status: "live", accent: "from-orange-500 to-amber-600", minRamGb: 4, freeAtFloor: true },
  { id: "ark", name: "ARK: Survival", href: "/games/ark", status: "soon", accent: "from-fuchsia-500 to-purple-600" },
];

export function getGame(id: string): Game | undefined {
  return GAMES.find((g) => g.id === id);
}

// Thème couleur par jeu — classes Tailwind LITTÉRALES (pas de construction
// dynamique, sinon Tailwind ne les génère pas).
export interface GameTheme {
  text: string;   // couleur d'accent (texte/icône)
  btn: string;    // bouton principal (bg + hover + ombre)
  border: string; // bordure d'accent
  glow: string;   // halo de fond (hero)
}

export const GAME_THEME: Record<string, GameTheme> = {
  minecraft:    { text: "text-green-400", btn: "bg-green-500 hover:bg-green-400 shadow-green-500/30", border: "border-green-500/30", glow: "bg-green-500/15" },
  hytale:       { text: "text-blue-400",  btn: "bg-blue-500 hover:bg-blue-400 shadow-blue-500/30",   border: "border-blue-500/30",  glow: "bg-blue-500/15" },
  satisfactory: { text: "text-amber-400", btn: "bg-amber-500 hover:bg-amber-400 shadow-amber-500/30", border: "border-amber-500/30", glow: "bg-amber-500/15" },
  rust:         { text: "text-orange-400", btn: "bg-orange-500 hover:bg-orange-400 shadow-orange-500/30", border: "border-orange-500/30", glow: "bg-orange-500/15" },
  ark:          { text: "text-fuchsia-400", btn: "bg-fuchsia-500 hover:bg-fuchsia-400 shadow-fuchsia-500/30", border: "border-fuchsia-500/30", glow: "bg-fuchsia-500/15" },
};

export const gameTheme = (id: string): GameTheme =>
  GAME_THEME[id] ?? GAME_THEME.minecraft;
