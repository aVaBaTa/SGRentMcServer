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
  // --- En ligne ---
  { id: "minecraft", name: "Minecraft", href: "/games/minecraft", status: "live", accent: "from-green-500 to-emerald-600" },
  { id: "hytale", name: "Hytale", href: "/games/hytale", status: "live", accent: "from-blue-500 to-indigo-600", minRamGb: 10, freeAtFloor: true },
  { id: "satisfactory", name: "Satisfactory", href: "/games/satisfactory", status: "live", accent: "from-orange-500 to-amber-600", minRamGb: 4, freeAtFloor: true },
  { id: "valheim", name: "Valheim", href: "/games/valheim", status: "live", accent: "from-sky-500 to-cyan-600", minRamGb: 4 },

  // --- Bientôt (jeux populaires avec serveur dédié — href "#" tant qu'il n'y a pas de page) ---
  { id: "rust", name: "Rust", href: "/games/rust", status: "soon", accent: "from-red-500 to-orange-600" },
  { id: "ark", name: "ARK: Survival", href: "/games/ark", status: "soon", accent: "from-fuchsia-500 to-purple-600" },
  { id: "palworld", name: "Palworld", href: "#", status: "soon", accent: "from-yellow-400 to-amber-500" },
  { id: "terraria", name: "Terraria", href: "#", status: "soon", accent: "from-emerald-500 to-teal-600" },
  { id: "7dtd", name: "7 Days to Die", href: "#", status: "soon", accent: "from-amber-700 to-red-800" },
  { id: "project-zomboid", name: "Project Zomboid", href: "#", status: "soon", accent: "from-rose-600 to-red-700" },
  { id: "v-rising", name: "V Rising", href: "#", status: "soon", accent: "from-purple-700 to-fuchsia-800" },
  { id: "enshrouded", name: "Enshrouded", href: "#", status: "soon", accent: "from-violet-500 to-purple-700" },
  { id: "sons-of-the-forest", name: "Sons of the Forest", href: "#", status: "soon", accent: "from-emerald-700 to-green-900" },
  { id: "conan-exiles", name: "Conan Exiles", href: "#", status: "soon", accent: "from-orange-600 to-amber-700" },
  { id: "factorio", name: "Factorio", href: "#", status: "soon", accent: "from-yellow-600 to-amber-700" },
  { id: "core-keeper", name: "Core Keeper", href: "#", status: "soon", accent: "from-indigo-500 to-violet-700" },
  { id: "dst", name: "Don't Starve Together", href: "#", status: "soon", accent: "from-stone-500 to-zinc-700" },
  { id: "cs2", name: "Counter-Strike 2", href: "#", status: "soon", accent: "from-amber-500 to-yellow-600" },
  { id: "gmod", name: "Garry's Mod", href: "#", status: "soon", accent: "from-sky-600 to-indigo-700" },
  { id: "unturned", name: "Unturned", href: "#", status: "soon", accent: "from-lime-500 to-green-600" },
  { id: "bannerlord", name: "Bannerlord (Together)", href: "#", status: "soon", accent: "from-rose-600 to-red-800" },
  { id: "vintage-story", name: "Vintage Story", href: "#", status: "soon", accent: "from-teal-600 to-emerald-700" },
  { id: "necesse", name: "Necesse", href: "#", status: "soon", accent: "from-cyan-500 to-sky-600" },
  { id: "eco", name: "Eco", href: "#", status: "soon", accent: "from-green-500 to-lime-600" },
  { id: "soulmask", name: "Soulmask", href: "#", status: "soon", accent: "from-lime-600 to-emerald-700" },
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
  valheim:      { text: "text-cyan-400",  btn: "bg-cyan-500 hover:bg-cyan-400 shadow-cyan-500/30",   border: "border-cyan-500/30",  glow: "bg-cyan-500/15" },
  rust:         { text: "text-orange-400", btn: "bg-orange-500 hover:bg-orange-400 shadow-orange-500/30", border: "border-orange-500/30", glow: "bg-orange-500/15" },
  ark:          { text: "text-fuchsia-400", btn: "bg-fuchsia-500 hover:bg-fuchsia-400 shadow-fuchsia-500/30", border: "border-fuchsia-500/30", glow: "bg-fuchsia-500/15" },
};

export const gameTheme = (id: string): GameTheme =>
  GAME_THEME[id] ?? GAME_THEME.minecraft;
