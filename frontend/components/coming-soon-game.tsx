"use client";

import { Sparkles, Check } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { getGame } from "@/lib/games";
import { SiteNav, SiteFooter } from "@/components/site-chrome";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// Page générique « Bientôt disponible » pour un jeu pas encore live, avec CTA
// « être prévenu » (login Discord). Le nom et l'accent viennent de games.ts.
export function ComingSoonGame({ gameId }: { gameId: string }) {
  const { t } = useI18n();
  const g = getGame(gameId);
  const name = g?.name ?? gameId;
  const accent = g?.accent ?? "from-zinc-500 to-zinc-600";

  return (
    <main className="flex flex-col min-h-screen">
      <SiteNav />

      <section className="relative overflow-hidden flex-1">
        <div className={`pointer-events-none absolute -top-40 -left-40 w-[40rem] h-[40rem] rounded-full bg-gradient-to-r ${accent} opacity-10 blur-3xl`} />

        <div className="relative max-w-3xl mx-auto px-6 py-24 flex flex-col gap-6 items-start">
          <a href="/games" className="text-sm text-zinc-400 hover:text-zinc-200 transition-colors">{t.comingSoon.back}</a>

          <div className="inline-flex items-center gap-2 bg-zinc-800/60 text-zinc-300 text-sm px-3 py-1 rounded-full border border-zinc-700">
            <Sparkles className="w-4 h-4" /> {t.comingSoon.badge}
          </div>

          <h1 className="text-5xl sm:text-6xl font-extrabold tracking-tight leading-[1.05] uppercase">
            {t.comingSoon.titlePre}<br />
            <span className={`bg-gradient-to-r ${accent} bg-clip-text text-transparent`}>{name}</span>
          </h1>

          <p className="text-zinc-400 text-lg max-w-2xl">{t.comingSoon.subtitle.replace("{game}", name)}</p>

          <ul className="flex flex-col gap-2 mt-2">
            {t.comingSoon.bullets.map((b) => (
              <li key={b} className="flex items-center gap-2 text-zinc-300">
                <Check className="w-5 h-5 text-zinc-400 shrink-0" /> {b}
              </li>
            ))}
          </ul>

          <div className="mt-4 flex flex-col gap-2">
            <a href={`${API}/auth/discord`} className="inline-flex bg-zinc-100 hover:bg-white text-black font-semibold px-6 py-3 rounded-lg transition-colors">
              {t.comingSoon.cta}
            </a>
            <p className="text-zinc-600 text-sm">{t.comingSoon.note}</p>
          </div>
        </div>
      </section>

      <SiteFooter />
    </main>
  );
}
