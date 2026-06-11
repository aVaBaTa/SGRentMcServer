"use client";

import { useEffect, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Play, Square, RefreshCw, Server, ArrowUpCircle, Users, Terminal, SendHorizontal, Shield, Ban, UserMinus, UserPlus, Folder, FileText, Upload, Download, Trash2, FolderPlus, Save, X, ChevronRight, KeyRound } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { LanguageSwitcher } from "@/components/site-chrome";

const VERSIONS = ["LATEST", "1.21.4", "1.21.1", "1.20.6", "1.20.4", "1.20.1", "1.19.4", "1.18.2", "1.16.5", "1.12.2", "1.8.9"];

const PLANS = [
  { id: "free", ram: "1 GB", cores: 1, price: "0$" },
  { id: "starter", ram: "2 GB", cores: 1, price: "2$/mo" },
  { id: "standard", ram: "4 GB", cores: 2, price: "5$/mo" },
  { id: "pro", ram: "8 GB", cores: 4, price: "9$/mo" },
  { id: "extreme", ram: "16 GB", cores: 6, price: "16$/mo" },
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

// Couleurs/animation par statut — les libellés viennent du dictionnaire i18n.
const STATUS_META: Record<string, { dot: string; text: string; pulse?: boolean }> = {
  running:       { dot: "bg-green-500",  text: "text-green-400" },
  creating:      { dot: "bg-yellow-500", text: "text-yellow-400", pulse: true },
  auth_required: { dot: "bg-amber-500",  text: "text-amber-400", pulse: true },
  stopped:       { dot: "bg-zinc-500",   text: "text-zinc-400" },
  error:         { dot: "bg-red-500",    text: "text-red-400" },
};

export default function ServerPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const { t } = useI18n();
  const [server, setServer] = useState<GameServer | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedPlan, setSelectedPlan] = useState("");
  const [upgrading, setUpgrading] = useState(false);
  const [selectedVersion, setSelectedVersion] = useState("");
  const [changingVersion, setChangingVersion] = useState(false);
  const [paypalReady, setPaypalReady] = useState(false);
  const [paypalEnabled, setPaypalEnabled] = useState(false);
  const paypalRef = useRef<HTMLDivElement>(null);

  const statusMeta = (s: string) => STATUS_META[s] ?? { dot: "bg-zinc-500", text: "text-zinc-400" };
  const statusLabel = (s: string) => t.dash.status[s] ?? s;

  const fmtBytes = (n: number) => {
    const u = t.srv.bytes;
    if (n < 1024) return n + " " + u.b;
    if (n < 1048576) return (n / 1024).toFixed(0) + " " + u.kb;
    if (n < 1073741824) return (n / 1048576).toFixed(1) + " " + u.mb;
    return (n / 1073741824).toFixed(2) + " " + u.gb;
  };

  // Joueurs + console
  const [players, setPlayers] = useState<{ online: number; max: number; players: string[] } | null>(null);
  const [logs, setLogs] = useState("");
  const [command, setCommand] = useState("");
  const [sending, setSending] = useState(false);
  const logRef = useRef<HTMLPreElement>(null);

  // Gestion des joueurs
  const [lists, setLists] = useState<{ whitelist: string[]; banned: string[] }>({ whitelist: [], banned: [] });
  const [newPlayer, setNewPlayer] = useState("");
  const [playerAct, setPlayerAct] = useState("whitelist_add");
  const [actBusy, setActBusy] = useState(false);

  // Fichiers
  interface FileEntry { name: string; is_dir: boolean; size: number; mtime: number; }
  const [filesPath, setFilesPath] = useState("/");
  const [entries, setEntries] = useState<FileEntry[]>([]);
  const [filesBusy, setFilesBusy] = useState(false);
  const [editPath, setEditPath] = useState<string | null>(null);
  const [editContent, setEditContent] = useState("");
  const [editTrunc, setEditTrunc] = useState(false);
  const [savingFile, setSavingFile] = useState(false);
  const [uploading, setUploading] = useState(false);

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
          alert(t.srv.errPrefix + (await res.text()));
        }
      },
      onError: (err: any) => alert(t.srv.errPaypal + err),
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

  // Authentification interactive (Hytale) : poll de l'URL+code OAuth
  const [authInfo, setAuthInfo] = useState<{ pending: boolean; step?: number; url?: string; code?: string; raw?: string[] } | null>(null);
  const authRequired = server?.status === "auth_required";
  useEffect(() => {
    if (!authRequired) { setAuthInfo(null); return; }
    const tick = async () => {
      try {
        const res = await fetch(`${API}/api/v1/servers/${id}/auth`, { credentials: "include" });
        if (res.ok) setAuthInfo(await res.json());
      } catch { /* serveur pas prêt */ }
    };
    tick();
    const t = setInterval(tick, 4000);
    return () => clearInterval(t);
  }, [id, authRequired]);

  // Joueurs + console : poll quand le serveur tourne
  const running = server?.status === "running";
  useEffect(() => {
    if (!running) { setPlayers(null); return; }
    const tick = async () => {
      try {
        const [p, l, pl] = await Promise.all([
          fetch(`${API}/api/v1/servers/${id}/players`, { credentials: "include" }),
          fetch(`${API}/api/v1/servers/${id}/logs?tail=200`, { credentials: "include" }),
          fetch(`${API}/api/v1/servers/${id}/playerlists`, { credentials: "include" }),
        ]);
        if (p.ok) setPlayers(await p.json());
        if (l.ok) {
          const stick = logRef.current && logRef.current.scrollTop + logRef.current.clientHeight >= logRef.current.scrollHeight - 40;
          setLogs((await l.json()).logs ?? "");
          if (stick) requestAnimationFrame(() => { if (logRef.current) logRef.current.scrollTop = logRef.current.scrollHeight; });
        }
        if (pl.ok) {
          const d = await pl.json();
          setLists({ whitelist: d.whitelist ?? [], banned: d.banned ?? [] });
        }
      } catch { /* serveur pas prêt */ }
    };
    tick();
    const t = setInterval(tick, 5000);
    return () => clearInterval(t);
  }, [id, running]);

  async function doPlayerAction(action: string, player: string) {
    const p = player.trim();
    if (!p || actBusy) return;
    setActBusy(true);
    try {
      const res = await fetch(`${API}/api/v1/servers/${id}/players/action`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action, player: p }),
      });
      if (res.ok) {
        const out = (await res.json()).output ?? "";
        if (out) setLogs((prev) => prev + `\n> ${action} ${p}\n${out}`.trimEnd() + "\n");
        setNewPlayer("");
      }
    } finally {
      setActBusy(false);
    }
  }

  // ---- Fichiers ----
  const joinPath = (dir: string, name: string) => (dir === "/" ? "/" + name : dir + "/" + name);

  useEffect(() => {
    if (!running) { setEntries([]); return; }
    loadFiles(filesPath);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, running, filesPath]);

  async function loadFiles(p: string) {
    setFilesBusy(true);
    try {
      const res = await fetch(`${API}/api/v1/servers/${id}/files?path=${encodeURIComponent(p)}`, { credentials: "include" });
      if (res.ok) {
        const d = await res.json();
        const list: FileEntry[] = d.entries ?? [];
        list.sort((a, b) => (a.is_dir !== b.is_dir ? (a.is_dir ? -1 : 1) : a.name.localeCompare(b.name)));
        setEntries(list);
      }
    } finally {
      setFilesBusy(false);
    }
  }

  async function openFile(name: string) {
    const p = joinPath(filesPath, name);
    const res = await fetch(`${API}/api/v1/servers/${id}/files/content?path=${encodeURIComponent(p)}`, { credentials: "include" });
    if (res.ok) {
      const d = await res.json();
      setEditPath(p);
      setEditContent(d.content ?? "");
      setEditTrunc(!!d.truncated);
    } else {
      alert(t.srv.openFileErr);
    }
  }

  async function saveFile() {
    if (!editPath) return;
    setSavingFile(true);
    try {
      const res = await fetch(`${API}/api/v1/servers/${id}/files/content`, {
        method: "PUT", credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ path: editPath, content: editContent }),
      });
      if (res.ok) setEditPath(null);
      else alert(t.srv.saveFail);
    } finally {
      setSavingFile(false);
    }
  }

  async function deleteEntry(e: FileEntry) {
    const msg = t.srv.delAsk
      .replace("{type}", e.is_dir ? t.srv.delFolder : t.srv.delFile)
      .replace("{name}", e.name)
      .replace("{contents}", e.is_dir ? t.srv.delContents : "");
    if (!confirm(msg)) return;
    const res = await fetch(`${API}/api/v1/servers/${id}/files`, {
      method: "DELETE", credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ path: joinPath(filesPath, e.name) }),
    });
    if (res.ok) loadFiles(filesPath);
  }

  async function uploadFile(f: File) {
    setUploading(true);
    try {
      const fd = new FormData();
      fd.append("file", f);
      const res = await fetch(`${API}/api/v1/servers/${id}/files/upload?path=${encodeURIComponent(filesPath)}`, {
        method: "POST", credentials: "include", body: fd,
      });
      if (res.ok) loadFiles(filesPath);
      else alert(t.srv.uploadFail);
    } finally {
      setUploading(false);
    }
  }

  async function mkdir() {
    const name = prompt(t.srv.mkdirPrompt);
    if (!name) return;
    const res = await fetch(`${API}/api/v1/servers/${id}/files/mkdir`, {
      method: "POST", credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ path: joinPath(filesPath, name) }),
    });
    if (res.ok) loadFiles(filesPath);
  }

  function downloadEntry(name: string) {
    const p = joinPath(filesPath, name);
    window.open(`${API}/api/v1/servers/${id}/files/download?path=${encodeURIComponent(p)}`, "_blank");
  }

  async function sendCommand(e: React.FormEvent) {
    e.preventDefault();
    const cmd = command.trim();
    if (!cmd || sending) return;
    setSending(true);
    try {
      const res = await fetch(`${API}/api/v1/servers/${id}/command`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ command: cmd }),
      });
      if (res.ok) {
        const out = (await res.json()).output ?? "";
        setLogs((prev) => prev + `\n> ${cmd}\n${out}`.trimEnd() + "\n");
        requestAnimationFrame(() => { if (logRef.current) logRef.current.scrollTop = logRef.current.scrollHeight; });
      }
      setCommand("");
    } finally {
      setSending(false);
    }
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
    else alert(t.srv.errPrefix + (await res.text()));
    setUpgrading(false);
  }

  async function changeVersion() {
    if (!selectedVersion) return;
    if (!confirm(t.srv.confirmVersion.replace("{v}", selectedVersion))) return;
    setChangingVersion(true);
    const res = await fetch(`${API}/api/v1/servers/${id}/version`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ version: selectedVersion }),
    });
    if (res.ok) { setSelectedVersion(""); fetchServer(); }
    else alert(t.srv.errPrefix + (await res.text()));
    setChangingVersion(false);
  }

  async function deleteServer() {
    if (!confirm(t.srv.confirmDelete)) return;
    await fetch(`${API}/api/v1/servers/${id}`, { method: "DELETE", credentials: "include" });
    router.push("/dashboard");
  }

  if (loading) return <div className="p-8 text-zinc-400">{t.dash.loading}</div>;
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
        <span className={`flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-md bg-zinc-800 ml-2 ${statusMeta(server.status).text}`}>
          <span className={`w-2 h-2 rounded-full ${statusMeta(server.status).dot} ${statusMeta(server.status).pulse ? "animate-pulse" : ""}`} />
          {statusLabel(server.status)}
        </span>
        {running && players && (
          <span className="flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-md bg-zinc-800 text-zinc-300">
            <Users className="w-3.5 h-3.5 text-green-400" />
            {players.online} / {players.max} {t.pricing.players}
          </span>
        )}
        <div className="ml-auto">
          <LanguageSwitcher />
        </div>
      </nav>

      <div className="flex-1 p-6 max-w-4xl mx-auto w-full flex flex-col gap-6">
        {/* Infos + actions */}
        <div className="border border-zinc-800 bg-zinc-900 rounded-xl p-6 flex flex-col sm:flex-row gap-6 justify-between">
          <div className="flex flex-col gap-2 text-sm">
            <div className="text-zinc-400">{t.srv.connAddress}</div>
            <code className="font-mono text-green-400">
              {server.game === "minecraft"
                ? `${server.subdomain}.servers.vbt-prog.com:${server.port}`
                : `${process.env.NEXT_PUBLIC_SERVER_HOST ?? "24.157.140.226"}:${server.port}`}
            </code>
            <div className="text-zinc-500 mt-2">
              {t.dash.plan} <span className="text-zinc-300 capitalize">{server.plan}</span>
              {" · "}{server.ram_mb / 1024} GB RAM
              {" · "}{server.cpu_cores} CPU
              {" · "}{t.srv.version} <span className="text-zinc-300">{server.version === "LATEST" ? t.srv.versionLatestShort : server.version}</span>
            </div>
          </div>
          <div className="flex gap-2 items-start">
            {server.status === "stopped" && (
              <button onClick={() => action("start")} className="flex items-center gap-2 bg-green-500 hover:bg-green-400 text-black font-medium px-4 py-2 rounded-lg text-sm transition-colors">
                <Play className="w-4 h-4" /> {t.dash.start}
              </button>
            )}
            {server.status === "running" && (
              <>
                <button onClick={() => action("restart")} className="flex items-center gap-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-100 font-medium px-4 py-2 rounded-lg text-sm transition-colors">
                  <RefreshCw className="w-4 h-4" /> {t.dash.restart}
                </button>
                <button onClick={() => action("stop")} className="flex items-center gap-2 bg-zinc-800 hover:bg-red-900 text-red-400 font-medium px-4 py-2 rounded-lg text-sm transition-colors">
                  <Square className="w-4 h-4" /> {t.dash.stop}
                </button>
              </>
            )}
          </div>
        </div>

        {/* Autorisation interactive (Hytale) : OAuth device-code */}
        {authRequired && (
          <div className="border border-amber-800/60 bg-amber-950/20 rounded-xl p-6 flex flex-col gap-3">
            <div className="flex items-center gap-2 font-semibold text-amber-300">
              <KeyRound className="w-5 h-5" /> {t.srv.authTitle}
              {authInfo?.step ? (
                <span className="text-xs font-normal text-amber-400/80">· {t.srv.authStep.replace("{n}", String(authInfo.step))}</span>
              ) : null}
            </div>
            <p className="text-sm text-zinc-300">{t.srv.authDesc}</p>
            {authInfo?.pending && authInfo.url ? (
              <div className="flex flex-col gap-3">
                <a
                  href={authInfo.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex w-fit items-center gap-2 bg-amber-500 hover:bg-amber-400 text-black font-semibold px-5 py-2.5 rounded-lg transition-colors"
                >
                  {t.srv.authVisit}
                </a>
                {authInfo.code && (
                  <div className="flex items-center gap-2 text-sm">
                    <span className="text-zinc-400">{t.srv.authCode}</span>
                    <code className="font-mono text-lg tracking-widest text-amber-300 bg-black/40 px-3 py-1 rounded-md">{authInfo.code}</code>
                  </div>
                )}
              </div>
            ) : null}
            {authInfo?.raw && authInfo.raw.length > 0 && (
              <div>
                <div className="text-xs text-zinc-500 mb-1">{t.srv.authRaw}</div>
                <pre className="max-h-40 overflow-auto bg-black/50 rounded-lg p-3 text-[11px] font-mono text-amber-200/90 whitespace-pre-wrap break-words">{authInfo.raw.join("\n")}</pre>
                <p className="text-xs text-zinc-500 mt-2">{t.srv.authManual}</p>
              </div>
            )}
            <p className="text-xs text-zinc-500">{t.srv.authWaiting}</p>
          </div>
        )}

        {/* Upgrade / changement de plan */}
        <div className="border border-zinc-800 bg-zinc-900 rounded-xl p-6">
          <div className="flex items-center gap-2 font-semibold mb-1">
            <ArrowUpCircle className="w-5 h-5 text-green-400" /> {t.srv.changePlan}
          </div>
          <p className="text-sm text-zinc-500 mb-4">
            {t.srv.currentPlan} <span className="text-zinc-300 capitalize">{server.plan}</span>.
            {" "}{t.srv.paidNeedSub}
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
                  <div className="font-semibold text-sm">{t.planNames[p.id]}</div>
                  <div className="text-zinc-400">{p.ram} · {p.cores} {p.cores > 1 ? t.srv.cores : t.srv.core}</div>
                  <div className="text-zinc-500 mt-1">{p.price}</div>
                  {current && <div className="text-[10px] text-green-400 mt-1">{t.srv.current}</div>}
                  {p.id !== "free" && <div className="text-[10px] text-zinc-600 mt-1">{t.srv.subscription}</div>}
                </button>
              );
            })}
          </div>
          {selectedPlan && selectedPlan !== "free" && (
            <div className="mb-4 rounded-lg border border-indigo-900 bg-indigo-950/30 px-4 py-3 text-sm text-indigo-300">
              {t.srv.planPayHint
                .replace("{plan}", t.planNames[selectedPlan] ?? selectedPlan)
                .replace("{price}", PLANS.find((p) => p.id === selectedPlan)?.price ?? "")}
            </div>
          )}
          {selectedPlan && selectedPlan !== "free" ? (
            paypalEnabled ? (
              <div ref={paypalRef} className="max-w-xs" />
            ) : (
              <button disabled className="bg-indigo-600/40 text-indigo-200 text-sm font-medium px-5 py-2 rounded-lg cursor-not-allowed">
                {t.srv.payComingSoon}
              </button>
            )
          ) : (
            <button
              onClick={upgrade}
              disabled={!selectedPlan || upgrading}
              className="bg-green-500 hover:bg-green-400 disabled:opacity-40 disabled:cursor-not-allowed text-black text-sm font-medium px-5 py-2 rounded-lg transition-colors"
            >
              {upgrading ? t.srv.applying : selectedPlan ? t.srv.backToFree : t.srv.choosePlan}
            </button>
          )}
        </div>

        {/* Version Minecraft (jeux MC uniquement) */}
        {server.game === "minecraft" && (
        <div className="border border-zinc-800 bg-zinc-900 rounded-xl p-6">
          <div className="flex items-center gap-2 font-semibold mb-1">
            <RefreshCw className="w-5 h-5 text-green-400" /> {t.srv.mcVersion}
          </div>
          <p className="text-sm text-zinc-500 mb-4">
            {t.srv.currentVersion} <span className="text-zinc-300">{server.version === "LATEST" ? t.srv.latestParens : server.version}</span>.
            {" "}{t.srv.vChangePre}<span className="text-zinc-400">{t.srv.vChangeBold}</span>{t.srv.vChangePost}
          </p>
          <div className="flex items-center gap-2">
            <select
              value={selectedVersion}
              onChange={(e) => setSelectedVersion(e.target.value)}
              className="select-dark cursor-pointer bg-zinc-950 border border-zinc-700 rounded-lg pl-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-green-500"
            >
              <option value="">{t.srv.chooseVersion}</option>
              {VERSIONS.filter((v) => v !== server.version).map((v) => (
                <option key={v} value={v}>{v === "LATEST" ? t.dash.versionLatest : v}</option>
              ))}
            </select>
            <button
              onClick={changeVersion}
              disabled={!selectedVersion || changingVersion}
              className="bg-green-500 hover:bg-green-400 disabled:opacity-40 disabled:cursor-not-allowed text-black text-sm font-medium px-5 py-2 rounded-lg transition-colors"
            >
              {changingVersion ? t.srv.applying : t.srv.apply}
            </button>
          </div>
        </div>
        )}

        {/* Console */}
        <div className="border border-zinc-800 bg-zinc-900 rounded-xl p-6">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2 font-semibold">
              <Terminal className="w-5 h-5 text-green-400" /> {t.srv.console}
            </div>
            {running && players && (
              <span className="flex items-center gap-1.5 text-xs text-zinc-400">
                <Users className="w-3.5 h-3.5" />
                {players.online}/{players.max}
                {players.players.length > 0 && <span className="text-zinc-500">· {players.players.join(", ")}</span>}
              </span>
            )}
          </div>
          {running ? (
            <>
              <pre ref={logRef} className="h-72 overflow-auto bg-black/60 rounded-lg p-3 text-xs font-mono text-zinc-300 whitespace-pre-wrap break-words">
                {logs || t.srv.logsLoading}
              </pre>
              {/* Saisie de commande : RCON, Minecraft uniquement */}
              {server.game === "minecraft" && (
              <form onSubmit={sendCommand} className="mt-3 flex gap-2">
                <span className="flex items-center text-zinc-500 font-mono text-sm">/</span>
                <input
                  value={command}
                  onChange={(e) => setCommand(e.target.value)}
                  placeholder={t.srv.cmdPlaceholder}
                  className="flex-1 bg-black/40 border border-zinc-800 rounded-lg px-3 py-2 text-sm font-mono text-zinc-100 focus:outline-none focus:border-green-500"
                />
                <button type="submit" disabled={sending || !command.trim()}
                  className="flex items-center gap-2 bg-green-500 hover:bg-green-400 disabled:opacity-40 text-black font-medium px-4 py-2 rounded-lg text-sm transition-colors">
                  <SendHorizontal className="w-4 h-4" /> {t.srv.send}
                </button>
              </form>
              )}
            </>
          ) : (
            <p className="text-sm text-zinc-500">{t.srv.consoleOffline}</p>
          )}
        </div>

        {/* Gestion des joueurs : RCON, Minecraft uniquement */}
        {server.game === "minecraft" && (
        <div className="border border-zinc-800 bg-zinc-900 rounded-xl p-6">
          <div className="flex items-center gap-2 font-semibold mb-4">
            <Shield className="w-5 h-5 text-green-400" /> {t.srv.playersTitle}
          </div>
          {running ? (
            <div className="flex flex-col gap-5">
              {/* Connectés */}
              <div>
                <div className="text-xs uppercase tracking-wide text-zinc-500 mb-2">{t.srv.connected} ({players?.online ?? 0})</div>
                {players && players.players.length > 0 ? (
                  <div className="flex flex-col gap-2">
                    {players.players.map((p) => (
                      <div key={p} className="flex items-center justify-between gap-2 rounded-lg bg-black/30 px-3 py-2">
                        <span className="font-medium text-sm truncate">{p}</span>
                        <div className="flex gap-1.5 shrink-0">
                          <button onClick={() => doPlayerAction("op", p)} disabled={actBusy} title={t.srv.giveOp}
                            className="flex items-center gap-1 text-xs px-2 py-1 rounded-md bg-zinc-800 hover:bg-green-900/60 text-green-400 disabled:opacity-40 transition-colors"><Shield className="w-3.5 h-3.5" />OP</button>
                          <button onClick={() => doPlayerAction("kick", p)} disabled={actBusy} title={t.srv.kick}
                            className="flex items-center gap-1 text-xs px-2 py-1 rounded-md bg-zinc-800 hover:bg-yellow-900/60 text-yellow-400 disabled:opacity-40 transition-colors"><UserMinus className="w-3.5 h-3.5" />Kick</button>
                          <button onClick={() => doPlayerAction("ban", p)} disabled={actBusy} title={t.srv.ban}
                            className="flex items-center gap-1 text-xs px-2 py-1 rounded-md bg-zinc-800 hover:bg-red-900/60 text-red-400 disabled:opacity-40 transition-colors"><Ban className="w-3.5 h-3.5" />Ban</button>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-zinc-600">{t.srv.noPlayersOnline}</p>
                )}
              </div>

              {/* Action par pseudo */}
              <div>
                <div className="text-xs uppercase tracking-wide text-zinc-500 mb-2">{t.srv.addManageByName}</div>
                <form onSubmit={(e) => { e.preventDefault(); doPlayerAction(playerAct, newPlayer); }} className="flex flex-wrap gap-2">
                  <input
                    value={newPlayer}
                    onChange={(e) => setNewPlayer(e.target.value)}
                    placeholder={t.srv.mcUsername}
                    className="flex-1 min-w-[140px] bg-black/40 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-green-500"
                  />
                  <select value={playerAct} onChange={(e) => setPlayerAct(e.target.value)}
                    className="select-dark cursor-pointer bg-zinc-950 border border-zinc-700 rounded-lg pl-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-green-500">
                    <option value="whitelist_add">Whitelist +</option>
                    <option value="whitelist_remove">Whitelist −</option>
                    <option value="op">{t.srv.giveOp}</option>
                    <option value="deop">{t.srv.deop}</option>
                    <option value="ban">{t.srv.ban}</option>
                    <option value="pardon">{t.srv.pardon}</option>
                  </select>
                  <button type="submit" disabled={actBusy || !newPlayer.trim()}
                    className="flex items-center gap-2 bg-green-500 hover:bg-green-400 disabled:opacity-40 text-black font-medium px-4 py-2 rounded-lg text-sm transition-colors">
                    <UserPlus className="w-4 h-4" /> {t.srv.apply}
                  </button>
                </form>
              </div>

              {/* Listes */}
              <div className="grid sm:grid-cols-2 gap-4">
                <div>
                  <div className="text-xs uppercase tracking-wide text-zinc-500 mb-2">{t.srv.whitelist} ({lists.whitelist.length})</div>
                  {lists.whitelist.length ? (
                    <div className="flex flex-wrap gap-1.5">
                      {lists.whitelist.map((p) => (
                        <button key={p} onClick={() => doPlayerAction("whitelist_remove", p)} title={t.srv.removeFromWhitelist}
                          className="group flex items-center gap-1 text-xs px-2 py-1 rounded-md bg-green-500/10 text-green-300 hover:bg-red-900/50 hover:text-red-300 transition-colors">
                          {p} <span className="opacity-50 group-hover:opacity-100">×</span>
                        </button>
                      ))}
                    </div>
                  ) : <p className="text-sm text-zinc-600">{t.srv.empty}</p>}
                </div>
                <div>
                  <div className="text-xs uppercase tracking-wide text-zinc-500 mb-2">{t.srv.banned} ({lists.banned.length})</div>
                  {lists.banned.length ? (
                    <div className="flex flex-wrap gap-1.5">
                      {lists.banned.map((p) => (
                        <button key={p} onClick={() => doPlayerAction("pardon", p)} title={t.srv.pardon}
                          className="group flex items-center gap-1 text-xs px-2 py-1 rounded-md bg-red-500/10 text-red-300 hover:bg-green-900/50 hover:text-green-300 transition-colors">
                          {p} <span className="opacity-50 group-hover:opacity-100">×</span>
                        </button>
                      ))}
                    </div>
                  ) : <p className="text-sm text-zinc-600">{t.srv.none}</p>}
                </div>
              </div>
            </div>
          ) : (
            <p className="text-sm text-zinc-500">{t.srv.playersOffline}</p>
          )}
        </div>
        )}

        {/* Fichiers */}
        <div className="border border-zinc-800 bg-zinc-900 rounded-xl p-6">
          <div className="flex flex-wrap items-center justify-between gap-3 mb-4">
            <div className="flex items-center gap-2 font-semibold">
              <Folder className="w-5 h-5 text-green-400" /> {t.srv.files}
            </div>
            {running && (
              <div className="flex items-center gap-2">
                <button onClick={mkdir} className="flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-200 transition-colors"><FolderPlus className="w-4 h-4" /> {t.srv.folder}</button>
                <label className={`flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-lg bg-green-500 hover:bg-green-400 text-black font-medium cursor-pointer transition-colors ${uploading ? "opacity-50 pointer-events-none" : ""}`}>
                  <Upload className="w-4 h-4" /> {uploading ? t.srv.uploading : t.srv.upload}
                  <input type="file" className="hidden" onChange={(e) => { const f = e.target.files?.[0]; if (f) uploadFile(f); e.currentTarget.value = ""; }} />
                </label>
              </div>
            )}
          </div>

          {running ? (
            <>
              {/* Fil d'Ariane */}
              <div className="flex items-center flex-wrap gap-1 text-sm mb-3">
                {(() => {
                  const parts = filesPath.split("/").filter(Boolean);
                  return (
                    <>
                      <button onClick={() => setFilesPath("/")} className="text-green-400 hover:underline">/data</button>
                      {parts.map((seg, i) => (
                        <span key={i} className="flex items-center gap-1">
                          <ChevronRight className="w-3.5 h-3.5 text-zinc-600" />
                          <button onClick={() => setFilesPath("/" + parts.slice(0, i + 1).join("/"))} className="text-zinc-300 hover:underline">{seg}</button>
                        </span>
                      ))}
                    </>
                  );
                })()}
              </div>

              {/* Liste */}
              <div className="rounded-lg border border-zinc-800 divide-y divide-zinc-800 overflow-hidden">
                {filesPath !== "/" && (
                  <button onClick={() => setFilesPath(filesPath.split("/").slice(0, -1).join("/") || "/")} className="w-full flex items-center gap-2 px-3 py-2 text-sm text-zinc-400 hover:bg-zinc-800/50 transition-colors">
                    <Folder className="w-4 h-4" /> ..
                  </button>
                )}
                {filesBusy && entries.length === 0 ? (
                  <div className="px-3 py-4 text-sm text-zinc-600">{t.srv.loadingShort}</div>
                ) : entries.length === 0 ? (
                  <div className="px-3 py-4 text-sm text-zinc-600">{t.srv.emptyFolder}</div>
                ) : entries.map((e) => (
                  <div key={e.name} className="flex items-center gap-2 px-3 py-2 text-sm hover:bg-zinc-800/40 transition-colors group">
                    <button onClick={() => e.is_dir ? setFilesPath(joinPath(filesPath, e.name)) : openFile(e.name)} className="flex items-center gap-2 min-w-0 flex-1 text-left">
                      {e.is_dir ? <Folder className="w-4 h-4 text-blue-400 shrink-0" /> : <FileText className="w-4 h-4 text-zinc-500 shrink-0" />}
                      <span className="truncate">{e.name}</span>
                    </button>
                    {!e.is_dir && <span className="text-xs text-zinc-600 shrink-0 hidden sm:block">{fmtBytes(e.size)}</span>}
                    {!e.is_dir && <button onClick={() => downloadEntry(e.name)} title={t.srv.download} className="p-1 rounded text-zinc-500 hover:text-green-400 shrink-0"><Download className="w-4 h-4" /></button>}
                    <button onClick={() => deleteEntry(e)} title={t.srv.delete} className="p-1 rounded text-zinc-500 hover:text-red-400 shrink-0"><Trash2 className="w-4 h-4" /></button>
                  </div>
                ))}
              </div>
              <p className="text-xs text-zinc-600 mt-2">{t.srv.filesHintPre}<code className="text-zinc-400">world</code>{t.srv.filesHintPost}</p>
            </>
          ) : (
            <p className="text-sm text-zinc-500">{t.srv.filesOffline}</p>
          )}
        </div>

        {/* Éditeur de fichier (overlay) */}
        {editPath && (
          <div className="fixed inset-0 z-50 bg-black/70 flex items-center justify-center p-4" onClick={() => setEditPath(null)}>
            <div className="bg-zinc-900 border border-zinc-700 rounded-xl w-full max-w-3xl max-h-[85vh] flex flex-col" onClick={(e) => e.stopPropagation()}>
              <div className="flex items-center justify-between px-4 py-3 border-b border-zinc-800">
                <div className="flex items-center gap-2 font-mono text-sm truncate"><FileText className="w-4 h-4 text-green-400 shrink-0" />{editPath}</div>
                <button onClick={() => setEditPath(null)} className="text-zinc-400 hover:text-zinc-100"><X className="w-5 h-5" /></button>
              </div>
              {editTrunc && <div className="px-4 py-2 text-xs text-yellow-400 bg-yellow-500/10">{t.srv.truncated}</div>}
              <textarea
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                spellCheck={false}
                className="flex-1 min-h-[50vh] bg-black/60 text-zinc-200 font-mono text-xs p-4 resize-none focus:outline-none"
              />
              <div className="flex justify-end gap-2 px-4 py-3 border-t border-zinc-800">
                <button onClick={() => setEditPath(null)} className="px-4 py-2 rounded-lg text-sm bg-zinc-800 hover:bg-zinc-700 text-zinc-200">{t.srv.cancel}</button>
                <button onClick={saveFile} disabled={savingFile || editTrunc} className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm bg-green-500 hover:bg-green-400 disabled:opacity-40 text-black font-medium">
                  <Save className="w-4 h-4" /> {savingFile ? t.srv.saving : t.srv.save}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Sections à venir */}
        <div className="border border-dashed border-zinc-800 rounded-xl p-6">
          <div className="font-semibold mb-1">{t.srv.modsTitle}</div>
          <div className="text-sm text-zinc-500">{t.srv.modsDesc}</div>
        </div>

        {/* Danger zone */}
        <div className="border border-red-900 rounded-xl p-6 mt-4">
          <div className="font-semibold text-red-400 mb-2">{t.srv.dangerZone}</div>
          <p className="text-sm text-zinc-400 mb-4">{t.srv.deleteDesc}</p>
          <button onClick={deleteServer} className="bg-red-900 hover:bg-red-800 text-red-300 text-sm font-medium px-4 py-2 rounded-lg transition-colors">
            {t.srv.deleteBtn}
          </button>
        </div>
      </div>
    </div>
  );
}
