#cloud-config
# Template cloud-init pour un serveur d'event Hytale jetable (loué à l'heure).
# Les variables ${...} sont substituées par provision.sh (envsubst).
# À la 1re création, le serveur Hytale demande un OAuth device-code : voir README
# (récupérer le code via `docker logs hytale-event` puis autoriser sur le navigateur).

package_update: true
packages:
  - docker.io
  - ufw

write_files:
  # Service systemd qui (re)lance le container Hytale au boot.
  - path: /etc/systemd/system/hytale-event.service
    content: |
      [Unit]
      Description=Hytale event server (Playrena)
      After=docker.service
      Requires=docker.service
      [Service]
      Restart=always
      ExecStartPre=-/usr/bin/docker rm -f hytale-event
      ExecStart=/usr/bin/docker run --rm --name hytale-event \
        -p ${SERVER_PORT}:${SERVER_PORT}/udp \
        -e SERVER_PORT=${SERVER_PORT} \
        -e MEMORY=${MEMORY_MB}M \
        -e MAX_PLAYERS=${MAX_PLAYERS} \
        -e SERVER_NAME="${SERVER_NAME}" \
        -e AUTO_DOWNLOAD=true -e AUTO_UPDATE=true \
        -v /opt/hytale-data:/data \
        ghcr.io/terkea/hytale-server:latest
      ExecStop=/usr/bin/docker stop hytale-event
      [Install]
      WantedBy=multi-user.target

runcmd:
  # Pare-feu : SSH + le port de jeu UDP (IP publique cloud, pas de NAT).
  - ufw allow OpenSSH
  - ufw allow ${SERVER_PORT}/udp
  - ufw --force enable
  - systemctl enable docker
  - mkdir -p /opt/hytale-data
  - systemctl daemon-reload
  - systemctl enable --now hytale-event.service

final_message: "Hytale event server prêt. Port UDP ${SERVER_PORT}. Récupère l'OAuth : docker logs hytale-event"
