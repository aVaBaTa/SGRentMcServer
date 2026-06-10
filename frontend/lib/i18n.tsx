"use client";

import { createContext, useContext, useEffect, useState } from "react";

export type Lang = "fr" | "en";

// ---- Plans & tarification ----
export interface PlanDef {
  id: string;
  usd: number; // prix mensuel de base en USD
  cad: number; // prix mensuel de base en CAD
  ram: string;
  cpu: number;
  slots: number; // 0 = illimité
  popular?: boolean;
}

export const PLANS: PlanDef[] = [
  { id: "free", usd: 0, cad: 0, ram: "1 GB", cpu: 1, slots: 5 },
  { id: "starter", usd: 3, cad: 4, ram: "2 GB", cpu: 1, slots: 20 },
  { id: "standard", usd: 7, cad: 9, ram: "4 GB", cpu: 2, slots: 50, popular: true },
  { id: "pro", usd: 14, cad: 19, ram: "8 GB", cpu: 4, slots: 100 },
  { id: "extreme", usd: 25, cad: 34, ram: "16 GB", cpu: 6, slots: 0 },
];

export type Period = "monthly" | "quarterly" | "annually";

export const PERIODS: Record<Period, { months: number; discount: number }> = {
  monthly: { months: 1, discount: 0 },
  quarterly: { months: 3, discount: 0.1 },
  annually: { months: 12, discount: 0.2 },
};

// Prix mensuel effectif (avec rabais) et total facturé pour la période.
// On garde les décimales pour que la réduction soit réellement visible
// (ex. 4 $ × -10 % = 3,60 $ et non arrondi à 4 $).
export function priceFor(plan: PlanDef, lang: Lang, period: Period) {
  const base = lang === "fr" ? plan.cad : plan.usd;
  const { months, discount } = PERIODS[period];
  const monthly = Math.round(base * (1 - discount) * 100) / 100;
  const total = Math.round(monthly * months * 100) / 100;
  return { monthly, total, months, discount };
}

// Formate un montant. Les prix entiers restent sans décimale (4 $),
// les prix réduits affichent 2 décimales (3,60 $ / 3.60$).
export function fmtMoney(amount: number, lang: Lang): string {
  const isInt = Number.isInteger(amount);
  const fr = isInt ? `${amount}` : amount.toFixed(2).replace(".", ",");
  const en = isInt ? `${amount}` : amount.toFixed(2);
  return lang === "fr" ? `${fr} $` : `$${en}`;
}

export const currencyCode = (lang: Lang) => (lang === "fr" ? "CAD" : "USD");

// ---- Dictionnaire ----
type Dict = typeof fr;
const fr = {
  langName: "Français (CA)",
  nav: { plans: "Plans", blog: "Blog", games: "Jeux — bientôt", login: "Se connecter" },
  rating: "Serveurs dédiés",
  ratingSuffix: "Fait au Québec",
  hero: {
    eyebrow: "Hébergement de jeux gratuit",
    title1: "Hébergement de serveurs",
    title2: "de jeux",
    subtitle:
      "Héberge ton propre serveur Minecraft avec Playrena. Des serveurs",
    subtitleBold: " fiables, entièrement personnalisables et prêts en quelques secondes,",
    subtitleEnd: " parfaits pour jouer entre amis ou lancer une communauté !",
    ctaStart: "Commencer gratuitement",
    ctaPlans: "Voir les plans",
    soon: "Bientôt : Satisfactory, Rust, ARK et plus.",
  },
  badges: {
    setup: "Installation instantanée",
    uptime: "99,9 % de disponibilité",
    support: "Support 24/7",
    mods: "Installateur de mods & plugins",
    ddos: "Protection DDoS",
    refund: "Remboursement 72 h",
  },
  features: [
    { title: "Démarrage instantané", desc: "Ton serveur est prêt en moins de 30 secondes. Paper, Fabric ou Forge." },
    { title: "Matériel dédié", desc: "Xeon 24 cœurs, 32 GB RAM par node. Limites CPU et RAM garanties par plan." },
    { title: "Données protégées", desc: "Tes mondes sont conservés même si tu supprimes ton serveur. Backups disponibles." },
  ],
  pricing: {
    heading: "Choisis ton plan",
    sub: "Change de plan à tout moment. Annule quand tu veux.",
    monthly: "Mensuel",
    quarterly: "Trimestriel",
    annually: "Annuel",
    save: "Économise",
    perMonth: "/mois",
    billedQuarterly: "facturé {x} / trimestre",
    billedAnnually: "facturé {x} / an",
    free: "Gratuit",
    popular: "Populaire",
    ramLabel: "RAM",
    players: "joueurs",
    unlimited: "Illimité",
    ctaFree: "Commencer",
    ctaChoose: "Choisir",
  },
  planNames: { free: "Gratuit", starter: "Starter", standard: "Standard", pro: "Pro", extreme: "Extreme" } as Record<string, string>,
  footer: "Hébergement de serveurs de jeux instantané",
};

const en: Dict = {
  langName: "English (US)",
  nav: { plans: "Plans", blog: "Blog", games: "Games — soon", login: "Sign in" },
  rating: "Dedicated servers",
  ratingSuffix: "Made in Canada",
  hero: {
    eyebrow: "Free Game Hosting",
    title1: "Game Server",
    title2: "Hosting",
    subtitle: "Easily host your own Minecraft server with Playrena. Enjoy",
    subtitleBold: " reliable, fully customizable and quick to set up servers,",
    subtitleEnd: " perfect for playing with friends or starting a community!",
    ctaStart: "Start for free",
    ctaPlans: "See plans",
    soon: "Coming soon: Satisfactory, Rust, ARK and more.",
  },
  badges: {
    setup: "Instant Setup",
    uptime: "99.9% Uptime",
    support: "24/7 Support",
    mods: "Instant Mod & Plugin Installer",
    ddos: "DDoS Protection",
    refund: "72h Self-Serve Refund",
  },
  features: [
    { title: "Instant start", desc: "Your server is ready in under 30 seconds. Paper, Fabric or Forge." },
    { title: "Dedicated hardware", desc: "24-core Xeon, 32 GB RAM per node. Guaranteed CPU and RAM per plan." },
    { title: "Protected data", desc: "Your worlds are kept even if you delete your server. Backups available." },
  ],
  pricing: {
    heading: "Choose your plan",
    sub: "Change plan anytime. Cancel whenever you want.",
    monthly: "Monthly",
    quarterly: "Quarterly",
    annually: "Annually",
    save: "Save",
    perMonth: "/mo",
    billedQuarterly: "billed {x} / quarter",
    billedAnnually: "billed {x} / year",
    free: "Free",
    popular: "Popular",
    ramLabel: "RAM",
    players: "players",
    unlimited: "Unlimited",
    ctaFree: "Start",
    ctaChoose: "Choose",
  },
  planNames: { free: "Free", starter: "Starter", standard: "Standard", pro: "Pro", extreme: "Extreme" },
  footer: "Instant game server hosting",
};

const dicts: Record<Lang, Dict> = { fr, en };

interface Ctx {
  lang: Lang;
  setLang: (l: Lang) => void;
  t: Dict;
}
const LangContext = createContext<Ctx>({ lang: "fr", setLang: () => {}, t: fr });

export function LanguageProvider({ children }: { children: React.ReactNode }) {
  const [lang, setLangState] = useState<Lang>("fr");

  useEffect(() => {
    const saved = (typeof localStorage !== "undefined" && localStorage.getItem("lang")) as Lang | null;
    if (saved === "fr" || saved === "en") setLangState(saved);
  }, []);

  const setLang = (l: Lang) => {
    setLangState(l);
    try { localStorage.setItem("lang", l); } catch {}
    if (typeof document !== "undefined") document.documentElement.lang = l;
  };

  return <LangContext.Provider value={{ lang, setLang, t: dicts[lang] }}>{children}</LangContext.Provider>;
}

export const useI18n = () => useContext(LangContext);
