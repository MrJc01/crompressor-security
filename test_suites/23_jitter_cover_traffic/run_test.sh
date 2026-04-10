#!/bin/bash
# Suite 23: Jitter Cover-Traffic Anti-NSA
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

$BIN/dummy_backend &
PID_BACK=$!
sleep 1.0

echo "$CROM_TENANT_SEED" | $BIN/proxy_out &
PID_OUT=$!
sleep 1.0

echo "$CROM_TENANT_SEED" | $BIN/proxy_in &
PID_IN=$!
sleep 2.0 # Aguarda alguns segundos para a Goroutine do proxy_in floodar o Omega com lixo JITT

# O trafego real flui pelo fluxo misturado:
RESP=$(curl -s --max-time 3 http://127.0.0.1:5432/api/data 2>/dev/null || echo "EMPTY")

if [[ "$RESP" == *"success"* ]]; then
    echo "23_jitter_cover_traffic: PASS (Névoa não atrapalhou o fluxo semântico)" > "$REPORTS/23_status.log"
    echo "PASS"
else
    echo "23_jitter_cover_traffic: FAIL" > "$REPORTS/23_status.log"
    echo "FAIL"
fi

kill -9 $PID_BACK $PID_OUT $PID_IN 2>/dev/null
