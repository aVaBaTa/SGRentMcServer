"use client";

import { Sparkles, Check, Boxes } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { SiteNav, SiteFooter } from "@/components/site-chrome";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export default function ModsPage() {
  const { t } = useI18n();
  return (
    <main className="flex flex-col min-h-screen">
      <SiteNav />

      <section className="relative overflow-hidden flex-1">
        <div className="pointer-events-none absolute -top-40 -left-40 w-[40rem] h-[40rem] rounded-full bg-gradient-to-r from-blue-500 to-indigo-600 opacity-10 blur-3xl" />

        <div className="relative max-w-3xl mx-auto px-6 py-24 flex flex-col gap-6 items-start">
          <a href="/" className="text-sm text-zinc-400 hover:text-zinc-200 transition-colors">{t.mods.back}</a>

          <div className="inline-flex items-center gap-2 bg-indigo-950/60 text-indigo-300 text-sm px-3 py-1 rounded-full border border-indigo-800">
            <Boxes className="w-4 h-4" /> {t.mods.badge}
          </div>

          <h1 className="text-5xl sm:text-6xl font-extrabold tracking-tight leading-[1.05] uppercase">
            {t.mods.title1}<br />
            <span className="bg-gradient-to-r from-blue-400 to-indigo-400 bg-clip-text text-transparent">{t.mods.title2}</span>
          </h1>

          <p className="text-zinc-400 text-lg max-w-2xl">{t.mods.subtitle}</p>

          <ul className="flex flex-col gap-2 mt-2">
            {t.mods.bullets.map((b) => (
              <li key={b} className="flex items-center gap-2 text-zinc-300">
                <Check className="w-5 h-5 text-indigo-400 shrink-0" /> {b}
              </li>
            ))}
          </ul>

          <div className="mt-4 flex flex-col gap-2">
            <a href={`${API}/auth/discord`} className="inline-flex items-center gap-2 bg-indigo-500 hover:bg-indigo-400 text-black font-semibold px-6 py-3 rounded-lg transition-colors">
              <Sparkles className="w-4 h-4" /> {t.mods.cta}
            </a>
            <p className="text-zinc-600 text-sm">{t.mods.note}</p>
          </div>
        </div>
      </section>

      <SiteFooter />
    </main>
  );
}
