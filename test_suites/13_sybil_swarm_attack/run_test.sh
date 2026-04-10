#!/bin/bash
# Suite 13: Sybil Swarm Attack
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

$BIN/dummy_backend &
PID1=$!
sleep 0.3
$BIN/proxy_out &
PID2=$!
sleep 0.3

# 10 ghost nodes com timeout muito curto, sem wait
for i in $(seq 1 10); do
    (echo "SYBIL_${i}" | nc -w 1 127.0.0.1 9999) &
done
sleep 3

if kill -0 $PID1 2>/dev/null; then
    echo "13_sybil_swarm_attack: PASS (backend sobreviveu)" > "$REPORTS/13_status.log"
    echo "PASS"
else
    echo "13_sybil_swarm_attack: FAIL" > "$REPORTS/13_status.log"
    echo "FAIL"
fi

kill $PID1 $PID2 $(jobs -p) 2>/dev/null
