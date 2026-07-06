#!/usr/bin/env bash
# Ré-applique les mappings UPnP Calradia sur la box Helix (XB7). Idempotent :
# AddPortMapping identique = refresh. Lancé par cron (toutes les 6 h) pour
# survivre aux redémarrages du routeur, qui vident la table UPnP.
set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
DESC="http://10.0.0.1:49152/IGDdevicedesc_brlan0.xml"
TO="10.0.0.2"

"$DIR/upnp-map.py" add 7777 both --to "$TO" --lease 0 --desc "$DESC"
"$DIR/upnp-map.py" add 7778 tcp  --to "$TO" --lease 0 --desc "$DESC"
# Blocs de jeu loués (25566-26565) : la plage complète dépasse la table UPnP du
# routeur → règle "port range forwarding" manuelle dans l'admin Helix, OU mapper
# ici le bloc d'un serveur précis, ex. :
#   "$DIR/upnp-map.py" add 25566-25581 both --to "$TO" --lease 0 --desc "$DESC"
