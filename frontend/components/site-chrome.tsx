"use client";

import { useState, useEffect } from "react";
import { Server } from "lucide-react";
import { useI18n, type Lang } from "@/lib/i18n";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// AuthButton : affiche « Tableau de bord » si l'utilisateur est connecté (cookie de
// session valide, vérifié via /api/v1/me), sinon « Se connecter » (OAuth Discord).
// → l'utilisateur reste connecté en naviguant et a toujours un lien vers son espace.
export function AuthButton() {
  const { t } = useI18n();
  const [authed, setAuthed] = useState<boolean | null>(null);
  useEffect(() => {
    fetch(`${API}/api/v1/user/me`, { credentials: "include" })
      .then((r) => setAuthed(r.ok))
      .catch(() => setAuthed(false));
  }, []);
  if (authed === null) return <span className="inline-block w-28 h-9" aria-hidden />;
  return authed ? (
    <a href="/dashboard" className="bg-green-600 hover:bg-green-500 text-white px-4 py-2 rounded-lg transition-colors font-medium">
      {t.nav.dashboard}
    </a>
  ) : (
    <a href={`${API}/auth/discord`} className="bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-lg transition-colors font-medium">
      {t.nav.login}
    </a>
  );
}
const DISCORD_INVITE = "https://discord.gg/3knNHXqpNG";

type IconProps = { className?: string };
const SVG = ({ className = "", d }: IconProps & { d: string }) => (
  <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden><path d={d} /></svg>
);
const DiscordIcon = (p: IconProps) => <SVG {...p} d="M20.317 4.369A19.79 19.79 0 0 0 15.885 3c-.21.375-.444.88-.608 1.283a18.27 18.27 0 0 0-5.487 0A12.6 12.6 0 0 0 9.18 3 19.74 19.74 0 0 0 4.745 4.37C1.6 9.057.743 13.63 1.17 18.138a19.9 19.9 0 0 0 6.073 3.058c.487-.667.92-1.377 1.292-2.122a12.9 12.9 0 0 1-2.035-.978c.171-.125.338-.255.5-.39a14.2 14.2 0 0 0 12 0c.164.135.331.265.5.39-.65.384-1.334.712-2.037.978.373.745.805 1.455 1.292 2.122a19.86 19.86 0 0 0 6.075-3.058c.5-5.224-.838-9.756-3.715-13.77ZM8.02 15.331c-1.182 0-2.157-1.086-2.157-2.42 0-1.333.955-2.42 2.157-2.42 1.21 0 2.176 1.096 2.157 2.42 0 1.334-.955 2.42-2.157 2.42Zm7.96 0c-1.183 0-2.157-1.086-2.157-2.42 0-1.333.955-2.42 2.157-2.42 1.21 0 2.176 1.096 2.157 2.42 0 1.334-.946 2.42-2.157 2.42Z" />;
const XIcon = (p: IconProps) => <SVG {...p} d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24h-6.66l-5.214-6.817-5.97 6.817H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231 5.45-6.231Zm-1.161 17.52h1.833L7.084 4.126H5.117L17.083 19.77Z" />;
const YoutubeIcon = (p: IconProps) => <SVG {...p} d="M23.5 6.2a3 3 0 0 0-2.1-2.1C19.5 3.6 12 3.6 12 3.6s-7.5 0-9.4.5A3 3 0 0 0 .5 6.2 31 31 0 0 0 0 12a31 31 0 0 0 .5 5.8 3 3 0 0 0 2.1 2.1c1.9.5 9.4.5 9.4.5s7.5 0 9.4-.5a3 3 0 0 0 2.1-2.1A31 31 0 0 0 24 12a31 31 0 0 0-.5-5.8ZM9.6 15.6V8.4l6.3 3.6-6.3 3.6Z" />;
const InstagramIcon = (p: IconProps) => <SVG {...p} d="M12 2.16c3.2 0 3.58.01 4.85.07 1.17.05 1.8.25 2.23.41.56.22.96.48 1.38.9.42.42.68.82.9 1.38.16.43.36 1.06.41 2.23.06 1.27.07 1.65.07 4.85s-.01 3.58-.07 4.85c-.05 1.17-.25 1.8-.41 2.23-.22.56-.48.96-.9 1.38-.42.42-.82.68-1.38.9-.43.16-1.06.36-2.23.41-1.27.06-1.65.07-4.85.07s-3.58-.01-4.85-.07c-1.17-.05-1.8-.25-2.23-.41a3.7 3.7 0 0 1-1.38-.9 3.7 3.7 0 0 1-.9-1.38c-.16-.43-.36-1.06-.41-2.23C2.17 15.58 2.16 15.2 2.16 12s.01-3.58.07-4.85c.05-1.17.25-1.8.41-2.23.22-.56.48-.96.9-1.38.42-.42.82-.68 1.38-.9.43-.16 1.06-.36 2.23-.41C8.42 2.17 8.8 2.16 12 2.16Zm0 4.92a4.92 4.92 0 1 0 0 9.84 4.92 4.92 0 0 0 0-9.84Zm0 8.12a3.2 3.2 0 1 1 0-6.4 3.2 3.2 0 0 1 0 6.4Zm6.27-8.31a1.15 1.15 0 1 1-2.3 0 1.15 1.15 0 0 1 2.3 0Z" />;
const TwitchIcon = (p: IconProps) => <SVG {...p} d="M4.265 0 1.5 4.143v15.43h5.18V24h2.765l2.765-4.427h4.15L21.75 14.4V0H4.265Zm15.64 13.38-3.318 3.318h-4.15l-2.765 2.765v-2.765H5.18V1.66h14.725v11.72ZM16.68 5.024v6.083h-1.844V5.024h1.844Zm-4.98 0v6.083H9.857V5.024H11.7Z" />;

export function LanguageSwitcher() {
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

export function PromoBanner() {
  const { t, promo } = useI18n();
  if (!promo) return null;
  return (
    <div className="bg-gradient-to-r from-green-600 to-emerald-500 text-black text-center text-sm font-semibold px-4 py-2">
      {t.promoBanner.replace("{pct}", String(promo))}
    </div>
  );
}

// FeedbackWidget : bouton flottant + sondage (canal d'acquisition, cas d'usage, jeu,
// note, commentaire) → POST /api/v1/feedback. Résultats agrégés dans /admin.
export function FeedbackWidget() {
  const { t } = useI18n();
  const s = t.survey;
  const [open, setOpen] = useState(false);
  const [sent, setSent] = useState(false);
  const [sending, setSending] = useState(false);
  const [rating, setRating] = useState(0);
  const [source, setSource] = useState("");
  const [usecase, setUsecase] = useState("");
  const [game, setGame] = useState("");
  const [message, setMessage] = useState("");
  const [email, setEmail] = useState("");

  const chip = (val: string, label: string, sel: string, set: (v: string) => void) => (
    <button key={val} type="button" onClick={() => set(sel === val ? "" : val)}
      className={`text-xs px-2.5 py-1 rounded-full border transition-colors ${sel === val ? "border-green-500 bg-green-500/15 text-green-300" : "border-zinc-700 text-zinc-400 hover:text-zinc-200"}`}>{label}</button>
  );

  async function submit() {
    if (sending) return;
    setSending(true);
    try {
      await fetch(`${API}/api/v1/feedback`, {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ rating, source, usecase, game, message, email, page: typeof location !== "undefined" ? location.pathname : "" }),
      });
      setSent(true);
      setTimeout(() => setOpen(false), 2500);
    } catch { /* ignore */ } finally { setSending(false); }
  }

  return (
    <>
      <button onClick={() => { setOpen(true); setSent(false); }}
        className="fixed bottom-4 right-4 z-40 bg-green-500 hover:bg-green-400 text-black font-semibold text-sm px-4 py-2.5 rounded-full shadow-lg shadow-green-500/20">
        💬 {s.button}
      </button>
      {open && (
        <div className="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/60 p-4" onClick={() => setOpen(false)}>
          <div className="w-full max-w-md bg-zinc-900 border border-zinc-700 rounded-2xl p-6 max-h-[90vh] overflow-auto" onClick={(e) => e.stopPropagation()}>
            {sent ? (
              <div className="text-center py-6">
                <div className="text-lg font-semibold text-green-400 mb-3">{s.thanks}</div>
                <button onClick={() => setOpen(false)} className="text-sm text-zinc-400 hover:text-zinc-200">{s.close}</button>
              </div>
            ) : (
              <>
                <div className="font-bold text-lg">{s.title}</div>
                <p className="text-sm text-zinc-400 mb-4">{s.intro}</p>
                <div className="flex flex-col gap-4">
                  <div>
                    <div className="text-sm text-zinc-300 mb-1.5">{s.rating}</div>
                    <div className="flex gap-1">{[1, 2, 3, 4, 5].map((n) => (
                      <button key={n} type="button" onClick={() => setRating(n)} className={`text-2xl ${n <= rating ? "text-amber-400" : "text-zinc-600"} hover:text-amber-300`}>★</button>
                    ))}</div>
                  </div>
                  <div><div className="text-sm text-zinc-300 mb-1.5">{s.source}</div><div className="flex flex-wrap gap-1.5">{Object.entries(s.sources).map(([k, l]) => chip(k, l as string, source, setSource))}</div></div>
                  <div><div className="text-sm text-zinc-300 mb-1.5">{s.usecase}</div><div className="flex flex-wrap gap-1.5">{Object.entries(s.usecases).map(([k, l]) => chip(k, l as string, usecase, setUsecase))}</div></div>
                  <div><div className="text-sm text-zinc-300 mb-1.5">{s.game}</div><div className="flex flex-wrap gap-1.5">{Object.entries(s.games).map(([k, l]) => chip(k, l as string, game, setGame))}</div></div>
                  <textarea value={message} onChange={(e) => setMessage(e.target.value)} placeholder={s.message} rows={3} className="bg-black/40 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-green-500" />
                  <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder={s.email} className="bg-black/40 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-green-500" />
                  <div className="flex justify-end gap-2">
                    <button onClick={() => setOpen(false)} className="text-sm text-zinc-400 hover:text-zinc-200 px-3 py-2">{s.close}</button>
                    <button onClick={submit} disabled={sending} className="bg-green-500 hover:bg-green-400 disabled:opacity-50 text-black font-semibold text-sm px-5 py-2 rounded-lg">{sending ? s.sending : s.send}</button>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </>
  );
}

export function SiteNav() {
  const { t } = useI18n();
  return (
    <>
    <PromoBanner />
    <nav className="border-b border-zinc-800 px-6 py-4 flex items-center justify-between sticky top-0 z-30 bg-zinc-950/80 backdrop-blur">
      <a href="/games" className="flex items-center gap-2 font-bold text-lg">
        <Server className="w-5 h-5 text-green-400" />
        Playrena
      </a>
      <div className="flex items-center gap-4 sm:gap-6 text-sm text-zinc-400">
        <a href="/games" className="hidden sm:inline hover:text-zinc-100 transition-colors">{t.navGames}</a>
        <a href="/mods" className="hidden sm:inline hover:text-zinc-100 transition-colors">{t.nav.mods}</a>
        <a href="/blog" className="hidden sm:inline hover:text-zinc-100 transition-colors">{t.nav.blog}</a>
        <LanguageSwitcher />
        <AuthButton />
      </div>
    </nav>
    </>
  );
}

function FootCol({ title, links }: { title: string; links: { label: string; href: string }[] }) {
  return (
    <div>
      <h3 className="font-semibold text-zinc-100 mb-4">{title}</h3>
      <ul className="flex flex-col gap-2.5">
        {links.map((lnk) => (
          <li key={lnk.label}>
            <a href={lnk.href} className="text-sm text-zinc-400 hover:text-green-400 transition-colors">{lnk.label}</a>
          </li>
        ))}
      </ul>
    </div>
  );
}

export function SiteFooter() {
  const { t } = useI18n();
  const L = t.foot.l;
  return (
    <footer className="mt-auto">
      {/* Bandeau Discord */}
      <div className="bg-indigo-600">
        <div className="max-w-6xl mx-auto px-6 py-4 flex flex-col sm:flex-row items-center justify-between gap-3">
          <p className="font-semibold text-white text-center sm:text-left">{t.foot.discordCta}</p>
          <a href={DISCORD_INVITE} className="shrink-0 inline-flex items-center gap-2 bg-white text-indigo-700 font-semibold px-5 py-2.5 rounded-lg hover:bg-indigo-50 transition-colors">
            <DiscordIcon className="w-5 h-5" /> {t.foot.joinDiscord}
          </a>
        </div>
      </div>

      {/* Corps du footer */}
      <div className="border-t border-zinc-800 bg-zinc-950">
        <div className="max-w-6xl mx-auto px-6 py-12 grid gap-10 sm:grid-cols-2 lg:grid-cols-5">
          {/* Marque + réseaux */}
          <div className="lg:col-span-1 sm:col-span-2">
            <div className="flex items-center gap-2 font-bold text-lg mb-3">
              <Server className="w-5 h-5 text-green-400" /> Playrena
            </div>
            <p className="text-sm text-zinc-400 max-w-xs mb-4">{t.foot.brandDesc}</p>
            <div className="flex items-center gap-2">
              {[
                { icon: <DiscordIcon className="w-4 h-4" />, href: DISCORD_INVITE, label: "Discord" },
                { icon: <XIcon className="w-4 h-4" />, href: "#", label: "X" },
                { icon: <YoutubeIcon className="w-4 h-4" />, href: "#", label: "YouTube" },
                { icon: <InstagramIcon className="w-4 h-4" />, href: "#", label: "Instagram" },
                { icon: <TwitchIcon className="w-4 h-4" />, href: "#", label: "Twitch" },
              ].map((s) => (
                <a key={s.label} href={s.href} aria-label={s.label}
                  className="w-9 h-9 flex items-center justify-center rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-white hover:border-zinc-600 transition-colors">
                  {s.icon}
                </a>
              ))}
            </div>
          </div>

          <FootCol title={t.foot.games} links={[
            { label: "Minecraft", href: "/games/minecraft" },
            { label: "Hytale", href: "/games/hytale" },
            { label: "Rust", href: "/games" },
            { label: "ARK", href: "/games" },
            { label: L.allGames, href: "/games" },
          ]} />

          <FootCol title={t.foot.company} links={[
            { label: L.about, href: "#" },
            { label: L.blog, href: "/blog" },
            { label: L.affiliates, href: "#" },
            { label: L.partners, href: "#" },
          ]} />

          <FootCol title={t.foot.support} links={[
            { label: L.contact, href: "/games/minecraft#support" },
            { label: L.kb, href: "#" },
            { label: L.aup, href: "#" },
            { label: L.tos, href: "#" },
            { label: L.privacy, href: "#" },
            { label: L.sla, href: "#" },
          ]} />

          <FootCol title={t.foot.tech} links={[
            { label: L.panel, href: "/dashboard" },
            { label: L.hardware, href: "#" },
            { label: L.clientArea, href: "/dashboard" },
          ]} />
        </div>

        {/* Bas */}
        <div className="border-t border-zinc-900">
          <div className="max-w-6xl mx-auto px-6 py-5 text-sm text-zinc-600 text-center sm:text-left">
            © {new Date().getFullYear()} Playrena — {t.footer}. {t.foot.rights}
          </div>
        </div>
      </div>
    </footer>
  );
}
