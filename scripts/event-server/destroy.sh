#!/usr/bin/env bash
# Détruit le serveur d'event (ARRÊTE la facturation horaire). À lancer dès l'event fini.
# Usage : ./destroy.sh <nom>   (défaut : playrena-event)
set -euo pipefail
NAME="${1:-playrena-event}"
echo "→ Suppression de $NAME (la facturation s'arrête)…"
hcloud server delete "$NAME"
echo "✅ Détruit. Pense à sauvegarder /opt/hytale-data avant si tu veux garder le monde."
