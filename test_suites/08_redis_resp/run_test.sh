#!/bin/bash
# Suite 08: Redis RESP via proxy CROM
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

# Mock Redis na porta 8080 (backend)
python3 -c "
import socket
s=socket.socket()
s.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1)
s.bind(('127.0.0.1',8080))
s.listen(1)
while True:
    c,_=s.accept()
    d=c.recv(1024)
    if d: c.sendall(b'+PONG\r\n')
    c.close()
" &
PID_RD=$!
sleep 0.3

$BIN/proxy_out &
PID_OUT=$!
sleep 0.3
$BIN/proxy_in &
PID_IN=$!
sleep 0.5

# Enviar via proxy (5432), NÃO direto na 9999
RESP=$(echo -e "PING\r\n" | nc -w 2 127.0.0.1 5432 2>/dev/null || echo "")

if [[ "$RESP" == *"PONG"* ]]; then
    echo "08_redis_resp: PASS (RESP PONG via proxy criptografado)" > "$REPORTS/08_status.log"
    echo "PASS"
else
    echo "08_redis_resp: FAIL (resp='$RESP')" > "$REPORTS/08_status.log"
    echo "FAIL"
fi

kill $PID_RD $PID_OUT $PID_IN 2>/dev/null
