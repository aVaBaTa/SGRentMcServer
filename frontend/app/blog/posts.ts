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
    slug: "hytale-vs-minecraft-serveur",
    title: "Hytale vs Minecraft : quel serveur de jeu héberger en 2026 ?",
    description:
      "Comparatif complet entre héberger un serveur Hytale et un serveur Minecraft en 2026 : ressources, gameplay, plugins, coût et facilité. Lequel choisir pour jouer avec tes amis ?",
    date: "2026-06-12",
    keywords: [
      "hytale vs minecraft",
      "serveur hytale ou minecraft",
      "héberger serveur hytale 2026",
      "comparatif hytale minecraft",
      "quel serveur de jeu héberger",
    ],
    sections: [
      {
        p: [
          "Tu hésites entre monter un serveur Hytale ou un serveur Minecraft pour jouer avec tes amis ? Les deux jeux sont jouables sur Playrena, mais ils ne demandent pas les mêmes ressources et n'offrent pas la même expérience. Voici un comparatif honnête pour t'aider à choisir.",
        ],
      },
      {
        h: "Le gameplay : maturité contre nouveauté",
        p: [
          "Minecraft est une valeur sûre : plus de dix ans de contenu, un écosystème de plugins et de mods gigantesque, une communauté énorme. Hytale, lui, est le petit nouveau — un univers plus structuré, orienté aventure et création, avec un moteur moderne. Si tu veux du contenu éprouvé et des milliers de plugins, Minecraft gagne. Si tu veux découvrir un nouveau monde avec tes amis, Hytale est l'option fraîche.",
        ],
      },
      {
        h: "Les ressources : Hytale est plus gourmand",
        p: [
          "C'est la différence la plus concrète. Un serveur Minecraft tourne très bien à partir de 1 Go de RAM pour un petit groupe, et reste léger grâce à Paper. Hytale, basé sur un moteur Java moderne, est nettement plus exigeant : on parle d'un plancher de 10 Go de RAM et 4 cœurs CPU pour un gameplay fluide. Concrètement, héberger Hytale demande une machine bien plus musclée que Minecraft.",
        ],
      },
      {
        h: "Plugins et personnalisation",
        p: [
          "Côté Minecraft, l'écosystème Paper te donne accès à des milliers de plugins (permissions, économie, mini-jeux, protection de zones) et à des versions de 1.8 jusqu'à la dernière. Hytale est encore jeune : la personnalisation est plus limitée pour l'instant, mais le moteur a été pensé pour le modding et ça arrivera. Pour un serveur très personnalisé aujourd'hui, Minecraft a une longueur d'avance.",
        ],
      },
      {
        h: "Le coût sur Playrena",
        p: [
          "Sur Playrena, un serveur Minecraft est gratuit dès 1 Go de RAM, sans carte de crédit. Hytale est lui aussi offert pour l'instant (promo de lancement), avec un serveur dédié de 10 Go et 4 cœurs — ce qui serait normalement coûteux ailleurs vu les ressources demandées. C'est le bon moment pour tester Hytale sans débourser.",
        ],
      },
      {
        h: "Alors, lequel choisir ?",
        p: [
          "Choisis Minecraft si tu veux du contenu mature, des plugins à volonté et un serveur ultra-léger qui démarre gratuitement. Choisis Hytale si tu veux explorer la nouveauté avec un serveur dédié puissant, profité de la promo de lancement. Le mieux ? Tu peux créer les deux sur Playrena en quelques minutes via Discord et te faire ta propre idée. Lance ton serveur dès maintenant.",
        ],
      },
    ],
  },
  {
    slug: "combien-de-ram-serveur-minecraft",
    title: "Combien de RAM pour un serveur Minecraft ? Le guide 2026",
    description:
      "Combien de Go de RAM allouer à ton serveur Minecraft selon le nombre de joueurs, les plugins et les mods ? Guide de dimensionnement clair et chiffré pour 2026.",
    date: "2026-06-12",
    keywords: [
      "combien de ram serveur minecraft",
      "ram serveur minecraft",
      "dimensionner serveur minecraft",
      "go ram minecraft joueurs",
      "ram minecraft plugins",
    ],
    sections: [
      {
        p: [
          "« Combien de RAM pour mon serveur Minecraft ? » C'est LA question qu'on se pose en premier. Trop peu et ton serveur lague ou crashe ; trop et tu paies pour rien. Voici des repères clairs selon ton type de serveur en 2026.",
        ],
      },
      {
        h: "Les repères par nombre de joueurs (Vanilla / Paper léger)",
        p: [
          "Pour un serveur survie ou créatif sans gros plugins : 1 Go suffit pour 1 à 5 joueurs entre amis (c'est l'offre gratuite de Playrena). Compte 2 Go pour 5 à 10 joueurs, 4 Go pour 10 à 20 joueurs, et 6 à 8 Go au-delà. Ces chiffres supposent un serveur Paper, déjà bien optimisé.",
        ],
      },
      {
        h: "Plugins et mods changent tout",
        p: [
          "La RAM ne dépend pas que du nombre de joueurs : un serveur lourdement modé (Forge, Fabric) ou bourré de plugins consomme beaucoup plus. Un gros modpack peut exiger 6 à 8 Go même à quelques joueurs. Règle simple : ajoute environ 1 à 2 Go par couche importante de mods/plugins par-dessus le repère de base.",
        ],
      },
      {
        h: "La distance de vue (view-distance), le facteur caché",
        p: [
          "Plus la view-distance est élevée, plus le serveur charge de chunks par joueur, et plus la RAM grimpe. Passer de 10 à 16 de view-distance peut faire exploser la consommation. Si ton serveur manque de mémoire, réduire la view-distance est souvent plus efficace que rajouter de la RAM.",
        ],
      },
      {
        h: "N'alloue pas toute la RAM de la machine",
        p: [
          "Erreur classique : donner 100 % de la RAM à la JVM. Le système d'exploitation et le garbage collector ont besoin de marge. Sur Playrena, c'est géré pour toi : les flags Aikar sont déjà configurés pour un usage mémoire optimal, tu n'as pas à bricoler les paramètres de la JVM.",
        ],
      },
      {
        h: "En résumé, et comment ajuster",
        p: [
          "Commence petit : 1 Go gratuit pour démarrer, puis monte si tu observes du lag ou des avertissements de mémoire. Sur Playrena, tu changes de plan (2, 4, 8, 16 Go) en quelques clics, sans perdre ton monde, et les nouvelles ressources s'appliquent au prochain démarrage. Crée ton serveur Minecraft gratuitement et ajuste la RAM quand tu en as besoin.",
        ],
      },
    ],
  },
  {
    slug: "reduire-lag-serveur-minecraft",
    title: "Comment réduire le lag de ton serveur Minecraft",
    description:
      "Ton serveur Minecraft lague ? Guide concret pour diagnostiquer et réduire le lag : Paper, RAM, CPU, view-distance, plugins et l'outil Spark pour trouver la cause.",
    date: "2026-06-12",
    keywords: [
      "réduire lag serveur minecraft",
      "serveur minecraft lague",
      "optimiser serveur minecraft",
      "tps minecraft",
      "spark minecraft",
    ],
    sections: [
      {
        p: [
          "Rien de plus frustrant qu'un serveur Minecraft qui rame : blocs qui mettent du temps à casser, mobs qui téléportent, TPS qui chute. La bonne nouvelle, c'est que le lag a presque toujours une cause identifiable. Voici comment le diagnostiquer et le réduire concrètement.",
        ],
      },
      {
        h: "Comprendre le TPS et le type de lag",
        p: [
          "Le serveur vise 20 TPS (ticks par seconde). En dessous, le monde ralentit pour tout le monde — c'est du lag serveur. Attention à ne pas confondre avec le lag réseau (ta latence/ping) ou la chute de FPS côté client : ces trois problèmes ont des causes différentes. Si la commande de TPS montre moins de 20, c'est bien ton serveur qui est saturé.",
        ],
      },
      {
        h: "Pars sur Paper, pas Vanilla",
        p: [
          "La première optimisation, c'est le logiciel. Paper réduit massivement le lag par rapport à un serveur Vanilla grâce à ses optimisations internes et ses options de configuration. Sur Playrena, tous les serveurs Minecraft sont déjà en Paper avec les flags Aikar — tu pars donc d'une base saine sans rien configurer.",
        ],
      },
      {
        h: "RAM, CPU : le bon coupable",
        p: [
          "Le lag vient souvent d'un manque de RAM (le serveur passe son temps à nettoyer la mémoire) ou d'un CPU saturé (génération de monde, redstone, fermes à mobs). Astuce : la plupart du lag Minecraft est lié au CPU mono-thread, donc ajouter de la RAM ne règle pas tout. Vérifie d'abord lequel des deux sature avant d'augmenter ton plan.",
        ],
      },
      {
        h: "Baisse la view-distance et surveille les entités",
        p: [
          "Réduire la view-distance (par exemple de 12 à 8) et la simulation-distance soulage énormément le serveur, surtout avec plusieurs joueurs. Méfie-toi aussi des fermes à mobs géantes, des hoppers en masse et des items qui s'accumulent au sol : ce sont des sources de lag classiques que tu peux limiter.",
        ],
      },
      {
        h: "Trouve la vraie cause avec Spark",
        p: [
          "Plutôt que de deviner, installe le plugin Spark : il profile ton serveur et te montre exactement quels plugins, entités ou tâches consomment le CPU. C'est l'outil de référence pour identifier la source d'un lag tenace. Sur Playrena, tu déposes le plugin via le gestionnaire de fichiers et tu redémarres en un clic.",
        ],
      },
      {
        h: "Quand augmenter les ressources",
        p: [
          "Si tu as optimisé Paper, la view-distance et les plugins et que ça lague encore avec beaucoup de joueurs, c'est le signe qu'il faut plus de ressources. Sur Playrena, tu montes de plan (RAM/CPU) en quelques clics sans perdre ton monde. Crée ou migre ton serveur Minecraft sur Playrena pour partir d'une base optimisée.",
        ],
      },
    ],
  },
  {
    slug: "installer-plugins-serveur-minecraft-paper",
    title: "Comment installer des plugins sur ton serveur Minecraft (Paper)",
    description:
      "Tutoriel pas à pas pour installer des plugins sur un serveur Minecraft Paper : où les télécharger, comment les déposer via le panel, et comment les configurer sans bug.",
    date: "2026-06-12",
    keywords: [
      "installer plugins minecraft",
      "plugins serveur paper",
      "ajouter plugin minecraft",
      "tutoriel plugin minecraft",
      "dossier plugins minecraft",
    ],
    sections: [
      {
        p: [
          "Les plugins transforment un serveur Minecraft basique en serveur sur-mesure : permissions, économie, protection de zones, mini-jeux, anti-grief… Sur un serveur Paper, les installer est simple une fois qu'on connaît la méthode. Voici le tutoriel complet.",
        ],
      },
      {
        h: "Plugins ou mods : ne pas confondre",
        p: [
          "Important : les plugins (Bukkit/Spigot/Paper) et les mods (Forge/Fabric) sont différents et incompatibles entre eux. Un serveur Paper utilise des plugins, pas des mods. Gros avantage : avec des plugins, tes amis n'ont rien à installer côté client — ils rejoignent avec un Minecraft normal. C'est le mode par défaut sur Playrena.",
        ],
      },
      {
        h: "1. Télécharger le bon plugin",
        p: [
          "Récupère tes plugins sur des sources fiables comme Modrinth, SpigotMC ou Hangar (le dépôt officiel de Paper). Vérifie toujours que la version du plugin est compatible avec la version de ton serveur Minecraft. Un plugin prévu pour la 1.20 peut planter sur la 1.21 — c'est la cause numéro un d'erreurs.",
        ],
      },
      {
        h: "2. Déposer le fichier .jar dans le dossier plugins",
        p: [
          "Un plugin est un fichier .jar. Il doit aller dans le dossier « plugins » de ton serveur. Sur Playrena, ouvre le gestionnaire de fichiers depuis ton panel, va dans le dossier plugins, et glisse-dépose ton .jar — pas besoin de FTP ni de ligne de commande.",
        ],
      },
      {
        h: "3. Redémarrer le serveur",
        p: [
          "Un plugin n'est chargé qu'au démarrage. Après l'avoir déposé, redémarre ton serveur (bouton « Redémarrer » en un clic dans le panel). Au redémarrage, Paper charge le plugin et crée automatiquement son dossier de configuration dans « plugins/<NomDuPlugin> ».",
        ],
      },
      {
        h: "4. Configurer et vérifier dans la console",
        p: [
          "Après le redémarrage, ouvre la console en direct du panel : tu y verras si le plugin s'est chargé correctement ou s'il signale une erreur de version. Tu peux ensuite éditer son fichier config.yml via le gestionnaire de fichiers pour ajuster les options, puis recharger ou redémarrer.",
        ],
      },
      {
        h: "Les indispensables pour démarrer",
        p: [
          "Pour un premier serveur, jette un œil à des plugins comme LuckPerms (permissions), EssentialsX (commandes de base, /home, /tpa), WorldGuard (protection de zones) et Spark (diagnostic de performance). Crée ton serveur Paper sur Playrena et installe tes premiers plugins en quelques minutes depuis le panel.",
        ],
      },
    ],
  },
  {
    slug: "creer-serveur-satisfactory-dedie",
    title: "Comment créer un serveur Satisfactory dédié en 2026",
    description:
      "Tutoriel pour créer et héberger un serveur Satisfactory dédié en 2026 : déploiement en quelques minutes, connexion directe en IP:port, et jeu 24/7 avec tes amis.",
    date: "2026-06-12",
    keywords: [
      "créer serveur satisfactory",
      "serveur satisfactory dédié",
      "héberger serveur satisfactory",
      "satisfactory serveur dédié 2026",
      "tutoriel serveur satisfactory",
    ],
    sections: [
      {
        p: [
          "Satisfactory est bien plus fun en multijoueur, mais laisser l'hôte allumé en permanence est contraignant : si l'hôte se déconnecte, tout le monde est éjecté et l'usine s'arrête. La solution, c'est un serveur dédié qui tourne 24/7 dans le cloud. Voici comment en monter un en quelques minutes sur Playrena.",
        ],
      },
      {
        h: "Pourquoi un serveur dédié plutôt que l'hébergement par un joueur",
        p: [
          "En partie hébergée par un joueur, l'usine ne progresse que quand l'hôte est connecté, et ses performances dépendent du PC de l'hôte. Un serveur dédié, lui, tourne en continu : tes machines produisent même quand tu dors, tout le monde rejoint quand il veut, et personne n'a besoin d'être « l'hôte ».",
        ],
      },
      {
        h: "1. Connecte-toi à Playrena",
        p: [
          "Rends-toi sur playrena.vbt-prog.com et connecte-toi avec Discord — un clic, pas de formulaire ni de carte de crédit. Tu arrives directement sur ton tableau de bord.",
        ],
      },
      {
        h: "2. Crée ton serveur Satisfactory",
        p: [
          "Clique sur « Créer un serveur » et choisis Satisfactory dans la liste des jeux. Le plancher recommandé est de 4 Go de RAM, et l'hébergement Satisfactory est gratuit pour l'instant sur Playrena. Donne un nom à ton serveur et valide : il se déploie automatiquement.",
        ],
      },
      {
        h: "3. Connecte-toi en IP:port direct",
        p: [
          "Une fois le serveur en ligne, le panel affiche son adresse de connexion sous forme d'IP:port. Dans Satisfactory, va dans « Serveurs », ajoute un serveur avec cette adresse, configure-le (mot de passe admin au premier lancement) puis charge ou crée ta sauvegarde. À noter : Satisfactory n'utilise pas de console RCON, toute la gestion se fait dans le jeu et via le panel.",
        ],
      },
      {
        h: "4. Invite tes amis et gère ton serveur",
        p: [
          "Partage la même adresse IP:port à tes amis : ils se connectent directement, sans rien installer de plus. Depuis le panel Playrena, tu accèdes aux fichiers et aux sauvegardes de ton serveur, et tu peux le redémarrer en un clic. Besoin de plus de RAM pour une méga-usine ? Tu ajustes ton plan facilement.",
        ],
      },
      {
        h: "Lance-toi",
        p: [
          "En quelques minutes, ton usine tourne 24/7 sans dépendre de ton PC. Crée ton serveur Satisfactory dédié gratuitement sur Playrena et reprends ta production là où tu l'avais laissée — même hors ligne.",
        ],
      },
    ],
  },
  {
    slug: "hebergement-serveur-jeu-quebec",
    title: "Pourquoi héberger ton serveur de jeu au Québec",
    description:
      "Latence plus basse, service en français, données au Québec : voici pourquoi héberger ton serveur de jeu (Minecraft, Satisfactory, Hytale) localement au Québec change l'expérience.",
    date: "2026-06-12",
    keywords: [
      "hébergement serveur jeu québec",
      "serveur minecraft québec",
      "hébergeur jeu québec",
      "serveur de jeu bas ping québec",
      "héberger serveur québec",
    ],
    sections: [
      {
        p: [
          "Quand tu choisis un hébergeur de serveur de jeu, l'emplacement compte plus que tu ne le penses. Beaucoup de services populaires hébergent en Europe ou aux États-Unis, ce qui ajoute de la latence pour un joueur québécois. Héberger localement, au Québec, change concrètement l'expérience. Voici pourquoi.",
        ],
      },
      {
        h: "Une latence plus basse, un jeu plus réactif",
        p: [
          "La latence (ton ping) dépend en grande partie de la distance physique entre toi et le serveur. Un serveur situé loin ajoute des dizaines de millisecondes : blocs qui réagissent en retard sur Minecraft, désynchronisations en multijoueur. Pour un groupe d'amis basés au Québec, un serveur hébergé au Québec offre un ping plus bas et un gameplay nettement plus réactif.",
        ],
      },
      {
        h: "Un service pensé et fait au Québec",
        p: [
          "Playrena est conçu et opéré au Québec. Concrètement, ça veut dire un support et une interface en français, pensés pour les joueurs d'ici, sans avoir à jongler avec un service étranger ou un fuseau horaire à l'opposé. Tu parles à un projet local plutôt qu'à un géant impersonnel.",
        ],
      },
      {
        h: "Tes données restent proches",
        p: [
          "Héberger localement, c'est aussi garder tes mondes et tes sauvegardes près de chez toi plutôt que sur un continent lointain. Avec le panel Playrena, tu gardes la main sur tes fichiers et tes sauvegardes, quel que soit le jeu.",
        ],
      },
      {
        h: "Tous tes jeux au même endroit",
        p: [
          "Que tu joues à Minecraft, Satisfactory ou Hytale, tu gères tout depuis un seul panel hébergé localement : console en direct, fichiers, sauvegardes, redémarrage en un clic. Pas besoin de multiplier les comptes chez plusieurs hébergeurs étrangers.",
        ],
      },
      {
        h: "Essaie par toi-même",
        p: [
          "Le meilleur moyen de constater la différence de latence, c'est de tester. Connecte-toi avec Discord et crée ton serveur Minecraft gratuit (dès 1 Go) sur Playrena — hébergé au Québec, sans carte de crédit. Tu sentiras la différence dès la première partie.",
        ],
      },
    ],
  },
  {
    slug: "creer-serveur-hytale-tutoriel",
    title: "Comment créer un serveur Hytale en 2026 (tutoriel pas à pas)",
    description:
      "Tutoriel complet pour créer et héberger ton propre serveur Hytale en quelques minutes sur Playrena : création, authentification avec ton compte Hytale, et connexion entre amis.",
    date: "2026-06-12",
    keywords: [
      "créer serveur hytale",
      "héberger serveur hytale",
      "tutoriel serveur hytale",
      "serveur hytale 2026",
      "comment faire un serveur hytale",
    ],
    sections: [
      {
        p: [
          "Hytale est enfin disponible, et rien de mieux qu'un serveur dédié pour explorer ses mondes entre amis, 24/7, sans laisser ton PC allumé. Avec Playrena, ton serveur Hytale se déploie en quelques minutes : le jeu est pré-téléchargé sur nos machines, tu n'as qu'à l'autoriser avec ton compte Hytale et inviter tes amis. Voici le tutoriel complet.",
        ],
      },
      {
        h: "Ce qu'il te faut avant de commencer",
        p: [
          "Un compte Discord (pour te connecter au panel Playrena en un clic) et un compte Hytale valide (celui avec lequel tu joues). C'est tout — pas de carte de crédit : l'hébergement Hytale est offert pour l'instant chez Playrena, avec un serveur dédié de 10 Go de RAM et 4 cœurs CPU.",
        ],
      },
      {
        h: "1. Connecte-toi à Playrena",
        p: [
          "Rends-toi sur playrena.vbt-prog.com et clique sur « Se connecter ». Choisis la connexion Discord : en deux clics, ton compte est créé et tu arrives sur ton tableau de bord. Aucun formulaire à remplir.",
        ],
      },
      {
        h: "2. Crée ton serveur Hytale",
        p: [
          "Sur le tableau de bord, clique sur « Créer un serveur », puis sélectionne Hytale dans la liste des jeux. Donne un nom à ton serveur et valide. Le serveur se prépare automatiquement : comme les fichiers du jeu sont déjà présents sur nos serveurs, il n'y a aucun long téléchargement — il démarre en quelques secondes.",
        ],
      },
      {
        h: "3. Autorise ton serveur avec ton compte Hytale",
        p: [
          "Au premier démarrage, Hytale demande d'autoriser le serveur avec ton compte (c'est une sécurité du jeu, comme chez tous les hébergeurs). Le panel affiche une carte « Autorisation Hytale » avec un bouton et un code : clique sur « Autoriser sur Hytale », connecte-toi avec ton compte, entre le code affiché, puis valide.",
          "Dès que tu as autorisé, le serveur passe tout seul en statut « En ligne ». Tu peux suivre la progression en direct dans la console juste en dessous. L'autorisation n'est à faire qu'une seule fois : tes identifiants sont conservés et le serveur redémarre ensuite sans rien te redemander.",
        ],
      },
      {
        h: "4. Connecte-toi et invite tes amis",
        p: [
          "Une fois le serveur en ligne, le panel affiche son adresse de connexion (IP et port). Lance Hytale, ajoute un serveur avec cette adresse, et rejoins ! Partage la même adresse à tes amis : ils se connectent directement, sans rien installer de plus.",
        ],
      },
      {
        h: "Apparaître dans le navigateur de serveurs Hytale (optionnel)",
        p: [
          "Tu veux que ton serveur soit visible publiquement dans la liste de serveurs intégrée à Hytale ? Crée ton listing sur le site officiel de Hytale, récupère ton « discovery token », et colle-le dans la section « Listing public » de ton serveur sur Playrena. En un clic, ton serveur est lié et apparaît dans le navigateur in-game.",
        ],
      },
      {
        h: "Gérer ton serveur au quotidien",
        p: [
          "Depuis le panel, tu accèdes à la console en direct, aux fichiers du serveur, aux sauvegardes, et au redémarrage en un clic. Besoin de plus de puissance pour un gros événement ? Les ressources s'ajustent facilement. Crée ton serveur Hytale gratuitement dès maintenant sur Playrena et commence à jouer en quelques minutes.",
        ],
      },
    ],
  },
  {
    slug: "configuration-serveur-hytale-ram-cpu",
    title: "Configuration requise pour un serveur Hytale : RAM, CPU et joueurs",
    description:
      "Combien de RAM et de CPU pour un serveur Hytale ? Guide des ressources recommandées selon le nombre de joueurs, et comment bien dimensionner ton serveur.",
    date: "2026-06-12",
    keywords: [
      "configuration serveur hytale",
      "ram serveur hytale",
      "cpu serveur hytale",
      "hytale server requirements",
    ],
    sections: [
      {
        p: [
          "Hytale est un serveur Java moderne (Java 25) plus gourmand qu'un serveur Minecraft. Bien dimensionner la RAM et le CPU évite les lags, surtout quand plusieurs joueurs explorent en même temps. Voici les repères.",
        ],
      },
      {
        h: "La RAM : le facteur principal",
        p: [
          "La recommandation officielle est d'environ 4 Go pour 4 joueurs, puis ~1 Go par joueur supplémentaire. En pratique, le principal facteur de consommation est la distance de vue (view distance) et le fait que les joueurs explorent ensemble ou chacun de leur côté. Pour un petit groupe d'amis, 10 Go offrent une marge confortable et stable — c'est l'allocation par défaut chez Playrena.",
        ],
      },
      {
        h: "Le CPU : ne le néglige pas",
        p: [
          "Le serveur Hytale est sensible au CPU : la génération de monde, la physique et les entités sollicitent fortement les cœurs. Un serveur bridé à 2 cœurs peut « laguer » dès que tu fais des actions, même sans autres joueurs. Chez Playrena, les serveurs Hytale disposent de 4 cœurs par défaut pour un gameplay fluide.",
        ],
      },
      {
        h: "Combien de joueurs ?",
        p: [
          "Pour un serveur entre amis (jusqu'à ~10 joueurs), une configuration 10 Go / 4 cœurs est largement suffisante. Pour des dizaines de joueurs (événements communautaires), il faut beaucoup plus de RAM — c'est là qu'un serveur dédié dimensionné pour l'occasion prend tout son sens.",
        ],
      },
      {
        h: "Sur Playrena",
        p: [
          "Pas besoin de calculer : nos serveurs Hytale sont pré-configurés en 10 Go / 4 cœurs, offerts pour l'instant. Tu peux te lancer gratuitement et te concentrer sur le jeu plutôt que sur le matériel.",
        ],
      },
    ],
  },
  {
    slug: "authentifier-serveur-hytale",
    title: "Comment authentifier ton serveur Hytale (device-code expliqué)",
    description:
      "Tout savoir sur l'authentification d'un serveur Hytale : pourquoi elle existe, comment fonctionne le code device, et comment l'effectuer facilement sur Playrena.",
    date: "2026-06-12",
    keywords: [
      "authentifier serveur hytale",
      "hytale server authentication",
      "hytale device code",
      "autorisation serveur hytale",
    ],
    sections: [
      {
        p: [
          "Au premier démarrage, un serveur Hytale doit être « autorisé » avec un compte Hytale valide. C'est une sécurité imposée par le jeu (pour lier le serveur à un compte et limiter les abus), commune à tous les hébergeurs. Bonne nouvelle : c'est rapide et tu n'as à le faire qu'une seule fois.",
        ],
      },
      {
        h: "Comment ça marche : le flux « device code »",
        p: [
          "Le serveur génère un lien d'autorisation et un court code. Tu ouvres le lien dans ton navigateur, tu te connectes avec ton compte Hytale, tu entres le code, et tu valides. Le serveur détecte automatiquement l'autorisation et finit de démarrer. Tes identifiants restent côté serveur — ils ne sont jamais partagés avec d'autres joueurs.",
        ],
      },
      {
        h: "Sur Playrena : tout est dans le panel",
        p: [
          "Pas besoin de toucher à une console technique : quand ton serveur Hytale attend l'autorisation, le tableau de bord affiche une carte claire avec un bouton « Autoriser sur Hytale » et le code à copier. Tu cliques, tu te connectes, tu valides — et le serveur passe en « En ligne » automatiquement. La page se met à jour toute seule.",
        ],
      },
      {
        h: "À retenir",
        p: [
          "L'autorisation est à faire une seule fois par serveur. Ensuite, les redémarrages sont automatiques. Si jamais tu recrées un serveur, il faudra simplement réautoriser — toujours en quelques secondes depuis le panel. Crée ton serveur Hytale sur Playrena et teste par toi-même.",
        ],
      },
    ],
  },
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

// ---- Catégorisation des articles par jeu (sections du blog) ----
export type GameCat = "minecraft" | "hytale" | "satisfactory" | "general";

const POST_GAME: Record<string, GameCat> = {
  "creer-serveur-hytale-tutoriel": "hytale",
  "configuration-serveur-hytale-ram-cpu": "hytale",
  "authentifier-serveur-hytale": "hytale",
  "creer-serveur-minecraft-gratuit": "minecraft",
  "meilleurs-mods-minecraft": "minecraft",
  "paper-vs-vanilla-minecraft": "minecraft",
  "combien-de-ram-serveur-minecraft": "minecraft",
  "reduire-lag-serveur-minecraft": "minecraft",
  "installer-plugins-serveur-minecraft-paper": "minecraft",
  "creer-serveur-satisfactory-dedie": "satisfactory",
  "hytale-vs-minecraft-serveur": "general",
  "hebergement-serveur-jeu-quebec": "general",
};

export function gameOf(slug: string): GameCat {
  return POST_GAME[slug] ?? "general";
}

// Ordre + libellé des sections. La page n'affiche que celles qui ont des articles.
export const BLOG_CATEGORIES: { game: GameCat; label: string }[] = [
  { game: "minecraft", label: "Minecraft" },
  { game: "hytale", label: "Hytale" },
  { game: "satisfactory", label: "Satisfactory" },
  { game: "general", label: "Guides généraux" },
];

export function postsByGame(game: GameCat): Post[] {
  return posts.filter((p) => gameOf(p.slug) === game);
}
