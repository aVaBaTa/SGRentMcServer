"use client";

import { Server, Clock } from "lucide-react";
import { useI18n, type Lang } from "@/lib/i18n";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

function LanguageSwitcher() {
  const { lang, setLang } = useI18n();
  const Btn = ({ l, label }: { l: Lang; label: string }) => (
    <button
      onClick={() => setLang(l)}
      className={`px-2.5 py-1 rounded-md text-xs font-semibold transition-colors ${
        lang === l ? "bg-zinc-700 text-white" : "text-zinc-400 hover:text-zinc-200"
      }`}
    >
      {label}
    </button>
  );
  return (
    <div className="flex items-center gap-0.5 rounded-lg bg-zinc-900 border border-zinc-800 p-0.5">
      <Btn l="fr" label="FR · CAD" />
      <Btn l="en" label="EN · USD" />
    </div>
  );
}

export function SiteNav() {
  const { t } = useI18n();
  return (
    <nav className="border-b border-zinc-800 px-6 py-4 flex items-center justify-between sticky top-0 z-30 bg-zinc-950/80 backdrop-blur">
      <a href="/games" className="flex items-center gap-2 font-bold text-lg">
        <Server className="w-5 h-5 text-green-400" />
        Playrena
      </a>
      <div className="flex items-center gap-4 sm:gap-6 text-sm text-zinc-400">
        <a href="/games" className="hidden sm:inline hover:text-zinc-100 transition-colors">{t.navGames}</a>
        <a href="/blog" className="hidden sm:inline hover:text-zinc-100 transition-colors">{t.nav.blog}</a>
        <LanguageSwitcher />
        <a href={`${API}/auth/discord`} className="bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-lg transition-colors font-medium">
          {t.nav.login}
        </a>
      </div>
    </nav>
  );
}

export function SiteFooter() {
  const { t } = useI18n();
  return (
    <footer className="border-t border-zinc-800 px-6 py-6 flex flex-col sm:flex-row items-center justify-between gap-3 text-sm text-zinc-600 mt-auto">
      <span>© {new Date().getFullYear()} Playrena — {t.footer}</span>
      <div className="flex items-center gap-4">
        <a href="/blog" className="hover:text-zinc-300 transition-colors flex items-center gap-1"><Clock className="w-3.5 h-3.5" /> {t.nav.blog}</a>
      </div>
    </footer>
  );
}
