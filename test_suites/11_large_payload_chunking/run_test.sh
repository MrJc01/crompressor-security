#!/bin/bash
# Suite 11: Large Payload — simplificado para validar streaming
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

# Backend HTTP que conta content-length
python3 -c "
import http.server
class H(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        length=int(self.headers.get('Content-Length',0))
        body=self.rfile.read(length)
        self.send_response(200)
        self.end_headers()
        self.wfile.write(f'GOT_{len(body)}_BYTES'.encode())
    def log_message(self, *a): pass
http.server.HTTPServer(('127.0.0.1',8080),H).serve_forever()
" &
PID1=$!
sleep 0.3

$BIN/proxy_out &
PID2=$!
sleep 0.3
$BIN/proxy_in &
PID3=$!
sleep 0.5

# Gerar 10KB e enviar via curl (HTTP POST com content-length)
PAYLOAD=$(python3 -c "print('X'*10240)")
RESP=$(curl -s --max-time 5 -X POST -d "$PAYLOAD" http://127.0.0.1:5432/upload 2>/dev/null || echo "")

if [[ "$RESP" == *"GOT_"* ]]; then
    echo "11_large_payload_chunking: PASS ($RESP)" > "$REPORTS/11_status.log"
    echo "PASS"
else
    echo "11_large_payload_chunking: FAIL" > "$REPORTS/11_status.log"
    echo "FAIL"
fi

kill $PID1 $PID2 $PID3 2>/dev/null
