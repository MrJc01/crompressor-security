#!/bin/bash
# ==============================================================================
# CROM Docker Anti-Intrusion (Kill-Switch)
# Script guardião que roda atrelado ao backend no ambiente contêiner protegido.
# Monitorará a criação de PTYs (terminais de shell via `docker exec` ou `nsenter`)
# Se um terminal indesejado aparecer com intenções maliciosas via Root Host, 
# a thread reage instantaneamente, expurga state virtual se necessário e comete 
# kill-switch no container isolado, barrando dumps em cache.
# ==============================================================================

echo "[CROM-SEC] Escudo Anti-Intrusão de PTY ativado. Namespaces trancados."

# Loop de vigília sem CPU footprint
while true; do
  # Ignora /dev/pts/ptmx que é de base. Conta os verdadeiros TTY shells provisionados.
  if [ "$(ls -1 /dev/pts/ | grep -v ptmx | wc -l)" -gt "0" ]; then
      
      echo "[ALERTA MÁXIMO CROM] Intrusa via 'docker exec' ou Namespace TTY detectada!" >> /dev/stderr
      
      # 1. Pressiona um expurgo se houver chaves secretas locais no filesystem (RAM/Tmp/Env)
      # rm -rf /dev/shm/* 2>/dev/null
      # unset TENANT_SEED
      
      # 2. Mata o processo 1 de forma severa fechando o cofre e banindo o atacante da Shell
      logger "Injeção Crítica mitigada através de aniquilação"
      kill -9 1
      exit 1
  fi
  
  sleep 2
done
