"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Server, Plus, Play, Square, RefreshCw } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { GAMES } from "@/lib/games";
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

// Couleurs/animation par statut — les libellés viennent du dictionnaire i18n.
const STATUS_META: Record<string, { dot: string; text: string; pulse?: boolean }> = {
  running:  { dot: "bg-green-500",  text: "text-green-400" },
  creating: { dot: "bg-yellow-500", text: "text-yellow-400", pulse: true },
  stopped:  { dot: "bg-zinc-500",   text: "text-zinc-400" },
  error:    { dot: "bg-red-500",    text: "text-red-400" },
};

export default function Dashboard() {
  const router = useRouter();
  const { t } = useI18n();
  const [servers, setServers] = useState<GameServer[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState("");
  const [newGame, setNewGame] = useState("minecraft");
  const [newVersion, setNewVersion] = useState("LATEST");
  const [error, setError] = useState("");

  const metaOf = (s: string) => STATUS_META[s] ?? { dot: "bg-zinc-500", text: "text-zinc-400" };
  const labelOf = (s: string) => t.dash.status[s] ?? s;
  const gameName = GAMES.find((g) => g.id === newGame)?.name ?? newGame;
  const supportsVersion = newGame === "minecraft";

  useEffect(() => {
    fetchServers();
    // Poll toutes les 3s pour suivre les statuts (creating → running)
    const interval = setInterval(fetchServers, 3000);
    return () => clearInterval(interval);
  }, []);

  async function fetchServers() {
    try {
      const res = await fetch(`${API}/api/v1/servers`, { credentials: "include" });
      if (res.status === 401) { router.push("/"); return; }
      const data = await res.json();
      setServers(Array.isArray(data) ? data : []);
    } catch {
      /* garde l'état précédent */
    } finally {
      setLoading(false);
    }
  }

  async function createServer() {
    if (!newName.trim()) {
      setError(t.dash.errName);
      return;
    }
    setCreating(true);
    setError("");
    try {
      const res = await fetch(`${API}/api/v1/servers`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: newName, plan: "free", game: newGame, version: supportsVersion ? newVersion : "LATEST" }),
      });
      if (res.ok) {
        setNewName("");
        fetchServers();
      } else {
        const msg = await res.text();
        setError(msg || `Erreur ${res.status}`);
      }
    } catch {
      setError(t.dash.errConn);
    } finally {
      setCreating(false);
    }
  }

  async function serverAction(id: string, action: "start" | "stop" | "restart") {
    await fetch(`${API}/api/v1/servers/${id}/${action}`, {
      method: "POST",
      credentials: "include",
    });
    fetchServers();
  }

  return (
    <div className="min-h-screen flex flex-col">
      {/* Nav */}
      <nav className="border-b border-zinc-800 px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-2 font-bold text-lg">
          <Server className="w-5 h-5 text-green-400" />
          Playrena
        </div>
        <LanguageSwitcher />
      </nav>

      <div className="flex-1 p-6 max-w-5xl mx-auto w-full">
        <h1 className="text-2xl font-bold mb-6 text-center">{t.dash.title}</h1>

        {/* Bloc de création — centré et mis en avant, toujours visible */}
        <div className="mb-10 flex flex-col items-center text-center gap-5 rounded-2xl border border-green-500/30 bg-gradient-to-b from-zinc-900 to-zinc-950 px-6 py-10">
          <p className="text-zinc-300 max-w-md">
            {t.dash.leadPre}{gameName}{t.dash.leadMid}
            <span className="text-green-400 font-medium">{t.dash.leadFree}</span>
            {t.dash.leadPost}
          </p>
          <form
            onSubmit={(e) => { e.preventDefault(); createServer(); }}
            className="flex flex-col sm:flex-row items-stretch sm:items-center justify-center gap-3 w-full max-w-2xl"
          >
            <select
              value={newGame}
              onChange={(e) => setNewGame(e.target.value)}
              aria-label={t.dash.gameLabel}
              className="select-dark cursor-pointer bg-zinc-900 border border-zinc-700 rounded-lg pl-4 py-3 text-sm text-zinc-100 focus:outline-none focus:border-green-500"
            >
              {GAMES.map((g) => (
                <option key={g.id} value={g.id} disabled={g.status !== "live"}>
                  {g.name}{g.status !== "live" ? ` — ${t.dash.soon}` : ""}
                </option>
              ))}
            </select>
            <input
              type="text"
              placeholder={t.dash.namePlaceholder}
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              className="flex-1 bg-zinc-900 border border-zinc-700 rounded-lg px-4 py-3 text-sm focus:outline-none focus:border-green-500"
            />
            {supportsVersion && (
              <select
                value={newVersion}
                onChange={(e) => setNewVersion(e.target.value)}
                className="select-dark cursor-pointer bg-zinc-900 border border-zinc-700 rounded-lg pl-4 py-3 text-sm text-zinc-100 focus:outline-none focus:border-green-500"
              >
                {VERSIONS.map((v) => <option key={v} value={v}>{v === "LATEST" ? t.dash.versionLatest : v}</option>)}
              </select>
            )}
          </form>
          <button
            type="button"
            onClick={createServer}
            disabled={creating}
            className="flex items-center justify-center gap-2 bg-green-500 hover:bg-green-400 disabled:opacity-50 text-black font-semibold text-base px-8 py-4 rounded-xl shadow-lg shadow-green-500/30 hover:shadow-green-400/40 transition-all hover:scale-[1.02]"
          >
            <Plus className="w-5 h-5" />
            {creating ? t.dash.creating : t.dash.createBtn}
          </button>
        </div>

        {error && (
          <div className="mb-4 rounded-lg border border-red-900 bg-red-950/40 px-4 py-3 text-sm text-red-300">
            {error}
          </div>
        )}

        {loading ? (
          <div className="text-zinc-400 text-center">{t.dash.loading}</div>
        ) : servers.length === 0 ? (
          <div className="border border-dashed border-zinc-700 rounded-xl p-10 text-center">
            <Server className="w-8 h-8 mx-auto mb-3 text-zinc-600" />
            <p className="text-zinc-400">{t.dash.empty}</p>
          </div>
        ) : (
          <div className="flex flex-col gap-3">
            {servers.map((s) => (
              <div key={s.id} className="border border-zinc-800 bg-zinc-900 rounded-xl p-5 flex items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                  <span className={`w-2.5 h-2.5 rounded-full shrink-0 ${metaOf(s.status).dot} ${metaOf(s.status).pulse ? "animate-pulse" : ""}`} />
                  <div>
                    <Link href={`/dashboard/servers/${s.id}`} className="font-semibold hover:text-green-400 transition-colors">
                      {s.name}
                    </Link>
                    <div className="text-xs text-zinc-500 mt-0.5">
                      {s.subdomain}.servers.vbt-prog.com · {s.ram_mb / 1024} GB · {t.dash.plan} {s.plan}
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2 shrink-0">
                  <span className={`text-xs font-medium px-2 py-1 rounded-md bg-zinc-800 ${metaOf(s.status).text}`}>
                    {labelOf(s.status)}
                  </span>
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
