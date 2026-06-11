"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Server, Play, Square, RefreshCw, ChevronRight } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { GAMES, getGame } from "@/lib/games";
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

  const metaOf = (s: string) => STATUS_META[s] ?? { dot: "bg-zinc-500", text: "text-zinc-400" };
  const labelOf = (s: string) => t.dash.status[s] ?? s;
  const gameNameOf = (id: string) => getGame(id)?.name ?? id;

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

  return (
    <div className="min-h-screen flex flex-col">
      <nav className="border-b border-zinc-800 px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-2 font-bold text-lg">
          <Server className="w-5 h-5 text-green-400" />
          Playrena
        </div>
        <LanguageSwitcher />
      </nav>

      <div className="flex-1 p-6 max-w-5xl mx-auto w-full">
        {/* Choix du jeu */}
        <h1 className="text-2xl font-bold mb-1 text-center">{t.dash.chooseGame}</h1>
        <p className="text-zinc-400 text-center mb-8">{t.dash.hubLead}</p>

        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4 mb-12">
          {GAMES.map((g) => {
            const live = g.status === "live";
            const card = (
              <div className={`group relative h-full rounded-xl border bg-zinc-900 p-5 flex flex-col gap-3 transition-colors ${live ? "border-zinc-800 hover:border-zinc-600 cursor-pointer" : "border-zinc-900 opacity-60"}`}>
                <div className={`h-1.5 w-10 rounded-full bg-gradient-to-r ${g.accent}`} />
                <div className="font-semibold">{g.name}</div>
                <div className="mt-auto flex items-center justify-between text-xs">
                  <span className={live ? "text-green-400" : "text-zinc-500"}>{live ? t.hub.live : t.hub.soon}</span>
                  {live && <ChevronRight className="w-4 h-4 text-zinc-500 group-hover:text-zinc-300 transition-colors" />}
                </div>
              </div>
            );
            return live
              ? <Link key={g.id} href={`/dashboard/${g.id}`} className="block h-full">{card}</Link>
              : <div key={g.id} className="h-full" aria-disabled>{card}</div>;
          })}
        </div>

        {/* Tous les serveurs */}
        <h2 className="text-lg font-semibold mb-3">{t.dash.allServers}</h2>
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
                    <div className="flex items-center gap-2">
                      <Link href={`/dashboard/servers/${s.id}`} className="font-semibold hover:text-white transition-colors">{s.name}</Link>
                      <span className="text-[10px] uppercase tracking-wide px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-400">{gameNameOf(s.game)}</span>
                    </div>
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
