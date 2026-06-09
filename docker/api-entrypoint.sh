#!/bin/sh
# Copie les fichiers SSH montés vers /root/.ssh avec l'ownership root (ssh l'exige).
if [ -d /etc/sgrent-ssh ]; then
  mkdir -p /root/.ssh
  cp /etc/sgrent-ssh/* /root/.ssh/ 2>/dev/null || true
  chown -R root:root /root/.ssh
  chmod 700 /root/.ssh
  chmod 600 /root/.ssh/* 2>/dev/null || true
fi
exec /api
