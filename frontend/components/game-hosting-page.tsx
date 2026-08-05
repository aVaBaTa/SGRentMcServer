"use client";

// Page de vente générique d'un jeu « standard » (grille de plans partagée,
// pas de gratuité freeAtFloor ni de flux spécial) : hero + plans. Les pages
// historiques (Minecraft, Hytale, Satisfactory, Valheim, Calradia) gardent
// leur fichier dédié ; les nouveaux jeux passent par ce composant.

import { useState } from "react";
import { Check, type LucideIcon } from "lucide-react";
import {
  useI18n, PLANS, PERIODS, priceFor, fmtMoney, currencyCode, type Period,
} from "@/lib/i18n";
import { getGame, gameTheme } from "@/lib/games";
import { SiteNav, SiteFooter } from "@/components/site-chrome";
import { PingBadge } from "@/components/ping-badge";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export interface GameHostingTexts {
  badge: string;
  title1: string;
  title2: string;
  subtitle: string;
  bullets: string[];
  ctaStart: string;
  back: string;
}

export function GameHostingPage({
  gameId,
  icon: Icon,
  texts,
}: {
  gameId: string;
  icon: LucideIcon;
  texts: GameHostingTexts;
}) {
  const { t, lang } = useI18n();
  const [period, setPeriod] = useState<Period>("monthly");
  const game = getGame(gameId);
  const theme = gameTheme(gameId);
  const floor = game?.minRamGb ?? 0;

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "Product",
    name: `Playrena — ${texts.title1} ${texts.title2}`,
    description: texts.subtitle,
    brand: { "@type": "Brand", name: "Playrena" },
    offers: { "@type": "AggregateOffer", priceCurrency: "USD", lowPrice: "0", highPrice: "22", offerCount: PLANS.length },
  };

  return (
    <main className="flex flex-col min-h-screen">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />

      <SiteNav />

      {/* Hero */}
      <section className="relative overflow-hidden">
        <div className={`pointer-events-none absolute -top-40 -left-40 w-[40rem] h-[40rem] rounded-full blur-3xl ${theme.glow}`} />
        <div className={`pointer-events-none absolute top-20 -right-40 w-[40rem] h-[40rem] rounded-full blur-3xl ${theme.glow}`} />

        <div className="relative max-w-3xl mx-auto px-6 py-24 flex flex-col gap-6 items-start">
          <a href="/games" className="text-sm text-zinc-400 hover:text-zinc-200 transition-colors">{texts.back}</a>

          <div className={`inline-flex items-center gap-2 bg-zinc-900/80 text-sm px-3 py-1 rounded-full border ${theme.border} ${theme.text}`}>
            <Icon className="w-4 h-4" /> {texts.badge}
          </div>

          <h1 className="text-5xl sm:text-6xl font-extrabold tracking-tight leading-[1.05] uppercase">
            {texts.title1}<br />
            <span className={`bg-gradient-to-r ${game?.accent ?? ""} bg-clip-text text-transparent`}>{texts.title2}</span>
          </h1>

          <p className="text-zinc-400 text-lg max-w-2xl">{texts.subtitle}</p>

          <ul className="flex flex-col gap-2 mt-2">
            {texts.bullets.map((b) => (
              <li key={b} className="flex items-center gap-2 text-zinc-300">
                <Check className={`w-5 h-5 shrink-0 ${theme.text}`} /> {b}
              </li>
            ))}
          </ul>

          <div className="mt-4 flex flex-wrap gap-3">
            <a href={`${API}/auth/discord`} className={`inline-flex text-black font-semibold px-6 py-3 rounded-lg transition-colors ${theme.btn}`}>
              {texts.ctaStart}
            </a>
            <a href="#plans" className="border border-zinc-600 hover:border-zinc-400 bg-zinc-900/50 text-zinc-200 px-6 py-3 rounded-lg transition-colors">
              {t.hero.ctaPlans}
            </a>
          </div>

          <PingBadge className="mt-2" />
        </div>
      </section>

      {/* Plans (partagés) */}
      <section id="plans" className="px-6 pb-32 max-w-6xl mx-auto w-full">
        <h2 className="text-3xl font-bold text-center mb-2">{t.pricing.heading}</h2>
        <p className="text-zinc-500 text-center mb-2">{t.pricing.sub}</p>
        {floor > 0 && (
          <p className={`text-sm text-center mb-8 ${theme.text}`}>{t.pricing.minIncluded.replace("{n}", String(floor))}</p>
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
                    period === p ? `${theme.btn} text-black` : "text-zinc-400 hover:text-zinc-200"
                  }`}
                >
                  {label}
                  {disc > 0 && (
                    <span className={`ml-1.5 text-[10px] font-bold px-1.5 py-0.5 rounded ${period === p ? "bg-black/20" : `bg-zinc-800 ${theme.text}`}`}>
                      -{Math.round(disc * 100)}%
                    </span>
                  )}
                </button>
              );
            })}
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
          {PLANS.map((plan) => {
            const { monthly, total, discount, originalMonthly, promoActive } = priceFor(plan, lang, period);
            const isFree = plan.id === "free";
            return (
              <div
                key={plan.id}
                className={`relative rounded-xl border p-5 flex flex-col gap-3 ${
                  plan.popular ? `${theme.border} bg-zinc-900 shadow-lg` : "border-zinc-800 bg-zinc-900"
                }`}
              >
                {plan.popular && (
                  <span className={`absolute -top-3 left-1/2 -translate-x-1/2 text-[10px] font-bold uppercase tracking-wide text-black px-2 py-0.5 rounded-full ${theme.btn}`}>
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
                    <div className={`text-xs font-medium mt-0.5 ${theme.text}`}>{t.pricing.save} {Math.round(discount * 100)}%</div>
                  )}
                </div>
                <ul className="text-sm text-zinc-400 flex flex-col gap-1.5 mt-1">
                  <li className="flex items-center gap-2"><Check className={`w-3.5 h-3.5 ${theme.text}`} /> {Math.max(parseInt(plan.ram), floor)} GB {t.pricing.ramLabel}</li>
                  <li className="flex items-center gap-2"><Check className={`w-3.5 h-3.5 ${theme.text}`} /> {plan.cpu} {plan.cpu > 1 ? "cœurs" : "cœur"}</li>
                  <li className="flex items-center gap-2"><Check className={`w-3.5 h-3.5 ${theme.text}`} /> {plan.slots === 0 ? t.pricing.unlimited : `${plan.slots} ${t.pricing.players}`}</li>
                </ul>
                <a
                  href={`${API}/auth/discord`}
                  className={`mt-auto text-center text-sm font-medium py-2 rounded-lg transition-colors ${
                    plan.popular ? `${theme.btn} text-black` : "bg-zinc-800 hover:bg-zinc-700 text-zinc-100"
                  }`}
                >
                  {isFree ? t.pricing.ctaFree : t.pricing.ctaChoose}
                </a>
              </div>
            );
          })}
        </div>
      </section>

      <SiteFooter />
    </main>
  );
}
