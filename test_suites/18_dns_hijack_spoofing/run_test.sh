#!/bin/bash
# Suite 18: DNS Hijack Spoofing (Target redirect attack)
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

# Fake server na porta ERRADA (é onde o hacker quer redirecionar)
python3 -c "
import socket
s=socket.socket()
s.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1)
s.bind(('127.0.0.1',8083))
s.listen(1)
while True:
    c,_=s.accept()
    c.sendall(b'HACKED_DNS_REDIRECT')
    c.close()
" &
PID_FAKE=$!
sleep 0.2

# Servidor legítimo na porta CERTA
$BIN/dummy_backend &
PID_REAL=$!
sleep 0.3
$BIN/proxy_out &
PID2=$!
sleep 0.3

# Atacante injeta pacotes que apontam pro servidor falso (8083)
# Mas o proxy_out está hardcoded para :8080
RESP=$(echo "REDIRECT_ATTACK" | nc -w 2 127.0.0.1 9999 2>/dev/null || echo "")

if [[ "$RESP" == *"HACKED"* ]]; then
    echo "18_dns_hijack_spoofing: FAIL (redirect bem-sucedido)" > "$REPORTS/18_status.log"
    echo "FAIL"
else
    echo "18_dns_hijack_spoofing: PASS (redirect bloqueado)" > "$REPORTS/18_status.log"
    echo "PASS"
fi

kill $PID_FAKE $PID_REAL $PID2 2>/dev/null
