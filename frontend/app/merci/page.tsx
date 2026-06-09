"use client";

import { useEffect, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { CheckCircle2, Server } from "lucide-react";

// ⚠️ Remplace par ton libellé de conversion Google Ads (Google Ads → Conversions →
// crée une action "Achat" → tu obtiens un send_to du type "AW-18226964787/AbCdEf...").
const CONVERSION_SEND_TO = "AW-18226964787";

function MerciContent() {
  const params = useSearchParams();
  const plan = params.get("plan") ?? "";
  const server = params.get("server") ?? "";

  useEffect(() => {
    // Déclenche la conversion Google Ads
    const w = window as any;
    if (typeof w.gtag === "function") {
      w.gtag("event", "conversion", {
        send_to: CONVERSION_SEND_TO,
        value: 0,
        currency: "USD",
      });
    }
  }, []);

  return (
    <main className="min-h-screen flex flex-col items-center justify-center text-center px-6 gap-6">
      <CheckCircle2 className="w-16 h-16 text-green-400" />
      <h1 className="text-3xl font-bold">Merci pour ton achat ! 🎉</h1>
      <p className="text-zinc-400 max-w-md">
        Ton paiement a été confirmé{plan && <> et le plan <span className="text-green-400 capitalize">{plan}</span> est activé</>}.
        Les nouvelles ressources sont déjà appliquées à ton serveur.
      </p>
      <div className="flex gap-3 mt-2">
        <Link
          href={server ? `/dashboard/servers/${server}` : "/dashboard"}
          className="flex items-center gap-2 bg-green-500 hover:bg-green-400 text-black font-semibold px-6 py-3 rounded-lg transition-colors"
        >
          <Server className="w-4 h-4" /> Retour à mon serveur
        </Link>
        <Link
          href="/dashboard"
          className="border border-zinc-700 hover:border-zinc-500 text-zinc-300 px-6 py-3 rounded-lg transition-colors"
        >
          Mes serveurs
        </Link>
      </div>
      <p className="text-xs text-zinc-600 mt-6">Playrena — Free Game Hosting</p>
    </main>
  );
}

export default function MerciPage() {
  return (
    <Suspense fallback={null}>
      <MerciContent />
    </Suspense>
  );
}
