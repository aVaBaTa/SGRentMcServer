"use client";

import { ArrowRight, Gamepad2 } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { GAMES } from "@/lib/games";
import { SiteNav, SiteFooter } from "@/components/site-chrome";

export default function GamesHub() {
  const { t } = useI18n();
  return (
    <main className="flex flex-col min-h-screen">
      <SiteNav />

      <section className="relative overflow-hidden">
        <div className="pointer-events-none absolute -top-40 left-1/2 -translate-x-1/2 w-[50rem] h-[40rem] rounded-full bg-green-500/10 blur-3xl" />
        <div className="relative max-w-5xl mx-auto px-6 pt-20 pb-10 text-center">
          <div className="inline-flex items-center gap-2 text-green-400 text-sm font-semibold mb-4">
            <Gamepad2 className="w-5 h-5" /> Playrena
          </div>
          <h1 className="text-4xl sm:text-5xl font-extrabold tracking-tight">{t.hub.title}</h1>
          <p className="text-zinc-400 text-lg mt-4 max-w-2xl mx-auto">{t.hub.subtitle}</p>
        </div>
      </section>

      <section className="px-6 pb-32 max-w-5xl mx-auto w-full grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {GAMES.map((g) => {
          const isLive = g.status === "live";
          const clickable = g.href !== "#";
          const Card = (
            <div className={`group relative h-full rounded-2xl border bg-zinc-900 overflow-hidden transition-colors ${
              clickable ? "border-zinc-800 hover:border-green-600" : "border-zinc-800 opacity-80"
            }`}>
              {/* Bandeau coloré */}
              <div className={`h-28 bg-gradient-to-br ${g.accent} relative`}>
                <div className="absolute inset-0 bg-black/20" />
                <span className="absolute top-3 right-3 text-[10px] font-bold uppercase tracking-wide px-2 py-0.5 rounded-full backdrop-blur bg-black/40 text-white">
                  {g.restricted ? t.hub.restricted : isLive ? t.hub.live : t.hub.soon}
                </span>
                <span className="absolute bottom-3 left-4 text-2xl font-extrabold text-white drop-shadow">{g.name}</span>
              </div>
              <div className="p-5 flex flex-col gap-4">
                <p className="text-sm text-zinc-400 min-h-[2.5rem]">{t.gameTag[g.id] ?? ""}</p>
                <span className={`inline-flex items-center gap-1.5 text-sm font-medium ${clickable ? "text-green-400" : "text-zinc-500"}`}>
                  {isLive ? t.hub.view : t.hub.notify}
                  {clickable && <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />}
                </span>
              </div>
            </div>
          );
          return clickable ? (
            <a key={g.id} href={g.href} className="block h-full">{Card}</a>
          ) : (
            <div key={g.id} className="h-full cursor-default">{Card}</div>
          );
        })}
      </section>

      <SiteFooter />
    </main>
  );
}
