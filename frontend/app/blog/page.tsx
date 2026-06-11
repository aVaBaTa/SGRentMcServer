import type { Metadata } from "next";
import Link from "next/link";
import { Server, ArrowLeft } from "lucide-react";
import { posts } from "./posts";

export const metadata: Metadata = {
  title: "Blog — Guides Minecraft | Playrena",
  description:
    "Guides et tutoriels pour créer, héberger et optimiser ton serveur Minecraft : mods, versions, performances et plus.",
  alternates: { canonical: "https://playrena.vbt-prog.com/blog" },
};

export default function BlogIndex() {
  return (
    <main className="min-h-screen">
      <nav className="border-b border-zinc-800 px-6 py-4 flex items-center gap-3">
        <Link href="/" className="text-zinc-400 hover:text-zinc-100 transition-colors">
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <Server className="w-5 h-5 text-green-400" />
        <span className="font-bold">Playrena — Blog</span>
      </nav>

      <div className="max-w-3xl mx-auto px-6 py-12">
        <h1 className="text-3xl font-bold mb-2">Guides Minecraft</h1>
        <p className="text-zinc-400 mb-10">
          Tout pour créer, héberger et optimiser ton serveur Minecraft.
        </p>

        <div className="flex flex-col gap-4">
          {posts.map((post) => (
            <Link
              key={post.slug}
              href={`/blog/${post.slug}`}
              className="block rounded-xl border border-zinc-800 bg-zinc-900 p-6 hover:border-green-700 transition-colors"
            >
              <h2 className="text-lg font-semibold mb-1">{post.title}</h2>
              <p className="text-sm text-zinc-400">{post.description}</p>
              <span className="text-xs text-green-400 mt-3 inline-block">Lire le guide →</span>
            </Link>
          ))}
        </div>
      </div>
    </main>
  );
}
