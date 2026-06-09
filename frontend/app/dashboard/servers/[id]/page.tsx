"use client";

import { useEffect, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Play, Square, RefreshCw, Server, ArrowUpCircle } from "lucide-react";

const VERSIONS = ["LATEST", "1.21.4", "1.21.1", "1.20.6", "1.20.4", "1.20.1", "1.19.4", "1.18.2", "1.16.5", "1.12.2", "1.8.9"];

const PLANS = [
  { id: "free", label: "Gratuit", specs: "1 GB · 1 cœur", price: "0$" },
  { id: "starter", label: "Starter", specs: "2 GB · 1 cœur", price: "3$/mo" },
  { id: "standard", label: "Standard", specs: "4 GB · 2 cœurs", price: "7$/mo" },
  { id: "pro", label: "Pro", specs: "8 GB · 4 cœurs", price: "14$/mo" },
  { id: "extreme", label: "Extreme", specs: "16 GB · 6 cœurs", price: "25$/mo" },
];

interface GameServer {
  id: string;
  name: string;
  subdomain: string;
  game: string;
  plan: string;
  version: string;
  status: string;
  ram_mb: number;
  cpu_cores: number;
  port: number;
}

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const STATUS: Record<string, { label: string; dot: string; text: string; pulse?: boolean }> = {
  running:  { label: "En ligne",        dot: "bg-green-500",  text: "text-green-400" },
  creating: { label: "Initialisation…", dot: "bg-yellow-500", text: "text-yellow-400", pulse: true },
  stopped:  { label: "Arrêté",          dot: "bg-zinc-500",   text: "text-zinc-400" },
  error:    { label: "Erreur",          dot: "bg-red-500",    text: "text-red-400" },
};
const statusOf = (s: string) => STATUS[s] ?? { label: s, dot: "bg-zinc-500", text: "text-zinc-400" };

export default function ServerPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const [server, setServer] = useState<GameServer | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedPlan, setSelectedPlan] = useState("");
  const [upgrading, setUpgrading] = useState(false);
  const [selectedVersion, setSelectedVersion] = useState("");
  const [changingVersion, setChangingVersion] = useState(false);
  const [paypalReady, setPaypalReady] = useState(false);
  const [paypalEnabled, setPaypalEnabled] = useState(false);
  const paypalRef = useRef<HTMLDivElement>(null);

  // Charge la config billing + le SDK PayPal
  useEffect(() => {
    (async () => {
      const res = await fetch(`${API}/api/v1/billing/config`, { credentials: "include" });
      if (!res.ok) return;
      const cfg = await res.json();
      if (!cfg.enabled || !cfg.paypal_client_id) return;
      setPaypalEnabled(true);
      if ((window as any).paypal) { setPaypalReady(true); return; }
      const sc = document.createElement("script");
      sc.src = `https://www.paypal.com/sdk/js?client-id=${cfg.paypal_client_id}&currency=USD`;
      sc.onload = () => setPaypalReady(true);
      document.body.appendChild(sc);
    })();
  }, []);

  // (Re)rend les boutons PayPal quand un plan payant est sélectionné
  useEffect(() => {
    const paypal = (window as any).paypal;
    if (!paypalReady || !paypal || !paypalRef.current) return;
    if (!selectedPlan || selectedPlan === "free") return;
    paypalRef.current.innerHTML = "";
    paypal.Buttons({
      style: { layout: "horizontal", color: "blue", height: 38 },
      createOrder: async () => {
        const res = await fetch(`${API}/api/v1/servers/${id}/checkout/paypal`, {
          method: "POST", credentials: "include",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ plan: selectedPlan }),
        });
        const d = await res.json();
        return d.order_id;
      },
      onApprove: async (data: any) => {
        const res = await fetch(`${API}/api/v1/servers/${id}/checkout/paypal/${data.orderID}/capture`, {
          method: "POST", credentials: "include",
        });
        if (res.ok) {
          router.push(`/merci?plan=${selectedPlan}&server=${id}&txn=${data.orderID}`);
        } else {
          alert("Erreur: " + (await res.text()));
        }
      },
      onError: (err: any) => alert("Erreur PayPal: " + err),
    }).render(paypalRef.current);
  }, [paypalReady, selectedPlan, id]);

  useEffect(() => {
    fetchServer();
    const t = setInterval(fetchServer, 4000);
    return () => clearInterval(t);
  }, [id]);

  async function fetchServer() {
    const res = await fetch(`${API}/api/v1/servers/${id}`, { credentials: "include" });
    if (res.status === 401) { router.push("/"); return; }
    if (res.status === 404) { router.push("/dashboard"); return; }
    setServer(await res.json());
    setLoading(false);
  }

  async function action(a: "start" | "stop" | "restart") {
    await fetch(`${API}/api/v1/servers/${id}/${a}`, { method: "POST", credentials: "include" });
    fetchServer();
  }

  async function upgrade() {
    if (!selectedPlan) return;
    setUpgrading(true);
    const res = await fetch(`${API}/api/v1/servers/${id}/upgrade`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ plan: selectedPlan }),
    });
    if (res.ok) { setSelectedPlan(""); fetchServer(); }
    else alert("Erreur: " + (await res.text()));
    setUpgrading(false);
  }

  async function changeVersion() {
    if (!selectedVersion) return;
    if (!confirm(`Changer la version vers ${selectedVersion} ? Le serveur va redémarrer (le monde est conservé).`)) return;
    setChangingVersion(true);
    const res = await fetch(`${API}/api/v1/servers/${id}/version`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ version: selectedVersion }),
    });
    if (res.ok) { setSelectedVersion(""); fetchServer(); }
    else alert("Erreur: " + (await res.text()));
    setChangingVersion(false);
  }

  async function deleteServer() {
    if (!confirm("Supprimer ce serveur ? Les données du monde seront conservées.")) return;
    await fetch(`${API}/api/v1/servers/${id}`, { method: "DELETE", credentials: "include" });
    router.push("/dashboard");
  }

  if (loading) return <div className="p-8 text-zinc-400">Chargement...</div>;
  if (!server) return null;

  return (
    <div className="min-h-screen flex flex-col">
      {/* Nav */}
      <nav className="border-b border-zinc-800 px-6 py-4 flex items-center gap-3">
        <Link href="/dashboard" className="text-zinc-400 hover:text-zinc-100 transition-colors">
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <Server className="w-5 h-5 text-green-400" />
        <span className="font-bold">{server.name}</span>
        <span className={`flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-md bg-zinc-800 ml-2 ${statusOf(server.status).text}`}>
          <span className={`w-2 h-2 rounded-full ${statusOf(server.status).dot} ${statusOf(server.status).pulse ? "animate-pulse" : ""}`} />
          {statusOf(server.status).label}
        </span>
      </nav>

      <div className="flex-1 p-6 max-w-4xl mx-auto w-full flex flex-col gap-6">
        {/* Infos + actions */}
        <div className="border border-zinc-800 bg-zinc-900 rounded-xl p-6 flex flex-col sm:flex-row gap-6 justify-between">
          <div className="flex flex-col gap-2 text-sm">
            <div className="text-zinc-400">Adresse de connexion</div>
            <code className="font-mono text-green-400">{server.subdomain}.servers.vbt-prog.com:{server.port}</code>
            <div className="text-zinc-500 mt-2">
              Plan <span className="text-zinc-300 capitalize">{server.plan}</span>
              {" · "}{server.ram_mb / 1024} GB RAM
              {" · "}{server.cpu_cores} CPU
              {" · "}version <span className="text-zinc-300">{server.version === "LATEST" ? "dernière" : server.version}</span>
            </div>
          </div>
          <div className="flex gap-2 items-start">
            {server.status === "stopped" && (
              <button onClick={() => action("start")} className="flex items-center gap-2 bg-green-500 hover:bg-green-400 text-black font-medium px-4 py-2 rounded-lg text-sm transition-colors">
                <Play className="w-4 h-4" /> Démarrer
              </button>
            )}
            {server.status === "running" && (
              <>
                <button onClick={() => action("restart")} className="flex items-center gap-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-100 font-medium px-4 py-2 rounded-lg text-sm transition-colors">
                  <RefreshCw className="w-4 h-4" /> Redémarrer
                </button>
                <button onClick={() => action("stop")} className="flex items-center gap-2 bg-zinc-800 hover:bg-red-900 text-red-400 font-medium px-4 py-2 rounded-lg text-sm transition-colors">
                  <Square className="w-4 h-4" /> Arrêter
                </button>
              </>
            )}
          </div>
        </div>

        {/* Upgrade / changement de plan */}
        <div className="border border-zinc-800 bg-zinc-900 rounded-xl p-6">
          <div className="flex items-center gap-2 font-semibold mb-1">
            <ArrowUpCircle className="w-5 h-5 text-green-400" /> Changer de plan
          </div>
          <p className="text-sm text-zinc-500 mb-4">
            Plan actuel : <span className="text-zinc-300 capitalize">{server.plan}</span>.
            Les plans payants nécessitent un abonnement.
          </p>
          <div className="grid grid-cols-2 sm:grid-cols-5 gap-2 mb-4">
            {PLANS.map((p) => {
              const current = p.id === server.plan;
              const sel = p.id === selectedPlan;
              return (
                <button
                  key={p.id}
                  disabled={current}
                  onClick={() => setSelectedPlan(p.id)}
                  className={`rounded-lg border p-3 text-left text-xs transition-colors ${
                    current ? "border-zinc-700 bg-zinc-800 opacity-50 cursor-default"
                    : sel ? "border-green-500 bg-green-950/30"
                    : "border-zinc-700 hover:border-zinc-500"
                  }`}
                >
                  <div className="font-semibold text-sm">{p.label}</div>
                  <div className="text-zinc-400">{p.specs}</div>
                  <div className="text-zinc-500 mt-1">{p.price}</div>
                  {current && <div className="text-[10px] text-green-400 mt-1">actuel</div>}
                  {p.id !== "free" && <div className="text-[10px] text-zinc-600 mt-1">abonnement</div>}
                </button>
              );
            })}
          </div>
          {selectedPlan && selectedPlan !== "free" && (
            <div className="mb-4 rounded-lg border border-indigo-900 bg-indigo-950/30 px-4 py-3 text-sm text-indigo-300">
              Le plan <b>{PLANS.find(p => p.id === selectedPlan)?.label}</b> à {PLANS.find(p => p.id === selectedPlan)?.price} —
              paie avec PayPal pour l'activer immédiatement.
            </div>
          )}
          {selectedPlan && selectedPlan !== "free" ? (
            paypalEnabled ? (
              <div ref={paypalRef} className="max-w-xs" />
            ) : (
              <button disabled className="bg-indigo-600/40 text-indigo-200 text-sm font-medium px-5 py-2 rounded-lg cursor-not-allowed">
                Paiement bientôt disponible
              </button>
            )
          ) : (
            <button
              onClick={upgrade}
              disabled={!selectedPlan || upgrading}
              className="bg-green-500 hover:bg-green-400 disabled:opacity-40 disabled:cursor-not-allowed text-black text-sm font-medium px-5 py-2 rounded-lg transition-colors"
            >
              {upgrading ? "Application..." : selectedPlan ? "Repasser au plan Gratuit" : "Choisis un plan"}
            </button>
          )}
        </div>

        {/* Version Minecraft */}
        <div className="border border-zinc-800 bg-zinc-900 rounded-xl p-6">
          <div className="flex items-center gap-2 font-semibold mb-1">
            <RefreshCw className="w-5 h-5 text-green-400" /> Version Minecraft
          </div>
          <p className="text-sm text-zinc-500 mb-4">
            Version actuelle : <span className="text-zinc-300">{server.version === "LATEST" ? "dernière (LATEST)" : server.version}</span>.
            Changer la version <span className="text-zinc-400">redémarre le serveur</span> — le monde et les configs sont conservés.
          </p>
          <div className="flex items-center gap-2">
            <select
              value={selectedVersion}
              onChange={(e) => setSelectedVersion(e.target.value)}
              className="bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-green-500"
            >
              <option value="">Choisir une version…</option>
              {VERSIONS.filter((v) => v !== server.version).map((v) => (
                <option key={v} value={v}>{v === "LATEST" ? "Dernière version" : v}</option>
              ))}
            </select>
            <button
              onClick={changeVersion}
              disabled={!selectedVersion || changingVersion}
              className="bg-green-500 hover:bg-green-400 disabled:opacity-40 disabled:cursor-not-allowed text-black text-sm font-medium px-5 py-2 rounded-lg transition-colors"
            >
              {changingVersion ? "Application..." : "Appliquer"}
            </button>
          </div>
        </div>

        {/* Sections à venir */}
        {[
          { title: "Console", desc: "Terminal temps réel — bientôt disponible" },
          { title: "Mods & Plugins", desc: "Gestion des mods — bientôt disponible" },
          { title: "Fichiers", desc: "Explorateur de fichiers — bientôt disponible" },
          { title: "Joueurs", desc: "Whitelist, banlist, ops — bientôt disponible" },
        ].map(({ title, desc }) => (
          <div key={title} className="border border-dashed border-zinc-800 rounded-xl p-6">
            <div className="font-semibold mb-1">{title}</div>
            <div className="text-sm text-zinc-500">{desc}</div>
          </div>
        ))}

        {/* Danger zone */}
        <div className="border border-red-900 rounded-xl p-6 mt-4">
          <div className="font-semibold text-red-400 mb-2">Zone de danger</div>
          <p className="text-sm text-zinc-400 mb-4">Supprimer le serveur arrête le container Docker. Les données du monde sont conservées.</p>
          <button onClick={deleteServer} className="bg-red-900 hover:bg-red-800 text-red-300 text-sm font-medium px-4 py-2 rounded-lg transition-colors">
            Supprimer ce serveur
          </button>
        </div>
      </div>
    </div>
  );
}
