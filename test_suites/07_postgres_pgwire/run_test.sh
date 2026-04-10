#!/bin/bash
# Suite 07: PostgreSQL PGWire via proxy CROM (stateful)
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

# Mock PGWire na porta 8080 (backend)
python3 -c "
import socket
s=socket.socket()
s.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1)
s.bind(('127.0.0.1',8080))
s.listen(1)
while True:
    c,_=s.accept()
    d=c.recv(1024)
    if d: c.sendall(b'R\x00\x00\x00\x08\x00\x00\x00\x00')
    c.close()
" &
PID_PG=$!
sleep 0.3

echo "$CROM_TENANT_SEED" | $BIN/proxy_out &
PID_OUT=$!
sleep 0.3
echo "$CROM_TENANT_SEED" | $BIN/proxy_in &
PID_IN=$!
sleep 0.5

# Enviar startup via proxy_in (porta 5432) — NÃO direto no 9999
HEX_BYTES=$(printf '\x00\x00\x00\x08\x00\x03\x00\x00' | nc -w 2 127.0.0.1 5432 2>/dev/null | wc -c)

if [ "$HEX_BYTES" -gt "4" ] 2>/dev/null; then
    echo "07_postgres_pgwire: PASS (handshake PGWire transitou via proxy)" > "$REPORTS/07_status.log"
    echo "PASS"
else
    echo "07_postgres_pgwire: FAIL (bytes recebidos: $HEX_BYTES)" > "$REPORTS/07_status.log"
    echo "FAIL"
fi

kill $PID_PG $PID_OUT $PID_IN 2>/dev/null
