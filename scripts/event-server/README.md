# POC — Serveur d'event Hytale jetable (cloud à l'heure)

Provisionne un serveur Hytale **25 / 50 / 100 joueurs** sur **Hetzner Cloud** facturé
**à l'heure**, le temps d'un event sponsorisé (cf. `PLAYRENA.md` #S), puis le détruit.
**Pourquoi pas xe80dell :** 31 Go de RAM total → incapable de tenir 25 j (~25 Go) et a
fortiori 50/100 j. Le cloud à l'heure coûte **~2-10 € par event**, c'est dérisoire.

## Pourquoi le cloud règle 2 problèmes de la maison
- **Pas de NAT** : IP publique directe → le port UDP s'ouvre via `ufw` (cloud-init le fait),
  fini le galère NAT routeur de xe80dell.
- **Anti-DDoS inclus** (Hetzner basique / OVH "Game") → ne pas exposer l'IP résidentielle (cf. #T).

## Coût (Hetzner Cloud, à l'heure, plafonné au mois)
| Joueurs | Type | RAM | €/h | Event 5 h | Week-end (48 h) |
|---|---|---|---|---|---|
| 25  | ccx33 | 32 Go  | 0,10 € | ~0,50 € | ~5 €  |
| 50  | ccx43 | 64 Go  | 0,20 € | ~1 €    | ~10 € |
| 100 | ccx53 | 128 Go | 0,40 € | ~2 €    | ~20 € |

> `ccx*` = vCPU **dédiés** (requis pour un jeu temps réel). Latence : `ash` (Ashburn, US-East)
> pour l'Amérique du Nord ; OVH Beauharnois (Montréal) serait encore mieux pour le QC (autre script à faire).

## Pré-requis (une fois)
1. Compte Hetzner Cloud + un **projet** → générer un **token API**.
2. `hcloud` CLI : `hcloud context create playrena` (colle le token).
3. Ajouter ta clé SSH : `hcloud ssh-key create --name playrena --public-key-from-file ~/.ssh/id_ed25519.pub`

## Lancer un event
```bash
cd scripts/event-server
./provision.sh 100 "Event-NomCreateur" ash     # crée la machine + démarre Hytale
# → suivre les instructions : ssh, docker logs, autoriser l'OAuth device-code
# ... event ...
./destroy.sh "Event-NomCreateur"               # STOPPE la facturation
```

## ⚠️ Limites connues du POC (à durcir avant prod)
- **OAuth manuel** : l'opérateur doit autoriser le device-code via `docker logs` au 1er boot.
  Pour du 0-click multi-serveurs → modèle **token provider** (cf. Server Provider Auth Guide,
  noté dans PLAYRENA.md) : pré-auth une fois → injecter `HYTALE_SERVER_SESSION_TOKEN` /
  `HYTALE_SERVER_IDENTITY_TOKEN`. Pas nécessaire pour un event ponctuel.
- **Non testé live** (pas de token Hetzner dans ce dépôt) : scripts complets et corrects en
  structure, à valider avec un vrai token avant tout démarchage de créateur.
- **Sauvegarde du monde** : `/opt/hytale-data` est perdu à la destruction → `scp` avant si besoin.
- **DNS** : on donne l'`IP:port` brute. Optionnel : pointer `event.playrena.vbt-prog.com` (DNS-only) dessus.
- **Intégration panel** : ce POC est *hors* du panel Playrena. L'industrialiser = endpoint
  orchestrateur "event" (lié #M CRUD + #S).
