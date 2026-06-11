"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { Server, Plus, Play, Square, RefreshCw, KeyRound } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { getGame, gameTheme } from "@/lib/games";
import { LanguageSwitcher } from "@/components/site-chrome";

interface GameServer {
  id: string;
  name: string;
  subdomain: string;
  game: string;
  plan: string;
  status: string;
  ram_mb: number;
  cpu_cores: number;
  port: number;
  created_at: string;
}

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
  const [error, setError] = useState("");

  const metaOf = (s: string) => STATUS_META[s] ?? { dot: "bg-zinc-500", text: "text-zinc-400" };
  const labelOf = (s: string) => t.dash.status[s] ?? s;

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
        body: JSON.stringify({ name: newName, plan: "free", game, version: supportsVersion ? newVersion : "LATEST" }),
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

  return (
    <div className="min-h-screen flex flex-col">
      <nav className="border-b border-zinc-800 px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link href="/dashboard" className="text-sm text-zinc-400 hover:text-zinc-200 transition-colors">{t.dash.backToHub}</Link>
          <span className="text-zinc-600">/</span>
          <div className={`flex items-center gap-2 font-bold ${theme.text}`}>
            <Server className="w-5 h-5" /> {def.name}
          </div>
        </div>
        <LanguageSwitcher />
      </nav>

      <div className="flex-1 p-6 max-w-5xl mx-auto w-full">
        {/* Bloc de création — adapté au jeu */}
        <div className={`relative overflow-hidden mb-10 flex flex-col items-center text-center gap-5 rounded-2xl border ${theme.border} bg-gradient-to-b from-zinc-900 to-zinc-950 px-6 py-10`}>
          <div className={`pointer-events-none absolute -top-32 -right-24 w-96 h-96 rounded-full blur-3xl ${theme.glow}`} />
          <p className="relative text-zinc-300 max-w-md">
            {t.dash.leadPre}{def.name}{t.dash.leadMid}
            <span className={`font-medium ${theme.text}`}>{t.dash.leadFree}</span>
            {t.dash.leadPost}
          </p>
          <form
            onSubmit={(e) => { e.preventDefault(); createServer(); }}
            className="relative flex flex-col sm:flex-row items-stretch sm:items-center justify-center gap-3 w-full max-w-2xl"
          >
            <input
              type="text"
              placeholder={t.dash.namePlaceholder}
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              className="flex-1 bg-zinc-900 border border-zinc-700 rounded-lg px-4 py-3 text-sm focus:outline-none focus:border-zinc-400"
            />
            {supportsVersion && (
              <select
                value={newVersion}
                onChange={(e) => setNewVersion(e.target.value)}
                className="select-dark cursor-pointer bg-zinc-900 border border-zinc-700 rounded-lg pl-4 py-3 text-sm text-zinc-100 focus:outline-none focus:border-zinc-400"
              >
                {VERSIONS.map((v) => <option key={v} value={v}>{v === "LATEST" ? t.dash.versionLatest : v}</option>)}
              </select>
            )}
          </form>
          <button
            type="button"
            onClick={createServer}
            disabled={creating}
            className={`relative flex items-center justify-center gap-2 disabled:opacity-50 text-black font-semibold text-base px-8 py-4 rounded-xl shadow-lg transition-all hover:scale-[1.02] ${theme.btn}`}
          >
            <Plus className="w-5 h-5" />
            {creating ? t.dash.creating : t.dash.createBtn}
          </button>

          {/* Note d'autorisation Hytale (OAuth au 1er démarrage) */}
          {def.id === "hytale" && (
            <div className="relative flex items-start gap-2 text-xs text-zinc-400 max-w-md">
              <KeyRound className="w-4 h-4 shrink-0 mt-0.5 text-blue-400" />
              <span>{t.hytale.authNote}</span>
            </div>
          )}
        </div>

        {error && (
          <div className="mb-4 rounded-lg border border-red-900 bg-red-950/40 px-4 py-3 text-sm text-red-300">{error}</div>
        )}

        <h2 className="text-lg font-semibold mb-3">{t.dash.gameServersTitle.replace("{game}", def.name)}</h2>

        {loading ? (
          <div className="text-zinc-400 text-center">{t.dash.loading}</div>
        ) : mine.length === 0 ? (
          <div className="border border-dashed border-zinc-700 rounded-xl p-10 text-center">
            <Server className="w-8 h-8 mx-auto mb-3 text-zinc-600" />
            <p className="text-zinc-400">{t.dash.empty}</p>
          </div>
        ) : (
          <div className="flex flex-col gap-3">
            {mine.map((s) => (
              <div key={s.id} className="border border-zinc-800 bg-zinc-900 rounded-xl p-5 flex items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                  <span className={`w-2.5 h-2.5 rounded-full shrink-0 ${metaOf(s.status).dot} ${metaOf(s.status).pulse ? "animate-pulse" : ""}`} />
                  <div>
                    <Link href={`/dashboard/servers/${s.id}`} className="font-semibold hover:text-white transition-colors">{s.name}</Link>
                    <div className="text-xs text-zinc-500 mt-0.5">
                      {s.subdomain}.servers.vbt-prog.com · {s.ram_mb / 1024} GB · {t.dash.plan} {s.plan}
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  <span className={`text-xs font-medium px-2 py-1 rounded-md bg-zinc-800 ${metaOf(s.status).text}`}>{labelOf(s.status)}</span>
                  {s.status === "stopped" && (
                    <button onClick={() => serverAction(s.id, "start")} className="p-2 rounded-lg hover:bg-zinc-800 text-green-400 transition-colors" title={t.dash.start}>
                      <Play className="w-4 h-4" />
                    </button>
                  )}
                  {s.status === "running" && (
                    <>
                      <button onClick={() => serverAction(s.id, "restart")} className="p-2 rounded-lg hover:bg-zinc-800 text-zinc-400 transition-colors" title={t.dash.restart}>
                        <RefreshCw className="w-4 h-4" />
                      </button>
                      <button onClick={() => serverAction(s.id, "stop")} className="p-2 rounded-lg hover:bg-zinc-800 text-red-400 transition-colors" title={t.dash.stop}>
                        <Square className="w-4 h-4" />
                      </button>
                    </>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
