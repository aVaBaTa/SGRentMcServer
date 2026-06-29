import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Script from "next/script";
import "./globals.css";
import { LanguageProvider } from "@/lib/i18n";
import { FeedbackWidget } from "@/components/site-chrome";
import { PageTracker } from "@/components/page-tracker";

const GADS_ID = "AW-18226964787";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  metadataBase: new URL("https://playrena.vbt-prog.com"),
  title: "Playrena — Hébergement de serveurs Minecraft gratuit",
  description:
    "Crée et héberge ton serveur Minecraft gratuitement en 30 secondes. Paper, mods, console, gestion des joueurs. Serveurs performants au Québec.",
  keywords: [
    "serveur minecraft gratuit",
    "hébergement minecraft",
    "héberger serveur minecraft",
    "minecraft server hosting",
    "serveur minecraft québec",
    "serveur minecraft canada",
    "paper minecraft",
  ],
  openGraph: {
    title: "Playrena — Ton serveur Minecraft gratuit en 30 secondes",
    description:
      "Hébergement de serveurs Minecraft performants. Gratuit pour commencer, mods inclus.",
    url: "https://playrena.vbt-prog.com",
    siteName: "Playrena",
    locale: "fr_CA",
    type: "website",
  },
  alternates: { canonical: "https://playrena.vbt-prog.com" },
};

// Récupère le rabais global côté serveur (SSR) → prix réduits dès le 1er rendu.
// URL interne si dispo (rapide), sinon URL publique. Cache 60 s (revalidate).
async function getInitialPromo(): Promise<number> {
  const base = process.env.INTERNAL_API_URL ?? process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
  try {
    // Timeout court : au BUILD l'API n'est pas joignable → on échoue vite (sinon le build
    // hang 60s/page). Au runtime l'URL interne répond en quelques ms.
    const res = await fetch(`${base}/api/v1/promo`, { next: { revalidate: 30 }, signal: AbortSignal.timeout(2000) });
    if (!res.ok) return 0;
    const d = await res.json();
    return Number(d?.percent) || 0;
  } catch {
    return 0;
  }
}

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const initialPromo = await getInitialPromo();
  return (
    <html
      lang="fr"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col bg-zinc-950 text-zinc-100">
        <LanguageProvider initialPromo={initialPromo}>
          {children}
          <PageTracker />
          <FeedbackWidget />
        </LanguageProvider>
      </body>

      {/* Google tag (gtag.js) — Google Ads */}
      <Script
        src={`https://www.googletagmanager.com/gtag/js?id=${GADS_ID}`}
        strategy="afterInteractive"
      />
      <Script id="google-ads" strategy="afterInteractive">
        {`
          window.dataLayer = window.dataLayer || [];
          function gtag(){dataLayer.push(arguments);}
          gtag('js', new Date());
          gtag('config', '${GADS_ID}');
        `}
      </Script>
    </html>
  );
}
