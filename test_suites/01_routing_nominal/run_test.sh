#!/bin/bash
# Suite 01: Routing Nominal (HTTP REST Drop-In)
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

RESP=$(curl -s --max-time 3 http://127.0.0.1:5432/api/data 2>/dev/null || echo "EMPTY")

if [[ "$RESP" == *"Legacy"* ]] || [[ "$RESP" == *"alienigena"* ]]; then
    echo "01_routing_nominal: PASS" > "$REPORTS/01_status.log"
    echo "PASS"
else
    echo "01_routing_nominal: FAIL (resp=$RESP)" > "$REPORTS/01_status.log"
    echo "FAIL"
fi

kill $PID1 $PID2 $PID3 2>/dev/null
