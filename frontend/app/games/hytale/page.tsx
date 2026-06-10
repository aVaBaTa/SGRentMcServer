"use client";

import { Check, Sparkles } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { SiteNav, SiteFooter } from "@/components/site-chrome";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export default function HytaleHosting() {
  const { t } = useI18n();
  return (
    <main className="flex flex-col min-h-screen">
      <SiteNav />

      <section className="relative overflow-hidden flex-1">
        <div className="pointer-events-none absolute -top-40 -left-40 w-[40rem] h-[40rem] rounded-full bg-blue-500/15 blur-3xl" />
        <div className="pointer-events-none absolute top-20 -right-40 w-[40rem] h-[40rem] rounded-full bg-indigo-500/15 blur-3xl" />

        <div className="relative max-w-3xl mx-auto px-6 py-24 flex flex-col gap-6 items-start">
          <a href="/games" className="text-sm text-zinc-400 hover:text-zinc-200 transition-colors">{t.hytale.back}</a>

          <div className="inline-flex items-center gap-2 bg-blue-950/60 text-blue-300 text-sm px-3 py-1 rounded-full border border-blue-800">
            <Sparkles className="w-4 h-4" /> {t.hytale.badge}
          </div>

          <h1 className="text-5xl sm:text-6xl font-extrabold tracking-tight leading-[1.05] uppercase">
            {t.hytale.title1}<br />
            <span className="bg-gradient-to-r from-blue-400 to-indigo-400 bg-clip-text text-transparent">{t.hytale.title2}</span>
          </h1>

          <p className="text-zinc-400 text-lg max-w-2xl">{t.hytale.subtitle}</p>

          <ul className="flex flex-col gap-2 mt-2">
            {t.hytale.bullets.map((b) => (
              <li key={b} className="flex items-center gap-2 text-zinc-300">
                <Check className="w-5 h-5 text-blue-400 shrink-0" /> {b}
              </li>
            ))}
          </ul>

          <div className="mt-4 flex flex-col gap-2">
            <a href={`${API}/auth/discord`} className="inline-flex bg-blue-500 hover:bg-blue-400 text-black font-semibold px-6 py-3 rounded-lg transition-colors">
              {t.hytale.cta}
            </a>
            <p className="text-zinc-600 text-sm">{t.hytale.note}</p>
          </div>
        </div>
      </section>

      <SiteFooter />
    </main>
  );
}
