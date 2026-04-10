#!/bin/bash
# ==========================================
# CROM-SECURITY MASTER AUDIT (20 SCENARIOS)
# ==========================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "=========================================================="
echo " INICIANDO AUDITORIA CROM-SEC AUTOMATIZADA (20 SCENARIOS)"
echo " Data: $(date)"
echo "=========================================================="

# 0. Matar TUDO de sessões anteriores
killall -9 proxy_in proxy_out dummy_backend alien_sniffer tcp_cannon proxy_onion_relay python3 nc 2>/dev/null || true
fuser -k 5432/tcp 8080/tcp 9999/tcp 9955/tcp 8082/tcp 8085/tcp 6379/tcp 2>/dev/null || true
sleep 1.0

# 1. COMPILAÇÃO AOT (Ahead-of-Time) dos binários Go
echo "[*] Compilando Simuladores em Binários Nativos..."
mkdir -p bin
go build -o bin/proxy_in ../simulators/dropin_tcp/proxy_universal_in.go 2>/dev/null
go build -o bin/proxy_out ../simulators/dropin_tcp/proxy_universal_out.go 2>/dev/null
go build -o bin/dummy_backend ../simulators/dropin_tcp/dummy_backend.go 2>/dev/null
go build -o bin/alien_sniffer ../simulators/pentest/alien_sniffer.go 2>/dev/null
go build -o bin/tcp_cannon ../simulators/pentest/tcp_cannon.go 2>/dev/null
go build -o bin/proxy_onion_relay ../simulators/dropin_tcp/proxy_onion_relay.go 2>/dev/null
echo "[*] Builds concluídas."

# Permissões
find . -name "run_test.sh" -exec chmod +x {} \;

# Limpar logs anteriores
rm -f reports/*.log
mkdir -p reports

# 2. ITERAÇÃO SEQUENCIAL COM TIMEOUT POR SUITE
SUITES=(
    01_routing_nominal
    02_pentest_mitm
    03_pentest_dos_cannon
    04_websocket_chat
    05_php_fpm_cgi
    06_python_grpc
    07_postgres_pgwire
    08_redis_resp
    09_iot_mqtt_broker
    10_cplusplus_raw_tcp
    11_large_payload_chunking
    12_high_concurrency
    13_sybil_swarm_attack
    14_silent_drop_validation
    15_split_brain_recovery
    16_nodejs_express_rest
    17_java_spring_boot_xml
    18_dns_hijack_spoofing
    19_payload_forgery
    20_vfs_fd_exhaust
    21_private_brain_system
    22_onion_multi_hop_route
    23_jitter_cover_traffic
)

echo ""
echo "=== ITERANDO ${#SUITES[@]} SUÍTES COM TIMEOUT DE 20s ==="

for suite in "${SUITES[@]}"; do
    if [ -d "$suite" ] && [ -f "$suite/run_test.sh" ]; then
        echo -n " -> [$suite] "
        timeout 20 bash "$suite/run_test.sh" 2>/dev/null
        EXIT_CODE=$?
        if [ $EXIT_CODE -eq 124 ]; then
            echo "TIMEOUT"
            # Extrair o número da suite
            NUM="${suite:0:2}"
            echo "${suite}: TIMEOUT" > "reports/${NUM}_status.log"
        fi
        # Cleanup agressivo entre suites
        killall -9 proxy_in proxy_out dummy_backend alien_sniffer tcp_cannon proxy_onion_relay python3 nc 2>/dev/null || true
        fuser -k 5432/tcp 8080/tcp 9999/tcp 9955/tcp 8082/tcp 8085/tcp 6379/tcp 2>/dev/null || true
        sleep 1.0
    else
        echo " -> [$suite] NÃO IMPLEMENTADO"
        NUM="${suite:0:2}"
        echo "${suite}: NAO_IMPLEMENTADO" > "reports/${NUM}_status.log"
    fi
done

# 3. CONSOLIDAÇÃO DO RELATÓRIO MESTRE
echo ""
echo "=== RELATÓRIO MESTRE (/reports/FINAL_AUDIT_REPORT.txt) ==="

REPORT_FILE="reports/FINAL_AUDIT_REPORT.txt"

{
  echo "================================================================"
  echo " Relatório de Auditoria P2P CROM-SEC"
  echo " Data: $(date)"
  echo " Hostname: $(hostname)"
  echo "================================================================"
  echo ""
  TOTAL=0
  PASS=0
  FAIL=0
  TIMEOUT=0
  DEV=0
  NI=0
  
  for i in $(seq -w 1 23); do
      STATUS_FILE="reports/${i}_status.log"
      if [ -f "$STATUS_FILE" ]; then
          LINE=$(cat "$STATUS_FILE")
          echo "  $LINE"
          TOTAL=$((TOTAL+1))
          case "$LINE" in
              *PASS*) PASS=$((PASS+1)) ;;
              *FAIL*) FAIL=$((FAIL+1)) ;;
              *TIMEOUT*) TIMEOUT=$((TIMEOUT+1)) ;;
              *DESENVOLVIMENTO*) DEV=$((DEV+1)) ;;
              *NAO_IMPLEMENTADO*) NI=$((NI+1)) ;;
          esac
      fi
  done
  
  echo ""
  echo "================================================================"
  echo " RESUMO EXECUTIVO"
  echo "================================================================"
  echo "  Total de Suítes:     $TOTAL"
  echo "  APROVADAS (PASS):    $PASS"
  echo "  REPROVADAS (FAIL):   $FAIL"
  echo "  TIMEOUT:             $TIMEOUT"
  echo "  EM DESENVOLVIMENTO:  $DEV"
  echo "  NÃO IMPLEMENTADAS:   $NI"
  echo "================================================================"
} > "$REPORT_FILE"

cat "$REPORT_FILE"

echo ""
echo "Auditoria CROM-SEC finalizada. Relatório salvo."
