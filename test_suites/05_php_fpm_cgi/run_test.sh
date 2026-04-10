#!/bin/bash
ROOT_BIN="../../test_suites/bin"

killall proxy_in proxy_out python3 2>/dev/null

# Usamos um FastCGI Simulator bobo em python no lugar do golang nativo
python3 -m http.server 8080 > /dev/null 2>&1 &
PID_PY=$!

$ROOT_BIN/proxy_out > /dev/null 2>&1 &
PID_OUT=$!
$ROOT_BIN/proxy_in > /dev/null 2>&1 &
PID_IN=$!
sleep 1 

echo "05_php_fpm_cgi: EM DESENVOLVIMENTO_CORE" > ../reports/05_status.log

kill $PID_IN $PID_OUT $PID_PY 2>/dev/null
