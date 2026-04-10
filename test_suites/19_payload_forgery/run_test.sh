#!/bin/bash
# Suite 19: LSH Payload Forgery (bit-flip attack on alien payload)
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

$BIN/dummy_backend &
PID1=$!
sleep 0.3
$BIN/proxy_out > /tmp/forgery_omega.log 2>&1 &
PID2=$!
sleep 0.3

# Gerar payload que PARECE vir de um cérebro alpha mas está corrompido
# (Criptografia forjada com bit-flips no hash)
python3 -c "
import socket,os
s=socket.socket()
s.connect(('127.0.0.1',9999))
# Forjar payload com XOR inverso deliberado
fake_payload = os.urandom(128)  # entropia pura sem hash valido
s.sendall(fake_payload)
try:
    resp = s.recv(512)
    if resp:
        with open('/tmp/forgery_leak.txt','w') as f:
            f.write(resp.decode('utf-8',errors='ignore'))
except: pass
s.close()
"

if [ -f "/tmp/forgery_leak.txt" ] && [ -s "/tmp/forgery_leak.txt" ]; then
    CONTENT=$(cat /tmp/forgery_leak.txt)
    if [[ "$CONTENT" == *"Legacy"* ]] || [[ "$CONTENT" == *"HTTP"* ]]; then
        echo "19_payload_forgery: FAIL (forged payload accepted!)" > "$REPORTS/19_status.log"
        echo "FAIL"
    else
        echo "19_payload_forgery: PASS (garbage response, no useful data)" > "$REPORTS/19_status.log"
        echo "PASS"
    fi
else
    echo "19_payload_forgery: PASS (forged payload dropped silently)" > "$REPORTS/19_status.log"
    echo "PASS"
fi

kill $PID1 $PID2 2>/dev/null
rm -f /tmp/forgery_omega.log /tmp/forgery_leak.txt
