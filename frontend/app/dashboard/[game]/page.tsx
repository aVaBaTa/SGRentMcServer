"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { Server, Plus, Play, Square, RefreshCw, KeyRound, ArrowLeft, Copy, Check, Cpu, MemoryStick, Settings, Sparkles } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { getGame, gameTheme } from "@/lib/games";
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

const SERVER_HOST = process.env.NEXT_PUBLIC_SERVER_HOST ?? "24.157.140.226";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
const VERSIONS = ["LATEST", "1.21.4", "1.21.1", "1.20.6", "1.20.4", "1.20.1", "1.19.4", "1.18.2", "1.16.5", "1.12.2", "1.8.9"];

const STATUS_META: Record<string, { dot: string; text: string; pulse?: boolean }> = {
  running:       { dot: "bg-green-500",  text: "text-green-400" },
  creating:      { dot: "bg-yellow-500", text: "text-yellow-400", pulse: true },
  auth_required: { dot: "bg-amber-500",  text: "text-amber-400", pulse: true },
  stopped:       { dot: "bg-zinc-500",   text: "text-zinc-400" },
  error:         { dot: "bg-red-500",    text: "text-red-400" },
};

export default function GameDashboard() {
  const router = useRouter();
  const { game } = useParams<{ game: string }>();
  const { t } = useI18n();

  const def = getGame(game);
  const theme = gameTheme(game);
  const supportsVersion = game === "minecraft";

  const [servers, setServers] = useState<GameServer[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState("");
  const [newVersion, setNewVersion] = useState("LATEST");
  const [newLoader, setNewLoader] = useState("paper"); // Minecraft : paper | fabric | forge
  const [error, setError] = useState("");
  const [copiedId, setCopiedId] = useState("");

  const metaOf = (s: string) => STATUS_META[s] ?? { dot: "bg-zinc-500", text: "text-zinc-400" };
  const labelOf = (s: string) => t.dash.status[s] ?? s;
  const addrOf = (s: GameServer) => s.game === "minecraft" ? `${s.subdomain}.servers.vbt-prog.com` : `${SERVER_HOST}:${s.port}`;
  const kindOf = (s: GameServer) => s.modpack ? "Modpack" : s.loader === "fabric" ? "Fabric" : s.loader === "forge" ? "Forge" : s.game === "minecraft" ? "Paper" : "";
  async function copyAddr(s: GameServer) {
    try { await navigator.clipboard.writeText(addrOf(s)); setCopiedId(s.id); setTimeout(() => setCopiedId(""), 1500); } catch { /* clipboard indispo */ }
  }

  // Jeu inconnu ou pas encore disponible → retour au hub.
  useEffect(() => {
    if (!def || def.status !== "live") router.replace("/dashboard");
  }, [def, router]);

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

  async function createServer() {
    if (!newName.trim()) { setError(t.dash.errName); return; }
    setCreating(true);
    setError("");
    try {
      const res = await fetch(`${API}/api/v1/servers`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: newName, plan: "free", game, version: supportsVersion ? newVersion : "LATEST", loader: game === "minecraft" ? newLoader : "paper" }),
      });
      if (res.ok) { setNewName(""); fetchServers(); }
      else { setError((await res.text()) || `Erreur ${res.status}`); }
    } catch { setError(t.dash.errConn); } finally { setCreating(false); }
  }

  async function serverAction(id: string, action: "start" | "stop" | "restart") {
    await fetch(`${API}/api/v1/servers/${id}/${action}`, { method: "POST", credentials: "include" });
    fetchServers();
  }

  if (!def || def.status !== "live") return null;

  const mine = servers.filter((s) => s.game === game);

  const onlineCount = mine.filter((s) => s.status === "running").length;

  return (
    <div className="min-h-screen flex flex-col">
      {/* Nav */}
      <nav className="sticky top-0 z-10 border-b border-zinc-800/80 bg-zinc-950/70 backdrop-blur px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-2.5">
          <Link href="/dashboard" className="flex items-center gap-1.5 text-sm text-zinc-400 hover:text-zinc-100 transition-colors">
            <ArrowLeft className="w-4 h-4" /> {t.dash.backToHub}
          </Link>
          <span className="text-zinc-700">/</span>
          <span className={`flex items-center gap-2 font-bold ${theme.text}`}>
            <Server className="w-5 h-5" /> {def.name}
          </span>
        </div>
        <LanguageSwitcher />
      </nav>

      <div className="flex-1 w-full p-6 max-w-5xl mx-auto">
        {/* Hero de création */}
        <section className={`relative overflow-hidden mb-10 rounded-3xl border ${theme.border} bg-gradient-to-br from-zinc-900 via-zinc-900 to-zinc-950 px-6 sm:px-10 py-10`}>
          <div className={`pointer-events-none absolute -top-40 -right-32 w-[30rem] h-[30rem] rounded-full blur-3xl ${theme.glow}`} />
          <div className={`pointer-events-none absolute -bottom-44 -left-28 w-96 h-96 rounded-full blur-3xl ${theme.glow} opacity-60`} />
          <div className="relative flex flex-col items-center text-center gap-6">
            <div className={`flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br ${def.accent} shadow-xl ring-1 ring-white/10`}>
              <GameArt game={game} className="w-9 h-9 text-white drop-shadow" />
            </div>
            <div className="space-y-2">
              <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">{def.name}</h1>
              <p className="text-zinc-400 max-w-md mx-auto">
                {t.dash.leadPre}{def.name}{t.dash.leadMid}
                <span className={`font-semibold ${theme.text}`}>{t.dash.leadFree}</span>{t.dash.leadPost}
              </p>
            </div>

            <form onSubmit={(e) => { e.preventDefault(); createServer(); }} className="w-full max-w-2xl flex flex-col gap-3">
              <div className="flex flex-col sm:flex-row gap-3">
                <input
                  type="text"
                  placeholder={t.dash.namePlaceholder}
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  className="flex-1 bg-zinc-950/70 border border-zinc-700 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-zinc-400 focus:ring-2 focus:ring-white/5 transition-colors"
                />
                {supportsVersion && (
                  <select
                    value={newVersion}
                    onChange={(e) => setNewVersion(e.target.value)}
                    className="select-dark cursor-pointer bg-zinc-950/70 border border-zinc-700 rounded-xl pl-4 py-3 text-sm text-zinc-100 focus:outline-none focus:border-zinc-400"
                  >
                    {VERSIONS.map((v) => <option key={v} value={v}>{v === "LATEST" ? t.dash.versionLatest : v}</option>)}
                  </select>
                )}
                {game === "minecraft" && (
                  <select
                    value={newLoader}
                    onChange={(e) => setNewLoader(e.target.value)}
                    title={t.dash.loaderLabel}
                    className="select-dark cursor-pointer bg-zinc-950/70 border border-zinc-700 rounded-xl pl-4 py-3 text-sm text-zinc-100 focus:outline-none focus:border-zinc-400"
                  >
                    <option value="paper">{t.dash.loaderPaper}</option>
                    <option value="fabric">{t.dash.loaderFabric}</option>
                    <option value="forge">{t.dash.loaderForge}</option>
                  </select>
                )}
              </div>
              {game === "minecraft" && (
                <p className="text-xs text-zinc-500 max-w-md mx-auto">{t.dash.loaderHint}</p>
              )}
              <button
                type="submit"
                disabled={creating}
                className={`mx-auto flex items-center justify-center gap-2 disabled:opacity-50 text-black font-semibold text-base px-8 py-3.5 rounded-xl shadow-lg transition-all hover:scale-[1.02] ${theme.btn}`}
              >
                <Plus className="w-5 h-5" />
                {creating ? t.dash.creating : t.dash.createBtn}
              </button>
            </form>

            {/* Note d'autorisation Hytale (OAuth au 1er démarrage) */}
            {def.id === "hytale" && (
              <div className="flex items-start gap-2 text-xs text-zinc-400 max-w-md rounded-lg bg-blue-500/5 border border-blue-500/20 px-3 py-2">
                <KeyRound className="w-4 h-4 shrink-0 mt-0.5 text-blue-400" />
                <span>{t.hytale.authNote}</span>
              </div>
            )}
          </div>
        </section>

        {error && (
          <div className="mb-4 rounded-xl border border-red-900 bg-red-950/40 px-4 py-3 text-sm text-red-300">{error}</div>
        )}

        {/* En-tête liste + résumé */}
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">{t.dash.gameServersTitle.replace("{game}", def.name)}</h2>
          {mine.length > 0 && (
            <span className="text-xs text-zinc-500">
              {mine.length} · <span className="text-green-400">{onlineCount} {t.dash.status.running}</span>
            </span>
          )}
        </div>

        {loading ? (
          <div className="text-zinc-400 text-center py-12">{t.dash.loading}</div>
        ) : mine.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-zinc-700 p-12 text-center">
            <div className="mx-auto mb-4 flex items-center justify-center w-14 h-14 rounded-2xl bg-zinc-800/60">
              <Sparkles className="w-7 h-7 text-zinc-500" />
            </div>
            <p className="text-zinc-400 max-w-sm mx-auto">{t.dash.empty}</p>
          </div>
        ) : (
          <div className="grid sm:grid-cols-2 gap-4">
            {mine.map((s) => {
              const m = metaOf(s.status);
              const kind = kindOf(s);
              return (
                <div key={s.id} className="group relative overflow-hidden rounded-2xl border border-zinc-800 bg-zinc-900/70 transition-all hover:border-zinc-600 hover:shadow-xl hover:shadow-black/30">
                  <div className={`h-1 w-full bg-gradient-to-r ${def.accent}`} />
                  <div className="p-5">
                    <div className="flex items-start justify-between gap-3 mb-3">
                      <div className="min-w-0">
                        <Link href={`/dashboard/servers/${s.id}`} className="font-semibold text-lg truncate block hover:text-white transition-colors">{s.name}</Link>
                        {kind && <span className="inline-block mt-1 text-[10px] font-medium uppercase tracking-wide text-zinc-400 bg-zinc-800 rounded px-1.5 py-0.5">{kind}</span>}
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
