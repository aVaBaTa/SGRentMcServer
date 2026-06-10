"use client";

import { useState } from "react";
import {
  Server, Zap, Cpu, Shield, Star, Check, Clock, Headphones,
  Puzzle, RefreshCw, ShieldCheck, Gauge,
} from "lucide-react";
import {
  useI18n, PLANS, PERIODS, priceFor, fmtMoney, currencyCode, type Period, type Lang,
} from "@/lib/i18n";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const jsonLd = {
  "@context": "https://schema.org",
  "@type": "Product",
  name: "Playrena — Hébergement de serveurs Minecraft",
  description:
    "Hébergement de serveurs Minecraft performants. Gratuit pour commencer, plans payants avec plus de RAM et de CPU.",
  brand: { "@type": "Brand", name: "Playrena" },
  offers: {
    "@type": "AggregateOffer", priceCurrency: "USD", lowPrice: "0", highPrice: "25", offerCount: 5,
  },
};

function LanguageSwitcher() {
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

export default function Home() {
  const { t, lang } = useI18n();
  const [period, setPeriod] = useState<Period>("monthly");

  const badges = [
    { icon: Zap, label: t.badges.setup },
    { icon: Gauge, label: t.badges.uptime },
    { icon: Headphones, label: t.badges.support },
    { icon: Puzzle, label: t.badges.mods },
    { icon: ShieldCheck, label: t.badges.ddos },
    { icon: RefreshCw, label: t.badges.refund },
  ];

  return (
    <main className="flex flex-col min-h-screen">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />

      {/* Nav */}
      <nav className="border-b border-zinc-800 px-6 py-4 flex items-center justify-between sticky top-0 z-30 bg-zinc-950/80 backdrop-blur">
        <div className="flex items-center gap-2 font-bold text-lg">
          <Server className="w-5 h-5 text-green-400" />
          Playrena
        </div>
        <div className="flex items-center gap-4 sm:gap-6 text-sm text-zinc-400">
          <a href="#plans" className="hidden sm:inline hover:text-zinc-100 transition-colors">{t.nav.plans}</a>
          <a href="/blog" className="hidden sm:inline hover:text-zinc-100 transition-colors">{t.nav.blog}</a>
          <LanguageSwitcher />
          <a href={`${API}/auth/discord`} className="bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-lg transition-colors font-medium">
            {t.nav.login}
          </a>
        </div>
      </nav>

      {/* Hero */}
      <section className="relative overflow-hidden">
        {/* Halo décoratif */}
        <div className="pointer-events-none absolute -top-40 -left-40 w-[40rem] h-[40rem] rounded-full bg-green-500/10 blur-3xl" />
        <div className="pointer-events-none absolute top-20 -right-40 w-[40rem] h-[40rem] rounded-full bg-blue-500/10 blur-3xl" />

        <div className="relative max-w-6xl mx-auto px-6 py-20 grid lg:grid-cols-2 gap-12 items-center">
          {/* Texte */}
          <div className="flex flex-col gap-6">
            <div className="flex items-center gap-2">
              <div className="flex">
                {[0, 1, 2, 3, 4].map((i) => (
                  <Star key={i} className="w-4 h-4 fill-green-400 text-green-400" />
                ))}
              </div>
              <span className="text-sm text-zinc-400"><b className="text-zinc-200">{t.rating}</b> · {t.ratingSuffix}</span>
            </div>

            <p className="text-sm font-semibold tracking-widest text-green-400 uppercase">{t.hero.eyebrow}</p>

            <h1 className="text-5xl sm:text-6xl font-extrabold tracking-tight leading-[1.05] uppercase">
              {t.hero.title1}<br />
              <span className="bg-gradient-to-r from-blue-400 to-green-400 bg-clip-text text-transparent">{t.hero.title2}</span>
            </h1>

            <p className="text-zinc-400 text-lg max-w-xl">
              {t.hero.subtitle}<b className="text-zinc-200">{t.hero.subtitleBold}</b>{t.hero.subtitleEnd}
            </p>

            <div className="flex flex-wrap gap-3">
              <a href={`${API}/auth/discord`} className="bg-green-500 hover:bg-green-400 text-black font-semibold px-6 py-3 rounded-lg transition-colors">
                {t.hero.ctaStart}
              </a>
              <a href="#plans" className="border border-zinc-700 hover:border-zinc-500 text-zinc-300 px-6 py-3 rounded-lg transition-colors">
                {t.hero.ctaPlans}
              </a>
            </div>

            {/* Badges */}
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-x-6 gap-y-3 mt-2">
              {badges.map(({ icon: Icon, label }) => (
                <div key={label} className="flex items-center gap-2 text-sm text-zinc-300">
                  <Icon className="w-4 h-4 text-green-400 shrink-0" />
                  {label}
                </div>
              ))}
            </div>
            <p className="text-zinc-600 text-sm">{t.hero.soon}</p>
          </div>

          {/* Visuel */}
          <div className="relative">
            <div className="absolute inset-0 bg-gradient-to-tr from-blue-500/20 to-green-500/20 blur-2xl rounded-3xl" />
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img src="/hero.png" alt="Playrena" className="relative w-full rounded-2xl border border-zinc-800 shadow-2xl" />
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="grid grid-cols-1 md:grid-cols-3 gap-6 px-6 max-w-5xl mx-auto w-full py-16">
        {[Zap, Cpu, Shield].map((Icon, i) => (
          <div key={i} className="border border-zinc-800 rounded-xl p-6 bg-zinc-900">
            <Icon className="w-6 h-6 text-green-400 mb-3" />
            <h3 className="font-semibold mb-1">{t.features[i].title}</h3>
            <p className="text-zinc-400 text-sm">{t.features[i].desc}</p>
          </div>
        ))}
      </section>

      {/* Plans */}
      <section id="plans" className="px-6 pb-32 max-w-6xl mx-auto w-full">
        <h2 className="text-3xl font-bold text-center mb-2">{t.pricing.heading}</h2>
        <p className="text-zinc-500 text-center mb-8">{t.pricing.sub}</p>

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
                    period === p ? "bg-green-500 text-black" : "text-zinc-400 hover:text-zinc-200"
                  }`}
                >
                  {label}
                  {disc > 0 && (
                    <span className={`ml-1.5 text-[10px] font-bold px-1.5 py-0.5 rounded ${period === p ? "bg-black/20" : "bg-green-500/15 text-green-400"}`}>
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
            const { monthly, total, discount } = priceFor(plan, lang, period);
            const isFree = plan.id === "free";
            return (
              <div
                key={plan.id}
                className={`relative rounded-xl border p-5 flex flex-col gap-3 ${
                  plan.popular ? "border-green-500 bg-green-950/30 shadow-lg shadow-green-500/10" : "border-zinc-800 bg-zinc-900"
                }`}
              >
                {plan.popular && (
                  <span className="absolute -top-3 left-1/2 -translate-x-1/2 text-[10px] font-bold uppercase tracking-wide bg-green-500 text-black px-2 py-0.5 rounded-full">
                    {t.pricing.popular}
                  </span>
                )}
                <div>
                  <div className="font-bold text-lg">{t.planNames[plan.id]}</div>
                  <div className="mt-1 flex items-end gap-1">
                    <span className="text-3xl font-extrabold">{isFree ? t.pricing.free : fmtMoney(monthly, lang)}</span>
                    {!isFree && <span className="text-zinc-500 text-sm mb-1">{t.pricing.perMonth}</span>}
                  </div>
                  {!isFree && period !== "monthly" && (
                    <div className="text-xs text-zinc-500 mt-1">
                      {(period === "quarterly" ? t.pricing.billedQuarterly : t.pricing.billedAnnually).replace("{x}", `${fmtMoney(total, lang)} ${currencyCode(lang)}`)}
                    </div>
                  )}
                  {!isFree && discount > 0 && (
                    <div className="text-xs text-green-400 font-medium mt-0.5">{t.pricing.save} {Math.round(discount * 100)}%</div>
                  )}
                </div>
                <ul className="text-sm text-zinc-400 flex flex-col gap-1.5 mt-1">
                  <li className="flex items-center gap-2"><Check className="w-3.5 h-3.5 text-green-400" /> {plan.ram} {t.pricing.ramLabel}</li>
                  <li className="flex items-center gap-2"><Check className="w-3.5 h-3.5 text-green-400" /> {plan.cpu} {plan.cpu > 1 ? "cœurs" : "cœur"}</li>
                  <li className="flex items-center gap-2"><Check className="w-3.5 h-3.5 text-green-400" /> {plan.slots === 0 ? t.pricing.unlimited : `${plan.slots} ${t.pricing.players}`}</li>
                </ul>
                <a
                  href={`${API}/auth/discord`}
                  className={`mt-auto text-center text-sm font-medium py-2 rounded-lg transition-colors ${
                    plan.popular ? "bg-green-500 hover:bg-green-400 text-black" : "bg-zinc-800 hover:bg-zinc-700 text-zinc-100"
                  }`}
                >
                  {isFree ? t.pricing.ctaFree : t.pricing.ctaChoose}
                </a>
              </div>
            );
          })}
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-zinc-800 px-6 py-6 flex flex-col sm:flex-row items-center justify-between gap-3 text-sm text-zinc-600 mt-auto">
        <span>© {new Date().getFullYear()} Playrena — {t.footer}</span>
        <div className="flex items-center gap-4">
          <a href="/blog" className="hover:text-zinc-300 transition-colors flex items-center gap-1"><Clock className="w-3.5 h-3.5" /> {t.nav.blog}</a>
        </div>
      </footer>
    </main>
  );
}
