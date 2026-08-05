"use client";

import { useState } from "react";
import { Check, Swords, Users, Shield, Globe, Zap, RefreshCw, Lock } from "lucide-react";
import {
  useI18n, PLANS, PERIODS, priceFor, fmtMoney, currencyCode, type Period,
} from "@/lib/i18n";
import { getGame } from "@/lib/games";
import { SiteNav, SiteFooter } from "@/components/site-chrome";
import { PingBadge } from "@/components/ping-badge";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
const MOD_SITE = "https://calradiacoop.vbt-prog.com";
const PORTFOLIO = "https://vbt-prog.com";

const jsonLd = {
  "@context": "https://schema.org",
  "@type": "Product",
  name: "Playrena — Serveurs Bannerlord Calradia-Coop",
  description:
    "Serveurs dédiés pour Calradia-Coop, le mod coop de campagne pour Mount & Blade II: Bannerlord. Monde persistant 24/7, sans PC hôte, sans configuration réseau.",
  brand: { "@type": "Brand", name: "Playrena" },
  offers: {
    "@type": "AggregateOffer", priceCurrency: "USD", lowPrice: "0", highPrice: "0", offerCount: 1,
  },
};

// Icônes de la grille de features (même ordre que t.calradia.features).
const FEATURE_ICONS = [Globe, Swords, Users, Zap, Shield, RefreshCw];

const SHOTS = [
  "/games/calradia/calradia-1.jpg",
  "/games/calradia/calradia-2.jpg",
  "/games/calradia/calradia-3.jpg",
  "/games/calradia/calradia-4.jpg",
];

export default function CalradiaCoopHosting() {
  const { t, lang } = useI18n();
  const [period, setPeriod] = useState<Period>("monthly");
  const game = getGame("calradia-coop");
  const floor = game?.minRamGb ?? 0;
  const freeAtFloor = game?.freeAtFloor ?? false;
  // Jeu offert (freeAtFloor) → seul le plan gratuit, au plancher. Sinon grille complète.
  const visiblePlans = freeAtFloor ? PLANS.filter((p) => p.id === "free") : PLANS;

  return (
    <main className="flex flex-col min-h-screen">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />

      <SiteNav />

      {/* Hero */}
      <section className="relative overflow-hidden">
        <div className="pointer-events-none absolute -top-40 -left-40 w-[40rem] h-[40rem] rounded-full bg-rose-500/15 blur-3xl" />
        <div className="pointer-events-none absolute top-20 -right-40 w-[40rem] h-[40rem] rounded-full bg-red-500/15 blur-3xl" />

        <div className="relative max-w-3xl mx-auto px-6 py-24 flex flex-col gap-6 items-start">
          <a href="/games" className="text-sm text-zinc-400 hover:text-zinc-200 transition-colors">{t.calradia.back}</a>

          <div className="inline-flex items-center gap-2 bg-rose-950/60 text-rose-300 text-sm px-3 py-1 rounded-full border border-rose-800">
            <Swords className="w-4 h-4" /> {t.calradia.badge}
          </div>

          <h1 className="text-5xl sm:text-6xl font-extrabold tracking-tight leading-[1.05] uppercase">
            {t.calradia.title1}<br />
            <span className="bg-gradient-to-r from-rose-400 to-red-400 bg-clip-text text-transparent">{t.calradia.title2}</span>
          </h1>

          <p className="text-zinc-400 text-lg max-w-2xl">{t.calradia.subtitle}</p>

          <ul className="flex flex-col gap-2 mt-2">
            {t.calradia.bullets.map((b) => (
              <li key={b} className="flex items-center gap-2 text-zinc-300">
                <Check className="w-5 h-5 text-rose-400 shrink-0" /> {b}
              </li>
            ))}
          </ul>

          <div className="mt-4 flex flex-wrap gap-3">
            <a href={`${API}/auth/discord`} className="inline-flex bg-rose-500 hover:bg-rose-400 text-black font-semibold px-6 py-3 rounded-lg transition-colors">
              {t.calradia.ctaStart}
            </a>
            <a href={MOD_SITE} className="border border-zinc-600 hover:border-zinc-400 bg-zinc-900/50 text-zinc-200 px-6 py-3 rounded-lg transition-colors">
              {t.calradia.ctaSite}
            </a>
          </div>

          <p className="flex items-start gap-2 text-sm text-zinc-400 max-w-xl rounded-lg bg-rose-500/5 border border-rose-500/20 px-3 py-2">
            <Lock className="w-4 h-4 shrink-0 mt-0.5 text-rose-400" /> {t.calradia.accessNote}
          </p>

          <PingBadge className="mt-2" />
        </div>
      </section>

      {/* C'est quoi le mod — description + grille de features */}
      <section className="px-6 pb-20 max-w-6xl mx-auto w-full">
        <h2 className="text-3xl font-bold text-center mb-3">{t.calradia.aboutTitle}</h2>
        <p className="text-zinc-400 text-center max-w-3xl mx-auto mb-10">{t.calradia.aboutText}</p>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {t.calradia.features.map((f, i) => {
            const Icon = FEATURE_ICONS[i % FEATURE_ICONS.length];
            return (
              <div key={f.title} className="rounded-xl border border-zinc-800 bg-zinc-900 p-5 flex flex-col gap-2">
                <Icon className="w-6 h-6 text-rose-400" />
                <div className="font-semibold">{f.title}</div>
                <p className="text-sm text-zinc-400">{f.desc}</p>
              </div>
            );
          })}
        </div>
      </section>

      {/* Galerie — captures du jeu */}
      <section className="px-6 pb-20 max-w-6xl mx-auto w-full">
        <h2 className="text-3xl font-bold text-center mb-8">{t.calradia.galleryTitle}</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {SHOTS.map((src, i) => (
            <img
              key={src}
              src={src}
              alt={`Mount & Blade II: Bannerlord — ${i + 1}`}
              loading="lazy"
              className="rounded-xl border border-zinc-800 w-full aspect-video object-cover"
            />
          ))}
        </div>
        <p className="text-xs text-zinc-600 text-center mt-4">{t.calradia.disclaimer}</p>
      </section>

      {/* Pourquoi louer ici */}
      <section className="px-6 pb-20 max-w-6xl mx-auto w-full">
        <h2 className="text-3xl font-bold text-center mb-8">{t.calradia.whyTitle}</h2>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          {t.calradia.whyItems.map((f) => (
            <div key={f.title} className="rounded-xl border border-rose-500/20 bg-rose-950/10 p-5 flex flex-col gap-2">
              <div className="font-semibold text-rose-300">{f.title}</div>
              <p className="text-sm text-zinc-400">{f.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Plans (partagés) */}
      <section id="plans" className="px-6 pb-20 max-w-6xl mx-auto w-full">
        <h2 className="text-3xl font-bold text-center mb-2">{t.pricing.heading}</h2>
        <p className="text-zinc-500 text-center mb-2">{t.pricing.sub}</p>
        {floor > 0 && (
          <p className="text-rose-400/90 text-sm text-center mb-8">{t.pricing.minIncluded.replace("{n}", String(floor))}</p>
        )}

        {/* Toggle période */}
        <div className="flex justify-center mb-10">
          <div className="inline-flex items-center gap-1 rounded-xl bg-zinc-900 border border-zinc-800 p-1">
            {(Object.keys(PERIODS) as Period[]).map((p) => {
              const label = p === "monthly" ? t.pricing.monthly : p === "quarterly" ? t.pricing.quarterly : t.pricing.annually;
              const disc = PERIODS[p].discount;
              return (
                <button
                  key={p}
                  onClick={() => setPeriod(p)}
                  className={`relative px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                    period === p ? "bg-rose-500 text-black" : "text-zinc-400 hover:text-zinc-200"
                  }`}
                >
                  {label}
                  {disc > 0 && (
                    <span className={`ml-1.5 text-[10px] font-bold px-1.5 py-0.5 rounded ${period === p ? "bg-black/20" : "bg-rose-500/15 text-rose-400"}`}>
                      -{Math.round(disc * 100)}%
                    </span>
                  )}
                </button>
              );
            })}
          </div>
        </div>

        <div className={freeAtFloor ? "max-w-sm mx-auto" : "grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4"}>
          {visiblePlans.map((plan) => {
            const { monthly, total, discount, originalMonthly, promoActive } = priceFor(plan, lang, period);
            const isFree = plan.id === "free";
            return (
              <div
                key={plan.id}
                className={`relative rounded-xl border p-5 flex flex-col gap-3 ${
                  plan.popular ? "border-rose-500 bg-rose-950/20 shadow-lg shadow-rose-500/10" : "border-zinc-800 bg-zinc-900"
                }`}
              >
                {plan.popular && (
                  <span className="absolute -top-3 left-1/2 -translate-x-1/2 text-[10px] font-bold uppercase tracking-wide bg-rose-500 text-black px-2 py-0.5 rounded-full">
                    {t.pricing.popular}
                  </span>
                )}
                <div>
                  <div className="font-bold text-lg">{t.planNames[plan.id]}</div>
                  <div className="mt-1 flex items-end gap-1.5 flex-wrap">
                    {!isFree && promoActive && (
                      <span className="text-zinc-500 text-lg line-through mb-0.5">{fmtMoney(originalMonthly, lang)}</span>
                    )}
                    <span className="text-3xl font-extrabold">{isFree ? t.pricing.free : fmtMoney(monthly, lang)}</span>
                    {!isFree && <span className="text-zinc-500 text-sm mb-1">{t.pricing.perMonth}</span>}
                  </div>
                  {!isFree && period !== "monthly" && (
                    <div className="text-xs text-zinc-500 mt-1">
                      {(period === "quarterly" ? t.pricing.billedQuarterly : t.pricing.billedAnnually).replace("{x}", `${fmtMoney(total, lang)} ${currencyCode(lang)}`)}
                    </div>
                  )}
                  {!isFree && discount > 0 && (
                    <div className="text-xs text-rose-400 font-medium mt-0.5">{t.pricing.save} {Math.round(discount * 100)}%</div>
                  )}
                </div>
                <ul className="text-sm text-zinc-400 flex flex-col gap-1.5 mt-1">
                  <li className="flex items-center gap-2"><Check className="w-3.5 h-3.5 text-rose-400" /> {Math.max(parseInt(plan.ram), floor)} GB {t.pricing.ramLabel}</li>
                  <li className="flex items-center gap-2"><Check className="w-3.5 h-3.5 text-rose-400" /> {plan.cpu} {plan.cpu > 1 ? "cœurs" : "cœur"}</li>
                  <li className="flex items-center gap-2"><Check className="w-3.5 h-3.5 text-rose-400" /> {plan.slots === 0 ? t.pricing.unlimited : `${Math.min(plan.slots, 8)} ${t.pricing.players}`}</li>
                </ul>
                <a
                  href={`${API}/auth/discord`}
                  className={`mt-auto text-center text-sm font-medium py-2 rounded-lg transition-colors ${
                    plan.popular ? "bg-rose-500 hover:bg-rose-400 text-black" : "bg-zinc-800 hover:bg-zinc-700 text-zinc-100"
                  }`}
                >
                  {isFree ? t.pricing.ctaFree : t.pricing.ctaChoose}
                </a>
              </div>
            );
          })}
        </div>
      </section>

      {/* Crédits — développeur (portfolio) + site du mod */}
      <section className="px-6 pb-24 max-w-3xl mx-auto w-full text-center">
        <p className="text-zinc-400">
          {t.calradia.modBy}{" "}
          <a href={PORTFOLIO} className="text-rose-400 hover:text-rose-300 underline underline-offset-4">{t.calradia.modByLink}</a>
          {" · "}
          <a href={MOD_SITE} className="text-rose-400 hover:text-rose-300 underline underline-offset-4">{t.calradia.modSite}</a>
        </p>
      </section>

      <SiteFooter />
    </main>
  );
}
