"use client";

import { useEffect, useState } from "react";
import { Download, Lock, LogIn, Check, MonitorPlay, AlertTriangle } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { SiteNav, SiteFooter } from "@/components/site-chrome";

// Page de téléchargement de l'application cliente (visionneur SGCoreLink).
// Même verrou que Calradia : l'API décide (session Discord + AdminUsers) ;
// ici on ne fait qu'afficher l'état et pointer vers /api/v1/app/download,
// qui refuse lui-même les comptes non autorisés (la page n'est pas la sécurité).

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

interface AppStatus {
  logged_in: boolean;
  username?: string;
  allowed: boolean;
  release?: { tag: string; name: string; published_at: string; asset_name: string; asset_size: number };
  error?: string;
}

function fmtBytes(n: number) {
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(0)} Ko`;
  return `${(n / (1024 * 1024)).toFixed(1)} Mo`;
}

export default function TelechargerPage() {
  const { t, lang } = useI18n();
  const tt = t.telecharger;
  const [status, setStatus] = useState<AppStatus | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    fetch(`${API}/api/v1/app/status`, { credentials: "include", cache: "no-store" })
      .then((r) => r.json())
      .then((d: AppStatus) => setStatus(d))
      .catch(() => setFailed(true));
  }, []);

  const fill = (s: string) => s.replace("{user}", status?.username ?? "");
  const loginHref = `${API}/auth/discord?next=${encodeURIComponent("/telecharger")}`;
  const publishedAt = status?.release
    ? new Date(status.release.published_at).toLocaleDateString(lang === "fr" ? "fr-CA" : "en-US", { year: "numeric", month: "long", day: "numeric" })
    : "";

  return (
    <main className="flex flex-col min-h-screen">
      <SiteNav />

      <section className="relative overflow-hidden flex-1">
        <div className="pointer-events-none absolute -top-40 -left-40 w-[40rem] h-[40rem] rounded-full bg-gradient-to-r from-emerald-500 to-cyan-600 opacity-10 blur-3xl" />

        <div className="relative max-w-3xl mx-auto px-6 py-24 flex flex-col gap-6 items-start">
          <a href="/" className="text-sm text-zinc-400 hover:text-zinc-200 transition-colors">{tt.back}</a>

          <div className="inline-flex items-center gap-2 bg-emerald-950/60 text-emerald-300 text-sm px-3 py-1 rounded-full border border-emerald-800">
            <Lock className="w-4 h-4" /> {tt.badge}
          </div>

          <h1 className="text-5xl sm:text-6xl font-extrabold tracking-tight leading-[1.05] uppercase">
            {tt.title1}<br />
            <span className="bg-gradient-to-r from-emerald-400 to-cyan-400 bg-clip-text text-transparent">{tt.title2}</span>
          </h1>

          <p className="text-zinc-400 text-lg max-w-2xl">{tt.subtitle}</p>

          {/* État : chargement / non connecté / refusé / autorisé */}
          <div className="mt-4 w-full rounded-xl border border-zinc-800 bg-zinc-900/60 p-6 flex flex-col gap-4">
            {failed ? (
              <p className="flex items-center gap-2 text-amber-300"><AlertTriangle className="w-5 h-5 shrink-0" /> {tt.unavailable} API</p>
            ) : !status ? (
              <p className="text-zinc-400">{tt.loading}</p>
            ) : !status.logged_in ? (
              <>
                <p className="text-zinc-300">{tt.loginNote}</p>
                <a href={loginHref} className="inline-flex w-fit items-center gap-2 bg-indigo-500 hover:bg-indigo-400 text-black font-semibold px-6 py-3 rounded-lg transition-colors">
                  <LogIn className="w-4 h-4" /> {tt.loginCta}
                </a>
              </>
            ) : !status.allowed ? (
              <>
                <p className="flex items-center gap-2 font-semibold text-white"><Lock className="w-5 h-5 text-amber-400" /> {tt.deniedTitle}</p>
                <p className="text-zinc-400">{fill(tt.deniedText)}</p>
              </>
            ) : (
              <>
                <p className="flex items-center gap-2 font-semibold text-white"><Check className="w-5 h-5 text-emerald-400" /> {tt.readyTitle}</p>
                <p className="text-zinc-400">{fill(tt.readyText)}</p>
                {status.release ? (
                  <>
                    <dl className="grid grid-cols-3 gap-3 text-sm max-w-md">
                      <div><dt className="text-zinc-500">{tt.version}</dt><dd className="text-zinc-200 font-mono">{status.release.tag}</dd></div>
                      <div><dt className="text-zinc-500">{tt.size}</dt><dd className="text-zinc-200 font-mono">{fmtBytes(status.release.asset_size)}</dd></div>
                      <div><dt className="text-zinc-500">{tt.published}</dt><dd className="text-zinc-200">{publishedAt}</dd></div>
                    </dl>
                    <a href={`${API}/api/v1/app/download`} className="inline-flex w-fit items-center gap-2 bg-emerald-500 hover:bg-emerald-400 text-black font-semibold px-6 py-3 rounded-lg transition-colors">
                      <Download className="w-4 h-4" /> {tt.downloadCta}
                      <span className="font-mono text-emerald-900/80 text-sm">{status.release.asset_name}</span>
                    </a>
                  </>
                ) : (
                  <p className="flex items-start gap-2 text-amber-300"><AlertTriangle className="w-5 h-5 shrink-0 mt-0.5" /> {tt.unavailable} {status.error}</p>
                )}
              </>
            )}
          </div>

          {/* Installation */}
          <h2 className="mt-6 flex items-center gap-2 text-xl font-bold text-white"><MonitorPlay className="w-5 h-5 text-emerald-400" /> {tt.stepsTitle}</h2>
          <ol className="flex flex-col gap-3">
            {tt.steps.map((s, i) => (
              <li key={i} className="flex items-start gap-3 text-zinc-300">
                <span className="mt-0.5 grid h-6 w-6 shrink-0 place-items-center rounded-full bg-emerald-500/15 text-emerald-300 text-sm font-semibold">{i + 1}</span>
                <span>{s}</span>
              </li>
            ))}
          </ol>
          <p className="text-zinc-500 text-sm">{tt.note}</p>
        </div>
      </section>

      <SiteFooter />
    </main>
  );
}
