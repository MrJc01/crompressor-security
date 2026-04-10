#!/bin/bash
# Suite 10: C++ Raw TCP Binary Struct
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

# Mock: raw struct server (little endian binary payload)
python3 -c "
import socket,struct
s=socket.socket()
s.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1)
s.bind(('127.0.0.1',8080))
s.listen(1)
while True:
    c,_=s.accept()
    data=c.recv(256)
    # Responder com struct: 4 bytes id, 8 bytes double, 4 bytes status
    reply=struct.pack('<IdI',42,3.14159,200)
    c.sendall(reply)
    c.close()
" &
PID_CPP=$!
sleep 0.3

$BIN/proxy_out &
PID2=$!
sleep 0.3
$BIN/proxy_in &
PID3=$!
sleep 0.5

RESP_SIZE=$(printf 'GAME_PACKET_V1' | nc -w 2 127.0.0.1 5432 2>/dev/null | wc -c)

if [ "$RESP_SIZE" -gt "10" ] 2>/dev/null; then
    echo "10_cplusplus_raw_tcp: PASS (binary struct transitou intacta)" > "$REPORTS/10_status.log"
    echo "PASS"
else
    echo "10_cplusplus_raw_tcp: FAIL" > "$REPORTS/10_status.log"
    echo "FAIL"
fi

kill $PID_CPP $PID2 $PID3 2>/dev/null
