#!/bin/bash
# Suite 04: WebSocket bidirecional via proxy CROM
# O nc echo agora vai atrás do proxy (porta 8080) e o client bate no proxy_in (5432)
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

# Backend echo server na porta 8080 (simula WebSocket echo)
python3 -c "
import socket
s=socket.socket()
s.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1)
s.bind(('127.0.0.1',8080))
s.listen(5)
while True:
    c,_=s.accept()
    try:
        d=c.recv(4096)
        if d: c.sendall(d)  # echo back
    except: pass
    c.close()
" &
PID_ECHO=$!
sleep 0.3

$BIN/proxy_out &
PID_OUT=$!
sleep 0.3
$BIN/proxy_in &
PID_IN=$!
sleep 0.5

RESP=$(echo "WS_HEARTBEAT_PING" | nc -w 3 127.0.0.1 5432 2>/dev/null || echo "")

if [[ "$RESP" == *"HEARTBEAT"* ]] || [[ "$RESP" == *"WS_"* ]]; then
    echo "04_websocket_chat: PASS (echo bidirecional funcionou)" > "$REPORTS/04_status.log"
    echo "PASS"
else
    echo "04_websocket_chat: FAIL (resp='$RESP')" > "$REPORTS/04_status.log"
    echo "FAIL"
fi

kill $PID_ECHO $PID_OUT $PID_IN 2>/dev/null
