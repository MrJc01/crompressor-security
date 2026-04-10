#!/bin/bash
# Suite 20: VFS File Descriptor Exhaustion
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

$BIN/dummy_backend &
PID1=$!
sleep 0.3
$BIN/proxy_out &
PID2=$!
sleep 0.3

# Tentar esgotar file descriptors abrindo centenas de conexões sem fechar
python3 -c "
import socket
conns=[]
for i in range(200):
    try:
        s=socket.socket()
        s.settimeout(0.5)
        s.connect(('127.0.0.1',9999))
        s.sendall(b'FD_EXHAUST_' + str(i).encode())
        conns.append(s)
    except: break
import time; time.sleep(1)
for c in conns:
    try: c.close()
    except: pass
"

# Backend ainda funciona?
if kill -0 $PID1 2>/dev/null && kill -0 $PID2 2>/dev/null; then
    echo "20_vfs_fd_exhaust: PASS (sobreviveu a 200 FDs abertos)" > "$REPORTS/20_status.log"
    echo "PASS"
else
    echo "20_vfs_fd_exhaust: FAIL (processo crashou)" > "$REPORTS/20_status.log"
    echo "FAIL"
fi

kill $PID1 $PID2 2>/dev/null
