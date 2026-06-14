# Kit de lancement — obtenir le premier utilisateur
*Playrena · 2026-06-13 · le produit est prêt ; il manque l'acquisition.*

## ✅ Readiness produit (vérifié 2026-06-13)
Le funnel **Minecraft gratuit** marche de bout en bout :
- Home → landing Minecraft (200, hero + CTA « Commencer gratuitement »).
- Inscription **Discord** OK ; création serveur gratuit (1 Go) OK.
- **Connexion externe prouvée** : `*.servers.vbt-prog.com` résout vers l'IP publique, port **25565 forwardé et joignable** (un vrai client externe a atteint mc-router), routes mc-router correctes (node1 par nom de conteneur, node2 par `10.0.0.110:port`).
- Formulaire **support** → courriel livré (testé, HTTP 200).
- **9 articles de blog SEO** en ligne (Minecraft, Hytale, Satisfactory, QC) → trafic organique.

⚠️ **Détail connexion** : il faut utiliser le **sous-domaine** (`nom.servers.vbt-prog.com`), PAS l'IP brute (mc-router route par hostname). Le panel affiche déjà la bonne adresse. (Des inconnus sondent déjà l'IP brute → il y a de l'intérêt organique.)

## 🎯 Actions de Simon pour décrocher le 1ᵉʳ user
1. **Poster les messages ci-dessous** (Reddit, Discord, groupes FB) — c'est LE levier manquant.
2. **Sécurité (5 min)** : révoquer le token Cloudflare exposé + régénérer le bot token Discord.
3. (Optionnel) Forward routeur UDP/TCP `25566-26565 → 10.0.0.2` pour débloquer Satisfactory/Hytale en plus de Minecraft.

---

## Texte de présentation du service (positionnement BÊTA = gratuit)

> **Angle clé** : on assume que le service est jeune. « En test » justifie la gratuité, fixe les
> attentes (pas encore parfait) et crée un sentiment d'exclusivité (« sois parmi les premiers »).

### Version longue (forum, page « à propos », Reddit body)
> **Playrena — hébergement de serveurs de jeux (en bêta, donc gratuit)**
>
> Playrena est un nouvel hébergeur de serveurs de jeux, **fait au Québec**. On est en pleine
> **phase de test (bêta)** — et c'est exactement pour ça que **tout est gratuit en ce moment** :
> on construit le service avec les retours des premiers utilisateurs, avant d'ouvrir les offres payantes.
>
> Concrètement, tu peux **créer ton serveur en 30 secondes** : connexion en un clic avec Discord,
> aucune carte de crédit. Minecraft est dispo (serveur Paper optimisé, 1 Go gratuit, versions 1.8
> à la dernière), et Hytale + Satisfactory aussi pendant la bêta. Ton serveur tourne 24/7, géré
> depuis un panel simple (console, fichiers, sauvegardes, redémarrage en un clic).
>
> En échange, **une seule chose** : ton **feedback honnête** — ce qui marche, ce qui manque, ce qui
> bugue. Les premiers testeurs façonnent le produit (et garderont des avantages quand le payant
> arrivera). 👉 **https://playrena.vbt-prog.com**

### Version courte (Discord, chat)
> 🚀 **Playrena** (bêta) — hébergeur de serveurs de jeux fait au Québec 🇨🇦. **Gratuit pendant la
> phase de test !** Serveur Minecraft prêt en 30 s via Discord, sans carte de crédit. Je cherche
> mes premiers testeurs pour du feedback honnête → https://playrena.vbt-prog.com 🎮

### Titre Reddit (r/MCServers, r/admincraft)
> [Free][Beta] Playrena — hébergement de serveurs Minecraft gratuit pendant la bêta, cherche ses premiers testeurs

---

## 📍 Liste concrète de lieux où poster

> ⚠️ **Règle d'or** : lis les règles de chaque communauté, **n'spamme pas**, présente-toi comme un
> fondateur qui cherche du feedback (pas une pub). Poste 1 endroit à la fois, réponds aux commentaires.

### Reddit (anglophone surtout, gros volume)
- **r/MCServers** — dédié à la promo de serveurs, le plus permissif. ← commencer ici.
- **r/admincraft** — admins de serveurs (ton public : ceux qui hébergent). Très ciblé.
- **r/MinecraftServer**, **r/MinecraftBuddies** (1 post / 7 j), **r/realms**.
- **r/HytaleInfo** / r/Hytale — pour l'angle Hytale (jeu récent, public chaud).

### Discord francophones / QC (ton meilleur terrain pour le 1ᵉʳ user)
- **HytaleQC** (hytaleqc.ca) — communauté **Hytale québécoise** 🎯 (parfait vu qu'on héberge Hytale).
- **Minecraft Francophone** (discord.gg/mcfr) — a un **salon de pub** serveurs/hébergement.
- **Communauté Minecraft Fr (CMFr)** — 3ᵉ plus grande communauté MC FR.
- **Minecraft-France** (discord.gg/mc-fr) — très grande communauté.
- **Discord Québec** (discord.me/quebec) — 7000+ membres QC, salon pub/gaming.
- Chercher d'autres via **disboard.org** (tags `minecraft` + `quebecois`/`quebec`/`français`).

### Annuaires / forums FR
- **serveur-prive.net** — annuaire FR (listing serveur + liste de Discords MC publics).
- **minecraft.fr** forums, groupes **Facebook** « Minecraft FR/Québec ».

### Le plus facile pour le TOUT premier user (commence par là)
- **Ton réseau perso** : amis, **clubs/assos gaming de l'UdeS**, Discord d'étudiants en génie info.
- Demande à 2-3 amis de créer un serveur et de te donner leur ressenti → tes 3 premiers users en 10 min.

---

## Messages prêts à coller

### Reddit — r/admincraft, r/QuebecGaming, r/MinecraftServer, r/Frenchmemes (adapter)
> **[Gratuit] J'ai monté un hébergeur de serveurs Minecraft, je cherche mes premiers testeurs**
>
> Salut ! Je suis un étudiant en génie info au Québec et j'ai créé **Playrena**, un hébergeur de serveurs Minecraft. Tu crées ton serveur **gratuitement en 30 secondes** (connexion Discord, 1 Go RAM, serveur Paper optimisé), prêt 24/7, sans carte de crédit.
>
> Je cherche des premiers utilisateurs pour avoir du feedback honnête. Si tu veux un petit serveur pour jouer avec tes amis, teste et dis-moi ce qui cloche : 👉 https://playrena.vbt-prog.com
>
> (Hytale et Satisfactory aussi dispo gratuitement en ce moment.)

### Discord (communautés MC/gaming FR-QC) — message court
> Hey ! J'ai lancé **Playrena**, un hébergeur de serveurs Minecraft gratuit fait au Québec 🇨🇦 — serveur prêt en 30 s via Discord, gratuit (1 Go), pas de CB. Je cherche mes premiers testeurs pour du feedback : https://playrena.vbt-prog.com 🎮

### Pitch ultra-court (Twitter/X, TikTok bio, signature)
> Playrena — ton serveur Minecraft gratuit, prêt en 30 secondes. Fait au Québec. https://playrena.vbt-prog.com

### Message d'approche micro-créateurs (cf. PARTENAIRES.md §1)
> Salut [Nom] ! Je lance Playrena (hébergement de serveurs de jeux, fait au QC). Je t'offre un **serveur Pro gratuit à vie** + un code promo pour ta communauté, en échange d'une petite mention. Ça t'intéresse de tester ? 🎮

---

## Où poster (ordre de priorité)
1. **Reddit** : r/admincraft (héberge des admins de serveurs), r/MinecraftServer, subs gaming FR/QC. ⚠️ Pas de spam : poste comme un fondateur qui cherche du feedback, pas une pub.
2. **Discord** : serveurs communautaires MC FR/QC (chercher via disboard.org, tag minecraft + fr). Salons #pub / #partenariats.
3. **Groupes Facebook** Minecraft FR/QC.
4. **Réseau perso** (UdeS, clubs gaming) — le plus facile pour un tout premier user.
5. **Micro-créateurs** (1k-50k) — cf. `PARTENAIRES.md`.

## Mesurer
- Surveiller `/admin` (inscriptions, serveurs créés) après chaque post.
- Conversion « douce » Google Ads à l'inscription = backlog **#F** (donner du signal aux ads).
- Demander activement le feedback au 1ᵉʳ user (DM/courriel) → corriger vite les frictions.
