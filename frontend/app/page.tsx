import { Server, Zap, Cpu, Shield } from "lucide-react";

const plans = [
  { name: "Gratuit", price: "0$", ram: "1 GB", cpu: "1 cœur", slots: "5 joueurs", highlight: false },
  { name: "Starter", price: "3$/mo", ram: "2 GB", cpu: "1 cœur", slots: "20 joueurs", highlight: false },
  { name: "Standard", price: "7$/mo", ram: "4 GB", cpu: "2 cœurs", slots: "50 joueurs", highlight: true },
  { name: "Pro", price: "14$/mo", ram: "8 GB", cpu: "4 cœurs", slots: "100 joueurs", highlight: false },
  { name: "Extreme", price: "25$/mo", ram: "16 GB", cpu: "6 cœurs", slots: "Illimité", highlight: false },
];

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// Tant que le backend (auth Discord) n'est pas déployé, on désactive le login.
const COMING_SOON = false;

const jsonLd = {
  "@context": "https://schema.org",
  "@type": "Product",
  name: "Playrena — Hébergement de serveurs Minecraft",
  description:
    "Hébergement de serveurs Minecraft performants. Gratuit pour commencer, plans payants avec plus de RAM et de CPU.",
  brand: { "@type": "Brand", name: "Playrena" },
  offers: {
    "@type": "AggregateOffer",
    priceCurrency: "USD",
    lowPrice: "0",
    highPrice: "25",
    offerCount: 5,
    offers: [
      { "@type": "Offer", name: "Gratuit", price: "0", priceCurrency: "USD" },
      { "@type": "Offer", name: "Starter", price: "3", priceCurrency: "USD" },
      { "@type": "Offer", name: "Standard", price: "7", priceCurrency: "USD" },
      { "@type": "Offer", name: "Pro", price: "14", priceCurrency: "USD" },
      { "@type": "Offer", name: "Extreme", price: "25", priceCurrency: "USD" },
    ],
  },
};

export default function Home() {
  return (
    <main className="flex flex-col min-h-screen">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      {/* Nav */}
      <nav className="border-b border-zinc-800 px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-2 font-bold text-lg">
          <Server className="w-5 h-5 text-green-400" />
          Playrena
        </div>
        <div className="flex items-center gap-6 text-sm text-zinc-400">
          <a href="#plans" className="hover:text-zinc-100 transition-colors">Plans</a>
          <a href="/blog" className="hover:text-zinc-100 transition-colors">Blog</a>
          <span className="text-zinc-600">Jeux — bientôt</span>
          {COMING_SOON ? (
            <span className="bg-zinc-800 text-zinc-500 px-4 py-2 rounded-lg font-medium cursor-not-allowed">
              Connexion — bientôt
            </span>
          ) : (
            <a
              href={`${API}/auth/discord`}
              className="bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-lg transition-colors font-medium"
            >
              Se connecter
            </a>
          )}
        </div>
      </nav>

      {/* Hero */}
      <section className="flex flex-col items-center justify-center text-center px-6 py-32 gap-6">
        <div className="inline-flex items-center gap-2 bg-green-950 text-green-400 text-sm px-3 py-1 rounded-full border border-green-800">
          <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
          Serveurs en ligne — essai gratuit disponible
        </div>
        <p className="text-sm font-semibold tracking-widest text-green-400 uppercase">
          Free Game Hosting — Hébergement de jeux gratuit
        </p>
        <h1 className="text-5xl font-bold tracking-tight max-w-2xl leading-tight">
          Ton serveur Minecraft,<br />
          <span className="text-green-400">en 30 secondes.</span>
        </h1>
        <p className="text-zinc-400 text-lg max-w-xl">
          Déploie un serveur Minecraft gratuitement. Upgrade quand tu es prêt.
          Mods, console, gestion des joueurs — tout inclus.
          <span className="block mt-2 text-zinc-500 text-base">Bientôt : Satisfactory, Rust, ARK et plus.</span>
        </p>
        <div className="flex gap-3 mt-2">
          {COMING_SOON ? (
            <span className="bg-zinc-800 text-zinc-500 font-semibold px-6 py-3 rounded-lg cursor-not-allowed">
              Lancement bientôt
            </span>
          ) : (
            <a
              href={`${API}/auth/discord`}
              className="bg-green-500 hover:bg-green-400 text-black font-semibold px-6 py-3 rounded-lg transition-colors"
            >
              Commencer gratuitement
            </a>
          )}
          <a
            href="#plans"
            className="border border-zinc-700 hover:border-zinc-500 text-zinc-300 px-6 py-3 rounded-lg transition-colors"
          >
            Voir les plans
          </a>
        </div>
      </section>

      {/* Features */}
      <section className="grid grid-cols-1 md:grid-cols-3 gap-6 px-6 max-w-5xl mx-auto w-full pb-24">
        {[
          { icon: Zap, title: "Démarrage instantané", desc: "Ton serveur est prêt en moins de 30 secondes. Paper, Fabric ou Forge." },
          { icon: Cpu, title: "Matériel dédié", desc: "Xeon 24 cœurs, 32 GB RAM par node. Limites CPU et RAM garanties par plan." },
          { icon: Shield, title: "Données protégées", desc: "Tes mondes sont conservés même si tu supprimes ton serveur. Backups disponibles." },
        ].map(({ icon: Icon, title, desc }) => (
          <div key={title} className="border border-zinc-800 rounded-xl p-6 bg-zinc-900">
            <Icon className="w-6 h-6 text-green-400 mb-3" />
            <h3 className="font-semibold mb-1">{title}</h3>
            <p className="text-zinc-400 text-sm">{desc}</p>
          </div>
        ))}
      </section>

      {/* Plans */}
      <section id="plans" className="px-6 pb-32 max-w-5xl mx-auto w-full">
        <h2 className="text-3xl font-bold text-center mb-12">Choisis ton plan</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
          {plans.map((plan) => (
            <div
              key={plan.name}
              className={`rounded-xl border p-5 flex flex-col gap-3 ${
                plan.highlight
                  ? "border-green-500 bg-green-950/30"
                  : "border-zinc-800 bg-zinc-900"
              }`}
            >
              {plan.highlight && (
                <span className="text-xs font-semibold text-green-400 uppercase tracking-wide">Populaire</span>
              )}
              <div>
                <div className="font-bold text-lg">{plan.name}</div>
                <div className="text-2xl font-bold mt-1">{plan.price}</div>
              </div>
              <ul className="text-sm text-zinc-400 flex flex-col gap-1">
                <li>🖥 {plan.ram} RAM</li>
                <li>⚡ {plan.cpu}</li>
                <li>👥 {plan.slots}</li>
              </ul>
              {COMING_SOON ? (
                <span className="mt-auto text-center text-sm font-medium py-2 rounded-lg bg-zinc-800 text-zinc-500 cursor-not-allowed">
                  Bientôt
                </span>
              ) : (
                <a
                  href={`${API}/auth/discord`}
                  className={`mt-auto text-center text-sm font-medium py-2 rounded-lg transition-colors ${
                    plan.highlight
                      ? "bg-green-500 hover:bg-green-400 text-black"
                      : "bg-zinc-800 hover:bg-zinc-700 text-zinc-100"
                  }`}
                >
                  {plan.name === "Gratuit" ? "Commencer" : "Choisir"}
                </a>
              )}
            </div>
          ))}
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-zinc-800 px-6 py-6 text-center text-sm text-zinc-600 mt-auto">
        © {new Date().getFullYear()} Playrena — vbt-prog.com
      </footer>
    </main>
  );
}
