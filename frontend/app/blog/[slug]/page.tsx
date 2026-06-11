import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { Server, ArrowLeft } from "lucide-react";
import { posts, getPost } from "../posts";

export function generateStaticParams() {
  return posts.map((p) => ({ slug: p.slug }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const { slug } = await params;
  const post = getPost(slug);
  if (!post) return {};
  const url = `https://playrena.vbt-prog.com/blog/${post.slug}`;
  return {
    title: `${post.title} | Playrena`,
    description: post.description,
    keywords: post.keywords,
    alternates: { canonical: url },
    openGraph: {
      title: post.title,
      description: post.description,
      url,
      type: "article",
    },
  };
}

export default async function BlogPost({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const post = getPost(slug);
  if (!post) notFound();

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "Article",
    headline: post.title,
    description: post.description,
    datePublished: post.date,
    author: { "@type": "Organization", name: "Playrena" },
  };

  return (
    <main className="min-h-screen">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <nav className="border-b border-zinc-800 px-6 py-4 flex items-center gap-3">
        <Link href="/blog" className="text-zinc-400 hover:text-zinc-100 transition-colors">
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <Server className="w-5 h-5 text-green-400" />
        <span className="font-bold">Playrena — Blog</span>
      </nav>

      <article className="max-w-2xl mx-auto px-6 py-12">
        <h1 className="text-3xl font-bold mb-6 leading-tight">{post.title}</h1>
        {post.sections.map((s, i) => (
          <section key={i} className="mb-6">
            {s.h && <h2 className="text-xl font-semibold mb-2 text-green-400">{s.h}</h2>}
            {s.p.map((para, j) => (
              <p key={j} className="text-zinc-300 leading-relaxed mb-3">{para}</p>
            ))}
          </section>
        ))}

        <div className="mt-10 rounded-xl border border-green-800 bg-green-950/20 p-6 text-center">
          <p className="text-zinc-200 mb-4 font-medium">Prêt à lancer ton serveur Minecraft ?</p>
          <Link
            href="/"
            className="inline-block bg-green-500 hover:bg-green-400 text-black font-semibold px-6 py-3 rounded-lg transition-colors"
          >
            Créer mon serveur gratuit
          </Link>
        </div>
      </article>
    </main>
  );
}
