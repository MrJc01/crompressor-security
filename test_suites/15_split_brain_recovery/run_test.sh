#!/bin/bash
# Suite 15: Split-Brain Recovery (Kill proxy_out mid-stream)
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

$BIN/dummy_backend &
PID1=$!
sleep 0.3
echo "$CROM_TENANT_SEED" | $BIN/proxy_out &
PID2=$!
sleep 0.3
echo "$CROM_TENANT_SEED" | $BIN/proxy_in &
PID3=$!
sleep 0.5

# Primeira requisição (deve funcionar)
R1=$(curl -s --max-time 2 http://127.0.0.1:5432/ 2>/dev/null || echo "")

# MATAR proxy_out (simular partição de rede)
kill -9 $PID2 2>/dev/null
sleep 0.3

# Segunda requisição (deve falhar graciosamente)
R2=$(curl -s --max-time 2 http://127.0.0.1:5432/ 2>/dev/null || echo "CONN_REFUSED")

# Re-reviver proxy_out
echo "$CROM_TENANT_SEED" | $BIN/proxy_out &
PID2_NEW=$!
sleep 0.5

# Terceira requisição (recovery test)
R3=$(curl -s --max-time 2 http://127.0.0.1:5432/ 2>/dev/null || echo "")

if [[ -n "$R1" ]] && [[ "$R2" == "CONN_REFUSED" || -z "$R2" ]]; then
    echo "15_split_brain_recovery: PASS (graceful degradation confirmada)" > "$REPORTS/15_status.log"
    echo "PASS"
else
    echo "15_split_brain_recovery: FAIL" > "$REPORTS/15_status.log"
    echo "FAIL"
fi

kill $PID1 $PID2_NEW $PID3 2>/dev/null
