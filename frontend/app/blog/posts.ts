export interface Section {
  h?: string;
  p: string[];
}

export interface Post {
  slug: string;
  title: string;
  description: string;
  date: string; // ISO
  keywords: string[];
  sections: Section[];
}

export const posts: Post[] = [
  {
    slug: "creer-serveur-minecraft-gratuit",
    title: "Comment créer un serveur Minecraft gratuit en 2026",
    description:
      "Guide simple pour créer et héberger ton serveur Minecraft gratuitement en quelques minutes, sans carte de crédit ni configuration compliquée.",
    date: "2026-06-09",
    keywords: [
      "créer serveur minecraft gratuit",
      "héberger serveur minecraft",
      "serveur minecraft gratuit 2026",
    ],
    sections: [
      {
        p: [
          "Créer un serveur Minecraft pour jouer avec tes amis n'a jamais été aussi simple. Plus besoin de laisser ton PC allumé toute la nuit ou de bidouiller des fichiers de configuration : avec un service d'hébergement, ton serveur tourne 24/7 dans le cloud et tes amis s'y connectent quand ils veulent.",
        ],
      },
      {
        h: "1. Choisir un hébergement gratuit",
        p: [
          "La première étape est de choisir un hébergeur qui offre un plan gratuit. Sur Playrena, tu peux créer un serveur de 1 Go de RAM gratuitement, suffisant pour 5 joueurs en mode survie ou créatif. Aucune carte de crédit n'est requise pour démarrer.",
        ],
      },
      {
        h: "2. Créer ton serveur en 30 secondes",
        p: [
          "Une fois connecté avec Discord, clique sur « Créer mon serveur », choisis un nom et la version de Minecraft souhaitée (de 1.8 à la dernière version). Ton serveur Paper se déploie automatiquement avec des performances optimisées (flags Aikar inclus).",
        ],
      },
      {
        h: "3. Se connecter et inviter tes amis",
        p: [
          "Ton serveur reçoit une adresse personnalisée du type ton-nom.servers.vbt-prog.com. Tes amis n'ont qu'à entrer cette adresse dans Minecraft pour rejoindre. Tu peux gérer la whitelist, les permissions et les règles depuis ton tableau de bord.",
        ],
      },
      {
        h: "Et si je veux plus de puissance ?",
        p: [
          "Le plan gratuit est parfait pour commencer. Si ton serveur grandit, tu peux passer à un plan supérieur (2, 4, 8 ou 16 Go de RAM) en quelques clics — les ressources s'appliquent immédiatement sans perdre ton monde.",
        ],
      },
    ],
  },
  {
    slug: "meilleurs-mods-minecraft",
    title: "Les meilleurs mods Minecraft à installer en 2026",
    description:
      "Sélection des meilleurs mods Minecraft pour améliorer tes performances, ton gameplay et ton serveur multijoueur en 2026.",
    date: "2026-06-09",
    keywords: ["meilleurs mods minecraft", "mods minecraft 2026", "mods serveur minecraft"],
    sections: [
      {
        p: [
          "Les mods transforment complètement l'expérience Minecraft. Que tu cherches à booster tes FPS, ajouter du contenu ou améliorer ton serveur, voici une sélection des mods incontournables en 2026.",
        ],
      },
      {
        h: "Mods de performance",
        p: [
          "Sur un serveur, les performances comptent. Paper (la base de nos serveurs) optimise déjà énormément le jeu côté serveur. Côté client, des mods comme Sodium et Lithium améliorent drastiquement les FPS et l'optimisation.",
        ],
      },
      {
        h: "Mods de gameplay",
        p: [
          "Pour enrichir ton aventure : Create (machines et automatisation), Twilight Forest (nouvelle dimension), ou des modpacks complets disponibles via Modrinth et CurseForge.",
        ],
      },
      {
        h: "Installer des mods sur ton serveur",
        p: [
          "Sur Playrena, tu peux choisir le type de serveur (Paper pour les plugins, Fabric ou Forge pour les mods) à la création. Les plans payants te permettent d'uploader tes propres mods et plugins directement.",
        ],
      },
    ],
  },
  {
    slug: "paper-vs-vanilla-minecraft",
    title: "Serveur Minecraft : Paper vs Vanilla, lequel choisir ?",
    description:
      "Comparatif entre un serveur Minecraft Paper et Vanilla : performances, plugins, compatibilité. Lequel choisir pour ton serveur ?",
    date: "2026-06-09",
    keywords: ["paper vs vanilla", "serveur paper minecraft", "type serveur minecraft"],
    sections: [
      {
        p: [
          "Quand tu crées un serveur Minecraft, tu dois choisir le « type » de serveur. Les deux plus courants sont Vanilla (officiel Mojang) et Paper. Voici comment choisir.",
        ],
      },
      {
        h: "Vanilla : l'expérience pure",
        p: [
          "Vanilla, c'est le serveur officiel de Mojang, sans modification. Compatibilité parfaite avec chaque mise à jour, mais des performances limitées dès que plusieurs joueurs se connectent.",
        ],
      },
      {
        h: "Paper : performances et plugins",
        p: [
          "Paper est une version optimisée de Spigot/Bukkit. Il offre de bien meilleures performances (moins de lag), supporte des milliers de plugins, et propose des options de configuration avancées. C'est le choix par défaut sur Playrena.",
        ],
      },
      {
        h: "Notre recommandation",
        p: [
          "Pour la majorité des serveurs multijoueurs, Paper est le meilleur choix : performances supérieures, plugins, et compatible avec les clients Vanilla (tes amis n'ont rien à installer). Choisis Vanilla seulement si tu veux une expérience 100% officielle.",
        ],
      },
    ],
  },
];

export function getPost(slug: string): Post | undefined {
  return posts.find((p) => p.slug === slug);
}
