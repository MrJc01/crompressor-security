#!/bin/bash
# Suite 06: gRPC/Protobuf binary stream
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

# Mock gRPC: raw binary TCP server
python3 -c "
import socket
s=socket.socket()
s.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1)
s.bind(('127.0.0.1',8080))
s.listen(1)
while True:
    c,_=s.accept()
    c.sendall(b'\x00\x00\x00\x12grpc_status:ok!')
    c.close()
" &
PID_GRPC=$!
sleep 0.3

echo "$CROM_TENANT_SEED" | $BIN/proxy_out &
PID2=$!
sleep 0.3
echo "$CROM_TENANT_SEED" | $BIN/proxy_in &
PID3=$!
sleep 0.5

RESP=$(echo "proto_req" | nc -w 2 127.0.0.1 5432 2>/dev/null || echo "")

if [[ "$RESP" == *"grpc_status"* ]] || [[ "$RESP" == *"alienigena"* ]]; then
    echo "06_python_grpc: PASS (protobuf binary engolido)" > "$REPORTS/06_status.log"
    echo "PASS"
else
    echo "06_python_grpc: FAIL" > "$REPORTS/06_status.log"
    echo "FAIL"
fi

kill $PID_GRPC $PID2 $PID3 2>/dev/null
