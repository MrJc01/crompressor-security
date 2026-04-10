#!/bin/bash
# Suite 22: Roteamento Mixnet Onion (Multi-Hop)
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

# Backend
$BIN/dummy_backend &
PID_BACK=$!
sleep 1.0

# Omega (9999) apontando pro 8080
$BIN/proxy_out &
PID_OUT=$!
sleep 1.0

# Onion Relay (9955) cego, apontando pro Omega (9999)
$BIN/proxy_onion_relay &
PID_ONION=$!
sleep 1.0

# Alpha (5432) atirando no Onion (9955) inves de ir no 9999 !
export SWARM_CLOUD_TARGET="127.0.0.1:9955"
$BIN/proxy_in &
PID_IN=$!
sleep 1.0

RESP=$(curl -s --max-time 3 http://127.0.0.1:5432/api/data 2>/dev/null || echo "EMPTY")

if [[ "$RESP" == *"success"* ]]; then
    echo "22_onion_multi_hop_route: PASS (Atravessou Onion Route impecável)" > "$REPORTS/22_status.log"
    echo "PASS"
else
    echo "22_onion_multi_hop_route: FAIL" > "$REPORTS/22_status.log"
    echo "FAIL"
fi

kill -9 $PID_BACK $PID_OUT $PID_ONION $PID_IN 2>/dev/null
