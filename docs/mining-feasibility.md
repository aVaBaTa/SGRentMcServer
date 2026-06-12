# Note de faisabilité — « Mining de serveurs » rémunéré (hébergement distribué)
*(backlog PLAYRENA.md #R — 2026-06-12, à valider/itérer)*

## Idée
Un réseau où des **particuliers prêtent leur machine** (CPU/RAM/bande passante) pour héberger
des serveurs de jeu Playrena, et sont **rémunérés** en retour. Modèle inspiré de Salad,
Honeygain, ou des pools de hosting communautaire. But : **capacité élastique sans capex**
(utile pour absorber les pics, ex. events Hytale 100 joueurs #S).

## Verdict court
**Faisable techniquement mais gros projet R&D, et l'économie est douteuse pour du jeu temps réel.**
👉 Ne PAS coder maintenant. Pertinent seulement si la demande dépasse durablement la capacité de
xe80dell **et** que le cloud à l'heure (#S) ne suffit plus. À ce stade : note de faisabilité only.

## Pourquoi c'est dur (les 4 murs)

### 1. Réseau — le mur principal
- Les contributeurs sont derrière **NAT résidentiel / CGNAT** → pas de port entrant.
- Le **trafic de jeu est UDP brut** et **ne passe pas par Cloudflare** (déjà notre blocage actuel).
- Il faut donc un **relais** : chaque node ouvre un **tunnel sortant** vers un point d'entrée
  public Playrena (WireGuard / FRP / `cloudflared`… mais CF Tunnel = TCP/HTTP surtout).
  → on paie quand même des **relais publics** (la "capacité gratuite" l'est moins qu'elle en a l'air),
  et chaque hop ajoute de la **latence** — rédhibitoire pour un jeu d'action.

### 2. Latence & fiabilité — incompatible avec le temps réel
- Machines résidentielles = uplink asymétrique faible, bufferbloat, uptime non garanti,
  PC éteint en pleine partie. Acceptable pour du **batch/rendu** (Salad), **pas** pour un
  serveur de jeu où 50 joueurs comptent sur la stabilité.
- Un host qui crashe en plein event = désastre réputationnel.

### 3. Sécurité — bidirectionnelle
- On exécute **notre** charge (et celle de clients) sur la machine **d'un tiers** non fiable
  (vol de données joueurs, triche, sniff réseau) **ET** le tiers exécute du code chez lui
  (risque pour lui). Isolation forte requise (gVisor/Kata/VM), supervision, signature d'images.

### 4. Économie & légal
- Combien payer le contributeur (par Go·h ? par uptime ?) vs un **VPS Hetzner à 0,40 €/h** qui
  est fiable, sans NAT, avec anti-DDoS ? Le cloud à l'heure est déjà **très bon marché** (#S).
- **Fiscalité / paiements** à des particuliers, CGU, responsabilité si un node héberge du contenu illicite.

## Ce qu'on a déjà qui aide
- `internal/orchestrator` gère **déjà N nodes** (socket local + SSH node2) → l'abstraction
  multi-node existe ; l'étendre à des nodes non fiables est une évolution, pas une réécriture.
- Modèle Docker par serveur → portable sur n'importe quel node avec Docker.

## Alternatives plus simples au même besoin (capacité élastique)
1. **Cloud à l'heure** (#S, déjà chiffré ~0,40 €/h pour 128 Go) — couvre les pics events **aujourd'hui**.
2. **Partenaires de confiance** (#O) : 2-3 amis/membres avec une bonne machine + bonne connexion,
   ajoutés comme **nodes SSH** (réutilise l'orchestrateur tel quel) — 90 % du bénéfice, 10 % du risque.
3. **Upgrade RAM xe80dell** (31 → 128 Go, capex modeste) si la demande Minecraft sature.

## Reco
- **Court terme** : ignorer le mining. Utiliser cloud à l'heure (#S) + éventuellement 1-2 nodes
  partenaires de confiance (#O) via l'orchestrateur existant.
- **Re-évaluer #R** seulement si : demande récurrente >> capacité, coûts cloud qui explosent,
  et tolérance à un gros chantier réseau/sécurité. Même alors, viser d'abord du **non-temps-réel**
  (backups, world-gen, rendu de cartes) plutôt que des serveurs de jeu live.
