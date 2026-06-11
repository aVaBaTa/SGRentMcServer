"use client";

import { useEffect, useState } from "react";
import { Wifi } from "lucide-react";
import { useI18n } from "@/lib/i18n";

// Endpoint de ping dédié, servi DEPUIS LE CACHE EDGE Cloudflare en HIT (#B) :
// nginx renvoie `/ping.ico` instantanément (pas d'upstream), avec des en-têtes
// `immutable` + extension `.ico` => Cloudflare le garde en cache et répond sans
// jamais toucher l'origine. On mesure donc le pur RTT navigateur<->edge (proximité
// réseau), pas un aller-retour gonflé vers l'origine.
//
// Chemin RELATIF => suit le domaine courant (playrena.vbt-prog.com ou l'ancien
// playrena.vbt-prog.com), même origine (zéro CORS).
// `cache: "no-store"` ne bypasse que le cache LOCAL du navigateur (sinon RTT ≈ 0) ;
// l'edge Cloudflare répond quand même en HIT à chaque requête.
//
// NB: un vrai ping ICMP vers le node de jeu est impossible côté navigateur — c'est
// une approximation de proximité réseau, d'où le préfixe « ~ » à l'affichage.
const TARGET = "/ping.ico";
const WARMUP = 1; // 1re requête écartée (DNS + TLS + 1er remplissage cache)
const SAMPLES = 5; // on garde le min des échantillons « à chaud »

export function PingBadge({ className = "" }: { className?: string }) {
  const { t } = useI18n();
  const [ms, setMs] = useState<number | null>(null);

  useEffect(() => {
    let cancelled = false;
    const ping = async () => {
      const start = performance.now();
      try {
        await fetch(TARGET, { mode: "no-cors", cache: "no-store" });
      } catch { /* opaque (no-cors), le timing reste valable */ }
      return performance.now() - start;
    };
    (async () => {
      for (let i = 0; i < WARMUP; i++) {
        await ping();
        if (cancelled) return;
      }
      const samples: number[] = [];
      for (let i = 0; i < SAMPLES; i++) {
        const d = await ping();
        if (cancelled) return;
        samples.push(d);
      }
      if (!cancelled && samples.length) setMs(Math.round(Math.min(...samples)));
    })();
    return () => { cancelled = true; };
  }, []);

  const color = ms == null ? "text-zinc-500" : ms < 60 ? "text-green-400" : ms < 150 ? "text-yellow-400" : "text-orange-400";
  const quality = ms == null ? "" : ms < 60 ? t.ping.excellent : ms < 150 ? t.ping.good : t.ping.high;

  return (
    <div className={`inline-flex items-center gap-2 rounded-lg border border-zinc-700 bg-zinc-900/60 px-3 py-1.5 text-sm ${className}`}>
      <Wifi className={`w-4 h-4 ${color}`} />
      <span className="text-zinc-300">{t.ping.label}</span>
      <span className={`font-mono font-semibold ${color}`}>{ms == null ? "…" : `~${ms} ms`}</span>
      {quality && <span className="text-zinc-500">· {quality}</span>}
    </div>
  );
}
