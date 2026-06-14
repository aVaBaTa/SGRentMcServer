#!/usr/bin/env bash
# Pré-télécharge les fichiers de jeu Hytale UNE FOIS sur l'hôte, pour les réutiliser
# à chaque création de serveur (plus de re-download ni d'OAuth "downloader" côté client).
#
# Ce que ça fait :
#   1) lance un container temporaire qui télécharge Hytale dans ./seeds/_hytale-dl
#   2) TU autorises l'OAuth "downloader" UNE fois (compte Hytale licencié) — voir logs
#   3) une fois HytaleServer.jar + Assets.zip présents, on les copie dans ./seeds/hytale/
#   4) le backend (SEED_DIR=/seeds) seede ces fichiers dans chaque nouveau volume Hytale
#      → le client n'a plus QUE l'auth SERVEUR à faire (device-code dans le panel).
#
# Usage :  ./scripts/hytale-seed/download.sh
# Pré-requis : Docker + un compte Hytale licencié pour l'autorisation unique.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
DL_DIR="$REPO_ROOT/seeds/_hytale-dl"      # zone de téléchargement temporaire
SEED_DIR="$REPO_ROOT/seeds/hytale"        # destination finale (lue par l'API)
IMAGE="ghcr.io/terkea/hytale-server:latest"
CONTAINER="hytale-seed-dl"
FILES=(HytaleServer.jar Assets.zip)

mkdir -p "$DL_DIR" "$SEED_DIR"
chmod 777 "$DL_DIR"                        # le container tourne en UID 1000

echo "→ Lancement du téléchargeur Hytale (container temporaire)…"
docker rm -f "$CONTAINER" >/dev/null 2>&1 || true
docker run -d --name "$CONTAINER" \
  -e AUTO_DOWNLOAD=true -e AUTO_UPDATE=false \
  -v "$DL_DIR:/data" \
  "$IMAGE" >/dev/null

cat <<EOF

================================================================
  AUTORISE LE TÉLÉCHARGEMENT (une seule fois)
  1) docker logs -f $CONTAINER
  2) ouvre l'URL affichée, connecte-toi (compte Hytale licencié), autorise
  3) laisse télécharger (plusieurs centaines de Mo)
================================================================
EOF

echo "→ Attente de ${FILES[*]} dans $DL_DIR (Ctrl-C pour annuler)…"
# L'image extrait HytaleServer.jar d'abord dans Server/ puis le déplace à la racine.
# On le cherche aux deux endroits.
find_seed() {
  if [ -s "$DL_DIR/$1" ]; then echo "$DL_DIR/$1"; return 0; fi
  if [ -s "$DL_DIR/Server/$1" ]; then echo "$DL_DIR/Server/$1"; return 0; fi
  return 1
}

while :; do
  ok=1
  for f in "${FILES[@]}"; do
    find_seed "$f" >/dev/null || ok=0
  done
  # Exige que l'extraction soit FINIE (game.zip supprimé) → évite de copier un
  # Assets.zip encore en cours d'écriture (fichier partiel/corrompu).
  [ -e "$DL_DIR/game.zip" ] && ok=0
  [ "$ok" = 1 ] && break
  sleep 5
done

echo "→ Fichiers prêts (extraction terminée). Copie vers $SEED_DIR…"
for f in "${FILES[@]}"; do
  cp -f "$(find_seed "$f")" "$SEED_DIR/$f"
  echo "   ✓ $f ($(du -h "$SEED_DIR/$f" | cut -f1))"
done

echo "→ Nettoyage du container temporaire…"
docker rm -f "$CONTAINER" >/dev/null 2>&1 || true

cat <<EOF

✅ Seed Hytale prêt dans $SEED_DIR
   Le backend le montera en /seeds (SEED_DIR). Rebuild/redeploy l'API si besoin :
     docker compose up -d --build api
   Dès lors, chaque serveur Hytale réutilise ces fichiers (pas de re-download) ;
   le client n'a que l'auth SERVEUR à faire (device-code via le panel).

   (Optionnel) supprime la zone temporaire : rm -rf "$DL_DIR"
EOF
