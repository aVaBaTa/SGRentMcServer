"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Server, Play, Square, RefreshCw, ChevronRight, ChevronDown, Copy, Check, Cpu, MemoryStick, Settings } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { GAMES, getGame } from "@/lib/games";
import { LanguageSwitcher } from "@/components/site-chrome";
import { GameArt } from "@/components/game-art";

interface GameServer {
  id: string;
  name: string;
  subdomain: string;
  game: string;
  plan: string;
  status: string;
  loader?: string;
  modpack?: string;
  ram_mb: number;
  cpu_cores: number;
  port: number;
  created_at: string;
}

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
const SERVER_HOST = process.env.NEXT_PUBLIC_SERVER_HOST ?? "24.157.140.226";

const STATUS_META: Record<string, { dot: string; text: string; pulse?: boolean }> = {
  running:       { dot: "bg-green-500",  text: "text-green-400" },
  creating:      { dot: "bg-yellow-500", text: "text-yellow-400", pulse: true },
  auth_required: { dot: "bg-amber-500",  text: "text-amber-400", pulse: true },
  stopped:       { dot: "bg-zinc-500",   text: "text-zinc-400" },
  error:         { dot: "bg-red-500",    text: "text-red-400" },
};

export default function DashboardHub() {
  const router = useRouter();
  const { t } = useI18n();
  const [servers, setServers] = useState<GameServer[]>([]);
  const [loading, setLoading] = useState(true);
  const [copiedId, setCopiedId] = useState("");
  const [showAllGames, setShowAllGames] = useState(false);
  const GAMES_PREVIEW = 10;

  const metaOf = (s: string) => STATUS_META[s] ?? { dot: "bg-zinc-500", text: "text-zinc-400" };
  const labelOf = (s: string) => t.dash.status[s] ?? s;
  const gameNameOf = (id: string) => getGame(id)?.name ?? id;
  const addrOf = (s: GameServer) => s.game === "minecraft" ? `${s.subdomain}.servers.vbt-prog.com` : `${SERVER_HOST}:${s.port}`;
  const kindOf = (s: GameServer) => s.modpack ? "Modpack" : s.loader === "fabric" ? "Fabric" : s.loader === "forge" ? "Forge" : s.game === "minecraft" ? "Paper" : "";
  async function copyAddr(s: GameServer) {
    try { await navigator.clipboard.writeText(addrOf(s)); setCopiedId(s.id); setTimeout(() => setCopiedId(""), 1500); } catch { /* clipboard indispo */ }
  }

  useEffect(() => {
    fetchServers();
    const interval = setInterval(fetchServers, 3000);
    return () => clearInterval(interval);
  }, []);

  async function fetchServers() {
    try {
      const res = await fetch(`${API}/api/v1/servers`, { credentials: "include" });
      if (res.status === 401) { router.push("/"); return; }
      const data = await res.json();
      setServers(Array.isArray(data) ? data : []);
    } catch { /* garde l'état précédent */ } finally { setLoading(false); }
  }

  async function serverAction(id: string, action: "start" | "stop" | "restart") {
    await fetch(`${API}/api/v1/servers/${id}/${action}`, { method: "POST", credentials: "include" });
    fetchServers();
  }

  async function logout() {
    try { await fetch(`${API}/auth/logout`, { method: "POST", credentials: "include" }); } catch { /* ignore */ }
    router.push("/");
  }

  const onlineCount = servers.filter((s) => s.status === "running").length;

  return (
    <div className="min-h-screen flex flex-col">
      {/* Nav */}
      <nav className="sticky top-0 z-10 border-b border-zinc-800/80 bg-zinc-950/70 backdrop-blur px-6 py-4 flex items-center justify-between">
        <Link href="/dashboard" className="flex items-center gap-2 font-bold text-lg">
          <Server className="w-5 h-5 text-green-400" />
          Playrena
        </Link>
        {/* Accès aux pages publiques (blog, jeux…) en restant connecté — la session persiste. */}
        <div className="flex items-center gap-4 sm:gap-5 text-sm text-zinc-400">
          <Link href="/games" className="hidden sm:inline hover:text-zinc-100 transition-colors">{t.navGames}</Link>
          <Link href="/blog" className="hover:text-zinc-100 transition-colors">{t.nav.blog}</Link>
          <Link href="/mods" className="hidden sm:inline hover:text-zinc-100 transition-colors">{t.nav.mods}</Link>
          <LanguageSwitcher />
          <button onClick={logout} className="text-zinc-400 hover:text-red-400 transition-colors">{t.nav.logout}</button>
        </div>
      </nav>

      <div className="flex-1 p-6 max-w-5xl mx-auto w-full">
        {/* Choix du jeu */}
        <h1 className="text-2xl sm:text-3xl font-bold mb-1 text-center tracking-tight">{t.dash.chooseGame}</h1>
        <p className="text-zinc-400 text-center mb-8">{t.dash.hubLead}</p>

        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4 mb-4">
          {(showAllGames ? GAMES : GAMES.slice(0, GAMES_PREVIEW)).map((g) => {
            const live = g.status === "live";
            const card = (
              <div className={`group relative h-full overflow-hidden rounded-2xl border bg-zinc-900/70 flex flex-col transition-all ${live ? "border-zinc-800 hover:border-zinc-600 hover:shadow-xl hover:shadow-black/30 cursor-pointer" : "border-zinc-900 opacity-60"}`}>
                <div className={`relative h-20 bg-gradient-to-br ${g.accent} flex items-center justify-center ${!live ? "grayscale-[35%]" : ""}`}>
                  <GameArt game={g.id} className={`w-12 h-12 text-white drop-shadow ${live ? "transition-transform group-hover:scale-110" : ""}`} />
                  <div className="pointer-events-none absolute inset-0 bg-gradient-to-t from-zinc-900/40 to-transparent" />
                </div>
                <div className="p-4 flex flex-col gap-1 flex-1">
                  <div className="font-semibold">{g.name}</div>
                  <div className="mt-auto flex items-center justify-between text-xs">
                    <span className={g.restricted ? "text-rose-300" : live ? "text-green-400" : "text-zinc-500"}>{g.restricted ? t.hub.restricted : live ? t.hub.live : t.hub.soon}</span>
                    {live && <ChevronRight className="w-4 h-4 text-zinc-500 group-hover:text-zinc-300 group-hover:translate-x-0.5 transition-all" />}
                  </div>
                </div>
              </div>
            );
            return live
              ? <Link key={g.id} href={`/dashboard/${g.id}`} className="block h-full">{card}</Link>
              : <div key={g.id} className="h-full" aria-disabled>{card}</div>;
          })}
        </div>

        {GAMES.length > GAMES_PREVIEW && (
          <div className="flex justify-center mb-12">
            <button
              onClick={() => setShowAllGames((v) => !v)}
              className="flex items-center gap-1.5 text-sm text-zinc-400 hover:text-zinc-100 bg-zinc-900/70 hover:bg-zinc-800 border border-zinc-800 hover:border-zinc-600 rounded-full px-4 py-2 transition-colors"
            >
              {showAllGames ? t.dash.showLess : t.dash.showMore.replace("{n}", String(GAMES.length - GAMES_PREVIEW))}
              <ChevronDown className={`w-4 h-4 transition-transform ${showAllGames ? "rotate-180" : ""}`} />
            </button>
          </div>
        )}

        {/* Tous les serveurs */}
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">{t.dash.allServers}</h2>
          {servers.length > 0 && (
            <span className="text-xs text-zinc-500">
              {servers.length} · <span className="text-green-400">{onlineCount} {t.dash.status.running}</span>
            </span>
          )}
        </div>

        {loading ? (
          <div className="text-zinc-400 text-center py-12">{t.dash.loading}</div>
        ) : servers.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-zinc-700 p-12 text-center">
            <div className="mx-auto mb-4 flex items-center justify-center w-14 h-14 rounded-2xl bg-zinc-800/60">
              <Server className="w-7 h-7 text-zinc-500" />
            </div>
            <p className="text-zinc-400 max-w-sm mx-auto">{t.dash.empty}</p>
          </div>
        ) : (
          <div className="grid sm:grid-cols-2 gap-4">
            {servers.map((s) => {
              const m = metaOf(s.status);
              const kind = kindOf(s);
              const accent = getGame(s.game)?.accent ?? "from-green-500 to-emerald-600";
              return (
                <div key={s.id} className="group relative overflow-hidden rounded-2xl border border-zinc-800 bg-zinc-900/70 transition-all hover:border-zinc-600 hover:shadow-xl hover:shadow-black/30">
                  <div className={`h-1 w-full bg-gradient-to-r ${accent}`} />
                  <div className="p-5">
                    <div className="flex items-start justify-between gap-3 mb-3">
                      <div className="min-w-0">
                        <Link href={`/dashboard/servers/${s.id}`} className="font-semibold text-lg truncate block hover:text-white transition-colors">{s.name}</Link>
                        <div className="flex flex-wrap gap-1.5 mt-1">
                          <span className="text-[10px] font-medium uppercase tracking-wide text-zinc-300 bg-zinc-800 rounded px-1.5 py-0.5">{gameNameOf(s.game)}</span>
                          {kind && <span className="text-[10px] font-medium uppercase tracking-wide text-zinc-400 bg-zinc-800/70 rounded px-1.5 py-0.5">{kind}</span>}
                        </div>
                      </div>
                      <span className={`shrink-0 flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-full bg-zinc-800/80 ${m.text}`}>
                        <span className={`w-2 h-2 rounded-full ${m.dot} ${m.pulse ? "animate-pulse" : ""}`} />
                        {labelOf(s.status)}
                      </span>
                    </div>

                    <button onClick={() => copyAddr(s)} title={addrOf(s)} className="w-full flex items-center justify-between gap-2 rounded-lg bg-black/30 border border-zinc-800 px-3 py-2 text-xs font-mono text-zinc-300 hover:border-zinc-600 transition-colors">
                      <span className="truncate">{addrOf(s)}</span>
                      {copiedId === s.id ? <Check className="w-3.5 h-3.5 text-green-400 shrink-0" /> : <Copy className="w-3.5 h-3.5 text-zinc-500 shrink-0" />}
                    </button>

                    <div className="flex flex-wrap gap-1.5 mt-3">
                      <span className="flex items-center gap-1 text-[11px] text-zinc-400 bg-zinc-800/60 rounded-md px-2 py-1"><MemoryStick className="w-3.5 h-3.5" /> {s.ram_mb / 1024} GB</span>
                      <span className="flex items-center gap-1 text-[11px] text-zinc-400 bg-zinc-800/60 rounded-md px-2 py-1"><Cpu className="w-3.5 h-3.5" /> {s.cpu_cores} vCPU</span>
                      <span className="flex items-center gap-1 text-[11px] text-zinc-400 bg-zinc-800/60 rounded-md px-2 py-1 capitalize">{t.dash.plan} {s.plan}</span>
                    </div>

                    <div className="flex items-center gap-2 mt-4">
                      {s.status === "stopped" && (
                        <button onClick={() => serverAction(s.id, "start")} className="flex items-center gap-1.5 text-sm bg-green-500 hover:bg-green-400 text-black font-medium px-3 py-1.5 rounded-lg transition-colors" title={t.dash.start}>
                          <Play className="w-4 h-4" /> {t.dash.start}
                        </button>
                      )}
                      {s.status === "running" && (
                        <>
                          <button onClick={() => serverAction(s.id, "restart")} className="p-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 transition-colors" title={t.dash.restart}>
                            <RefreshCw className="w-4 h-4" />
                          </button>
                          <button onClick={() => serverAction(s.id, "stop")} className="p-2 rounded-lg bg-zinc-800 hover:bg-red-900/60 text-red-400 transition-colors" title={t.dash.stop}>
                            <Square className="w-4 h-4" />
                          </button>
                        </>
                      )}
                      <Link href={`/dashboard/servers/${s.id}`} className="ml-auto flex items-center gap-1.5 text-sm text-zinc-300 hover:text-white bg-zinc-800 hover:bg-zinc-700 px-3 py-1.5 rounded-lg transition-colors">
                        <Settings className="w-4 h-4" /> {t.dash.manage}
                      </Link>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
