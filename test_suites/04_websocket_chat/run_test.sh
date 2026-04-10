#!/bin/bash

# Simulando WebSocker (Long lived TCP connection) usando um netcat server rodando wsh/echo
ROOT_BIN="../../test_suites/bin"

killall proxy_in proxy_out nc 2>/dev/null

# Sobe um Dummy Node (Socket bruto com echo simulando ws handshake)
nc -l -p 8081 -k -c 'xargs -n1 echo' > /dev/null 2>&1 &
PID_NC=$!

# Modifica o Backend Target temporariamente
# (Apenas conceitual pois os binarios buildados hardcodaram 8080, 
# mas este teste exemplifica o check de falhas de timeout contínuas).
$ROOT_BIN/proxy_out > /dev/null 2>&1 &
PID_OUT=$!

$ROOT_BIN/proxy_in > /dev/null 2>&1 &
PID_IN=$!
sleep 0.5 

echo "04_websocket_chat: EM DESENVOLVIMENTO_CORE" > ../reports/04_status.log

kill $PID_IN $PID_OUT $PID_NC 2>/dev/null
