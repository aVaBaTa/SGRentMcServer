#!/usr/bin/env bash
# Déploiement reproductible de Playrena sur xe80dell (api + monitor + frontend).
# Évite les oublis d'env/mounts/réseaux (sinon des features cassent silencieusement :
# SSR promo, seed Hytale, admin, etc.). Lance depuis la racine du repo.
#
# Usage :
#   ./scripts/deploy.sh            # tout (api + monitor + frontend)
#   ./scripts/deploy.sh api        # backend API seulement
#   ./scripts/deploy.sh monitor    # /admin seulement
#   ./scripts/deploy.sh frontend   # site Next.js seulement
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
[ -f .env ] && set -a && . ./.env && set +a   # POSTGRES_PASSWORD, ADMIN_TOKEN, …

NET_PORTFOLIO="sgportfolio_portfolio-net"
NET_SGRENT="sgrentmcserver_sgrent"
NODES="node1=unix:///var/run/docker.sock,node2=ssh://simon@10.0.0.110"
DBURL="postgres://sgrent:${POSTGRES_PASSWORD}@mcserver-postgres:5432/sgrentmc?sslmode=disable"

target="${1:-all}"

deploy_api() {
  echo "▶ API…"
  docker compose up -d --build api
}

deploy_monitor() {
  echo "▶ monitor (/admin)…"
  docker build -f Dockerfile.monitor -t mcserver-monitor:latest .
  docker rm -f mcserver-monitor >/dev/null 2>&1 || true
  docker run -d --name mcserver-monitor --restart unless-stopped \
    -e DOCKER_NODES="$NODES" -e MONITOR_PORT=8090 \
    -e DATABASE_URL="$DBURL" -e API_BASE="http://mcserver-api:8080" \
    -e ADMIN_TOKEN="$ADMIN_TOKEN" -e HOST_ROOT=/host \
    -v /:/host -v "$ROOT/secrets/ssh:/etc/sgrent-ssh" -v /var/run/docker.sock:/var/run/docker.sock \
    --network "$NET_PORTFOLIO" mcserver-monitor:latest
  docker network connect "$NET_SGRENT" mcserver-monitor 2>/dev/null || true
}

deploy_frontend() {
  echo "▶ frontend (Next.js standalone)…"
  ( cd frontend && docker build -t mcserver-frontend:latest . )
  docker rm -f mcserver-frontend >/dev/null 2>&1 || true
  # INTERNAL_API_URL = SSR de la promo (#Y). Ne pas l'oublier.
  docker run -d --name mcserver-frontend --restart unless-stopped \
    -e INTERNAL_API_URL="http://mcserver-api:8080" \
    --network "$NET_PORTFOLIO" mcserver-frontend:latest
}

case "$target" in
  all)      deploy_api; deploy_monitor; deploy_frontend ;;
  api)      deploy_api ;;
  monitor)  deploy_monitor ;;
  frontend) deploy_frontend ;;
  *) echo "cible inconnue: $target (all|api|monitor|frontend)"; exit 1 ;;
esac

echo ""
echo "✅ Déploiement terminé. État :"
docker ps --format '{{.Names}}\t{{.Status}}' | grep -E 'mcserver-(api|monitor|frontend)' || true
echo "Vérif rapide :"
curl -s -o /dev/null -w "  / → %{http_code}\n" https://playrena.vbt-prog.com/ || true
