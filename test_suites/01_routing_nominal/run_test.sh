#!/bin/bash
# NOME: 01_routing_nominal
# Run test ajustado

ROOT_BIN="../../test_suites/bin"

killall proxy_in proxy_out dummy_backend 2>/dev/null

# Subindo Binários Nativos Instantaneos
$ROOT_BIN/dummy_backend > /dev/null 2>&1 &
PID_DUMMY=$!

$ROOT_BIN/proxy_out > /dev/null 2>&1 &
PID_OUT=$!

$ROOT_BIN/proxy_in > /dev/null 2>&1 &
PID_IN=$!

sleep 0.5 # Apenas meio segundo pois não há build JIT

CURL_OUTPUT=$(curl -s http://127.0.0.1:5432/api/data)

if [[ "$CURL_OUTPUT" == *"Legacy_App"* ]]; then
    echo "01_routing_nominal: PASS" > ../reports/01_status.log
else
    echo "01_routing_nominal: FAIL" > ../reports/01_status.log
fi

kill $PID_IN $PID_OUT $PID_DUMMY 2>/dev/null
