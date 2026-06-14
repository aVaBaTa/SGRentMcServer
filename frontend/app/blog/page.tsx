import type { Metadata } from "next";
import Link from "next/link";
import { Server, ArrowLeft } from "lucide-react";
import { BLOG_CATEGORIES, postsByGame } from "./posts";
import { AuthButton } from "@/components/site-chrome";

export const metadata: Metadata = {
  title: "Blog — Guides serveurs de jeux (Minecraft, Hytale, Satisfactory) | Playrena",
  description:
    "Guides et tutoriels pour créer, héberger et optimiser tes serveurs de jeux : Minecraft, Hytale, Satisfactory — mods, ressources, performances et plus.",
  alternates: { canonical: "https://playrena.vbt-prog.com/blog" },
};

export default function BlogIndex() {
  return (
    <main className="min-h-screen">
      <nav className="border-b border-zinc-800 px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link href="/" className="text-zinc-400 hover:text-zinc-100 transition-colors">
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <Server className="w-5 h-5 text-green-400" />
          <span className="font-bold">Playrena — Blog</span>
        </div>
        <AuthButton />
      </nav>

      <div className="max-w-3xl mx-auto px-6 py-12">
        <h1 className="text-3xl font-bold mb-2">Guides & tutoriels</h1>
        <p className="text-zinc-400 mb-10">
          Tout pour créer, héberger et optimiser tes serveurs de jeux — par jeu.
        </p>

        <div className="flex flex-col gap-12">
          {BLOG_CATEGORIES.map(({ game, label }) => {
            const items = postsByGame(game);
            if (items.length === 0) return null;
            return (
              <section key={game}>
                <h2 className="text-2xl font-bold mb-4 border-b border-zinc-800 pb-2">{label}</h2>
                <div className="flex flex-col gap-4">
                  {items.map((post) => (
                    <Link
                      key={post.slug}
                      href={`/blog/${post.slug}`}
                      className="block rounded-xl border border-zinc-800 bg-zinc-900 p-6 hover:border-green-700 transition-colors"
                    >
                      <h3 className="text-lg font-semibold mb-1">{post.title}</h3>
                      <p className="text-sm text-zinc-400">{post.description}</p>
                      <span className="text-xs text-green-400 mt-3 inline-block">Lire le guide →</span>
                    </Link>
                  ))}
                </div>
              </section>
            );
          })}
        </div>
      </div>
    </main>
  );
}
