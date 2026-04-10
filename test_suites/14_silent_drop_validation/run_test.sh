#!/bin/bash
# Suite 14: Silent Drop Validation
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

$BIN/dummy_backend &
PID1=$!
sleep 0.3
$BIN/proxy_out &
PID2=$!
sleep 0.3

# Enviar lixo sem assinatura CROM válida direto no Omega (porta 9999)
# e aferir se ele devolve ALGO pro atacante (deveria: NADA)
LEAK=$(echo "GARBAGE_INJECTION_ATTEMPT" | nc -w 2 127.0.0.1 9999 2>/dev/null || echo "")

if [ -z "$LEAK" ]; then
    echo "14_silent_drop_validation: PASS (zero leaks, silent drop)" > "$REPORTS/14_status.log"
    echo "PASS"
else
    echo "14_silent_drop_validation: FAIL (leaked: ${#LEAK} bytes)" > "$REPORTS/14_status.log"
    echo "FAIL"
fi

kill $PID1 $PID2 2>/dev/null
