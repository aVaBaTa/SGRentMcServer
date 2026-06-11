"use client";

import { useEffect, useState } from "react";
import { Wifi } from "lucide-react";
import { useI18n } from "@/lib/i18n";

// Cible mesurée : un petit asset statique servi DEPUIS LE CACHE EDGE Cloudflare.
// Le navigateur ne peut pas faire de vrai ICMP ping vers l'IP de jeu ; on mesure
// le RTT HTTP (best-of-N) comme approximation de la proximité réseau.
//
// IMPORTANT (backlog #B) : on NE met PAS de query de cache-busting. Avec un
// cache-buster, Cloudflare contourne son cache edge et fait un aller-retour complet
// Full-Strict jusqu'à l'origine → valeur gonflée et non représentative. Sans
// cache-buster, l'edge répond depuis son cache → on mesure browser↔edge, honnête.
// `cache: "no-store"` ne sert qu'à bypasser le cache LOCAL du navigateur (sinon RTT ≈ 0).
const TARGET = "https://mcserver.vbt-prog.com/favicon.ico";

export function PingBadge({ className = "" }: { className?: string }) {
  const { t } = useI18n();
  const [ms, setMs] = useState<number | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      const samples: number[] = [];
      for (let i = 0; i < 4; i++) {
        const start = performance.now();
        try {
          await fetch(TARGET, { mode: "no-cors", cache: "no-store" });
        } catch { /* opaque (no-cors), le timing reste valable */ }
        if (cancelled) return;
        samples.push(performance.now() - start);
      }
      // On ignore le 1er échantillon (DNS+TLS) si possible, puis on prend le min.
      const considered = samples.length > 1 ? samples.slice(1) : samples;
      if (!cancelled) setMs(Math.round(Math.min(...considered)));
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
