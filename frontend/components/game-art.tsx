// Vignettes SVG originales par jeu (motifs iconiques, en `currentColor`) affichées sur
// le dégradé du jeu dans les cartes du hub. Aucune image externe → légal, net, instantané.

const MOTIFS: Record<string, React.ReactNode> = {
  // Minecraft : visage de creeper (ultra reconnaissable).
  minecraft: (
    <g fill="currentColor">
      <rect x="14" y="14" width="36" height="36" rx="2" opacity="0.18" />
      <rect x="20" y="22" width="8" height="8" />
      <rect x="36" y="22" width="8" height="8" />
      <rect x="28" y="30" width="8" height="14" />
      <rect x="22" y="36" width="6" height="8" />
      <rect x="36" y="36" width="6" height="8" />
    </g>
  ),
  // Hytale : gemme/cristal facetté.
  hytale: (
    <>
      <path d="M22 14 H42 L52 26 L32 54 L12 26 Z" fill="currentColor" opacity="0.92" />
      <path d="M12 26 H52 M22 14 L27 26 L32 54 M42 14 L37 26 L32 54" stroke="#0b1220" strokeOpacity="0.25" strokeWidth="1.6" fill="none" />
    </>
  ),
  // Satisfactory : cheminées d'usine.
  satisfactory: (
    <g fill="currentColor">
      <rect x="14" y="30" width="8" height="22" rx="1" />
      <rect x="28" y="20" width="8" height="32" rx="1" />
      <rect x="42" y="34" width="8" height="18" rx="1" />
      <rect x="10" y="50" width="44" height="4" rx="1" />
      <circle cx="32" cy="15" r="2.5" opacity="0.6" />
      <circle cx="18" cy="25" r="2" opacity="0.5" />
    </g>
  ),
  // Rust : baril avec cerclages.
  rust: (
    <>
      <rect x="20" y="12" width="24" height="40" rx="5" fill="currentColor" />
      <g stroke="#1a1208" strokeOpacity="0.28" strokeWidth="2.4">
        <line x1="20" y1="24" x2="44" y2="24" />
        <line x1="20" y1="40" x2="44" y2="40" />
      </g>
    </>
  ),
  // ARK : empreinte de dino (3 doigts).
  ark: (
    <g fill="currentColor">
      <ellipse cx="32" cy="42" rx="9" ry="11" />
      <ellipse cx="21" cy="26" rx="4.5" ry="10" transform="rotate(-22 21 26)" />
      <ellipse cx="32" cy="21" rx="4.5" ry="11" />
      <ellipse cx="43" cy="26" rx="4.5" ry="10" transform="rotate(22 43 26)" />
    </g>
  ),
  // Valheim : casque viking à cornes.
  valheim: (
    <g fill="currentColor">
      <path d="M20 42 a12 12 0 0 1 24 0 v2 h-24 z" />
      <rect x="30" y="32" width="4" height="12" />
      <path d="M21 37 q-11 -2 -13 -14 q9 4 14 9 z" />
      <path d="M43 37 q11 -2 13 -14 q-9 4 -14 9 z" />
    </g>
  ),
  // Palworld : sphère de capture.
  palworld: (
    <g>
      <circle cx="32" cy="32" r="16" fill="currentColor" />
      <rect x="16" y="30" width="32" height="4" fill="#0b1220" fillOpacity="0.3" />
      <circle cx="32" cy="32" r="5.5" fill="#0b1220" fillOpacity="0.28" />
      <circle cx="32" cy="32" r="2.5" fill="currentColor" />
    </g>
  ),
  // Terraria : épée.
  terraria: (
    <g fill="currentColor">
      <polygon points="29,10 35,10 32,5" />
      <rect x="29" y="10" width="6" height="32" rx="1" />
      <rect x="22" y="40" width="20" height="5" rx="1" />
      <rect x="30" y="45" width="4" height="9" rx="1" />
      <circle cx="32" cy="55" r="3" />
    </g>
  ),
  // Bannerlord : bannière médiévale.
  bannerlord: (
    <g fill="currentColor">
      <rect x="19" y="8" width="3.5" height="48" rx="1" />
      <path d="M22.5 12 H46 V36 L40 31 L34 36 L28 31 L22.5 36 Z" />
    </g>
  ),
  // Counter-Strike : réticule.
  cs2: (
    <g>
      <circle cx="32" cy="32" r="15" fill="none" stroke="currentColor" strokeWidth="3" />
      <g stroke="currentColor" strokeWidth="3" strokeLinecap="round">
        <line x1="32" y1="9" x2="32" y2="21" />
        <line x1="32" y1="43" x2="32" y2="55" />
        <line x1="9" y1="32" x2="21" y2="32" />
        <line x1="43" y1="32" x2="55" y2="32" />
      </g>
      <circle cx="32" cy="32" r="2.5" fill="currentColor" />
    </g>
  ),
  // Factorio : engrenage.
  factorio: (
    <g>
      <rect x="20" y="20" width="24" height="24" rx="3" fill="currentColor" />
      <rect x="20" y="20" width="24" height="24" rx="3" fill="currentColor" transform="rotate(45 32 32)" />
      <circle cx="32" cy="32" r="13" fill="currentColor" />
      <circle cx="32" cy="32" r="5" fill="#0b1220" fillOpacity="0.3" />
    </g>
  ),
  // Défaut : manette (jeux à venir sans motif dédié).
  default: (
    <g>
      <rect x="8" y="26" width="48" height="20" rx="10" fill="currentColor" />
      <g fill="#0b1220" fillOpacity="0.28">
        <rect x="16" y="34" width="11" height="4" rx="1" />
        <rect x="19.5" y="30.5" width="4" height="11" rx="1" />
        <circle cx="41" cy="33" r="2.6" />
        <circle cx="47" cy="38" r="2.6" />
        <circle cx="35" cy="38" r="2.6" />
        <circle cx="41" cy="43" r="2.6" />
      </g>
    </g>
  ),
};

export function GameArt({ game, className = "" }: { game: string; className?: string }) {
  return (
    <svg viewBox="0 0 64 64" className={className} aria-hidden="true" fill="none">
      {MOTIFS[game] ?? MOTIFS.default}
    </svg>
  );
}
