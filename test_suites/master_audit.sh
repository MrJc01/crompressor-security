#!/bin/bash
# ==========================================
# CROM-SECURITY MASTER AUDIT (SUPER SUITE)
# ==========================================

echo "=========================================================="
echo " INICIANDO AUDITORIA CROM-SEC AUTOMATIZADA (15 SCENARIOS)"
echo "=========================================================="

cd "$(dirname "$0")"

# 1. PERFOMANCE TUNING: Compilando os simuladores nativos UMA VEZ. (Ahead of time)
echo "[*] [BUILD START] Compilando Simuladores proxy do CROM-SEC em Binários Nativos Ligeiros..."
mkdir -p bin
go build -o bin/proxy_in ../simulators/dropin_tcp/proxy_universal_in.go
go build -o bin/proxy_out ../simulators/dropin_tcp/proxy_universal_out.go
go build -o bin/dummy_backend ../simulators/dropin_tcp/dummy_backend.go
go build -o bin/alien_sniffer ../simulators/pentest/alien_sniffer.go
go build -o bin/tcp_cannon ../simulators/pentest/tcp_cannon.go
echo "[*] [BUILD DONE] Binários em cache. Velocidade Máxima Destravada."

# Permissões Mestre
chmod +x */*.sh 2>/dev/null || true
rm -f reports/*.log
rm -f reports/FINAL_AUDIT_REPORT.txt
mkdir -p reports

# =================
# SUITES RUNNERS LOOP
# =================
echo ""
echo "=== INICIANDO ITERAÇÕES DA MATRIX CROM-SEC ==="

for ext_dir in 01_routing_nominal 02_pentest_mitm 03_pentest_dos_cannon 04_websocket_chat 05_php_fpm_cgi 06_python_grpc 07_postgres_pgwire 08_redis_resp 09_iot_mqtt_broker 10_cplusplus_raw_tcp 11_large_payload_chunking 12_high_concurrency 13_sybil_swarm_attack 14_silent_drop_validation 15_split_brain_recovery; do
  
  if [ -d "$ext_dir" ] && [ -f "$ext_dir/run_test.sh" ]; then
    echo " -> Disparando Suite: $ext_dir"
    cd "$ext_dir"
    ./run_test.sh
    cd ..
  else
     echo " -> Suite $ext_dir aguardando deployment de scripts..."
  fi
done

# =================
# CONSOLIDAÇÃO RELATÓRIO
# =================
echo ""
echo "=== GERANDO RELATÓRIO MESTRE (/reports/FINAL_AUDIT_REPORT.txt) ==="

REPORT_FILE="reports/FINAL_AUDIT_REPORT.txt"

{
  echo "Relatório Automático do Ecossistema P2P CROM-SEC - $(date)"
  echo "==================================================================="
  for suite in $(ls -d 0* 1* 2>/dev/null | sort); do
     status_file="reports/${suite:0:2}_status.log"
     if [ -f "$status_file" ]; then
         cat "$status_file"
     else
         echo "${suite:0:2}_${suite:3}: EM DESENVOLVIMENTO"
     fi
  done
  echo "==================================================================="
} > "$REPORT_FILE"

cat "$REPORT_FILE"

echo "Limpeza Geral de Segurança SRE Killall"
killall proxy_in proxy_out dummy_backend alien_sniffer tcp_cannon 2>/dev/null

echo "Super Auditoria Concluída."
