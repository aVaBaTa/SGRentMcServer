export type GameStatus = "live" | "soon";

export interface Game {
  id: string;
  name: string;
  href: string; // "#" si pas de page dédiée
  status: GameStatus;
  accent: string; // dégradé Tailwind
}

export const GAMES: Game[] = [
  { id: "minecraft", name: "Minecraft", href: "/games/minecraft", status: "live", accent: "from-green-500 to-emerald-600" },
  { id: "hytale", name: "Hytale", href: "/games/hytale", status: "live", accent: "from-blue-500 to-indigo-600" },
  { id: "rust", name: "Rust", href: "#", status: "soon", accent: "from-red-500 to-orange-600" },
  { id: "satisfactory", name: "Satisfactory", href: "/games/satisfactory", status: "live", accent: "from-orange-500 to-amber-600" },
  { id: "ark", name: "ARK: Survival", href: "#", status: "soon", accent: "from-fuchsia-500 to-purple-600" },
];
