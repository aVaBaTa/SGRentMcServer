#!/usr/bin/env bash
# Provisionne un serveur d'event Hytale JETABLE sur Hetzner Cloud (facturé à l'heure).
# Stratégie PLAYRENA.md #S : louer la machine juste le temps de l'event (~2-10 €),
# puis la détruire avec destroy.sh. Pas sur xe80dell (31 Go insuffisant).
#
# Pré-requis :
#   - hcloud CLI installé (https://github.com/hetznercloud/cli) et `hcloud context create`
#     OU exporter HCLOUD_TOKEN (token API d'un projet Hetzner Cloud).
#   - Une clé SSH ajoutée au projet (hcloud ssh-key list) → SSH_KEY ci-dessous.
#
# Usage : ./provision.sh <25|50|100> [nom] [location]
#   ex.  ./provision.sh 100 "Event-CaseOh" ash   # 100 joueurs, datacenter Ashburn (US-East)
#
# Coût indicatif (Hetzner Cloud, à l'heure, plafonné au mois) :
#   25 j  → ccx33  32 Go  ~0,10 €/h     50 j → ccx43  64 Go ~0,20 €/h
#   100 j → ccx53 128 Go  ~0,40 €/h    (CCX = vCPU dédiés, requis pour un jeu temps réel)
set -euo pipefail

PLAYERS="${1:?Usage: ./provision.sh <25|50|100> [nom] [location]}"
NAME="${2:-playrena-event}"
LOCATION="${3:-ash}"            # ash=Ashburn(US-East), fsn1=Falkenstein(DE), hel1=Helsinki(FI)
SSH_KEY="${SSH_KEY:-playrena}"  # nom de la clé SSH dans Hetzner Cloud
SERVER_PORT="${SERVER_PORT:-5520}"
SERVER_NAME="${NAME}"

# Map joueurs → type d'instance + RAM allouée au JVM (laisse de la marge à l'OS).
case "$PLAYERS" in
  25)  TYPE="ccx33"; MEMORY_MB=24576;  MAX_PLAYERS=25  ;;   # 32 Go box
  50)  TYPE="ccx43"; MEMORY_MB=52224;  MAX_PLAYERS=50  ;;   # 64 Go box
  100) TYPE="ccx53"; MEMORY_MB=114688; MAX_PLAYERS=100 ;;   # 128 Go box
  *)   echo "Joueurs doit être 25, 50 ou 100"; exit 1 ;;
esac

export SERVER_PORT MEMORY_MB MAX_PLAYERS SERVER_NAME
TPL_DIR="$(cd "$(dirname "$0")" && pwd)"
USERDATA="$(mktemp)"
trap 'rm -f "$USERDATA"' EXIT
# Substitue uniquement nos variables (pas le reste du YAML).
envsubst '${SERVER_PORT} ${MEMORY_MB} ${MAX_PLAYERS} ${SERVER_NAME}' \
  < "$TPL_DIR/cloud-init.yaml.tpl" > "$USERDATA"

echo "→ Création $NAME ($TYPE, $PLAYERS joueurs, $LOCATION, port UDP $SERVER_PORT)…"
hcloud server create \
  --name "$NAME" \
  --type "$TYPE" \
  --image ubuntu-24.04 \
  --location "$LOCATION" \
  --ssh-key "$SSH_KEY" \
  --user-data-from-file "$USERDATA"

IP="$(hcloud server ip "$NAME")"
cat <<EOF

✅ Serveur créé : $NAME — IP $IP
   Adresse de jeu à donner au créateur : $IP:$SERVER_PORT (UDP)

⏳ ~3-5 min pour le 1er téléchargement Hytale, puis OAuth device-code :
   1) ssh root@$IP
   2) docker logs -f hytale-event        # copie l'URL + le code
   3) ouvre l'URL, connecte-toi (compte Hytale licencié), autorise
   (creds persistés dans /opt/hytale-data → pas de re-auth au redémarrage)

💸 N'OUBLIE PAS après l'event :  ./destroy.sh "$NAME"
EOF
