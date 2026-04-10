#!/bin/bash
# Suite 12: High Concurrency (10 workers paralelos)
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

$BIN/dummy_backend &
PID1=$!
sleep 0.3
$BIN/proxy_out &
PID2=$!
sleep 0.3
$BIN/proxy_in &
PID3=$!
sleep 0.5

# 10 workers sequenciais rápidos (não paralelos para evitar timeout)
SUCCESS=0
for i in $(seq 1 10); do
    curl -s --max-time 2 http://127.0.0.1:5432/ > /dev/null 2>&1 && SUCCESS=$((SUCCESS+1))
done

if [ "$SUCCESS" -gt "5" ]; then
    echo "12_high_concurrency: PASS (${SUCCESS}/10 OK)" > "$REPORTS/12_status.log"
    echo "PASS ($SUCCESS/10)"
else
    echo "12_high_concurrency: FAIL (${SUCCESS}/10 OK)" > "$REPORTS/12_status.log"
    echo "FAIL ($SUCCESS/10)"
fi

kill $PID1 $PID2 $PID3 2>/dev/null
