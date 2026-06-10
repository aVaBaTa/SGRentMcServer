import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Script from "next/script";
import "./globals.css";
import { LanguageProvider } from "@/lib/i18n";

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
  metadataBase: new URL("https://mcserver.vbt-prog.com"),
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
    url: "https://mcserver.vbt-prog.com",
    siteName: "Playrena",
    locale: "fr_CA",
    type: "website",
  },
  alternates: { canonical: "https://mcserver.vbt-prog.com" },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="fr"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col bg-zinc-950 text-zinc-100">
        <LanguageProvider>{children}</LanguageProvider>
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
