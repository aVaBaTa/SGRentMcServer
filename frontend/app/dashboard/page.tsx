"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Server, Plus, Play, Square, RefreshCw } from "lucide-react";

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

const STATUS: Record<string, { label: string; dot: string; text: string; pulse?: boolean }> = {
  running:  { label: "En ligne",        dot: "bg-green-500",  text: "text-green-400" },
  creating: { label: "Initialisation…", dot: "bg-yellow-500", text: "text-yellow-400", pulse: true },
  stopped:  { label: "Arrêté",          dot: "bg-zinc-500",   text: "text-zinc-400" },
  error:    { label: "Erreur",          dot: "bg-red-500",    text: "text-red-400" },
};

function statusOf(s: string) {
  return STATUS[s] ?? { label: s, dot: "bg-zinc-500", text: "text-zinc-400" };
}

export default function Dashboard() {
  const router = useRouter();
  const [servers, setServers] = useState<GameServer[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState("");
  const [newVersion, setNewVersion] = useState("LATEST");
  const [error, setError] = useState("");

  useEffect(() => {
    fetchServers();
    // Poll toutes les 3s pour suivre les statuts (creating → running)
    const t = setInterval(fetchServers, 3000);
    return () => clearInterval(t);
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
      setError("Entre un nom pour ton serveur.");
      return;
    }
    setCreating(true);
    setError("");
    try {
      const res = await fetch(`${API}/api/v1/servers`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: newName, plan: "free", game: "minecraft", version: newVersion }),
      });
      if (res.ok) {
        setNewName("");
        fetchServers();
      } else {
        const msg = await res.text();
        setError(msg || `Erreur ${res.status}`);
      }
    } catch (e) {
      setError("Connexion au serveur impossible.");
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
        <form
          onSubmit={(e) => { e.preventDefault(); createServer(); }}
          className="flex items-center gap-2"
        >
          <input
            type="text"
            placeholder="Nom du serveur"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            className="bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-green-500"
          />
          <select
            value={newVersion}
            onChange={(e) => setNewVersion(e.target.value)}
            className="bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-green-500"
          >
            {VERSIONS.map((v) => <option key={v} value={v}>{v === "LATEST" ? "Dernière version" : v}</option>)}
          </select>
          <button
            type="submit"
            disabled={creating}
            className="flex items-center gap-1 bg-green-500 hover:bg-green-400 disabled:opacity-50 text-black font-medium px-4 py-2 rounded-lg text-sm transition-colors"
          >
            <Plus className="w-4 h-4" />
            {creating ? "Création..." : "Nouveau serveur"}
          </button>
        </form>
      </nav>

      <div className="flex-1 p-6 max-w-5xl mx-auto w-full">
        <h1 className="text-2xl font-bold mb-6">Mes serveurs</h1>

        {error && (
          <div className="mb-4 rounded-lg border border-red-900 bg-red-950/40 px-4 py-3 text-sm text-red-300">
            {error}
          </div>
        )}

        {loading ? (
          <div className="text-zinc-400">Chargement...</div>
        ) : servers.length === 0 ? (
          <div className="border border-dashed border-zinc-700 rounded-xl p-10 text-center">
            <Server className="w-8 h-8 mx-auto mb-3 text-zinc-600" />
            <p className="text-zinc-400 mb-5">Aucun serveur pour l'instant. Crée ton premier serveur Minecraft gratuitement.</p>
            <form
              onSubmit={(e) => { e.preventDefault(); createServer(); }}
              className="flex items-center justify-center gap-2"
            >
              <input
                type="text"
                placeholder="Nom de ton serveur"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                className="bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-green-500"
              />
              <select
                value={newVersion}
                onChange={(e) => setNewVersion(e.target.value)}
                className="bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-green-500"
              >
                {VERSIONS.map((v) => <option key={v} value={v}>{v === "LATEST" ? "Dernière version" : v}</option>)}
              </select>
              <button
                type="submit"
                disabled={creating}
                className="flex items-center gap-1 bg-green-500 hover:bg-green-400 disabled:opacity-50 text-black font-medium px-4 py-2 rounded-lg text-sm transition-colors"
              >
                <Plus className="w-4 h-4" />
                {creating ? "Création..." : "Créer mon serveur"}
              </button>
            </form>
          </div>
        ) : (
          <div className="flex flex-col gap-3">
            {servers.map((s) => (
              <div key={s.id} className="border border-zinc-800 bg-zinc-900 rounded-xl p-5 flex items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                  <span className={`w-2.5 h-2.5 rounded-full shrink-0 ${statusOf(s.status).dot} ${statusOf(s.status).pulse ? "animate-pulse" : ""}`} />
                  <div>
                    <Link href={`/dashboard/servers/${s.id}`} className="font-semibold hover:text-green-400 transition-colors">
                      {s.name}
                    </Link>
                    <div className="text-xs text-zinc-500 mt-0.5">
                      {s.subdomain}.servers.vbt-prog.com · {s.ram_mb / 1024} GB · Plan {s.plan}
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2 shrink-0">
                  <span className={`text-xs font-medium px-2 py-1 rounded-md bg-zinc-800 ${statusOf(s.status).text}`}>
                    {statusOf(s.status).label}
                  </span>
                  {s.status === "stopped" && (
                    <button onClick={() => serverAction(s.id, "start")} className="p-2 rounded-lg hover:bg-zinc-800 text-green-400 transition-colors" title="Démarrer">
                      <Play className="w-4 h-4" />
                    </button>
                  )}
                  {s.status === "running" && (
                    <>
                      <button onClick={() => serverAction(s.id, "restart")} className="p-2 rounded-lg hover:bg-zinc-800 text-zinc-400 transition-colors" title="Redémarrer">
                        <RefreshCw className="w-4 h-4" />
                      </button>
                      <button onClick={() => serverAction(s.id, "stop")} className="p-2 rounded-lg hover:bg-zinc-800 text-red-400 transition-colors" title="Arrêter">
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
