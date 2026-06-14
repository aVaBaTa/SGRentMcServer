"use client";

import { useState } from "react";
import {
  Zap, Cpu, Shield, Star, Check, Headphones,
  Puzzle, RefreshCw, ShieldCheck, Gauge, ChevronDown,
} from "lucide-react";
import {
  useI18n, PLANS, PERIODS, priceFor, fmtMoney, currencyCode, type Period,
} from "@/lib/i18n";
import { SiteNav, SiteFooter } from "@/components/site-chrome";
import { PingBadge } from "@/components/ping-badge";

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

export default function MinecraftHosting() {
  const { t, lang } = useI18n();
  const [period, setPeriod] = useState<Period>("monthly");
  const [faqOpen, setFaqOpen] = useState<number | null>(0);

  // Formulaire de support
  const [sup, setSup] = useState({ name: "", email: "", subject: "", message: "" });
  const [supStatus, setSupStatus] = useState<"idle" | "sending" | "sent" | "error">("idle");

  async function submitSupport(e: React.FormEvent) {
    e.preventDefault();
    setSupStatus("sending");
    try {
      const res = await fetch(`${API}/api/v1/support`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(sup),
      });
      if (res.ok) {
        setSup({ name: "", email: "", subject: "", message: "" });
        setSupStatus("sent");
      } else {
        setSupStatus("error");
      }
    } catch {
      setSupStatus("error");
    }
  }

  const faqLd = {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: t.faq.items.map((it) => ({
      "@type": "Question",
      name: it.q,
      acceptedAnswer: { "@type": "Answer", text: it.a },
    })),
  };

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

      <SiteNav />

      {/* Hero avec image en arrière-plan */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 -z-10">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src="/hero.png" alt="" aria-hidden className="w-full h-full object-cover object-center opacity-40" />
          {/* Voiles pour la lisibilité du texte */}
          <div className="absolute inset-0 bg-gradient-to-r from-zinc-950 via-zinc-950/85 to-zinc-950/40" />
          <div className="absolute inset-0 bg-gradient-to-b from-zinc-950/40 via-transparent to-zinc-950" />
        </div>

        <div className="relative max-w-6xl mx-auto px-6 py-28 sm:py-36">
          <div className="flex flex-col gap-6 max-w-2xl">
            <div className="flex items-center gap-2">
              <div className="flex">
                {[0, 1, 2, 3, 4].map((i) => (
                  <Star key={i} className="w-4 h-4 fill-green-400 text-green-400" />
                ))}
              </div>
              <span className="text-sm text-zinc-300"><b className="text-zinc-100">{t.rating}</b> · {t.ratingSuffix}</span>
            </div>

            <p className="text-sm font-semibold tracking-widest text-green-400 uppercase">{t.hero.eyebrow}</p>

            <h1 className="text-5xl sm:text-6xl font-extrabold tracking-tight leading-[1.05] uppercase drop-shadow-lg">
              {t.hero.title1}<br />
              <span className="bg-gradient-to-r from-blue-400 to-green-400 bg-clip-text text-transparent">{t.hero.title2}</span>
            </h1>

            <p className="text-zinc-300 text-lg max-w-xl">
              {t.hero.subtitle}<b className="text-white">{t.hero.subtitleBold}</b>{t.hero.subtitleEnd}
            </p>

            <div className="flex flex-wrap gap-3">
              <a href={`${API}/auth/discord`} className="bg-green-500 hover:bg-green-400 text-black font-semibold px-6 py-3 rounded-lg transition-colors">
                {t.hero.ctaStart}
              </a>
              <a href="#plans" className="border border-zinc-600 hover:border-zinc-400 bg-zinc-900/50 text-zinc-200 px-6 py-3 rounded-lg transition-colors">
                {t.hero.ctaPlans}
              </a>
            </div>

            <PingBadge />

            {/* Badges */}
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-x-6 gap-y-3 mt-2">
              {badges.map(({ icon: Icon, label }) => (
                <div key={label} className="flex items-center gap-2 text-sm text-zinc-200">
                  <Icon className="w-4 h-4 text-green-400 shrink-0" />
                  {label}
                </div>
              ))}
            </div>
            <p className="text-zinc-400 text-sm">{t.hero.soon}</p>
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
            const { monthly, total, discount, originalMonthly, promoActive } = priceFor(plan, lang, period);
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

      {/* FAQ */}
      <section className="px-6 pb-32 max-w-3xl mx-auto w-full">
        <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(faqLd) }} />
        <h2 className="text-3xl font-bold text-center mb-2">{t.faq.heading}</h2>
        <p className="text-zinc-500 text-center mb-10">{t.faq.sub}</p>

        <div className="flex flex-col gap-3">
          {t.faq.items.map((item, i) => {
            const open = faqOpen === i;
            return (
              <div key={i} className="border border-zinc-800 rounded-xl bg-zinc-900 overflow-hidden">
                <button
                  type="button"
                  onClick={() => setFaqOpen(open ? null : i)}
                  aria-expanded={open}
                  className="w-full flex items-center justify-between gap-4 text-left px-5 py-4 hover:bg-zinc-800/50 transition-colors"
                >
                  <span className="font-medium">{item.q}</span>
                  <ChevronDown className={`w-5 h-5 shrink-0 text-green-400 transition-transform duration-200 ${open ? "rotate-180" : ""}`} />
                </button>
                <div className={`grid transition-all duration-200 ${open ? "grid-rows-[1fr]" : "grid-rows-[0fr]"}`}>
                  <div className="overflow-hidden">
                    <p className="px-5 pb-4 text-sm text-zinc-400 leading-relaxed">{item.a}</p>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </section>

      {/* Support */}
      <section id="support" className="px-6 pb-32 max-w-2xl mx-auto w-full">
        <h2 className="text-3xl font-bold text-center mb-2">{t.support.heading}</h2>
        <p className="text-zinc-500 text-center mb-8">{t.support.sub}</p>

        {supStatus === "sent" ? (
          <div className="border border-green-500/30 bg-green-950/20 rounded-xl p-8 text-center">
            <p className="text-green-400 font-semibold text-lg mb-1">{t.support.sentTitle}</p>
            <p className="text-zinc-400 text-sm mb-5">{t.support.sentDesc}</p>
            <button onClick={() => setSupStatus("idle")} className="text-sm text-green-400 hover:underline">
              {t.support.sendAnother}
            </button>
          </div>
        ) : (
          <form onSubmit={submitSupport} className="flex flex-col gap-3">
            <div className="grid sm:grid-cols-2 gap-3">
              <input
                required value={sup.name} onChange={(e) => setSup({ ...sup, name: e.target.value })}
                placeholder={t.support.name} maxLength={100}
                className="bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-3 text-sm focus:outline-none focus:border-green-500"
              />
              <input
                required type="email" value={sup.email} onChange={(e) => setSup({ ...sup, email: e.target.value })}
                placeholder={t.support.email}
                className="bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-3 text-sm focus:outline-none focus:border-green-500"
              />
            </div>
            <input
              value={sup.subject} onChange={(e) => setSup({ ...sup, subject: e.target.value })}
              placeholder={t.support.subject} maxLength={150}
              className="bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-3 text-sm focus:outline-none focus:border-green-500"
            />
            <textarea
              required value={sup.message} onChange={(e) => setSup({ ...sup, message: e.target.value })}
              placeholder={t.support.message} rows={5} maxLength={5000}
              className="bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-3 text-sm focus:outline-none focus:border-green-500 resize-y"
            />
            {supStatus === "error" && <p className="text-sm text-red-400">{t.support.error}</p>}
            <button
              type="submit" disabled={supStatus === "sending"}
              className="self-center flex items-center justify-center gap-2 bg-green-500 hover:bg-green-400 disabled:opacity-50 text-black font-semibold px-8 py-3 rounded-xl transition-colors"
            >
              {supStatus === "sending" ? t.support.sending : t.support.send}
            </button>
          </form>
        )}
      </section>

      <SiteFooter />
    </main>
  );
}
